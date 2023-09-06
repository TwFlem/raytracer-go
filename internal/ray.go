package internal

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
	unit := Unit(r.dir)
	a := 0.5 * (unit.Y + 1)

	white := NewVec3[float32](1, 1, 1)
	blue := NewVec3[float32](0.5, 0.7, 1)

	return Add(Scale(white, (1.0-a)), Scale(blue, a))
}
