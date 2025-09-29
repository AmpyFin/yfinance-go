package scrape

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics handles Prometheus metrics for scraping operations
type Metrics struct {
	requestsTotal     *prometheus.CounterVec
	retriesTotal      *prometheus.CounterVec
	robotsDeniedTotal *prometheus.CounterVec
	fetchLatency      *prometheus.HistogramVec
	pageBytes         *prometheus.HistogramVec
	inflightGauge     *prometheus.GaugeVec
	backoffTotal      *prometheus.CounterVec
	backoffSleep      *prometheus.HistogramVec
	// News-specific metrics
	newsTotal         *prometheus.CounterVec
	newsParseLatency  *prometheus.HistogramVec
}

var (
	// Global metrics instances to avoid duplicate registration
	requestsTotal     *prometheus.CounterVec
	retriesTotal      *prometheus.CounterVec
	robotsDeniedTotal *prometheus.CounterVec
	fetchLatency      *prometheus.HistogramVec
	pageBytes         *prometheus.HistogramVec
	inflightGauge     *prometheus.GaugeVec
	backoffTotal      *prometheus.CounterVec
	backoffSleep      *prometheus.HistogramVec
	// News-specific global metrics
	newsTotal         *prometheus.CounterVec
	newsParseLatency  *prometheus.HistogramVec
	metricsOnce       sync.Once
)

// NewMetrics creates a new metrics instance
func NewMetrics() *Metrics {
	metricsOnce.Do(func() {
		requestsTotal = promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "yfin_scrape_requests_total",
				Help: "Total number of scraping requests",
			},
			[]string{"host", "outcome", "code"},
		)
		retriesTotal = promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "yfin_scrape_retries_total",
				Help: "Total number of retries",
			},
			[]string{"host", "reason"},
		)
		robotsDeniedTotal = promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "yfin_scrape_robots_denied_total",
				Help: "Total number of requests denied by robots.txt",
			},
			[]string{"host"},
		)
		fetchLatency = promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "yfin_scrape_fetch_latency_ms",
				Help:    "Latency of scraping requests in milliseconds",
				Buckets: prometheus.ExponentialBuckets(10, 2, 12), // 10ms to ~40s
			},
			[]string{"host"},
		)
		pageBytes = promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "yfin_scrape_page_bytes",
				Help:    "Size of scraped pages in bytes",
				Buckets: prometheus.ExponentialBuckets(1024, 2, 16), // 1KB to ~64MB
			},
			[]string{"host"},
		)
		inflightGauge = promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "yfin_scrape_inflight",
				Help: "Number of in-flight scraping requests",
			},
			[]string{"host"},
		)
		backoffTotal = promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "yfin_scrape_backoff_total",
				Help: "Total number of backoff events",
			},
			[]string{"host", "reason"},
		)
		backoffSleep = promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "yfin_scrape_backoff_sleep_ms",
				Help:    "Backoff sleep duration in milliseconds",
				Buckets: prometheus.ExponentialBuckets(1, 2, 16), // 1ms to ~32s
			},
			[]string{"host", "reason"},
		)
		newsTotal = promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "yfin_scrape_news_total",
				Help: "Total number of news parsing operations",
			},
			[]string{"outcome"},
		)
		newsParseLatency = promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "yfin_scrape_news_parse_latency_ms",
				Help:    "News parsing latency in milliseconds",
				Buckets: prometheus.ExponentialBuckets(1, 2, 12), // 1ms to ~4s
			},
			[]string{},
		)
	})

	return &Metrics{
		requestsTotal:     requestsTotal,
		retriesTotal:      retriesTotal,
		robotsDeniedTotal: robotsDeniedTotal,
		fetchLatency:      fetchLatency,
		pageBytes:         pageBytes,
		inflightGauge:     inflightGauge,
		backoffTotal:      backoffTotal,
		backoffSleep:      backoffSleep,
		newsTotal:         newsTotal,
		newsParseLatency:  newsParseLatency,
	}
}

