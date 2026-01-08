// Package profile provides user behavior profiling for fraud detection.
package profile

import "time"

// UserProfile represents a user's behavioral profile for fraud detection.
type UserProfile struct {
	UserID string `json:"user_id" gorm:"primaryKey;type:uuid"`

	// Booking behavior statistics
	TotalBookings      int     `json:"total_bookings"`
	AvgBookingValue    float64 `json:"avg_booking_value"`
	BookingVelocity24h float64 `json:"booking_velocity_24h"` // Avg bookings per 24h
	BookingVelocity7d  float64 `json:"booking_velocity_7d"`  // Avg bookings per week

	// Pattern data (stored as JSON)
	CommonRoutes       []string `json:"common_routes" gorm:"type:jsonb;serializer:json"`
	CommonTimes        []int    `json:"common_times" gorm:"type:jsonb;serializer:json"` // Hours of day (0-23)
	DeviceFingerprints []string `json:"device_fingerprints" gorm:"type:jsonb;serializer:json"`
	CommonIPs          []string `json:"common_ips" gorm:"type:jsonb;serializer:json"`

	// Risk history
	RiskScores   []float64 `json:"risk_scores" gorm:"type:jsonb;serializer:json"` // Last 10 scores
	AvgRiskScore float64   `json:"avg_risk_score"`
	FraudFlags   int       `json:"fraud_flags"`   // Number of times flagged
	BlockedCount int       `json:"blocked_count"` // Number of times blocked

	// Embedding for similarity search (768 dimensions for text-embedding-005)
	Embedding []float32 `json:"embedding" gorm:"type:vector(768)"`

	// Timestamps
	FirstSeen time.Time `json:"first_seen"`
	LastSeen  time.Time `json:"last_seen"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
}

// BookingEvent represents a booking event for profile updates.
type BookingEvent struct {
	UserID      string
	OrderID     string
	TripID      string
	Route       string // "origin-destination"
	AmountPaisa int64
	BookingTime time.Time
	IPAddress   string
	UserAgent   string
	RiskScore   float64
	WasBlocked  bool
}

// DeviationResult represents the result of behavior deviation analysis.
type DeviationResult struct {
	// Overall deviation score (0-100)
	Score float64 `json:"score"`

	// Individual deviation factors
	ValueDeviation    float64 `json:"value_deviation"`    // Booking value vs avg
	VelocityDeviation float64 `json:"velocity_deviation"` // Booking frequency vs avg
	TimeDeviation     float64 `json:"time_deviation"`     // Booking time vs common times
	RouteDeviation    float64 `json:"route_deviation"`    // New route flag
	IPDeviation       float64 `json:"ip_deviation"`       // New IP flag
	DeviceDeviation   float64 `json:"device_deviation"`   // New device flag

	// Flags
	IsNewUser       bool `json:"is_new_user"`
	IsAnomalous     bool `json:"is_anomalous"`
	HighRiskHistory bool `json:"high_risk_history"`
}

// ProfileStats holds statistical data for deviation calculation.
type ProfileStats struct {
	Mean   float64
	StdDev float64
	Count  int
}
