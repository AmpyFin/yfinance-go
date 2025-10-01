# Complete Working Examples

This document provides comprehensive, working examples that demonstrate how to use yfinance-go effectively, including data processing, formatting, error handling, and real-world use cases.

## Table of Contents

1. [Basic Data Fetching](#basic-data-fetching)
2. [Data Processing & Formatting](#data-processing--formatting)
3. [Error Handling & Retry Logic](#error-handling--retry-logic)
4. [Batch Processing](#batch-processing)
5. [HTML Report Generation](#html-report-generation)
6. [Data Pipeline Integration](#data-pipeline-integration)
7. [Performance Optimization](#performance-optimization)

## Basic Data Fetching

### Example 1: Fetch Quote with Proper Field Access

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/AmpyFin/yfinance-go"
)

func main() {
    client := yfinance.NewClient()
    ctx := context.Background()
    runID := fmt.Sprintf("example-%d", time.Now().Unix())
    
    // Fetch quote with proper field access
    quote, err := client.FetchQuote(ctx, "AAPL", runID)
    if err != nil {
        log.Fatal(err)
    }
    
    // Access price correctly using scaled decimal
    if quote.RegularMarketPrice != nil {
        price := float64(quote.RegularMarketPrice.Scaled) / 
                float64(1<<uint(quote.RegularMarketPrice.Scale))
        fmt.Printf("Price: $%.2f\n", price)
    }
    
    // Access other fields
    if quote.RegularMarketVolume != nil {
        fmt.Printf("Volume: %d\n", *quote.RegularMarketVolume)
    }
    
    fmt.Printf("Currency: %s\n", quote.CurrencyCode)
    fmt.Printf("Event Time: %s\n", quote.EventTime.Format("2006-01-02 15:04:05"))
}
```

### Example 2: Fetch Historical Data

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/AmpyFin/yfinance-go"
)

func main() {
    client := yfinance.NewClient()
    ctx := context.Background()
    runID := fmt.Sprintf("historical-%d", time.Now().Unix())
    
    // Define date range
    start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
    end := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)
    
    // Fetch daily bars
    bars, err := client.FetchDailyBars(ctx, "AAPL", start, end, true, runID)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Fetched %d bars for %s\n", len(bars.Bars), bars.Security.Symbol)
    
    // Process each bar
    for _, bar := range bars.Bars {
        // Convert scaled decimals to floats
        open := float64(bar.Open.Scaled) / float64(bar.Open.Scale)
        high := float64(bar.High.Scaled) / float64(bar.High.Scale)
        low := float64(bar.Low.Scaled) / float64(bar.Low.Scale)
        close := float64(bar.Close.Scaled) / float64(bar.Close.Scale)
        
        fmt.Printf("Date: %s, OHLC: %.2f/%.2f/%.2f/%.2f, Volume: %d\n", 
            bar.EventTime.Format("2006-01-02"),
            open, high, low, close, bar.Volume)
    }
}
```

### Example 3: Fetch Company Information

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/AmpyFin/yfinance-go"
)

func main() {
    client := yfinance.NewClient()
    ctx := context.Background()
    runID := fmt.Sprintf("company-%d", time.Now().Unix())
    
    // Fetch company info
    companyInfo, err := client.FetchCompanyInfo(ctx, "AAPL", runID)
    if err != nil {
        log.Fatal(err)
    }
    
    // Display company information
    fmt.Printf("Company: %s\n", companyInfo.LongName)
    fmt.Printf("Short Name: %s\n", companyInfo.ShortName)
    fmt.Printf("Exchange: %s\n", companyInfo.Exchange)
    fmt.Printf("Full Exchange: %s\n", companyInfo.FullExchangeName)
    fmt.Printf("Currency: %s\n", companyInfo.Currency)
    fmt.Printf("Instrument Type: %s\n", companyInfo.InstrumentType)
    fmt.Printf("Timezone: %s\n", companyInfo.Timezone)
    
    if companyInfo.FirstTradeDate != nil {
        fmt.Printf("First Trade Date: %s\n", 
            companyInfo.FirstTradeDate.Format("2006-01-02"))
    }
    
    // Note: This does NOT include address, executives, website, etc.
    fmt.Println("\nNote: For detailed company profiles, use alternative data sources.")
}
```

## Data Processing & Formatting

### Example 4: Scaled Decimal Helper Functions

```go
package main

import (
    "fmt"
    "math"
)

// Helper function for scaled decimal conversion
func formatScaledDecimal(scaled int64, scale int32) string {
    if scale == 0 {
        return fmt.Sprintf("%d", scaled)
    }
    divisor := math.Pow(10, float64(scale))
    value := float64(scaled) / divisor
    return fmt.Sprintf("%.2f", value)
}

// Convert scaled decimal to float64
func scaledDecimalToFloat(scaled int64, scale int32) float64 {
    if scale == 0 {
        return float64(scaled)
    }
    divisor := math.Pow(10, float64(scale))
    return float64(scaled) / divisor
}

// Format currency with proper precision
func formatCurrency(scaled int64, scale int32, currency string) string {
    value := scaledDecimalToFloat(scaled, scale)
    return fmt.Sprintf("%.2f %s", value, currency)
}

func main() {
    // Example usage
    price := struct {
        Scaled int64
        Scale  int32
    }{
        Scaled: 25503,
        Scale:  2,
    }
    
    fmt.Printf("Formatted: %s\n", formatScaledDecimal(price.Scaled, price.Scale))
    fmt.Printf("Float: %.4f\n", scaledDecimalToFloat(price.Scaled, price.Scale))
    fmt.Printf("Currency: %s\n", formatCurrency(price.Scaled, price.Scale, "USD"))
}
```

### Example 5: Financial Data Processing

```go
package main

import (
    "context"
    "fmt"
    "log"
    "sort"
    "time"

    "github.com/AmpyFin/yfinance-go"
)

type FinancialMetric struct {
    Key   string
    Value float64
    Unit  string
}

func main() {
    client := yfinance.NewClient()
    ctx := context.Background()
    runID := fmt.Sprintf("financials-%d", time.Now().Unix())
    
    // Scrape financial data
    financials, err := client.ScrapeFinancials(ctx, "AAPL", runID)
    if err != nil {
        log.Fatal(err)
    }
    
    // Process and sort financial metrics
    var metrics []FinancialMetric
    
    for _, line := range financials.Lines {
        value := float64(line.Value.Scaled) / float64(line.Value.Scale)
        metrics = append(metrics, FinancialMetric{
            Key:   line.Key,
            Value: value,
            Unit:  line.CurrencyCode,
        })
    }
    
    // Sort by value (descending)
    sort.Slice(metrics, func(i, j int) bool {
        return metrics[i].Value > metrics[j].Value
    })
    
    // Display top 10 metrics
    fmt.Printf("Top 10 Financial Metrics for %s:\n", financials.Meta.Source)
    fmt.Println("=====================================")
    
    for i, metric := range metrics {
        if i >= 10 {
            break
        }
        fmt.Printf("%-20s: %15.2f %s\n", metric.Key, metric.Value, metric.Unit)
    }
}
```

### Example 6: News Data Processing

```go
package main

import (
    "context"
    "fmt"
    "log"
    "sort"
    "strings"
    "time"

    "github.com/AmpyFin/yfinance-go"
)

type NewsSummary struct {
    Title       string
    Summary     string
    PublishedAt time.Time
    Source      string
    URL         string
}

func main() {
    client := yfinance.NewClient()
    ctx := context.Background()
    runID := fmt.Sprintf("news-%d", time.Now().Unix())
    
    // Scrape news
    news, err := client.ScrapeNews(ctx, "AAPL", runID)
    if err != nil {
        log.Fatal(err)
    }
    
    if len(news) == 0 {
        fmt.Println("No news articles found")
        return
    }
    
    // Process news articles
    var summaries []NewsSummary
    
    for _, article := range news {
        summaries = append(summaries, NewsSummary{
            Title:       article.Title,
            Summary:     article.Summary,
            PublishedAt: article.PublishedAt,
            Source:      article.Source,
            URL:         article.Url,
        })
    }
    
    // Sort by publication date (newest first)
    sort.Slice(summaries, func(i, j int) bool {
        return summaries[i].PublishedAt.After(summaries[j].PublishedAt)
    })
    
    // Display news summary
    fmt.Printf("Latest News for AAPL (%d articles):\n", len(summaries))
    fmt.Println("=====================================")
    
    for i, summary := range summaries {
        if i >= 5 { // Show top 5
            break
        }
        
        // Truncate summary if too long
        summaryText := summary.Summary
        if len(summaryText) > 100 {
            summaryText = summaryText[:100] + "..."
        }
        
        fmt.Printf("\n%d. %s\n", i+1, summary.Title)
        fmt.Printf("   Source: %s\n", summary.Source)
        fmt.Printf("   Published: %s\n", summary.PublishedAt.Format("2006-01-02 15:04"))
        fmt.Printf("   Summary: %s\n", summaryText)
        fmt.Printf("   URL: %s\n", summary.URL)
    }
}
```

## Error Handling & Retry Logic

### Example 7: Robust Error Handling

```go
package main

import (
    "context"
    "fmt"
    "log"
    "strings"
    "time"

    "github.com/AmpyFin/yfinance-go"
)

// Error classification
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
    default:
        return "unknown"
    }
}

// Retry with exponential backoff
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
            delay := time.Duration(i+1) * time.Second
            if errorType == "rate_limit" {
                delay = delay * 2 // Longer delay for rate limits
            }
            log.Printf("Retrying in %v...", delay)
            time.Sleep(delay)
        }
    }
    
    return nil, fmt.Errorf("failed after %d retries", maxRetries)
}

