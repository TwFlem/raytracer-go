package internal

import (
	"fmt"
	"io"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"sync"
)

type Camera struct {
	aspectRatio       float32
	imageWidth        float32
	imageHeight       float32
	viewportHeight    float32
	viewportWidth     float32
	focalLength       float32
	samplesPerPixel   int
	bounceDepth       int
	viewportU         Vec3[float32]
	viewportV         Vec3[float32]
	pixelDu           Vec3[float32]
	pixelDv           Vec3[float32]
	viewportUpperLeft Vec3[float32]
	pixel00           Vec3[float32]
	center            Vec3[float32]
	lookAt            Vec3[float32]
	lookFrom          Vec3[float32]
	vup               Vec3[float32]
	u                 Vec3[float32]
	v                 Vec3[float32]
	w                 Vec3[float32]
	fovRadians        float64
	once              sync.Once
}

type CameraOpt func(*Camera)

func WithFOV(fov float64) CameraOpt {
	return func(c *Camera) {
		c.fovRadians = fov
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

func NewCamera(aspectRatio float32, imageWidth int, opts ...CameraOpt) *Camera {
	c := &Camera{
		aspectRatio:     aspectRatio,
		imageWidth:      float32(imageWidth),
		fovRadians:      PiO2,
		samplesPerPixel: 100,
		bounceDepth:     50,
		lookAt:          NewVec3[float32](0, 0, 0),
		lookFrom:        NewVec3[float32](0, 0, -1),
		vup:             NewVec3[float32](0, 1, 0),
	}

	for _, fn := range opts {
		fn(c)
	}

	c.init()

	return c
}

func (c *Camera) init() {
	c.once.Do(func() {
		c.center = c.lookAt.Cpy()

		dist := Sub(c.lookFrom, c.lookAt)
		c.focalLength = dist.Len()

		h := float32(math.Tan(c.fovRadians / 2.0))
		c.viewportHeight = 2.0 * h * c.focalLength

		c.imageHeight = float32(math.Floor(float64(c.imageWidth)) / float64(c.aspectRatio))
		if c.imageHeight < 1 {
			c.imageHeight = 1
		}
		c.viewportWidth = c.viewportHeight * (float32(c.imageWidth) / float32(c.imageHeight))

		c.w = Unit(Scale(dist, -1))
		c.u = Unit(Cross(c.vup, c.w))
		c.v = Cross(c.w, c.u)

		c.viewportU = Scale(c.u, c.viewportWidth)
		c.viewportV = Scale(c.v, -c.viewportHeight)

		c.pixelDu = c.viewportU.Cpy()
		c.pixelDu.Scale(1 / float32(c.imageWidth))
		c.pixelDv = c.viewportV.Cpy()
		c.pixelDv.Scale(1 / c.imageHeight)

		c.viewportUpperLeft = c.center.Cpy()
		c.viewportUpperLeft.Sub(Scale(c.w, c.focalLength))
		c.viewportUpperLeft.Sub(Scale(c.viewportU, 0.5))
		c.viewportUpperLeft.Sub(Scale(c.viewportV, 0.5))

		c.pixel00 = c.viewportUpperLeft.Cpy()
		c.pixel00.Add(Scale(Add(c.pixelDu, c.pixelDv), 0.5))
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
		fmt.Printf("computing %d of %d\n", j, h-1)
		for i := 0; i < w; i++ {
			sample := NewVec3Zero[float32]()
			for k := 0; k < c.samplesPerPixel; k++ {
				ray := c.GetRay(i, j)
				sample.Add(ray.GetColor(world, c.bounceDepth))
			}
			sample.Scale(1.0 / float32(c.samplesPerPixel))
			sample.ToGamma2()
			sample.ToRGB()

			ppm = append(ppm, sample.String())
		}
	}

	_, err := io.WriteString(writer, strings.Join(ppm, "\n"))
	return err
}

func (c *Camera) GetRay(i, j int) *Ray {
	duOffset := c.pixelDu.Cpy()
	duOffset.Scale(float32(i))

	dvOffset := c.pixelDv.Cpy()
	dvOffset.Scale(float32(j))

	pixelCenter := c.pixel00.Cpy()
	pixelCenter.Add(duOffset)
	pixelCenter.Add(dvOffset)
	pixelCenter.Add(c.sampleUnitSquare())

	rayDir := pixelCenter.Cpy()
	rayDir.Sub(c.center)

	return NewRay(c.center, rayDir)
}

func (c *Camera) sampleUnitSquare() Vec3[float32] {
	dx := -0.5 + rand.Float32()
	dy := -0.5 + rand.Float32()

	du := c.pixelDu.Cpy()
	du.Scale(dx)
	dv := c.pixelDv.Cpy()
	dv.Scale(dy)

	return Add(du, dv)
}
