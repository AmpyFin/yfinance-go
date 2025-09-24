package fx

import (
	"context"
	"testing"
	"time"
)

func TestNoneProvider(t *testing.T) {
	provider := NewNoneProvider()

	ctx := context.Background()
	base := "EUR"
	symbols := []string{"USD", "GBP"}
	at := time.Now().UTC()

	// Should always return error
	rates, asOf, err := provider.Rates(ctx, base, symbols, at)
	if err == nil {
		t.Error("Expected error from none provider")
	}

	// Verify error message
	expectedError := "FX conversion not enabled (provider: none). To enable conversion, configure fx.provider to 'yahoo-web' in your config"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
	}

	// Verify return values
	if rates != nil {
		t.Error("Expected nil rates")
	}
	if !asOf.IsZero() {
		t.Error("Expected zero time")
	}
}

func TestNoneProviderWithDifferentInputs(t *testing.T) {
	provider := NewNoneProvider()
	ctx := context.Background()

	testCases := []struct {
		name    string
		base    string
		symbols []string
		at      time.Time
	}{
		{
			name:    "USD to EUR",
			base:    "USD",
			symbols: []string{"EUR"},
			at:      time.Now().UTC(),
		},
		{
			name:    "JPY to USD",
			base:    "JPY",
			symbols: []string{"USD"},
			at:      time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		},
		{
			name:    "multiple symbols",
			base:    "EUR",
			symbols: []string{"USD", "GBP", "JPY"},
			at:      time.Now().UTC(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rates, asOf, err := provider.Rates(ctx, tc.base, tc.symbols, tc.at)
			if err == nil {
				t.Error("Expected error from none provider")
			}

			if rates != nil {
				t.Error("Expected nil rates")
			}
			if !asOf.IsZero() {
				t.Error("Expected zero time")
			}
		})
	}
}
