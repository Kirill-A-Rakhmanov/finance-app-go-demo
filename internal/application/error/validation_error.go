package error

import (
	"fmt"
	"unicode/utf8"
)

type ValidationError struct {
	Message string
	Fields  map[string][]string
}

func NewValidationError() *ValidationError {
	return &ValidationError{
		Message: "validation failed",
		Fields:  make(map[string][]string),
	}
}

func (e *ValidationError) Add(field, message string) {
	e.Fields[field] = append(e.Fields[field], message)
}

func (e *ValidationError) AddIf(condition bool, field, message string) {
	if condition {
		e.Add(field, message)
	}
}

func (e *ValidationError) MaxLength(field, value string, max int) {
	e.AddIf(utf8.RuneCountInString(value) > max, field, fmt.Sprintf("must be at most %d characters", max))
}

func (e *ValidationError) MaxLengthPtr(field string, value *string, max int) {
	if value != nil {
		e.MaxLength(field, *value, max)
	}
}

func (e *ValidationError) HasErrors() bool {
	return len(e.Fields) > 0
}

func (e *ValidationError) ErrorOrNil() error {
	if !e.HasErrors() {
		return nil
	}
	return e
}

func (e *ValidationError) Error() string {
	return e.Message
}
