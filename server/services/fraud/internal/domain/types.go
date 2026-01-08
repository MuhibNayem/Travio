// Package domain defines fraud detection types.
package domain

import "time"

// RiskLevel represents the fraud risk level.
type RiskLevel string

const (
	RiskLevelLow      RiskLevel = "low"
	RiskLevelMedium   RiskLevel = "medium"
	RiskLevelHigh     RiskLevel = "high"
	RiskLevelCritical RiskLevel = "critical"
)

// FraudResult represents the result of a fraud analysis.
type FraudResult struct {
	// Risk score from 0-100 (higher = more risky)
	RiskScore int `json:"risk_score"`
	// Risk level derived from score
	RiskLevel RiskLevel `json:"risk_level"`
	// Confidence in the analysis (0-100)
	Confidence int `json:"confidence"`
	// Specific risk factors identified
	RiskFactors []RiskFactor `json:"risk_factors"`
	// Whether the transaction should be blocked
	ShouldBlock bool `json:"should_block"`
	// Human-readable summary
	Summary string `json:"summary"`
	// Analysis timestamp
	AnalyzedAt time.Time `json:"analyzed_at"`
	// Model used for analysis
	Model string `json:"model"`
}

// RiskFactor represents a specific fraud indicator.
type RiskFactor struct {
	Code        string `json:"code"`
	Description string `json:"description"`
	Severity    string `json:"severity"` // low, medium, high
	Score       int    `json:"score"`    // Contribution to total score
}

// BookingAnalysisRequest represents a booking fraud analysis request.
type BookingAnalysisRequest struct {
	// Booking details
	OrderID        string `json:"order_id"`
	OrganizationID string `json:"organization_id"`
	UserID         string `json:"user_id"`
	TripID         string `json:"trip_id"`

	// Passenger details
	PassengerNIDs  []string `json:"passenger_nids"`
	PassengerNames []string `json:"passenger_names"`
	PassengerCount int      `json:"passenger_count"`

	// Booking behavior
	BookingTimestamp time.Time `json:"booking_timestamp"`
	IPAddress        string    `json:"ip_address"`
	UserAgent        string    `json:"user_agent"`
	PaymentMethod    string    `json:"payment_method"`
	TotalAmountPaisa int64     `json:"total_amount_paisa"`

	// Historical context
	BookingsLast24Hours int `json:"bookings_last_24_hours"`
	BookingsLastWeek    int `json:"bookings_last_week"`
	PreviousFraudFlags  int `json:"previous_fraud_flags"`
}

// DocumentVerificationRequest represents a document verification request.
type DocumentVerificationRequest struct {
	DocumentType   string `json:"document_type"` // "nid", "passport"
	DocumentImage  []byte `json:"document_image"`
	ImageMimeType  string `json:"image_mime_type"`
	ExpectedNID    string `json:"expected_nid,omitempty"`
	ExpectedName   string `json:"expected_name,omitempty"`
	OrganizationID string `json:"organization_id"`
}

// DocumentVerificationResult represents document verification result.
type DocumentVerificationResult struct {
	IsAuthentic    bool     `json:"is_authentic"`
	Confidence     int      `json:"confidence"`
	ExtractedNID   string   `json:"extracted_nid,omitempty"`
	ExtractedName  string   `json:"extracted_name,omitempty"`
	TamperingScore int      `json:"tampering_score"` // 0-100, higher = more tampering
	Issues         []string `json:"issues"`
	Summary        string   `json:"summary"`
}

// RiskScoreToLevel converts a risk score to a risk level.
func RiskScoreToLevel(score int) RiskLevel {
	switch {
	case score < 30:
		return RiskLevelLow
	case score < 60:
		return RiskLevelMedium
	case score < 85:
		return RiskLevelHigh
	default:
		return RiskLevelCritical
	}
}
