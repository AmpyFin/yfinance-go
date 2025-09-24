package bus

import (
	"context"
	"fmt"

	"github.com/AmpyFin/ampy-bus/pkg/ampybus"
	"github.com/AmpyFin/ampy-bus/pkg/ampybus/natsbinding"
	"google.golang.org/protobuf/proto"
)

// BusPublisher implements the Publisher interface using ampy-bus
type BusPublisher struct {
	config       *Config
	bus          *natsbinding.Bus
	topicBuilder *TopicBuilder
	envelopeBuilder *EnvelopeBuilder
	chunking     *ChunkingStrategy
}

// NewBusPublisher creates a new bus publisher
func NewBusPublisher(config *Config) (*BusPublisher, error) {
	if !config.Enabled {
		return nil, fmt.Errorf("bus publishing is disabled")
	}
	
	// Create topic builder
	topicBuilder := NewTopicBuilder(config.Env, config.TopicPrefix)
	
	// Create envelope builder
	producer := fmt.Sprintf("yfinance-go@%s", getHostname())
	envelopeBuilder := NewEnvelopeBuilder(producer, "yfinance-go")
	
	// Create chunking strategy
	chunking := NewChunkingStrategy(config.MaxPayloadBytes)
	
	// Create bus based on backend
	var bus *natsbinding.Bus
	var err error
	
	switch config.Publisher.Backend {
	case "nats":
		bus, err = createNATSBus(config)
	case "kafka":
		return nil, fmt.Errorf("Kafka backend not implemented yet")
	default:
		return nil, fmt.Errorf("unsupported backend: %s", config.Publisher.Backend)
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to create bus: %w", err)
	}
	
	return &BusPublisher{
		config:          config,
		bus:             bus,
		topicBuilder:    topicBuilder,
		envelopeBuilder: envelopeBuilder,
		chunking:        chunking,
	}, nil
}

// PublishBars publishes a bar batch to the bus
func (p *BusPublisher) PublishBars(ctx context.Context, batch *BarBatchMessage) error {
	// Build topic
	topic := p.topicBuilder.BuildBarsTopic(batch.Key, "v1")
	
	// Marshal the batch to protobuf
	payload, err := proto.Marshal(batch.Batch.(proto.Message))
	if err != nil {
		return fmt.Errorf("failed to marshal bar batch: %w", err)
	}
	
	// Build envelope
	envelope := p.envelopeBuilder.BuildEnvelope(
		"ampy.bars.v1.BarBatch",
		"1.0.0",
		batch.Key,
		batch.RunID,
		"", // traceID - could be extracted from context
		nil,
	)
	
	// Check if chunking is needed
	chunkResult, err := p.chunking.ChunkPayload(payload)
	if err != nil {
		return fmt.Errorf("failed to chunk payload: %w", err)
	}
	
	// Publish chunks
	for i, chunk := range chunkResult.Chunks {
		chunkEnvelope := envelope
		
		// If we have multiple chunks, create a chunked envelope
		if len(chunkResult.Chunks) > 1 {
			chunkEnvelope = p.envelopeBuilder.BuildChunkedEnvelope(
				"ampy.bars.v1.BarBatch",
				"1.0.0",
				batch.Key,
				batch.RunID,
				"",
				i,
				len(chunkResult.Chunks),
				nil,
			)
		}
		
		// Create ampy-bus envelope
		ampyEnvelope := ampybus.Envelope{
			Topic: topic,
			Headers: ampybus.Headers{
				MessageID:     chunkEnvelope.MessageID,
				SchemaFQDN:    chunkEnvelope.SchemaFQDN,
				SchemaVersion: chunkEnvelope.SchemaVersion,
				ContentType:   chunkEnvelope.ContentType,
				ProducedAt:    chunkEnvelope.ProducedAt,
				Producer:      chunkEnvelope.Producer,
				Source:        chunkEnvelope.Source,
				RunID:         chunkEnvelope.RunID,
				TraceID:       chunkEnvelope.TraceID,
				PartitionKey:  chunkEnvelope.PartitionKey,
			},
			Payload: chunk,
		}
		
		// Publish to bus
		_, err = p.bus.PublishEnvelope(ctx, ampyEnvelope, map[string]string{})
		if err != nil {
			return fmt.Errorf("failed to publish bar batch chunk %d: %w", i, err)
		}
	}
	
	return nil
}

