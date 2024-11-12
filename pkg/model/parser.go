package model

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/codingconcepts/drk/pkg/random"
	"github.com/samber/lo"
)

func parseArgTypeGen(raw map[string]any) (genFunc, dependencyFunc, error) {
	value, err := parseField[string](raw, "value")
	if err != nil {
		return nil, nil, fmt.Errorf("parsing value: %w", err)
	}

	return func(vu *VU) (any, error) {
		g, ok := random.Replacements[value]
		if !ok {
			return nil, fmt.Errorf("missing generator: %q", value)
		}
		return g(), nil
	}, dependencyFuncNoop, nil
}

func parseArgTypeScalar(argType string, raw map[string]any) (genFunc, dependencyFunc, error) {
	return func(vu *VU) (any, error) {
		switch strings.ToLower(argType) {
		case "int":
			min, err := parseField[int](raw, "min")
			if err != nil {
				return nil, fmt.Errorf("parsing min: %w", err)
			}

			max, err := parseField[int](raw, "max")
			if err != nil {
				return nil, fmt.Errorf("parsing max: %w", err)
			}

			return Int(min, max), nil

		case "float":
			min, err := parseField[float64](raw, "min")
			if err != nil {
				return nil, fmt.Errorf("parsing min: %w", err)
			}

			max, err := parseField[float64](raw, "max")
			if err != nil {
				return nil, fmt.Errorf("parsing max: %w", err)
			}

			return Float(min, max), nil

		case "timestamp":
			minStr, err := parseField[string](raw, "min")
			if err != nil {
				return nil, fmt.Errorf("parsing min: %w", err)
			}

			maxStr, err := parseField[string](raw, "max")
			if err != nil {
				return nil, fmt.Errorf("parsing max: %w", err)
			}

			min, err := time.Parse(time.RFC3339, minStr)
			if err != nil {
				return nil, fmt.Errorf("parsing max as timestamp: %w", err)
			}

			max, err := time.Parse(time.RFC3339, maxStr)
			if err != nil {
				return nil, fmt.Errorf("parsing max as timestamp: %w", err)
			}

			return Timestamp(min, max), nil

		default:
			return nil, fmt.Errorf("invalid scalar generator: %q", argType)
		}
	}, dependencyFuncNoop, nil
}

func parseArgTypeRef(raw map[string]any) (genFunc, dependencyFunc, error) {
	queryRef, err := parseField[string](raw, "query")
	if err != nil {
		return nil, nil, fmt.Errorf("parsing table: %w", err)
	}

	columnRef, err := parseField[string](raw, "column")
	if err != nil {
		return nil, nil, fmt.Errorf("parsing column: %w", err)
	}

	genFunc := func(vu *VU) (any, error) {
		vu.logger.Debug().Msgf("[REF] gen %s - %s", queryRef, columnRef)

		vu.dataMu.RLock()
		defer vu.dataMu.RUnlock()
		query, ok := vu.data[queryRef]
		if !ok {
			return nil, fmt.Errorf("missing query: %q", query)
		}

		if len(query) == 0 {
			return nil, fmt.Errorf("no data found for %s - %s", queryRef, columnRef)
		}

		row := rand.IntN(len(query))
		cell, ok := query[row][columnRef]
		if !ok {
			return nil, fmt.Errorf("missing column: %q", cell)
		}

		return cell, nil
	}

	depFunc := func(vu *VU) bool {
		data, ok := vu.data[queryRef]
		if !ok || len(data) == 0 {
			vu.logger.Info().Str("query", queryRef).Bool("found", ok).Any("data", vu.data).Msg("missing table data")
			return false
		}

		_, ok = data[0][columnRef]
		if !ok {
			vu.logger.Info().Str("column", columnRef).Bool("found", ok).Msg("missing cell data")
		}

		return ok
	}

	return genFunc, depFunc, err
}

func parseArgTypeSet(raw map[string]any) (genFunc, dependencyFunc, error) {
	values, err := parseField[[]any](raw, "values")
	if err != nil {
		return nil, nil, fmt.Errorf("parsing values: %w", err)
	}

	var weights []int
	rawWeights, err := parseField[[]any](raw, "weights")
	if err != nil {
		if _, ok := err.(FieldMissingErr); ok {
			weights = defaultWeights(len(values))
		} else {
			return nil, nil, fmt.Errorf("parsing values: %w", err)
		}
	} else {
		weights = lo.Map(rawWeights, func(w any, _ int) int {
			return w.(int)
		})
	}

	weightedItems, err := buildWeightedItems(values, weights)
	if err != nil {
		return nil, nil, fmt.Errorf("building weighted items: %w", err)
	}

	genFunc := func(vu *VU) (any, error) {
		vu.logger.Debug().Msgf("[SET] gen %v", values)

		return weightedItems.choose(), nil
	}

	return genFunc, dependencyFuncNoop, nil
}

func parseField[T any](m map[string]any, key string) (T, error) {
	valueRaw, ok := m[key]
	if !ok {
		return *new(T), FieldMissingErr{Name: key}
	}

	value, ok := valueRaw.(T)
	if !ok {
		return *new(T), fmt.Errorf("field type mismatch (got: %T exp: %T)", value, *new(T))
	}

	return value, nil
}
