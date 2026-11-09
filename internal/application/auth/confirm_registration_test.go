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
	transactionmock "finance-app-go-demo/pkg/transaction/mock"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type confirmMocks struct {
	userRepo            *repomock.MockUserRepository
	categoryRepo        *repomock.MockTransactionCategoryRepository
	txManager           *transactionmock.MockManager
	registrationStorage *authmock.MockRegistrationStorage
}

func TestService_ConfirmRegistration(t *testing.T) {
	pending := &auth.PendingRegistration{
		FirstName: "Jane", LastName: "Doe", Email: "jane@example.com", PasswordHash: "hashed-password",
	}

	runTx := func(m *confirmMocks) {
		m.txManager.EXPECT().
			Wrap(mock.Anything, mock.Anything).
			RunAndReturn(func(ctx context.Context, txFunc func(context.Context) error) error {
				return txFunc(ctx)
			})
	}

	tests := []struct {
		name       string
		input      auth.ConfirmRegistrationInput
		setupMocks func(m *confirmMocks)
		wantErr    func(t *testing.T, err error)
	}{
		{
			name:  "creates the user with default categories and consumes the token",
			input: auth.ConfirmRegistrationInput{Token: "token"},
			setupMocks: func(m *confirmMocks) {
				created := &model.User{ID: uuid.New(), FirstName: "Jane"}

				m.registrationStorage.EXPECT().Get(mock.Anything, "token").Return(pending, nil)
				m.userRepo.EXPECT().ExistsByEmail(mock.Anything, "jane@example.com").Return(false, nil)
				runTx(m)
				m.userRepo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(u *model.User) bool {
					return u.FirstName == "Jane" && u.LastName == "Doe" &&
						u.Email.ToString() == "jane@example.com" && u.Password == "hashed-password"
				})).Return(created, nil)
				m.categoryRepo.EXPECT().BulkCreate(mock.Anything, mock.MatchedBy(func(c []model.TransactionCategory) bool {
					return len(c) > 0 && c[0].UserID == created.ID
				})).Return(nil, nil)
				m.registrationStorage.EXPECT().Delete(mock.Anything, "token").Return(nil)
			},
		},
		{
			name:  "a failing cleanup does not fail the confirmation",
			input: auth.ConfirmRegistrationInput{Token: "token"},
			setupMocks: func(m *confirmMocks) {
				created := &model.User{ID: uuid.New()}

				m.registrationStorage.EXPECT().Get(mock.Anything, "token").Return(pending, nil)
				m.userRepo.EXPECT().ExistsByEmail(mock.Anything, "jane@example.com").Return(false, nil)
				runTx(m)
				m.userRepo.EXPECT().Create(mock.Anything, mock.Anything).Return(created, nil)
				m.categoryRepo.EXPECT().BulkCreate(mock.Anything, mock.Anything).Return(nil, nil)
				m.registrationStorage.EXPECT().Delete(mock.Anything, "token").Return(errors.New("redis down"))
			},
		},
		{
			name:  "unknown or expired token",
			input: auth.ConfirmRegistrationInput{Token: "token"},
			setupMocks: func(m *confirmMocks) {
				m.registrationStorage.EXPECT().Get(mock.Anything, "token").Return(nil, nil)
			},
			wantErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, auth.ErrInvalidRegistrationToken)
			},
		},
		{
			name:  "already confirmed: conflict, and the leftover entry is removed",
			input: auth.ConfirmRegistrationInput{Token: "token"},
			setupMocks: func(m *confirmMocks) {
				m.registrationStorage.EXPECT().Get(mock.Anything, "token").Return(pending, nil)
				m.userRepo.EXPECT().ExistsByEmail(mock.Anything, "jane@example.com").Return(true, nil)
				m.registrationStorage.EXPECT().Delete(mock.Anything, "token").Return(nil)
			},
			wantErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, auth.ErrUserExists)
			},
		},
		{
			name:       "validation error - empty token",
			input:      auth.ConfirmRegistrationInput{},
			setupMocks: func(m *confirmMocks) {},
			wantErr: func(t *testing.T, err error) {
				var validationErr *apperror.ValidationError
				require.True(t, errors.As(err, &validationErr))
				assert.Contains(t, validationErr.Fields, "token")
			},
		},
		{
			name:  "storage error propagates",
			input: auth.ConfirmRegistrationInput{Token: "token"},
			setupMocks: func(m *confirmMocks) {
				m.registrationStorage.EXPECT().Get(mock.Anything, "token").Return(nil, errors.New("redis down"))
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "redis down")
			},
		},
		{
			name:  "user lookup error propagates",
			input: auth.ConfirmRegistrationInput{Token: "token"},
			setupMocks: func(m *confirmMocks) {
				m.registrationStorage.EXPECT().Get(mock.Anything, "token").Return(pending, nil)
				m.userRepo.EXPECT().ExistsByEmail(mock.Anything, "jane@example.com").Return(false, errors.New("db down"))
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "db down")
			},
		},
		{
			name:  "failure while creating the user keeps the token usable",
			input: auth.ConfirmRegistrationInput{Token: "token"},
			setupMocks: func(m *confirmMocks) {
				m.registrationStorage.EXPECT().Get(mock.Anything, "token").Return(pending, nil)
				m.userRepo.EXPECT().ExistsByEmail(mock.Anything, "jane@example.com").Return(false, nil)
				runTx(m)
				m.userRepo.EXPECT().Create(mock.Anything, mock.Anything).Return(nil, errors.New("insert failed"))
				// Delete is not expected: the person can click the link again.
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "insert failed")
			},
		},
		{
			name:  "failure while creating categories keeps the token usable",
			input: auth.ConfirmRegistrationInput{Token: "token"},
			setupMocks: func(m *confirmMocks) {
				m.registrationStorage.EXPECT().Get(mock.Anything, "token").Return(pending, nil)
				m.userRepo.EXPECT().ExistsByEmail(mock.Anything, "jane@example.com").Return(false, nil)
				runTx(m)
				m.userRepo.EXPECT().Create(mock.Anything, mock.Anything).Return(&model.User{ID: uuid.New()}, nil)
				m.categoryRepo.EXPECT().BulkCreate(mock.Anything, mock.Anything).Return(nil, errors.New("insert failed"))
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "insert failed")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &confirmMocks{
				userRepo:            repomock.NewMockUserRepository(t),
				categoryRepo:        repomock.NewMockTransactionCategoryRepository(t),
				txManager:           transactionmock.NewMockManager(t),
				registrationStorage: authmock.NewMockRegistrationStorage(t),
			}
			tt.setupMocks(m)

			service := auth.NewService(m.userRepo, m.categoryRepo, nil, nil, nil, m.txManager, nil, nil, nil, m.registrationStorage, nil, nil)

			result, err := service.ConfirmRegistration(context.Background(), tt.input)

			if tt.wantErr != nil {
				tt.wantErr(t, err)
				assert.Nil(t, result)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, "registration confirmed, you can log in now", result.Message)
		})
	}
}
