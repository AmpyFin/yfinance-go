# Observability Guide

## Overview

The yfinance-go scrape fallback system provides comprehensive observability through metrics, logging, and distributed tracing. This guide covers all monitoring capabilities, alerting strategies, and operational dashboards.

## Metrics Collection

### Prometheus Metrics

The system exposes Prometheus-compatible metrics on `/metrics` endpoint (default port 8080).

#### Core Request Metrics

```prometheus
# Request rates and success
yfin_scrape_requests_total{endpoint, ticker, outcome, source}
yfin_scrape_request_duration_seconds{endpoint, source}
yfin_scrape_requests_in_flight{endpoint}

# Fallback behavior
yfin_fallback_decisions_total{from_source, to_source, reason}
yfin_fallback_rate{endpoint}

# Error tracking
yfin_scrape_errors_total{endpoint, error_type, source}
yfin_scrape_parse_failures_total{endpoint, failure_type}
```

#### Rate Limiting Metrics

```prometheus
# Rate limit tracking
yfin_rate_limit_hits_total{host, endpoint}
yfin_rate_limit_wait_duration_seconds{host}
yfin_backoff_events_total{host, reason}

# Circuit breaker
yfin_circuit_breaker_state{host}  # 0=closed, 1=open, 2=half-open
yfin_circuit_breaker_failures_total{host}
```

#### Robots.txt Compliance

```prometheus
# Robots compliance
yfin_robots_blocked_total{host, path}
yfin_robots_cache_hits_total{host}
yfin_robots_policy_violations_total{host}
```

#### Data Quality Metrics

```prometheus
# Schema validation
yfin_schema_validation_failures_total{endpoint, field}
yfin_data_completeness_ratio{endpoint, ticker}

# Parsing success
yfin_parse_success_rate{endpoint}
yfin_field_extraction_success_total{endpoint, field}
```

### Metric Examples with Queries

#### Request Success Rate
```promql
# Overall success rate (last 5 minutes)
rate(yfin_scrape_requests_total{outcome="success"}[5m]) / 
rate(yfin_scrape_requests_total[5m]) * 100

# Per-endpoint success rate
rate(yfin_scrape_requests_total{outcome="success"}[5m]) by (endpoint) / 
rate(yfin_scrape_requests_total[5m]) by (endpoint) * 100
```

#### Fallback Rate
```promql
# Percentage of requests using fallback
rate(yfin_fallback_decisions_total[5m]) / 
rate(yfin_scrape_requests_total[5m]) * 100

# Fallback reasons breakdown
rate(yfin_fallback_decisions_total[5m]) by (reason)
```

#### Response Time Percentiles
```promql
# 95th percentile response time by endpoint
histogram_quantile(0.95, 
  rate(yfin_scrape_request_duration_seconds_bucket[5m])) by (endpoint)

# Response time comparison: API vs Scrape
histogram_quantile(0.50, 
  rate(yfin_scrape_request_duration_seconds_bucket{source="api"}[5m])) by (endpoint)
vs
histogram_quantile(0.50, 
  rate(yfin_scrape_request_duration_seconds_bucket{source="scrape"}[5m])) by (endpoint)
```

#### Error Rate Analysis
```promql
# Error rate by type
rate(yfin_scrape_errors_total[5m]) by (error_type)

# Rate limit hit frequency
rate(yfin_rate_limit_hits_total[5m]) by (host)
```

## Logging

### Structured Logging Format

All logs use structured JSON format with consistent fields:

```json
{
  "timestamp": "2024-01-03T10:30:15Z",
  "level": "info",
  "source": "yfinance-go/scrape",
  "message": "scrape request completed",
  "fields": {
    "ticker": "AAPL",
    "endpoint": "key-statistics",
    "url": "https://finance.yahoo.com/quote/AAPL/key-statistics",
    "status": 200,
    "duration_ms": 1250,
    "bytes": 2048576,
    "attempt": 1,
    "source": "scrape",
    "run_id": "batch-20240103-001",
    "gzip": true,
    "redirects": 0
  }
}
```

### Log Levels and Content

