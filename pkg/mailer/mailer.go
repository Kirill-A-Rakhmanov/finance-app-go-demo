package mailer

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"fmt"
	"mime"
	"mime/quotedprintable"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
	"time"
)

const (
	SecuritySTARTTLS = "starttls"
	SecurityTLS      = "tls"
	SecurityNone     = "none"
	sessionTimeout   = 20 * time.Second
)

type Config struct {
	Host     string `envconfig:"SMTP_HOST"`
	Port     string `envconfig:"SMTP_PORT" default:"587"`
	Username string `envconfig:"SMTP_USERNAME"`
	Password string `envconfig:"SMTP_PASSWORD"`
	From     string `envconfig:"SMTP_FROM"`
	FromName string `envconfig:"SMTP_FROM_NAME" default:"Finance App"`
	Security string `envconfig:"SMTP_SECURITY" default:"starttls"`
}

func (c Config) Enabled() bool {
	return c.Host != ""
}

func (c Config) Validate() error {
	if !c.Enabled() {
		return nil
	}

	if _, err := mail.ParseAddress(c.From); err != nil {
		return fmt.Errorf("SMTP_FROM must be a valid email address: %w", err)
	}

	switch c.Security {
	case SecuritySTARTTLS, SecurityTLS, SecurityNone:
	default:
		return fmt.Errorf("SMTP_SECURITY must be one of %s, %s, %s, got %q", SecuritySTARTTLS, SecurityTLS, SecurityNone, c.Security)
	}

	if (c.Username == "") != (c.Password == "") {
		return errors.New("SMTP_USERNAME and SMTP_PASSWORD must be set together")
	}

	return nil
}

type Message struct {
	To      string
	Subject string
	Body    string
}

type Mailer struct {
	cfg Config
}

func New(cfg Config) *Mailer {
	return &Mailer{cfg: cfg}
}

func (m *Mailer) Send(ctx context.Context, msg Message) error {
	to, err := mail.ParseAddress(msg.To)
	if err != nil {
		return fmt.Errorf("invalid recipient: %w", err)
	}

	from := mail.Address{Name: m.cfg.FromName, Address: m.cfg.From}

	data, err := buildMessage(from, *to, msg)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, sessionTimeout)
	defer cancel()

	client, err := m.connect(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	if m.cfg.Username != "" {
		auth := smtp.PlainAuth("", m.cfg.Username, m.cfg.Password, m.cfg.Host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
	}

	if err := client.Mail(from.Address); err != nil {
		return fmt.Errorf("smtp MAIL FROM: %w", err)
	}

	if err := client.Rcpt(to.Address); err != nil {
		return fmt.Errorf("smtp RCPT TO: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp DATA: %w", err)
	}

	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("smtp send: %w", err)
	}

	return client.Quit()
}

func (m *Mailer) connect(ctx context.Context) (*smtp.Client, error) {
	addr := net.JoinHostPort(m.cfg.Host, m.cfg.Port)
	tlsConfig := &tls.Config{ServerName: m.cfg.Host, MinVersion: tls.VersionTLS12}

	var dialer net.Dialer

	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("smtp dial %s: %w", addr, err)
	}

	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
	}

	if m.cfg.Security == SecurityTLS {
		conn = tls.Client(conn, tlsConfig)
	}

	client, err := smtp.NewClient(conn, m.cfg.Host)
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("smtp handshake: %w", err)
	}

	if m.cfg.Security == SecuritySTARTTLS {
		if err := client.StartTLS(tlsConfig); err != nil {
			_ = client.Close()
			return nil, fmt.Errorf("smtp starttls: %w", err)
		}
	}

	return client, nil
}

func buildMessage(from, to mail.Address, msg Message) ([]byte, error) {
	if strings.ContainsAny(msg.Subject, "\r\n") {
		return nil, errors.New("subject must not contain line breaks")
	}

	messageID, err := newMessageID(from.Address)
	if err != nil {
		return nil, err
	}

	var b strings.Builder

	writeHeader := func(name, value string) {
		b.WriteString(name + ": " + value + "\r\n")
	}

	writeHeader("From", from.String())
	writeHeader("To", to.String())
	writeHeader("Subject", mime.QEncoding.Encode("utf-8", msg.Subject))
	writeHeader("Date", time.Now().Format(time.RFC1123Z))
	writeHeader("Message-ID", messageID)
	writeHeader("MIME-Version", "1.0")
	writeHeader("Content-Type", `text/plain; charset="utf-8"`)
	writeHeader("Content-Transfer-Encoding", "quoted-printable")
	b.WriteString("\r\n")

	qp := quotedprintable.NewWriter(&b)
	if _, err := qp.Write([]byte(normalizeNewlines(msg.Body))); err != nil {
		return nil, err
	}
	if err := qp.Close(); err != nil {
		return nil, err
	}

	return []byte(b.String()), nil
}

func normalizeNewlines(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return strings.ReplaceAll(s, "\n", "\r\n")
}

func newMessageID(fromAddress string) (string, error) {
	domain := "localhost"
	if at := strings.LastIndex(fromAddress, "@"); at >= 0 {
		domain = fromAddress[at+1:]
	}

	random := make([]byte, 12)
	if _, err := rand.Read(random); err != nil {
		return "", err
	}

	return fmt.Sprintf("<%s@%s>", hex.EncodeToString(random), domain), nil
}
