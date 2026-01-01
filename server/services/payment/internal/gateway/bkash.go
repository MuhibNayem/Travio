package gateway

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// BKash implements Gateway for bKash Payment Gateway (leading mobile wallet in BD)
// API Docs: https://developer.bka.sh/docs
type BKash struct {
	appKey      string
	appSecret   string
	username    string
	password    string
	baseURL     string
	client      *http.Client
	isSandbox   bool
	token       string
	tokenExpiry time.Time
}

type BKashConfig struct {
	AppKey    string
	AppSecret string
	Username  string
	Password  string
	IsSandbox bool
	Timeout   time.Duration
}

func NewBKash(cfg BKashConfig) *BKash {
	baseURL := "https://tokenized.pay.bka.sh/v1.2.0-beta"
	if cfg.IsSandbox {
		baseURL = "https://tokenized.sandbox.bka.sh/v1.2.0-beta"
	}

	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &BKash{
		appKey:    cfg.AppKey,
		appSecret: cfg.AppSecret,
		username:  cfg.Username,
		password:  cfg.Password,
		baseURL:   baseURL,
		client:    &http.Client{Timeout: timeout},
		isSandbox: cfg.IsSandbox,
	}
}

func (g *BKash) Name() string {
	return "bkash"
}

func (g *BKash) CreatePayment(ctx context.Context, req *CreatePaymentRequest) (*CreatePaymentResponse, error) {
	// Get or refresh token
	if err := g.ensureToken(ctx); err != nil {
		return nil, fmt.Errorf("token error: %w", err)
	}

	// Create payment request
	bkashReq := bkashCreateRequest{
		Mode:                  "0011", // Checkout URL mode
		PayerReference:        req.CustomerPhone,
		CallbackURL:           req.ReturnURL,
		Amount:                fmt.Sprintf("%.2f", req.Amount.AmountTaka()),
		Currency:              req.Currency,
		Intent:                "sale",
		MerchantInvoiceNumber: req.OrderID,
	}

	body, _ := json.Marshal(bkashReq)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", g.baseURL+"/tokenized/checkout/create", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	g.setHeaders(httpReq)

	resp, err := g.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var bkashResp bkashCreateResponse
	if err := json.Unmarshal(respBody, &bkashResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if bkashResp.StatusCode != "0000" {
		return nil, fmt.Errorf("%w: %s", ErrGatewayError, bkashResp.StatusMessage)
	}

	return &CreatePaymentResponse{
		TransactionID: req.OrderID,
		SessionID:     bkashResp.PaymentID,
		RedirectURL:   bkashResp.BkashURL,
		GatewayRef:    bkashResp.PaymentID,
		Status:        string(StatusPending),
	}, nil
}

func (g *BKash) VerifyPayment(ctx context.Context, transactionID string) (*PaymentStatus, error) {
	if err := g.ensureToken(ctx); err != nil {
		return nil, err
	}

	body, _ := json.Marshal(map[string]string{"paymentID": transactionID})

	httpReq, err := http.NewRequestWithContext(ctx, "POST", g.baseURL+"/tokenized/checkout/payment/status", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	g.setHeaders(httpReq)

	resp, err := g.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var statusResp bkashStatusResponse
	if err := json.Unmarshal(respBody, &statusResp); err != nil {
		return nil, err
	}

	return &PaymentStatus{
		TransactionID: transactionID,
		GatewayRef:    statusResp.TrxID,
		Status:        mapBKashStatus(statusResp.TransactionStatus),
		AmountPaisa:   int64(statusResp.Amount * 100),
		Currency:      "BDT",
	}, nil
}

func (g *BKash) CapturePayment(ctx context.Context, transactionID string) (*PaymentStatus, error) {
	if err := g.ensureToken(ctx); err != nil {
		return nil, err
	}

	body, _ := json.Marshal(map[string]string{"paymentID": transactionID})

	httpReq, err := http.NewRequestWithContext(ctx, "POST", g.baseURL+"/tokenized/checkout/execute", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	g.setHeaders(httpReq)

	resp, err := g.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var execResp bkashExecuteResponse
	if err := json.Unmarshal(respBody, &execResp); err != nil {
		return nil, err
	}

	if execResp.StatusCode != "0000" {
		return nil, fmt.Errorf("%w: %s", ErrGatewayError, execResp.StatusMessage)
	}

	return &PaymentStatus{
		TransactionID: transactionID,
		GatewayRef:    execResp.TrxID,
		Status:        StatusCaptured,
		AmountPaisa:   int64(execResp.Amount * 100),
		Currency:      "BDT",
	}, nil
}

func (g *BKash) RefundPayment(ctx context.Context, transactionID string, amountPaisa int64, reason string) (*RefundResponse, error) {
	if err := g.ensureToken(ctx); err != nil {
		return nil, err
	}

	refundReq := map[string]interface{}{
		"paymentID": transactionID,
		"amount":    fmt.Sprintf("%.2f", float64(amountPaisa)/100),
		"trxID":     transactionID,
		"sku":       "refund",
		"reason":    reason,
	}

	body, _ := json.Marshal(refundReq)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", g.baseURL+"/tokenized/checkout/payment/refund", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	g.setHeaders(httpReq)

	resp, err := g.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var refundResp bkashRefundResponse
	if err := json.Unmarshal(respBody, &refundResp); err != nil {
		return nil, err
	}

	if refundResp.StatusCode != "0000" {
		return nil, fmt.Errorf("%w: %s", ErrRefundFailed, refundResp.StatusMessage)
	}

	return &RefundResponse{
		RefundID:      refundResp.RefundTrxID,
		TransactionID: transactionID,
		AmountPaisa:   amountPaisa,
		Status:        "completed",
		Reason:        reason,
		ProcessedAt:   time.Now().Unix(),
	}, nil
}

func (g *BKash) ValidateIPN(ctx context.Context, payload []byte) (*IPNData, error) {
	var ipn bkashIPN
	if err := json.Unmarshal(payload, &ipn); err != nil {
		return nil, err
	}

	// Verify by querying status
	status, err := g.VerifyPayment(ctx, ipn.PaymentID)
	if err != nil {
		return nil, err
	}

	return &IPNData{
		TransactionID: ipn.PaymentID,
		OrderID:       ipn.MerchantInvoiceNumber,
		Status:        status.Status,
		AmountPaisa:   status.AmountPaisa,
		GatewayRef:    status.GatewayRef,
		IsValid:       true,
	}, nil
}

func (g *BKash) HealthCheck(ctx context.Context) error {
	return g.ensureToken(ctx)
}

// --- Token Management ---

func (g *BKash) ensureToken(ctx context.Context) error {
	if g.token != "" && time.Now().Before(g.tokenExpiry) {
		return nil
	}

	body, _ := json.Marshal(map[string]string{
		"app_key":    g.appKey,
		"app_secret": g.appSecret,
	})

	httpReq, err := http.NewRequestWithContext(ctx, "POST", g.baseURL+"/tokenized/checkout/token/grant", bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("username", g.username)
	httpReq.Header.Set("password", g.password)

	resp, err := g.client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var tokenResp bkashTokenResponse
	if err := json.Unmarshal(respBody, &tokenResp); err != nil {
		return err
	}

	if tokenResp.StatusCode != "0000" {
		return fmt.Errorf("token grant failed: %s", tokenResp.StatusMessage)
	}

	g.token = tokenResp.IDToken
	g.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	return nil
}

func (g *BKash) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", g.token)
	req.Header.Set("X-APP-Key", g.appKey)
}

func (g *BKash) generateSignature(payload string) string {
	hash := sha256.Sum256([]byte(payload + g.appSecret))
	return base64.StdEncoding.EncodeToString(hash[:])
}

func mapBKashStatus(status string) Status {
	switch status {
	case "Completed":
		return StatusCaptured
	case "Initiated", "Pending":
		return StatusPending
	case "Cancelled":
		return StatusCancelled
	default:
		return StatusFailed
	}
}

// --- bKash API Types ---

type bkashTokenResponse struct {
	StatusCode    string `json:"statusCode"`
	StatusMessage string `json:"statusMessage"`
	IDToken       string `json:"id_token"`
	TokenType     string `json:"token_type"`
	ExpiresIn     int    `json:"expires_in"`
}

type bkashCreateRequest struct {
	Mode                  string `json:"mode"`
	PayerReference        string `json:"payerReference"`
	CallbackURL           string `json:"callbackURL"`
	Amount                string `json:"amount"`
	Currency              string `json:"currency"`
	Intent                string `json:"intent"`
	MerchantInvoiceNumber string `json:"merchantInvoiceNumber"`
}

type bkashCreateResponse struct {
	StatusCode        string `json:"statusCode"`
	StatusMessage     string `json:"statusMessage"`
	PaymentID         string `json:"paymentID"`
	BkashURL          string `json:"bkashURL"`
	PaymentCreateTime string `json:"paymentCreateTime"`
	TransactionStatus string `json:"transactionStatus"`
	Amount            string `json:"amount"`
	Intent            string `json:"intent"`
	Currency          string `json:"currency"`
}

type bkashExecuteResponse struct {
	StatusCode        string  `json:"statusCode"`
	StatusMessage     string  `json:"statusMessage"`
	PaymentID         string  `json:"paymentID"`
	TrxID             string  `json:"trxID"`
	Amount            float64 `json:"amount,string"`
	TransactionStatus string  `json:"transactionStatus"`
}

type bkashStatusResponse struct {
	StatusCode        string  `json:"statusCode"`
	StatusMessage     string  `json:"statusMessage"`
	PaymentID         string  `json:"paymentID"`
	TrxID             string  `json:"trxID"`
	Amount            float64 `json:"amount,string"`
	TransactionStatus string  `json:"transactionStatus"`
}

type bkashRefundResponse struct {
	StatusCode    string `json:"statusCode"`
	StatusMessage string `json:"statusMessage"`
	RefundTrxID   string `json:"refundTrxID"`
}

type bkashIPN struct {
	PaymentID             string `json:"paymentID"`
	Status                string `json:"status"`
	Amount                string `json:"amount"`
	MerchantInvoiceNumber string `json:"merchantInvoiceNumber"`
	TrxID                 string `json:"trxID"`
}

var _ Gateway = (*BKash)(nil)
