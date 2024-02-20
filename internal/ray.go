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

func (r *Ray) GetColor(cw *CameraWorker, world Hittable, backgroundColor Color, bounceDepth int) Color {
	currRay := r
	var currRayColor Color
	stopped := 0
	for j := 0; j < bounceDepth; j++ {
		hitInfo, ok := world.Hit(currRay, Interval{0.001, float32(math.Inf(1))})
		if !ok {
			currRayColor = backgroundColor
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
		return NewVec3Zero()
	}

	currRaySample := currRayColor.GetColor()
	for j := 0; j < stopped; j++ {
		currRaySample.Mul(cw.attenuationStack[stopped-1-j].GetColor())
		currRaySample.Add(cw.emissionStack[stopped-1-j].GetColor())
	}
	return currRaySample
}

type Color interface {
	GetColor() Vec3
}
