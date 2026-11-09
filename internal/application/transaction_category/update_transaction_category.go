package transaction_category

import (
	"context"
	apperror "finance-app-go-demo/internal/application/error"
	"finance-app-go-demo/internal/domain/model"

	"github.com/google/uuid"
)

type UpdateTransactionCategoryInput struct {
	TransactionCategoryID uuid.UUID
	UserID                uuid.UUID
	ParentID              *uuid.UUID
	Name                  string
	Icon                  string
	Color                 string
}

type UpdateTransactionCategoryResult struct {
	TransactionCategory *model.TransactionCategory
}

func (s *Service) UpdateTransactionCategory(
	ctx context.Context,
	input UpdateTransactionCategoryInput,
) (*UpdateTransactionCategoryResult, error) {
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

	if input.ParentID != nil {
		if category.ID == *input.ParentID {
			return nil, ErrTransactionCategorySelfParent
		}

		children, err := s.transactionCategoryRepo.GetAllByParentID(ctx, category.ID)
		if err != nil {
			return nil, err
		}
		if len(children) > 0 {
			return nil, ErrTransactionCategoryHasChildren
		}

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
		if parentCategory.Type != category.Type {
			return nil, ErrTransactionCategoryTypeMismatch
		}
	}

	if category.Name != input.Name {
		exists, err := s.transactionCategoryRepo.ExistsByUserIDAndTypeAndName(ctx, input.UserID, category.Type, input.Name)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrTransactionCategoryAlreadyExists
		}
	}

	category.ParentID = input.ParentID
	category.Name = input.Name
	category.Icon = input.Icon
	category.Color = input.Color

	category, err = s.transactionCategoryRepo.Update(ctx, category)
	if err != nil {
		return nil, err
	}

	return &UpdateTransactionCategoryResult{
		TransactionCategory: category,
	}, nil
}

func (i *UpdateTransactionCategoryInput) Validate() error {
	v := apperror.NewValidationError()
	v.AddIf(i.TransactionCategoryID == uuid.Nil, "transactionCategoryId", "is required")
	v.AddIf(i.UserID == uuid.Nil, "userId", "is required")
	v.AddIf(i.Name == "", "name", "is required")
	v.MaxLength("name", i.Name, model.TransactionCategoryNameMaxLen)
	v.MaxLength("icon", i.Icon, model.TransactionCategoryIconMaxLen)
	v.MaxLength("color", i.Color, model.TransactionCategoryColorMaxLen)
	return v.ErrorOrNil()
}
