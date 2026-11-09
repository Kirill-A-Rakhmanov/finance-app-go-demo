package auth_test

import (
	"testing"

	"finance-app-go-demo/internal/application/auth"
	apperror "finance-app-go-demo/internal/application/error"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The password rules themselves are covered in the validation package;
// these tests only check that the auth inputs apply them to the right field.
func TestPasswordRulesAppliedToAuthInputs(t *testing.T) {
	tests := []struct {
		name      string
		validate  func(password string) error
		wantField string
	}{
		{
			name: "register",
			validate: func(password string) error {
				input := auth.RegisterInput{FirstName: "John", LastName: "Doe", Email: "john@example.com", Password: password}
				return input.Validate()
			},
			wantField: "password",
		},
		{
			name: "reset password",
			validate: func(password string) error {
				input := auth.ResetPasswordInput{Code: "123456", NewPassword: password}
				return input.Validate()
			},
			wantField: "newPassword",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, tt.validate("Password123"))

			err := tt.validate("Пароль123")

			var validationErr *apperror.ValidationError
			require.ErrorAs(t, err, &validationErr)
			assert.Equal(t, []string{"must not contain cyrillic characters"}, validationErr.Fields[tt.wantField])
		})
	}
}
