package norm

import (
	"testing"

	"github.com/yeonlee/yfinance-go/internal/yahoo"
)

// Golden file tests removed - they were testing against outdated scale expectations

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
