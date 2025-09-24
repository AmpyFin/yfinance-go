package httpx

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

// TestConcurrencyAndQPSShaping tests the concurrency and QPS shaping requirements
func TestConcurrencyAndQPSShaping(t *testing.T) {
	// Create test server that tracks requests
	requestCount := 0
	requestTimes := make([]time.Time, 0)
	var mu sync.Mutex
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requestCount++
		requestTimes = append(requestTimes, time.Now())
		mu.Unlock()
		
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()
	
	// Create config with specific settings for testing
	config := &Config{
		BaseURL:            server.URL,
		Timeout:            30 * time.Second,
		IdleTimeout:        90 * time.Second,
		MaxConnsPerHost:    5,
		MaxAttempts:        3,
		BackoffBaseMs:      100,
		BackoffJitterMs:    50,
		MaxDelayMs:         1000,
		QPS:                2.0, // 2 QPS
		Burst:              2,
		CircuitWindow:      60 * time.Second,
		FailureThreshold:   3,
		ResetTimeout:       30 * time.Second,
		UserAgent:          "test-agent",
		EnableSessionRotation: false,
		NumSessions:        1,
	}
	
	// Create client
	client := NewClient(config)
	
	// Create request
	req, err := http.NewRequest("GET", server.URL+"/test", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	
	// Make multiple requests to test QPS shaping
	startTime := time.Now()
	numRequests := 10
	
	for i := 0; i < numRequests; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		resp, err := client.Do(ctx, req)
		cancel()
		
		if err != nil {
			t.Errorf("Request %d failed: %v", i, err)
			continue
		}
		
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Request %d returned status %d", i, resp.StatusCode)
		}
		
		resp.Body.Close()
	}
	
	elapsed := time.Since(startTime)
	
	// Check that we made the expected number of requests
	mu.Lock()
	actualRequests := requestCount
	mu.Unlock()
	
	if actualRequests != numRequests {
		t.Errorf("Expected %d requests, got %d", numRequests, actualRequests)
	}
	
	// Check that QPS shaping is working (should take at least some time)
	// With 2 QPS, 10 requests should take at least 4.5 seconds (9 intervals of 0.5s each)
	expectedMinTime := time.Duration(4.5 * float64(time.Second))
	if elapsed < expectedMinTime {
		t.Errorf("QPS shaping not working properly. Expected at least %v, got %v", expectedMinTime, elapsed)
	}
}

// TestCircuitBreakerIntegration tests circuit breaker behavior
func TestCircuitBreakerIntegration(t *testing.T) {
	// Create test server that fails requests
	failureCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		failureCount++
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Server Error"))
	}))
	defer server.Close()
	
	// Create config with aggressive circuit breaker settings
	config := &Config{
		BaseURL:            server.URL,
		Timeout:            5 * time.Second,
		IdleTimeout:        30 * time.Second,
		MaxConnsPerHost:    1,
		MaxAttempts:        1, // No retries
		BackoffBaseMs:      100,
		BackoffJitterMs:    50,
		MaxDelayMs:         1000,
		QPS:                10.0, // High QPS to trigger failures quickly
		Burst:              10,
		CircuitWindow:      1 * time.Second, // Short window
		FailureThreshold:   2, // Open after 2 failures
		ResetTimeout:       100 * time.Millisecond, // Quick reset
		UserAgent:          "test-agent",
		EnableSessionRotation: false,
		NumSessions:        1,
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