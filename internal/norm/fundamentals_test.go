package norm

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/yeonlee/yfinance-go/internal/yahoo"
)

func TestNormalizeFundamentalsGolden(t *testing.T) {
	// Read source data
	sourceData, err := os.ReadFile(filepath.Join("../../testdata/source/yahoo/fundamentals/AAPL_quarterly_sample.json"))
	if err != nil {
		t.Fatalf("Failed to read source file: %v", err)
	}

	// Decode Yahoo response
	yahooResp, err := yahoo.DecodeFundamentalsResponse(sourceData)
	if err != nil {
		t.Fatalf("Failed to decode Yahoo response: %v", err)
	}

	// Extract fundamentals
	fundamentals, err := yahooResp.GetFundamentals()
	if err != nil {
		t.Fatalf("Failed to get fundamentals: %v", err)
	}

	// Normalize fundamentals
	normalized, err := NormalizeFundamentals(fundamentals, "AAPL", "golden_fund_v1")
	if err != nil {
		t.Fatalf("Failed to normalize fundamentals: %v", err)
	}

	// Read golden data
	goldenData, err := os.ReadFile(filepath.Join("../../testdata/golden/ampy/fundamentals/AAPL_quarterly_snapshot.json"))
	if err != nil {
		t.Fatalf("Failed to read golden file: %v", err)
	}

	// Parse golden data
	var golden NormalizedFundamentalsSnapshot
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

func TestNormalizeFundamentalsValidation(t *testing.T) {
	tests := []struct {
		name          string
		fundamentals  *yahoo.Fundamentals
		symbol        string
		runID         string
		wantErr       bool
	}{
		{
			name: "valid fundamentals",
			fundamentals: &yahoo.Fundamentals{
				IncomeStatements: []yahoo.IncomeStatement{
					{
						EndDate: yahoo.DateValue{
							Raw: 1719705600, // 2024-06-29
							Fmt: "2024-06-29",
						},
						TotalRevenue: &yahoo.Value{
							Raw: func() *int64 { v := int64(1198700000000); return &v }(),
							Fmt: func() *string { v := "1.2T"; return &v }(),
						},
						NetIncome: &yahoo.Value{
							Raw: func() *int64 { v := int64(23860000000); return &v }(),
							Fmt: func() *string { v := "23.86B"; return &v }(),
						},
						EPS: &yahoo.Value{
							Raw: func() *int64 { v := int64(1525); return &v }(),
							Fmt: func() *string { v := "1.53"; return &v }(),
						},
					},
				},
			},
			symbol:  "AAPL",
			runID:   "test_run",
			wantErr: false,
		},
		{
			name:          "nil fundamentals",
			fundamentals:  nil,
			symbol:        "AAPL",
			runID:         "test_run",
			wantErr:       true,
		},
		{
			name: "empty fundamentals",
			fundamentals: &yahoo.Fundamentals{
				IncomeStatements: []yahoo.IncomeStatement{},
			},
			symbol:  "AAPL",
			runID:   "test_run",
			wantErr: true,
		},
		{
			name: "missing symbol",
			fundamentals: &yahoo.Fundamentals{
				IncomeStatements: []yahoo.IncomeStatement{
					{
						EndDate: yahoo.DateValue{
							Raw: 1719705600,
							Fmt: "2024-06-29",
						},
						TotalRevenue: &yahoo.Value{
							Raw: func() *int64 { v := int64(1198700000000); return &v }(),
						},
					},
				},
			},
			symbol:  "",
			runID:   "test_run",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NormalizeFundamentals(tt.fundamentals, tt.symbol, tt.runID)
			if (err != nil) != tt.wantErr {
				t.Errorf("NormalizeFundamentals() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
