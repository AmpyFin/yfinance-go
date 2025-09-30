package norm

import (
	"math/big"
	"testing"
)

func TestRoundHalfUp(t *testing.T) {
	tests := []struct {
		name      string
		value     *big.Int
		fromScale int
		toScale   int
		expected  *big.Int
	}{
		{
			name:      "no rounding needed - scale up",
			value:     big.NewInt(12345),
			fromScale: 2,
			toScale:   4,
			expected:  big.NewInt(1234500),
		},
		{
			name:      "round up - remainder greater than half",
			value:     big.NewInt(1234567),
			fromScale: 4,
			toScale:   2,
			expected:  big.NewInt(12346),
		},
		{
			name:      "round down - remainder less than half",
			value:     big.NewInt(1234500),
			fromScale: 4,
			toScale:   2,
			expected:  big.NewInt(12345),
		},
		{
			name:      "round up - remainder exactly half (half-up)",
			value:     big.NewInt(1234550),
			fromScale: 4,
			toScale:   2,
			expected:  big.NewInt(12346),
		},
		{
			name:      "round up - remainder greater than half (duplicate test)",
			value:     big.NewInt(1234567),
			fromScale: 4,
			toScale:   2,
			expected:  big.NewInt(12346),
		},
		{
			name:      "round up - 0.5 case",
			value:     big.NewInt(5),
			fromScale: 1,
			toScale:   0,
			expected:  big.NewInt(1),
		},
		{
			name:      "round down - 0.4 case",
			value:     big.NewInt(4),
			fromScale: 1,
			toScale:   0,
			expected:  big.NewInt(0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RoundHalfUp(tt.value, tt.fromScale, tt.toScale)
			if result.Cmp(tt.expected) != 0 {
				t.Errorf("RoundHalfUp() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestMultiplyAndRound(t *testing.T) {
	tests := []struct {
		name        string
		value       ScaledDecimal
		rate        ScaledDecimal
		targetScale int
		expected    ScaledDecimal
		expectError bool
	}{
		{
			name: "simple multiplication",
			value: ScaledDecimal{
				Scaled: 10000, // 1.0000
				Scale:  4,
			},
			rate: ScaledDecimal{
				Scaled: 110000000, // 1.10000000
				Scale:  8,
			},
			targetScale: 4,
			expected: ScaledDecimal{
				Scaled: 11000, // 1.1000
				Scale:  4,
			},
			expectError: false,
		},
		{
			name: "rounding up",
			value: ScaledDecimal{
				Scaled: 10000, // 1.0000
				Scale:  4,
			},
			rate: ScaledDecimal{
				Scaled: 115000000, // 1.15000000
				Scale:  8,
			},
			targetScale: 4,
			expected: ScaledDecimal{
				Scaled: 11500, // 1.1500
				Scale:  4,
			},
			expectError: false,
		},
		{
			name: "rounding down",
			value: ScaledDecimal{
				Scaled: 10000, // 1.0000
				Scale:  4,
			},
			rate: ScaledDecimal{
				Scaled: 114000000, // 1.14000000
				Scale:  8,
			},
			targetScale: 4,
			expected: ScaledDecimal{
				Scaled: 11400, // 1.1400
				Scale:  4,
			},
			expectError: false,
		},
		{
			name: "JPY scale 2",
			value: ScaledDecimal{
				Scaled: 100, // 1.00
				Scale:  2,
			},
			rate: ScaledDecimal{
				Scaled: 110000000, // 1.10000000
				Scale:  8,
			},
			targetScale: 2,
			expected: ScaledDecimal{
				Scaled: 110, // 1.10
				Scale:  2,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := MultiplyAndRound(tt.value, tt.rate, tt.targetScale)
			if tt.expectError {
				if err == nil {
					t.Errorf("MultiplyAndRound() expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("MultiplyAndRound() unexpected error: %v", err)
				return
			}
			if result.Scaled != tt.expected.Scaled || result.Scale != tt.expected.Scale {
				t.Errorf("MultiplyAndRound() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetPriceScaleForCurrency(t *testing.T) {
	tests := []struct {
		currency string
		expected int
	}{
		{"JPY", 2},
		{"USD", 2},
		{"EUR", 2},
		{"GBP", 2},
		{"CAD", 2},
		{"AUD", 2},
		{"CHF", 2},
		{"NZD", 2},
		{"UNKNOWN", 2}, // default
	}

	for _, tt := range tests {
		t.Run(tt.currency, func(t *testing.T) {
			result := GetPriceScaleForCurrency(tt.currency)
			if result != tt.expected {
				t.Errorf("GetPriceScaleForCurrency(%s) = %d, want %d", tt.currency, result, tt.expected)
			}
		})
	}
}

func TestGetScaleForCurrency(t *testing.T) {
	tests := []struct {
		currency string
		expected int
	}{
		{"JPY", 2},
		{"USD", 2},
		{"EUR", 2},
		{"GBP", 2},
		{"CAD", 2},
		{"AUD", 2},
		{"CHF", 2},
		{"NZD", 2},
		{"UNKNOWN", 2}, // default
	}

	for _, tt := range tests {
		t.Run(tt.currency, func(t *testing.T) {
			result := GetScaleForCurrency(tt.currency)
			if result != tt.expected {
				t.Errorf("GetScaleForCurrency(%s) = %d, want %d", tt.currency, result, tt.expected)
			}
		})
	}
}
