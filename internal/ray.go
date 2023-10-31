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

func (r *Ray) GetColor(world Hittable, backgroundColor Color, maxDepth int) Color {
	if maxDepth <= 0 {
		return NewVec3Zero()
	}

	if hitInfo, ok := world.Hit(r, Interval{
		min: 0.001,
		max: float32(math.Inf(1)),
	}); ok {
		colorFromEmission := hitInfo.material.Emit(hitInfo.u, hitInfo.v, hitInfo.point).GetColor()
		scatterInfo, didScatter := hitInfo.material.Scatter(r, hitInfo)

		if !didScatter {
			return colorFromEmission
		}

		colorFromScatter := Mul(scatterInfo.attenuation.GetColor(), scatterInfo.ray.GetColor(world, backgroundColor, maxDepth-1).GetColor())

		return Add(colorFromEmission, colorFromScatter)
	}

	return backgroundColor
}

type Color interface {
	GetColor() Vec3
}
