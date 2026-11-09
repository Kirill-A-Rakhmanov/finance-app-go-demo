package auth

import "context"

type RegistrationConfirmation struct {
	Email     string
	FirstName string
	LastName  string
	Token     string
}

type RegistrationNotifier interface {
	SendRegistrationConfirmation(ctx context.Context, confirmation RegistrationConfirmation) error
	SendRegistrationAttempt(ctx context.Context, email string) error
}
