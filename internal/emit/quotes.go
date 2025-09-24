package emit

import (
	"fmt"

	"github.com/AmpyFin/yfinance-go/internal/norm"
	ticksv1 "github.com/AmpyFin/ampy-proto/v2/gen/go/ampy/ticks/v1"
	commonv1 "github.com/AmpyFin/ampy-proto/v2/gen/go/ampy/common/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// EmitQuote converts a NormalizedQuote to ampy.ticks.v1.QuoteTick
func EmitQuote(n *norm.NormalizedQuote) (*ticksv1.QuoteTick, error) {
	if n == nil {
		return nil, fmt.Errorf("normalized quote cannot be nil")
	}
	
	// Validate security
	if err := ValidateSecurity(n.Security); err != nil {
		return nil, fmt.Errorf("security validation failed: %w", err)
	}
	
	// Validate currency
	if err := ValidateCurrency(n.CurrencyCode); err != nil {
		return nil, fmt.Errorf("currency validation failed: %w", err)
	}
	
	// Convert security
	ampySecurity := emitSecurity(&n.Security)
	
	// Convert bid/ask prices if present
	var bid, ask *commonv1.Decimal
	var err error
	
	if n.Bid != nil {
		bid, err = emitDecimal(n.Bid)
		if err != nil {
			return nil, fmt.Errorf("bid price validation failed: %w", err)
		}
	}
	
	if n.Ask != nil {
		ask, err = emitDecimal(n.Ask)
		if err != nil {
			return nil, fmt.Errorf("ask price validation failed: %w", err)
		}
	}
	
	// Convert timestamps
	eventTime := timestamppb.New(n.EventTime)
	ingestTime := timestamppb.New(n.IngestTime)
	
	// Convert metadata
	ampyMeta := emitMeta(&n.Meta)
	
	return &ticksv1.QuoteTick{
		Security:   ampySecurity,
		Bid:        bid,
		BidSize:    getInt64Value(n.BidSize),
		Ask:        ask,
		AskSize:    getInt64Value(n.AskSize),
		Venue:      n.Venue,
		EventTime:  eventTime,
		IngestTime: ingestTime,
		Meta:       ampyMeta,
	}, nil
}

// emitMeta converts a Meta to ampy.common.v1.Meta
func emitMeta(m *norm.Meta) *commonv1.Meta {
	if m == nil {
		return nil
	}
	
	return &commonv1.Meta{
		RunId:         m.RunID,
		Source:        m.Source,
		Producer:      m.Producer,
		SchemaVersion: m.SchemaVersion,
		// Checksum is optional and not available in our normalized types
	}
}

// getInt64Value safely gets int64 value from pointer
func getInt64Value(ptr *int64) int64 {
	if ptr == nil {
		return 0
	}
	return *ptr
}
