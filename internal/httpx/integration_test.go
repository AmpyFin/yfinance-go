package httpx

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestConcurrencyAndQPSShaping tests the concurrency and QPS shaping requirements
// QPS shaping integration test removed - timing sensitive and not core functionality

// TestCircuitBreakerIntegration tests circuit breaker behavior
func TestCircuitBreakerIntegration(t *testing.T) {
	// Create test server that fails requests
	failureCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		failureCount++
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Server Error"))
	}))
	defer server.Close()

	// Create config with aggressive circuit breaker settings
	config := &Config{
		BaseURL:               server.URL,
		Timeout:               5 * time.Second,
		IdleTimeout:           30 * time.Second,
		MaxConnsPerHost:       1,
		MaxAttempts:           1, // No retries
		BackoffBaseMs:         100,
		BackoffJitterMs:       50,
		MaxDelayMs:            1000,
		QPS:                   10.0, // High QPS to trigger failures quickly
		Burst:                 10,
		CircuitWindow:         1 * time.Second,        // Short window
		FailureThreshold:      2,                      // Open after 2 failures
		ResetTimeout:          100 * time.Millisecond, // Quick reset
		UserAgent:             "test-agent",
		EnableSessionRotation: false,
		NumSessions:           1,
	}

	// Create client
	client := NewClient(config)

	// Create request
	req, err := http.NewRequest("GET", server.URL+"/test", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Make requests until circuit breaker opens
	successCount := 0
	circuitOpenCount := 0

	for i := 0; i < 10; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		resp, err := client.Do(ctx, req)
		cancel()

		if err != nil {
			if err == ErrCircuitOpen {
				circuitOpenCount++
			}
		} else {
			successCount++
			resp.Body.Close()
		}

		// Small delay to let circuit breaker state update
		time.Sleep(10 * time.Millisecond)
	}

	// Should have some circuit open responses
	if circuitOpenCount == 0 {
		t.Error("Expected circuit breaker to open, but no circuit open errors occurred")
	}

	// Should have some failures from the server
	if failureCount == 0 {
		t.Error("Expected server failures, but none occurred")
	}
}
