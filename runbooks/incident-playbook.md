# Incident Response Playbook

## Overview

This playbook provides structured response procedures for common incidents in the yfinance-go scrape fallback system, including rate-limit spikes, layout drift, DLQ events, and system outages.

## Incident Classification

### Severity Levels

**P0 - Critical (< 15 min response)**
- Complete system outage
- Data corruption or loss
- Security breach
- Compliance violations (robots.txt violations in production)

**P1 - High (< 1 hour response)**
- Partial system outage (>50% error rate)
- Significant performance degradation
- Rate limit violations causing service impact

**P2 - Medium (< 4 hours response)**
- Intermittent failures (10-50% error rate)
- Schema drift causing parse failures
- High fallback rates

**P3 - Low (< 24 hours response)**
- Minor performance issues
- Non-critical parse warnings
- Documentation or configuration issues

---

## Incident 1: Rate-Limit Spikes

### Symptoms
- High 429/503 error rates (>5% sustained)
- Increased response times
- Circuit breakers opening
- Fallback rate increases

### Immediate Response (< 5 minutes)

1. **Assess Impact**
   ```bash
   # Check current error rate
   curl -s "http://prometheus:9090/api/v1/query?query=rate(yfin_rate_limit_hits_total[5m])" | jq .
   
   # Check affected endpoints
   curl -s "http://prometheus:9090/api/v1/query?query=rate(yfin_rate_limit_hits_total[5m]) by (endpoint)" | jq .
   ```

2. **Emergency Mitigation**
   ```bash
   # Immediately reduce QPS
   kubectl patch configmap yfinance-config --patch '{"data":{"qps":"0.5"}}'
   
   # Or via CLI override
   yfin scrape --qps 0.5 --concurrency 2 --ticker AAPL --endpoint key-statistics --check
   ```

3. **Enable Session Rotation**
   ```bash
   # Increase session pool size
   kubectl patch configmap yfinance-config --patch '{"data":{"session_pool_size":"20"}}'
   
   # Verify session rotation is working
   yfin scrape --session-status --debug
   ```

### Investigation (< 15 minutes)

1. **Identify Root Cause**
   ```bash
   # Check QPS over time
   curl -s "http://prometheus:9090/api/v1/query_range?query=rate(yfin_scrape_requests_total[1m])&start=$(date -d '1 hour ago' +%s)&end=$(date +%s)&step=60" | jq .
   
   # Check for traffic spikes
   curl -s "http://prometheus:9090/api/v1/query?query=increase(yfin_scrape_requests_total[1h])" | jq .
   
   # Review recent deployments
   kubectl get events --sort-by='.lastTimestamp' | grep yfinance
   ```

2. **Check Configuration Changes**
   ```bash
   # Review recent config changes
   git log --oneline --since="1 hour ago" configs/
   
   # Check effective configuration
   yfin config --print-effective | grep -A 10 rate_limit
   ```

### Resolution Actions

1. **Short-term Fix**
   ```yaml
   # Update configuration
   scrape:
     qps: 1.0  # Reduced from previous value
     burst: 3  # Reduced burst allowance
     per_host_qps: 0.5  # Add per-host limits
     
   session_rotation:
     enabled: true
     pool_size: 15  # Increased pool size
     rotation_interval_minutes: 10  # More frequent rotation
   ```

2. **Monitor Recovery**
   ```bash
   # Watch error rate decrease
   watch "curl -s 'http://prometheus:9090/api/v1/query?query=rate(yfin_rate_limit_hits_total[5m])' | jq '.data.result[0].value[1]'"
   
   # Verify normal operation
   yfin scrape --ticker AAPL --endpoint key-statistics --check
   ```

3. **Gradual QPS Increase**
   ```bash
   # Slowly increase QPS if system is stable
   # Wait 10 minutes between increases
   kubectl patch configmap yfinance-config --patch '{"data":{"qps":"1.0"}}'
   # Wait and monitor...
   kubectl patch configmap yfinance-config --patch '{"data":{"qps":"1.5"}}'
   ```

### Post-Incident Actions

1. **Root Cause Analysis**
   - Review traffic patterns leading to spike
   - Identify configuration or deployment triggers
   - Document lessons learned

2. **Prevention Measures**
   - Implement gradual QPS ramping
   - Add pre-deployment rate limit testing
   - Improve monitoring and alerting

---

## Incident 2: Layout Drift (Schema Changes)

