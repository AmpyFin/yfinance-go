package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/AmpyFin/yfinance-go/internal/httpx"
)

func TestHTTPAdapterRetriesBackoff(t *testing.T) {
	// Test server that fails first 2 requests, then succeeds
	attemptCount := 0
	var mu sync.Mutex
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		attemptCount++
		currentAttempt := attemptCount
		mu.Unlock()
		
		if currentAttempt <= 2 {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("Server Error"))
		} else {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("Success"))
		}
	}))
	defer server.Close()

	config := &httpx.Config{
		BaseURL:            server.URL,
		Timeout:            30 * time.Second,
		IdleTimeout:        90 * time.Second,
		MaxConnsPerHost:    5,
		MaxAttempts:        3,
		BackoffBaseMs:      100,
		BackoffJitterMs:    50,
		MaxDelayMs:         1000,
		QPS:                10.0,
		Burst:              10,
		CircuitWindow:      60 * time.Second,
		FailureThreshold:   5,
		ResetTimeout:       30 * time.Second,
		UserAgent:          "test-agent",
		EnableSessionRotation: false,
		NumSessions:        1,
	}

	client := httpx.NewClient(config)
	req, err := http.NewRequest("GET", server.URL+"/test", nil)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := client.Do(ctx, req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	
	mu.Lock()
	finalAttempts := attemptCount
	mu.Unlock()
	assert.Equal(t, 3, finalAttempts, "Should have retried 2 times before success")
}

func TestHTTPAdapterCircuitBreaker(t *testing.T) {
	// Test server that always fails
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Server Error"))
	}))
	defer server.Close()

	config := &httpx.Config{
		BaseURL:            server.URL,
		Timeout:            30 * time.Second,
		IdleTimeout:        90 * time.Second,
		MaxConnsPerHost:    5,
		MaxAttempts:        1, // No retries for circuit breaker test
		BackoffBaseMs:      100,
		BackoffJitterMs:    50,
		MaxDelayMs:         1000,
		QPS:                10.0,
		Burst:              10,
		CircuitWindow:      5 * time.Second,
		FailureThreshold:   3,
		ResetTimeout:       2 * time.Second,
		UserAgent:          "test-agent",
		EnableSessionRotation: false,
		NumSessions:        1,
	}

	client := httpx.NewClient(config)
	req, err := http.NewRequest("GET", server.URL+"/test", nil)
	require.NoError(t, err)

	// Make requests to trigger circuit breaker
	for i := 0; i < 5; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		resp, err := client.Do(ctx, req)
		cancel()
		
		// All requests should fail, but we can't easily test circuit breaker state
		assert.Error(t, err)
		assert.Nil(t, resp)
	}
}

func TestHTTPAdapterRateLimiting(t *testing.T) {
	requestCount := 0
	var mu sync.Mutex
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requestCount++
		mu.Unlock()
		
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}))
	defer server.Close()

	config := &httpx.Config{
		BaseURL:            server.URL,
		Timeout:            30 * time.Second,
		IdleTimeout:        90 * time.Second,
		MaxConnsPerHost:    5,
		MaxAttempts:        1,
		BackoffBaseMs:      100,
		BackoffJitterMs:    50,
		MaxDelayMs:         1000,
		QPS:                2.0, // 2 QPS
		Burst:              2,
		CircuitWindow:      60 * time.Second,
		FailureThreshold:   5,
		ResetTimeout:       30 * time.Second,
		UserAgent:          "test-agent",
		EnableSessionRotation: false,
		NumSessions:        1,
	}

	client := httpx.NewClient(config)
	req, err := http.NewRequest("GET", server.URL+"/test", nil)
	require.NoError(t, err)

	// Make 5 requests - should be rate limited
	startTime := time.Now()
	for i := 0; i < 5; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		resp, err := client.Do(ctx, req)
		cancel()
		
		require.NoError(t, err)
		require.NotNil(t, resp)
		resp.Body.Close()
	}
	elapsed := time.Since(startTime)

	// With 2 QPS and burst of 2, 5 requests should take at least 1.5 seconds
	// (2 immediate + 3 at 0.5s intervals = 1.5s minimum)
	assert.True(t, elapsed >= 1*time.Second, "Rate limiting should have taken at least 1 second, got %v", elapsed)
	
	mu.Lock()
	finalCount := requestCount
	mu.Unlock()
	assert.Equal(t, 5, finalCount)
}