func main() {
    client := yfinance.NewClientWithSessionRotation() // Use session rotation for better reliability
    symbol := "AAPL"
    
    quote, err := fetchWithRetry(client, symbol, 3)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Successfully fetched quote for %s\n", symbol)
    if quote.RegularMarketPrice != nil {
        price := float64(quote.RegularMarketPrice.Scaled) / 
                float64(quote.RegularMarketPrice.Scale)
        fmt.Printf("Price: $%.2f %s\n", price, quote.CurrencyCode)
    }
}
```

### Example 8: Graceful Degradation

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/AmpyFin/yfinance-go"
)

type StockData struct {
    Symbol      string
    Quote       *yfinance.NormalizedQuote
    CompanyInfo *yfinance.NormalizedCompanyInfo
    Financials  *yfinance.FundamentalsSnapshot
    News        []*yfinance.NewsItem
    Errors      []string
}

func fetchStockData(client *yfinance.Client, symbol string) *StockData {
    ctx := context.Background()
    runID := fmt.Sprintf("stock-data-%d", time.Now().Unix())
    
    data := &StockData{
        Symbol: symbol,
        Errors: make([]string, 0),
    }
    
    // Fetch quote (required)
    quote, err := client.FetchQuote(ctx, symbol, runID)
    if err != nil {
        data.Errors = append(data.Errors, fmt.Sprintf("Quote: %v", err))
    } else {
        data.Quote = quote
    }
    
    // Fetch company info (optional)
    companyInfo, err := client.FetchCompanyInfo(ctx, symbol, runID)
    if err != nil {
        data.Errors = append(data.Errors, fmt.Sprintf("Company Info: %v", err))
    } else {
        data.CompanyInfo = companyInfo
    }
    
    // Fetch financials (optional)
    financials, err := client.ScrapeFinancials(ctx, symbol, runID)
    if err != nil {
        data.Errors = append(data.Errors, fmt.Sprintf("Financials: %v", err))
    } else {
        data.Financials = financials
    }
    
    // Fetch news (optional)
    news, err := client.ScrapeNews(ctx, symbol, runID)
    if err != nil {
        data.Errors = append(data.Errors, fmt.Sprintf("News: %v", err))
    } else {
        data.News = news
    }
    
    return data
}

func main() {
    client := yfinance.NewClientWithSessionRotation()
    symbol := "AAPL"
    
    data := fetchStockData(client, symbol)
    
    // Display results
    fmt.Printf("Stock Data for %s:\n", symbol)
    fmt.Println("==================")
    
    if data.Quote != nil {
        if data.Quote.RegularMarketPrice != nil {
            price := float64(data.Quote.RegularMarketPrice.Scaled) / 
                    float64(data.Quote.RegularMarketPrice.Scale)
            fmt.Printf("Price: $%.2f %s\n", price, data.Quote.CurrencyCode)
        }
    }
    
    if data.CompanyInfo != nil {
        fmt.Printf("Company: %s\n", data.CompanyInfo.LongName)
        fmt.Printf("Exchange: %s\n", data.CompanyInfo.Exchange)
    }
    
    if data.Financials != nil {
        fmt.Printf("Financial Metrics: %d available\n", len(data.Financials.Lines))
    }
    
    if data.News != nil {
        fmt.Printf("News Articles: %d available\n", len(data.News))
    }
    
    // Display errors
    if len(data.Errors) > 0 {
        fmt.Println("\nErrors encountered:")
        for _, err := range data.Errors {
            fmt.Printf("  - %s\n", err)
        }
    }
}
```