### Symptoms
- Sudden increase in parse failures
- ErrSchemaDrift errors in logs
- Missing or malformed data fields
- Schema validation failures

### Immediate Response (< 10 minutes)

1. **Assess Impact**
   ```bash
   # Check parse failure rate
   curl -s "http://prometheus:9090/api/v1/query?query=rate(yfin_scrape_parse_failures_total[5m])" | jq .
   
   # Identify affected endpoints
   curl -s "http://prometheus:9090/api/v1/query?query=rate(yfin_scrape_parse_failures_total[5m]) by (endpoint)" | jq .
   ```

2. **Capture Evidence**
   ```bash
   # Save current HTML for analysis
   mkdir -p incident/layout-drift-$(date +%Y%m%d-%H%M%S)
   
   # Capture HTML from multiple tickers
   for ticker in AAPL MSFT GOOGL; do
     yfin scrape --ticker $ticker --endpoint key-statistics --save-html incident/layout-drift-$(date +%Y%m%d-%H%M%S)/
   done
   ```

3. **Enable Fallback**
   ```bash
   # Switch to API-only for affected endpoints (if available)
   yfin scrape --fallback api-only --ticker AAPL --endpoint quote --check
   
   # Or enable lenient parsing
   yfin scrape --strict-mode=false --ticker AAPL --endpoint key-statistics --preview
   ```

### Investigation (< 30 minutes)

1. **Compare HTML Structure**
   ```bash
   # Compare current vs. previous HTML
   diff -u incident/previous-good.html incident/layout-drift-*/AAPL-key-statistics.html
   
   # Look for structural changes
   grep -n "quote-summary\|key-statistics\|financials" incident/layout-drift-*/AAPL-key-statistics.html
   ```

2. **Test Multiple Tickers**
   ```bash
   # Check if issue is universal or ticker-specific
   for ticker in AAPL MSFT GOOGL TSLA AMZN; do
     echo "Testing $ticker:"
     yfin scrape --ticker $ticker --endpoint key-statistics --check
   done
   ```

3. **Identify Changed Fields**
   ```bash
   # Test parsing with debug output
   yfin scrape --ticker AAPL --endpoint key-statistics --debug --preview 2>&1 | grep -E "ERROR|WARN|missing|failed"
   ```

### Resolution Actions

1. **Immediate Workaround**
   ```bash
   # Use alternative endpoints if available
   yfin scrape --ticker AAPL --endpoint profile --preview  # May have some key stats
   
   # Enable lenient parsing temporarily
   kubectl patch configmap yfinance-config --patch '{"data":{"strict_parsing":"false"}}'
   ```

2. **Update Parsing Logic**
   ```go
   // Update regex patterns in internal/scrape/regex/key_statistics.yaml
   market_cap:
     selectors:
       - "td[data-test='MARKET_CAP-value']"  # New selector
       - "span[data-field='marketCap']"      # Fallback
       - "div[class*='market-cap']"          # Generic fallback
   ```

3. **Test Fix**
   ```bash
   # Test updated parsing
   go build -o bin/yfin ./cmd/yfin
   ./bin/yfin scrape --ticker AAPL --endpoint key-statistics --preview
   
   # Test with multiple tickers
   for ticker in AAPL MSFT GOOGL; do
     ./bin/yfin scrape --ticker $ticker --endpoint key-statistics --check
   done
   ```

### Deployment and Verification

1. **Deploy Fix**
   ```bash
   # Build and deploy updated version
   docker build -t yfinance-go:fix-layout-drift .
   kubectl set image deployment/yfinance-go yfinance-go=yfinance-go:fix-layout-drift
   ```

2. **Monitor Recovery**
   ```bash
   # Watch parse success rate
   watch "curl -s 'http://prometheus:9090/api/v1/query?query=rate(yfin_scrape_requests_total{outcome=\"success\"}[5m]) / rate(yfin_scrape_requests_total[5m]) * 100' | jq '.data.result[0].value[1]'"
   ```

3. **Re-enable Strict Mode**
   ```bash
   # Once parsing is stable, re-enable strict mode
   kubectl patch configmap yfinance-config --patch '{"data":{"strict_parsing":"true"}}'
   ```

---

## Incident 3: DLQ Events (Dead Letter Queue)

### Symptoms
- Messages appearing in dead letter queue
- Publishing failures
- Message ordering violations
- Bus connectivity issues

