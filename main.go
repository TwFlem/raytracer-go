package main

import (
	"fmt"
	"math/rand"
	"os"
	"raytracer/internal"
	"runtime/pprof"
	"time"
)

const profileEnabled = true

func main() {
	now := time.Now()
	camera := internal.NewCamera(
		16.0/9.0,
		400.0,
		internal.WithSamplesPerPixel(500),
		internal.WithMaxRayDepth(50),
		internal.WithLookFrom(internal.NewVec3(13, 2, 3)),
		internal.WithLookAt(internal.NewVec3(0, 0, 0)),
		internal.WithFOVDegrees(20),
		internal.WithDefocusAngleDegrees(0.6),
		internal.WithFocusDist(10),
	)

	var cpuPprofF *os.File
	var MemPprofF *os.File
	var pprofErr error
	if profileEnabled {
		cpuPprofF, pprofErr = internal.Overwrite("out/cpu.pprof")
		if pprofErr != nil {
			panic(pprofErr)
		}
		defer cpuPprofF.Close()

		MemPprofF, pprofErr = internal.Overwrite("out/mem.pprof")
		if pprofErr != nil {
			panic(pprofErr)
		}
		defer MemPprofF.Close()

	}

	f, err := internal.Overwrite("out/img.ppm")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	world := internal.NewWorld()

	matGround := internal.NewLambertian(internal.NewVec3(0.5, 0.5, 0.5))
	world.Add(internal.NewSphere(internal.NewVec3(0, -1000, 0), 1000, &matGround))

	src := rand.NewSource(time.Now().Unix())
	randCtx := rand.New(src)
	p := internal.NewVec3(4, 0.2, 0)
	for i := -11; i < 11; i++ {
		for j := -11; j < 11; j++ {
			matPer := rand.Float32()
			center := internal.NewVec3(float32(i)+0.9*rand.Float32(), 0.2, float32(j)+0.9*rand.Float32())

			dist := internal.Sub(center, p)
			ln := dist.Len()
			if ln > 0.9 {
				var sphereMat internal.Material
				if matPer < 0.8 {
					albedo := internal.Mul(internal.NewVec3Rand32(randCtx), internal.NewVec3Rand32(randCtx))
					mat := internal.NewLambertian(albedo)
					sphereMat = &mat
				} else if matPer < 0.95 {
					albedo := internal.NewVec3RandRange32(randCtx, 0.5, 1)
					fuzz := internal.RandF32N(randCtx, 0, 0.5)
					mat := internal.NewMetal(albedo, fuzz)
					sphereMat = &mat
				} else {
					mat := internal.NewDielectric(1.5)
					sphereMat = &mat
				}
				world.Add(internal.NewSphere(center, 0.2, sphereMat))
			}

		}
	}

	m1 := internal.NewDielectric(1.5)
	world.Add(internal.NewSphere(internal.NewVec3(0, 1, 0), 1, &m1))

	m2 := internal.NewLambertian(internal.NewVec3(0.4, 0.2, 0.1))
	world.Add(internal.NewSphere(internal.NewVec3(-4, 1, 0), 1, &m2))

	m3 := internal.NewMetal(internal.NewVec3(0.7, 0.6, 0.5), 0)
	world.Add(internal.NewSphere(internal.NewVec3(4, 1, 0), 1, &m3))

	if profileEnabled {
		pprof.StartCPUProfile(cpuPprofF)
	}

	worldTree := internal.NewBVHFromWorld(world)
	err = camera.Render(worldTree, f)
	if err != nil {
		panic(err)
	}
	if profileEnabled {
		pprof.StopCPUProfile()
		pprof.WriteHeapProfile(MemPprofF)
	}
	fmt.Println("Finished in: " + time.Since(now).String())
}
