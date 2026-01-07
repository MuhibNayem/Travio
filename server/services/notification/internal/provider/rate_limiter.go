package provider

import (
	"context"

	"golang.org/x/time/rate"
)

// RateLimitedEmailProvider wraps an EmailProvider with rate limiting
type RateLimitedEmailProvider struct {
	next    EmailProvider
	limiter *rate.Limiter
}

// EmailProvider interface (must match service.EmailProvider)
type EmailProvider interface {
	Send(ctx context.Context, to, subject, body string) error
}

// NewRateLimitedEmailProvider creates a rate-limited email provider
// rps = requests per second
func NewRateLimitedEmailProvider(next EmailProvider, rps int) *RateLimitedEmailProvider {
	return &RateLimitedEmailProvider{
		next:    next,
		limiter: rate.NewLimiter(rate.Limit(rps), rps), // Bucket size = rps
	}
}

func (p *RateLimitedEmailProvider) Send(ctx context.Context, to, subject, body string) error {
	// Block until allowed
	if err := p.limiter.Wait(ctx); err != nil {
		return err
	}
	return p.next.Send(ctx, to, subject, body)
}

// RateLimitedSMSProvider wraps an SMSProvider with rate limiting
type RateLimitedSMSProvider struct {
	next    SMSProvider
	limiter *rate.Limiter
}

// SMSProvider interface (must match service.SMSProvider)
type SMSProvider interface {
	Send(ctx context.Context, to, message string) error
}

// NewRateLimitedSMSProvider creates a rate-limited SMS provider
func NewRateLimitedSMSProvider(next SMSProvider, rps int) *RateLimitedSMSProvider {
	return &RateLimitedSMSProvider{
		next:    next,
		limiter: rate.NewLimiter(rate.Limit(rps), rps),
	}
}

func (p *RateLimitedSMSProvider) Send(ctx context.Context, to, message string) error {
	if err := p.limiter.Wait(ctx); err != nil {
		return err
	}
	return p.next.Send(ctx, to, message)
}
