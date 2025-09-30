# CLI Usage Guide

## Overview

The `yfin` CLI provides comprehensive access to the scrape fallback system with intuitive commands and extensive configuration options. This guide covers all scraping-related CLI functionality with practical examples.

## Command Structure

```bash
yfin [global-flags] <command> [command-flags] [arguments]
```

### Global Flags
- `--config`: Configuration file path
- `--log-level`: Logging level (debug, info, warn, error)
- `--run-id`: Unique identifier for request tracking
- `--concurrency`: Number of concurrent workers
- `--qps`: Global queries per second limit
- `--timeout`: HTTP timeout duration

## Scrape Command

The primary command for web scraping operations.

### Basic Syntax
```bash
yfin scrape [flags]
```

### Core Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--ticker` | string | - | Single ticker symbol to scrape |
| `--endpoint` | string | - | Specific endpoint to scrape |
| `--endpoints` | string | - | Comma-separated list of endpoints |
| `--fallback` | string | `auto` | Fallback strategy (auto/api-only/scrape-only) |
| `--preview` | bool | `false` | Show data preview without processing |
| `--preview-json` | bool | `false` | Show JSON preview of multiple endpoints |
| `--preview-news` | bool | `false` | Preview news articles without proto conversion |
| `--preview-proto` | bool | `false` | Preview proto summaries without full output |
| `--check` | bool | `false` | Validate endpoint accessibility |
| `--force` | bool | `false` | Override robots.txt restrictions (testing only) |

## Usage Examples

### Example 1: Basic Key Statistics Scraping

```bash
# Scrape key statistics for Apple
yfin scrape --config configs/dev.yaml --ticker AAPL --endpoint key-statistics --preview

# Expected output:
# Key Statistics for AAPL:
# Market Cap: $2.85T
# P/E Ratio: 28.45
# EPS: $6.13
# Revenue: $394.33B
# Employees: 164,000
```

**Why this example**: Demonstrates the most common scraping use case - getting key financial metrics not available through free APIs.

**When to use**: 
- Need P/E ratios, market cap, or other key metrics
- API access is rate-limited or requires subscription
- Building financial dashboards or screening tools

### Example 2: Multiple Endpoint Preview

```bash
# Preview multiple endpoints for comprehensive data
yfin scrape --config configs/dev.yaml --ticker MSFT --preview-json --endpoints key-statistics,financials,analysis,profile

# Expected output:
# {
#   "key-statistics": {
#     "market_cap": "2.8T",
#     "pe_ratio": 32.1,
#     "eps": 11.05
#   },
#   "financials": {
#     "revenue": "211.9B",
#     "net_income": "72.4B",
#     "currency": "USD"
#   },
#   "analysis": {
#     "recommendation": "BUY",
#     "target_price": 420.0,
#     "analyst_count": 45
#   },
#   "profile": {
#     "company_name": "Microsoft Corporation",
#     "sector": "Technology",
#     "employees": 221000
#   }
# }
```

**Why this example**: Shows how to efficiently gather comprehensive company data in a single command.

**When to use**:
- Building company profiles or research reports
- Need data from multiple sources simultaneously
- Comparing data consistency across endpoints

### Example 3: News Article Scraping

```bash
# Scrape and preview news articles
yfin scrape --config configs/dev.yaml --ticker TSLA --preview-news

# Expected output:
# News Articles for TSLA (15 articles found):
# 
# 1. "Tesla Delivers Record Q4 Numbers" 
#    Source: Reuters | Published: 2024-01-03T14:30:00Z
#    Summary: Tesla reported record quarterly deliveries...
# 
# 2. "Musk Announces New Gigafactory Plans"
#    Source: Bloomberg | Published: 2024-01-02T09:15:00Z
#    Summary: CEO Elon Musk revealed plans for expansion...
# 
# 3. "Tesla Stock Rises on Delivery Beat"
#    Source: CNBC | Published: 2024-01-03T16:45:00Z
#    Summary: Shares jumped 8% in after-hours trading...
```

**Why this example**: Demonstrates news scraping capabilities for sentiment analysis and market intelligence.

**When to use**:
- Building news aggregation systems
- Performing sentiment analysis on stocks
- Tracking company-specific news events

## Advanced Usage Patterns

### Example 4: Fallback Strategy Control

```bash
# Force API-only mode (will fail if API unavailable)
yfin scrape --config configs/prod.yaml --ticker AAPL --endpoint quote --fallback api-only

# Force scrape-only mode (bypass API entirely)
yfin scrape --config configs/prod.yaml --ticker AAPL --endpoint key-statistics --fallback scrape-only

# Automatic fallback (recommended for production)
yfin scrape --config configs/prod.yaml --ticker AAPL --endpoint financials --fallback auto
```

**Why this example**: Shows how to control data source selection for different operational scenarios.

**When to use**:
- Testing API vs scrape data consistency
- Debugging fallback behavior
- Ensuring specific data source usage

