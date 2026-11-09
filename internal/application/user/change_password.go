package user

import (
	"context"
	"finance-app-go-demo/internal/application/auth"
	apperror "finance-app-go-demo/internal/application/error"
	"finance-app-go-demo/internal/application/validation"

	"github.com/google/uuid"
)

type ChangePasswordInput struct {
	UserID      uuid.UUID
	SessionID   string
	OldPassword string
	NewPassword string
}

type ChangePasswordResult struct {
	Message string
}

func (s *Service) ChangePassword(ctx context.Context, input ChangePasswordInput) (*ChangePasswordResult, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	failuresKey := auth.ChangePasswordFailuresKey(input.UserID)

	if err := auth.EnsureNotLocked(ctx, s.failureLimiter, failuresKey); err != nil {
		return nil, err
	}

	u, err := s.userRepo.GetByID(ctx, input.UserID)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, apperror.ErrUserNotFound
	}

	if !s.passwordService.Verify(input.OldPassword, u.Password) {
		if err := s.failureLimiter.RecordFailure(ctx, failuresKey, auth.FailedAttemptsWindow); err != nil {
			return nil, err
		}

		return nil, ErrInvalidCurrentPassword
	}

	if err := s.failureLimiter.Reset(ctx, failuresKey); err != nil {
		return nil, err
	}

	hashedPassword, err := s.passwordService.Hash(input.NewPassword)
	if err != nil {
		return nil, err
	}

	u.Password = hashedPassword

	if _, err := s.userRepo.Update(ctx, u); err != nil {
		return nil, err
	}

	if err := s.sessionRevoker.DeleteSessionsByUserID(ctx, u.ID, input.SessionID); err != nil {
		return nil, err
	}

	return &ChangePasswordResult{
		Message: "password changed successfully",
	}, nil
}

func (i *ChangePasswordInput) Validate() error {
	v := apperror.NewValidationError()
	v.AddIf(i.UserID == uuid.Nil, "userId", "is required")
	v.AddIf(i.SessionID == "", "sessionId", "is required")
	v.AddIf(i.OldPassword == "", "oldPassword", "is required")
	validation.Password(v, "newPassword", i.NewPassword)
	return v.ErrorOrNil()
}
