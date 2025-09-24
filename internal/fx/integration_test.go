package fx

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/AmpyFin/yfinance-go/internal/norm"
)

// TestYahooWebProviderIntegration tests the yahoo-web provider with a fake server
func TestYahooWebProviderIntegration(t *testing.T) {
	// Create provider with test configuration
	config := &YahooWebConfig{
		QPS:               0.5,
		Burst:             1,
		Timeout:           5 * time.Second,
		BackoffAttempts:   3,
		BackoffBase:       250 * time.Millisecond,
		BackoffMaxDelay:   2 * time.Second,
		CircuitReset:      30 * time.Second,
	}
	
	provider := NewYahooWebProvider(config, 8)
	
	// Test rate extraction from mock response
	mockResponse := &YahooChartResponse{
		Chart: &YahooChart{
			Result: []YahooChartResult{
				{
					Meta: &YahooChartMeta{
						RegularMarketPrice: func() *float64 { v := 1.0850; return &v }(),
						Currency:           "USD",
						ExchangeName:       "CCY",
						InstrumentType:     "CURRENCY",
						Timezone:           "UTC",
						ExchangeTimezone:   "UTC",
						GMTOffset:          0,
						LastUpdateTime:     func() *int64 { v := time.Now().Unix(); return &v }(),
						RegularMarketTime:  func() *int64 { v := time.Now().Unix(); return &v }(),
					},
				},
			},
			Error: nil,
		},
	}
	
	rate, err := provider.extractRateFromResponse(mockResponse)
	if err != nil {
		t.Fatalf("Failed to extract rate: %v", err)
	}
	
	expectedRate := 1.0850
	if rate != expectedRate {
		t.Errorf("Expected rate %f, got %f", expectedRate, rate)
	}
	
	// Test conversion to scaled decimal
	scaledRate, err := norm.ToScaledDecimal(rate, 8)
	if err != nil {
		t.Fatalf("Failed to convert to scaled decimal: %v", err)
	}
	
	expectedScaled := int64(108500000) // 1.0850 * 10^8
	if scaledRate.Scaled != expectedScaled {
		t.Errorf("Expected scaled rate %d, got %d", expectedScaled, scaledRate.Scaled)
	}
	
	if scaledRate.Scale != 8 {
		t.Errorf("Expected scale 8, got %d", scaledRate.Scale)
	}
}

// TestFXManagerIntegration tests the FX manager with different providers
func TestFXManagerIntegration(t *testing.T) {
	// Test with none provider
	config := &Config{
		Provider:   "none",
		Target:     "USD",
		RateScale:  8,
		Rounding:   "half_up",
	}
	
	manager, err := NewManager(config)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	
	// Test that none provider returns error
	_, _, err = manager.GetRates(context.Background(), "EUR", []string{"USD"}, time.Now())
	if err == nil {
		t.Error("Expected error from none provider, got nil")
	}
	
	// Test with yahoo-web provider (but don't actually make HTTP calls)
	config.Provider = "yahoo-web"
	config.YahooWeb = YahooWebConfig{
		QPS:               0.5,
		Burst:             1,
		Timeout:           5 * time.Second,
		BackoffAttempts:   3,
		BackoffBase:       250 * time.Millisecond,
		BackoffMaxDelay:   2 * time.Second,
		CircuitReset:      30 * time.Second,
	}
	
	manager, err = NewManager(config)
	if err != nil {
		t.Fatalf("Failed to create manager with yahoo-web: %v", err)
	}
	
	// Test that yahoo-web provider is created (but will fail on actual HTTP call)
	// This tests the provider instantiation logic
	if manager == nil {
		t.Error("Expected manager to be created")
	}
}

// TestFXCacheIntegration tests the FX cache with realistic usage patterns
func TestFXCacheIntegration(t *testing.T) {
	cache := NewFXCache(1 * time.Second) // 1 second TTL
	
	base := "EUR"
	symbols := []string{"USD", "JPY"}
	at := time.Now()
	
	// Create mock rates
	rates := map[string]norm.ScaledDecimal{
		"USD": {Scaled: 108500000, Scale: 8}, // 1.0850
		"JPY": {Scaled: 162750000, Scale: 8}, // 1.6275
	}
	asOf := time.Now()
	
	// Test cache miss
	if _, _, hit := cache.Get(base, symbols, at); hit {
		t.Error("Expected cache miss, got hit")
	}
	
	// Set cache
	cache.Set(base, symbols, at, rates, asOf)
	
	// Test cache hit
	cachedRates, cachedAsOf, hit := cache.Get(base, symbols, at)
	if !hit {
		t.Error("Expected cache hit, got miss")
	}
	
	if len(cachedRates) != len(rates) {
		t.Errorf("Expected %d rates, got %d", len(rates), len(cachedRates))
	}
	
	if cachedAsOf != asOf {
		t.Errorf("Expected asOf %v, got %v", asOf, cachedAsOf)
	}
	
	// Test cache expiration
	time.Sleep(2 * time.Second)
	if _, _, hit := cache.Get(base, symbols, at); hit {
		t.Error("Expected cache miss after expiration, got hit")
	}
}

