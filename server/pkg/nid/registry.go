package nid

import (
	"context"
	"errors"
	"sync"
)

// Registry manages multiple NID providers (Factory + Registry Pattern)
// Allows runtime provider selection based on country
type Registry struct {
	providers map[string]Provider // country code -> provider
	mu        sync.RWMutex
	fallback  Provider
}

func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]Provider),
	}
}

// Register adds a provider for a country
func (r *Registry) Register(country string, provider Provider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[country] = provider
}

// SetFallback sets a fallback provider for unknown countries
func (r *Registry) SetFallback(provider Provider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.fallback = provider
}

// Get returns the provider for a country
func (r *Registry) Get(country string) (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if provider, ok := r.providers[country]; ok {
		return provider, nil
	}

	if r.fallback != nil {
		return r.fallback, nil
	}

	return nil, ErrProviderNotFound
}

// GetAll returns all registered providers
func (r *Registry) GetAll() map[string]Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]Provider, len(r.providers))
	for k, v := range r.providers {
		result[k] = v
	}
	return result
}

// HealthCheckAll checks health of all providers
func (r *Registry) HealthCheckAll(ctx context.Context) map[string]error {
	providers := r.GetAll()
	results := make(map[string]error, len(providers))

	var wg sync.WaitGroup
	var mu sync.Mutex

	for country, provider := range providers {
		wg.Add(1)
		go func(c string, p Provider) {
			defer wg.Done()
			err := p.HealthCheck(ctx)
			mu.Lock()
			results[c] = err
			mu.Unlock()
		}(country, provider)
	}

	wg.Wait()
	return results
}

var ErrProviderNotFound = errors.New("no provider registered for country")

// --- Service ---

// Service provides high-level NID verification operations
type Service struct {
	registry *Registry
}

func NewService(registry *Registry) *Service {
	return &Service{registry: registry}
}

// Verify verifies an NID using the appropriate country provider
func (s *Service) Verify(ctx context.Context, country string, req *VerifyRequest) (*VerifyResponse, error) {
	provider, err := s.registry.Get(country)
	if err != nil {
		return nil, err
	}
	return provider.Verify(ctx, req)
}

// AutoDetectCountry attempts to detect country from NID format
func (s *Service) AutoDetectCountry(nidStr string) string {
	// Bangladesh: 10 or 17 digits
	if len(nidStr) == 10 || len(nidStr) == 17 {
		valid := true
		for _, c := range nidStr {
			if c < '0' || c > '9' {
				valid = false
				break
			}
		}
		if valid {
			return "BD"
		}
	}

	// India Aadhaar: 12 digits, doesn't start with 0 or 1
	if len(nidStr) == 12 {
		valid := true
		for _, c := range nidStr {
			if c < '0' || c > '9' {
				valid = false
				break
			}
		}
		if valid && nidStr[0] != '0' && nidStr[0] != '1' {
			return "IN"
		}
	}

	return ""
}

// VerifyAuto verifies with auto-detected country
func (s *Service) VerifyAuto(ctx context.Context, req *VerifyRequest) (*VerifyResponse, error) {
	country := s.AutoDetectCountry(req.NID)
	if country == "" {
		return &VerifyResponse{
			IsValid:      false,
			ErrorCode:    ErrorCodeInvalidFormat,
			ErrorMessage: "could not detect country from NID format",
		}, nil
	}
	return s.Verify(ctx, country, req)
}

// HealthCheck returns health status of all providers
func (s *Service) HealthCheck(ctx context.Context) map[string]error {
	return s.registry.HealthCheckAll(ctx)
}
