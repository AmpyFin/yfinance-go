# Troubleshooting Guide

## Overview

This guide provides comprehensive troubleshooting information for the scrape fallback system, including common errors, their causes, and detailed remediation steps.

## Error Categories

### 1. Parse Errors
### 2. Network and Rate Limiting Errors  
### 3. Robots.txt and Compliance Errors
### 4. Configuration Errors
### 5. Data Quality and Schema Errors

---

## Parse Errors

### ErrNoQuoteSummary

**Error Message:**
```
failed to parse key statistics: ErrNoQuoteSummary: quote summary section not found in HTML
```

**Cause:** The HTML structure of Yahoo Finance has changed, and the expected quote summary section is missing or has been renamed.

**Symptoms:**
- Consistent failures across multiple tickers
- Error occurs immediately after successful HTTP request
- HTML content is retrieved but parsing fails

**Diagnosis Steps:**
```bash
# 1. Check if the issue is widespread
yfin scrape --ticker AAPL --endpoint key-statistics --check
yfin scrape --ticker MSFT --endpoint key-statistics --check
yfin scrape --ticker GOOGL --endpoint key-statistics --check

# 2. Enable debug logging to see HTML structure
yfin scrape --log-level debug --ticker AAPL --endpoint key-statistics --preview

# 3. Capture HTML for analysis
yfin scrape --ticker AAPL --endpoint key-statistics --save-html debug/aapl-keystat.html
```

**Remediation:**
1. **Immediate Fix (Production):**
   ```bash
   # Switch to API-only mode if available
   yfin scrape --fallback api-only --ticker AAPL --endpoint quote
   
   # Or use alternative endpoints
   yfin scrape --ticker AAPL --endpoint profile --preview  # May have some key stats
   ```

2. **Comprehensive Fix (Development):**
   ```go
   // Update regex patterns in internal/scrape/regex/key_statistics.yaml
   quote_summary:
     selectors:
       - "div[data-testid='quote-summary']"  # New selector
       - "div[class*='quote-summary']"       # Fallback pattern
       - "section[data-module='QuoteSummary']"  # Alternative
   ```

3. **Testing Fix:**
   ```bash
   # Test with multiple tickers
   yfin scrape --ticker AAPL --endpoint key-statistics --preview
   yfin scrape --ticker MSFT --endpoint key-statistics --preview
   
   # Run soak test to validate
   yfin soak --duration 10m --endpoints key-statistics --concurrency 2
   ```

### ErrJSONUnescape

**Error Message:**
```
failed to parse financials: ErrJSONUnescape: invalid JSON escape sequence in embedded data
```

**Cause:** Yahoo Finance embedded JSON contains malformed escape sequences or encoding issues.

**Symptoms:**
- Intermittent failures (not all tickers affected)
- Error occurs during JSON parsing phase
- Some tickers work, others fail consistently

**Diagnosis Steps:**
```bash
# 1. Test multiple tickers to identify pattern
yfin scrape --ticker AAPL --endpoint financials --preview  # Works
yfin scrape --ticker "BRK-A" --endpoint financials --preview  # Fails (special chars)
yfin scrape --ticker "0700.HK" --endpoint financials --preview  # Fails (international)

# 2. Check specific ticker patterns
yfin scrape --ticker "BRK.A" --endpoint financials --preview  # Try normalized symbol
```

**Remediation:**
1. **Immediate Fix:**
   ```bash
   # Use symbol normalization
   yfin scrape --ticker "BRK-A" --normalize-symbol --endpoint financials
   
   # Or try alternative endpoints
   yfin scrape --ticker "BRK-A" --endpoint key-statistics --preview
   ```

2. **Comprehensive Fix:**
   ```go
   // Update JSON unescaping in internal/scrape/extract.go
   func unescapeJSON(raw string) (string, error) {
       // Add better handling for malformed escape sequences
       raw = strings.ReplaceAll(raw, `\"`, `"`)
       raw = strings.ReplaceAll(raw, `\\`, `\`)
       
       // Handle special cases
       raw = regexp.MustCompile(`\\u[0-9a-fA-F]{4}`).ReplaceAllStringFunc(raw, func(match string) string {
           // Proper Unicode unescaping
           return unescapeUnicode(match)
       })
       
       return raw, nil
   }
   ```

### ErrSchemaDrift

**Error Message:**
```
schema validation failed: ErrSchemaDrift: field 'marketCap' expected number, got string
```

