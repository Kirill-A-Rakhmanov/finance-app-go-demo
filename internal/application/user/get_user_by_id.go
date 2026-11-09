package user

import (
	"context"
	apperror "finance-app-go-demo/internal/application/error"
	"finance-app-go-demo/internal/domain/model"

	"github.com/google/uuid"
)

type GetUserByIdInput struct {
	UserID uuid.UUID
}

type GetUserByIdResult struct {
	User *model.User
}

func (s *Service) GetUserById(ctx context.Context, input GetUserByIdInput) (*GetUserByIdResult, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByID(ctx, input.UserID)

	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, apperror.ErrUserNotFound
	}

	return &GetUserByIdResult{
		User: user,
	}, nil
}

func (i *GetUserByIdInput) Validate() error {
	v := apperror.NewValidationError()
	v.AddIf(i.UserID == uuid.Nil, "userId", "is required")
	return v.ErrorOrNil()
}
