package emit

import (
	"crypto/sha256"
	"fmt"
	"time"

	commonv1 "github.com/AmpyFin/ampy-proto/v2/gen/go/ampy/common/v1"
	"google.golang.org/protobuf/proto"
)

// MetaConfig holds configuration for metadata creation
type MetaConfig struct {
	RunID         string
	Producer      string
	Source        string
	TraceID       string
	IncludeChecksum bool
	IncludeProducedAt bool
}

// StampMeta creates standardized metadata for ampy-proto messages
func StampMeta(config MetaConfig, schemaVersion string) *commonv1.Meta {
	meta := &commonv1.Meta{
		RunId:         config.RunID,
		Source:        config.Source,
		Producer:      config.Producer,
		SchemaVersion: schemaVersion,
	}

	// Note: ProducedAt field not available in ampy-proto v2.1.0 Meta message
	// Timestamp information can be included in the message-specific fields if needed

	return meta
}

// StampMetaWithChecksum creates metadata with message checksum
func StampMetaWithChecksum(config MetaConfig, schemaVersion string, message proto.Message) (*commonv1.Meta, error) {
	meta := StampMeta(config, schemaVersion)

	// Add checksum if requested
	if config.IncludeChecksum && message != nil {
		checksum, err := calculateMessageChecksum(message)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate message checksum: %w", err)
		}
		meta.Checksum = checksum
	}

	return meta, nil
}

// calculateMessageChecksum calculates SHA-256 checksum of protobuf message
func calculateMessageChecksum(message proto.Message) (string, error) {
	// Marshal message to bytes
	data, err := proto.Marshal(message)
	if err != nil {
		return "", fmt.Errorf("failed to marshal message: %w", err)
	}

	// Calculate SHA-256 hash
	hash := sha256.Sum256(data)
	return fmt.Sprintf("sha256:%x", hash), nil
}

// CreateScrapeMetaConfig creates a MetaConfig for scrape operations
func CreateScrapeMetaConfig(runID, producer, traceID string) MetaConfig {
	return MetaConfig{
		RunID:             runID,
		Producer:          producer,
		Source:            "yfinance-go/scrape",
		TraceID:           traceID,
		IncludeChecksum:   false, // Disabled by default for performance
		IncludeProducedAt: true,  // Include timestamp by default
	}
}

// SchemaVersions contains the schema versions for different message types
var SchemaVersions = struct {
	Fundamentals string
	News         string
	Bars         string
	Ticks        string
	Common       string
}{
	Fundamentals: "ampy.fundamentals.v1:2.1.0",
	News:         "ampy.news.v1:2.1.0",
	Bars:         "ampy.bars.v1:2.1.0",
	Ticks:        "ampy.ticks.v1:2.1.0",
	Common:       "ampy.common.v1:2.1.0",
}

// ValidateMetaConfig validates metadata configuration
func ValidateMetaConfig(config MetaConfig) error {
	if config.RunID == "" {
		return fmt.Errorf("run_id cannot be empty")
	}

	if config.Producer == "" {
		return fmt.Errorf("producer cannot be empty")
	}

	if config.Source == "" {
		return fmt.Errorf("source cannot be empty")
	}

	return nil
}

// MetaSummary provides a summary of metadata for logging/preview
type MetaSummary struct {
	RunID         string    `json:"run_id"`
	Producer      string    `json:"producer"`
	Source        string    `json:"source"`
	SchemaVersion string    `json:"schema_version"`
	HasChecksum   bool      `json:"has_checksum"`
	ProducedAt    *time.Time `json:"produced_at,omitempty"`
	TraceID       string    `json:"trace_id,omitempty"`
}

// SummarizeMeta creates a summary of metadata for logging/preview
func SummarizeMeta(meta *commonv1.Meta) *MetaSummary {
	if meta == nil {
		return nil
	}

	summary := &MetaSummary{
		RunID:         meta.RunId,
		Producer:      meta.Producer,
		Source:        meta.Source,
		SchemaVersion: meta.SchemaVersion,
		HasChecksum:   meta.Checksum != "",
	}

	// Note: ProducedAt field not available in ampy-proto v2.1.0 Meta message
	// summary.ProducedAt remains nil

	return summary
}

// LineageInfo represents lineage information for data tracing
type LineageInfo struct {
	SourceSystem   string                 `json:"source_system"`
	SourceEndpoint string                 `json:"source_endpoint"`
	ProcessingStep string                 `json:"processing_step"`
	Timestamp      time.Time             `json:"timestamp"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// CreateLineageInfo creates lineage information for data tracing
func CreateLineageInfo(sourceSystem, sourceEndpoint, processingStep string, metadata map[string]interface{}) *LineageInfo {
	return &LineageInfo{
		SourceSystem:   sourceSystem,
		SourceEndpoint: sourceEndpoint,
		ProcessingStep: processingStep,
		Timestamp:      time.Now().UTC(),
		Metadata:       metadata,
	}
}

// AddLineageToMeta adds lineage information to metadata (as custom field if supported)
func AddLineageToMeta(meta *commonv1.Meta, lineage *LineageInfo) *commonv1.Meta {
	// Note: ampy-proto v2 common.Meta doesn't have custom fields
	// This is a placeholder for potential future extension
	// For now, we can encode lineage in the source field
	if meta != nil && lineage != nil {
		// Encode basic lineage info in source field
		meta.Source = fmt.Sprintf("%s/%s->%s", meta.Source, lineage.SourceEndpoint, lineage.ProcessingStep)
	}
	return meta
}

// MetricsInfo represents metrics information for observability
type MetricsInfo struct {
	MessageType   string        `json:"message_type"`
	MessageCount  int           `json:"message_count"`
	TotalBytes    int64         `json:"total_bytes"`
	ProcessingTime time.Duration `json:"processing_time"`
	SuccessCount  int           `json:"success_count"`
	ErrorCount    int           `json:"error_count"`
}

// CreateMetricsInfo creates metrics information for observability
func CreateMetricsInfo(messageType string) *MetricsInfo {
	return &MetricsInfo{
		MessageType:    messageType,
		MessageCount:   0,
		TotalBytes:     0,
		ProcessingTime: 0,
		SuccessCount:   0,
		ErrorCount:     0,
	}
}

// UpdateMetrics updates metrics with message information
func (m *MetricsInfo) UpdateMetrics(messageSize int64, processingTime time.Duration, success bool) {
	m.MessageCount++
	m.TotalBytes += messageSize
	m.ProcessingTime += processingTime

	if success {
		m.SuccessCount++
	} else {
		m.ErrorCount++
	}
}

// GetSuccessRate returns the success rate as a percentage
func (m *MetricsInfo) GetSuccessRate() float64 {
	if m.MessageCount == 0 {
		return 0.0
	}
	return float64(m.SuccessCount) / float64(m.MessageCount) * 100.0
}

// GetAverageMessageSize returns the average message size in bytes
func (m *MetricsInfo) GetAverageMessageSize() float64 {
	if m.MessageCount == 0 {
		return 0.0
	}
	return float64(m.TotalBytes) / float64(m.MessageCount)
}

// GetAverageProcessingTime returns the average processing time per message
func (m *MetricsInfo) GetAverageProcessingTime() time.Duration {
	if m.MessageCount == 0 {
		return 0
	}
	return m.ProcessingTime / time.Duration(m.MessageCount)
}
