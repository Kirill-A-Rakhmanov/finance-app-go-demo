package user

import (
	"context"
	apperror "finance-app-go-demo/internal/application/error"
	"finance-app-go-demo/internal/domain/model"

	"github.com/google/uuid"
)

type UpdateUserInput struct {
	UserID    uuid.UUID
	FirstName string
	LastName  string
}

type UpdateUserResult struct {
	User *model.User
}

func (s *Service) UpdateUser(ctx context.Context, input UpdateUserInput) (*UpdateUserResult, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	u, err := s.userRepo.GetByID(ctx, input.UserID)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, apperror.ErrUserNotFound
	}

	u.FirstName = input.FirstName
	u.LastName = input.LastName

	updated, err := s.userRepo.Update(ctx, u)
	if err != nil {
		return nil, err
	}

	return &UpdateUserResult{User: updated}, nil
}

func (i *UpdateUserInput) Validate() error {
	v := apperror.NewValidationError()
	v.AddIf(i.UserID == uuid.Nil, "userId", "is required")
	v.AddIf(i.FirstName == "", "firstName", "is required")
	v.AddIf(i.LastName == "", "lastName", "is required")
	v.MaxLength("firstName", i.FirstName, model.UserFirstNameMaxLen)
	v.MaxLength("lastName", i.LastName, model.UserLastNameMaxLen)
	return v.ErrorOrNil()
}
