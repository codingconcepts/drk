package model

import (
	"fmt"
	"testing"
	"time"

	"github.com/codingconcepts/drk/pkg/test"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestParseArgTypeGen(t *testing.T) {
	cases := []struct {
		name             string
		raw              map[string]any
		genFuncValidator func(t *testing.T, f genFunc)
		depFuncValidator func(t *testing.T, f dependencyFunc)
		expErr           error
	}{
		{
			name: "valid gen value",
			raw: map[string]any{
				"value": "email",
			},
			genFuncValidator: func(t *testing.T, f genFunc) {
				raw, err := f(nil)
				assert.NoError(t, err)

				value := raw.(string)
				assert.Contains(t, value, "@")
			},
			depFuncValidator: func(t *testing.T, f dependencyFunc) {
				assert.True(t, f(nil))
			},
		},
		{
			name:   "missing value",
			raw:    map[string]any{},
			expErr: fmt.Errorf("parsing value: %w", FieldMissingErr{Name: "value"}),
		},
		{
			name: "missing generator",
			raw: map[string]any{
				"value": "invalid_generator",
			},
			genFuncValidator: func(t *testing.T, f genFunc) {
				_, err := f(nil)

				expErr := fmt.Errorf("missing generator: \"invalid_generator\"")
				assert.Equal(t, expErr, err)
			},
			depFuncValidator: func(t *testing.T, f dependencyFunc) {
				assert.True(t, f(nil))
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			gen, dep, err := parseArgTypeGen(c.raw)
			assert.Equal(t, c.expErr, err)
			if err != nil {
				return
			}

			c.genFuncValidator(t, gen)
			c.depFuncValidator(t, dep)
		})
	}
}

func TestParseArgTypeScalar(t *testing.T) {
	cases := []struct {
		name             string
		argType          string
		raw              map[string]any
		genFuncValidator func(t *testing.T, f genFunc)
		depFuncValidator func(t *testing.T, f dependencyFunc)
		expErr           error
	}{
		{
			name:    "missing min",
			argType: "int",
			raw: map[string]any{
				"max": 10,
			},
			genFuncValidator: func(t *testing.T, f genFunc) {
				raw, err := f(nil)
				exp := fmt.Errorf("parsing min: %w", FieldMissingErr{Name: "min"})
				assert.Equal(t, exp, err)
				assert.Nil(t, raw)
			},
			depFuncValidator: func(t *testing.T, f dependencyFunc) {
				assert.True(t, f(nil))
			},
		},
		{
			name:    "invalid min",
			argType: "int",
			raw: map[string]any{
				"min": "invalid",
				"max": 10,
			},
			genFuncValidator: func(t *testing.T, f genFunc) {
				raw, err := f(nil)
				exp := "parsing min: field type mismatch (got: string exp: int)"
				assert.Equal(t, exp, err.Error())
				assert.Nil(t, raw)
			},
			depFuncValidator: func(t *testing.T, f dependencyFunc) {
				assert.True(t, f(nil))
			},
		},
		{
			name:    "missing max",
			argType: "int",
			raw: map[string]any{
				"min": 10,
			},
			genFuncValidator: func(t *testing.T, f genFunc) {
				raw, err := f(nil)
				exp := fmt.Errorf("parsing max: %w", FieldMissingErr{Name: "max"})
				assert.Equal(t, exp, err)
				assert.Nil(t, raw)
			},
			depFuncValidator: func(t *testing.T, f dependencyFunc) {
				assert.True(t, f(nil))
			},
		},
		{
			name:    "invalid max",
			argType: "int",
			raw: map[string]any{
				"min": 10,
				"max": "invalid",
			},
			genFuncValidator: func(t *testing.T, f genFunc) {
				raw, err := f(nil)
				exp := "parsing max: field type mismatch (got: string exp: int)"
				assert.Equal(t, exp, err.Error())
				assert.Nil(t, raw)
			},
			depFuncValidator: func(t *testing.T, f dependencyFunc) {
				assert.True(t, f(nil))
			},
		},
		{
			name:    "valid int generator - min eq max",
			argType: "int",
			raw: map[string]any{
				"min": 10,
				"max": 10,
			},
			genFuncValidator: func(t *testing.T, f genFunc) {
				raw, err := f(nil)
				assert.NoError(t, err)

				assert.Equal(t, 10, raw.(int))
			},
			depFuncValidator: func(t *testing.T, f dependencyFunc) {
				assert.True(t, f(nil))
			},
		},
		{
			name:    "missing min",
			argType: "float",
			raw: map[string]any{
				"max": 10.0,
			},
			genFuncValidator: func(t *testing.T, f genFunc) {
				raw, err := f(nil)
				exp := fmt.Errorf("parsing min: %w", FieldMissingErr{Name: "min"})
				assert.Equal(t, exp, err)
				assert.Nil(t, raw)
			},
			depFuncValidator: func(t *testing.T, f dependencyFunc) {
				assert.True(t, f(nil))
			},
		},
		{
			name:    "invalid min",
			argType: "float",
			raw: map[string]any{
				"min": "invalid",
				"max": 10.0,
			},
			genFuncValidator: func(t *testing.T, f genFunc) {
				raw, err := f(nil)
				exp := "parsing min: field type mismatch (got: string exp: float64)"
				assert.Equal(t, exp, err.Error())
				assert.Nil(t, raw)
			},
			depFuncValidator: func(t *testing.T, f dependencyFunc) {
				assert.True(t, f(nil))
			},
		},
		{
			name:    "missing max",
			argType: "float",
			raw: map[string]any{
				"min": 10.0,
			},
			genFuncValidator: func(t *testing.T, f genFunc) {
				raw, err := f(nil)
				exp := fmt.Errorf("parsing max: %w", FieldMissingErr{Name: "max"})
				assert.Equal(t, exp, err)
				assert.Nil(t, raw)
			},
			depFuncValidator: func(t *testing.T, f dependencyFunc) {
				assert.True(t, f(nil))
			},
		},
		{
			name:    "invalid max",
			argType: "float",
			raw: map[string]any{
				"min": 10.0,
				"max": "invalid",
			},
			genFuncValidator: func(t *testing.T, f genFunc) {
				raw, err := f(nil)
				exp := "parsing max: field type mismatch (got: string exp: float64)"
				assert.Equal(t, exp, err.Error())
				assert.Nil(t, raw)
			},
			depFuncValidator: func(t *testing.T, f dependencyFunc) {
				assert.True(t, f(nil))
			},
		},
		{
			name:    "valid float generator - min eq max",
			argType: "float",
			raw: map[string]any{
				"min": 10.0,
				"max": 10.0,
			},
			genFuncValidator: func(t *testing.T, f genFunc) {
				raw, err := f(nil)
				assert.NoError(t, err)

				assert.Equal(t, 10.0, raw.(float64))
			},
			depFuncValidator: func(t *testing.T, f dependencyFunc) {
				assert.True(t, f(nil))
			},
		},
		{
			name:    "missing min",
			argType: "timestamp",
			raw: map[string]any{
				"max": "2024-11-12T19:13:07Z",
			},
			genFuncValidator: func(t *testing.T, f genFunc) {
				raw, err := f(nil)
				exp := fmt.Errorf("parsing min: %w", FieldMissingErr{Name: "min"})
				assert.Equal(t, exp, err)
				assert.Nil(t, raw)
			},
			depFuncValidator: func(t *testing.T, f dependencyFunc) {
				assert.True(t, f(nil))
			},
		},
		{
			name:    "invalid min",
			argType: "timestamp",
			raw: map[string]any{
				"min": "invalid",
				"max": "2024-11-12T19:13:07Z",
			},
			genFuncValidator: func(t *testing.T, f genFunc) {
				raw, err := f(nil)
				exp := "parsing max as timestamp: parsing time \"invalid\" as \"2006-01-02T15:04:05Z07:00\": cannot parse \"invalid\" as \"2006\""
				assert.Equal(t, exp, err.Error())
				assert.Nil(t, raw)
			},
			depFuncValidator: func(t *testing.T, f dependencyFunc) {
				assert.True(t, f(nil))
			},
		},
		{
			name:    "missing max",
			argType: "timestamp",
			raw: map[string]any{
				"min": "2024-11-12T19:13:07Z",
			},
			genFuncValidator: func(t *testing.T, f genFunc) {
				raw, err := f(nil)
				exp := fmt.Errorf("parsing max: %w", FieldMissingErr{Name: "max"})
				assert.Equal(t, exp, err)
				assert.Nil(t, raw)
			},
			depFuncValidator: func(t *testing.T, f dependencyFunc) {
				assert.True(t, f(nil))
			},
		},
		{
			name:    "invalid max",
			argType: "timestamp",
			raw: map[string]any{
				"min": "2024-11-12T19:13:07Z",
				"max": "invalid",
			},
			genFuncValidator: func(t *testing.T, f genFunc) {
				raw, err := f(nil)
				exp := "parsing max as timestamp: parsing time \"invalid\" as \"2006-01-02T15:04:05Z07:00\": cannot parse \"invalid\" as \"2006\""
				assert.Equal(t, exp, err.Error())
				assert.Nil(t, raw)
			},
			depFuncValidator: func(t *testing.T, f dependencyFunc) {
				assert.True(t, f(nil))
			},
		},
		{
			name:    "valid timestamp generator - min eq max",
			argType: "timestamp",
			raw: map[string]any{
				"min": "2024-11-12T19:13:07Z",
				"max": "2024-11-12T19:13:07Z",
			},
			genFuncValidator: func(t *testing.T, f genFunc) {
				raw, err := f(nil)
				assert.NoError(t, err)

				exp := time.Date(2024, 11, 12, 19, 13, 7, 0, time.UTC)
				assert.Equal(t, exp, raw.(time.Time))
			},
			depFuncValidator: func(t *testing.T, f dependencyFunc) {
				assert.True(t, f(nil))
			},
		},
		{
			name:    "missing min",
			argType: "interval",
			raw: map[string]any{
				"max": "1h",
			},
			genFuncValidator: func(t *testing.T, f genFunc) {
				raw, err := f(nil)
				exp := fmt.Errorf("parsing min: %w", FieldMissingErr{Name: "min"})
				assert.Equal(t, exp, err)
				assert.Nil(t, raw)
			},
			depFuncValidator: func(t *testing.T, f dependencyFunc) {
				assert.True(t, f(nil))
			},
		},
		{
			name:    "invalid min",
			argType: "interval",
			raw: map[string]any{
				"min": "invalid",
				"max": "1h",
			},
			genFuncValidator: func(t *testing.T, f genFunc) {
				raw, err := f(nil)
				exp := "parsing min as duration: time: invalid duration \"invalid\""
				assert.Equal(t, exp, err.Error())
				assert.Nil(t, raw)
			},
			depFuncValidator: func(t *testing.T, f dependencyFunc) {
				assert.True(t, f(nil))
			},
		},
		{
			name:    "missing max",
			argType: "interval",
			raw: map[string]any{
				"min": "1h",
			},
			genFuncValidator: func(t *testing.T, f genFunc) {
				raw, err := f(nil)
				exp := fmt.Errorf("parsing max: %w", FieldMissingErr{Name: "max"})
				assert.Equal(t, exp, err)
				assert.Nil(t, raw)
			},
			depFuncValidator: func(t *testing.T, f dependencyFunc) {
				assert.True(t, f(nil))
			},
		},
		{
			name:    "invalid max",
			argType: "interval",
			raw: map[string]any{
				"min": "1h",
				"max": "invalid",
			},
			genFuncValidator: func(t *testing.T, f genFunc) {
				raw, err := f(nil)
				exp := "parsing max as duration: time: invalid duration \"invalid\""
				assert.Equal(t, exp, err.Error())
				assert.Nil(t, raw)
			},
			depFuncValidator: func(t *testing.T, f dependencyFunc) {
				assert.True(t, f(nil))
			},
		},
		{
			name:    "valid interval generator - min eq max",
			argType: "interval",
			raw: map[string]any{
				"min": "1h2m3s",
				"max": "1h2m3s",
			},
			genFuncValidator: func(t *testing.T, f genFunc) {
				raw, err := f(nil)
				assert.NoError(t, err)

				exp := time.Duration(1*time.Hour + 2*time.Minute + 3*time.Second)
				assert.Equal(t, exp, raw.(time.Duration))
			},
			depFuncValidator: func(t *testing.T, f dependencyFunc) {
				assert.True(t, f(nil))
			},
		},
		{
			name:    "valid int generator - min lt max",
			argType: "int",
			raw: map[string]any{
				"min": 10,
				"max": 100,
			},
			genFuncValidator: func(t *testing.T, f genFunc) {
				raw, err := f(nil)
				assert.NoError(t, err)

				act := raw.(int)
				test.NumberBetween(t, act, 10, 100)
			},
			depFuncValidator: func(t *testing.T, f dependencyFunc) {
				assert.True(t, f(nil))
			},
		},
		{
			name:    "valid float generator - min lt max",
			argType: "float",
			raw: map[string]any{
				"min": 10.0,
				"max": 100.0,
			},
			genFuncValidator: func(t *testing.T, f genFunc) {
				raw, err := f(nil)
				assert.NoError(t, err)

				act := raw.(float64)
				test.NumberBetween(t, act, 10.0, 100.0)
			},
			depFuncValidator: func(t *testing.T, f dependencyFunc) {
				assert.True(t, f(nil))
			},
		},
		{
			name:    "valid timestamp generator - min lt max",
			argType: "timestamp",
			raw: map[string]any{
				"min": "2023-11-12T19:13:07Z",
				"max": "2024-11-12T19:13:07Z",
			},
			genFuncValidator: func(t *testing.T, f genFunc) {
				raw, err := f(nil)
				assert.NoError(t, err)

				act := raw.(time.Time)
				min := time.Date(2023, 11, 12, 19, 13, 7, 0, time.UTC)
				max := time.Date(2024, 11, 12, 19, 13, 7, 0, time.UTC)
				test.TimestampBetween(t, act, min, max)
			},
			depFuncValidator: func(t *testing.T, f dependencyFunc) {
				assert.True(t, f(nil))
			},
		},
		{
			name:    "valid interval generator - min lt max",
			argType: "interval",
			raw: map[string]any{
				"min": "1h2m3s",
				"max": "2h3m4s",
			},
			genFuncValidator: func(t *testing.T, f genFunc) {
				raw, err := f(nil)
				assert.NoError(t, err)

				act := raw.(time.Duration)
				min := time.Hour*1 + time.Minute*2 + time.Second*3
				max := time.Hour*2 + time.Minute*3 + time.Second*4
				test.NumberBetween(t, act, min, max)
			},
			depFuncValidator: func(t *testing.T, f dependencyFunc) {
				assert.True(t, f(nil))
			},
		},
		{
			name:    "unsupported scalar type",
			argType: "unsupported",
			genFuncValidator: func(t *testing.T, f genFunc) {
				raw, err := f(nil)
				exp := "invalid scalar generator: \"unsupported\""
				assert.Equal(t, exp, err.Error())
				assert.Nil(t, raw)
			},
			depFuncValidator: func(t *testing.T, f dependencyFunc) {
				assert.True(t, f(nil))
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			gen, dep, err := parseArgTypeScalar(c.argType, c.raw)
			assert.Equal(t, c.expErr, err)
			if err != nil {
				return
			}

			c.genFuncValidator(t, gen)
			c.depFuncValidator(t, dep)
		})
	}
}

func TestParseArgTypeRef(t *testing.T) {
	cases := []struct {
		name             string
		raw              map[string]any
		genFuncValidator func(*testing.T, genFunc, *VU)
		depFuncValidator func(*testing.T, dependencyFunc, *VU)
		expErr           error
	}{
		{
			name: "valid ref value",
			raw: map[string]any{
				"query":  "table",
				"column": "column",
			},
			genFuncValidator: func(t *testing.T, f genFunc, vu *VU) {
				raw, err := f(vu)
				assert.NoError(t, err)

				value := raw.(string)
				assert.Equal(t, value, "a")
			},
			depFuncValidator: func(t *testing.T, f dependencyFunc, vu *VU) {
				assert.True(t, f(vu))
			},
		},
		{
			name:   "missing query config",
			raw:    map[string]any{},
			expErr: fmt.Errorf("parsing table: %w", FieldMissingErr{Name: "query"}),
		},
		{
			name: "missing column config",
			raw: map[string]any{
				"query": "test",
			},
			expErr: fmt.Errorf("parsing column: %w", FieldMissingErr{Name: "column"}),
		},
		{
			name: "missing data - query",
			raw: map[string]any{
				"query":  "non_existent",
				"column": "column",
			},
			genFuncValidator: func(t *testing.T, f genFunc, vu *VU) {
				raw, err := f(vu)
				assert.Nil(t, raw)
				t.Log(err)
				assert.Equal(t, fmt.Errorf("missing query: \"non_existent\""), err)
			},
			depFuncValidator: func(t *testing.T, f dependencyFunc, vu *VU) {
				assert.False(t, f(vu))
			},
		},
		{
			name: "missing data - column",
			raw: map[string]any{
				"query":  "table",
				"column": "non_existent_column",
			},
			genFuncValidator: func(t *testing.T, f genFunc, vu *VU) {
				raw, err := f(vu)
				assert.Nil(t, raw)
				assert.Equal(t, fmt.Errorf("missing column: \"non_existent_column\""), err)
			},
			depFuncValidator: func(t *testing.T, f dependencyFunc, vu *VU) {
				assert.False(t, f(vu))
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			gen, dep, err := parseArgTypeRef(c.raw)
			assert.Equal(t, c.expErr, err)
			if err != nil {
				return
			}

			vu := NewVU(&zerolog.Logger{})
			vu.data = map[string][]map[string]any{
				"table": {
					{
						"column": "a",
					},
				},
			}

			c.genFuncValidator(t, gen, vu)
			c.depFuncValidator(t, dep, vu)
		})
	}
}

func TestParseArgTypeSet(t *testing.T) {
	cases := []struct {
		name             string
		raw              map[string]any
		genFuncValidator func(*testing.T, genFunc, *VU)
		depFuncValidator func(*testing.T, dependencyFunc, *VU)
		expErr           error
	}{
		{
			name: "valid string values without weights",
			raw: map[string]any{
				"values": []any{"a", "b", "c"},
			},
			genFuncValidator: func(t *testing.T, f genFunc, vu *VU) {
				raw, err := f(vu)
				assert.NoError(t, err)

				assert.Contains(t, []any{"a", "b", "c"}, raw)
			},
			depFuncValidator: func(t *testing.T, f dependencyFunc, vu *VU) {
				assert.True(t, f(vu))
			},
		},
		{
			name: "valid string values with weights",
			raw: map[string]any{
				"values":  []any{"a", "b", "c"},
				"weights": []any{100, 0, 0},
			},
			genFuncValidator: func(t *testing.T, f genFunc, vu *VU) {
				raw, err := f(vu)
				assert.NoError(t, err)

				assert.Equal(t, "a", raw)
			},
			depFuncValidator: func(t *testing.T, f dependencyFunc, vu *VU) {
				assert.True(t, f(vu))
			},
		},
		{
			name: "valid int values with weights",
			raw: map[string]any{
				"values": []any{1, 2, 3},
			},
			genFuncValidator: func(t *testing.T, f genFunc, vu *VU) {
				raw, err := f(vu)
				assert.NoError(t, err)

				assert.Contains(t, []any{1, 2, 3}, raw)
			},
			depFuncValidator: func(t *testing.T, f dependencyFunc, vu *VU) {
				assert.True(t, f(vu))
			},
		},
		{
			name:   "missing values",
			raw:    map[string]any{},
			expErr: fmt.Errorf("parsing values: %w", FieldMissingErr{Name: "values"}),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			gen, dep, err := parseArgTypeSet(c.raw)
			assert.Equal(t, c.expErr, err)
			if err != nil {
				return
			}

			vu := NewVU(&zerolog.Logger{})

			c.genFuncValidator(t, gen, vu)
			c.depFuncValidator(t, dep, vu)
		})
	}
}

