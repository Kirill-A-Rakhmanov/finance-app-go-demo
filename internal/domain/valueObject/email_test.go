package valueObject_test

import (
	"testing"

	"finance-app-go-demo/internal/domain/valueObject"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEmail(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"valid email", "john@example.com", "john@example.com", false},
		{"valid email with subdomain", "john@mail.example.co", "john@mail.example.co", false},
		{"valid email with plus tag", "john+finance@example.com", "john+finance@example.com", false},
		{"upper case is lowered", "John.Doe@Example.COM", "john.doe@example.com", false},
		{"surrounding spaces are trimmed", "  john@example.com\t", "john@example.com", false},
		{"missing at sign", "johnexample.com", "", true},
		{"missing domain", "john@", "", true},
		{"missing tld", "john@example", "", true},
		{"empty string", "", "", true},
		{"only spaces", "   ", "", true},
		{"spaces inside not allowed", "john doe@example.com", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			email, err := valueObject.NewEmail(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, email)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, email.ToString())
		})
	}
}

func TestNormalizeEmail(t *testing.T) {
	assert.Equal(t, "john@example.com", valueObject.NormalizeEmail(" John@Example.com "))
	assert.Equal(t, "", valueObject.NormalizeEmail("   "))
}
