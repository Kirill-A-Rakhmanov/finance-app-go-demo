package notifier

import (
	"context"
	"fmt"
	"time"

	"finance-app-go-demo/internal/application/auth"
	"finance-app-go-demo/pkg/logger"
	"finance-app-go-demo/pkg/mailer"
)

type logRegistrationNotifier struct {
	log       logger.Logger
	verifyURL string
}

func NewLogRegistrationNotifier(log logger.Logger, verifyURL string) auth.RegistrationNotifier {
	return &logRegistrationNotifier{log: log, verifyURL: verifyURL}
}

func (n *logRegistrationNotifier) SendRegistrationConfirmation(_ context.Context, c auth.RegistrationConfirmation) error {
	n.log.Info("registration confirmation requested", "email", c.Email, "link", buildLink(n.verifyURL, "token", c.Token))

	return nil
}

func (n *logRegistrationNotifier) SendRegistrationAttempt(_ context.Context, email string) error {
	n.log.Info("registration attempted for an existing account", "email", email)

	return nil
}

type smtpRegistrationNotifier struct {
	sender    Sender
	log       logger.Logger
	verifyURL string
	tokenTTL  time.Duration
}

func NewSMTPRegistrationNotifier(sender Sender, log logger.Logger, verifyURL string, tokenTTL time.Duration) auth.RegistrationNotifier {
	return &smtpRegistrationNotifier{
		sender:    sender,
		log:       log,
		verifyURL: verifyURL,
		tokenTTL:  tokenTTL,
	}
}

func (n *smtpRegistrationNotifier) SendRegistrationConfirmation(ctx context.Context, c auth.RegistrationConfirmation) error {
	body := fmt.Sprintf(`Hi %s,

Someone registered a Finance App account for this email address as %s %s.
To finish the registration, open the link below:

%s

The link is valid for %s. If this was not you, ignore this email:
no account is created until the link is opened.
`, c.FirstName, c.FirstName, c.LastName, buildLink(n.verifyURL, "token", c.Token), humanDuration(n.tokenTTL))

	deliverAsync(ctx, n.sender, n.log, mailer.Message{
		To:      c.Email,
		Subject: "Confirm your registration",
		Body:    body,
	}, "registration confirmation")

	return nil
}

func (n *smtpRegistrationNotifier) SendRegistrationAttempt(ctx context.Context, email string) error {
	body := `Hi,

Someone tried to register a Finance App account with this email address,
but an account with it already exists.

If it was you, just log in. Use "Forgot password?" on the login page if you do not
remember the password. If it was not you, ignore this email: nothing has changed.
`

	deliverAsync(ctx, n.sender, n.log, mailer.Message{
		To:      email,
		Subject: "Finance App registration attempt",
		Body:    body,
	}, "registration attempt")

	return nil
}
