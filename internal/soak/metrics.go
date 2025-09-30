package soak

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds soak test specific metrics
type Metrics struct {
	// Request metrics
	RequestsTotal     *prometheus.CounterVec
	RequestDuration   *prometheus.HistogramVec
	RequestsInFlight  *prometheus.GaugeVec
	
	// Fallback metrics
	FallbackDecisions *prometheus.CounterVec
	
	// Error metrics
	ErrorsTotal       *prometheus.CounterVec
	RateLimitHits     *prometheus.CounterVec
	RobotsBlocked     *prometheus.CounterVec
	
	// Correctness metrics
	ProbesTotal       *prometheus.CounterVec
	ProbeLatency      *prometheus.HistogramVec
	
	// Memory metrics
	MemoryUsage       prometheus.Gauge
	GoroutineCount    prometheus.Gauge
	
	// Worker metrics
	WorkerUtilization *prometheus.GaugeVec
	
	mu sync.RWMutex
}

// NewMetrics creates a new metrics instance
func NewMetrics() *Metrics {
	return &Metrics{
		RequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "yfin_soak_requests_total",
				Help: "Total number of requests made during soak test",
			},
			[]string{"endpoint", "ticker", "outcome", "source"},
		),
		
		RequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "yfin_soak_request_duration_seconds",
				Help:    "Request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"endpoint", "source"},
		),
		
		RequestsInFlight: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "yfin_soak_requests_in_flight",
				Help: "Number of requests currently in flight",
			},
			[]string{"endpoint"},
		),
		
		FallbackDecisions: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "yfin_soak_fallback_decisions_total",
				Help: "Total number of fallback decisions made",
			},
			[]string{"from_source", "to_source", "reason"},
		),
		
		ErrorsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "yfin_soak_errors_total",
				Help: "Total number of errors encountered",
			},
			[]string{"endpoint", "error_type", "source"},
		),
		
		RateLimitHits: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "yfin_soak_rate_limit_hits_total",
				Help: "Total number of rate limit hits",
			},
			[]string{"host", "endpoint"},
		),
		
		RobotsBlocked: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "yfin_soak_robots_blocked_total",
				Help: "Total number of requests blocked by robots.txt",
			},
			[]string{"host", "path"},
		),
		
		ProbesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "yfin_soak_probes_total",
				Help: "Total number of correctness probes executed",
			},
			[]string{"ticker", "outcome"},
		),
		
		ProbeLatency: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "yfin_soak_probe_duration_seconds",
				Help:    "Correctness probe duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"ticker", "probe_type"},
		),
		
		MemoryUsage: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "yfin_soak_memory_usage_bytes",
				Help: "Current memory usage in bytes",
			},
		),
		
		GoroutineCount: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "yfin_soak_goroutines_count",
				Help: "Current number of goroutines",
			},
		),
		
		WorkerUtilization: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "yfin_soak_worker_utilization",
				Help: "Worker utilization percentage",
			},
			[]string{"worker_id"},
		),
	}
}

// RecordRequest records a request metric
func (m *Metrics) RecordRequest(endpoint, ticker, outcome, source string, duration time.Duration) {
	m.RequestsTotal.WithLabelValues(endpoint, ticker, outcome, source).Inc()
	m.RequestDuration.WithLabelValues(endpoint, source).Observe(duration.Seconds())
}

// RecordFallback records a fallback decision
func (m *Metrics) RecordFallback(fromSource, toSource, reason string) {
	m.FallbackDecisions.WithLabelValues(fromSource, toSource, reason).Inc()
}

// RecordError records an error
func (m *Metrics) RecordError(endpoint, errorType, source string) {
	m.ErrorsTotal.WithLabelValues(endpoint, errorType, source).Inc()
}

// RecordRateLimit records a rate limit hit
func (m *Metrics) RecordRateLimit(host, endpoint string) {
	m.RateLimitHits.WithLabelValues(host, endpoint).Inc()
}

// RecordRobotsBlocked records a robots.txt block
func (m *Metrics) RecordRobotsBlocked(host, path string) {
	m.RobotsBlocked.WithLabelValues(host, path).Inc()
}

// RecordProbe records a correctness probe
func (m *Metrics) RecordProbe(ticker, outcome string, duration time.Duration, probeType string) {
	m.ProbesTotal.WithLabelValues(ticker, outcome).Inc()
	m.ProbeLatency.WithLabelValues(ticker, probeType).Observe(duration.Seconds())
}

// UpdateMemoryUsage updates memory usage metric
func (m *Metrics) UpdateMemoryUsage(bytes uint64) {
	m.MemoryUsage.Set(float64(bytes))
}

// UpdateGoroutineCount updates goroutine count metric
func (m *Metrics) UpdateGoroutineCount(count int) {
	m.GoroutineCount.Set(float64(count))
}

// UpdateWorkerUtilization updates worker utilization
func (m *Metrics) UpdateWorkerUtilization(workerID string, utilization float64) {
	m.WorkerUtilization.WithLabelValues(workerID).Set(utilization)
}

// SetRequestsInFlight sets the number of in-flight requests
func (m *Metrics) SetRequestsInFlight(endpoint string, count int) {
	m.RequestsInFlight.WithLabelValues(endpoint).Set(float64(count))
}
