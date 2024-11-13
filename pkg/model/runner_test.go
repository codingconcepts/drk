package model

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestRunQuery(t *testing.T) {
	cases := []struct {
		name      string
		query     Query
		queryImpl func(string, ...any) ([]map[string]any, time.Duration, error)
		execImpl  func(string, ...any) (time.Duration, error)
		exp       []map[string]any
		expError  error
	}{
		{
			name: "error generating args",
			query: Query{
				Args: []Arg{
					{
						generator: func(_ *VU) (any, error) {
							return nil, errors.New("bad things happened")
						},
					},
				},
			},
			expError: fmt.Errorf("generating args: %w", fmt.Errorf("generating value for arg: %w", errors.New("bad things happened"))),
		},
		{
			name: "unsupported query type",
			query: Query{
				Type: "invalid",
			},
			expError: fmt.Errorf("unsupported query type: \"invalid\""),
		},
		{
			name: "error running query",
			query: Query{
				Type: "query",
			},
			queryImpl: func(s string, a ...any) ([]map[string]any, time.Duration, error) {
				return nil, 0, fmt.Errorf("bad things happened")
			},
			expError: errors.New("bad things happened"),
		},
		{
			name: "error running exec",
			query: Query{
				Type: "exec",
			},
			execImpl: func(s string, a ...any) (time.Duration, error) {
				return 0, fmt.Errorf("bad things happened")
			},
			expError: errors.New("bad things happened"),
		},
		{
			name: "query returns data",
			query: Query{
				Type: "query",
			},
			queryImpl: func(s string, a ...any) ([]map[string]any, time.Duration, error) {
				return []map[string]any{
					{"id": "a", "age": 1},
					{"id": "b", "age": 2},
					{"id": "c", "age": 3},
				}, 0, nil
			},
			exp: []map[string]any{
				{"id": "a", "age": 1},
				{"id": "b", "age": 2},
				{"id": "c", "age": 3},
			},
		},
		{
			name: "exec returns no data",
			query: Query{
				Type: "exec",
			},
			execImpl: func(s string, a ...any) (time.Duration, error) {
				return 0, nil
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			queryer := mockQueryer{
				query: c.queryImpl,
				exec:  c.execImpl,
			}

			r, err := NewRunner(nil, &queryer, "", "", 0, &zerolog.Logger{})
			assert.NoError(t, err)

			vu := NewVU(&zerolog.Logger{})
			act, _, err := r.runQuery(vu, c.query)

			if c.expError != nil {
				assert.Equal(t, c.expError, err)
				return
			}

			assert.Equal(t, c.exp, act)
		})
	}
}
