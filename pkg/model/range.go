package model

import (
	"math/rand/v2"
	"time"
)

func Int(min, max int) int {
	if min == max {
		return min
	}

	if min > max {
		min, max = max, min
	}

	return rand.IntN(max-min) + min
}

func Float(min, max float64) float64 {
	if min == max {
		return min
	}

	if min > max {
		min, max = max, min
	}

	return min + rand.Float64()*(max-min)
}

func Timestamp(min, max time.Time) time.Time {
	if min.Equal(max) {
		return min
	}

	if min.After(max) {
		min, max = max, min
	}

	minUnix := min.Unix()
	maxUnix := max.Unix()
	delta := maxUnix - minUnix

	randUnix := minUnix + rand.Int64N(delta)
	return time.Unix(randUnix, 0)
}

func Interval(min, max time.Duration) time.Duration {
	if min == max {
		return min
	}

	if min > max {
		min, max = max, min
	}

	diff := max - min
	randomDiff := time.Duration(rand.Int64N(int64(diff)))

	return min + randomDiff
}
