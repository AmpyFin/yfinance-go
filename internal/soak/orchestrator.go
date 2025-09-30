package soak

import (
	"bufio"
	"context"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/AmpyFin/yfinance-go"
	"github.com/AmpyFin/yfinance-go/internal/bus"
	"github.com/AmpyFin/yfinance-go/internal/config"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

// SoakConfig holds configuration for soak testing
type SoakConfig struct {
	UniverseFile  string
	Endpoints     string
	Fallback      string
	Duration      time.Duration
	Concurrency   int
	QPS           float64
	Preview       bool
	Publish       bool
	Env           string
	TopicPrefix   string
	ProbeInterval time.Duration
	FailureRate   float64
	MemoryCheck   bool
}

// Orchestrator manages the soak test execution
type Orchestrator struct {
	config        *config.Config
	soakConfig    *SoakConfig
	client        *yfinance.Client
	bus           *bus.Bus
	logger        *zap.Logger
	metrics       *Metrics
	probes        *CorrectnessProbes
	memMonitor    *MemoryMonitor
	failureServer *FailureServer

	// Runtime state
	tickers     []string
	endpoints   []string
	rateLimiter *rate.Limiter
	workers     []*Worker
	stats       *Stats

	// Control channels
	stopCh chan struct{}
	doneCh chan struct{}
	wg     sync.WaitGroup
}

// Stats tracks soak test statistics
type Stats struct {
	StartTime         time.Time
	TotalRequests     int64
	SuccessfulReqs    int64
	FailedRequests    int64
	APIRequests       int64
	ScrapeRequests    int64
	FallbackDecisions int64
	RateLimitHits     int64
	RobotsBlocked     int64
	ProbesPassed      int64
	ProbesFailed      int64

	// Per-endpoint stats
	EndpointStats map[string]*EndpointStats

	// Memory stats
	InitialMemory uint64
	PeakMemory    uint64

	// Goroutine stats
	InitialGoroutines int
	PeakGoroutines    int

	mu sync.RWMutex
}

// EndpointStats tracks per-endpoint statistics
type EndpointStats struct {
	Requests     int64
	Successes    int64
	Failures     int64
	AvgLatency   time.Duration
	TotalLatency time.Duration
	mu           sync.RWMutex
}

// NewOrchestrator creates a new soak test orchestrator
func NewOrchestrator(cfg *config.Config, soakCfg *SoakConfig) (*Orchestrator, error) {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	// Load ticker universe
	tickers, err := loadTickerUniverse(soakCfg.UniverseFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load ticker universe: %w", err)
	}

	// Parse endpoints
	endpoints := parseEndpoints(soakCfg.Endpoints)

	// Create client with session rotation for better resilience
	client := yfinance.NewClientWithSessionRotation()

	// Create bus if publishing is enabled
	var busInstance *bus.Bus
	if soakCfg.Publish {
		busConfig := &bus.Config{
			Enabled:     true,
			Env:         soakCfg.Env,
			TopicPrefix: soakCfg.TopicPrefix,
		}
		busInstance, err = bus.NewBus(busConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create bus: %w", err)
		}
	}

	// Initialize metrics
	metrics := NewMetrics()

	// Initialize correctness probes
	probes := NewCorrectnessProbes(client, logger)

	// Initialize memory monitor
	memMonitor := NewMemoryMonitor(logger)

	// Initialize failure injection server
	failureServer := NewFailureServer(soakCfg.FailureRate, logger)

	// Create rate limiter
	rateLimiter := rate.NewLimiter(rate.Limit(soakCfg.QPS), int(soakCfg.QPS)+1)

	// Initialize stats
	stats := &Stats{
		StartTime:     time.Now(),
		EndpointStats: make(map[string]*EndpointStats),
	}

	// Initialize per-endpoint stats
	for _, endpoint := range endpoints {
		stats.EndpointStats[endpoint] = &EndpointStats{}
	}

	return &Orchestrator{
		config:        cfg,
		soakConfig:    soakCfg,
		client:        client,
		bus:           busInstance,
		logger:        logger,
		metrics:       metrics,
		probes:        probes,
		memMonitor:    memMonitor,
		failureServer: failureServer,
		tickers:       tickers,
		endpoints:     endpoints,
		rateLimiter:   rateLimiter,
		stats:         stats,
		stopCh:        make(chan struct{}),
		doneCh:        make(chan struct{}),
	}, nil
}

// Run executes the soak test
func (o *Orchestrator) Run(ctx context.Context) error {
	o.logger.Info("Starting soak test",
		zap.String("universe_file", o.soakConfig.UniverseFile),
		zap.Int("ticker_count", len(o.tickers)),
		zap.Strings("endpoints", o.endpoints),
		zap.Duration("duration", o.soakConfig.Duration),
		zap.Int("concurrency", o.soakConfig.Concurrency),
		zap.Float64("qps", o.soakConfig.QPS),
		zap.Bool("publish", o.soakConfig.Publish),
	)

	// Record initial memory and goroutines
	if o.soakConfig.MemoryCheck {
		o.recordInitialState()
	}

	// Start failure injection server
	if err := o.failureServer.Start(); err != nil {
		return fmt.Errorf("failed to start failure server: %w", err)
	}
	defer func() {
		_ = o.failureServer.Stop()
	}()

	// Start memory monitoring
	if o.soakConfig.MemoryCheck {
		o.wg.Add(1)
		go o.memMonitor.Monitor(ctx, &o.wg, o.stopCh)
	}

	// Start correctness probes
	o.wg.Add(1)
	go o.runCorrectnessProbes(ctx)

	// Start metrics collection
	o.wg.Add(1)
	go o.collectMetrics(ctx)

	// Start workers
	o.startWorkers(ctx)

	// Wait for duration or cancellation
	select {
	case <-time.After(o.soakConfig.Duration):
		o.logger.Info("Soak test duration completed")
	case <-ctx.Done():
		o.logger.Info("Soak test canceled")
	}

	// Stop all workers
	close(o.stopCh)
	o.wg.Wait()
	close(o.doneCh)

	// Print final results
	o.printResults()

	return nil
}

// startWorkers initializes and starts worker goroutines
func (o *Orchestrator) startWorkers(ctx context.Context) {
	o.workers = make([]*Worker, o.soakConfig.Concurrency)

	for i := 0; i < o.soakConfig.Concurrency; i++ {
		worker := NewWorker(i, o.client, o.bus, o.rateLimiter, o.logger, o.stats)
		o.workers[i] = worker

		o.wg.Add(1)
		go worker.Run(ctx, &o.wg, o.stopCh, o.tickers, o.endpoints, o.soakConfig)
	}
}

// runCorrectnessProbes runs periodic correctness validation
func (o *Orchestrator) runCorrectnessProbes(ctx context.Context) {
	defer o.wg.Done()

	ticker := time.NewTicker(o.soakConfig.ProbeInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			o.runProbeRound(ctx)
		case <-o.stopCh:
			return
		case <-ctx.Done():
			return
		}
	}
}

