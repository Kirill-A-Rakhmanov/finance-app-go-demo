package mailer_test

import (
	"bufio"
	"context"
	"io"
	"mime"
	"mime/quotedprintable"
	"net"
	"net/mail"
	"strings"
	"testing"
	"time"

	"finance-app-go-demo/pkg/mailer"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type capturedMail struct {
	from string
	rcpt []string
	data string
}

// startFakeSMTP runs a minimal SMTP server without TLS or auth that records one message.
func startFakeSMTP(t *testing.T) (host, port string, received <-chan capturedMail) {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = ln.Close() })

	ch := make(chan capturedMail, 1)

	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		r := bufio.NewReader(conn)
		write := func(s string) { _, _ = io.WriteString(conn, s+"\r\n") }

		var got capturedMail
		write("220 fake ESMTP")

		for {
			line, err := r.ReadString('\n')
			if err != nil {
				return
			}
			line = strings.TrimRight(line, "\r\n")
			upper := strings.ToUpper(line)

			switch {
			case strings.HasPrefix(upper, "EHLO"), strings.HasPrefix(upper, "HELO"):
				write("250 fake")
			case strings.HasPrefix(upper, "MAIL FROM:"):
				got.from = line
				write("250 OK")
			case strings.HasPrefix(upper, "RCPT TO:"):
				got.rcpt = append(got.rcpt, line)
				write("250 OK")
			case upper == "DATA":
				write("354 go ahead")
				var sb strings.Builder
				for {
					l, err := r.ReadString('\n')
					if err != nil {
						return
					}
					if l == ".\r\n" {
						break
					}
					sb.WriteString(l)
				}
				got.data = sb.String()
				write("250 queued")
			case upper == "QUIT":
				write("221 bye")
				ch <- got
				return
			default:
				write("250 OK")
			}
		}
	}()

	host, port, err = net.SplitHostPort(ln.Addr().String())
	require.NoError(t, err)

	return host, port, ch
}

func TestMailer_Send(t *testing.T) {
	host, port, received := startFakeSMTP(t)

	m := mailer.New(mailer.Config{
		Host: host, Port: port, Security: mailer.SecurityNone,
		From: "noreply@example.com", FromName: "Finance App",
	})

	err := m.Send(context.Background(), mailer.Message{
		To:      "john@example.com",
		Subject: "Восстановление пароля",
		Body:    "Привет!\nСсылка: https://example.com/reset?code=abc",
	})
	require.NoError(t, err)

	var got capturedMail
	select {
	case got = <-received:
	case <-time.After(5 * time.Second):
		t.Fatal("the fake server received nothing")
	}

	assert.Contains(t, got.from, "<noreply@example.com>")
	require.Len(t, got.rcpt, 1)
	assert.Contains(t, got.rcpt[0], "<john@example.com>")

	msg, err := mail.ReadMessage(strings.NewReader(got.data))
	require.NoError(t, err)

	subject, err := new(mime.WordDecoder).DecodeHeader(msg.Header.Get("Subject"))
	require.NoError(t, err)
	assert.Equal(t, "Восстановление пароля", subject)
	assert.Equal(t, `"Finance App" <noreply@example.com>`, msg.Header.Get("From"))
	assert.Equal(t, "<john@example.com>", msg.Header.Get("To"))
	assert.Equal(t, "1.0", msg.Header.Get("Mime-Version"))
	assert.Contains(t, msg.Header.Get("Message-Id"), "@example.com>")

	body, err := io.ReadAll(quotedprintable.NewReader(msg.Body))
	require.NoError(t, err)
	// SMTP's DATA terminator adds a final line break to the body.
	assert.Equal(t, "Привет!\r\nСсылка: https://example.com/reset?code=abc", strings.TrimRight(string(body), "\r\n"))
}

func TestMailer_Send_RejectsHeaderInjection(t *testing.T) {
	m := mailer.New(mailer.Config{Host: "127.0.0.1", Port: "1", Security: mailer.SecurityNone, From: "noreply@example.com"})

	err := m.Send(context.Background(), mailer.Message{
		To:      "john@example.com",
		Subject: "hello\r\nBcc: attacker@example.com",
		Body:    "x",
	})

	assert.ErrorContains(t, err, "line breaks")
}

func TestMailer_Send_RejectsInvalidRecipient(t *testing.T) {
	m := mailer.New(mailer.Config{Host: "127.0.0.1", Port: "1", Security: mailer.SecurityNone, From: "noreply@example.com"})

	for _, to := range []string{"", "not-an-email", "a@b.com\r\nBcc: x@y.com"} {
		err := m.Send(context.Background(), mailer.Message{To: to, Subject: "s", Body: "b"})
		assert.Error(t, err, to)
	}
}

func TestMailer_Send_DialFailureIsReported(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	_, port, _ := net.SplitHostPort(ln.Addr().String())
	require.NoError(t, ln.Close()) // nothing listens on this port any more

	m := mailer.New(mailer.Config{Host: "127.0.0.1", Port: port, Security: mailer.SecurityNone, From: "noreply@example.com"})

	err = m.Send(context.Background(), mailer.Message{To: "john@example.com", Subject: "s", Body: "b"})

	assert.ErrorContains(t, err, "smtp dial")
}

func TestMailer_Send_RespectsContextDeadline(t *testing.T) {
	// A server that accepts the connection but never speaks.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = ln.Close() })
	go func() {
		conn, err := ln.Accept()
		if err == nil {
			t.Cleanup(func() { _ = conn.Close() })
		}
	}()
	host, port, _ := net.SplitHostPort(ln.Addr().String())

	m := mailer.New(mailer.Config{Host: host, Port: port, Security: mailer.SecurityNone, From: "noreply@example.com"})

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	start := time.Now()
	err = m.Send(ctx, mailer.Message{To: "john@example.com", Subject: "s", Body: "b"})

	assert.Error(t, err)
	assert.Less(t, time.Since(start), 3*time.Second)
}

func TestConfig_Validate(t *testing.T) {
	valid := mailer.Config{Host: "smtp-pulse.com", Port: "587", From: "noreply@example.com", Security: mailer.SecuritySTARTTLS, Username: "u", Password: "p"}

	tests := []struct {
		name    string
		mutate  func(c *mailer.Config)
		wantErr string
	}{
		{"valid", func(c *mailer.Config) {}, ""},
		{"disabled needs nothing", func(c *mailer.Config) { *c = mailer.Config{} }, ""},
		{"invalid from", func(c *mailer.Config) { c.From = "nope" }, "SMTP_FROM"},
		{"empty from", func(c *mailer.Config) { c.From = "" }, "SMTP_FROM"},
		{"unknown security", func(c *mailer.Config) { c.Security = "ssl" }, "SMTP_SECURITY"},
		{"username without password", func(c *mailer.Config) { c.Password = "" }, "together"},
		{"no credentials is fine", func(c *mailer.Config) { c.Username, c.Password = "", "" }, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := valid
			tt.mutate(&c)

			err := c.Validate()

			if tt.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}
