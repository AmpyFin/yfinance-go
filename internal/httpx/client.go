package httpx

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/AmpyFin/yfinance-go/internal/obsv"
)

// Config holds HTTP client configuration
type Config struct {
	BaseURL            string
	Timeout            time.Duration
	IdleTimeout        time.Duration
	MaxConnsPerHost    int
	MaxAttempts        int
	BackoffBaseMs      int
	BackoffJitterMs    int
	MaxDelayMs         int
	QPS                float64
	Burst              int
	CircuitWindow      time.Duration
	FailureThreshold   int
	ResetTimeout       time.Duration
	UserAgent          string
	EnableSessionRotation bool
	NumSessions        int
}

// DefaultConfig returns a sensible default configuration
func DefaultConfig() *Config {
	return &Config{
		BaseURL:            "https://query1.finance.yahoo.com",
		Timeout:            30 * time.Second,
		IdleTimeout:        90 * time.Second,
		MaxConnsPerHost:    10,
		MaxAttempts:        3, // Reduced to avoid overwhelming the API
		BackoffBaseMs:      200, // Increased base delay
		BackoffJitterMs:    100, // Increased jitter
		MaxDelayMs:         10000, // Increased max delay
		QPS:                1.0, // Reduced QPS to be more conservative
		Burst:              3, // Reduced burst size
		CircuitWindow:      60 * time.Second,
		FailureThreshold:   3, // Reduced failure threshold
		ResetTimeout:       30 * time.Second,
		UserAgent:          "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
		EnableSessionRotation: false, // Disabled by default
		NumSessions:        5, // Default number of sessions
	}
}

// SessionRotationConfig returns a configuration optimized for session rotation
func SessionRotationConfig() *Config {
	return &Config{
		BaseURL:            "https://query1.finance.yahoo.com",
		Timeout:            30 * time.Second,
		IdleTimeout:        90 * time.Second,
		MaxConnsPerHost:    10,
		MaxAttempts:        2, // Reduced since we have multiple sessions
		BackoffBaseMs:      100, // Reduced since sessions distribute load
		BackoffJitterMs:    50, // Reduced jitter
		MaxDelayMs:         5000, // Reduced max delay
		QPS:                5.0, // Increased QPS since we have session rotation
		Burst:              10, // Increased burst size
		CircuitWindow:      60 * time.Second,
		FailureThreshold:   5, // Increased failure threshold
		ResetTimeout:       30 * time.Second,
		UserAgent:          "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
		EnableSessionRotation: true, // Enable session rotation
		NumSessions:        7, // Use 7 sessions for good distribution
	}
}

// Client provides a robust HTTP client with retry, backoff, rate limiting, circuit breaker, and session rotation
type Client struct {
	config        *Config
	httpClient    *http.Client
	rateLimiter   *RateLimiter
	circuitBreaker *CircuitBreaker
	sessionManager *SessionManager
}

// NewClient creates a new HTTP client with the given configuration
func NewClient(config *Config) *Client {
	if config == nil {
		config = DefaultConfig()
	}

	// Initialize session manager if session rotation is enabled
	var sessionManager *SessionManager
	if config.EnableSessionRotation {
		sessionManager = NewSessionManager(config.BaseURL, config.NumSessions)
		// Initialize sessions to get initial cookies
		_ = sessionManager.InitializeSessions()
	}

	// Create HTTP client with timeouts and connection pooling
	httpClient := &http.Client{
		Timeout: config.Timeout,
		Transport: &http.Transport{
			IdleConnTimeout:     config.IdleTimeout,
			MaxConnsPerHost:     config.MaxConnsPerHost,
			DisableCompression:  false,
			DisableKeepAlives:   false,
		},
	}

	return &Client{
		config:        config,
		httpClient:    httpClient,
		rateLimiter:   NewRateLimiter(int(config.QPS), config.Burst),
		circuitBreaker: NewCircuitBreaker(config.CircuitWindow, config.FailureThreshold, config.ResetTimeout),
		sessionManager: sessionManager,
	}
}