// runProbeRound executes a round of correctness probes
func (o *Orchestrator) runProbeRound(ctx context.Context) {
	// Select random sample of tickers for probing
	sampleSize := minInt(10, len(o.tickers))
	sample := make([]string, sampleSize)

	// Random sampling without replacement
	indices := rand.Perm(len(o.tickers))
	for i := 0; i < sampleSize; i++ {
		sample[i] = o.tickers[indices[i]]
	}

	o.logger.Info("Running correctness probe round", zap.Strings("sample", sample))

	for _, ticker := range sample {
		if err := o.probes.ValidateTicker(ctx, ticker); err != nil {
			o.logger.Warn("Correctness probe failed",
				zap.String("ticker", ticker),
				zap.Error(err))
			atomic.AddInt64(&o.stats.ProbesFailed, 1)
		} else {
			atomic.AddInt64(&o.stats.ProbesPassed, 1)
		}
	}
}

// collectMetrics periodically collects and logs metrics
func (o *Orchestrator) collectMetrics(ctx context.Context) {
	defer o.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			o.logCurrentMetrics()
		case <-o.stopCh:
			return
		case <-ctx.Done():
			return
		}
	}
}

// logCurrentMetrics logs current performance metrics
func (o *Orchestrator) logCurrentMetrics() {
	o.stats.mu.RLock()
	defer o.stats.mu.RUnlock()

	elapsed := time.Since(o.stats.StartTime)
	totalReqs := atomic.LoadInt64(&o.stats.TotalRequests)
	successReqs := atomic.LoadInt64(&o.stats.SuccessfulReqs)
	failedReqs := atomic.LoadInt64(&o.stats.FailedRequests)

	successRate := float64(0)
	if totalReqs > 0 {
		successRate = float64(successReqs) / float64(totalReqs) * 100
	}

	qps := float64(totalReqs) / elapsed.Seconds()

	o.logger.Info("Soak test metrics",
		zap.Duration("elapsed", elapsed),
		zap.Int64("total_requests", totalReqs),
		zap.Int64("successful_requests", successReqs),
		zap.Int64("failed_requests", failedReqs),
		zap.Float64("success_rate", successRate),
		zap.Float64("actual_qps", qps),
		zap.Int("goroutines", runtime.NumGoroutine()),
	)
}

