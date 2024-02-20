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
	u         float32
	v         float32
	material  Material
	frontFace bool
}

func NewHitInfo(t, u, v float32, intersectingRayDirection, point, unitOutwardNormal Vec3, material Material) HitInfo {
	frontFace := Dot(intersectingRayDirection, unitOutwardNormal) < 0
	if !frontFace {
		unitOutwardNormal.Scale(-1)
	}

	return HitInfo{
		point:     point,
		normal:    unitOutwardNormal,
		t:         t,
		u:         u,
		v:         v,
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

	theta := float32(math.Acos(-float64(norm.Y)))
	phi := float32(math.Atan2(-float64(norm.Z), float64(norm.X)) + math.Pi)
	// TODO: this 5*PiF32 should not be here. Why is it needed to match the example earth texture? Is the camera actually incorrect?
	u := (phi + 5*PiF32/12) / (2 * PiF32)
	v := theta / (PiF32)

	hi := NewHitInfo(t, u, v, r.dir, point, norm, s.Material)

	return hi, true

}

func (s *Sphere) GetBounds() Aabb {
	return s.bBox
}

type Quad struct {
	Q        Vec3
	u        Vec3
	v        Vec3
	w        Vec3
	normal   Vec3
	D        float32
	material Material
	bBox     Aabb
}

func NewQuad(Q, u, v Vec3, material Material) Quad {
	n := Cross(u, v)
	norm := Unit(n)
	D := Dot(norm, Q)
	w := Scale(n, 1/Dot(n, n))

	return Quad{
		Q:        Q,
		u:        u,
		v:        v,
		w:        w,
		material: material,
		bBox:     NewAabb(Q, Add(Add(Q, u), v)).GetPaddedAabb(),
		D:        D,
		normal:   norm,
	}
}

func (q Quad) Hit(r *Ray, rayT Interval) (HitInfo, bool) {
	denom := Dot(r.dir, q.normal)

	if math.Abs(float64(denom)) < 1e-8 {
		return HitInfo{}, false
	}

	t := (q.D - Dot(q.normal, r.origin)) / denom

	if !rayT.In(t, 0) {
		return HitInfo{}, false
	}

	intersection := r.At(t)
	planeHitPoint := Sub(intersection, q.Q)
	alpha := Dot(q.w, Cross(planeHitPoint, q.v))
	beta := Dot(q.w, Cross(q.u, planeHitPoint))

	if !q.InPlane(alpha, beta) {
		return HitInfo{}, false
	}

	return NewHitInfo(t, alpha, beta, r.dir, intersection, q.normal, q.material), true
}

func (q *Quad) InPlane(alpha, beta float32) bool {
	return !(alpha < 0 || 1 < alpha || beta < 0 || 1 < beta)
}

func (q Quad) GetBounds() Aabb {
	return q.bBox
}

func Box(a, b Vec3, mat Material) []Hittable {
	min := NewVec3(MinF32(a.X, b.X), MinF32(a.Y, b.Y), MinF32(a.Z, b.Z))
	max := NewVec3(MaxF32(a.X, b.X), MaxF32(a.Y, b.Y), MaxF32(a.Z, b.Z))

	dx := NewVec3(max.X-min.X, 0, 0)
	dy := NewVec3(0, max.Y-min.Y, 0)
	dz := NewVec3(0, 0, max.Z-min.Z)

	return []Hittable{
		NewQuad(NewVec3(min.X, min.Y, max.Z), dx, dy, mat),
		NewQuad(NewVec3(max.X, min.Y, max.Z), Scale(dz, -1), dy, mat),
		NewQuad(NewVec3(max.X, min.Y, min.Z), Scale(dx, -1), dy, mat),
		NewQuad(NewVec3(min.X, min.Y, min.Z), dz, dy, mat),
		NewQuad(NewVec3(min.X, max.Y, max.Z), dx, Scale(dz, -1), mat),
		NewQuad(NewVec3(min.X, min.Y, min.Z), dx, dz, mat),
	}
}
