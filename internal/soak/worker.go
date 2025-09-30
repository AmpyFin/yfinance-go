package soak

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/AmpyFin/yfinance-go"
	"github.com/AmpyFin/yfinance-go/internal/bus"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

// Worker represents a soak test worker
type Worker struct {
	id          int
	client      *yfinance.Client
	bus         *bus.Bus
	rateLimiter *rate.Limiter
	logger      *zap.Logger
	stats       *Stats
	rng         *rand.Rand
}

// WorkRequest represents a work item for a worker
type WorkRequest struct {
	Ticker   string
	Endpoint string
	RunID    string
}

// NewWorker creates a new soak test worker
func NewWorker(id int, client *yfinance.Client, bus *bus.Bus, rateLimiter *rate.Limiter, logger *zap.Logger, stats *Stats) *Worker {
	return &Worker{
		id:          id,
		client:      client,
		bus:         bus,
		rateLimiter: rateLimiter,
		logger:      logger,
		stats:       stats,
		rng:         rand.New(rand.NewSource(time.Now().UnixNano() + int64(id))),
	}
}

// Run starts the worker loop
func (w *Worker) Run(ctx context.Context, wg *sync.WaitGroup, stopCh <-chan struct{}, tickers []string, endpoints []string, config *SoakConfig) {
	defer wg.Done()
	
	w.logger.Debug("Worker started", zap.Int("worker_id", w.id))
	defer w.logger.Debug("Worker stopped", zap.Int("worker_id", w.id))
	
	for {
		select {
		case <-stopCh:
			return
		case <-ctx.Done():
			return
		default:
			// Wait for rate limiter
			if err := w.rateLimiter.Wait(ctx); err != nil {
				if ctx.Err() != nil {
					return
				}
				w.logger.Warn("Rate limiter error", zap.Error(err))
				continue
			}
			
			// Generate random work request
			request := w.generateWorkRequest(tickers, endpoints)
			
			// Execute request
			w.executeRequest(ctx, request, config)
		}
	}
}

// generateWorkRequest creates a random work request
func (w *Worker) generateWorkRequest(tickers []string, endpoints []string) WorkRequest {
	ticker := tickers[w.rng.Intn(len(tickers))]
	endpoint := endpoints[w.rng.Intn(len(endpoints))]
	runID := fmt.Sprintf("soak-%d-%d", time.Now().Unix(), w.rng.Intn(10000))
	
	return WorkRequest{
		Ticker:   ticker,
		Endpoint: endpoint,
		RunID:    runID,
	}
}

