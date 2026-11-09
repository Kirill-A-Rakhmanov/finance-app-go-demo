package auth

import (
	"context"
	apperror "finance-app-go-demo/internal/application/error"
)

type RefreshInput struct {
	RefreshToken string
}

type RefreshResult struct {
	TokenPair TokenPair
}

func (s *Service) RefreshTokens(ctx context.Context, input RefreshInput) (*RefreshResult, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	token, err := s.tokenService.VerifyToken(input.RefreshToken, RefreshTokenType)
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

	session, err = s.sessionStorage.CreateSession(ctx, token.UserID)
	if err != nil {
		return nil, err
	}

	tokenData := NewTokenData(token.UserID, session.ID)

	tokenPair, err := s.tokenService.IssueTokens(tokenData)
	if err != nil {
		return nil, err
	}

	if err := s.sessionStorage.ShortenSession(ctx, token.SessionID, RefreshGracePeriod); err != nil {
		return nil, err
	}

	return &RefreshResult{
		TokenPair: *tokenPair,
	}, nil
}

func (i *RefreshInput) Validate() error {
	v := apperror.NewValidationError()
	v.AddIf(i.RefreshToken == "", "refreshToken", "is required")
	return v.ErrorOrNil()
}
