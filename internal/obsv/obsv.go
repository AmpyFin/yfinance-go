package obsv

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/AmpyFin/ampy-observability/go/ampyobs"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

// Config represents observability configuration
type Config struct {
	ServiceName       string
	ServiceVersion    string
	Environment       string
	CollectorEndpoint string
	TraceProtocol     string
	SampleRatio       float64
	LogLevel          string
	MetricsAddr       string
	MetricsEnabled    bool
	TracingEnabled    bool
}

// PrometheusConfig for Prometheus exporter
type PrometheusConfig struct {
	Enabled bool
	Addr    string
}

// Observability represents the main observability interface
type Observability struct {
	config   *Config
	ampyConfig ampyobs.Config
	initialized bool
}

// Global observability instance
var (
	globalObsv *Observability
	globalMux  sync.RWMutex
)

// Init initializes the observability system using ampy-observability
func Init(ctx context.Context, config *Config) error {
	globalMux.Lock()
	defer globalMux.Unlock()
	
	if globalObsv != nil {
		return fmt.Errorf("observability already initialized")
	}

	// Create ampy-observability config
	ampyConfig := ampyobs.Config{
		ServiceName:       config.ServiceName,
		ServiceVersion:    config.ServiceVersion,
		Environment:       config.Environment,
		CollectorEndpoint: config.CollectorEndpoint,
		TraceProtocol:     config.TraceProtocol,
	}

	// Initialize ampy-observability
	if err := ampyobs.Init(ampyConfig); err != nil {
		return fmt.Errorf("failed to initialize ampy-observability: %w", err)
	}

	// Initialize metrics
	if config.MetricsEnabled {
		if err := initMetrics(PrometheusConfig{
			Enabled: config.MetricsEnabled,
			Addr:    config.MetricsAddr,
		}); err != nil {
			Logger().Error("Failed to initialize Prometheus metrics", "error", err)
			return fmt.Errorf("failed to init metrics: %w", err)
		}
	}

	globalObsv = &Observability{
		config:     config,
		ampyConfig: ampyConfig,
		initialized: true,
	}

	return nil
}

// Shutdown shuts down the observability system
func Shutdown(ctx context.Context) error {
	globalMux.Lock()
	defer globalMux.Unlock()
	
	if globalObsv == nil {
		return nil
	}

	// Shutdown metrics server
	if err := shutdownMetrics(ctx); err != nil {
		Logger().Error("Failed to shutdown metrics server", "error", err)
	}

	err := ampyobs.Shutdown(ctx)
	globalObsv = nil
	return err
}

// Reset resets the global observability state (for testing)
func Reset() {
	globalMux.Lock()
	defer globalMux.Unlock()
	globalObsv = nil
}

// Logger returns the ampy-observability logger
func Logger() *slog.Logger {
	globalMux.RLock()
	defer globalMux.RUnlock()
	
	if globalObsv == nil || !globalObsv.initialized {
		return slog.Default()
	}
	return ampyobs.L()
}

// Tracer returns the ampy-observability tracer
func Tracer() trace.Tracer {
	globalMux.RLock()
	defer globalMux.RUnlock()
	
	if globalObsv == nil || !globalObsv.initialized {
		return noop.NewTracerProvider().Tracer("yfinance-go")
	}
	// ampy-observability doesn't expose tracer directly, use context logger
	return noop.NewTracerProvider().Tracer("yfinance-go")
}

// StartSpan creates a new span using ampy-observability
func StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	// Use ampy-observability StartSpan with proper signature
	return ampyobs.StartSpan(ctx, name, trace.SpanKindInternal)
}

// Metrics helpers are now implemented in metrics.go

// Span names for different operations
const (
	SpanNameRun           = "yfin.run"
	SpanNameIngestFetch   = "ingest.fetch"
	SpanNameIngestDecode  = "ingest.decode"
	SpanNameIngestNormalize = "ingest.normalize"
	SpanNameEmitProto     = "emit.proto"
	SpanNamePublishBus    = "publish.bus"
	SpanNameFXRates       = "fx.rates"
)

// StartRunSpan creates the root span for a yfin run
func StartRunSpan(ctx context.Context, runID, env string, args []string) (context.Context, trace.Span) {
	attrs := []attribute.KeyValue{
		attribute.String("run_id", runID),
		attribute.String("env", env),
		attribute.StringSlice("args", args),
	}
	
	return StartSpan(ctx, SpanNameRun, trace.WithAttributes(attrs...))
}

