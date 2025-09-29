package emit

import (
	"fmt"
	"math"
	"strings"

	"github.com/AmpyFin/yfinance-go/internal/scrape"
	commonv1 "github.com/AmpyFin/ampy-proto/v2/gen/go/ampy/common/v1"
)

// ScaledDecimalConfig holds configuration for decimal conversion
type ScaledDecimalConfig struct {
	DefaultScale       int  // Default scale for monetary values (typically 2)
	PercentScale       int  // Scale for percentage values (typically 6)
	AllowNegativeScale bool // Whether to allow negative scales
	MaxScale           int  // Maximum allowed scale (typically 9)
}

// DefaultScaledDecimalConfig returns sensible defaults for decimal conversion
func DefaultScaledDecimalConfig() *ScaledDecimalConfig {
	return &ScaledDecimalConfig{
		DefaultScale:       2,
		PercentScale:       6,
		AllowNegativeScale: false,
		MaxScale:           9,
	}
}

// MoneyToScaled converts a Scaled value to ampy.common.v1.Decimal with monetary semantics
func MoneyToScaled(v *scrape.Scaled, config *ScaledDecimalConfig) (*commonv1.Decimal, error) {
	if v == nil {
		return nil, fmt.Errorf("scaled value cannot be nil")
	}

	if config == nil {
		config = DefaultScaledDecimalConfig()
	}

	// Validate scale
	if err := validateScale(v.Scale, config); err != nil {
		return nil, fmt.Errorf("money scale validation failed: %w", err)
	}

	return &commonv1.Decimal{
		Scaled: v.Scaled,
		Scale:  int32(v.Scale),
	}, nil
}

// PercentToScaled converts a Scaled value to ampy.common.v1.Decimal with percentage semantics
func PercentToScaled(v *scrape.Scaled, config *ScaledDecimalConfig) (*commonv1.Decimal, error) {
	if v == nil {
		return nil, fmt.Errorf("scaled value cannot be nil")
	}

	if config == nil {
		config = DefaultScaledDecimalConfig()
	}

	// For percentages, we typically want higher precision (scale=6)
	// Convert to percentage scale if needed
	targetScale := config.PercentScale
	if v.Scale != targetScale {
		converted, err := convertScale(v, targetScale)
		if err != nil {
			return nil, fmt.Errorf("percent scale conversion failed: %w", err)
		}
		v = converted
	}

	// Validate scale
	if err := validateScale(v.Scale, config); err != nil {
		return nil, fmt.Errorf("percent scale validation failed: %w", err)
	}

	return &commonv1.Decimal{
		Scaled: v.Scaled,
		Scale:  int32(v.Scale),
	}, nil
}

// FloatToScaled converts a float64 to ampy.common.v1.Decimal with the specified scale
func FloatToScaled(value float64, scale int, config *ScaledDecimalConfig) (*commonv1.Decimal, error) {
	if config == nil {
		config = DefaultScaledDecimalConfig()
	}

	// Validate scale
	if err := validateScale(scale, config); err != nil {
		return nil, fmt.Errorf("float scale validation failed: %w", err)
	}

	// Check for special float values
	if math.IsNaN(value) {
		return nil, fmt.Errorf("cannot convert NaN to scaled decimal")
	}
	if math.IsInf(value, 0) {
		return nil, fmt.Errorf("cannot convert infinity to scaled decimal")
	}

	// Convert to scaled integer
	multiplier := math.Pow10(scale)
	scaled := int64(math.Round(value * multiplier))

	// Check for overflow
	if math.Abs(value*multiplier) > math.MaxInt64 {
		return nil, fmt.Errorf("value %f with scale %d would overflow int64", value, scale)
	}

	return &commonv1.Decimal{
		Scaled: scaled,
		Scale:  int32(scale),
	}, nil
}

