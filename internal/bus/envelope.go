package bus

import (
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// EnvelopeBuilder builds ampy-bus envelopes
type EnvelopeBuilder struct {
	producer string
	source   string
}

// NewEnvelopeBuilder creates a new envelope builder
func NewEnvelopeBuilder(producer, source string) *EnvelopeBuilder {
	return &EnvelopeBuilder{
		producer: producer,
		source:   source,
	}
}

// BuildEnvelope builds an envelope for the given message type and key
func (b *EnvelopeBuilder) BuildEnvelope(
	schemaFQDN string,
	schemaVersion string,
	key *Key,
	runID string,
	traceID string,
	extensions map[string]string,
) *Envelope {
	messageID := uuid.Must(uuid.NewV7()).String()
	
	envelope := &Envelope{
		MessageID:      messageID,
		SchemaFQDN:     schemaFQDN,
		SchemaVersion:  schemaVersion,
		ContentType:    "application/x-protobuf",
		ProducedAt:     time.Now().UTC(),
		Producer:       b.producer,
		Source:         b.source,
		RunID:          runID,
		TraceID:        traceID,
		PartitionKey:   key.PartitionKey(),
		Extensions:     extensions,
	}
	
	// Set content encoding if specified in extensions
	if encoding, ok := extensions["content_encoding"]; ok {
		envelope.ContentEncoding = encoding
	}
	
	// Set dedupe key if specified in extensions
	if dedupeKey, ok := extensions["dedupe_key"]; ok {
		envelope.DedupeKey = dedupeKey
	}
	
	// Set retry count if specified in extensions
	if retryCount, ok := extensions["retry_count"]; ok {
		if retryCount == "1" {
			envelope.RetryCount = 1
		}
	}
	
	return envelope
}

// BuildChunkedEnvelope builds an envelope for a chunked message
func (b *EnvelopeBuilder) BuildChunkedEnvelope(
	schemaFQDN string,
	schemaVersion string,
	key *Key,
	runID string,
	traceID string,
	chunkIndex int,
	totalChunks int,
	extensions map[string]string,
) *Envelope {
	envelope := b.BuildEnvelope(schemaFQDN, schemaVersion, key, runID, traceID, extensions)
	
	// Add chunking information to extensions
	if envelope.Extensions == nil {
		envelope.Extensions = make(map[string]string)
	}
	envelope.Extensions["chunk_index"] = fmt.Sprintf("%d", chunkIndex)
	envelope.Extensions["total_chunks"] = fmt.Sprintf("%d", totalChunks)
	
	return envelope
}

// ComputeSchemaHash computes a hash of the schema for validation
func ComputeSchemaHash(schemaFQDN string) string {
	hash := sha256.Sum256([]byte(schemaFQDN))
	return fmt.Sprintf("%x", hash[:8]) // Use first 8 bytes for brevity
}

// ValidateEnvelope validates an envelope
func ValidateEnvelope(envelope *Envelope) error {
	if envelope == nil {
		return fmt.Errorf("envelope cannot be nil")
	}
	
	if envelope.MessageID == "" {
		return fmt.Errorf("message_id is required")
	}
	
	if envelope.SchemaFQDN == "" {
		return fmt.Errorf("schema_fqdn is required")
	}
	
	if envelope.SchemaVersion == "" {
		return fmt.Errorf("schema_version is required")
	}
	
	if envelope.Producer == "" {
		return fmt.Errorf("producer is required")
	}
	
	if envelope.Source == "" {
		return fmt.Errorf("source is required")
	}
	
	if envelope.RunID == "" {
		return fmt.Errorf("run_id is required")
	}
	
	if envelope.PartitionKey == "" {
		return fmt.Errorf("partition_key is required")
	}
	
	// Validate UUID format for message_id
	if _, err := uuid.Parse(envelope.MessageID); err != nil {
		return fmt.Errorf("invalid message_id format: %w", err)
	}
	
	return nil
}
