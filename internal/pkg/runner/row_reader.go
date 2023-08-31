package runner

import (
	"fmt"

	"github.com/jackc/pgx/v5"
)

func read(dbRows pgx.Rows) ([]map[string]any, error) {
	fields := dbRows.FieldDescriptions()

	rows := []map[string]any{}
	for dbRows.Next() {
		scans := make([]any, len(fields))
		for i := range scans {
			scans[i] = &scans[i]
		}

		if err := dbRows.Scan(scans...); err != nil {
			return nil, fmt.Errorf("scaning values: %w", err)
		}

		row := make(map[string]any)
		for i, v := range scans {
			if v != nil {
				row[fields[i].Name] = v
			}
		}
		rows = append(rows, row)
	}

	return rows, nil
}