// ScaledToFloat converts ampy.common.v1.Decimal to float64
func ScaledToFloat(decimal *commonv1.Decimal) (float64, error) {
	if decimal == nil {
		return 0, fmt.Errorf("decimal cannot be nil")
	}

	if decimal.Scale < 0 || decimal.Scale > 15 { // Reasonable bounds for float64 precision
		return 0, fmt.Errorf("scale %d out of reasonable range for float64 conversion", decimal.Scale)
	}

	divisor := math.Pow10(int(decimal.Scale))
	return float64(decimal.Scaled) / divisor, nil
}

// convertScale converts a Scaled value from one scale to another
func convertScale(v *scrape.Scaled, targetScale int) (*scrape.Scaled, error) {
	if v.Scale == targetScale {
		return v, nil // No conversion needed
	}

	if targetScale < 0 || targetScale > 15 {
		return nil, fmt.Errorf("target scale %d out of reasonable range", targetScale)
	}

	var newScaled int64

	if targetScale > v.Scale {
		// Increasing scale (more precision)
		scaleDiff := targetScale - v.Scale
		multiplier := int64(math.Pow10(scaleDiff))
		
		// Check for overflow
		if v.Scaled > math.MaxInt64/multiplier {
			return nil, fmt.Errorf("scale conversion would overflow: %d * 10^%d", v.Scaled, scaleDiff)
		}
		
		newScaled = v.Scaled * multiplier
	} else {
		// Decreasing scale (less precision)
		scaleDiff := v.Scale - targetScale
		divisor := int64(math.Pow10(scaleDiff))
		
		// Round to nearest
		newScaled = (v.Scaled + divisor/2) / divisor
		if v.Scaled < 0 {
			newScaled = (v.Scaled - divisor/2) / divisor
		}
	}

	return &scrape.Scaled{
		Scaled: newScaled,
		Scale:  targetScale,
	}, nil
}

// validateScale validates that a scale is within acceptable bounds
func validateScale(scale int, config *ScaledDecimalConfig) error {
	if !config.AllowNegativeScale && scale < 0 {
		return fmt.Errorf("negative scale %d not allowed", scale)
	}

	if scale > config.MaxScale {
		return fmt.Errorf("scale %d exceeds maximum %d", scale, config.MaxScale)
	}

	return nil
}

// AttachCurrency validates and normalizes ISO-4217 currency codes
func AttachCurrency(code string) (string, error) {
	if code == "" {
		return "", fmt.Errorf("currency code cannot be empty")
	}

	// Normalize to uppercase
	normalized := strings.ToUpper(strings.TrimSpace(code))

	// Basic validation: 3 uppercase letters
	if len(normalized) != 3 {
		return "", fmt.Errorf("currency code must be 3 characters, got %d", len(normalized))
	}

	// Check that all characters are letters
	for _, char := range normalized {
		if char < 'A' || char > 'Z' {
			return "", fmt.Errorf("currency code must contain only uppercase letters, got '%s'", normalized)
		}
	}

	// Validate against known major currencies
	if err := validateCurrencyCode(normalized); err != nil {
		// Return warning but allow pass-through
		return normalized, fmt.Errorf("currency validation warning: %w", err)
	}

	return normalized, nil
}

