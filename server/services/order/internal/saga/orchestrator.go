package saga

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Orchestrator implements the Saga Orchestration pattern
// It coordinates distributed transactions across multiple services
// and handles compensating transactions on failures
type Orchestrator struct {
	mu        sync.Mutex
	sagas     map[string]*Saga
	listeners []StatusListener
}

func NewOrchestrator() *Orchestrator {
	return &Orchestrator{
		sagas: make(map[string]*Saga),
	}
}

// Saga represents a distributed transaction
type Saga struct {
	ID            string       `json:"id"`
	Name          string       `json:"name"`
	Status        Status       `json:"status"`
	CurrentStep   int          `json:"current_step"`
	Steps         []*Step      `json:"steps"`
	Context       *SagaContext `json:"context"`
	StartedAt     time.Time    `json:"started_at"`
	CompletedAt   time.Time    `json:"completed_at,omitempty"`
	FailureReason string       `json:"failure_reason,omitempty"`
	mu            sync.Mutex
}

// Status represents the saga lifecycle state
type Status string

const (
	StatusPending      Status = "pending"
	StatusRunning      Status = "running"
	StatusCompleted    Status = "completed"
	StatusCompensating Status = "compensating"
	StatusCompensated  Status = "compensated"
	StatusFailed       Status = "failed"
)

// Step represents a single step in the saga
type Step struct {
	Name         string    `json:"name"`
	Status       Status    `json:"status"`
	Error        string    `json:"error,omitempty"`
	StartedAt    time.Time `json:"started_at,omitempty"`
	CompletedAt  time.Time `json:"completed_at,omitempty"`
	Compensated  bool      `json:"compensated"`
	ExecuteFn    StepFunc  `json:"-"`
	CompensateFn StepFunc  `json:"-"`
}

// StepFunc is the function signature for step execution
type StepFunc func(ctx context.Context, sagaCtx *SagaContext) error

// SagaContext holds data shared across saga steps
type SagaContext struct {
	data map[string]interface{}
	mu   sync.RWMutex
}

func NewSagaContext() *SagaContext {
	return &SagaContext{
		data: make(map[string]interface{}),
	}
}

func (c *SagaContext) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
}

func (c *SagaContext) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.data[key]
	return v, ok
}

func (c *SagaContext) GetString(key string) string {
	v, ok := c.Get(key)
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
}

func (c *SagaContext) GetInt64(key string) int64 {
	v, ok := c.Get(key)
	if !ok {
		return 0
	}
	i, _ := v.(int64)
	return i
}

// StatusListener receives saga status updates
type StatusListener func(saga *Saga, step *Step, event string)

// RegisterListener adds a status listener
func (o *Orchestrator) RegisterListener(listener StatusListener) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.listeners = append(o.listeners, listener)
}

func (o *Orchestrator) notify(saga *Saga, step *Step, event string) {
	for _, l := range o.listeners {
		go l(saga, step, event)
	}
}

// CreateSaga creates a new saga with the given steps
func (o *Orchestrator) CreateSaga(name string, steps []*Step) *Saga {
	saga := &Saga{
		ID:        uuid.New().String(),
		Name:      name,
		Status:    StatusPending,
		Steps:     steps,
		Context:   NewSagaContext(),
		StartedAt: time.Now(),
	}

	for _, step := range steps {
		step.Status = StatusPending
	}

	o.mu.Lock()
	o.sagas[saga.ID] = saga
	o.mu.Unlock()

	return saga
}

// Execute runs the saga to completion
func (o *Orchestrator) Execute(ctx context.Context, saga *Saga) error {
	saga.mu.Lock()
	saga.Status = StatusRunning
	saga.mu.Unlock()
	o.notify(saga, nil, "saga_started")

	for i, step := range saga.Steps {
		saga.mu.Lock()
		saga.CurrentStep = i
		step.Status = StatusRunning
		step.StartedAt = time.Now()
		saga.mu.Unlock()

		o.notify(saga, step, "step_started")

		// Execute the step
		err := step.ExecuteFn(ctx, saga.Context)

		saga.mu.Lock()
		step.CompletedAt = time.Now()

		if err != nil {
			step.Status = StatusFailed
			step.Error = err.Error()
			saga.FailureReason = fmt.Sprintf("step '%s' failed: %s", step.Name, err.Error())
			saga.mu.Unlock()

			o.notify(saga, step, "step_failed")

			// Start compensation
			return o.compensate(ctx, saga, i)
		}

		step.Status = StatusCompleted
		saga.mu.Unlock()
		o.notify(saga, step, "step_completed")
	}

	// All steps completed successfully
	saga.mu.Lock()
	saga.Status = StatusCompleted
	saga.CompletedAt = time.Now()
	saga.mu.Unlock()
	o.notify(saga, nil, "saga_completed")

	return nil
}

// compensate runs compensation for all completed steps in reverse order
func (o *Orchestrator) compensate(ctx context.Context, saga *Saga, failedStepIdx int) error {
	saga.mu.Lock()
	saga.Status = StatusCompensating
	saga.mu.Unlock()
	o.notify(saga, nil, "compensation_started")

	var compensationErrors []error

	// Compensate in reverse order, starting from the step before the failed one
	for i := failedStepIdx - 1; i >= 0; i-- {
		step := saga.Steps[i]

		if step.CompensateFn == nil {
			continue // No compensation defined
		}

		o.notify(saga, step, "compensating_step")

		err := step.CompensateFn(ctx, saga.Context)

		saga.mu.Lock()
		if err != nil {
			compensationErrors = append(compensationErrors, fmt.Errorf("compensation for '%s' failed: %w", step.Name, err))
			// Continue compensating other steps even if one fails
		} else {
			step.Compensated = true
			step.Status = StatusCompensated
		}
		saga.mu.Unlock()

		o.notify(saga, step, "step_compensated")
	}

	saga.mu.Lock()
	if len(compensationErrors) > 0 {
		saga.Status = StatusFailed
		saga.FailureReason = fmt.Sprintf("%s; compensation errors: %v", saga.FailureReason, compensationErrors)
	} else {
		saga.Status = StatusCompensated
	}
	saga.CompletedAt = time.Now()
	saga.mu.Unlock()

	o.notify(saga, nil, "saga_compensated")

	if len(compensationErrors) > 0 {
		return errors.Join(compensationErrors...)
	}

	return errors.New(saga.FailureReason)
}

// GetSaga retrieves a saga by ID
func (o *Orchestrator) GetSaga(id string) (*Saga, bool) {
	o.mu.Lock()
	defer o.mu.Unlock()
	saga, ok := o.sagas[id]
	return saga, ok
}

// Retry attempts to resume a failed saga
func (o *Orchestrator) Retry(ctx context.Context, sagaID string) error {
	saga, ok := o.GetSaga(sagaID)
	if !ok {
		return ErrSagaNotFound
	}

	if saga.Status != StatusFailed && saga.Status != StatusCompensated {
		return ErrSagaNotRetryable
	}

	// Reset from the failed step
	saga.mu.Lock()
	for i := saga.CurrentStep; i < len(saga.Steps); i++ {
		saga.Steps[i].Status = StatusPending
		saga.Steps[i].Error = ""
		saga.Steps[i].Compensated = false
	}
	saga.Status = StatusPending
	saga.FailureReason = ""
	saga.mu.Unlock()

	return o.Execute(ctx, saga)
}

var (
	ErrSagaNotFound     = errors.New("saga not found")
	ErrSagaNotRetryable = errors.New("saga is not in a retryable state")
)