#### DEBUG Level
```json
{
  "level": "debug",
  "message": "robots.txt check",
  "fields": {
    "host": "finance.yahoo.com",
    "path": "/quote/AAPL/key-statistics",
    "allowed": true,
    "cache_hit": true,
    "policy": "enforce"
  }
}
```

#### INFO Level
```json
{
  "level": "info", 
  "message": "fallback decision",
  "fields": {
    "ticker": "AAPL",
    "endpoint": "key-statistics",
    "from_source": "api",
    "to_source": "scrape",
    "reason": "rate_limited",
    "api_error": "HTTP 429: Too Many Requests"
  }
}
```

#### WARN Level
```json
{
  "level": "warn",
  "message": "parse warning",
  "fields": {
    "ticker": "AAPL",
    "endpoint": "key-statistics", 
    "warning": "missing_field",
    "field": "market_cap",
    "impact": "field_skipped"
  }
}
```

#### ERROR Level
```json
{
  "level": "error",
  "message": "schema validation failed",
  "fields": {
    "ticker": "AAPL",
    "endpoint": "key-statistics",
    "error": "ErrSchemaDrift",
    "field": "pe_ratio",
    "expected_type": "number",
    "actual_type": "string",
    "value": "N/A"
  }
}
```

### Log Aggregation Queries

#### Splunk Queries
```splunk
# High error rate detection
index=yfinance source=scrape level=error 
| stats count by endpoint, error_type 
| where count > 10

# Fallback pattern analysis
index=yfinance source=scrape message="fallback decision"
| stats count by reason, endpoint
| sort -count

# Performance analysis
index=yfinance source=scrape message="scrape request completed"
| stats avg(duration_ms), p95(duration_ms) by endpoint, source
```

#### ELK Stack Queries
```json
// High latency requests
{
  "query": {
    "bool": {
      "must": [
        {"term": {"source": "yfinance-go/scrape"}},
        {"range": {"fields.duration_ms": {"gte": 5000}}}
      ]
    }
  }
}

// Parse failure analysis
{
  "query": {
    "bool": {
      "must": [
        {"term": {"level": "error"}},
        {"term": {"fields.error": "ErrSchemaDrift"}}
      ]
    }
  },
  "aggs": {
    "by_endpoint": {
      "terms": {"field": "fields.endpoint"}
    }
  }
}
```

## Distributed Tracing

### OpenTelemetry Integration

The system supports OpenTelemetry tracing with comprehensive span coverage:

#### Trace Structure
```
Request Span (yfin.scrape.request)
├── Robots Check Span (yfin.scrape.robots)
├── Rate Limit Span (yfin.scrape.ratelimit)
├── HTTP Request Span (yfin.scrape.http)
│   ├── DNS Resolution Span
│   ├── TCP Connection Span
│   └── HTTP Response Span
├── Parse Span (yfin.scrape.parse)
│   ├── HTML Parse Span
│   ├── JSON Extract Span
│   └── Field Mapping Span
├── Validation Span (yfin.scrape.validate)
└── Emit Span (yfin.scrape.emit)
```

#### Span Attributes
```json
{
  "trace_id": "1234567890abcdef",
  "span_id": "abcdef1234567890",
  "operation_name": "yfin.scrape.request",
  "start_time": "2024-01-03T10:30:15.123Z",
  "duration": "1.25s",
  "tags": {
    "ticker": "AAPL",
    "endpoint": "key-statistics",
    "source": "scrape",
    "run_id": "batch-20240103-001",
    "http.method": "GET",
    "http.url": "https://finance.yahoo.com/quote/AAPL/key-statistics",
    "http.status_code": 200,
    "robots.allowed": true,
    "parse.success": true,
    "fallback.used": false
  }
}
```

### Jaeger Trace Analysis

#### Common Trace Patterns

**Successful Request:**
```
Request [1.2s]
├── Robots [5ms] ✓
├── Rate Limit [2ms] ✓  
├── HTTP [800ms] ✓
├── Parse [350ms] ✓
├── Validate [25ms] ✓
└── Emit [18ms] ✓
```

**Fallback Request:**
```
Request [2.1s]
├── API Attempt [450ms] ✗ (429 Rate Limited)
├── Fallback Decision [2ms]
├── Robots [3ms] ✓
├── Rate Limit [1ms] ✓
├── HTTP [1.2s] ✓
├── Parse [380ms] ✓
├── Validate [22ms] ✓
└── Emit [15ms] ✓
```

