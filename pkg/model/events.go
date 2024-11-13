package model

import "time"

// Event is published whenever an operation of type Name is performed.
type Event struct {
	Workflow string
	Name     string
	Duration time.Duration
}
