package provider

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"gopkg.in/gomail.v2"
)

// SMTPProvider implements EmailProvider using SMTP
type SMTPProvider struct {
	host     string
	port     int
	username string
	password string
	from     string
	useTLS   bool
	dialer   *gomail.Dialer
}

// NewSMTPProvider creates a new SMTP email provider
func NewSMTPProvider(host string, port int, username, password, from string, useTLS bool) (*SMTPProvider, error) {
	dialer := gomail.NewDialer(host, port, username, password)
	dialer.TLSConfig = &tls.Config{
		ServerName: host,
		MinVersion: tls.VersionTLS12,
	}

	p := &SMTPProvider{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
		useTLS:   useTLS,
		dialer:   dialer,
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := p.testConnection(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to SMTP server: %w", err)
	}

	logger.Info("SMTP provider initialized",
		"host", host,
		"port", port,
		"from", from,
		"tls", useTLS,
	)

	return p, nil
}

// testConnection verifies SMTP connectivity
func (p *SMTPProvider) testConnection(ctx context.Context) error {
	ch := make(chan error, 1)
	go func() {
		conn, err := p.dialer.Dial()
		if err != nil {
			ch <- err
			return
		}
		conn.Close()
		ch <- nil
	}()

	select {
	case err := <-ch:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// NewSMTPProviderFromEnv creates SMTP provider from environment variables
func NewSMTPProviderFromEnv() (*SMTPProvider, error) {
	host := os.Getenv("SMTP_HOST")
	port := 587
	username := os.Getenv("SMTP_USER")
	password := os.Getenv("SMTP_PASSWORD")
	from := os.Getenv("SMTP_FROM")
	useTLS := os.Getenv("SMTP_USE_TLS") != "false"

	if host == "" || username == "" || password == "" || from == "" {
		return nil, fmt.Errorf("SMTP configuration incomplete: SMTP_HOST, SMTP_USER, SMTP_PASSWORD, SMTP_FROM are required")
	}

	return NewSMTPProvider(host, port, username, password, from, useTLS)
}

// Send sends an email
func (p *SMTPProvider) Send(ctx context.Context, to, subject, body string) error {
	return p.SendWithAttachments(ctx, to, subject, body, nil, nil)
}

// SendWithAttachments sends an email with optional attachments
func (p *SMTPProvider) SendWithAttachments(ctx context.Context, to, subject, body string, attachments []Attachment, headers map[string]string) error {
	msg := gomail.NewMessage()
	msg.SetHeader("From", p.from)
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", subject)
	msg.SetHeader("Content-Type", "text/html; charset=UTF-8")

	// Add custom headers
	for key, value := range headers {
		msg.SetHeader(key, value)
	}

	msg.SetBody("text/html", body)

	// Add attachments
	for _, att := range attachments {
		file := gomail.SetCopyFunc(func(w io.Writer) error {
			_, err := io.Copy(w, bytes.NewReader(att.Data))
			return err
		})
		msg.Attach(att.Filename, file)
	}

	// Retry logic with exponential backoff
	maxRetries := 3
	for attempt := 0; attempt < maxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err := p.dialer.DialAndSend(msg)
		if err == nil {
			logger.Info("Email sent successfully",
				"to", to,
				"subject", subject,
				"attempt", attempt+1,
			)
			return nil
		}

		logger.Warn("SMTP send failed, retrying",
			"to", to,
			"attempt", attempt+1,
			"max_retries", maxRetries,
			"error", err,
		)

		if attempt < maxRetries-1 {
			backoff := time.Duration(attempt+1) * 2 * time.Second
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
		}
	}

	return fmt.Errorf("failed to send email after %d attempts", maxRetries)
}

// HealthCheck verifies SMTP connectivity
func (p *SMTPProvider) HealthCheck(ctx context.Context) error {
	return p.testConnection(ctx)
}

// Attachment represents an email attachment
type Attachment struct {
	Filename string
	Data     []byte
}
