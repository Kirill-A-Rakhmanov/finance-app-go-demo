package transaction_category

import (
	"context"
	apperror "finance-app-go-demo/internal/application/error"
	"finance-app-go-demo/internal/domain/enum"
	"finance-app-go-demo/internal/domain/model"

	"github.com/google/uuid"
)

type CreateTransactionCategoryInput struct {
	UserID   uuid.UUID
	ParentID *uuid.UUID
	Name     string
	Icon     string
	Color    string
	Type     enum.TransactionCategoryType
}

type CreateTransactionCategoryResult struct {
	TransactionCategory *model.TransactionCategory
}

func (s *Service) CreateTransactionCategory(ctx context.Context, input CreateTransactionCategoryInput) (*CreateTransactionCategoryResult, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	if input.ParentID != nil {
		parentCategory, err := s.transactionCategoryRepo.GetByIDAndUserID(ctx, *input.ParentID, input.UserID)
		if err != nil {
			return nil, err
		}
		if parentCategory == nil {
			return nil, apperror.ErrTransactionCategoryNotFound
		}
		if parentCategory.ParentID != nil {
			return nil, ErrParentCategoryAlreadyHasParent
		}
		if parentCategory.Type != input.Type {
			return nil, ErrTransactionCategoryTypeMismatch
		}
	}

	exists, err := s.transactionCategoryRepo.ExistsByUserIDAndTypeAndName(ctx, input.UserID, input.Type, input.Name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrTransactionCategoryAlreadyExists
	}

	category := model.NewTransactionCategory(
		input.UserID,
		input.ParentID,
		input.Name,
		input.Icon,
		input.Color,
		input.Type,
	)

	category, err = s.transactionCategoryRepo.Create(ctx, category)
	if err != nil {
		return nil, err
	}

	return &CreateTransactionCategoryResult{
		TransactionCategory: category,
	}, nil
}

func (i *CreateTransactionCategoryInput) Validate() error {
	v := apperror.NewValidationError()
	v.AddIf(i.UserID == uuid.Nil, "userId", "is required")
	v.AddIf(i.Name == "", "name", "is required")
	v.MaxLength("name", i.Name, model.TransactionCategoryNameMaxLen)
	v.MaxLength("icon", i.Icon, model.TransactionCategoryIconMaxLen)
	v.MaxLength("color", i.Color, model.TransactionCategoryColorMaxLen)
	v.AddIf(i.Type != enum.Income && i.Type != enum.Expense, "type", "must be income or expense")
	return v.ErrorOrNil()
}
