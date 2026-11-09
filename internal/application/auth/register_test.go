package auth_test

import (
	"context"
	"errors"
	"testing"

	"finance-app-go-demo/internal/application/auth"
	authmock "finance-app-go-demo/internal/application/auth/mock"
	apperror "finance-app-go-demo/internal/application/error"
	"finance-app-go-demo/internal/domain/model"
	repomock "finance-app-go-demo/internal/domain/repository/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type registerMocks struct {
	userRepo             *repomock.MockUserRepository
	featureFlagRepo      *repomock.MockFeatureFlagRepository
	passwordService      *authmock.MockPasswordService
	registrationStorage  *authmock.MockRegistrationStorage
	registrationNotifier *authmock.MockRegistrationNotifier
}

func TestService_Register(t *testing.T) {
	validInput := auth.RegisterInput{
		FirstName: "Jane",
		LastName:  "Doe",
		Email:     "jane@example.com",
		Password:  "Password123",
	}

	registrationOpen := func(m *registerMocks) {
		m.featureFlagRepo.EXPECT().IsEnabled(mock.Anything, model.FeatureFlagRegistration).Return(true, nil)
	}
	hashed := func(m *registerMocks) {
		m.passwordService.EXPECT().Hash("Password123").Return("hashed-password", nil)
	}
	cooldownFree := func(m *registerMocks) {
		m.registrationStorage.EXPECT().AcquireCooldown(mock.Anything, "jane@example.com", auth.RegistrationCooldown).Return(true, nil)
	}

	tests := []struct {
		name       string
		input      auth.RegisterInput
		setupMocks func(m *registerMocks)
		wantErr    func(t *testing.T, err error)
	}{
		{
			name:  "free email: pending registration is stored and the confirmation is sent",
			input: validInput,
			setupMocks: func(m *registerMocks) {
				registrationOpen(m)
				hashed(m)
				cooldownFree(m)
				m.userRepo.EXPECT().ExistsByEmail(mock.Anything, "jane@example.com").Return(false, nil)
				m.registrationStorage.EXPECT().Create(mock.Anything, auth.PendingRegistration{
					FirstName: "Jane", LastName: "Doe", Email: "jane@example.com", PasswordHash: "hashed-password",
				}).Return("confirmation-token", nil)
				m.registrationNotifier.EXPECT().SendRegistrationConfirmation(mock.Anything, auth.RegistrationConfirmation{
					Email: "jane@example.com", FirstName: "Jane", LastName: "Doe", Token: "confirmation-token",
				}).Return(nil)
			},
		},
		{
			name: "email is normalized: case and spaces do not create a second registration",
			input: auth.RegisterInput{
				FirstName: "Jane", LastName: "Doe", Email: "  Jane@Example.COM ", Password: "Password123",
			},
			setupMocks: func(m *registerMocks) {
				registrationOpen(m)
				hashed(m)
				cooldownFree(m)
				m.userRepo.EXPECT().ExistsByEmail(mock.Anything, "jane@example.com").Return(false, nil)
				m.registrationStorage.EXPECT().Create(mock.Anything, auth.PendingRegistration{
					FirstName: "Jane", LastName: "Doe", Email: "jane@example.com", PasswordHash: "hashed-password",
				}).Return("confirmation-token", nil)
				m.registrationNotifier.EXPECT().SendRegistrationConfirmation(mock.Anything, mock.MatchedBy(func(c auth.RegistrationConfirmation) bool {
					return c.Email == "jane@example.com"
				})).Return(nil)
			},
		},
		{
			name:  "taken email: same answer, the owner is told instead and nothing is stored",
			input: validInput,
			setupMocks: func(m *registerMocks) {
				registrationOpen(m)
				hashed(m)
				cooldownFree(m)
				m.userRepo.EXPECT().ExistsByEmail(mock.Anything, "jane@example.com").Return(true, nil)
				m.registrationNotifier.EXPECT().SendRegistrationAttempt(mock.Anything, "jane@example.com").Return(nil)
			},
		},
		{
			name:  "within the cooldown: same answer, nothing is checked, stored or sent",
			input: validInput,
			setupMocks: func(m *registerMocks) {
				registrationOpen(m)
				hashed(m)
				m.registrationStorage.EXPECT().AcquireCooldown(mock.Anything, "jane@example.com", auth.RegistrationCooldown).Return(false, nil)
			},
		},
		{
			name:  "registration closed",
			input: validInput,
			setupMocks: func(m *registerMocks) {
				m.featureFlagRepo.EXPECT().IsEnabled(mock.Anything, model.FeatureFlagRegistration).Return(false, nil)
			},
			wantErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, auth.ErrRegistrationClosed)
			},
		},
		{
			name:       "validation error - empty fields",
			input:      auth.RegisterInput{Password: "short"},
			setupMocks: func(m *registerMocks) {},
			wantErr: func(t *testing.T, err error) {
				var validationErr *apperror.ValidationError
				require.True(t, errors.As(err, &validationErr))
				assert.Contains(t, validationErr.Fields, "firstName")
				assert.Contains(t, validationErr.Fields, "lastName")
				assert.Contains(t, validationErr.Fields, "email")
				assert.Contains(t, validationErr.Fields, "password")
			},
		},
		{
			name: "validation error - password missing digit and uppercase",
			input: auth.RegisterInput{
				FirstName: "Jane", LastName: "Doe", Email: "jane@example.com", Password: "lowercaseonly",
			},
			setupMocks: func(m *registerMocks) {},
			wantErr: func(t *testing.T, err error) {
				var validationErr *apperror.ValidationError
				require.True(t, errors.As(err, &validationErr))
				require.Contains(t, validationErr.Fields, "password")
				assert.Contains(t, validationErr.Fields["password"], "must contain at least one digit")
				assert.Contains(t, validationErr.Fields["password"], "must contain at least one uppercase letter")
			},
		},
		{
			name: "invalid email format",
			input: auth.RegisterInput{
				FirstName: "Jane", LastName: "Doe", Email: "not-an-email", Password: "Password123",
			},
			setupMocks: func(m *registerMocks) {
				registrationOpen(m)
			},
			wantErr: func(t *testing.T, err error) {
				var appErr *apperror.Error
				require.True(t, errors.As(err, &appErr))
				assert.Equal(t, apperror.KindInvalid, appErr.Kind)
			},
		},
		{
			name:  "feature flag lookup fails",
			input: validInput,
			setupMocks: func(m *registerMocks) {
				m.featureFlagRepo.EXPECT().IsEnabled(mock.Anything, model.FeatureFlagRegistration).Return(false, errors.New("db down"))
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "db down")
			},
		},
		{
			name:  "password hashing fails",
			input: validInput,
			setupMocks: func(m *registerMocks) {
				registrationOpen(m)
				m.passwordService.EXPECT().Hash("Password123").Return("", errors.New("bcrypt failure"))
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "bcrypt failure")
			},
		},
		{
			name:  "cooldown lookup fails",
			input: validInput,
			setupMocks: func(m *registerMocks) {
				registrationOpen(m)
				hashed(m)
				m.registrationStorage.EXPECT().AcquireCooldown(mock.Anything, "jane@example.com", auth.RegistrationCooldown).Return(false, errors.New("redis down"))
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "redis down")
			},
		},
		{
			name:  "user lookup fails",
			input: validInput,
			setupMocks: func(m *registerMocks) {
				registrationOpen(m)
				hashed(m)
				cooldownFree(m)
				m.userRepo.EXPECT().ExistsByEmail(mock.Anything, "jane@example.com").Return(false, errors.New("db down"))
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "db down")
			},
		},
		{
			name:  "storing the registration fails",
			input: validInput,
			setupMocks: func(m *registerMocks) {
				registrationOpen(m)
				hashed(m)
				cooldownFree(m)
				m.userRepo.EXPECT().ExistsByEmail(mock.Anything, "jane@example.com").Return(false, nil)
				m.registrationStorage.EXPECT().Create(mock.Anything, mock.Anything).Return("", errors.New("redis down"))
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "redis down")
			},
		},
		{
			name:  "confirmation notifier error propagates",
			input: validInput,
			setupMocks: func(m *registerMocks) {
				registrationOpen(m)
				hashed(m)
				cooldownFree(m)
				m.userRepo.EXPECT().ExistsByEmail(mock.Anything, "jane@example.com").Return(false, nil)
				m.registrationStorage.EXPECT().Create(mock.Anything, mock.Anything).Return("confirmation-token", nil)
				m.registrationNotifier.EXPECT().SendRegistrationConfirmation(mock.Anything, mock.Anything).Return(errors.New("send failed"))
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "send failed")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &registerMocks{
				userRepo:             repomock.NewMockUserRepository(t),
				featureFlagRepo:      repomock.NewMockFeatureFlagRepository(t),
				passwordService:      authmock.NewMockPasswordService(t),
				registrationStorage:  authmock.NewMockRegistrationStorage(t),
				registrationNotifier: authmock.NewMockRegistrationNotifier(t),
			}
			tt.setupMocks(m)

			service := auth.NewService(m.userRepo, nil, m.featureFlagRepo, nil, m.passwordService, nil, nil, nil, nil, m.registrationStorage, m.registrationNotifier, nil)

			result, err := service.Register(context.Background(), tt.input)

			if tt.wantErr != nil {
				tt.wantErr(t, err)
				assert.Nil(t, result)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, "check your email to finish the registration", result.Message)
		})
	}
}

