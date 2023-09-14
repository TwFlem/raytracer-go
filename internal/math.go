package internal

import "math/rand"

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
