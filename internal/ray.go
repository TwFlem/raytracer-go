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

func (r *Ray) GetColor() Vec3[float32] {
	sphereCenter := NewVec3[float32](0, 0, -1)
	if t := getClosestRootToTheCamera(sphereCenter, 0.5, r); t > 0 {
		n := Sub(r.At(t), sphereCenter)
		n.Unit()
		color := NewVec3[float32](1+n.X, 1+n.Y, 1+n.Z)
		color.Scale(0.5)
		return color
	}

	unit := Unit(r.dir)
	a := 0.5 * (unit.Y + 1)

	white := NewVec3[float32](1, 1, 1)
	blue := NewVec3[float32](0.5, 0.7, 1)

	return Add(Scale(white, (1.0-a)), Scale(blue, a))
}

func getClosestRootToTheCamera(center Vec3[float32], radius float32, r *Ray) float32 {
	ASubC := Sub(r.origin, center)
	a := Dot(r.dir, r.dir)
	b := 2 * Dot(ASubC, r.dir)
	c := Dot(ASubC, ASubC) - radius*radius

	discriminate := (b*b - 4*a*c)

	if discriminate < 0 {
		return -1
	}

	return (-b - float32(math.Sqrt(float64(discriminate)))) / (2 * a)
}
