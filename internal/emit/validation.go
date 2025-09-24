package emit

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/AmpyFin/yfinance-go/internal/norm"
)

// Validation errors
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error in field '%s': %s", e.Field, e.Message)
}

// ValidateSecurity validates a security identifier
func ValidateSecurity(sec norm.Security) error {
	if sec.Symbol == "" {
		return ValidationError{Field: "symbol", Message: "symbol cannot be empty"}
	}
	
	if sec.MIC != "" {
		// MIC must be uppercase and 4 characters
		if len(sec.MIC) != 4 {
			return ValidationError{Field: "mic", Message: "MIC must be exactly 4 characters"}
		}
		micRegex := regexp.MustCompile(`^[A-Z0-9]{4}$`)
		if !micRegex.MatchString(sec.MIC) {
			return ValidationError{Field: "mic", Message: "MIC must be uppercase alphanumeric (4 chars)"}
		}
	}
	
	return nil
}

// ValidateTimeWindow validates time window for daily bars
func ValidateTimeWindow(start, end, event time.Time) error {
	// For daily bars: end = start + 24h, event = end
	expectedEnd := start.Add(24 * time.Hour)
	if !end.Equal(expectedEnd) {
		return ValidationError{
			Field:   "time_window",
			Message: fmt.Sprintf("end time must be start + 24h, got %v, expected %v", end, expectedEnd),
		}
	}
	
	if !event.Equal(end) {
		return ValidationError{
			Field:   "event_time",
			Message: fmt.Sprintf("event_time must equal end time for daily bars, got %v, expected %v", event, end),
		}
	}
	
	return nil
}

// ValidateDecimal validates a scaled decimal
func ValidateDecimal(d norm.ScaledDecimal) error {
	if d.Scale < 0 || d.Scale > 9 {
		return ValidationError{
			Field:   "scale",
			Message: fmt.Sprintf("scale must be between 0 and 9, got %d", d.Scale),
		}
	}
	
	// Note: d.Scaled is already int64, so it's guaranteed to be within int64 range
	
	return nil
}

// ValidateCurrency validates ISO-4217 currency code
func ValidateCurrency(code string) error {
	if code == "" {
		return ValidationError{Field: "currency_code", Message: "currency code cannot be empty"}
	}
	
	// Basic validation: 3 uppercase letters
	if len(code) != 3 {
		return ValidationError{
			Field:   "currency_code",
			Message: fmt.Sprintf("currency code must be 3 characters, got %d", len(code)),
		}
	}
	
	currencyRegex := regexp.MustCompile(`^[A-Z]{3}$`)
	if !currencyRegex.MatchString(code) {
		return ValidationError{
			Field:   "currency_code",
			Message: "currency code must be 3 uppercase letters",
		}
	}
	
	// Check against known major currencies (pass-through for others)
	majorCurrencies := map[string]bool{
		"USD": true, "EUR": true, "GBP": true, "JPY": true, "CHF": true,
		"CAD": true, "AUD": true, "NZD": true, "SEK": true, "NOK": true,
		"DKK": true, "PLN": true, "CZK": true, "HUF": true, "RUB": true,
		"CNY": true, "KRW": true, "SGD": true, "HKD": true, "INR": true,
		"BRL": true, "MXN": true, "ZAR": true, "TRY": true, "ILS": true,
	}
	
	if !majorCurrencies[code] {
		// Allow pass-through for other currencies but log a warning
		// In production, you might want to maintain a more comprehensive list
		_ = code // Suppress unused variable warning
	}
	
	return nil
}

// ValidateAdjustments validates adjustment policy consistency
func ValidateAdjustments(adjusted bool, policyID string) error {
	validPolicies := map[string]bool{
		"raw":             true,
		"split_only":      true,
		"split_dividend":  true,
	}
	
	if !validPolicies[policyID] {
		return ValidationError{
			Field:   "adjustment_policy_id",
			Message: fmt.Sprintf("invalid adjustment policy: %s, must be one of: raw, split_only, split_dividend", policyID),
		}
	}
	
	// Validate consistency
	if adjusted && policyID == "raw" {
		return ValidationError{
			Field:   "adjustment_consistency",
			Message: "cannot have adjusted=true with policy=raw",
		}
	}
	
	if !adjusted && (policyID == "split_only" || policyID == "split_dividend") {
		return ValidationError{
			Field:   "adjustment_consistency",
			Message: fmt.Sprintf("cannot have adjusted=false with policy=%s", policyID),
		}
	}
	
	return nil
}

// ValidateFundamentals validates fundamentals line items
func ValidateFundamentals(lines []norm.NormalizedFundamentalsLine) error {
	// Whitelist of allowed keys (can be extended)
	allowedKeys := map[string]bool{
		"revenue":                 true,
		"total_revenue":           true,
		"gross_profit":            true,
		"operating_income":        true,
		"net_income":              true,
		"eps_basic":               true,
		"total_assets":            true,
		"total_liabilities":       true,
		"shareholders_equity":     true,
		"cash_and_equivalents":    true,
		"total_debt":              true,
		"free_cash_flow":          true,
		"operating_cash_flow":     true,
		"investing_cash_flow":     true,
		"financing_cash_flow":     true,
		"earnings_per_share":      true,
		"book_value_per_share":    true,
		"price_to_earnings":       true,
		"price_to_book":           true,
		"debt_to_equity":          true,
		"return_on_equity":        true,
		"return_on_assets":        true,
	}
	
	for i, line := range lines {
		// Check if key is whitelisted or has allowed prefix
		if !allowedKeys[line.Key] && !strings.HasPrefix(line.Key, "custom_") {
			return ValidationError{
				Field:   fmt.Sprintf("lines[%d].key", i),
				Message: fmt.Sprintf("key '%s' not in whitelist and doesn't have 'custom_' prefix", line.Key),
			}
		}
		
		// Validate decimal
		if err := ValidateDecimal(line.Value); err != nil {
			return fmt.Errorf("lines[%d].value: %w", i, err)
		}
		
		// Validate currency
		if err := ValidateCurrency(line.CurrencyCode); err != nil {
			return fmt.Errorf("lines[%d].currency_code: %w", i, err)
		}
		
		// Validate period
		if line.PeriodStart.After(line.PeriodEnd) {
			return ValidationError{
				Field:   fmt.Sprintf("lines[%d].period", i),
				Message: "period_start must be before period_end",
			}
		}
	}
	
	return nil
}
