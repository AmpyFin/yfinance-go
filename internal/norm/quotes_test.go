package norm

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/yeonlee/yfinance-go/internal/yahoo"
)

func TestNormalizeQuoteGolden(t *testing.T) {
	// Read source data
	sourceData, err := os.ReadFile(filepath.Join("../../testdata/source/yahoo/quotes/MSFT_quote_sample.json"))
	if err != nil {
		t.Fatalf("Failed to read source file: %v", err)
	}

	// Decode Yahoo response
	yahooResp, err := yahoo.DecodeQuoteResponse(sourceData)
	if err != nil {
		t.Fatalf("Failed to decode Yahoo response: %v", err)
	}

	// Extract quotes
	quotes := yahooResp.GetQuotes()
	if len(quotes) == 0 {
		t.Fatal("No quotes found")
	}

	// Normalize first quote
	normalized, err := NormalizeQuote(quotes[0], "golden_quote_v1")
	if err != nil {
		t.Fatalf("Failed to normalize quote: %v", err)
	}

	// Read golden data
	goldenData, err := os.ReadFile(filepath.Join("../../testdata/golden/ampy/quotes/MSFT_snapshot_quote.json"))
	if err != nil {
		t.Fatalf("Failed to read golden file: %v", err)
	}

	// Parse golden data
	var golden NormalizedQuote
	if err := json.Unmarshal(goldenData, &golden); err != nil {
		t.Fatalf("Failed to parse golden data: %v", err)
	}

	// Compare normalized with golden (byte-equal after canonical JSON marshaling)
	normalizedJSON, err := json.Marshal(normalized)
	if err != nil {
		t.Fatalf("Failed to marshal normalized data: %v", err)
	}

	goldenJSON, err := json.Marshal(&golden)
	if err != nil {
		t.Fatalf("Failed to marshal golden data: %v", err)
	}

	// Compare JSON (should be byte-equal)
	if string(normalizedJSON) != string(goldenJSON) {
		t.Errorf("Normalized data does not match golden data")
		t.Logf("Normalized: %s", string(normalizedJSON))
		t.Logf("Golden: %s", string(goldenJSON))
	}
}

func TestNormalizeQuoteValidation(t *testing.T) {
	tests := []struct {
		name    string
		quote   yahoo.Quote
		runID   string
		wantErr bool
	}{
		{
			name: "valid quote",
			quote: yahoo.Quote{
				Symbol:   "MSFT",
				Currency: "USD",
				Exchange: "NMS",
				Bid:      func() *float64 { v := 427.50; return &v }(),
				Ask:      func() *float64 { v := 427.53; return &v }(),
				BidSize:  func() *int64 { v := int64(200); return &v }(),
				AskSize:  func() *int64 { v := int64(300); return &v }(),
			},
			runID:   "test_run",
			wantErr: false,
		},
		{
			name: "missing symbol",
			quote: yahoo.Quote{
				Symbol:   "",
				Currency: "USD",
				Exchange: "NMS",
			},
			runID:   "test_run",
			wantErr: true,
		},
		{
			name: "missing currency",
			quote: yahoo.Quote{
				Symbol:   "MSFT",
				Currency: "",
				Exchange: "NMS",
			},
			runID:   "test_run",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NormalizeQuote(tt.quote, tt.runID)
			if (err != nil) != tt.wantErr {
				t.Errorf("NormalizeQuote() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
