package auth

import apperror "finance-app-go-demo/internal/application/error"

var (
	ErrRegistrationClosed       = apperror.Forbidden("registration is closed")
	ErrInvalidRegistrationToken = apperror.Invalid("invalid or expired confirmation link")
	ErrUserExists               = apperror.Conflict("user already exists")
	ErrInvalidCredentials       = apperror.Unauthorized("invalid credentials")
	ErrInvalidTokenType         = apperror.Unauthorized("invalid token type")
	ErrSessionNotFound          = apperror.Unauthorized("session not found")
	ErrInvalidSession           = apperror.Unauthorized("invalid session")
	ErrInvalidRecoveryCode      = apperror.Unauthorized("invalid or expired recovery code")
)
