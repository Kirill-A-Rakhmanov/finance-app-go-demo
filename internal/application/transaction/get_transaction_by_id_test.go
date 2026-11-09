package transaction_test

import (
	"context"
	"errors"
	"testing"

	apperror "finance-app-go-demo/internal/application/error"
	"finance-app-go-demo/internal/application/transaction"
	"finance-app-go-demo/internal/domain/model"
	repomock "finance-app-go-demo/internal/domain/repository/mock"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_GetTransactionById(t *testing.T) {
	transactionID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name       string
		input      transaction.GetTrainsactionByIdInput
		setupMocks func(transactionRepo *repomock.MockTransactionRepository)
		wantErr    func(t *testing.T, err error)
		wantResult func(t *testing.T, result *transaction.GetTransactionByIdResult)
	}{
		{
			name:  "successful fetch",
			input: transaction.GetTrainsactionByIdInput{TransactionID: transactionID, UserID: userID},
			setupMocks: func(transactionRepo *repomock.MockTransactionRepository) {
				transactionRepo.EXPECT().GetByIDAndUserID(mock.Anything, transactionID, userID).
					Return(&model.Transaction{ID: transactionID}, nil)
			},
			wantResult: func(t *testing.T, result *transaction.GetTransactionByIdResult) {
				assert.Equal(t, transactionID, result.Transaction.ID)
			},
		},
		{
			name:  "validation error - empty fields",
			input: transaction.GetTrainsactionByIdInput{TransactionID: uuid.Nil, UserID: uuid.Nil},
			setupMocks: func(transactionRepo *repomock.MockTransactionRepository) {
			},
			wantErr: func(t *testing.T, err error) {
				var validationErr *apperror.ValidationError
				require.True(t, errors.As(err, &validationErr))
				assert.Contains(t, validationErr.Fields, "transactionId")
				assert.Contains(t, validationErr.Fields, "userId")
			},
		},
		{
			name:  "transaction not found",
			input: transaction.GetTrainsactionByIdInput{TransactionID: transactionID, UserID: userID},
			setupMocks: func(transactionRepo *repomock.MockTransactionRepository) {
				transactionRepo.EXPECT().GetByIDAndUserID(mock.Anything, transactionID, userID).Return(nil, nil)
			},
			wantErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, transaction.ErrTransactionNotFound)
			},
		},
		{
			name:  "repository error propagates",
			input: transaction.GetTrainsactionByIdInput{TransactionID: transactionID, UserID: userID},
			setupMocks: func(transactionRepo *repomock.MockTransactionRepository) {
				transactionRepo.EXPECT().GetByIDAndUserID(mock.Anything, transactionID, userID).Return(nil, errors.New("connection refused"))
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "connection refused")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transactionRepo := repomock.NewMockTransactionRepository(t)
			tt.setupMocks(transactionRepo)

			service := transaction.NewService(transactionRepo, nil, nil, nil, nil, nil)

			result, err := service.GetTransactionById(context.Background(), tt.input)

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
