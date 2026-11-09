package account

import (
	"context"
	apperror "finance-app-go-demo/internal/application/error"
	"finance-app-go-demo/internal/domain/valueObject"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type GetBalanceByAccountIdInput struct {
	AccountID uuid.UUID
	UserID    uuid.UUID
}

type GetBalanceByAccountIdResult struct {
	Balance  decimal.Decimal
	Currency valueObject.Currency
}

func (s *Service) GetBalanceByAccountId(ctx context.Context, input GetBalanceByAccountIdInput) (*GetBalanceByAccountIdResult, error) {
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

	balance, err := s.transactionRepo.GetBalanceByAccountID(ctx, input.AccountID)
	if err != nil {
		return nil, err
	}

	return &GetBalanceByAccountIdResult{
		Balance:  balance,
		Currency: account.Currency,
	}, nil
}

func (i *GetBalanceByAccountIdInput) Validate() error {
	v := apperror.NewValidationError()
	v.AddIf(i.AccountID == uuid.Nil, "accountId", "is required")
	v.AddIf(i.UserID == uuid.Nil, "userId", "is required")
	return v.ErrorOrNil()
}
