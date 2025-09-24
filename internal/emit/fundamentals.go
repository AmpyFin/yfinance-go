package emit

import (
	"fmt"

	"github.com/AmpyFin/yfinance-go/internal/norm"
	fundamentalsv1 "github.com/AmpyFin/ampy-proto/v2/gen/go/ampy/fundamentals/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// EmitFundamentals converts a NormalizedFundamentalsSnapshot to ampy.fundamentals.v1.FundamentalsSnapshot
func EmitFundamentals(n *norm.NormalizedFundamentalsSnapshot) (*fundamentalsv1.FundamentalsSnapshot, error) {
	if n == nil {
		return nil, fmt.Errorf("normalized fundamentals snapshot cannot be nil")
	}
	
	// Validate security
	if err := ValidateSecurity(n.Security); err != nil {
		return nil, fmt.Errorf("security validation failed: %w", err)
	}
	
	// Validate fundamentals lines
	if err := ValidateFundamentals(n.Lines); err != nil {
		return nil, fmt.Errorf("fundamentals validation failed: %w", err)
	}
	
	// Convert security
	ampySecurity := emitSecurity(&n.Security)
	
	// Convert line items
	ampyLines := make([]*fundamentalsv1.LineItem, 0, len(n.Lines))
	for i, line := range n.Lines {
		ampyLine, err := emitLineItem(&line)
		if err != nil {
			return nil, fmt.Errorf("failed to emit line item %d: %w", i, err)
		}
		ampyLines = append(ampyLines, ampyLine)
	}
	
	// Convert timestamps
	asOf := timestamppb.New(n.AsOf)
	
	// Convert metadata
	ampyMeta := emitMeta(&n.Meta)
	
	return &fundamentalsv1.FundamentalsSnapshot{
		Security: ampySecurity,
		Lines:    ampyLines,
		Source:   n.Source,
		AsOf:     asOf,
		Meta:     ampyMeta,
	}, nil
}

// emitLineItem converts a NormalizedFundamentalsLine to ampy.fundamentals.v1.LineItem
func emitLineItem(n *norm.NormalizedFundamentalsLine) (*fundamentalsv1.LineItem, error) {
	if n == nil {
		return nil, fmt.Errorf("normalized fundamentals line cannot be nil")
	}
	
	// Validate and convert decimal value
	value, err := emitDecimal(&n.Value)
	if err != nil {
		return nil, fmt.Errorf("value validation failed: %w", err)
	}
	
	// Validate currency
	if err := ValidateCurrency(n.CurrencyCode); err != nil {
		return nil, fmt.Errorf("currency validation failed: %w", err)
	}
	
	// Convert timestamps
	periodStart := timestamppb.New(n.PeriodStart)
	periodEnd := timestamppb.New(n.PeriodEnd)
	
	return &fundamentalsv1.LineItem{
		Key:          n.Key,
		Value:        value,
		CurrencyCode: n.CurrencyCode,
		PeriodStart:  periodStart,
		PeriodEnd:    periodEnd,
	}, nil
}
