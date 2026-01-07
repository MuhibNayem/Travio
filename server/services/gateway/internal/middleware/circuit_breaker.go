package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/sony/gobreaker"
)

// CircuitBreaker wraps gobreaker.CircuitBreaker
type CircuitBreaker struct {
	cb *gobreaker.CircuitBreaker
}

// NewCircuitBreaker creates a new circuit breaker with default settings
func NewCircuitBreaker(name string) *CircuitBreaker {
	settings := gobreaker.Settings{
		Name:        name,
		MaxRequests: 5,                // Max requests allowed in half-open state
		Interval:    60 * time.Second, // Cyclic period of closed state
		Timeout:     60 * time.Second, // Period of open state
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 5 && failureRatio >= 0.6
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			fmt.Printf("Circuit Breaker '%s' changed from %s to %s\n", name, from, to)
		},
	}
	return &CircuitBreaker{
		cb: gobreaker.NewCircuitBreaker(settings),
	}
}

// Execute wraps a function with the circuit breaker
func (cb *CircuitBreaker) Execute(req func() (interface{}, error)) (interface{}, error) {
	return cb.cb.Execute(req)
}

// HTTPClient returns an http.Client that uses the circuit breaker
func (cb *CircuitBreaker) HTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
		Transport: &cbTransport{
			cb:        cb.cb,
			transport: http.DefaultTransport,
		},
	}
}

// cbTransport implements http.RoundTripper using the circuit breaker
type cbTransport struct {
	cb        *gobreaker.CircuitBreaker
	transport http.RoundTripper
}

func (t *cbTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.cb.Execute(func() (interface{}, error) {
		resp, err := t.transport.RoundTrip(req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode >= 500 {
			return nil, fmt.Errorf("upstream error: %d", resp.StatusCode)
		}
		return resp, nil
	})

	if err != nil {
		return nil, err
	}
	return resp.(*http.Response), nil
}
