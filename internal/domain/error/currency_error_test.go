package error_test

import (
	"errors"
	"testing"

	domainerror "finance-app-go-demo/internal/domain/error"

	"github.com/stretchr/testify/assert"
)

func TestCurrencyError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  *domainerror.CurrencyError
		want string
	}{
		{
			name: "without comment",
			err:  &domainerror.CurrencyError{Message: "unknown currency code"},
			want: "unknown currency code",
		},
		{
			name: "with comment",
			err:  &domainerror.CurrencyError{Message: "unknown currency code", Comment: "XXX"},
			want: "unknown currency code: XXX",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.err.Error())
		})
	}
}

func TestCurrencyError_New(t *testing.T) {
	err := domainerror.ErrUnknownCurrencyCode.New("XXX")

	assert.Equal(t, "unknown currency code", err.Message)
	assert.Equal(t, "XXX", err.Comment)
	assert.Equal(t, "unknown currency code: XXX", err.Error())
}

func TestCurrencyError_Is(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		target error
		want   bool
	}{
		{
			name:   "same message with different comment matches",
			err:    domainerror.ErrUnknownCurrencyCode.New("XXX"),
			target: domainerror.ErrUnknownCurrencyCode,
			want:   true,
		},
		{
			name:   "different error type does not match",
			err:    domainerror.ErrUnknownCurrencyCode,
			target: errors.New("unknown currency code"),
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, errors.Is(tt.err, tt.target))
		})
	}
}