**Failed Request:**
```
Request [5.2s]
├── Robots [4ms] ✓
├── Rate Limit [1ms] ✓
├── HTTP [3.8s] ✗ (Timeout)
├── Retry 1 [1.2s] ✗ (503 Service Unavailable)
└── Retry 2 [200ms] ✗ (Connection Reset)
```

## Dashboard Configuration

### Grafana Dashboard

#### Overview Dashboard

```json
{
  "dashboard": {
    "title": "yfinance-go Scrape Overview",
    "panels": [
      {
        "title": "Request Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(yfin_scrape_requests_total[5m])",
            "legendFormat": "{{endpoint}}"
          }
        ]
      },
      {
        "title": "Success Rate",
        "type": "singlestat",
        "targets": [
          {
            "expr": "rate(yfin_scrape_requests_total{outcome=\"success\"}[5m]) / rate(yfin_scrape_requests_total[5m]) * 100"
          }
        ]
      },
      {
        "title": "Response Time P95",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(yfin_scrape_request_duration_seconds_bucket[5m]))",
            "legendFormat": "{{endpoint}}"
          }
        ]
      },
      {
        "title": "Fallback Rate",
        "type": "graph", 
        "targets": [
          {
            "expr": "rate(yfin_fallback_decisions_total[5m]) / rate(yfin_scrape_requests_total[5m]) * 100",
            "legendFormat": "Fallback %"
          }
        ]
      }
    ]
  }
}
```

#### Error Analysis Dashboard

```json
{
  "dashboard": {
    "title": "yfinance-go Error Analysis",
    "panels": [
      {
        "title": "Error Rate by Type",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(yfin_scrape_errors_total[5m])",
            "legendFormat": "{{error_type}}"
          }
        ]
      },
      {
        "title": "Rate Limit Hits",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(yfin_rate_limit_hits_total[5m])",
            "legendFormat": "{{host}}"
          }
        ]
      },
      {
        "title": "Parse Failures",
        "type": "table",
        "targets": [
          {
            "expr": "rate(yfin_scrape_parse_failures_total[5m])",
            "format": "table"
          }
        ]
      }
    ]
  }
}
```

### DataDog Dashboard

```yaml
# datadog-dashboard.yaml
dashboard:
  title: "yfinance-go Scrape Monitoring"
  widgets:
    - title: "Request Throughput"
      type: "timeseries"
      requests:
        - q: "sum:yfin.scrape.requests.total{*}.as_rate()"
          
    - title: "Error Rate"
      type: "query_value"
      requests:
        - q: "sum:yfin.scrape.errors.total{*}.as_rate() / sum:yfin.scrape.requests.total{*}.as_rate() * 100"
          
    - title: "Response Time Distribution"
      type: "distribution"
      requests:
        - q: "avg:yfin.scrape.request.duration{*} by {endpoint}"
```

## Alerting Configuration

### Prometheus Alerting Rules

```yaml
# alerts.yml
groups:
  - name: yfinance-scrape
    rules:
      - alert: HighErrorRate
        expr: |
          (
            rate(yfin_scrape_errors_total[5m]) / 
            rate(yfin_scrape_requests_total[5m])
          ) * 100 > 5
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High error rate in yfinance scraping"
          description: "Error rate is {{ $value }}% for the last 5 minutes"
          
      - alert: CriticalErrorRate  
        expr: |
          (
            rate(yfin_scrape_errors_total[5m]) / 
            rate(yfin_scrape_requests_total[5m])
          ) * 100 > 10
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "Critical error rate in yfinance scraping"
          description: "Error rate is {{ $value }}% for the last 2 minutes"
          
      - alert: HighFallbackRate
        expr: |
          (
            rate(yfin_fallback_decisions_total[5m]) / 
            rate(yfin_scrape_requests_total[5m])
          ) * 100 > 50
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "High fallback rate detected"
          description: "{{ $value }}% of requests are using fallback"
          
      - alert: RateLimitViolation
        expr: rate(yfin_rate_limit_hits_total[5m]) > 0.1
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Rate limit violations detected"
          description: "{{ $value }} rate limit hits per second"
          
      - alert: RobotsViolation
        expr: rate(yfin_robots_blocked_total[5m]) > 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Robots.txt violations detected"
          description: "{{ $value }} robots.txt violations per second"
          
      - alert: HighLatency
        expr: |
          histogram_quantile(0.95, 
            rate(yfin_scrape_request_duration_seconds_bucket[5m])
          ) > 5
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High response latency"
          description: "95th percentile latency is {{ $value }}s"
          
      - alert: ParseFailures
        expr: rate(yfin_scrape_parse_failures_total[5m]) > 0.05
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Parse failures detected"
          description: "{{ $value }} parse failures per second"
```