// recordInitialState records initial memory and goroutine state
func (o *Orchestrator) recordInitialState() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	o.stats.mu.Lock()
	o.stats.InitialMemory = m.Alloc
	o.stats.PeakMemory = m.Alloc
	o.stats.InitialGoroutines = runtime.NumGoroutine()
	o.stats.PeakGoroutines = runtime.NumGoroutine()
	o.stats.mu.Unlock()
}

// printResults prints comprehensive soak test results
func (o *Orchestrator) printResults() {
	o.stats.mu.RLock()
	defer o.stats.mu.RUnlock()

	elapsed := time.Since(o.stats.StartTime)
	totalReqs := atomic.LoadInt64(&o.stats.TotalRequests)
	successReqs := atomic.LoadInt64(&o.stats.SuccessfulReqs)
	failedReqs := atomic.LoadInt64(&o.stats.FailedRequests)

	fmt.Printf("\n=== SOAK TEST RESULTS ===\n")
	fmt.Printf("Duration: %v\n", elapsed)
	fmt.Printf("Total Requests: %d\n", totalReqs)
	fmt.Printf("Successful Requests: %d\n", successReqs)
	fmt.Printf("Failed Requests: %d\n", failedReqs)

	if totalReqs > 0 {
		successRate := float64(successReqs) / float64(totalReqs) * 100
		fmt.Printf("Success Rate: %.2f%%\n", successRate)
		fmt.Printf("Actual QPS: %.2f\n", float64(totalReqs)/elapsed.Seconds())
	}

	fmt.Printf("API Requests: %d\n", atomic.LoadInt64(&o.stats.APIRequests))
	fmt.Printf("Scrape Requests: %d\n", atomic.LoadInt64(&o.stats.ScrapeRequests))
	fmt.Printf("Fallback Decisions: %d\n", atomic.LoadInt64(&o.stats.FallbackDecisions))
	fmt.Printf("Rate Limit Hits: %d\n", atomic.LoadInt64(&o.stats.RateLimitHits))
	fmt.Printf("Robots Blocked: %d\n", atomic.LoadInt64(&o.stats.RobotsBlocked))

	probesPassed := atomic.LoadInt64(&o.stats.ProbesPassed)
	probesFailed := atomic.LoadInt64(&o.stats.ProbesFailed)
	fmt.Printf("Correctness Probes Passed: %d\n", probesPassed)
	fmt.Printf("Correctness Probes Failed: %d\n", probesFailed)

	if o.soakConfig.MemoryCheck {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		fmt.Printf("\n=== MEMORY ANALYSIS ===\n")
		fmt.Printf("Initial Memory: %d KB\n", o.stats.InitialMemory/1024)
		fmt.Printf("Peak Memory: %d KB\n", o.stats.PeakMemory/1024)
		fmt.Printf("Final Memory: %d KB\n", m.Alloc/1024)
		fmt.Printf("Memory Growth: %d KB\n", int64(m.Alloc-o.stats.InitialMemory)/1024)

		fmt.Printf("Initial Goroutines: %d\n", o.stats.InitialGoroutines)
		fmt.Printf("Peak Goroutines: %d\n", o.stats.PeakGoroutines)
		fmt.Printf("Final Goroutines: %d\n", runtime.NumGoroutine())
		fmt.Printf("Goroutine Growth: %d\n", runtime.NumGoroutine()-o.stats.InitialGoroutines)
	}

	fmt.Printf("\n=== ENDPOINT BREAKDOWN ===\n")
	for endpoint, stats := range o.stats.EndpointStats {
		stats.mu.RLock()
		avgLatency := time.Duration(0)
		if stats.Requests > 0 {
			avgLatency = time.Duration(stats.TotalLatency.Nanoseconds() / stats.Requests)
		}
		fmt.Printf("%s: %d requests, %d successes, %d failures, avg latency: %v\n",
			endpoint, stats.Requests, stats.Successes, stats.Failures, avgLatency)
		stats.mu.RUnlock()
	}
}

// Close cleans up resources
func (o *Orchestrator) Close() {
	if o.bus != nil {
		o.bus.Close(context.Background())
	}
	if o.logger != nil {
		_ = o.logger.Sync()
	}
}

// loadTickerUniverse loads tickers from a file
func loadTickerUniverse(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open universe file: %w", err)
	}
	defer file.Close()

	var tickers []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			tickers = append(tickers, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read universe file: %w", err)
	}

	if len(tickers) == 0 {
		return nil, fmt.Errorf("no tickers found in universe file")
	}

	return tickers, nil
}

// parseEndpoints parses comma-separated endpoint list
func parseEndpoints(endpointsStr string) []string {
	endpoints := strings.Split(endpointsStr, ",")
	for i, endpoint := range endpoints {
		endpoints[i] = strings.TrimSpace(endpoint)
	}
	return endpoints
}

// min returns the minimum of two integers
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
