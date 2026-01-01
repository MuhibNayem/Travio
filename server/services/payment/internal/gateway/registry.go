package gateway

import (
	"errors"
	"sync"
)

// Registry manages multiple payment gateways
type Registry struct {
	gateways map[string]Gateway
	mu       sync.RWMutex
	fallback Gateway
}

func NewRegistry() *Registry {
	return &Registry{gateways: make(map[string]Gateway)}
}

func (r *Registry) Register(name string, gateway Gateway) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.gateways[name] = gateway
}

func (r *Registry) SetFallback(gateway Gateway) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.fallback = gateway
}

func (r *Registry) Get(name string) (Gateway, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if gw, ok := r.gateways[name]; ok {
		return gw, nil
	}
	if r.fallback != nil {
		return r.fallback, nil
	}
	return nil, ErrGatewayNotFound
}

func (r *Registry) SelectByMethod(method string) (Gateway, error) {
	methodMap := map[string]string{
		"card": "sslcommerz", "bkash": "bkash", "nagad": "nagad",
		"bank": "sslcommerz", "mobile_bank": "sslcommerz",
	}
	if name, ok := methodMap[method]; ok {
		return r.Get(name)
	}
	if r.fallback != nil {
		return r.fallback, nil
	}
	return nil, ErrPaymentMethodNotSupported
}

var ErrGatewayNotFound = errors.New("gateway not found")
var ErrPaymentMethodNotSupported = errors.New("payment method not supported")
