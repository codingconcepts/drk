package model

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
	Load(vu *VU, batch Batch, args [][]any) (time.Duration, error)
}

type DBRepo struct {
	db      *sql.DB
	driver  string
	timeout time.Duration
}

func NewDBRepo(db *sql.DB, driver string, timeout time.Duration) *DBRepo {
	return &DBRepo{
		db:      db,
		driver:  driver,
		timeout: timeout,
	}
}

func (r *DBRepo) Query(query string, args ...any) (values []map[string]any, taken time.Duration, err error) {
	start := time.Now()

	defer func() {
		taken = time.Since(start)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		err = fmt.Errorf("running query: %w", err)
	}

	values, err = readRows(rows)
	if err != nil {
		err = fmt.Errorf("reading rows: %w", err)
	}

	return
}

func (r *DBRepo) Exec(query string, args ...any) (taken time.Duration, err error) {
	start := time.Now()

	defer func() {
		taken = time.Since(start)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	_, err = r.db.ExecContext(ctx, query, args...)
	if err != nil {
		err = fmt.Errorf("running query: %w", err)
	}

	return
}

func (r *DBRepo) Load(vu *VU, batch Batch, rows [][]any) (time.Duration, error) {
	argGenerator := argGenerator(r.driver)
	stmt, err := insertStatement(argGenerator, batch, rows)
	if err != nil {
		return 0, fmt.Errorf("generating insert statement: %w", err)
	}

	return r.Exec(stmt, lo.Flatten(rows)...)
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
