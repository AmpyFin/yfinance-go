package norm

import (
	"fmt"
	"math"
)

// GetScaleForCurrency returns the appropriate scale for a given currency
// USD/EUR/GBP use scale 2 (cents), JPY uses scale 2
func GetScaleForCurrency(currency string) int {
	switch currency {
	case "JPY":
		return 2
	case "USD", "EUR", "GBP", "CAD", "AUD", "CHF", "NZD":
		return 2  // Use scale 2 for cents
	default:
		// Default to scale 2 for most currencies
		return 2
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

// GetPriceScaleForCurrency returns the appropriate scale for price data in a given currency
// This is an alias for GetScaleForCurrency for compatibility
func GetPriceScaleForCurrency(currency string) int {
	return GetScaleForCurrency(currency)
}

// MultiplyAndRound multiplies two scaled decimals and rounds to the target scale
func MultiplyAndRound(a ScaledDecimal, b ScaledDecimal, targetScale int) (ScaledDecimal, error) {
	// Validate inputs
	if err := ValidateScaledDecimal(a); err != nil {
		return ScaledDecimal{}, fmt.Errorf("invalid first operand: %w", err)
	}
	if err := ValidateScaledDecimal(b); err != nil {
		return ScaledDecimal{}, fmt.Errorf("invalid second operand: %w", err)
	}
	if targetScale < 0 || targetScale > 8 {
		return ScaledDecimal{}, fmt.Errorf("invalid target scale: %d", targetScale)
	}
	
	// Perform multiplication using int64 to avoid floating point precision issues
	// result = (a.Scaled * b.Scaled) / (10^(a.Scale + b.Scale - targetScale))
	
	// Calculate the total scale of the multiplication
	totalScale := a.Scale + b.Scale
	
	// Calculate the divisor to get to target scale
	scaleDiff := totalScale - targetScale
	
	if scaleDiff < 0 {
		// Need to multiply by 10^(-scaleDiff)
		multiplier := int64(math.Pow(10, float64(-scaleDiff)))
		result := a.Scaled * b.Scaled * multiplier
		return ScaledDecimal{
			Scaled: result,
			Scale:  targetScale,
		}, nil
	} else if scaleDiff > 0 {
		// Need to divide by 10^scaleDiff
		divisor := int64(math.Pow(10, float64(scaleDiff)))
		result := a.Scaled * b.Scaled / divisor
		
		// Handle rounding for the remainder
		remainder := (a.Scaled * b.Scaled) % divisor
		if remainder >= divisor/2 {
			result++
		}
		
		return ScaledDecimal{
			Scaled: result,
			Scale:  targetScale,
		}, nil
	} else {
		// No scaling needed
		return ScaledDecimal{
			Scaled: a.Scaled * b.Scaled,
			Scale:  targetScale,
		}, nil
	}
}
