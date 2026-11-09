package account_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"finance-app-go-demo/internal/application/account"
	apperror "finance-app-go-demo/internal/application/error"
	"finance-app-go-demo/internal/domain/model"
	repomock "finance-app-go-demo/internal/domain/repository/mock"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_CreateAccount(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name       string
		input      account.CreateAccountInput
		setupMocks func(
			accountRepo *repomock.MockAccountRepository,
			userRepo *repomock.MockUserRepository,
		)
		wantErr    func(t *testing.T, err error)
		wantResult func(t *testing.T, result *account.CreateAccountResult)
	}{
		{
			name: "successful creation",
			input: account.CreateAccountInput{
				UserID:         userID,
				Name:           "My Account",
				Icon:           "wallet",
				Currency:       "USD",
				InitialBalance: decimal.NewFromInt(100),
			},
			setupMocks: func(
				accountRepo *repomock.MockAccountRepository,
				userRepo *repomock.MockUserRepository,
			) {
				userRepo.EXPECT().ExistsByID(mock.Anything, userID).Return(true, nil)
				accountRepo.EXPECT().Create(mock.Anything, mock.Anything).
					RunAndReturn(func(ctx context.Context, acc *model.Account) (*model.Account, error) {
						acc.ID = uuid.New()
						return acc, nil
					})
			},
			wantResult: func(t *testing.T, result *account.CreateAccountResult) {
				assert.Equal(t, "My Account", result.Account.Name)
				assert.Equal(t, "USD", string(result.Account.Currency.Code))
			},
		},
		{
			name: "validation error - name and icon too long",
			input: account.CreateAccountInput{
				UserID: uuid.New(),
				Name:   strings.Repeat("a", model.AccountNameMaxLen+1),
				Icon:   strings.Repeat("a", model.AccountIconMaxLen+1),
			},
			setupMocks: func(
				accountRepo *repomock.MockAccountRepository,
				userRepo *repomock.MockUserRepository,
			) {
			},
			wantErr: func(t *testing.T, err error) {
				var validationErr *apperror.ValidationError
				require.True(t, errors.As(err, &validationErr))
				assert.Contains(t, validationErr.Fields, "name")
				assert.Contains(t, validationErr.Fields, "icon")
			},
		},
		{
			name: "validation error - empty fields",
			input: account.CreateAccountInput{
				UserID: uuid.Nil,
				Name:   "",
			},
			setupMocks: func(
				accountRepo *repomock.MockAccountRepository,
				userRepo *repomock.MockUserRepository,
			) {
			},
			wantErr: func(t *testing.T, err error) {
				var validationErr *apperror.ValidationError
				require.True(t, errors.As(err, &validationErr))
				assert.Contains(t, validationErr.Fields, "userId")
				assert.Contains(t, validationErr.Fields, "name")
			},
		},
		{
			name: "user not found",
			input: account.CreateAccountInput{
				UserID: userID,
				Name:   "My Account",
			},
			setupMocks: func(
				accountRepo *repomock.MockAccountRepository,
				userRepo *repomock.MockUserRepository,
			) {
				userRepo.EXPECT().ExistsByID(mock.Anything, userID).Return(false, nil)
			},
			wantErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, apperror.ErrUserNotFound)
			},
		},
		{
			name: "invalid currency code",
			input: account.CreateAccountInput{
				UserID:   userID,
				Name:     "My Account",
				Currency: "XXX",
			},
			setupMocks: func(
				accountRepo *repomock.MockAccountRepository,
				userRepo *repomock.MockUserRepository,
			) {
				userRepo.EXPECT().ExistsByID(mock.Anything, userID).Return(true, nil)
			},
			wantErr: func(t *testing.T, err error) {
				var appErr *apperror.Error
				require.True(t, errors.As(err, &appErr))
				assert.Equal(t, apperror.KindInvalid, appErr.Kind)
			},
		},
		{
			name: "repository create error",
			input: account.CreateAccountInput{
				UserID:   userID,
				Name:     "My Account",
				Currency: "USD",
			},
			setupMocks: func(
				accountRepo *repomock.MockAccountRepository,
				userRepo *repomock.MockUserRepository,
			) {
				userRepo.EXPECT().ExistsByID(mock.Anything, userID).Return(true, nil)
				accountRepo.EXPECT().Create(mock.Anything, mock.Anything).Return(nil, errors.New("insert failed"))
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "insert failed")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accountRepo := repomock.NewMockAccountRepository(t)
			userRepo := repomock.NewMockUserRepository(t)
			tt.setupMocks(accountRepo, userRepo)

			service := account.NewService(nil, accountRepo, nil, userRepo)

			result, err := service.CreateAccount(context.Background(), tt.input)

			if tt.wantErr != nil {
				tt.wantErr(t, err)
				assert.Nil(t, result)
				return
			}

			require.NoError(t, err)
			tt.wantResult(t, result)
		})
	}
}