### Example 5: Batch Processing with Universe Files

```bash
# Create universe file
echo -e "AAPL\nMSFT\nGOOGL\nTSLA\nAMZN" > universe.txt

# Process multiple tickers
yfin scrape --config configs/prod.yaml --universe-file universe.txt --endpoint key-statistics --preview-proto

# Expected output:
# Processing 5 tickers...
# 
# AAPL: ✓ Key statistics extracted (45 fields)
# MSFT: ✓ Key statistics extracted (43 fields)  
# GOOGL: ✓ Key statistics extracted (44 fields)
# TSLA: ✓ Key statistics extracted (42 fields)
# AMZN: ✓ Key statistics extracted (46 fields)
# 
# Summary: 5/5 successful, 0 failures
# Total fields extracted: 220
# Average processing time: 1.2s per ticker
```

**Why this example**: Demonstrates efficient batch processing for multiple securities.

**When to use**:
- Building market screening tools
- Bulk data collection for analysis
- Regular data updates for portfolios

### Example 6: Endpoint Validation and Health Checks

```bash
# Check endpoint accessibility
yfin scrape --config configs/prod.yaml --ticker AAPL --endpoint profile --check

# Expected output:
# Endpoint Health Check for AAPL/profile:
# ✓ Robots.txt allows access
# ✓ Page loads successfully (1.2s)
# ✓ Required data fields present
# ✓ Parsing successful
# ✓ Schema validation passed
# 
# Status: HEALTHY
# Response time: 1.2s
# Data completeness: 95%
# Last updated: 2024-01-03T10:30:00Z
```

**Why this example**: Shows how to validate endpoint health before production use.

**When to use**:
- Pre-deployment health checks
- Monitoring endpoint availability
- Debugging parsing issues

## Soak Testing Command

Comprehensive load testing and robustness validation.

### Basic Soak Test

```bash
# Short smoke test (10 minutes)
yfin soak --config configs/dev.yaml \
  --universe-file testdata/universe/soak.txt \
  --endpoints key-statistics,financials,analysis,profile,news \
  --fallback auto \
  --duration 10m \
  --concurrency 8 \
  --qps 5 \
  --preview
```

**Expected Output:**
```
Starting soak test...
Universe: 64 tickers
Endpoints: 5 (key-statistics, financials, analysis, profile, news)
Duration: 10m0s
Workers: 8
Target QPS: 5.0

[10:30:15] INFO: Soak test started
[10:30:45] INFO: 150 requests completed, 98.7% success rate
[10:31:15] INFO: 300 requests completed, 97.3% success rate
[10:32:15] INFO: Memory usage stable: 45MB (+2MB from start)
...

=== SOAK TEST RESULTS ===
Duration: 10m15s
Total Requests: 3,045
Success Rate: 96.8%
Actual QPS: 4.95
Fallback Rate: 23.4%
Memory Growth: +8MB (stable)
Goroutine Growth: +2 (stable)

=== ENDPOINT BREAKDOWN ===
key-statistics: 612 req, 98.2% success, avg 1.2s
financials: 608 req, 95.1% success, avg 1.8s
analysis: 615 req, 97.8% success, avg 1.1s
profile: 605 req, 94.9% success, avg 2.1s
news: 605 req, 98.5% success, avg 1.4s
```

### Production Soak Test

```bash
# Full production soak test (2 hours)
yfin soak --config configs/prod.yaml \
  --universe-file production-universe.txt \
  --endpoints key-statistics,financials,analysis,profile,news \
  --fallback auto \
  --duration 2h \
  --concurrency 12 \
  --qps 5 \
  --memory-check \
  --probe-interval 1h \
  --preview
```

### Soak Test with Publishing

```bash
# Test with actual message publishing
yfin soak --config configs/staging.yaml \
  --universe-file small-universe.txt \
  --endpoints news \
  --fallback auto \
  --duration 30m \
  --concurrency 8 \
  --qps 3 \
  --publish \
  --env staging \
  --topic-prefix ampy.staging
```

## Error Handling and Debugging

### Debug Mode

```bash
# Enable debug logging
yfin scrape --config configs/dev.yaml --log-level debug --ticker AAPL --endpoint key-statistics

# Expected debug output:
# [DEBUG] Loading configuration from configs/dev.yaml
# [DEBUG] Creating HTTP client with timeout 30s
# [DEBUG] Checking robots.txt for finance.yahoo.com
# [DEBUG] Robots.txt allows /quote/AAPL/key-statistics
# [DEBUG] Making request to https://finance.yahoo.com/quote/AAPL/key-statistics
# [DEBUG] Response received: 200 OK (1.2s)
# [DEBUG] Parsing HTML content (2.1MB)
# [DEBUG] Extracted 45 data fields
# [DEBUG] Schema validation passed
# [INFO] Key statistics scraped successfully
```

### Force Mode (Testing Only)

