package notifier

import (
	"context"
	"errors"
	"testing"
	"time"

	"finance-app-go-demo/internal/application/auth"
	"finance-app-go-demo/pkg/mailer"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSMTPRegistrationNotifier_ConfirmationEmail(t *testing.T) {
	sender := &fakeSender{sent: make(chan mailer.Message, 1)}
	n := NewSMTPRegistrationNotifier(sender, discardLogger(), "https://example.com/verify-email", 24*time.Hour)

	err := n.SendRegistrationConfirmation(context.Background(), auth.RegistrationConfirmation{
		Email: "jane@example.com", FirstName: "Jane", LastName: "Doe", Token: "tok-123",
	})
	require.NoError(t, err)

	select {
	case msg := <-sender.sent:
		assert.Equal(t, "jane@example.com", msg.To)
		assert.Equal(t, "Confirm your registration", msg.Subject)
		assert.Contains(t, msg.Body, "https://example.com/verify-email?token=tok-123")
		assert.Contains(t, msg.Body, "Jane Doe", "the registered name lets the owner spot a registration made by someone else")
		assert.Contains(t, msg.Body, "valid for 24 hours")
	case <-time.After(5 * time.Second):
		t.Fatal("no email was sent")
	}
}

func TestSMTPRegistrationNotifier_AttemptEmail(t *testing.T) {
	sender := &fakeSender{sent: make(chan mailer.Message, 1)}
	n := NewSMTPRegistrationNotifier(sender, discardLogger(), "https://example.com/verify-email", 24*time.Hour)

	require.NoError(t, n.SendRegistrationAttempt(context.Background(), "jane@example.com"))

	select {
	case msg := <-sender.sent:
		assert.Equal(t, "jane@example.com", msg.To)
		assert.Equal(t, "Finance App registration attempt", msg.Subject)
		assert.NotContains(t, msg.Body, "token=", "the email carries no link that could create or change anything")
	case <-time.After(5 * time.Second):
		t.Fatal("no email was sent")
	}
}

func TestSMTPRegistrationNotifier_DoesNotWaitForDeliveryAndHidesFailures(t *testing.T) {
	sender := &fakeSender{sent: make(chan mailer.Message, 2), release: make(chan struct{}), err: errors.New("smtp down")}
	n := NewSMTPRegistrationNotifier(sender, discardLogger(), "https://example.com/verify-email", time.Hour)

	done := make(chan error, 2)
	go func() {
		done <- n.SendRegistrationConfirmation(context.Background(), auth.RegistrationConfirmation{Email: "jane@example.com", Token: "t"})
		done <- n.SendRegistrationAttempt(context.Background(), "jane@example.com")
	}()

	for i := 0; i < 2; i++ {
		select {
		case err := <-done:
			require.NoError(t, err, "the call returns while delivery is still blocked, and delivery errors are not reported")
		case <-time.After(2 * time.Second):
			t.Fatal("the notifier blocked on delivery")
		}
	}

	close(sender.release)
	<-sender.sent
	<-sender.sent
}

func TestLogRegistrationNotifier_DoesNotFail(t *testing.T) {
	n := NewLogRegistrationNotifier(discardLogger(), "https://example.com/verify-email")

	assert.NoError(t, n.SendRegistrationConfirmation(context.Background(), auth.RegistrationConfirmation{Email: "jane@example.com", Token: "t"}))
	assert.NoError(t, n.SendRegistrationAttempt(context.Background(), "jane@example.com"))
}

func TestBuildLink(t *testing.T) {
	assert.Equal(t, "https://example.com/verify-email?token=abc", buildLink("https://example.com/verify-email", "token", "abc"))
	assert.Equal(t, "https://example.com/p?lang=en&token=a%2Bb", buildLink("https://example.com/p?lang=en", "token", "a+b"))
}
