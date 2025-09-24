package emit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/AmpyFin/yfinance-go/internal/norm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmitBarBatch_Golden(t *testing.T) {
	tests := []struct {
		name     string
		golden   string
		input    *norm.NormalizedBarBatch
	}{
		{
			name:   "AAPL_1d_adjusted",
			golden: "testdata/golden/ampy/bars/AAPL_1d_adjusted.json",
			input: &norm.NormalizedBarBatch{
				Security: norm.Security{
					Symbol: "AAPL",
					MIC:    "XNAS",
				},
				Bars: []norm.NormalizedBar{
					{
						Start:              time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
						End:                time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC),
						Open:               norm.ScaledDecimal{Scaled: 1892300, Scale: 4},
						High:               norm.ScaledDecimal{Scaled: 1910000, Scale: 4},
						Low:                norm.ScaledDecimal{Scaled: 1889000, Scale: 4},
						Close:              norm.ScaledDecimal{Scaled: 1904500, Scale: 4},
						CurrencyCode:       "USD",
						Volume:             43210000,
						Adjusted:           true,
						AdjustmentPolicyID: "split_dividend",
						EventTime:          time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC),
						IngestTime:         time.Date(2024, 1, 3, 0, 0, 1, 0, time.UTC),
						AsOf:               time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC),
					},
				},
				Meta: norm.Meta{
					RunID:         "golden_bars_v1",
					Source:        "yfinance-go",
					Producer:      "local",
					SchemaVersion: "ampy.bars.v1:1.0.0",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Emit the bar batch
			barBatch, err := EmitBarBatch(tt.input)
			require.NoError(t, err)

			// Convert to golden format
			goldenBarBatch := ToGoldenBarBatch(barBatch, tt.input.Security, tt.input.Meta)

			// Convert to canonical JSON
			actualJSON, err := MarshalToGoldenJSON(goldenBarBatch)
			require.NoError(t, err)

			// Read golden file
			goldenPath := filepath.Join("../../", tt.golden)
			goldenData, err := os.ReadFile(goldenPath)
			require.NoError(t, err)

			// Parse golden JSON to normalize it
			var goldenObj map[string]interface{}
			err = json.Unmarshal(goldenData, &goldenObj)
			require.NoError(t, err)

			// Convert golden to canonical JSON
			goldenCanonical, err := CanonicalMarshaler.Marshal(goldenObj)
			require.NoError(t, err)

			// Compare
			assert.Equal(t, string(goldenCanonical), string(actualJSON), "Emitted JSON should match golden file")
		})
	}
}