// TestFXConversionIntegration tests the complete FX conversion flow
func TestFXConversionIntegration(t *testing.T) {
	// Create a mock FX converter for testing
	mockConverter := &MockFXConverter{
		rates: map[string]map[string]norm.ScaledDecimal{
			"EUR": {
				"USD": {Scaled: 108500000, Scale: 8}, // 1.0850
			},
		},
		asOf: time.Now(),
	}
	
	// Create a normalized bar batch
	bars := &norm.NormalizedBarBatch{
		Security: norm.Security{Symbol: "SAP"},
		Bars: []norm.NormalizedBar{
			{
				Start:         time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				End:           time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
				Open:          norm.ScaledDecimal{Scaled: 1200000, Scale: 4}, // 120.00 EUR
				High:          norm.ScaledDecimal{Scaled: 1210000, Scale: 4}, // 121.00 EUR
				Low:           norm.ScaledDecimal{Scaled: 1190000, Scale: 4}, // 119.00 EUR
				Close:         norm.ScaledDecimal{Scaled: 1205000, Scale: 4}, // 120.50 EUR
				Volume:        1000000,
				CurrencyCode:  "EUR",
				Adjusted:      true,
				AdjustmentPolicyID: "split_dividend",
				EventTime:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				IngestTime:    time.Now(),
				AsOf:          time.Now(),
			},
		},
		Meta: norm.Meta{
			RunID: "test-run",
		},
	}
	
	// Convert to USD
	convertedBars, fxMeta, err := bars.ConvertTo(context.Background(), "USD", mockConverter)
	if err != nil {
		t.Fatalf("Failed to convert bars: %v", err)
	}
	
	if convertedBars == nil {
		t.Fatal("Expected converted bars, got nil")
	}
	
	if fxMeta == nil {
		t.Fatal("Expected FX metadata, got nil")
	}
	
	// Verify conversion
	convertedBar := convertedBars.Bars[0]
	
	// Expected: 120.50 EUR * 1.0850 = 130.7425 USD
	expectedClose := int64(1307425) // 130.7425 * 10^4
	if convertedBar.Close.Scaled != expectedClose {
		t.Errorf("Expected converted close %d, got %d", expectedClose, convertedBar.Close.Scaled)
	}
	
	if convertedBar.OriginalCurrency != "EUR" {
		t.Errorf("Expected original currency EUR, got %s", convertedBar.OriginalCurrency)
	}
	
	if convertedBar.ConvertedCurrency != "USD" {
		t.Errorf("Expected converted currency USD, got %s", convertedBar.ConvertedCurrency)
	}
	
	if fxMeta.Provider != "mock" {
		t.Errorf("Expected provider mock, got %s", fxMeta.Provider)
	}
}

// MockFXConverter implements the FXConverter interface for testing
type MockFXConverter struct {
	rates map[string]map[string]norm.ScaledDecimal
	asOf  time.Time
}

func (m *MockFXConverter) Rates(ctx context.Context, base string, symbols []string, at time.Time) (map[string]norm.ScaledDecimal, time.Time, error) {
	result := make(map[string]norm.ScaledDecimal)
	
	if baseRates, exists := m.rates[base]; exists {
		for _, symbol := range symbols {
			if rate, exists := baseRates[symbol]; exists {
				result[symbol] = rate
			}
		}
	}
	
	return result, m.asOf, nil
}

func (m *MockFXConverter) ConvertValue(ctx context.Context, value norm.ScaledDecimal, fromCurrency, toCurrency string, at time.Time) (norm.ScaledDecimal, *norm.FXMeta, error) {
	if fromCurrency == toCurrency {
		return value, &norm.FXMeta{
			Provider:  "mock",
			Base:      fromCurrency,
			Symbols:   []string{toCurrency},
			AsOf:      m.asOf,
			RateScale: 8,
			CacheHit:  false,
			Attempts:  1,
		}, nil
	}
	
	rates, _, err := m.Rates(ctx, fromCurrency, []string{toCurrency}, at)
	if err != nil {
		return norm.ScaledDecimal{}, nil, err
	}
	
	rate, exists := rates[toCurrency]
	if !exists {
		return norm.ScaledDecimal{}, nil, fmt.Errorf("no rate available for %s/%s", fromCurrency, toCurrency)
	}
	
	// Convert using the rate
	converted, err := norm.MultiplyAndRound(value, rate, 4) // Target scale 4 for USD
	if err != nil {
		return norm.ScaledDecimal{}, nil, err
	}
	
	return converted, &norm.FXMeta{
		Provider:  "mock",
		Base:      fromCurrency,
		Symbols:   []string{toCurrency},
		AsOf:      m.asOf,
		RateScale: 8,
		CacheHit:  false,
		Attempts:  1,
	}, nil
}
