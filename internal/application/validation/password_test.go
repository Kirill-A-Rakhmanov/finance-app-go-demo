package validation_test

import (
	"strings"
	"testing"

	apperror "finance-app-go-demo/internal/application/error"
	"finance-app-go-demo/internal/application/validation"

	"github.com/stretchr/testify/assert"
)

func TestPassword(t *testing.T) {
	const (
		errTooShort = "must be at least 8 characters"
		errTooLong  = "must be at most 32 characters"
		errNoDigit  = "must contain at least one digit"
		errNoUpper  = "must contain at least one uppercase letter"
		errCyrillic = "must not contain cyrillic characters"
	)

	tests := []struct {
		name     string
		password string
		wantErrs []string
	}{
		{name: "valid password", password: "Password123"},
		{name: "valid with special characters", password: "Pass-word_123!"},
		{name: "exactly min length", password: "Passwor1"},
		{name: "exactly max length", password: "Aa1" + strings.Repeat("a", 29)},
		{name: "one below min length", password: "Passwo1", wantErrs: []string{errTooShort}},
		{name: "one above max length", password: "Aa1" + strings.Repeat("a", 30), wantErrs: []string{errTooLong}},
		{name: "missing digit", password: "Passwordddd", wantErrs: []string{errNoDigit}},
		{name: "missing uppercase", password: "password123", wantErrs: []string{errNoUpper}},
		{name: "missing digit and uppercase", password: "lowercaseonly", wantErrs: []string{errNoDigit, errNoUpper}},
		{name: "empty password", password: "", wantErrs: []string{errTooShort, errNoDigit, errNoUpper}},
		{name: "too long and missing digit", password: "P" + strings.Repeat("a", 40), wantErrs: []string{errTooLong, errNoDigit}},
		{name: "only cyrillic letters", password: "Пароль123", wantErrs: []string{errCyrillic}},
		{name: "one cyrillic letter among latin", password: "Passwordд1", wantErrs: []string{errCyrillic}},
		{name: "only lowercase cyrillic letters", password: "пароль123", wantErrs: []string{errNoUpper, errCyrillic}},
		{name: "lowercase cyrillic and latin uppercase", password: "Password1я", wantErrs: []string{errCyrillic}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := apperror.NewValidationError()

			validation.Password(v, "password", tt.password)

			if len(tt.wantErrs) == 0 {
				assert.False(t, v.HasErrors(), "unexpected errors: %v", v.Fields)
				return
			}
			assert.ElementsMatch(t, tt.wantErrs, v.Fields["password"])
			assert.Len(t, v.Fields, 1)
		})
	}
}
