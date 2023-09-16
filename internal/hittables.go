package internal

import "math"

type Hittable interface {
	Hit(r *Ray, tMin, tMax float32) (HitInfo, bool)
}

type HitInfo struct {
	point     Vec3[float32]
	normal    Vec3[float32]
	t         float32
	material  Material
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
	Center   Vec3[float32]
	Radius   float32
	Material Material
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
		point:    point,
		normal:   norm,
		t:        t,
		material: s.Material,
	}

	hi.SetFaceNormal(r, norm)
	return hi, true

}
