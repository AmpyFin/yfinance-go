package obsv

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestObservabilityIntegration tests the full observability integration
func TestObservabilityIntegration(t *testing.T) {
	// Reset global state
	Reset()
	
	// Initialize observability with all features enabled
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
		MetricsEnabled:    true,
		TracingEnabled:    true,
	}
	
	err := Init(ctx, config)
	require.NoError(t, err)
	defer func() {
		_ = Shutdown(ctx)
	}()
	
	// Test that observability is properly initialized
	assert.NotNil(t, globalObsv)
	assert.True(t, globalObsv.initialized)
	assert.Equal(t, config, globalObsv.config)
	
	// Test logger
	logger := Logger()
	assert.NotNil(t, logger)
	
	// Test tracer
	tracer := Tracer()
	assert.NotNil(t, tracer)
	
	// Test span creation
	ctx, span := StartSpan(ctx, "test.operation")
	assert.NotNil(t, span)
	assert.NotNil(t, ctx)
	span.End()
	
	// Test metrics recording (should not panic)
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

// TestObservabilityDisabled tests that observability works when disabled
func TestObservabilityDisabled(t *testing.T) {
	// Reset global state
	Reset()
	
	// Initialize observability with features disabled
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
		MetricsEnabled:    false, // Disabled
		TracingEnabled:    false, // Disabled
	}
	
	err := Init(ctx, config)
	require.NoError(t, err)
	defer func() {
		_ = Shutdown(ctx)
	}()
	
	// Test that observability is initialized but features are disabled
	assert.NotNil(t, globalObsv)
	assert.False(t, globalObsv.config.MetricsEnabled)
	assert.False(t, globalObsv.config.TracingEnabled)
	
	// Test that metrics functions are no-ops when disabled
	// These should not panic even when metrics are disabled
	RecordRequest("bars_1d", "success", "200")
	RecordRequestLatency("bars_1d", 100*time.Millisecond)
	RecordPublish("bars", "ack")
	RecordPublishLatency("bars", 50*time.Millisecond)
}

// TestObservabilityNotInitialized tests that functions work when not initialized
func TestObservabilityNotInitialized(t *testing.T) {
	// Reset global state
	Reset()
	
	// Test that functions work when not initialized (should be no-ops)
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
	
	// Test that logger and tracer return fallback implementations
	logger := Logger()
	assert.NotNil(t, logger)
	
	tracer := Tracer()
	assert.NotNil(t, tracer)
	
	// Test that span creation works with noop tracer
	ctx := context.Background()
	ctx, span := StartSpan(ctx, "test.operation")
	assert.NotNil(t, span)
	assert.NotNil(t, ctx)
	span.End()
}

// TestSpanHierarchy tests the span hierarchy for yfinance-go operations
func TestSpanHierarchy(t *testing.T) {
	// Reset global state
	Reset()
	
	// Initialize observability
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
		MetricsEnabled:    false,
		TracingEnabled:    true,
	}
	
	err := Init(ctx, config)
	require.NoError(t, err)
	defer func() {
		_ = Shutdown(ctx)
	}()
	
	// Test the complete span hierarchy
	ctx, runSpan := StartRunSpan(ctx, "test-run-123", "test", []string{"arg1", "arg2"})
	defer runSpan.End()
	
	ctx, fetchSpan := StartIngestFetchSpan(ctx, "bars_1d", "AAPL", "XNAS", "https://example.com", 1)
	defer fetchSpan.End()
	
	UpdateIngestFetchSpan(fetchSpan, 200, 1024, 100*time.Millisecond)
	
	ctx, decodeSpan := StartIngestDecodeSpan(ctx, "bars_1d", "AAPL")
	defer decodeSpan.End()
	
	ctx, normalizeSpan := StartIngestNormalizeSpan(ctx, "bars_1d", "AAPL", "XNAS")
	defer normalizeSpan.End()
	
	ctx, emitSpan := StartEmitProtoSpan(ctx, "bars", "AAPL")
	defer emitSpan.End()
	
	ctx, publishSpan := StartPublishBusSpan(ctx, "ampy.bars", "AAPL", 0, 1024)
	defer publishSpan.End()
	
	_, fxSpan := StartFXRatesSpan(ctx, "USD", "EUR")
	defer fxSpan.End()
	
	// Test error recording
	RecordSpanError(fetchSpan, assert.AnError)
	RecordSpanError(decodeSpan, assert.AnError)
	
	// All spans should be created successfully
	assert.NotNil(t, runSpan)
	assert.NotNil(t, fetchSpan)
	assert.NotNil(t, decodeSpan)
	assert.NotNil(t, normalizeSpan)
	assert.NotNil(t, emitSpan)
	assert.NotNil(t, publishSpan)
	assert.NotNil(t, fxSpan)
}

// TestLoggingIntegration tests the logging integration
func TestLoggingIntegration(t *testing.T) {
	// Reset global state
	Reset()
	
	// Initialize observability
	ctx := context.Background()
	config := &Config{
		ServiceName:       "test-service",
		ServiceVersion:    "1.0.0",
		Environment:       "test",
		CollectorEndpoint: "localhost:4317",
		TraceProtocol:     "grpc",
		SampleRatio:       0.1,
		LogLevel:          "debug",
		MetricsAddr:       ":9090",
		MetricsEnabled:    false,
		TracingEnabled:    true,
	}
	
	err := Init(ctx, config)
	require.NoError(t, err)
	defer func() {
		_ = Shutdown(ctx)
	}()
	
	// Test logging with trace context
	ctx, span := StartSpan(ctx, "test.operation")
	defer span.End()
	
	// Test common log attributes
	attrs := CommonLogAttrs("test-run-123", "AAPL", "XNAS", "bars_1d")
	assert.Contains(t, attrs, "source")
	assert.Contains(t, attrs, "yfinance-go")
	assert.Contains(t, attrs, "run_id")
	assert.Contains(t, attrs, "test-run-123")
	assert.Contains(t, attrs, "symbol")
	assert.Contains(t, attrs, "AAPL")
	assert.Contains(t, attrs, "mic")
	assert.Contains(t, attrs, "XNAS")
	assert.Contains(t, attrs, "endpoint")
	assert.Contains(t, attrs, "bars_1d")
	
	// Test log with trace context
	logAttrs := LogWithTrace(ctx, "key1", "value1", "key2", "value2")
	assert.GreaterOrEqual(t, len(logAttrs), 4) // At least the original attrs
	
	// Test empty log attributes
	emptyAttrs := CommonLogAttrs("", "", "", "")
	assert.Contains(t, emptyAttrs, "source")
	assert.Contains(t, emptyAttrs, "yfinance-go")
	assert.Len(t, emptyAttrs, 2) // Only source and value
}
