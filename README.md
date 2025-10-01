# yfinance-go — Yahoo Finance Client for Go

[![Go Version](https://img.shields.io/badge/go-1.23+-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/AmpyFin/yfinance-go)](https://goreportcard.com/report/github.com/AmpyFin/yfinance-go)
[![GoDoc](https://godoc.org/github.com/AmpyFin/yfinance-go?status.svg)](https://godoc.org/github.com/AmpyFin/yfinance-go)

> ⚠️ **IMPORTANT DISCLAIMER** ⚠️
> 
> **This project is NOT affiliated with, endorsed by, or sponsored by Yahoo Finance or Yahoo Inc.**
> 
> This is an **independent, open-source Go client** that accesses publicly available financial data from Yahoo Finance's website. Yahoo Finance does not provide an official API for this data, and this client operates by scraping publicly accessible web pages.
> 
> **Use at your own risk.** Yahoo Finance may change their website structure at any time, which could break this client. We make no guarantees about data accuracy, availability, or compliance with Yahoo Finance's terms of service.
> 
> **Legal Notice:** Users are responsible for ensuring their use of this software complies with Yahoo Finance's terms of service and applicable laws in their jurisdiction.

---

## 🎯 Problem We're Solving

**The Challenge:** Most financial data clients suffer from inconsistent data formats, unreliable APIs, and poor error handling. When building financial applications, developers often face:

- **Inconsistent Data Formats**: Different APIs return data in various shapes and formats
- **Floating Point Precision Issues**: Financial calculations require exact decimal precision
- **Rate Limiting Problems**: Unbounded requests lead to API bans and throttling
- **Poor Error Handling**: Limited retry logic and circuit breaking
- **Currency Conversion Complexity**: Multi-currency support is often missing or buggy
- **No Standardization**: Each client has its own data structures and conventions

**Our Solution:** A production-grade Go client that provides:

✅ **Standardized Data Formats** - Consistent `ampy-proto` message structures  
✅ **High Precision Decimals** - Scaled decimal arithmetic for financial accuracy  
✅ **Robust Rate Limiting** - Built-in backoff, circuit breakers, and session rotation  
✅ **Multi-Currency Support** - Automatic currency conversion with FX providers  
✅ **Production Ready** - Comprehensive error handling, observability, and monitoring  
✅ **Easy Integration** - Simple API with both library and CLI interfaces  

---

## 🚀 Installation

### As a Go Module

```bash
go get github.com/AmpyFin/yfinance-go
```

### From Source

```bash
git clone https://github.com/AmpyFin/yfinance-go.git
cd yfinance-go
go build ./cmd/yfin
```

---

## 📖 Quick Start

### Basic Usage

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
    // Create a new client
    client := yfinance.NewClient()
    ctx := context.Background()
    
    // Fetch daily bars for Apple
    start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
    end := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)
    
    bars, err := client.FetchDailyBars(ctx, "AAPL", start, end, true, "my-run-id")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Fetched %d bars for AAPL\n", len(bars.Bars))
    for _, bar := range bars.Bars {
        price := float64(bar.Close.Scaled) / float64(bar.Close.Scale)
        fmt.Printf("Date: %s, Close: %.4f %s\n", 
            bar.EventTime.Format("2006-01-02"),
            price, bar.CurrencyCode)
    }
}
```

---

## 🔧 API Reference

### Client Creation

```go
// Default client with standard configuration
client := yfinance.NewClient()

// Client with session rotation (recommended for production)
client := yfinance.NewClientWithSessionRotation()
```

### Available Functions

#### 📊 Historical Data

**FetchDailyBars** - Get daily OHLCV data
```go
bars, err := client.FetchDailyBars(ctx, "AAPL", start, end, adjusted, runID)
```

**FetchIntradayBars** - Get intraday data (1m, 5m, 15m, 30m, 60m)
```go
bars, err := client.FetchIntradayBars(ctx, "AAPL", start, end, "1m", runID)
```
> **Note:** Intraday data may not be available for all symbols and may return HTTP 422 errors for some requests.

**FetchWeeklyBars** - Get weekly OHLCV data
```go
bars, err := client.FetchWeeklyBars(ctx, "AAPL", start, end, adjusted, runID)
```

**FetchMonthlyBars** - Get monthly OHLCV data
```go
bars, err := client.FetchMonthlyBars(ctx, "AAPL", start, end, adjusted, runID)
```

#### 💰 Real-time Data

**FetchQuote** - Get current market quote
```go
quote, err := client.FetchQuote(ctx, "AAPL", runID)
```

**FetchMarketData** - Get comprehensive market data
```go
marketData, err := client.FetchMarketData(ctx, "AAPL", runID)
```

#### 🏢 Company Information

**FetchCompanyInfo** - Get basic company information
```go
companyInfo, err := client.FetchCompanyInfo(ctx, "AAPL", runID)
```

**FetchFundamentalsQuarterly** - Get quarterly financials (requires paid subscription)
```go
fundamentals, err := client.FetchFundamentalsQuarterly(ctx, "AAPL", runID)
```

---

## 🕸️ Scrape Fallback System

When Yahoo Finance API endpoints are unavailable, rate-limited, or require paid subscriptions, yfinance-go automatically falls back to web scraping with full data consistency guarantees.

### Key Features
- **Automatic Fallback**: Seamlessly switches between API and scraping
- **Data Consistency**: Identical output formats regardless of source
- **Production Safety**: Respects robots.txt, implements proper rate limiting
- **Comprehensive Coverage**: Access data not available through APIs

### Supported Scrape Endpoints
- **Key Statistics**: P/E ratios, market cap, financial metrics
- **Financials**: Income statements, balance sheets, cash flow
- **Analysis**: Analyst recommendations and price targets
- **Profile**: Company information, executives, business summary
- **News**: Recent news articles and press releases

### Quick Scrape Examples

```go
// Scrape key statistics (not available through free API)
keyStats, err := client.ScrapeKeyStatistics(ctx, "AAPL", runID)

// Scrape financial statements
financials, err := client.ScrapeFinancials(ctx, "AAPL", runID)

// Scrape news articles
news, err := client.ScrapeNews(ctx, "AAPL", runID)
```

### CLI Scraping

```bash
# Scrape key statistics with preview
yfin scrape --ticker AAPL --endpoint key-statistics --preview

# Multiple endpoints with JSON output
yfin scrape --ticker AAPL --endpoints key-statistics,financials,news --preview-json

# Soak testing for production validation
yfin soak --universe-file universe.txt --duration 2h --qps 5 --preview
```

📖 **[Complete Scrape Documentation →](docs/scrape/overview.md)**

---

## 📝 Usage Examples

### Example 1: Fetch Daily Bars for Multiple Symbols

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
    client := yfinance.NewClient()
    ctx := context.Background()
    
    symbols := []string{"AAPL", "GOOGL", "MSFT", "TSLA"}
    start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
    end := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)
    
    var wg sync.WaitGroup
    results := make(chan string, len(symbols))
    
    for _, symbol := range symbols {
        wg.Add(1)
        go func(sym string) {
            defer wg.Done()
            
            bars, err := client.FetchDailyBars(ctx, sym, start, end, true, "batch-run")
            if err != nil {
                results <- fmt.Sprintf("Error fetching %s: %v", sym, err)
                return
            }
            
            results <- fmt.Sprintf("%s: %d bars fetched", sym, len(bars.Bars))
        }(symbol)
    }
    
    wg.Wait()
    close(results)
    
    for result := range results {
        fmt.Println(result)
    }
}
```

### Example 2: Get Current Market Quote

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
    
    quote, err := client.FetchQuote(ctx, "AAPL", "quote-run")
    if err != nil {
        log.Fatal(err)
    }
    
	fmt.Printf("Symbol: %s\n", quote.Security.Symbol)
	if quote.RegularMarketPrice != nil {
		price := float64(quote.RegularMarketPrice.Scaled) / float64(quote.RegularMarketPrice.Scale)
		fmt.Printf("Price: %.4f %s\n", price, quote.CurrencyCode)
	}
	if quote.RegularMarketVolume != nil {
		fmt.Printf("Volume: %d\n", *quote.RegularMarketVolume)
	}
	fmt.Printf("Event Time: %s\n", quote.EventTime.Format("2006-01-02 15:04:05"))
}
```

### Example 3: Fetch Company Information

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/AmpyFin/yfinance-go"
)

func main() {
    client := yfinance.NewClient()
    ctx := context.Background()
    
    companyInfo, err := client.FetchCompanyInfo(ctx, "AAPL", "company-run")
    if err != nil {
        log.Fatal(err)
    }
    
	fmt.Printf("Company: %s\n", companyInfo.LongName)
	fmt.Printf("Exchange: %s\n", companyInfo.Exchange)
	fmt.Printf("Full Exchange: %s\n", companyInfo.FullExchangeName)
	fmt.Printf("Currency: %s\n", companyInfo.Currency)
	fmt.Printf("Instrument Type: %s\n", companyInfo.InstrumentType)
	fmt.Printf("Timezone: %s\n", companyInfo.Timezone)
}
```

### Example 4: Error Handling

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
	
	// Fetch data with proper error handling
	bars, err := client.FetchDailyBars(ctx, "AAPL", 
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
		true, "error-handling-run")
	
	if err != nil {
		log.Printf("Error fetching bars: %v", err)
		return
	}
	
	fmt.Printf("Successfully fetched %d bars\n", len(bars.Bars))
	
	// Handle empty results
	if len(bars.Bars) == 0 {
		fmt.Println("No data available for the specified date range")
		return
	}
	
	// Process the data
	for _, bar := range bars.Bars {
		price := float64(bar.Close.Scaled) / float64(bar.Close.Scale)
		fmt.Printf("Date: %s, Close: %.4f %s\n", 
			bar.EventTime.Format("2006-01-02"),
			price, bar.CurrencyCode)
	}
}
```

---

## 📚 Documentation

### Core Documentation
- **[Installation Guide](docs/install.md)** - Setup and installation instructions
- **[Usage Guide](docs/usage.md)** - Comprehensive usage examples and patterns
- **[API Reference](docs/api-reference.md)** - Complete API documentation with method capabilities and limitations
- **[Data Structures](docs/data-structures.md)** - Detailed data structure guide with field naming conventions
- **[Complete Examples](docs/examples.md)** - Working code examples with data processing and error handling

### Method Comparison & Migration
- **[Method Comparison](docs/method-comparison.md)** - Method comparison table and use case guidance
- **[Migration Guide](docs/migration-guide.md)** - Migration from Python yfinance with feature comparison

### Error Handling & Quality
- **[Error Handling Guide](docs/error-handling.md)** - Comprehensive error handling and troubleshooting
- **[Data Quality Guide](docs/data-quality.md)** - Data quality expectations and validation guidelines
- **[Performance Guide](docs/performance.md)** - Performance optimization and best practices

### Scrape Fallback System
- **[Scrape Overview](docs/scrape/overview.md)** - Architecture and data flow
- **[Configuration Guide](docs/scrape/config.md)** - All configuration options and best practices
- **[CLI Usage](docs/scrape/cli.md)** - Command-line interface examples
- **[Troubleshooting](docs/scrape/troubleshooting.md)** - Common issues and solutions

### Operations & Monitoring
- **[Observability Guide](docs/observability.md)** - Metrics, logging, and monitoring
- **[Soak Testing Guide](docs/soak-testing.md)** - Load testing and validation

### Operator Runbooks
- **[Scrape Fallback Runbook](runbooks/scrape-fallback.md)** - Operational procedures
- **[Incident Response Playbook](runbooks/incident-playbook.md)** - Emergency response procedures

### Examples & Code Samples
- **[Library Examples](examples/library/)** - Go code examples and patterns
- **[CLI Examples](examples/cli/)** - Ready-to-run shell scripts
- **[API Usage Example](examples/api_usage.go)** - Basic API usage
- **[Historical Data Example](examples/historical_data_example.go)** - Time series data

---

## 🖥️ CLI Usage

The `yfin` CLI tool provides command-line access to all functionality:

> **Note:** All CLI commands require a configuration file. Use `--config configs/effective.yaml` or set up your own config file.

### Installation

```bash
# Build from source
go build -o yfin ./cmd/yfin

