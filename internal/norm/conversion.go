package norm

import (
	"context"
	"fmt"
	"time"
)

// ConvertedBar represents a bar with converted currency values
type ConvertedBar struct {
	Start               time.Time     `json:"start"`
	End                 time.Time     `json:"end"`
	Open                ScaledDecimal `json:"open"`
	High                ScaledDecimal `json:"high"`
	Low                 ScaledDecimal `json:"low"`
	Close               ScaledDecimal `json:"close"`
	OriginalCurrency    string        `json:"original_currency"`    // Source currency
	ConvertedCurrency   string        `json:"converted_currency"`   // Target currency
	Volume              int64         `json:"volume"`
	Adjusted            bool          `json:"adjusted"`
	AdjustmentPolicyID  string        `json:"adjustment_policy_id"`
	EventTime           time.Time     `json:"event_time"`
	IngestTime          time.Time     `json:"ingest_time"`
	AsOf                time.Time     `json:"as_of"`
}

// ConvertedBarBatch represents a batch of converted bars
type ConvertedBarBatch struct {
	Security Security         `json:"security"`
	Bars     []ConvertedBar   `json:"bars"`
	Meta     Meta            `json:"meta"`
	FXMeta   *FXMeta         `json:"fx_meta,omitempty"`
}

// ConvertedQuote represents a quote with converted currency values
type ConvertedQuote struct {
	Security           Security       `json:"security"`
	Type               string         `json:"type"`
	Bid                *ScaledDecimal `json:"bid,omitempty"`
	BidSize            *int64         `json:"bid_size,omitempty"`
	Ask                *ScaledDecimal `json:"ask,omitempty"`
	AskSize            *int64         `json:"ask_size,omitempty"`
	RegularMarketPrice *ScaledDecimal `json:"regular_market_price,omitempty"`
	RegularMarketHigh  *ScaledDecimal `json:"regular_market_high,omitempty"`
	RegularMarketLow   *ScaledDecimal `json:"regular_market_low,omitempty"`
	RegularMarketVolume *int64        `json:"regular_market_volume,omitempty"`
	OriginalCurrency   string         `json:"original_currency"`    // Source currency
	ConvertedCurrency  string         `json:"converted_currency"`   // Target currency
	Venue              string         `json:"venue,omitempty"`
	EventTime          time.Time      `json:"event_time"`
	IngestTime         time.Time      `json:"ingest_time"`
	Meta               Meta           `json:"meta"`
	FXMeta             *FXMeta        `json:"fx_meta,omitempty"`
}

// ConvertedFundamentalsLine represents a fundamentals line with converted currency
type ConvertedFundamentalsLine struct {
	Key                string        `json:"key"`
	Value              ScaledDecimal `json:"value"`
	OriginalCurrency   string        `json:"original_currency"`    // Source currency
	ConvertedCurrency  string        `json:"converted_currency"`   // Target currency
	PeriodStart        time.Time     `json:"period_start"`
	PeriodEnd          time.Time     `json:"period_end"`
}

// ConvertedFundamentalsSnapshot represents a fundamentals snapshot with converted currency
type ConvertedFundamentalsSnapshot struct {
	Security Security                      `json:"security"`
	Lines    []ConvertedFundamentalsLine   `json:"lines"`
	Source   string                        `json:"source"`
	AsOf     time.Time                     `json:"as_of"`
	Meta     Meta                          `json:"meta"`
	FXMeta   *FXMeta                       `json:"fx_meta,omitempty"`
}

// ConvertedMarketData represents market data with converted currency
type ConvertedMarketData struct {
	Security               Security       `json:"security"`
	RegularMarketPrice     *ScaledDecimal `json:"regular_market_price,omitempty"`
	RegularMarketHigh      *ScaledDecimal `json:"regular_market_high,omitempty"`
	RegularMarketLow       *ScaledDecimal `json:"regular_market_low,omitempty"`
	RegularMarketVolume    *int64         `json:"regular_market_volume,omitempty"`
	FiftyTwoWeekHigh       *ScaledDecimal `json:"fifty_two_week_high,omitempty"`
	FiftyTwoWeekLow        *ScaledDecimal `json:"fifty_two_week_low,omitempty"`
	PreviousClose          *ScaledDecimal `json:"previous_close,omitempty"`
	ChartPreviousClose     *ScaledDecimal `json:"chart_previous_close,omitempty"`
	OriginalCurrency       string         `json:"original_currency"`    // Source currency
	ConvertedCurrency      string         `json:"converted_currency"`   // Target currency
	RegularMarketTime      *time.Time     `json:"regular_market_time,omitempty"`
	HasPrePostMarketData   bool           `json:"has_pre_post_market_data"`
	EventTime              time.Time      `json:"event_time"`
	IngestTime             time.Time      `json:"ingest_time"`
	Meta                   Meta           `json:"meta"`
	FXMeta                 *FXMeta        `json:"fx_meta,omitempty"`
}

