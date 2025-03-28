package model

import (
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/samber/lo"
)

type VU struct {
	// Link to runner for global arg fetching.
	r *Runner

	// Map of query names to columns to rows.
	dataMu sync.RWMutex
	data   map[string][]map[string]any

	envMapper envMappingGenerator

	logger *zerolog.Logger
}

func NewVU(r *Runner) *VU {
	return &VU{
		r:         r,
		data:      map[string][]map[string]any{},
		envMapper: r.envMappings,
		logger:    r.logger,
	}
}

func (vu *VU) stagger(queries []WorkflowQuery) {
	// Stagger using any time between now and the max query tick.
	maxTicks := lo.MaxBy(queries, func(a, b WorkflowQuery) bool {
		return a.Rate.tickerInterval > b.Rate.tickerInterval
	})

	staggerDuration := Interval(0, maxTicks.Rate.tickerInterval)
	time.Sleep(staggerDuration)
}

func (vu *VU) applyData(query string, data []map[string]any) {
	vu.dataMu.Lock()
	defer vu.dataMu.Unlock()

	vu.data[query] = data
}

func (vu *VU) generateArgs(args []Arg) ([]any, error) {
	var values []any

	for _, arg := range args {
		v, err := arg.generator(vu)
		if err != nil {
			return nil, fmt.Errorf("generating value for arg: %w", err)
		}

		values = append(values, v)
	}

	return values, nil
}

func (vu *VU) generateNamedArgs(args map[string]Arg) (map[string]any, error) {
	values := map[string]any{}

	for name, arg := range args {
		v, err := arg.generator(vu)
		if err != nil {
			return nil, fmt.Errorf("generating value for arg: %w", err)
		}

		values[name] = v
	}

	return values, nil
}
