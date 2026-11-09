package transaction_category_test

import (
	"context"
	"errors"
	"testing"

	apperror "finance-app-go-demo/internal/application/error"
	"finance-app-go-demo/internal/application/transaction_category"
	"finance-app-go-demo/internal/domain/model"
	repomock "finance-app-go-demo/internal/domain/repository/mock"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_GetTransactionCategoriesByUserID(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name       string
		input      transaction_category.GetTransactionCategoriesByUserIDInput
		setupMocks func(repo *repomock.MockTransactionCategoryRepository)
		wantErr    func(t *testing.T, err error)
		wantResult func(t *testing.T, result *transaction_category.GetTransactionCategoriesByUserIDResult)
	}{
		{
			name:  "successful fetch",
			input: transaction_category.GetTransactionCategoriesByUserIDInput{UserID: userID},
			setupMocks: func(repo *repomock.MockTransactionCategoryRepository) {
				repo.EXPECT().GetByUserID(mock.Anything, userID).Return([]model.TransactionCategory{
					{ID: uuid.New(), UserID: userID, Name: "Groceries"},
					{ID: uuid.New(), UserID: userID, Name: "Salary"},
				}, nil)
			},
			wantResult: func(t *testing.T, result *transaction_category.GetTransactionCategoriesByUserIDResult) {
				assert.Len(t, result.TransactionCategories, 2)
			},
		},
		{
			name:  "validation error - empty user id",
			input: transaction_category.GetTransactionCategoriesByUserIDInput{UserID: uuid.Nil},
			setupMocks: func(repo *repomock.MockTransactionCategoryRepository) {
			},
			wantErr: func(t *testing.T, err error) {
				var validationErr *apperror.ValidationError
				require.True(t, errors.As(err, &validationErr))
				assert.Contains(t, validationErr.Fields, "userId")
			},
		},
		{
			name:  "repository error propagates",
			input: transaction_category.GetTransactionCategoriesByUserIDInput{UserID: userID},
			setupMocks: func(repo *repomock.MockTransactionCategoryRepository) {
				repo.EXPECT().GetByUserID(mock.Anything, userID).Return(nil, errors.New("connection refused"))
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "connection refused")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := repomock.NewMockTransactionCategoryRepository(t)
			tt.setupMocks(repo)

			service := transaction_category.New(repo)

			result, err := service.GetTransactionCategoriesByUserID(context.Background(), tt.input)

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
