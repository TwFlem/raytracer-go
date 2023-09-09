package main

import (
	"fmt"
	"io"
	"math"
	"raytracer/internal"
	"strconv"
	"strings"
)

func main() {
	aspectRatio := float32(16.0 / 9.0)
	imageWidth := float32(400)
	viewportHeight := float32(2.0)
	imageHeight := float32(math.Floor(float64(imageWidth)) / float64(aspectRatio))
	if imageHeight < 1 {
		imageHeight = 1
	}
	viewportWidth := viewportHeight * (float32(imageWidth) / float32(imageHeight))

	focalLength := float32(1.0)
	cameraCenter := internal.NewVec3[float32](0, 0, 0)

	viewportU := internal.NewVec3[float32](viewportWidth, 0, 0)
	viewportV := internal.NewVec3[float32](0, -viewportHeight, 0)

	pixelDu := viewportU.Cpy()
	pixelDu.Scale(1 / imageWidth)
	pixelDv := viewportV.Cpy()
	pixelDv.Scale(1 / imageHeight)

	viewportUHalf := viewportU.Cpy()
	viewportUHalf.Scale(0.5)
	viewportVHalf := viewportV.Cpy()
	viewportVHalf.Scale(0.5)

	viewportUpperLeft := cameraCenter.Cpy()
	viewportUpperLeft.Sub(internal.NewVec3[float32](0, 0, focalLength))
	viewportUpperLeft.Sub(viewportUHalf)
	viewportUpperLeft.Sub(viewportVHalf)

	pixel00 := internal.NewVec3[float32](0, 0, 0)
	pixel00.Add(pixelDu)
	pixel00.Add(pixelDv)
	pixel00.Scale(0.5)
	pixel00.Add(viewportUpperLeft)

	f, err := internal.Overwrite("out/img.ppm")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	w := int(imageWidth)
	h := int(imageHeight)
	ppm := []string{
		"P3",
		strconv.Itoa(w) + " " + strconv.Itoa(h),
		"255",
	}
	for j := 0; j < h; j++ {
		fmt.Printf("computing %d of %d\n", j, h-1)
		for i := 0; i < w; i++ {
			duOffset := pixelDu.Cpy()
			duOffset.Scale(float32(i))

			dvOffset := pixelDv.Cpy()
			dvOffset.Scale(float32(j))

			pixelCenter := pixel00.Cpy()
			pixelCenter.Add(duOffset)
			pixelCenter.Add(dvOffset)

			rayDir := pixelCenter.Cpy()
			rayDir.Sub(cameraCenter)

			ray := internal.NewRay(cameraCenter, rayDir)
			color := ray.GetColor()
			color.ToRGB()
			colorStr := color.String()

			ppm = append(ppm, colorStr)
		}
	}

	_, err = io.WriteString(f, strings.Join(ppm, "\n"))
	if err != nil {
		panic(err)
	}
}
