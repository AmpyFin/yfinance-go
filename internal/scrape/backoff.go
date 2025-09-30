package scrape

import (
	"math"
	"math/rand"
	"time"
)

// BackoffPolicy implements exponential backoff with jitter
type BackoffPolicy struct {
	BaseDelay    time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
	JitterFactor float64
}

// DefaultBackoffPolicy returns a sensible default backoff policy
func DefaultBackoffPolicy() *BackoffPolicy {
	return &BackoffPolicy{
		BaseDelay:    300 * time.Millisecond,
		MaxDelay:     4 * time.Second,
		Multiplier:   2.0,
		JitterFactor: 0.2, // ±20% jitter
	}
}

// CalculateDelay calculates the backoff delay for a given attempt
func (bp *BackoffPolicy) CalculateDelay(attempt int) time.Duration {
	// Exponential backoff: base * multiplier^attempt
	exponentialDelay := float64(bp.BaseDelay) * math.Pow(bp.Multiplier, float64(attempt))

	// Add jitter (±jitterFactor)
	jitter := exponentialDelay * bp.JitterFactor * (2*rand.Float64() - 1) // -1 to 1
	totalDelay := exponentialDelay + jitter

	// Ensure minimum delay
	if totalDelay < float64(bp.BaseDelay) {
		totalDelay = float64(bp.BaseDelay)
	}

	// Cap at maximum delay
	if totalDelay > float64(bp.MaxDelay) {
		totalDelay = float64(bp.MaxDelay)
	}

	return time.Duration(totalDelay)
}

// CalculateDelayWithRetryAfter calculates backoff delay considering Retry-After header
func (bp *BackoffPolicy) CalculateDelayWithRetryAfter(attempt int, retryAfter time.Duration) time.Duration {
	// Use Retry-After if it's reasonable (not too long)
	if retryAfter > 0 && retryAfter < bp.MaxDelay*2 {
		// Add some jitter to Retry-After
		jitter := retryAfter * time.Duration(bp.JitterFactor*(2*rand.Float64()-1))
		delay := retryAfter + jitter

		// Ensure it's not too short
		if delay < bp.BaseDelay {
			delay = bp.BaseDelay
		}

		return delay
	}

	// Fall back to normal exponential backoff
	return bp.CalculateDelay(attempt)
}

// NewBackoffPolicy creates a new backoff policy with custom parameters
func NewBackoffPolicy(baseDelay, maxDelay time.Duration, multiplier, jitterFactor float64) *BackoffPolicy {
	return &BackoffPolicy{
		BaseDelay:    baseDelay,
		MaxDelay:     maxDelay,
		Multiplier:   multiplier,
		JitterFactor: jitterFactor,
	}
}

// Validate validates the backoff policy parameters
func (bp *BackoffPolicy) Validate() error {
	if bp.BaseDelay <= 0 {
		return &ScrapeError{
			Type:    "invalid_config",
			Message: "base delay must be positive",
		}
	}

	if bp.MaxDelay <= 0 {
		return &ScrapeError{
			Type:    "invalid_config",
			Message: "max delay must be positive",
		}
	}

	if bp.BaseDelay > bp.MaxDelay {
		return &ScrapeError{
			Type:    "invalid_config",
			Message: "base delay cannot be greater than max delay",
		}
	}

	if bp.Multiplier <= 1.0 {
		return &ScrapeError{
			Type:    "invalid_config",
			Message: "multiplier must be greater than 1.0",
		}
	}

	if bp.JitterFactor < 0 || bp.JitterFactor > 1.0 {
		return &ScrapeError{
			Type:    "invalid_config",
			Message: "jitter factor must be between 0 and 1",
		}
	}

	return nil
}

// GetStats returns statistics about the backoff policy
func (bp *BackoffPolicy) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"base_delay_ms": bp.BaseDelay.Milliseconds(),
		"max_delay_ms":  bp.MaxDelay.Milliseconds(),
		"multiplier":    bp.Multiplier,
		"jitter_factor": bp.JitterFactor,
	}
}

// CalculateDelays calculates delays for multiple attempts (useful for testing)
func (bp *BackoffPolicy) CalculateDelays(maxAttempts int) []time.Duration {
	delays := make([]time.Duration, maxAttempts)
	for i := 0; i < maxAttempts; i++ {
		delays[i] = bp.CalculateDelay(i)
	}
	return delays
}
