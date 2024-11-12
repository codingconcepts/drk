package model

import "fmt"

// FieldMissingErr is returned when a fields isn't found in
// the config file.
type FieldMissingErr struct {
	Name string
}

func (err FieldMissingErr) Error() string {
	return fmt.Sprintf("%q field is missing:", err.Name)
}
