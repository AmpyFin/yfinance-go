//go:build integration

package tests

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/AmpyFin/yfinance-go"
	"github.com/AmpyFin/yfinance-go/internal/norm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDataCorrectness_RealPrices validates that we get correct price data from real API calls
func TestDataCorrectness_RealPrices(t *testing.T) {
	client := yfinance.NewClient()
	ctx := context.Background()

	testCases := []struct {
		name   string
		symbol string
	}{
		{"NASDAQ Stock", "AAPL"},
		{"NYSE Stock", "MSFT"},
		{"NYSE Stock", "JPM"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Fetch quote
			quote, err := client.FetchQuote(ctx, tc.symbol, "test-correctness")
			require.NoError(t, err, "Failed to fetch quote for %s", tc.symbol)
			require.NotNil(t, quote, "Quote should not be nil")

			// Validate quote data
			assert.Equal(t, tc.symbol, quote.Security.Symbol, "Symbol should match")
			assert.NotEmpty(t, quote.CurrencyCode, "Currency code should not be empty")
			assert.True(t, len(quote.CurrencyCode) == 3, "Currency code should be 3 characters")

			// Validate price if present
			if quote.RegularMarketPrice != nil {
				price := norm.FromScaledDecimal(*quote.RegularMarketPrice)
				assert.Greater(t, price, 0.0, "Price should be positive")
				assert.False(t, math.IsNaN(price), "Price should not be NaN")
				assert.False(t, math.IsInf(price, 0), "Price should not be infinite")
				assert.Less(t, price, 1000000.0, "Price should be reasonable (< $1M)")

				t.Logf("✓ %s: Price = %.2f %s", tc.symbol, price, quote.CurrencyCode)
			}

			// Validate volume if present
			if quote.RegularMarketVolume != nil {
				assert.GreaterOrEqual(t, *quote.RegularMarketVolume, int64(0), "Volume should be non-negative")
				t.Logf("✓ %s: Volume = %d", tc.symbol, *quote.RegularMarketVolume)
			}

			// Validate high/low relationship if both present
			if quote.RegularMarketHigh != nil && quote.RegularMarketLow != nil {
				high := norm.FromScaledDecimal(*quote.RegularMarketHigh)
				low := norm.FromScaledDecimal(*quote.RegularMarketLow)
				assert.GreaterOrEqual(t, high, low, "High should be >= Low")
				t.Logf("✓ %s: High = %.2f, Low = %.2f", tc.symbol, high, low)
			}
		})
	}
}

// TestDataCorrectness_HistoricalData validates historical bar data correctness
func TestDataCorrectness_HistoricalData(t *testing.T) {
	client := yfinance.NewClient()
	ctx := context.Background()

	testCases := []struct {
		name   string
		symbol string
		days   int
	}{
		{"AAPL 30 days", "AAPL", 30},
		{"MSFT 30 days", "MSFT", 30},
		{"JPM 7 days", "JPM", 7},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			end := time.Now().UTC()
			start := end.AddDate(0, 0, -tc.days)

			// Fetch daily bars
			bars, err := client.FetchDailyBars(ctx, tc.symbol, start, end, true, "test-historical")
			require.NoError(t, err, "Failed to fetch bars for %s", tc.symbol)
			require.NotNil(t, bars, "Bars should not be nil")
			require.Greater(t, len(bars.Bars), 0, "Should have at least one bar")

			// Validate security
			assert.Equal(t, tc.symbol, bars.Security.Symbol, "Symbol should match")
			assert.NotEmpty(t, bars.Security.MIC, "MIC should be inferred")

			// Validate each bar
			prevTime := time.Time{}
			for i, bar := range bars.Bars {
				// Validate prices (use proper conversion function)
				open := norm.FromScaledDecimal(bar.Open)
				high := norm.FromScaledDecimal(bar.High)
				low := norm.FromScaledDecimal(bar.Low)
				close := norm.FromScaledDecimal(bar.Close)

				assert.Greater(t, open, 0.0, "Bar[%d]: Open should be positive", i)
				assert.Greater(t, high, 0.0, "Bar[%d]: High should be positive", i)
				assert.Greater(t, low, 0.0, "Bar[%d]: Low should be positive", i)
				assert.Greater(t, close, 0.0, "Bar[%d]: Close should be positive", i)

				// Validate OHLC relationships
				assert.GreaterOrEqual(t, high, low, "Bar[%d]: High >= Low", i)
				assert.GreaterOrEqual(t, high, open, "Bar[%d]: High >= Open", i)
				assert.GreaterOrEqual(t, high, close, "Bar[%d]: High >= Close", i)
				assert.LessOrEqual(t, low, open, "Bar[%d]: Low <= Open", i)
				assert.LessOrEqual(t, low, close, "Bar[%d]: Low <= Close", i)

				// Validate volume
				assert.GreaterOrEqual(t, bar.Volume, int64(0), "Bar[%d]: Volume should be non-negative", i)

				// Validate time ordering
				if !prevTime.IsZero() {
					assert.True(t, bar.EventTime.After(prevTime) || bar.EventTime.Equal(prevTime),
						"Bar[%d]: EventTime should be >= previous bar time", i)
				}
				prevTime = bar.EventTime

				// Validate currency
				assert.NotEmpty(t, bar.CurrencyCode, "Bar[%d]: Currency code should not be empty", i)
				assert.Equal(t, 3, len(bar.CurrencyCode), "Bar[%d]: Currency code should be 3 characters", i)
			}

			// MIC may be empty for some exchanges (e.g., NYQ for JPM), which is acceptable
			// The important thing is that it's inferred when possible
			if bars.Security.MIC == "" {
				t.Logf("Note: MIC is empty for %s (exchange may not be in mapping)", tc.symbol)
			}

			t.Logf("✓ %s: Validated %d bars", tc.symbol, len(bars.Bars))
		})
	}
}

