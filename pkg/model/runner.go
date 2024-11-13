package model

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"golang.org/x/sync/errgroup"
)

type Runner struct {
	db       *sql.DB
	cfg      *Drk
	duration time.Duration
	events   chan Event
	logger   *zerolog.Logger
}

func NewRunner(cfg *Drk, url, driver string, duration time.Duration, logger *zerolog.Logger) (*Runner, error) {
	db, err := sql.Open(driver, url)
	if err != nil {
		return nil, fmt.Errorf("connecting to database: %w", err)
	}

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

	start := time.Now()

	switch query.Type {
	case "query":
		rows, err := r.db.Query(query.Query, args...)
		if err != nil {
			return nil, 0, fmt.Errorf("running query: %w", err)
		}

		data, err := readRows(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("reading rows: %w", err)
		}
		return data, time.Since(start), nil

	case "exec":
		_, err = r.db.Exec(query.Query, args...)
		if err != nil {
			return nil, 0, fmt.Errorf("running query: %w", err)
		}
		return nil, time.Since(start), nil

	default:
		return nil, 0, fmt.Errorf("unsupported query type: %q", query.Type)
	}
}

func readRows(rows *sql.Rows) ([]map[string]any, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("getting column names: %w", err)
	}

	values := make([]any, len(columns))
	scanArgs := make([]any, len(columns))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	var results []map[string]any

	for rows.Next() {
		if err := rows.Scan(scanArgs...); err != nil {
			return nil, fmt.Errorf("scaning row: %w", err)
		}

		result := map[string]any{}

		for i, c := range columns {
			cellPtr := scanArgs[i]
			result[c] = *cellPtr.(*any)
		}

		results = append(results, result)
	}

	return results, nil
}