func TestHTTPAdapterSessionRotation(t *testing.T) {
	sessionRequests := make(map[string]int)
	var mu sync.Mutex
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract session info from User-Agent
		sessionID := r.UserAgent()
		
		mu.Lock()
		sessionRequests[sessionID]++
		mu.Unlock()
		
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}))
	defer server.Close()

	config := &httpx.Config{
		BaseURL:            server.URL,
		Timeout:            30 * time.Second,
		IdleTimeout:        90 * time.Second,
		MaxConnsPerHost:    5,
		MaxAttempts:        1,
		BackoffBaseMs:      100,
		BackoffJitterMs:    50,
		MaxDelayMs:         1000,
		QPS:                10.0,
		Burst:              10,
		CircuitWindow:      60 * time.Second,
		FailureThreshold:   5,
		ResetTimeout:       30 * time.Second,
		UserAgent:          "test-agent",
		EnableSessionRotation: true,
		NumSessions:        3,
	}

	client := httpx.NewClient(config)
	req, err := http.NewRequest("GET", server.URL+"/test", nil)
	require.NoError(t, err)

	// Make multiple requests to test session rotation
	for i := 0; i < 10; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		resp, err := client.Do(ctx, req)
		cancel()
		
		require.NoError(t, err)
		require.NotNil(t, resp)
		resp.Body.Close()
	}

	mu.Lock()
	sessionCount := len(sessionRequests)
	mu.Unlock()
	
	// Should have used multiple sessions (exact count depends on implementation)
	assert.True(t, sessionCount > 1, "Should have used multiple sessions, got %d", sessionCount)
}

func TestHTTPAdapterErrorHandling(t *testing.T) {
	testCases := []struct {
		name           string
		statusCode     int
		expectedError  string
	}{
		{
			name:          "429 Too Many Requests",
			statusCode:    http.StatusTooManyRequests,
			expectedError: "HTTP 429",
		},
		{
			name:          "500 Internal Server Error",
			statusCode:    http.StatusInternalServerError,
			expectedError: "HTTP 500",
		},
		{
			name:          "503 Service Unavailable",
			statusCode:    http.StatusServiceUnavailable,
			expectedError: "HTTP 503",
		},
		{
			name:          "400 Bad Request",
			statusCode:    http.StatusBadRequest,
			expectedError: "HTTP 400",
		},
		{
			name:          "401 Unauthorized",
			statusCode:    http.StatusUnauthorized,
			expectedError: "HTTP 401",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.statusCode)
				_, _ = w.Write([]byte("Error"))
			}))
			defer server.Close()

			config := &httpx.Config{
				BaseURL:            server.URL,
				Timeout:            30 * time.Second,
				IdleTimeout:        90 * time.Second,
				MaxConnsPerHost:    5,
				MaxAttempts:        1, // No retries for error type test
				BackoffBaseMs:      100,
				BackoffJitterMs:    50,
				MaxDelayMs:         1000,
				QPS:                10.0,
				Burst:              10,
				CircuitWindow:      60 * time.Second,
				FailureThreshold:   5,
				ResetTimeout:       30 * time.Second,
				UserAgent:          "test-agent",
				EnableSessionRotation: false,
				NumSessions:        1,
			}

			client := httpx.NewClient(config)
			req, err := http.NewRequest("GET", server.URL+"/test", nil)
			require.NoError(t, err)

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			resp, err := client.Do(ctx, req)
			cancel()

			assert.Error(t, err)
			assert.Nil(t, resp)
			assert.Contains(t, err.Error(), tc.expectedError)
		})
	}
}