## Batch Processing

### Example 9: Concurrent Batch Processing

```go
package main

import (
    "context"
    "fmt"
    "log"
    "sync"
    "time"

    "github.com/AmpyFin/yfinance-go"
)

type BatchResult struct {
    Symbol    string
    Quote     *yfinance.NormalizedQuote
    Error     error
    Duration  time.Duration
}

func processSymbol(client *yfinance.Client, symbol string) BatchResult {
    start := time.Now()
    ctx := context.Background()
    runID := fmt.Sprintf("batch-%d", time.Now().Unix())
    
    quote, err := client.FetchQuote(ctx, symbol, runID)
    
    return BatchResult{
        Symbol:   symbol,
        Quote:    quote,
        Error:    err,
        Duration: time.Since(start),
    }
}

func main() {
    client := yfinance.NewClientWithSessionRotation()
    
    // Define symbols to process
    symbols := []string{"AAPL", "MSFT", "GOOGL", "AMZN", "TSLA", "META", "NVDA", "NFLX"}
    
    // Process symbols concurrently
    results := make(chan BatchResult, len(symbols))
    var wg sync.WaitGroup
    
    // Limit concurrency to avoid rate limiting
    semaphore := make(chan struct{}, 3) // Max 3 concurrent requests
    
    for _, symbol := range symbols {
        wg.Add(1)
        go func(sym string) {
            defer wg.Done()
            
            // Acquire semaphore
            semaphore <- struct{}{}
            defer func() { <-semaphore }()
            
            result := processSymbol(client, sym)
            results <- result
        }(symbol)
    }
    
    // Wait for all goroutines to complete
    go func() {
        wg.Wait()
        close(results)
    }()
    
    // Collect results
    var successful, failed int
    var totalDuration time.Duration
    
    fmt.Printf("Processing %d symbols...\n", len(symbols))
    fmt.Println("================================")
    
    for result := range results {
        totalDuration += result.Duration
        
        if result.Error != nil {
            failed++
            fmt.Printf("✗ %s: %v (%.2fs)\n", result.Symbol, result.Error, result.Duration.Seconds())
        } else {
            successful++
            if result.Quote.RegularMarketPrice != nil {
                price := float64(result.Quote.RegularMarketPrice.Scaled) / 
                        float64(result.Quote.RegularMarketPrice.Scale)
                fmt.Printf("✓ %s: $%.2f %s (%.2fs)\n", 
                    result.Symbol, price, result.Quote.CurrencyCode, result.Duration.Seconds())
            }
        }
    }
    
    // Summary
    fmt.Println("\nBatch Processing Summary:")
    fmt.Printf("Successful: %d\n", successful)
    fmt.Printf("Failed: %d\n", failed)
    fmt.Printf("Success Rate: %.1f%%\n", float64(successful)/float64(len(symbols))*100)
    fmt.Printf("Average Duration: %.2fs\n", totalDuration.Seconds()/float64(len(symbols)))
}
```

