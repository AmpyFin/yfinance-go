package obsv

import (
	"context"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Prometheus metrics following Step 10 specifications
var (
	// Counters
	requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "yfin_requests_total",
			Help: "Total number of requests.",
		},
		[]string{"endpoint", "outcome", "code"},
	)

	retriesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "yfin_retries_total",
			Help: "Total number of retries.",
		},
		[]string{"endpoint", "reason"},
	)

	backoffTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "yfin_backoff_total",
			Help: "Total number of backoffs.",
		},
		[]string{"endpoint", "reason"},
	)

	decodeFailTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "yfin_decode_fail_total",
			Help: "Total number of decode failures.",
		},
		[]string{"reason"},
	)

	cbOpenTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "yfin_cb_open_total",
			Help: "Total number of circuit breaker opens.",
		},
		[]string{"scope"},
	)

	sessionEjectTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "yfin_session_eject_total",
			Help: "Total number of session ejections.",
		},
	)

	publishTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "yfin_publish_total",
			Help: "Total number of publish operations.",
		},
		[]string{"type", "outcome"},
	)

	// Gauges
	inflightRequests = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "yfin_inflight_requests",
			Help: "Number of inflight requests.",
		},
		[]string{"endpoint"},
	)

	cbState = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "yfin_cb_state",
			Help: "Circuit breaker state (0=closed, 1=half-open, 2=open).",
		},
		[]string{"scope"},
	)

	// Histograms
	requestLatencyMs = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "yfin_request_latency_ms",
			Help:    "Request latency in milliseconds.",
			Buckets: []float64{10, 25, 50, 100, 200, 500, 1000, 2000},
		},
		[]string{"endpoint"},
	)

	backoffSleepMs = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "yfin_backoff_sleep_ms",
			Help:    "Backoff sleep duration in milliseconds.",
			Buckets: []float64{10, 25, 50, 100, 200, 500, 1000, 2000, 5000, 10000},
		},
		[]string{"endpoint"},
	)

	batchBytes = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "yfin_batch_bytes",
			Help:    "Batch size in bytes.",
			Buckets: []float64{1024, 4096, 16384, 65536, 262144, 1048576, 4194304},
		},
		[]string{"type"},
	)

	publishLatencyMs = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "yfin_publish_latency_ms",
			Help:    "Publish latency in milliseconds.",
			Buckets: []float64{10, 25, 50, 100, 200, 500, 1000, 2000},
		},
		[]string{"type"},
	)
)

var (
	metricsRegistered = false
	metricsServer     *http.Server
)

// initMetrics initializes Prometheus metrics and starts the HTTP server for exposition.
func initMetrics(cfg PrometheusConfig) error {
	if !cfg.Enabled {
		return nil
	}

	// Register metrics only once
	if !metricsRegistered {
		prometheus.MustRegister(
			requestsTotal,
			retriesTotal,
			backoffTotal,
			decodeFailTotal,
			cbOpenTotal,
			sessionEjectTotal,
			publishTotal,
			inflightRequests,
			cbState,
			requestLatencyMs,
			backoffSleepMs,
			batchBytes,
			publishLatencyMs,
		)
		metricsRegistered = true
	}

	// Start HTTP server for Prometheus metrics
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	metricsServer = &http.Server{
		Addr:    cfg.Addr,
		Handler: mux,
	}

	go func() {
		Logger().Info("Prometheus metrics exporter started", "addr", cfg.Addr)
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			Logger().Error("Failed to start Prometheus metrics exporter", "error", err, "addr", cfg.Addr)
		}
	}()

	return nil
}

// shutdownMetrics shuts down the metrics server
func shutdownMetrics(ctx context.Context) error {
	if metricsServer != nil {
		return metricsServer.Shutdown(ctx)
	}
	return nil
}

// Metrics recording functions following Step 10 specifications

func RecordRequest(endpoint, outcome, code string) {
	if globalObsv == nil || !globalObsv.config.MetricsEnabled {
		return
	}
	requestsTotal.WithLabelValues(endpoint, outcome, code).Inc()
}

func RecordRequestLatency(endpoint string, duration time.Duration) {
	if globalObsv == nil || !globalObsv.config.MetricsEnabled {
		return
	}
	requestLatencyMs.WithLabelValues(endpoint).Observe(float64(duration.Milliseconds()))
}

func RecordRetry(endpoint, reason string) {
	if globalObsv == nil || !globalObsv.config.MetricsEnabled {
		return
	}
	retriesTotal.WithLabelValues(endpoint, reason).Inc()
}

func RecordBackoff(endpoint, reason string) {
	if globalObsv == nil || !globalObsv.config.MetricsEnabled {
		return
	}
	backoffTotal.WithLabelValues(endpoint, reason).Inc()
}

func RecordBackoffSleep(endpoint string, duration time.Duration) {
	if globalObsv == nil || !globalObsv.config.MetricsEnabled {
		return
	}
	backoffSleepMs.WithLabelValues(endpoint).Observe(float64(duration.Milliseconds()))
}

func RecordCBOpen(scope string) {
	if globalObsv == nil || !globalObsv.config.MetricsEnabled {
		return
	}
	cbOpenTotal.WithLabelValues(scope).Inc()
}

func SetCBState(scope string, state int) {
	if globalObsv == nil || !globalObsv.config.MetricsEnabled {
		return
	}
	cbState.WithLabelValues(scope).Set(float64(state))
}

func RecordDecodeFail(reason string) {
	if globalObsv == nil || !globalObsv.config.MetricsEnabled {
		return
	}
	decodeFailTotal.WithLabelValues(reason).Inc()
}

func RecordSessionEject() {
	if globalObsv == nil || !globalObsv.config.MetricsEnabled {
		return
	}
	sessionEjectTotal.Inc()
}

func SetInflightRequests(endpoint string, count int) {
	if globalObsv == nil || !globalObsv.config.MetricsEnabled {
		return
	}
	inflightRequests.WithLabelValues(endpoint).Set(float64(count))
}

func RecordPublish(publishType, outcome string) {
	if globalObsv == nil || !globalObsv.config.MetricsEnabled {
		return
	}
	publishTotal.WithLabelValues(publishType, outcome).Inc()
}

func RecordPublishLatency(publishType string, duration time.Duration) {
	if globalObsv == nil || !globalObsv.config.MetricsEnabled {
		return
	}
	publishLatencyMs.WithLabelValues(publishType).Observe(float64(duration.Milliseconds()))
}

func RecordBatchBytes(batchType string, bytes int64) {
	if globalObsv == nil || !globalObsv.config.MetricsEnabled {
		return
	}
	batchBytes.WithLabelValues(batchType).Observe(float64(bytes))
}
