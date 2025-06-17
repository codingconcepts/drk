package model

import (
	"math"
	"math/rand/v2"
	"time"
)

const (
	earthRadiusKM = 6_378
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

func Point(lat, lon, radiusKM float64) (float64, float64) {
	randomDistance := (rand.Float64() * radiusKM) / earthRadiusKM
	randomBearing := rand.Float64() * 2 * math.Pi

	latRad := degreesToRadians(lat)
	lonRad := degreesToRadians(lon)

	sinLatRad := math.Sin(latRad)
	cosLatRad := math.Cos(latRad)
	sinRandomDistance := math.Sin(randomDistance)
	cosRandomDistance := math.Cos(randomDistance)
	cosRandomBearing := math.Cos(randomBearing)
	sinRandomBearing := math.Sin(randomBearing)

	newLatRad := math.Asin(sinLatRad*cosRandomDistance + cosLatRad*sinRandomDistance*cosRandomBearing)

	newLonRad := lonRad + math.Atan2(
		sinRandomBearing*sinRandomDistance*cosLatRad,
		cosRandomDistance-sinLatRad*math.Sin(newLatRad),
	)

	return radiansToDegrees(newLatRad), radiansToDegrees(newLonRad)
}

func degreesToRadians(degrees float64) float64 {
	return degrees * math.Pi / 180
}

func radiansToDegrees(radians float64) float64 {
	return radians * 180 / math.Pi
}
