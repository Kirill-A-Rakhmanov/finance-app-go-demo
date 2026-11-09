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

func TestService_DeleteTransactionCategory(t *testing.T) {
	categoryID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name       string
		input      transaction_category.DeleteTransactionCategoryInput
		setupMocks func(repo *repomock.MockTransactionCategoryRepository)
		wantErr    func(t *testing.T, err error)
		wantResult func(t *testing.T, result *transaction_category.DeleteTransactionCategoryResult)
	}{
		{
			name:  "successful deletion",
			input: transaction_category.DeleteTransactionCategoryInput{TransactionCategoryID: categoryID, UserID: userID},
			setupMocks: func(repo *repomock.MockTransactionCategoryRepository) {
				repo.EXPECT().GetByIDAndUserID(mock.Anything, categoryID, userID).
					Return(&model.TransactionCategory{ID: categoryID, UserID: userID}, nil)
				repo.EXPECT().DeleteByID(mock.Anything, categoryID).Return(nil)
			},
			wantResult: func(t *testing.T, result *transaction_category.DeleteTransactionCategoryResult) {
				assert.Equal(t, "Transaction category deleted successfully", result.Message)
			},
		},
		{
			name:  "validation error - empty fields",
			input: transaction_category.DeleteTransactionCategoryInput{TransactionCategoryID: uuid.Nil, UserID: uuid.Nil},
			setupMocks: func(repo *repomock.MockTransactionCategoryRepository) {
			},
			wantErr: func(t *testing.T, err error) {
				var validationErr *apperror.ValidationError
				require.True(t, errors.As(err, &validationErr))
				assert.Contains(t, validationErr.Fields, "transactionCategoryId")
				assert.Contains(t, validationErr.Fields, "userId")
			},
		},
		{
			name:  "category not found",
			input: transaction_category.DeleteTransactionCategoryInput{TransactionCategoryID: categoryID, UserID: userID},
			setupMocks: func(repo *repomock.MockTransactionCategoryRepository) {
				repo.EXPECT().GetByIDAndUserID(mock.Anything, categoryID, userID).Return(nil, nil)
			},
			wantErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, apperror.ErrTransactionCategoryNotFound)
			},
		},
		{
			name:  "repository delete error",
			input: transaction_category.DeleteTransactionCategoryInput{TransactionCategoryID: categoryID, UserID: userID},
			setupMocks: func(repo *repomock.MockTransactionCategoryRepository) {
				repo.EXPECT().GetByIDAndUserID(mock.Anything, categoryID, userID).
					Return(&model.TransactionCategory{ID: categoryID, UserID: userID}, nil)
				repo.EXPECT().DeleteByID(mock.Anything, categoryID).Return(errors.New("delete failed"))
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "delete failed")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := repomock.NewMockTransactionCategoryRepository(t)
			tt.setupMocks(repo)

			service := transaction_category.New(repo)

			result, err := service.DeleteTransactionCategory(context.Background(), tt.input)

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
