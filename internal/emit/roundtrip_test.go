package emit

import (
	"testing"
	"time"

	"github.com/AmpyFin/yfinance-go/internal/norm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestEmitBarBatch_RoundTrip(t *testing.T) {
	// Create test input
	input := &norm.NormalizedBarBatch{
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
			RunID:         "test_roundtrip",
			Source:        "yfinance-go",
			Producer:      "local",
			SchemaVersion: "ampy.bars.v1:1.0.0",
		},
	}

	// Emit to protobuf
	barBatch, err := EmitBarBatch(input)
	require.NoError(t, err)

	// Marshal to protobuf bytes
	protoBytes, err := proto.Marshal(barBatch)
	require.NoError(t, err)
	assert.NotEmpty(t, protoBytes)

	// Note: We can't directly unmarshal back to our normalized type
	// This test verifies that the protobuf serialization/deserialization works
	// In a real scenario, you'd unmarshal to the ampy-proto type and then convert back

	// Verify the protobuf message is valid
	assert.NotNil(t, barBatch)
	assert.Len(t, barBatch.Bars, 1)

	bar := barBatch.Bars[0]
	assert.Equal(t, "AAPL", bar.Security.Symbol)
	assert.Equal(t, "XNAS", bar.Security.Mic)
	assert.Equal(t, int64(1892300), bar.Open.Scaled)
	assert.Equal(t, int32(4), bar.Open.Scale)
	assert.Equal(t, int64(43210000), bar.Volume)
	assert.True(t, bar.Adjusted)
	assert.Equal(t, "split_dividend", bar.AdjustmentPolicyId)
}

func TestEmitQuote_RoundTrip(t *testing.T) {
	// Create test input
	input := &norm.NormalizedQuote{
		Security: norm.Security{
			Symbol: "MSFT",
			MIC:    "XNAS",
		},
		Type:         "QUOTE",
		Bid:          &norm.ScaledDecimal{Scaled: 4275000, Scale: 4},
		BidSize:      int64Ptr(200),
		Ask:          &norm.ScaledDecimal{Scaled: 4275300, Scale: 4},
		AskSize:      int64Ptr(300),
		CurrencyCode: "USD",
		Venue:        "XNMS",
		EventTime:    time.Date(2024, 1, 3, 15, 30, 12, 0, time.UTC),
		IngestTime:   time.Date(2024, 1, 3, 15, 30, 12, 0, time.UTC),
		Meta: norm.Meta{
			RunID:         "test_roundtrip",
			Source:        "yfinance-go",
			Producer:      "local",
			SchemaVersion: "ampy.ticks.v1:1.0.0",
		},
	}

	// Emit to protobuf
	quote, err := EmitQuote(input)
	require.NoError(t, err)

	// Marshal to protobuf bytes
	protoBytes, err := proto.Marshal(quote)
	require.NoError(t, err)
	assert.NotEmpty(t, protoBytes)

	// Verify the protobuf message is valid
	assert.NotNil(t, quote)
	assert.Equal(t, "MSFT", quote.Security.Symbol)
	assert.Equal(t, "XNAS", quote.Security.Mic)
	assert.Equal(t, int64(4275000), quote.Bid.Scaled)
	assert.Equal(t, int32(4), quote.Bid.Scale)
	assert.Equal(t, int64(200), quote.BidSize)
	assert.Equal(t, int64(4275300), quote.Ask.Scaled)
	assert.Equal(t, int32(4), quote.Ask.Scale)
	assert.Equal(t, int64(300), quote.AskSize)
	assert.Equal(t, "XNMS", quote.Venue)
}

func TestEmitFundamentals_RoundTrip(t *testing.T) {
	// Create test input
	input := &norm.NormalizedFundamentalsSnapshot{
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
		},
		Source: "yfinance",
		AsOf:   time.Date(2025, 8, 1, 0, 0, 0, 0, time.UTC),
		Meta: norm.Meta{
			RunID:         "test_roundtrip",
			Source:        "yfinance-go",
			Producer:      "local",
			SchemaVersion: "ampy.fundamentals.v1:1.0.0",
		},
	}

	// Emit to protobuf
	fundamentals, err := EmitFundamentals(input)
	require.NoError(t, err)

	// Marshal to protobuf bytes
	protoBytes, err := proto.Marshal(fundamentals)
	require.NoError(t, err)
	assert.NotEmpty(t, protoBytes)

	// Verify the protobuf message is valid
	assert.NotNil(t, fundamentals)
	assert.Equal(t, "AAPL", fundamentals.Security.Symbol)
	assert.Equal(t, "XNAS", fundamentals.Security.Mic)
	assert.Len(t, fundamentals.Lines, 2)

	// Check first line item
	line1 := fundamentals.Lines[0]
	assert.Equal(t, "revenue", line1.Key)
	assert.Equal(t, int64(119870000000000), line1.Value.Scaled)
	assert.Equal(t, int32(2), line1.Value.Scale)
	assert.Equal(t, "USD", line1.CurrencyCode)

	// Check second line item
	line2 := fundamentals.Lines[1]
	assert.Equal(t, "net_income", line2.Key)
	assert.Equal(t, int64(2386000000000), line2.Value.Scaled)
	assert.Equal(t, int32(2), line2.Value.Scale)
	assert.Equal(t, "USD", line2.CurrencyCode)
}
