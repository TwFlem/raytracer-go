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
		dir := Add(hitInfo.normal, NewVec3UnitRandOnUnitSphere32())
		bouncedRay := NewRay(hitInfo.point, dir)
		color := bouncedRay.GetColor(world, remainingBounces-1)
		color.Scale(0.5)
		return color
	}

	unit := Unit(r.dir)
	a := 0.5 * (unit.Y + 1)

	white := NewVec3[float32](1, 1, 1)
	blue := NewVec3[float32](0.5, 0.7, 1)

	return Add(Scale(white, (1.0-a)), Scale(blue, a))
}

type HitInfo struct {
	point     Vec3[float32]
	normal    Vec3[float32]
	t         float32
	frontFace bool
}

func (hi *HitInfo) SetFaceNormal(r *Ray, unitOutwardNormal Vec3[float32]) {
	frontFace := Dot(r.dir, unitOutwardNormal) < 0
	hi.frontFace = frontFace
	if frontFace {
		hi.normal = unitOutwardNormal
	} else {
		hi.normal = Scale(unitOutwardNormal, -1)
	}
}

// Hittable anything that can be hit by a Ray
type Hittable interface {
	Hit(r *Ray, tMin, tMax float32) (HitInfo, bool)
}

type World struct {
	hittables []Hittable
}

func NewWorld(hittables []Hittable) *World {
	return &World{
		hittables: hittables,
	}
}

func (w *World) Hit(r *Ray, tMin float32, tMax float32) (HitInfo, bool) {
	for i := range w.hittables {
		hi, ok := w.hittables[i].Hit(r, tMin, tMax)
		if ok {
			return hi, true
		}
	}

	return HitInfo{}, false
}

type Sphere struct {
	Center Vec3[float32]
	Radius float32
}

func (s *Sphere) Hit(r *Ray, tMin float32, tMax float32) (HitInfo, bool) {
	ASubC := Sub(r.origin, s.Center)
	a := r.dir.LenSq()
	halfB := Dot(r.dir, ASubC)
	c := ASubC.LenSq() - s.Radius*s.Radius

	discriminate := (halfB*halfB - a*c)

	if discriminate < 0 {
		return HitInfo{}, false
	}

	sqt := float32(math.Sqrt(float64(discriminate)))
	var t float32
	if r := (-halfB - sqt) / a; Contains(tMin, tMax, r) {
		t = r
	} else if r := (-halfB + sqt) / a; Contains(tMin, tMax, r) {
		t = r
	} else {
		return HitInfo{}, false
	}

	point := r.At(t)
	norm := Scale(Sub(point, s.Center), 1/s.Radius)

	hi := HitInfo{
		point:  point,
		normal: norm,
		t:      t,
	}

	hi.SetFaceNormal(r, norm)
	return hi, true

}
