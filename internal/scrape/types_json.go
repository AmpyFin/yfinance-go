package scrape

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Scaled represents a scaled decimal number with precision preservation
type Scaled struct {
	Scaled int64 `json:"scaled"`
	Scale  int   `json:"scale"` // e.g., 2 for cents, 6 for micro-units
}

// Currency represents an ISO-4217 currency code
type Currency = string

// YahooNum represents Yahoo's numeric format with raw, fmt, and longFmt
type YahooNum struct {
	Raw     *float64 `json:"raw,omitempty"`
	Fmt     string   `json:"fmt,omitempty"`
	LongFmt string   `json:"longFmt,omitempty"`
}

// YahooInt represents Yahoo's integer format with raw, fmt, and longFmt
type YahooInt struct {
	Raw     *int64  `json:"raw,omitempty"`
	Fmt     string  `json:"fmt,omitempty"`
	LongFmt string  `json:"longFmt,omitempty"`
}

// YahooString represents Yahoo's string format that might contain numbers
type YahooString struct {
	Raw     *string `json:"raw,omitempty"`
	Fmt     string  `json:"fmt,omitempty"`
	LongFmt string  `json:"longFmt,omitempty"`
}

// KeyStatisticsDTO represents extracted key statistics data
type KeyStatisticsDTO struct {
	Symbol                    string   `json:"symbol"`
	Market                    string   `json:"market"`
	Currency                  Currency `json:"currency"`
	
	// Market metrics (from summaryDetail - real-time data)
	MarketCap                 *Scaled  `json:"market_cap,omitempty"`
	ForwardPE                 *Scaled  `json:"forward_pe,omitempty"`
	TrailingPE                *Scaled  `json:"trailing_pe,omitempty"`
	Beta                      *Scaled  `json:"beta,omitempty"`
	PriceToSales              *Scaled  `json:"price_to_sales,omitempty"`
	
	// Share data
	SharesOutstanding         *int64   `json:"shares_outstanding,omitempty"`
	FloatShares               *int64   `json:"float_shares,omitempty"`
	ShortInterest             *int64   `json:"short_interest,omitempty"`
	
	// Financial metrics (from financialData)
	EnterpriseValue           *Scaled  `json:"enterprise_value,omitempty"`
	TotalCash                 *Scaled  `json:"total_cash,omitempty"`
	TotalDebt                 *Scaled  `json:"total_debt,omitempty"`
	QuickRatio                *Scaled  `json:"quick_ratio,omitempty"`
	CurrentRatio              *Scaled  `json:"current_ratio,omitempty"`
	DebtToEquity              *Scaled  `json:"debt_to_equity,omitempty"`
	ReturnOnAssets            *Scaled  `json:"return_on_assets,omitempty"`
	ReturnOnEquity            *Scaled  `json:"return_on_equity,omitempty"`
	GrossMargins              *Scaled  `json:"gross_margins,omitempty"`
	OperatingMargins          *Scaled  `json:"operating_margins,omitempty"`
	ProfitMargins             *Scaled  `json:"profit_margins,omitempty"`
	RevenueGrowth             *Scaled  `json:"revenue_growth,omitempty"`
	EarningsGrowth            *Scaled  `json:"earnings_growth,omitempty"`
	
	// Price data
	FiftyTwoWeekHigh          *Scaled  `json:"fifty_two_week_high,omitempty"`
	FiftyTwoWeekLow           *Scaled  `json:"fifty_two_week_low,omitempty"`
	AverageVolume             *int64   `json:"average_volume,omitempty"`
	AverageVolume10Day        *int64   `json:"average_volume_10_day,omitempty"`
	
	AsOf                      time.Time `json:"as_of"`
}

// PeriodLine represents a financial statement line item for a specific period
type PeriodLine struct {
	PeriodStart time.Time `json:"period_start"`
	PeriodEnd   time.Time `json:"period_end"`
	Key         string    `json:"key"`
	Value       Scaled    `json:"value"`
	Currency    Currency  `json:"currency"`
}

// FinancialsDTO represents extracted financial statements data
type FinancialsDTO struct {
	Symbol   string        `json:"symbol"`
	Market   string        `json:"market"`
	Lines    []PeriodLine  `json:"lines"`
	AsOf     time.Time     `json:"as_of"`
}

// Recommendation represents analyst recommendation data for a period
type Recommendation struct {
	Period     string `json:"period"`
	StrongBuy  int    `json:"strong_buy"`
	Buy        int    `json:"buy"`
	Hold       int    `json:"hold"`
	Sell       int    `json:"sell"`
	StrongSell int    `json:"strong_sell"`
}

// QuarterlyEPS represents quarterly EPS estimates and actuals
type QuarterlyEPS struct {
	Date     string  `json:"date"`
	Actual   *Scaled `json:"actual,omitempty"`
	Estimate *Scaled `json:"estimate,omitempty"`
}

// AnalysisDTO represents extracted analysis data
type AnalysisDTO struct {
	Symbol       string          `json:"symbol"`
	Market       string          `json:"market"`
	Currency     Currency        `json:"currency"`
	RecTrends    []Recommendation `json:"rec_trends"`
	EPSQuarterly []QuarterlyEPS  `json:"eps_quarterly"`
	AsOf         time.Time       `json:"as_of"`
}

// Officer represents a company officer/executive
type Officer struct {
	Name  string  `json:"name"`
	Title string  `json:"title"`
	Age   *int    `json:"age,omitempty"`
	Pay   *Scaled `json:"pay,omitempty"`
}

