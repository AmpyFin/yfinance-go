# Performance Optimization & Best Practices

This guide provides comprehensive information about performance optimization, best practices, and production-ready configurations for yfinance-go.

## Table of Contents

1. [Performance Overview](#performance-overview)
2. [Client Configuration](#client-configuration)
3. [Concurrency & Rate Limiting](#concurrency--rate-limiting)
4. [Caching Strategies](#caching-strategies)
5. [Batch Processing](#batch-processing)
6. [Memory Optimization](#memory-optimization)
7. [Network Optimization](#network-optimization)
8. [Production Best Practices](#production-best-practices)
9. [Monitoring & Metrics](#monitoring--metrics)

## Performance Overview

### Performance Characteristics

| Operation | Typical Latency | Throughput | Notes |
|-----------|----------------|------------|-------|
| **FetchQuote()** | 100-500ms | 10-50 req/s | Fast, API-based |
| **FetchDailyBars()** | 200-800ms | 5-20 req/s | Depends on date range |
| **FetchCompanyInfo()** | 100-400ms | 10-50 req/s | Fast, API-based |
| **ScrapeFinancials()** | 1-3s | 1-5 req/s | Slower, scraping-based |
| **ScrapeNews()** | 500ms-2s | 2-10 req/s | Variable, depends on content |
| **ScrapeKeyStatistics()** | 800ms-2s | 2-8 req/s | Slower, scraping-based |

### Performance Factors

1. **Network Latency**: Distance to Yahoo Finance servers
2. **Rate Limiting**: Yahoo Finance's rate limits
3. **Data Size**: Amount of data requested
4. **Concurrency**: Number of concurrent requests
5. **Client Configuration**: Timeouts, retries, session rotation

## Client Configuration

### 1. Production Configuration

```go
// Production-ready configuration
func createProductionClient() *yfinance.Client {
    config := &httpx.Config{
        // Timeouts
        Timeout:     30 * time.Second,
        DialTimeout: 10 * time.Second,
        
        // Rate limiting
        QPS:         2.0,  // 2 requests per second
        Burst:       5,    // Allow burst of 5 requests
        
        // Retry configuration
        MaxAttempts: 3,
        BackoffBase: 1 * time.Second,
        BackoffMax:  10 * time.Second,
        
        // Connection pooling
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
        
        // User agent
        UserAgent: "MyApp/1.0 (contact@mycompany.com)",
    }
    
    return yfinance.NewClientWithConfig(config)
}
```

### 2. High-Volume Configuration

```go
// High-volume configuration with session rotation
func createHighVolumeClient() *yfinance.Client {
    config := &httpx.Config{
        // Aggressive timeouts for high volume
        Timeout:     15 * time.Second,
        DialTimeout: 5 * time.Second,
        
        // Higher rate limits
        QPS:         5.0,  // 5 requests per second
        Burst:       10,   // Allow burst of 10 requests
        
        // More retries
        MaxAttempts: 5,
        BackoffBase: 500 * time.Millisecond,
        BackoffMax:  5 * time.Second,
        
        // Larger connection pool
        MaxIdleConns:        200,
        MaxIdleConnsPerHost: 20,
        IdleConnTimeout:     60 * time.Second,
        
        // Session rotation
        EnableSessionRotation: true,
        SessionRotationInterval: 5 * time.Minute,
    }
    
    return yfinance.NewClientWithConfig(config)
}
```

### 3. Development Configuration

```go
// Development configuration with relaxed limits
func createDevelopmentClient() *yfinance.Client {
    config := &httpx.Config{
        // Longer timeouts for debugging
        Timeout:     60 * time.Second,
        DialTimeout: 30 * time.Second,
        
        // Conservative rate limiting
        QPS:         1.0,  // 1 request per second
        Burst:       2,    // Allow burst of 2 requests
        
        // More retries for debugging
        MaxAttempts: 5,
        BackoffBase: 2 * time.Second,
        BackoffMax:  30 * time.Second,
        
        // Smaller connection pool
        MaxIdleConns:        50,
        MaxIdleConnsPerHost: 5,
        IdleConnTimeout:     120 * time.Second,
    }
    
    return yfinance.NewClientWithConfig(config)
}
```

## Concurrency & Rate Limiting

### 1. Controlled Concurrency

```go
type ConcurrentProcessor struct {
    client      *yfinance.Client
    semaphore   chan struct{}
    rateLimiter *time.Ticker
}

func NewConcurrentProcessor(client *yfinance.Client, maxConcurrency int, rateLimit time.Duration) *ConcurrentProcessor {
    return &ConcurrentProcessor{
        client:      client,
        semaphore:   make(chan struct{}, maxConcurrency),
        rateLimiter: time.NewTicker(rateLimit),
    }
}

func (cp *ConcurrentProcessor) ProcessSymbol(symbol string) (*yfinance.NormalizedQuote, error) {
    // Acquire semaphore
    cp.semaphore <- struct{}{}
    defer func() { <-cp.semaphore }()
    
    // Wait for rate limit
    <-cp.rateLimiter.C
    
    // Process symbol
    ctx := context.Background()
    runID := fmt.Sprintf("concurrent-%d", time.Now().Unix())
    return cp.client.FetchQuote(ctx, symbol, runID)
}

func (cp *ConcurrentProcessor) ProcessSymbols(symbols []string) []*yfinance.NormalizedQuote {
    results := make([]*yfinance.NormalizedQuote, len(symbols))
    var wg sync.WaitGroup
    
    for i, symbol := range symbols {
        wg.Add(1)
        go func(index int, sym string) {
            defer wg.Done()
            
            quote, err := cp.ProcessSymbol(sym)
            if err != nil {
                log.Printf("Error processing %s: %v", sym, err)
                return
            }
            
            results[index] = quote
        }(i, symbol)
    }
    
    wg.Wait()
    return results
}
```

### 2. Adaptive Rate Limiting

```go
type AdaptiveRateLimiter struct {
    currentRate    float64
    minRate        float64
    maxRate        float64
    successCount   int
    failureCount   int
    lastAdjustment time.Time
    mutex          sync.RWMutex
}

func NewAdaptiveRateLimiter(minRate, maxRate float64) *AdaptiveRateLimiter {
    return &AdaptiveRateLimiter{
        currentRate:    minRate,
        minRate:        minRate,
        maxRate:        maxRate,
        lastAdjustment: time.Now(),
    }
}

func (arl *AdaptiveRateLimiter) RecordSuccess() {
    arl.mutex.Lock()
    defer arl.mutex.Unlock()
    
    arl.successCount++
    arl.adjustRate()
}

func (arl *AdaptiveRateLimiter) RecordFailure() {
    arl.mutex.Lock()
    defer arl.mutex.Unlock()
    
    arl.failureCount++
    arl.adjustRate()
}

func (arl *AdaptiveRateLimiter) adjustRate() {
    // Only adjust every 10 requests
    if arl.successCount+arl.failureCount < 10 {
        return
    }
    
    successRate := float64(arl.successCount) / float64(arl.successCount+arl.failureCount)
    
    if successRate > 0.95 && arl.currentRate < arl.maxRate {
        // Increase rate
        arl.currentRate = math.Min(arl.currentRate*1.1, arl.maxRate)
    } else if successRate < 0.8 && arl.currentRate > arl.minRate {
        // Decrease rate
        arl.currentRate = math.Max(arl.currentRate*0.9, arl.minRate)
    }
    
    // Reset counters
    arl.successCount = 0
    arl.failureCount = 0
    arl.lastAdjustment = time.Now()
}

func (arl *AdaptiveRateLimiter) GetRate() float64 {
    arl.mutex.RLock()
    defer arl.mutex.RUnlock()
    return arl.currentRate
}
```

### 3. Circuit Breaker Pattern

```go
type CircuitBreaker struct {
    failureCount   int
    successCount   int
    lastFailure    time.Time
    threshold      int
    timeout        time.Duration
    state          string // "closed", "open", "half-open"
    mutex          sync.RWMutex
}

func NewCircuitBreaker(threshold int, timeout time.Duration) *CircuitBreaker {
    return &CircuitBreaker{
        threshold: threshold,
        timeout:   timeout,
        state:     "closed",
    }
}

func (cb *CircuitBreaker) Call(fn func() error) error {
    cb.mutex.Lock()
    defer cb.mutex.Unlock()
    
    if cb.state == "open" {
        if time.Since(cb.lastFailure) > cb.timeout {
            cb.state = "half-open"
        } else {
            return fmt.Errorf("circuit breaker is open")
        }
    }
    
    err := fn()
    if err != nil {
        cb.failureCount++
        cb.lastFailure = time.Now()
        
        if cb.failureCount >= cb.threshold {
            cb.state = "open"
        }
        return err
    }
    
    // Success
    cb.successCount++
    if cb.state == "half-open" {
        cb.state = "closed"
        cb.failureCount = 0
    }
    
    return nil
}
```

## Caching Strategies

### 1. In-Memory Caching

```go
type CacheEntry struct {
    Data      interface{}
    ExpiresAt time.Time
}

type InMemoryCache struct {
    cache map[string]CacheEntry
    mutex sync.RWMutex
}

func NewInMemoryCache() *InMemoryCache {
    cache := &InMemoryCache{
        cache: make(map[string]CacheEntry),
    }
    
    // Start cleanup goroutine
    go cache.cleanup()
    
    return cache
}

func (c *InMemoryCache) Get(key string) (interface{}, bool) {
    c.mutex.RLock()
    defer c.mutex.RUnlock()
    
    entry, exists := c.cache[key]
    if !exists {
        return nil, false
    }
    
    if time.Now().After(entry.ExpiresAt) {
        return nil, false
    }
    
    return entry.Data, true
}

func (c *InMemoryCache) Set(key string, data interface{}, ttl time.Duration) {
    c.mutex.Lock()
    defer c.mutex.Unlock()
    
    c.cache[key] = CacheEntry{
        Data:      data,
        ExpiresAt: time.Now().Add(ttl),
    }
}

func (c *InMemoryCache) cleanup() {
    ticker := time.NewTicker(1 * time.Minute)
    defer ticker.Stop()
    
    for range ticker.C {
        c.mutex.Lock()
        now := time.Now()
        for key, entry := range c.cache {
            if now.After(entry.ExpiresAt) {
                delete(c.cache, key)
            }
        }
        c.mutex.Unlock()
    }
}
```

### 2. Cached Client Wrapper

```go
type CachedClient struct {
    client *yfinance.Client
    cache  *InMemoryCache
}

func NewCachedClient(client *yfinance.Client) *CachedClient {
    return &CachedClient{
        client: client,
        cache:  NewInMemoryCache(),
    }
}

func (cc *CachedClient) FetchQuote(ctx context.Context, symbol string, runID string) (*yfinance.NormalizedQuote, error) {
    // Check cache first
    cacheKey := fmt.Sprintf("quote:%s", symbol)
    if cached, exists := cc.cache.Get(cacheKey); exists {
        return cached.(*yfinance.NormalizedQuote), nil
    }
    
    // Fetch from API
    quote, err := cc.client.FetchQuote(ctx, symbol, runID)
    if err != nil {
        return nil, err
    }
    
    // Cache for 5 minutes
    cc.cache.Set(cacheKey, quote, 5*time.Minute)
    
    return quote, nil
}

func (cc *CachedClient) FetchDailyBars(ctx context.Context, symbol string, start, end time.Time, adjusted bool, runID string) (*yfinance.NormalizedBarBatch, error) {
    // Check cache first
    cacheKey := fmt.Sprintf("bars:%s:%s:%s:%t", symbol, start.Format("2006-01-02"), end.Format("2006-01-02"), adjusted)
    if cached, exists := cc.cache.Get(cacheKey); exists {
        return cached.(*yfinance.NormalizedBarBatch), nil
    }
    
    // Fetch from API
    bars, err := cc.client.FetchDailyBars(ctx, symbol, start, end, adjusted, runID)
    if err != nil {
        return nil, err
    }
    
    // Cache for 1 hour
    cc.cache.Set(cacheKey, bars, 1*time.Hour)
    
    return bars, nil
}
```

## Batch Processing

### 1. Efficient Batch Processing

```go
type BatchProcessor struct {
    client      *yfinance.Client
    concurrency int
    rateLimit   time.Duration
}

func NewBatchProcessor(client *yfinance.Client, concurrency int, rateLimit time.Duration) *BatchProcessor {
    return &BatchProcessor{
        client:      client,
        concurrency: concurrency,
        rateLimit:   rateLimit,
    }
}

func (bp *BatchProcessor) ProcessSymbols(symbols []string) []BatchResult {
    results := make([]BatchResult, len(symbols))
    semaphore := make(chan struct{}, bp.concurrency)
    rateLimiter := time.NewTicker(bp.rateLimit)
    defer rateLimiter.Stop()
    
    var wg sync.WaitGroup
    
    for i, symbol := range symbols {
        wg.Add(1)
        go func(index int, sym string) {
            defer wg.Done()
            
            // Acquire semaphore
            semaphore <- struct{}{}
            defer func() { <-semaphore }()
            
            // Wait for rate limit
            <-rateLimiter.C
            
            // Process symbol
            start := time.Now()
            ctx := context.Background()
            runID := fmt.Sprintf("batch-%d", time.Now().Unix())
            
            quote, err := bp.client.FetchQuote(ctx, sym, runID)
            
            results[index] = BatchResult{
                Symbol:   sym,
                Quote:    quote,
                Error:    err,
                Duration: time.Since(start),
            }
        }(i, symbol)
    }
    
    wg.Wait()
    return results
}
```

### 2. Streaming Batch Processing

```go
func (bp *BatchProcessor) ProcessSymbolsStream(symbols []string, resultChan chan<- BatchResult) {
    semaphore := make(chan struct{}, bp.concurrency)
    rateLimiter := time.NewTicker(bp.rateLimit)
    defer rateLimiter.Stop()
    
    var wg sync.WaitGroup
    
    for _, symbol := range symbols {
        wg.Add(1)
        go func(sym string) {
            defer wg.Done()
            
            // Acquire semaphore
            semaphore <- struct{}{}
            defer func() { <-semaphore }()
            
            // Wait for rate limit
            <-rateLimiter.C
            
            // Process symbol
            start := time.Now()
            ctx := context.Background()
            runID := fmt.Sprintf("stream-%d", time.Now().Unix())
            
            quote, err := bp.client.FetchQuote(ctx, sym, runID)
            
            resultChan <- BatchResult{
                Symbol:   sym,
                Quote:    quote,
                Error:    err,
                Duration: time.Since(start),
            }
        }(symbol)
    }
    
    wg.Wait()
    close(resultChan)
}
```

## Memory Optimization

### 1. Memory-Efficient Data Processing

```go
func processBarsEfficiently(bars *yfinance.NormalizedBarBatch) {
    // Process bars in chunks to avoid memory spikes
    chunkSize := 1000
    
    for i := 0; i < len(bars.Bars); i += chunkSize {
        end := i + chunkSize
        if end > len(bars.Bars) {
            end = len(bars.Bars)
        }
        
        chunk := bars.Bars[i:end]
        processBarChunk(chunk)
        
        // Force garbage collection if needed
        if i%10000 == 0 {
            runtime.GC()
        }
    }
}

func processBarChunk(bars []yfinance.NormalizedBar) {
    for _, bar := range bars {
        // Process bar
        price := float64(bar.Close.Scaled) / float64(bar.Close.Scale)
        // ... do something with price
        _ = price
    }
}
```

### 2. Object Pooling

```go
var barPool = sync.Pool{
    New: func() interface{} {
        return &yfinance.NormalizedBar{}
    },
}

func getBar() *yfinance.NormalizedBar {
    return barPool.Get().(*yfinance.NormalizedBar)
}

func putBar(bar *yfinance.NormalizedBar) {
    // Reset bar
    *bar = yfinance.NormalizedBar{}
    barPool.Put(bar)
}
```

## Network Optimization

### 1. Connection Pooling

```go
func createOptimizedClient() *yfinance.Client {
    config := &httpx.Config{
        // Connection pooling
        MaxIdleConns:        200,
        MaxIdleConnsPerHost: 20,
        IdleConnTimeout:     90 * time.Second,
        
        // Keep-alive
        DisableKeepAlives: false,
        
        // Compression
        DisableCompression: false,
        
        // Timeouts
        Timeout:     30 * time.Second,
        DialTimeout: 10 * time.Second,
        
        // Rate limiting
        QPS:   3.0,
        Burst: 10,
    }
    
    return yfinance.NewClientWithConfig(config)
}
```

### 2. HTTP/2 Support

```go
func createHTTP2Client() *yfinance.Client {
    config := &httpx.Config{
        // Enable HTTP/2
        ForceHTTP2: true,
        
        // Connection pooling
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     60 * time.Second,
        
        // Timeouts
        Timeout:     20 * time.Second,
        DialTimeout: 5 * time.Second,
    }
    
    return yfinance.NewClientWithConfig(config)
}
```

## Production Best Practices

### 1. Health Checks

```go
type HealthChecker struct {
    client *yfinance.Client
}

func (hc *HealthChecker) CheckHealth() error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    runID := fmt.Sprintf("health-%d", time.Now().Unix())
    
    // Test with a known symbol
    _, err := hc.client.FetchQuote(ctx, "AAPL", runID)
    if err != nil {
        return fmt.Errorf("health check failed: %w", err)
    }
    
    return nil
}

func (hc *HealthChecker) StartHealthChecks(interval time.Duration) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()
    
    for range ticker.C {
        if err := hc.CheckHealth(); err != nil {
            log.Printf("Health check failed: %v", err)
            // Send alert
        }
    }
}
```

### 2. Graceful Shutdown

```go
type GracefulShutdown struct {
    client    *yfinance.Client
    ctx       context.Context
    cancel    context.CancelFunc
    wg        sync.WaitGroup
    shutdown  chan struct{}
}

func NewGracefulShutdown(client *yfinance.Client) *GracefulShutdown {
    ctx, cancel := context.WithCancel(context.Background())
    
    return &GracefulShutdown{
        client:   client,
        ctx:      ctx,
        cancel:   cancel,
        shutdown: make(chan struct{}),
    }
}

func (gs *GracefulShutdown) Start() {
    gs.wg.Add(1)
    go func() {
        defer gs.wg.Done()
        
        // Start processing
        for {
            select {
            case <-gs.ctx.Done():
                return
            case <-gs.shutdown:
                return
            default:
                // Do work
                time.Sleep(100 * time.Millisecond)
            }
        }
    }()
}

func (gs *GracefulShutdown) Stop() {
    close(gs.shutdown)
    gs.cancel()
    gs.wg.Wait()
}
```

### 3. Configuration Management

```go
type Config struct {
    YahooFinance YahooFinanceConfig `yaml:"yahoo_finance"`
    RateLimit    RateLimitConfig    `yaml:"rate_limit"`
    Retry        RetryConfig        `yaml:"retry"`
    Cache        CacheConfig        `yaml:"cache"`
}

type YahooFinanceConfig struct {
    Timeout      time.Duration `yaml:"timeout"`
    MaxAttempts  int           `yaml:"max_attempts"`
    QPS          float64       `yaml:"qps"`
    Burst        int           `yaml:"burst"`
}

type RateLimitConfig struct {
    Enabled bool          `yaml:"enabled"`
    Rate    float64       `yaml:"rate"`
    Burst   int           `yaml:"burst"`
    Window  time.Duration `yaml:"window"`
}

type RetryConfig struct {
    Enabled     bool          `yaml:"enabled"`
    MaxAttempts int           `yaml:"max_attempts"`
    BackoffBase time.Duration `yaml:"backoff_base"`
    BackoffMax  time.Duration `yaml:"backoff_max"`
}

type CacheConfig struct {
    Enabled bool          `yaml:"enabled"`
    TTL     time.Duration `yaml:"ttl"`
    Size    int           `yaml:"size"`
}

func LoadConfig(filename string) (*Config, error) {
    data, err := os.ReadFile(filename)
    if err != nil {
        return nil, err
    }
    
    var config Config
    if err := yaml.Unmarshal(data, &config); err != nil {
        return nil, err
    }
    
    return &config, nil
}
```

## Monitoring & Metrics

### 1. Performance Metrics

```go
type PerformanceMetrics struct {
    RequestCount     int64
    SuccessCount     int64
    FailureCount     int64
    TotalLatency     time.Duration
    AverageLatency   time.Duration
    P95Latency       time.Duration
    P99Latency       time.Duration
    mutex            sync.RWMutex
}

func (pm *PerformanceMetrics) RecordRequest(success bool, latency time.Duration) {
    pm.mutex.Lock()
    defer pm.mutex.Unlock()
    
    pm.RequestCount++
    pm.TotalLatency += latency
    
    if success {
        pm.SuccessCount++
    } else {
        pm.FailureCount++
    }
    
    // Update average latency
    pm.AverageLatency = pm.TotalLatency / time.Duration(pm.RequestCount)
}

func (pm *PerformanceMetrics) GetMetrics() map[string]interface{} {
    pm.mutex.RLock()
    defer pm.mutex.RUnlock()
    
    return map[string]interface{}{
        "request_count":   pm.RequestCount,
        "success_count":   pm.SuccessCount,
        "failure_count":   pm.FailureCount,
        "success_rate":    float64(pm.SuccessCount) / float64(pm.RequestCount),
        "average_latency": pm.AverageLatency.Milliseconds(),
        "total_latency":   pm.TotalLatency.Milliseconds(),
    }
}
```

### 2. Prometheus Metrics

```go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    requestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "yfinance_request_duration_seconds",
            Help: "Duration of yfinance requests",
        },
        []string{"method", "symbol", "status"},
    )
    
    requestCount = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "yfinance_requests_total",
            Help: "Total number of yfinance requests",
        },
        []string{"method", "symbol", "status"},
    )
    
    activeConnections = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "yfinance_active_connections",
            Help: "Number of active connections",
        },
    )
)

func recordMetrics(method, symbol, status string, duration time.Duration) {
    requestDuration.WithLabelValues(method, symbol, status).Observe(duration.Seconds())
    requestCount.WithLabelValues(method, symbol, status).Inc()
}
```

### 3. Logging

```go
import (
    "go.uber.org/zap"
)

type Logger struct {
    logger *zap.Logger
}

func NewLogger() *Logger {
    logger, _ := zap.NewProduction()
    return &Logger{logger: logger}
}

func (l *Logger) LogRequest(method, symbol string, duration time.Duration, err error) {
    fields := []zap.Field{
        zap.String("method", method),
        zap.String("symbol", symbol),
        zap.Duration("duration", duration),
    }
    
    if err != nil {
        fields = append(fields, zap.Error(err))
        l.logger.Error("Request failed", fields...)
    } else {
        l.logger.Info("Request completed", fields...)
    }
}
```

## Next Steps

- [API Reference](api-reference.md) - Complete API documentation
- [Data Structures](data-structures.md) - Detailed data structure guide
- [Complete Examples](examples.md) - Working code examples
- [Error Handling Guide](error-handling.md) - Comprehensive error handling
- [Method Comparison](method-comparison.md) - Method comparison and use cases
- [Data Quality Guide](data-quality.md) - Data quality and validation
