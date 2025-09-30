package scrape

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"time"

	"gopkg.in/yaml.v3"
)

// AnalystInsightsDTO represents analyst insights data from Yahoo Finance
type AnalystInsightsDTO struct {
	Symbol string    `json:"symbol"`
	Market string    `json:"market"`
	AsOf   time.Time `json:"as_of"`

	// Price Targets
	CurrentPrice      *float64 `json:"current_price,omitempty"`
	TargetMeanPrice   *float64 `json:"target_mean_price,omitempty"`
	TargetMedianPrice *float64 `json:"target_median_price,omitempty"`
	TargetHighPrice   *float64 `json:"target_high_price,omitempty"`
	TargetLowPrice    *float64 `json:"target_low_price,omitempty"`

	// Analyst Opinions
	NumberOfAnalysts   *int     `json:"number_of_analysts,omitempty"`
	RecommendationMean *float64 `json:"recommendation_mean,omitempty"`
	RecommendationKey  *string  `json:"recommendation_key,omitempty"`
}

// AnalystInsightsRegexConfig holds the regex patterns for analyst insights extraction
type AnalystInsightsRegexConfig struct {
	FinancialData struct {
		CombinedPattern string `yaml:"combined_pattern"`
	} `yaml:"financial_data"`

	IndividualFields struct {
		CurrentPrice       string `yaml:"current_price"`
		TargetMeanPrice    string `yaml:"target_mean_price"`
		TargetMedianPrice  string `yaml:"target_median_price"`
		TargetHighPrice    string `yaml:"target_high_price"`
		TargetLowPrice     string `yaml:"target_low_price"`
		RecommendationMean string `yaml:"recommendation_mean"`
		RecommendationKey  string `yaml:"recommendation_key"`
		NumberOfAnalysts   string `yaml:"number_of_analysts"`
	} `yaml:"individual_fields"`
}

var analystInsightsRegexConfig *AnalystInsightsRegexConfig

// LoadAnalystInsightsRegexConfig loads the regex patterns from YAML file
func LoadAnalystInsightsRegexConfig() error {
	if analystInsightsRegexConfig != nil {
		return nil // Already loaded
	}

	// Get the directory of the current file
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return fmt.Errorf("unable to get current file path")
	}

	configPath := filepath.Join(filepath.Dir(filename), "regex", "analyst_insights.yaml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read analyst insights regex config file: %w", err)
	}

	analystInsightsRegexConfig = &AnalystInsightsRegexConfig{}
	if err := yaml.Unmarshal(data, analystInsightsRegexConfig); err != nil {
		return fmt.Errorf("failed to parse analyst insights regex config YAML: %w", err)
	}

	return nil
}

// ParseAnalystInsights parses analyst insights data from Yahoo Finance HTML
func ParseAnalystInsights(html []byte, symbol, market string) (*AnalystInsightsDTO, error) {
	if err := LoadAnalystInsightsRegexConfig(); err != nil {
		return nil, fmt.Errorf("failed to load analyst insights regex config: %w", err)
	}

	dto := &AnalystInsightsDTO{
		Symbol: symbol,
		Market: market,
		AsOf:   time.Now(),
	}

	htmlStr := string(html)

	// Extract financial data from embedded JSON
	if err := extractFinancialDataFromJSON(htmlStr, dto); err != nil {
		return nil, fmt.Errorf("failed to extract financial data: %w", err)
	}

	return dto, nil
}

// extractFinancialDataFromJSON extracts analyst insights from embedded JSON data
func extractFinancialDataFromJSON(html string, dto *AnalystInsightsDTO) error {
	// Find the financialData section in the embedded JSON
	// Pattern to match: "financialData":{"maxAge":86400,"currentPrice":...}
	re := regexp.MustCompile(analystInsightsRegexConfig.FinancialData.CombinedPattern)
	matches := re.FindStringSubmatch(html)

	if len(matches) < 9 {
		// Try a more flexible approach by searching for individual fields
		return extractFinancialDataFlexible(html, dto)
	}

	// Parse the extracted values
	if currentPrice := parseFloat(matches[1]); currentPrice != nil {
		dto.CurrentPrice = currentPrice
	}
	if targetMeanPrice := parseFloat(matches[2]); targetMeanPrice != nil {
		dto.TargetMeanPrice = targetMeanPrice
	}
	if targetMedianPrice := parseFloat(matches[3]); targetMedianPrice != nil {
		dto.TargetMedianPrice = targetMedianPrice
	}
	if targetHighPrice := parseFloat(matches[4]); targetHighPrice != nil {
		dto.TargetHighPrice = targetHighPrice
	}
	if targetLowPrice := parseFloat(matches[5]); targetLowPrice != nil {
		dto.TargetLowPrice = targetLowPrice
	}
	if recommendationMean := parseFloat(matches[6]); recommendationMean != nil {
		dto.RecommendationMean = recommendationMean
	}
	if recommendationKey := matches[7]; recommendationKey != "" {
		dto.RecommendationKey = &recommendationKey
	}
	if numberOfAnalysts := parseInt(matches[8]); numberOfAnalysts != nil {
		dto.NumberOfAnalysts = numberOfAnalysts
	}

	return nil
}

// extractFinancialDataFlexible tries to extract fields individually if the combined pattern fails
func extractFinancialDataFlexible(html string, dto *AnalystInsightsDTO) error {
	// Extract individual fields with more flexible patterns (handle escaped quotes and actual format)
	patterns := map[string]string{
		"currentPrice":       analystInsightsRegexConfig.IndividualFields.CurrentPrice,
		"targetMeanPrice":    analystInsightsRegexConfig.IndividualFields.TargetMeanPrice,
		"targetMedianPrice":  analystInsightsRegexConfig.IndividualFields.TargetMedianPrice,
		"targetHighPrice":    analystInsightsRegexConfig.IndividualFields.TargetHighPrice,
		"targetLowPrice":     analystInsightsRegexConfig.IndividualFields.TargetLowPrice,
		"recommendationMean": analystInsightsRegexConfig.IndividualFields.RecommendationMean,
		"recommendationKey":  analystInsightsRegexConfig.IndividualFields.RecommendationKey,
		"numberOfAnalysts":   analystInsightsRegexConfig.IndividualFields.NumberOfAnalysts,
	}

	for field, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(html)
		if len(matches) > 1 {
			value := matches[1]
			switch field {
			case "currentPrice":
				if parsed := parseFloat(value); parsed != nil {
					dto.CurrentPrice = parsed
				}
			case "targetMeanPrice":
				if parsed := parseFloat(value); parsed != nil {
					dto.TargetMeanPrice = parsed
				}
			case "targetMedianPrice":
				if parsed := parseFloat(value); parsed != nil {
					dto.TargetMedianPrice = parsed
				}
			case "targetHighPrice":
				if parsed := parseFloat(value); parsed != nil {
					dto.TargetHighPrice = parsed
				}
			case "targetLowPrice":
				if parsed := parseFloat(value); parsed != nil {
					dto.TargetLowPrice = parsed
				}
			case "recommendationMean":
				if parsed := parseFloat(value); parsed != nil {
					dto.RecommendationMean = parsed
				}
			case "recommendationKey":
				if value != "" {
					dto.RecommendationKey = &value
				}
			case "numberOfAnalysts":
				if parsed := parseInt(value); parsed != nil {
					dto.NumberOfAnalysts = parsed
				}
			}
		}
	}

	return nil
}