// ConvertTo converts a NormalizedBarBatch to a target currency
func (b *NormalizedBarBatch) ConvertTo(ctx context.Context, target string, fxConverter FXConverter) (*ConvertedBarBatch, *FXMeta, error) {
	if len(b.Bars) == 0 {
		return nil, nil, fmt.Errorf("no bars to convert")
	}

	// Get the source currency from the first bar
	sourceCurrency := b.Bars[0].CurrencyCode
	if sourceCurrency == "" {
		return nil, nil, fmt.Errorf("no source currency found in bars")
	}

	// If source and target are the same, no conversion needed
	if sourceCurrency == target {
		convertedBars := make([]ConvertedBar, len(b.Bars))
		for i, bar := range b.Bars {
			convertedBars[i] = ConvertedBar{
				Start:              bar.Start,
				End:                bar.End,
				Open:               bar.Open,
				High:               bar.High,
				Low:                bar.Low,
				Close:              bar.Close,
				OriginalCurrency:   sourceCurrency,
				ConvertedCurrency:  target,
				Volume:             bar.Volume,
				Adjusted:           bar.Adjusted,
				AdjustmentPolicyID: bar.AdjustmentPolicyID,
				EventTime:          bar.EventTime,
				IngestTime:         bar.IngestTime,
				AsOf:               bar.AsOf,
			}
		}

		fxMeta := &FXMeta{
			Provider: "none",
			Base:     sourceCurrency,
			Symbols:  []string{target},
			AsOf:     time.Now().UTC(),
		}
		return &ConvertedBarBatch{
			Security: b.Security,
			Bars:     convertedBars,
			Meta:     b.Meta,
			FXMeta:   fxMeta,
		}, fxMeta, nil
	}

	// Convert each bar
	convertedBars := make([]ConvertedBar, len(b.Bars))
	var fxMeta *FXMeta

	for i, bar := range b.Bars {
		// Convert prices
		open, meta, err := fxConverter.ConvertValue(ctx, bar.Open, sourceCurrency, target, bar.AsOf)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to convert open price for bar %d: %w", i, err)
		}
		fxMeta = meta

		high, _, err := fxConverter.ConvertValue(ctx, bar.High, sourceCurrency, target, bar.AsOf)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to convert high price for bar %d: %w", i, err)
		}

		low, _, err := fxConverter.ConvertValue(ctx, bar.Low, sourceCurrency, target, bar.AsOf)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to convert low price for bar %d: %w", i, err)
		}

		close, _, err := fxConverter.ConvertValue(ctx, bar.Close, sourceCurrency, target, bar.AsOf)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to convert close price for bar %d: %w", i, err)
		}

		convertedBars[i] = ConvertedBar{
			Start:              bar.Start,
			End:                bar.End,
			Open:               open,
			High:               high,
			Low:                low,
			Close:              close,
			OriginalCurrency:   sourceCurrency,
			ConvertedCurrency:  target,
			Volume:             bar.Volume,
			Adjusted:           bar.Adjusted,
			AdjustmentPolicyID: bar.AdjustmentPolicyID,
			EventTime:          bar.EventTime,
			IngestTime:         bar.IngestTime,
			AsOf:               bar.AsOf,
		}
	}

	return &ConvertedBarBatch{
		Security: b.Security,
		Bars:     convertedBars,
		Meta:     b.Meta,
		FXMeta:   fxMeta,
	}, fxMeta, nil
}

