package model

import (
	"testing"
	"time"

	"github.com/codingconcepts/drk/pkg/test"
	"github.com/stretchr/testify/assert"
)

func TestInt(t *testing.T) {
	cases := []struct {
		name    string
		min     int
		max     int
		expFunc func(t *testing.T, val int)
	}{
		{
			name: "min and max are equal",
			min:  1,
			max:  1,
			expFunc: func(t *testing.T, val int) {
				assert.Equal(t, 1, val)
			},
		},
		{
			name: "min lt max",
			min:  1,
			max:  10,
			expFunc: func(t *testing.T, val int) {
				test.NumberBetween(t, val, 1, 100)
			},
		},
		{
			name: "min gt max",
			min:  10,
			max:  1,
			expFunc: func(t *testing.T, val int) {
				test.NumberBetween(t, val, 1, 100)
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			act := Int(c.min, c.max)
			c.expFunc(t, act)
		})
	}
}

func TestFloat(t *testing.T) {
	cases := []struct {
		name    string
		min     float64
		max     float64
		expFunc func(t *testing.T, val float64)
	}{
		{
			name: "min and max are equal",
			min:  1.0,
			max:  1.0,
			expFunc: func(t *testing.T, val float64) {
				assert.Equal(t, 1.0, val)
			},
		},
		{
			name: "min lt max",
			min:  1.0,
			max:  10.0,
			expFunc: func(t *testing.T, val float64) {
				test.NumberBetween(t, val, 1.0, 100.00)
			},
		},
		{
			name: "min gt max",
			min:  10.0,
			max:  1.0,
			expFunc: func(t *testing.T, val float64) {
				test.NumberBetween(t, val, 1.0, 100.00)
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			act := Float(c.min, c.max)
			c.expFunc(t, act)
		})
	}
}

func TestTimestamp(t *testing.T) {
	cases := []struct {
		name    string
		min     time.Time
		max     time.Time
		expFunc func(t *testing.T, val time.Time)
	}{
		{
			name: "min and max are equal",
			min:  time.Date(2024, 11, 13, 9, 54, 32, 0, time.UTC),
			max:  time.Date(2024, 11, 13, 9, 54, 32, 0, time.UTC),
			expFunc: func(t *testing.T, val time.Time) {
				assert.Equal(t, time.Date(2024, 11, 13, 9, 54, 32, 0, time.UTC), val)
			},
		},
		{
			name: "min lt max",
			min:  time.Date(2023, 11, 13, 9, 54, 32, 0, time.UTC),
			max:  time.Date(2024, 11, 13, 9, 54, 32, 0, time.UTC),
			expFunc: func(t *testing.T, val time.Time) {
				min := time.Date(2023, 11, 13, 9, 54, 32, 0, time.UTC)
				max := time.Date(2024, 11, 13, 9, 54, 32, 0, time.UTC)
				test.TimestampBetween(t, val, min, max)
			},
		},
		{
			name: "min gt max",
			min:  time.Date(2024, 11, 13, 9, 54, 32, 0, time.UTC),
			max:  time.Date(2023, 11, 13, 9, 54, 32, 0, time.UTC),
			expFunc: func(t *testing.T, val time.Time) {
				min := time.Date(2023, 11, 13, 9, 54, 32, 0, time.UTC)
				max := time.Date(2024, 11, 13, 9, 54, 32, 0, time.UTC)
				test.TimestampBetween(t, val, min, max)
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			act := Timestamp(c.min, c.max)
			c.expFunc(t, act)
		})
	}
}

func TestInterval(t *testing.T) {
	cases := []struct {
		name    string
		min     time.Duration
		max     time.Duration
		expFunc func(t *testing.T, val time.Duration)
	}{
		{
			name: "min and max are equal",
			min:  time.Hour * 1,
			max:  time.Hour * 1,
			expFunc: func(t *testing.T, val time.Duration) {
				assert.Equal(t, time.Hour*1, val)
			},
		},
		{
			name: "min lt max",
			min:  time.Hour * 1,
			max:  time.Hour * 10,
			expFunc: func(t *testing.T, val time.Duration) {
				test.NumberBetween(t, val, time.Hour*1, time.Hour*10)
			},
		},
		{
			name: "min gt max",
			min:  time.Hour * 10,
			max:  time.Hour * 1,
			expFunc: func(t *testing.T, val time.Duration) {
				test.NumberBetween(t, val, time.Hour*1, time.Hour*10)
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			act := Interval(c.min, c.max)
			c.expFunc(t, act)
		})
	}
}
