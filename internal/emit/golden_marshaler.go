package emit

import (
	"time"

	"github.com/yeonlee/yfinance-go/internal/norm"
	barsv1 "github.com/AmpyFin/ampy-proto/v2/gen/go/ampy/bars/v1"
	ticksv1 "github.com/AmpyFin/ampy-proto/v2/gen/go/ampy/ticks/v1"
	fundamentalsv1 "github.com/AmpyFin/ampy-proto/v2/gen/go/ampy/fundamentals/v1"
)

// GoldenBarBatch represents the expected golden format for bar batches
type GoldenBarBatch struct {
	Security norm.Security `json:"security"`
	Bars     []GoldenBar   `json:"bars"`
	Meta     norm.Meta     `json:"meta"`
}

// GoldenBar represents the expected golden format for bars
type GoldenBar struct {
	Start               time.Time     `json:"start"`
	End                 time.Time     `json:"end"`
	Open                norm.ScaledDecimal `json:"open"`
	High                norm.ScaledDecimal `json:"high"`
	Low                 norm.ScaledDecimal `json:"low"`
	Close               norm.ScaledDecimal `json:"close"`
	Volume              int64         `json:"volume"`
	Adjusted            bool          `json:"adjusted"`
	AdjustmentPolicyID  string        `json:"adjustment_policy_id"`
	EventTime           time.Time     `json:"event_time"`
	IngestTime          time.Time     `json:"ingest_time"`
	AsOf                time.Time     `json:"as_of"`
}

// GoldenQuote represents the expected golden format for quotes
type GoldenQuote struct {
	Security   norm.Security `json:"security"`
	Type       string        `json:"type"`
	Bid        *norm.ScaledDecimal `json:"bid,omitempty"`
	BidSize    *int64        `json:"bid_size,omitempty"`
	Ask        *norm.ScaledDecimal `json:"ask,omitempty"`
	AskSize    *int64        `json:"ask_size,omitempty"`
	Venue      string        `json:"venue"`
	EventTime  time.Time     `json:"event_time"`
	IngestTime time.Time     `json:"ingest_time"`
	Meta       norm.Meta     `json:"meta"`
}

// GoldenFundamentals represents the expected golden format for fundamentals
type GoldenFundamentals struct {
	Security norm.Security                `json:"security"`
	Lines    []norm.NormalizedFundamentalsLine `json:"lines"`
	Source   string                       `json:"source"`
	AsOf     time.Time                    `json:"as_of"`
	Meta     norm.Meta                    `json:"meta"`
}

// ToGoldenBarBatch converts ampy-proto BarBatch to golden format
func ToGoldenBarBatch(barBatch *barsv1.BarBatch, security norm.Security, meta norm.Meta) *GoldenBarBatch {
	goldenBars := make([]GoldenBar, 0, len(barBatch.Bars))
	
	for _, bar := range barBatch.Bars {
		goldenBar := GoldenBar{
			Start:               bar.Start.AsTime(),
			End:                 bar.End.AsTime(),
			Open:                norm.ScaledDecimal{Scaled: bar.Open.Scaled, Scale: int(bar.Open.Scale)},
			High:                norm.ScaledDecimal{Scaled: bar.High.Scaled, Scale: int(bar.High.Scale)},
			Low:                 norm.ScaledDecimal{Scaled: bar.Low.Scaled, Scale: int(bar.Low.Scale)},
			Close:               norm.ScaledDecimal{Scaled: bar.Close.Scaled, Scale: int(bar.Close.Scale)},
			Volume:              bar.Volume,
			Adjusted:            bar.Adjusted,
			AdjustmentPolicyID:  bar.AdjustmentPolicyId,
			EventTime:           bar.EventTime.AsTime(),
			IngestTime:          bar.IngestTime.AsTime(),
			AsOf:                bar.AsOf.AsTime(),
		}
		goldenBars = append(goldenBars, goldenBar)
	}
	
	return &GoldenBarBatch{
		Security: security,
		Bars:     goldenBars,
		Meta:     meta,
	}
}

// ToGoldenQuote converts ampy-proto QuoteTick to golden format
func ToGoldenQuote(quote *ticksv1.QuoteTick, meta norm.Meta) *GoldenQuote {
	var bid, ask *norm.ScaledDecimal
	var bidSize, askSize *int64
	
	if quote.Bid != nil {
		bid = &norm.ScaledDecimal{Scaled: quote.Bid.Scaled, Scale: int(quote.Bid.Scale)}
	}
	if quote.Ask != nil {
		ask = &norm.ScaledDecimal{Scaled: quote.Ask.Scaled, Scale: int(quote.Ask.Scale)}
	}
	if quote.BidSize != 0 {
		bidSize = &quote.BidSize
	}
	if quote.AskSize != 0 {
		askSize = &quote.AskSize
	}
	
	return &GoldenQuote{
		Security:   norm.Security{Symbol: quote.Security.Symbol, MIC: quote.Security.Mic},
		Type:       "QUOTE",
		Bid:        bid,
		BidSize:    bidSize,
		Ask:        ask,
		AskSize:    askSize,
		Venue:      quote.Venue,
		EventTime:  quote.EventTime.AsTime(),
		IngestTime: quote.IngestTime.AsTime(),
		Meta:       meta,
	}
}

// ToGoldenFundamentals converts ampy-proto FundamentalsSnapshot to golden format
func ToGoldenFundamentals(fundamentals *fundamentalsv1.FundamentalsSnapshot, meta norm.Meta) *GoldenFundamentals {
	goldenLines := make([]norm.NormalizedFundamentalsLine, 0, len(fundamentals.Lines))
	
	for _, line := range fundamentals.Lines {
		goldenLine := norm.NormalizedFundamentalsLine{
			Key:          line.Key,
			Value:        norm.ScaledDecimal{Scaled: line.Value.Scaled, Scale: int(line.Value.Scale)},
			CurrencyCode: line.CurrencyCode,
			PeriodStart:  line.PeriodStart.AsTime(),
			PeriodEnd:    line.PeriodEnd.AsTime(),
		}
		goldenLines = append(goldenLines, goldenLine)
	}
	
	return &GoldenFundamentals{
		Security: norm.Security{Symbol: fundamentals.Security.Symbol, MIC: fundamentals.Security.Mic},
		Lines:    goldenLines,
		Source:   fundamentals.Source,
		AsOf:     fundamentals.AsOf.AsTime(),
		Meta:     meta,
	}
}

// MarshalToGoldenJSON marshals to the expected golden JSON format
func MarshalToGoldenJSON(v interface{}) ([]byte, error) {
	return CanonicalMarshaler.Marshal(v)
}
