package bus

import (
	"context"
	"time"
)

// Publisher defines the interface for publishing messages to the bus
type Publisher interface {
	// PublishBars publishes a bar batch to the bus
	PublishBars(ctx context.Context, batch *BarBatchMessage) error
	
	// PublishQuote publishes a quote to the bus
	PublishQuote(ctx context.Context, quote *QuoteMessage) error
	
	// PublishFundamentals publishes fundamentals to the bus
	PublishFundamentals(ctx context.Context, fundamentals *FundamentalsMessage) error
	
	// Close closes the publisher and cleans up resources
	Close(ctx context.Context) error
}

// Message represents a message to be published
type Message struct {
	Topic     string
	Envelope  *Envelope
	Payload   []byte
	Key       *Key
}

// Envelope represents the ampy-bus envelope structure
type Envelope struct {
	MessageID     string            `json:"message_id"`
	SchemaFQDN    string            `json:"schema_fqdn"`
	SchemaVersion string            `json:"schema_version"`
	ContentType   string            `json:"content_type"`
	ContentEncoding string          `json:"content_encoding,omitempty"`
	ProducedAt    time.Time         `json:"produced_at"`
	Producer      string            `json:"producer"`
	Source        string            `json:"source"`
	RunID         string            `json:"run_id"`
	TraceID       string            `json:"trace_id,omitempty"`
	SpanID        string            `json:"span_id,omitempty"`
	PartitionKey  string            `json:"partition_key"`
	DedupeKey     string            `json:"dedupe_key,omitempty"`
	RetryCount    int               `json:"retry_count,omitempty"`
	DlqReason     string            `json:"dlq_reason,omitempty"`
	SchemaHash    string            `json:"schema_hash,omitempty"`
	BlobRef       string            `json:"blob_ref,omitempty"`
	BlobHash      string            `json:"blob_hash,omitempty"`
	BlobSize      int64             `json:"blob_size,omitempty"`
	Extensions    map[string]string `json:"extensions,omitempty"`
}

// Key represents a partition key for ordering
type Key struct {
	Symbol string
	MIC    string
}

// PartitionKey returns the formatted partition key
func (k *Key) PartitionKey() string {
	if k.MIC == "" {
		return k.Symbol
	}
	return k.MIC + "." + k.Symbol
}

// BarBatchMessage represents a bar batch message
type BarBatchMessage struct {
	Batch   interface{} // *ampy.bars.v1.BarBatch
	Key     *Key
	RunID   string
	Env     string
}

// QuoteMessage represents a quote message
type QuoteMessage struct {
	Quote   interface{} // *ampy.ticks.v1.QuoteTick
	Key     *Key
	RunID   string
	Env     string
}

// FundamentalsMessage represents a fundamentals message
type FundamentalsMessage struct {
	Fundamentals interface{} // *ampy.fundamentals.v1.FundamentalsSnapshot
	Key          *Key
	RunID        string
	Env          string
}

// Config represents the bus configuration
type Config struct {
	Enabled           bool          `yaml:"enabled"`
	Env               string        `yaml:"env"`
	TopicPrefix       string        `yaml:"topic_prefix"`
	MaxPayloadBytes   int64         `yaml:"max_payload_bytes"`
	Publisher         PublisherConfig `yaml:"publisher"`
	Retry             RetryConfig   `yaml:"retry"`
	CircuitBreaker    CircuitBreakerConfig `yaml:"circuit_breaker"`
}

// PublisherConfig represents publisher-specific configuration
type PublisherConfig struct {
	Backend string      `yaml:"backend"`
	NATS    NATSConfig  `yaml:"nats"`
	Kafka   KafkaConfig `yaml:"kafka"`
}

// NATSConfig represents NATS-specific configuration
type NATSConfig struct {
	URL           string `yaml:"url"`
	SubjectStyle  string `yaml:"subject_style"`
	AckWaitMs     int    `yaml:"ack_wait_ms"`
}

// KafkaConfig represents Kafka-specific configuration
type KafkaConfig struct {
	Brokers     []string `yaml:"brokers"`
	Acks        string   `yaml:"acks"`
	Compression string   `yaml:"compression"`
}

// RetryConfig represents retry configuration
type RetryConfig struct {
	Attempts     int           `yaml:"attempts"`
	BaseMs       int           `yaml:"base_ms"`
	MaxDelayMs   int           `yaml:"max_delay_ms"`
}

// CircuitBreakerConfig represents circuit breaker configuration
type CircuitBreakerConfig struct {
	Window          int `yaml:"window"`
	FailureThreshold float64 `yaml:"failure_threshold"`
	ResetTimeoutMs  int `yaml:"reset_timeout_ms"`
	HalfOpenProbes  int `yaml:"half_open_probes"`
}

// PreviewSummary represents a preview summary for dry-run mode
type PreviewSummary struct {
	Topic         string
	Envelope      *Envelope
	PartitionKey  string
	Chunking      ChunkingInfo
	Span          string
	PayloadBytes  int
	MessageCount  int
}

// ChunkingInfo represents chunking information
type ChunkingInfo struct {
	ChunkCount    int
	MaxPayload    int64
	ChunkSizes    []int
}
