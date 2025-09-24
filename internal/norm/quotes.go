package norm

import (
	"fmt"
	"time"

	"github.com/yeonlee/yfinance-go/internal/yahoo"
)

// NormalizeQuote converts a Yahoo Finance quote to a normalized quote
func NormalizeQuote(quote yahoo.Quote, runID string) (*NormalizedQuote, error) {
	// Validate required fields
	if quote.Symbol == "" {
		return nil, fmt.Errorf("missing symbol")
	}
	if quote.Currency == "" {
		return nil, fmt.Errorf("missing currency")
	}
	
	// Create security
	security := CreateSecurity(quote.Symbol, quote.Exchange, quote.FullExchangeName)
	if err := ValidateSecurity(security); err != nil {
		return nil, fmt.Errorf("invalid security: %w", err)
	}
	
	// Get currency scale
	scale := GetScaleForCurrency(quote.Currency)
	
	// Convert event time - use current time for real-time data
	eventTime := time.Now().UTC()
	
	// Convert bid/ask prices if present
	var bid, ask *ScaledDecimal
	
	if quote.Bid != nil {
		bidScaled, err := ToScaledDecimal(*quote.Bid, scale)
		if err != nil {
			return nil, fmt.Errorf("invalid bid price: %w", err)
		}
		bid = &bidScaled
	}
	
	if quote.Ask != nil {
		askScaled, err := ToScaledDecimal(*quote.Ask, scale)
		if err != nil {
			return nil, fmt.Errorf("invalid ask price: %w", err)
		}
		ask = &askScaled
	}
	
	// Convert regular market data if present
	var regularMarketPrice, regularMarketHigh, regularMarketLow *ScaledDecimal
	
	if quote.RegularMarketPrice != nil {
		priceScaled, err := ToScaledDecimal(*quote.RegularMarketPrice, scale)
		if err != nil {
			return nil, fmt.Errorf("invalid regular market price: %w", err)
		}
		regularMarketPrice = &priceScaled
	}
	
	if quote.RegularMarketDayHigh != nil {
		highScaled, err := ToScaledDecimal(*quote.RegularMarketDayHigh, scale)
		if err != nil {
			return nil, fmt.Errorf("invalid regular market high: %w", err)
		}
		regularMarketHigh = &highScaled
	}
	
	if quote.RegularMarketDayLow != nil {
		lowScaled, err := ToScaledDecimal(*quote.RegularMarketDayLow, scale)
		if err != nil {
			return nil, fmt.Errorf("invalid regular market low: %w", err)
		}
		regularMarketLow = &lowScaled
	}
	
	// Determine venue - use exchange MIC mapping
	venue := ""
	if quote.Exchange != "" {
		venue = InferMIC(quote.Exchange, "")
		if venue == "" {
			venue = quote.Exchange
		}
	}
	
	// Create metadata
	meta := Meta{
		RunID:         runID,
		Source:        "yfinance-go",
		Producer:      "local",
		SchemaVersion: "ampy.ticks.v1:1.0.0",
	}
	
	return &NormalizedQuote{
		Security:            security,
		Type:                "QUOTE",
		Bid:                 bid,
		BidSize:             quote.BidSize,
		Ask:                 ask,
		AskSize:             quote.AskSize,
		RegularMarketPrice:  regularMarketPrice,
		RegularMarketHigh:   regularMarketHigh,
		RegularMarketLow:    regularMarketLow,
		RegularMarketVolume: quote.RegularMarketVolume,
		Venue:               venue,
		CurrencyCode:        quote.Currency,
		EventTime:           eventTime,
		IngestTime:          eventTime,
		Meta:                meta,
	}, nil
}
