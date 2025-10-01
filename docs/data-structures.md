# Data Structures Guide

This guide provides detailed documentation of all data structures returned by yfinance-go, including field naming conventions, data types, and usage examples.

## Overview

yfinance-go returns data in standardized, normalized formats that ensure consistency across different data sources. All monetary values use scaled decimal arithmetic for financial precision, and timestamps are in UTC.

## Core Data Types

### ScaledDecimal

All monetary values in yfinance-go use a scaled decimal format to ensure financial precision and avoid floating-point errors.

```go
type ScaledDecimal struct {
    Scaled int64 `json:"scaled"`
    Scale  int   `json:"scale"`
}
```

**Purpose**: Represents decimal numbers with explicit precision for financial calculations.

**Conversion Formula**: `value = scaled / (10^scale)`

**Examples**:
```go
// Price: $255.03
{
    "scaled": 25503,
    "scale": 2
}

// Price: $1,234.5678
{
    "scaled": 12345678,
    "scale": 4
}

// Price: $100.00
{
    "scaled": 10000,
    "scale": 2
}
```

**Conversion Helper Function**:
```go
func formatScaledDecimal(scaled int64, scale int32) string {
    if scale == 0 {
        return fmt.Sprintf("%d", scaled)
    }
    divisor := 1.0
    for i := int32(0); i < scale; i++ {
        divisor *= 10.0
    }
    value := float64(scaled) / divisor
    return fmt.Sprintf("%.2f", value)
}

// Usage
price := float64(quote.RegularMarketPrice.Scaled) / 
        float64(1<<uint(quote.RegularMarketPrice.Scale))
fmt.Printf("Price: $%.2f\n", price)
```

### Security

Represents a financial security with identification information.

```go
type Security struct {
    Symbol string `json:"symbol"`
    MIC    string `json:"mic,omitempty"`
}
```

**Fields**:
- `Symbol`: Stock symbol (e.g., "AAPL", "MSFT")
- `MIC`: Market Identifier Code (e.g., "XNAS" for NASDAQ)

### Meta

Contains metadata for tracking and lineage.

```go
type Meta struct {
    RunID         string `json:"run_id"`
    Source        string `json:"source"`
    Producer      string `json:"producer"`
    SchemaVersion string `json:"schema_version"`
    IngestTime    time.Time `json:"ingest_time"`
}
```

## Historical Data Structures

### NormalizedBar

Represents a single price bar (OHLCV data).

