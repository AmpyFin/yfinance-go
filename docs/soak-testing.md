# Soak Testing Guide

This document describes the comprehensive soak testing functionality implemented in yfinance-go for validating stability, throughput, and robustness under production-like conditions.

## Overview

The soak testing system provides end-to-end validation of the entire data pipeline including:

- **API → Fallback → Mapping → Publishing** pipeline robustness
- **Rate limiting and backoff** behavior under sustained load
- **Session rotation and robots.txt compliance**
- **Memory and goroutine leak detection**
- **API vs scrape correctness validation**
- **Failure injection and recovery testing**

## Quick Start

### Basic Smoke Test (10 minutes)
```bash
./bin/yfin soak \
  --config configs/example.dev.yaml \
  --universe-file testdata/universe/soak.txt \
  --endpoints key-statistics,financials,analysis,profile,news \
  --fallback auto \
  --duration 10m \
  --concurrency 8 \
  --qps 5 \
  --preview
```

### Full Production Soak Test (2 hours)
```bash
./bin/yfin soak \
  --config configs/example.prod.yaml \
  --universe-file testdata/universe/soak.txt \
  --endpoints key-statistics,financials,analysis,profile,news \
  --fallback auto \
  --duration 2h \
  --concurrency 12 \
  --qps 5 \
  --preview
```

### With Publishing Enabled
```bash
./bin/yfin soak \
  --config configs/example.dev.yaml \
  --universe-file testdata/universe/soak.txt \
  --endpoints news \
  --fallback auto \
  --duration 30m \
  --concurrency 8 \
  --qps 3 \
  --publish \
  --env dev \
  --topic-prefix ampy.dev
```

## Configuration Options

### Core Parameters

| Flag | Default | Description |
|------|---------|-------------|
| `--universe-file` | *required* | File containing list of tickers to test |
| `--endpoints` | `key-statistics,financials,analysis,profile,news` | Comma-separated list of endpoints |
| `--fallback` | `auto` | Fallback strategy: `auto`, `api-only`, `scrape-only` |
| `--duration` | `2h` | Duration to run soak test |
| `--concurrency` | `12` | Number of concurrent workers |
| `--qps` | `5.0` | Target queries per second |

### Observability & Testing

| Flag | Default | Description |
|------|---------|-------------|
| `--preview` | `false` | Enable preview mode (no actual publishing) |
| `--memory-check` | `true` | Enable memory and goroutine leak detection |
| `--probe-interval` | `1h` | Interval for correctness probes |
| `--failure-rate` | `0.1` | Simulated failure rate (0.0-1.0) |

### Publishing

| Flag | Default | Description |
|------|---------|-------------|
| `--publish` | `false` | Enable publishing to bus |
| `--env` | `dev` | Environment for publishing |
| `--topic-prefix` | `ampy.dev` | Topic prefix for publishing |

## Ticker Universe

The soak test uses a diverse ticker universe defined in `testdata/universe/soak.txt`:

- **US Large Cap**: AAPL, MSFT, GOOGL, AMZN, META, etc.
- **US Mid/Small Cap**: ROKU, PLTR, RBLX, COIN, SNOW, etc.
- **International**: ASML, SAP, 005930.KS, 0700.HK, etc.
- **ADRs**: NIO, LI, XPEV, PDD, JD, etc.
- **Specialized**: REITs, Crypto-related, Healthcare, Financial Services

Total: 64+ diverse tickers across markets, currencies, and sectors.

## Supported Endpoints

### API Endpoints (with fallback to scraping)
- `quote` - Real-time quote data
- `daily-bars` - Historical daily price data
- `fundamentals` - Quarterly fundamentals (requires subscription)

### Scrape-Only Endpoints
- `key-statistics` - Key financial metrics
- `financials` - Income statement data
- `analysis` - Analyst recommendations
- `profile` - Company profile information
- `news` - News articles
- `balance-sheet` - Balance sheet data
- `cash-flow` - Cash flow statement
- `analyst-insights` - Analyst insights

## Fallback Strategies

### Auto Fallback (`--fallback auto`)
- Attempts API first for supported endpoints
- Falls back to scraping on:
  - 401 (Authentication required)
  - 429 (Rate limiting)
  - 5xx (Server errors)
- Uses scraping directly for scrape-only endpoints

### API Only (`--fallback api-only`)
- Only uses API endpoints
- Fails if API is unavailable
- Useful for testing API reliability

### Scrape Only (`--fallback scrape-only`)
- Only uses web scraping
- Tests scraping infrastructure in isolation
- Validates robots.txt compliance

## Observability Features