// StartIngestFetchSpan creates a span for HTTP fetch operations
func StartIngestFetchSpan(ctx context.Context, endpoint, symbol, mic, url string, attempt int) (context.Context, trace.Span) {
	attrs := []attribute.KeyValue{
		attribute.String("endpoint", endpoint),
		attribute.String("symbol", symbol),
		attribute.String("mic", mic),
		attribute.String("url", url),
		attribute.Int("attempt", attempt),
	}
	
	return StartSpan(ctx, SpanNameIngestFetch, trace.WithAttributes(attrs...))
}

// UpdateIngestFetchSpan updates the fetch span with response details
func UpdateIngestFetchSpan(span trace.Span, status int, bytes int64, elapsed time.Duration) {
	attrs := []attribute.KeyValue{
		attribute.Int("status", status),
		attribute.Int64("bytes", bytes),
		attribute.Int64("elapsed_ms", elapsed.Milliseconds()),
	}
	
	span.SetAttributes(attrs...)
}

// StartIngestDecodeSpan creates a span for decode operations
func StartIngestDecodeSpan(ctx context.Context, endpoint, symbol string) (context.Context, trace.Span) {
	attrs := []attribute.KeyValue{
		attribute.String("endpoint", endpoint),
		attribute.String("symbol", symbol),
	}
	
	return StartSpan(ctx, SpanNameIngestDecode, trace.WithAttributes(attrs...))
}

// StartIngestNormalizeSpan creates a span for normalization operations
func StartIngestNormalizeSpan(ctx context.Context, endpoint, symbol, mic string) (context.Context, trace.Span) {
	attrs := []attribute.KeyValue{
		attribute.String("endpoint", endpoint),
		attribute.String("symbol", symbol),
		attribute.String("mic", mic),
	}
	
	return StartSpan(ctx, SpanNameIngestNormalize, trace.WithAttributes(attrs...))
}

// StartEmitProtoSpan creates a span for protobuf emission
func StartEmitProtoSpan(ctx context.Context, messageType, symbol string) (context.Context, trace.Span) {
	attrs := []attribute.KeyValue{
		attribute.String("message_type", messageType),
		attribute.String("symbol", symbol),
	}
	
	return StartSpan(ctx, SpanNameEmitProto, trace.WithAttributes(attrs...))
}

// StartPublishBusSpan creates a span for bus publishing
func StartPublishBusSpan(ctx context.Context, topic, partitionKey string, chunkIndex int, bytes int64) (context.Context, trace.Span) {
	attrs := []attribute.KeyValue{
		attribute.String("topic", topic),
		attribute.String("partition_key", partitionKey),
		attribute.Int("chunk_index", chunkIndex),
		attribute.Int64("bytes", bytes),
	}
	
	return StartSpan(ctx, SpanNamePublishBus, trace.WithAttributes(attrs...))
}

// StartFXRatesSpan creates a span for FX rate operations
func StartFXRatesSpan(ctx context.Context, fromCurrency, toCurrency string) (context.Context, trace.Span) {
	attrs := []attribute.KeyValue{
		attribute.String("from_currency", fromCurrency),
		attribute.String("to_currency", toCurrency),
	}
	
	return StartSpan(ctx, SpanNameFXRates, trace.WithAttributes(attrs...))
}

// RecordSpanError records an error in a span
func RecordSpanError(span trace.Span, err error) {
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
}

// LogWithTrace adds trace context to log attributes
func LogWithTrace(ctx context.Context, attrs ...any) []any {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		attrs = append(attrs, "trace_id", span.SpanContext().TraceID().String())
		attrs = append(attrs, "span_id", span.SpanContext().SpanID().String())
	}
	return attrs
}

// Common log attributes for yfinance-go
func CommonLogAttrs(runID, symbol, mic, endpoint string) []any {
	attrs := []any{
		"source", "yfinance-go",
	}
	
	if runID != "" {
		attrs = append(attrs, "run_id", runID)
	}
	if symbol != "" {
		attrs = append(attrs, "symbol", symbol)
	}
	if mic != "" {
		attrs = append(attrs, "mic", mic)
	}
	if endpoint != "" {
		attrs = append(attrs, "endpoint", endpoint)
	}
	
	return attrs
}
