# Scrape Configuration Guide

## Configuration Overview

The scrape fallback system provides extensive configuration options to control behavior, performance, and safety. Configuration can be set through YAML files, environment variables, or CLI flags with a clear precedence hierarchy.

## Configuration Precedence

```
CLI Flags > Environment Variables > Config File > Defaults
```

## Core Configuration Structure

### Complete Configuration Example

```yaml
# config/production.yaml
scrape:
  # Basic settings
  enabled: true
  user_agent: "yfinance-go/1.0 (+https://github.com/AmpyFin/yfinance-go)"
  timeout_ms: 30000
  
  # Rate limiting
  qps: 2.0
  burst: 5
  per_host_qps: 1.0
  
  # Robots.txt compliance
  robots_policy: "enforce"  # enforce, warn, ignore
  robots_cache_ttl_hours: 24
  
  # Retry configuration
  retry:
    attempts: 3
    backoff_base_ms: 1000
    backoff_max_ms: 10000
    backoff_multiplier: 2.0
    jitter: true
  
  # Session management
  session_rotation:
    enabled: true
    pool_size: 5
    rotation_interval_minutes: 30
  
  # Fallback behavior
  fallback:
    strategy: "auto"  # auto, api-only, scrape-only
    triggers:
      - "401"  # Authentication required
      - "403"  # Forbidden
      - "429"  # Rate limited
      - "500"  # Internal server error
      - "502"  # Bad gateway
      - "503"  # Service unavailable
    retry_before_fallback: 2
    timeout_ms: 30000
  
  # Parsing configuration
  parsing:
    strict_mode: true
    ignore_missing_fields: false
    max_news_articles: 50
    date_formats:
      - "2006-01-02"
      - "Jan 2, 2006"
      - "2006-01-02T15:04:05Z"
  
  # Caching
  cache:
    enabled: true
    ttl_minutes: 15
    max_entries: 1000
  
  # Observability
  observability:
    log_requests: true
    log_responses: false  # Only for debugging
    metrics_enabled: true
    trace_sampling_rate: 0.1

# HTTP client configuration
http:
  timeout_ms: 30000
  keep_alive_ms: 30000
  max_idle_conns: 100
  max_idle_conns_per_host: 10
  
# Circuit breaker
circuit_breaker:
  enabled: true
  failure_threshold: 5
  success_threshold: 3
  timeout_ms: 60000
  max_requests: 10
```

## Configuration Sections

### 1. Basic Settings

#### `scrape.enabled`
- **Type**: `boolean`
- **Default**: `true`
- **Description**: Enable/disable scraping functionality
- **Example**:
  ```yaml
  scrape:
    enabled: true
  ```

#### `scrape.user_agent`
- **Type**: `string`
- **Default**: `"yfinance-go/1.0 (+https://github.com/AmpyFin/yfinance-go)"`
- **Description**: User-Agent header for HTTP requests
- **Best Practice**: Include contact information for responsible scraping
- **Example**:
  ```yaml
  scrape:
    user_agent: "MyApp/2.0 (contact@mycompany.com)"
  ```

#### `scrape.timeout_ms`
- **Type**: `integer`
- **Default**: `30000` (30 seconds)
- **Description**: HTTP request timeout in milliseconds
- **Range**: `1000` - `300000` (1 second to 5 minutes)
- **Example**:
  ```yaml
  scrape:
    timeout_ms: 45000  # 45 seconds for slow connections
  ```

### 2. Rate Limiting Configuration

#### `scrape.qps`
- **Type**: `float`
- **Default**: `2.0`
- **Description**: Global queries per second limit
- **Range**: `0.1` - `10.0`
- **Production Recommendation**: `1.0` - `3.0`
- **Example**:
  ```yaml
  scrape:
    qps: 1.5  # Conservative rate for production
  ```

#### `scrape.burst`
- **Type**: `integer`
- **Default**: `5`
- **Description**: Maximum burst requests allowed
- **Range**: `1` - `20`
- **Example**:
  ```yaml
  scrape:
    burst: 3  # Allow small bursts
  ```

