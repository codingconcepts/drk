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
	events   chan string
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
		events:   make(chan string, 1000),
		logger:   logger,
	}

	logger.Info().Float64("duration", r.duration.Seconds()).Msgf("runner")

	return &r, nil
}

func (r *Runner) Run() error {
	var eg errgroup.Group

	for _, workflow := range r.cfg.Workflows {
		eg.Go(func() error {
			return r.runWorkflow(workflow)
		})
	}

	return eg.Wait()
}

func (r *Runner) GetEventStream() <-chan string {
	return r.events
}

func (r *Runner) runWorkflow(workflow Workflow) error {
	var eg errgroup.Group

	for vu := 0; vu < workflow.Vus; vu++ {
		eg.Go(func() error {
			return r.runVU(workflow)
		})
	}

	return eg.Wait()
}

func (r *Runner) runVU(workflow Workflow) error {
	// Prepare VU.
	vu := NewVU(r.logger)

	for _, query := range workflow.SetupQueries {
		act, ok := r.cfg.Activities[query]
		if !ok {
			return fmt.Errorf("missing activity: %q", query)
		}

		data, err := r.runQuery(vu, act)
		if err != nil {
			return fmt.Errorf("running query %q: %w", query, err)
		}

		r.events <- query
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
			return r.runActivity(vu, query.Name, act, query.Rate, deadline)
		})
	}

	return eg.Wait()
}

func (r *Runner) runActivity(vu *VU, name string, query Query, rate Rate, fin <-chan time.Time) error {
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

			r.logger.Debug().Str("query", name).Msg("starting")

			data, err := r.runQuery(vu, query)
			if err != nil {
				r.logger.Error().Str("query", name).Msgf("error: %v", err)
				continue
			}
			r.logger.Debug().Str("query", name).Msgf("[DATA] %+v", data)

			r.events <- name
			vu.applyData(name, data)

		case <-fin:
			r.logger.Info().Str("query", name).Msg("received termination signal")
			return nil
		}
	}
}

func (r *Runner) runQuery(vu *VU, query Query) ([]map[string]any, error) {
	args, err := vu.generateArgs(query.Args)
	if err != nil {
		return nil, fmt.Errorf("generating args: %w", err)
	}

	r.logger.Debug().Msgf("[STMT] %s", query.Query)
	r.logger.Debug().Msgf("\t[ARGS] %v", args)

	switch query.Type {
	case "query":
		rows, err := r.db.Query(query.Query, args...)
		if err != nil {
			return nil, fmt.Errorf("running query: %w", err)
		}

		data, err := readRows(rows)
		if err != nil {
			return nil, fmt.Errorf("reading rows: %w", err)
		}
		return data, nil

	case "exec":
		_, err = r.db.Exec(query.Query, args...)
		if err != nil {
			return nil, fmt.Errorf("running query: %w", err)
		}
		return nil, nil

	default:
		return nil, fmt.Errorf("unsupported query type: %q", query.Type)
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
