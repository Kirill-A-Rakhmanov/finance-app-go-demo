package notifier

import (
	"context"
	"fmt"
	"time"

	"finance-app-go-demo/internal/application/auth"
	"finance-app-go-demo/pkg/logger"
	"finance-app-go-demo/pkg/mailer"
)

type smtpRecoveryNotifier struct {
	sender           Sender
	log              logger.Logger
	resetPasswordURL string
	codeTTL          time.Duration
}

func NewSMTPRecoveryNotifier(sender Sender, log logger.Logger, resetPasswordURL string, codeTTL time.Duration) auth.RecoveryNotifier {
	return &smtpRecoveryNotifier{
		sender:           sender,
		log:              log,
		resetPasswordURL: resetPasswordURL,
		codeTTL:          codeTTL,
	}
}

func (n *smtpRecoveryNotifier) SendPasswordRecovery(ctx context.Context, email string, code string) error {
	msg := mailer.Message{
		To:      email,
		Subject: "Reset your password",
		Body:    n.body(code),
	}

	deliverAsync(ctx, n.sender, n.log, msg, "password recovery")

	return nil
}

func (n *smtpRecoveryNotifier) body(code string) string {
	return fmt.Sprintf(`Hi,

We received a request to reset the password of your Finance App account.
Open the link below to choose a new password:

%s

The link is valid for %s. If you did not ask for this, ignore this email:
your password stays the same.
`, buildLink(n.resetPasswordURL, "code", code), humanDuration(n.codeTTL))
}
