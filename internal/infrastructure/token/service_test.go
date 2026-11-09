package token_test

import (
	"testing"
	"time"

	"finance-app-go-demo/internal/application/auth"
	"finance-app-go-demo/internal/infrastructure/token"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSecret = "0123456789abcdef0123456789abcdef"

func newService() auth.TokenService {
	return token.NewTokenService(testSecret, 15*time.Minute, 14*24*time.Hour)
}

func newData() *auth.TokenData {
	return auth.NewTokenData(uuid.New(), "session-1")
}

func TestTokenService_IssuedTokensVerifyAsTheirOwnType(t *testing.T) {
	service := newService()
	data := newData()

	pair, err := service.IssueTokens(data)
	require.NoError(t, err)

	access, err := service.VerifyToken(pair.AccessToken, auth.AccessTokenType)
	require.NoError(t, err)
	assert.Equal(t, data, access)

	refresh, err := service.VerifyToken(pair.RefreshToken, auth.RefreshTokenType)
	require.NoError(t, err)
	assert.Equal(t, data, refresh)
}

func TestTokenService_TokensOfOneTypeAreRefusedAsTheOther(t *testing.T) {
	service := newService()

	pair, err := service.IssueTokens(newData())
	require.NoError(t, err)

	_, err = service.VerifyToken(pair.AccessToken, auth.RefreshTokenType)
	assert.ErrorIs(t, err, auth.ErrInvalidTokenType, "an access token must not work as a refresh token")

	_, err = service.VerifyToken(pair.RefreshToken, auth.AccessTokenType)
	assert.ErrorIs(t, err, auth.ErrInvalidTokenType, "a refresh token must not work as an access token")
}

func TestTokenService_TokenWithoutTypeIsRefused(t *testing.T) {
	claims := jwt.MapClaims{
		token.UserIDClaim:    uuid.NewString(),
		token.SessionIDClaim: "session-1",
		"exp":                time.Now().Add(time.Hour).Unix(),
	}
	legacy, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(testSecret))
	require.NoError(t, err)

	service := newService()

	_, err = service.VerifyToken(legacy, auth.AccessTokenType)
	assert.ErrorIs(t, err, auth.ErrInvalidTokenType)

	_, err = service.VerifyToken(legacy, auth.RefreshTokenType)
	assert.ErrorIs(t, err, auth.ErrInvalidTokenType)
}

func TestTokenService_RejectsInvalidTokens(t *testing.T) {
	service := newService()
	data := newData()

	valid, err := service.GenerateToken(data, auth.AccessTokenType, time.Hour)
	require.NoError(t, err)

	expired, err := service.GenerateToken(data, auth.AccessTokenType, -time.Minute)
	require.NoError(t, err)

	otherSecret, err := token.NewTokenService("ffffffffffffffffffffffffffffffff", time.Hour, time.Hour).
		GenerateToken(data, auth.AccessTokenType, time.Hour)
	require.NoError(t, err)

	unsigned, err := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{
		token.UserIDClaim:    data.UserID.String(),
		token.SessionIDClaim: data.SessionID,
		token.TokenTypeClaim: string(auth.AccessTokenType),
		"exp":                time.Now().Add(time.Hour).Unix(),
	}).SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	tests := map[string]string{
		"empty":             "",
		"garbage":           "not-a-token",
		"expired":           expired,
		"signed elsewhere":  otherSecret,
		"alg none":          unsigned,
		"tampered payload":  valid[:len(valid)-3] + "abc",
		"truncated":         valid[:len(valid)/2],
		"valid with suffix": valid + "x",
	}

	for name, raw := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := service.VerifyToken(raw, auth.AccessTokenType)
			assert.Error(t, err)
		})
	}

	_, err = service.VerifyToken(valid, auth.AccessTokenType)
	assert.NoError(t, err, "the unmodified token is still accepted")
}

func TestTokenService_RejectsMalformedClaims(t *testing.T) {
	service := newService()

	sign := func(claims jwt.MapClaims) string {
		claims["exp"] = time.Now().Add(time.Hour).Unix()
		claims[token.TokenTypeClaim] = string(auth.AccessTokenType)
		signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(testSecret))
		require.NoError(t, err)

		return signed
	}

	tests := map[string]jwt.MapClaims{
		"no user":             {token.SessionIDClaim: "s"},
		"user is not a uuid":  {token.UserIDClaim: "42", token.SessionIDClaim: "s"},
		"no session":          {token.UserIDClaim: uuid.NewString()},
		"session is not text": {token.UserIDClaim: uuid.NewString(), token.SessionIDClaim: 7},
	}

	for name, claims := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := service.VerifyToken(sign(claims), auth.AccessTokenType)
			assert.Error(t, err)
		})
	}
}
