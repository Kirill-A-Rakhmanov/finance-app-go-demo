package auth

import (
	"context"
	apperror "finance-app-go-demo/internal/application/error"
	"finance-app-go-demo/internal/application/validation"
)

type ResetPasswordInput struct {
	Code        string
	NewPassword string
}

type ResetPasswordResult struct {
	Message string
}

func (s *Service) ResetPassword(ctx context.Context, input ResetPasswordInput) (*ResetPasswordResult, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	recovery, err := s.recoveryStorage.GetRecovery(ctx, input.Code)
	if err != nil {
		return nil, err
	}
	if recovery == nil {
		return nil, ErrInvalidRecoveryCode
	}

	user, err := s.userRepo.GetByID(ctx, recovery.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidRecoveryCode
	}

	hashedPassword, err := s.passwordService.Hash(input.NewPassword)
	if err != nil {
		return nil, err
	}

	user.Password = hashedPassword

	if _, err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	if err := s.sessionStorage.DeleteSessionsByUserID(ctx, user.ID, ""); err != nil {
		return nil, err
	}

	if err := s.failureLimiter.Reset(ctx, LoginFailuresKey(user.Email.ToString())); err != nil {
		return nil, err
	}

	if err := s.recoveryStorage.DeleteRecovery(ctx, input.Code); err != nil {
		return nil, err
	}

	return &ResetPasswordResult{
		Message: "password reset successfully",
	}, nil
}

func (i *ResetPasswordInput) Validate() error {
	v := apperror.NewValidationError()
	v.AddIf(i.Code == "", "code", "is required")
	validation.Password(v, "newPassword", i.NewPassword)
	return v.ErrorOrNil()
}
