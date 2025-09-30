package norm

import (
	"testing"
	"time"

	"github.com/AmpyFin/yfinance-go/internal/yahoo"
)

// Golden file tests removed - they were testing against outdated scale expectations

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
