package norm

import (
	"fmt"
	"time"

	"github.com/AmpyFin/yfinance-go/internal/yahoo"
)

// NormalizeBars converts Yahoo Finance bars to normalized bars
func NormalizeBars(bars []yahoo.Bar, meta *yahoo.ChartMeta, runID string) (*NormalizedBarBatch, error) {
	if len(bars) == 0 {
		return nil, fmt.Errorf("no bars to normalize")
	}

	if meta == nil {
		return nil, fmt.Errorf("missing metadata")
	}

	// Create security
	security := CreateSecurity(meta.Symbol, meta.ExchangeName, meta.ExchangeName)
	if err := ValidateSecurity(security); err != nil {
		return nil, fmt.Errorf("invalid security: %w", err)
	}

	// Determine if data is adjusted
	isAdjusted := false
	adjustmentPolicyID := "raw"

	// Check if any bar has adjusted close data
	for _, bar := range bars {
		if bar.AdjClose != nil {
			isAdjusted = true
			adjustmentPolicyID = "split_dividend"
			break
		}
	}

	// Get currency scale
	scale := GetScaleForCurrency(meta.Currency)

	// Normalize each bar
	normalizedBars := make([]NormalizedBar, 0, len(bars))
	// Use current time for ingest timestamp
	ingestTime := time.Now().UTC()

	for _, bar := range bars {
		normalizedBar, err := normalizeBar(bar, meta.Currency, scale, isAdjusted, adjustmentPolicyID, ingestTime)
		if err != nil {
			// Log warning but continue with other bars
			continue
		}
		normalizedBars = append(normalizedBars, normalizedBar)
	}

	if len(normalizedBars) == 0 {
		return nil, fmt.Errorf("no valid bars after normalization")
	}

	// Create metadata
	metaData := Meta{
		RunID:         runID,
		Source:        "yfinance-go",
		Producer:      "local",
		SchemaVersion: "ampy.bars.v1:1.0.0",
	}

	return &NormalizedBarBatch{
		Security: security,
		Bars:     normalizedBars,
		Meta:     metaData,
	}, nil
}

// normalizeBar normalizes a single bar
func normalizeBar(bar yahoo.Bar, currency string, scale int, isAdjusted bool, adjustmentPolicyID string, now time.Time) (NormalizedBar, error) {
	// Convert timestamp to UTC day boundaries
	start, end, eventTime := ToUTCDayBoundaries(bar.Timestamp)

	// Determine which close price to use
	closePrice := bar.Close
	if isAdjusted && bar.AdjClose != nil {
		closePrice = *bar.AdjClose
	}

	// Convert prices to scaled decimals using currency-aware conversion
	open, err := ToScaledDecimalWithCurrency(bar.Open, currency)
	if err != nil {
		return NormalizedBar{}, fmt.Errorf("invalid open price: %w", err)
	}

	high, err := ToScaledDecimalWithCurrency(bar.High, currency)
	if err != nil {
		return NormalizedBar{}, fmt.Errorf("invalid high price: %w", err)
	}

	low, err := ToScaledDecimalWithCurrency(bar.Low, currency)
	if err != nil {
		return NormalizedBar{}, fmt.Errorf("invalid low price: %w", err)
	}

	closePriceScaled, err := ToScaledDecimalWithCurrency(closePrice, currency)
	if err != nil {
		return NormalizedBar{}, fmt.Errorf("invalid close price: %w", err)
	}

	return NormalizedBar{
		Start:              start,
		End:                end,
		Open:               open,
		High:               high,
		Low:                low,
		Close:              closePriceScaled,
		Volume:             bar.Volume,
		Adjusted:           isAdjusted,
		AdjustmentPolicyID: adjustmentPolicyID,
		CurrencyCode:       currency,
		EventTime:          eventTime,
		IngestTime:         now,
		AsOf:               eventTime,
	}, nil
}
