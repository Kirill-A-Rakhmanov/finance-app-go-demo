package counterparty

import (
	"context"
	"strings"

	apperror "finance-app-go-demo/internal/application/error"
	"finance-app-go-demo/internal/domain/model"

	"github.com/google/uuid"
)

type UpdateCounterpartyInput struct {
	CounterpartyID uuid.UUID
	UserID         uuid.UUID
	Name           string
}

type UpdateCounterpartyResult struct {
	Counterparty *model.Counterparty
}

func (s *Service) UpdateCounterparty(ctx context.Context, input UpdateCounterpartyInput) (*UpdateCounterpartyResult, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	name := strings.TrimSpace(input.Name)

	counterparty, err := s.counterpartyRepo.GetByIDAndUserID(ctx, input.CounterpartyID, input.UserID)
	if err != nil {
		return nil, err
	}
	if counterparty == nil {
		return nil, apperror.ErrCounterpartyNotFound
	}

	if counterparty.Name != name {
		exists, err := s.counterpartyRepo.ExistsByUserIDAndName(ctx, input.UserID, name)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrCounterpartyAlreadyExists
		}
	}

	counterparty.Name = name

	counterparty, err = s.counterpartyRepo.Update(ctx, counterparty)
	if err != nil {
		return nil, err
	}

	return &UpdateCounterpartyResult{
		Counterparty: counterparty,
	}, nil
}

func (i *UpdateCounterpartyInput) Validate() error {
	v := apperror.NewValidationError()
	v.AddIf(i.CounterpartyID == uuid.Nil, "counterpartyId", "is required")
	v.AddIf(i.UserID == uuid.Nil, "userId", "is required")
	v.AddIf(strings.TrimSpace(i.Name) == "", "name", "is required")
	v.MaxLength("name", strings.TrimSpace(i.Name), model.CounterpartyNameMaxLen)
	return v.ErrorOrNil()
}