func TestEmitQuote_Golden(t *testing.T) {
	tests := []struct {
		name     string
		golden   string
		input    *norm.NormalizedQuote
	}{
		{
			name:   "MSFT_snapshot_quote",
			golden: "testdata/golden/ampy/quotes/MSFT_snapshot_quote.json",
			input: &norm.NormalizedQuote{
				Security: norm.Security{
					Symbol: "MSFT",
					MIC:    "XNAS",
				},
				Type:                "QUOTE",
				Bid:                 &norm.ScaledDecimal{Scaled: 4275000, Scale: 4},
				BidSize:             int64Ptr(200),
				Ask:                 &norm.ScaledDecimal{Scaled: 4275300, Scale: 4},
				AskSize:             int64Ptr(300),
				CurrencyCode:        "USD",
				Venue:               "XNMS",
				EventTime:           time.Date(2024, 1, 3, 15, 30, 12, 0, time.UTC),
				IngestTime:          time.Date(2024, 1, 3, 15, 30, 12, 0, time.UTC),
				Meta: norm.Meta{
					RunID:         "golden_quote_v1",
					Source:        "yfinance-go",
					Producer:      "local",
					SchemaVersion: "ampy.ticks.v1:1.0.0",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Emit the quote
			quote, err := EmitQuote(tt.input)
			require.NoError(t, err)

			// Convert to golden format
			goldenQuote := ToGoldenQuote(quote, tt.input.Meta)

			// Convert to canonical JSON
			actualJSON, err := MarshalToGoldenJSON(goldenQuote)
			require.NoError(t, err)

			// Read golden file
			goldenPath := filepath.Join("../../", tt.golden)
			goldenData, err := os.ReadFile(goldenPath)
			require.NoError(t, err)

			// Parse golden JSON to normalize it
			var goldenObj map[string]interface{}
			err = json.Unmarshal(goldenData, &goldenObj)
			require.NoError(t, err)

			// Convert golden to canonical JSON
			goldenCanonical, err := CanonicalMarshaler.Marshal(goldenObj)
			require.NoError(t, err)

			// Compare
			assert.Equal(t, string(goldenCanonical), string(actualJSON), "Emitted JSON should match golden file")
		})
	}
}

func TestEmitFundamentals_Golden(t *testing.T) {
	tests := []struct {
		name     string
		golden   string
		input    *norm.NormalizedFundamentalsSnapshot
	}{
		{
			name:   "AAPL_quarterly_snapshot",
			golden: "testdata/golden/ampy/fundamentals/AAPL_quarterly_snapshot.json",
			input: &norm.NormalizedFundamentalsSnapshot{
				Security: norm.Security{
					Symbol: "AAPL",
					MIC:    "XNAS",
				},
				Lines: []norm.NormalizedFundamentalsLine{
					{
						Key:          "revenue",
						Value:        norm.ScaledDecimal{Scaled: 119870000000000, Scale: 2},
						CurrencyCode: "USD",
						PeriodStart:  time.Date(2025, 3, 30, 0, 0, 0, 0, time.UTC),
						PeriodEnd:    time.Date(2025, 6, 29, 0, 0, 0, 0, time.UTC),
					},
					{
						Key:          "net_income",
						Value:        norm.ScaledDecimal{Scaled: 2386000000000, Scale: 2},
						CurrencyCode: "USD",
						PeriodStart:  time.Date(2025, 3, 30, 0, 0, 0, 0, time.UTC),
						PeriodEnd:    time.Date(2025, 6, 29, 0, 0, 0, 0, time.UTC),
					},
					{
						Key:          "eps_basic",
						Value:        norm.ScaledDecimal{Scaled: 1525, Scale: 2},
						CurrencyCode: "USD",
						PeriodStart:  time.Date(2025, 3, 30, 0, 0, 0, 0, time.UTC),
						PeriodEnd:    time.Date(2025, 6, 29, 0, 0, 0, 0, time.UTC),
					},
				},
				Source: "yfinance",
				AsOf:   time.Date(2025, 8, 1, 0, 0, 0, 0, time.UTC),
				Meta: norm.Meta{
					RunID:         "golden_fund_v1",
					Source:        "yfinance-go",
					Producer:      "local",
					SchemaVersion: "ampy.fundamentals.v1:1.0.0",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Emit the fundamentals
			fundamentals, err := EmitFundamentals(tt.input)
			require.NoError(t, err)

			// Convert to golden format
			goldenFundamentals := ToGoldenFundamentals(fundamentals, tt.input.Meta)

			// Convert to canonical JSON
			actualJSON, err := MarshalToGoldenJSON(goldenFundamentals)
			require.NoError(t, err)

			// Read golden file
			goldenPath := filepath.Join("../../", tt.golden)
			goldenData, err := os.ReadFile(goldenPath)
			require.NoError(t, err)

			// Parse golden JSON to normalize it
			var goldenObj map[string]interface{}
			err = json.Unmarshal(goldenData, &goldenObj)
			require.NoError(t, err)

			// Convert golden to canonical JSON
			goldenCanonical, err := CanonicalMarshaler.Marshal(goldenObj)
			require.NoError(t, err)

			// Compare
			assert.Equal(t, string(goldenCanonical), string(actualJSON), "Emitted JSON should match golden file")
		})
	}
}

// Helper function to create int64 pointer
func int64Ptr(v int64) *int64 {
	return &v
}