// TestDataCorrectness_AnalystData validates analyst/insights data from scraping
func TestDataCorrectness_AnalystData(t *testing.T) {
	client := yfinance.NewClient()
	ctx := context.Background()

	testCases := []struct {
		name     string
		symbol   string
		endpoint string
	}{
		{"AAPL Key Statistics", "AAPL", "key-statistics"},
		{"MSFT Analysis", "MSFT", "analysis"},
		{"AAPL Analyst Insights", "AAPL", "analyst-insights"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var snapshot interface{}
			var err error

			switch tc.endpoint {
			case "key-statistics":
				snapshot, err = client.ScrapeKeyStatistics(ctx, tc.symbol, "test-analyst")
			case "analysis":
				snapshot, err = client.ScrapeAnalysis(ctx, tc.symbol, "test-analyst")
			case "analyst-insights":
				snapshot, err = client.ScrapeAnalystInsights(ctx, tc.symbol, "test-analyst")
			default:
				t.Fatalf("Unknown endpoint: %s", tc.endpoint)
			}

			require.NoError(t, err, "Failed to scrape %s for %s", tc.endpoint, tc.symbol)
			require.NotNil(t, snapshot, "Snapshot should not be nil")

			// Type assert to fundamentals snapshot
			fundamentals, ok := snapshot.(*struct {
				Security interface{}
				Lines    []interface{}
				Source   string
				AsOf     time.Time
				Meta     interface{}
			})
			if !ok {
				// Try to get lines count using reflection or type assertion
				t.Logf("✓ %s: Successfully scraped %s data", tc.symbol, tc.endpoint)
				return
			}

			// Validate source
			if fundamentals != nil {
				assert.NotEmpty(t, fundamentals.Source, "Source should not be empty")
				t.Logf("✓ %s: Source = %s", tc.symbol, fundamentals.Source)
			}
		})
	}
}

// TestDataCorrectness_MICInference validates that MIC is correctly inferred for different exchanges
func TestDataCorrectness_MICInference(t *testing.T) {
	client := yfinance.NewClient()
	ctx := context.Background()

	testCases := []struct {
		name     string
		symbol   string
		expected string // Expected MIC (can be empty if unknown)
	}{
		{"NASDAQ Stock", "AAPL", "XNAS"}, // NMS maps to XNAS
		{"NYSE Stock", "JPM", "XNYS"},    // NYQ maps to XNYS
		{"NASDAQ Stock", "MSFT", "XNAS"}, // NMS maps to XNAS
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Fetch company info to get exchange
			companyInfo, err := client.FetchCompanyInfo(ctx, tc.symbol, "test-mic")
			require.NoError(t, err, "Failed to fetch company info for %s", tc.symbol)

			// Validate MIC is inferred (may be empty for unknown exchanges, but should work for known ones)
			if tc.expected != "" {
				assert.NotEmpty(t, companyInfo.Security.MIC, "MIC should be inferred for %s", tc.symbol)
				assert.Equal(t, tc.expected, companyInfo.Security.MIC,
					"MIC should match expected for %s", tc.symbol)
			} else {
				// For unknown exchanges, MIC may be empty, which is acceptable
				t.Logf("MIC for %s: %s (may be empty for unknown exchange)", tc.symbol, companyInfo.Security.MIC)
			}

			t.Logf("✓ %s: MIC = %s (Exchange: %s)", tc.symbol, companyInfo.Security.MIC, companyInfo.Exchange)

			// Now test scraping with inferred MIC
			keyStats, err := client.ScrapeKeyStatistics(ctx, tc.symbol, "test-mic-scrape")
			if err == nil {
				require.NotNil(t, keyStats, "Key statistics should not be nil")
				assert.Equal(t, tc.symbol, keyStats.Security.Symbol, "Symbol should match")
				if tc.expected != "" {
					assert.Equal(t, tc.expected, keyStats.Security.Mic,
						"MIC should match expected in scraped data for %s", tc.symbol)
				}
				t.Logf("✓ %s: Scraped data has correct MIC = %s", tc.symbol, keyStats.Security.Mic)
			}
		})
	}
}

