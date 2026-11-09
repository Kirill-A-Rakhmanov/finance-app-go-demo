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

func TestService_CreateTransactionCategory(t *testing.T) {
	userID := uuid.New()
	parentID := uuid.New()

	tests := []struct {
		name       string
		input      transaction_category.CreateTransactionCategoryInput
		setupMocks func(repo *repomock.MockTransactionCategoryRepository)
		wantErr    func(t *testing.T, err error)
		wantResult func(t *testing.T, result *transaction_category.CreateTransactionCategoryResult)
	}{
		{
			name: "successful creation without parent",
			input: transaction_category.CreateTransactionCategoryInput{
				UserID: userID,
				Name:   "Groceries",
				Type:   enum.Expense,
			},
			setupMocks: func(repo *repomock.MockTransactionCategoryRepository) {
				repo.EXPECT().ExistsByUserIDAndTypeAndName(mock.Anything, userID, enum.Expense, "Groceries").Return(false, nil)
				repo.EXPECT().Create(mock.Anything, mock.Anything).
					RunAndReturn(func(ctx context.Context, c *model.TransactionCategory) (*model.TransactionCategory, error) {
						c.ID = uuid.New()
						return c, nil
					})
			},
			wantResult: func(t *testing.T, result *transaction_category.CreateTransactionCategoryResult) {
				assert.Equal(t, "Groceries", result.TransactionCategory.Name)
			},
		},
		{
			name: "successful creation with valid parent",
			input: transaction_category.CreateTransactionCategoryInput{
				UserID:   userID,
				ParentID: &parentID,
				Name:     "Restaurants",
				Type:     enum.Expense,
			},
			setupMocks: func(repo *repomock.MockTransactionCategoryRepository) {
				repo.EXPECT().GetByIDAndUserID(mock.Anything, parentID, userID).
					Return(&model.TransactionCategory{ID: parentID, UserID: userID, Type: enum.Expense}, nil)
				repo.EXPECT().ExistsByUserIDAndTypeAndName(mock.Anything, userID, enum.Expense, "Restaurants").Return(false, nil)
				repo.EXPECT().Create(mock.Anything, mock.Anything).
					RunAndReturn(func(ctx context.Context, c *model.TransactionCategory) (*model.TransactionCategory, error) {
						c.ID = uuid.New()
						return c, nil
					})
			},
			wantResult: func(t *testing.T, result *transaction_category.CreateTransactionCategoryResult) {
				assert.Equal(t, parentID, *result.TransactionCategory.ParentID)
			},
		},
		{
			name: "validation error - empty fields",
			input: transaction_category.CreateTransactionCategoryInput{
				UserID: uuid.Nil,
				Name:   "",
				Type:   "unknown",
			},
			setupMocks: func(repo *repomock.MockTransactionCategoryRepository) {
			},
			wantErr: func(t *testing.T, err error) {
				var validationErr *apperror.ValidationError
				require.True(t, errors.As(err, &validationErr))
				assert.Contains(t, validationErr.Fields, "userId")
				assert.Contains(t, validationErr.Fields, "name")
				assert.Contains(t, validationErr.Fields, "type")
			},
		},
		{
			name: "parent category not found",
			input: transaction_category.CreateTransactionCategoryInput{
				UserID:   userID,
				ParentID: &parentID,
				Name:     "Restaurants",
				Type:     enum.Expense,
			},
			setupMocks: func(repo *repomock.MockTransactionCategoryRepository) {
				repo.EXPECT().GetByIDAndUserID(mock.Anything, parentID, userID).Return(nil, nil)
			},
			wantErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, apperror.ErrTransactionCategoryNotFound)
			},
		},
		{
			name: "parent already has a parent",
			input: transaction_category.CreateTransactionCategoryInput{
				UserID:   userID,
				ParentID: &parentID,
				Name:     "Restaurants",
				Type:     enum.Expense,
			},
			setupMocks: func(repo *repomock.MockTransactionCategoryRepository) {
				grandParentID := uuid.New()
				repo.EXPECT().GetByIDAndUserID(mock.Anything, parentID, userID).
					Return(&model.TransactionCategory{ID: parentID, UserID: userID, ParentID: &grandParentID, Type: enum.Expense}, nil)
			},
			wantErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, transaction_category.ErrParentCategoryAlreadyHasParent)
			},
		},
		{
			name: "parent type mismatch",
			input: transaction_category.CreateTransactionCategoryInput{
				UserID:   userID,
				ParentID: &parentID,
				Name:     "Restaurants",
				Type:     enum.Expense,
			},
			setupMocks: func(repo *repomock.MockTransactionCategoryRepository) {
				repo.EXPECT().GetByIDAndUserID(mock.Anything, parentID, userID).
					Return(&model.TransactionCategory{ID: parentID, UserID: userID, Type: enum.Income}, nil)
			},
			wantErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, transaction_category.ErrTransactionCategoryTypeMismatch)
			},
		},
		{
			name: "category with the same name and type already exists",
			input: transaction_category.CreateTransactionCategoryInput{
				UserID: userID,
				Name:   "Groceries",
				Type:   enum.Expense,
			},
			setupMocks: func(repo *repomock.MockTransactionCategoryRepository) {
				repo.EXPECT().ExistsByUserIDAndTypeAndName(mock.Anything, userID, enum.Expense, "Groceries").Return(true, nil)
			},
			wantErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, transaction_category.ErrTransactionCategoryAlreadyExists)
			},
		},
		{
			name: "uniqueness check error",
			input: transaction_category.CreateTransactionCategoryInput{
				UserID: userID,
				Name:   "Groceries",
				Type:   enum.Expense,
			},
			setupMocks: func(repo *repomock.MockTransactionCategoryRepository) {
				repo.EXPECT().ExistsByUserIDAndTypeAndName(mock.Anything, userID, enum.Expense, "Groceries").Return(false, errors.New("db is down"))
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "db is down")
			},
		},
		{
			name: "repository create error",
			input: transaction_category.CreateTransactionCategoryInput{
				UserID: userID,
				Name:   "Groceries",
				Type:   enum.Expense,
			},
			setupMocks: func(repo *repomock.MockTransactionCategoryRepository) {
				repo.EXPECT().ExistsByUserIDAndTypeAndName(mock.Anything, userID, enum.Expense, "Groceries").Return(false, nil)
				repo.EXPECT().Create(mock.Anything, mock.Anything).Return(nil, errors.New("insert failed"))
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "insert failed")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := repomock.NewMockTransactionCategoryRepository(t)
			tt.setupMocks(repo)

			service := transaction_category.New(repo)

			result, err := service.CreateTransactionCategory(context.Background(), tt.input)

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
