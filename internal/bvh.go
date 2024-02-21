package internal

import (
	"math/rand"

	"golang.org/x/exp/slices"
)

type Interval struct {
	min float32
	max float32
}

func NewInterval(min, max float32) Interval {
	return Interval{min, max}
}

func (i Interval) In(v, padding float32) bool {
	return i.min-padding < v && v < i.max+padding
}

type Aabb struct {
	x Interval
	y Interval
	z Interval
}

func NewAabb(p1, p2 Vec3) Aabb {
	return Aabb{
		x: NewInterval(MinF32(p1.X, p2.X), MaxF32(p1.X, p2.X)),
		y: NewInterval(MinF32(p1.Y, p2.Y), MaxF32(p1.Y, p2.Y)),
		z: NewInterval(MinF32(p1.Z, p2.Z), MaxF32(p1.Z, p2.Z)),
	}
}

func NewAabbFromIntervals(x, y, z Interval) Aabb {
	return Aabb{
		x: x,
		y: y,
		z: z,
	}
}

func NewAabbFromBoxes(b1, b2 Aabb) Aabb {
	return Aabb{
		x: NewInterval(MinF32(b1.x.min, b2.x.min), MaxF32(b1.x.max, b2.x.max)),
		y: NewInterval(MinF32(b1.y.min, b2.y.min), MaxF32(b1.y.max, b2.y.max)),
		z: NewInterval(MinF32(b1.z.min, b2.z.min), MaxF32(b1.z.max, b2.z.max)),
	}
}

func (a *Aabb) Hit(r *Ray, rT Interval) bool {
	if ok := InBoundary(r.dir.X, r.origin.X, a.x.min, a.x.max, &rT); ok {
		if ok := InBoundary(r.dir.Y, r.origin.Y, a.y.min, a.y.max, &rT); ok {
			if ok = InBoundary(r.dir.Z, r.origin.Z, a.z.min, a.z.max, &rT); ok {
				return true
			}
		}
	}
	return false
}

func (a Aabb) GetPaddedAabb() Aabb {
	eps := float32(0.0001)
	x := a.x
	if x.max-x.min < eps {
		x.min -= eps
		x.max += eps
	}
	y := a.y
	if y.max-y.min < eps {
		y.min -= eps
		y.max += eps
	}
	z := a.z
	if z.max-z.min < eps {
		z.min -= eps
		z.max += eps
	}

	return NewAabbFromIntervals(x, y, z)
}

func InBoundary(dir, origin, aMin, aMax float32, rT *Interval) bool {
	invD := 1 / dir
	t0 := (aMin - origin) * invD
	t1 := (aMax - origin) * invD

	if invD < 0 {
		t0, t1 = t1, t0
	}

	if t0 > rT.min {
		rT.min = t0
	}

	if t1 < rT.max {
		rT.max = t1
	}

	return rT.min < rT.max
}

func (a Aabb) GetBounds() Aabb {
	return a
}

type BVH struct {
	left  Hittable
	right Hittable
	bBox  Aabb
}

func NewBVHFromWorld(w *World) *BVH {
	return NewBVH(w.hittables)
}

func NewBVH(hittables []Hittable) *BVH {
	bvh := &BVH{}
	h := make([]Hittable, len(hittables))
	copy(h, hittables)

	axis := rand.Intn(3)
	var compare func(h1, h2 Hittable) int
	switch axis {
	case 0:
		compare = HittableCompareX
		break
	case 1:
		compare = HittableCompareY
		break
	case 2:
		compare = HittableCompareZ
		break
	}

	switch len(h) {
	case 1:
		bvh.left = h[0]
		bvh.right = h[0]
		break
	case 2:
		if compare(h[0], h[1]) > 0 {
			bvh.left = h[1]
			bvh.right = h[0]
		} else {
			bvh.left = h[0]
			bvh.right = h[1]
		}
		break
	default:
		slices.SortFunc(h, compare)
		mid := len(h) / 2
		bvh.left = NewBVH(h[:mid])
		bvh.right = NewBVH(h[mid:])
	}

	bvh.bBox = NewAabbFromBoxes(bvh.left.GetBounds(), bvh.right.GetBounds())

	return bvh
}

func HittableCompareX(h1, h2 Hittable) int {
	diff := h2.GetBounds().x.min - h1.GetBounds().x.min
	if diff > 0 {
		return 1
	}
	if diff < 0 {
		return -1
	}
	return 0
}

func HittableCompareY(h1, h2 Hittable) int {
	diff := h2.GetBounds().y.min - h1.GetBounds().y.min
	if diff > 0 {
		return 1
	}
	if diff < 0 {
		return -1
	}
	return 0
}

func HittableCompareZ(h1, h2 Hittable) int {
	diff := h2.GetBounds().z.min - h1.GetBounds().z.min
	if diff > 0 {
		return 1
	}
	if diff < 0 {
		return -1
	}
	return 0
}

func (b *BVH) Hit(r *Ray, rayT Interval) (HitInfo, bool) {
	if !b.bBox.Hit(r, rayT) {
		return HitInfo{}, false
	}

	leftHitInfo, hitLeft := b.left.Hit(r, rayT)

	rightInterval := rayT
	if hitLeft {
		rightInterval.max = leftHitInfo.t
	}

	rightHitInfo, hitRight := b.right.Hit(r, rightInterval)

	if hitLeft && hitRight {
		if leftHitInfo.t < rightHitInfo.t {
			return leftHitInfo, true
		}
		return rightHitInfo, true
	}

	if hitRight {
		return rightHitInfo, true
	}

	if hitLeft {
		return leftHitInfo, true
	}
	return HitInfo{}, false
}

func (b *BVH) GetBounds() Aabb {
	return b.bBox
}
