package model

import (
	"errors"
	"sort"
	"strings"
)

var (
	// ErrNotFound is returned when an employee does not exist.
	ErrNotFound = errors.New("employee not found")
	// ErrDuplicateEmail is returned when the email is already in use.
	ErrDuplicateEmail = errors.New("an employee with this email already exists")
)

// ValidationError carries per-field validation messages.
type ValidationError struct {
	Fields map[string]string `json:"fields"`
}

func (e *ValidationError) Error() string {
	keys := make([]string, 0, len(e.Fields))
	for k := range e.Fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+": "+e.Fields[k])
	}
	return "validation failed: " + strings.Join(parts, "; ")
}
