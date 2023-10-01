package internal

import (
	"fmt"
	"math"
	"math/rand"
)

type Vec3 struct {
	D [4]float32
}

func NewVec3(x, y, z, a float32) Vec3 {
	return Vec3{[4]float32{x, y, z, a}}
}

func NewVec3Zero() Vec3 {
	return Vec3{[4]float32{0, 0, 0, 0}}
}

func NewVec3Unit() Vec3 {
	return Vec3{[4]float32{1, 1, 1, 1}}
}

func (v *Vec3) Cpy() Vec3 {
	return *v
}

func (v *Vec3) Add(in Vec3) {
	v.D[0] += in.D[0]
	v.D[1] += in.D[1]
	v.D[2] += in.D[2]
}

func Add(a, b Vec3) Vec3 {
	c := a.Cpy()
	c.Add(b)
	return c
}

func (v *Vec3) Mul(in Vec3) {
	v.D[0] *= in.D[0]
	v.D[1] *= in.D[1]
	v.D[2] *= in.D[2]
}

func Mul(a, b Vec3) Vec3 {
	c := a.Cpy()
	c.Mul(b)
	return c
}

func (v *Vec3) Sub(in Vec3) {
	v.D[0] -= in.D[0]
	v.D[1] -= in.D[1]
	v.D[2] -= in.D[2]
}

func Sub(a, b Vec3) Vec3 {
	c := a.Cpy()
	c.Sub(b)
	return c
}

func (v *Vec3) Div(in Vec3) {
	v.D[0] /= in.D[0]
	v.D[1] /= in.D[1]
	v.D[2] /= in.D[2]
}

func Div(a, b Vec3) Vec3 {
	c := a.Cpy()
	c.Div(b)
	return c
}

func (v *Vec3) Scale(in float32) {
	v.D[0] *= in
	v.D[1] *= in
	v.D[2] *= in
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
	return v.D[0]*v.D[0] + v.D[1]*v.D[1] + v.D[2]*v.D[2]
}

func (v *Vec3) Len() float32 {
	return float32(math.Sqrt(float64(v.LenSq())))
}

func (v *Vec3) Cross(l, r Vec3) {
	v.D[0] = l.D[1]*r.D[2] - l.D[2]*r.D[1]
	v.D[1] = l.D[2]*r.D[0] - l.D[0]*r.D[2]
	v.D[2] = l.D[0]*r.D[1] - l.D[1]*r.D[0]
}

func Cross(l, r Vec3) Vec3 {
	v := NewVec3Zero()
	v.D[0] = l.D[1]*r.D[2] - l.D[2]*r.D[1]
	v.D[1] = l.D[2]*r.D[0] - l.D[0]*r.D[2]
	v.D[2] = l.D[0]*r.D[1] - l.D[1]*r.D[0]
	return v
}

func Dot(l, r Vec3) float32 {
	return l.D[0]*r.D[0] + l.D[1]*r.D[1] + l.D[2]*r.D[2]
}

func (v *Vec3) String() string {
	return fmt.Sprintf("%d %d %d", int(v.D[0]), int(v.D[1]), int(v.D[2]))
}

func (v *Vec3) ToRGB() {
	v.D[0] = Clamp(0, 1, v.D[0])
	v.D[1] = Clamp(0, 1, v.D[1])
	v.D[2] = Clamp(0, 1, v.D[2])
	v.D[0] *= 255.999
	v.D[1] *= 255.999
	v.D[2] *= 255.999
}

func (v *Vec3) ToGamma2() {
	v.D[0] = float32(math.Sqrt(float64(v.D[0])))
	v.D[1] = float32(math.Sqrt(float64(v.D[1])))
	v.D[2] = float32(math.Sqrt(float64(v.D[2])))
}

const nearZeroEpsilon float32 = 1e-8

func (v *Vec3) NearZero() bool {
	return float32(math.Abs(float64(v.D[0]))) < float32(nearZeroEpsilon) && float32(math.Abs(float64(v.D[1]))) < float32(nearZeroEpsilon) && float32(math.Abs(float64(v.D[2]))) < float32(nearZeroEpsilon)
}

func NewVec3Rand32(randCtx *rand.Rand) Vec3 {
	return NewVec3(randCtx.Float32(), randCtx.Float32(), randCtx.Float32(), randCtx.Float32())
}

func NewVec3RandRange32(randCtx *rand.Rand, min, max float32) Vec3 {
	return NewVec3(RandF32N(randCtx, min, max), RandF32N(randCtx, min, max), RandF32N(randCtx, min, max), RandF32N(randCtx, min, max))
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
		v := NewVec3(RandF32N(randCtx, -1, 1), RandF32N(randCtx, -1, 1), 0, 0)
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
