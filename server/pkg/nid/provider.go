package nid

import (
	"context"
	"errors"
	"time"
)

// Provider interface for NID verification (Strategy Pattern)
// Implement this for each country/provider: Porichoy (BD), Aadhaar (IN), etc.
type Provider interface {
	// Name returns the provider identifier
	Name() string

	// Country returns the ISO country code this provider handles
	Country() string

	// Verify performs NID verification and returns citizen data
	Verify(ctx context.Context, req *VerifyRequest) (*VerifyResponse, error)

	// HealthCheck verifies the provider is operational
	HealthCheck(ctx context.Context) error
}

// VerifyRequest contains NID verification input
type VerifyRequest struct {
	NID         string    `json:"nid"`
	DateOfBirth time.Time `json:"date_of_birth"`
	Name        string    `json:"name,omitempty"`  // Optional: for name matching
	Phone       string    `json:"phone,omitempty"` // Optional: for OTP verification

	// Request metadata
	RequestID   string `json:"request_id"`
	RequesterIP string `json:"requester_ip"`
}

// VerifyResponse contains NID verification result
type VerifyResponse struct {
	// Verification Status
	IsValid    bool        `json:"is_valid"`
	Confidence float64     `json:"confidence"` // 0.0 to 1.0
	MatchScore *MatchScore `json:"match_score,omitempty"`

	// Citizen Data (only if verified)
	Citizen *CitizenData `json:"citizen,omitempty"`

	// Verification Metadata
	ProviderName string    `json:"provider_name"`
	VerifiedAt   time.Time `json:"verified_at"`
	ExpiresAt    time.Time `json:"expires_at"` // Cache expiry

	// Error details (if verification failed)
	ErrorCode    string `json:"error_code,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// CitizenData contains verified citizen information
type CitizenData struct {
	NID         string    `json:"nid"`
	NameBN      string    `json:"name_bn"` // Bengali name
	NameEN      string    `json:"name_en"` // English name
	FatherName  string    `json:"father_name"`
	MotherName  string    `json:"mother_name"`
	DateOfBirth time.Time `json:"date_of_birth"`
	Gender      string    `json:"gender"` // male, female, other
	BloodGroup  string    `json:"blood_group,omitempty"`

	// Address
	PresentAddress   *Address `json:"present_address,omitempty"`
	PermanentAddress *Address `json:"permanent_address,omitempty"`

	// Photo (Base64 encoded)
	Photo string `json:"photo,omitempty"`

	// Voter Info
	VoterArea string `json:"voter_area,omitempty"`
}

// Address represents a citizen's address
type Address struct {
	Division string `json:"division"`
	District string `json:"district"`
	Upazila  string `json:"upazila"`
	Union    string `json:"union,omitempty"`
	Village  string `json:"village,omitempty"`
	PostCode string `json:"post_code,omitempty"`
	FullText string `json:"full_text,omitempty"`
}

// MatchScore indicates how well the provided data matches the verified data
type MatchScore struct {
	NameMatch    float64 `json:"name_match"` // 0.0 to 1.0
	DOBMatch     bool    `json:"dob_match"`
	OverallMatch float64 `json:"overall_match"`
}

// --- Errors ---

var (
	ErrInvalidNIDFormat    = errors.New("invalid NID format")
	ErrNIDNotFound         = errors.New("NID not found in database")
	ErrDOBMismatch         = errors.New("date of birth does not match")
	ErrProviderUnavailable = errors.New("verification provider unavailable")
	ErrRateLimited         = errors.New("rate limited by provider")
	ErrInvalidCredentials  = errors.New("invalid API credentials")
	ErrVerificationFailed  = errors.New("verification failed")
)

// ErrorCode constants for structured error handling
const (
	ErrorCodeInvalidFormat = "INVALID_FORMAT"
	ErrorCodeNotFound      = "NOT_FOUND"
	ErrorCodeDOBMismatch   = "DOB_MISMATCH"
	ErrorCodeProviderDown  = "PROVIDER_DOWN"
	ErrorCodeRateLimited   = "RATE_LIMITED"
	ErrorCodeUnauthorized  = "UNAUTHORIZED"
	ErrorCodeInternalError = "INTERNAL_ERROR"
)

// --- Validation ---

// Validator provides NID format validation per country
type Validator interface {
	Validate(nid string) error
	Normalize(nid string) string
}

// BangladeshNIDValidator validates Bangladesh NIDs
type BangladeshNIDValidator struct{}

func (v *BangladeshNIDValidator) Validate(nid string) error {
	// Bangladesh NID: 10 digits (old) or 17 digits (smart card)
	if len(nid) != 10 && len(nid) != 17 {
		return ErrInvalidNIDFormat
	}

	for _, c := range nid {
		if c < '0' || c > '9' {
			return ErrInvalidNIDFormat
		}
	}

	// Additional checksum validation for 17-digit NIDs
	if len(nid) == 17 {
		// First 2 digits: birth year (last 2 digits)
		// Next 2 digits: district code
		// Remaining: unique identifier
		// TODO: Add checksum algorithm
	}

	return nil
}

func (v *BangladeshNIDValidator) Normalize(nid string) string {
	// Remove any spaces or dashes
	result := ""
	for _, c := range nid {
		if c >= '0' && c <= '9' {
			result += string(c)
		}
	}
	return result
}

// IndiaAadhaarValidator validates Indian Aadhaar numbers
type IndiaAadhaarValidator struct{}

func (v *IndiaAadhaarValidator) Validate(nid string) error {
	// Aadhaar: 12 digits with Verhoeff checksum
	if len(nid) != 12 {
		return ErrInvalidNIDFormat
	}

	for _, c := range nid {
		if c < '0' || c > '9' {
			return ErrInvalidNIDFormat
		}
	}

	// First digit cannot be 0 or 1
	if nid[0] == '0' || nid[0] == '1' {
		return ErrInvalidNIDFormat
	}

	// TODO: Add Verhoeff checksum validation

	return nil
}

func (v *IndiaAadhaarValidator) Normalize(nid string) string {
	result := ""
	for _, c := range nid {
		if c >= '0' && c <= '9' {
			result += string(c)
		}
	}
	return result
}
