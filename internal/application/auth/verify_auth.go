package auth

import (
	"context"
	apperror "finance-app-go-demo/internal/application/error"

	"github.com/google/uuid"
)

type VerifyInput struct {
	AccessToken string
}

type VerifyResult struct {
	UserID    uuid.UUID
	SessionID string
}

func (s *Service) VerifyAuth(ctx context.Context, input VerifyInput) (*VerifyResult, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	token, err := s.tokenService.VerifyToken(input.AccessToken, AccessTokenType)
	if err != nil {
		return nil, err
	}

	session, err := s.sessionStorage.GetSession(ctx, token.SessionID)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, ErrSessionNotFound
	}
	if session.UserID != token.UserID {
		_ = s.sessionStorage.DeleteSession(ctx, token.SessionID)

		return nil, ErrInvalidSession
	}

	return &VerifyResult{
		UserID:    token.UserID,
		SessionID: token.SessionID,
	}, nil
}

func (i *VerifyInput) Validate() error {
	v := apperror.NewValidationError()
	v.AddIf(i.AccessToken == "", "accessToken", "is required")
	return v.ErrorOrNil()
}
