package auth

import (
	"time"

	"github.com/google/uuid"
)

type TokenType string

const (
	AccessTokenType  TokenType = "access"
	RefreshTokenType TokenType = "refresh"
)

type TokenService interface {
	GenerateToken(data *TokenData, tokenType TokenType, ttl time.Duration) (string, error)
	IssueTokens(data *TokenData) (*TokenPair, error)
	VerifyToken(token string, tokenType TokenType) (*TokenData, error)
}

type TokenData struct {
	UserID    uuid.UUID
	SessionID string
}

func NewTokenData(userID uuid.UUID, sessionID string) *TokenData {
	return &TokenData{
		UserID:    userID,
		SessionID: sessionID,
	}
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}