# Or install globally
go install github.com/AmpyFin/yfinance-go/cmd/yfin@latest
```

### Basic Commands

```bash
# Fetch daily bars for a single symbol
yfin pull --ticker AAPL --start 2024-01-01 --end 2024-12-31 --adjusted split_dividend --preview --config configs/effective.yaml

# Fetch data for multiple symbols from a file
yfin pull --universe-file symbols.txt --start 2024-01-01 --end 2024-12-31 --publish --env prod --config configs/effective.yaml

# Get current quote
yfin quote --tickers AAPL --preview --config configs/effective.yaml

# Get fundamentals (requires paid subscription)
yfin fundamentals --ticker AAPL --preview --config configs/effective.yaml
```

### Scraping Commands

```bash
# Scrape key statistics (not available through free API)
yfin scrape --ticker AAPL --endpoint key-statistics --preview --config configs/effective.yaml

# Multiple endpoints with JSON preview
yfin scrape --ticker AAPL --endpoints key-statistics,financials,analysis --preview-json --config configs/effective.yaml

# News articles preview
yfin scrape --ticker AAPL --endpoint news --preview-news --config configs/effective.yaml

# Health check for endpoints
yfin scrape --ticker AAPL --endpoint key-statistics --check --config configs/effective.yaml
```

### Soak Testing Commands

```bash
# Quick smoke test (10 minutes)
yfin soak --universe-file testdata/universe/soak.txt --endpoints key-statistics,news --duration 10m --concurrency 8 --qps 5 --preview --config configs/effective.yaml

