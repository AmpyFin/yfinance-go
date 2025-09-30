package norm

import (
	"context"
	"fmt"
	"time"
)

// MockFXConverter is a mock implementation of FXConverter for testing
type MockFXConverter struct {
	// Mock behavior - can be customized in tests
	ConvertValueFunc func(ctx context.Context, value ScaledDecimal, fromCurrency, toCurrency string, at time.Time) (ScaledDecimal, *FXMeta, error)
}

// ConvertValue implements the FXConverter interface
func (m *MockFXConverter) ConvertValue(ctx context.Context, value ScaledDecimal, fromCurrency, toCurrency string, at time.Time) (ScaledDecimal, *FXMeta, error) {
	if m.ConvertValueFunc != nil {
		return m.ConvertValueFunc(ctx, value, fromCurrency, toCurrency, at)
	}

	// Default behavior - return error for different currencies
	if fromCurrency != toCurrency {
		return ScaledDecimal{}, nil, fmt.Errorf("FX conversion not enabled (provider: none)")
	}

	// Same currency - return unchanged value
	return value, &FXMeta{
		Provider: "none",
		Base:     fromCurrency,
		Symbols:  []string{toCurrency},
		AsOf:     at,
	}, nil
}