// validateCurrencyCode validates against known ISO-4217 currency codes
func validateCurrencyCode(code string) error {
	// Major world currencies (not exhaustive, but covers most common ones)
	majorCurrencies := map[string]string{
		"USD": "US Dollar",
		"EUR": "Euro",
		"GBP": "British Pound",
		"JPY": "Japanese Yen",
		"CHF": "Swiss Franc",
		"CAD": "Canadian Dollar",
		"AUD": "Australian Dollar",
		"NZD": "New Zealand Dollar",
		"SEK": "Swedish Krona",
		"NOK": "Norwegian Krone",
		"DKK": "Danish Krone",
		"PLN": "Polish Zloty",
		"CZK": "Czech Koruna",
		"HUF": "Hungarian Forint",
		"RUB": "Russian Ruble",
		"CNY": "Chinese Yuan",
		"KRW": "South Korean Won",
		"SGD": "Singapore Dollar",
		"HKD": "Hong Kong Dollar",
		"INR": "Indian Rupee",
		"BRL": "Brazilian Real",
		"MXN": "Mexican Peso",
		"ZAR": "South African Rand",
		"TRY": "Turkish Lira",
		"ILS": "Israeli Shekel",
		"THB": "Thai Baht",
		"MYR": "Malaysian Ringgit",
		"IDR": "Indonesian Rupiah",
		"PHP": "Philippine Peso",
		"VND": "Vietnamese Dong",
		"TWD": "Taiwan Dollar",
		"AED": "UAE Dirham",
		"SAR": "Saudi Riyal",
		"EGP": "Egyptian Pound",
		"CLP": "Chilean Peso",
		"COP": "Colombian Peso",
		"PEN": "Peruvian Sol",
		"ARS": "Argentine Peso",
	}

	if _, exists := majorCurrencies[code]; !exists {
		return fmt.Errorf("currency '%s' not in major currencies list", code)
	}

	return nil
}

// CurrencyInfo provides information about a currency
type CurrencyInfo struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Symbol      string `json:"symbol,omitempty"`
	MinorUnits  int    `json:"minor_units"` // Number of minor units (e.g., 2 for USD cents)
}

// GetCurrencyInfo returns information about a currency code
func GetCurrencyInfo(code string) (*CurrencyInfo, error) {
	normalized, err := AttachCurrency(code)
	if err != nil && !strings.Contains(err.Error(), "warning") {
		return nil, err
	}

	// Currency information map
	currencyInfoMap := map[string]CurrencyInfo{
		"USD": {Code: "USD", Name: "US Dollar", Symbol: "$", MinorUnits: 2},
		"EUR": {Code: "EUR", Name: "Euro", Symbol: "€", MinorUnits: 2},
		"GBP": {Code: "GBP", Name: "British Pound", Symbol: "£", MinorUnits: 2},
		"JPY": {Code: "JPY", Name: "Japanese Yen", Symbol: "¥", MinorUnits: 0},
		"CHF": {Code: "CHF", Name: "Swiss Franc", Symbol: "CHF", MinorUnits: 2},
		"CAD": {Code: "CAD", Name: "Canadian Dollar", Symbol: "C$", MinorUnits: 2},
		"AUD": {Code: "AUD", Name: "Australian Dollar", Symbol: "A$", MinorUnits: 2},
		"CNY": {Code: "CNY", Name: "Chinese Yuan", Symbol: "¥", MinorUnits: 2},
		"KRW": {Code: "KRW", Name: "South Korean Won", Symbol: "₩", MinorUnits: 0},
		"INR": {Code: "INR", Name: "Indian Rupee", Symbol: "₹", MinorUnits: 2},
		"BRL": {Code: "BRL", Name: "Brazilian Real", Symbol: "R$", MinorUnits: 2},
		"MXN": {Code: "MXN", Name: "Mexican Peso", Symbol: "MX$", MinorUnits: 2},
		"HKD": {Code: "HKD", Name: "Hong Kong Dollar", Symbol: "HK$", MinorUnits: 2},
		"SGD": {Code: "SGD", Name: "Singapore Dollar", Symbol: "S$", MinorUnits: 2},
		"TWD": {Code: "TWD", Name: "Taiwan Dollar", Symbol: "NT$", MinorUnits: 2},
	}

	if info, exists := currencyInfoMap[normalized]; exists {
		return &info, nil
	}

	// Return basic info for unknown currencies
	return &CurrencyInfo{
		Code:       normalized,
		Name:       fmt.Sprintf("Unknown Currency (%s)", normalized),
		MinorUnits: 2, // Default to 2 minor units
	}, nil
}

// RecommendedScale returns the recommended scale for a given currency
func RecommendedScale(currencyCode string) (int, error) {
	info, err := GetCurrencyInfo(currencyCode)
	if err != nil {
		return 2, err // Default to 2 on error
	}

	return info.MinorUnits, nil
}
