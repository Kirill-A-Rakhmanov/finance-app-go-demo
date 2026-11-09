package transaction_category

import (
	"context"
	apperror "finance-app-go-demo/internal/application/error"
	"finance-app-go-demo/internal/domain/model"

	"github.com/google/uuid"
)

type GetTransactionCategoriesByUserIDInput struct {
	UserID uuid.UUID
}

type GetTransactionCategoriesByUserIDResult struct {
	TransactionCategories []model.TransactionCategory
}

func (s *Service) GetTransactionCategoriesByUserID(
	ctx context.Context,
	input GetTransactionCategoriesByUserIDInput,
) (*GetTransactionCategoriesByUserIDResult, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	categories, err := s.transactionCategoryRepo.GetByUserID(ctx, input.UserID)
	if err != nil {
		return nil, err
	}

	return &GetTransactionCategoriesByUserIDResult{
		TransactionCategories: categories,
	}, nil
}

func (i *GetTransactionCategoriesByUserIDInput) Validate() error {
	v := apperror.NewValidationError()
	v.AddIf(i.UserID == uuid.Nil, "userId", "is required")
	return v.ErrorOrNil()
}
