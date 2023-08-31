package runner

import (
	"fmt"

	"github.com/jackc/pgx/v5"
)

func scan(rows pgx.Rows) ([]map[string]any, error) {
	fields := rows.FieldDescriptions()

	var values []map[string]any
	for rows.Next() {
		scans := make([]any, len(fields))
		row := make(map[string]any)

		for i := range scans {
			scans[i] = &scans[i]
		}

		if err := rows.Scan(scans...); err != nil {
			return nil, fmt.Errorf("scaning values: %w", err)
		}

		for i, v := range scans {
			if v != nil {
				if fields[i].DataTypeOID == 2950 {
					b := v.([16]byte)
					v = fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
				}
				row[fields[i].Name] = v
			}
		}
		values = append(values, row)
	}

	return values, nil
}
