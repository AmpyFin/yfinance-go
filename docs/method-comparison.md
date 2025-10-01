# Method Comparison & Use Cases

This guide provides a comprehensive comparison of all yfinance-go methods, their capabilities, limitations, and recommended use cases to help you choose the right method for your needs.

## Overview

yfinance-go provides two main categories of methods:
1. **API Methods**: Direct access to Yahoo Finance API endpoints
2. **Scraping Methods**: Web scraping fallback for data not available through APIs

## Method Comparison Table

| Method | Category | Use Case | Data Returned | Limitations | Performance |
|--------|----------|----------|---------------|-------------|-------------|
| `FetchQuote()` | API | Real-time prices | Current price, volume, market data | No historical data | Fast |
| `FetchDailyBars()` | API | Historical analysis | OHLCV data for date range | Limited to available history | Fast |
| `FetchWeeklyBars()` | API | Long-term trends | Weekly OHLCV data | Limited historical depth | Fast |
| `FetchMonthlyBars()` | API | Long-term analysis | Monthly OHLCV data | Limited historical depth | Fast |
| `FetchIntradayBars()` | API | Short-term analysis | Intraday OHLCV data | May return 422 errors | Fast |
| `ScrapeFinancials()` | Scraping | Fundamental analysis | Revenue, EPS, financial metrics | Quarterly/annual data only | Slower |
| `ScrapeBalanceSheet()` | Scraping | Balance sheet analysis | Assets, liabilities, equity | Quarterly/annual data only | Slower |
| `ScrapeCashFlow()` | Scraping | Cash flow analysis | Operating, investing, financing cash flows | Quarterly/annual data only | Slower |
| `ScrapeKeyStatistics()` | Scraping | Key metrics | P/E ratios, market cap, financial metrics | May be limited for some stocks | Slower |
| `ScrapeAnalysis()` | Scraping | Analyst insights | Estimates, projections | May be limited for smaller companies | Slower |
| `ScrapeAnalystInsights()` | Scraping | Detailed analysis | Comprehensive analyst data | May be limited for smaller companies | Slower |
| `ScrapeNews()` | Scraping | Market sentiment | Recent news articles | May be empty for some stocks | Slower |
| `FetchCompanyInfo()` | API | Basic company data | Name, exchange, currency | No address/executives | Fast |
| `FetchMarketData()` | API | Comprehensive market data | Price, volume, 52-week ranges | No historical data | Fast |
| `FetchFundamentalsQuarterly()` | API | Quarterly fundamentals | Financial statement data | Requires paid subscription | Fast |

## Detailed Method Analysis

### Real-time Data Methods

#### FetchQuote()
**Best for**: Getting current market prices and basic market data

**Returns**:
- Current market price
- Bid/ask prices and sizes
- Daily high/low
- Volume
- Market venue information

**Limitations**:
- No historical data
- May be delayed (not real-time)
- Some fields may be nil

**Use Cases**:
- Real-time price monitoring
- Portfolio valuation
- Market data dashboards
- Trading applications

**Example**:
```go
quote, err := client.FetchQuote(ctx, "AAPL", runID)
if quote.RegularMarketPrice != nil {
    price := float64(quote.RegularMarketPrice.Scaled) / 
            float64(quote.RegularMarketPrice.Scale)
    fmt.Printf("Current price: $%.2f\n", price)
}
```

#### FetchMarketData()
**Best for**: Comprehensive market data including 52-week ranges

**Returns**:
- Current market price
- Daily high/low
- 52-week high/low
- Previous close
- Market time information

**Limitations**:
- No historical data
- No fundamental metrics

**Use Cases**:
- Market analysis dashboards
- Price range analysis
- Market timing applications

### Historical Data Methods

#### FetchDailyBars()
**Best for**: Daily price analysis and backtesting

**Returns**:
- Daily OHLCV data
- Split/dividend adjustments
- Volume information
- Currency information

**Limitations**:
- Limited by Yahoo Finance's historical data availability
- Some symbols may have incomplete data

**Use Cases**:
- Technical analysis
- Backtesting strategies
- Portfolio performance analysis
- Risk management