**Cause:** Yahoo Finance changed the data format for specific fields, breaking our parsing expectations.

**Symptoms:**
- Sudden failures after period of stability
- Affects specific fields across all tickers
- Schema validation fails after successful parsing

**Diagnosis Steps:**
```bash
# 1. Compare current vs. expected schema
yfin scrape --ticker AAPL --endpoint key-statistics --validate-schema --debug

# 2. Check multiple tickers for consistency
yfin scrape --ticker AAPL --endpoint key-statistics --schema-check
yfin scrape --ticker MSFT --endpoint key-statistics --schema-check

# 3. Capture sample data for analysis
yfin scrape --ticker AAPL --endpoint key-statistics --save-parsed debug/aapl-parsed.json
```

**Remediation:**
1. **Immediate Fix:**
   ```bash
   # Enable lenient parsing mode
   yfin scrape --strict-mode=false --ticker AAPL --endpoint key-statistics
   
   # Use field mapping overrides
   yfin scrape --field-mapping "marketCap:string->number" --ticker AAPL --endpoint key-statistics
   ```

2. **Comprehensive Fix:**
   ```go
   // Update field mappings in internal/emit/map_financials.go
   func mapMarketCap(raw interface{}) (*fundamentalsv1.ScaledDecimal, error) {
       switch v := raw.(type) {
       case string:
           // Handle new string format
           return parseMarketCapString(v)
       case float64:
           // Handle legacy number format  
           return parseMarketCapNumber(v)
       case map[string]interface{}:
           // Handle new object format
           return parseMarketCapObject(v)
       default:
           return nil, fmt.Errorf("unexpected marketCap type: %T", v)
       }
   }
   ```

---

## Network and Rate Limiting Errors

### High 429/503 Rates

**Error Message:**
```
HTTP 429: Too Many Requests - rate limit exceeded
HTTP 503: Service Temporarily Unavailable
```

**Cause:** Exceeding Yahoo Finance's rate limits or server capacity issues.

**Symptoms:**
- Sustained high error rates (>5%)
- Errors increase with higher QPS settings
- Temporary recovery followed by more errors

**Diagnosis Steps:**
```bash
# 1. Check current rate limit settings
yfin config --print-effective | grep -A 5 rate_limit

# 2. Test with lower QPS
yfin scrape --qps 0.5 --ticker AAPL --endpoint key-statistics

# 3. Monitor error rates over time
yfin soak --duration 10m --qps 1.0 --concurrency 2 --endpoints key-statistics
```

**Remediation:**
1. **Immediate Fix:**
   ```bash
   # Reduce QPS and concurrency
   yfin scrape --qps 1.0 --concurrency 4 --ticker AAPL --endpoint key-statistics
   
   # Enable session rotation
   yfin scrape --sessions 8 --ticker AAPL --endpoint key-statistics
   
   # Increase backoff times
   yfin scrape --retry-backoff-base 2s --retry-backoff-max 30s --ticker AAPL
   ```

2. **Configuration Fix:**
   ```yaml
   # Update config/production.yaml
   scrape:
     qps: 1.0  # Reduced from 2.0
     burst: 3  # Reduced from 5
     per_host_qps: 0.5  # Add per-host limit
     
   session_rotation:
     enabled: true
     pool_size: 10  # Increased from 5
     rotation_interval_minutes: 15  # More frequent rotation
     
   retry:
     backoff_base_ms: 2000  # Increased from 1000
     backoff_max_ms: 30000  # Increased from 10000
   ```

3. **Monitoring Setup:**
   ```bash
   # Set up alerts for high error rates
   # Alert if 429/503 rate > 5% for 5 minutes
   # Critical if > 10% for 2 minutes
   ```

### Session Rotation Issues

**Error Message:**
```
session rotation failed: all sessions exhausted or rate limited
```

**Cause:** All sessions in the rotation pool are rate-limited or blocked.

**Diagnosis Steps:**
```bash
# 1. Check session pool status
yfin scrape --session-status --debug

# 2. Test individual sessions
yfin scrape --session-id 0 --ticker AAPL --endpoint key-statistics
yfin scrape --session-id 1 --ticker AAPL --endpoint key-statistics

# 3. Monitor session health
yfin scrape --session-health-check
```

**Remediation:**
1. **Immediate Fix:**
   ```bash
   # Increase session pool size
   yfin scrape --sessions 15 --ticker AAPL --endpoint key-statistics
   
   # Reset session pool
   yfin scrape --reset-sessions --ticker AAPL --endpoint key-statistics
   ```