// ConvertTo converts a NormalizedQuote to a target currency
func (q *NormalizedQuote) ConvertTo(ctx context.Context, target string, fxConverter FXConverter) (*ConvertedQuote, *FXMeta, error) {
	sourceCurrency := q.CurrencyCode
	if sourceCurrency == "" {
		return nil, nil, fmt.Errorf("no source currency found in quote")
	}

	// If source and target are the same, no conversion needed
	if sourceCurrency == target {
		fxMeta := &FXMeta{
			Provider: "none",
			Base:     sourceCurrency,
			Symbols:  []string{target},
			AsOf:     time.Now().UTC(),
		}
		converted := &ConvertedQuote{
			Security:            q.Security,
			Type:                q.Type,
			Bid:                 q.Bid,
			BidSize:             q.BidSize,
			Ask:                 q.Ask,
			AskSize:             q.AskSize,
			RegularMarketPrice:  q.RegularMarketPrice,
			RegularMarketHigh:   q.RegularMarketHigh,
			RegularMarketLow:    q.RegularMarketLow,
			RegularMarketVolume: q.RegularMarketVolume,
			OriginalCurrency:    sourceCurrency,
			ConvertedCurrency:   target,
			Venue:               q.Venue,
			EventTime:           q.EventTime,
			IngestTime:          q.IngestTime,
			Meta:                q.Meta,
			FXMeta:              fxMeta,
		}
		return converted, fxMeta, nil
	}

	// Convert prices
	converted := &ConvertedQuote{
		Security:            q.Security,
		Type:                q.Type,
		BidSize:             q.BidSize,
		AskSize:             q.AskSize,
		RegularMarketVolume: q.RegularMarketVolume,
		OriginalCurrency:    sourceCurrency,
		ConvertedCurrency:   target,
		Venue:               q.Venue,
		EventTime:           q.EventTime,
		IngestTime:          q.IngestTime,
		Meta:                q.Meta,
	}

	var fxMeta *FXMeta

	// Convert bid if present
	if q.Bid != nil {
		bid, meta, err := fxConverter.ConvertValue(ctx, *q.Bid, sourceCurrency, target, q.EventTime)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to convert bid: %w", err)
		}
		converted.Bid = &bid
		fxMeta = meta
	}

	// Convert ask if present
	if q.Ask != nil {
		ask, meta, err := fxConverter.ConvertValue(ctx, *q.Ask, sourceCurrency, target, q.EventTime)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to convert ask: %w", err)
		}
		converted.Ask = &ask
		fxMeta = meta
	}

	// Convert regular market price if present
	if q.RegularMarketPrice != nil {
		price, meta, err := fxConverter.ConvertValue(ctx, *q.RegularMarketPrice, sourceCurrency, target, q.EventTime)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to convert regular market price: %w", err)
		}
		converted.RegularMarketPrice = &price
		fxMeta = meta
	}

	// Convert regular market high if present
	if q.RegularMarketHigh != nil {
		high, meta, err := fxConverter.ConvertValue(ctx, *q.RegularMarketHigh, sourceCurrency, target, q.EventTime)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to convert regular market high: %w", err)
		}
		converted.RegularMarketHigh = &high
		fxMeta = meta
	}

	// Convert regular market low if present
	if q.RegularMarketLow != nil {
		low, meta, err := fxConverter.ConvertValue(ctx, *q.RegularMarketLow, sourceCurrency, target, q.EventTime)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to convert regular market low: %w", err)
		}
		converted.RegularMarketLow = &low
		fxMeta = meta
	}

	converted.FXMeta = fxMeta
	return converted, fxMeta, nil
}

// ConvertTo converts a NormalizedFundamentalsSnapshot to a target currency
func (f *NormalizedFundamentalsSnapshot) ConvertTo(ctx context.Context, target string, fxConverter FXConverter) (*ConvertedFundamentalsSnapshot, *FXMeta, error) {
	if len(f.Lines) == 0 {
		return nil, nil, fmt.Errorf("no fundamentals lines to convert")
	}

	// Get the source currency from the first line
	sourceCurrency := f.Lines[0].CurrencyCode
	if sourceCurrency == "" {
		return nil, nil, fmt.Errorf("no source currency found in fundamentals")
	}

	// If source and target are the same, no conversion needed
	if sourceCurrency == target {
		convertedLines := make([]ConvertedFundamentalsLine, len(f.Lines))
		for i, line := range f.Lines {
			convertedLines[i] = ConvertedFundamentalsLine{
				Key:               line.Key,
				Value:             line.Value,
				OriginalCurrency:  sourceCurrency,
				ConvertedCurrency: target,
				PeriodStart:       line.PeriodStart,
				PeriodEnd:         line.PeriodEnd,
			}
		}

		fxMeta := &FXMeta{
			Provider: "none",
			Base:     sourceCurrency,
			Symbols:  []string{target},
			AsOf:     time.Now().UTC(),
		}
		return &ConvertedFundamentalsSnapshot{
			Security: f.Security,
			Lines:    convertedLines,
			Source:   f.Source,
			AsOf:     f.AsOf,
			Meta:     f.Meta,
			FXMeta:   fxMeta,
		}, fxMeta, nil
	}

	// Convert each line
	convertedLines := make([]ConvertedFundamentalsLine, len(f.Lines))
	var fxMeta *FXMeta

	for i, line := range f.Lines {
		value, meta, err := fxConverter.ConvertValue(ctx, line.Value, sourceCurrency, target, f.AsOf)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to convert value for line %d (%s): %w", i, line.Key, err)
		}
		fxMeta = meta

		convertedLines[i] = ConvertedFundamentalsLine{
			Key:               line.Key,
			Value:             value,
			OriginalCurrency:  sourceCurrency,
			ConvertedCurrency: target,
			PeriodStart:       line.PeriodStart,
			PeriodEnd:         line.PeriodEnd,
		}
	}

	return &ConvertedFundamentalsSnapshot{
		Security: f.Security,
		Lines:    convertedLines,
		Source:   f.Source,
		AsOf:     f.AsOf,
		Meta:     f.Meta,
		FXMeta:   fxMeta,
	}, fxMeta, nil
}

