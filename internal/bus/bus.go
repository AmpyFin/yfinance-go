package bus

import (
	"context"
	"fmt"
	"os"
)

// Bus represents the main bus interface
type Bus struct {
	config          *Config
	publisher       Publisher
	previewPublisher *PreviewPublisher
	retryPolicy     *RetryPolicy
	circuitBreaker  *CircuitBreaker
}

// NewBus creates a new bus instance
func NewBus(config *Config) (*Bus, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	
	// Validate config
	if err := ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	
	// Create retry policy
	retryPolicy := NewRetryPolicy(&config.Retry)
	
	// Create circuit breaker
	circuitBreaker := NewCircuitBreaker(&config.CircuitBreaker)
	
	// Create preview publisher
	previewPublisher := NewPreviewPublisher(config)
	
	// Create actual publisher if enabled
	var publisher Publisher
	if config.Enabled {
		busPublisher, err := NewBusPublisher(config)
		if err != nil {
			return nil, fmt.Errorf("failed to create bus publisher: %w", err)
		}
		publisher = busPublisher
	}
	
	return &Bus{
		config:          config,
		publisher:       publisher,
		previewPublisher: previewPublisher,
		retryPolicy:     retryPolicy,
		circuitBreaker:  circuitBreaker,
	}, nil
}

// PublishBars publishes bars with retry and circuit breaker protection
func (b *Bus) PublishBars(ctx context.Context, batch *BarBatchMessage) error {
	if !b.config.Enabled {
		return fmt.Errorf("bus publishing is disabled")
	}
	
	// Execute with retry and circuit breaker
	return b.retryPolicy.ExecuteWithRetry(ctx, func() error {
		return b.circuitBreaker.Execute(ctx, func() error {
			return b.publisher.PublishBars(ctx, batch)
		})
	})
}

// PublishQuote publishes quote with retry and circuit breaker protection
func (b *Bus) PublishQuote(ctx context.Context, quote *QuoteMessage) error {
	if !b.config.Enabled {
		return fmt.Errorf("bus publishing is disabled")
	}
	
	// Execute with retry and circuit breaker
	return b.retryPolicy.ExecuteWithRetry(ctx, func() error {
		return b.circuitBreaker.Execute(ctx, func() error {
			return b.publisher.PublishQuote(ctx, quote)
		})
	})
}

// PublishFundamentals publishes fundamentals with retry and circuit breaker protection
func (b *Bus) PublishFundamentals(ctx context.Context, fundamentals *FundamentalsMessage) error {
	if !b.config.Enabled {
		return fmt.Errorf("bus publishing is disabled")
	}
	
	// Execute with retry and circuit breaker
	return b.retryPolicy.ExecuteWithRetry(ctx, func() error {
		return b.circuitBreaker.Execute(ctx, func() error {
			return b.publisher.PublishFundamentals(ctx, fundamentals)
		})
	})
}

// PreviewBars generates a preview for bar batch publishing
func (b *Bus) PreviewBars(batch *BarBatchMessage, payloadSize int) (*PreviewSummary, error) {
	return b.previewPublisher.PreviewBars(batch, payloadSize)
}

// PreviewQuote generates a preview for quote publishing
func (b *Bus) PreviewQuote(quote *QuoteMessage, payloadSize int) (*PreviewSummary, error) {
	return b.previewPublisher.PreviewQuote(quote, payloadSize)
}

// PreviewFundamentals generates a preview for fundamentals publishing
func (b *Bus) PreviewFundamentals(fundamentals *FundamentalsMessage, payloadSize int) (*PreviewSummary, error) {
	return b.previewPublisher.PreviewFundamentals(fundamentals, payloadSize)
}

// Close closes the bus
func (b *Bus) Close(ctx context.Context) error {
	if b.publisher != nil {
		return b.publisher.Close(ctx)
	}
	return nil
}

// GetConfig returns the bus configuration
func (b *Bus) GetConfig() *Config {
	return b.config
}

// GetCircuitBreakerStats returns circuit breaker statistics
func (b *Bus) GetCircuitBreakerStats() CircuitBreakerStats {
	return b.circuitBreaker.GetStats()
}

// ValidateConfig validates the bus configuration
func ValidateConfig(config *Config) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}
	
	// Validate environment
	if config.Env == "" {
		return fmt.Errorf("environment cannot be empty")
	}
	
	validEnvs := map[string]bool{
		"dev":     true,
		"staging": true,
		"prod":    true,
	}
	
	if !validEnvs[config.Env] {
		return fmt.Errorf("invalid environment: %s (must be dev, staging, or prod)", config.Env)
	}
	
	// Validate topic prefix
	if config.TopicPrefix == "" {
		return fmt.Errorf("topic prefix cannot be empty")
	}
	
	// Validate max payload bytes
	if config.MaxPayloadBytes < 256*1024 { // 256 KiB
		return fmt.Errorf("max payload bytes must be at least 256 KiB")
	}
	
	if config.MaxPayloadBytes > 10*1024*1024 { // 10 MiB
		return fmt.Errorf("max payload bytes must be at most 10 MiB")
	}
	
	// Validate publisher config if enabled
	if config.Enabled {
		if err := validatePublisherConfig(&config.Publisher); err != nil {
			return fmt.Errorf("invalid publisher config: %w", err)
		}
	}
	
	// Validate retry config
	if err := validateRetryConfig(&config.Retry); err != nil {
		return fmt.Errorf("invalid retry config: %w", err)
	}
	
	// Validate circuit breaker config
	if err := validateCircuitBreakerConfig(&config.CircuitBreaker); err != nil {
		return fmt.Errorf("invalid circuit breaker config: %w", err)
	}
	
	return nil
}

