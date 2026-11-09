package user

import (
	"context"
	apperror "finance-app-go-demo/internal/application/error"
	"finance-app-go-demo/internal/domain/valueObject"

	"github.com/google/uuid"
)

type GetBalanceByUserIDInput struct {
	UserID uuid.UUID
}

type GetBalanceByUserIDResult struct {
	Balances []valueObject.Balance
}

func (s *Service) GetBalanceByUserID(ctx context.Context, input GetBalanceByUserIDInput) (*GetBalanceByUserIDResult, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	balances, err := s.transactionRepo.GetBalanceByUserID(ctx, input.UserID)
	if err != nil {
		return nil, err
	}

	return &GetBalanceByUserIDResult{
		Balances: balances,
	}, nil
}

func (i *GetBalanceByUserIDInput) Validate() error {
	v := apperror.NewValidationError()
	v.AddIf(i.UserID == uuid.Nil, "userId", "is required")
	return v.ErrorOrNil()
}
