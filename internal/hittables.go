package internal

import (
	"math"
)

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

func NewHitInfo(t float32, intersectingRayDirection, point, unitOutwardNormal Vec3[float32], material Material) HitInfo {
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
}

func NewWorld(hittables []Hittable) *World {
	return &World{
		hittables: hittables,
	}
}

func (w *World) Hit(r *Ray, tMin float32, tMax float32) (HitInfo, bool) {
	hitAny := false
	closest := tMax
	closestRecord := HitInfo{}
	for i := range w.hittables {
		hi, ok := w.hittables[i].Hit(r, tMin, closest)
		if ok {
			hitAny = true
			closestRecord = hi
			closest = hi.t
		}
	}

	return closestRecord, hitAny
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
	norm := Scale(Sub(point, s.Center), s.Radius)
	norm.Unit()

	hi := NewHitInfo(t, r.dir, point, norm, s.Material)

	return hi, true

}
