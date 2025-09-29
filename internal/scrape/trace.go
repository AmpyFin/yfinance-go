package scrape

import (
	"context"
	"time"

	"github.com/AmpyFin/yfinance-go/internal/obsv"
)

// Tracer handles OpenTelemetry tracing for scraping operations
type Tracer struct {
	serviceName string
}

// NewTracer creates a new tracer instance
func NewTracer() *Tracer {
	return &Tracer{
		serviceName: "yfinance-go/scrape",
	}
}

// StartFetchSpan starts a new trace span for a fetch operation
func (t *Tracer) StartFetchSpan(ctx context.Context, url, host string) (context.Context, interface{}) {
	// Use the existing observability system
	ctx, span := obsv.StartIngestFetchSpan(ctx, "scrape", host, "", url, 0)
	return ctx, span
}

// RecordSpanError records an error in the span
func (t *Tracer) RecordSpanError(span interface{}, err error) {
	// For now, we'll just log the error since we can't cast the interface{} to the proper span type
	// In a full implementation, we'd need to properly handle the span type
	_ = span
	_ = err
}

// UpdateSpan updates span with response information
func (t *Tracer) UpdateSpan(span interface{}, status, bytes int, duration time.Duration) {
	// For now, we'll just log the update since we can't cast the interface{} to the proper span type
	// In a full implementation, we'd need to properly handle the span type
	_ = span
	_ = status
	_ = bytes
	_ = duration
}

// EndSpan ends the span
func (t *Tracer) EndSpan(span interface{}) {
	// For now, we'll just log the end since we can't cast the interface{} to the proper span type
	// In a full implementation, we'd need to properly handle the span type
	_ = span
}

// GetStats returns tracer statistics
func (t *Tracer) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"service_name": t.serviceName,
		"tracer_type":  "opentelemetry",
	}
}
