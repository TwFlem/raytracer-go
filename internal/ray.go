package internal

import "math"

type Ray struct {
	origin Vec3[float32]
	dir    Vec3[float32]
}

func NewRay(origin, dir Vec3[float32]) *Ray {
	return &Ray{
		origin: origin,
		dir:    dir,
	}
}

func (r *Ray) At(t float32) Vec3[float32] {
	dir := r.dir.Cpy()
	dir.Scale(t)
	dir.Add(r.origin)
	return dir
}

func (r *Ray) GetColor(world *World, remainingBounces int) Vec3[float32] {
	if remainingBounces == 0 {
		return NewVec3Zero[float32]()
	}
	if hitInfo, ok := world.Hit(r, 0.001, float32(math.Inf(1))); ok {
		if scatterInfo, ok := hitInfo.material.Scatter(r, hitInfo); ok {
			color := scatterInfo.ray.GetColor(world, remainingBounces-1)
			color.Mul(scatterInfo.attenuation)
			return color
		}
		return NewVec3[float32](0, 0, 0)
	}

	unit := Unit(r.dir)
	a := 0.5 * (unit.Y + 1)

	white := NewVec3[float32](1, 1, 1)
	blue := NewVec3[float32](0.5, 0.7, 1)

	return Add(Scale(white, (1.0-a)), Scale(blue, a))
}