## HTML Report Generation

### Example 10: Generate HTML Report

```go
package main

import (
    "context"
    "fmt"
    "html/template"
    "log"
    "os"
    "time"

    "github.com/AmpyFin/yfinance-go"
)

type ReportData struct {
    Symbol      string
    Quote       *yfinance.NormalizedQuote
    CompanyInfo *yfinance.NormalizedCompanyInfo
    Financials  *yfinance.FundamentalsSnapshot
    News        []*yfinance.NewsItem
    GeneratedAt time.Time
}

const htmlTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>Stock Report - {{.Symbol}}</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .header { background-color: #f0f0f0; padding: 20px; border-radius: 5px; }
        .section { margin: 20px 0; }
        .metric { display: inline-block; margin: 10px; padding: 10px; background-color: #e8f4f8; border-radius: 3px; }
        .news-item { border-bottom: 1px solid #eee; padding: 10px 0; }
        .error { color: red; }
        .success { color: green; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Stock Report: {{.Symbol}}</h1>
        <p>Generated: {{.GeneratedAt.Format "2006-01-02 15:04:05"}}</p>
    </div>

    {{if .Quote}}
    <div class="section">
        <h2>Current Quote</h2>
        {{if .Quote.RegularMarketPrice}}
        <div class="metric">
            <strong>Price:</strong> ${{printf "%.2f" (div .Quote.RegularMarketPrice.Scaled (pow 10 .Quote.RegularMarketPrice.Scale))}} {{.Quote.CurrencyCode}}
        </div>
        {{end}}
        {{if .Quote.RegularMarketVolume}}
        <div class="metric">
            <strong>Volume:</strong> {{.Quote.RegularMarketVolume}}
        </div>
        {{end}}
        <div class="metric">
            <strong>Event Time:</strong> {{.Quote.EventTime.Format "2006-01-02 15:04:05"}}
        </div>
    </div>
    {{end}}

    {{if .CompanyInfo}}
    <div class="section">
        <h2>Company Information</h2>
        <div class="metric">
            <strong>Name:</strong> {{.CompanyInfo.LongName}}
        </div>
        <div class="metric">
            <strong>Exchange:</strong> {{.CompanyInfo.Exchange}}
        </div>
        <div class="metric">
            <strong>Currency:</strong> {{.CompanyInfo.Currency}}
        </div>
        <div class="metric">
            <strong>Instrument Type:</strong> {{.CompanyInfo.InstrumentType}}
        </div>
    </div>
    {{end}}

    {{if .Financials}}
    <div class="section">
        <h2>Financial Metrics</h2>
        {{range .Financials.Lines}}
        <div class="metric">
            <strong>{{.Key}}:</strong> {{printf "%.2f" (div .Value.Scaled (pow 10 .Value.Scale))}} {{.CurrencyCode}}
        </div>
        {{end}}
    </div>
    {{end}}

    {{if .News}}
    <div class="section">
        <h2>Recent News ({{len .News}} articles)</h2>
        {{range .News}}
        <div class="news-item">
            <h3>{{.Title}}</h3>
            <p><strong>Source:</strong> {{.Source}} | <strong>Published:</strong> {{.PublishedAt.Format "2006-01-02 15:04"}}</p>
            <p>{{.Summary}}</p>
            <p><a href="{{.Url}}" target="_blank">Read more</a></p>
        </div>
        {{end}}
    </div>
    {{end}}
</body>
</html>
`

