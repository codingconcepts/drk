package model

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultWeights(t *testing.T) {
	cases := []struct {
		name  string
		items int
		exp   []int
	}{
		{
			name:  "1 item",
			items: 1,
			exp:   []int{100},
		},
		{
			name:  "2 items",
			items: 2,
			exp:   []int{50, 50},
		},
		{
			name:  "3 items",
			items: 3,
			exp:   []int{34, 33, 33},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			act := defaultWeights(c.items)
			assert.Equal(t, c.exp, act)
		})
	}
}

func TestBuildWeightedItems(t *testing.T) {
	cases := []struct {
		name    string
		values  []any
		weights []int
		exp     weightedItems
		expErr  error
	}{
		{
			name:    "unequal values and weights",
			values:  []any{"a", "b", "c", "d"},
			weights: []int{100},
			expErr:  fmt.Errorf("set values and weights need to be the same"),
		},
		{
			name:    "equal values and weights",
			values:  []any{"a", "b", "c", "d"},
			weights: []int{10, 20, 30, 40},
			exp: weightedItems{
				items: []weightedItem{
					{Value: "a", Weight: 10},
					{Value: "b", Weight: 20},
					{Value: "c", Weight: 30},
					{Value: "d", Weight: 40},
				},
				totalWeight: 100,
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			act, err := buildWeightedItems(c.values, c.weights)
			assert.Equal(t, c.expErr, err)
			if err != nil {
				return
			}

			assert.Equal(t, c.exp, act)
		})
	}
}
