# Scrape Fallback Operator Runbook

## Overview

This runbook provides step-by-step procedures for managing the scrape fallback system, including forcing API or scrape modes, handling authentication errors, managing robots.txt decisions, and safe use of override flags.

## Quick Reference

### Emergency Commands
```bash
# Force API-only mode (bypass scraping)
yfin scrape --fallback api-only --ticker AAPL --endpoint quote

# Force scrape-only mode (bypass API)  
yfin scrape --fallback scrape-only --ticker AAPL --endpoint key-statistics

# Check system health
yfin scrape --ticker AAPL --endpoint key-statistics --check

# Override robots.txt (TESTING ONLY)
yfin scrape --force --ticker AAPL --endpoint profile --preview
```

### Key Metrics to Monitor
- Fallback rate: Should be < 30% under normal conditions
- Error rate: Should be < 5% sustained
- Rate limit hits: Should be < 1% in production
- Robots violations: Should be 0 in production

---

## Procedure 1: Forcing API-Only Mode

### When to Use
- Scraping endpoints are experiencing issues
- Need to validate API data quality
- Debugging fallback behavior
- Compliance requirements restrict web scraping

### Steps

1. **Identify Affected Endpoints**
   ```bash
   # Check which endpoints support API mode
   yfin scrape --list-endpoints --show-sources
   
   # Expected output:
   # quote: API + Scrape
   # daily-bars: API only
   # key-statistics: Scrape only
   # financials: Scrape only
   ```

2. **Test API Availability**
   ```bash
   # Test API endpoint health
   yfin scrape --fallback api-only --ticker AAPL --endpoint quote --check
   
   # Expected output:
   # ✓ API endpoint accessible
   # ✓ Authentication valid
   # ✓ Rate limits OK
   # ✓ Response format valid
   ```

3. **Force API Mode**
   ```bash
   # Single ticker
   yfin scrape --config configs/prod.yaml \
     --fallback api-only \
     --ticker AAPL \
     --endpoint quote \
     --preview
   
   # Multiple tickers
   yfin scrape --config configs/prod.yaml \
     --fallback api-only \
     --universe-file api-supported-tickers.txt \
     --endpoint quote \
     --publish \
     --env prod
   ```

4. **Monitor Results**
   ```bash
   # Check success rate
   curl -s "http://prometheus:9090/api/v1/query?query=rate(yfin_scrape_requests_total{source=\"api\",outcome=\"success\"}[5m])" | jq .
   
   # Check for fallback attempts (should be 0)
   curl -s "http://prometheus:9090/api/v1/query?query=rate(yfin_fallback_decisions_total[5m])" | jq .
   ```

### Troubleshooting API-Only Mode

**Issue: High failure rate in API-only mode**
```bash
# Diagnosis
yfin scrape --fallback api-only --ticker AAPL --endpoint quote --debug

# Common causes and fixes:
# 1. Authentication issues (401/403)
#    - Check API credentials
#    - Verify subscription status
# 2. Rate limiting (429)
#    - Reduce QPS: --qps 1.0
#    - Enable session rotation
# 3. Service unavailable (5xx)
#    - Check Yahoo Finance status
#    - Implement retry with backoff
```

---

## Procedure 2: Forcing Scrape-Only Mode

### When to Use
- API endpoints are down or rate-limited
- Need data not available through APIs
- Testing scrape infrastructure
- API subscription issues

### Steps

1. **Verify Scrape Endpoint Availability**
   ```bash
   # Check robots.txt compliance
   yfin scrape --ticker AAPL --endpoint key-statistics --check
   
   # Expected output:
   # ✓ Robots.txt allows access
   # ✓ Page loads successfully
   # ✓ Required data fields present
   # ✓ Parsing successful
   ```

2. **Force Scrape Mode**
   ```bash
   # Single endpoint
   yfin scrape --config configs/prod.yaml \
     --fallback scrape-only \
     --ticker AAPL \
     --endpoint key-statistics \
     --preview
   
   # Multiple endpoints
   yfin scrape --config configs/prod.yaml \
     --fallback scrape-only \
     --ticker AAPL \
     --endpoints key-statistics,financials,analysis,profile,news \
     --preview-json
   ```

3. **Monitor Scrape Performance**
   ```bash
   # Check response times (expect higher latency)
   curl -s "http://prometheus:9090/api/v1/query?query=histogram_quantile(0.95, rate(yfin_scrape_request_duration_seconds_bucket{source=\"scrape\"}[5m]))" | jq .
   
   # Check parse success rate
   curl -s "http://prometheus:9090/api/v1/query?query=rate(yfin_scrape_requests_total{source=\"scrape\",outcome=\"success\"}[5m])" | jq .
   ```

### Troubleshooting Scrape-Only Mode

