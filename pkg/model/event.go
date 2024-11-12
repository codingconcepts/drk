package model

import "time"

// Event is published whenever an operation of type Name is performed.
type Event struct {
	Name     string
	Duration time.Duration
}
