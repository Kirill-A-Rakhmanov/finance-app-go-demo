package counterparty

import (
	"context"

	apperror "finance-app-go-demo/internal/application/error"

	"github.com/google/uuid"
)

type DeleteCounterpartyInput struct {
	UserID         uuid.UUID
	CounterpartyID uuid.UUID
}

type DeleteCounterpartyResult struct {
	Message string
}

func (s *Service) DeleteCounterparty(ctx context.Context, input DeleteCounterpartyInput) (*DeleteCounterpartyResult, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	counterparty, err := s.counterpartyRepo.GetByIDAndUserID(ctx, input.CounterpartyID, input.UserID)
	if err != nil {
		return nil, err
	}
	if counterparty == nil {
		return nil, apperror.ErrCounterpartyNotFound
	}

	err = s.counterpartyRepo.DeleteByID(ctx, counterparty.ID)
	if err != nil {
		return nil, err
	}

	return &DeleteCounterpartyResult{
		Message: "Counterparty deleted successfully",
	}, nil
}

func (i *DeleteCounterpartyInput) Validate() error {
	v := apperror.NewValidationError()
	v.AddIf(i.CounterpartyID == uuid.Nil, "counterpartyId", "is required")
	v.AddIf(i.UserID == uuid.Nil, "userId", "is required")
	return v.ErrorOrNil()
}
