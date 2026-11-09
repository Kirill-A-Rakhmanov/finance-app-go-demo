package account

import (
	"context"
	apperror "finance-app-go-demo/internal/application/error"
	"finance-app-go-demo/internal/domain/model"

	"github.com/google/uuid"
)

type GetAccountByIdInput struct {
	UserID    uuid.UUID
	AccountID uuid.UUID
}

type GetAccountByIdResult struct {
	Account model.Account
}

func (s *Service) GetAccountById(ctx context.Context, input GetAccountByIdInput) (*GetAccountByIdResult, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	account, err := s.accountRepo.GetByIDAndUserID(ctx, input.AccountID, input.UserID)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, apperror.ErrAccountNotFound
	}

	return &GetAccountByIdResult{
		Account: *account,
	}, nil
}

func (i *GetAccountByIdInput) Validate() error {
	v := apperror.NewValidationError()
	v.AddIf(i.AccountID == uuid.Nil, "accountId", "is required")
	v.AddIf(i.UserID == uuid.Nil, "userId", "is required")
	return v.ErrorOrNil()
}
