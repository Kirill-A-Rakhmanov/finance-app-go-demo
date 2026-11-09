package valueObject

import (
	"errors"
	"regexp"
	"strings"
)

var emailRegex = regexp.MustCompile(`^[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}$`)

type Email string

func NormalizeEmail(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func NewEmail(value string) (Email, error) {
	normalized := NormalizeEmail(value)

	if !emailRegex.MatchString(normalized) {
		return "", errors.New("invalid email format")
	}

	return Email(normalized), nil
}

func (email Email) ToString() string {
	return string(email)
}
