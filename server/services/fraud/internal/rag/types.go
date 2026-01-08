// Package rag provides Retrieval Augmented Generation for fraud detection.
package rag

import "time"

// FraudCase represents a historical fraud case for RAG.
type FraudCase struct {
	ID             string `json:"id" gorm:"primaryKey;type:uuid"`
	OrganizationID string `json:"organization_id" gorm:"type:uuid;index"`

	// Booking data snapshot
	UserID         string   `json:"user_id"`
	OrderID        string   `json:"order_id"`
	TripID         string   `json:"trip_id"`
	Route          string   `json:"route"`
	PassengerCount int      `json:"passenger_count"`
	AmountPaisa    int64    `json:"amount_paisa"`
	IPAddress      string   `json:"ip_address"`
	UserAgent      string   `json:"user_agent"`
	PassengerNIDs  []string `json:"passenger_nids" gorm:"type:jsonb;serializer:json"`

	// Analysis results
	RiskScore   int    `json:"risk_score"`
	RiskFactors string `json:"risk_factors" gorm:"type:text"` // JSON string

	// Outcome
	Outcome     string    `json:"outcome"` // "fraud", "legitimate", "review", "blocked"
	ReviewedBy  string    `json:"reviewed_by,omitempty"`
	ReviewNotes string    `json:"review_notes,omitempty"`
	OutcomeAt   time.Time `json:"outcome_at,omitempty"`

	// Embedding for similarity search (768 dimensions for text-embedding-005)
	Embedding []float32 `json:"embedding" gorm:"type:vector(768)"`

	// Metadata
	CreatedAt time.Time `json:"created_at"`
}

// CaseOutcome constants
const (
	OutcomeFraud      = "fraud"
	OutcomeLegitimate = "legitimate"
	OutcomeReview     = "review"
	OutcomeBlocked    = "blocked"
)

// SimilarCase represents a retrieved similar case with similarity score.
type SimilarCase struct {
	Case       *FraudCase
	Similarity float64 // 0-1, higher = more similar
}

// RAGContext represents the context built from similar cases for the LLM.
type RAGContext struct {
	Cases   []SimilarCase
	Summary string
}

// EmbeddingRequest represents a request to generate embeddings.
type EmbeddingRequest struct {
	Text string
}

// BookingToText converts booking data to text for embedding.
func BookingToText(orderID, userID, route string, amount int64, passengerCount int, ipAddress, userAgent string, riskScore int) string {
	return "Booking: " +
		"route=" + route +
		", amount=" + formatAmount(amount) +
		", passengers=" + formatInt(passengerCount) +
		", risk_score=" + formatInt(riskScore) +
		", ip=" + ipAddress
}

func formatAmount(paisa int64) string {
	taka := float64(paisa) / 100
	return formatFloat(taka) + " BDT"
}

func formatInt(n int) string {
	if n == 0 {
		return "0"
	}
	result := ""
	for n > 0 {
		result = string(rune('0'+n%10)) + result
		n /= 10
	}
	return result
}

func formatFloat(f float64) string {
	whole := int(f)
	frac := int((f - float64(whole)) * 100)
	return formatInt(whole) + "." + formatInt(frac)
}
