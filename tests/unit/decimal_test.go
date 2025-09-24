package unit

import (
	"math"
	"math/big"
	"testing"

	"github.com/AmpyFin/yfinance-go/internal/norm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoundHalfUp(t *testing.T) {
	tests := []struct {
		name       string
		value      *big.Int
		fromScale  int
		toScale    int
		expected   *big.Int
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
		{
			name:      "edge case - very small remainder",
			value:     big.NewInt(1234501),
			fromScale: 4,
			toScale:   2,
			expected:  big.NewInt(12345),
		},
		{
			name:      "edge case - very large remainder",
			value:     big.NewInt(1234599),
			fromScale: 4,
			toScale:   2,
			expected:  big.NewInt(12346),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := norm.RoundHalfUp(tt.value, tt.fromScale, tt.toScale)
			assert.Equal(t, tt.expected.Cmp(result), 0, "Expected %v, got %v", tt.expected, result)
		})
	}
}

func TestMultiplyAndRound(t *testing.T) {
	tests := []struct {
		name         string
		value        norm.ScaledDecimal
		rate         norm.ScaledDecimal
		targetScale  int
		expected     norm.ScaledDecimal
		expectError  bool
	}{
		{
			name: "simple multiplication",
			value: norm.ScaledDecimal{
				Scaled: 10000, // 1.0000
				Scale:  4,
			},
			rate: norm.ScaledDecimal{
				Scaled: 110000000, // 1.10000000
				Scale:  8,
			},
			targetScale: 4,
			expected: norm.ScaledDecimal{
				Scaled: 11000, // 1.1000
				Scale:  4,
			},
			expectError: false,
		},
		{
			name: "rounding up",
			value: norm.ScaledDecimal{
				Scaled: 10000, // 1.0000
				Scale:  4,
			},
			rate: norm.ScaledDecimal{
				Scaled: 115000000, // 1.15000000
				Scale:  8,
			},
			targetScale: 4,
			expected: norm.ScaledDecimal{
				Scaled: 11500, // 1.1500
				Scale:  4,
			},
			expectError: false,
		},
		{
			name: "rounding down",
			value: norm.ScaledDecimal{
				Scaled: 10000, // 1.0000
				Scale:  4,
			},
			rate: norm.ScaledDecimal{
				Scaled: 114000000, // 1.14000000
				Scale:  8,
			},
			targetScale: 4,
			expected: norm.ScaledDecimal{
				Scaled: 11400, // 1.1400
				Scale:  4,
			},
			expectError: false,
		},
		{
			name: "JPY scale 2",
			value: norm.ScaledDecimal{
				Scaled: 100, // 1.00
				Scale:  2,
			},
			rate: norm.ScaledDecimal{
				Scaled: 110000000, // 1.10000000
				Scale:  8,
			},
			targetScale: 2,
			expected: norm.ScaledDecimal{
				Scaled: 110, // 1.10
				Scale:  2,
			},
			expectError: false,
		},
		{
			name: "edge case - half-up rounding",
			value: norm.ScaledDecimal{
				Scaled: 10000, // 1.0000
				Scale:  4,
			},
			rate: norm.ScaledDecimal{
				Scaled: 115500000, // 1.15500000 (exactly 0.5 remainder)
				Scale:  8,
			},
			targetScale: 4,
			expected: norm.ScaledDecimal{
				Scaled: 11550, // 1.1550 (rounds up)
				Scale:  4,
			},
			expectError: false,
		},
		{
			name: "invalid scale - negative",
			value: norm.ScaledDecimal{
				Scaled: 10000,
				Scale:  -1,
			},
			rate: norm.ScaledDecimal{
				Scaled: 110000000,
				Scale:  8,
			},
			targetScale: 4,
			expectError: true,
		},
		{
			name: "invalid target scale",
			value: norm.ScaledDecimal{
				Scaled: 10000,
				Scale:  4,
			},
			rate: norm.ScaledDecimal{
				Scaled: 110000000,
				Scale:  8,
			},
			targetScale: -1,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := norm.MultiplyAndRound(tt.value, tt.rate, tt.targetScale)
			if tt.expectError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
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
			result := norm.GetPriceScaleForCurrency(tt.currency)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestToScaledDecimal(t *testing.T) {
	tests := []struct {
		name      string
		price     float64
		scale     int
		expected  norm.ScaledDecimal
		expectErr bool
	}{
		{
			name:     "normal price",
			price:    123.45,
			scale:    2,
			expected: norm.ScaledDecimal{Scaled: 12345, Scale: 2},
		},
		{
			name:     "price with more precision",
			price:    123.4567,
			scale:    4,
			expected: norm.ScaledDecimal{Scaled: 1234567, Scale: 4},
		},
		{
			name:     "zero price",
			price:    0.0,
			scale:    2,
			expected: norm.ScaledDecimal{Scaled: 0, Scale: 2},
		},
		{
			name:      "NaN price",
			price:     math.NaN(),
			scale:     2,
			expectErr: true,
		},
		{
			name:      "infinite price",
			price:     math.Inf(1),
			scale:     2,
			expectErr: true,
		},
		{
			name:      "negative infinite price",
			price:     math.Inf(-1),
			scale:     2,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := norm.ToScaledDecimal(tt.price, tt.scale)
			if tt.expectErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateScaledDecimal(t *testing.T) {
	tests := []struct {
		name      string
		decimal   norm.ScaledDecimal
		expectErr bool
	}{
		{
			name:      "valid decimal",
			decimal:   norm.ScaledDecimal{Scaled: 12345, Scale: 2},
			expectErr: false,
		},
		{
			name:      "negative scale",
			decimal:   norm.ScaledDecimal{Scaled: 12345, Scale: -1},
			expectErr: true,
		},
		{
			name:      "scale too large",
			decimal:   norm.ScaledDecimal{Scaled: 12345, Scale: 10},
			expectErr: true,
		},
		{
			name:      "zero scale",
			decimal:   norm.ScaledDecimal{Scaled: 12345, Scale: 0},
			expectErr: false,
		},
		{
			name:      "max scale",
			decimal:   norm.ScaledDecimal{Scaled: 12345, Scale: 8},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := norm.ValidateScaledDecimal(tt.decimal)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
