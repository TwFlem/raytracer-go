package internal

import (
	"math"
)

type Hittable interface {
	Hit(r *Ray, rayT Interval) (HitInfo, bool)
	GetBounds() Aabb
}

type HitInfo struct {
	point     Vec3
	normal    Vec3
	t         float32
	material  Material
	frontFace bool
}

func NewHitInfo(t float32, intersectingRayDirection, point, unitOutwardNormal Vec3, material Material) HitInfo {
	frontFace := Dot(intersectingRayDirection, unitOutwardNormal) < 0
	if !frontFace {
		unitOutwardNormal.Scale(-1)
	}

	return HitInfo{
		point:     point,
		normal:    unitOutwardNormal,
		t:         t,
		material:  material,
		frontFace: frontFace,
	}
}

type World struct {
	hittables []Hittable
	bBox      Aabb
}

func NewWorld() *World {
	return &World{}
}

func (w *World) Add(hittables ...Hittable) {
	for i := range hittables {
		w.hittables = append(w.hittables, hittables[i])
		w.bBox = NewAabbFromBoxes(w.bBox, hittables[i].GetBounds())
	}
}

func (w *World) Hit(r *Ray, rayT Interval) (HitInfo, bool) {
	hitAny := false
	closest := rayT.max
	closestRecord := HitInfo{}
	for i := range w.hittables {
		hi, ok := w.hittables[i].Hit(r, Interval{
			min: rayT.min,
			max: closest,
		})
		if ok {
			hitAny = true
			closestRecord = hi
			closest = hi.t
		}
	}

	return closestRecord, hitAny
}

func (w *World) GetBounds() Aabb {
	return w.bBox
}

type Sphere struct {
	Center   Vec3
	Radius   float32
	Material Material
	bBox     Aabb
}

func NewSphere(center Vec3, radius float32, mat Material) *Sphere {
	rvec := NewVec3(radius, radius, radius)
	return &Sphere{
		Center:   center,
		Radius:   radius,
		Material: mat,
		bBox:     NewAabb(Add(center, Scale(rvec, -1)), Add(center, rvec)),
	}

}

func (s *Sphere) Hit(r *Ray, rayT Interval) (HitInfo, bool) {
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
	if r := (-halfB - sqt) / a; rayT.In(r, 0) {
		t = r
	} else if r := (-halfB + sqt) / a; rayT.In(r, 0) {
		t = r
	} else {
		return HitInfo{}, false
	}

	point := r.At(t)
	norm := Scale(Sub(point, s.Center), s.Radius)
	norm.Unit()

	hi := NewHitInfo(t, r.dir, point, norm, s.Material)

	return hi, true

}

func (s *Sphere) GetBounds() Aabb {
	return s.bBox
}
