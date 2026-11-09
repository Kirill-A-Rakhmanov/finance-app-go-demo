package token

import (
	"finance-app-go-demo/internal/application/auth"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type tokenService struct {
	secret          []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewTokenService(secret string, accessTokenTTL, refreshTokenTTL time.Duration) auth.TokenService {
	return &tokenService{
		secret:          []byte(secret),
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
	}
}

func (s *tokenService) GenerateToken(data *auth.TokenData, tokenType auth.TokenType, ttl time.Duration) (string, error) {
	claims := jwt.MapClaims{
		UserIDClaim:    data.UserID.String(),
		SessionIDClaim: data.SessionID,
		TokenTypeClaim: string(tokenType),
		"exp":          time.Now().Add(ttl).Unix(),
		"iat":          time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(s.secret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *tokenService) IssueTokens(data *auth.TokenData) (*auth.TokenPair, error) {
	accessToken, err := s.GenerateToken(data, auth.AccessTokenType, s.accessTokenTTL)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.GenerateToken(data, auth.RefreshTokenType, s.refreshTokenTTL)
	if err != nil {
		return nil, err
	}

	return &auth.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *tokenService) VerifyToken(token string, tokenType auth.TokenType) (*auth.TokenData, error) {
	parsed, err := jwt.Parse(token, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok || !parsed.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	if claimedType, _ := claims[TokenTypeClaim].(string); claimedType != string(tokenType) {
		return nil, auth.ErrInvalidTokenType
	}

	userIDStr, ok := claims[UserIDClaim].(string)
	if !ok {
		return nil, fmt.Errorf("%s not found in token", UserIDClaim)
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid user_id format: %w", err)
	}

	sessionID, ok := claims[SessionIDClaim].(string)
	if !ok {
		return nil, fmt.Errorf("%s not found in token", SessionIDClaim)
	}

	return auth.NewTokenData(userID, sessionID), nil
}