2. **Configuration Fix:**
   ```yaml
   session_rotation:
     enabled: true
     pool_size: 20  # Increased pool size
     rotation_interval_minutes: 10  # More frequent rotation
     session_timeout_minutes: 60  # Longer session timeout
     health_check_interval_minutes: 5  # Regular health checks
   ```

---

## Robots.txt and Compliance Errors

### Robots Denied

**Error Message:**
```
robots.txt denied: path /quote/AAPL/key-statistics disallowed for user-agent yfinance-go
```

**Cause:** Yahoo Finance's robots.txt file has been updated to disallow access to specific paths.

**Symptoms:**
- Sudden failures for previously working endpoints
- Affects all tickers for specific endpoints
- Error occurs before HTTP request is made

**Diagnosis Steps:**
```bash
# 1. Check current robots.txt
curl -s https://finance.yahoo.com/robots.txt | grep -A 10 -B 10 "key-statistics"

# 2. Test different endpoints
yfin scrape --ticker AAPL --endpoint profile --check  # May still be allowed
yfin scrape --ticker AAPL --endpoint news --check     # Check news access

# 3. Verify user-agent handling
yfin scrape --user-agent "Mozilla/5.0..." --ticker AAPL --endpoint key-statistics --check
```

**Remediation:**
1. **Immediate Fix (Production):**
   ```bash
   # Switch to allowed endpoints
   yfin scrape --ticker AAPL --endpoint profile --preview  # If allowed
   
   # Use API fallback if available
   yfin scrape --fallback api-only --ticker AAPL --endpoint quote
   ```

2. **Configuration Fix:**
   ```yaml
   # Update user-agent to be more generic
   scrape:
     user_agent: "Mozilla/5.0 (compatible; yfinance-go/1.0; +https://github.com/AmpyFin/yfinance-go)"
     
   # Or use robots policy override for testing only
   scrape:
     robots_policy: "warn"  # Development only, never in production
   ```

3. **Alternative Approach:**
   ```bash
   # Find alternative data sources
   yfin scrape --ticker AAPL --endpoint profile --preview  # May contain key stats
   yfin scrape --ticker AAPL --endpoint financials --preview  # Alternative metrics
   ```

**CRITICAL WARNING:** Never use `--force` or `robots_policy: "ignore"` in production. This violates terms of service and can lead to IP blocking.

---

## Configuration Errors

### Invalid Configuration Values

**Error Message:**
```
configuration validation failed: qps value 15.0 exceeds maximum allowed value of 10.0
```

**Cause:** Configuration values outside acceptable ranges.

**Diagnosis Steps:**
```bash
# 1. Validate configuration
yfin config --file configs/prod.yaml --validate

# 2. Check effective configuration
yfin config --file configs/prod.yaml --print-effective

# 3. Test with default values
yfin scrape --ticker AAPL --endpoint key-statistics  # Uses defaults
```

**Remediation:**
```yaml
# Fix configuration values
scrape:
  qps: 5.0  # Within acceptable range (0.1-10.0)
  timeout_ms: 30000  # Within range (1000-300000)
  retry:
    attempts: 3  # Within range (1-10)
```

### Missing Required Configuration

**Error Message:**
```
configuration error: scrape.enabled is required but not specified
```

**Remediation:**
```yaml
# Ensure all required fields are present
scrape:
  enabled: true  # Required
  robots_policy: "enforce"  # Required in production
```

---

## Data Quality and Schema Errors

### News Deduplication Issues

**Error Message:**
```
news deduplication failed: duplicate article detected with same URL but different timestamps
```

**Cause:** Yahoo Finance serves the same article with updated timestamps, causing deduplication conflicts.

**Diagnosis Steps:**
```bash
# 1. Check news articles for duplicates
yfin scrape --ticker AAPL --endpoint news --preview-news --debug

# 2. Test deduplication settings
yfin scrape --ticker AAPL --endpoint news --dedupe-strategy url --preview
yfin scrape --ticker AAPL --endpoint news --dedupe-strategy title --preview
```

**Remediation:**
1. **Configuration Fix:**
   ```yaml
   scrape:
     parsing:
       news_deduplication:
         strategy: "url_and_title"  # More robust deduplication
         time_window_hours: 24      # Consider articles within 24h as potential duplicates
         similarity_threshold: 0.8   # 80% title similarity for duplicates
   ```

