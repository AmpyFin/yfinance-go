package emit

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	fundamentalsv1 "github.com/AmpyFin/ampy-proto/v2/gen/go/ampy/fundamentals/v1"
	newsv1 "github.com/AmpyFin/ampy-proto/v2/gen/go/ampy/news/v1"
	"github.com/AmpyFin/yfinance-go/internal/scrape"
)

// Metrics for proto emission
var (
	// Counter for total proto emissions by type and outcome
	protoEmissionTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "yfin_emit_proto_total",
			Help: "Total number of proto emissions by type and outcome",
		},
		[]string{"type", "outcome"},
	)

	// Histogram for proto message sizes in bytes
	protoEmissionBytes = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "yfin_emit_proto_bytes",
			Help:    "Size of emitted proto messages in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 2, 15), // 100B to ~3MB
		},
		[]string{"type"},
	)

	// Histogram for proto emission latency in milliseconds
	protoEmissionLatencyMs = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "yfin_emit_proto_latency_ms",
			Help:    "Latency of proto emission operations in milliseconds",
			Buckets: prometheus.ExponentialBuckets(1, 2, 12), // 1ms to ~4s
		},
		[]string{"type"},
	)

	// Counter for validation errors by type and field
	protoValidationErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "yfin_emit_proto_validation_errors_total",
			Help: "Total number of proto validation errors by type and field",
		},
		[]string{"type", "field"},
	)

	// Gauge for current mapping operations in progress
	protoMappingInProgress = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "yfin_emit_proto_mapping_in_progress",
			Help: "Number of proto mapping operations currently in progress",
		},
		[]string{"type"},
	)
)

// ObservableMapper wraps a ScrapeMapper with observability
type ObservableMapper struct {
	mapper *ScrapeMapper
	logger *slog.Logger
}

// NewObservableMapper creates a new observable mapper
func NewObservableMapper(mapper *ScrapeMapper, logger *slog.Logger) *ObservableMapper {
	if logger == nil {
		logger = slog.Default()
	}
	
	return &ObservableMapper{
		mapper: mapper,
		logger: logger,
	}
}

// MapFinancialsWithObservability maps financials with observability
func (o *ObservableMapper) MapFinancialsWithObservability(ctx context.Context, dto interface{}, runID, producer string) (interface{}, error) {
	start := time.Now()
	messageType := "fundamentals"
	
	// Increment in-progress gauge
	protoMappingInProgress.WithLabelValues(messageType).Inc()
	defer protoMappingInProgress.WithLabelValues(messageType).Dec()

	o.logger.InfoContext(ctx, "Starting fundamentals mapping",
		slog.String("run_id", runID),
		slog.String("producer", producer),
		slog.String("type", messageType))

	// Perform mapping based on DTO type
	var result interface{}
	var err error
	var resultSize int64

	switch v := dto.(type) {
	case *scrape.FinancialsDTO:
		snapshot, mapErr := MapFinancialsDTO(v, runID, producer)
		if mapErr != nil {
			err = mapErr
		} else {
			result = snapshot
			if snapshot != nil {
				resultSize = estimateProtoSize(snapshot)
			}
		}
	case *scrape.ComprehensiveFinancialsDTO:
		snapshots, mapErr := MapComprehensiveFinancialsDTO(v, runID, producer)
		if mapErr != nil {
			err = mapErr
		} else {
			result = snapshots
			for _, snapshot := range snapshots {
				resultSize += estimateProtoSize(snapshot)
			}
		}
	default:
		err = fmt.Errorf("unsupported financials DTO type: %T", dto)
	}

	// Record metrics
	duration := time.Since(start)
	outcome := "success"
	if err != nil {
		outcome = "error"
		o.logger.ErrorContext(ctx, "Fundamentals mapping failed",
			slog.String("run_id", runID),
			slog.String("error", err.Error()),
			slog.Duration("duration", duration))
	} else {
		o.logger.InfoContext(ctx, "Fundamentals mapping completed",
			slog.String("run_id", runID),
			slog.Int64("size_bytes", resultSize),
			slog.Duration("duration", duration))
	}

	protoEmissionTotal.WithLabelValues(messageType, outcome).Inc()
	protoEmissionLatencyMs.WithLabelValues(messageType).Observe(float64(duration.Nanoseconds()) / 1e6)
	
	if resultSize > 0 {
		protoEmissionBytes.WithLabelValues(messageType).Observe(float64(resultSize))
	}

	return result, err
}

// MapProfileWithObservability maps profile with observability
func (o *ObservableMapper) MapProfileWithObservability(ctx context.Context, dto *scrape.ComprehensiveProfileDTO, runID, producer string) (*ProfileMappingResult, error) {
	start := time.Now()
	messageType := "profile"
	
	// Increment in-progress gauge
	protoMappingInProgress.WithLabelValues(messageType).Inc()
	defer protoMappingInProgress.WithLabelValues(messageType).Dec()

	o.logger.InfoContext(ctx, "Starting profile mapping",
		slog.String("run_id", runID),
		slog.String("producer", producer),
		slog.String("symbol", dto.Symbol))

	result, err := MapProfileDTO(dto, runID, producer)
	
	// Record metrics
	duration := time.Since(start)
	outcome := "success"
	var resultSize int64
	
	if err != nil {
		outcome = "error"
		o.logger.ErrorContext(ctx, "Profile mapping failed",
			slog.String("run_id", runID),
			slog.String("symbol", dto.Symbol),
			slog.String("error", err.Error()),
			slog.Duration("duration", duration))
	} else {
		resultSize = int64(len(result.JSONBytes))
		o.logger.InfoContext(ctx, "Profile mapping completed",
			slog.String("run_id", runID),
			slog.String("symbol", dto.Symbol),
			slog.String("content_type", result.ContentType),
			slog.Int64("size_bytes", resultSize),
			slog.Duration("duration", duration))
	}

	protoEmissionTotal.WithLabelValues(messageType, outcome).Inc()
	protoEmissionLatencyMs.WithLabelValues(messageType).Observe(float64(duration.Nanoseconds()) / 1e6)
	
	if resultSize > 0 {
		protoEmissionBytes.WithLabelValues(messageType).Observe(float64(resultSize))
	}

	return result, err
}

