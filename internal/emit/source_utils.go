package emit

import (
	"strings"
)

// SourceType represents the different types of fundamental data
type SourceType string

const (
	SourceTypeFinancials             SourceType = "financials"
	SourceTypeKeyStatistics          SourceType = "key-statistics"
	SourceTypeAnalysis               SourceType = "analysis"
	SourceTypeAnalystInsights        SourceType = "analyst-insights"
	SourceTypeBalanceSheet           SourceType = "balance-sheet"
	SourceTypeCashFlow               SourceType = "cash-flow"
	SourceTypeComprehensiveFinancials SourceType = "comprehensive-financials"
	SourceTypeProfile                SourceType = "profile"
	SourceTypeNews                   SourceType = "news"
	SourceTypeUnknown                SourceType = "unknown"
)

// GetSourceType extracts the fundamental type from a source string
func GetSourceType(source string) SourceType {
	if source == "" {
		return SourceTypeUnknown
	}

	// Handle different source patterns
	if strings.Contains(source, "/scrape/") {
		// Extract the part after "/scrape/"
		parts := strings.Split(source, "/scrape/")
		if len(parts) > 1 {
			sourceType := strings.TrimSpace(parts[1])
			return SourceType(sourceType)
		}
	}

	// Handle other patterns - extract the last part after "/"
	if strings.Contains(source, "/") {
		parts := strings.Split(source, "/")
		lastPart := strings.TrimSpace(parts[len(parts)-1])
		if lastPart != "" {
			return SourceType(lastPart)
		}
	}

	// If no "/" found, return the whole string as type
	return SourceType(strings.TrimSpace(source))
}

// GetSourceBase extracts the base source (before the type)
func GetSourceBase(source string) string {
	if source == "" {
		return ""
	}

	if strings.Contains(source, "/scrape/") {
		parts := strings.Split(source, "/scrape/")
		if len(parts) > 0 {
			return parts[0] + "/scrape"
		}
	}

	if strings.Contains(source, "/") {
		parts := strings.Split(source, "/")
		if len(parts) > 1 {
			return strings.Join(parts[:len(parts)-1], "/")
		}
	}

	return source
}

// IsScrapedSource checks if the source indicates scraped data
func IsScrapedSource(source string) bool {
	return strings.Contains(source, "/scrape/")
}

// IsAPI Source checks if the source indicates API data
func IsAPISource(source string) bool {
	return !IsScrapedSource(source) && (strings.Contains(source, "yfinance") || strings.Contains(source, "yahoo"))
}

// GetSourceTypeDescription returns a human-readable description of the source type
func GetSourceTypeDescription(sourceType SourceType) string {
	switch sourceType {
	case SourceTypeFinancials:
		return "Income Statement Financials"
	case SourceTypeKeyStatistics:
		return "Key Statistics and Metrics"
	case SourceTypeAnalysis:
		return "Financial Analysis Data"
	case SourceTypeAnalystInsights:
		return "Analyst Insights and Estimates"
	case SourceTypeBalanceSheet:
		return "Balance Sheet Data"
	case SourceTypeCashFlow:
		return "Cash Flow Statement Data"
	case SourceTypeComprehensiveFinancials:
		return "Comprehensive Financial Data"
	case SourceTypeProfile:
		return "Company Profile Information"
	case SourceTypeNews:
		return "News and Articles"
	default:
		return "Unknown Data Type"
	}
}

// ValidateSourceType checks if a source type is valid
func ValidateSourceType(sourceType SourceType) bool {
	switch sourceType {
	case SourceTypeFinancials, SourceTypeKeyStatistics, SourceTypeAnalysis,
		SourceTypeAnalystInsights, SourceTypeBalanceSheet, SourceTypeCashFlow,
		SourceTypeComprehensiveFinancials, SourceTypeProfile, SourceTypeNews:
		return true
	default:
		return false
	}
}

