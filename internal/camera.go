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
	center            Vec3[float32]
	viewportU         Vec3[float32]
	viewportV         Vec3[float32]
	viewportUHalf     Vec3[float32]
	viewportVHalf     Vec3[float32]
	pixelDu           Vec3[float32]
	pixelDv           Vec3[float32]
	viewportUpperLeft Vec3[float32]
	pixel00           Vec3[float32]
	once              sync.Once
}

func NewCamera(aspectRatio float32, imageWidth int) *Camera {
	c := &Camera{
		aspectRatio: aspectRatio,
		imageWidth:  float32(imageWidth),
	}

	c.init()

	return c
}

func (c *Camera) init() {
	c.once.Do(func() {
		c.samplesPerPixel = 100.0
		c.bounceDepth = 50.0
		c.viewportHeight = float32(2.0)
		c.imageHeight = float32(math.Floor(float64(c.imageWidth)) / float64(c.aspectRatio))
		if c.imageHeight < 1 {
			c.imageHeight = 1
		}
		c.viewportWidth = c.viewportHeight * (float32(c.imageWidth) / float32(c.imageHeight))

		c.focalLength = float32(1.0)
		c.center = NewVec3[float32](0, 0, 0)

		c.viewportU = NewVec3[float32](c.viewportWidth, 0, 0)
		c.viewportV = NewVec3[float32](0, -c.viewportHeight, 0)

		c.pixelDu = c.viewportU.Cpy()
		c.pixelDu.Scale(1 / float32(c.imageWidth))
		c.pixelDv = c.viewportV.Cpy()
		c.pixelDv.Scale(1 / c.imageHeight)

		c.viewportUHalf = c.viewportU.Cpy()
		c.viewportUHalf.Scale(0.5)
		c.viewportVHalf = c.viewportV.Cpy()
		c.viewportVHalf.Scale(0.5)

		c.viewportUpperLeft = c.center.Cpy()
		c.viewportUpperLeft.Sub(NewVec3[float32](0, 0, c.focalLength))
		c.viewportUpperLeft.Sub(c.viewportUHalf)
		c.viewportUpperLeft.Sub(c.viewportVHalf)

		c.pixel00 = NewVec3[float32](0, 0, 0)
		c.pixel00.Add(c.pixelDu)
		c.pixel00.Add(c.pixelDv)
		c.pixel00.Scale(0.5)
		c.pixel00.Add(c.viewportUpperLeft)
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
