package obsv

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	ctx := context.Background()
	
	config := &Config{
		ServiceName:       "test-service",
		ServiceVersion:    "1.0.0",
		Environment:       "test",
		CollectorEndpoint: "localhost:4317",
		TraceProtocol:     "grpc",
		SampleRatio:       0.1,
		LogLevel:          "info",
		MetricsAddr:       ":9090",
		MetricsEnabled:    false, // Disable for testing
		TracingEnabled:    false, // Disable for testing
	}
	
	err := Init(ctx, config)
	require.NoError(t, err)
	
	// Test that global instance is set
	assert.NotNil(t, globalObsv)
	assert.Equal(t, config, globalObsv.config)
	
	// Test shutdown
	err = Shutdown(ctx)
	require.NoError(t, err)
}

func TestLogger(t *testing.T) {
	// Test fallback logger when not initialized
	logger := Logger()
	assert.NotNil(t, logger)
}

func TestTracer(t *testing.T) {
	// Test fallback tracer when not initialized
	tracer := Tracer()
	assert.NotNil(t, tracer)
}

func TestStartSpan(t *testing.T) {
	ctx := context.Background()
	
	// Test with noop tracer
	ctx, span := StartSpan(ctx, "test.operation")
	assert.NotNil(t, span)
	assert.NotNil(t, ctx)
	
	span.End()
}

func TestStartRunSpan(t *testing.T) {
	ctx := context.Background()
	
	ctx, span := StartRunSpan(ctx, "test-run-123", "test", []string{"arg1", "arg2"})
	assert.NotNil(t, span)
	assert.NotNil(t, ctx)
	
	span.End()
}

func TestStartIngestFetchSpan(t *testing.T) {
	ctx := context.Background()
	
	ctx, span := StartIngestFetchSpan(ctx, "bars_1d", "AAPL", "XNAS", "https://example.com", 1)
	assert.NotNil(t, span)
	assert.NotNil(t, ctx)
	
	span.End()
}

func TestUpdateIngestFetchSpan(t *testing.T) {
	ctx := context.Background()
	ctx, span := StartIngestFetchSpan(ctx, "bars_1d", "AAPL", "XNAS", "https://example.com", 1)
	defer span.End()
	
	UpdateIngestFetchSpan(span, 200, 1024, 100*time.Millisecond)
	// No assertion needed as this just sets attributes
}

func TestStartIngestDecodeSpan(t *testing.T) {
	ctx := context.Background()
	
	ctx, span := StartIngestDecodeSpan(ctx, "bars_1d", "AAPL")
	assert.NotNil(t, span)
	assert.NotNil(t, ctx)
	
	span.End()
}

func TestStartIngestNormalizeSpan(t *testing.T) {
	ctx := context.Background()
	
	ctx, span := StartIngestNormalizeSpan(ctx, "bars_1d", "AAPL", "XNAS")
	assert.NotNil(t, span)
	assert.NotNil(t, ctx)
	
	span.End()
}

func TestStartEmitProtoSpan(t *testing.T) {
	ctx := context.Background()
	
	ctx, span := StartEmitProtoSpan(ctx, "bars", "AAPL")
	assert.NotNil(t, span)
	assert.NotNil(t, ctx)
	
	span.End()
}

func TestStartPublishBusSpan(t *testing.T) {
	ctx := context.Background()
	
	ctx, span := StartPublishBusSpan(ctx, "ampy.bars", "AAPL", 0, 1024)
	assert.NotNil(t, span)
	assert.NotNil(t, ctx)
	
	span.End()
}

func TestStartFXRatesSpan(t *testing.T) {
	ctx := context.Background()
	
	ctx, span := StartFXRatesSpan(ctx, "USD", "EUR")
	assert.NotNil(t, span)
	assert.NotNil(t, ctx)
	
	span.End()
}

func TestRecordSpanError(t *testing.T) {
	ctx := context.Background()
	ctx, span := StartSpan(ctx, "test.operation")
	defer span.End()
	
	// Test with nil error
	RecordSpanError(span, nil)
	
	// Test with error
	err := assert.AnError
	RecordSpanError(span, err)
}

func TestLogWithTrace(t *testing.T) {
	ctx := context.Background()
	ctx, span := StartSpan(ctx, "test.operation")
	defer span.End()
	
	attrs := []any{"key1", "value1", "key2", "value2"}
	result := LogWithTrace(ctx, attrs...)
	
	// Should have added trace_id and span_id (or not if noop tracer)
	assert.GreaterOrEqual(t, len(result), len(attrs)) // At least the original attrs
}

func TestCommonLogAttrs(t *testing.T) {
	attrs := CommonLogAttrs("run-123", "AAPL", "XNAS", "bars_1d")
	
	// Should contain source and provided values
	assert.Contains(t, attrs, "source")
	assert.Contains(t, attrs, "yfinance-go")
	assert.Contains(t, attrs, "run_id")
	assert.Contains(t, attrs, "run-123")
	assert.Contains(t, attrs, "symbol")
	assert.Contains(t, attrs, "AAPL")
	assert.Contains(t, attrs, "mic")
	assert.Contains(t, attrs, "XNAS")
	assert.Contains(t, attrs, "endpoint")
	assert.Contains(t, attrs, "bars_1d")
}

func TestCommonLogAttrsEmpty(t *testing.T) {
	attrs := CommonLogAttrs("", "", "", "")
	
	// Should only contain source
	assert.Contains(t, attrs, "source")
	assert.Contains(t, attrs, "yfinance-go")
	assert.Len(t, attrs, 2) // Only source and value
}

func TestMetricsFunctions(t *testing.T) {
	// Test that metrics functions don't panic when not initialized
	// (they should be no-ops when globalObsv is nil)
	RecordRequest("bars_1d", "success", "200")
	RecordRequestLatency("bars_1d", 100*time.Millisecond)
	RecordRetry("bars_1d", "network_error")
	RecordBackoff("bars_1d", "retry")
	RecordBackoffSleep("bars_1d", 250*time.Millisecond)
	RecordCBOpen("host")
	SetCBState("host", 1)
	RecordDecodeFail("json_parse")
	RecordSessionEject()
	SetInflightRequests("bars_1d", 5)
	RecordPublish("bars", "ack")
	RecordPublishLatency("bars", 50*time.Millisecond)
	RecordBatchBytes("bars", 1024)
}

func TestSpanNames(t *testing.T) {
	// Test that span name constants are defined
	assert.Equal(t, "yfin.run", SpanNameRun)
	assert.Equal(t, "ingest.fetch", SpanNameIngestFetch)
	assert.Equal(t, "ingest.decode", SpanNameIngestDecode)
	assert.Equal(t, "ingest.normalize", SpanNameIngestNormalize)
	assert.Equal(t, "emit.proto", SpanNameEmitProto)
	assert.Equal(t, "publish.bus", SpanNamePublishBus)
	assert.Equal(t, "fx.rates", SpanNameFXRates)
}
