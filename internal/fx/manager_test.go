package fx

import (
	"context"
	"testing"
	"time"

	"github.com/AmpyFin/yfinance-go/internal/norm"
)

func TestNewManager(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
	}{
		{
			name:        "default config",
			config:      nil,
			expectError: false,
		},
		{
			name: "valid none provider",
			config: &Config{
				Provider:  "none",
				RateScale: 8,
				Rounding:  "half_up",
			},
			expectError: false,
		},
		{
			name: "valid yahoo-web provider",
			config: &Config{
				Provider:  "yahoo-web",
				RateScale: 8,
				Rounding:  "half_up",
				YahooWeb: YahooWebConfig{
					QPS:             0.5,
					Burst:           1,
					Timeout:         5 * time.Second,
					BackoffAttempts: 4,
					BackoffBase:     250 * time.Millisecond,
					BackoffMaxDelay: 4 * time.Second,
					CircuitReset:    30 * time.Second,
				},
			},
			expectError: false,
		},
		{
			name: "invalid provider",
			config: &Config{
				Provider:  "invalid",
				RateScale: 8,
				Rounding:  "half_up",
			},
			expectError: true,
		},
		{
			name: "rate scale too low",
			config: &Config{
				Provider:  "none",
				RateScale: 4,
				Rounding:  "half_up",
			},
			expectError: true,
		},
		{
			name: "invalid rounding",
			config: &Config{
				Provider:  "none",
				RateScale: 8,
				Rounding:  "invalid",
			},
			expectError: true,
		},
		{
			name: "yahoo-web with invalid QPS",
			config: &Config{
				Provider:  "yahoo-web",
				RateScale: 8,
				Rounding:  "half_up",
				YahooWeb: YahooWebConfig{
					QPS:     0,
					Burst:   1,
					Timeout: 5 * time.Second,
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager, err := NewManager(tt.config)
			if tt.expectError {
				if err == nil {
					t.Errorf("NewManager() expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("NewManager() unexpected error: %v", err)
				return
			}
			if manager == nil {
				t.Error("NewManager() returned nil manager")
			}
		})
	}
}

func TestManagerGetRates(t *testing.T) {
	// Test with none provider
	config := &Config{
		Provider:  "none",
		RateScale: 8,
		Rounding:  "half_up",
	}

	manager, err := NewManager(config)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	ctx := context.Background()
	base := "EUR"
	symbols := []string{"USD"}
	at := time.Now().UTC()

	// Should return error for none provider
	_, meta, err := manager.GetRates(ctx, base, symbols, at)
	if err == nil {
		t.Error("Expected error for none provider")
	}

	// Verify metadata
	if meta.Provider != "none" {
		t.Errorf("Expected provider 'none', got '%s'", meta.Provider)
	}
	if meta.Base != base {
		t.Errorf("Expected base '%s', got '%s'", base, meta.Base)
	}
	if len(meta.Symbols) != len(symbols) || meta.Symbols[0] != symbols[0] {
		t.Errorf("Expected symbols %v, got %v", symbols, meta.Symbols)
	}
}

func TestManagerConvertValue(t *testing.T) {
	config := &Config{
		Provider:  "none",
		RateScale: 8,
		Rounding:  "half_up",
	}

	manager, err := NewManager(config)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	ctx := context.Background()
	value := norm.ScaledDecimal{Scaled: 10000, Scale: 4} // 1.0000
	fromCurrency := "EUR"
	toCurrency := "USD"
	at := time.Now().UTC()

	// Should return error for none provider
	_, meta, err := manager.ConvertValue(ctx, value, fromCurrency, toCurrency, at)
	if err == nil {
		t.Error("Expected error for none provider")
	}

	// Verify metadata
	if meta.Provider != "none" {
		t.Errorf("Expected provider 'none', got '%s'", meta.Provider)
	}
	if meta.Base != fromCurrency {
		t.Errorf("Expected base '%s', got '%s'", fromCurrency, meta.Base)
	}
	if len(meta.Symbols) != 1 || meta.Symbols[0] != toCurrency {
		t.Errorf("Expected symbols [%s], got %v", toCurrency, meta.Symbols)
	}
}

func TestManagerConvertValueSameCurrency(t *testing.T) {
	config := &Config{
		Provider:  "none",
		RateScale: 8,
		Rounding:  "half_up",
	}

	manager, err := NewManager(config)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	ctx := context.Background()
	value := norm.ScaledDecimal{Scaled: 10000, Scale: 4} // 1.0000
	currency := "USD"
	at := time.Now().UTC()

	// Should succeed for same currency
	converted, meta, err := manager.ConvertValue(ctx, value, currency, currency, at)
	if err != nil {
		t.Errorf("Unexpected error for same currency: %v", err)
	}

	// Verify value is unchanged
	if converted.Scaled != value.Scaled || converted.Scale != value.Scale {
		t.Errorf("Expected unchanged value %v, got %v", value, converted)
	}

	// Verify metadata
	if meta.Provider != "none" {
		t.Errorf("Expected provider 'none', got '%s'", meta.Provider)
	}
	if meta.Base != currency {
		t.Errorf("Expected base '%s', got '%s'", currency, meta.Base)
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Provider != "none" {
		t.Errorf("Expected provider 'none', got '%s'", config.Provider)
	}
	if config.Target != "" {
		t.Errorf("Expected empty target, got '%s'", config.Target)
	}
	if config.CacheTTL != 60*time.Second {
		t.Errorf("Expected cache TTL 60s, got %v", config.CacheTTL)
	}
	if config.RateScale != 8 {
		t.Errorf("Expected rate scale 8, got %d", config.RateScale)
	}
	if config.Rounding != "half_up" {
		t.Errorf("Expected rounding 'half_up', got '%s'", config.Rounding)
	}

	// Check yahoo-web config
	if config.YahooWeb.QPS != 0.5 {
		t.Errorf("Expected yahoo-web QPS 0.5, got %f", config.YahooWeb.QPS)
	}
	if config.YahooWeb.Burst != 1 {
		t.Errorf("Expected yahoo-web burst 1, got %d", config.YahooWeb.Burst)
	}
	if config.YahooWeb.Timeout != 5*time.Second {
		t.Errorf("Expected yahoo-web timeout 5s, got %v", config.YahooWeb.Timeout)
	}
}