#### `scrape.per_host_qps`
- **Type**: `float`
- **Default**: `1.0`
- **Description**: Per-host rate limit (overrides global for specific hosts)
- **Example**:
  ```yaml
  scrape:
    per_host_qps: 0.5  # Very conservative for Yahoo Finance
  ```

### 3. Robots.txt Compliance

#### `scrape.robots_policy`
- **Type**: `string`
- **Default**: `"enforce"`
- **Options**:
  - `"enforce"`: Strictly follow robots.txt (Production)
  - `"warn"`: Log violations but proceed (Development)
  - `"ignore"`: Skip robots.txt checks (Testing only)
- **Example**:
  ```yaml
  scrape:
    robots_policy: "enforce"  # Production setting
  ```

#### `scrape.robots_cache_ttl_hours`
- **Type**: `integer`
- **Default**: `24`
- **Description**: How long to cache robots.txt files
- **Range**: `1` - `168` (1 hour to 1 week)
- **Example**:
  ```yaml
  scrape:
    robots_cache_ttl_hours: 12  # Refresh twice daily
  ```

### 4. Retry Configuration

#### Complete Retry Example
```yaml
scrape:
  retry:
    attempts: 3              # Maximum retry attempts
    backoff_base_ms: 1000    # Base backoff time (1 second)
    backoff_max_ms: 10000    # Maximum backoff time (10 seconds)
    backoff_multiplier: 2.0  # Exponential backoff multiplier
    jitter: true             # Add random jitter to prevent thundering herd
```

#### Retry Strategy Explanation
```
Attempt 1: Immediate
Attempt 2: 1000ms + jitter
Attempt 3: 2000ms + jitter
Attempt 4: 4000ms + jitter (capped at backoff_max_ms)
```

### 5. Session Management

#### `scrape.session_rotation.enabled`
- **Type**: `boolean`
- **Default**: `true`
- **Description**: Enable session rotation to avoid rate limiting
- **Production Recommendation**: Always `true`

#### `scrape.session_rotation.pool_size`
- **Type**: `integer`
- **Default**: `5`
- **Description**: Number of sessions in rotation pool
- **Range**: `2` - `20`
- **Example**:
  ```yaml
  scrape:
    session_rotation:
      enabled: true
      pool_size: 8  # Larger pool for high-volume usage
      rotation_interval_minutes: 15
  ```

### 6. Fallback Configuration

#### `scrape.fallback.strategy`
- **Type**: `string`
- **Default**: `"auto"`
- **Options**:
  - `"auto"`: Try API first, fallback to scraping
  - `"api-only"`: Only use API endpoints
  - `"scrape-only"`: Only use web scraping
- **Example**:
  ```yaml
  scrape:
    fallback:
      strategy: "auto"
      retry_before_fallback: 1  # Quick fallback
  ```

#### `scrape.fallback.triggers`
- **Type**: `array of strings`
- **Default**: `["401", "403", "429", "5xx"]`
- **Description**: HTTP status codes that trigger fallback
- **Example**:
  ```yaml
  scrape:
    fallback:
      triggers:
        - "401"  # Authentication required
        - "429"  # Rate limited
        - "503"  # Service unavailable
  ```

### 7. Parsing Configuration

#### `scrape.parsing.strict_mode`
- **Type**: `boolean`
- **Default**: `true`
- **Description**: Fail on any parsing errors vs. best-effort parsing
- **Production Recommendation**: `true` for data quality

#### `scrape.parsing.max_news_articles`
- **Type**: `integer`
- **Default**: `50`
- **Description**: Maximum news articles to parse per request
- **Range**: `1` - `200`

#### `scrape.parsing.date_formats`
- **Type**: `array of strings`
- **Description**: Supported date formats for parsing
- **Example**:
  ```yaml
  scrape:
    parsing:
      date_formats:
        - "2006-01-02"
        - "Jan 2, 2006"
        - "2006-01-02T15:04:05Z"
        - "Mon, 02 Jan 2006 15:04:05 MST"
  ```

