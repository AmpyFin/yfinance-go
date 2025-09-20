package norm

import (
	"fmt"
	"math"
)

// GetScaleForCurrency returns the appropriate scale for a given currency
// USD/EUR/GBP use scale 4, JPY uses scale 2
func GetScaleForCurrency(currency string) int {
	switch currency {
	case "JPY":
		return 2
	case "USD", "EUR", "GBP", "CAD", "AUD", "CHF", "NZD":
		return 4
	default:
		// Default to scale 4 for most currencies
		return 4
	}
}

// ToScaledDecimal converts a float64 price to a scaled decimal
func ToScaledDecimal(price float64, scale int) (ScaledDecimal, error) {
	// Validate input
	if math.IsNaN(price) {
		return ScaledDecimal{}, fmt.Errorf("NaN price")
	}
	if math.IsInf(price, 0) {
		return ScaledDecimal{}, fmt.Errorf("infinite price")
	}
	
	// Calculate multiplier
	multiplier := math.Pow(10, float64(scale))
	
	// Round to avoid floating point precision issues
	scaled := int64(math.Round(price * multiplier))
	
	return ScaledDecimal{
		Scaled: scaled,
		Scale:  scale,
	}, nil
}

// ToScaledDecimalWithCurrency converts a float64 price to a scaled decimal using currency-appropriate scale
func ToScaledDecimalWithCurrency(price float64, currency string) (ScaledDecimal, error) {
	scale := GetScaleForCurrency(currency)
	
	// Note: JPY scaling may need adjustment based on data source
	// Currently using scale 2 for JPY, but this may need refinement
	// based on actual Yahoo Finance data format
	
	return ToScaledDecimal(price, scale)
}

// FromScaledDecimal converts a scaled decimal back to float64
func FromScaledDecimal(sd ScaledDecimal) float64 {
	multiplier := math.Pow(10, float64(sd.Scale))
	return float64(sd.Scaled) / multiplier
}

// ValidateScaledDecimal validates a scaled decimal
func ValidateScaledDecimal(sd ScaledDecimal) error {
	if sd.Scale < 0 {
		return fmt.Errorf("negative scale: %d", sd.Scale)
	}
	if sd.Scale > 8 {
		return fmt.Errorf("scale too large: %d", sd.Scale)
	}
	return nil
}
