package fx

import (
	"context"
	"fmt"
	"time"

	"github.com/AmpyFin/yfinance-go/internal/norm"
)

// Manager coordinates FX operations and provider selection
type Manager struct {
	config   *Config
	provider FX
}

// NewManager creates a new FX manager
func NewManager(config *Config) (*Manager, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Validate configuration
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid FX config: %w", err)
	}

	// Create provider based on configuration
	var provider FX
	switch config.Provider {
	case "none":
		provider = NewNoneProvider()
	case "yahoo-web":
		provider = NewYahooWebProvider(&config.YahooWeb, config.RateScale)
	default:
		return nil, fmt.Errorf("unsupported FX provider: %s", config.Provider)
	}

	return &Manager{
		config:   config,
		provider: provider,
	}, nil
}

// GetRates fetches FX rates using the configured provider
func (m *Manager) GetRates(ctx context.Context, base string, symbols []string, at time.Time) (map[string]norm.ScaledDecimal, *FXMeta, error) {
	start := time.Now()
	attempts := 0

	// Create metadata
	meta := &FXMeta{
		Provider:       m.config.Provider,
		Base:           base,
		Symbols:        make([]string, len(symbols)),
		RateScale:      m.config.RateScale,
		BackoffProfile: "default",
	}

	// Copy symbols to avoid mutation
	copy(meta.Symbols, symbols)

	// Get rates from provider
	rates, asOf, err := m.provider.Rates(ctx, base, symbols, at)
	attempts++

	// Update metadata
	meta.AsOf = asOf
	meta.Attempts = attempts

	// Check if rates are stale (older than cache TTL)
	if !asOf.IsZero() && time.Since(asOf) > m.config.CacheTTL {
		meta.Stale = true
	}

	// For yahoo-web provider, check if this was a cache hit
	if _, ok := m.provider.(*YahooWebProvider); ok {
		// This is a simplified check - in a real implementation, we'd track cache hits more precisely
		meta.CacheHit = time.Since(start) < 10*time.Millisecond // Assume cache hit if very fast
	}

	return rates, meta, err
}

// ConvertValue converts a monetary value using FX rates
func (m *Manager) ConvertValue(ctx context.Context, value norm.ScaledDecimal, fromCurrency, toCurrency string, at time.Time) (norm.ScaledDecimal, *FXMeta, error) {
	if fromCurrency == toCurrency {
		// No conversion needed
		return value, &FXMeta{
			Provider:  m.config.Provider,
			Base:      fromCurrency,
			Symbols:   []string{toCurrency},
			AsOf:      at,
			RateScale: m.config.RateScale,
		}, nil
	}

	// Get FX rate
	rates, meta, err := m.GetRates(ctx, fromCurrency, []string{toCurrency}, at)
	if err != nil {
		return norm.ScaledDecimal{}, meta, err
	}

	rate, exists := rates[toCurrency]
	if !exists {
		return norm.ScaledDecimal{}, meta, fmt.Errorf("no rate available for %s/%s", fromCurrency, toCurrency)
	}

	// Convert using high-precision math
	targetScale := norm.GetPriceScaleForCurrency(toCurrency)
	converted, err := norm.MultiplyAndRound(value, rate, targetScale)
	if err != nil {
		return norm.ScaledDecimal{}, meta, fmt.Errorf("conversion failed: %w", err)
	}

	return converted, meta, nil
}

// validateConfig validates the FX configuration
func validateConfig(config *Config) error {
	// Validate provider
	validProviders := []string{"none", "yahoo-web"}
	valid := false
	for _, p := range validProviders {
		if config.Provider == p {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid provider: %s (must be one of %v)", config.Provider, validProviders)
	}

	// Validate rate scale
	if config.RateScale < 6 {
		return fmt.Errorf("rate scale too low: %d (minimum 6 recommended)", config.RateScale)
	}

	// Validate rounding mode
	if config.Rounding != "half_up" {
		return fmt.Errorf("invalid rounding mode: %s (only 'half_up' supported)", config.Rounding)
	}

	// Validate yahoo-web specific config if provider is yahoo-web
	if config.Provider == "yahoo-web" {
		if config.YahooWeb.QPS <= 0 {
			return fmt.Errorf("yahoo-web QPS must be positive")
		}
		if config.YahooWeb.Burst <= 0 {
			return fmt.Errorf("yahoo-web burst must be positive")
		}
		if config.YahooWeb.Timeout <= 0 {
			return fmt.Errorf("yahoo-web timeout must be positive")
		}
	}

	return nil
}
