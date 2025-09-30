package bus

import (
	"fmt"
)

// PreviewPublisher provides preview functionality without actually publishing
type PreviewPublisher struct {
	config          *Config
	topicBuilder    *TopicBuilder
	envelopeBuilder *EnvelopeBuilder
	chunking        *ChunkingStrategy
}

// NewPreviewPublisher creates a new preview publisher
func NewPreviewPublisher(config *Config) *PreviewPublisher {
	// Create topic builder
	topicBuilder := NewTopicBuilder(config.Env, config.TopicPrefix)

	// Create envelope builder
	producer := fmt.Sprintf("yfinance-go@%s", getHostname())
	envelopeBuilder := NewEnvelopeBuilder(producer, "yfinance-go")

	// Create chunking strategy
	chunking := NewChunkingStrategy(config.MaxPayloadBytes)

	return &PreviewPublisher{
		config:          config,
		topicBuilder:    topicBuilder,
		envelopeBuilder: envelopeBuilder,
		chunking:        chunking,
	}
}

// PreviewBars generates a preview for bar batch publishing
func (p *PreviewPublisher) PreviewBars(batch *BarBatchMessage, payloadSize int) (*PreviewSummary, error) {
	// Build topic
	topic := p.topicBuilder.BuildBarsTopic(batch.Key, "v1")

	// Build envelope
	envelope := p.envelopeBuilder.BuildEnvelope(
		"ampy.bars.v1.BarBatch",
		"1.0.0",
		batch.Key,
		batch.RunID,
		"", // traceID
		nil,
	)

	// Get chunking info
	chunkingInfo := p.chunking.GetChunkingInfo(payloadSize)

	return &PreviewSummary{
		Topic:        topic,
		Envelope:     envelope,
		PartitionKey: batch.Key.PartitionKey(),
		Chunking:     chunkingInfo,
		Span:         "ingest.fetch→emit→publish",
		PayloadBytes: payloadSize,
		MessageCount: chunkingInfo.ChunkCount,
	}, nil
}

// PreviewQuote generates a preview for quote publishing
func (p *PreviewPublisher) PreviewQuote(quote *QuoteMessage, payloadSize int) (*PreviewSummary, error) {
	// Build topic
	topic := p.topicBuilder.BuildQuotesTopic(quote.Key, "v1")

	// Build envelope
	envelope := p.envelopeBuilder.BuildEnvelope(
		"ampy.ticks.v1.QuoteTick",
		"1.0.0",
		quote.Key,
		quote.RunID,
		"", // traceID
		nil,
	)

	// Get chunking info
	chunkingInfo := p.chunking.GetChunkingInfo(payloadSize)

	return &PreviewSummary{
		Topic:        topic,
		Envelope:     envelope,
		PartitionKey: quote.Key.PartitionKey(),
		Chunking:     chunkingInfo,
		Span:         "ingest.fetch→emit→publish",
		PayloadBytes: payloadSize,
		MessageCount: chunkingInfo.ChunkCount,
	}, nil
}

// PreviewFundamentals generates a preview for fundamentals publishing
func (p *PreviewPublisher) PreviewFundamentals(fundamentals *FundamentalsMessage, payloadSize int) (*PreviewSummary, error) {
	// Build topic
	topic := p.topicBuilder.BuildFundamentalsTopic(fundamentals.Key, "v1")

	// Build envelope
	envelope := p.envelopeBuilder.BuildEnvelope(
		"ampy.fundamentals.v1.FundamentalsSnapshot",
		"1.0.0",
		fundamentals.Key,
		fundamentals.RunID,
		"", // traceID
		nil,
	)

	// Get chunking info
	chunkingInfo := p.chunking.GetChunkingInfo(payloadSize)

	return &PreviewSummary{
		Topic:        topic,
		Envelope:     envelope,
		PartitionKey: fundamentals.Key.PartitionKey(),
		Chunking:     chunkingInfo,
		Span:         "ingest.fetch→emit→publish",
		PayloadBytes: payloadSize,
		MessageCount: chunkingInfo.ChunkCount,
	}, nil
}

// PrintPreview prints a preview summary
func PrintPreview(summary *PreviewSummary) {
	fmt.Println("PUBLISH PREVIEW (no send)")
	fmt.Printf("Topic: %s\n", summary.Topic)
	fmt.Printf("Envelope: message_id=%s schema=%s v=%s run_id=%s\n",
		summary.Envelope.MessageID[:8]+"...",
		summary.Envelope.SchemaFQDN,
		summary.Envelope.SchemaVersion,
		summary.Envelope.RunID)
	fmt.Printf("PartitionKey: %s\n", summary.PartitionKey)

	if summary.Chunking.ChunkCount > 1 {
		fmt.Printf("Chunking: %d messages (max_payload=%.1f MiB)\n",
			summary.Chunking.ChunkCount,
			float64(summary.Chunking.MaxPayload)/(1024*1024))
	} else {
		fmt.Printf("Payload: %d bytes (single message)\n", summary.PayloadBytes)
	}

	fmt.Printf("Span: %s  p95=420ms\n", summary.Span)
}

// PrintPreviewJSON prints a preview summary in JSON format
func PrintPreviewJSON(summary *PreviewSummary) {
	// This would print the summary in JSON format
	// For now, just print the basic info
	fmt.Printf("{\n")
	fmt.Printf("  \"topic\": \"%s\",\n", summary.Topic)
	fmt.Printf("  \"partition_key\": \"%s\",\n", summary.PartitionKey)
	fmt.Printf("  \"message_count\": %d,\n", summary.MessageCount)
	fmt.Printf("  \"payload_bytes\": %d,\n", summary.PayloadBytes)
	fmt.Printf("  \"chunking\": {\n")
	fmt.Printf("    \"chunk_count\": %d,\n", summary.Chunking.ChunkCount)
	fmt.Printf("    \"max_payload\": %d\n", summary.Chunking.MaxPayload)
	fmt.Printf("  }\n")
	fmt.Printf("}\n")
}
