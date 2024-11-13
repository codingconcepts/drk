package model

import (
	"fmt"

	"github.com/samber/lo"
)

func defaultWeights(items int) []int {
	result := make([]int, items)
	baseValue := 100 / items
	remainder := 100 % items

	for i := range result {
		result[i] = baseValue
	}

	for i := 0; i < remainder; i++ {
		result[i]++
	}

	return result
}

func buildWeightedItems(values []any, weights []int) (weightedItems, error) {
	if len(values) != len(weights) {
		return weightedItems{}, fmt.Errorf("set values and weights need to be the same")
	}

	weightedItems := make([]weightedItem, len(values))
	for i, v := range values {
		weightedItems[i] = weightedItem{
			Value:  v,
			Weight: weights[i],
		}
	}

	return makeWeightedItems(weightedItems), nil
}

type weightedItem struct {
	Value  any
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

func (wi weightedItems) choose() any {
	randomWeight := Int(1, wi.totalWeight)
	for _, i := range wi.items {
		randomWeight -= i.Weight
		if randomWeight <= 0 {
			return i.Value
		}
	}

	return nil
}
