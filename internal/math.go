package internal

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