**Issue: Parse failures in scrape mode**
```bash
# Diagnosis
yfin scrape --fallback scrape-only --ticker AAPL --endpoint key-statistics --debug

# Common causes and fixes:
# 1. HTML structure changes
#    - Update regex patterns
#    - Capture HTML for analysis: --save-html debug/
# 2. Rate limiting
#    - Reduce QPS: --qps 0.5
#    - Enable session rotation
# 3. Robots.txt blocks
#    - Check robots.txt policy
#    - Use alternative endpoints
```

---

## Procedure 3: Handling 401/403 Authentication Errors

### Immediate Response

1. **Assess Impact**
   ```bash
   # Check error rate for 401/403
   curl -s "http://prometheus:9090/api/v1/query?query=rate(yfin_scrape_errors_total{error_type=\"auth_error\"}[5m])" | jq .
   
   # Identify affected endpoints
   yfin scrape --ticker AAPL --endpoint fundamentals --check  # Likely to fail
   yfin scrape --ticker AAPL --endpoint quote --check         # Should work
   ```

2. **Enable Automatic Fallback**
   ```bash
   # Ensure fallback is enabled for auth errors
   yfin scrape --config configs/prod.yaml \
     --fallback auto \
     --ticker AAPL \
     --endpoint fundamentals \
     --preview
   
   # Expected behavior: API fails with 401 → automatically falls back to scraping
   ```

3. **Monitor Fallback Success**
   ```bash
   # Check fallback decisions due to auth errors
   curl -s "http://prometheus:9090/api/v1/query?query=rate(yfin_fallback_decisions_total{reason=\"auth_error\"}[5m])" | jq .
   ```

### Long-term Resolution

1. **Subscription Management**
   ```bash
   # Check subscription status (if applicable)
   # Contact Yahoo Finance support
   # Update API credentials if needed
   ```

2. **Endpoint Strategy Update**
   ```yaml
   # Update configuration to prefer scraping for paid endpoints
   scrape:
     fallback:
       strategy: "scrape-first"  # For paid endpoints
       endpoints:
         fundamentals: "scrape-only"
         key-statistics: "scrape-only"
         quote: "auto"  # Keep API for free endpoints
   ```

---

## Procedure 4: Managing Robots.txt Decisions

### Understanding Robots.txt Policies

```yaml
# Configuration options
scrape:
  robots_policy: "enforce"  # Production (strict compliance)
  robots_policy: "warn"     # Development (log violations)
  robots_policy: "ignore"   # Testing only (bypass checks)
```

### Procedure: Robots.txt Violation Response

1. **Immediate Assessment**
   ```bash
   # Check for robots violations
   curl -s "http://prometheus:9090/api/v1/query?query=rate(yfin_robots_blocked_total[5m])" | jq .
   
   # Identify blocked paths
   yfin scrape --ticker AAPL --endpoint key-statistics --check
   ```

2. **Verify Current Robots.txt**
   ```bash
   # Download and analyze current robots.txt
   curl -s https://finance.yahoo.com/robots.txt > robots.txt
   
   # Check specific paths
   grep -A 5 -B 5 "key-statistics" robots.txt
   grep -A 5 -B 5 "financials" robots.txt
   ```

3. **Response Actions**

   **If robots.txt blocks our paths:**
   ```bash
   # Option 1: Switch to allowed endpoints
   yfin scrape --ticker AAPL --endpoint profile --check  # Check if allowed
   
   # Option 2: Use API fallback
   yfin scrape --fallback api-only --ticker AAPL --endpoint quote
   
   # Option 3: Update user-agent (if that's the issue)
   yfin scrape --user-agent "Mozilla/5.0 (compatible; yfinance-go/1.0)" --ticker AAPL --endpoint key-statistics --check
   ```

   **If robots.txt is overly restrictive:**
   ```bash
   # NEVER use --force in production
   # Instead, find alternative data sources or endpoints
   yfin scrape --ticker AAPL --endpoint profile --preview  # May contain similar data
   ```

### Safe Use of --force Flag

**⚠️ CRITICAL WARNING: --force should NEVER be used in production**

**Acceptable use cases (testing only):**
1. Local development and testing
2. Debugging parsing issues
3. Validating data extraction logic
4. Research and analysis (non-commercial)

**Safe testing procedure:**
```bash
# 1. Use only in development environment
export ENVIRONMENT=development

# 2. Use with preview mode only
yfin scrape --force --ticker AAPL --endpoint profile --preview --config configs/dev.yaml

# 3. Log the usage
echo "$(date): Used --force for testing ticker AAPL endpoint profile" >> force-usage.log

# 4. Never use with --publish
# NEVER: yfin scrape --force --publish  # This violates terms of service
```

---

## Procedure 5: Rate Limit Management

### Detecting Rate Limit Issues

1. **Monitor Rate Limit Metrics**
   ```bash
   # Check rate limit hit frequency
   curl -s "http://prometheus:9090/api/v1/query?query=rate(yfin_rate_limit_hits_total[5m])" | jq .
   
   # Check backoff events
   curl -s "http://prometheus:9090/api/v1/query?query=rate(yfin_backoff_events_total[5m])" | jq .
   ```

