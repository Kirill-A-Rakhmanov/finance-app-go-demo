package auth

import "context"

type RecoveryNotifier interface {
	SendPasswordRecovery(ctx context.Context, email string, code string) error
}
