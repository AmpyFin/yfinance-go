package scrape

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/AmpyFin/yfinance-go/internal/httpx"
)

// Client interface for web scraping operations
type Client interface {
	Fetch(ctx context.Context, url string) ([]byte, *FetchMeta, error)
}

// client implements the Client interface
type client struct {
	config        *Config
	httpClient    *httpx.Client
	rateLimiter   *RateLimiter
	robotsManager *RobotsManager
	backoffPolicy *BackoffPolicy
	metrics       *Metrics
	logger        *Logger
	tracer        *Tracer
}

// NewClient creates a new scraping client
func NewClient(config *Config, httpxPool *httpx.Client) *client {
	if config == nil {
		config = DefaultConfig()
	}

	// Create HTTP client if not provided
	var httpClient *httpx.Client
	if httpxPool != nil {
		httpClient = httpxPool
	} else {
		// Create a new httpx client with scraping-optimized config
		httpxConfig := &httpx.Config{
			BaseURL:            "https://finance.yahoo.com",
			Timeout:            time.Duration(config.TimeoutMs) * time.Millisecond,
			IdleTimeout:        90 * time.Second,
			MaxConnsPerHost:    10,
			MaxAttempts:        config.Retry.Attempts,
			BackoffBaseMs:      config.Retry.BaseMs,
			BackoffJitterMs:    config.Retry.BaseMs / 2,
			MaxDelayMs:         config.Retry.MaxDelayMs,
			QPS:                config.QPS,
			Burst:              config.Burst,
			CircuitWindow:      60 * time.Second,
			FailureThreshold:   5,
			ResetTimeout:       30 * time.Second,
			UserAgent:          config.UserAgent,
			EnableSessionRotation: true,
			NumSessions:        3,
		}
		httpClient = httpx.NewClient(httpxConfig)
	}

	// Initialize components
	rateLimiter := NewRateLimiter(config.QPS, config.Burst)
	robotsManager := NewRobotsManager(config.RobotsPolicy, time.Duration(config.CacheTTLMs)*time.Millisecond)
	backoffPolicy := DefaultBackoffPolicy()
	metrics := NewMetrics()
	logger := NewLogger()
	tracer := NewTracer()

	return &client{
		config:        config,
		httpClient:    httpClient,
		rateLimiter:   rateLimiter,
		robotsManager: robotsManager,
		backoffPolicy: backoffPolicy,
		metrics:       metrics,
		logger:        logger,
		tracer:        tracer,
	}
}

