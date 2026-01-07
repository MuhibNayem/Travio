package gateway

import (
	"encoding/json"
	"errors"
	"sync"
)

// Factory creates a Gateway instance from configuration
type Factory interface {
	Create(credentials json.RawMessage, isSandbox bool) (Gateway, error)
	ParseOrderID(payload map[string]string) (string, error)
}

// Registry manages payment gateway factories
type Registry struct {
	factories map[string]Factory
	mu        sync.RWMutex
}

func NewRegistry() *Registry {
	return &Registry{factories: make(map[string]Factory)}
}

func (r *Registry) Register(name string, factory Factory) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.factories[name] = factory
}

func (r *Registry) GetFactory(name string) (Factory, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if f, ok := r.factories[name]; ok {
		return f, nil
	}
	return nil, ErrGatewayNotFound
}

// Helper to map generic method names (card, mobile_bank) to specific providers
func (r *Registry) ResolveProvider(method string) string {
	methodMap := map[string]string{
		"card":        "sslcommerz",
		"bank":        "sslcommerz",
		"mobile_bank": "sslcommerz", // Default aggregator, often overridden
		"bkash":       "bkash",
		"nagad":       "nagad",
	}
	if provider, ok := methodMap[method]; ok {
		return provider
	}
	return method // Return as-is if no mapping
}

var ErrGatewayNotFound = errors.New("gateway factory not found")

// --- Factory Implementations ---

// SSLCommerzFactory
type SSLCommerzFactory struct{}

func (f *SSLCommerzFactory) Create(credentials json.RawMessage, isSandbox bool) (Gateway, error) {
	var cfg SSLCommerzConfig
	if err := json.Unmarshal(credentials, &cfg); err != nil {
		return nil, err
	}
	cfg.IsSandbox = isSandbox
	return NewSSLCommerz(cfg), nil
}

func (f *SSLCommerzFactory) ParseOrderID(payload map[string]string) (string, error) {
	if val, ok := payload["tran_id"]; ok {
		return val, nil
	}
	return "", errors.New("tran_id not found in payload")
}

// BKashFactory
type BKashFactory struct{}

func (f *BKashFactory) Create(credentials json.RawMessage, isSandbox bool) (Gateway, error) {
	var cfg BKashConfig
	if err := json.Unmarshal(credentials, &cfg); err != nil {
		return nil, err
	}
	cfg.IsSandbox = isSandbox
	return NewBKash(cfg), nil
}

func (f *BKashFactory) ParseOrderID(payload map[string]string) (string, error) {
	// bKash usually sends paymentID via callback, but merchantInvoiceNumber corresponds to OrderID
	if val, ok := payload["merchantInvoiceNumber"]; ok {
		return val, nil
	}
	return "", errors.New("merchantInvoiceNumber not found in payload")
}

// NagadFactory
type NagadFactory struct{}

func (f *NagadFactory) Create(credentials json.RawMessage, isSandbox bool) (Gateway, error) {
	var cfg NagadConfig
	if err := json.Unmarshal(credentials, &cfg); err != nil {
		return nil, err
	}
	cfg.IsSandbox = isSandbox
	return NewNagad(cfg), nil
}

func (f *NagadFactory) ParseOrderID(payload map[string]string) (string, error) {
	if val, ok := payload["order_id"]; ok {
		return val, nil
	}
	return "", errors.New("order_id not found in payload")
}
