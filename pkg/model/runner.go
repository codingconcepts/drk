package model

import (
	"fmt"
	"time"

	"github.com/codingconcepts/drk/pkg/repo"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"golang.org/x/sync/errgroup"
)

const (
	initWorkflow = "init"
)

type Runner struct {
	db       repo.Queryer
	cfg      *Drk
	duration time.Duration
	events   chan Event
	logger   *zerolog.Logger
}

func NewRunner(cfg *Drk, db repo.Queryer, url, driver string, duration time.Duration, logger *zerolog.Logger) (*Runner, error) {
	r := Runner{
		db:       db,
		cfg:      cfg,
		duration: duration,
		events:   make(chan Event, 1000),
		logger:   logger,
	}

	logger.Info().Float64("duration", r.duration.Seconds()).Msgf("runner")

	return &r, nil
}

func (r *Runner) Run() error {
	var eg errgroup.Group

	// Run init workflow if provided, using a single VU.
	init, ok := r.cfg.Workflows[initWorkflow]
	if ok {
		r.logger.Info().Msg("running init workflow")
		time.Sleep(time.Second)

		init.Vus = 1
		if err := r.runWorkflow(initWorkflow, init); err != nil {
			return fmt.Errorf("running init workflow: %w", err)
		}
	}

	for name, workflow := range r.cfg.Workflows {
		eg.Go(func() error {
			return r.runWorkflow(name, workflow)
		})
	}

	return eg.Wait()
}

func (r *Runner) GetEventStream() <-chan Event {
	return r.events
}

func (r *Runner) runWorkflow(name string, workflow Workflow) error {
	var eg errgroup.Group

	for vu := 0; vu < workflow.Vus; vu++ {
		eg.Go(func() error {
			return r.runVU(name, workflow)
		})
	}

	return eg.Wait()
}

func (r *Runner) runVU(workflowName string, workflow Workflow) error {
	// Prepare VU.
	vu := NewVU(r.logger)

	for _, query := range workflow.SetupQueries {
		act, ok := r.cfg.Activities[query]
		if !ok {
			return fmt.Errorf("missing activity: %q", query)
		}

		data, taken, err := r.runQuery(vu, act)
		if err != nil {
			return fmt.Errorf("running query %q: %w", query, err)
		}

		r.events <- Event{Workflow: "*" + workflowName, Name: query, Duration: taken}
		vu.applyData(query, data)
	}

	// Stagger VU.
	vu.stagger(workflow.Queries)

	// Start VU.
	var eg errgroup.Group

	deadline := time.After(r.duration)

	for _, query := range workflow.Queries {
		act, ok := r.cfg.Activities[query.Name]
		if !ok {
			return fmt.Errorf("missing activity: %q", query)
		}

		eg.Go(func() error {
			return r.runActivity(vu, workflowName, query.Name, act, query.Rate, deadline)
		})
	}

	return eg.Wait()
}

func (r *Runner) runActivity(vu *VU, workflowName, queryName string, query Query, rate Rate, fin <-chan time.Time) error {
	ticks := time.NewTicker(rate.tickerInterval).C

	for {
		select {
		case <-ticks:
			depencenciesMet := lo.EveryBy(query.Args, func(a Arg) bool {
				return a.dependencyCheck(vu)
			})
			if !depencenciesMet {
				continue
			}

			r.logger.Debug().Str("query", queryName).Msg("starting")

			data, taken, err := r.runQuery(vu, query)
			if err != nil {
				r.logger.Error().Str("query", queryName).Msgf("error: %v", err)
				continue
			}
			r.logger.Debug().Str("query", queryName).Msgf("[DATA] %+v", data)

			r.events <- Event{Workflow: workflowName, Name: queryName, Duration: taken}
			vu.applyData(queryName, data)

		case <-fin:
			r.logger.Info().Str("query", queryName).Msg("received termination signal")
			return nil
		}
	}
}

func (r *Runner) runQuery(vu *VU, query Query) ([]map[string]any, time.Duration, error) {
	args, err := vu.generateArgs(query.Args)
	if err != nil {
		return nil, 0, fmt.Errorf("generating args: %w", err)
	}

	r.logger.Debug().Msgf("[STMT] %s", query.Query)
	r.logger.Debug().Msgf("\t[ARGS] %v", args)

	switch query.Type {
	case "query":
		return r.db.Query(query.Query, args...)

	case "exec":
		taken, err := r.db.Exec(query.Query, args...)
		return nil, taken, err

	default:
		return nil, 0, fmt.Errorf("unsupported query type: %q", query.Type)
	}
}