# Full production soak test (2 hours)
yfin soak --universe-file testdata/universe/soak.txt --endpoints key-statistics,financials,analysis,profile,news --duration 2h --concurrency 12 --qps 5 --preview --config configs/effective.yaml
```

### CLI Options

#### Core Options
- `--ticker` - Single symbol to fetch
- `--universe-file` - File containing list of symbols
- `--start`, `--end` - Date range (UTC)
- `--adjusted` - Adjustment policy (raw, split_only, split_dividend)
- `--publish` - Publish to ampy-bus
- `--env` - Environment (dev, staging, prod)
- `--preview` - Show data preview without publishing
- `--concurrency` - Number of concurrent requests
- `--qps` - Requests per second limit

#### Scraping Options
- `--endpoint` - Single endpoint to scrape
- `--endpoints` - Comma-separated list of endpoints
- `--fallback` - Fallback strategy (auto, api-only, scrape-only)
- `--preview-json` - JSON preview of multiple endpoints
- `--preview-news` - Preview news articles
- `--preview-proto` - Preview proto summaries
- `--check` - Validate endpoint accessibility
- `--force` - Override robots.txt (testing only)

#### Soak Testing Options
- `--duration` - Test duration (e.g., 2h, 30m)
- `--memory-check` - Enable memory leak detection
- `--probe-interval` - Correctness probe interval
- `--failure-rate` - Simulated failure rate for testing

📖 **[Complete CLI Documentation →](docs/scrape/cli.md)**

---

## 🎯 Mission & Success Criteria

**Mission**  
Provide a **reliable, consistent, and fast** Yahoo Finance client in Go that speaks **Ampy's canonical contracts** (`ampy-proto`) and optionally **emits** to `ampy-bus`, so ingestion pipelines and research tools work identically across providers.

**Success looks like**
- Library returns **validated `ampy-proto` messages** with correct UTC times, currency semantics, and adjustment flags.
- CLI supports on-demand pulls and batch backfills; ops can **dry‑run**, **preview**, and **publish** with a single command.
- Concurrency and backoff keep **error rates** and **429/503** responses within policy; throughput is tunable and predictable.
- Golden samples round‑trip across **Go → Python/C++** consumers without shape drift.
- Observability shows latency/throughput, decode failures, and backoff behavior; alerts catch regressions.

---

## 📊 Data Coverage

### ✅ Supported Data Types

- **Historical Bars** - Daily, weekly, monthly, and intraday OHLCV data
- **Real-time Quotes** - Current market prices, bid/ask, volume
- **Company Information** - Basic company details, exchange info, industry/sector
- **Market Data** - 52-week ranges, market state, trading hours
- **Multi-Currency Support** - Automatic currency conversion with FX providers

### ❌ API Limitations (Available via Scraping)

These data types require paid Yahoo Finance subscriptions through the API, but are **available through the scrape fallback system**:

- **Financial Statements** - Income statement, balance sheet, cash flow ✅ *Available via scraping*
- **Analyst Recommendations** - Price targets, ratings ✅ *Available via scraping*
- **Key Statistics** - P/E ratios, market cap, financial metrics ✅ *Available via scraping*
- **Company Profiles** - Business summary, executives, sector info ✅ *Available via scraping*
- **News Articles** - Recent news and press releases ✅ *Available via scraping*

### ❌ Not Supported

- **Options Data** - Options chains and pricing
- **Insider Trading** - Insider transactions  
- **Institutional Holdings** - Major shareholders
- **Level 2 Market Data** - Order book, bid/ask depth

### 🌍 Supported Markets

- **US Markets** - NYSE, NASDAQ, AMEX
- **International** - Major exchanges worldwide
- **Currencies** - Forex pairs and cryptocurrency
- **Commodities** - Gold, oil, agricultural products
- **Indices** - S&P 500, Dow Jones, NASDAQ Composite

---

## ⚡ Key Features

### 🛡️ Production Ready
- **Rate Limiting** - Built-in QPS limits and burst control
- **Circuit Breakers** - Automatic failure detection and recovery
- **Retry Logic** - Exponential backoff with jitter
- **Session Rotation** - Prevents IP blocking and rate limits
- **Scrape Fallback** - Automatic API→scrape fallback with robots.txt compliance
- **Observability** - Comprehensive metrics, logs, and tracing
- **Soak Testing** - Built-in load testing and robustness validation

### 💰 Financial Accuracy
- **High Precision Decimals** - Scaled decimal arithmetic for exact calculations
- **Currency Support** - Multi-currency with automatic conversion
- **Corporate Actions** - Split and dividend adjustments
- **Market Hours** - Proper handling of trading sessions and holidays

### 🚀 Performance
- **Concurrent Requests** - Configurable goroutine pools
- **Connection Pooling** - Efficient HTTP connection reuse
- **Caching** - Built-in response caching for FX rates
- **Batching** - Efficient data batching and chunking

### 🔧 Developer Experience
- **Simple API** - Clean, intuitive Go interface
- **Type Safety** - Strongly typed data structures
- **Error Handling** - Comprehensive error types and messages
- **CLI Tool** - Command-line interface for operations
- **Documentation** - Extensive examples and API docs

---

## 📋 Data Formats & Conventions

1) **Time**: All timestamps **UTC** ISO‑8601. Bars use `start` inclusive, `end` exclusive; `event_time` at bar close.  
2) **Precision**: Prices/amounts are **scaled decimals** (`scaled`, `scale`). Volumes are integers.  
3) **Currency**: Attach **ISO‑4217** code to monetary fields and fundamentals lines.  
4) **Identity**: Use `SecurityId` = `{ symbol, mic?, figi?, isin? }`. If MIC is unknown, prefer primary listing inference; document fallback rules.  
5) **Adjustments**: Bars declare `adjusted: true|false` and `adjustment_policy_id: "raw" | "split_only" | "split_dividend"`.  
6) **Lineage**: Every message has `meta.run_id`, `meta.source="yfinance-go"`, `meta.producer="<host|pod>"`, `schema_version`.  
7) **Batching**: Prefer `BarBatch` for efficiency. Maintain **in‑batch order** by `event_time` ascending.  
8) **Compatibility**: Additive evolution only; breaking changes require new major (`bars.v2`, `fundamentals.v2`).

### 💰 Price Formatting

All prices are stored as `ScaledDecimal` with explicit scale for financial precision:

```go
// ✅ CORRECT: Use the scale field to convert back to decimal
price := float64(bar.Close.Scaled) / float64(bar.Close.Scale)
fmt.Printf("Price: %.4f %s\n", price, bar.CurrencyCode)

