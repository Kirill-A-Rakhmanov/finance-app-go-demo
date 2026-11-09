package account_test

import (
	"context"
	"errors"
	"testing"

	"finance-app-go-demo/internal/application/account"
	apperror "finance-app-go-demo/internal/application/error"
	"finance-app-go-demo/internal/domain/model"
	repomock "finance-app-go-demo/internal/domain/repository/mock"
	transactionmock "finance-app-go-demo/pkg/transaction/mock"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func runTx(txManager *transactionmock.MockManager) {
	txManager.EXPECT().
		Wrap(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, txFunc func(context.Context) error) error {
			return txFunc(ctx)
		})
}

func TestService_UpdateAccount(t *testing.T) {
	accountID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name       string
		input      account.UpdateAccountInput
		setupMocks func(
			accountRepo *repomock.MockAccountRepository,
			txManager *transactionmock.MockManager,
		)
		wantErr    func(t *testing.T, err error)
		wantResult func(t *testing.T, result *account.UpdateAccountResult)
	}{
		{
			name: "successful update",
			input: account.UpdateAccountInput{
				AccountID: accountID,
				UserID:    userID,
				Name:      "Renamed Account",
				Icon:      "bank",
			},
			setupMocks: func(
				accountRepo *repomock.MockAccountRepository,
				txManager *transactionmock.MockManager,
			) {
				runTx(txManager)
				existing := &model.Account{ID: accountID, UserID: userID, Name: "Old Name"}
				accountRepo.EXPECT().GetByIDAndUserID(mock.Anything, accountID, userID).Return(existing, nil)
				accountRepo.EXPECT().Update(mock.Anything, mock.Anything).
					RunAndReturn(func(ctx context.Context, acc *model.Account) (*model.Account, error) {
						return acc, nil
					})
			},
			wantResult: func(t *testing.T, result *account.UpdateAccountResult) {
				assert.Equal(t, "Renamed Account", result.Account.Name)
				assert.Equal(t, "bank", result.Account.Icon)
			},
		},
		{
			name: "validation error - empty fields",
			input: account.UpdateAccountInput{
				AccountID: uuid.Nil,
				UserID:    uuid.Nil,
				Name:      "",
			},
			setupMocks: func(
				accountRepo *repomock.MockAccountRepository,
				txManager *transactionmock.MockManager,
			) {
			},
			wantErr: func(t *testing.T, err error) {
				var validationErr *apperror.ValidationError
				require.True(t, errors.As(err, &validationErr))
				assert.Contains(t, validationErr.Fields, "accountId")
				assert.Contains(t, validationErr.Fields, "userId")
				assert.Contains(t, validationErr.Fields, "name")
			},
		},
		{
			name: "account not found",
			input: account.UpdateAccountInput{
				AccountID: accountID,
				UserID:    userID,
				Name:      "Renamed Account",
			},
			setupMocks: func(
				accountRepo *repomock.MockAccountRepository,
				txManager *transactionmock.MockManager,
			) {
				runTx(txManager)
				accountRepo.EXPECT().GetByIDAndUserID(mock.Anything, accountID, userID).Return(nil, nil)
			},
			wantErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, apperror.ErrAccountNotFound)
			},
		},
		{
			name: "repository update error",
			input: account.UpdateAccountInput{
				AccountID: accountID,
				UserID:    userID,
				Name:      "Renamed Account",
			},
			setupMocks: func(
				accountRepo *repomock.MockAccountRepository,
				txManager *transactionmock.MockManager,
			) {
				runTx(txManager)
				existing := &model.Account{ID: accountID, UserID: userID, Name: "Old Name"}
				accountRepo.EXPECT().GetByIDAndUserID(mock.Anything, accountID, userID).Return(existing, nil)
				accountRepo.EXPECT().Update(mock.Anything, mock.Anything).Return(nil, errors.New("update failed"))
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "update failed")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accountRepo := repomock.NewMockAccountRepository(t)
			txManager := transactionmock.NewMockManager(t)
			tt.setupMocks(accountRepo, txManager)

			service := account.NewService(txManager, accountRepo, nil, nil)

			result, err := service.UpdateAccount(context.Background(), tt.input)

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
