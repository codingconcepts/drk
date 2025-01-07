package repo

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/samber/lo"
)

type Queryer interface {
	Query(query string, args ...any) ([]map[string]any, time.Duration, error)
	Exec(query string, args ...any) (time.Duration, error)
}

type DBRepo struct {
	db      *sql.DB
	timeout time.Duration
}

func NewDBRepo(db *sql.DB, timeout time.Duration) *DBRepo {
	return &DBRepo{
		db:      db,
		timeout: timeout,
	}
}

func (r *DBRepo) Query(query string, args ...any) ([]map[string]any, time.Duration, error) {
	start := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("running query: %w", err)
	}

	data, err := readRows(rows)
	if err != nil {
		return nil, 0, fmt.Errorf("reading rows: %w", err)
	}
	return data, time.Since(start), nil
}

func (r *DBRepo) Exec(query string, args ...any) (time.Duration, error) {
	start := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("running query: %w", err)
	}

	return time.Since(start), nil
}

func readRows(rows *sql.Rows) ([]map[string]any, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("getting column names: %w", err)
	}

	// Convert to lower case to handle all databases.
	columns = lo.Map(columns, func(s string, _ int) string {
		return strings.ToLower(s)
	})

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
