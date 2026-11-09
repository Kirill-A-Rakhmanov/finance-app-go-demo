package notifier

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"finance-app-go-demo/pkg/logger"
	"finance-app-go-demo/pkg/mailer"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeSender struct {
	sent chan mailer.Message
	err  error
	// release, when set, blocks Send until it is closed.
	release chan struct{}
}

func (f *fakeSender) Send(_ context.Context, msg mailer.Message) error {
	if f.release != nil {
		<-f.release
	}
	f.sent <- msg
	return f.err
}

func discardLogger() logger.Logger {
	return logger.New(slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func TestSMTPRecoveryNotifier_SendsLinkAndTTL(t *testing.T) {
	sender := &fakeSender{sent: make(chan mailer.Message, 1)}
	n := NewSMTPRecoveryNotifier(sender, discardLogger(), "https://example.com/reset-password", 15*time.Minute)

	require.NoError(t, n.SendPasswordRecovery(context.Background(), "john@example.com", "abc123"))

	select {
	case msg := <-sender.sent:
		assert.Equal(t, "john@example.com", msg.To)
		assert.Equal(t, "Reset your password", msg.Subject)
		assert.Contains(t, msg.Body, "https://example.com/reset-password?code=abc123")
		assert.Contains(t, msg.Body, "valid for 15 minutes")
	case <-time.After(5 * time.Second):
		t.Fatal("no email was sent")
	}
}

func TestSMTPRecoveryNotifier_DoesNotWaitForDelivery(t *testing.T) {
	sender := &fakeSender{sent: make(chan mailer.Message, 1), release: make(chan struct{})}
	n := NewSMTPRecoveryNotifier(sender, discardLogger(), "https://example.com/reset-password", time.Hour)

	done := make(chan error, 1)
	go func() { done <- n.SendPasswordRecovery(context.Background(), "john@example.com", "abc123") }()

	select {
	case err := <-done:
		require.NoError(t, err, "the call returns while the email is still being delivered")
	case <-time.After(2 * time.Second):
		t.Fatal("SendPasswordRecovery blocked on delivery")
	}

	close(sender.release)
	<-sender.sent
}

func TestSMTPRecoveryNotifier_DeliveryFailureIsNotReturned(t *testing.T) {
	sender := &fakeSender{sent: make(chan mailer.Message, 1), err: errors.New("smtp down")}
	n := NewSMTPRecoveryNotifier(sender, discardLogger(), "https://example.com/reset-password", time.Hour)

	assert.NoError(t, n.SendPasswordRecovery(context.Background(), "john@example.com", "abc123"),
		"the caller must not learn whether the account exists or whether delivery works")

	<-sender.sent
}

func TestHumanDuration(t *testing.T) {
	assert.Equal(t, "1 minute", humanDuration(time.Minute))
	assert.Equal(t, "15 minutes", humanDuration(15*time.Minute))
	assert.Equal(t, "1 hour", humanDuration(time.Hour))
	assert.Equal(t, "24 hours", humanDuration(24*time.Hour))
	assert.Equal(t, "90 minutes", humanDuration(90*time.Minute))
}
