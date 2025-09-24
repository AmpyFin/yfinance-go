package crosslang

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/AmpyFin/yfinance-go/internal/emit"
	"github.com/AmpyFin/yfinance-go/internal/norm"
	"google.golang.org/protobuf/proto"
)

func TestCrossLanguageRoundTripBars(t *testing.T) {
	// Test Go → Python round-trip for bars
	testCases := []struct {
		name     string
		currency string
		scale    int
	}{
		{
			name:     "USD adjusted bars",
			currency: "USD",
			scale:    2,
		},
		{
			name:     "EUR adjusted bars", 
			currency: "EUR",
			scale:    2,
		},
		{
			name:     "JPY adjusted bars",
			currency: "JPY", 
			scale:    2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create test bar batch with scale 2
			barBatch := createTestBarBatch(tc.currency, tc.scale)
			
			// Emit to protobuf
			protobufData, err := emit.EmitBarBatch(barBatch)
			require.NoError(t, err)

			// Marshal to bytes
			protobufBytes, err := proto.Marshal(protobufData)
			require.NoError(t, err)

			// Write protobuf to file for Python to read
			outputDir := "_rt"
			_ = os.MkdirAll(outputDir, 0755)
			
			pbFile := filepath.Join(outputDir, "bars_"+tc.currency+".pb")
			err = os.WriteFile(pbFile, protobufBytes, 0644)
			require.NoError(t, err)

			// Create metadata file for Python
			metadata := createBarMetadata(barBatch)
			metadataFile := filepath.Join(outputDir, "bars_"+tc.currency+"_metadata.json")
			metadataBytes, err := json.Marshal(metadata)
			require.NoError(t, err)
			err = os.WriteFile(metadataFile, metadataBytes, 0644)
			require.NoError(t, err)

			// Run Python test
			runPythonRoundTripTest(t, "bars", tc.currency)
		})
	}
}

func TestCrossLanguageRoundTripQuotes(t *testing.T) {
	// Create test quote
	quote := createTestQuote()
	
	// Emit to protobuf
	protobufData, err := emit.EmitQuote(quote)
	require.NoError(t, err)

	// Marshal to bytes
	protobufBytes, err := proto.Marshal(protobufData)
	require.NoError(t, err)

	// Write protobuf to file
	outputDir := "_rt"
	_ = os.MkdirAll(outputDir, 0755)
	
	pbFile := filepath.Join(outputDir, "quote.pb")
	err = os.WriteFile(pbFile, protobufBytes, 0644)
	require.NoError(t, err)

	// Create metadata file
	metadata := createQuoteMetadata(quote)
	metadataFile := filepath.Join(outputDir, "quote_metadata.json")
	metadataBytes, err := json.Marshal(metadata)
	require.NoError(t, err)
	err = os.WriteFile(metadataFile, metadataBytes, 0644)
	require.NoError(t, err)

	// Run Python test
	runPythonRoundTripTest(t, "quote", "USD")
}

func TestCrossLanguageRoundTripFundamentals(t *testing.T) {
	// Create test fundamentals
	fundamentals := createTestFundamentals()
	
	// Emit to protobuf
	protobufData, err := emit.EmitFundamentals(fundamentals)
	require.NoError(t, err)

	// Marshal to bytes
	protobufBytes, err := proto.Marshal(protobufData)
	require.NoError(t, err)

	// Write protobuf to file
	outputDir := "_rt"
	_ = os.MkdirAll(outputDir, 0755)
	
	pbFile := filepath.Join(outputDir, "fundamentals.pb")
	err = os.WriteFile(pbFile, protobufBytes, 0644)
	require.NoError(t, err)

	// Create metadata file
	metadata := createFundamentalsMetadata(fundamentals)
	metadataFile := filepath.Join(outputDir, "fundamentals_metadata.json")
	metadataBytes, err := json.Marshal(metadata)
	require.NoError(t, err)
	err = os.WriteFile(metadataFile, metadataBytes, 0644)
	require.NoError(t, err)

	// Run Python test
	runPythonRoundTripTest(t, "fundamentals", "USD")
}

