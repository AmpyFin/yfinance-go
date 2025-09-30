# Scrape Fallback System Overview

## Architecture & Data Flow

The yfinance-go scraper fallback system provides a robust, production-ready alternative to Yahoo Finance's API endpoints. When API access is unavailable, rate-limited, or requires paid subscriptions, the system automatically falls back to web scraping with full data consistency guarantees.

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   User Request  │    │   Orchestrator  │    │  Data Pipeline  │
│                 │    │                 │    │                 │
│ • CLI Command   │───▶│ • Fallback      │───▶│ • Normalization │
│ • Library Call  │    │   Logic         │    │ • Validation    │
│ • Batch Job     │    │ • Rate Limiting │    │ • Mapping       │
└─────────────────┘    │ • Session Mgmt  │    │ • Publishing    │
                       └─────────────────┘    └─────────────────┘
                                │                       │
                                ▼                       ▼
                       ┌─────────────────┐    ┌─────────────────┐
                       │   Data Sources  │    │     Output      │
                       │                 │    │                 │
                       │ • Yahoo API     │    │ • ampy-proto    │
                       │ • Web Scraping  │    │ • JSON Export   │
                       │ • FX Providers  │    │ • Bus Messages  │
                       └─────────────────┘    └─────────────────┘
```

## Why Scrape Fallback Was Created

### Problem Statement

Yahoo Finance's API has several limitations that affect production systems:

1. **Authentication Requirements**: Many endpoints require paid subscriptions
2. **Rate Limiting**: Aggressive throttling can block legitimate requests
3. **Data Availability**: Some data is only available through web interfaces
4. **Reliability**: API endpoints can be unstable or temporarily unavailable

### Solution Benefits

The scrape fallback system addresses these challenges by:

- **Automatic Fallback**: Seamlessly switches between API and scraping
- **Data Consistency**: Maintains identical output formats regardless of source
- **Production Safety**: Respects robots.txt, implements proper rate limiting
- **Comprehensive Coverage**: Accesses data not available through APIs

## Supported Data Sources

### API Endpoints (Primary)
- **Quotes**: Real-time and delayed market quotes
- **Historical Bars**: Daily, weekly, monthly OHLCV data
- **Fundamentals**: Quarterly financial statements (requires subscription)

### Scrape Endpoints (Fallback)
- **Key Statistics**: P/E ratios, market cap, financial metrics
- **Financials**: Income statements, balance sheets, cash flow
- **Analysis**: Analyst recommendations and price targets
- **Profile**: Company information, executives, business summary
- **News**: Recent news articles and press releases

## Data Flow Architecture

### 1. Request Orchestration

```go
// Simplified orchestration flow
func (o *Orchestrator) ProcessRequest(ctx context.Context, symbol string, endpoint string) (*Data, error) {
    // 1. Determine strategy
    strategy := o.determineStrategy(endpoint)
    
    // 2. Execute with fallback
    switch strategy {
    case "api-first":
        data, err := o.tryAPI(ctx, symbol, endpoint)
        if shouldFallback(err) {
            return o.tryScrape(ctx, symbol, endpoint)
        }
        return data, err
    case "scrape-only":
        return o.tryScrape(ctx, symbol, endpoint)
    }
}
```

### 2. Data Normalization Pipeline

```
Raw Data → Parser → Validator → Mapper → ampy-proto Message
    │         │         │         │            │
    │         │         │         │            ▼
    │         │         │         │    ┌─────────────────┐
    │         │         │         │    │   Standardized  │
    │         │         │         │    │     Output      │
    │         │         │         │    │                 │
    │         │         │         │    │ • UTC Times     │
    │         │         │         │    │ • Scaled Decimals│
    │         │         │         │    │ • ISO Currency  │
    │         │         │         │    │ • Lineage Meta  │
    │         │         │         │    └─────────────────┘
    │         │         │         │
    │         │         │         ▼
    │         │         │  ┌─────────────────┐
    │         │         │  │   Field Mapper  │
    │         │         │  │                 │
    │         │         │  │ • Type Conversion│
    │         │         │  │ • Unit Scaling  │
    │         │         │  │ • Currency Norm │
    │         │         │  └─────────────────┘
    │         │         │
    │         │         ▼
    │         │  ┌─────────────────┐
    │         │  │    Validator    │
    │         │  │                 │
    │         │  │ • Schema Check  │
    │         │  │ • Range Limits  │
    │         │  │ • Required Fields│
    │         │  └─────────────────┘
    │         │
    │         ▼
    │  ┌─────────────────┐
    │  │     Parser      │
    │  │                 │
    │  │ • HTML/JSON     │
    │  │ • Regex Extract │
    │  │ • Error Recovery│
    │  └─────────────────┘
    │
    ▼
