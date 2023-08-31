package runner

import (
	"fmt"

	"github.com/samber/lo"
)

// RefGenerator provides additional context to a ref column.
type RefGenerator struct {
	Table  string `yaml:"table"`
	Column string `yaml:"column"`
}

// Generate looks to previously generated table data and references that when generating data
// for the given table.
func (g RefGenerator) Generate(data map[string]table) (any, error) {
	table, ok := data[g.Table]
	if !ok {
		return nil, fmt.Errorf("missing table %q for ref lookup", g.Table)
	}

	column, ok := table[g.Column]
	if !ok {
		return nil, fmt.Errorf("missing column %q for ref lookup", g.Column)
	}

	return lo.Sample(column), nil
}
