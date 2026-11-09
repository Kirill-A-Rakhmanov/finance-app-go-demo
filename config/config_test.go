package config_test

import (
	"strings"
	"testing"

	"finance-app-go-demo/config"

	"github.com/stretchr/testify/assert"
)

func TestApp_Validate_JWTSecretLength(t *testing.T) {
	tests := []struct {
		name    string
		secret  string
		wantErr bool
	}{
		{"empty", "", true},
		{"example placeholder", "secret", true},
		{"one short of the minimum", strings.Repeat("a", config.JWTSecretMinLength-1), true},
		{"exactly the minimum", strings.Repeat("a", config.JWTSecretMinLength), false},
		{"long random secret", strings.Repeat("0123456789abcdef", 8), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := config.App{JWTSecret: tt.secret}.Validate()

			if tt.wantErr {
				assert.ErrorContains(t, err, "AUTH_JWT_SECRET")
				return
			}

			assert.NoError(t, err)
		})
	}
}