// Do executes an HTTP request with retry, backoff, rate limiting, and circuit breaker
func (c *Client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	// Set User-Agent
	req.Header.Set("User-Agent", c.config.UserAgent)

	// Extract endpoint from URL path for observability
	endpoint := extractEndpoint(req.URL.Path)
	
	// Start fetch span
	ctx, span := obsv.StartIngestFetchSpan(ctx, endpoint, "", "", req.URL.String(), 0)
	defer span.End()

	// Check circuit breaker
	if !c.circuitBreaker.Allow() {
		obsv.RecordRequest(endpoint, "error", "circuit_open")
		obsv.RecordSpanError(span, ErrCircuitOpen)
		return nil, ErrCircuitOpen
	}

	// Rate limiting
	if err := c.rateLimiter.Wait(ctx); err != nil {
		obsv.RecordRequest(endpoint, "error", "rate_limit")
		obsv.RecordSpanError(span, err)
		return nil, fmt.Errorf("rate limiter: %w", err)
	}

	var lastErr error
	startTime := time.Now()
	
	for attempt := 0; attempt < c.config.MaxAttempts; attempt++ {
		// Get session for this attempt if session rotation is enabled
		var clientToUse *http.Client = c.httpClient
		if c.sessionManager != nil {
			clientToUse = c.sessionManager.GetNextSession()
		}
		
		// Execute request with the selected client (either default or rotated session)
		resp, err := clientToUse.Do(req.WithContext(ctx))
		if err != nil {
			lastErr = err
			c.circuitBreaker.RecordFailure()
			
			// Record retry
			if attempt > 0 {
				obsv.RecordRetry(endpoint, "network_error")
			}
			
			if !c.shouldRetry(err, attempt) {
				obsv.RecordRequest(endpoint, "error", "network_error")
				obsv.RecordRequestLatency(endpoint, time.Since(startTime))
				obsv.RecordSpanError(span, err)
				return nil, err
			}
		} else {
			// Check if response indicates retry
			if c.shouldRetryResponse(resp, attempt) {
				resp.Body.Close()
				lastErr = fmt.Errorf("HTTP %d", resp.StatusCode)
				c.circuitBreaker.RecordFailure()
				
				// Record retry
				if attempt > 0 {
					obsv.RecordRetry(endpoint, fmt.Sprintf("http_%d", resp.StatusCode))
				}
				
				// Don't return here, continue to backoff and retry
			} else {
				// Check if this is actually a success or a failure we can't retry
				if c.isSuccessResponse(resp) {
					// Success
					c.circuitBreaker.RecordSuccess()
					obsv.RecordRequest(endpoint, "success", fmt.Sprintf("%d", resp.StatusCode))
					obsv.RecordRequestLatency(endpoint, time.Since(startTime))
					obsv.UpdateIngestFetchSpan(span, resp.StatusCode, resp.ContentLength, time.Since(startTime))
					return resp, nil
				} else {
					// Failure that we can't retry (e.g., 400, 404, etc.)
					resp.Body.Close()
					lastErr = fmt.Errorf("HTTP %d", resp.StatusCode)
					
					// Don't count 401 errors as circuit breaker failures
					// 401 errors are expected for paid endpoints like fundamentals
					if resp.StatusCode != 401 {
						c.circuitBreaker.RecordFailure()
					}
					
					obsv.RecordRequest(endpoint, "error", fmt.Sprintf("%d", resp.StatusCode))
					obsv.RecordRequestLatency(endpoint, time.Since(startTime))
					obsv.RecordSpanError(span, lastErr)
					return nil, lastErr
				}
			}
		}

		// Calculate backoff delay
		delay := c.calculateBackoff(attempt)
		
		// Record backoff
		obsv.RecordBackoff(endpoint, "retry")
		obsv.RecordBackoffSleep(endpoint, delay)
		
		// Wait with context cancellation support
		select {
		case <-ctx.Done():
			obsv.RecordRequest(endpoint, "error", "context_cancelled")
			obsv.RecordSpanError(span, ctx.Err())
			return nil, ctx.Err()
		case <-time.After(delay):
			// Continue to next attempt
		}
	}

	c.circuitBreaker.RecordFailure()
	obsv.RecordRequest(endpoint, "error", "max_attempts")
	obsv.RecordRequestLatency(endpoint, time.Since(startTime))
	obsv.RecordSpanError(span, fmt.Errorf("max attempts exceeded: %w", lastErr))
	return nil, fmt.Errorf("max attempts exceeded: %w", lastErr)
}

// shouldRetry determines if an error should trigger a retry
func (c *Client) shouldRetry(err error, attempt int) bool {
	if attempt >= c.config.MaxAttempts-1 {
		return false
	}

	// Network errors, timeouts, and context cancellation should be retried
	if err == context.DeadlineExceeded || err == context.Canceled {
		return true
	}

	// Check for network errors (connection refused, etc.)
	if _, ok := err.(*TransportError); ok {
		return true
	}

	return false
}

// shouldRetryResponse determines if an HTTP response should trigger a retry
func (c *Client) shouldRetryResponse(resp *http.Response, attempt int) bool {
	if attempt >= c.config.MaxAttempts-1 {
		return false
	}

	// Retry on server errors and rate limiting
	switch resp.StatusCode {
	case 429, 500, 502, 503, 504:
		return true
	case 400, 401, 403, 404, 422:
		return false // Fatal errors
	default:
		return false
	}
}