// PublishQuote publishes a quote to the bus
func (p *BusPublisher) PublishQuote(ctx context.Context, quote *QuoteMessage) error {
	// Build topic
	topic := p.topicBuilder.BuildQuotesTopic(quote.Key, "v1")
	
	// Marshal the quote to protobuf
	payload, err := proto.Marshal(quote.Quote.(proto.Message))
	if err != nil {
		return fmt.Errorf("failed to marshal quote: %w", err)
	}
	
	// Build envelope
	envelope := p.envelopeBuilder.BuildEnvelope(
		"ampy.ticks.v1.QuoteTick",
		"1.0.0",
		quote.Key,
		quote.RunID,
		"", // traceID - could be extracted from context
		nil,
	)
	
	// Check if chunking is needed
	chunkResult, err := p.chunking.ChunkPayload(payload)
	if err != nil {
		return fmt.Errorf("failed to chunk payload: %w", err)
	}
	
	// Publish chunks
	for i, chunk := range chunkResult.Chunks {
		chunkEnvelope := envelope
		
		// If we have multiple chunks, create a chunked envelope
		if len(chunkResult.Chunks) > 1 {
			chunkEnvelope = p.envelopeBuilder.BuildChunkedEnvelope(
				"ampy.ticks.v1.QuoteTick",
				"1.0.0",
				quote.Key,
				quote.RunID,
				"",
				i,
				len(chunkResult.Chunks),
				nil,
			)
		}
		
		// Create ampy-bus envelope
		ampyEnvelope := ampybus.Envelope{
			Topic: topic,
			Headers: ampybus.Headers{
				MessageID:     chunkEnvelope.MessageID,
				SchemaFQDN:    chunkEnvelope.SchemaFQDN,
				SchemaVersion: chunkEnvelope.SchemaVersion,
				ContentType:   chunkEnvelope.ContentType,
				ProducedAt:    chunkEnvelope.ProducedAt,
				Producer:      chunkEnvelope.Producer,
				Source:        chunkEnvelope.Source,
				RunID:         chunkEnvelope.RunID,
				TraceID:       chunkEnvelope.TraceID,
				PartitionKey:  chunkEnvelope.PartitionKey,
			},
			Payload: chunk,
		}
		
		// Publish to bus
		_, err = p.bus.PublishEnvelope(ctx, ampyEnvelope, map[string]string{})
		if err != nil {
			return fmt.Errorf("failed to publish quote chunk %d: %w", i, err)
		}
	}
	
	return nil
}

// PublishFundamentals publishes fundamentals to the bus
func (p *BusPublisher) PublishFundamentals(ctx context.Context, fundamentals *FundamentalsMessage) error {
	// Build topic
	topic := p.topicBuilder.BuildFundamentalsTopic(fundamentals.Key, "v1")
	
	// Marshal the fundamentals to protobuf
	payload, err := proto.Marshal(fundamentals.Fundamentals.(proto.Message))
	if err != nil {
		return fmt.Errorf("failed to marshal fundamentals: %w", err)
	}
	
	// Build envelope
	envelope := p.envelopeBuilder.BuildEnvelope(
		"ampy.fundamentals.v1.FundamentalsSnapshot",
		"1.0.0",
		fundamentals.Key,
		fundamentals.RunID,
		"", // traceID - could be extracted from context
		nil,
	)
	
	// Check if chunking is needed
	chunkResult, err := p.chunking.ChunkPayload(payload)
	if err != nil {
		return fmt.Errorf("failed to chunk payload: %w", err)
	}
	
	// Publish chunks
	for i, chunk := range chunkResult.Chunks {
		chunkEnvelope := envelope
		
		// If we have multiple chunks, create a chunked envelope
		if len(chunkResult.Chunks) > 1 {
			chunkEnvelope = p.envelopeBuilder.BuildChunkedEnvelope(
				"ampy.fundamentals.v1.FundamentalsSnapshot",
				"1.0.0",
				fundamentals.Key,
				fundamentals.RunID,
				"",
				i,
				len(chunkResult.Chunks),
				nil,
			)
		}
		
		// Create ampy-bus envelope
		ampyEnvelope := ampybus.Envelope{
			Topic: topic,
			Headers: ampybus.Headers{
				MessageID:     chunkEnvelope.MessageID,
				SchemaFQDN:    chunkEnvelope.SchemaFQDN,
				SchemaVersion: chunkEnvelope.SchemaVersion,
				ContentType:   chunkEnvelope.ContentType,
				ProducedAt:    chunkEnvelope.ProducedAt,
				Producer:      chunkEnvelope.Producer,
				Source:        chunkEnvelope.Source,
				RunID:         chunkEnvelope.RunID,
				TraceID:       chunkEnvelope.TraceID,
				PartitionKey:  chunkEnvelope.PartitionKey,
			},
			Payload: chunk,
		}
		
		// Publish to bus
		_, err = p.bus.PublishEnvelope(ctx, ampyEnvelope, map[string]string{})
		if err != nil {
			return fmt.Errorf("failed to publish fundamentals chunk %d: %w", i, err)
		}
	}
	
	return nil
}

// Close closes the publisher
func (p *BusPublisher) Close(ctx context.Context) error {
	if p.bus != nil {
		p.bus.Close()
	}
	return nil
}

// createNATSBus creates a NATS bus
func createNATSBus(config *Config) (*natsbinding.Bus, error) {
	natsConfig := natsbinding.Config{
		URLs:          config.Publisher.NATS.URL,
		StreamName:    "AMPY_TRADING",
		Subjects:      []string{fmt.Sprintf("%s.%s.>", config.TopicPrefix, config.Env)},
		DurablePrefix: "ampy-trading",
	}
	
	return natsbinding.Connect(natsConfig)
}

// createKafkaBus creates a Kafka bus
func createKafkaBus(config *Config) (*natsbinding.Bus, error) {
	// Note: Kafka binding would be implemented here
	// For now, return an error as it's not implemented
	return nil, fmt.Errorf("Kafka backend not implemented yet")
}

// getHostname returns the hostname for the producer field
func getHostname() string {
	// In a real implementation, you would get the actual hostname
	// For now, return a default value
	return "localhost"
}
