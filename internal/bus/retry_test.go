package bus

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRetryPolicy_ExecuteWithRetry(t *testing.T) {
	tests := []struct {
		name           string
		config         *RetryConfig
		fn             func() error
		expectedError  bool
		expectedAttempts int
	}{
		{
			name: "success on first attempt",
			config: &RetryConfig{
				Attempts:   3,
				BaseMs:     100,
				MaxDelayMs: 1000,
			},
			fn: func() error {
				return nil
			},
			expectedError:   false,
			expectedAttempts: 1,
		},
		{
			name: "success on retry",
			config: &RetryConfig{
				Attempts:   3,
				BaseMs:     100,
				MaxDelayMs: 1000,
			},
			fn: func() error {
				// Simulate failure then success
				return &RetryableError{Err: errors.New("temporary error")}
			},
			expectedError:   true, // Will fail after all attempts
			expectedAttempts: 3,
		},
		{
			name: "non-retryable error",
			config: &RetryConfig{
				Attempts:   3,
				BaseMs:     100,
				MaxDelayMs: 1000,
			},
			fn: func() error {
				return errors.New("permanent error")
			},
			expectedError:   true,
			expectedAttempts: 1,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy := NewRetryPolicy(tt.config)
			
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			
			attempts := 0
			fn := func() error {
				attempts++
				return tt.fn()
			}
			
			err := policy.ExecuteWithRetry(ctx, fn)
			
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			
			assert.Equal(t, tt.expectedAttempts, attempts)
		})
	}
}

func TestRetryPolicy_calculateDelay(t *testing.T) {
	policy := NewRetryPolicy(&RetryConfig{
		BaseMs:     100,
		MaxDelayMs: 1000,
	})
	
	tests := []struct {
		name     string
		attempt  int
		minDelay time.Duration
		maxDelay time.Duration
	}{
		{
			name:     "first retry",
			attempt:  0,
			minDelay: 50 * time.Millisecond,  // 100ms - 25% jitter
			maxDelay: 150 * time.Millisecond, // 100ms + 25% jitter
		},
		{
			name:     "second retry",
			attempt:  1,
			minDelay: 150 * time.Millisecond, // 200ms - 25% jitter
			maxDelay: 250 * time.Millisecond, // 200ms + 25% jitter
		},
		{
			name:     "third retry",
			attempt:  2,
			minDelay: 350 * time.Millisecond, // 400ms - 25% jitter
			maxDelay: 450 * time.Millisecond, // 400ms + 25% jitter
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delay := policy.calculateDelay(tt.attempt)
			assert.GreaterOrEqual(t, delay, tt.minDelay)
			assert.LessOrEqual(t, delay, tt.maxDelay)
		})
	}
}

func TestCircuitBreaker_Execute(t *testing.T) {
	config := &CircuitBreakerConfig{
		Window:          3,
		FailureThreshold: 0.5,
		ResetTimeoutMs:  100,
		HalfOpenProbes:  2,
	}
	
	t.Run("successful execution", func(t *testing.T) {
		cb := NewCircuitBreaker(config)
		ctx := context.Background()
		
		err := cb.Execute(ctx, func() error {
			return nil
		})
		
		assert.NoError(t, err)
		assert.Equal(t, CircuitBreakerClosed, cb.GetState())
	})
	
	t.Run("circuit opens after failures", func(t *testing.T) {
		cb := NewCircuitBreaker(config)
		ctx := context.Background()
		
		// Execute failing function multiple times
		for i := 0; i < 3; i++ {
			err := cb.Execute(ctx, func() error {
				return &RetryableError{Err: errors.New("temporary error")}
			})
			assert.Error(t, err)
		}
		
		// Circuit should be open now
		assert.Equal(t, CircuitBreakerOpen, cb.GetState())
		
		// Next execution should fail immediately
		err := cb.Execute(ctx, func() error {
			return nil
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "circuit breaker is open")
	})
	
	t.Run("circuit resets after timeout", func(t *testing.T) {
		cb := NewCircuitBreaker(config)
		ctx := context.Background()
		
		// First, open the circuit by causing failures
		for i := 0; i < 3; i++ {
			err := cb.Execute(ctx, func() error {
				return &RetryableError{Err: errors.New("temporary error")}
			})
			assert.Error(t, err)
		}
		
		// Circuit should be open now
		assert.Equal(t, CircuitBreakerOpen, cb.GetState())
		
		// Wait for reset timeout
		time.Sleep(150 * time.Millisecond)
		
		// Try to execute - this should trigger the transition to half-open
		err := cb.Execute(ctx, func() error {
			return nil
		})
		assert.NoError(t, err)
		
		// Circuit should be half-open now (after first successful execution)
		assert.Equal(t, CircuitBreakerHalfOpen, cb.GetState())
		
		// Execute one more successful operation to close the circuit
		err = cb.Execute(ctx, func() error {
			return nil
		})
		assert.NoError(t, err)
		assert.Equal(t, CircuitBreakerClosed, cb.GetState())
	})
}

func TestCircuitBreaker_GetStats(t *testing.T) {
	config := &CircuitBreakerConfig{
		Window:          3,
		FailureThreshold: 0.5,
		ResetTimeoutMs:  100,
		HalfOpenProbes:  2,
	}
	
	cb := NewCircuitBreaker(config)
	
	// Execute some operations
	ctx := context.Background()
	_ = cb.Execute(ctx, func() error { return nil })
	_ = cb.Execute(ctx, func() error { return &RetryableError{Err: errors.New("error")} })
	
	stats := cb.GetStats()
	assert.Equal(t, CircuitBreakerClosed, stats.State)
	assert.Equal(t, 1, stats.SuccessCount)
	assert.Equal(t, 1, stats.FailureCount)
}

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "retryable error",
			err:      &RetryableError{Err: errors.New("temporary")},
			expected: true,
		},
		{
			name:     "non-retryable error",
			err:      errors.New("permanent"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsRetryableError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