```go
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

**Field Descriptions**:
- `Start`: Bar start time (inclusive)
- `End`: Bar end time (exclusive)
- `Open/High/Low/Close`: Price data as scaled decimals
- `Volume`: Trading volume as integer
- `Adjusted`: Whether prices are adjusted for splits/dividends
- `AdjustmentPolicyID`: "raw", "split_only", or "split_dividend"
- `CurrencyCode`: ISO-4217 currency code (e.g., "USD", "EUR")
- `EventTime`: When the bar was recorded
- `IngestTime`: When data was ingested
- `AsOf`: Data as-of timestamp

**Usage Example**:
```go
for _, bar := range bars.Bars {
    // Convert scaled decimal to float
    open := float64(bar.Open.Scaled) / float64(bar.Open.Scale)
    high := float64(bar.High.Scaled) / float64(bar.High.Scale)
    low := float64(bar.Low.Scaled) / float64(bar.Low.Scale)
    close := float64(bar.Close.Scaled) / float64(bar.Close.Scale)
    
    fmt.Printf("Date: %s, OHLC: %.2f/%.2f/%.2f/%.2f, Volume: %d\n",
        bar.EventTime.Format("2006-01-02"),
        open, high, low, close, bar.Volume)
}
```

### NormalizedBarBatch

Contains a collection of bars for a single security.

```go
type NormalizedBarBatch struct {
    Security Security        `json:"security"`
    Bars     []NormalizedBar `json:"bars"`
    Meta     Meta            `json:"meta"`
}
```

## Real-time Data Structures

### NormalizedQuote

Represents current market quote data.

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

**Usage Example**:
```go
quote, err := client.FetchQuote(ctx, "AAPL", runID)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Symbol: %s\n", quote.Security.Symbol)
if quote.RegularMarketPrice != nil {
    price := float64(quote.RegularMarketPrice.Scaled) / 
            float64(quote.RegularMarketPrice.Scale)
    fmt.Printf("Price: %.4f %s\n", price, quote.CurrencyCode)
}
if quote.RegularMarketVolume != nil {
    fmt.Printf("Volume: %d\n", *quote.RegularMarketVolume)
}
fmt.Printf("Event Time: %s\n", quote.EventTime.Format("2006-01-02 15:04:05"))
```

### NormalizedMarketData

Comprehensive market data including 52-week ranges.

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

## Company Information Structures

### NormalizedCompanyInfo

Basic company information from chart metadata.

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

**Usage Example**:
```go
companyInfo, err := client.FetchCompanyInfo(ctx, "AAPL", runID)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Company: %s\n", companyInfo.LongName)
fmt.Printf("Exchange: %s\n", companyInfo.Exchange)
fmt.Printf("Full Exchange: %s\n", companyInfo.FullExchangeName)
fmt.Printf("Currency: %s\n", companyInfo.Currency)
fmt.Printf("Instrument Type: %s\n", companyInfo.InstrumentType)
fmt.Printf("Timezone: %s\n", companyInfo.Timezone)
```

## Fundamentals Structures

### NormalizedFundamentalsLine

Single line item in financial statements.

```go
type NormalizedFundamentalsLine struct {
    Key          string        `json:"key"`
    Value        ScaledDecimal `json:"value"`
    CurrencyCode string        `json:"currency_code"`
    PeriodStart  time.Time     `json:"period_start"`
    PeriodEnd    time.Time     `json:"period_end"`
}
```

**Common Keys**:
- `eps_basic`: Basic earnings per share
- `eps_diluted`: Diluted earnings per share
- `revenue`: Total revenue
- `ebitda`: Earnings before interest, taxes, depreciation, and amortization
- `net_income`: Net income
- `total_assets`: Total assets
- `total_liabilities`: Total liabilities
- `market_cap`: Market capitalization
- `pe_ratio`: Price-to-earnings ratio

### NormalizedFundamentalsSnapshot

Collection of financial line items.

```go
type NormalizedFundamentalsSnapshot struct {
    Security Security                     `json:"security"`
    Lines    []NormalizedFundamentalsLine `json:"lines"`
    Source   string                       `json:"source"`
    AsOf     time.Time                    `json:"as_of"`
    Meta     Meta                         `json:"meta"`
}
```

**Usage Example**:
```go
financials, err := client.ScrapeFinancials(ctx, "AAPL", runID)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Found %d financial line items:\n", len(financials.Lines))
for _, line := range financials.Lines {
    value := float64(line.Value.Scaled) / float64(line.Value.Scale)
    fmt.Printf("%s: %.2f %s\n", line.Key, value, line.CurrencyCode)
}
```

## AMPY-PROTO Structures

The scraping methods return structured `ampy-proto` data that follows the same patterns but with additional metadata and validation.

### FundamentalsSnapshot (AMPY-PROTO)

```go
type FundamentalsSnapshot struct {
    Meta    *Meta                    `protobuf:"bytes,1,opt,name=meta,proto3"`
    Lines   []*FundamentalsLine      `protobuf:"bytes,2,rep,name=lines,proto3"`
    // ... additional fields
}
```

### NewsItem (AMPY-PROTO)

```go
type NewsItem struct {
    Title       string    `json:"title"`
    Summary     string    `json:"summary"`
    Url         string    `json:"url"`
    PublishedAt time.Time `json:"published_at"`
    Source      string    `json:"source"`
    // ... additional fields
}
```

## Field Naming Conventions

### Consistent Patterns

All data structures follow consistent naming conventions:

1. **Snake Case**: All field names use snake_case
   - `regular_market_price`
   - `regular_market_volume`
   - `fifty_two_week_high`

2. **Pointer Fields**: Optional fields are pointers (`*ScaledDecimal`, `*int64`)
   - Allows distinction between zero values and missing data

3. **Time Fields**: All timestamps are `time.Time` in UTC
   - `event_time`: When the data event occurred
   - `ingest_time`: When data was ingested
   - `as_of`: Data as-of timestamp

4. **Currency Fields**: All monetary fields include currency code
   - `currency_code`: ISO-4217 code (e.g., "USD", "EUR")

## Data Validation

### Common Validation Patterns

```go
func validateBarData(bar *NormalizedBar) error {
    if bar.Open.Scale < 0 || bar.High.Scale < 0 || 
       bar.Low.Scale < 0 || bar.Close.Scale < 0 {
        return fmt.Errorf("invalid scale values")
    }
    
    if bar.Volume < 0 {
        return fmt.Errorf("negative volume")
    }
    
    if bar.Start.After(bar.End) {
        return fmt.Errorf("start time after end time")
    }
    
    return nil
}

