package transaction_category

import (
	"context"
	apperror "finance-app-go-demo/internal/application/error"

	"github.com/google/uuid"
)

type DeleteTransactionCategoryInput struct {
	UserID                uuid.UUID
	TransactionCategoryID uuid.UUID
}

type DeleteTransactionCategoryResult struct {
	Message string
}

func (s *Service) DeleteTransactionCategory(ctx context.Context, input DeleteTransactionCategoryInput) (*DeleteTransactionCategoryResult, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	category, err := s.transactionCategoryRepo.GetByIDAndUserID(ctx, input.TransactionCategoryID, input.UserID)
	if err != nil {
		return nil, err
	}
	if category == nil {
		return nil, apperror.ErrTransactionCategoryNotFound
	}

	err = s.transactionCategoryRepo.DeleteByID(ctx, category.ID)
	if err != nil {
		return nil, err
	}

	return &DeleteTransactionCategoryResult{
		Message: "Transaction category deleted successfully",
	}, nil
}

func (i *DeleteTransactionCategoryInput) Validate() error {
	v := apperror.NewValidationError()
	v.AddIf(i.TransactionCategoryID == uuid.Nil, "transactionCategoryId", "is required")
	v.AddIf(i.UserID == uuid.Nil, "userId", "is required")
	return v.ErrorOrNil()
}
