package error

import "fmt"

var ErrUnknownCurrencyCode = &CurrencyError{Message: "unknown currency code"}

type CurrencyError struct {
	Message string
	Comment string
}

func (e *CurrencyError) Error() string {
	if e.Comment == "" {
		return e.Message
	}
	return fmt.Sprintf("%s: %s", e.Message, e.Comment)
}

func (e *CurrencyError) Is(target error) bool {
	t, ok := target.(*CurrencyError)
	if !ok {
		return false
	}
	return e.Message == t.Message
}

func (e *CurrencyError) New(comment string) *CurrencyError {
	return &CurrencyError{
		Message: e.Message,
		Comment: comment,
	}
}