func generateReport(client *yfinance.Client, symbol string) error {
    ctx := context.Background()
    runID := fmt.Sprintf("report-%d", time.Now().Unix())
    
    // Collect data
    data := &ReportData{
        Symbol:      symbol,
        GeneratedAt: time.Now(),
    }
    
    // Fetch quote
    if quote, err := client.FetchQuote(ctx, symbol, runID); err == nil {
        data.Quote = quote
    }
    
    // Fetch company info
    if companyInfo, err := client.FetchCompanyInfo(ctx, symbol, runID); err == nil {
        data.CompanyInfo = companyInfo
    }
    
    // Fetch financials
    if financials, err := client.ScrapeFinancials(ctx, symbol, runID); err == nil {
        data.Financials = financials
    }
    
    // Fetch news
    if news, err := client.ScrapeNews(ctx, symbol, runID); err == nil {
        data.News = news
    }
    
    // Generate HTML
    tmpl, err := template.New("report").Funcs(template.FuncMap{
        "div": func(a, b float64) float64 { return a / b },
        "pow": func(a, b float64) float64 { 
            result := 1.0
            for i := 0; i < int(b); i++ {
                result *= a
            }
            return result
        },
    }).Parse(htmlTemplate)
    if err != nil {
        return err
    }
    
    filename := fmt.Sprintf("stock_report_%s_%s.html", symbol, time.Now().Format("20060102_150405"))
    file, err := os.Create(filename)
    if err != nil {
        return err
    }
    defer file.Close()
    
    return tmpl.Execute(file, data)
}

