package auth

import (
	"context"
	apperror "finance-app-go-demo/internal/application/error"
	"finance-app-go-demo/internal/domain/model"
	"finance-app-go-demo/internal/domain/valueObject"
)

type LoginInput struct {
	Email    string
	Password string
}

type LoginResult struct {
	User         *model.User
	AccessToken  string
	RefreshToken string
}

func (s *Service) Login(ctx context.Context, input LoginInput) (*LoginResult, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	email := valueObject.NormalizeEmail(input.Email)
	failuresKey := LoginFailuresKey(email)

	if err := EnsureNotLocked(ctx, s.failureLimiter, failuresKey); err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	hashed := ""
	if user != nil {
		hashed = user.Password
	}

	if !s.passwordService.Verify(input.Password, hashed) || user == nil {
		if err := s.failureLimiter.RecordFailure(ctx, failuresKey, FailedAttemptsWindow); err != nil {
			return nil, err
		}

		return nil, ErrInvalidCredentials
	}

	if err := s.failureLimiter.Reset(ctx, failuresKey); err != nil {
		return nil, err
	}

	session, err := s.sessionStorage.CreateSession(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	tokenData := NewTokenData(user.ID, session.ID)

	tokenPair, err := s.tokenService.IssueTokens(tokenData)
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		User:         user,
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}, nil
}

func (i *LoginInput) Validate() error {
	v := apperror.NewValidationError()
	v.AddIf(i.Email == "", "email", "is required")
	v.AddIf(i.Password == "", "password", "is required")
	return v.ErrorOrNil()
}
