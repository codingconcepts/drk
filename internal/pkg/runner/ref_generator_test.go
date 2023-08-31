package runner

import (
	"testing"
)

func TestGenerateRefColumn(t *testing.T) {
	// table := model.Table{
	// 	Name:  "pet",
	// 	Count: 2,
	// }

	// column := model.Column{
	// 	Name: "person_id",
	// }

	// g := RefGenerator{
	// 	Table:  "person",
	// 	Column: "id",
	// }

	// files := map[string]model.CSVFile{
	// 	"person": {
	// 		Header: []string{"id"},
	// 		Lines:  [][]string{{"ce9af887-37eb-4e08-9790-4f481b0fa594"}},
	// 	},
	// }
	// err := g.Generate(table, column, files)
	// assert.Nil(t, err)
	// assert.Equal(t, "ce9af887-37eb-4e08-9790-4f481b0fa594", files["pet"].Lines[0][0])
	// assert.Equal(t, "ce9af887-37eb-4e08-9790-4f481b0fa594", files["pet"].Lines[0][1])
}
