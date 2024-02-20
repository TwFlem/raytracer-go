package internal

import (
	"context"
	"io"
	"math"
	"math/rand"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/TwFlem/pipe/stage"
)

// CameraWorker workers that concurrently generate colors of pixels
type CameraWorker struct {
	rand             *rand.Rand
	emissionStack    []Color
	attenuationStack []Color
}

type Camera struct {
	aspectRatio         float32
	imageWidth          float32
	imageHeight         float32
	viewportHeight      float32
	viewportWidth       float32
	samplesPerPixel     int
	bounceDepth         int
	defocusAngleRadians float32
	focusDistance       float32
	viewportU           Vec3
	viewportV           Vec3
	pixelDu             Vec3
	pixelDv             Vec3
	viewportUpperLeft   Vec3
	pixel00             Vec3
	center              Vec3
	lookAt              Vec3
	lookFrom            Vec3
	vup                 Vec3
	u                   Vec3
	v                   Vec3
	w                   Vec3
	defocusDiskU        Vec3
	defocusDiskV        Vec3
	fovRadians          float32
	once                sync.Once
	background          Color
}

type CameraOpt func(*Camera)

func WithSamplesPerPixel(samples int) CameraOpt {
	return func(c *Camera) {
		c.samplesPerPixel = samples
	}
}

func WithMaxRayDepth(depth int) CameraOpt {
	return func(c *Camera) {
		c.bounceDepth = depth
	}
}

func WithFOVDegrees(fov float32) CameraOpt {
	return func(c *Camera) {
		c.fovRadians = ToRadians(fov)
	}
}

func WithLookAt(lookAt Vec3) CameraOpt {
	return func(c *Camera) {
		c.lookAt = lookAt
	}
}

func WithLookFrom(lookFrom Vec3) CameraOpt {
	return func(c *Camera) {
		c.lookFrom = lookFrom
	}
}

func WithDefocusAngleDegrees(degrees float32) CameraOpt {
	return func(c *Camera) {
		c.defocusAngleRadians = ToRadians(degrees)
	}
}

func WithFocusDist(dist float32) CameraOpt {
	return func(c *Camera) {
		c.focusDistance = dist
	}
}

func WithBackgroundColor(color Color) CameraOpt {
	return func(c *Camera) {
		c.background = color
	}
}

func NewCamera(aspectRatio float32, imageWidth int, opts ...CameraOpt) *Camera {
	c := &Camera{
		aspectRatio:         aspectRatio,
		imageWidth:          float32(imageWidth),
		fovRadians:          float32(PiO2),
		samplesPerPixel:     100,
		bounceDepth:         50,
		focusDistance:       10,
		defocusAngleRadians: 0,
		lookAt:              NewVec3(0, 0, 0),
		lookFrom:            NewVec3(0, 0, -1),
		vup:                 NewVec3(0, 1, 0),
		background:          NewVec3(0, 0, 0),
	}

	for _, fn := range opts {
		fn(c)
	}

	c.init()

	return c
}

func (c *Camera) init() {
	c.once.Do(func() {
		c.center = c.lookFrom.Cpy()

		dist := Sub(c.lookFrom, c.lookAt)

		h := float32(math.Tan(float64(c.fovRadians / 2.0)))
		c.viewportHeight = 2.0 * h * c.focusDistance

		c.imageHeight = float32(math.Floor(float64(c.imageWidth)) / float64(c.aspectRatio))
		if c.imageHeight < 1 {
			c.imageHeight = 1
		}
		c.viewportWidth = c.viewportHeight * (float32(c.imageWidth) / float32(c.imageHeight))

		c.w = Unit(dist)
		c.u = Unit(Cross(c.vup, c.w))
		c.v = Cross(c.w, c.u)

		c.viewportU = Scale(c.u, c.viewportWidth)
		c.viewportV = Scale(c.v, -c.viewportHeight)

		c.pixelDu = c.viewportU.Cpy()
		c.pixelDu.Scale(1 / float32(c.imageWidth))
		c.pixelDv = c.viewportV.Cpy()
		c.pixelDv.Scale(1 / c.imageHeight)

		c.viewportUpperLeft = c.center.Cpy()
		c.viewportUpperLeft.Sub(Scale(c.w, c.focusDistance))
		c.viewportUpperLeft.Sub(Scale(c.viewportU, 0.5))
		c.viewportUpperLeft.Sub(Scale(c.viewportV, 0.5))

		c.pixel00 = c.viewportUpperLeft.Cpy()
		c.pixel00.Add(Scale(Add(c.pixelDu, c.pixelDv), 0.5))

		defocusRadius := c.focusDistance * float32(math.Tan(float64(c.defocusAngleRadians/2.0)))
		c.defocusDiskU = Scale(c.u, defocusRadius)
		c.defocusDiskV = Scale(c.v, defocusRadius)
	})
}

