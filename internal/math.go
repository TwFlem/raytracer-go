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

func ToRadians(degrees float32) float32 {
	return degrees * radRatio
}

const (
	PiO2 float64 = math.Pi / 2
)
