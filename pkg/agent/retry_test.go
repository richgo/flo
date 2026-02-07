package agent

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestCircuitBreaker_Call(t *testing.T) {
	tests := []struct {
		name             string
		failureThreshold int
		resetTimeout     time.Duration
		calls            []bool // true = success, false = failure
		wantState        CircuitState
		wantErrors       int
	}{
		{
			name:             "success keeps circuit closed",
			failureThreshold: 3,
			resetTimeout:     time.Second,
			calls:            []bool{true, true, true},
			wantState:        CircuitClosed,
			wantErrors:       0,
		},
		{
			name:             "failures below threshold keeps circuit closed",
			failureThreshold: 3,
			resetTimeout:     time.Second,
			calls:            []bool{false, false, true},
			wantState:        CircuitClosed,
			wantErrors:       2,
		},
		{
			name:             "failures at threshold opens circuit",
			failureThreshold: 3,
			resetTimeout:     time.Second,
			calls:            []bool{false, false, false},
			wantState:        CircuitOpen,
			wantErrors:       3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cb := NewCircuitBreaker(tt.failureThreshold, tt.resetTimeout)
			errorCount := 0

			for _, shouldSucceed := range tt.calls {
				err := cb.Call(func() error {
					if shouldSucceed {
						return nil
					}
					return errors.New("simulated failure")
				})
				if err != nil {
					errorCount++
				}
			}

			if errorCount != tt.wantErrors {
				t.Errorf("got %d errors, want %d", errorCount, tt.wantErrors)
			}

			if cb.State() != tt.wantState {
				t.Errorf("circuit state = %v, want %v", cb.State(), tt.wantState)
			}
		})
	}
}

func TestCircuitBreaker_Reset(t *testing.T) {
	cb := NewCircuitBreaker(2, time.Second)

	// Trigger failures to open circuit
	cb.Call(func() error { return errors.New("fail") })
	cb.Call(func() error { return errors.New("fail") })

	if cb.State() != CircuitOpen {
		t.Errorf("circuit state = %v, want CircuitOpen", cb.State())
	}

	// Reset should close circuit
	cb.Reset()

	if cb.State() != CircuitClosed {
		t.Errorf("circuit state after reset = %v, want CircuitClosed", cb.State())
	}
}

func TestCircuitBreaker_HalfOpen(t *testing.T) {
	resetTimeout := 100 * time.Millisecond
	cb := NewCircuitBreaker(2, resetTimeout)

	// Open the circuit
	cb.Call(func() error { return errors.New("fail") })
	cb.Call(func() error { return errors.New("fail") })

	if cb.State() != CircuitOpen {
		t.Fatalf("circuit state = %v, want CircuitOpen", cb.State())
	}

	// Wait for reset timeout
	time.Sleep(resetTimeout + 10*time.Millisecond)

	// Next call should transition to half-open
	err := cb.Call(func() error { return nil })
	if err != nil {
		t.Errorf("call in half-open state failed: %v", err)
	}

	// Successful call should close circuit
	if cb.State() != CircuitClosed {
		t.Errorf("circuit state after success = %v, want CircuitClosed", cb.State())
	}
}

func TestRetryableBackend_RetryLogic(t *testing.T) {
	tests := []struct {
		name          string
		maxRetries    int
		failures      int
		wantAttempts  int
		wantSuccess   bool
	}{
		{
			name:         "success on first try",
			maxRetries:   3,
			failures:     0,
			wantAttempts: 1,
			wantSuccess:  true,
		},
		{
			name:         "success on second try",
			maxRetries:   3,
			failures:     1,
			wantAttempts: 2,
			wantSuccess:  true,
		},
		{
			name:         "success on last try",
			maxRetries:   3,
			failures:     3,
			wantAttempts: 4,
			wantSuccess:  true,
		},
		{
			name:         "failure after max retries",
			maxRetries:   2,
			failures:     10,
			wantAttempts: 3,
			wantSuccess:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBackend := NewMockBackend()
			config := RetryConfig{
				MaxRetries:       tt.maxRetries,
				InitialBackoff:   time.Millisecond,
				MaxBackoff:       10 * time.Millisecond,
				BackoffFactor:    2.0,
				FailureThreshold: 100, // High threshold to avoid circuit breaker interference
				ResetTimeout:     time.Second,
			}

			rb := NewRetryableBackend(mockBackend, config)

			attempts := 0
			err := rb.retryWithBackoff(context.Background(), func() error {
				attempts++
				if attempts <= tt.failures {
					return errors.New("simulated failure")
				}
				return nil
			})

			if attempts != tt.wantAttempts {
				t.Errorf("attempts = %d, want %d", attempts, tt.wantAttempts)
			}

			if (err == nil) != tt.wantSuccess {
				t.Errorf("success = %v, want %v (error: %v)", err == nil, tt.wantSuccess, err)
			}
		})
	}
}

func TestRetryableBackend_ContextCancellation(t *testing.T) {
	mockBackend := NewMockBackend()
	config := RetryConfig{
		MaxRetries:       10,
		InitialBackoff:   100 * time.Millisecond,
		MaxBackoff:       time.Second,
		BackoffFactor:    2.0,
		FailureThreshold: 100,
		ResetTimeout:     time.Second,
	}

	rb := NewRetryableBackend(mockBackend, config)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	attempts := 0
	err := rb.retryWithBackoff(ctx, func() error {
		attempts++
		return errors.New("simulated failure")
	})

	if err == nil {
		t.Error("expected error due to context cancellation")
	}

	// Should have attempted at least once but stopped early due to cancellation
	if attempts > 3 {
		t.Errorf("too many attempts %d before cancellation", attempts)
	}
}

func TestRetryableBackend_ExponentialBackoff(t *testing.T) {
	mockBackend := NewMockBackend()
	config := RetryConfig{
		MaxRetries:       3,
		InitialBackoff:   10 * time.Millisecond,
		MaxBackoff:       100 * time.Millisecond,
		BackoffFactor:    2.0,
		FailureThreshold: 100,
		ResetTimeout:     time.Second,
	}

	rb := NewRetryableBackend(mockBackend, config)

	start := time.Now()
	attempts := 0

	rb.retryWithBackoff(context.Background(), func() error {
		attempts++
		return errors.New("simulated failure")
	})

	elapsed := time.Since(start)

	// With initial backoff of 10ms and factor 2.0:
	// Attempt 1: immediate
	// Attempt 2: wait 10ms
	// Attempt 3: wait 20ms
	// Attempt 4: wait 40ms
	// Total wait time: 10 + 20 + 40 = 70ms
	expectedMin := 70 * time.Millisecond

	if elapsed < expectedMin {
		t.Errorf("elapsed time %v is less than expected minimum %v", elapsed, expectedMin)
	}
}

func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()

	if config.MaxRetries <= 0 {
		t.Error("MaxRetries should be > 0")
	}
	if config.InitialBackoff <= 0 {
		t.Error("InitialBackoff should be > 0")
	}
	if config.MaxBackoff <= config.InitialBackoff {
		t.Error("MaxBackoff should be > InitialBackoff")
	}
	if config.BackoffFactor <= 1.0 {
		t.Error("BackoffFactor should be > 1.0")
	}
	if config.FailureThreshold <= 0 {
		t.Error("FailureThreshold should be > 0")
	}
	if config.ResetTimeout <= 0 {
		t.Error("ResetTimeout should be > 0")
	}
}