// MapNewsWithObservability maps news with observability
func (o *ObservableMapper) MapNewsWithObservability(ctx context.Context, items []scrape.NewsItem, symbol, runID, producer string) ([]*newsv1.NewsItem, error) {
	start := time.Now()
	messageType := "news"
	
	// Increment in-progress gauge
	protoMappingInProgress.WithLabelValues(messageType).Inc()
	defer protoMappingInProgress.WithLabelValues(messageType).Dec()

	o.logger.InfoContext(ctx, "Starting news mapping",
		slog.String("run_id", runID),
		slog.String("producer", producer),
		slog.String("symbol", symbol),
		slog.Int("article_count", len(items)))

	articles, err := MapNewsItems(items, symbol, runID, producer)
	
	// Record metrics
	duration := time.Since(start)
	outcome := "success"
	var totalSize int64
	
	if err != nil {
		outcome = "error"
		o.logger.ErrorContext(ctx, "News mapping failed",
			slog.String("run_id", runID),
			slog.String("symbol", symbol),
			slog.String("error", err.Error()),
			slog.Duration("duration", duration))
	} else {
		for _, article := range articles {
			totalSize += estimateProtoSize(article)
		}
		
		o.logger.InfoContext(ctx, "News mapping completed",
			slog.String("run_id", runID),
			slog.String("symbol", symbol),
			slog.Int("mapped_articles", len(articles)),
			slog.Int("original_articles", len(items)),
			slog.Int64("total_size_bytes", totalSize),
			slog.Duration("duration", duration))
	}

	protoEmissionTotal.WithLabelValues(messageType, outcome).Inc()
	protoEmissionLatencyMs.WithLabelValues(messageType).Observe(float64(duration.Nanoseconds()) / 1e6)
	
	if totalSize > 0 {
		protoEmissionBytes.WithLabelValues(messageType).Observe(float64(totalSize))
	}

	return articles, err
}

// RecordValidationError records a validation error metric
func RecordValidationError(messageType, field string) {
	protoValidationErrors.WithLabelValues(messageType, field).Inc()
}

// LogMappingSummary logs a summary of mapping operations
func (o *ObservableMapper) LogMappingSummary(ctx context.Context, summary MappingSummary) {
	o.logger.InfoContext(ctx, "Mapping summary",
		slog.String("run_id", summary.RunID),
		slog.Int("total_operations", summary.TotalOperations),
		slog.Int("successful_operations", summary.SuccessfulOperations),
		slog.Int("failed_operations", summary.FailedOperations),
		slog.Duration("total_duration", summary.TotalDuration),
		slog.Int64("total_bytes", summary.TotalBytes),
		slog.Float64("success_rate", summary.GetSuccessRate()),
		slog.Float64("avg_latency_ms", summary.GetAverageLatencyMs()))
}

// MappingSummary represents a summary of mapping operations
type MappingSummary struct {
	RunID                string
	TotalOperations      int
	SuccessfulOperations int
	FailedOperations     int
	TotalDuration        time.Duration
	TotalBytes           int64
	OperationsByType     map[string]int
}

// GetSuccessRate returns the success rate as a percentage
func (s *MappingSummary) GetSuccessRate() float64 {
	if s.TotalOperations == 0 {
		return 0.0
	}
	return float64(s.SuccessfulOperations) / float64(s.TotalOperations) * 100.0
}

// GetAverageLatencyMs returns the average latency in milliseconds
func (s *MappingSummary) GetAverageLatencyMs() float64 {
	if s.TotalOperations == 0 {
		return 0.0
	}
	return float64(s.TotalDuration.Nanoseconds()) / float64(s.TotalOperations) / 1e6
}

// estimateProtoSize estimates the size of a protobuf message
func estimateProtoSize(msg interface{}) int64 {
	// This is a rough estimation - in practice you'd use proto.Size()
	// For now, we'll use a simple heuristic based on message type
	switch v := msg.(type) {
	case *fundamentalsv1.FundamentalsSnapshot:
		// Estimate based on number of lines and average field sizes
		baseSize := int64(200) // Base message overhead
		lineSize := int64(100) // Average size per line item
		return baseSize + int64(len(v.Lines))*lineSize
		
	case *newsv1.NewsItem:
		// Estimate based on content sizes
		baseSize := int64(100) // Base message overhead
		headlineSize := int64(len(v.Headline))
		urlSize := int64(len(v.Url))
		sourceSize := int64(len(v.Source))
		bodySize := int64(len(v.Body))
		tickerSize := int64(len(v.Tickers) * 10) // Average ticker length
		return baseSize + headlineSize + urlSize + sourceSize + bodySize + tickerSize
		
	default:
		// Default estimation
		return 500
	}
}

// NewMappingSummary creates a new mapping summary
func NewMappingSummary(runID string) *MappingSummary {
	return &MappingSummary{
		RunID:            runID,
		OperationsByType: make(map[string]int),
	}
}

// RecordOperation records a mapping operation in the summary
func (s *MappingSummary) RecordOperation(messageType string, duration time.Duration, bytes int64, success bool) {
	s.TotalOperations++
	s.TotalDuration += duration
	s.TotalBytes += bytes
	s.OperationsByType[messageType]++
	
	if success {
		s.SuccessfulOperations++
	} else {
		s.FailedOperations++
	}
}

