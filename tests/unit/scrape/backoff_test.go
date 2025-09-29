package scrape_test

import (
	"testing"
	"time"

	"github.com/AmpyFin/yfinance-go/internal/scrape"
)

func TestBackoffPolicy_CalculateDelay(t *testing.T) {
	policy := scrape.DefaultBackoffPolicy()

	// Test first few attempts
	delays := policy.CalculateDelays(5)
	
	// Verify delays are increasing
	for i := 1; i < len(delays); i++ {
		if delays[i] <= delays[i-1] {
			t.Errorf("Delay should increase with attempts: attempt %d = %v, attempt %d = %v", 
				i-1, delays[i-1], i, delays[i])
		}
	}

	// Verify all delays are within bounds
	for i, delay := range delays {
		if delay < policy.BaseDelay {
			t.Errorf("Attempt %d: delay %v is less than base delay %v", i, delay, policy.BaseDelay)
		}
		if delay > policy.MaxDelay {
			t.Errorf("Attempt %d: delay %v exceeds max delay %v", i, delay, policy.MaxDelay)
		}
	}
}

func TestBackoffPolicy_CalculateDelayWithRetryAfter(t *testing.T) {
	policy := scrape.DefaultBackoffPolicy()
	
	// Test with reasonable Retry-After
	retryAfter := 2 * time.Second
	delay := policy.CalculateDelayWithRetryAfter(0, retryAfter)
	
	// Should be close to retry-after with some jitter
	if delay < retryAfter/2 || delay > retryAfter*2 {
		t.Errorf("Delay %v should be close to retry-after %v", delay, retryAfter)
	}

	// Test with very long Retry-After (should fall back to normal backoff)
	longRetryAfter := 1 * time.Hour
	delay = policy.CalculateDelayWithRetryAfter(0, longRetryAfter)
	normalDelay := policy.CalculateDelay(0)
	
	// Should be similar to normal delay
	if delay < normalDelay/2 || delay > normalDelay*2 {
		t.Errorf("Delay %v should be similar to normal delay %v for long retry-after", delay, normalDelay)
	}
}

func TestBackoffPolicy_Validate(t *testing.T) {
	tests := []struct {
		name    string
		policy  *scrape.BackoffPolicy
		wantErr bool
	}{
		{
			name:    "valid policy",
			policy:  scrape.DefaultBackoffPolicy(),
			wantErr: false,
		},
		{
			name: "invalid base delay",
			policy: &scrape.BackoffPolicy{
				BaseDelay:    -1 * time.Second,
				MaxDelay:     4 * time.Second,
				Multiplier:   2.0,
				JitterFactor: 0.2,
			},
			wantErr: true,
		},
		{
			name: "invalid max delay",
			policy: &scrape.BackoffPolicy{
				BaseDelay:    300 * time.Millisecond,
				MaxDelay:     -1 * time.Second,
				Multiplier:   2.0,
				JitterFactor: 0.2,
			},
			wantErr: true,
		},
		{
			name: "base delay greater than max delay",
			policy: &scrape.BackoffPolicy{
				BaseDelay:    5 * time.Second,
				MaxDelay:     4 * time.Second,
				Multiplier:   2.0,
				JitterFactor: 0.2,
			},
			wantErr: true,
		},
		{
			name: "invalid multiplier",
			policy: &scrape.BackoffPolicy{
				BaseDelay:    300 * time.Millisecond,
				MaxDelay:     4 * time.Second,
				Multiplier:   1.0, // Should be > 1.0
				JitterFactor: 0.2,
			},
			wantErr: true,
		},
		{
			name: "invalid jitter factor",
			policy: &scrape.BackoffPolicy{
				BaseDelay:    300 * time.Millisecond,
				MaxDelay:     4 * time.Second,
				Multiplier:   2.0,
				JitterFactor: 1.5, // Should be <= 1.0
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.policy.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("BackoffPolicy.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBackoffPolicy_GetStats(t *testing.T) {
	policy := scrape.DefaultBackoffPolicy()
	stats := policy.GetStats()

	// Verify all expected fields are present
	expectedFields := []string{"base_delay_ms", "max_delay_ms", "multiplier", "jitter_factor"}
	for _, field := range expectedFields {
		if _, exists := stats[field]; !exists {
			t.Errorf("Stats missing field: %s", field)
		}
	}

	// Verify values
	if stats["base_delay_ms"] != policy.BaseDelay.Milliseconds() {
		t.Errorf("base_delay_ms: expected %v, got %v", policy.BaseDelay.Milliseconds(), stats["base_delay_ms"])
	}
	if stats["max_delay_ms"] != policy.MaxDelay.Milliseconds() {
		t.Errorf("max_delay_ms: expected %v, got %v", policy.MaxDelay.Milliseconds(), stats["max_delay_ms"])
	}
	if stats["multiplier"] != policy.Multiplier {
		t.Errorf("multiplier: expected %v, got %v", policy.Multiplier, stats["multiplier"])
	}
	if stats["jitter_factor"] != policy.JitterFactor {
		t.Errorf("jitter_factor: expected %v, got %v", policy.JitterFactor, stats["jitter_factor"])
	}
}

func TestBackoffPolicy_CalculateDelays(t *testing.T) {
	policy := scrape.DefaultBackoffPolicy()
	delays := policy.CalculateDelays(3)

	if len(delays) != 3 {
		t.Errorf("Expected 3 delays, got %d", len(delays))
	}

	// Verify delays are reasonable
	for i, delay := range delays {
		if delay <= 0 {
			t.Errorf("Delay %d should be positive, got %v", i, delay)
		}
		if delay < policy.BaseDelay {
			t.Errorf("Delay %d %v should be >= base delay %v", i, delay, policy.BaseDelay)
		}
		if delay > policy.MaxDelay {
			t.Errorf("Delay %d %v should be <= max delay %v", i, delay, policy.MaxDelay)
		}
	}
}
