package transfer

import (
	"context"
	apperror "finance-app-go-demo/internal/application/error"
	"finance-app-go-demo/internal/domain/model"

	"github.com/google/uuid"
)

type DeleteTransferInput struct {
	UserID     uuid.UUID
	TransferID uuid.UUID
}

type DeleteTransferResult struct {
	Message string
}

func (s *Service) DeleteTransfer(ctx context.Context, input DeleteTransferInput) (*DeleteTransferResult, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	err := s.txManager.Wrap(ctx, func(ctx context.Context) error {
		transfer, err := s.transferRepo.GetByID(ctx, input.TransferID)
		if err != nil {
			return err
		}
		if transfer == nil {
			return ErrTransferNotFound
		}

		toAccount, err := s.accountRepo.GetByIDAndUserID(ctx, transfer.ToAccountID, input.UserID)
		if err != nil {
			return err
		}
		if toAccount == nil {
			return apperror.ErrAccountNotFound
		}

		if !toAccount.IsNegativeBalanceEnabled {
			if err := s.txManager.Lock(ctx, model.AccountBalanceLockKey(toAccount.ID)); err != nil {
				return err
			}

			balance, err := s.transactionRepo.GetBalanceByAccountID(ctx, toAccount.ID)
			if err != nil {
				return err
			}
			if balance.Sub(transfer.ToAmount).IsNegative() {
				return apperror.ErrNegativeBalanceNotAllowed
			}
		}

		return s.transferRepo.DeleteByID(ctx, input.TransferID)
	})
	if err != nil {
		return nil, err
	}

	return &DeleteTransferResult{
		Message: "Transfer deleted successfully",
	}, nil
}

func (i *DeleteTransferInput) Validate() error {
	v := apperror.NewValidationError()

	v.AddIf(i.UserID == uuid.Nil, "userID", "is required")
	v.AddIf(i.TransferID == uuid.Nil, "transferID", "is required")

	return v.ErrorOrNil()
}