// ProfileDTO represents extracted company profile data
type ProfileDTO struct {
	Symbol    string   `json:"symbol"`
	Market    string   `json:"market"`
	Company   string   `json:"company"`
	Address1  string   `json:"address1"`
	City      string   `json:"city"`
	State     string   `json:"state"`
	Country   string   `json:"country"`
	Phone     string   `json:"phone"`
	Website   string   `json:"website"`
	Industry  string   `json:"industry"`
	Sector    string   `json:"sector"`
	Employees *int     `json:"employees,omitempty"`
	Officers  []Officer `json:"officers"`
	AsOf      time.Time `json:"as_of"`
}

// Numeric coercion functions

// NumToScaled converts a YahooNum to a Scaled value with the given scale
func NumToScaled(n YahooNum, scale int) (Scaled, bool) {
	if n.Raw == nil {
		return Scaled{}, false
	}
	
	// Convert to scaled integer
	multiplier := int64(1)
	for i := 0; i < scale; i++ {
		multiplier *= 10
	}
	
	scaled := int64(*n.Raw * float64(multiplier))
	return Scaled{Scaled: scaled, Scale: scale}, true
}

// IntToScaled converts a YahooInt to a Scaled value with the given scale
func IntToScaled(i YahooInt, scale int) (Scaled, bool) {
	if i.Raw == nil {
		return Scaled{}, false
	}
	
	// For integers, we typically use scale 0 unless it's a percentage
	actualScale := scale
	if scale == 0 {
		actualScale = 0
	}
	
	multiplier := int64(1)
	for j := 0; j < actualScale; j++ {
		multiplier *= 10
	}
	
	scaled := *i.Raw * multiplier
	return Scaled{Scaled: scaled, Scale: actualScale}, true
}

// CoerceCurrency extracts currency from various Yahoo formats
func CoerceCurrency(v any) (Currency, bool) {
	switch val := v.(type) {
	case string:
		// Clean up currency string
		currency := strings.TrimSpace(strings.ToUpper(val))
		if len(currency) == 3 {
			return currency, true
		}
		return "", false
	case map[string]any:
		// Check for currency field in nested objects
		if curr, ok := val["currency"].(string); ok {
			return CoerceCurrency(curr)
		}
		return "", false
	default:
		return "", false
	}
}

// ParseYahooDate parses various Yahoo date formats
func ParseYahooDate(ts any) (time.Time, bool) {
	switch val := ts.(type) {
	case float64:
		// Unix timestamp in seconds
		return time.Unix(int64(val), 0).UTC(), true
	case int64:
		// Unix timestamp in seconds
		return time.Unix(val, 0).UTC(), true
	case string:
		// Try various date formats
		formats := []string{
			"2006-01-02",
			"2006-01-02T15:04:05Z",
			"2006-01-02T15:04:05.000Z",
			"2006-01-02 15:04:05",
		}
		
		for _, format := range formats {
			if t, err := time.Parse(format, val); err == nil {
				return t.UTC(), true
			}
		}
		return time.Time{}, false
	default:
		return time.Time{}, false
	}
}

// ParseYahooPeriod parses Yahoo's period format (e.g., "2023-12-31")
func ParseYahooPeriod(periodStr string) (time.Time, time.Time, bool) {
	// Yahoo periods are typically year-end dates
	if t, ok := ParseYahooDate(periodStr); ok {
		// Assume it's the end of the period, start is beginning of year
		year := t.Year()
		start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
		return start, t, true
	}
	return time.Time{}, time.Time{}, false
}

// StringToInt64 safely converts a string to int64
func StringToInt64(s string) (int64, bool) {
	if s == "" {
		return 0, false
	}
	
	// Remove common formatting
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, " ", "")
	
	val, err := strconv.ParseInt(s, 10, 64)
	return val, err == nil
}

// StringToFloat64 safely converts a string to float64
func StringToFloat64(s string) (float64, bool) {
	if s == "" {
		return 0, false
	}
	
	// Remove common formatting
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, " ", "")
	
	val, err := strconv.ParseFloat(s, 64)
	return val, err == nil
}

// Format helpers

// String returns a human-readable representation of Scaled
func (s Scaled) String() string {
	if s.Scale == 0 {
		return fmt.Sprintf("%d", s.Scaled)
	}
	
	divisor := float64(1)
	for i := 0; i < s.Scale; i++ {
		divisor *= 10
	}
	
	value := float64(s.Scaled) / divisor
	return fmt.Sprintf("%.6f", value)
}

// Float64 returns the float64 value of Scaled
func (s Scaled) Float64() float64 {
	if s.Scale == 0 {
		return float64(s.Scaled)
	}
	
	divisor := float64(1)
	for i := 0; i < s.Scale; i++ {
		divisor *= 10
	}
	
	return float64(s.Scaled) / divisor
}

// Helper functions to convert raw struct fields to YahooNum/YahooInt

// ToYahooNum converts a raw struct to YahooNum
func ToYahooNum(raw *float64, fmt, longFmt string) YahooNum {
	return YahooNum{
		Raw:     raw,
		Fmt:     fmt,
		LongFmt: longFmt,
	}
}

// ToYahooInt converts a raw struct to YahooInt
func ToYahooInt(raw *int64, fmt, longFmt string) YahooInt {
	return YahooInt{
		Raw:     raw,
		Fmt:     fmt,
		LongFmt: longFmt,
	}
}
