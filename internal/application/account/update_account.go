package account

import (
	"context"
	apperror "finance-app-go-demo/internal/application/error"
	"finance-app-go-demo/internal/domain/model"

	"github.com/google/uuid"
)

type UpdateAccountInput struct {
	AccountID                uuid.UUID
	UserID                   uuid.UUID
	Name                     string
	Icon                     string
	IsNegativeBalanceEnabled bool
	IsHidden                 bool
	IsExcludedFromCount      bool
}

type UpdateAccountResult struct {
	Account *model.Account
}

func (s *Service) UpdateAccount(ctx context.Context, input UpdateAccountInput) (*UpdateAccountResult, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	var result UpdateAccountResult

	err := s.txManager.Wrap(ctx, func(ctx context.Context) error {
		account, err := s.accountRepo.GetByIDAndUserID(ctx, input.AccountID, input.UserID)
		if err != nil {
			return err
		}
		if account == nil {
			return apperror.ErrAccountNotFound
		}

		account.Name = input.Name
		account.Icon = input.Icon
		account.IsNegativeBalanceEnabled = input.IsNegativeBalanceEnabled
		account.IsHidden = input.IsHidden
		account.IsExcludedFromCount = input.IsExcludedFromCount

		account, err = s.accountRepo.Update(ctx, account)
		if err != nil {
			return err
		}

		result.Account = account

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (i *UpdateAccountInput) Validate() error {
	v := apperror.NewValidationError()
	v.AddIf(i.AccountID == uuid.Nil, "accountId", "is required")
	v.AddIf(i.UserID == uuid.Nil, "userId", "is required")
	v.AddIf(i.Name == "", "name", "is required")
	v.MaxLength("name", i.Name, model.AccountNameMaxLen)
	v.MaxLength("icon", i.Icon, model.AccountIconMaxLen)
	return v.ErrorOrNil()
}
