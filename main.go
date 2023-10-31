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
const (
	randSphereScene      = 0
	earthScene           = 1
	perlinDemoScene      = 2
	quadDemoScene        = 3
	simpleLightDemoScene = 4
	cornellBoxDemoScene  = 5
)

func main() {
	now := time.Now()

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

	if err != nil {
		panic(err)
	}
	if profileEnabled {
		pprof.StartCPUProfile(cpuPprofF)
	}
	scene := cornellBoxDemoScene
	switch scene {
	case randSphereScene:
		err = randSpheres(f)
	case earthScene:
		err = earth(f)
	case perlinDemoScene:
		err = perlinDemo(f)
	case quadDemoScene:
		err = quadDemo(f)
	case simpleLightDemoScene:
		err = simpleLightDemo(f)
	case cornellBoxDemoScene:
		err = cornellBox(f)
	}
	if profileEnabled {
		pprof.StopCPUProfile()
		pprof.WriteHeapProfile(MemPprofF)
	}
	if err != nil {
		panic(err)
	}
	fmt.Println("Finished in: " + time.Since(now).String())
}

func earth(f *os.File) error {
	camera := internal.NewCamera(
		16.0/9.0,
		400.0,
		internal.WithSamplesPerPixel(100),
		internal.WithMaxRayDepth(50),
		internal.WithLookFrom(internal.NewVec3(0, 0, 12)),
		internal.WithLookAt(internal.NewVec3(0, 0, 0)),
		internal.WithFOVDegrees(20),
		internal.WithDefocusAngleDegrees(0),
		internal.WithBackgroundColor(internal.NewVec3(0.7, 0.8, 1)),
	)
	world := internal.NewWorld()

	earthImg, err := internal.LoadJPEG("textures/earthmap.jpg")
	if err != nil {
		panic(err)
	}
	earthTex := internal.NewImageTexture(earthImg)
	mat := internal.NewLambertian(&earthTex)
	world.Add(internal.NewSphere(internal.NewVec3(0, 0, 0), 2, &mat))

	worldTree := internal.NewBVHFromWorld(world)
	return camera.Render(worldTree, f)
}

func perlinDemo(f *os.File) error {
	camera := internal.NewCamera(
		16.0/9.0,
		400.0,
		internal.WithSamplesPerPixel(100),
		internal.WithMaxRayDepth(50),
		internal.WithLookFrom(internal.NewVec3(13, 2, 3)),
		internal.WithLookAt(internal.NewVec3(0, 0, 0)),
		internal.WithFOVDegrees(20),
		internal.WithDefocusAngleDegrees(0),
		internal.WithBackgroundColor(internal.NewVec3(0.7, 0.8, 1)),
	)
	world := internal.NewWorld()

	src := rand.NewSource(time.Now().Unix())
	randCtx := rand.New(src)

	perlinTex := internal.NewNoiseTexture(randCtx, 4)
	mat := internal.NewLambertian(&perlinTex)
	world.Add(internal.NewSphere(internal.NewVec3(0, -1000, 0), 1000, &mat))
	world.Add(internal.NewSphere(internal.NewVec3(0, 2, 0), 2, &mat))

	worldTree := internal.NewBVHFromWorld(world)
	return camera.Render(worldTree, f)
}

func quadDemo(f *os.File) error {
	camera := internal.NewCamera(
		16.0/9.0,
		400.0,
		internal.WithSamplesPerPixel(100),
		internal.WithMaxRayDepth(50),
		internal.WithLookFrom(internal.NewVec3(0, 0, 9)),
		internal.WithLookAt(internal.NewVec3(0, 0, 0)),
		internal.WithFOVDegrees(80),
		internal.WithDefocusAngleDegrees(0),
		internal.WithBackgroundColor(internal.NewVec3(0.7, 0.8, 1)),
	)
	world := internal.NewWorld()

	leftRed := internal.NewLambertian(internal.NewSolidColor(1, 0.2, 0.2))
	backGreen := internal.NewLambertian(internal.NewSolidColor(0.2, 1, 0.2))
	rightBlue := internal.NewLambertian(internal.NewSolidColor(0.2, 0.2, 1))
	upperOrange := internal.NewLambertian(internal.NewSolidColor(1, 0.5, 0))
	lowerTeal := internal.NewLambertian(internal.NewSolidColor(0.2, 0.8, 0.8))

	world.Add(internal.NewQuad(internal.NewVec3(-3, -2, 5), internal.NewVec3(0, 0, -4), internal.NewVec3(0, 4, 0), &leftRed))
	world.Add(internal.NewQuad(internal.NewVec3(-2, -2, 0), internal.NewVec3(4, 0, 0), internal.NewVec3(0, 4, 0), &backGreen))
	world.Add(internal.NewQuad(internal.NewVec3(3, -2, 1), internal.NewVec3(0, 0, 4), internal.NewVec3(0, 4, 0), &rightBlue))
	world.Add(internal.NewQuad(internal.NewVec3(-2, 3, 1), internal.NewVec3(4, 0, 0), internal.NewVec3(0, 0, 4), &upperOrange))
	world.Add(internal.NewQuad(internal.NewVec3(-2, -3, 5), internal.NewVec3(4, 0, 0), internal.NewVec3(0, 0, -4), &lowerTeal))

	worldTree := internal.NewBVHFromWorld(world)
	return camera.Render(worldTree, f)
}

