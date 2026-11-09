package account

import (
	"context"
	apperror "finance-app-go-demo/internal/application/error"
	"finance-app-go-demo/internal/domain/model"

	"github.com/google/uuid"
)

type GetAccountsByUserIdInput struct {
	UserID uuid.UUID
}

type GetAccountsByUserIdResult struct {
	Accounts []model.Account
}

func (s *Service) GetAccountsByUserId(ctx context.Context, input GetAccountsByUserIdInput) (*GetAccountsByUserIdResult, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	accounts, err := s.accountRepo.GetByUserID(ctx, input.UserID)
	if err != nil {
		return nil, err
	}

	return &GetAccountsByUserIdResult{
		Accounts: accounts,
	}, nil
}

func (i *GetAccountsByUserIdInput) Validate() error {
	v := apperror.NewValidationError()
	v.AddIf(i.UserID == uuid.Nil, "userId", "is required")
	return v.ErrorOrNil()
}
