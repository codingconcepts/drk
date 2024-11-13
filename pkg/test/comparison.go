package test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/constraints"
)

func NumberBetween[T constraints.Integer | constraints.Float](t *testing.T, val, min, max T) {
	assert.Truef(
		t, val >= min && val <= max,
		"out of range (%d <= %d <= %d)",
		min, val, max,
	)
}

func TimestampBetween(t *testing.T, val, min, max time.Time) {
	assert.Truef(t,
		(val.Equal(min) || val.After(min)) && (val.Equal(max) || val.Before(max)),
		"out of range (%s <= %s <= %s)",
		min, val, max,
	)
}
