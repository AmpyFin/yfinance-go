package emit

import (
	"math"
	"testing"

	"github.com/AmpyFin/yfinance-go/internal/scrape"
	commonv1 "github.com/AmpyFin/ampy-proto/v2/gen/go/ampy/common/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMoneyToScaled(t *testing.T) {
	config := DefaultScaledDecimalConfig()

	// Test valid money value
	value := &scrape.Scaled{
		Scaled: 12345,
		Scale:  2,
	}

	decimal, err := MoneyToScaled(value, config)
	require.NoError(t, err)
	assert.Equal(t, int64(12345), decimal.Scaled)
	assert.Equal(t, int32(2), decimal.Scale)

	// Test nil value
	_, err = MoneyToScaled(nil, config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "scaled value cannot be nil")

	// Test invalid scale
	value.Scale = 15 // Too high
	_, err = MoneyToScaled(value, config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "scale validation failed")

	// Test negative scale (should fail with default config)
	value.Scale = -1
	_, err = MoneyToScaled(value, config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "negative scale")
}

func TestPercentToScaled(t *testing.T) {
	config := DefaultScaledDecimalConfig()

	// Test percent value that needs scale conversion
	value := &scrape.Scaled{
		Scaled: 1234, // 12.34%
		Scale:  2,
	}

	decimal, err := PercentToScaled(value, config)
	require.NoError(t, err)
	// Should be converted to scale=6 (multiply by 10^4 to go from scale 2 to scale 6)
	assert.Equal(t, int64(12340000), decimal.Scaled) // 12.34% -> 12340000 at scale 6
	assert.Equal(t, int32(6), decimal.Scale)

	// Test percent value already at correct scale
	value.Scale = 6
	value.Scaled = 12340000

	decimal, err = PercentToScaled(value, config)
	require.NoError(t, err)
	assert.Equal(t, int64(12340000), decimal.Scaled)
	assert.Equal(t, int32(6), decimal.Scale)
}

