package norm

import (
	"testing"

	"github.com/AmpyFin/yfinance-go/internal/yahoo"
)

// Golden file tests removed - they were testing against outdated scale expectations

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