### PagerDuty Integration

```yaml
# pagerduty-config.yml
route:
  group_by: ['alertname', 'severity']
  group_wait: 30s
  group_interval: 5m
  repeat_interval: 12h
  receiver: 'yfinance-team'
  routes:
    - match:
        severity: critical
      receiver: 'yfinance-oncall'
      
receivers:
  - name: 'yfinance-team'
    pagerduty_configs:
      - service_key: 'your-service-key'
        description: '{{ range .Alerts }}{{ .Annotations.summary }}{{ end }}'
        
  - name: 'yfinance-oncall'
    pagerduty_configs:
      - service_key: 'your-oncall-key'
        description: '{{ range .Alerts }}{{ .Annotations.summary }}{{ end }}'
        severity: 'critical'
```

## Health Checks

### Application Health Endpoint

```bash
# Health check endpoint
curl http://localhost:8080/health

# Response
{
  "status": "healthy",
  "timestamp": "2024-01-03T10:30:15Z",
  "version": "1.0.0",
  "checks": {
    "scrape_client": "healthy",
    "rate_limiter": "healthy", 
    "circuit_breaker": "closed",
    "robots_cache": "healthy"
  },
  "metrics": {
    "uptime_seconds": 3600,
    "total_requests": 15420,
    "success_rate": 0.967,
    "avg_response_time_ms": 1250
  }
}
```

### Kubernetes Health Checks

```yaml
# k8s-deployment.yaml
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
        - name: yfinance-go
          livenessProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 30
            periodSeconds: 10
            
          readinessProbe:
            httpGet:
              path: /ready
              port: 8080
            initialDelaySeconds: 5
            periodSeconds: 5
```

## Operational Runbooks

### Daily Health Check

```bash
#!/bin/bash
# daily-health-check.sh

echo "=== yfinance-go Daily Health Check ==="
echo "Date: $(date)"

# Check overall health
echo "1. Application Health:"
curl -s http://localhost:8080/health | jq .

# Check key metrics
echo "2. Key Metrics (last 24h):"
curl -s "http://prometheus:9090/api/v1/query?query=rate(yfin_scrape_requests_total[24h])" | jq .

# Check error rates
echo "3. Error Rates:"
curl -s "http://prometheus:9090/api/v1/query?query=rate(yfin_scrape_errors_total[24h])" | jq .

# Test key endpoints
echo "4. Endpoint Tests:"
yfin scrape --ticker AAPL --endpoint key-statistics --check
yfin scrape --ticker MSFT --endpoint financials --check
yfin scrape --ticker GOOGL --endpoint news --check
```

### Performance Monitoring

```bash
#!/bin/bash
# performance-monitor.sh

# Monitor response times
echo "Response Time Analysis:"
curl -s "http://prometheus:9090/api/v1/query?query=histogram_quantile(0.95, rate(yfin_scrape_request_duration_seconds_bucket[1h]))" | jq .

# Monitor throughput
echo "Throughput Analysis:"
curl -s "http://prometheus:9090/api/v1/query?query=rate(yfin_scrape_requests_total[1h])" | jq .

# Monitor resource usage
echo "Resource Usage:"
curl -s "http://prometheus:9090/api/v1/query?query=process_resident_memory_bytes" | jq .
```

This observability guide provides comprehensive monitoring coverage for the scrape fallback system. Use these metrics, logs, and dashboards to maintain operational excellence and quickly identify issues.
