package scrape

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// RegexConfig holds the regex patterns for statistics extraction
type RegexConfig struct {
	Current struct {
		MarketCap              string `yaml:"market_cap"`
		EnterpriseValue        string `yaml:"enterprise_value"`
		TrailingPE             string `yaml:"trailing_pe"`
		ForwardPE              string `yaml:"forward_pe"`
		PEGRatio               string `yaml:"peg_ratio"`
		PriceSales             string `yaml:"price_sales"`
		PriceBook              string `yaml:"price_book"`
		EnterpriseValueRevenue string `yaml:"enterprise_value_revenue"`
		EnterpriseValueEBITDA  string `yaml:"enterprise_value_ebitda"`
	} `yaml:"current"`

	Additional struct {
		Beta              string `yaml:"beta"`
		SharesOutstanding string `yaml:"shares_outstanding"`
		ProfitMargin      string `yaml:"profit_margin"`
		OperatingMargin   string `yaml:"operating_margin"`
		ReturnOnAssets    string `yaml:"return_on_assets"`
		ReturnOnEquity    string `yaml:"return_on_equity"`
	} `yaml:"additional"`

	HistoricalColumns struct {
		Column2 ColumnPatterns `yaml:"column_2"`
		Column3 ColumnPatterns `yaml:"column_3"`
		Column4 ColumnPatterns `yaml:"column_4"`
		Column5 ColumnPatterns `yaml:"column_5"`
		Column6 ColumnPatterns `yaml:"column_6"`
	} `yaml:"historical_columns"`

	DateHeaders string `yaml:"date_headers"`
}

type ColumnPatterns struct {
	MarketCap              string `yaml:"market_cap"`
	EnterpriseValue        string `yaml:"enterprise_value"`
	TrailingPE             string `yaml:"trailing_pe"`
	ForwardPE              string `yaml:"forward_pe"`
	PEGRatio               string `yaml:"peg_ratio"`
	PriceSales             string `yaml:"price_sales"`
	PriceBook              string `yaml:"price_book"`
	EnterpriseValueRevenue string `yaml:"enterprise_value_revenue"`
	EnterpriseValueEBITDA  string `yaml:"enterprise_value_ebitda"`
}

// ComprehensiveKeyStatisticsDTO holds all key statistics data
type ComprehensiveKeyStatisticsDTO struct {
	Symbol   string    `json:"symbol"`
	Market   string    `json:"market"`
	Currency string    `json:"currency"`
	AsOf     time.Time `json:"as_of"`

	// Current values (most recent data)
	Current struct {
		MarketCap              *Scaled `json:"market_cap,omitempty"`
		EnterpriseValue        *Scaled `json:"enterprise_value,omitempty"`
		TrailingPE             *Scaled `json:"trailing_pe,omitempty"`
		ForwardPE              *Scaled `json:"forward_pe,omitempty"`
		PEGRatio               *Scaled `json:"peg_ratio,omitempty"`
		PriceSales             *Scaled `json:"price_sales,omitempty"`
		PriceBook              *Scaled `json:"price_book,omitempty"`
		EnterpriseValueRevenue *Scaled `json:"enterprise_value_revenue,omitempty"`
		EnterpriseValueEBITDA  *Scaled `json:"enterprise_value_ebitda,omitempty"`
	} `json:"current"`

	// Additional statistics (from other parts of the page)
	Additional struct {
		Beta              *Scaled `json:"beta,omitempty"`
		SharesOutstanding *int64  `json:"shares_outstanding,omitempty"`
		ProfitMargin      *Scaled `json:"profit_margin,omitempty"`
		OperatingMargin   *Scaled `json:"operating_margin,omitempty"`
		ReturnOnAssets    *Scaled `json:"return_on_assets,omitempty"`
		ReturnOnEquity    *Scaled `json:"return_on_equity,omitempty"`
	} `json:"additional"`

	// Historical values - dynamic quarters
	Historical []HistoricalQuarter `json:"historical,omitempty"`
}

type HistoricalQuarter struct {
	Date                   string  `json:"date"`
	MarketCap              *Scaled `json:"market_cap,omitempty"`
	EnterpriseValue        *Scaled `json:"enterprise_value,omitempty"`
	TrailingPE             *Scaled `json:"trailing_pe,omitempty"`
	ForwardPE              *Scaled `json:"forward_pe,omitempty"`
	PEGRatio               *Scaled `json:"peg_ratio,omitempty"`
	PriceSales             *Scaled `json:"price_sales,omitempty"`
	PriceBook              *Scaled `json:"price_book,omitempty"`
	EnterpriseValueRevenue *Scaled `json:"enterprise_value_revenue,omitempty"`
	EnterpriseValueEBITDA  *Scaled `json:"enterprise_value_ebitda,omitempty"`
}

