package model

import (
	"fmt"
	"strings"
)

// FieldMissingErr is returned when a fields isn't found in
// the config file.
type FieldMissingErr struct {
	Name string
}

func (err FieldMissingErr) Error() string {
	return fmt.Sprintf("%q field is missing:", err.Name)
}

// ErrBuilder wraps strings.Builder and simplifies the process of error handling.
type ErrBuilder struct {
	b   strings.Builder
	err error
}

func (b *ErrBuilder) WriteString(format string, args ...any) {
	if b.err != nil {
		return
	}

	_, b.err = b.b.WriteString(fmt.Sprintf(format, args...))
}

func (b *ErrBuilder) String() string {
	return b.b.String()
}

func (b *ErrBuilder) Error() error {
	return b.err
}
