package model

import (
	"fmt"
	"sync"
	"time"

	"github.com/codingconcepts/drk/pkg/random"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
)

type VU struct {
	// Map of query names to columns to rows.
	dataMu sync.RWMutex
	data   map[string][]map[string]any

	logger *zerolog.Logger
}

func NewVU(logger *zerolog.Logger) *VU {
	return &VU{
		data:   map[string][]map[string]any{},
		logger: logger,
	}
}

func (vu *VU) stagger(queries []WorkflowQuery) {
	// Stagger using any time between now and the average query tick.
	sumTicks := lo.SumBy(queries, func(a WorkflowQuery) time.Duration {
		return a.Rate.tickerInterval
	})

	avgTicks := sumTicks / time.Duration(len(queries))

	staggerDuration := random.Interval(0, avgTicks)
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