func main() {
    client := yfinance.NewClientWithSessionRotation()
    symbol := "AAPL"
    
    fmt.Printf("Generating report for %s...\n", symbol)
    
    if err := generateReport(client, symbol); err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Report generated successfully!\n")
}
```

## Data Pipeline Integration

### Example 11: Data Pipeline with Validation

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/AmpyFin/yfinance-go"
)

type DataPipeline struct {
    client *yfinance.Client
    runID  string
}

type PipelineResult struct {
    Symbol      string
    Quote       *yfinance.NormalizedQuote
    CompanyInfo *yfinance.NormalizedCompanyInfo
    Financials  *yfinance.FundamentalsSnapshot
    News        []*yfinance.NewsItem
    Errors      []string
    Warnings    []string
}

func NewDataPipeline() *DataPipeline {
    return &DataPipeline{
        client: yfinance.NewClientWithSessionRotation(),
        runID:  fmt.Sprintf("pipeline-%d", time.Now().Unix()),
    }
}

func (p *DataPipeline) ProcessSymbol(symbol string) *PipelineResult {
    ctx := context.Background()
    result := &PipelineResult{
        Symbol:   symbol,
        Errors:   make([]string, 0),
        Warnings: make([]string, 0),
    }
    
    // Step 1: Fetch quote (required)
    quote, err := p.client.FetchQuote(ctx, symbol, p.runID)
    if err != nil {
        result.Errors = append(result.Errors, fmt.Sprintf("Quote fetch failed: %v", err))
        return result
    }
    result.Quote = quote
    
    // Validate quote data
    if quote.RegularMarketPrice == nil {
        result.Warnings = append(result.Warnings, "No market price available")
    }
    
    // Step 2: Fetch company info (optional)
    companyInfo, err := p.client.FetchCompanyInfo(ctx, symbol, p.runID)
    if err != nil {
        result.Warnings = append(result.Warnings, fmt.Sprintf("Company info unavailable: %v", err))
    } else {
        result.CompanyInfo = companyInfo
    }
    
    // Step 3: Fetch financials (optional)
    financials, err := p.client.ScrapeFinancials(ctx, symbol, p.runID)
    if err != nil {
        result.Warnings = append(result.Warnings, fmt.Sprintf("Financials unavailable: %v", err))
    } else {
        result.Financials = financials
        
        // Validate financial data
        if len(financials.Lines) == 0 {
            result.Warnings = append(result.Warnings, "No financial metrics available")
        }
    }
    
    // Step 4: Fetch news (optional)
    news, err := p.client.ScrapeNews(ctx, symbol, p.runID)
    if err != nil {
        result.Warnings = append(result.Warnings, fmt.Sprintf("News unavailable: %v", err))
    } else {
        result.News = news
        
        // Validate news data
        if len(news) == 0 {
            result.Warnings = append(result.Warnings, "No news articles available")
        }
    }
    
    return result
}

func (p *DataPipeline) ProcessSymbols(symbols []string) []*PipelineResult {
    results := make([]*PipelineResult, 0, len(symbols))
    
    for i, symbol := range symbols {
        fmt.Printf("[%d/%d] Processing %s...", i+1, len(symbols), symbol)
        
        result := p.ProcessSymbol(symbol)
        results = append(results, result)
        
        if len(result.Errors) > 0 {
            fmt.Printf(" ✗ (%d errors)\n", len(result.Errors))
        } else {
            fmt.Printf(" ✓ (%d warnings)\n", len(result.Warnings))
        }
        
        // Rate limiting
        if i < len(symbols)-1 {
            time.Sleep(1 * time.Second)
        }
    }
    
    return results
}

func main() {
    pipeline := NewDataPipeline()
    symbols := []string{"AAPL", "MSFT", "GOOGL"}
    
    fmt.Printf("Starting data pipeline for %d symbols...\n", len(symbols))
    fmt.Println("==========================================")
    
    results := pipeline.ProcessSymbols(symbols)
    
    // Summary
    fmt.Println("\nPipeline Summary:")
    fmt.Println("=================")
    
    successful := 0
    totalWarnings := 0
    
    for _, result := range results {
        if len(result.Errors) == 0 {
            successful++
        }
        totalWarnings += len(result.Warnings)
        
        fmt.Printf("%s: ", result.Symbol)
        if len(result.Errors) > 0 {
            fmt.Printf("FAILED (%d errors)\n", len(result.Errors))
        } else {
            fmt.Printf("SUCCESS (%d warnings)\n", len(result.Warnings))
        }
    }
    
    fmt.Printf("\nOverall: %d/%d successful (%.1f%%)\n", 
        successful, len(symbols), float64(successful)/float64(len(symbols))*100)
    fmt.Printf("Total warnings: %d\n", totalWarnings)
}
```