func TestParseArgTypeConst(t *testing.T) {
	cases := []struct {
		name             string
		raw              map[string]any
		genFuncValidator func(*testing.T, genFunc, *VU)
		depFuncValidator func(*testing.T, dependencyFunc, *VU)
		expErr           error
	}{
		{
			name: "valid string value",
			raw: map[string]any{
				"value": "a",
			},
			genFuncValidator: func(t *testing.T, f genFunc, vu *VU) {
				raw, err := f(vu)
				assert.NoError(t, err)

				assert.Equal(t, "a", raw)
			},
			depFuncValidator: func(t *testing.T, f dependencyFunc, vu *VU) {
				assert.True(t, f(vu))
			},
		},
		{
			name: "valid int value",
			raw: map[string]any{
				"value": 1,
			},
			genFuncValidator: func(t *testing.T, f genFunc, vu *VU) {
				raw, err := f(vu)
				assert.NoError(t, err)

				assert.Equal(t, 1, raw)
			},
			depFuncValidator: func(t *testing.T, f dependencyFunc, vu *VU) {
				assert.True(t, f(vu))
			},
		},
		{
			name: "valid float value",
			raw: map[string]any{
				"value": 1.0,
			},
			genFuncValidator: func(t *testing.T, f genFunc, vu *VU) {
				raw, err := f(vu)
				assert.NoError(t, err)

				assert.Equal(t, 1.0, raw)
			},
			depFuncValidator: func(t *testing.T, f dependencyFunc, vu *VU) {
				assert.True(t, f(vu))
			},
		},
		{
			name: "valid bool value",
			raw: map[string]any{
				"value": true,
			},
			genFuncValidator: func(t *testing.T, f genFunc, vu *VU) {
				raw, err := f(vu)
				assert.NoError(t, err)

				assert.Equal(t, true, raw)
			},
			depFuncValidator: func(t *testing.T, f dependencyFunc, vu *VU) {
				assert.True(t, f(vu))
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			gen, dep, err := parseArgTypeConst(c.raw)
			assert.Equal(t, c.expErr, err)
			if err != nil {
				return
			}

			vu := NewVU(&zerolog.Logger{})

			c.genFuncValidator(t, gen, vu)
			c.depFuncValidator(t, dep, vu)
		})
	}
}
