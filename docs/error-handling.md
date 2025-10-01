# Error Handling & Troubleshooting Guide

This guide provides comprehensive information about error handling, common issues, and troubleshooting strategies for yfinance-go.

## Table of Contents

1. [Error Types & Classification](#error-types--classification)
2. [Common Error Scenarios](#common-error-scenarios)
3. [Error Handling Strategies](#error-handling-strategies)
4. [Retry Logic & Backoff](#retry-logic--backoff)
5. [Rate Limiting & Throttling](#rate-limiting--throttling)
6. [Data Quality Issues](#data-quality-issues)
7. [Troubleshooting Checklist](#troubleshooting-checklist)
8. [Best Practices](#best-practices)

## Error Types & Classification

### Network Errors
**Symptoms**: Connection timeouts, network failures, DNS resolution issues

**Common Causes**:
- Network connectivity problems
- Yahoo Finance server issues
- Firewall or proxy blocking requests
- DNS resolution failures

**Error Messages**:
```
context deadline exceeded
connection refused
no such host
network is unreachable
```

**Solutions**:
- Check network connectivity
- Verify DNS resolution
- Check firewall/proxy settings
- Increase timeout values
- Use session rotation

### Rate Limiting Errors
**Symptoms**: HTTP 429 responses, slow responses, request throttling

**Common Causes**:
- Too many requests per second
- IP address being throttled
- Missing rate limiting controls

**Error Messages**:
```
429 Too Many Requests
rate limit exceeded
too many requests
```

**Solutions**:
- Implement rate limiting
- Use session rotation
- Add delays between requests
- Use `NewClientWithSessionRotation()`

### Authentication Errors
**Symptoms**: HTTP 401 responses, subscription required errors

**Common Causes**:
- Paid subscription required for some endpoints
- Invalid authentication credentials
- API key issues

**Error Messages**:
```
401 Unauthorized
authentication required
paid subscription required
```

**Solutions**:
- Use scraping methods as fallback
- Check subscription status
- Verify authentication credentials

### Parse Errors
**Symptoms**: JSON parsing failures, data structure mismatches

**Common Causes**:
- Yahoo Finance website structure changes
- Invalid response format
- Data corruption during transmission

**Error Messages**:
```
invalid character
unexpected end of JSON input
cannot unmarshal
parse error
```

**Solutions**:
- Update to latest version
- Check for website structure changes
- Implement graceful degradation
- Report issues to maintainers

### Data Not Found Errors
**Symptoms**: Empty results, "no data found" messages

**Common Causes**:
- Invalid symbol
- Data not available for symbol
- Date range issues
- Market closed

**Error Messages**:
```
no quotes found
no bars found
no news articles found
symbol not found
```

**Solutions**:
- Verify symbol validity
- Check date ranges
- Handle empty results gracefully
- Use alternative data sources

## Common Error Scenarios

### Scenario 1: "no news articles found" Error

**Problem**: News scraping returns empty results

**Cause**: Yahoo Finance may not have news for certain symbols

**Solution**: Handle gracefully, this is not a critical error

```go
news, err := client.ScrapeNews(ctx, "AAPL", runID)
if err != nil {
    log.Printf("Warning: News collection failed: %v", err)
    // Continue without news - this is not critical
    news = nil
}

if news == nil || len(news) == 0 {
    log.Printf("No news articles available for %s", symbol)
    // Continue processing other data
}
```

### Scenario 2: Rate Limiting Issues

**Problem**: Requests are being throttled or blocked

**Symptoms**: Slow responses, timeouts, 429 errors

**Solution**: Use session rotation and rate limiting

```go
// Use session rotation for better reliability
client := yfinance.NewClientWithSessionRotation()

// Implement rate limiting
func fetchWithRateLimit(client *yfinance.Client, symbol string) (*yfinance.NormalizedQuote, error) {
    // Add delay between requests
    time.Sleep(1 * time.Second)
    
    return client.FetchQuote(ctx, symbol, runID)
}
```

### Scenario 3: Empty Company Information

**Problem**: `FetchCompanyInfo()` returns limited data

**Cause**: `FetchCompanyInfo()` only returns basic security data

**Solution**: Use alternative data sources for detailed company profiles

```go
companyInfo, err := client.FetchCompanyInfo(ctx, "AAPL", runID)
if err != nil {
    log.Fatal(err)
}

// Note: This only returns basic info
fmt.Printf("Company: %s\n", companyInfo.LongName)
fmt.Printf("Exchange: %s\n", companyInfo.Exchange)

// For detailed company profiles, use alternative data sources
// or consider contributing to expose internal profile scraping functionality
```

### Scenario 4: Data Structure Access Issues

**Problem**: Field names don't match expectations

**Cause**: Field names use snake_case convention

**Solution**: Always check the actual JSON structure first

```go
// ✅ CORRECT: Use proper field names
if quote.RegularMarketPrice != nil {
    price := float64(quote.RegularMarketPrice.Scaled) / 
            float64(quote.RegularMarketPrice.Scale)
    fmt.Printf("Price: $%.2f\n", price)
}

// ❌ WRONG: Don't assume field names
// if quote.Price != nil { // This field doesn't exist
```

### Scenario 5: Paid Subscription Required

**Problem**: Some endpoints require paid subscription

**Cause**: Yahoo Finance paid features

**Solution**: Use scraping methods as fallback

```go
// Try API first
fundamentals, err := client.FetchFundamentalsQuarterly(ctx, "AAPL", runID)
if err != nil {
    // Check if it's a subscription error
    if strings.Contains(err.Error(), "paid subscription") {
        // Fall back to scraping
        fundamentals, err = client.ScrapeFinancials(ctx, "AAPL", runID)
        if err != nil {
            return fmt.Errorf("both API and scraping failed: %w", err)
        }
    } else {
        return err
    }
}
```

## Error Handling Strategies

### 1. Graceful Degradation

Handle errors gracefully and continue processing when possible:

```go
func fetchStockData(client *yfinance.Client, symbol string) *StockData {
    data := &StockData{Symbol: symbol}
    
    // Fetch quote (required)
    quote, err := client.FetchQuote(ctx, symbol, runID)
    if err != nil {
        data.Errors = append(data.Errors, fmt.Sprintf("Quote: %v", err))
        return data // Cannot continue without quote
    }
    data.Quote = quote
    
    // Fetch company info (optional)
    companyInfo, err := client.FetchCompanyInfo(ctx, symbol, runID)
    if err != nil {
        data.Warnings = append(data.Warnings, fmt.Sprintf("Company Info: %v", err))
        // Continue without company info
    } else {
        data.CompanyInfo = companyInfo
    }
    
    // Fetch financials (optional)
    financials, err := client.ScrapeFinancials(ctx, symbol, runID)
    if err != nil {
        data.Warnings = append(data.Warnings, fmt.Sprintf("Financials: %v", err))
        // Continue without financials
    } else {
        data.Financials = financials
    }
    
    return data
}
```

### 2. Error Classification

Classify errors to handle them appropriately:

```go
func classifyError(err error) string {
    if err == nil {
        return "none"
    }
    
    errStr := err.Error()
    switch {
    case strings.Contains(errStr, "429") || strings.Contains(errStr, "rate limit"):
        return "rate_limit"
    case strings.Contains(errStr, "401") || strings.Contains(errStr, "Unauthorized"):
        return "authentication"
    case strings.Contains(errStr, "timeout") || strings.Contains(errStr, "connection"):
        return "network"
    case strings.Contains(errStr, "parse") || strings.Contains(errStr, "schema"):
        return "parse"
    case strings.Contains(errStr, "no") && strings.Contains(errStr, "found"):
        return "not_found"
    default:
        return "unknown"
    }
}

func handleError(err error, symbol string) {
    errorType := classifyError(err)
    
    switch errorType {
    case "rate_limit":
        log.Printf("Rate limit hit for %s, implementing backoff", symbol)
        time.Sleep(5 * time.Second)
    case "authentication":
        log.Printf("Authentication error for %s, trying scraping fallback", symbol)
        // Try scraping methods
    case "network":
        log.Printf("Network error for %s, retrying with longer timeout", symbol)
        // Retry with longer timeout
    case "not_found":
        log.Printf("No data found for %s, this may be normal", symbol)
        // Handle gracefully
    default:
        log.Printf("Unknown error for %s: %v", symbol, err)
    }
}
```

### 3. Circuit Breaker Pattern

Implement circuit breaker to prevent cascading failures:

```go
type CircuitBreaker struct {
    failureCount int
    lastFailure  time.Time
    threshold    int
    timeout      time.Duration
    state        string // "closed", "open", "half-open"
}

func (cb *CircuitBreaker) Call(fn func() error) error {
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
    cb.failureCount = 0
    cb.state = "closed"
    return nil
}
```

## Retry Logic & Backoff

### Exponential Backoff with Jitter

```go
func fetchWithRetry(client *yfinance.Client, symbol string, maxRetries int) (*yfinance.NormalizedQuote, error) {
    ctx := context.Background()
    runID := fmt.Sprintf("retry-%d", time.Now().Unix())
    
    for i := 0; i < maxRetries; i++ {
        quote, err := client.FetchQuote(ctx, symbol, runID)
        if err == nil {
            return quote, nil
        }
        
        errorType := classifyError(err)
        log.Printf("Attempt %d failed for %s: %s (%s)", 
            i+1, symbol, err.Error(), errorType)
        
        if i < maxRetries-1 {
            // Exponential backoff with jitter
            baseDelay := time.Duration(i+1) * time.Second
            if errorType == "rate_limit" {
                baseDelay = baseDelay * 2 // Longer delay for rate limits
            }
            
            // Add jitter to prevent thundering herd
            jitter := time.Duration(rand.Intn(1000)) * time.Millisecond
            delay := baseDelay + jitter
            
            log.Printf("Retrying in %v...", delay)
            time.Sleep(delay)
        }
    }
    
    return nil, fmt.Errorf("failed after %d retries", maxRetries)
}
```

### Retry with Different Strategies

```go
func fetchWithMultipleStrategies(client *yfinance.Client, symbol string) (*yfinance.NormalizedQuote, error) {
    strategies := []func() (*yfinance.NormalizedQuote, error){
        func() (*yfinance.NormalizedQuote, error) {
            return client.FetchQuote(ctx, symbol, runID)
        },
        func() (*yfinance.NormalizedQuote, error) {
            // Try with session rotation
            rotatedClient := yfinance.NewClientWithSessionRotation()
            return rotatedClient.FetchQuote(ctx, symbol, runID)
        },
        func() (*yfinance.NormalizedQuote, error) {
            // Try with different timeout
            time.Sleep(2 * time.Second)
            return client.FetchQuote(ctx, symbol, runID)
        },
    }
    
    for i, strategy := range strategies {
        quote, err := strategy()
        if err == nil {
            return quote, nil
        }
        
        log.Printf("Strategy %d failed: %v", i+1, err)
        if i < len(strategies)-1 {
            time.Sleep(time.Duration(i+1) * time.Second)
        }
    }
    
    return nil, fmt.Errorf("all strategies failed")
}
```

## Rate Limiting & Throttling

### Client Configuration for Rate Limiting

```go
// Use session rotation for better rate limit handling
client := yfinance.NewClientWithSessionRotation()

// Custom configuration with rate limiting
config := &httpx.Config{
    QPS:         1.0,              // 1 request per second
    Burst:       3,                // Allow burst of 3 requests
    Timeout:     30 * time.Second, // 30 second timeout
    MaxAttempts: 3,                // 3 retry attempts
}

client := yfinance.NewClientWithConfig(config)
```

### Manual Rate Limiting

```go
type RateLimiter struct {
    tokens   chan struct{}
    interval time.Duration
}

func NewRateLimiter(rate int, interval time.Duration) *RateLimiter {
    rl := &RateLimiter{
        tokens:   make(chan struct{}, rate),
        interval: interval,
    }
    
    // Fill the token bucket
    for i := 0; i < rate; i++ {
        rl.tokens <- struct{}{}
    }
    
    // Start refilling tokens
    go func() {
        ticker := time.NewTicker(interval)
        for range ticker.C {
            select {
            case rl.tokens <- struct{}{}:
            default:
                // Bucket is full
            }
        }
    }()
    
    return rl
}

func (rl *RateLimiter) Wait() {
    <-rl.tokens
}

// Usage
rateLimiter := NewRateLimiter(10, time.Second) // 10 requests per second

for _, symbol := range symbols {
    rateLimiter.Wait() // Wait for rate limit
    
    quote, err := client.FetchQuote(ctx, symbol, runID)
    if err != nil {
        log.Printf("Error fetching %s: %v", symbol, err)
    }
}
```

## Data Quality Issues

### Handling Missing Data

```go
func validateDataQuality(data interface{}) error {
    switch v := data.(type) {
    case *yfinance.NormalizedBarBatch:
        if len(v.Bars) == 0 {
            return fmt.Errorf("no bars in batch")
        }
        
        // Check for data quality issues
        for _, bar := range v.Bars {
            if bar.Volume < 0 {
                return fmt.Errorf("negative volume detected")
            }
            if bar.Open.Scale < 0 || bar.High.Scale < 0 || 
               bar.Low.Scale < 0 || bar.Close.Scale < 0 {
                return fmt.Errorf("invalid scale values")
            }
        }
        
    case *yfinance.NormalizedQuote:
        if v.RegularMarketPrice == nil {
            return fmt.Errorf("missing market price")
        }
        
        if v.CurrencyCode == "" {
            return fmt.Errorf("missing currency code")
        }
        
    case *yfinance.NormalizedFundamentalsSnapshot:
        if len(v.Lines) == 0 {
            return fmt.Errorf("no financial lines")
        }
        
        // Check for zero values
        for _, line := range v.Lines {
            if line.Value.Scale == 0 && line.Value.Scaled == 0 {
                log.Printf("Warning: Zero value for %s", line.Key)
            }
        }
    }
    
    return nil
}
```

### Data Validation

```go
func validateStockData(data *StockData) error {
    if data.Quote == nil {
        return fmt.Errorf("quote data is required")
    }
    
    if data.HistoricalData == nil {
        log.Printf("Warning: No historical data available")
    }
    
    if data.Financials == nil {
        log.Printf("Warning: No financial data available")
    }
    
    // Validate quote data
    if err := validateDataQuality(data.Quote); err != nil {
        return fmt.Errorf("quote validation failed: %w", err)
    }
    
    return nil
}
```

## Troubleshooting Checklist

### 1. Network Issues
- [ ] Check internet connectivity
- [ ] Verify DNS resolution
- [ ] Check firewall/proxy settings
- [ ] Test with different network
- [ ] Increase timeout values

### 2. Rate Limiting Issues
- [ ] Implement rate limiting
- [ ] Use session rotation
- [ ] Add delays between requests
- [ ] Check request frequency
- [ ] Monitor for 429 errors

### 3. Data Issues
- [ ] Verify symbol validity
- [ ] Check date ranges
- [ ] Handle empty results
- [ ] Validate data quality
- [ ] Check for nil pointers

### 4. Authentication Issues
- [ ] Check subscription status
- [ ] Use scraping fallback
- [ ] Verify API credentials
- [ ] Check for 401 errors

### 5. Parse Issues
- [ ] Update to latest version
- [ ] Check for website changes
- [ ] Implement graceful degradation
- [ ] Report issues to maintainers

## Best Practices

### 1. Always Check for Nil Pointers
```go
if quote.RegularMarketPrice != nil {
    price := float64(quote.RegularMarketPrice.Scaled) / 
            float64(quote.RegularMarketPrice.Scale)
    fmt.Printf("Price: %.2f\n", price)
}
```

### 2. Implement Proper Error Handling
```go
quote, err := client.FetchQuote(ctx, "AAPL", runID)
if err != nil {
    log.Printf("Error fetching quote: %v", err)
    // Handle error appropriately
    return
}
```

### 3. Use Session Rotation for Production
```go
// For production use with high volume
client := yfinance.NewClientWithSessionRotation()

// For development/testing
client := yfinance.NewClient()
```

### 4. Implement Retry Logic
```go
func fetchWithRetry(client *yfinance.Client, symbol string, maxRetries int) (*yfinance.NormalizedQuote, error) {
    for i := 0; i < maxRetries; i++ {
        quote, err := client.FetchQuote(ctx, symbol, runID)
        if err == nil {
            return quote, nil
        }
        
        if i < maxRetries-1 {
            time.Sleep(time.Duration(i+1) * time.Second)
        }
    }
    return nil, fmt.Errorf("failed after %d retries", maxRetries)
}
```

### 5. Monitor and Log Errors
```go
func logError(err error, context string) {
    log.Printf("Error in %s: %v", context, err)
    
    // Send to monitoring system
    // metrics.IncrementErrorCounter(context, classifyError(err))
}
```

### 6. Handle Empty Results Gracefully
```go
if len(bars.Bars) == 0 {
    log.Printf("No data available for the specified date range")
    return
}
```

## Next Steps

- [API Reference](api-reference.md) - Complete API documentation
- [Data Structures](data-structures.md) - Detailed data structure guide
- [Complete Examples](examples.md) - Working code examples
- [Method Comparison](method-comparison.md) - Method comparison and use cases
