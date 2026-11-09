package auth

import (
	"context"
	apperror "finance-app-go-demo/internal/application/error"

	"github.com/google/uuid"
)

type LogoutInput struct {
	UserID    uuid.UUID
	SessionID string
}

func (s *Service) Logout(ctx context.Context, input LogoutInput) error {
	if err := input.Validate(); err != nil {
		return err
	}

	err := s.sessionStorage.DeleteSession(ctx, input.SessionID)
	if err != nil {
		return err
	}

	return nil
}

func (i *LogoutInput) Validate() error {
	v := apperror.NewValidationError()
	v.AddIf(i.UserID == uuid.Nil, "userId", "is required")
	v.AddIf(i.SessionID == "", "sessionId", "is required")
	return v.ErrorOrNil()
}
