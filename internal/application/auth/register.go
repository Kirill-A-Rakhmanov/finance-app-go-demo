package auth

import (
	"context"
	apperror "finance-app-go-demo/internal/application/error"
	"finance-app-go-demo/internal/application/validation"
	"finance-app-go-demo/internal/domain/model"
	"finance-app-go-demo/internal/domain/valueObject"
)

const registerMessage = "check your email to finish the registration"

type RegisterInput struct {
	FirstName string
	LastName  string
	Email     string
	Password  string
}

type RegisterResult struct {
	Message string
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (*RegisterResult, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	registrationOpen, err := s.featureFlagRepo.IsEnabled(ctx, model.FeatureFlagRegistration)
	if err != nil {
		return nil, err
	}
	if !registrationOpen {
		return nil, ErrRegistrationClosed
	}

	email, err := valueObject.NewEmail(input.Email)
	if err != nil {
		return nil, apperror.Invalid(err.Error())
	}

	passwordHash, err := s.passwordService.Hash(input.Password)
	if err != nil {
		return nil, err
	}

	allowed, err := s.registrationStorage.AcquireCooldown(ctx, email.ToString(), RegistrationCooldown)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return &RegisterResult{Message: registerMessage}, nil
	}

	userExists, err := s.userRepo.ExistsByEmail(ctx, email.ToString())
	if err != nil {
		return nil, err
	}
	if userExists {
		if err := s.registrationNotifier.SendRegistrationAttempt(ctx, email.ToString()); err != nil {
			return nil, err
		}

		return &RegisterResult{Message: registerMessage}, nil
	}

	token, err := s.registrationStorage.Create(ctx, PendingRegistration{
		FirstName:    input.FirstName,
		LastName:     input.LastName,
		Email:        email.ToString(),
		PasswordHash: passwordHash,
	})
	if err != nil {
		return nil, err
	}

	err = s.registrationNotifier.SendRegistrationConfirmation(ctx, RegistrationConfirmation{
		Email:     email.ToString(),
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Token:     token,
	})
	if err != nil {
		return nil, err
	}

	return &RegisterResult{Message: registerMessage}, nil
}

func (i *RegisterInput) Validate() error {
	v := apperror.NewValidationError()
	v.AddIf(i.FirstName == "", "firstName", "is required")
	v.AddIf(i.LastName == "", "lastName", "is required")
	v.AddIf(i.Email == "", "email", "is required")
	v.MaxLength("firstName", i.FirstName, model.UserFirstNameMaxLen)
	v.MaxLength("lastName", i.LastName, model.UserLastNameMaxLen)
	v.MaxLength("email", i.Email, model.UserEmailMaxLen)
	validation.Password(v, "password", i.Password)
	return v.ErrorOrNil()
}
