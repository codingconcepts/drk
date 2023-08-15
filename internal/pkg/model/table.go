package model

import (
	"sync"
)

// Table represents what has been written to a table.
type Table struct {
	Name        string
	ColumnNames []string
	RowValues   [][]any
}

// Tables is a thread-safe collection of Table.
type Tables struct {
	Mu     sync.RWMutex
	Tables map[string]Table
}

// Init adds a table to the collection.
func (t *Tables) Init(tableName string, columns []string) {
	t.Mu.Lock()
	defer t.Mu.Unlock()

	t.Tables[tableName] = Table{
		Name:        tableName,
		ColumnNames: columns,
	}
}
