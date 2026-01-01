package captcha

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Verifier interface for CAPTCHA verification (supports multiple providers)
type Verifier interface {
	Verify(ctx context.Context, token string, remoteIP string) (*VerifyResult, error)
}

// VerifyResult contains the verification outcome
type VerifyResult struct {
	Success     bool
	Score       float64 // 0.0 to 1.0 (for reCAPTCHA v3)
	Action      string
	ChallengeTS time.Time
	Hostname    string
	ErrorCodes  []string
}

// --- Google reCAPTCHA Implementation ---

type ReCAPTCHAConfig struct {
	SecretKey string
	MinScore  float64 // Minimum score to pass (default 0.5)
	VerifyURL string
}

type ReCAPTCHAVerifier struct {
	config ReCAPTCHAConfig
	client *http.Client
}

// NewReCAPTCHAVerifier creates a Google reCAPTCHA verifier
func NewReCAPTCHAVerifier(secretKey string, minScore float64) *ReCAPTCHAVerifier {
	if minScore == 0 {
		minScore = 0.5
	}
	return &ReCAPTCHAVerifier{
		config: ReCAPTCHAConfig{
			SecretKey: secretKey,
			MinScore:  minScore,
			VerifyURL: "https://www.google.com/recaptcha/api/siteverify",
		},
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

type recaptchaResponse struct {
	Success     bool     `json:"success"`
	Score       float64  `json:"score"`
	Action      string   `json:"action"`
	ChallengeTS string   `json:"challenge_ts"`
	Hostname    string   `json:"hostname"`
	ErrorCodes  []string `json:"error-codes"`
}

func (v *ReCAPTCHAVerifier) Verify(ctx context.Context, token string, remoteIP string) (*VerifyResult, error) {
	if token == "" {
		return &VerifyResult{Success: false, ErrorCodes: []string{"missing-input-response"}}, nil
	}

	// Build request
	form := url.Values{}
	form.Add("secret", v.config.SecretKey)
	form.Add("response", token)
	if remoteIP != "" {
		form.Add("remoteip", remoteIP)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", v.config.VerifyURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.URL.RawQuery = form.Encode()

	resp, err := v.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("verification request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result recaptchaResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Parse timestamp
	var challengeTS time.Time
	if result.ChallengeTS != "" {
		challengeTS, _ = time.Parse(time.RFC3339, result.ChallengeTS)
	}

	verifyResult := &VerifyResult{
		Success:     result.Success && result.Score >= v.config.MinScore,
		Score:       result.Score,
		Action:      result.Action,
		ChallengeTS: challengeTS,
		Hostname:    result.Hostname,
		ErrorCodes:  result.ErrorCodes,
	}

	return verifyResult, nil
}

// --- Cloudflare Turnstile Implementation ---

type TurnstileConfig struct {
	SecretKey string
	VerifyURL string
}

type TurnstileVerifier struct {
	config TurnstileConfig
	client *http.Client
}

// NewTurnstileVerifier creates a Cloudflare Turnstile verifier
func NewTurnstileVerifier(secretKey string) *TurnstileVerifier {
	return &TurnstileVerifier{
		config: TurnstileConfig{
			SecretKey: secretKey,
			VerifyURL: "https://challenges.cloudflare.com/turnstile/v0/siteverify",
		},
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

type turnstileResponse struct {
	Success     bool     `json:"success"`
	ErrorCodes  []string `json:"error-codes"`
	ChallengeTS string   `json:"challenge_ts"`
	Hostname    string   `json:"hostname"`
}

func (v *TurnstileVerifier) Verify(ctx context.Context, token string, remoteIP string) (*VerifyResult, error) {
	if token == "" {
		return &VerifyResult{Success: false, ErrorCodes: []string{"missing-input-response"}}, nil
	}

	form := url.Values{}
	form.Add("secret", v.config.SecretKey)
	form.Add("response", token)
	if remoteIP != "" {
		form.Add("remoteip", remoteIP)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", v.config.VerifyURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.URL.RawQuery = form.Encode()

	resp, err := v.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("verification request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result turnstileResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	var challengeTS time.Time
	if result.ChallengeTS != "" {
		challengeTS, _ = time.Parse(time.RFC3339, result.ChallengeTS)
	}

	return &VerifyResult{
		Success:     result.Success,
		Score:       1.0, // Turnstile is binary pass/fail
		ChallengeTS: challengeTS,
		Hostname:    result.Hostname,
		ErrorCodes:  result.ErrorCodes,
	}, nil
}

// --- Middleware ---

var ErrCAPTCHARequired = errors.New("CAPTCHA verification required")
var ErrCAPTCHAFailed = errors.New("CAPTCHA verification failed")

// Middleware creates an HTTP middleware that requires CAPTCHA for protected routes
func Middleware(verifier Verifier, headerName string) func(http.Handler) http.Handler {
	if headerName == "" {
		headerName = "X-Captcha-Token"
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get(headerName)
			if token == "" {
				http.Error(w, "CAPTCHA token required", http.StatusBadRequest)
				return
			}

			result, err := verifier.Verify(r.Context(), token, r.RemoteAddr)
			if err != nil {
				http.Error(w, "CAPTCHA verification error", http.StatusInternalServerError)
				return
			}

			if !result.Success {
				w.Header().Set("X-Captcha-Error", fmt.Sprintf("%v", result.ErrorCodes))
				http.Error(w, "CAPTCHA verification failed", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
