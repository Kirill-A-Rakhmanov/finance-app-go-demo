package transaction

import (
	"context"
	apperror "finance-app-go-demo/internal/application/error"
	"finance-app-go-demo/internal/domain/model"

	"github.com/google/uuid"
)

type GetTrainsactionByIdInput struct {
	UserID        uuid.UUID
	TransactionID uuid.UUID
}

type GetTransactionByIdResult struct {
	Transaction *model.Transaction
}

func (s *Service) GetTransactionById(ctx context.Context, input GetTrainsactionByIdInput) (*GetTransactionByIdResult, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	transaction, err := s.transactionRepo.GetByIDAndUserID(ctx, input.TransactionID, input.UserID)
	if err != nil {
		return nil, err
	}
	if transaction == nil {
		return nil, ErrTransactionNotFound
	}

	return &GetTransactionByIdResult{
		Transaction: transaction,
	}, nil
}

func (i *GetTrainsactionByIdInput) Validate() error {
	v := apperror.NewValidationError()
	v.AddIf(i.TransactionID == uuid.Nil, "transactionId", "is required")
	v.AddIf(i.UserID == uuid.Nil, "userId", "is required")
	return v.ErrorOrNil()
}
