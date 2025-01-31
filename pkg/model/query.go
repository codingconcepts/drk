package model

import (
	"fmt"
	"strings"

	"github.com/samber/lo"
)

func argGenerator(driver string) func() string {
	index := 0

	switch strings.ToLower(driver) {
	case "mysql":
		return func() string {
			return "?"
		}

	case "postgres", "pgx":
		return func() string {
			index++
			return fmt.Sprintf("$%d", index)
		}

	case "oracle":
		return func() string {
			index++
			return fmt.Sprintf(":a_%d", index)
		}

	default:
		panic(fmt.Sprintf("unsupported database driver: %q", driver))
	}
}

func insertStatement(argGenerator func() string, batch Batch, rows [][]any) (string, error) {
	var b ErrBuilder

	b.WriteString(
		"INSERT INTO %s (%s) VALUES ",
		batch.Table,
		strings.Join(batch.Columns, ","),
	)

	for i, row := range rows {
		columnValues := lo.Times(len(row), func(_ int) string {
			return argGenerator()
		})

		b.WriteString("(%s)", strings.Join(columnValues, ","))

		if i < len(rows)-1 {
			b.WriteString(",")
		}
	}

	if err := b.Error(); err != nil {
		return "", err
	}

	return b.String(), nil
}
