package notifier

import (
	"context"
	"finance-app-go-demo/internal/application/auth"
	"finance-app-go-demo/pkg/logger"
	"net/url"
)

type logRecoveryNotifier struct {
	log              logger.Logger
	resetPasswordURL string
}

func NewLogRecoveryNotifier(log logger.Logger, resetPasswordURL string) auth.RecoveryNotifier {
	return &logRecoveryNotifier{log: log, resetPasswordURL: resetPasswordURL}
}

func (n *logRecoveryNotifier) SendPasswordRecovery(_ context.Context, email string, code string) error {
	link := buildLink(n.resetPasswordURL, "code", code)

	n.log.Info("password recovery requested", "email", email, "code", code, "link", link)

	return nil
}

func buildLink(pageURL, param, value string) string {
	u, err := url.Parse(pageURL)
	if err != nil {
		return pageURL
	}

	q := u.Query()
	q.Set(param, value)
	u.RawQuery = q.Encode()

	return u.String()
}
