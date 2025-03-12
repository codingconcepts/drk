package model

import (
	"fmt"
	"os"
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
	db          repo.Queryer
	cfg         *Drk
	envMappings envMappingGenerator
	duration    time.Duration
	events      chan Event
	logger      *zerolog.Logger
}

func NewRunner(cfg *Drk, db repo.Queryer, url, driver string, duration time.Duration, logger *zerolog.Logger) (*Runner, error) {
	r := Runner{
		db:          db,
		cfg:         cfg,
		envMappings: createEnvMappingGenerator(cfg),
		duration:    duration,
		events:      make(chan Event, 1000),
		logger:      logger,
	}

	logger.Info().Float64("duration", r.duration.Seconds()).Msgf("runner")

	return &r, nil
}

func createEnvMappingGenerator(cfg *Drk) func(env, value string) (string, bool) {
	return func(env, value string) (string, bool) {
		mapping, ok := cfg.EnvMappings[env]
		if !ok {
			return "", false
		}

		envVarValue, ok := os.LookupEnv(env)
		if !ok {
			return "", false
		}

		to, ok := mapping[envVarValue]
		return to, ok
	}
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
	r.logger.Info().Msg("finished init workflow")

	for name, workflow := range r.cfg.Workflows {
		eg.Go(func() error {
			return lo.Ternary(
				workflow.RampFor > 0,
				r.rampWorkflow(name, workflow),
				r.runWorkflow(name, workflow),
			)
		})
	}

	return eg.Wait()
}

func (r *Runner) GetEventStream() <-chan Event {
	return r.events
}

func (r *Runner) rampWorkflow(name string, workflow Workflow) error {
	var eg errgroup.Group

	stagger := workflow.RampFor / time.Duration(workflow.Vus)
	for range workflow.Vus {
		time.Sleep(stagger)

		eg.Go(func() error {
			return r.runVU(name, workflow)
		})
	}

	return eg.Wait()
}

func (r *Runner) runWorkflow(name string, workflow Workflow) error {
	var eg errgroup.Group

	for range workflow.Vus {
		eg.Go(func() error {
			return r.runVU(name, workflow)
		})
	}

	return eg.Wait()
}

func (r *Runner) runVU(workflowName string, workflow Workflow) error {
	// Delay start if required.
	if workflow.RunAfter > 0 {
		r.logger.Debug().Str("workflow", workflowName).Dur("for", workflow.RunAfter).Msgf("delaying")
		time.Sleep(workflow.RunAfter)
	}

	// Prepare VU.
	vu := NewVU(r)

	r.logger.Debug().Str("workflow", workflowName).Msgf("running setup queries")

	for _, query := range workflow.SetupQueries {
		act, ok := r.cfg.Activities[query]
		if !ok {
			return fmt.Errorf("missing activity: %q", query)
		}

		data, taken, err := r.runQuery(vu, act)
		if err != nil {
			r.logger.Warn().Str("query", query).Any("error", err.Error()).Msg("running query")
			r.events <- Event{Workflow: workflowName, Name: query, Duration: taken, Err: err}

			return fmt.Errorf("running query %q: %w", query, err)
		}

		r.events <- Event{Workflow: "*" + workflowName, Name: query, Duration: taken}
		vu.applyData(query, data)
	}

	r.logger.Debug().Str("workflow", workflowName).Msgf("finished setup queries")

	// Stagger VU.
	vu.stagger(workflow.Queries)

	// Start VU.
	var eg errgroup.Group

	// Finish early if required, otherwise, run until end of test.
	deadline := time.After(lo.CoalesceOrEmpty(workflow.RunFor, r.duration))

	r.logger.Debug().Str("workflow", workflowName).Msgf("preparing workflow queries")

	var activities int
	for _, query := range workflow.Queries {
		act, ok := r.cfg.Activities[query.Name]
		if !ok {
			r.logger.Warn().Str("name", query.Name).Msgf("missing activity")
			return fmt.Errorf("missing activity: %q", query)
		}

		activities++
		eg.Go(func() error {
			return r.runActivity(vu, workflowName, query.Name, act, query.Rate, deadline)
		})
	}

	r.logger.Debug().
		Str("workflow", workflowName).
		Int("activites", activities).
		Msgf("running workflow queries")

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
				r.logger.Debug().Str("workflow", workflowName).Str("query", queryName).Msg("dependencies not met")
				continue
			}

			data, taken, err := r.runQuery(vu, query)
			if err != nil {
				r.logger.Warn().Str("workflow", workflowName).Str("query", queryName).Err(err).Msg("")
				r.events <- Event{Workflow: workflowName, Name: queryName, Duration: taken, Err: err}
				continue
			}

			r.events <- Event{Workflow: workflowName, Name: queryName, Duration: taken}
			vu.applyData(queryName, data)

		case <-fin:
			r.logger.Debug().Str("query", queryName).Msg("received termination signal")
			return nil
		}
	}
}

func (r *Runner) runQuery(vu *VU, query Query) ([]map[string]any, time.Duration, error) {
	args, err := vu.generateArgs(query.Args)
	if err != nil {
		return nil, 0, fmt.Errorf("generating args: %w", err)
	}

	r.logger.Debug().Str("type", query.Type).Msgf("[STMT] %s", query.Query)
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