// RecordRequest records a scraping request
func (m *Metrics) RecordRequest(host, outcome, code string) {
	m.requestsTotal.WithLabelValues(host, outcome, code).Inc()
}

// RecordRetry records a retry event
func (m *Metrics) RecordRetry(host, reason string) {
	m.retriesTotal.WithLabelValues(host, reason).Inc()
}

// RecordRobotsDenied records a robots.txt denial
func (m *Metrics) RecordRobotsDenied(host string) {
	m.robotsDeniedTotal.WithLabelValues(host).Inc()
}

// RecordLatency records request latency
func (m *Metrics) RecordLatency(host string, duration time.Duration) {
	m.fetchLatency.WithLabelValues(host).Observe(float64(duration.Milliseconds()))
}

// RecordPageBytes records page size
func (m *Metrics) RecordPageBytes(host string, bytes int) {
	m.pageBytes.WithLabelValues(host).Observe(float64(bytes))
}

// RecordInflight records in-flight requests
func (m *Metrics) RecordInflight(host string, count int) {
	m.inflightGauge.WithLabelValues(host).Set(float64(count))
}

// RecordBackoff records a backoff event
func (m *Metrics) RecordBackoff(host, reason string) {
	m.backoffTotal.WithLabelValues(host, reason).Inc()
}

// RecordBackoffSleep records backoff sleep duration
func (m *Metrics) RecordBackoffSleep(host, reason string, duration time.Duration) {
	m.backoffSleep.WithLabelValues(host, reason).Observe(float64(duration.Milliseconds()))
}

// RecordNews records a news parsing operation
func (m *Metrics) RecordNews(outcome string) {
	m.newsTotal.WithLabelValues(outcome).Inc()
}

// RecordNewsParseLatency records news parsing latency
func (m *Metrics) RecordNewsParseLatency(duration time.Duration) {
	m.newsParseLatency.WithLabelValues().Observe(float64(duration.Milliseconds()))
}

// GetStats returns current metrics statistics
func (m *Metrics) GetStats() map[string]interface{} {
	// This would typically collect current metric values
	// For now, return a placeholder
	return map[string]interface{}{
		"metrics_enabled": true,
		"collectors": []string{
			"yfin_scrape_requests_total",
			"yfin_scrape_retries_total",
			"yfin_scrape_robots_denied_total",
			"yfin_scrape_fetch_latency_ms",
			"yfin_scrape_page_bytes",
			"yfin_scrape_inflight",
			"yfin_scrape_backoff_total",
			"yfin_scrape_backoff_sleep_ms",
			"yfin_scrape_news_total",
			"yfin_scrape_news_parse_latency_ms",
		},
	}
}

// InflightTracker tracks in-flight requests per host
type InflightTracker struct {
	mu    sync.RWMutex
	count map[string]int
}

// NewInflightTracker creates a new in-flight tracker
func NewInflightTracker() *InflightTracker {
	return &InflightTracker{
		count: make(map[string]int),
	}
}

// Increment increments the in-flight count for a host
func (it *InflightTracker) Increment(host string) {
	it.mu.Lock()
	defer it.mu.Unlock()
	it.count[host]++
}

// Decrement decrements the in-flight count for a host
func (it *InflightTracker) Decrement(host string) {
	it.mu.Lock()
	defer it.mu.Unlock()
	if it.count[host] > 0 {
		it.count[host]--
	}
}

// GetCount returns the current in-flight count for a host
func (it *InflightTracker) GetCount(host string) int {
	it.mu.RLock()
	defer it.mu.RUnlock()
	return it.count[host]
}

// GetAllCounts returns all in-flight counts
func (it *InflightTracker) GetAllCounts() map[string]int {
	it.mu.RLock()
	defer it.mu.RUnlock()
	
	result := make(map[string]int)
	for host, count := range it.count {
		result[host] = count
	}
	return result
}