**Example**:
```go
start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
end := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
bars, err := client.FetchDailyBars(ctx, "AAPL", start, end, true, runID)
```

#### FetchIntradayBars()
**Best for**: Short-term trading and intraday analysis

**Returns**:
- Intraday OHLCV data (1m, 5m, 15m, 30m, 60m)
- High-frequency price data

**Limitations**:
- May return HTTP 422 errors for some symbols
- Limited historical depth (typically 60 days for 1m data)
- Data availability varies by symbol

**Use Cases**:
- Day trading
- Intraday analysis
- High-frequency trading
- Market microstructure analysis

#### FetchWeeklyBars() / FetchMonthlyBars()
**Best for**: Long-term trend analysis

**Returns**:
- Weekly/monthly OHLCV data
- Long-term price trends

**Limitations**:
- Limited historical depth
- Less granular than daily data

**Use Cases**:
- Long-term investment analysis
- Trend identification
- Macro analysis
- Portfolio rebalancing

### Company Information Methods

#### FetchCompanyInfo()
**Best for**: Basic company identification and exchange information

**Returns**:
- Company name (long and short)
- Exchange information
- Currency
- Instrument type
- Timezone information
- First trade date

**⚠️ Important Limitations**:
- **Only returns basic security information**
- **Does NOT include**: Address, executives, website, employees, business summary
- **Use case**: Basic identification only

**Use Cases**:
- Symbol validation
- Exchange identification
- Basic company lookup
- Data source identification

**Example**:
```go
companyInfo, err := client.FetchCompanyInfo(ctx, "AAPL", runID)
fmt.Printf("Company: %s\n", companyInfo.LongName)
fmt.Printf("Exchange: %s\n", companyInfo.Exchange)
// Note: No address, executives, or business summary available
```

### Fundamentals Methods

#### FetchFundamentalsQuarterly()
**Best for**: Quarterly financial data (requires paid subscription)

**Returns**:
- Quarterly financial statements
- EPS, revenue, net income
- Financial ratios

**⚠️ Important Limitations**:
- **Requires Yahoo Finance paid subscription**
- Returns error with exit code 2 if subscription required
- Limited to quarterly data only

**Use Cases**:
- Financial analysis (if you have paid subscription)
- Earnings analysis
- Financial modeling

#### ScrapeFinancials()
**Best for**: Income statement data without paid subscription

**Returns**:
- Revenue, expenses, net income
- EPS (basic and diluted)
- Financial statement line items

**Limitations**:
- Quarterly/annual data only
- Slower than API methods
- May be limited for some stocks

**Use Cases**:
- Financial analysis
- Earnings analysis
- Financial modeling
- Investment research

#### ScrapeBalanceSheet()
**Best for**: Balance sheet analysis

**Returns**:
- Assets, liabilities, equity
- Balance sheet line items
- Financial position data

**Use Cases**:
- Financial health analysis
- Debt analysis
- Asset analysis
- Financial ratio calculations

#### ScrapeCashFlow()
**Best for**: Cash flow analysis

**Returns**:
- Operating, investing, financing cash flows
- Cash flow statement data
- Liquidity analysis

**Use Cases**:
- Cash flow analysis
- Liquidity assessment
- Financial health evaluation
- Investment analysis

#### ScrapeKeyStatistics()
**Best for**: Key financial metrics and ratios

**Returns**:
- P/E ratio, market cap
- Enterprise value
- Financial ratios
- Key performance indicators

**Use Cases**:
- Valuation analysis
- Financial ratio analysis
- Investment screening
- Performance metrics

### Analysis Methods

#### ScrapeAnalysis()
**Best for**: Analyst recommendations and price targets

**Returns**:
- Analyst recommendations
- Price targets
- Earnings estimates
- Analyst ratings

**Limitations**:
- May be limited for smaller companies
- Slower than API methods

**Use Cases**:
- Investment research
- Analyst sentiment analysis
- Price target analysis
- Market sentiment

#### ScrapeAnalystInsights()
**Best for**: Detailed analyst insights

**Returns**:
- Comprehensive analyst data
- Detailed insights
- Analyst reports