var regexConfig *RegexConfig

// LoadRegexConfig loads the regex patterns from YAML file
func LoadRegexConfig() error {
	if regexConfig != nil {
		return nil // Already loaded
	}

	// Get the directory of the current file
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return fmt.Errorf("unable to get current file path")
	}

	configPath := filepath.Join(filepath.Dir(filename), "regex", "statistics.yaml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read regex config file: %w", err)
	}

	regexConfig = &RegexConfig{}
	if err := yaml.Unmarshal(data, regexConfig); err != nil {
		return fmt.Errorf("failed to parse regex config YAML: %w", err)
	}

	return nil
}

// ParseComprehensiveKeyStatistics extracts comprehensive key statistics data from HTML
func ParseComprehensiveKeyStatistics(html []byte, symbol, market string) (*ComprehensiveKeyStatisticsDTO, error) {
	if err := LoadRegexConfig(); err != nil {
		return nil, fmt.Errorf("failed to load regex config: %w", err)
	}

	dto := &ComprehensiveKeyStatisticsDTO{
		Symbol:   symbol,
		Market:   market,
		Currency: "USD", // Default, will be updated from actual data
		AsOf:     time.Now().UTC(),
	}

	htmlStr := string(html)

	// Extract current values
	extractCurrentValues(htmlStr, dto)

	// Extract additional statistics
	extractAdditionalValues(htmlStr, dto)

	// Extract historical values dynamically
	extractHistoricalValues(htmlStr, dto)

	return dto, nil
}

// extractCurrentValues extracts current (most recent) statistics values
func extractCurrentValues(html string, dto *ComprehensiveKeyStatisticsDTO) {
	dto.Current.MarketCap = extractScaledValue(html, regexConfig.Current.MarketCap)
	dto.Current.EnterpriseValue = extractScaledValue(html, regexConfig.Current.EnterpriseValue)
	dto.Current.TrailingPE = extractScaledValue(html, regexConfig.Current.TrailingPE)
	dto.Current.ForwardPE = extractScaledValue(html, regexConfig.Current.ForwardPE)
	dto.Current.PEGRatio = extractScaledValue(html, regexConfig.Current.PEGRatio)
	dto.Current.PriceSales = extractScaledValue(html, regexConfig.Current.PriceSales)
	dto.Current.PriceBook = extractScaledValue(html, regexConfig.Current.PriceBook)
	dto.Current.EnterpriseValueRevenue = extractScaledValue(html, regexConfig.Current.EnterpriseValueRevenue)
	dto.Current.EnterpriseValueEBITDA = extractScaledValue(html, regexConfig.Current.EnterpriseValueEBITDA)
}

// extractAdditionalValues extracts additional statistics from other parts of the page
func extractAdditionalValues(html string, dto *ComprehensiveKeyStatisticsDTO) {
	dto.Additional.Beta = extractScaledValue(html, regexConfig.Additional.Beta)
	dto.Additional.ProfitMargin = extractScaledValue(html, regexConfig.Additional.ProfitMargin)
	dto.Additional.OperatingMargin = extractScaledValue(html, regexConfig.Additional.OperatingMargin)
	dto.Additional.ReturnOnAssets = extractScaledValue(html, regexConfig.Additional.ReturnOnAssets)
	dto.Additional.ReturnOnEquity = extractScaledValue(html, regexConfig.Additional.ReturnOnEquity)

	// Shares Outstanding needs special handling since it's an integer, not a scaled value
	if sharesStr := extractStringValue(html, regexConfig.Additional.SharesOutstanding); sharesStr != "" {
		dto.Additional.SharesOutstanding = parseSharesOutstanding(sharesStr)
	}
}

// extractHistoricalValues extracts historical values dynamically
func extractHistoricalValues(html string, dto *ComprehensiveKeyStatisticsDTO) {
	// Extract dates from table headers
	dates := extractDates(html)

	// Define column patterns in order
	columnPatterns := []ColumnPatterns{
		regexConfig.HistoricalColumns.Column2,
		regexConfig.HistoricalColumns.Column3,
		regexConfig.HistoricalColumns.Column4,
		regexConfig.HistoricalColumns.Column5,
		regexConfig.HistoricalColumns.Column6,
	}

	// Extract data for each available column/date
	for i, patterns := range columnPatterns {
		var date string
		if i < len(dates) {
			date = dates[i]
		} else {
			// If we don't have a date, skip this column
			continue
		}

		quarter := HistoricalQuarter{
			Date:                   date,
			MarketCap:              extractScaledValue(html, patterns.MarketCap),
			EnterpriseValue:        extractScaledValue(html, patterns.EnterpriseValue),
			TrailingPE:             extractScaledValue(html, patterns.TrailingPE),
			ForwardPE:              extractScaledValue(html, patterns.ForwardPE),
			PEGRatio:               extractScaledValue(html, patterns.PEGRatio),
			PriceSales:             extractScaledValue(html, patterns.PriceSales),
			PriceBook:              extractScaledValue(html, patterns.PriceBook),
			EnterpriseValueRevenue: extractScaledValue(html, patterns.EnterpriseValueRevenue),
			EnterpriseValueEBITDA:  extractScaledValue(html, patterns.EnterpriseValueEBITDA),
		}

		dto.Historical = append(dto.Historical, quarter)
	}
}

