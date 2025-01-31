package model

import (
	"time"
)

type mockQueryer struct {
	query func(query string, args ...any) ([]map[string]any, time.Duration, error)
	exec  func(query string, args ...any) (time.Duration, error)
	load  func(vu *VU, batch Batch, args [][]any) (time.Duration, error)
}

func (m *mockQueryer) Query(query string, args ...any) ([]map[string]any, time.Duration, error) {
	return m.query(query, args...)
}

func (m *mockQueryer) Exec(query string, args ...any) (time.Duration, error) {
	return m.exec(query, args...)
}

func (m *mockQueryer) Load(vu *VU, batch Batch, args [][]any) (time.Duration, error) {
	return m.load(vu, batch, args)
}
