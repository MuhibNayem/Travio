package gateway

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Nagad implements Gateway for Nagad Payment Gateway (digital wallet by BD Post)
// API Docs: https://developer.nagad.com.bd
type Nagad struct {
	merchantID     string
	merchantNumber string
	publicKey      string
	privateKey     string
	baseURL        string
	client         *http.Client
	isSandbox      bool
}

type NagadConfig struct {
	MerchantID     string
	MerchantNumber string
	PublicKey      string // Nagad's public key for encryption
	PrivateKey     string // Merchant's private key for signing
	IsSandbox      bool
	Timeout        time.Duration
}

func NewNagad(cfg NagadConfig) *Nagad {
	baseURL := "https://api.mynagad.com/api/dfs"
	if cfg.IsSandbox {
		baseURL = "https://sandbox.mynagad.com/api/dfs"
	}

	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &Nagad{
		merchantID:     cfg.MerchantID,
		merchantNumber: cfg.MerchantNumber,
		publicKey:      cfg.PublicKey,
		privateKey:     cfg.PrivateKey,
		baseURL:        baseURL,
		client:         &http.Client{Timeout: timeout},
		isSandbox:      cfg.IsSandbox,
	}
}

func (g *Nagad) Name() string {
	return "nagad"
}

func (g *Nagad) CreatePayment(ctx context.Context, req *CreatePaymentRequest) (*CreatePaymentResponse, error) {
	// Step 1: Initialize checkout
	dateTime := time.Now().Format("20060102150405")
	orderId := req.OrderID

	initData := nagadInitRequest{
		MerchantID: g.merchantID,
		OrderID:    orderId,
		DateTime:   dateTime,
		Challenge:  g.generateChallenge(),
	}

	// Sign the sensitive data
	sensitiveData := map[string]string{
		"merchantId": g.merchantID,
		"datetime":   dateTime,
		"orderId":    orderId,
		"challenge":  initData.Challenge,
	}
	sensitiveJSON, _ := json.Marshal(sensitiveData)
	signature := g.signData(sensitiveJSON)

	initData.SensitiveData = g.encryptData(sensitiveJSON)
	initData.Signature = signature

	body, _ := json.Marshal(initData)

	initURL := fmt.Sprintf("%s/check-out/initialize/%s/%s", g.baseURL, g.merchantID, orderId)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", initURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-KM-IP-V4", "127.0.0.1")
	httpReq.Header.Set("X-KM-MC-Id", g.merchantID)
	httpReq.Header.Set("X-KM-Api-Version", "v-0.2.0")

	resp, err := g.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("init request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var initResp nagadInitResponse
	if err := json.Unmarshal(respBody, &initResp); err != nil {
		return nil, err
	}

	if initResp.Status != "Success" {
		return nil, fmt.Errorf("%w: %s", ErrGatewayError, initResp.Message)
	}

	// Step 2: Complete checkout with payment details
	paymentData := map[string]interface{}{
		"merchantId":   g.merchantID,
		"orderId":      orderId,
		"currencyCode": "050", // BDT
		"amount":       fmt.Sprintf("%.2f", req.Amount.AmountTaka()),
		"challenge":    initResp.Challenge,
	}
	paymentJSON, _ := json.Marshal(paymentData)

	completeReq := nagadCompleteRequest{
		SensitiveData:       g.encryptData(paymentJSON),
		Signature:           g.signData(paymentJSON),
		MerchantCallbackURL: req.ReturnURL,
		AdditionalMerchantInfo: map[string]string{
			"customer_phone": req.CustomerPhone,
			"customer_email": req.CustomerEmail,
		},
	}

	completeBody, _ := json.Marshal(completeReq)
	completeURL := fmt.Sprintf("%s/check-out/complete/%s", g.baseURL, initResp.PaymentReferenceID)

	completeHttpReq, _ := http.NewRequestWithContext(ctx, "POST", completeURL, bytes.NewReader(completeBody))
	completeHttpReq.Header.Set("Content-Type", "application/json")
	completeHttpReq.Header.Set("X-KM-IP-V4", "127.0.0.1")
	completeHttpReq.Header.Set("X-KM-MC-Id", g.merchantID)
	completeHttpReq.Header.Set("X-KM-Api-Version", "v-0.2.0")

	completeResp, err := g.client.Do(completeHttpReq)
	if err != nil {
		return nil, err
	}
	defer completeResp.Body.Close()

	completeRespBody, _ := io.ReadAll(completeResp.Body)

	var checkout nagadCheckoutResponse
	if err := json.Unmarshal(completeRespBody, &checkout); err != nil {
		return nil, err
	}

	if checkout.Status != "Success" {
		return nil, fmt.Errorf("%w: %s", ErrGatewayError, checkout.Message)
	}

	return &CreatePaymentResponse{
		TransactionID: orderId,
		SessionID:     checkout.PaymentReferenceID,
		RedirectURL:   checkout.CallBackURL,
		GatewayRef:    checkout.PaymentReferenceID,
		Status:        string(StatusPending),
	}, nil
}

