package domain

import (
	"encoding/json"
	"time"
)

// AuditLog represents an immutable audit log entry
type AuditLog struct {
	ID           string          `json:"id"`
	AggregateID  string          `json:"aggregate_id"`
	EventType    string          `json:"event_type"`
	ResourceType string          `json:"resource_type"`
	Action       string          `json:"action"`
	ActorID      string          `json:"actor_id"`
	ActorType    string          `json:"actor_type"` // "user", "system", "service"
	OrganizationID string        `json:"organization_id"`
	IPAddress    string          `json:"ip_address"`
	UserAgent    string          `json:"user_agent"`
	Details      json.RawMessage `json:"details,omitempty"`
	Status       string          `json:"status"` // "success", "failure"
	FailureReason string         `json:"failure_reason,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	Archived     bool            `json:"archived"`
	ArchivedAt   *time.Time      `json:"archived_at,omitempty"`
}

// AuditEventType constants
const (
	EventUserLogin       = "user.login"
	EventUserLogout      = "user.logout"
	EventUserRegister    = "user.register"
	EventOrgCreate       = "org.create"
	EventOrgUpdate       = "org.update"
	EventOrgInvite       = "org.invite"
	EventOrderCreate     = "order.create"
	EventOrderCancel     = "order.cancel"
	EventPaymentProcess  = "payment.process"
	EventPaymentRefund   = "payment.refund"
	EventSeatHold        = "seat.hold"
	EventSeatBook        = "seat.book"
	EventSeatRelease     = "seat.release"
	EventConfigChange    = "config.change"
	EventDataExport      = "data.export"
	EventDataDelete     = "data.delete"
)

// AuditStatus constants
const (
	StatusSuccess = "success"
	StatusFailure = "failure"
)

// ActorType constants
const (
	ActorUser   = "user"
	ActorSystem = "system"
	ActorService = "service"
)
