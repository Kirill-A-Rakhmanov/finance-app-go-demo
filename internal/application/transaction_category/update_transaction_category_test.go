package transaction_category_test

import (
	"context"
	"errors"
	"testing"

	apperror "finance-app-go-demo/internal/application/error"
	"finance-app-go-demo/internal/application/transaction_category"
	"finance-app-go-demo/internal/domain/enum"
	"finance-app-go-demo/internal/domain/model"
	repomock "finance-app-go-demo/internal/domain/repository/mock"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_UpdateTransactionCategory(t *testing.T) {
	categoryID := uuid.New()
	userID := uuid.New()
	newParentID := uuid.New()

	tests := []struct {
		name       string
		input      transaction_category.UpdateTransactionCategoryInput
		setupMocks func(repo *repomock.MockTransactionCategoryRepository)
		wantErr    func(t *testing.T, err error)
		wantResult func(t *testing.T, result *transaction_category.UpdateTransactionCategoryResult)
	}{
		{
			name: "successful update without parent change",
			input: transaction_category.UpdateTransactionCategoryInput{
				TransactionCategoryID: categoryID,
				UserID:                userID,
				Name:                  "Renamed",
			},
			setupMocks: func(repo *repomock.MockTransactionCategoryRepository) {
				repo.EXPECT().GetByIDAndUserID(mock.Anything, categoryID, userID).
					Return(&model.TransactionCategory{ID: categoryID, UserID: userID, Type: enum.Expense}, nil)
				repo.EXPECT().ExistsByUserIDAndTypeAndName(mock.Anything, userID, enum.Expense, "Renamed").Return(false, nil)
				repo.EXPECT().Update(mock.Anything, mock.Anything).
					RunAndReturn(func(ctx context.Context, c *model.TransactionCategory) (*model.TransactionCategory, error) {
						return c, nil
					})
			},
			wantResult: func(t *testing.T, result *transaction_category.UpdateTransactionCategoryResult) {
				assert.Equal(t, "Renamed", result.TransactionCategory.Name)
			},
		},
		{
			name: "successful update with valid new parent",
			input: transaction_category.UpdateTransactionCategoryInput{
				TransactionCategoryID: categoryID,
				UserID:                userID,
				ParentID:              &newParentID,
				Name:                  "Renamed",
			},
			setupMocks: func(repo *repomock.MockTransactionCategoryRepository) {
				repo.EXPECT().GetByIDAndUserID(mock.Anything, categoryID, userID).
					Return(&model.TransactionCategory{ID: categoryID, UserID: userID, Type: enum.Expense}, nil)
				repo.EXPECT().GetAllByParentID(mock.Anything, categoryID).Return(nil, nil)
				repo.EXPECT().GetByIDAndUserID(mock.Anything, newParentID, userID).
					Return(&model.TransactionCategory{ID: newParentID, UserID: userID, Type: enum.Expense}, nil)
				repo.EXPECT().ExistsByUserIDAndTypeAndName(mock.Anything, userID, enum.Expense, "Renamed").Return(false, nil)
				repo.EXPECT().Update(mock.Anything, mock.Anything).
					RunAndReturn(func(ctx context.Context, c *model.TransactionCategory) (*model.TransactionCategory, error) {
						return c, nil
					})
			},
			wantResult: func(t *testing.T, result *transaction_category.UpdateTransactionCategoryResult) {
				assert.Equal(t, newParentID, *result.TransactionCategory.ParentID)
			},
		},
		{
			name: "validation error - empty fields",
			input: transaction_category.UpdateTransactionCategoryInput{
				TransactionCategoryID: uuid.Nil,
				UserID:                uuid.Nil,
				Name:                  "",
			},
			setupMocks: func(repo *repomock.MockTransactionCategoryRepository) {
			},
			wantErr: func(t *testing.T, err error) {
				var validationErr *apperror.ValidationError
				require.True(t, errors.As(err, &validationErr))
				assert.Contains(t, validationErr.Fields, "transactionCategoryId")
				assert.Contains(t, validationErr.Fields, "userId")
				assert.Contains(t, validationErr.Fields, "name")
			},
		},
		{
			name: "category not found",
			input: transaction_category.UpdateTransactionCategoryInput{
				TransactionCategoryID: categoryID,
				UserID:                userID,
				Name:                  "Renamed",
			},
			setupMocks: func(repo *repomock.MockTransactionCategoryRepository) {
				repo.EXPECT().GetByIDAndUserID(mock.Anything, categoryID, userID).Return(nil, nil)
			},
			wantErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, apperror.ErrTransactionCategoryNotFound)
			},
		},
		{
			name: "cannot become its own parent",
			input: transaction_category.UpdateTransactionCategoryInput{
				TransactionCategoryID: categoryID,
				UserID:                userID,
				ParentID:              &categoryID,
				Name:                  "Renamed",
			},
			setupMocks: func(repo *repomock.MockTransactionCategoryRepository) {
				repo.EXPECT().GetByIDAndUserID(mock.Anything, categoryID, userID).
					Return(&model.TransactionCategory{ID: categoryID, UserID: userID, Type: enum.Expense}, nil)
			},
			wantErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, transaction_category.ErrTransactionCategorySelfParent)
			},
		},
		{
			name: "category with children cannot become a child itself",
			input: transaction_category.UpdateTransactionCategoryInput{
				TransactionCategoryID: categoryID,
				UserID:                userID,
				ParentID:              &newParentID,
				Name:                  "Renamed",
			},
			setupMocks: func(repo *repomock.MockTransactionCategoryRepository) {
				repo.EXPECT().GetByIDAndUserID(mock.Anything, categoryID, userID).
					Return(&model.TransactionCategory{ID: categoryID, UserID: userID, Type: enum.Expense}, nil)
				repo.EXPECT().GetAllByParentID(mock.Anything, categoryID).
					Return([]model.TransactionCategory{{ID: uuid.New(), ParentID: &categoryID}}, nil)
			},
			wantErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, transaction_category.ErrTransactionCategoryHasChildren)
			},
		},
		{
			name: "new parent not found",
			input: transaction_category.UpdateTransactionCategoryInput{
				TransactionCategoryID: categoryID,
				UserID:                userID,
				ParentID:              &newParentID,
				Name:                  "Renamed",
			},
			setupMocks: func(repo *repomock.MockTransactionCategoryRepository) {
				repo.EXPECT().GetByIDAndUserID(mock.Anything, categoryID, userID).
					Return(&model.TransactionCategory{ID: categoryID, UserID: userID, Type: enum.Expense}, nil)
				repo.EXPECT().GetAllByParentID(mock.Anything, categoryID).Return(nil, nil)
				repo.EXPECT().GetByIDAndUserID(mock.Anything, newParentID, userID).Return(nil, nil)
			},
			wantErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, apperror.ErrTransactionCategoryNotFound)
			},
		},
		{
			name: "new parent already has a parent",
			input: transaction_category.UpdateTransactionCategoryInput{
				TransactionCategoryID: categoryID,
				UserID:                userID,
				ParentID:              &newParentID,
				Name:                  "Renamed",
			},
			setupMocks: func(repo *repomock.MockTransactionCategoryRepository) {
				grandParentID := uuid.New()
				repo.EXPECT().GetByIDAndUserID(mock.Anything, categoryID, userID).
					Return(&model.TransactionCategory{ID: categoryID, UserID: userID, Type: enum.Expense}, nil)
				repo.EXPECT().GetAllByParentID(mock.Anything, categoryID).Return(nil, nil)
				repo.EXPECT().GetByIDAndUserID(mock.Anything, newParentID, userID).
					Return(&model.TransactionCategory{ID: newParentID, UserID: userID, ParentID: &grandParentID, Type: enum.Expense}, nil)
			},
			wantErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, transaction_category.ErrParentCategoryAlreadyHasParent)
			},
		},
		{
			name: "new parent type mismatch",
			input: transaction_category.UpdateTransactionCategoryInput{
				TransactionCategoryID: categoryID,
				UserID:                userID,
				ParentID:              &newParentID,
				Name:                  "Renamed",
			},
			setupMocks: func(repo *repomock.MockTransactionCategoryRepository) {
				repo.EXPECT().GetByIDAndUserID(mock.Anything, categoryID, userID).
					Return(&model.TransactionCategory{ID: categoryID, UserID: userID, Type: enum.Expense}, nil)
				repo.EXPECT().GetAllByParentID(mock.Anything, categoryID).Return(nil, nil)
				repo.EXPECT().GetByIDAndUserID(mock.Anything, newParentID, userID).
					Return(&model.TransactionCategory{ID: newParentID, UserID: userID, Type: enum.Income}, nil)
			},
			wantErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, transaction_category.ErrTransactionCategoryTypeMismatch)
			},
		},
		{
			name: "renaming to a name that already exists",
			input: transaction_category.UpdateTransactionCategoryInput{
				TransactionCategoryID: categoryID,
				UserID:                userID,
				Name:                  "Renamed",
			},
			setupMocks: func(repo *repomock.MockTransactionCategoryRepository) {
				repo.EXPECT().GetByIDAndUserID(mock.Anything, categoryID, userID).
					Return(&model.TransactionCategory{ID: categoryID, UserID: userID, Name: "Old", Type: enum.Expense}, nil)
				repo.EXPECT().ExistsByUserIDAndTypeAndName(mock.Anything, userID, enum.Expense, "Renamed").Return(true, nil)
			},
			wantErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, transaction_category.ErrTransactionCategoryAlreadyExists)
			},
		},
		{
			name: "unchanged name skips the uniqueness check",
			input: transaction_category.UpdateTransactionCategoryInput{
				TransactionCategoryID: categoryID,
				UserID:                userID,
				Name:                  "Same",
				Color:                 "#000000",
			},
			setupMocks: func(repo *repomock.MockTransactionCategoryRepository) {
				repo.EXPECT().GetByIDAndUserID(mock.Anything, categoryID, userID).
					Return(&model.TransactionCategory{ID: categoryID, UserID: userID, Name: "Same", Type: enum.Expense}, nil)
				repo.EXPECT().Update(mock.Anything, mock.Anything).
					RunAndReturn(func(ctx context.Context, c *model.TransactionCategory) (*model.TransactionCategory, error) {
						return c, nil
					})
			},
			wantResult: func(t *testing.T, result *transaction_category.UpdateTransactionCategoryResult) {
				assert.Equal(t, "#000000", result.TransactionCategory.Color)
			},
		},
		{
			name: "uniqueness check error",
			input: transaction_category.UpdateTransactionCategoryInput{
				TransactionCategoryID: categoryID,
				UserID:                userID,
				Name:                  "Renamed",
			},
			setupMocks: func(repo *repomock.MockTransactionCategoryRepository) {
				repo.EXPECT().GetByIDAndUserID(mock.Anything, categoryID, userID).
					Return(&model.TransactionCategory{ID: categoryID, UserID: userID, Name: "Old", Type: enum.Expense}, nil)
				repo.EXPECT().ExistsByUserIDAndTypeAndName(mock.Anything, userID, enum.Expense, "Renamed").Return(false, errors.New("db is down"))
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "db is down")
			},
		},
		{
			name: "repository update error",
			input: transaction_category.UpdateTransactionCategoryInput{
				TransactionCategoryID: categoryID,
				UserID:                userID,
				Name:                  "Renamed",
			},
			setupMocks: func(repo *repomock.MockTransactionCategoryRepository) {
				repo.EXPECT().GetByIDAndUserID(mock.Anything, categoryID, userID).
					Return(&model.TransactionCategory{ID: categoryID, UserID: userID, Type: enum.Expense}, nil)
				repo.EXPECT().ExistsByUserIDAndTypeAndName(mock.Anything, userID, enum.Expense, "Renamed").Return(false, nil)
				repo.EXPECT().Update(mock.Anything, mock.Anything).Return(nil, errors.New("update failed"))
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "update failed")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := repomock.NewMockTransactionCategoryRepository(t)
			tt.setupMocks(repo)

			service := transaction_category.New(repo)

			result, err := service.UpdateTransactionCategory(context.Background(), tt.input)

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