// The three outcomes for a valid request must be indistinguishable to the caller.
func TestService_Register_AnswersAreIdentical(t *testing.T) {
	input := auth.RegisterInput{FirstName: "Jane", LastName: "Doe", Email: "jane@example.com", Password: "Password123"}

	outcomes := map[string]func(m *registerMocks){
		"free": func(m *registerMocks) {
			m.registrationStorage.EXPECT().AcquireCooldown(mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
			m.userRepo.EXPECT().ExistsByEmail(mock.Anything, mock.Anything).Return(false, nil)
			m.registrationStorage.EXPECT().Create(mock.Anything, mock.Anything).Return("token", nil)
			m.registrationNotifier.EXPECT().SendRegistrationConfirmation(mock.Anything, mock.Anything).Return(nil)
		},
		"taken": func(m *registerMocks) {
			m.registrationStorage.EXPECT().AcquireCooldown(mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
			m.userRepo.EXPECT().ExistsByEmail(mock.Anything, mock.Anything).Return(true, nil)
			m.registrationNotifier.EXPECT().SendRegistrationAttempt(mock.Anything, mock.Anything).Return(nil)
		},
		"cooldown": func(m *registerMocks) {
			m.registrationStorage.EXPECT().AcquireCooldown(mock.Anything, mock.Anything, mock.Anything).Return(false, nil)
		},
	}

	var results []*auth.RegisterResult
	for name, setup := range outcomes {
		m := &registerMocks{
			userRepo:             repomock.NewMockUserRepository(t),
			featureFlagRepo:      repomock.NewMockFeatureFlagRepository(t),
			passwordService:      authmock.NewMockPasswordService(t),
			registrationStorage:  authmock.NewMockRegistrationStorage(t),
			registrationNotifier: authmock.NewMockRegistrationNotifier(t),
		}
		m.featureFlagRepo.EXPECT().IsEnabled(mock.Anything, model.FeatureFlagRegistration).Return(true, nil)
		// The hash is computed on every path, so the response time does not depend on the outcome.
		m.passwordService.EXPECT().Hash("Password123").Return("hashed-password", nil).Once()
		setup(m)

		service := auth.NewService(m.userRepo, nil, m.featureFlagRepo, nil, m.passwordService, nil, nil, nil, nil, m.registrationStorage, m.registrationNotifier, nil)

		result, err := service.Register(context.Background(), input)
		require.NoError(t, err, name)
		results = append(results, result)
	}

	for _, r := range results[1:] {
		assert.Equal(t, results[0], r)
	}
}
