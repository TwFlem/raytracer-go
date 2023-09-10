package main

import (
	"raytracer/internal"
)

func main() {
	camera := internal.NewCamera(16.0/9.0, 400.0)

	f, err := internal.Overwrite("out/img.ppm")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	hittables := []internal.Hittable{
		&internal.Sphere{
			Center: internal.NewVec3[float32](0, 0, -1),
			Radius: 0.5,
		},
		&internal.Sphere{
			Center: internal.NewVec3[float32](0, -100.5, -1),
			Radius: 100,
		},
	}

	world := internal.NewWorld(hittables)

	err = camera.Render(world, f)
	if err != nil {
		panic(err)
	}
}
