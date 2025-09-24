package norm

import (
	"fmt"
	"time"

	"github.com/AmpyFin/yfinance-go/internal/yahoo"
)

// NormalizeMarketData normalizes comprehensive market data from chart metadata
func NormalizeMarketData(meta *yahoo.ChartMeta, runID string) (*NormalizedMarketData, error) {
	if meta == nil {
		return nil, fmt.Errorf("metadata is nil")
	}

	// Create security
	security := Security{
		Symbol: meta.Symbol,
		MIC:    InferMIC(meta.ExchangeName, ""),
	}

	// Convert regular market time if available
	var regularMarketTime *time.Time
	if meta.RegularMarketTime != 0 {
		rmt := time.Unix(meta.RegularMarketTime, 0).UTC()
		regularMarketTime = &rmt
	}

	// Create normalized market data
	marketData := &NormalizedMarketData{
		Security:               security,
		RegularMarketPrice:     ToScaledDecimalPtr(meta.RegularMarketPrice, meta.Currency),
		RegularMarketHigh:      ToScaledDecimalPtr(meta.RegularMarketDayHigh, meta.Currency),
		RegularMarketLow:       ToScaledDecimalPtr(meta.RegularMarketDayLow, meta.Currency),
		RegularMarketVolume:    meta.RegularMarketVolume,
		FiftyTwoWeekHigh:       ToScaledDecimalPtr(meta.FiftyTwoWeekHigh, meta.Currency),
		FiftyTwoWeekLow:        ToScaledDecimalPtr(meta.FiftyTwoWeekLow, meta.Currency),
		PreviousClose:          ToScaledDecimalPtr(meta.PreviousClose, meta.Currency),
		ChartPreviousClose:     ToScaledDecimalPtr(meta.ChartPreviousClose, meta.Currency),
		RegularMarketTime:      regularMarketTime,
		HasPrePostMarketData:   meta.HasPrePostMarketData,
		EventTime:              time.Now().UTC(),
		IngestTime:             time.Now().UTC(),
		Meta: Meta{
			RunID:         runID,
			Source:        "yahoo",
			Producer:      "yfinance-go",
			SchemaVersion: "1.0",
		},
	}

	return marketData, nil
}

// ToScaledDecimalPtr converts a float64 pointer to a ScaledDecimal pointer
func ToScaledDecimalPtr(value *float64, currency string) *ScaledDecimal {
	if value == nil {
		return nil
	}
	
	scaled, _ := ToScaledDecimalWithCurrency(*value, currency)
	return &scaled
}