**Use Cases**:
- Investment research
- Analyst sentiment analysis
- Market research
- Investment decision support

### News Methods

#### ScrapeNews()
**Best for**: Market sentiment and news analysis

**Returns**:
- Recent news articles
- Press releases
- Market news

**Limitations**:
- May be empty for some stocks
- News availability varies by company

**Use Cases**:
- Market sentiment analysis
- News monitoring
- Event-driven analysis
- Risk assessment

## Recommended Data Fetching Strategy

### 1. Start with Quote
Get current market data for real-time applications:
```go
quote, err := client.FetchQuote(ctx, "AAPL", runID)
```

### 2. Add Historical Data
Get price history for analysis:
```go
bars, err := client.FetchDailyBars(ctx, "AAPL", start, end, true, runID)
```

### 3. Include Fundamentals
Get financial metrics for analysis:
```go
financials, err := client.ScrapeFinancials(ctx, "AAPL", runID)
keyStats, err := client.ScrapeKeyStatistics(ctx, "AAPL", runID)
```

### 4. Add Analysis
Get analyst estimates and recommendations:
```go
analysis, err := client.ScrapeAnalysis(ctx, "AAPL", runID)
```

### 5. Include News
Get market sentiment:
```go
news, err := client.ScrapeNews(ctx, "AAPL", runID)
```

### 6. Handle Company Info
Use for basic identification only:
```go
companyInfo, err := client.FetchCompanyInfo(ctx, "AAPL", runID)
// Note: Limited data, use alternative sources for detailed profiles
```

## Performance Considerations

### Fast Methods (API-based)
- `FetchQuote()`
- `FetchDailyBars()`
- `FetchWeeklyBars()`
- `FetchMonthlyBars()`
- `FetchIntradayBars()`
- `FetchCompanyInfo()`
- `FetchMarketData()`

### Slower Methods (Scraping-based)
- `ScrapeFinancials()`
- `ScrapeBalanceSheet()`
- `ScrapeCashFlow()`
- `ScrapeKeyStatistics()`
- `ScrapeAnalysis()`
- `ScrapeAnalystInsights()`
- `ScrapeNews()`

### Performance Optimization Tips

1. **Use API methods when possible** for better performance
2. **Use session rotation** for high-volume requests
3. **Implement rate limiting** to avoid being blocked
4. **Cache results** when appropriate
5. **Use concurrent processing** for multiple symbols

## Error Handling by Method Type

### API Methods
- Network errors
- Rate limiting (429)
- Invalid symbols
- Data not available

### Scraping Methods
- Parse errors
- Website structure changes
- Rate limiting
- Empty results

### Paid Subscription Methods
- Authentication errors (401)
- Subscription required (exit code 2)

## Data Quality Expectations

### High Quality (API Methods)
- Quotes: Generally available for all active stocks
- Historical Data: Available for most stocks
- Company Info: Basic info only

### Variable Quality (Scraping Methods)
- Financials: Available for most public companies
- Analysis: May be limited for smaller companies
- News: Highly variable, may be empty

## Use Case Recommendations

### Trading Applications
- Use `FetchQuote()` for real-time prices
- Use `FetchIntradayBars()` for short-term analysis
- Use `FetchDailyBars()` for daily analysis

### Investment Research
- Use `ScrapeFinancials()` for financial analysis
- Use `ScrapeKeyStatistics()` for valuation metrics
- Use `ScrapeAnalysis()` for analyst insights
- Use `ScrapeNews()` for market sentiment

### Portfolio Management
- Use `FetchDailyBars()` for performance analysis
- Use `FetchQuote()` for current valuations
- Use `ScrapeKeyStatistics()` for risk metrics

### Market Analysis
- Use `FetchMarketData()` for comprehensive market data
- Use `ScrapeNews()` for market sentiment
- Use `ScrapeAnalysis()` for analyst sentiment

## Next Steps

- [API Reference](api-reference.md) - Complete API documentation
- [Data Structures](data-structures.md) - Detailed data structure guide
- [Complete Examples](examples.md) - Working code examples
- [Error Handling Guide](error-handling.md) - Comprehensive error handling
