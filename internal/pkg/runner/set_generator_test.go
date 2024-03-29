package runner

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateSetColumn(t *testing.T) {
	g := SetGenerator{
		Values:  []string{"a", "b", "c"},
		Weights: []int{0, 1, 0},
	}

	act := g.Generate()

	assert.Equal(t,
		[][]string{{"b", "b", "b", "b", "b", "b", "b", "b", "b", "b"}},
		act,
	)
}

func TestMakeWeightedItems(t *testing.T) {
	items := makeWeightedItems(
		[]weightedItem{
			{Value: "a", Weight: 10},
			{Value: "b", Weight: 20},
			{Value: "c", Weight: 30},
		},
	)

	assert.Equal(t, 60, items.totalWeight)
}

func TestChoose(t *testing.T) {
	cases := []struct {
		name  string
		items []weightedItem
		exp   []string
	}{
		{
			name: "3 items 1 has all the weight",
			items: []weightedItem{
				{Value: "a", Weight: 100},
				{Value: "b", Weight: 0},
				{Value: "c", Weight: 0},
			},
			exp: []string{"a", "a", "a", "a", "a", "a", "a", "a", "a", "a"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			items := makeWeightedItems(c.items)

			var act []string
			for i := 0; i < 10; i++ {
				act = append(act, items.choose())
			}

			assert.Equal(t, c.exp, act)
		})
	}
}