## Performance Optimization

### Example 12: Optimized Batch Processing

```go
package main

import (
    "context"
    "fmt"
    "log"
    "sync"
    "time"

    "github.com/AmpyFin/yfinance-go"
)

type OptimizedBatchProcessor struct {
    client      *yfinance.Client
    concurrency int
    rateLimit   time.Duration
}

func NewOptimizedBatchProcessor(concurrency int, rateLimit time.Duration) *OptimizedBatchProcessor {
    return &OptimizedBatchProcessor{
        client:      yfinance.NewClientWithSessionRotation(),
        concurrency: concurrency,
        rateLimit:   rateLimit,
    }
}

func (p *OptimizedBatchProcessor) ProcessSymbols(symbols []string) []BatchResult {
    results := make([]BatchResult, len(symbols))
    semaphore := make(chan struct{}, p.concurrency)
    var wg sync.WaitGroup
    
    for i, symbol := range symbols {
        wg.Add(1)
        go func(index int, sym string) {
            defer wg.Done()
            
            // Acquire semaphore
            semaphore <- struct{}{}
            defer func() { <-semaphore }()
            
            // Process symbol
            start := time.Now()
            ctx := context.Background()
            runID := fmt.Sprintf("optimized-%d", time.Now().Unix())
            
            quote, err := p.client.FetchQuote(ctx, sym, runID)
            
            results[index] = BatchResult{
                Symbol:   sym,
                Quote:    quote,
                Error:    err,
                Duration: time.Since(start),
            }
            
            // Rate limiting
            time.Sleep(p.rateLimit)
        }(i, symbol)
    }
    
    wg.Wait()
    return results
}

func main() {
    // Create optimized processor
    processor := NewOptimizedBatchProcessor(3, 500*time.Millisecond)
    
    // Large symbol list
    symbols := []string{
        "AAPL", "MSFT", "GOOGL", "AMZN", "TSLA", "META", "NVDA", "NFLX",
        "ORCL", "CRM", "ADBE", "INTC", "CSCO", "IBM", "AMD", "QCOM",
    }
    
    fmt.Printf("Processing %d symbols with optimized batch processor...\n", len(symbols))
    fmt.Printf("Concurrency: %d, Rate limit: %v\n", processor.concurrency, processor.rateLimit)
    fmt.Println("=====================================================")
    
    start := time.Now()
    results := processor.ProcessSymbols(symbols)
    totalDuration := time.Since(start)
    
    // Analyze results
    successful := 0
    totalRequestTime := time.Duration(0)
    
    for _, result := range results {
        if result.Error == nil {
            successful++
        }
        totalRequestTime += result.Duration
    }
    
    fmt.Printf("\nOptimized Batch Processing Results:\n")
    fmt.Printf("Total time: %.2fs\n", totalDuration.Seconds())
    fmt.Printf("Successful: %d/%d (%.1f%%)\n", 
        successful, len(symbols), float64(successful)/float64(len(symbols))*100)
    fmt.Printf("Average request time: %.2fs\n", totalRequestTime.Seconds()/float64(len(symbols)))
    fmt.Printf("Throughput: %.2f requests/second\n", float64(len(symbols))/totalDuration.Seconds())
}
```

## Next Steps

- [API Reference](api-reference.md) - Complete API documentation
- [Data Structures](data-structures.md) - Detailed data structure guide
- [Error Handling Guide](error-handling.md) - Comprehensive error handling
- [Performance Guide](performance.md) - Performance optimization tips
