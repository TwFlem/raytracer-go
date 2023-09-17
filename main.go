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

	matGround := internal.NewLambertian(internal.NewVec3[float32](0.8, 0.8, 0))
	matCenter := internal.NewLambertian(internal.NewVec3[float32](0.1, 0.2, 0.5))
	matLeft := internal.NewDielectric(1.5)
	matRight := internal.NewMetal(internal.NewVec3[float32](0.8, 0.6, 0.2), 0)

	hittables := []internal.Hittable{
		&internal.Sphere{
			Center:   internal.NewVec3[float32](0, 0, -1),
			Radius:   0.5,
			Material: &matCenter,
		},
		&internal.Sphere{
			Center:   internal.NewVec3[float32](-1, 0, -1),
			Radius:   0.5,
			Material: &matLeft,
		},
		&internal.Sphere{
			Center:   internal.NewVec3[float32](1, 0, -1),
			Radius:   0.5,
			Material: &matRight,
		},
		&internal.Sphere{
			Center:   internal.NewVec3[float32](0, -100.5, -1),
			Radius:   100,
			Material: &matGround,
		},
	}

	world := internal.NewWorld(hittables)

	err = camera.Render(world, f)
	if err != nil {
		panic(err)
	}
}