func TestCrossLanguageNumericPrecision(t *testing.T) {
	// Test edge cases for decimal precision
	testCases := []struct {
		name  string
		price float64
		scale int
	}{
		{
			name:  "high precision decimal",
			price: 123.456789,
			scale: 2,
		},
		{
			name:  "edge case - exactly 0.5",
			price: 123.50,
			scale: 2,
		},
		{
			name:  "edge case - very small value",
			price: 0.01,
			scale: 2,
		},
		{
			name:  "edge case - large value",
			price: 999999.99,
			scale: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create bar with specific price
			barBatch := createTestBarBatchWithPrice(tc.price, tc.scale)
			
			// Emit to protobuf
			protobufData, err := emit.EmitBarBatch(barBatch)
			require.NoError(t, err)

			// Marshal to bytes
			protobufBytes, err := proto.Marshal(protobufData)
			require.NoError(t, err)

			// Write protobuf to file
			outputDir := "_rt"
			_ = os.MkdirAll(outputDir, 0755)
			
			pbFile := filepath.Join(outputDir, "precision_"+tc.name+".pb")
			err = os.WriteFile(pbFile, protobufBytes, 0644)
			require.NoError(t, err)

			// Create metadata with expected values
			metadata := map[string]interface{}{
				"expected_price": tc.price,
				"expected_scale": tc.scale,
				"test_case":      tc.name,
			}
			metadataFile := filepath.Join(outputDir, "precision_"+tc.name+"_metadata.json")
			metadataBytes, err := json.Marshal(metadata)
			require.NoError(t, err)
			err = os.WriteFile(metadataFile, metadataBytes, 0644)
			require.NoError(t, err)

			// Run Python test
			runPythonRoundTripTest(t, "precision", tc.name)
		})
	}
}

// Helper functions

func createTestBarBatch(currency string, scale int) *norm.NormalizedBarBatch {
	now := time.Now()
	return &norm.NormalizedBarBatch{
		Security: norm.Security{
			Symbol: "AAPL",
			MIC:    "XNAS",
		},
		Bars: []norm.NormalizedBar{
			{
				Open:   norm.ScaledDecimal{Scaled: 15000, Scale: scale},
				High:   norm.ScaledDecimal{Scaled: 15100, Scale: scale},
				Low:    norm.ScaledDecimal{Scaled: 14900, Scale: scale},
				Close:  norm.ScaledDecimal{Scaled: 15050, Scale: scale},
				Volume: 1000000,
				Start:  now.Add(-24 * time.Hour),
				End:    now,
				EventTime: now,
				IngestTime: now,
				AsOf: now,
				CurrencyCode: currency,
				Adjusted: true,
				AdjustmentPolicyID: "split_dividend",
			},
		},
		Meta: norm.Meta{
			RunID: "crosslang_test",
			Source: "yfinance-go",
			Producer: "test",
			SchemaVersion: "ampy.bars.v1:1.0.0",
		},
	}
}

func createTestBarBatchWithPrice(price float64, scale int) *norm.NormalizedBarBatch {
	now := time.Now()
	multiplier := int64(1)
	for i := 0; i < scale; i++ {
		multiplier *= 10
	}
	scaled := int64(price * float64(multiplier))
	return &norm.NormalizedBarBatch{
		Security: norm.Security{
			Symbol: "TEST",
			MIC:    "XNAS",
		},
		Bars: []norm.NormalizedBar{
			{
				Open:   norm.ScaledDecimal{Scaled: scaled, Scale: scale},
				High:   norm.ScaledDecimal{Scaled: scaled + 100, Scale: scale},
				Low:    norm.ScaledDecimal{Scaled: scaled - 100, Scale: scale},
				Close:  norm.ScaledDecimal{Scaled: scaled, Scale: scale},
				Volume: 1000,
				Start:  now.Add(-24 * time.Hour),
				End:    now,
				EventTime: now,
				IngestTime: now,
				AsOf: now,
				CurrencyCode: "USD",
				Adjusted: false,
				AdjustmentPolicyID: "raw",
			},
		},
		Meta: norm.Meta{
			RunID: "precision_test",
			Source: "yfinance-go",
			Producer: "test",
			SchemaVersion: "ampy.bars.v1:1.0.0",
		},
	}
}

func createTestQuote() *norm.NormalizedQuote {
	now := time.Now()
	bidSize := int64(200)
	askSize := int64(300)
	return &norm.NormalizedQuote{
		Security: norm.Security{
			Symbol: "MSFT",
			MIC:    "XNAS",
		},
		Bid:     &norm.ScaledDecimal{Scaled: 42750, Scale: 2},
		Ask:     &norm.ScaledDecimal{Scaled: 42753, Scale: 2},
		BidSize: &bidSize,
		AskSize: &askSize,
		EventTime: now,
		IngestTime: now,
		Venue: "XNMS",
		CurrencyCode: "USD",
		Meta: norm.Meta{
			RunID: "crosslang_test",
			Source: "yfinance-go",
			Producer: "test",
			SchemaVersion: "ampy.ticks.v1:1.0.0",
		},
	}
}

