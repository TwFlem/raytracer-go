package internal

import (
	"fmt"
	"math"
	"math/rand"
)

type Vec3 struct {
	X float32
	Y float32
	Z float32
}

func NewVec3(x, y, z float32) Vec3 {
	return Vec3{
		X: x,
		Y: y,
		Z: z,
	}
}

func NewVec3Zero() Vec3 {
	return Vec3{
		X: 0,
		Y: 0,
		Z: 0,
	}
}

func NewVec3Unit() Vec3 {
	return Vec3{
		X: 1,
		Y: 1,
		Z: 1,
	}
}

func (v *Vec3) Cpy() Vec3 {
	return *v
}

func (v *Vec3) Add(in Vec3) {
	v.X += in.X
	v.Y += in.Y
	v.Z += in.Z
}

func Add(a, b Vec3) Vec3 {
	c := a.Cpy()
	c.Add(b)
	return c
}

func (v *Vec3) Mul(in Vec3) {
	v.X *= in.X
	v.Y *= in.Y
	v.Z *= in.Z
}

func Mul(a, b Vec3) Vec3 {
	c := a.Cpy()
	c.Mul(b)
	return c
}

func (v *Vec3) Sub(in Vec3) {
	v.X -= in.X
	v.Y -= in.Y
	v.Z -= in.Z
}

func Sub(a, b Vec3) Vec3 {
	c := a.Cpy()
	c.Sub(b)
	return c
}

func (v *Vec3) Div(in Vec3) {
	v.X /= in.X
	v.Y /= in.Y
	v.Z /= in.Z
}

func Div(a, b Vec3) Vec3 {
	c := a.Cpy()
	c.Div(b)
	return c
}

func (v *Vec3) Scale(in float32) {
	v.X *= in
	v.Y *= in
	v.Z *= in
}

func Scale(a Vec3, scaler float32) Vec3 {
	c := a.Cpy()
	c.Scale(scaler)
	return c
}

func (v *Vec3) Unit() {
	lensq := v.LenSq()
	l := float32(math.Sqrt(float64(lensq)))
	v.Scale(1 / l)
}

func Unit(a Vec3) Vec3 {
	c := a.Cpy()
	c.Unit()
	return c
}

func (v *Vec3) LenSq() float32 {
	return v.X*v.X + v.Y*v.Y + v.Z*v.Z
}

func (v *Vec3) Len() float32 {
	return float32(math.Sqrt(float64(v.LenSq())))
}

func (v *Vec3) Cross(l, r Vec3) {
	v.X = l.Y*r.Z - l.Z*r.Y
	v.Y = l.Z*r.X - l.X*r.Z
	v.Z = l.X*r.Y - l.Y*r.X
}

func Cross(l, r Vec3) Vec3 {
	v := NewVec3Zero()
	v.X = l.Y*r.Z - l.Z*r.Y
	v.Y = l.Z*r.X - l.X*r.Z
	v.Z = l.X*r.Y - l.Y*r.X
	return v
}

func Dot(l, r Vec3) float32 {
	return l.X*r.X + l.Y*r.Y + l.Z*r.Z
}

func (v *Vec3) String() string {
	return fmt.Sprintf("%d %d %d", int(v.X), int(v.Y), int(v.Z))
}

func (v *Vec3) ToRGB() {
	v.X = Clamp(0, 1, v.X)
	v.Y = Clamp(0, 1, v.Y)
	v.Z = Clamp(0, 1, v.Z)
	v.X *= 255.999
	v.Y *= 255.999
	v.Z *= 255.999
}

func (v *Vec3) ToGamma2() {
	v.X = float32(math.Sqrt(float64(v.X)))
	v.Y = float32(math.Sqrt(float64(v.Y)))
	v.Z = float32(math.Sqrt(float64(v.Z)))
}

const nearZeroEpsilon float32 = 1e-8

func (v *Vec3) NearZero() bool {
	return float32(math.Abs(float64(v.X))) < float32(nearZeroEpsilon) && float32(math.Abs(float64(v.Y))) < float32(nearZeroEpsilon) && float32(math.Abs(float64(v.Z))) < float32(nearZeroEpsilon)
}

func NewVec3Rand32(randCtx *rand.Rand) Vec3 {
	return NewVec3(randCtx.Float32(), randCtx.Float32(), randCtx.Float32())
}

func NewVec3RandRange32(randCtx *rand.Rand, min, max float32) Vec3 {
	return NewVec3(RandF32N(randCtx, min, max), RandF32N(randCtx, min, max), RandF32N(randCtx, min, max))
}

func NewVec3UnitRandOnUnitSphere32(randCtx *rand.Rand) Vec3 {
	for {
		v := NewVec3RandRange32(randCtx, -1, 1)
		if v.LenSq() < 1.0 {
			v.Unit()
			return v
		}
	}
}

// NewVec3RandInHemisphereOfSurroundingUnitSphere32 gives a random vector that lies on a unit sphere that is in the same
// hemisphere as the surface normal provided
func NewVec3RandInHemisphereOfSurroundingUnitSphere32(randCtx *rand.Rand, norm Vec3) Vec3 {
	v := NewVec3UnitRandOnUnitSphere32(randCtx)
	if Dot(v, norm) < 0 {
		v.Scale(-1)
		return v
	}
	return v
}

func NewVec3RandInUnitDisk(randCtx *rand.Rand) Vec3 {
	for {
		v := NewVec3(RandF32N(randCtx, -1, 1), RandF32N(randCtx, -1, 1), 0)
		if v.LenSq() < 1 {
			return v
		}
	}
}

func reflect(v, n Vec3) Vec3 {
	return Sub(v, Scale(n, 2*Dot(v, n)))
}

func refract(uv, n Vec3, etaOEtaPrime float32) Vec3 {
	cosfloat32heta := Dot(Scale(uv, -1), n)
	perp := Scale(Add(uv, Scale(n, cosfloat32heta)), etaOEtaPrime)
	par := Scale(n, -1*float32(math.Sqrt(math.Abs(float64(1.0-perp.LenSq())))))
	return Add(par, perp)
}
