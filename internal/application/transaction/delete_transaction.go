package transaction

import (
	"context"
	apperror "finance-app-go-demo/internal/application/error"
	"finance-app-go-demo/internal/domain/model"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type DeleteTransactionInput struct {
	TransactionID uuid.UUID
	UserID        uuid.UUID
}

type DeleteTransactionResult struct {
	Message string
	Balance decimal.Decimal
}

func (s *Service) DeleteTransaction(ctx context.Context, input DeleteTransactionInput) (*DeleteTransactionResult, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	var result *DeleteTransactionResult

	err := s.txManager.Wrap(ctx, func(ctx context.Context) error {
		transaction, err := s.transactionRepo.GetByIDAndUserID(ctx, input.TransactionID, input.UserID)
		if err != nil {
			return err
		}
		if transaction == nil {
			return ErrTransactionNotFound
		}
		if transaction.TransferID != nil {
			return ErrDeleteTransferNotAllowed
		}

		account, err := s.accountRepo.GetByIDAndUserID(ctx, transaction.AccountID, input.UserID)
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
		}

		currentBalance, err := s.transactionRepo.GetBalanceByAccountID(ctx, transaction.AccountID)
		if err != nil {
			return err
		}

		balanceAfterDeletion := currentBalance.Sub(transaction.BalanceEffect())

		if !account.IsNegativeBalanceEnabled && balanceAfterDeletion.IsNegative() {
			return apperror.ErrNegativeBalanceNotAllowed
		}

		if err := s.transactionRepo.DeleteByID(ctx, transaction.ID); err != nil {
			return err
		}

		// todo: add ws emit balance

		result = &DeleteTransactionResult{
			Message: "Transaction deleted successfully",
			Balance: balanceAfterDeletion,
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (i *DeleteTransactionInput) Validate() error {
	v := apperror.NewValidationError()
	v.AddIf(i.TransactionID == uuid.Nil, "transactionId", "is required")
	v.AddIf(i.UserID == uuid.Nil, "userId", "is required")
	return v.ErrorOrNil()
}