func createTestFundamentals() *norm.NormalizedFundamentalsSnapshot {
	now := time.Now()
	return &norm.NormalizedFundamentalsSnapshot{
		Security: norm.Security{
			Symbol: "AAPL",
			MIC:    "XNAS",
		},
		Lines: []norm.NormalizedFundamentalsLine{
			{
				Key:          "revenue",
				Value:        norm.ScaledDecimal{Scaled: 119870000000000, Scale: 2},
				CurrencyCode: "USD",
				PeriodStart:  now.Add(-90 * 24 * time.Hour),
				PeriodEnd:    now,
			},
			{
				Key:          "net_income",
				Value:        norm.ScaledDecimal{Scaled: 2386000000000, Scale: 2},
				CurrencyCode: "USD",
				PeriodStart:  now.Add(-90 * 24 * time.Hour),
				PeriodEnd:    now,
			},
		},
		Source: "yfinance",
		AsOf:   now,
		Meta: norm.Meta{
			RunID: "crosslang_test",
			Source: "yfinance-go",
			Producer: "test",
			SchemaVersion: "ampy.fundamentals.v1:1.0.0",
		},
	}
}

func createBarMetadata(barBatch *norm.NormalizedBarBatch) map[string]interface{} {
	bar := barBatch.Bars[0]
	return map[string]interface{}{
		"security": map[string]interface{}{
			"symbol": barBatch.Security.Symbol,
			"mic":    barBatch.Security.MIC,
		},
		"bar": map[string]interface{}{
			"open": map[string]interface{}{
				"scaled": bar.Open.Scaled,
				"scale":  bar.Open.Scale,
			},
			"high": map[string]interface{}{
				"scaled": bar.High.Scaled,
				"scale":  bar.High.Scale,
			},
			"low": map[string]interface{}{
				"scaled": bar.Low.Scaled,
				"scale":  bar.Low.Scale,
			},
			"close": map[string]interface{}{
				"scaled": bar.Close.Scaled,
				"scale":  bar.Close.Scale,
			},
			"volume": bar.Volume,
			"currency": bar.CurrencyCode,
		},
		"run_id": barBatch.Meta.RunID,
	}
}

func createQuoteMetadata(quote *norm.NormalizedQuote) map[string]interface{} {
	return map[string]interface{}{
		"security": map[string]interface{}{
			"symbol": quote.Security.Symbol,
			"mic":    quote.Security.MIC,
		},
		"quote": map[string]interface{}{
			"bid": map[string]interface{}{
				"scaled": quote.Bid.Scaled,
				"scale":  quote.Bid.Scale,
			},
			"ask": map[string]interface{}{
				"scaled": quote.Ask.Scaled,
				"scale":  quote.Ask.Scale,
			},
			"bid_size": quote.BidSize,
			"ask_size": quote.AskSize,
			"venue": quote.Venue,
		},
		"run_id": quote.Meta.RunID,
	}
}

func createFundamentalsMetadata(fundamentals *norm.NormalizedFundamentalsSnapshot) map[string]interface{} {
	lines := make([]map[string]interface{}, len(fundamentals.Lines))
	for i, line := range fundamentals.Lines {
		lines[i] = map[string]interface{}{
			"key": line.Key,
			"value": map[string]interface{}{
				"scaled": line.Value.Scaled,
				"scale":  line.Value.Scale,
			},
			"currency_code": line.CurrencyCode,
		}
	}

	return map[string]interface{}{
		"security": map[string]interface{}{
			"symbol": fundamentals.Security.Symbol,
			"mic":    fundamentals.Security.MIC,
		},
		"lines": lines,
		"source": fundamentals.Source,
		"as_of": fundamentals.AsOf.Unix(),
		"meta": map[string]interface{}{
			"run_id": fundamentals.Meta.RunID,
		},
	}
}

func runPythonRoundTripTest(t *testing.T, testType, identifier string) {
	// Check if Python test script exists
	pythonScript := filepath.Join("tests", "crosslang", "python", "roundtrip_test.py")
	if _, err := os.Stat(pythonScript); os.IsNotExist(err) {
		t.Skipf("Python test script not found: %s", pythonScript)
		return
	}

	// Run Python test
	cmd := exec.Command("python3", pythonScript, testType, identifier)
	cmd.Dir = "."
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Python round-trip test failed: %v\nOutput: %s", err, string(output))
	}
	
	t.Logf("Python round-trip test passed for %s/%s", testType, identifier)
}