// Fetch retrieves content from a URL with proper error handling, rate limiting, and observability
func (c *client) Fetch(ctx context.Context, urlStr string) ([]byte, *FetchMeta, error) {
	// Parse URL to extract host
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, nil, &ScrapeError{
			Type:    "invalid_url",
			Message: fmt.Sprintf("failed to parse URL: %v", err),
			URL:     urlStr,
		}
	}

	host := parsedURL.Host
	startTime := time.Now()

	// Start tracing span
	ctx, span := c.tracer.StartFetchSpan(ctx, urlStr, host)
	defer func() {
		if span != nil {
			c.tracer.EndSpan(span)
		}
	}()

	// Check robots.txt policy
	if err := c.robotsManager.CheckRobots(ctx, host, parsedURL.Path); err != nil {
		c.metrics.RecordRobotsDenied(host)
		c.logger.LogRobotsDenied(urlStr, host, err.Error())
		c.tracer.RecordSpanError(span, err)
		return nil, nil, err
	}

	// Rate limiting
	if err := c.rateLimiter.Wait(ctx); err != nil {
		c.metrics.RecordRequest(host, "error", "rate_limit")
		c.logger.LogRateLimit(urlStr, host, err.Error())
		c.tracer.RecordSpanError(span, err)
		return nil, nil, fmt.Errorf("rate limiter: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		return nil, nil, &ScrapeError{
			Type:    "invalid_request",
			Message: fmt.Sprintf("failed to create request: %v", err),
			URL:     urlStr,
		}
	}

	// Set browser-like headers
	c.setBrowserHeaders(req)

	// Execute request with retries
	var fetchMeta *FetchMeta

	for attempt := 0; attempt < c.config.Retry.Attempts; attempt++ {
		attemptStart := time.Now()
		
		// Execute HTTP request
		resp, err := c.httpClient.Do(ctx, req)
		if err != nil {
			c.metrics.RecordRetry(host, "network_error")
			c.logger.LogRetry(urlStr, host, attempt+1, "network_error", err.Error())
			
			if !IsRetryableError(err) || attempt >= c.config.Retry.Attempts-1 {
				c.metrics.RecordRequest(host, "error", "network_error")
				c.logger.LogRequest(urlStr, host, 0, attempt+1, time.Since(attemptStart), 0, false, 0, err.Error())
				c.tracer.RecordSpanError(span, err)
				return nil, nil, err
			}
		} else {
			// Process response
			body, meta, err := c.processResponse(resp, urlStr, host, attempt+1, time.Since(attemptStart))
			if err != nil {
				resp.Body.Close()
				
				if !IsRetryableError(err) || attempt >= c.config.Retry.Attempts-1 {
					c.metrics.RecordRequest(host, "error", fmt.Sprintf("http_%d", meta.Status))
					c.logger.LogRequest(urlStr, host, meta.Status, attempt+1, meta.Duration, meta.Bytes, meta.Gzip, meta.Redirects, err.Error())
					c.tracer.RecordSpanError(span, err)
					return nil, nil, err
				}
				
				c.metrics.RecordRetry(host, fmt.Sprintf("http_%d", meta.Status))
				c.logger.LogRetry(urlStr, host, attempt+1, fmt.Sprintf("http_%d", meta.Status), err.Error())
			} else {
				// Success
				fetchMeta = meta
				fetchMeta.Duration = time.Since(startTime)
				fetchMeta.RobotsPolicy = c.config.RobotsPolicy
				
				c.metrics.RecordRequest(host, "success", fmt.Sprintf("%d", meta.Status))
				c.metrics.RecordLatency(host, meta.Duration)
				c.metrics.RecordPageBytes(host, meta.Bytes)
				c.logger.LogRequest(urlStr, host, meta.Status, attempt+1, meta.Duration, meta.Bytes, meta.Gzip, meta.Redirects, "")
				c.tracer.UpdateSpan(span, meta.Status, meta.Bytes, meta.Duration)
				
				return body, fetchMeta, nil
			}
		}

		// Calculate backoff delay
		delay := c.backoffPolicy.CalculateDelay(attempt)
		c.metrics.RecordBackoff(host, "retry")
		c.logger.LogBackoff(urlStr, host, delay)
		
		// Wait with context cancellation support
		select {
		case <-ctx.Done():
			c.metrics.RecordRequest(host, "error", "context_cancelled")
			c.logger.LogRequest(urlStr, host, 0, attempt+1, time.Since(startTime), 0, false, 0, ctx.Err().Error())
			c.tracer.RecordSpanError(span, ctx.Err())
			return nil, nil, ctx.Err()
		case <-time.After(delay):
			// Continue to next attempt
		}
	}

	// All retries exhausted
	c.metrics.RecordRequest(host, "error", "max_attempts")
	c.logger.LogRequest(urlStr, host, 0, c.config.Retry.Attempts, time.Since(startTime), 0, false, 0, "max attempts exceeded")
	c.tracer.RecordSpanError(span, ErrRetryExhausted)
	return nil, nil, ErrRetryExhausted
}

