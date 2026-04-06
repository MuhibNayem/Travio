package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/MuhibNayem/Travio/server/pkg/logger"
)

// TwilioProvider implements SMSProvider using Twilio API
type TwilioProvider struct {
	accountSid string
	authToken  string
	fromNumber string
	client     *http.Client
	baseURL    string
}

// NewTwilioProvider creates a new Twilio SMS provider
func NewTwilioProvider(accountSid, authToken, fromNumber string) (*TwilioProvider, error) {
	if accountSid == "" || authToken == "" || fromNumber == "" {
		return nil, fmt.Errorf("Twilio configuration incomplete: accountSid, authToken, and fromNumber are required")
	}

	p := &TwilioProvider{
		accountSid: accountSid,
		authToken:  authToken,
		fromNumber: fromNumber,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://api.twilio.com/2010-04-01",
	}

	logger.Info("Twilio SMS provider initialized",
		"account_sid", accountSid,
		"from_number", fromNumber,
	)

	return p, nil
}

// NewTwilioProviderFromEnv creates Twilio provider from environment variables
func NewTwilioProviderFromEnv() (*TwilioProvider, error) {
	accountSid := os.Getenv("TWILIO_ACCOUNT_SID")
	authToken := os.Getenv("TWILIO_AUTH_TOKEN")
	fromNumber := os.Getenv("TWILIO_FROM_NUMBER")

	return NewTwilioProvider(accountSid, authToken, fromNumber)
}

// Send sends an SMS message
func (p *TwilioProvider) Send(ctx context.Context, to, message string) error {
	formData := url.Values{}
	formData.Set("To", to)
	formData.Set("From", p.fromNumber)
	formData.Set("Body", message)

	reqURL := fmt.Sprintf("%s/Accounts/%s/Messages.json", p.baseURL, p.accountSid)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewBufferString(formData.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(p.accountSid, p.authToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Retry logic with exponential backoff
	maxRetries := 3
	for attempt := 0; attempt < maxRetries; attempt++ {
		resp, err := p.client.Do(req)
		if err != nil {
			if attempt < maxRetries-1 {
				backoff := time.Duration(attempt+1) * 2 * time.Second
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(backoff):
				}
				continue
			}
			return fmt.Errorf("failed to send SMS after %d attempts: %w", maxRetries, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			var result twilioMessageResponse
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				logger.Warn("Failed to decode Twilio response", "error", err)
			} else {
				logger.Info("SMS sent successfully via Twilio",
					"to", to,
					"sid", result.Sid,
					"status", result.Status,
				)
			}
			return nil
		}

		// Parse error response
		var twilioErr twilioErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&twilioErr); err == nil {
			logger.Error("Twilio API error",
				"status", resp.StatusCode,
				"code", twilioErr.Code,
				"message", twilioErr.Message,
			)
			return fmt.Errorf("twilio API error %d: %s", twilioErr.Code, twilioErr.Message)
		}

		if attempt < maxRetries-1 {
			backoff := time.Duration(attempt+1) * 2 * time.Second
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
		}
	}

	return fmt.Errorf("failed to send SMS after %d attempts", maxRetries)
}

// HealthCheck verifies Twilio API connectivity
func (p *TwilioProvider) HealthCheck(ctx context.Context) error {
	reqURL := fmt.Sprintf("%s/Accounts/%s.json", p.baseURL, p.accountSid)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return err
	}

	req.SetBasicAuth(p.accountSid, p.authToken)

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("Twilio health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Twilio health check returned status: %d", resp.StatusCode)
	}

	return nil
}

type twilioMessageResponse struct {
	Sid         string `json:"sid"`
	AccountSid  string `json:"account_sid"`
	To          string `json:"to"`
	From        string `json:"from"`
	Body        string `json:"body"`
	Status      string `json:"status"`
	DateCreated string `json:"date_created"`
}

type twilioErrorResponse struct {
	Code     int    `json:"code"`
	Message  string `json:"message"`
	Status   int    `json:"status"`
	MoreInfo string `json:"more_info"`
}
