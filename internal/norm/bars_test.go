package norm

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yeonlee/yfinance-go/internal/yahoo"
)

func TestNormalizeBarsGolden(t *testing.T) {
	tests := []struct {
		name           string
		sourceFile     string
		goldenFile     string
		runID          string
	}{
		{
			name:       "AAPL adjusted bars",
			sourceFile: "AAPL_1d_sample.json",
			goldenFile: "AAPL_1d_adjusted.json",
			runID:      "golden_bars_v1",
		},
		{
			name:       "SAP EUR bars",
			sourceFile: "SAP_XETR_1d_eur.json",
			goldenFile: "SAP_XETR_1d_adjusted_eur.json",
			runID:      "golden_bars_v1",
		},
		{
			name:       "TM JPY bars",
			sourceFile: "TM_XTKS_1d_jpy.json",
			goldenFile: "TM_XTKS_1d_adjusted_jpy.json",
			runID:      "golden_bars_v1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Read source data
			sourceData, err := os.ReadFile(filepath.Join("../../testdata/source/yahoo/bars", tt.sourceFile))
			if err != nil {
				t.Fatalf("Failed to read source file: %v", err)
			}

			// Decode Yahoo response
			yahooResp, err := yahoo.DecodeBarsResponse(sourceData)
			if err != nil {
				t.Fatalf("Failed to decode Yahoo response: %v", err)
			}

			// Extract bars and metadata
			bars, err := yahooResp.GetBars()
			if err != nil {
				t.Fatalf("Failed to get bars: %v", err)
			}

			meta := yahooResp.GetMetadata()
			if meta == nil {
				t.Fatal("Missing metadata")
			}

			// Normalize bars
			normalized, err := NormalizeBars(bars, meta, tt.runID)
			if err != nil {
				t.Fatalf("Failed to normalize bars: %v", err)
			}

			// Read golden data
			goldenData, err := os.ReadFile(filepath.Join("../../testdata/golden/ampy/bars", tt.goldenFile))
			if err != nil {
				t.Fatalf("Failed to read golden file: %v", err)
			}

			// Parse golden data
			var golden NormalizedBarBatch
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
		})
	}
}

func TestNormalizeBarsValidation(t *testing.T) {
	tests := []struct {
		name    string
		bars    []yahoo.Bar
		meta    *yahoo.ChartMeta
		runID   string
		wantErr bool
	}{
		{
			name: "valid bars",
			bars: []yahoo.Bar{
				{
					Timestamp: 1704326400,
					Open:      189.23,
					High:      191.0,
					Low:       188.9,
					Close:     190.45,
					Volume:    43210000,
				},
			},
			meta: &yahoo.ChartMeta{
				Symbol:   "AAPL",
				Currency: "USD",
			},
			runID:   "test_run",
			wantErr: false,
		},
		{
			name: "empty bars",
			bars: []yahoo.Bar{},
			meta: &yahoo.ChartMeta{
				Symbol:   "AAPL",
				Currency: "USD",
			},
			runID:   "test_run",
			wantErr: true,
		},
		{
			name: "nil metadata",
			bars: []yahoo.Bar{
				{
					Timestamp: 1704326400,
					Open:      189.23,
					High:      191.0,
					Low:       188.9,
					Close:     190.45,
					Volume:    43210000,
				},
			},
			meta:    nil,
			runID:   "test_run",
			wantErr: true,
		},
		{
			name: "missing symbol",
			bars: []yahoo.Bar{
				{
					Timestamp: 1704326400,
					Open:      189.23,
					High:      191.0,
					Low:       188.9,
					Close:     190.45,
					Volume:    43210000,
				},
			},
			meta: &yahoo.ChartMeta{
				Symbol:   "",
				Currency: "USD",
			},
			runID:   "test_run",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NormalizeBars(tt.bars, tt.meta, tt.runID)
			if (err != nil) != tt.wantErr {
				t.Errorf("NormalizeBars() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestToUTCDayBoundaries(t *testing.T) {
	tests := []struct {
		name      string
		timestamp int64
		wantStart time.Time
		wantEnd   time.Time
		wantEvent time.Time
	}{
		{
			name:      "AAPL timestamp",
			timestamp: 1704326400, // 2024-01-04 00:00:00 UTC (end of Jan 3 EST trading day)
			wantStart: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			wantEnd:   time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC),
			wantEvent: time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end, eventTime := ToUTCDayBoundaries(tt.timestamp)
			
			if !start.Equal(tt.wantStart) {
				t.Errorf("ToUTCDayBoundaries() start = %v, want %v", start, tt.wantStart)
			}
			if !end.Equal(tt.wantEnd) {
				t.Errorf("ToUTCDayBoundaries() end = %v, want %v", end, tt.wantEnd)
			}
			if !eventTime.Equal(tt.wantEvent) {
				t.Errorf("ToUTCDayBoundaries() eventTime = %v, want %v", eventTime, tt.wantEvent)
			}
		})
	}
}