// ConvertTo converts a NormalizedMarketData to a target currency
func (m *NormalizedMarketData) ConvertTo(ctx context.Context, target string, fxConverter FXConverter) (*ConvertedMarketData, *FXMeta, error) {
	sourceCurrency := m.CurrencyCode
	if sourceCurrency == "" {
		return nil, nil, fmt.Errorf("no source currency found in market data")
	}

	// If source and target are the same, no conversion needed
	if sourceCurrency == target {
		fxMeta := &FXMeta{
			Provider: "none",
			Base:     sourceCurrency,
			Symbols:  []string{target},
			AsOf:     time.Now().UTC(),
		}
		converted := &ConvertedMarketData{
			Security:               m.Security,
			RegularMarketPrice:     m.RegularMarketPrice,
			RegularMarketHigh:      m.RegularMarketHigh,
			RegularMarketLow:       m.RegularMarketLow,
			RegularMarketVolume:    m.RegularMarketVolume,
			FiftyTwoWeekHigh:       m.FiftyTwoWeekHigh,
			FiftyTwoWeekLow:        m.FiftyTwoWeekLow,
			PreviousClose:          m.PreviousClose,
			ChartPreviousClose:     m.ChartPreviousClose,
			OriginalCurrency:       sourceCurrency,
			ConvertedCurrency:      target,
			RegularMarketTime:      m.RegularMarketTime,
			HasPrePostMarketData:   m.HasPrePostMarketData,
			EventTime:              m.EventTime,
			IngestTime:             m.IngestTime,
			Meta:                   m.Meta,
			FXMeta:                 fxMeta,
		}
		return converted, fxMeta, nil
	}

	// Convert prices
	converted := &ConvertedMarketData{
		Security:               m.Security,
		RegularMarketVolume:    m.RegularMarketVolume,
		OriginalCurrency:       sourceCurrency,
		ConvertedCurrency:      target,
		RegularMarketTime:      m.RegularMarketTime,
		HasPrePostMarketData:   m.HasPrePostMarketData,
		EventTime:              m.EventTime,
		IngestTime:             m.IngestTime,
		Meta:                   m.Meta,
	}

	var fxMeta *FXMeta

	// Convert all price fields
	priceFields := []struct {
		source *ScaledDecimal
		target **ScaledDecimal
		name   string
	}{
		{m.RegularMarketPrice, &converted.RegularMarketPrice, "regular market price"},
		{m.RegularMarketHigh, &converted.RegularMarketHigh, "regular market high"},
		{m.RegularMarketLow, &converted.RegularMarketLow, "regular market low"},
		{m.FiftyTwoWeekHigh, &converted.FiftyTwoWeekHigh, "52-week high"},
		{m.FiftyTwoWeekLow, &converted.FiftyTwoWeekLow, "52-week low"},
		{m.PreviousClose, &converted.PreviousClose, "previous close"},
		{m.ChartPreviousClose, &converted.ChartPreviousClose, "chart previous close"},
	}

	for _, field := range priceFields {
		if field.source != nil {
			convertedValue, meta, err := fxConverter.ConvertValue(ctx, *field.source, sourceCurrency, target, m.EventTime)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to convert %s: %w", field.name, err)
			}
			*field.target = &convertedValue
			fxMeta = meta
		}
	}

	converted.FXMeta = fxMeta
	return converted, fxMeta, nil
}