2. **Code Fix:**
   ```go
   // Update deduplication logic in internal/scrape/extract_news.go
   func deduplicateNews(articles []NewsArticle) []NewsArticle {
       seen := make(map[string]bool)
       var result []NewsArticle
       
       for _, article := range articles {
           // Create composite key for better deduplication
           key := fmt.Sprintf("%s|%s", normalizeURL(article.URL), normalizeTitle(article.Title))
           
           if !seen[key] {
               seen[key] = true
               result = append(result, article)
           }
       }
       
       return result
   }
   ```

### Time Normalization Caveats

**Error Message:**
```
time parsing failed: unable to parse date "Q3 2023" into standard format
```

**Cause:** Yahoo Finance uses various date formats that don't map to standard timestamps.

**Diagnosis Steps:**
```bash
# 1. Check date formats in use
yfin scrape --ticker AAPL --endpoint financials --debug | grep "date format"

# 2. Test with different date parsing modes
yfin scrape --ticker AAPL --endpoint financials --date-parsing lenient --preview
```

**Remediation:**
1. **Configuration Fix:**
   ```yaml
   scrape:
     parsing:
       date_formats:
         - "2006-01-02"
         - "Jan 2, 2006"
         - "2006-01-02T15:04:05Z"
         - "Q1 2006"  # Add quarterly format
         - "2006"     # Add yearly format
       date_parsing_mode: "lenient"  # Allow partial date parsing
   ```

2. **Code Fix:**
   ```go
   // Update date parsing in internal/scrape/time.go
   func parseFlexibleDate(dateStr string) (time.Time, error) {
       // Handle quarterly dates
       if matched := quarterlyRegex.FindStringSubmatch(dateStr); matched != nil {
           return parseQuarterlyDate(matched[1], matched[2])
       }
       
       // Handle yearly dates
       if matched := yearlyRegex.FindStringSubmatch(dateStr); matched != nil {
           return parseYearlyDate(matched[1])
       }
       
       // Standard date parsing
       return parseStandardDate(dateStr)
   }
   ```

---

## Comprehensive Debugging Workflow

### Step 1: Initial Diagnosis

```bash
# 1. Test basic connectivity
yfin scrape --ticker AAPL --endpoint key-statistics --check

# 2. Enable debug logging
yfin scrape --log-level debug --ticker AAPL --endpoint key-statistics --preview

# 3. Check configuration
yfin config --print-effective | grep -A 20 scrape
```

### Step 2: Isolate the Issue

```bash
# Test different components
yfin scrape --fallback api-only --ticker AAPL --endpoint quote      # Test API path
yfin scrape --fallback scrape-only --ticker AAPL --endpoint profile # Test scrape path
yfin scrape --dry-run --ticker AAPL --endpoint key-statistics       # Test config only
```

### Step 3: Gather Evidence

```bash
# Capture debugging information
yfin scrape --ticker AAPL --endpoint key-statistics --save-html debug/
yfin scrape --ticker AAPL --endpoint key-statistics --save-parsed debug/
yfin scrape --ticker AAPL --endpoint key-statistics --trace-requests debug/
```

### Step 4: Test Fix

```bash
# Test fix with single ticker
yfin scrape --ticker AAPL --endpoint key-statistics --preview

# Test with multiple tickers
yfin scrape --universe-file test-universe.txt --endpoint key-statistics --preview

# Run soak test to validate
yfin soak --duration 5m --endpoints key-statistics --concurrency 2
```

## Prevention Best Practices

### 1. Monitoring Setup

```bash
# Set up comprehensive monitoring
# - Parse success rate > 95%
# - Rate limit errors < 1%
# - Response time < 3s P95
# - Schema validation success > 99%
```

### 2. Configuration Management

```yaml
# Use environment-specific configurations
# - Conservative production settings
# - Aggressive development settings
# - Comprehensive logging in staging
```

### 3. Regular Health Checks

```bash
#!/bin/bash
# health-check.sh - Run daily
yfin scrape --ticker AAPL --endpoint key-statistics --check
yfin scrape --ticker MSFT --endpoint financials --check
yfin scrape --ticker GOOGL --endpoint news --check
```

### 4. Schema Evolution Monitoring

```bash
# Monitor for schema changes
yfin scrape --schema-diff --ticker AAPL --endpoint key-statistics --baseline schema/baseline.json
```

This troubleshooting guide provides comprehensive coverage of common issues and their solutions. For additional support, capture debug information and provide detailed error logs when reporting issues.
