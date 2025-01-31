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