### Immediate Response (< 5 minutes)

1. **Check DLQ Status**
   ```bash
   # Check DLQ message count
   curl -s "http://bus-admin:8080/api/v1/queues/yfinance-dlq/stats" | jq .
   
   # Check recent DLQ messages
   curl -s "http://bus-admin:8080/api/v1/queues/yfinance-dlq/messages?limit=10" | jq .
   ```

2. **Assess Publishing Health**
   ```bash
   # Check bus connectivity
   yfin scrape --ticker AAPL --endpoint news --publish --env staging --dry-run
   
   # Check publishing metrics
   curl -s "http://prometheus:9090/api/v1/query?query=rate(yfin_bus_publish_failures_total[5m])" | jq .
   ```

### Investigation (< 15 minutes)

1. **Analyze DLQ Messages**
   ```bash
   # Download DLQ messages for analysis
   curl -s "http://bus-admin:8080/api/v1/queues/yfinance-dlq/messages?limit=100" > dlq-messages.json
   
   # Analyze failure reasons
   jq '.messages[] | .dlq_reason' dlq-messages.json | sort | uniq -c
   ```

2. **Check Message Format**
   ```bash
   # Validate message schema
   jq '.messages[0] | .payload' dlq-messages.json | yfin validate-schema --stdin
   
   # Check for malformed messages
   jq '.messages[] | select(.dlq_reason == "schema_validation_failed")' dlq-messages.json
   ```

3. **Verify Bus Configuration**
   ```bash
   # Check bus connection settings
   yfin config --print-effective | grep -A 20 bus
   
   # Test bus connectivity
   yfin bus --health-check
   ```

### Resolution Actions

1. **Fix Schema Issues**
   ```bash
   # If schema validation failures
   # Update message mapping logic
   git diff HEAD~1 internal/emit/map_*.go
   
   # Test message generation
   yfin scrape --ticker AAPL --endpoint news --preview-proto --validate-schema
   ```

2. **Retry DLQ Messages**
   ```bash
   # Replay messages from DLQ (after fixing root cause)
   curl -X POST "http://bus-admin:8080/api/v1/queues/yfinance-dlq/replay" \
     -H "Content-Type: application/json" \
     -d '{"max_messages": 100, "target_topic": "ampy.prod.news"}'
   ```

3. **Fix Ordering Issues**
   ```bash
   # Check partition key generation
   yfin scrape --ticker AAPL --endpoint news --debug | grep "partition_key"
   
   # Verify ordering preservation
   yfin scrape --ticker AAPL --endpoint news --check-ordering
   ```

### Prevention Measures

1. **Enhanced Validation**
   ```yaml
   # Add pre-publish validation
   bus:
     validation:
       enabled: true
       strict_schema: true
       validate_before_publish: true
   ```

2. **Monitoring Improvements**
   ```bash
   # Add DLQ monitoring
   # Alert on any DLQ messages
   # Monitor schema validation failures
   ```

---

## Incident 4: Complete System Outage

### Symptoms
- All endpoints returning errors
- No successful requests
- Health checks failing
- Circuit breakers open

### Immediate Response (< 2 minutes)

1. **Confirm Outage Scope**
   ```bash
   # Check overall health
   curl -f http://localhost:8080/health || echo "Health check failed"
   
   # Test basic functionality
   yfin scrape --ticker AAPL --endpoint key-statistics --check --timeout 10s
   ```

2. **Check External Dependencies**
   ```bash
   # Test Yahoo Finance availability
   curl -I https://finance.yahoo.com/ --max-time 10
   
   # Check DNS resolution
   nslookup finance.yahoo.com
   
   # Test network connectivity
   ping -c 3 finance.yahoo.com
   ```

### Investigation (< 5 minutes)

1. **Check System Resources**
   ```bash
   # Check CPU and memory
   top -n 1 | head -20
   
   # Check disk space
   df -h
   
   # Check network connections
   netstat -tuln | grep :8080
   ```

2. **Review Recent Changes**
   ```bash
   # Check recent deployments
   kubectl get events --sort-by='.lastTimestamp' | head -20
   
   # Check configuration changes
   git log --oneline --since="2 hours ago"
   ```

3. **Analyze Logs**
   ```bash
   # Check application logs
   kubectl logs -l app=yfinance-go --tail=100
   
   # Check for error patterns
   kubectl logs -l app=yfinance-go --tail=1000 | grep -E "ERROR|FATAL|panic"
   ```

