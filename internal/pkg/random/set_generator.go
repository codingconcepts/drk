package random

import (
	"fmt"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/samber/lo"
)

// SetGenerator provides additional context to a set column.
type SetGenerator struct {
	Values  []string `yaml:"values"`
	Weights []int    `yaml:"weights"`
}

// Generate selects between a set of values for a given table.
func (g SetGenerator) Generate() any {
	if len(g.Weights) > 0 {
		items, err := g.buildWeightedItems()
		if err != nil {
			return fmt.Errorf("making weighted items collection: %w", err)
		}

		return items.choose()
	}

	return g.Values[Intn(len(g.Values))]
}

func (g SetGenerator) buildWeightedItems() (weightedItems, error) {
	if len(g.Values) != len(g.Weights) {
		return weightedItems{}, fmt.Errorf("set values and weights need to be the same")
	}

	weightedItems := make([]weightedItem, len(g.Values))
	for i, v := range g.Values {
		weightedItems = append(weightedItems, weightedItem{
			Value:  v,
			Weight: g.Weights[i],
		})
	}

	return makeWeightedItems(weightedItems), nil
}

type weightedItem struct {
	Value  string
	Weight int
}

type weightedItems struct {
	items       []weightedItem
	totalWeight int
}

func makeWeightedItems(items []weightedItem) weightedItems {
	wi := weightedItems{
		items: items,
	}

	wi.totalWeight = lo.SumBy(items, func(wi weightedItem) int {
		return wi.Weight
	})

	return wi
}

func (wi weightedItems) choose() string {
	randomWeight := gofakeit.IntRange(1, wi.totalWeight)
	for _, i := range wi.items {
		randomWeight -= i.Weight
		if randomWeight <= 0 {
			return i.Value
		}
	}

	return ""
}