func TestFloatToScaled(t *testing.T) {
	config := DefaultScaledDecimalConfig()

	// Test normal float conversion
	decimal, err := FloatToScaled(123.45, 2, config)
	require.NoError(t, err)
	assert.Equal(t, int64(12345), decimal.Scaled)
	assert.Equal(t, int32(2), decimal.Scale)

	// Test rounding
	decimal, err = FloatToScaled(123.456, 2, config)
	require.NoError(t, err)
	assert.Equal(t, int64(12346), decimal.Scaled) // Rounded up
	assert.Equal(t, int32(2), decimal.Scale)

	// Test zero
	decimal, err = FloatToScaled(0.0, 2, config)
	require.NoError(t, err)
	assert.Equal(t, int64(0), decimal.Scaled)
	assert.Equal(t, int32(2), decimal.Scale)

	// Test negative value
	decimal, err = FloatToScaled(-123.45, 2, config)
	require.NoError(t, err)
	assert.Equal(t, int64(-12345), decimal.Scaled)
	assert.Equal(t, int32(2), decimal.Scale)

	// Test NaN
	_, err = FloatToScaled(math.NaN(), 2, config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot convert NaN")

	// Test infinity
	_, err = FloatToScaled(math.Inf(1), 2, config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot convert infinity")

	// Test overflow
	_, err = FloatToScaled(math.MaxFloat64, 9, config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "would overflow")
}

func TestScaledToFloat(t *testing.T) {
	// Test normal conversion
	decimal := &commonv1.Decimal{
		Scaled: 12345,
		Scale:  2,
	}

	value, err := ScaledToFloat(decimal)
	require.NoError(t, err)
	assert.InDelta(t, 123.45, value, 0.001)

	// Test zero
	decimal.Scaled = 0
	value, err = ScaledToFloat(decimal)
	require.NoError(t, err)
	assert.Equal(t, 0.0, value)

	// Test negative
	decimal.Scaled = -12345
	value, err = ScaledToFloat(decimal)
	require.NoError(t, err)
	assert.InDelta(t, -123.45, value, 0.001)

	// Test nil decimal
	_, err = ScaledToFloat(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "decimal cannot be nil")

	// Test invalid scale
	decimal = &commonv1.Decimal{
		Scaled: 12345,
		Scale:  20, // Too high for float64 precision
	}
	_, err = ScaledToFloat(decimal)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "out of reasonable range")
}

func TestAttachCurrency(t *testing.T) {
	testCases := []struct {
		input       string
		expected    string
		shouldError bool
		errorMsg    string
	}{
		{"USD", "USD", false, ""},
		{"usd", "USD", false, ""},
		{"eur", "EUR", false, ""},
		{"gbp", "GBP", false, ""},
		{"", "", true, "currency code cannot be empty"},
		{"US", "", true, "currency code must be 3 characters"},
		{"USDX", "", true, "currency code must be 3 characters"},
		{"123", "", true, "currency code must contain only uppercase letters"},
		{"XYZ", "XYZ", true, "currency validation warning"}, // Unknown but valid format
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result, err := AttachCurrency(tc.input)

			if tc.shouldError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMsg)
			} else {
				assert.NoError(t, err)
			}

			if tc.expected != "" {
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestGetCurrencyInfo(t *testing.T) {
	// Test major currency
	info, err := GetCurrencyInfo("USD")
	require.NoError(t, err)
	assert.Equal(t, "USD", info.Code)
	assert.Equal(t, "US Dollar", info.Name)
	assert.Equal(t, "$", info.Symbol)
	assert.Equal(t, 2, info.MinorUnits)

	// Test Japanese Yen (no minor units)
	info, err = GetCurrencyInfo("JPY")
	require.NoError(t, err)
	assert.Equal(t, "JPY", info.Code)
	assert.Equal(t, "Japanese Yen", info.Name)
	assert.Equal(t, "¥", info.Symbol)
	assert.Equal(t, 0, info.MinorUnits)

	// Test unknown currency
	info, err = GetCurrencyInfo("XYZ")
	require.NoError(t, err) // Should not error, but return basic info
	assert.Equal(t, "XYZ", info.Code)
	assert.Contains(t, info.Name, "Unknown Currency")
	assert.Equal(t, 2, info.MinorUnits) // Default

	// Test invalid format
	_, err = GetCurrencyInfo("US")
	assert.Error(t, err)
}

func TestRecommendedScale(t *testing.T) {
	// Test major currencies
	scale, err := RecommendedScale("USD")
	assert.NoError(t, err)
	assert.Equal(t, 2, scale)

	scale, err = RecommendedScale("JPY")
	assert.NoError(t, err)
	assert.Equal(t, 0, scale)

	// Test unknown currency (should default to 2)
	scale, err = RecommendedScale("XYZ")
	assert.Equal(t, 2, scale) // Should return default even on error
}

func TestConvertScale(t *testing.T) {
	// Test increasing scale (more precision)
	value := &scrape.Scaled{
		Scaled: 12345,
		Scale:  2,
	}

	converted, err := convertScale(value, 4)
	require.NoError(t, err)
	assert.Equal(t, int64(1234500), converted.Scaled) // 123.45 -> 123.4500
	assert.Equal(t, 4, converted.Scale)

	// Test decreasing scale (less precision)
	value = &scrape.Scaled{
		Scaled: 1234567,
		Scale:  4,
	}

	converted, err = convertScale(value, 2)
	require.NoError(t, err)
	assert.Equal(t, int64(12346), converted.Scaled) // 123.4567 -> 123.46 (rounded)
	assert.Equal(t, 2, converted.Scale)

	// Test same scale (no conversion)
	converted, err = convertScale(value, 4)
	require.NoError(t, err)
	assert.Equal(t, value.Scaled, converted.Scaled)
	assert.Equal(t, value.Scale, converted.Scale)

	// Test rounding negative numbers
	value = &scrape.Scaled{
		Scaled: -1234567,
		Scale:  4,
	}

	converted, err = convertScale(value, 2)
	require.NoError(t, err)
	assert.Equal(t, int64(-12346), converted.Scaled) // -123.4567 -> -123.46 (rounded)

	// Test overflow
	value = &scrape.Scaled{
		Scaled: math.MaxInt64 / 10, // Large number
		Scale:  2,
	}

	_, err = convertScale(value, 4) // Would overflow
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "scale conversion would overflow")

	// Test invalid target scale
	_, err = convertScale(value, -1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "target scale")

	_, err = convertScale(value, 20)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "target scale")
}

func TestDefaultScaledDecimalConfig(t *testing.T) {
	config := DefaultScaledDecimalConfig()
	
	assert.Equal(t, 2, config.DefaultScale)
	assert.Equal(t, 6, config.PercentScale)
	assert.False(t, config.AllowNegativeScale)
	assert.Equal(t, 9, config.MaxScale)
}