// validatePublisherConfig validates publisher configuration
func validatePublisherConfig(config *PublisherConfig) error {
	if config.Backend == "" {
		return fmt.Errorf("backend cannot be empty")
	}
	
	validBackends := map[string]bool{
		"nats":  true,
		"kafka": true,
	}
	
	if !validBackends[config.Backend] {
		return fmt.Errorf("invalid backend: %s (must be nats or kafka)", config.Backend)
	}
	
	// Validate backend-specific config
	switch config.Backend {
	case "nats":
		if err := validateNATSConfig(&config.NATS); err != nil {
			return fmt.Errorf("invalid NATS config: %w", err)
		}
	case "kafka":
		if err := validateKafkaConfig(&config.Kafka); err != nil {
			return fmt.Errorf("invalid Kafka config: %w", err)
		}
	}
	
	return nil
}

// validateNATSConfig validates NATS configuration
func validateNATSConfig(config *NATSConfig) error {
	if config.URL == "" {
		return fmt.Errorf("NATS URL cannot be empty")
	}
	
	if config.AckWaitMs <= 0 {
		return fmt.Errorf("NATS ack wait must be positive")
	}
	
	return nil
}

// validateKafkaConfig validates Kafka configuration
func validateKafkaConfig(config *KafkaConfig) error {
	if len(config.Brokers) == 0 {
		return fmt.Errorf("Kafka brokers cannot be empty")
	}
	
	if config.Acks == "" {
		return fmt.Errorf("Kafka acks cannot be empty")
	}
	
	validAcks := map[string]bool{
		"all":    true,
		"1":      true,
		"0":      true,
	}
	
	if !validAcks[config.Acks] {
		return fmt.Errorf("invalid Kafka acks: %s (must be all, 1, or 0)", config.Acks)
	}
	
	return nil
}

// validateRetryConfig validates retry configuration
func validateRetryConfig(config *RetryConfig) error {
	if config.Attempts <= 0 {
		return fmt.Errorf("retry attempts must be positive")
	}
	
	if config.BaseMs <= 0 {
		return fmt.Errorf("retry base delay must be positive")
	}
	
	if config.MaxDelayMs <= 0 {
		return fmt.Errorf("retry max delay must be positive")
	}
	
	if config.MaxDelayMs < config.BaseMs {
		return fmt.Errorf("retry max delay must be >= base delay")
	}
	
	return nil
}

// validateCircuitBreakerConfig validates circuit breaker configuration
func validateCircuitBreakerConfig(config *CircuitBreakerConfig) error {
	if config.Window <= 0 {
		return fmt.Errorf("circuit breaker window must be positive")
	}
	
	if config.FailureThreshold <= 0 || config.FailureThreshold > 1 {
		return fmt.Errorf("circuit breaker failure threshold must be between 0 and 1")
	}
	
	if config.ResetTimeoutMs <= 0 {
		return fmt.Errorf("circuit breaker reset timeout must be positive")
	}
	
	if config.HalfOpenProbes <= 0 {
		return fmt.Errorf("circuit breaker half-open probes must be positive")
	}
	
	return nil
}

// LoadConfigFromFile loads configuration from a YAML file
func LoadConfigFromFile(filename string) (*Config, error) {
	// This would load from YAML file
	// For now, return a default config
	return GetDefaultConfig(), nil
}

// GetDefaultConfig returns a default configuration
func GetDefaultConfig() *Config {
	return &Config{
		Enabled:         false,
		Env:             "dev",
		TopicPrefix:     "ampy",
		MaxPayloadBytes: 1024 * 1024, // 1 MiB
		Publisher: PublisherConfig{
			Backend: "nats",
			NATS: NATSConfig{
				URL:          "nats://localhost:4222",
				SubjectStyle: "topic",
				AckWaitMs:    5000,
			},
			Kafka: KafkaConfig{
				Brokers:     []string{"localhost:9092"},
				Acks:        "all",
				Compression: "snappy",
			},
		},
		Retry: RetryConfig{
			Attempts:   5,
			BaseMs:     250,
			MaxDelayMs: 8000,
		},
		CircuitBreaker: CircuitBreakerConfig{
			Window:          50,
			FailureThreshold: 0.30,
			ResetTimeoutMs:  30000,
			HalfOpenProbes:  3,
		},
	}
}

// GetConfigFromEnv loads configuration from environment variables
func GetConfigFromEnv() *Config {
	config := GetDefaultConfig()
	
	if env := os.Getenv("AMPY_BUS_ENABLED"); env != "" {
		config.Enabled = env == "true"
	}
	
	if env := os.Getenv("AMPY_BUS_ENV"); env != "" {
		config.Env = env
	}
	
	if env := os.Getenv("AMPY_BUS_TOPIC_PREFIX"); env != "" {
		config.TopicPrefix = env
	}
	
	if env := os.Getenv("AMPY_BUS_NATS_URL"); env != "" {
		config.Publisher.NATS.URL = env
	}
	
	return config
}
