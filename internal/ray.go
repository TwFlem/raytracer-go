package internal

import (
	"math"
	"math/rand"
	"time"
)

type Ray struct {
	origin    Vec3
	dir       Vec3
	rand      *rand.Rand
	startTime time.Time
}

func NewRay(origin, dir Vec3, randCtx *rand.Rand) *Ray {
	return &Ray{
		origin:    origin,
		dir:       dir,
		rand:      randCtx,
		startTime: time.Now(),
	}
}

func (r *Ray) At(t float32) Vec3 {
	dir := r.dir.Cpy()
	dir.Scale(t)
	dir.Add(r.origin)
	return dir
}

type GetColorInfo struct {
	color   Vec3
	nextRay *Ray
}

func (r *Ray) GetColor(world Hittable, maxDepth int) Vec3 {
	initColorInfo := getNextColor(r, world)
	color := initColorInfo.color
	if initColorInfo.nextRay != nil {
		nextRay := initColorInfo.nextRay
		bounce := 1
		for nextRay != nil && bounce < maxDepth {
			cInfo := getNextColor(nextRay, world)
			color.Mul(cInfo.color)
			nextRay = cInfo.nextRay
			bounce++
		}
		if bounce >= maxDepth {
			return NewVec3Zero()
		}
	}
	return color
}

func getNextColor(r *Ray, world Hittable) GetColorInfo {
	if hitInfo, ok := world.Hit(r, Interval{
		min: 0.001,
		max: float32(math.Inf(1)),
	}); ok {
		if scatterInfo, ok := hitInfo.material.Scatter(r, hitInfo); ok {
			return GetColorInfo{
				color:   scatterInfo.attenuation,
				nextRay: &scatterInfo.ray,
			}
		}
		return GetColorInfo{
			color:   NewVec3Zero(),
			nextRay: nil,
		}
	}

	unit := Unit(r.dir)
	a := 0.5 * (unit.Y + 1)

	white := NewVec3(1, 1, 1)
	blue := NewVec3(0.5, 0.7, 1)

	sky := Add(Scale(white, (1.0-a)), Scale(blue, a))
	return GetColorInfo{
		color:   sky,
		nextRay: nil,
	}
}
