package counterparty

import (
	"context"

	apperror "finance-app-go-demo/internal/application/error"
	"finance-app-go-demo/internal/domain/model"

	"github.com/google/uuid"
)

type GetCounterpartiesByUserIDInput struct {
	UserID uuid.UUID
}

type GetCounterpartiesByUserIDResult struct {
	Counterparties []model.Counterparty
}

func (s *Service) GetCounterpartiesByUserID(
	ctx context.Context,
	input GetCounterpartiesByUserIDInput,
) (*GetCounterpartiesByUserIDResult, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	counterparties, err := s.counterpartyRepo.GetByUserID(ctx, input.UserID)
	if err != nil {
		return nil, err
	}

	return &GetCounterpartiesByUserIDResult{
		Counterparties: counterparties,
	}, nil
}

func (i *GetCounterpartiesByUserIDInput) Validate() error {
	v := apperror.NewValidationError()
	v.AddIf(i.UserID == uuid.Nil, "userId", "is required")
	return v.ErrorOrNil()
}
