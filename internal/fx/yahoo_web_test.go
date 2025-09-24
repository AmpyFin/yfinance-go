package fx

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AmpyFin/yfinance-go/internal/norm"
)

func TestYahooWebProvider(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Mock response for EURUSD=X
		response := YahooChartResponse{
			Chart: &YahooChart{
				Result: []YahooChartResult{
					{
						Meta: &YahooChartMeta{
							RegularMarketPrice: floatPtr(1.1000),
							Currency:           "USD",
							ExchangeName:       "CCY",
							InstrumentType:     "CURRENCY",
							Timezone:           "UTC",
							ExchangeTimezone:   "UTC",
						},
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create provider with mock server URL
	config := &YahooWebConfig{
		QPS:             1.0,
		Burst:           2,
		Timeout:         5 * time.Second,
		BackoffAttempts: 3,
		BackoffBase:     100 * time.Millisecond,
		BackoffMaxDelay: 1 * time.Second,
		CircuitReset:    10 * time.Second,
	}

	provider := NewYahooWebProvider(config, 8)

	// Override the URL for testing
	// Note: In a real implementation, we'd need to modify the provider to accept a base URL
	// For now, this test demonstrates the structure

	base := "EUR"
	symbols := []string{"USD"}
	at := time.Now().UTC()

	// Test that provider is created successfully
	if provider == nil {
		t.Fatal("Failed to create yahoo-web provider")
	}

	// Test cache functionality
	cache := provider.cache
	if cache == nil {
		t.Fatal("Provider cache is nil")
	}

	// Test cache miss initially
	_, _, hit := cache.Get(base, symbols, at)
	if hit {
		t.Error("Expected cache miss initially")
	}

	// Test cache set
	rates := map[string]norm.ScaledDecimal{
		"USD": {Scaled: 110000000, Scale: 8},
	}
	cache.Set(base, symbols, at, rates, at)

	// Test cache hit
	cachedRates, cachedAsOf, hit := cache.Get(base, symbols, at)
	if !hit {
		t.Error("Expected cache hit after set")
	}

	if len(cachedRates) != 1 {
		t.Errorf("Expected 1 cached rate, got %d", len(cachedRates))
	}

	if usdRate, exists := cachedRates["USD"]; !exists {
		t.Error("Missing USD rate in cache")
	} else if usdRate.Scaled != 110000000 || usdRate.Scale != 8 {
		t.Errorf("Unexpected USD rate: %v", usdRate)
	}

	if !cachedAsOf.Equal(at) {
		t.Errorf("Expected asOf %v, got %v", at, cachedAsOf)
	}
}

func TestYahooWebProviderExtractRate(t *testing.T) {
	provider := &YahooWebProvider{}

	tests := []struct {
		name     string
		response *YahooChartResponse
		expected float64
		hasError bool
	}{
		{
			name: "regular market price available",
			response: &YahooChartResponse{
				Chart: &YahooChart{
					Result: []YahooChartResult{
						{
							Meta: &YahooChartMeta{
								RegularMarketPrice: floatPtr(1.1000),
							},
						},
					},
				},
			},
			expected: 1.1000,
			hasError: false,
		},
		{
			name: "previous close available",
			response: &YahooChartResponse{
				Chart: &YahooChart{
					Result: []YahooChartResult{
						{
							Meta: &YahooChartMeta{
								PreviousClose: floatPtr(1.0950),
							},
						},
					},
				},
			},
			expected: 1.0950,
			hasError: false,
		},
		{
			name: "chart previous close available",
			response: &YahooChartResponse{
				Chart: &YahooChart{
					Result: []YahooChartResult{
						{
							Meta: &YahooChartMeta{
								ChartPreviousClose: floatPtr(1.0900),
							},
						},
					},
				},
			},
			expected: 1.0900,
			hasError: false,
		},
		{
			name: "no chart data",
			response: &YahooChartResponse{
				Chart: &YahooChart{},
			},
			expected: 0,
			hasError: true,
		},
		{
			name: "no meta data",
			response: &YahooChartResponse{
				Chart: &YahooChart{
					Result: []YahooChartResult{
						{},
					},
				},
			},
			expected: 0,
			hasError: true,
		},
		{
			name: "no price data",
			response: &YahooChartResponse{
				Chart: &YahooChart{
					Result: []YahooChartResult{
						{
							Meta: &YahooChartMeta{},
						},
					},
				},
			},
			expected: 0,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rate, err := provider.extractRateFromResponse(tt.response)
			if tt.hasError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			if rate != tt.expected {
				t.Errorf("Expected rate %f, got %f", tt.expected, rate)
			}
		})
	}
}

func TestYahooWebProviderNewProvider(t *testing.T) {
	config := &YahooWebConfig{
		QPS:             0.5,
		Burst:           1,
		Timeout:         5 * time.Second,
		BackoffAttempts: 4,
		BackoffBase:     250 * time.Millisecond,
		BackoffMaxDelay: 4 * time.Second,
		CircuitReset:    30 * time.Second,
	}

	provider := NewYahooWebProvider(config, 8)

	if provider == nil {
		t.Fatal("Provider is nil")
	}

	if provider.config != config {
		t.Error("Config not set correctly")
	}

	if provider.httpClient == nil {
		t.Error("HTTP client is nil")
	}

	if provider.cache == nil {
		t.Error("Cache is nil")
	}

	if provider.rateScale != 8 {
		t.Errorf("Expected rate scale 8, got %d", provider.rateScale)
	}
}

// Helper function to create float pointers
func floatPtr(f float64) *float64 {
	return &f
}