### Resolution Actions

1. **Emergency Rollback**
   ```bash
   # Rollback to previous version
   kubectl rollout undo deployment/yfinance-go
   
   # Wait for rollback to complete
   kubectl rollout status deployment/yfinance-go
   ```

2. **Service Restart**
   ```bash
   # If rollback doesn't work, restart service
   kubectl delete pods -l app=yfinance-go
   
   # Wait for pods to restart
   kubectl get pods -l app=yfinance-go -w
   ```

3. **Configuration Reset**
   ```bash
   # Reset to known-good configuration
   kubectl apply -f configs/last-known-good.yaml
   
   # Restart with clean state
   kubectl delete pods -l app=yfinance-go
   ```

### Verification

1. **Test Core Functionality**
   ```bash
   # Test each endpoint type
   yfin scrape --ticker AAPL --endpoint key-statistics --check
   yfin scrape --ticker AAPL --endpoint financials --check
   yfin scrape --ticker AAPL --endpoint news --check
   ```

2. **Monitor Recovery**
   ```bash
   # Watch success rate
   watch "curl -s 'http://prometheus:9090/api/v1/query?query=rate(yfin_scrape_requests_total{outcome=\"success\"}[5m])' | jq '.data.result[0].value[1]'"
   ```

---

## Communication Templates

### Internal Incident Notification

```
INCIDENT: yfinance-go P1 - High Error Rate
STATUS: Investigating
IMPACT: 45% of scraping requests failing
START TIME: 2024-01-03 10:30 UTC
NEXT UPDATE: 2024-01-03 11:00 UTC

ACTIONS TAKEN:
- Reduced QPS from 2.0 to 1.0
- Enabled session rotation
- Investigating rate limit causes

WORKAROUND:
- API-only mode available for quotes
- Fallback working for key-statistics

TEAM: @yfinance-oncall
```

### Customer Communication

```
We are currently experiencing elevated error rates in our Yahoo Finance data collection service. 

IMPACT: Some financial data requests may be delayed or temporarily unavailable.
WORKAROUND: API-based data (quotes, historical prices) remains fully functional.
ETA: Resolution expected within 2 hours.

We will provide updates every 30 minutes until resolved.
```

### Resolution Notification

```
RESOLVED: yfinance-go P1 - High Error Rate
DURATION: 45 minutes
ROOT CAUSE: Configuration change increased QPS beyond rate limits
RESOLUTION: Reverted QPS settings and enabled session rotation

PREVENTION:
- Added QPS validation to deployment pipeline
- Enhanced rate limit monitoring
- Updated runbook procedures

POST-MORTEM: Scheduled for 2024-01-04 14:00 UTC
```

---

## Post-Incident Procedures

### 1. Immediate Post-Resolution (< 1 hour)

```bash
# Document incident timeline
echo "$(date): Incident resolved - $(git rev-parse HEAD)" >> incidents.log

# Capture final state
yfin scrape --ticker AAPL --endpoint key-statistics --check > post-incident-check.log

# Update monitoring dashboards
# Add any new metrics discovered during incident
```

### 2. Post-Mortem Preparation (< 24 hours)

1. **Gather Evidence**
   - Timeline of events
   - Metrics and logs
   - Actions taken
   - Impact assessment

2. **Root Cause Analysis**
   - Technical root cause
   - Process failures
   - Contributing factors

3. **Action Items**
   - Prevention measures
   - Monitoring improvements
   - Process updates

### 3. Follow-up Actions (< 1 week)

1. **Implement Fixes**
   - Code changes
   - Configuration updates
   - Monitoring enhancements

2. **Update Documentation**
   - Runbook improvements
   - Alert threshold adjustments
   - Process refinements

3. **Team Training**
   - Share lessons learned
   - Update incident response procedures
   - Conduct tabletop exercises

---

## Emergency Contacts

### On-Call Rotation
- Primary: @yfinance-primary
- Secondary: @yfinance-secondary
- Escalation: @engineering-manager

### External Dependencies
- Yahoo Finance Status: https://status.yahoo.com/
- Cloud Provider Status: [Provider Status Page]
- Bus Service Team: @bus-team

### Communication Channels
- Incident Channel: #yfinance-incidents
- General Updates: #yfinance-alerts
- Customer Communication: @customer-success

This incident playbook provides structured response procedures for all major incident types. Regular drills and updates ensure effective incident response.