2. **Identify Rate Limit Patterns**
   ```bash
   # Check which hosts are rate limiting
   curl -s "http://prometheus:9090/api/v1/query?query=rate(yfin_rate_limit_hits_total[5m]) by (host)" | jq .
   ```

### Rate Limit Mitigation

1. **Immediate Actions**
   ```bash
   # Reduce QPS globally
   yfin scrape --qps 1.0 --ticker AAPL --endpoint key-statistics
   
   # Reduce concurrency
   yfin scrape --concurrency 4 --ticker AAPL --endpoint key-statistics
   
   # Enable session rotation
   yfin scrape --sessions 10 --ticker AAPL --endpoint key-statistics
   ```

2. **Configuration Updates**
   ```yaml
   # Update production config
   scrape:
     qps: 1.0  # Reduced from 2.0
     burst: 3  # Reduced from 5
     
   session_rotation:
     enabled: true
     pool_size: 15  # Increased from 5
     rotation_interval_minutes: 10  # More frequent rotation
     
   retry:
     backoff_base_ms: 2000  # Increased backoff
     backoff_max_ms: 30000
   ```

---

## Procedure 6: Emergency Procedures

### Complete System Shutdown

```bash
# 1. Stop all scraping operations
pkill -f "yfin scrape"

# 2. Verify no active connections
netstat -an | grep :443 | grep finance.yahoo.com

# 3. Clear session pools
rm -rf /tmp/yfin-sessions-*

# 4. Reset rate limiters
curl -X POST http://localhost:8080/admin/reset-rate-limiters
```

### Emergency Fallback to API-Only

```bash
# 1. Update configuration
cat > emergency-config.yaml << EOF
scrape:
  fallback:
    strategy: "api-only"
  enabled: false  # Disable scraping entirely
EOF

# 2. Restart with emergency config
yfin scrape --config emergency-config.yaml --ticker AAPL --endpoint quote --check

# 3. Monitor API-only performance
curl -s "http://prometheus:9090/api/v1/query?query=rate(yfin_scrape_requests_total{source=\"api\"}[5m])" | jq .
```

### Data Quality Incident Response

```bash
# 1. Capture evidence
yfin scrape --ticker AAPL --endpoint key-statistics --save-html incident/
yfin scrape --ticker AAPL --endpoint key-statistics --save-parsed incident/

# 2. Switch to known-good configuration
yfin scrape --config configs/last-known-good.yaml --ticker AAPL --endpoint key-statistics

# 3. Test with multiple tickers
for ticker in AAPL MSFT GOOGL; do
  yfin scrape --ticker $ticker --endpoint key-statistics --check
done

# 4. Document incident
echo "$(date): Data quality incident - see incident/ directory" >> incidents.log
```

---

## Monitoring and Alerting

### Key Metrics to Watch

```bash
# Success rate (should be > 95%)
curl -s "http://prometheus:9090/api/v1/query?query=rate(yfin_scrape_requests_total{outcome=\"success\"}[5m]) / rate(yfin_scrape_requests_total[5m]) * 100" | jq .

# Fallback rate (should be < 30%)
curl -s "http://prometheus:9090/api/v1/query?query=rate(yfin_fallback_decisions_total[5m]) / rate(yfin_scrape_requests_total[5m]) * 100" | jq .

# Rate limit violations (should be 0 in production)
curl -s "http://prometheus:9090/api/v1/query?query=rate(yfin_rate_limit_hits_total[5m])" | jq .

# Robots violations (should be 0 in production)
curl -s "http://prometheus:9090/api/v1/query?query=rate(yfin_robots_blocked_total[5m])" | jq .
```

### Alert Thresholds

- **Warning**: Error rate > 5% for 5 minutes
- **Critical**: Error rate > 10% for 2 minutes
- **Critical**: Any robots.txt violations in production
- **Warning**: Fallback rate > 50% for 10 minutes
- **Critical**: Rate limit hits > 0.1/second for 1 minute

---

## Best Practices

### 1. Configuration Management
- Use environment-specific configurations
- Never use `--force` in production
- Keep robots_policy as "enforce" in production
- Regularly review and update rate limits

### 2. Monitoring
- Set up comprehensive alerting
- Monitor key metrics continuously
- Regular health checks on all endpoints
- Track configuration changes

### 3. Incident Response
- Capture evidence before making changes
- Document all incidents and resolutions
- Test fixes thoroughly before production deployment
- Maintain rollback procedures

### 4. Compliance
- Respect robots.txt at all times in production
- Keep rate limits conservative
- Monitor for terms of service changes
- Regular compliance audits

This runbook provides comprehensive procedures for managing the scrape fallback system safely and effectively. Always prioritize compliance and system stability over data availability.
