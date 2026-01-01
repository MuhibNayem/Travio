package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/MuhibNayem/Travio/server/pkg/nid"
)

// PorichoyProvider implements nid.Provider for Bangladesh NID via Porichoy API
// Porichoy is the official Bangladesh government NID verification platform
// API Docs: https://porichoy.gov.bd (requires registration)
type PorichoyProvider struct {
	baseURL     string
	apiKey      string
	apiSecret   string
	client      *http.Client
	validator   *nid.BangladeshNIDValidator
	rateLimiter *rateLimiter
}

// PorichoyConfig holds configuration for Porichoy API
type PorichoyConfig struct {
	BaseURL    string        // Default: https://api.porichoy.gov.bd
	APIKey     string        // Provided by Porichoy
	APISecret  string        // Provided by Porichoy
	Timeout    time.Duration // Default: 30s
	MaxRetries int           // Default: 3
	RateLimit  int           // Requests per minute
}

func NewPorichoyProvider(cfg PorichoyConfig) *PorichoyProvider {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.porichoy.gov.bd"
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}
	if cfg.RateLimit == 0 {
		cfg.RateLimit = 100 // 100 requests per minute
	}

	return &PorichoyProvider{
		baseURL:   cfg.BaseURL,
		apiKey:    cfg.APIKey,
		apiSecret: cfg.APISecret,
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
		validator:   &nid.BangladeshNIDValidator{},
		rateLimiter: newRateLimiter(cfg.RateLimit, time.Minute),
	}
}

func (p *PorichoyProvider) Name() string {
	return "porichoy"
}

func (p *PorichoyProvider) Country() string {
	return "BD"
}

func (p *PorichoyProvider) Verify(ctx context.Context, req *nid.VerifyRequest) (*nid.VerifyResponse, error) {
	// Validate format first
	if err := p.validator.Validate(req.NID); err != nil {
		return &nid.VerifyResponse{
			IsValid:      false,
			ProviderName: p.Name(),
			VerifiedAt:   time.Now(),
			ErrorCode:    nid.ErrorCodeInvalidFormat,
			ErrorMessage: err.Error(),
		}, nil
	}

	// Check rate limit
	if !p.rateLimiter.Allow() {
		return nil, nid.ErrRateLimited
	}

	// Build API request
	apiReq := porichoyRequest{
		NID: p.validator.Normalize(req.NID),
		DOB: req.DateOfBirth.Format("2006-01-02"),
	}

	body, err := json.Marshal(apiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/api/v1/nid/verify", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-API-Key", p.apiKey)
	httpReq.Header.Set("X-API-Secret", p.apiSecret)
	httpReq.Header.Set("X-Request-ID", req.RequestID)

	// Execute request
	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Handle HTTP errors
	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, nid.ErrRateLimited
	}
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, nid.ErrInvalidCredentials
	}
	if resp.StatusCode >= 500 {
		return nil, nid.ErrProviderUnavailable
	}

	// Parse response
	var apiResp porichoyResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Convert to our format
	return p.convertResponse(&apiResp), nil
}

func (p *PorichoyProvider) HealthCheck(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", p.baseURL+"/api/v1/health", nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-API-Key", p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return nid.ErrProviderUnavailable
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nid.ErrProviderUnavailable
	}
	return nil
}

// --- Porichoy API DTOs ---

type porichoyRequest struct {
	NID string `json:"nid"`
	DOB string `json:"dob"` // YYYY-MM-DD
}

type porichoyResponse struct {
	Success bool             `json:"success"`
	Code    string           `json:"code"`
	Message string           `json:"message"`
	Data    *porichoyCitizen `json:"data,omitempty"`
}

type porichoyCitizen struct {
	NationalID    string `json:"national_id"`
	NameBN        string `json:"name_bn"`
	NameEN        string `json:"name_en"`
	FatherNameBN  string `json:"father_name_bn"`
	MotherNameBN  string `json:"mother_name_bn"`
	DOB           string `json:"dob"`
	Gender        string `json:"gender"`
	BloodGroup    string `json:"blood_group"`
	Photo         string `json:"photo"` // Base64
	PresentAddr   string `json:"present_address"`
	PermanentAddr string `json:"permanent_address"`
	VoterArea     string `json:"voter_area"`
}

func (p *PorichoyProvider) convertResponse(resp *porichoyResponse) *nid.VerifyResponse {
	result := &nid.VerifyResponse{
		ProviderName: p.Name(),
		VerifiedAt:   time.Now(),
		ExpiresAt:    time.Now().Add(24 * time.Hour),
	}

	if !resp.Success {
		result.IsValid = false
		result.ErrorCode = mapPorichoyErrorCode(resp.Code)
		result.ErrorMessage = resp.Message
		return result
	}

	if resp.Data == nil {
		result.IsValid = false
		result.ErrorCode = nid.ErrorCodeNotFound
		result.ErrorMessage = "No citizen data returned"
		return result
	}

	// Parse DOB
	dob, _ := time.Parse("2006-01-02", resp.Data.DOB)

	result.IsValid = true
	result.Confidence = 1.0
	result.Citizen = &nid.CitizenData{
		NID:         resp.Data.NationalID,
		NameBN:      resp.Data.NameBN,
		NameEN:      resp.Data.NameEN,
		FatherName:  resp.Data.FatherNameBN,
		MotherName:  resp.Data.MotherNameBN,
		DateOfBirth: dob,
		Gender:      resp.Data.Gender,
		BloodGroup:  resp.Data.BloodGroup,
		Photo:       resp.Data.Photo,
		PresentAddress: &nid.Address{
			FullText: resp.Data.PresentAddr,
		},
		PermanentAddress: &nid.Address{
			FullText: resp.Data.PermanentAddr,
		},
		VoterArea: resp.Data.VoterArea,
	}

	return result
}

func mapPorichoyErrorCode(code string) string {
	switch code {
	case "NID_NOT_FOUND":
		return nid.ErrorCodeNotFound
	case "DOB_MISMATCH":
		return nid.ErrorCodeDOBMismatch
	case "INVALID_NID":
		return nid.ErrorCodeInvalidFormat
	case "RATE_LIMITED":
		return nid.ErrorCodeRateLimited
	default:
		return nid.ErrorCodeInternalError
	}
}

// --- Simple Rate Limiter ---

type rateLimiter struct {
	tokens     int
	maxTokens  int
	refillRate time.Duration
	lastRefill time.Time
}

func newRateLimiter(maxTokens int, refillPeriod time.Duration) *rateLimiter {
	return &rateLimiter{
		tokens:     maxTokens,
		maxTokens:  maxTokens,
		refillRate: refillPeriod / time.Duration(maxTokens),
		lastRefill: time.Now(),
	}
}

func (r *rateLimiter) Allow() bool {
	now := time.Now()
	elapsed := now.Sub(r.lastRefill)

	// Refill tokens
	tokensToAdd := int(elapsed / r.refillRate)
	if tokensToAdd > 0 {
		r.tokens = min(r.tokens+tokensToAdd, r.maxTokens)
		r.lastRefill = now
	}

	if r.tokens > 0 {
		r.tokens--
		return true
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