// ❌ WRONG: Don't hardcode division by 10000
// price := bar.Close.Scaled / 10000  // This is incorrect!
```

**Example:**
- Raw Yahoo price: $221.74
- Stored as: `Scaled: 22174, Scale: 2`
- Converted back: `22174 / 100 = 221.74` ✅

---

---

## ⚙️ Configuration

### Environment Variables

```bash
# Rate limiting
export YFIN_QPS=2.0
export YFIN_BURST=5
export YFIN_CONCURRENCY=32

# Timeouts
export YFIN_TIMEOUT=30s
export YFIN_BACKOFF_BASE=1s
export YFIN_BACKOFF_MAX=10s

# Circuit breaker
export YFIN_CIRCUIT_THRESHOLD=5
export YFIN_CIRCUIT_RESET=30s

# Observability
export YFIN_LOG_LEVEL=info
export YFIN_METRICS_ENABLED=true
```

### Configuration File

```yaml
# config.yaml
yahoo:
  timeout_ms: 30000
  base_url: "https://query1.finance.yahoo.com"

concurrency:
  global_workers: 32
  max_inflight: 64

rate_limit:
  per_host_qps: 2.0
  burst: 5

retry:
  attempts: 3
  backoff_base_ms: 1000
  backoff_max_ms: 10000