func validateQuoteData(quote *NormalizedQuote) error {
    if quote.RegularMarketPrice != nil {
        if quote.RegularMarketPrice.Scale < 0 {
            return fmt.Errorf("invalid price scale")
        }
    }
    
    if quote.CurrencyCode == "" {
        return fmt.Errorf("missing currency code")
    }
    
    return nil
}
```

## Common Field Mappings

| Expected Field | Actual Field | Data Type | Notes |
|----------------|--------------|-----------|-------|
| `name` | `long_name` | string | Company name |
| `price` | `regular_market_price.scaled` | scaled decimal | Current price |
| `volume` | `regular_market_volume` | number | Trading volume |
| `high` | `regular_market_high.scaled` | scaled decimal | Daily high |
| `low` | `regular_market_low.scaled` | scaled decimal | Daily low |
| `52_week_high` | `fifty_two_week_high.scaled` | scaled decimal | 52-week high |
| `52_week_low` | `fifty_two_week_low.scaled` | scaled decimal | 52-week low |
| `address` | Not available | N/A | Use alternative sources |
| `executives` | Not available | N/A | Use alternative sources |
| `website` | Not available | N/A | Use alternative sources |
| `employees` | Not available | N/A | Use alternative sources |
| `business_summary` | Not available | N/A | Use alternative sources |

## Best Practices

### 1. Always Check for Nil Pointers
```go
if quote.RegularMarketPrice != nil {
    price := float64(quote.RegularMarketPrice.Scaled) / 
            float64(quote.RegularMarketPrice.Scale)
    fmt.Printf("Price: %.2f\n", price)
}
```

### 2. Use Proper Scale Conversion
```go
// ✅ CORRECT: Use the scale field
price := float64(bar.Close.Scaled) / float64(bar.Close.Scale)

// ❌ WRONG: Don't hardcode division
// price := bar.Close.Scaled / 10000  // This is incorrect!
```

### 3. Handle Missing Data Gracefully
```go
func processFinancials(financials *NormalizedFundamentalsSnapshot) {
    if len(financials.Lines) == 0 {
        log.Printf("Warning: No financial data available")
        return
    }
    
    for _, line := range financials.Lines {
        if line.Value.Scale == 0 && line.Value.Scaled == 0 {
            log.Printf("Warning: Zero value for %s", line.Key)
            continue
        }
        
        // Process the line item
        value := float64(line.Value.Scaled) / float64(line.Value.Scale)
        fmt.Printf("%s: %.2f %s\n", line.Key, value, line.CurrencyCode)
    }
}
```

### 4. Validate Data Quality
```go
func validateDataQuality(data interface{}) error {
    switch v := data.(type) {
    case *NormalizedBarBatch:
        if len(v.Bars) == 0 {
            return fmt.Errorf("no bars in batch")
        }
    case *NormalizedQuote:
        if v.RegularMarketPrice == nil {
            return fmt.Errorf("missing market price")
        }
    case *NormalizedFundamentalsSnapshot:
        if len(v.Lines) == 0 {
            return fmt.Errorf("no financial lines")
        }
    }
    return nil
}
```

## Next Steps

- [Complete Examples](examples.md) - Working code examples with data processing
- [Error Handling Guide](error-handling.md) - Comprehensive error handling
- [API Reference](api-reference.md) - Complete API documentation
