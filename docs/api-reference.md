# API Reference

This document provides comprehensive documentation for the yfinance-go library API, including method capabilities, data structures, and usage examples.

## Client Creation

### NewClient()
Creates a new Yahoo Finance client with default configuration.

```go
client := yfinance.NewClient()
```

**Configuration**: Standard HTTP client with default timeouts and rate limiting.

### NewClientWithConfig(config *httpx.Config)
Creates a new Yahoo Finance client with custom configuration.

```go
config := &httpx.Config{
    Timeout:     30 * time.Second,
    MaxAttempts: 3,
    QPS:         2.0,
}
client := yfinance.NewClientWithConfig(config)
```

### NewClientWithSessionRotation()
Creates a new Yahoo Finance client with session rotation enabled (recommended for production).

```go
client := yfinance.NewClientWithSessionRotation()
```

**Benefits**: Prevents IP blocking and rate limiting issues in high-volume scenarios.

## Historical Data Methods

### FetchDailyBars()

**Purpose**: Fetch daily OHLCV (Open, High, Low, Close, Volume) data for a symbol.

```go
bars, err := client.FetchDailyBars(ctx, "AAPL", start, end, adjusted, runID)
```

**Parameters**:
- `ctx`: Context for cancellation and timeouts
- `symbol`: Stock symbol (e.g., "AAPL", "MSFT")
- `start`: Start date (inclusive)
- `end`: End date (exclusive)
- `adjusted`: Whether to apply split/dividend adjustments
- `runID`: Unique identifier for tracking

**Returns**: `*norm.NormalizedBarBatch`

**Data Structure**:
```go
type NormalizedBarBatch struct {
    Security Security        `json:"security"`
    Bars     []NormalizedBar `json:"bars"`
    Meta     Meta            `json:"meta"`
}

type NormalizedBar struct {
    Start              time.Time     `json:"start"`
    End                time.Time     `json:"end"`
    Open               ScaledDecimal `json:"open"`
    High               ScaledDecimal `json:"high"`
    Low                ScaledDecimal `json:"low"`
    Close              ScaledDecimal `json:"close"`
    Volume             int64         `json:"volume"`
    Adjusted           bool          `json:"adjusted"`
    AdjustmentPolicyID string        `json:"adjustment_policy_id"`
    CurrencyCode       string        `json:"currency_code"`
    EventTime          time.Time     `json:"event_time"`
    IngestTime         time.Time     `json:"ingest_time"`
    AsOf               time.Time     `json:"as_of"`
}
```

**Limitations**:
- Limited by Yahoo Finance's historical data availability
- Some symbols may have incomplete data for certain date ranges
- Intraday data not available through this method

### FetchIntradayBars()

**Purpose**: Fetch intraday bars with configurable intervals.

```go
bars, err := client.FetchIntradayBars(ctx, "AAPL", start, end, interval, runID)
```

**Parameters**:
- `interval`: "1m", "5m", "15m", "30m", "60m"

**Returns**: `*norm.NormalizedBarBatch` (same structure as daily bars)

**Limitations**:
- May return HTTP 422 errors for some symbols
- Data availability varies by symbol and exchange
- Limited historical depth (typically 60 days for 1m data)

### FetchWeeklyBars()

**Purpose**: Fetch weekly OHLCV data.

```go
bars, err := client.FetchWeeklyBars(ctx, "AAPL", start, end, adjusted, runID)
```

**Returns**: `*norm.NormalizedBarBatch`

### FetchMonthlyBars()

**Purpose**: Fetch monthly OHLCV data.

```go
bars, err := client.FetchMonthlyBars(ctx, "AAPL", start, end, adjusted, runID)
```

**Returns**: `*norm.NormalizedBarBatch`

## Real-time Data Methods

### FetchQuote()

**Purpose**: Get current market quote data.

```go
quote, err := client.FetchQuote(ctx, "AAPL", runID)
```

**Returns**: `*norm.NormalizedQuote`

**Data Structure**:
```go
type NormalizedQuote struct {
    Security            Security       `json:"security"`
    Type                string         `json:"type"`
    Bid                 *ScaledDecimal `json:"bid,omitempty"`
    BidSize             *int64         `json:"bid_size,omitempty"`
    Ask                 *ScaledDecimal `json:"ask,omitempty"`
    AskSize             *int64         `json:"ask_size,omitempty"`
    RegularMarketPrice  *ScaledDecimal `json:"regular_market_price,omitempty"`
    RegularMarketHigh   *ScaledDecimal `json:"regular_market_high,omitempty"`
    RegularMarketLow    *ScaledDecimal `json:"regular_market_low,omitempty"`
    RegularMarketVolume *int64         `json:"regular_market_volume,omitempty"`
    Venue               string         `json:"venue,omitempty"`
    CurrencyCode        string         `json:"currency_code"`
    EventTime           time.Time      `json:"event_time"`
    IngestTime          time.Time      `json:"ingest_time"`
    Meta                Meta           `json:"meta"`
}
```

