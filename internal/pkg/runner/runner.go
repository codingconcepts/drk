package runner

import (
	"context"
	"fmt"

	"github.com/codingconcepts/drk/internal/pkg/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Runner holds any data that has been extracted from the database for later use.
type Runner struct {
	db   *pgxpool.Pool
	data map[string]table
}

type table map[string][]any

// New returns a pointer to a new instance of Runner.
func New(db *pgxpool.Pool) *Runner {
	return &Runner{
		db:   db,
		data: map[string]table{},
	}
}

func (r *Runner) Run(q model.Query) error {
	args, err := r.generateArguments(q.Arguments)
	if err != nil {
		return fmt.Errorf("generating arguments: %w", err)
	}

	return r.Query(q.Table, q.Statement, args)
}

func (r *Runner) Query(table, stmt string, args []any) error {
	if _, ok := r.data[table]; !ok {
		r.data[table] = map[string][]any{}
	}

	dbRows, err := r.db.Query(context.Background(), stmt, args...)
	if err != nil {
		return fmt.Errorf("making query: %w", err)
	}

	rows, err := scan(dbRows)
	if err != nil {
		return fmt.Errorf("scanning rows: %v", err)
	}

	// Add to the runner's in-memory cache of data.
	r.addRows(table, rows)
	return nil
}

func (r *Runner) addRows(table string, rows []map[string]any) {
	for _, row := range rows {
		for col, val := range row {
			r.data[table][col] = append(r.data[table][col], val)
		}
	}
}

func (r *Runner) generateArguments(args []model.Argument) ([]any, error) {
	out := make([]any, len(args))

	for i, arg := range args {
		switch arg.Type {
		case "gen":
			var g GenGenerator
			if err := arg.Processor.UnmarshalFunc(&g); err != nil {
				return nil, fmt.Errorf("parsing gen processor for %s: %w", arg.Name, err)
			}
			out[i] = g.Generate()
		case "set":
			var g SetGenerator
			if err := arg.Processor.UnmarshalFunc(&g); err != nil {
				return nil, fmt.Errorf("parsing gen processor for %s: %w", arg.Name, err)
			}
			out[i] = g.Generate()
		case "ref":
			var g RefGenerator
			if err := arg.Processor.UnmarshalFunc(&g); err != nil {
				return nil, fmt.Errorf("parsing ref processor for %s: %w", arg.Name, err)
			}

			val, err := g.Generate(r.data)
			if err != nil {
				return nil, fmt.Errorf("running ref processor for %s: %w", arg.Name, err)
			}
			out[i] = val
		}
	}

	return out, nil
}
