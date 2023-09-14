package internal

import (
	"fmt"
	"math"
	"math/rand"
)

type Vec3[T Float] struct {
	X T
	Y T
	Z T
}

func NewVec3[T Float](x, y, z T) Vec3[T] {
	return Vec3[T]{
		X: x,
		Y: y,
		Z: z,
	}
}

func NewVec3Zero[T Float]() Vec3[T] {
	return Vec3[T]{
		X: 0,
		Y: 0,
		Z: 0,
	}
}

func NewVec3Unit[T Float]() Vec3[T] {
	return Vec3[T]{
		X: 1,
		Y: 1,
		Z: 1,
	}
}

func (v *Vec3[T]) Cpy() Vec3[T] {
	return *v
}

func (v *Vec3[T]) Add(in Vec3[T]) {
	v.X += in.X
	v.Y += in.Y
	v.Z += in.Z
}

func Add[T Float](a, b Vec3[T]) Vec3[T] {
	c := a.Cpy()
	c.Add(b)
	return c
}

func (v *Vec3[T]) Mul(in Vec3[T]) {
	v.X *= in.X
	v.Y *= in.Y
	v.Z *= in.Z
}

func Mul[T Float](a, b Vec3[T]) Vec3[T] {
	c := a.Cpy()
	c.Mul(b)
	return c
}

func (v *Vec3[T]) Sub(in Vec3[T]) {
	v.X -= in.X
	v.Y -= in.Y
	v.Z -= in.Z
}

func Sub[T Float](a, b Vec3[T]) Vec3[T] {
	c := a.Cpy()
	c.Sub(b)
	return c
}

func (v *Vec3[T]) Div(in Vec3[T]) {
	v.X *= 1 / in.X
	v.Y *= 1 / in.Y
	v.Z *= 1 / in.Z
}

func Div[T Float](a, b Vec3[T]) Vec3[T] {
	c := a.Cpy()
	c.Div(b)
	return c
}

func (v *Vec3[T]) Scale(in T) {
	v.X *= in
	v.Y *= in
	v.Z *= in
}

func Scale[T Float](a Vec3[T], scaler T) Vec3[T] {
	c := a.Cpy()
	c.Scale(scaler)
	return c
}

func (v *Vec3[T]) Unit() {
	lensq := v.LenSq()
	l := T(math.Sqrt(float64(lensq)))
	v.Scale(1 / l)
}

func Unit[T Float](a Vec3[T]) Vec3[T] {
	c := a.Cpy()
	c.Unit()
	return c
}

func (v *Vec3[T]) LenSq() T {
	return v.X*v.X + v.Y*v.Y + v.Z*v.Z
}

func (v *Vec3[T]) Cross(l, r Vec3[T]) {
	v.X = l.Y*r.Z - l.Z*r.Y
	v.Y = l.Z*r.X - l.X*r.Z
	v.Z = l.X*r.Y - l.Y*r.X
}

func Dot[T Float](l, r Vec3[T]) T {
	return l.X*r.X + l.Y*r.Y + l.Z*r.Z
}

func (v *Vec3[T]) String() string {
	return fmt.Sprintf("%d %d %d", int(v.X), int(v.Y), int(v.Z))
}

func (v *Vec3[T]) ToRGB() {
	v.X = Clamp(0, 1, v.X)
	v.Y = Clamp(0, 1, v.Y)
	v.Z = Clamp(0, 1, v.Z)
	v.X *= 255.999
	v.Y *= 255.999
	v.Z *= 255.999
}

func NewVec3Rand32() Vec3[float32] {
	return NewVec3[float32](rand.Float32(), rand.Float32(), rand.Float32())
}

func NewVec3RandRange32(min, max float32) Vec3[float32] {
	return NewVec3[float32](RandF32N(min, max), RandF32N(min, max), RandF32N(min, max))
}

func NewVec3UnitRandOnUnitSphere32() Vec3[float32] {
	for {
		v := NewVec3RandRange32(-1, 1)
		if v.LenSq() < 1.0 {
			v.Unit()
			return v
		}
	}
}

// NewVec3RandInHemisphereOfSurroundingUnitSphere32 gives a random vector that lies on a unit sphere that is in the same
// hemisphere as the surface normal provided
func NewVec3RandInHemisphereOfSurroundingUnitSphere32(norm Vec3[float32]) Vec3[float32] {
	v := NewVec3UnitRandOnUnitSphere32()
	if Dot(v, norm) < 0 {
		v.Scale(-1)
		return v
	}
	return v
}
