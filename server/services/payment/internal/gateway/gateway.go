package gateway

import (
	"context"
	"errors"
	"fmt"
)

// Gateway defines the interface for payment providers
// Implement this for each payment gateway: SSLCommerz, bKash, Nagad, Stripe
type Gateway interface {
	// Name returns the gateway identifier
	Name() string

	// CreatePayment initiates a payment and returns redirect URL or session
	CreatePayment(ctx context.Context, req *CreatePaymentRequest) (*CreatePaymentResponse, error)

	// VerifyPayment verifies payment status after redirect/IPN
	VerifyPayment(ctx context.Context, transactionID string) (*PaymentStatus, error)

	// CapturePayment captures an authorized payment
	CapturePayment(ctx context.Context, transactionID string) (*PaymentStatus, error)

	// RefundPayment initiates a refund
	RefundPayment(ctx context.Context, transactionID string, amountPaisa int64, reason string) (*RefundResponse, error)

	// ValidateIPN validates Instant Payment Notification
	ValidateIPN(ctx context.Context, payload []byte) (*IPNData, error)

	// HealthCheck verifies gateway is operational
	HealthCheck(ctx context.Context) error
}

// CreatePaymentRequest contains payment initiation data
type CreatePaymentRequest struct {
	OrderID       string            `json:"order_id"`
	Amount        Money             `json:"amount"`
	Currency      string            `json:"currency"`
	CustomerName  string            `json:"customer_name"`
	CustomerEmail string            `json:"customer_email"`
	CustomerPhone string            `json:"customer_phone"`
	Description   string            `json:"description"`
	ReturnURL     string            `json:"return_url"`
	CancelURL     string            `json:"cancel_url"`
	IPNURL        string            `json:"ipn_url"`
	Metadata      map[string]string `json:"metadata"`
}

type Money struct {
	AmountPaisa int64  `json:"amount_paisa"`
	Currency    string `json:"currency"` // BDT, USD
}

func (m Money) AmountTaka() float64 {
	return float64(m.AmountPaisa) / 100
}

func (m Money) AmountString() string {
	return formatAmount(m.AmountPaisa, m.Currency)
}

// CreatePaymentResponse contains payment session data
type CreatePaymentResponse struct {
	TransactionID string            `json:"transaction_id"`
	SessionID     string            `json:"session_id"`
	RedirectURL   string            `json:"redirect_url"`
	GatewayRef    string            `json:"gateway_ref"`
	ExpiresAt     int64             `json:"expires_at"`
	Status        string            `json:"status"`
	Metadata      map[string]string `json:"metadata"`
}

// PaymentStatus represents current payment state
type PaymentStatus struct {
	TransactionID string `json:"transaction_id"`
	GatewayRef    string `json:"gateway_ref"`
	Status        Status `json:"status"`
	AmountPaisa   int64  `json:"amount_paisa"`
	Currency      string `json:"currency"`
	CardBrand     string `json:"card_brand,omitempty"`
	CardLast4     string `json:"card_last4,omitempty"`
	BankTranID    string `json:"bank_tran_id,omitempty"`
	RiskLevel     string `json:"risk_level,omitempty"`
	FailureReason string `json:"failure_reason,omitempty"`
	ProcessedAt   int64  `json:"processed_at"`
}

type Status string

const (
	StatusPending    Status = "pending"
	StatusProcessing Status = "processing"
	StatusAuthorized Status = "authorized"
	StatusCaptured   Status = "captured"
	StatusFailed     Status = "failed"
	StatusCancelled  Status = "cancelled"
	StatusRefunded   Status = "refunded"
)

// RefundResponse contains refund result
type RefundResponse struct {
	RefundID      string `json:"refund_id"`
	TransactionID string `json:"transaction_id"`
	AmountPaisa   int64  `json:"amount_paisa"`
	Status        string `json:"status"`
	Reason        string `json:"reason"`
	ProcessedAt   int64  `json:"processed_at"`
}

// IPNData contains parsed IPN payload
type IPNData struct {
	TransactionID string `json:"transaction_id"`
	OrderID       string `json:"order_id"`
	Status        Status `json:"status"`
	AmountPaisa   int64  `json:"amount_paisa"`
	GatewayRef    string `json:"gateway_ref"`
	BankTranID    string `json:"bank_tran_id"`
	CardType      string `json:"card_type"`
	RiskLevel     string `json:"risk_level"`
	Signature     string `json:"signature"`
	IsValid       bool   `json:"is_valid"`
}

// --- Errors ---

var (
	ErrPaymentNotFound      = errors.New("payment not found")
	ErrPaymentAlreadyExists = errors.New("payment already exists for order")
	ErrInsufficientFunds    = errors.New("insufficient funds")
	ErrCardDeclined         = errors.New("card declined")
	ErrInvalidCard          = errors.New("invalid card details")
	ErrGatewayError         = errors.New("payment gateway error")
	ErrIPNValidationFailed  = errors.New("IPN signature validation failed")
	ErrRefundFailed         = errors.New("refund failed")
)

// --- Helpers ---

func formatAmount(paisa int64, currency string) string {
	switch currency {
	case "BDT":
		return "৳" + formatNumber(float64(paisa)/100)
	case "USD":
		return "$" + formatNumber(float64(paisa)/100)
	default:
		return formatNumber(float64(paisa)/100) + " " + currency
	}
}

func formatNumber(n float64) string {
	if n == float64(int(n)) {
		return fmt.Sprintf("%.0f", n)
	}
	return fmt.Sprintf("%.2f", n)
}
