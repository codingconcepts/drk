package model

import (
	"fmt"
	"math/rand/v2"
	"os"
	"strings"
	"time"

	"github.com/codingconcepts/drk/pkg/random"
	"github.com/expr-lang/expr"
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
			min, max, err := parseMinMax[int](raw)
			if err != nil {
				return nil, err
			}

			return Int(min, max), nil

		case "float":
			min, max, err := parseMinMax[float64](raw)
			if err != nil {
				return nil, err
			}

			return Float(min, max), nil

		case "timestamp":
			minStr, maxStr, err := parseMinMax[string](raw)
			if err != nil {
				return nil, err
			}

			// Parse any optional time formats but fallback to RFC3339.
			format, err := parseField[string](raw, "fmt")
			if err != nil {
				if _, ok := err.(FieldMissingErr); ok {
					format = time.RFC3339
				} else {
					return nil, err
				}
			}

			min, err := time.Parse(format, minStr)
			if err != nil {
				return nil, fmt.Errorf("parsing max as timestamp: %w", err)
			}

			max, err := time.Parse(format, maxStr)
			if err != nil {
				return nil, fmt.Errorf("parsing max as timestamp: %w", err)
			}

			return Timestamp(min, max), nil

		case "interval", "duration":
			minStr, maxStr, err := parseMinMax[string](raw)
			if err != nil {
				return nil, err
			}

			min, err := time.ParseDuration(minStr)
			if err != nil {
				return nil, fmt.Errorf("parsing min as duration: %w", err)
			}

			max, err := time.ParseDuration(maxStr)
			if err != nil {
				return nil, fmt.Errorf("parsing max as duration: %w", err)
			}

			return Interval(min, max), nil

		case "location", "point":
			lat, err := parseField[float64](raw, "lat")
			if err != nil {
				return nil, err
			}

			lon, err := parseField[float64](raw, "lon")
			if err != nil {
				return nil, err
			}

			distanceKM, err := parseField[float64](raw, "distance_km")
			if err != nil {
				return nil, err
			}

			randomLat, randomLon := Point(lat, lon, distanceKM)
			return LatLon{Lat: randomLat, Lon: randomLon}, nil

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
		vu.dataMu.RLock()
		defer vu.dataMu.RUnlock()

		query, ok := vu.data[queryRef]
		if !ok {
			return nil, fmt.Errorf("missing query: %q", queryRef)
		}

		if len(query) == 0 {
			return nil, fmt.Errorf("no data found for %s - %s", queryRef, columnRef)
		}

		row := rand.IntN(len(query))
		cell, ok := query[row][columnRef]
		if !ok {
			return nil, fmt.Errorf("missing column: %q", columnRef)
		}

		return cell, nil
	}

	depFunc := func(vu *VU) bool {
		vu.dataMu.RLock()
		defer vu.dataMu.RUnlock()

		data, ok := vu.data[queryRef]
		if !ok || len(data) == 0 {
			return false
		}

		_, ok = data[0][columnRef]
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
		return weightedItems.choose(), nil
	}

	return genFunc, dependencyFuncNoop, nil
}

func parseArgTypeConst(raw map[string]any) (genFunc, dependencyFunc, error) {
	value, err := parseField[any](raw, "value")
	if err != nil {
		return nil, nil, fmt.Errorf("parsing value: %w", err)
	}

	genFunc := func(vu *VU) (any, error) {
		return value, nil
	}

	return genFunc, dependencyFuncNoop, nil
}

func parseArgTypeEnv(raw map[string]any) (genFunc, dependencyFunc, error) {
	envVarName, err := parseField[string](raw, "name")
	if err != nil {
		return nil, nil, fmt.Errorf("parsing env var name: %w", err)
	}

	value, ok := os.LookupEnv(envVarName)
	if !ok {
		return nil, nil, fmt.Errorf("missing env var: %q", envVarName)
	}

	genFunc := func(vu *VU) (any, error) {
		to, ok := vu.envMapper(envVarName, value)
		if ok {
			return to, nil
		}

		return value, nil
	}

	return genFunc, dependencyFuncNoop, nil
}

func parseArgTypeExpr(raw map[string]any) (genFunc, dependencyFunc, error) {
	value, err := parseField[string](raw, "value")
	if err != nil {
		return nil, nil, fmt.Errorf("parsing value: %w", err)
	}

	env := map[string]any{
		"env": func(name string) string {
			return os.Getenv(name)
		},
	}

	program, err := expr.Compile(value, expr.Env(env))
	if err != nil {
		return nil, nil, fmt.Errorf("compiling expression: %w", err)
	}

	genFunc := func(vu *VU) (any, error) {
		return expr.Run(program, env)
	}

	return genFunc, dependencyFuncNoop, nil
}

func parseArgTypeGlobal(raw map[string]any) (genFunc, dependencyFunc, error) {
	argName, err := parseField[string](raw, "name")
	if err != nil {
		return nil, nil, fmt.Errorf("parsing env var name: %w", err)
	}

	genFunc := func(vu *VU) (any, error) {
		to, ok := vu.r.globalArgs.get(argName)
		if !ok {
			return nil, fmt.Errorf("missing global arg: %q", argName)
		}

		return to, nil
	}

	return genFunc, dependencyFuncNoop, nil
}

func parseMinMax[T any](raw map[string]any) (T, T, error) {
	min, err := parseField[T](raw, "min")
	if err != nil {
		return *new(T), *new(T), fmt.Errorf("parsing min: %w", err)
	}

	max, err := parseField[T](raw, "max")
	if err != nil {
		return *new(T), *new(T), fmt.Errorf("parsing max: %w", err)
	}

	return min, max, nil
}

func parseField[T any](m map[string]any, key string) (T, error) {
	valueRaw, ok := m[key]
	if !ok {
		return *new(T), FieldMissingErr{Name: key}
	}

	value, ok := valueRaw.(T)
	if !ok {
		return *new(T), fmt.Errorf("field type mismatch (got: %T exp: %T)", valueRaw, *new(T))
	}

	return value, nil
}