## Environment Variable Overrides

All configuration options can be overridden with environment variables using the pattern:
`YFIN_<SECTION>_<KEY>=value`

### Examples

```bash
# Basic settings
export YFIN_SCRAPE_ENABLED=true
export YFIN_SCRAPE_QPS=1.5
export YFIN_SCRAPE_TIMEOUT_MS=45000

# Robots policy
export YFIN_SCRAPE_ROBOTS_POLICY=enforce

# Retry configuration
export YFIN_SCRAPE_RETRY_ATTEMPTS=2
export YFIN_SCRAPE_RETRY_BACKOFF_BASE_MS=2000

# Fallback strategy
export YFIN_SCRAPE_FALLBACK_STRATEGY=auto
```

## CLI Flag Overrides

CLI flags have the highest precedence and override all other configuration:

```bash
# Override QPS and timeout
yfin scrape --qps 0.5 --timeout 60s --ticker AAPL --endpoint key-statistics

# Override robots policy (testing only)
yfin scrape --robots-policy ignore --ticker AAPL --endpoint profile --force

# Override fallback strategy
yfin scrape --fallback scrape-only --ticker AAPL --endpoint news
```

## Configuration Validation

### Required Fields
- `scrape.enabled` must be boolean
- `scrape.qps` must be positive number
- `scrape.robots_policy` must be valid option

### Range Validation
```yaml
# These will be validated at startup
scrape:
  qps: 15.0  # ERROR: exceeds maximum of 10.0
  timeout_ms: 500  # ERROR: below minimum of 1000
  retry:
    attempts: 0  # ERROR: must be at least 1
```

### Configuration Testing

```bash
# Validate configuration file
yfin config --file config/production.yaml --validate

# Print effective configuration
yfin config --file config/production.yaml --print-effective

# Test configuration with dry run
yfin scrape --config config/production.yaml --ticker AAPL --endpoint key-statistics --dry-run
```

## Production Configuration Examples

### High-Volume Production
```yaml
scrape:
  qps: 3.0
  burst: 10
  timeout_ms: 45000
  robots_policy: "enforce"
  session_rotation:
    enabled: true
    pool_size: 10
  retry:
    attempts: 2
    backoff_base_ms: 2000
  fallback:
    strategy: "auto"
    retry_before_fallback: 1
```

### Conservative Production
```yaml
scrape:
  qps: 1.0
  burst: 3
  timeout_ms: 30000
  robots_policy: "enforce"
  session_rotation:
    enabled: true
    pool_size: 5
  retry:
    attempts: 3
    backoff_base_ms: 1000
  fallback:
    strategy: "auto"
    retry_before_fallback: 2
```

### Development/Testing
```yaml
scrape:
  qps: 5.0
  burst: 15
  timeout_ms: 60000
  robots_policy: "warn"
  session_rotation:
    enabled: false
  retry:
    attempts: 1
    backoff_base_ms: 500
  fallback:
    strategy: "auto"
  parsing:
    strict_mode: false
```

## Configuration Best Practices

### 1. Production Safety
- Always use `robots_policy: "enforce"`
- Keep QPS conservative (1.0-3.0)
- Enable session rotation
- Use reasonable timeouts (30-45 seconds)

### 2. Performance Optimization
- Tune burst settings for traffic patterns
- Adjust retry attempts based on error rates
- Use caching for repeated requests
- Monitor and adjust based on metrics

### 3. Error Handling
- Enable strict parsing in production
- Configure appropriate fallback triggers
- Set reasonable retry limits
- Log configuration changes

### 4. Monitoring
- Enable metrics collection
- Use appropriate trace sampling
- Log request patterns
- Monitor configuration drift

This configuration guide provides comprehensive control over the scrape fallback system behavior. For usage examples and troubleshooting, see the related documentation sections.
