package counterparty

import (
	"context"
	"strings"

	apperror "finance-app-go-demo/internal/application/error"
	"finance-app-go-demo/internal/domain/model"

	"github.com/google/uuid"
)

type CreateCounterpartyInput struct {
	UserID uuid.UUID
	Name   string
}

type CreateCounterpartyResult struct {
	Counterparty *model.Counterparty
}

func (s *Service) CreateCounterparty(ctx context.Context, input CreateCounterpartyInput) (*CreateCounterpartyResult, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	name := strings.TrimSpace(input.Name)

	exists, err := s.counterpartyRepo.ExistsByUserIDAndName(ctx, input.UserID, name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrCounterpartyAlreadyExists
	}

	counterparty, err := s.counterpartyRepo.Create(ctx, model.NewCounterparty(input.UserID, name))
	if err != nil {
		return nil, err
	}

	return &CreateCounterpartyResult{
		Counterparty: counterparty,
	}, nil
}

func (i *CreateCounterpartyInput) Validate() error {
	v := apperror.NewValidationError()
	v.AddIf(i.UserID == uuid.Nil, "userId", "is required")
	v.AddIf(strings.TrimSpace(i.Name) == "", "name", "is required")
	v.MaxLength("name", strings.TrimSpace(i.Name), model.CounterpartyNameMaxLen)
	return v.ErrorOrNil()
}
