package internal

import (
	"math"
	"math/rand"
)

type Float interface {
	float32 | float64
}

func Contains(min, max, val float32) bool {
	return min <= val && val <= max
}

func StrictlyWithin(min, max, val float32) bool {
	return min < val && val < max
}

func Clamp[T Float](min, max, val T) T {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

func RandF32N(randCtx *rand.Rand, min, max float32) float32 {
	return min + randCtx.Float32()*(max-min)
}

func AbsF32(in float32) float32 {
	return math.Float32frombits(math.Float32bits(in) &^ (1 << 31))
}

func MinF32(a, b float32) float32 {
	return float32(math.Min(float64(a), float64(b)))
}

func MaxF32(a, b float32) float32 {
	return float32(math.Max(float64(a), float64(b)))
}

const radRatio float32 = math.Pi / 180.0

const PiF32 = float32(math.Pi)

func ToRadians(degrees float32) float32 {
	return degrees * radRatio
}

const (
	PiO2 float64 = math.Pi / 2
)

func Lerp(t, x, y float32) float32 {
	return x*(1-t) + y*t
}

func Vec3BiLinearLerp(tx, ty float32, c00, c10, c01, c11 Vec3) Vec3 {
	a := Add(Scale(c00, 1-tx), Scale(c10, tx))
	b := Add(Scale(c01, 1-tx), Scale(c11, tx))
	return Add(Scale(a, 1-ty), Scale(b, ty))
}

func Vec3TriLinearLerp(
	tx, ty, tz float32,
	c000, c100, c010, c110 Vec3,
	c001, c101, c011, c111 Vec3,
) Vec3 {
	e := Vec3BiLinearLerp(tx, ty, c000, c100, c010, c110)
	f := Vec3BiLinearLerp(tx, ty, c001, c101, c011, c111)
	return Add(Scale(e, 1-tz), Scale(f, tz))
}

func BiLinearLerp(tx, ty float32, c00, c10, c01, c11 float32) float32 {
	a := Lerp(tx, c00, c10)
	b := Lerp(tx, c01, c11)
	return Lerp(ty, a, b)
}

func TriLinearLerp(
	tx, ty, tz float32,
	c000, c100, c010, c110 float32,
	c001, c101, c011, c111 float32,
) float32 {
	e := BiLinearLerp(tx, ty, c000, c100, c010, c110)
	f := BiLinearLerp(tx, ty, c001, c101, c011, c111)
	return Lerp(tz, e, f)
}
