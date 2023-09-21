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

func RandF32N(min, max float32) float32 {
	return min + rand.Float32()*(max-min)
}

func AbsF32(in float32) float32 {
	return math.Float32frombits(math.Float32bits(in) &^ (1 << 31))
}

const (
	PiO2 float64 = math.Pi / 2
)
