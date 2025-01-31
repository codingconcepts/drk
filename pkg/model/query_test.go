package model

import (
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestNextArgPlaceholder(t *testing.T) {
	cases := []struct {
		driver string
		count  int
		exp    []string
	}{
		{
			driver: "mysql",
			count:  5,
			exp:    []string{"?", "?", "?", "?", "?"},
		},
		{
			driver: "postgres",
			count:  5,
			exp:    []string{"$1", "$2", "$3", "$4", "$5"},
		},
		{
			driver: "oracle",
			count:  5,
			exp:    []string{":a_1", ":a_2", ":a_3", ":a_4", ":a_5"},
		},
	}

	for _, c := range cases {
		t.Run(c.driver, func(t *testing.T) {
			sut := argGenerator(c.driver)

			placeholders := lo.Times(c.count, func(_ int) string {
				return sut()
			})

			assert.Equal(t, c.exp, placeholders)
		})
	}
}

func TestInsertStatement(t *testing.T) {
	cases := []struct {
		driver string
		exp    string
	}{
		{
			driver: "mysql",
			exp:    "INSERT INTO t (a,b,c) VALUES (?,?,?),(?,?,?),(?,?,?)",
		},
		{
			driver: "pgx",
			exp:    "INSERT INTO t (a,b,c) VALUES ($1,$2,$3),($4,$5,$6),($7,$8,$9)",
		},
		{
			driver: "postgres",
			exp:    "INSERT INTO t (a,b,c) VALUES ($1,$2,$3),($4,$5,$6),($7,$8,$9)",
		},
		{
			driver: "oracle",
			exp:    "INSERT INTO t (a,b,c) VALUES (:a_1,:a_2,:a_3),(:a_4,:a_5,:a_6),(:a_7,:a_8,:a_9)",
		},
	}

	batch := Batch{
		Table:   "t",
		Columns: []string{"a", "b", "c"},
	}

	rows := [][]any{
		{1, 2, 3},
		{4, 5, 6},
		{7, 8, 9},
	}

	for _, c := range cases {
		t.Run(c.driver, func(t *testing.T) {
			argGenerator := argGenerator(c.driver)
			act, err := insertStatement(argGenerator, batch, rows)
			assert.NoError(t, err)
			assert.Equal(t, c.exp, act)
		})
	}
}

func TestExtractColumnValues(t *testing.T) {
	cases := []struct {
		name      string
		columns   []string
		returning []string
		exp       []map[string]any
	}{
		{
			name:      "all columns",
			columns:   []string{"id", "name", "price"},
			returning: []string{"id", "name", "price"},
			exp: []map[string]any{
				{"id": "a403ae29-4d41-468a-a359-ac0e898772de", "name": "a", "price": 1.99},
				{"id": "bf7a83bc-69ea-41fa-a218-a0522b08c245", "name": "b", "price": 2.99},
				{"id": "c9b86e0b-94ea-467a-b803-459e179fa82b", "name": "c", "price": 3.99},
			},
		},
		{
			name:      "first column",
			columns:   []string{"id", "name", "price"},
			returning: []string{"id"},
			exp: []map[string]any{
				{"id": "a403ae29-4d41-468a-a359-ac0e898772de"},
				{"id": "bf7a83bc-69ea-41fa-a218-a0522b08c245"},
				{"id": "c9b86e0b-94ea-467a-b803-459e179fa82b"},
			},
		},
		{
			name:      "middle column",
			columns:   []string{"id", "name", "price"},
			returning: []string{"name"},
			exp: []map[string]any{
				{"name": "a"},
				{"name": "b"},
				{"name": "c"},
			},
		},
		{
			name:      "last column",
			columns:   []string{"id", "name", "price"},
			returning: []string{"price"},
			exp: []map[string]any{
				{"price": 1.99},
				{"price": 2.99},
				{"price": 3.99},
			},
		},
	}

	data := [][]any{
		{"a403ae29-4d41-468a-a359-ac0e898772de", "a", 1.99},
		{"bf7a83bc-69ea-41fa-a218-a0522b08c245", "b", 2.99},
		{"c9b86e0b-94ea-467a-b803-459e179fa82b", "c", 3.99},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			act := extractColumnValues(c.columns, c.returning, data)
			assert.Equal(t, c.exp, act)
		})
	}
}