// extractDates extracts dates from table headers dynamically
func extractDates(html string) []string {
	re := regexp.MustCompile(regexConfig.DateHeaders)
	matches := re.FindAllStringSubmatch(html, -1)

	var dates []string
	for _, match := range matches {
		if len(match) > 1 {
			// Convert MM/DD/YYYY to YYYY-MM-DD format
			dateStr := match[1]
			if parsedDate, err := time.Parse("1/2/2006", dateStr); err == nil {
				dates = append(dates, parsedDate.Format("2006-01-02"))
			} else if parsedDate, err := time.Parse("01/02/2006", dateStr); err == nil {
				dates = append(dates, parsedDate.Format("2006-01-02"))
			} else {
				// If parsing fails, use original format
				dates = append(dates, dateStr)
			}
		}
	}

	return dates
}

// extractScaledValue extracts and converts a value using the given regex pattern
func extractScaledValue(html, pattern string) *Scaled {
	if pattern == "" {
		return nil
	}

	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(html)

	if len(matches) > 1 {
		value := strings.TrimSpace(matches[1])
		return parseFinancialValue(value)
	}

	return nil
}

// parseFinancialValue converts a financial string value to Scaled format
func parseFinancialValue(value string) *Scaled {
	if value == "" || value == "--" || value == "N/A" {
		return nil
	}

	// Remove any currency symbols and commas
	cleanValue := strings.ReplaceAll(value, ",", "")
	cleanValue = strings.ReplaceAll(cleanValue, "$", "")
	cleanValue = strings.TrimSpace(cleanValue)

	// Handle percentage values
	if strings.HasSuffix(cleanValue, "%") {
		cleanValue = strings.TrimSuffix(cleanValue, "%")
		if val, err := strconv.ParseFloat(cleanValue, 64); err == nil {
			// Convert percentage to basis points (multiply by 100)
			return &Scaled{Scaled: int64(val * 100), Scale: 2}
		}
	}

	// Handle suffixed values (B, M, K)
	var multiplier int64 = 1
	if strings.HasSuffix(cleanValue, "B") {
		multiplier = 1000000000 // Billion
		cleanValue = strings.TrimSuffix(cleanValue, "B")
	} else if strings.HasSuffix(cleanValue, "M") {
		multiplier = 1000000 // Million
		cleanValue = strings.TrimSuffix(cleanValue, "M")
	} else if strings.HasSuffix(cleanValue, "K") {
		multiplier = 1000 // Thousand
		cleanValue = strings.TrimSuffix(cleanValue, "K")
	}

	// Parse the numeric value
	if val, err := strconv.ParseFloat(cleanValue, 64); err == nil {
		// Use scale 2 for ratios and percentages, scale 0 for large numbers
		scale := 2
		if multiplier > 1 {
			scale = 0 // Large numbers don't need decimal precision
		}
		return &Scaled{Scaled: int64(val * float64(multiplier) * 100), Scale: scale}
	}

	return nil
}

// extractStringValue extracts a string value using the given regex pattern
func extractStringValue(html, pattern string) string {
	if pattern == "" {
		return ""
	}

	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(html)

	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	return ""
}

// parseSharesOutstanding converts a shares outstanding string to integer
func parseSharesOutstanding(value string) *int64 {
	if value == "" || value == "--" || value == "N/A" {
		return nil
	}

	// Remove any commas and spaces
	cleanValue := strings.ReplaceAll(value, ",", "")
	cleanValue = strings.TrimSpace(cleanValue)

	// Handle suffixed values (B, M, K)
	var multiplier int64 = 1
	if strings.HasSuffix(cleanValue, "B") {
		multiplier = 1000000000 // Billion
		cleanValue = strings.TrimSuffix(cleanValue, "B")
	} else if strings.HasSuffix(cleanValue, "M") {
		multiplier = 1000000 // Million
		cleanValue = strings.TrimSuffix(cleanValue, "M")
	} else if strings.HasSuffix(cleanValue, "K") {
		multiplier = 1000 // Thousand
		cleanValue = strings.TrimSuffix(cleanValue, "K")
	}

	// Parse the numeric value
	if val, err := strconv.ParseFloat(cleanValue, 64); err == nil {
		result := int64(val * float64(multiplier))
		return &result
	}

	return nil
}
