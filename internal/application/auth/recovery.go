package auth

import (
	"context"
	apperror "finance-app-go-demo/internal/application/error"
	"finance-app-go-demo/internal/domain/valueObject"
)

const recoveryMessage = "if an account with this email exists, a recovery code has been sent"

type RecoveryInput struct {
	Email string
}

type RecoveryResult struct {
	Message string
}

func (s *Service) Recovery(ctx context.Context, input RecoveryInput) (*RecoveryResult, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	email := valueObject.NormalizeEmail(input.Email)

	allowed, err := s.recoveryStorage.AcquireCooldown(ctx, email, RecoveryCooldown)
	if err != nil {
		return nil, err
	}

	if allowed {
		if err := s.sendRecovery(ctx, email); err != nil {
			return nil, err
		}
	}

	return &RecoveryResult{
		Message: recoveryMessage,
	}, nil
}

func (s *Service) sendRecovery(ctx context.Context, email string) error {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return err
	}
	if user == nil {
		return nil
	}

	recovery, err := s.recoveryStorage.CreateRecovery(ctx, user.ID)
	if err != nil {
		return err
	}

	return s.recoveryNotifier.SendPasswordRecovery(ctx, user.Email.ToString(), recovery.Code)
}

func (i *RecoveryInput) Validate() error {
	v := apperror.NewValidationError()
	v.AddIf(i.Email == "", "email", "is required")
	return v.ErrorOrNil()
}
