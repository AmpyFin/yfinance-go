package httpx

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClientRetry(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	}))
	defer server.Close()

	config := DefaultConfig()
	config.BaseURL = server.URL
	config.MaxAttempts = 5
	config.BackoffBaseMs = 10 // Fast backoff for testing

	client := NewClient(config)

	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.Do(ctx, req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

func TestClientCircuitBreaker(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	config := DefaultConfig()
	config.BaseURL = server.URL
	config.MaxAttempts = 1
	config.FailureThreshold = 2
	config.CircuitWindow = 100 * time.Millisecond
	config.ResetTimeout = 50 * time.Millisecond

	client := NewClient(config)

	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// First request should fail
	_, err = client.Do(ctx, req)
	if err == nil {
		t.Error("Request 1 should have failed")
	}

	// Check circuit is still closed (1 failure, threshold is 2)
	if client.circuitBreaker.State() != StateClosed {
		t.Errorf("Expected circuit to be closed after 1 failure, got state %v", client.circuitBreaker.State())
	}

	// Second request should fail and open the circuit
	_, err = client.Do(ctx, req)
	if err == nil {
		t.Error("Request 2 should have failed")
	}

	// Check circuit is now open (2 failures, threshold is 2)
	if client.circuitBreaker.State() != StateOpen {
		t.Errorf("Expected circuit to be open after 2 failures, got state %v", client.circuitBreaker.State())
	}

	// Third request should be rejected by circuit breaker
	_, err = client.Do(ctx, req)
	if err == nil {
		t.Error("Request should have been rejected by circuit breaker")
	}
	if !errors.Is(err, ErrCircuitOpen) {
		t.Errorf("Expected ErrCircuitOpen, got %v", err)
	}

	// Wait for circuit to reset (transition to half-open)
	time.Sleep(60 * time.Millisecond) // Wait longer than resetTimeout

	// Trigger transition to half-open by calling Allow()
	allowed := client.circuitBreaker.Allow()
	if !allowed {
		t.Error("Circuit should allow requests in half-open state")
	}

	// Check circuit is now half-open
	if client.circuitBreaker.State() != StateHalfOpen {
		t.Errorf("Expected circuit to be half-open after reset timeout, got state %v", client.circuitBreaker.State())
	}

	// Next request should be allowed (half-open state) but still fail
	_, err = client.Do(ctx, req)
	if err == nil {
		t.Error("Request should still fail (server still returns 500)")
	}

	// After failure in half-open state, circuit should be open again
	if client.circuitBreaker.State() != StateOpen {
		t.Errorf("Expected circuit to be open again after failure in half-open state, got state %v", client.circuitBreaker.State())
	}
}

func TestRateLimiter(t *testing.T) {
	limiter := NewRateLimiter(2.0, 2) // 2 QPS, burst of 2

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	start := time.Now()

	// First two requests should be immediate (burst)
	for i := 0; i < 2; i++ {
		err := limiter.Wait(ctx)
		if err != nil {
			t.Errorf("Request %d failed: %v", i+1, err)
		}
	}

	// Third request should be throttled
	err := limiter.Wait(ctx)
	if err != nil {
		t.Errorf("Request 3 failed: %v", err)
	}

	elapsed := time.Since(start)
	if elapsed < 500*time.Millisecond {
		t.Errorf("Expected at least 500ms delay, got %v", elapsed)
	}
}

func TestErrorTypes(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		retryable bool
		fatal     bool
	}{
		{
			name:      "429 error",
			err:       NewHTTPError(429, "Too Many Requests", nil),
			retryable: true,
			fatal:     false,
		},
		{
			name:      "500 error",
			err:       NewHTTPError(500, "Internal Server Error", nil),
			retryable: true,
			fatal:     false,
		},
		{
			name:      "400 error",
			err:       NewHTTPError(400, "Bad Request", nil),
			retryable: false,
			fatal:     true,
		},
		{
			name:      "404 error",
			err:       NewHTTPError(404, "Not Found", nil),
			retryable: false,
			fatal:     true,
		},
		{
			name:      "timeout error",
			err:       ErrTimeout,
			retryable: true,
			fatal:     false,
		},
		{
			name:      "decode error",
			err:       ErrDecode,
			retryable: false,
			fatal:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if IsRetryableError(tt.err) != tt.retryable {
				t.Errorf("IsRetryableError() = %v, want %v", IsRetryableError(tt.err), tt.retryable)
			}
			if IsFatalError(tt.err) != tt.fatal {
				t.Errorf("IsFatalError() = %v, want %v", IsFatalError(tt.err), tt.fatal)
			}
		})
	}
}
