package validation

import (
	apperror "finance-app-go-demo/internal/application/error"
	"unicode"
)

const (
	PasswordMinLen = 8
	PasswordMaxLen = 32
)

func Password(v *apperror.ValidationError, field, password string) {
	v.AddIf(len(password) < PasswordMinLen, field, "must be at least 8 characters")
	v.AddIf(len(password) > PasswordMaxLen, field, "must be at most 32 characters")

	hasDigit := false
	hasUpper := false
	hasCyrillic := false

	for _, r := range password {
		switch {
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsUpper(r):
			hasUpper = true
		}

		if unicode.Is(unicode.Cyrillic, r) {
			hasCyrillic = true
		}
	}

	v.AddIf(!hasDigit, field, "must contain at least one digit")
	v.AddIf(!hasUpper, field, "must contain at least one uppercase letter")
	v.AddIf(hasCyrillic, field, "must not contain cyrillic characters")
}
