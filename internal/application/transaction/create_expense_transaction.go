package transaction

import (
	"context"
	apperror "finance-app-go-demo/internal/application/error"
	"finance-app-go-demo/internal/application/validation"
	"finance-app-go-demo/internal/domain/enum"
	"finance-app-go-demo/internal/domain/model"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CreateExpenseTransactionInput struct {
	UserID                  uuid.UUID
	AccountID               uuid.UUID
	TransactionCategoryName *string
	CounterpartyName        *string
	Amount                  decimal.Decimal
	Name                    *string
	Date                    time.Time
}

type CreateExpenseTransactionResult struct {
	Transaction *model.Transaction
}

func (s *Service) CreateExpenseTransaction(ctx context.Context, input CreateExpenseTransactionInput) (*CreateExpenseTransactionResult, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	var result *CreateExpenseTransactionResult

	err := s.txManager.Wrap(ctx, func(ctx context.Context) error {
		account, err := s.accountRepo.GetByIDAndUserID(ctx, input.AccountID, input.UserID)
		if err != nil {
			return err
		}
		if account == nil {
			return apperror.ErrAccountNotFound
		}

		if !account.IsNegativeBalanceEnabled {
			if err := s.txManager.Lock(ctx, model.AccountBalanceLockKey(account.ID)); err != nil {
				return err
			}

			balance, err := s.transactionRepo.GetBalanceByAccountID(ctx, account.ID)
			if err != nil {
				return err
			}

			if balance.Sub(input.Amount).IsNegative() {
				return apperror.ErrNegativeBalanceNotAllowed
			}
		}

		categoryID, err := s.resolveTransactionCategoryID(ctx, input.UserID, enum.TransactionExpense, input.TransactionCategoryName)
		if err != nil {
			return err
		}

		counterpartyID, err := s.resolveCounterpartyID(ctx, input.UserID, input.CounterpartyName)
		if err != nil {
			return err
		}

		transaction, err := s.transactionRepo.Create(ctx, model.NewExpenseTransaction(
			account.ID,
			categoryID,
			counterpartyID,
			input.Amount,
			input.Name,
			input.Date,
			true,
		))
		if err != nil {
			return err
		}

		result = &CreateExpenseTransactionResult{Transaction: transaction}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (i *CreateExpenseTransactionInput) Validate() error {
	v := apperror.NewValidationError()
	v.AddIf(i.AccountID == uuid.Nil, "accountId", "is required")
	validation.Amount(v, "amount", i.Amount)
	v.MaxLengthPtr("name", i.Name, model.TransactionNameMaxLen)
	v.MaxLengthPtr("transactionCategoryName", i.TransactionCategoryName, model.TransactionCategoryNameMaxLen)
	v.MaxLengthPtr("counterpartyName", i.CounterpartyName, model.CounterpartyNameMaxLen)
	return v.ErrorOrNil()
}
