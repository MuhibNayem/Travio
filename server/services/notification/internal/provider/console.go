package provider

import (
	"context"

	"github.com/MuhibNayem/Travio/server/pkg/logger"
)

// ConsoleEmailProvider logs emails to console (for development)
type ConsoleEmailProvider struct{}

// NewConsoleEmailProvider creates a console email provider
func NewConsoleEmailProvider() *ConsoleEmailProvider {
	return &ConsoleEmailProvider{}
}

// Send logs the email to console
func (p *ConsoleEmailProvider) Send(ctx context.Context, to, subject, body string) error {
	logger.Info("EMAIL",
		"to", to,
		"subject", subject,
		"body", body,
	)
	return nil
}

// ConsoleSMSProvider logs SMS to console (for development)
type ConsoleSMSProvider struct{}

// NewConsoleSMSProvider creates a console SMS provider
func NewConsoleSMSProvider() *ConsoleSMSProvider {
	return &ConsoleSMSProvider{}
}

// Send logs the SMS to console
func (p *ConsoleSMSProvider) Send(ctx context.Context, to, message string) error {
	logger.Info("SMS",
		"to", to,
		"message", message,
	)
	return nil
}