// isSuccessResponse determines if an HTTP response represents success
func (c *Client) isSuccessResponse(resp *http.Response) bool {
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

// calculateBackoff calculates the backoff delay with exponential backoff and jitter
func (c *Client) calculateBackoff(attempt int) time.Duration {
	// Exponential backoff: base * 2^attempt
	baseDelay := time.Duration(c.config.BackoffBaseMs) * time.Millisecond
	exponentialDelay := baseDelay * time.Duration(math.Pow(2, float64(attempt)))
	
	// Add jitter
	var jitter time.Duration
	if c.config.BackoffJitterMs > 0 {
		jitter = time.Duration(rand.Intn(c.config.BackoffJitterMs)) * time.Millisecond
	}
	
	// Cap at max delay
	totalDelay := exponentialDelay + jitter
	maxDelay := time.Duration(c.config.MaxDelayMs) * time.Millisecond
	
	if totalDelay > maxDelay {
		totalDelay = maxDelay
	}
	
	return totalDelay
}

// RateLimiter implements a token bucket rate limiter
type RateLimiter struct {
	tokens   float64
	capacity float64
	rate     float64
	lastTime time.Time
	mu       sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(qps, burst int) *RateLimiter {
	return &RateLimiter{
		tokens:   float64(burst),
		capacity: float64(burst),
		rate:     float64(qps),
		lastTime: time.Now(),
	}
}

// Wait blocks until a token is available
func (r *RateLimiter) Wait(ctx context.Context) error {
	r.mu.Lock()
	
	now := time.Now()
	elapsed := now.Sub(r.lastTime)
	
	// Add tokens based on elapsed time
	r.tokens += r.rate * elapsed.Seconds()
	if r.tokens > r.capacity {
		r.tokens = r.capacity
	}
	
	r.lastTime = now
	
	// Check if we have a token available
	if r.tokens >= 1.0 {
		r.tokens -= 1.0
		r.mu.Unlock()
		return nil
	}
	
	// Calculate wait time needed to get a token
	waitTime := time.Duration((1.0-r.tokens)/r.rate * float64(time.Second))
	
	// Release the lock before waiting
	r.mu.Unlock()
	
	// Wait with context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(waitTime):
		// Reacquire lock and consume token
		r.mu.Lock()
		r.tokens = 0.0 // Token consumed by waiting
		r.mu.Unlock()
		return nil
	}
}

// CircuitBreaker implements a circuit breaker pattern
type CircuitBreaker struct {
	window         time.Duration
	failureThreshold int
	resetTimeout   time.Duration
	
	state          CircuitState
	failures       int
	lastFailure    time.Time
	mu             sync.RWMutex
}

// CircuitState represents the state of the circuit breaker
type CircuitState int

const (
	StateClosed CircuitState = iota
	StateOpen
	StateHalfOpen
)

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(window time.Duration, failureThreshold int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		window:           window,
		failureThreshold: failureThreshold,
		resetTimeout:     resetTimeout,
		state:            StateClosed,
	}
}

// Allow checks if the circuit breaker allows the request
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	now := time.Now()
	
	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		// Check if we should transition to half-open
		if now.Sub(cb.lastFailure) >= cb.resetTimeout {
			cb.mu.RUnlock()
			cb.mu.Lock()
			if cb.state == StateOpen && now.Sub(cb.lastFailure) >= cb.resetTimeout {
				cb.state = StateHalfOpen
			}
			cb.mu.Unlock()
			cb.mu.RLock()
			return cb.state == StateHalfOpen
		}
		return false
	case StateHalfOpen:
		return true
	default:
		return false
	}
}

// RecordSuccess records a successful request
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == StateHalfOpen {
		cb.state = StateClosed
		cb.failures = 0
	}
}

// RecordFailure records a failed request
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	cb.lastFailure = time.Now()
	
	if cb.failures >= cb.failureThreshold {
		cb.state = StateOpen
	}
}

// State returns the current state of the circuit breaker
func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Failures returns the current failure count (for testing)
func (cb *CircuitBreaker) Failures() int {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.failures
}

// GetSessionStats returns session usage statistics
func (c *Client) GetSessionStats() map[string]interface{} {
	if c.sessionManager == nil {
		return map[string]interface{}{
			"session_rotation_enabled": false,
		}
	}
	
	stats := c.sessionManager.GetSessionStats()
	stats["session_rotation_enabled"] = true
	return stats
}

// extractEndpoint extracts the endpoint name from a URL path
func extractEndpoint(path string) string {
	// Map common Yahoo Finance paths to endpoint names
	switch {
	case strings.Contains(path, "/v8/finance/chart/"):
		return "bars_1d"
	case strings.Contains(path, "/v1/finance/quote"):
		return "quote"
	case strings.Contains(path, "/v11/finance/quoteSummary/"):
		return "fundamentals"
	case strings.Contains(path, "/v1/finance/quoteSummary/"):
		return "fundamentals"
	default:
		return "unknown"
	}
}