**Field Naming Convention**: Uses snake_case (`regular_market_price`, `regular_market_volume`)

**Limitations**:
- No historical data
- May be delayed (not real-time)
- Some fields may be nil for certain symbols

### FetchMarketData()

**Purpose**: Get comprehensive market data including 52-week ranges.

```go
marketData, err := client.FetchMarketData(ctx, "AAPL", runID)
```

**Returns**: `*norm.NormalizedMarketData`

**Data Structure**:
```go
type NormalizedMarketData struct {
    Security             Security       `json:"security"`
    RegularMarketPrice   *ScaledDecimal `json:"regular_market_price,omitempty"`
    RegularMarketHigh    *ScaledDecimal `json:"regular_market_high,omitempty"`
    RegularMarketLow     *ScaledDecimal `json:"regular_market_low,omitempty"`
    RegularMarketVolume  *int64         `json:"regular_market_volume,omitempty"`
    FiftyTwoWeekHigh     *ScaledDecimal `json:"fifty_two_week_high,omitempty"`
    FiftyTwoWeekLow      *ScaledDecimal `json:"fifty_two_week_low,omitempty"`
    PreviousClose        *ScaledDecimal `json:"previous_close,omitempty"`
    ChartPreviousClose   *ScaledDecimal `json:"chart_previous_close,omitempty"`
    RegularMarketTime    *time.Time     `json:"regular_market_time,omitempty"`
    HasPrePostMarketData bool           `json:"has_pre_post_market_data"`
    CurrencyCode         string         `json:"currency_code"`
    EventTime            time.Time      `json:"event_time"`
    IngestTime           time.Time      `json:"ingest_time"`
    Meta                 Meta           `json:"meta"`
}
```

## Company Information Methods

### FetchCompanyInfo()

**Purpose**: Get basic company information from chart metadata.

```go
companyInfo, err := client.FetchCompanyInfo(ctx, "AAPL", runID)
```

**Returns**: `*norm.NormalizedCompanyInfo`

**Data Structure**:
```go
type NormalizedCompanyInfo struct {
    Security         Security
    LongName         string     `json:"long_name"`
    ShortName        string     `json:"short_name"`
    Exchange         string     `json:"exchange"`
    FullExchangeName string     `json:"full_exchange_name"`
    Currency         string     `json:"currency"`
    InstrumentType   string     `json:"instrument_type"`
    FirstTradeDate   *time.Time `json:"first_trade_date,omitempty"`
    Timezone         string     `json:"timezone"`
    ExchangeTimezone string     `json:"exchange_timezone"`
    EventTime        time.Time  `json:"event_time"`
    IngestTime       time.Time  `json:"ingest_time"`
    Meta             Meta       `json:"meta"`
}
```

**⚠️ Important Limitations**:
- **Only returns basic security information**
- **Does NOT include**: Company address, executives, website, employees, business summary
- **Use case**: Basic identification and exchange information only
- **For detailed company profiles**: Use alternative data sources or consider contributing to expose internal profile scraping functionality

## Fundamentals Methods

### FetchFundamentalsQuarterly()

**Purpose**: Fetch quarterly financial fundamentals (requires paid subscription).

```go
fundamentals, err := client.FetchFundamentalsQuarterly(ctx, "AAPL", runID)
```

**Returns**: `*norm.NormalizedFundamentalsSnapshot`

**Data Structure**:
```go
type NormalizedFundamentalsSnapshot struct {
    Security Security                     `json:"security"`
    Lines    []NormalizedFundamentalsLine `json:"lines"`
    Source   string                       `json:"source"`
    AsOf     time.Time                    `json:"as_of"`
    Meta     Meta                         `json:"meta"`
}

type NormalizedFundamentalsLine struct {
    Key          string        `json:"key"`
    Value        ScaledDecimal `json:"value"`
    CurrencyCode string        `json:"currency_code"`
    PeriodStart  time.Time     `json:"period_start"`
    PeriodEnd    time.Time     `json:"period_end"`
}
```

**⚠️ Important Limitations**:
- **Requires Yahoo Finance paid subscription**
- Returns error with exit code 2 if subscription required
- Limited to quarterly data only

## Scraping Methods (AMPY-PROTO Data)

These methods return structured `ampy-proto` data and are available even when API endpoints require paid subscriptions.

### ScrapeFinancials()

**Purpose**: Scrape financial statements (income statement).

```go
financials, err := client.ScrapeFinancials(ctx, "AAPL", runID)
```

**Returns**: `*fundamentalsv1.FundamentalsSnapshot`

**Data Available**: Revenue, expenses, net income, EPS, etc.

### ScrapeBalanceSheet()

**Purpose**: Scrape balance sheet data.

```go
balanceSheet, err := client.ScrapeBalanceSheet(ctx, "AAPL", runID)
```

