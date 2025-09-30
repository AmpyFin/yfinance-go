package bus

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"
)

// RetryPolicy implements retry logic with exponential backoff
type RetryPolicy struct {
	config *RetryConfig
}

// NewRetryPolicy creates a new retry policy
func NewRetryPolicy(config *RetryConfig) *RetryPolicy {
	return &RetryPolicy{
		config: config,
	}
}

// RetryableError represents an error that can be retried
type RetryableError struct {
	Err error
}

func (e *RetryableError) Error() string {
	return fmt.Sprintf("retryable error: %v", e.Err)
}

func (e *RetryableError) Unwrap() error {
	return e.Err
}

// IsRetryableError checks if an error is retryable
func IsRetryableError(err error) bool {
	_, ok := err.(*RetryableError)
	return ok
}

// ExecuteWithRetry executes a function with retry logic
func (rp *RetryPolicy) ExecuteWithRetry(ctx context.Context, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt < rp.config.Attempts; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if !IsRetryableError(err) {
			return err
		}

		// Don't sleep on the last attempt
		if attempt == rp.config.Attempts-1 {
			break
		}

		// Calculate delay with exponential backoff and jitter
		delay := rp.calculateDelay(attempt)

		// Wait with context cancellation support
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			// Continue to next attempt
		}
	}

	return fmt.Errorf("failed after %d attempts: %w", rp.config.Attempts, lastErr)
}

// calculateDelay calculates the delay for the given attempt
func (rp *RetryPolicy) calculateDelay(attempt int) time.Duration {
	// Exponential backoff: base * 2^attempt
	delay := float64(rp.config.BaseMs) * math.Pow(2, float64(attempt))

	// Cap at max delay
	if delay > float64(rp.config.MaxDelayMs) {
		delay = float64(rp.config.MaxDelayMs)
	}

	// Add jitter (±25%)
	jitter := delay * 0.25 * (rand.Float64() - 0.5)
	delay += jitter

	// Ensure minimum delay
	if delay < 0 {
		delay = 1
	}

	return time.Duration(delay) * time.Millisecond
}

// CircuitBreaker implements circuit breaker pattern
type CircuitBreaker struct {
	config          *CircuitBreakerConfig
	state           CircuitBreakerState
	failureCount    int
	successCount    int
	lastFailureTime time.Time
	nextAttemptTime time.Time
}

// CircuitBreakerState represents the state of the circuit breaker
type CircuitBreakerState int

const (
	CircuitBreakerClosed CircuitBreakerState = iota
	CircuitBreakerOpen
	CircuitBreakerHalfOpen
)

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(config *CircuitBreakerConfig) *CircuitBreaker {
	return &CircuitBreaker{
		config: config,
		state:  CircuitBreakerClosed,
	}
}

// Execute executes a function with circuit breaker protection
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func() error) error {
	// Check if circuit breaker allows execution
	if !cb.canExecute() {
		return fmt.Errorf("circuit breaker is open")
	}

	// Execute the function
	err := fn()

	// Update circuit breaker state based on result
	cb.recordResult(err)

	return err
}

// canExecute checks if the circuit breaker allows execution
func (cb *CircuitBreaker) canExecute() bool {
	now := time.Now()

	switch cb.state {
	case CircuitBreakerClosed:
		return true
	case CircuitBreakerOpen:
		// Check if reset timeout has passed
		if now.After(cb.nextAttemptTime) {
			cb.state = CircuitBreakerHalfOpen
			cb.successCount = 0
			return true
		}
		return false
	case CircuitBreakerHalfOpen:
		// Allow limited number of probes
		return cb.successCount < cb.config.HalfOpenProbes
	default:
		return false
	}
}

// recordResult records the result of an execution
func (cb *CircuitBreaker) recordResult(err error) {
	now := time.Now()

	if err != nil {
		cb.failureCount++
		cb.lastFailureTime = now

		// Check if we should open the circuit
		if cb.shouldOpenCircuit() {
			cb.state = CircuitBreakerOpen
			cb.nextAttemptTime = now.Add(time.Duration(cb.config.ResetTimeoutMs) * time.Millisecond)
		}
	} else {
		cb.successCount++

		// Reset failure count on success
		if cb.state == CircuitBreakerClosed {
			cb.failureCount = 0
		}

		// Check if we should close the circuit (half-open -> closed)
		if cb.state == CircuitBreakerHalfOpen && cb.successCount >= cb.config.HalfOpenProbes {
			cb.state = CircuitBreakerClosed
			cb.failureCount = 0
		}
	}
}

// shouldOpenCircuit checks if the circuit should be opened
func (cb *CircuitBreaker) shouldOpenCircuit() bool {
	// Check if we have enough failures to open the circuit
	if cb.failureCount < cb.config.Window {
		return false
	}

	// Calculate failure rate
	failureRate := float64(cb.failureCount) / float64(cb.config.Window)
	return failureRate >= cb.config.FailureThreshold
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	return cb.state
}

// GetStats returns statistics about the circuit breaker
func (cb *CircuitBreaker) GetStats() CircuitBreakerStats {
	return CircuitBreakerStats{
		State:           cb.state,
		FailureCount:    cb.failureCount,
		SuccessCount:    cb.successCount,
		LastFailureTime: cb.lastFailureTime,
		NextAttemptTime: cb.nextAttemptTime,
	}
}

// CircuitBreakerStats represents statistics about the circuit breaker
type CircuitBreakerStats struct {
	State           CircuitBreakerState
	FailureCount    int
	SuccessCount    int
	LastFailureTime time.Time
	NextAttemptTime time.Time
}
