package domain

import (
	"time"
)

// QueueEntry represents a user in the virtual queue
type QueueEntry struct {
	ID            string        `json:"id"`
	UserID        string        `json:"user_id"`
	SessionID     string        `json:"session_id"`
	EventID       string        `json:"event_id"` // The high-demand event/trip
	Position      int           `json:"position"`
	Token         string        `json:"token"`
	JoinedAt      time.Time     `json:"joined_at"`
	EstimatedWait time.Duration `json:"estimated_wait"`
	Status        QueueStatus   `json:"status"`
	ExpiresAt     time.Time     `json:"expires_at"`
}

// QueueStatus represents the state of a queue entry
type QueueStatus string

const (
	QueueStatusWaiting   QueueStatus = "waiting"
	QueueStatusReady     QueueStatus = "ready" // Admitted to purchase
	QueueStatusExpired   QueueStatus = "expired"
	QueueStatusCompleted QueueStatus = "completed"
)

// QueueStats provides real-time queue statistics
type QueueStats struct {
	EventID       string        `json:"event_id"`
	TotalWaiting  int           `json:"total_waiting"`
	TotalAdmitted int           `json:"total_admitted"`
	AvgWaitTime   time.Duration `json:"avg_wait_time"`
	AdmissionRate int           `json:"admission_rate_per_min"`
	EstimatedWait time.Duration `json:"estimated_wait"`
}

// AdmissionConfig controls queue admission behavior
type AdmissionConfig struct {
	EventID            string        `json:"event_id"`
	MaxConcurrent      int           `json:"max_concurrent"`       // Max users purchasing at once
	AdmissionBatchSize int           `json:"admission_batch_size"` // Users to admit per batch
	AdmissionInterval  time.Duration `json:"admission_interval"`   // Time between batches
	TokenTTL           time.Duration `json:"token_ttl"`            // How long admitted token is valid
	QueueEnabled       bool          `json:"queue_enabled"`
}
