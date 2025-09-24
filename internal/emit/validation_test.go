package emit

import (
	"testing"
	"time"

	"github.com/AmpyFin/yfinance-go/internal/norm"
	"github.com/stretchr/testify/assert"
)

func TestValidateSecurity(t *testing.T) {
	tests := []struct {
		name    string
		security norm.Security
		wantErr bool
	}{
		{
			name:    "valid security",
			security: norm.Security{Symbol: "AAPL", MIC: "XNAS"},
			wantErr: false,
		},
		{
			name:    "empty symbol",
			security: norm.Security{Symbol: "", MIC: "XNAS"},
			wantErr: true,
		},
		{
			name:    "invalid MIC length",
			security: norm.Security{Symbol: "AAPL", MIC: "XN"},
			wantErr: true,
		},
		{
			name:    "invalid MIC format",
			security: norm.Security{Symbol: "AAPL", MIC: "xnas"},
			wantErr: true,
		},
		{
			name:    "valid without MIC",
			security: norm.Security{Symbol: "AAPL"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSecurity(tt.security)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateTimeWindow(t *testing.T) {
	start := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
	validEnd := start.Add(24 * time.Hour)
	validEvent := validEnd

	tests := []struct {
		name    string
		start   time.Time
		end     time.Time
		event   time.Time
		wantErr bool
	}{
		{
			name:    "valid daily window",
			start:   start,
			end:     validEnd,
			event:   validEvent,
			wantErr: false,
		},
		{
			name:    "invalid end time",
			start:   start,
			end:     start.Add(25 * time.Hour),
			event:   validEvent,
			wantErr: true,
		},
		{
			name:    "invalid event time",
			start:   start,
			end:     validEnd,
			event:   start.Add(12 * time.Hour),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTimeWindow(tt.start, tt.end, tt.event)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateDecimal(t *testing.T) {
	tests := []struct {
		name    string
		decimal norm.ScaledDecimal
		wantErr bool
	}{
		{
			name:    "valid decimal",
			decimal: norm.ScaledDecimal{Scaled: 12345, Scale: 2},
			wantErr: false,
		},
		{
			name:    "negative scale",
			decimal: norm.ScaledDecimal{Scaled: 12345, Scale: -1},
			wantErr: true,
		},
		{
			name:    "scale too large",
			decimal: norm.ScaledDecimal{Scaled: 12345, Scale: 10},
			wantErr: true,
		},
		{
			name:    "zero scale",
			decimal: norm.ScaledDecimal{Scaled: 12345, Scale: 0},
			wantErr: false,
		},
		{
			name:    "max scale",
			decimal: norm.ScaledDecimal{Scaled: 12345, Scale: 9},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDecimal(tt.decimal)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateCurrency(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		wantErr bool
	}{
		{
			name:    "valid USD",
			code:    "USD",
			wantErr: false,
		},
		{
			name:    "valid EUR",
			code:    "EUR",
			wantErr: false,
		},
		{
			name:    "empty currency",
			code:    "",
			wantErr: true,
		},
		{
			name:    "wrong length",
			code:    "US",
			wantErr: true,
		},
		{
			name:    "lowercase",
			code:    "usd",
			wantErr: true,
		},
		{
			name:    "unknown currency (pass-through)",
			code:    "XYZ",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCurrency(tt.code)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateAdjustments(t *testing.T) {
	tests := []struct {
		name     string
		adjusted bool
		policy   string
		wantErr  bool
	}{
		{
			name:     "valid raw",
			adjusted: false,
			policy:   "raw",
			wantErr:  false,
		},
		{
			name:     "valid split_only",
			adjusted: true,
			policy:   "split_only",
			wantErr:  false,
		},
		{
			name:     "valid split_dividend",
			adjusted: true,
			policy:   "split_dividend",
			wantErr:  false,
		},
		{
			name:     "invalid policy",
			adjusted: true,
			policy:   "invalid",
			wantErr:  true,
		},
		{
			name:     "inconsistent adjusted=true with raw",
			adjusted: true,
			policy:   "raw",
			wantErr:  true,
		},
		{
			name:     "inconsistent adjusted=false with split_only",
			adjusted: false,
			policy:   "split_only",
			wantErr:  true,
		},
		{
			name:     "inconsistent adjusted=false with split_dividend",
			adjusted: false,
			policy:   "split_dividend",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAdjustments(tt.adjusted, tt.policy)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateFundamentals(t *testing.T) {
	validLine := norm.NormalizedFundamentalsLine{
		Key:          "revenue",
		Value:        norm.ScaledDecimal{Scaled: 1000000, Scale: 2},
		CurrencyCode: "USD",
		PeriodStart:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:    time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
	}

	tests := []struct {
		name    string
		lines   []norm.NormalizedFundamentalsLine
		wantErr bool
	}{
		{
			name:    "valid fundamentals",
			lines:   []norm.NormalizedFundamentalsLine{validLine},
			wantErr: false,
		},
		{
			name: "invalid key",
			lines: []norm.NormalizedFundamentalsLine{
				{
					Key:          "invalid_key",
					Value:        norm.ScaledDecimal{Scaled: 1000000, Scale: 2},
					CurrencyCode: "USD",
					PeriodStart:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					PeriodEnd:    time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
				},
			},
			wantErr: true,
		},
		{
			name: "custom key (valid)",
			lines: []norm.NormalizedFundamentalsLine{
				{
					Key:          "custom_metric",
					Value:        norm.ScaledDecimal{Scaled: 1000000, Scale: 2},
					CurrencyCode: "USD",
					PeriodStart:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					PeriodEnd:    time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
				},
			},
			wantErr: false,
		},
		{
			name: "invalid period",
			lines: []norm.NormalizedFundamentalsLine{
				{
					Key:          "revenue",
					Value:        norm.ScaledDecimal{Scaled: 1000000, Scale: 2},
					CurrencyCode: "USD",
					PeriodStart:  time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
					PeriodEnd:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFundamentals(tt.lines)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