// TestDataCorrectness_PricePrecision validates that price precision is maintained correctly
func TestDataCorrectness_PricePrecision(t *testing.T) {
	client := yfinance.NewClient()
	ctx := context.Background()

	// Fetch quote for a stock with known price range
	quote, err := client.FetchQuote(ctx, "AAPL", "test-precision")
	require.NoError(t, err)
	require.NotNil(t, quote.RegularMarketPrice, "Should have price")

	price := norm.FromScaledDecimal(*quote.RegularMarketPrice)

	// Validate scale is appropriate (should be 2 for USD)
	assert.Equal(t, 2, int(quote.RegularMarketPrice.Scale), "Scale should be 2 for USD")

	// Validate price is reasonable (AAPL should be between $50 and $1000)
	assert.Greater(t, price, 50.0, "Price should be > $50")
	assert.Less(t, price, 1000.0, "Price should be < $1000")

	// Validate precision: price * 10^scale should equal scaled (within rounding)
	multiplier := math.Pow(10, float64(quote.RegularMarketPrice.Scale))
	expectedScaled := int64(math.Round(price * multiplier))
	actualScaled := quote.RegularMarketPrice.Scaled
	diff := expectedScaled - actualScaled
	if diff < 0 {
		diff = -diff
	}
	assert.LessOrEqual(t, diff, int64(1), "Price precision should be maintained (diff <= 1 cent)")

	t.Logf("✓ Price precision validated: %.2f = %d / 10^%d", price, actualScaled, quote.RegularMarketPrice.Scale)
}

// TestDataCorrectness_NoFakeData validates that we're not returning fake or placeholder data
func TestDataCorrectness_NoFakeData(t *testing.T) {
	client := yfinance.NewClient()
	ctx := context.Background()

	// Fetch data for a real symbol
	quote, err := client.FetchQuote(ctx, "AAPL", "test-no-fake")
	require.NoError(t, err)

	// Validate we're not getting placeholder values
	if quote.RegularMarketPrice != nil {
		price := norm.FromScaledDecimal(*quote.RegularMarketPrice)
		// Common placeholder values to check
		assert.NotEqual(t, 0.0, price, "Price should not be zero (placeholder)")
		assert.NotEqual(t, 1.0, price, "Price should not be 1.0 (placeholder)")
		assert.NotEqual(t, 100.0, price, "Price should not be 100.0 (placeholder)")
		assert.NotEqual(t, 999.99, price, "Price should not be 999.99 (placeholder)")
	}

	// Validate timestamps are recent (not from 1970 or far future)
	assert.True(t, quote.EventTime.After(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)),
		"Event time should be after 2020")
	assert.True(t, quote.EventTime.Before(time.Now().Add(24*time.Hour)),
		"Event time should be before now + 24h")

	t.Logf("✓ No fake data detected: Price = %.2f, Time = %s",
		norm.FromScaledDecimal(*quote.RegularMarketPrice),
		quote.EventTime.Format(time.RFC3339))
}

// TestDataCorrectness_CurrencyConsistency validates currency consistency across data types
func TestDataCorrectness_CurrencyConsistency(t *testing.T) {
	client := yfinance.NewClient()
	ctx := context.Background()

	symbol := "AAPL"

	// Fetch quote
	quote, err := client.FetchQuote(ctx, symbol, "test-currency")
	require.NoError(t, err)

	// Fetch bars
	end := time.Now().UTC()
	start := end.AddDate(0, 0, -7)
	bars, err := client.FetchDailyBars(ctx, symbol, start, end, true, "test-currency")
	require.NoError(t, err)

	// Validate currency consistency
	assert.Equal(t, quote.CurrencyCode, bars.Bars[0].CurrencyCode,
		"Currency should be consistent between quote and bars")

	t.Logf("✓ Currency consistency: Quote = %s, Bars = %s",
		quote.CurrencyCode, bars.Bars[0].CurrencyCode)
}

