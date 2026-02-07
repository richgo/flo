// Package agent provides retry logic with exponential backoff and circuit breaker.
package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/richgo/flo/pkg/task"
)

// RetryConfig configures retry behavior.
type RetryConfig struct {
	MaxRetries     int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
	BackoffFactor  float64
	// Circuit breaker settings
	FailureThreshold int
	ResetTimeout     time.Duration
}

// DefaultRetryConfig returns sensible defaults.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:       3,
		InitialBackoff:   time.Second,
		MaxBackoff:       30 * time.Second,
		BackoffFactor:    2.0,
		FailureThreshold: 5,
		ResetTimeout:     60 * time.Second,
	}
}

// CircuitState represents the state of the circuit breaker.
type CircuitState int

const (
	CircuitClosed CircuitState = iota
	CircuitOpen
	CircuitHalfOpen
)

// CircuitBreaker implements the circuit breaker pattern.
type CircuitBreaker struct {
	mu               sync.Mutex
	state            CircuitState
	failures         int
	lastFailureTime  time.Time
	failureThreshold int
	resetTimeout     time.Duration
}

// NewCircuitBreaker creates a new circuit breaker.
func NewCircuitBreaker(failureThreshold int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:            CircuitClosed,
		failures:         0,
		failureThreshold: failureThreshold,
		resetTimeout:     resetTimeout,
	}
}

// Call executes a function through the circuit breaker.
func (cb *CircuitBreaker) Call(fn func() error) error {
	cb.mu.Lock()

	// Check if circuit should transition from open to half-open
	if cb.state == CircuitOpen {
		if time.Since(cb.lastFailureTime) > cb.resetTimeout {
			cb.state = CircuitHalfOpen
			cb.failures = 0
		} else {
			cb.mu.Unlock()
			return fmt.Errorf("circuit breaker is open")
		}
	}

	cb.mu.Unlock()

	// Execute the function
	err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failures++
		cb.lastFailureTime = time.Now()

		if cb.failures >= cb.failureThreshold {
			cb.state = CircuitOpen
		}
		return err
	}

	// Success - reset circuit
	if cb.state == CircuitHalfOpen {
		cb.state = CircuitClosed
	}
	cb.failures = 0
	return nil
}

// State returns the current circuit state.
func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

// Reset resets the circuit breaker to closed state.
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.state = CircuitClosed
	cb.failures = 0
}

// RetryableBackend wraps a Backend with retry logic and circuit breaker.
type RetryableBackend struct {
	backend        Backend
	config         RetryConfig
	circuitBreaker *CircuitBreaker
}

// NewRetryableBackend wraps a backend with retry capabilities.
func NewRetryableBackend(backend Backend, config RetryConfig) *RetryableBackend {
	return &RetryableBackend{
		backend: backend,
		config:  config,
		circuitBreaker: NewCircuitBreaker(
			config.FailureThreshold,
			config.ResetTimeout,
		),
	}
}

// Name returns the backend name.
func (r *RetryableBackend) Name() string {
	return r.backend.Name()
}

// Start starts the backend with retry.
func (r *RetryableBackend) Start(ctx context.Context) error {
	return r.retryWithBackoff(ctx, func() error {
		return r.backend.Start(ctx)
	})
}

// Stop stops the backend.
func (r *RetryableBackend) Stop() error {
	return r.backend.Stop()
}

// CreateSession creates a session with retry.
func (r *RetryableBackend) CreateSession(ctx context.Context, t *task.Task, worktree string) (Session, error) {
	var session Session
	err := r.retryWithBackoff(ctx, func() error {
		var err error
		session, err = r.backend.CreateSession(ctx, t, worktree)
		return err
	})
	return session, err
}

// retryWithBackoff implements exponential backoff retry logic.
func (r *RetryableBackend) retryWithBackoff(ctx context.Context, fn func() error) error {
	var lastErr error
	backoff := r.config.InitialBackoff

	for attempt := 0; attempt <= r.config.MaxRetries; attempt++ {
		// Check circuit breaker
		err := r.circuitBreaker.Call(fn)
		if err == nil {
			return nil
		}

		lastErr = err

		// Don't sleep after last attempt
		if attempt == r.config.MaxRetries {
			break
		}

		// Check context cancellation
		select {
		case <-ctx.Done():
			return fmt.Errorf("retry cancelled: %w", ctx.Err())
		case <-time.After(backoff):
		}

		// Calculate next backoff
		backoff = time.Duration(float64(backoff) * r.config.BackoffFactor)
		if backoff > r.config.MaxBackoff {
			backoff = r.config.MaxBackoff
		}
	}

	return fmt.Errorf("max retries exceeded: %w", lastErr)
}

// RetryableSession wraps a Session with retry logic.
type RetryableSession struct {
	session        Session
	config         RetryConfig
	circuitBreaker *CircuitBreaker
}

// NewRetryableSession wraps a session with retry capabilities.
func NewRetryableSession(session Session, config RetryConfig) *RetryableSession {
	return &RetryableSession{
		session: session,
		config:  config,
		circuitBreaker: NewCircuitBreaker(
			config.FailureThreshold,
			config.ResetTimeout,
		),
	}
}

// Run executes the session with retry.
func (r *RetryableSession) Run(ctx context.Context, prompt string) (*Result, error) {
	var result *Result
	err := r.retryWithBackoff(ctx, func() error {
		var err error
		result, err = r.session.Run(ctx, prompt)
		return err
	})
	return result, err
}

// Events returns the event channel.
func (r *RetryableSession) Events() <-chan Event {
	return r.session.Events()
}

// Destroy destroys the session.
func (r *RetryableSession) Destroy(ctx context.Context) error {
	return r.session.Destroy(ctx)
}

// retryWithBackoff implements exponential backoff retry logic.
func (r *RetryableSession) retryWithBackoff(ctx context.Context, fn func() error) error {
	var lastErr error
	backoff := r.config.InitialBackoff

	for attempt := 0; attempt <= r.config.MaxRetries; attempt++ {
		// Check circuit breaker
		err := r.circuitBreaker.Call(fn)
		if err == nil {
			return nil
		}

		lastErr = err

		// Don't sleep after last attempt
		if attempt == r.config.MaxRetries {
			break
		}

		// Check context cancellation
		select {
		case <-ctx.Done():
			return fmt.Errorf("retry cancelled: %w", ctx.Err())
		case <-time.After(backoff):
		}

		// Calculate next backoff
		backoff = time.Duration(float64(backoff) * r.config.BackoffFactor)
		if backoff > r.config.MaxBackoff {
			backoff = r.config.MaxBackoff
		}
	}

	return fmt.Errorf("max retries exceeded: %w", lastErr)
}
