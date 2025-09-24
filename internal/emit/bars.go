package emit

import (
	"fmt"

	"github.com/AmpyFin/yfinance-go/internal/norm"
	barsv1 "github.com/AmpyFin/ampy-proto/v2/gen/go/ampy/bars/v1"
	commonv1 "github.com/AmpyFin/ampy-proto/v2/gen/go/ampy/common/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// EmitBarBatch converts a NormalizedBarBatch to ampy.bars.v1.BarBatch
func EmitBarBatch(n *norm.NormalizedBarBatch) (*barsv1.BarBatch, error) {
	if n == nil {
		return nil, fmt.Errorf("normalized bar batch cannot be nil")
	}
	
	// Validate security
	if err := ValidateSecurity(n.Security); err != nil {
		return nil, fmt.Errorf("security validation failed: %w", err)
	}
	
	// Convert bars
	ampyBars := make([]*barsv1.Bar, 0, len(n.Bars))
	for i, bar := range n.Bars {
		ampyBar, err := emitBar(&bar, &n.Security)
		if err != nil {
			return nil, fmt.Errorf("failed to emit bar %d: %w", i, err)
		}
		ampyBars = append(ampyBars, ampyBar)
	}
	
	return &barsv1.BarBatch{
		Bars: ampyBars,
	}, nil
}

// emitBar converts a single NormalizedBar to ampy.bars.v1.Bar
func emitBar(n *norm.NormalizedBar, security *norm.Security) (*barsv1.Bar, error) {
	if n == nil {
		return nil, fmt.Errorf("normalized bar cannot be nil")
	}
	
	// Validate time window for daily bars
	if err := ValidateTimeWindow(n.Start, n.End, n.EventTime); err != nil {
		return nil, fmt.Errorf("time window validation failed: %w", err)
	}
	
	// Validate currency
	if err := ValidateCurrency(n.CurrencyCode); err != nil {
		return nil, fmt.Errorf("currency validation failed: %w", err)
	}
	
	// Validate adjustments
	if err := ValidateAdjustments(n.Adjusted, n.AdjustmentPolicyID); err != nil {
		return nil, fmt.Errorf("adjustment validation failed: %w", err)
	}
	
	// Validate and convert decimals
	open, err := emitDecimal(&n.Open)
	if err != nil {
		return nil, fmt.Errorf("open price validation failed: %w", err)
	}
	
	high, err := emitDecimal(&n.High)
	if err != nil {
		return nil, fmt.Errorf("high price validation failed: %w", err)
	}
	
	low, err := emitDecimal(&n.Low)
	if err != nil {
		return nil, fmt.Errorf("low price validation failed: %w", err)
	}
	
	close, err := emitDecimal(&n.Close)
	if err != nil {
		return nil, fmt.Errorf("close price validation failed: %w", err)
	}
	
	// Convert security
	ampySecurity := emitSecurity(security)
	
	// Convert timestamps
	start := timestamppb.New(n.Start)
	end := timestamppb.New(n.End)
	eventTime := timestamppb.New(n.EventTime)
	ingestTime := timestamppb.New(n.IngestTime)
	asOf := timestamppb.New(n.AsOf)
	
	// Convert adjustment policy
	adjustmentPolicy := convertAdjustmentPolicy(n.AdjustmentPolicyID)
	
	return &barsv1.Bar{
		Security:            ampySecurity,
		Start:               start,
		End:                 end,
		Open:                open,
		High:                high,
		Low:                 low,
		Close:               close,
		Volume:              n.Volume,
		Adjusted:            n.Adjusted,
		AdjustmentPolicyId:  n.AdjustmentPolicyID,
		AdjustmentPolicy:    adjustmentPolicy,
		EventTime:           eventTime,
		IngestTime:          ingestTime,
		AsOf:                asOf,
		// Note: Meta is not included in individual bars in ampy-proto
		// It's typically handled at the batch level
	}, nil
}

// emitDecimal converts a ScaledDecimal to ampy.common.v1.Decimal
func emitDecimal(d *norm.ScaledDecimal) (*commonv1.Decimal, error) {
	if d == nil {
		return nil, fmt.Errorf("decimal cannot be nil")
	}
	
	if err := ValidateDecimal(*d); err != nil {
		return nil, err
	}
	
	return &commonv1.Decimal{
		Scaled: d.Scaled,
		Scale:  int32(d.Scale),
	}, nil
}

// emitSecurity converts a Security to ampy.common.v1.SecurityId
func emitSecurity(s *norm.Security) *commonv1.SecurityId {
	if s == nil {
		return nil
	}
	
	return &commonv1.SecurityId{
		Symbol: s.Symbol,
		Mic:    s.MIC,
		// Figi and Isin are optional and not available in our normalized types
	}
}

// convertAdjustmentPolicy converts string policy to enum
func convertAdjustmentPolicy(policyID string) commonv1.AdjustmentPolicy {
	switch policyID {
	case "raw":
		return commonv1.AdjustmentPolicy_ADJUSTMENT_POLICY_RAW
	case "split_only":
		return commonv1.AdjustmentPolicy_ADJUSTMENT_POLICY_SPLIT_ONLY
	case "split_dividend":
		return commonv1.AdjustmentPolicy_ADJUSTMENT_POLICY_SPLIT_DIVIDEND
	default:
		return commonv1.AdjustmentPolicy_ADJUSTMENT_POLICY_UNSPECIFIED
	}
}
