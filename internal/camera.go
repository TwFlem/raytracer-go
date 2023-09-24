package internal

import (
	"fmt"
	"io"
	"math"
	"math/rand"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// CameraWorker workers that concurrently generate colors of pixels
type CameraWorker struct {
	rand *rand.Rand
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
	viewportU           Vec3[float32]
	viewportV           Vec3[float32]
	pixelDu             Vec3[float32]
	pixelDv             Vec3[float32]
	viewportUpperLeft   Vec3[float32]
	pixel00             Vec3[float32]
	center              Vec3[float32]
	lookAt              Vec3[float32]
	lookFrom            Vec3[float32]
	vup                 Vec3[float32]
	u                   Vec3[float32]
	v                   Vec3[float32]
	w                   Vec3[float32]
	defocusDiskU        Vec3[float32]
	defocusDiskV        Vec3[float32]
	fovRadians          float32
	workers             chan *CameraWorker
	once                sync.Once
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

func WithLookAt(lookAt Vec3[float32]) CameraOpt {
	return func(c *Camera) {
		c.lookAt = lookAt
	}
}

func WithLookFrom(lookFrom Vec3[float32]) CameraOpt {
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

func NewCamera(aspectRatio float32, imageWidth int, opts ...CameraOpt) *Camera {
	c := &Camera{
		aspectRatio:         aspectRatio,
		imageWidth:          float32(imageWidth),
		fovRadians:          float32(PiO2),
		samplesPerPixel:     100,
		bounceDepth:         50,
		focusDistance:       10,
		defocusAngleRadians: 0,
		lookAt:              NewVec3[float32](0, 0, 0),
		lookFrom:            NewVec3[float32](0, 0, -1),
		vup:                 NewVec3[float32](0, 1, 0),
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

		numWorkers := runtime.NumCPU() * 5
		c.workers = make(chan *CameraWorker, numWorkers)
		for i := 0; i < numWorkers; i++ {
			src := rand.NewSource(time.Now().UnixNano())
			randCtx := rand.New(src)
			c.workers <- &CameraWorker{
				rand: randCtx,
			}
		}

	})
}

func (c *Camera) Render(world *World, writer io.Writer) error {
	w := int(c.imageWidth)
	h := int(c.imageHeight)
	ppm := []string{
		"P3",
		strconv.Itoa(w) + " " + strconv.Itoa(h),
		"255",
	}
	for j := 0; j < h; j++ {
		fmt.Printf("coloring line %d out of %d\n", j+1, h)
		for i := 0; i < w; i++ {
			sampleOut := c.GetPixelColor(world, i, j)

			sample := <-sampleOut
			ppm = append(ppm, sample.String())
		}
	}

	_, err := io.WriteString(writer, strings.Join(ppm, "\n"))
	return err
}

func (c *Camera) GetPixelColor(world *World, i, j int) <-chan Vec3[float32] {
	out := make(chan Vec3[float32])
	go func() {
		defer close(out)
		sample := NewVec3Zero[float32]()
		samplesOut := make(chan Vec3[float32])
		go func() {
			var wg sync.WaitGroup
			wg.Add(c.samplesPerPixel)
			for k := 0; k < c.samplesPerPixel; k++ {
				cameraWorker := <-c.workers
				go func(cw *CameraWorker) {
					ray := c.GetRay(cw, i, j)
					samplesOut <- ray.GetColor(world, c.bounceDepth)
					c.workers <- cw
					wg.Done()
				}(cameraWorker)
			}
			wg.Wait()
			close(samplesOut)
		}()
		for s := range samplesOut {
			sample.Add(s)
		}
		sample.Scale(1.0 / float32(c.samplesPerPixel))
		sample.ToGamma2()
		sample.ToRGB()
		out <- sample
	}()

	return out

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

func (c *Camera) sampleUnitSquare(randCtx *rand.Rand) Vec3[float32] {
	dx := -0.5 + randCtx.Float32()
	dy := -0.5 + randCtx.Float32()

	du := c.pixelDu.Cpy()
	du.Scale(dx)
	dv := c.pixelDv.Cpy()
	dv.Scale(dy)

	return Add(du, dv)
}
