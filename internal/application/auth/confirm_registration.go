package auth

import (
	"context"
	apperror "finance-app-go-demo/internal/application/error"
	"finance-app-go-demo/internal/domain/model"
	"finance-app-go-demo/internal/domain/valueObject"
)

type ConfirmRegistrationInput struct {
	Token string
}

type ConfirmRegistrationResult struct {
	Message string
}

func (s *Service) ConfirmRegistration(ctx context.Context, input ConfirmRegistrationInput) (*ConfirmRegistrationResult, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	registration, err := s.registrationStorage.Get(ctx, input.Token)
	if err != nil {
		return nil, err
	}
	if registration == nil {
		return nil, ErrInvalidRegistrationToken
	}

	userExists, err := s.userRepo.ExistsByEmail(ctx, registration.Email)
	if err != nil {
		return nil, err
	}
	if userExists {
		_ = s.registrationStorage.Delete(ctx, input.Token)

		return nil, ErrUserExists
	}

	email, err := valueObject.NewEmail(registration.Email)
	if err != nil {
		return nil, apperror.Invalid(err.Error())
	}

	user := model.NewUser(
		registration.FirstName,
		registration.LastName,
		email,
		registration.PasswordHash,
		nil,
	)

	err = s.txManager.Wrap(ctx, func(ctx context.Context) error {
		user, err = s.userRepo.Create(ctx, user)
		if err != nil {
			return err
		}

		_, err = s.transactionCategoryRepo.BulkCreate(ctx, getDefaultCategories(user.ID))

		return err
	})
	if err != nil {
		return nil, err
	}

	_ = s.registrationStorage.Delete(ctx, input.Token)

	return &ConfirmRegistrationResult{
		Message: "registration confirmed, you can log in now",
	}, nil
}

func (i *ConfirmRegistrationInput) Validate() error {
	v := apperror.NewValidationError()
	v.AddIf(i.Token == "", "token", "is required")
	return v.ErrorOrNil()
}