circuit_breaker:
  failure_threshold: 5
  reset_timeout_ms: 30000

observability:
  log_level: "info"
  metrics_enabled: true
  tracing_enabled: false
```

---

## 🚀 Quick Start Examples

### Run CLI Examples

```bash
# Make scripts executable
chmod +x examples/cli/*.sh

# Run AAPL preview examples
./examples/cli/preview_aapl.sh

# Run soak testing examples  
./examples/cli/soak_smoke.sh

# Run batch processing examples
./examples/cli/batch_processing.sh
```

### Build Library Examples

```bash
# Build and run library examples
go build -o examples/library/scrape_fallback examples/library/scrape_fallback.go
./examples/library/scrape_fallback
```

📖 **[All Examples →](examples/)**

---

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Development Setup

```bash
# Clone the repository
git clone https://github.com/AmpyFin/yfinance-go.git
cd yfinance-go

# Install dependencies
go mod download

# Run tests
go test ./...

# Build CLI
go build -o yfin ./cmd/yfin

# Run integration tests
go test -tags=integration ./...
```

### Code Style

- Follow Go standard formatting (`gofmt`)
- Use meaningful variable and function names
- Add tests for new functionality
- Update documentation for API changes

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 🙏 Acknowledgments

- **Yahoo Finance** for providing publicly accessible financial data
- **AmpyFin** for the ampy-proto schemas and infrastructure
- **Go Community** for excellent libraries and tools
- **Contributors** who help improve this project

---

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/AmpyFin/yfinance-go/issues)
- **Discussions**: [GitHub Discussions](https://github.com/AmpyFin/yfinance-go/discussions)
- **Documentation**: [Complete Documentation](docs/)
- **API Reference**: [GoDoc](https://godoc.org/github.com/AmpyFin/yfinance-go)

### Getting Help

1. **Check Documentation**: Start with [docs/](docs/) for comprehensive guides
2. **Review Examples**: See [examples/](examples/) for code samples
3. **Search Issues**: Check existing [GitHub Issues](https://github.com/AmpyFin/yfinance-go/issues)
4. **Troubleshooting**: See [docs/scrape/troubleshooting.md](docs/scrape/troubleshooting.md)
5. **Runbooks**: For operational issues, see [runbooks/](runbooks/)

---

**⭐ If you find this project useful, please give it a star on GitHub!**