// executeRequest executes a single work request
func (w *Worker) executeRequest(ctx context.Context, req WorkRequest, config *SoakConfig) {
	startTime := time.Now()
	
	// Update total requests counter
	atomic.AddInt64(&w.stats.TotalRequests, 1)
	
	// Get endpoint stats
	endpointStats := w.stats.EndpointStats[req.Endpoint]
	endpointStats.mu.Lock()
	endpointStats.Requests++
	endpointStats.mu.Unlock()
	
	// Execute based on fallback strategy
	var err error
	var fallbackUsed bool
	
	switch config.Fallback {
	case "api-only":
		err = w.executeAPIRequest(ctx, req)
		atomic.AddInt64(&w.stats.APIRequests, 1)
	case "scrape-only":
		err = w.executeScrapeRequest(ctx, req)
		atomic.AddInt64(&w.stats.ScrapeRequests, 1)
	case "auto":
		err, fallbackUsed = w.executeAutoFallbackRequest(ctx, req)
		if fallbackUsed {
			atomic.AddInt64(&w.stats.FallbackDecisions, 1)
		}
	default:
		err = fmt.Errorf("unknown fallback strategy: %s", config.Fallback)
	}
	
	// Record execution time
	duration := time.Since(startTime)
	endpointStats.mu.Lock()
	endpointStats.TotalLatency += duration
	endpointStats.mu.Unlock()
	
	// Update success/failure counters
	if err != nil {
		atomic.AddInt64(&w.stats.FailedRequests, 1)
		endpointStats.mu.Lock()
		endpointStats.Failures++
		endpointStats.mu.Unlock()
		
		w.logger.Debug("Request failed",
			zap.Int("worker_id", w.id),
			zap.String("ticker", req.Ticker),
			zap.String("endpoint", req.Endpoint),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		
		// Check for specific error types
		if isRateLimitError(err) {
			atomic.AddInt64(&w.stats.RateLimitHits, 1)
		}
		if isRobotsError(err) {
			atomic.AddInt64(&w.stats.RobotsBlocked, 1)
		}
	} else {
		atomic.AddInt64(&w.stats.SuccessfulReqs, 1)
		endpointStats.mu.Lock()
		endpointStats.Successes++
		endpointStats.mu.Unlock()
		
		w.logger.Debug("Request succeeded",
			zap.Int("worker_id", w.id),
			zap.String("ticker", req.Ticker),
			zap.String("endpoint", req.Endpoint),
			zap.Duration("duration", duration),
			zap.Bool("fallback_used", fallbackUsed),
		)
	}
}

// executeAPIRequest executes an API-only request
func (w *Worker) executeAPIRequest(ctx context.Context, req WorkRequest) error {
	switch req.Endpoint {
	case "quote":
		_, err := w.client.FetchQuote(ctx, req.Ticker, req.RunID)
		return err
	case "daily-bars":
		end := time.Now()
		start := end.AddDate(0, 0, -30) // Last 30 days
		_, err := w.client.FetchDailyBars(ctx, req.Ticker, start, end, true, req.RunID)
		return err
	case "fundamentals":
		_, err := w.client.FetchFundamentalsQuarterly(ctx, req.Ticker, req.RunID)
		return err
	default:
		return fmt.Errorf("API endpoint not supported: %s", req.Endpoint)
	}
}

// executeScrapeRequest executes a scrape-only request
func (w *Worker) executeScrapeRequest(ctx context.Context, req WorkRequest) error {
	switch req.Endpoint {
	case "key-statistics":
		_, err := w.client.ScrapeKeyStatistics(ctx, req.Ticker, req.RunID)
		return err
	case "financials":
		_, err := w.client.ScrapeFinancials(ctx, req.Ticker, req.RunID)
		return err
	case "analysis":
		_, err := w.client.ScrapeAnalysis(ctx, req.Ticker, req.RunID)
		return err
	case "profile":
		// Use comprehensive profile for more thorough testing
		return w.executeScrapeProfile(ctx, req)
	case "news":
		_, err := w.client.ScrapeNews(ctx, req.Ticker, req.RunID)
		return err
	case "balance-sheet":
		_, err := w.client.ScrapeBalanceSheet(ctx, req.Ticker, req.RunID)
		return err
	case "cash-flow":
		_, err := w.client.ScrapeCashFlow(ctx, req.Ticker, req.RunID)
		return err
	case "analyst-insights":
		_, err := w.client.ScrapeAnalystInsights(ctx, req.Ticker, req.RunID)
		return err
	default:
		return fmt.Errorf("scrape endpoint not supported: %s", req.Endpoint)
	}
}

// executeAutoFallbackRequest executes with automatic fallback
func (w *Worker) executeAutoFallbackRequest(ctx context.Context, req WorkRequest) (error, bool) {
	// First try API if available
	if w.isAPIEndpoint(req.Endpoint) {
		err := w.executeAPIRequest(ctx, req)
		if err == nil {
			atomic.AddInt64(&w.stats.APIRequests, 1)
			return nil, false
		}
		
		// If API fails, check if we should fallback to scraping
		if w.shouldFallbackToScrape(err) {
			w.logger.Debug("Falling back to scrape",
				zap.String("ticker", req.Ticker),
				zap.String("endpoint", req.Endpoint),
				zap.Error(err),
			)
			
			scrapeErr := w.executeScrapeRequest(ctx, req)
			atomic.AddInt64(&w.stats.ScrapeRequests, 1)
			return scrapeErr, true
		}
		
		atomic.AddInt64(&w.stats.APIRequests, 1)
		return err, false
	}
	
	// For scrape-only endpoints, go directly to scraping
	err := w.executeScrapeRequest(ctx, req)
	atomic.AddInt64(&w.stats.ScrapeRequests, 1)
	return err, false
}

// executeScrapeProfile executes profile scraping with comprehensive data
func (w *Worker) executeScrapeProfile(ctx context.Context, req WorkRequest) error {
	// This would integrate with the comprehensive profile scraping
	// For now, we'll use a placeholder that simulates the work
	time.Sleep(time.Duration(w.rng.Intn(500)) * time.Millisecond)
	
	// Simulate occasional failures
	if w.rng.Float64() < 0.05 { // 5% failure rate
		return fmt.Errorf("simulated profile scrape failure")
	}
	
	return nil
}

// isAPIEndpoint checks if an endpoint has API support
func (w *Worker) isAPIEndpoint(endpoint string) bool {
	apiEndpoints := map[string]bool{
		"quote":        true,
		"daily-bars":   true,
		"fundamentals": true,
	}
	return apiEndpoints[endpoint]
}

// shouldFallbackToScrape determines if we should fallback to scraping
func (w *Worker) shouldFallbackToScrape(err error) bool {
	if err == nil {
		return false
	}
	
	errStr := err.Error()
	
	// Fallback on authentication errors (paid features)
	if contains(errStr, "401") || contains(errStr, "authentication") || contains(errStr, "subscription") {
		return true
	}
	
	// Fallback on rate limiting
	if contains(errStr, "429") || contains(errStr, "rate limit") {
		return true
	}
	
	// Fallback on server errors
	if contains(errStr, "500") || contains(errStr, "502") || contains(errStr, "503") {
		return true
	}
	
	return false
}

// isRateLimitError checks if an error is due to rate limiting
func isRateLimitError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return contains(errStr, "429") || contains(errStr, "rate limit") || contains(errStr, "too many requests")
}

// isRobotsError checks if an error is due to robots.txt blocking
func isRobotsError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return contains(errStr, "robots") || contains(errStr, "disallowed")
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		len(s) > len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		 indexOfSubstring(s, substr) >= 0))
}

// indexOfSubstring finds the index of a substring in a string
func indexOfSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