func (g *Nagad) VerifyPayment(ctx context.Context, transactionID string) (*PaymentStatus, error) {
	verifyURL := fmt.Sprintf("%s/verify/payment/%s", g.baseURL, transactionID)

	httpReq, _ := http.NewRequestWithContext(ctx, "GET", verifyURL, nil)
	httpReq.Header.Set("X-KM-IP-V4", "127.0.0.1")
	httpReq.Header.Set("X-KM-MC-Id", g.merchantID)
	httpReq.Header.Set("X-KM-Api-Version", "v-0.2.0")

	resp, err := g.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var verifyResp nagadVerifyResponse
	if err := json.Unmarshal(respBody, &verifyResp); err != nil {
		return nil, err
	}

	return &PaymentStatus{
		TransactionID: transactionID,
		GatewayRef:    verifyResp.PaymentRefID,
		Status:        mapNagadStatus(verifyResp.Status),
		AmountPaisa:   int64(verifyResp.Amount * 100),
		Currency:      "BDT",
	}, nil
}

func (g *Nagad) CapturePayment(ctx context.Context, transactionID string) (*PaymentStatus, error) {
	// Nagad is direct capture
	return g.VerifyPayment(ctx, transactionID)
}

func (g *Nagad) RefundPayment(ctx context.Context, transactionID string, amountPaisa int64, reason string) (*RefundResponse, error) {
	refundReq := map[string]interface{}{
		"originalTrxId": transactionID,
		"amount":        fmt.Sprintf("%.2f", float64(amountPaisa)/100),
		"reason":        reason,
		"reference":     fmt.Sprintf("REF-%s-%d", transactionID, time.Now().Unix()),
	}

	body, _ := json.Marshal(refundReq)

	httpReq, _ := http.NewRequestWithContext(ctx, "POST", g.baseURL+"/payment/refund", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-KM-IP-V4", "127.0.0.1")
	httpReq.Header.Set("X-KM-MC-Id", g.merchantID)
	httpReq.Header.Set("X-KM-Api-Version", "v-0.2.0")

	resp, err := g.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var refundResp nagadRefundResponse
	if err := json.Unmarshal(respBody, &refundResp); err != nil {
		return nil, err
	}

	if refundResp.Status != "Success" {
		return nil, fmt.Errorf("%w: %s", ErrRefundFailed, refundResp.Message)
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

func (g *Nagad) ValidateIPN(ctx context.Context, payload []byte) (*IPNData, error) {
	var ipn nagadIPN
	if err := json.Unmarshal(payload, &ipn); err != nil {
		return nil, err
	}

	// Verify from Nagad API
	status, err := g.VerifyPayment(ctx, ipn.PaymentRefID)
	if err != nil {
		return nil, err
	}

	return &IPNData{
		TransactionID: ipn.PaymentRefID,
		OrderID:       ipn.OrderID,
		Status:        status.Status,
		AmountPaisa:   status.AmountPaisa,
		GatewayRef:    ipn.PaymentRefID,
		IsValid:       true,
	}, nil
}

func (g *Nagad) HealthCheck(ctx context.Context) error {
	req, _ := http.NewRequestWithContext(ctx, "GET", g.baseURL+"/health", nil)
	resp, err := g.client.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

// --- Crypto helpers (simplified - in production use proper RSA) ---

func (g *Nagad) generateChallenge() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func (g *Nagad) encryptData(data []byte) string {
	// In production: RSA encrypt with Nagad's public key
	// For now, return base64 (placeholder)
	return hex.EncodeToString(data)
}

func (g *Nagad) signData(data []byte) string {
	// HMAC-SHA256 signature
	h := hmac.New(sha256.New, []byte(g.privateKey))
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

func mapNagadStatus(status string) Status {
	switch status {
	case "Success":
		return StatusCaptured
	case "Pending":
		return StatusPending
	case "Failed":
		return StatusFailed
	case "Cancelled":
		return StatusCancelled
	default:
		return StatusFailed
	}
}

// --- Nagad API Types ---

type nagadInitRequest struct {
	MerchantID    string `json:"merchantId"`
	OrderID       string `json:"orderId"`
	DateTime      string `json:"datetime"`
	Challenge     string `json:"challenge"`
	SensitiveData string `json:"sensitiveData"`
	Signature     string `json:"signature"`
}

type nagadInitResponse struct {
	Status             string `json:"status"`
	Message            string `json:"message"`
	SensitiveData      string `json:"sensitiveData"`
	Signature          string `json:"signature"`
	Challenge          string `json:"challenge"`
	PaymentReferenceID string `json:"paymentReferenceId"`
}

type nagadCompleteRequest struct {
	SensitiveData          string            `json:"sensitiveData"`
	Signature              string            `json:"signature"`
	MerchantCallbackURL    string            `json:"merchantCallbackURL"`
	AdditionalMerchantInfo map[string]string `json:"additionalMerchantInfo"`
}

type nagadCheckoutResponse struct {
	Status             string `json:"status"`
	Message            string `json:"message"`
	PaymentReferenceID string `json:"paymentReferenceId"`
	CallBackURL        string `json:"callBackUrl"`
}

type nagadVerifyResponse struct {
	Status       string  `json:"status"`
	Message      string  `json:"message"`
	PaymentRefID string  `json:"paymentRefId"`
	Amount       float64 `json:"amount,string"`
	OrderID      string  `json:"orderId"`
}

type nagadRefundResponse struct {
	Status      string `json:"status"`
	Message     string `json:"message"`
	RefundTrxID string `json:"refundTrxId"`
}

type nagadIPN struct {
	PaymentRefID string `json:"paymentRefId"`
	OrderID      string `json:"orderId"`
	Status       string `json:"status"`
	Amount       string `json:"amount"`
}

var _ Gateway = (*Nagad)(nil)
