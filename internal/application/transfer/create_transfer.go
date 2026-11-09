package transfer

import (
	"context"
	apperror "finance-app-go-demo/internal/application/error"
	"finance-app-go-demo/internal/application/validation"
	"finance-app-go-demo/internal/domain/model"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CreateTransferInput struct {
	UserID        uuid.UUID
	FromAccountID uuid.UUID
	ToAccountID   uuid.UUID
	FromAmount    decimal.Decimal
	ToAmount      decimal.Decimal
	Date          time.Time
}

type CreateTransferResult struct {
	Transfer     *model.Transfer
	Transactions []model.Transaction
}

func (s *Service) CreateTransfer(ctx context.Context, input CreateTransferInput) (*CreateTransferResult, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	if input.FromAccountID == input.ToAccountID {
		return nil, ErrTransferToSameAccountNotAllowed
	}

	var result CreateTransferResult

	err := s.txManager.Wrap(ctx, func(ctx context.Context) error {
		fromAccount, err := s.accountRepo.GetByIDAndUserID(ctx, input.FromAccountID, input.UserID)
		if err != nil {
			return err
		}
		if fromAccount == nil {
			return apperror.ErrAccountNotFound
		}

		toAccount, err := s.accountRepo.GetByIDAndUserID(ctx, input.ToAccountID, input.UserID)
		if err != nil {
			return err
		}
		if toAccount == nil {
			return apperror.ErrAccountNotFound
		}

		if fromAccount.Currency.Code == toAccount.Currency.Code && !input.FromAmount.Equal(input.ToAmount) {
			return ErrAmountMismatchForSameCurrency
		}

		if !fromAccount.IsNegativeBalanceEnabled {
			if err := s.txManager.Lock(ctx, model.AccountBalanceLockKey(fromAccount.ID)); err != nil {
				return err
			}

			balance, err := s.transactionRepo.GetBalanceByAccountID(ctx, fromAccount.ID)
			if err != nil {
				return err
			}
			if balance.Sub(input.FromAmount).IsNegative() {
				return apperror.ErrNegativeBalanceNotAllowed
			}
		}

		transfer := model.NewTransfer(
			input.FromAccountID,
			input.ToAccountID,
			input.FromAmount,
			input.ToAmount,
			input.Date,
		)

		transfer, err = s.transferRepo.Create(ctx, transfer)
		if err != nil {
			return err
		}

		result.Transfer = transfer

		transactionName := buildTransactionName(fromAccount, toAccount)

		transactions := []model.Transaction{
			*model.NewTransferOutTransaction(
				transfer.FromAccountID,
				transfer.FromAmount,
				transactionName,
				transfer.Date,
				true,
				transfer.ID,
			),
			*model.NewTransferInTransaction(
				transfer.ToAccountID,
				transfer.ToAmount,
				transactionName,
				transfer.Date,
				true,
				transfer.ID,
			),
		}

		transactions, err = s.transactionRepo.BulkCreate(ctx, transactions)
		if err != nil {
			return err
		}
		result.Transactions = transactions

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (i *CreateTransferInput) Validate() error {
	v := apperror.NewValidationError()

	v.AddIf(i.FromAccountID == uuid.Nil, "fromAccountId", "is required")
	v.AddIf(i.ToAccountID == uuid.Nil, "toAccountId", "is required")
	validation.Amount(v, "fromAmount", i.FromAmount)
	validation.Amount(v, "toAmount", i.ToAmount)

	return v.ErrorOrNil()
}

func buildTransactionName(fromAccout *model.Account, toAccount *model.Account) string {
	if fromAccout == nil || toAccount == nil {
		return ""
	}
	return fmt.Sprintf("Transfer from %s to %s", fromAccout.Name, toAccount.Name)
}
