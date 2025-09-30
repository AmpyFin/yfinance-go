package bus

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBus(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		wantError bool
	}{
		{
			name:      "nil config",
			config:    nil,
			wantError: true,
		},
		{
			name: "valid config with bus disabled",
			config: &Config{
				Enabled:         false,
				Env:             "dev",
				TopicPrefix:     "ampy",
				MaxPayloadBytes: 1024 * 1024,
				Retry: RetryConfig{
					Attempts:   5,
					BaseMs:     250,
					MaxDelayMs: 8000,
				},
				CircuitBreaker: CircuitBreakerConfig{
					Window:           50,
					FailureThreshold: 0.30,
					ResetTimeoutMs:   30000,
					HalfOpenProbes:   3,
				},
			},
			wantError: false,
		},
		{
			name: "invalid environment",
			config: &Config{
				Enabled:         false,
				Env:             "invalid",
				TopicPrefix:     "ampy",
				MaxPayloadBytes: 1024 * 1024,
				Retry: RetryConfig{
					Attempts:   5,
					BaseMs:     250,
					MaxDelayMs: 8000,
				},
				CircuitBreaker: CircuitBreakerConfig{
					Window:           50,
					FailureThreshold: 0.30,
					ResetTimeoutMs:   30000,
					HalfOpenProbes:   3,
				},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bus, err := NewBus(tt.config)
			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, bus)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, bus)
			}
		})
	}
}

func TestBus_PreviewBars(t *testing.T) {
	config := GetDefaultConfig()
	config.Enabled = false // Disable actual publishing

	bus, err := NewBus(config)
	require.NoError(t, err)
	require.NotNil(t, bus)

	message := &BarBatchMessage{
		Key: &Key{
			Symbol: "AAPL",
			MIC:    "XNAS",
		},
		RunID: "test-run-id",
		Env:   "dev",
	}

	preview, err := bus.PreviewBars(message, 1000)
	require.NoError(t, err)
	require.NotNil(t, preview)

	assert.Equal(t, "ampy.dev.bars.v1.XNAS.AAPL", preview.Topic)
	assert.Equal(t, "XNAS.AAPL", preview.PartitionKey)
	assert.Equal(t, 1000, preview.PayloadBytes)
	assert.Equal(t, 1, preview.MessageCount)
}

func TestBus_PreviewQuote(t *testing.T) {
	config := GetDefaultConfig()
	config.Enabled = false // Disable actual publishing

	bus, err := NewBus(config)
	require.NoError(t, err)
	require.NotNil(t, bus)

	message := &QuoteMessage{
		Key: &Key{
			Symbol: "MSFT",
			MIC:    "XNAS",
		},
		RunID: "test-run-id",
		Env:   "dev",
	}

	preview, err := bus.PreviewQuote(message, 500)
	require.NoError(t, err)
	require.NotNil(t, preview)

	assert.Equal(t, "ampy.dev.ticks.v1.XNAS.MSFT", preview.Topic)
	assert.Equal(t, "XNAS.MSFT", preview.PartitionKey)
	assert.Equal(t, 500, preview.PayloadBytes)
	assert.Equal(t, 1, preview.MessageCount)
}

func TestBus_PreviewFundamentals(t *testing.T) {
	config := GetDefaultConfig()
	config.Enabled = false // Disable actual publishing

	bus, err := NewBus(config)
	require.NoError(t, err)
	require.NotNil(t, bus)

	message := &FundamentalsMessage{
		Key: &Key{
			Symbol: "GOOGL",
			MIC:    "XNAS",
		},
		RunID: "test-run-id",
		Env:   "dev",
	}

	preview, err := bus.PreviewFundamentals(message, 2000)
	require.NoError(t, err)
	require.NotNil(t, preview)

	assert.Equal(t, "ampy.dev.fundamentals.v1.GOOGL", preview.Topic)
	assert.Equal(t, "XNAS.GOOGL", preview.PartitionKey)
	assert.Equal(t, 2000, preview.PayloadBytes)
	assert.Equal(t, 1, preview.MessageCount)
}

func TestBus_PublishBars_Disabled(t *testing.T) {
	config := GetDefaultConfig()
	config.Enabled = false // Disable actual publishing

	bus, err := NewBus(config)
	require.NoError(t, err)
	require.NotNil(t, bus)

	message := &BarBatchMessage{
		Key: &Key{
			Symbol: "AAPL",
			MIC:    "XNAS",
		},
		RunID: "test-run-id",
		Env:   "dev",
	}

	ctx := context.Background()
	err = bus.PublishBars(ctx, message)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bus publishing is disabled")
}

func TestBus_GetConfig(t *testing.T) {
	config := GetDefaultConfig()
	bus, err := NewBus(config)
	require.NoError(t, err)
	require.NotNil(t, bus)

	retrievedConfig := bus.GetConfig()
	assert.Equal(t, config, retrievedConfig)
}

func TestBus_GetCircuitBreakerStats(t *testing.T) {
	config := GetDefaultConfig()
	bus, err := NewBus(config)
	require.NoError(t, err)
	require.NotNil(t, bus)

	stats := bus.GetCircuitBreakerStats()
	assert.Equal(t, CircuitBreakerClosed, stats.State)
	assert.Equal(t, 0, stats.SuccessCount)
	assert.Equal(t, 0, stats.FailureCount)
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		wantError bool
	}{
		{
			name:      "nil config",
			config:    nil,
			wantError: true,
		},
		{
			name: "valid config",
			config: &Config{
				Enabled:         false,
				Env:             "dev",
				TopicPrefix:     "ampy",
				MaxPayloadBytes: 1024 * 1024,
				Retry: RetryConfig{
					Attempts:   5,
					BaseMs:     250,
					MaxDelayMs: 8000,
				},
				CircuitBreaker: CircuitBreakerConfig{
					Window:           50,
					FailureThreshold: 0.30,
					ResetTimeoutMs:   30000,
					HalfOpenProbes:   3,
				},
			},
			wantError: false,
		},
		{
			name: "invalid environment",
			config: &Config{
				Env: "invalid",
			},
			wantError: true,
		},
		{
			name: "empty topic prefix",
			config: &Config{
				Env:         "dev",
				TopicPrefix: "",
			},
			wantError: true,
		},
		{
			name: "payload too small",
			config: &Config{
				Env:             "dev",
				TopicPrefix:     "ampy",
				MaxPayloadBytes: 1000, // Too small
			},
			wantError: true,
		},
		{
			name: "payload too large",
			config: &Config{
				Env:             "dev",
				TopicPrefix:     "ampy",
				MaxPayloadBytes: 20 * 1024 * 1024, // Too large
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.config)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetDefaultConfig(t *testing.T) {
	config := GetDefaultConfig()

	assert.NotNil(t, config)
	assert.False(t, config.Enabled)
	assert.Equal(t, "dev", config.Env)
	assert.Equal(t, "ampy", config.TopicPrefix)
	assert.Equal(t, int64(1024*1024), config.MaxPayloadBytes)
	assert.Equal(t, "nats", config.Publisher.Backend)
	assert.Equal(t, "nats://localhost:4222", config.Publisher.NATS.URL)
}

func TestGetConfigFromEnv(t *testing.T) {
	// This test would require setting environment variables
	// For now, just test that it returns a valid config
	config := GetConfigFromEnv()

	assert.NotNil(t, config)
	assert.Equal(t, "dev", config.Env)          // Default value
	assert.Equal(t, "ampy", config.TopicPrefix) // Default value
}
