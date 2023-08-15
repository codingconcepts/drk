package random

import (
	"fmt"
	"strings"

	"github.com/codingconcepts/drk/internal/pkg/model"
)

// GenGenerator provides additional context to a gen column.
type GenGenerator struct {
	Value          string `yaml:"value"`
	NullPercentage int    `yaml:"null_percentage"`
	Format         string `yaml:"format"`
}

func (g GenGenerator) GetFormat() string {
	return g.Format
}

// Generate generates random data for a given column.
func (g GenGenerator) Generate() any {
	return g.replacePlaceholders()
}

func (g GenGenerator) replacePlaceholders() string {
	r := Intn(100)
	if r < g.NullPercentage {
		return ""
	}

	s := g.Value

	// Look for quick single-replacements.
	if v, ok := replacements[s]; ok {
		return formatValue(g, v())
	}

	// Process multipe-replacements.
	for k, v := range replacements {
		if strings.Contains(s, k) {
			valueStr := formatValue(g, v())
			s = strings.ReplaceAll(s, k, valueStr)
		}
	}

	return s
}

func formatValue(g GenGenerator, value any) string {
	format := g.GetFormat()
	if format != "" {
		// Check if the value implements the formatter interface and use that first,
		// otherwise, just perform a simple string format.
		if f, ok := value.(model.Formatter); ok {
			return f.Format(format)
		} else {
			return fmt.Sprintf(format, value)
		}
	} else {
		return fmt.Sprintf("%v", value)
	}
}
