package gateway

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// SSLCommerz implements Gateway for SSLCommerz (Bangladesh's largest payment gateway)
// API Docs: https://developer.sslcommerz.com/doc/v4/
type SSLCommerz struct {
	storeID     string
	storePasswd string
	baseURL     string
	client      *http.Client
	isSandbox   bool
}

type SSLCommerzConfig struct {
	StoreID     string
	StorePasswd string
	IsSandbox   bool
	Timeout     time.Duration
}

func NewSSLCommerz(cfg SSLCommerzConfig) *SSLCommerz {
	baseURL := "https://securepay.sslcommerz.com"
	if cfg.IsSandbox {
		baseURL = "https://sandbox.sslcommerz.com"
	}

	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &SSLCommerz{
		storeID:     cfg.StoreID,
		storePasswd: cfg.StorePasswd,
		baseURL:     baseURL,
		client:      &http.Client{Timeout: timeout},
		isSandbox:   cfg.IsSandbox,
	}
}

func (g *SSLCommerz) Name() string {
	return "sslcommerz"
}

func (g *SSLCommerz) CreatePayment(ctx context.Context, req *CreatePaymentRequest) (*CreatePaymentResponse, error) {
	// Build SSLCommerz request
	formData := url.Values{}
	formData.Set("store_id", g.storeID)
	formData.Set("store_passwd", g.storePasswd)
	formData.Set("tran_id", req.OrderID)
	formData.Set("total_amount", fmt.Sprintf("%.2f", req.Amount.AmountTaka()))
	formData.Set("currency", req.Currency)
	formData.Set("success_url", req.ReturnURL)
	formData.Set("fail_url", req.CancelURL)
	formData.Set("cancel_url", req.CancelURL)
	formData.Set("ipn_url", req.IPNURL)
	formData.Set("cus_name", req.CustomerName)
	formData.Set("cus_email", req.CustomerEmail)
	formData.Set("cus_phone", req.CustomerPhone)
	formData.Set("product_name", req.Description)
	formData.Set("product_category", "transportation")
	formData.Set("product_profile", "general")
	formData.Set("shipping_method", "NO")
	formData.Set("num_of_item", "1")

	// Optional: EMI settings
	formData.Set("emi_option", "0")
	formData.Set("emi_max_inst_option", "0")

	// Shipping address (required but can be empty for digital goods)
	formData.Set("ship_name", req.CustomerName)
	formData.Set("ship_add1", "N/A")
	formData.Set("ship_city", "Dhaka")
	formData.Set("ship_country", "Bangladesh")

	// Customer address
	formData.Set("cus_add1", "N/A")
	formData.Set("cus_city", "Dhaka")
	formData.Set("cus_country", "Bangladesh")

	// API call
	httpReq, err := http.NewRequestWithContext(ctx, "POST", g.baseURL+"/gwprocess/v4/api.php", strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := g.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var sslResp sslCommerzInitResponse
	if err := json.Unmarshal(body, &sslResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if sslResp.Status != "SUCCESS" {
		return nil, fmt.Errorf("%w: %s", ErrGatewayError, sslResp.FailedReason)
	}

	return &CreatePaymentResponse{
		TransactionID: req.OrderID,
		SessionID:     sslResp.SessionKey,
		RedirectURL:   sslResp.GatewayPageURL,
		GatewayRef:    sslResp.SessionKey,
		Status:        string(StatusPending),
	}, nil
}

func (g *SSLCommerz) VerifyPayment(ctx context.Context, transactionID string) (*PaymentStatus, error) {
	// Use transaction validation API
	formData := url.Values{}
	formData.Set("store_id", g.storeID)
	formData.Set("store_passwd", g.storePasswd)
	formData.Set("tran_id", transactionID)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", g.baseURL+"/validator/api/merchantTransIDvalidationAPI.php", strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := g.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("validation request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var validationResp sslCommerzValidationResponse
	if err := json.Unmarshal(body, &validationResp); err != nil {
		return nil, fmt.Errorf("failed to parse validation response: %w", err)
	}

	return &PaymentStatus{
		TransactionID: transactionID,
		GatewayRef:    validationResp.ValID,
		Status:        mapSSLCommerzStatus(validationResp.Status),
		AmountPaisa:   int64(validationResp.Amount * 100),
		Currency:      validationResp.Currency,
		CardBrand:     validationResp.CardBrand,
		CardLast4:     validationResp.CardNo,
		BankTranID:    validationResp.BankTranID,
		RiskLevel:     validationResp.RiskLevel,
	}, nil
}

func (g *SSLCommerz) CapturePayment(ctx context.Context, transactionID string) (*PaymentStatus, error) {
	// SSLCommerz is direct capture (not auth-then-capture)
	return g.VerifyPayment(ctx, transactionID)
}

func (g *SSLCommerz) RefundPayment(ctx context.Context, transactionID string, amountPaisa int64, reason string) (*RefundResponse, error) {
	formData := url.Values{}
	formData.Set("store_id", g.storeID)
	formData.Set("store_passwd", g.storePasswd)
	formData.Set("refund_amount", fmt.Sprintf("%.2f", float64(amountPaisa)/100))
	formData.Set("refund_remarks", reason)
	formData.Set("bank_tran_id", transactionID) // Use bank_tran_id for refund

	httpReq, err := http.NewRequestWithContext(ctx, "POST", g.baseURL+"/validator/api/merchantTransIDvalidationAPI.php", strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := g.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("refund request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var refundResp sslCommerzRefundResponse
	if err := json.Unmarshal(body, &refundResp); err != nil {
		return nil, err
	}

	if refundResp.Status != "success" {
		return nil, fmt.Errorf("%w: %s", ErrRefundFailed, refundResp.ErrorReason)
	}

	return &RefundResponse{
		RefundID:      refundResp.RefundRefID,
		TransactionID: transactionID,
		AmountPaisa:   amountPaisa,
		Status:        "completed",
		Reason:        reason,
		ProcessedAt:   time.Now().Unix(),
	}, nil
}

func (g *SSLCommerz) ValidateIPN(ctx context.Context, payload []byte) (*IPNData, error) {
	var ipn sslCommerzIPN
	if err := json.Unmarshal(payload, &ipn); err != nil {
		return nil, fmt.Errorf("failed to parse IPN: %w", err)
	}

	// Validate signature using verify_sign
	expectedSign := g.generateVerifySign(ipn.TranID, ipn.ValID, ipn.Amount, ipn.StoreAmount)

	if ipn.VerifySign != expectedSign {
		return nil, ErrIPNValidationFailed
	}

	return &IPNData{
		TransactionID: ipn.TranID,
		OrderID:       ipn.TranID,
		Status:        mapSSLCommerzStatus(ipn.Status),
		AmountPaisa:   int64(ipn.Amount * 100),
		GatewayRef:    ipn.ValID,
		BankTranID:    ipn.BankTranID,
		CardType:      ipn.CardType,
		RiskLevel:     ipn.RiskLevel,
		IsValid:       true,
	}, nil
}

func (g *SSLCommerz) HealthCheck(ctx context.Context) error {
	// Simple connectivity check
	req, _ := http.NewRequestWithContext(ctx, "GET", g.baseURL, nil)
	resp, err := g.client.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func (g *SSLCommerz) generateVerifySign(tranID, valID string, amount, storeAmount float64) string {
	data := fmt.Sprintf("%s%s%.2f%.2f%s", tranID, valID, amount, storeAmount, g.storePasswd)
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}

func mapSSLCommerzStatus(status string) Status {
	switch strings.ToUpper(status) {
	case "VALID", "VALIDATED":
		return StatusCaptured
	case "PENDING":
		return StatusPending
	case "FAILED":
		return StatusFailed
	case "CANCELLED":
		return StatusCancelled
	default:
		return StatusPending
	}
}

// --- SSLCommerz API Response Types ---

type sslCommerzInitResponse struct {
	Status         string `json:"status"`
	FailedReason   string `json:"failedreason"`
	SessionKey     string `json:"sessionkey"`
	GatewayPageURL string `json:"GatewayPageURL"`
	RedirectURL    string `json:"redirectGatewayURL"`
	StoreBanner    string `json:"storeBanner"`
	StoreLogo      string `json:"storeLogo"`
}

type sslCommerzValidationResponse struct {
	Status      string  `json:"status"`
	TranDate    string  `json:"tran_date"`
	TranID      string  `json:"tran_id"`
	ValID       string  `json:"val_id"`
	Amount      float64 `json:"amount,string"`
	StoreAmount float64 `json:"store_amount,string"`
	Currency    string  `json:"currency"`
	BankTranID  string  `json:"bank_tran_id"`
	CardType    string  `json:"card_type"`
	CardNo      string  `json:"card_no"`
	CardIssuer  string  `json:"card_issuer"`
	CardBrand   string  `json:"card_brand"`
	RiskLevel   string  `json:"risk_level"`
	RiskTitle   string  `json:"risk_title"`
	ValidatedOn string  `json:"validated_on"`
}

type sslCommerzRefundResponse struct {
	Status      string `json:"status"`
	RefundRefID string `json:"refund_ref_id"`
	ErrorReason string `json:"errorReason"`
}

type sslCommerzIPN struct {
	TranID      string  `json:"tran_id"`
	ValID       string  `json:"val_id"`
	Amount      float64 `json:"amount,string"`
	StoreAmount float64 `json:"store_amount,string"`
	Currency    string  `json:"currency"`
	BankTranID  string  `json:"bank_tran_id"`
	Status      string  `json:"status"`
	CardType    string  `json:"card_type"`
	CardNo      string  `json:"card_no"`
	CardIssuer  string  `json:"card_issuer"`
	CardBrand   string  `json:"card_brand"`
	RiskLevel   string  `json:"risk_level"`
	VerifySign  string  `json:"verify_sign"`
	VerifyKey   string  `json:"verify_key"`
}

// Ensure SSLCommerz implements Gateway
var _ Gateway = (*SSLCommerz)(nil)