func (c *Camera) Render(world Hittable, writer io.Writer) error {
	w := int(c.imageWidth)
	h := int(c.imageHeight)
	ppm := []string{
		"P3",
		strconv.Itoa(w) + " " + strconv.Itoa(h),
		"255\n",
	}
	_, err := io.WriteString(writer, strings.Join(ppm, "\n"))
	if err != nil {
		return err
	}

	ctx := context.Background()

	numThreads := runtime.NumCPU()
	colorsOut := make([]<-chan string, numThreads)
	area := h * w
	for t := 0; t < numThreads; t++ {
		threadInput := make(chan string)
		go func(threadNum int, in chan string, innerWorld Hittable) {
			defer close(in)
			src := rand.NewSource(time.Now().UnixNano())
			randCtx := rand.New(src)
			cw := CameraWorker{
				rand:             randCtx,
				attenuationStack: make([]Color, c.bounceDepth),
				emissionStack:    make([]Color, c.bounceDepth),
			}
			for i := threadNum; i < area; i = i + numThreads {
				x := i % w
				y := i / h
				col := c.GetPixelColor(&cw, innerWorld, x, y)
				colorVec := col.GetColor()
				colorVec.ToGamma2()
				colorVec.ToRGB()
				in <- colorVec.String()
			}
		}(t, threadInput, world)
		colorsOut[t] = threadInput
	}

	orderedPixelsOut := Interleave(ctx.Done(), colorsOut...)
	chunksOut := stage.Agg(ctx.Done(), orderedPixelsOut, 5000)
	bufChunksOut := stage.Buf(ctx.Done(), chunksOut, 2)
	chunkResultOut := c.StartChunkRenderer(writer, bufChunksOut)

	res := <-chunkResultOut
	return res.err
}

// Bridge takes a channel of channels as input and streams out the values of each of those
// channels on a single output. This is similar
// to how FanIn works except that with the channels of channel input, order is implicitly
// maintained.
func Interleave[T any](done <-chan struct{}, ins ...<-chan T) <-chan T {
	out := make(chan T)
	go func() {
		defer close(out)
		closed := make([]bool, len(ins))
		closedCount := 0
		for {
			for i, in := range ins {
				select {
				case <-done:
					return
				case v, ok := <-in:
					if !ok {
						if !closed[i] {
							closed[i] = true
							closedCount++
						}
						continue
					}
					out <- v
				}
			}
			if closedCount >= len(ins) {
				return
			}
		}
	}()
	return out
}

type RenderChunkResult struct {
	err error
}

func (c *Camera) StartChunkRenderer(writer io.Writer, chunksIn <-chan []string) <-chan RenderChunkResult {
	out := make(chan RenderChunkResult)
	go func() {
		defer close(out)
		for c := range chunksIn {
			_, err := io.WriteString(writer, strings.Join(c, "\n")+"\n")
			if err != nil {
				out <- RenderChunkResult{err: err}
				break
			}
		}
		out <- RenderChunkResult{err: nil}
	}()
	return out

}

// func (c *Camera) GetPixelColor(cw *CameraWorker, world Hittable, i, j int) Color {
// 	sample := NewVec3Zero()
// 	for k := 0; k < c.samplesPerPixel; k++ {
// 		ray := c.GetRay(cw, i, j)
// 		s := ray.GetColor(world, c.background, c.bounceDepth).GetColor()
//
// 		sample.Add(s)
// 	}
// 	sample.Scale(1.0 / float32(c.samplesPerPixel))
// 	return sample
// }

// TODO: can we allocate a fix sized slice to CameraWorker an reuse it over and over again?
func (c *Camera) GetPixelColor(cw *CameraWorker, world Hittable, pixelX, pixelY int) Color {
	sample := NewVec3Zero()
	for i := 0; i < c.samplesPerPixel; i++ {
		currRay := c.GetRay(cw, pixelX, pixelY)

		var currRayColor Color
		stopped := 0
		for j := 0; j < c.bounceDepth; j++ {
			hitInfo, ok := world.Hit(currRay, Interval{0.001, float32(math.Inf(1))})
			if !ok {
				currRayColor = c.background
				stopped = j
				break
			}

			colorFromEmission := hitInfo.material.Emit(hitInfo.u, hitInfo.v, hitInfo.point)
			scatterInfo, didScatter := hitInfo.material.Scatter(currRay, hitInfo)

			if !didScatter {
				currRayColor = colorFromEmission
				stopped = j
				break
			}

			cw.attenuationStack[j] = scatterInfo.attenuation
			cw.emissionStack[j] = colorFromEmission
			currRay = &scatterInfo.ray
		}

		if currRayColor == nil {
			continue
		}

		currRaySample := currRayColor.GetColor()
		for j := 0; j < stopped; j++ {
			currRaySample.Mul(cw.attenuationStack[stopped-1-j].GetColor())
			currRaySample.Add(cw.emissionStack[stopped-1-j].GetColor())
		}

		sample.Add(currRaySample)
	}
	sample.Scale(1.0 / float32(c.samplesPerPixel))
	return sample
}

func (c *Camera) GetRay(cw *CameraWorker, i, j int) *Ray {
	duOffset := c.pixelDu.Cpy()
	duOffset.Scale(float32(i))

	dvOffset := c.pixelDv.Cpy()
	dvOffset.Scale(float32(j))

	pixelCenter := c.pixel00.Cpy()
	pixelCenter.Add(duOffset)
	pixelCenter.Add(dvOffset)
	pixelCenter.Add(c.sampleUnitSquare(cw.rand))

	discSample := NewVec3RandInUnitDisk(cw.rand)
	origin := c.center.Cpy()
	if c.defocusAngleRadians > 0 {
		origin = Add(c.center, Add(Scale(c.defocusDiskU, discSample.X), Scale(c.defocusDiskV, discSample.Y)))
	}

	rayDir := pixelCenter.Cpy()
	rayDir.Sub(origin)

	return NewRay(origin, rayDir, cw.rand)
}

func (c *Camera) sampleUnitSquare(randCtx *rand.Rand) Vec3 {
	dx := -0.5 + randCtx.Float32()
	dy := -0.5 + randCtx.Float32()

	du := c.pixelDu.Cpy()
	du.Scale(dx)
	dv := c.pixelDv.Cpy()
	dv.Scale(dy)

	return Add(du, dv)
}