```bash
# Override robots.txt restrictions (use with caution)
yfin scrape --config configs/test.yaml --ticker AAPL --endpoint profile --force --preview

# Warning output:
# ⚠️  WARNING: --force flag overrides robots.txt restrictions
# ⚠️  This should only be used for testing purposes
# ⚠️  Production usage may violate terms of service
# 
# Proceeding with forced request...
```

### Dry Run Mode

```bash
# Simulate requests without actually making them
yfin scrape --config configs/prod.yaml --ticker AAPL --endpoint key-statistics --dry-run

# Expected output:
# DRY RUN MODE - No actual requests will be made
# 
# Would execute:
# - Ticker: AAPL
# - Endpoint: key-statistics  
# - URL: https://finance.yahoo.com/quote/AAPL/key-statistics
# - Robots.txt: ✓ Allowed
# - Rate limit: ✓ Within limits
# - Fallback strategy: auto
# 
# Configuration validated successfully
```

## CLI Best Practices

### 1. Configuration Management

```bash
# Use environment-specific configs
yfin scrape --config configs/dev.yaml     # Development
yfin scrape --config configs/staging.yaml # Staging  
yfin scrape --config configs/prod.yaml    # Production

# Validate config before use
yfin config --file configs/prod.yaml --validate
```

### 2. Rate Limiting

```bash
# Conservative production settings
yfin scrape --qps 1.0 --concurrency 4 --timeout 45s

# Development/testing settings  
yfin scrape --qps 5.0 --concurrency 8 --timeout 30s
```

### 3. Error Recovery

```bash
# Enable retries and fallback
yfin scrape --retry-max 3 --fallback auto --timeout 60s

# Log errors for debugging
yfin scrape --log-level debug --ticker AAPL --endpoint key-statistics 2>&1 | tee scrape.log
```

### 4. Monitoring and Observability

```bash
# Enable metrics collection
yfin scrape --metrics-enabled --ticker AAPL --endpoint key-statistics

# Track request IDs for debugging
yfin scrape --run-id "batch-$(date +%s)" --ticker AAPL --endpoint key-statistics
```

## Common CLI Patterns

### Data Pipeline Integration

```bash
#!/bin/bash
# daily-scrape.sh - Daily data collection script

TICKERS="AAPL MSFT GOOGL TSLA AMZN"
DATE=$(date +%Y%m%d)
RUN_ID="daily-scrape-$DATE"

for ticker in $TICKERS; do
  echo "Processing $ticker..."
  
  yfin scrape \
    --config configs/prod.yaml \
    --ticker "$ticker" \
    --endpoints key-statistics,financials,news \
    --fallback auto \
    --run-id "$RUN_ID" \
    --publish \
    --env prod \
    --topic-prefix ampy.prod
    
  if [ $? -eq 0 ]; then
    echo "✓ $ticker completed successfully"
  else
    echo "✗ $ticker failed"
  fi
  
  # Rate limiting
  sleep 2
done
```

### Health Check Script

```bash
#!/bin/bash
# health-check.sh - Endpoint health monitoring

ENDPOINTS="key-statistics financials analysis profile news"
TICKER="AAPL"  # Use stable ticker for health checks

for endpoint in $ENDPOINTS; do
  echo "Checking $endpoint..."
  
  yfin scrape \
    --config configs/prod.yaml \
    --ticker "$TICKER" \
    --endpoint "$endpoint" \
    --check \
    --timeout 30s
    
  if [ $? -eq 0 ]; then
    echo "✓ $endpoint is healthy"
  else
    echo "✗ $endpoint has issues"
  fi
done
```

### Batch Processing Script

```bash
#!/bin/bash
# batch-process.sh - Process universe file with error handling

UNIVERSE_FILE="$1"
ENDPOINT="$2"
CONFIG="${3:-configs/prod.yaml}"

if [ ! -f "$UNIVERSE_FILE" ]; then
  echo "Error: Universe file not found: $UNIVERSE_FILE"
  exit 1
fi

TOTAL=$(wc -l < "$UNIVERSE_FILE")
CURRENT=0
FAILED=0

while IFS= read -r ticker; do
  CURRENT=$((CURRENT + 1))
  echo "[$CURRENT/$TOTAL] Processing $ticker..."
  
  yfin scrape \
    --config "$CONFIG" \
    --ticker "$ticker" \
    --endpoint "$ENDPOINT" \
    --fallback auto \
    --timeout 60s \
    --preview
    
  if [ $? -ne 0 ]; then
    FAILED=$((FAILED + 1))
    echo "✗ Failed: $ticker"
  else
    echo "✓ Success: $ticker"
  fi
  
  # Rate limiting
  sleep 1
done < "$UNIVERSE_FILE"

echo "Batch complete: $((TOTAL - FAILED))/$TOTAL successful"
```

This CLI guide provides comprehensive coverage of all scraping functionality. For configuration details and troubleshooting, see the related documentation sections.