┌─────────────────┐
│    Raw Data     │
│                 │
│ • HTML Pages    │
│ • JSON APIs     │
│ • CSV Exports   │
└─────────────────┘
```

## Data Contracts & Guarantees

### 1. Time Standardization
- **All timestamps in UTC**: Consistent timezone handling
- **ISO-8601 format**: Standard time representation
- **Event time semantics**: Bar close times, news publication dates

### 2. Precision & Currency
- **Scaled decimals**: Exact financial arithmetic
- **ISO-4217 currency codes**: Standard currency representation
- **Consistent scaling**: Prices, volumes, and amounts properly scaled

### 3. Lineage & Metadata
- **Run ID tracking**: Every request has unique identifier
- **Source attribution**: Clear data provenance
- **Schema versioning**: Forward/backward compatibility

### 4. Error Handling
- **Typed errors**: Specific error types for different failure modes
- **Graceful degradation**: Partial data better than no data
- **Retry semantics**: Intelligent retry with backoff

## Fallback Decision Logic

### Automatic Fallback Triggers

1. **Authentication Errors (401/403)**
   ```
   API Request → 401 Unauthorized → Scrape Fallback
   ```

2. **Rate Limiting (429)**
   ```
   API Request → 429 Too Many Requests → Backoff → Scrape Fallback
   ```

3. **Server Errors (5xx)**
   ```
   API Request → 500/502/503 → Retry → Scrape Fallback
   ```

4. **Timeout/Network Errors**
   ```
   API Request → Timeout → Retry → Scrape Fallback
   ```

### Fallback Strategy Configuration

```yaml
# Example configuration
scrape:
  fallback_strategy: "auto"  # auto, api-only, scrape-only
  fallback_triggers:
    - "401"  # Authentication required
    - "429"  # Rate limited
    - "5xx"  # Server errors
  retry_before_fallback: 2
  fallback_timeout_ms: 30000
```

## Safety & Compliance

### Robots.txt Compliance

The system respects robots.txt policies with three enforcement levels:

1. **Enforce Mode** (Production)
   - Strictly follows robots.txt rules
   - Blocks disallowed requests
   - Logs compliance violations

2. **Warn Mode** (Development)
   - Logs robots.txt violations
   - Allows requests to proceed
   - Useful for testing

3. **Ignore Mode** (Testing Only)
   - Bypasses robots.txt checks
   - Only for controlled testing environments

### Rate Limiting

- **Per-host QPS limits**: Configurable request rates
- **Burst allowances**: Handle traffic spikes
- **Exponential backoff**: Intelligent retry timing
- **Session rotation**: Distribute load across sessions

### Error Recovery

- **Circuit breakers**: Prevent cascade failures
- **Graceful degradation**: Partial data over failures
- **Health checks**: Monitor endpoint availability
- **Automatic recovery**: Resume normal operation when possible

## Performance Characteristics

### Throughput Benchmarks

| Endpoint | API (req/s) | Scrape (req/s) | Fallback Overhead |
|----------|-------------|----------------|-------------------|
| Quotes | 10-15 | 2-3 | ~200ms |
| Key Stats | N/A | 1-2 | N/A |
| Financials | 5-8 | 1-2 | ~500ms |
| News | N/A | 0.5-1 | N/A |

### Latency Profiles

- **API Requests**: 100-500ms typical
- **Scrape Requests**: 800-2000ms typical
- **Fallback Decision**: <10ms
- **Data Normalization**: 10-50ms

## Integration Points

### Library Usage

```go
// Automatic fallback
client := yfinance.NewClient()
data, err := client.ScrapeKeyStatistics(ctx, "AAPL", runID)

// Explicit fallback control
client := yfinance.NewClientWithConfig(&httpx.Config{
    FallbackStrategy: "scrape-only",
})
```

### CLI Usage

```bash
# Automatic fallback
yfin scrape --ticker AAPL --endpoint key-statistics --fallback auto

# Force scraping
yfin scrape --ticker AAPL --endpoint key-statistics --fallback scrape-only

# Preview mode
yfin scrape --preview-json --ticker AAPL --endpoints key-statistics,financials
```

### Bus Integration

```bash
# Publish scraped data
yfin scrape --ticker AAPL --endpoint news --publish --env prod --topic-prefix ampy
```

## Monitoring & Observability

### Key Metrics

- **Fallback rate**: Percentage of requests using scrape fallback
- **Success rate**: Overall request success percentage
- **Latency percentiles**: P50, P95, P99 response times
- **Error rates**: By error type and endpoint

### Health Indicators

- **Robots compliance**: Zero violations in enforce mode
- **Rate limit hits**: Below configured thresholds
- **Parse success**: High success rate for data extraction
- **Schema validation**: All outputs pass validation

### Alerting Thresholds

- **High fallback rate**: >50% for sustained periods
- **Parse failures**: >5% error rate
- **Rate limit violations**: Any in production
- **Schema drift**: New parsing errors

This overview provides the foundation for understanding the scrape fallback system. For detailed configuration, usage examples, and troubleshooting, see the related documentation sections.