func simpleLightDemo(f *os.File) error {
	camera := internal.NewCamera(
		16.0/9.0,
		400.0,
		internal.WithSamplesPerPixel(500),
		internal.WithMaxRayDepth(50),
		internal.WithLookFrom(internal.NewVec3(26, 3, 6)),
		internal.WithLookAt(internal.NewVec3(0, 2, 0)),
		internal.WithFOVDegrees(20),
		internal.WithDefocusAngleDegrees(0),
		internal.WithBackgroundColor(internal.NewVec3Zero()),
	)
	world := internal.NewWorld()

	src := rand.NewSource(time.Now().Unix())
	randCtx := rand.New(src)

	perlinTex := internal.NewNoiseTexture(randCtx, 4)
	mat := internal.NewLambertian(&perlinTex)
	world.Add(internal.NewSphere(internal.NewVec3(0, -1000, 0), 1000, &mat))
	world.Add(internal.NewSphere(internal.NewVec3(0, 2, 0), 2, &mat))

	red := internal.NewLambertian(internal.NewSolidColor(1, 0, 0))
	world.Add(internal.NewSphere(internal.NewVec3(-4, 2, 4), 2, &red))

	diffLight := internal.NewDiffuseLight(internal.NewSolidColor(4, 4, 4))
	world.Add(internal.NewSphere(internal.NewVec3(0, 7, 0), 2, &diffLight))

	worldTree := internal.NewBVHFromWorld(world)
	return camera.Render(worldTree, f)
}

func cornellBox(f *os.File) error {
	camera := internal.NewCamera(
		1,
		600.0,
		internal.WithSamplesPerPixel(200),
		internal.WithMaxRayDepth(50),
		internal.WithLookFrom(internal.NewVec3(278, 278, -800)),
		internal.WithLookAt(internal.NewVec3(278, 278, 0)),
		internal.WithFOVDegrees(40),
		internal.WithDefocusAngleDegrees(0),
		internal.WithBackgroundColor(internal.NewVec3Zero()),
	)
	world := internal.NewWorld()

	red := internal.NewLambertian(internal.NewSolidColor(.65, .05, .05))
	white := internal.NewLambertian(internal.NewSolidColor(.73, .73, .73))
	green := internal.NewLambertian(internal.NewSolidColor(.12, .45, .15))
	light := internal.NewDiffuseLight(internal.NewSolidColor(15, 15, 15))

	world.Add(internal.NewQuad(internal.NewVec3(555, 0, 0), internal.NewVec3(0, 555, 0), internal.NewVec3(0, 0, 555), &green))
	world.Add(internal.NewQuad(internal.NewVec3(0, 0, 0), internal.NewVec3(0, 555, 0), internal.NewVec3(0, 0, 555), &red))
	world.Add(internal.NewQuad(internal.NewVec3(343, 554, 332), internal.NewVec3(-130, 0, 0), internal.NewVec3(0, 0, -105), &light))
	world.Add(internal.NewQuad(internal.NewVec3(0, 0, 0), internal.NewVec3(555, 0, 0), internal.NewVec3(0, 0, 555), &white))
	world.Add(internal.NewQuad(internal.NewVec3(555, 555, 555), internal.NewVec3(-555, 0, 0), internal.NewVec3(0, 0, -555), &white))
	world.Add(internal.NewQuad(internal.NewVec3(0, 0, 555), internal.NewVec3(555, 0, 0), internal.NewVec3(0, 555, 0), &white))

	worldTree := internal.NewBVHFromWorld(world)
	return camera.Render(worldTree, f)
}

func randSpheres(f *os.File) error {
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
		internal.WithBackgroundColor(internal.NewVec3(0.7, 0.8, 1)),
	)
	world := internal.NewWorld()

	checkered := internal.NewCheckered(0.32, internal.NewVec3(0.2, 0.3, 0.1), internal.NewVec3(0.9, 0.9, 0.9))
	matGround := internal.NewLambertian(&checkered)
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
					randCol := internal.Mul(internal.NewVec3Rand32(randCtx), internal.NewVec3Rand32(randCtx))
					tex := internal.NewSolidColor(randCol.X, randCol.Y, randCol.Z)
					mat := internal.NewLambertian(tex)
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

	m2 := internal.NewLambertian(internal.NewSolidColor(0.4, 0.2, 0.1))
	world.Add(internal.NewSphere(internal.NewVec3(-4, 1, 0), 1, &m2))

	m3 := internal.NewMetal(internal.NewVec3(0.7, 0.6, 0.5), 0)
	world.Add(internal.NewSphere(internal.NewVec3(4, 1, 0), 1, &m3))

	worldTree := internal.NewBVHFromWorld(world)
	return camera.Render(worldTree, f)
}