### Real-Time Metrics
- Request success/failure rates
- Latency histograms per endpoint
- Fallback decision tracking
- Rate limit and robots.txt blocks
- Memory usage and goroutine counts

### Correctness Probes
Periodic validation comparing API vs scrape data:
- **Market Cap** comparison (5% tolerance)
- **P/E Ratio** validation (10% tolerance)
- **Employee Count** verification (15% tolerance)
- **Sector** information consistency
- **Currency** code validation

### Memory Leak Detection
- Continuous memory usage monitoring
- Goroutine count tracking
- Growth rate analysis
- Leak detection with recommendations
- GC behavior analysis

### Failure Injection
Built-in failure server simulating:
- Rate limiting (429 responses)
- Server errors (500, 502, 503)
- Authentication failures (401)
- Connection timeouts
- Bad gateway responses

## Output Analysis

### Success Criteria
✅ **Zero Memory Leaks**: Heap/GC steady within ±10% after warmup  
✅ **Rate Limit Safety**: 429/503 rates < 1% sustained  
✅ **Robots Adherence**: No disallowed fetches in enforce mode  
✅ **Correctness**: API vs scrape parity within tolerance  
✅ **No DLQ**: Clean publishing with proper ordering  

### Sample Output
```
=== SOAK TEST RESULTS ===
Duration: 2h0m15s
Total Requests: 7,234
Successful Requests: 7,198
Failed Requests: 36
Success Rate: 99.50%
Actual QPS: 1.00
API Requests: 2,156
Scrape Requests: 5,078
Fallback Decisions: 234
Rate Limit Hits: 12
Robots Blocked: 0
Correctness Probes Passed: 24
Correctness Probes Failed: 1

=== MEMORY ANALYSIS ===
Initial Memory: 2,313 KB
Peak Memory: 8,456 KB
Final Memory: 3,127 KB
Memory Growth: 814 KB
Initial Goroutines: 8
Peak Goroutines: 24
Final Goroutines: 12
Goroutine Growth: 4

=== ENDPOINT BREAKDOWN ===
key-statistics: 1,445 requests, 1,440 successes, 5 failures, avg latency: 1.2s
financials: 1,434 requests, 1,429 successes, 5 failures, avg latency: 1.8s
analysis: 1,445 requests, 1,441 successes, 4 failures, avg latency: 1.1s
profile: 1,456 requests, 1,444 successes, 12 failures, avg latency: 2.1s
news: 1,454 requests, 1,444 successes, 10 failures, avg latency: 1.5s
```

## Troubleshooting

### High Failure Rate
- Check network connectivity
- Verify robots.txt compliance
- Reduce QPS if rate limited
- Check Yahoo Finance service status

### Memory Leaks
- Review goroutine lifecycle management
- Check for unclosed channels
- Verify proper context cancellation
- Monitor object retention patterns

### Correctness Probe Failures
- Expected due to timing differences
- Currency conversion variations
- Reporting period mismatches
- Data source update delays

### Rate Limiting
- Reduce `--qps` parameter
- Increase `--concurrency` for better distribution
- Enable session rotation in config
- Check backoff behavior in logs

## Integration with CI/CD

### Nightly Soak Tests
```bash
# Short validation (30 minutes)
./bin/yfin soak \
  --config configs/ci.yaml \
  --universe-file testdata/universe/soak.txt \
  --endpoints key-statistics,news \
  --duration 30m \
  --concurrency 4 \
  --qps 2 \
  --preview
```

### Pre-Release Validation
```bash
# Comprehensive test (1 hour)
./bin/yfin soak \
  --config configs/staging.yaml \
  --universe-file testdata/universe/soak.txt \
  --endpoints key-statistics,financials,analysis,profile,news \
  --duration 1h \
  --concurrency 8 \
  --qps 3 \
  --preview \
  --memory-check
```

## Best Practices

1. **Start Small**: Begin with short durations and low QPS
2. **Monitor Resources**: Watch memory and goroutine growth
3. **Validate Correctness**: Review probe results regularly
4. **Respect Rate Limits**: Keep QPS reasonable (< 10)
5. **Use Preview Mode**: Test without publishing first
6. **Check Logs**: Monitor for robots.txt violations
7. **Gradual Scaling**: Increase load incrementally

## Architecture

The soak testing system consists of:

- **Orchestrator**: Manages test execution and coordination
- **Workers**: Execute randomized requests across endpoints
- **Metrics**: Prometheus-compatible metrics collection
- **Probes**: Correctness validation between data sources
- **Memory Monitor**: Leak detection and analysis
- **Failure Server**: Simulated failure injection
- **Rate Limiter**: QPS control and backoff management

This comprehensive system ensures production readiness and validates the entire data pipeline under realistic load conditions.