**Returns**: `*fundamentalsv1.FundamentalsSnapshot`

### ScrapeCashFlow()

**Purpose**: Scrape cash flow statement data.

```go
cashFlow, err := client.ScrapeCashFlow(ctx, "AAPL", runID)
```

**Returns**: `*fundamentalsv1.FundamentalsSnapshot`

### ScrapeKeyStatistics()

**Purpose**: Scrape key financial metrics and ratios.

```go
keyStats, err := client.ScrapeKeyStatistics(ctx, "AAPL", runID)
```

**Returns**: `*fundamentalsv1.FundamentalsSnapshot`

**Data Available**: P/E ratio, market cap, enterprise value, etc.

### ScrapeAnalysis()

**Purpose**: Scrape analyst recommendations and price targets.

```go
analysis, err := client.ScrapeAnalysis(ctx, "AAPL", runID)
```

**Returns**: `*fundamentalsv1.FundamentalsSnapshot`

### ScrapeAnalystInsights()

**Purpose**: Scrape detailed analyst insights.

```go
insights, err := client.ScrapeAnalystInsights(ctx, "AAPL", runID)
```

**Returns**: `*fundamentalsv1.FundamentalsSnapshot`

### ScrapeNews()

**Purpose**: Scrape recent news articles and press releases.

```go
news, err := client.ScrapeNews(ctx, "AAPL", runID)
```

**Returns**: `[]*newsv1.NewsItem`

**Data Structure**:
```go
type NewsItem struct {
    Title       string    `json:"title"`
    Summary     string    `json:"summary"`
    Url         string    `json:"url"`
    PublishedAt time.Time `json:"published_at"`
    Source      string    `json:"source"`
    // ... other fields
}
```

**Limitations**:
- May return empty results for some symbols
- News availability varies by company and market activity

### ScrapeAllFundamentals()

**Purpose**: Scrape all available fundamentals data types.

```go
snapshots, err := client.ScrapeAllFundamentals(ctx, "AAPL", runID)
```

**Returns**: `[]*fundamentalsv1.FundamentalsSnapshot`

**Includes**: Financials, balance sheet, cash flow, key statistics, analysis, and analyst insights.

## Data Structure Conventions

### Field Naming
- **Quote Data**: Uses snake_case (`regular_market_price`, `regular_market_volume`)
- **Company Info**: Uses snake_case (`long_name`, `first_trade_date`)
- **Financial Data**: Uses snake_case (`eps_basic`, `eps_diluted`)

### Scaled Decimal Format
All monetary values use a scaled decimal format for financial precision:

```go
type ScaledDecimal struct {
    Scaled int64 `json:"scaled"`
    Scale  int   `json:"scale"`
}
```

**Conversion Example**:
```json
{
  "scaled": 25503,
  "scale": 2
}
```
**Value**: `25503 / (10^2) = 255.03`

### Common Field Mappings

| Expected Field | Actual Field | Data Type | Notes |
|----------------|--------------|-----------|-------|
| `name` | `long_name` | string | Company name |
| `price` | `regular_market_price.scaled` | scaled decimal | Current price |
| `volume` | `regular_market_volume` | number | Trading volume |
| `address` | Not available | N/A | Use alternative sources |
| `executives` | Not available | N/A | Use alternative sources |
| `website` | Not available | N/A | Use alternative sources |
| `employees` | Not available | N/A | Use alternative sources |
| `business_summary` | Not available | N/A | Use alternative sources |

## Error Handling

### Common Error Scenarios

1. **"no news articles found" Error**
   - **Cause**: Yahoo Finance may not have news for certain symbols
   - **Solution**: Handle gracefully, this is not a critical error

2. **Rate Limiting**
   - **Symptoms**: Slow responses, timeouts
   - **Solution**: Use `NewClientWithSessionRotation()` for high-volume requests

3. **Empty Company Information**
   - **Cause**: `FetchCompanyInfo()` only returns basic security data
   - **Solution**: Use alternative data sources for detailed company profiles

4. **Data Structure Access**
   - **Issue**: Field names don't match expectations
   - **Solution**: Always check the actual JSON structure first

5. **Paid Subscription Required**
   - **Cause**: Some endpoints require Yahoo Finance paid subscription
   - **Solution**: Use scraping methods as fallback

## Best Practices

### Client Configuration
```go
// For production use with high volume
client := yfinance.NewClientWithSessionRotation()

// For development/testing
client := yfinance.NewClient()
```

### Rate Limiting
- Yahoo Finance has rate limits
- Use session rotation for high-volume requests
- Implement backoff strategies for production use

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
    
    return nil
}
```

## Next Steps

- [Data Structure Guide](data-structures.md) - Detailed explanation of all data types
- [Complete Examples](examples.md) - Working code examples
- [Error Handling Guide](error-handling.md) - Comprehensive error handling
- [Migration Guide](migration.md) - From Python yfinance