// setBrowserHeaders sets browser-like headers on the request
func (c *client) setBrowserHeaders(req *http.Request) {
	req.Header.Set("User-Agent", c.config.UserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Cache-Control", "max-age=0")
}

// processResponse processes the HTTP response and extracts content
func (c *client) processResponse(resp *http.Response, urlStr, host string, attempt int, duration time.Duration) ([]byte, *FetchMeta, error) {
	defer resp.Body.Close()

	// Create fetch metadata
	meta := &FetchMeta{
		URL:       urlStr,
		Host:      host,
		Status:    resp.StatusCode,
		Attempt:   attempt,
		Duration:  duration,
		FromCache: false,
	}

	// Check for redirects
	if resp.Header.Get("Location") != "" {
		meta.Redirects = 1 // Simplified - in practice you'd track the full redirect chain
	}

	// Check if response is successful
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if resp.StatusCode == 429 || (resp.StatusCode >= 500 && resp.StatusCode < 600) {
			return nil, meta, ErrHTTP(resp.StatusCode, urlStr)
		}
		return nil, meta, ErrHTTP(resp.StatusCode, urlStr)
	}

	// Check content length
	if resp.ContentLength > 0 {
		meta.Bytes = int(resp.ContentLength)
	}

	// Check for gzip encoding
	contentEncoding := resp.Header.Get("Content-Encoding")
	meta.Gzip = strings.Contains(contentEncoding, "gzip")

	// Read response body with size limit
	maxSize := 8 * 1024 * 1024 // 8 MiB limit
	body, err := c.readResponseBody(resp, maxSize)
	if err != nil {
		return nil, meta, err
	}

	meta.Bytes = len(body)
	return body, meta, nil
}

// readResponseBody reads the response body with size limits and gzip support
func (c *client) readResponseBody(resp *http.Response, maxSize int) ([]byte, error) {
	var reader io.Reader = resp.Body

	// Handle gzip decompression
	contentEncoding := resp.Header.Get("Content-Encoding")
	if strings.Contains(contentEncoding, "gzip") {
		gzReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzReader.Close()
		reader = gzReader
	}

	// Create a limited reader
	limitedReader := &limitedReader{
		Reader: reader,
		Limit:  maxSize,
	}

	// Read all content
	body, err := limitedReader.ReadAll()
	if err != nil {
		if err == ErrContentTooLarge {
			return nil, err
		}
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return body, nil
}

// limitedReader implements a reader with size limits
type limitedReader struct {
	Reader interface{ Read([]byte) (int, error) }
	Limit  int
	Read   int
}

// ReadAll reads all content from the limited reader
func (lr *limitedReader) ReadAll() ([]byte, error) {
	var result []byte
	buffer := make([]byte, 32*1024) // 32KB buffer

	for {
		n, err := lr.Reader.Read(buffer)
		if n > 0 {
			lr.Read += n
			if lr.Read > lr.Limit {
				return nil, ErrContentTooLarge
			}
			result = append(result, buffer[:n]...)
		}
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return result, err
		}
	}

	return result, nil
}

// RateLimiter implements per-host rate limiting
type RateLimiter struct {
	limiters map[string]*tokenBucket
	mu       sync.RWMutex
	qps      float64
	burst    int
}

// tokenBucket implements a token bucket rate limiter
type tokenBucket struct {
	tokens   float64
	capacity float64
	rate     float64
	lastTime time.Time
	mu       sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(qps float64, burst int) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*tokenBucket),
		qps:      qps,
		burst:    burst,
	}
}

// Wait blocks until a token is available for the given host
func (rl *RateLimiter) Wait(ctx context.Context) error {
	// For now, we'll use a global rate limiter
	// In a full implementation, this would be per-host
	rl.mu.Lock()
	bucket, exists := rl.limiters["global"]
	if !exists {
		bucket = &tokenBucket{
			tokens:   float64(rl.burst),
			capacity: float64(rl.burst),
			rate:     rl.qps,
			lastTime: time.Now(),
		}
		rl.limiters["global"] = bucket
	}
	rl.mu.Unlock()

	return bucket.Wait(ctx)
}

// Wait blocks until a token is available
func (tb *tokenBucket) Wait(ctx context.Context) error {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastTime)

	// Add tokens based on elapsed time
	tb.tokens += tb.rate * elapsed.Seconds()
	if tb.tokens > tb.capacity {
		tb.tokens = tb.capacity
	}

	tb.lastTime = now

	// Check if we have a token available
	if tb.tokens >= 1.0 {
		tb.tokens -= 1.0
		return nil
	}

	// Calculate wait time needed to get a token
	waitTime := time.Duration((1.0-tb.tokens)/tb.rate * float64(time.Second))

	// Wait with context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(waitTime):
		tb.tokens = 0.0 // Token consumed by waiting
		return nil
	}
}
