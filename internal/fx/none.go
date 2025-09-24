package fx

import (
	"context"
	"fmt"
	"time"

	"github.com/yeonlee/yfinance-go/internal/norm"
)

// NoneProvider is the default FX provider that returns errors for conversion requests
type NoneProvider struct{}

// NewNoneProvider creates a new none provider
func NewNoneProvider() *NoneProvider {
	return &NoneProvider{}
}

// Rates returns an error indicating that FX conversion is not enabled
func (p *NoneProvider) Rates(ctx context.Context, base string, symbols []string, at time.Time) (map[string]norm.ScaledDecimal, time.Time, error) {
	return nil, time.Time{}, fmt.Errorf("FX conversion not enabled (provider: none). To enable conversion, configure fx.provider to 'yahoo-web' in your config")
}
