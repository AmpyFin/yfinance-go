# Migration Guide: Python yfinance to Go yfinance-go

This guide helps you migrate from Python's `yfinance` library to Go's `yfinance-go` library, including feature comparisons, code examples, and best practices.

## Table of Contents

1. [Feature Comparison](#feature-comparison)
2. [Key Differences](#key-differences)
3. [Migration Examples](#migration-examples)
4. [Data Structure Mapping](#data-structure-mapping)
5. [Error Handling Differences](#error-handling-differences)
6. [Performance Considerations](#performance-considerations)
7. [Best Practices](#best-practices)

## Feature Comparison

| Feature | Python yfinance | Go yfinance-go | Notes |
|---------|-----------------|----------------|-------|
| **Installation** | `pip install yfinance` | `go get github.com/AmpyFin/yfinance-go` | Go requires Go 1.23+ |
| **Basic Usage** | `yf.Ticker("AAPL")` | `yfinance.NewClient()` | Different API design |
| **Historical Data** | `ticker.history()` | `FetchDailyBars()` | Similar functionality |
| **Real-time Quotes** | `ticker.info` | `FetchQuote()` | Go has structured types |
| **Company Info** | `ticker.info` | `FetchCompanyInfo()` | Go has limited data |
| **Financials** | `ticker.financials` | `ScrapeFinancials()` | Go uses scraping |
| **News** | `ticker.news` | `ScrapeNews()` | Similar functionality |
| **Analyst Data** | `ticker.analysts` | `ScrapeAnalysis()` | Similar functionality |
| **Options Data** | `ticker.options` | ❌ Not supported | Not available in Go version |
| **Institutional Holders** | `ticker.institutional_holders` | ❌ Not supported | Not available in Go version |
| **Insider Trading** | `ticker.insider_transactions` | ❌ Not supported | Not available in Go version |
| **Data Types** | Pandas DataFrames | Structured Go types | Go uses strongly typed structs |
| **Error Handling** | Exceptions | Explicit error returns | Go uses explicit error handling |
| **Concurrency** | Limited | Built-in goroutines | Go has better concurrency support |
| **Performance** | Slower for large datasets | Faster for concurrent requests | Go is generally faster |

## Key Differences

### 1. API Design Philosophy

**Python yfinance**:
```python
import yfinance as yf

# Create ticker object
ticker = yf.Ticker("AAPL")

# Access data through properties
info = ticker.info
history = ticker.history(period="1mo")
news = ticker.news
```

**Go yfinance-go**:
```go
import "github.com/AmpyFin/yfinance-go"

// Create client
client := yfinance.NewClient()

// Call methods explicitly
quote, err := client.FetchQuote(ctx, "AAPL", runID)
bars, err := client.FetchDailyBars(ctx, "AAPL", start, end, true, runID)
news, err := client.ScrapeNews(ctx, "AAPL", runID)
```

### 2. Data Structure Differences

**Python yfinance** returns pandas DataFrames and dictionaries:
```python
# Returns pandas DataFrame
history = ticker.history(period="1mo")
print(history.head())

# Returns dictionary
info = ticker.info
print(info['longName'])
print(info['currentPrice'])
```

**Go yfinance-go** returns structured types:
```go
// Returns structured type
bars, err := client.FetchDailyBars(ctx, "AAPL", start, end, true, runID)
for _, bar := range bars.Bars {
    price := float64(bar.Close.Scaled) / float64(bar.Close.Scale)
    fmt.Printf("Date: %s, Close: %.2f\n", bar.EventTime.Format("2006-01-02"), price)
}

// Returns structured type
quote, err := client.FetchQuote(ctx, "AAPL", runID)
if quote.RegularMarketPrice != nil {
    price := float64(quote.RegularMarketPrice.Scaled) / float64(quote.RegularMarketPrice.Scale)
    fmt.Printf("Price: $%.2f\n", price)
}
```

### 3. Error Handling

**Python yfinance** uses exceptions:
```python
try:
    ticker = yf.Ticker("AAPL")
    info = ticker.info
    print(info['longName'])
except Exception as e:
    print(f"Error: {e}")
```

**Go yfinance-go** uses explicit error returns:
```go
client := yfinance.NewClient()
quote, err := client.FetchQuote(ctx, "AAPL", runID)
if err != nil {
    log.Fatal(err)
}
// Use quote data
```

## Migration Examples

### Example 1: Basic Data Fetching

**Python yfinance**:
```python
import yfinance as yf
import pandas as pd

# Create ticker
ticker = yf.Ticker("AAPL")

# Get company info
info = ticker.info
print(f"Company: {info['longName']}")
print(f"Price: ${info['currentPrice']}")

# Get historical data
history = ticker.history(period="1mo")
print(f"Latest close: ${history['Close'].iloc[-1]:.2f}")

# Get news
news = ticker.news
print(f"Latest news: {news[0]['title']}")
```

**Go yfinance-go**:
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
    
    // Get company info
    companyInfo, err := client.FetchCompanyInfo(ctx, "AAPL", runID)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Company: %s\n", companyInfo.LongName)
    
    // Get current quote
    quote, err := client.FetchQuote(ctx, "AAPL", runID)
    if err != nil {
        log.Fatal(err)
    }
    if quote.RegularMarketPrice != nil {
        price := float64(quote.RegularMarketPrice.Scaled) / float64(quote.RegularMarketPrice.Scale)
        fmt.Printf("Price: $%.2f\n", price)
    }
    
    // Get historical data
    end := time.Now()
    start := end.AddDate(0, 0, -30) // 1 month ago
    bars, err := client.FetchDailyBars(ctx, "AAPL", start, end, true, runID)
    if err != nil {
        log.Fatal(err)
    }
    if len(bars.Bars) > 0 {
        latestBar := bars.Bars[len(bars.Bars)-1]
        closePrice := float64(latestBar.Close.Scaled) / float64(latestBar.Close.Scale)
        fmt.Printf("Latest close: $%.2f\n", closePrice)
    }
    
    // Get news
    news, err := client.ScrapeNews(ctx, "AAPL", runID)
    if err != nil {
        log.Printf("Warning: News failed: %v", err)
    } else if len(news) > 0 {
        fmt.Printf("Latest news: %s\n", news[0].Title)
    }
}
```

### Example 2: Multiple Symbols

**Python yfinance**:
```python
import yfinance as yf

symbols = ["AAPL", "MSFT", "GOOGL"]
tickers = yf.Tickers(" ".join(symbols))

for symbol in symbols:
    ticker = tickers.tickers[symbol]
    info = ticker.info
    print(f"{symbol}: ${info['currentPrice']}")
```

**Go yfinance-go**:
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

func main() {
    client := yfinance.NewClientWithSessionRotation()
    ctx := context.Background()
    symbols := []string{"AAPL", "MSFT", "GOOGL"}
    
    var wg sync.WaitGroup
    results := make(chan string, len(symbols))
    
    for _, symbol := range symbols {
        wg.Add(1)
        go func(sym string) {
            defer wg.Done()
            
            runID := fmt.Sprintf("multi-%d", time.Now().Unix())
            quote, err := client.FetchQuote(ctx, sym, runID)
            if err != nil {
                results <- fmt.Sprintf("%s: Error - %v", sym, err)
                return
            }
            
            if quote.RegularMarketPrice != nil {
                price := float64(quote.RegularMarketPrice.Scaled) / float64(quote.RegularMarketPrice.Scale)
                results <- fmt.Sprintf("%s: $%.2f", sym, price)
            } else {
                results <- fmt.Sprintf("%s: No price data", sym)
            }
        }(symbol)
    }
    
    wg.Wait()
    close(results)
    
    for result := range results {
        fmt.Println(result)
    }
}
```

### Example 3: Financial Data

**Python yfinance**:
```python
import yfinance as yf

ticker = yf.Ticker("AAPL")

# Get financials
financials = ticker.financials
print("Revenue:", financials.loc['Total Revenue'].iloc[0])

# Get balance sheet
balance_sheet = ticker.balance_sheet
print("Total Assets:", balance_sheet.loc['Total Assets'].iloc[0])

# Get key statistics
info = ticker.info
print("Market Cap:", info['marketCap'])
print("P/E Ratio:", info['trailingPE'])
```

**Go yfinance-go**:
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
    runID := fmt.Sprintf("financials-%d", time.Now().Unix())
    
    // Get financials
    financials, err := client.ScrapeFinancials(ctx, "AAPL", runID)
    if err != nil {
        log.Fatal(err)
    }
    
    // Find revenue
    for _, line := range financials.Lines {
        if line.Key == "revenue" {
            revenue := float64(line.Value.Scaled) / float64(line.Value.Scale)
            fmt.Printf("Revenue: %.2f %s\n", revenue, line.CurrencyCode)
            break
        }
    }
    
    // Get balance sheet
    balanceSheet, err := client.ScrapeBalanceSheet(ctx, "AAPL", runID)
    if err != nil {
        log.Printf("Warning: Balance sheet failed: %v", err)
    } else {
        for _, line := range balanceSheet.Lines {
            if line.Key == "total_assets" {
                assets := float64(line.Value.Scaled) / float64(line.Value.Scale)
                fmt.Printf("Total Assets: %.2f %s\n", assets, line.CurrencyCode)
                break
            }
        }
    }
    
    // Get key statistics
    keyStats, err := client.ScrapeKeyStatistics(ctx, "AAPL", runID)
    if err != nil {
        log.Printf("Warning: Key statistics failed: %v", err)
    } else {
        for _, line := range keyStats.Lines {
            if line.Key == "market_cap" {
                marketCap := float64(line.Value.Scaled) / float64(line.Value.Scale)
                fmt.Printf("Market Cap: %.2f %s\n", marketCap, line.CurrencyCode)
            } else if line.Key == "pe_ratio" {
                peRatio := float64(line.Value.Scaled) / float64(line.Value.Scale)
                fmt.Printf("P/E Ratio: %.2f\n", peRatio)
            }
        }
    }
}
```

### Example 4: Data Processing

**Python yfinance**:
```python
import yfinance as yf
import pandas as pd

ticker = yf.Ticker("AAPL")
history = ticker.history(period="1y")

# Calculate moving averages
history['MA20'] = history['Close'].rolling(window=20).mean()
history['MA50'] = history['Close'].rolling(window=50).mean()

# Calculate returns
history['Returns'] = history['Close'].pct_change()

# Display results
print(history[['Close', 'MA20', 'MA50', 'Returns']].tail())
```

**Go yfinance-go**:
```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/AmpyFin/yfinance-go"
)

type PriceData struct {
    Date   time.Time
    Close  float64
    MA20   float64
    MA50   float64
    Return float64
}

func calculateMovingAverage(prices []float64, window int) []float64 {
    if len(prices) < window {
        return make([]float64, len(prices))
    }
    
    ma := make([]float64, len(prices))
    for i := window - 1; i < len(prices); i++ {
        sum := 0.0
        for j := i - window + 1; j <= i; j++ {
            sum += prices[j]
        }
        ma[i] = sum / float64(window)
    }
    
    return ma
}

func calculateReturns(prices []float64) []float64 {
    returns := make([]float64, len(prices))
    for i := 1; i < len(prices); i++ {
        returns[i] = (prices[i] - prices[i-1]) / prices[i-1]
    }
    return returns
}

func main() {
    client := yfinance.NewClient()
    ctx := context.Background()
    runID := fmt.Sprintf("analysis-%d", time.Now().Unix())
    
    // Get historical data
    end := time.Now()
    start := end.AddDate(-1, 0, 0) // 1 year ago
    bars, err := client.FetchDailyBars(ctx, "AAPL", start, end, true, runID)
    if err != nil {
        log.Fatal(err)
    }
    
    // Extract close prices
    prices := make([]float64, len(bars.Bars))
    for i, bar := range bars.Bars {
        prices[i] = float64(bar.Close.Scaled) / float64(bar.Close.Scale)
    }
    
    // Calculate moving averages
    ma20 := calculateMovingAverage(prices, 20)
    ma50 := calculateMovingAverage(prices, 50)
    returns := calculateReturns(prices)
    
    // Create processed data
    processedData := make([]PriceData, len(bars.Bars))
    for i, bar := range bars.Bars {
        processedData[i] = PriceData{
            Date:   bar.EventTime,
            Close:  prices[i],
            MA20:   ma20[i],
            MA50:   ma50[i],
            Return: returns[i],
        }
    }
    
    // Display last 5 results
    fmt.Println("Date\t\tClose\tMA20\tMA50\tReturn")
    fmt.Println("==================================================")
    for i := len(processedData) - 5; i < len(processedData); i++ {
        data := processedData[i]
        fmt.Printf("%s\t%.2f\t%.2f\t%.2f\t%.4f\n",
            data.Date.Format("2006-01-02"),
            data.Close,
            data.MA20,
            data.MA50,
            data.Return)
    }
}
```

## Data Structure Mapping

### Company Information

| Python yfinance | Go yfinance-go | Notes |
|-----------------|----------------|-------|
| `info['longName']` | `companyInfo.LongName` | Company name |
| `info['shortName']` | `companyInfo.ShortName` | Short name |
| `info['exchange']` | `companyInfo.Exchange` | Exchange |
| `info['currency']` | `companyInfo.Currency` | Currency |
| `info['sector']` | ❌ Not available | Use alternative sources |
| `info['industry']` | ❌ Not available | Use alternative sources |
| `info['website']` | ❌ Not available | Use alternative sources |
| `info['address1']` | ❌ Not available | Use alternative sources |
| `info['city']` | ❌ Not available | Use alternative sources |
| `info['state']` | ❌ Not available | Use alternative sources |
| `info['zip']` | ❌ Not available | Use alternative sources |
| `info['country']` | ❌ Not available | Use alternative sources |
| `info['phone']` | ❌ Not available | Use alternative sources |
| `info['employees']` | ❌ Not available | Use alternative sources |
| `info['businessSummary']` | ❌ Not available | Use alternative sources |

### Quote Data

| Python yfinance | Go yfinance-go | Notes |
|-----------------|----------------|-------|
| `info['currentPrice']` | `quote.RegularMarketPrice.Scaled` | Current price (scaled decimal) |
| `info['previousClose']` | `quote.PreviousClose.Scaled` | Previous close (scaled decimal) |
| `info['open']` | `quote.RegularMarketOpen.Scaled` | Open price (scaled decimal) |
| `info['dayHigh']` | `quote.RegularMarketHigh.Scaled` | Day high (scaled decimal) |
| `info['dayLow']` | `quote.RegularMarketLow.Scaled` | Day low (scaled decimal) |
| `info['volume']` | `quote.RegularMarketVolume` | Volume |
| `info['marketCap']` | `keyStats.Lines["market_cap"]` | Market cap (from key stats) |
| `info['trailingPE']` | `keyStats.Lines["pe_ratio"]` | P/E ratio (from key stats) |

### Historical Data

| Python yfinance | Go yfinance-go | Notes |
|-----------------|----------------|-------|
| `history['Open']` | `bar.Open.Scaled` | Open price (scaled decimal) |
| `history['High']` | `bar.High.Scaled` | High price (scaled decimal) |
| `history['Low']` | `bar.Low.Scaled` | Low price (scaled decimal) |
| `history['Close']` | `bar.Close.Scaled` | Close price (scaled decimal) |
| `history['Volume']` | `bar.Volume` | Volume |
| `history.index` | `bar.EventTime` | Date/time |

## Error Handling Differences

### Python yfinance Error Handling
```python
import yfinance as yf

try:
    ticker = yf.Ticker("INVALID_SYMBOL")
    info = ticker.info
    print(info['longName'])
except Exception as e:
    print(f"Error: {e}")
    # Handle error
```

### Go yfinance-go Error Handling
```go
client := yfinance.NewClient()
quote, err := client.FetchQuote(ctx, "INVALID_SYMBOL", runID)
if err != nil {
    log.Printf("Error: %v", err)
    // Handle error
    return
}
// Use quote data
```

## Performance Considerations

### Python yfinance Performance
- Slower for large datasets
- Limited concurrency support
- Memory usage can be high with pandas DataFrames

### Go yfinance-go Performance
- Faster for concurrent requests
- Better memory efficiency
- Built-in goroutine support for concurrency

**Example: Concurrent Processing in Go**
```go
func processSymbolsConcurrently(symbols []string) {
    client := yfinance.NewClientWithSessionRotation()
    ctx := context.Background()
    
    var wg sync.WaitGroup
    semaphore := make(chan struct{}, 10) // Limit concurrency
    
    for _, symbol := range symbols {
        wg.Add(1)
        go func(sym string) {
            defer wg.Done()
            
            semaphore <- struct{}{} // Acquire semaphore
            defer func() { <-semaphore }() // Release semaphore
            
            runID := fmt.Sprintf("concurrent-%d", time.Now().Unix())
            quote, err := client.FetchQuote(ctx, sym, runID)
            if err != nil {
                log.Printf("Error fetching %s: %v", sym, err)
                return
            }
            
            // Process quote
            processQuote(quote)
        }(symbol)
    }
    
    wg.Wait()
}
```

## Best Practices

### 1. Use Session Rotation for Production
```go
// For production use
client := yfinance.NewClientWithSessionRotation()

// For development/testing
client := yfinance.NewClient()
```

### 2. Implement Proper Error Handling
```go
quote, err := client.FetchQuote(ctx, "AAPL", runID)
if err != nil {
    // Handle error appropriately
    log.Printf("Error fetching quote: %v", err)
    return
}
// Use quote data
```

### 3. Handle Scaled Decimals Properly
```go
// ✅ CORRECT: Use the scale field
price := float64(quote.RegularMarketPrice.Scaled) / float64(quote.RegularMarketPrice.Scale)

// ❌ WRONG: Don't hardcode division
// price := quote.RegularMarketPrice.Scaled / 10000
```

### 4. Use Context for Cancellation
```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

quote, err := client.FetchQuote(ctx, "AAPL", runID)
```

### 5. Implement Rate Limiting
```go
// Add delays between requests
time.Sleep(1 * time.Second)

// Or use semaphore for concurrency control
semaphore := make(chan struct{}, 5) // Max 5 concurrent requests
```

### 6. Validate Data Quality
```go
func validateQuote(quote *yfinance.NormalizedQuote) error {
    if quote.RegularMarketPrice == nil {
        return fmt.Errorf("missing market price")
    }
    
    if quote.CurrencyCode == "" {
        return fmt.Errorf("missing currency code")
    }
    
    return nil
}
```

## Migration Checklist

- [ ] **Install Go yfinance-go**: `go get github.com/AmpyFin/yfinance-go`
- [ ] **Update imports**: Replace `import yfinance as yf` with Go imports
- [ ] **Change API calls**: Replace ticker properties with method calls
- [ ] **Update data access**: Replace dictionary access with struct field access
- [ ] **Handle scaled decimals**: Convert scaled decimal format to floats
- [ ] **Implement error handling**: Replace try/catch with explicit error handling
- [ ] **Add context**: Use context for cancellation and timeouts
- [ ] **Implement concurrency**: Use goroutines for concurrent processing
- [ ] **Add rate limiting**: Implement rate limiting for production use
- [ ] **Validate data**: Add data validation and quality checks
- [ ] **Test thoroughly**: Test all functionality with real data
- [ ] **Monitor performance**: Monitor performance and error rates

## Next Steps

- [API Reference](api-reference.md) - Complete API documentation
- [Data Structures](data-structures.md) - Detailed data structure guide
- [Complete Examples](examples.md) - Working code examples
- [Error Handling Guide](error-handling.md) - Comprehensive error handling
- [Method Comparison](method-comparison.md) - Method comparison and use cases
