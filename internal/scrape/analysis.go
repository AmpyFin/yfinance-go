package scrape

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// ComprehensiveAnalysisDTO represents comprehensive analysis data from Yahoo Finance
type ComprehensiveAnalysisDTO struct {
	Symbol   string    `json:"symbol"`
	Market   string    `json:"market"`
	AsOf     time.Time `json:"as_of"`

	// Earnings Estimate
	EarningsEstimate struct {
		Currency string `json:"currency"`
		CurrentQtr struct {
			NoOfAnalysts   *int     `json:"no_of_analysts,omitempty"`
			AvgEstimate    *float64 `json:"avg_estimate,omitempty"`
			LowEstimate    *float64 `json:"low_estimate,omitempty"`
			HighEstimate   *float64 `json:"high_estimate,omitempty"`
			YearAgoEPS     *float64 `json:"year_ago_eps,omitempty"`
		} `json:"current_qtr"`
		NextQtr struct {
			NoOfAnalysts   *int     `json:"no_of_analysts,omitempty"`
			AvgEstimate    *float64 `json:"avg_estimate,omitempty"`
			LowEstimate    *float64 `json:"low_estimate,omitempty"`
			HighEstimate   *float64 `json:"high_estimate,omitempty"`
			YearAgoEPS     *float64 `json:"year_ago_eps,omitempty"`
		} `json:"next_qtr"`
		CurrentYear struct {
			NoOfAnalysts   *int     `json:"no_of_analysts,omitempty"`
			AvgEstimate    *float64 `json:"avg_estimate,omitempty"`
			LowEstimate    *float64 `json:"low_estimate,omitempty"`
			HighEstimate   *float64 `json:"high_estimate,omitempty"`
			YearAgoEPS     *float64 `json:"year_ago_eps,omitempty"`
		} `json:"current_year"`
		NextYear struct {
			NoOfAnalysts   *int     `json:"no_of_analysts,omitempty"`
			AvgEstimate    *float64 `json:"avg_estimate,omitempty"`
			LowEstimate    *float64 `json:"low_estimate,omitempty"`
			HighEstimate   *float64 `json:"high_estimate,omitempty"`
			YearAgoEPS     *float64 `json:"year_ago_eps,omitempty"`
		} `json:"next_year"`
	} `json:"earnings_estimate"`

	// Revenue Estimate
	RevenueEstimate struct {
		Currency string `json:"currency"`
		CurrentQtr struct {
			NoOfAnalysts       *int     `json:"no_of_analysts,omitempty"`
			AvgEstimate        *string  `json:"avg_estimate,omitempty"`        // Keep as string due to "B" suffix
			LowEstimate        *string  `json:"low_estimate,omitempty"`
			HighEstimate       *string  `json:"high_estimate,omitempty"`
			YearAgoSales       *string  `json:"year_ago_sales,omitempty"`
			SalesGrowthYearEst *string  `json:"sales_growth_year_est,omitempty"`
		} `json:"current_qtr"`
		NextQtr struct {
			NoOfAnalysts       *int     `json:"no_of_analysts,omitempty"`
			AvgEstimate        *string  `json:"avg_estimate,omitempty"`
			LowEstimate        *string  `json:"low_estimate,omitempty"`
			HighEstimate       *string  `json:"high_estimate,omitempty"`
			YearAgoSales       *string  `json:"year_ago_sales,omitempty"`
			SalesGrowthYearEst *string  `json:"sales_growth_year_est,omitempty"`
		} `json:"next_qtr"`
		CurrentYear struct {
			NoOfAnalysts       *int     `json:"no_of_analysts,omitempty"`
			AvgEstimate        *string  `json:"avg_estimate,omitempty"`
			LowEstimate        *string  `json:"low_estimate,omitempty"`
			HighEstimate       *string  `json:"high_estimate,omitempty"`
			YearAgoSales       *string  `json:"year_ago_sales,omitempty"`
			SalesGrowthYearEst *string  `json:"sales_growth_year_est,omitempty"`
		} `json:"current_year"`
		NextYear struct {
			NoOfAnalysts       *int     `json:"no_of_analysts,omitempty"`
			AvgEstimate        *string  `json:"avg_estimate,omitempty"`
			LowEstimate        *string  `json:"low_estimate,omitempty"`
			HighEstimate       *string  `json:"high_estimate,omitempty"`
			YearAgoSales       *string  `json:"year_ago_sales,omitempty"`
			SalesGrowthYearEst *string  `json:"sales_growth_year_est,omitempty"`
		} `json:"next_year"`
	} `json:"revenue_estimate"`

	// Earnings History (dynamic dates)
	EarningsHistory struct {
		Currency string `json:"currency"`
		Data     []struct {
		Date        string   `json:"date"`
		EPSEst      *float64 `json:"eps_est,omitempty"`
		EPSActual   *float64 `json:"eps_actual,omitempty"`
		Difference  *float64 `json:"difference,omitempty"`
		SurprisePercent *string `json:"surprise_percent,omitempty"`
		} `json:"data"`
	} `json:"earnings_history"`

	// EPS Trend
	EPSTrend struct {
		Currency string `json:"currency"`
		CurrentQtr struct {
			CurrentEstimate *float64 `json:"current_estimate,omitempty"`
			Days7Ago        *float64 `json:"days_7_ago,omitempty"`
			Days30Ago       *float64 `json:"days_30_ago,omitempty"`
			Days60Ago       *float64 `json:"days_60_ago,omitempty"`
			Days90Ago       *float64 `json:"days_90_ago,omitempty"`
		} `json:"current_qtr"`
		NextQtr struct {
			CurrentEstimate *float64 `json:"current_estimate,omitempty"`
			Days7Ago        *float64 `json:"days_7_ago,omitempty"`
			Days30Ago       *float64 `json:"days_30_ago,omitempty"`
			Days60Ago       *float64 `json:"days_60_ago,omitempty"`
			Days90Ago       *float64 `json:"days_90_ago,omitempty"`
		} `json:"next_qtr"`
		CurrentYear struct {
			CurrentEstimate *float64 `json:"current_estimate,omitempty"`
			Days7Ago        *float64 `json:"days_7_ago,omitempty"`
			Days30Ago       *float64 `json:"days_30_ago,omitempty"`
			Days60Ago       *float64 `json:"days_60_ago,omitempty"`
			Days90Ago       *float64 `json:"days_90_ago,omitempty"`
		} `json:"current_year"`
		NextYear struct {
			CurrentEstimate *float64 `json:"current_estimate,omitempty"`
			Days7Ago        *float64 `json:"days_7_ago,omitempty"`
			Days30Ago       *float64 `json:"days_30_ago,omitempty"`
			Days60Ago       *float64 `json:"days_60_ago,omitempty"`
			Days90Ago       *float64 `json:"days_90_ago,omitempty"`
		} `json:"next_year"`
	} `json:"eps_trend"`

	// EPS Revisions
	EPSRevisions struct {
		Currency string `json:"currency"`
		CurrentQtr struct {
			UpLast7Days    *int `json:"up_last_7_days,omitempty"`
			UpLast30Days   *int `json:"up_last_30_days,omitempty"`
			DownLast7Days  *int `json:"down_last_7_days,omitempty"`
			DownLast30Days *int `json:"down_last_30_days,omitempty"`
		} `json:"current_qtr"`
		NextQtr struct {
			UpLast7Days    *int `json:"up_last_7_days,omitempty"`
			UpLast30Days   *int `json:"up_last_30_days,omitempty"`
			DownLast7Days  *int `json:"down_last_7_days,omitempty"`
			DownLast30Days *int `json:"down_last_30_days,omitempty"`
		} `json:"next_qtr"`
		CurrentYear struct {
			UpLast7Days    *int `json:"up_last_7_days,omitempty"`
			UpLast30Days   *int `json:"up_last_30_days,omitempty"`
			DownLast7Days  *int `json:"down_last_7_days,omitempty"`
			DownLast30Days *int `json:"down_last_30_days,omitempty"`
		} `json:"current_year"`
		NextYear struct {
			UpLast7Days    *int `json:"up_last_7_days,omitempty"`
			UpLast30Days   *int `json:"up_last_30_days,omitempty"`
			DownLast7Days  *int `json:"down_last_7_days,omitempty"`
			DownLast30Days *int `json:"down_last_30_days,omitempty"`
		} `json:"next_year"`
	} `json:"eps_revisions"`

	// Growth Estimates (only ticker data, not S&P 500)
	GrowthEstimate struct {
		CurrentQtr  *string `json:"current_qtr,omitempty"`
		NextQtr     *string `json:"next_qtr,omitempty"`
		CurrentYear *string `json:"current_year,omitempty"`
		NextYear    *string `json:"next_year,omitempty"`
	} `json:"growth_estimate"`
}

// AnalysisRegexConfig holds the regex patterns for analysis extraction
type AnalysisRegexConfig struct {
	EarningsEstimate struct {
		SectionPattern  string `yaml:"section_pattern"`
		CurrencyPattern string `yaml:"currency_pattern"`
		TableRowPattern string `yaml:"table_row_pattern"`
	} `yaml:"earnings_estimate"`
	
	RevenueEstimate struct {
		SectionPattern  string `yaml:"section_pattern"`
		CurrencyPattern string `yaml:"currency_pattern"`
		TableRowPattern string `yaml:"table_row_pattern"`
	} `yaml:"revenue_estimate"`
	
	EarningsHistory struct {
		SectionPattern   string `yaml:"section_pattern"`
		CurrencyPattern  string `yaml:"currency_pattern"`
		HeaderPattern    string `yaml:"header_pattern"`
		TableRowPattern  string `yaml:"table_row_pattern"`
		TableCellPattern string `yaml:"table_cell_pattern"`
	} `yaml:"earnings_history"`
	
	EPSTrend struct {
		SectionPattern  string `yaml:"section_pattern"`
		CurrencyPattern string `yaml:"currency_pattern"`
		TableRowPattern string `yaml:"table_row_pattern"`
	} `yaml:"eps_trend"`
	
	EPSRevisions struct {
		SectionPattern  string `yaml:"section_pattern"`
		CurrencyPattern string `yaml:"currency_pattern"`
		TableRowPattern string `yaml:"table_row_pattern"`
	} `yaml:"eps_revisions"`
	
	GrowthEstimate struct {
		SectionPattern  string `yaml:"section_pattern"`
		TableRowPattern string `yaml:"table_row_pattern"`
	} `yaml:"growth_estimate"`
}

var analysisRegexConfig *AnalysisRegexConfig

// LoadAnalysisRegexConfig loads the regex patterns from YAML file
func LoadAnalysisRegexConfig() error {
	if analysisRegexConfig != nil {
		return nil // Already loaded
	}
	
	// Get the directory of the current file
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return fmt.Errorf("unable to get current file path")
	}
	
	configPath := filepath.Join(filepath.Dir(filename), "regex", "analysis.yaml")
	
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read analysis regex config file: %w", err)
	}
	
	analysisRegexConfig = &AnalysisRegexConfig{}
	if err := yaml.Unmarshal(data, analysisRegexConfig); err != nil {
		return fmt.Errorf("failed to parse analysis regex config YAML: %w", err)
	}
	
	return nil
}

// ParseAnalysis parses analysis data from Yahoo Finance HTML
func ParseAnalysis(html []byte, symbol, market string) (*ComprehensiveAnalysisDTO, error) {
	if err := LoadAnalysisRegexConfig(); err != nil {
		return nil, fmt.Errorf("failed to load analysis regex config: %w", err)
	}
	
	dto := &ComprehensiveAnalysisDTO{
		Symbol: symbol,
		Market: market,
		AsOf:   time.Now(),
	}

	htmlStr := string(html)

	// Extract analysis data from HTML tables
	if err := extractEarningsEstimate(htmlStr, dto); err != nil {
		return nil, fmt.Errorf("failed to extract earnings estimate: %w", err)
	}

	if err := extractRevenueEstimate(htmlStr, dto); err != nil {
		return nil, fmt.Errorf("failed to extract revenue estimate: %w", err)
	}

	if err := extractEarningsHistory(htmlStr, dto); err != nil {
		return nil, fmt.Errorf("failed to extract earnings history: %w", err)
	}

	if err := extractEPSTrend(htmlStr, dto); err != nil {
		return nil, fmt.Errorf("failed to extract EPS trend: %w", err)
	}

	if err := extractEPSRevisions(htmlStr, dto); err != nil {
		return nil, fmt.Errorf("failed to extract EPS revisions: %w", err)
	}

	if err := extractGrowthEstimate(htmlStr, dto); err != nil {
		return nil, fmt.Errorf("failed to extract growth estimate: %w", err)
	}

	return dto, nil
}

// Helper function to parse float from string, handling "--" and empty values
func parseFloat(s string) *float64 {
	s = strings.TrimSpace(s)
	if s == "" || s == "--" || s == "-" {
		return nil
	}
	if val, err := strconv.ParseFloat(s, 64); err == nil {
		return &val
	}
	return nil
}

// Helper function to parse int from string, handling "--" and empty values
func parseInt(s string) *int {
	s = strings.TrimSpace(s)
	if s == "" || s == "--" || s == "-" {
		return nil
	}
	if val, err := strconv.Atoi(s); err == nil {
		return &val
	}
	return nil
}

// Helper function to parse string, handling "--" and empty values
func parseString(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" || s == "--" || s == "-" {
		return nil
	}
	return &s
}

// extractEarningsEstimate extracts earnings estimate data from HTML
func extractEarningsEstimate(html string, dto *ComprehensiveAnalysisDTO) error {
	// Find the earnings estimate table section
	re := regexp.MustCompile(analysisRegexConfig.EarningsEstimate.SectionPattern)
	match := re.FindString(html)
	if match == "" {
		return fmt.Errorf("earnings estimate section not found")
	}

	// Extract currency from table header
	re = regexp.MustCompile(analysisRegexConfig.EarningsEstimate.CurrencyPattern)
	currencyMatch := re.FindStringSubmatch(match)
	if len(currencyMatch) > 1 {
		dto.EarningsEstimate.Currency = currencyMatch[1]
	} else {
		dto.EarningsEstimate.Currency = "USD" // Default fallback
	}

	// Extract table rows - we know the order: No. of Analysts, Avg. Estimate, Low Estimate, High Estimate, Year Ago EPS
	re = regexp.MustCompile(analysisRegexConfig.EarningsEstimate.TableRowPattern)
	matches := re.FindAllStringSubmatch(match, -1)

	for _, rowMatch := range matches {
		if len(rowMatch) < 6 {
			continue
		}
		
		rowTitle := strings.TrimSpace(rowMatch[1])
		currentQtr := strings.TrimSpace(rowMatch[2])
		nextQtr := strings.TrimSpace(rowMatch[3])
		currentYear := strings.TrimSpace(rowMatch[4])
		nextYear := strings.TrimSpace(rowMatch[5])

		switch rowTitle {
		case "No. of Analysts":
			dto.EarningsEstimate.CurrentQtr.NoOfAnalysts = parseInt(currentQtr)
			dto.EarningsEstimate.NextQtr.NoOfAnalysts = parseInt(nextQtr)
			dto.EarningsEstimate.CurrentYear.NoOfAnalysts = parseInt(currentYear)
			dto.EarningsEstimate.NextYear.NoOfAnalysts = parseInt(nextYear)
		case "Avg. Estimate":
			dto.EarningsEstimate.CurrentQtr.AvgEstimate = parseFloat(currentQtr)
			dto.EarningsEstimate.NextQtr.AvgEstimate = parseFloat(nextQtr)
			dto.EarningsEstimate.CurrentYear.AvgEstimate = parseFloat(currentYear)
			dto.EarningsEstimate.NextYear.AvgEstimate = parseFloat(nextYear)
		case "Low Estimate":
			dto.EarningsEstimate.CurrentQtr.LowEstimate = parseFloat(currentQtr)
			dto.EarningsEstimate.NextQtr.LowEstimate = parseFloat(nextQtr)
			dto.EarningsEstimate.CurrentYear.LowEstimate = parseFloat(currentYear)
			dto.EarningsEstimate.NextYear.LowEstimate = parseFloat(nextYear)
		case "High Estimate":
			dto.EarningsEstimate.CurrentQtr.HighEstimate = parseFloat(currentQtr)
			dto.EarningsEstimate.NextQtr.HighEstimate = parseFloat(nextQtr)
			dto.EarningsEstimate.CurrentYear.HighEstimate = parseFloat(currentYear)
			dto.EarningsEstimate.NextYear.HighEstimate = parseFloat(nextYear)
		case "Year Ago EPS":
			dto.EarningsEstimate.CurrentQtr.YearAgoEPS = parseFloat(currentQtr)
			dto.EarningsEstimate.NextQtr.YearAgoEPS = parseFloat(nextQtr)
			dto.EarningsEstimate.CurrentYear.YearAgoEPS = parseFloat(currentYear)
			dto.EarningsEstimate.NextYear.YearAgoEPS = parseFloat(nextYear)
		}
	}

	return nil
}

// extractRevenueEstimate extracts revenue estimate data from HTML
func extractRevenueEstimate(html string, dto *ComprehensiveAnalysisDTO) error {
	// Find the revenue estimate table section
	re := regexp.MustCompile(analysisRegexConfig.RevenueEstimate.SectionPattern)
	match := re.FindString(html)
	if match == "" {
		return fmt.Errorf("revenue estimate section not found")
	}

	// Extract currency from table header
	re = regexp.MustCompile(analysisRegexConfig.RevenueEstimate.CurrencyPattern)
	currencyMatch := re.FindStringSubmatch(match)
	if len(currencyMatch) > 1 {
		dto.RevenueEstimate.Currency = currencyMatch[1]
	} else {
		dto.RevenueEstimate.Currency = "USD" // Default fallback
	}

	// Extract table rows
	re = regexp.MustCompile(analysisRegexConfig.RevenueEstimate.TableRowPattern)
	matches := re.FindAllStringSubmatch(match, -1)

	for _, rowMatch := range matches {
		if len(rowMatch) < 6 {
			continue
		}
		
		rowTitle := strings.TrimSpace(rowMatch[1])
		currentQtr := strings.TrimSpace(rowMatch[2])
		nextQtr := strings.TrimSpace(rowMatch[3])
		currentYear := strings.TrimSpace(rowMatch[4])
		nextYear := strings.TrimSpace(rowMatch[5])

		switch rowTitle {
		case "No. of Analysts":
			dto.RevenueEstimate.CurrentQtr.NoOfAnalysts = parseInt(currentQtr)
			dto.RevenueEstimate.NextQtr.NoOfAnalysts = parseInt(nextQtr)
			dto.RevenueEstimate.CurrentYear.NoOfAnalysts = parseInt(currentYear)
			dto.RevenueEstimate.NextYear.NoOfAnalysts = parseInt(nextYear)
		case "Avg. Estimate":
			dto.RevenueEstimate.CurrentQtr.AvgEstimate = parseString(currentQtr)
			dto.RevenueEstimate.NextQtr.AvgEstimate = parseString(nextQtr)
			dto.RevenueEstimate.CurrentYear.AvgEstimate = parseString(currentYear)
			dto.RevenueEstimate.NextYear.AvgEstimate = parseString(nextYear)
		case "Low Estimate":
			dto.RevenueEstimate.CurrentQtr.LowEstimate = parseString(currentQtr)
			dto.RevenueEstimate.NextQtr.LowEstimate = parseString(nextQtr)
			dto.RevenueEstimate.CurrentYear.LowEstimate = parseString(currentYear)
			dto.RevenueEstimate.NextYear.LowEstimate = parseString(nextYear)
		case "High Estimate":
			dto.RevenueEstimate.CurrentQtr.HighEstimate = parseString(currentQtr)
			dto.RevenueEstimate.NextQtr.HighEstimate = parseString(nextQtr)
			dto.RevenueEstimate.CurrentYear.HighEstimate = parseString(currentYear)
			dto.RevenueEstimate.NextYear.HighEstimate = parseString(nextYear)
		case "Year Ago Sales":
			dto.RevenueEstimate.CurrentQtr.YearAgoSales = parseString(currentQtr)
			dto.RevenueEstimate.NextQtr.YearAgoSales = parseString(nextQtr)
			dto.RevenueEstimate.CurrentYear.YearAgoSales = parseString(currentYear)
			dto.RevenueEstimate.NextYear.YearAgoSales = parseString(nextYear)
		case "Sales Growth (year/est)":
			dto.RevenueEstimate.CurrentQtr.SalesGrowthYearEst = parseString(currentQtr)
			dto.RevenueEstimate.NextQtr.SalesGrowthYearEst = parseString(nextQtr)
			dto.RevenueEstimate.CurrentYear.SalesGrowthYearEst = parseString(currentYear)
			dto.RevenueEstimate.NextYear.SalesGrowthYearEst = parseString(nextYear)
		}
	}

	return nil
}

// extractEarningsHistory extracts earnings history data with dynamic dates from HTML
func extractEarningsHistory(html string, dto *ComprehensiveAnalysisDTO) error {
	// Find the earnings history table section
	re := regexp.MustCompile(analysisRegexConfig.EarningsHistory.SectionPattern)
	match := re.FindString(html)
	if match == "" {
		return fmt.Errorf("earnings history section not found")
	}

	// Extract currency from table header
	re = regexp.MustCompile(analysisRegexConfig.EarningsHistory.CurrencyPattern)
	currencyMatch := re.FindStringSubmatch(match)
	if len(currencyMatch) > 1 {
		dto.EarningsHistory.Currency = currencyMatch[1]
	} else {
		dto.EarningsHistory.Currency = "USD" // Default fallback
	}

	// Extract header to get dates
	re = regexp.MustCompile(analysisRegexConfig.EarningsHistory.HeaderPattern)
	headerMatches := re.FindAllStringSubmatch(match, -1)
	
	var dates []string
	for i, headerMatch := range headerMatches {
		if i == 0 {
			continue // Skip "Currency in USD" header
		}
		if len(headerMatch) >= 2 {
			dates = append(dates, strings.TrimSpace(headerMatch[1]))
		}
	}

	// Extract table rows
	re = regexp.MustCompile(analysisRegexConfig.EarningsHistory.TableRowPattern)
	rowMatches := re.FindAllStringSubmatch(match, -1)

	// Parse each row (EPS Est., EPS Actual, Difference, Surprise %)
	var epsEstValues, epsActualValues, differenceValues, surpriseValues []string

	for _, rowMatch := range rowMatches {
		if len(rowMatch) < 3 {
			continue
		}
		
		rowTitle := strings.TrimSpace(rowMatch[1])
		rowData := rowMatch[2]
		
		// Extract cell values from this row
		re = regexp.MustCompile(analysisRegexConfig.EarningsHistory.TableCellPattern)
		cellMatches := re.FindAllStringSubmatch(rowData, -1)
		
		var cellValues []string
		for _, cellMatch := range cellMatches {
			if len(cellMatch) >= 2 {
				cellValues = append(cellValues, strings.TrimSpace(cellMatch[1]))
			}
		}

		switch rowTitle {
		case "EPS Est.":
			epsEstValues = cellValues
		case "EPS Actual":
			epsActualValues = cellValues
		case "Difference":
			differenceValues = cellValues
		case "Surprise %":
			surpriseValues = cellValues
		}
	}

	// Create earnings history entries
	for i, date := range dates {
		if i >= len(epsEstValues) {
			break
		}
		
		entry := struct {
			Date            string   `json:"date"`
			EPSEst          *float64 `json:"eps_est,omitempty"`
			EPSActual       *float64 `json:"eps_actual,omitempty"`
			Difference      *float64 `json:"difference,omitempty"`
			SurprisePercent *string  `json:"surprise_percent,omitempty"`
		}{
			Date: date,
		}
		
		if i < len(epsEstValues) {
			entry.EPSEst = parseFloat(epsEstValues[i])
		}
		if i < len(epsActualValues) {
			entry.EPSActual = parseFloat(epsActualValues[i])
		}
		if i < len(differenceValues) {
			entry.Difference = parseFloat(differenceValues[i])
		}
		if i < len(surpriseValues) {
			entry.SurprisePercent = parseString(surpriseValues[i])
		}
		
		dto.EarningsHistory.Data = append(dto.EarningsHistory.Data, entry)
	}

	return nil
}

// extractEPSTrend extracts EPS trend data from HTML
func extractEPSTrend(html string, dto *ComprehensiveAnalysisDTO) error {
	// Find the EPS trend table section
	re := regexp.MustCompile(analysisRegexConfig.EPSTrend.SectionPattern)
	match := re.FindString(html)
	if match == "" {
		return fmt.Errorf("EPS trend section not found")
	}

	// Extract currency from table header
	re = regexp.MustCompile(analysisRegexConfig.EPSTrend.CurrencyPattern)
	currencyMatch := re.FindStringSubmatch(match)
	if len(currencyMatch) > 1 {
		dto.EPSTrend.Currency = currencyMatch[1]
	} else {
		dto.EPSTrend.Currency = "USD" // Default fallback
	}

	// Extract table rows
	re = regexp.MustCompile(analysisRegexConfig.EPSTrend.TableRowPattern)
	matches := re.FindAllStringSubmatch(match, -1)

	for _, rowMatch := range matches {
		if len(rowMatch) < 6 {
			continue
		}
		
		rowTitle := strings.TrimSpace(rowMatch[1])
		currentQtr := strings.TrimSpace(rowMatch[2])
		nextQtr := strings.TrimSpace(rowMatch[3])
		currentYear := strings.TrimSpace(rowMatch[4])
		nextYear := strings.TrimSpace(rowMatch[5])

		switch rowTitle {
		case "Current Estimate":
			dto.EPSTrend.CurrentQtr.CurrentEstimate = parseFloat(currentQtr)
			dto.EPSTrend.NextQtr.CurrentEstimate = parseFloat(nextQtr)
			dto.EPSTrend.CurrentYear.CurrentEstimate = parseFloat(currentYear)
			dto.EPSTrend.NextYear.CurrentEstimate = parseFloat(nextYear)
		case "7 Days Ago":
			dto.EPSTrend.CurrentQtr.Days7Ago = parseFloat(currentQtr)
			dto.EPSTrend.NextQtr.Days7Ago = parseFloat(nextQtr)
			dto.EPSTrend.CurrentYear.Days7Ago = parseFloat(currentYear)
			dto.EPSTrend.NextYear.Days7Ago = parseFloat(nextYear)
		case "30 Days Ago":
			dto.EPSTrend.CurrentQtr.Days30Ago = parseFloat(currentQtr)
			dto.EPSTrend.NextQtr.Days30Ago = parseFloat(nextQtr)
			dto.EPSTrend.CurrentYear.Days30Ago = parseFloat(currentYear)
			dto.EPSTrend.NextYear.Days30Ago = parseFloat(nextYear)
		case "60 Days Ago":
			dto.EPSTrend.CurrentQtr.Days60Ago = parseFloat(currentQtr)
			dto.EPSTrend.NextQtr.Days60Ago = parseFloat(nextQtr)
			dto.EPSTrend.CurrentYear.Days60Ago = parseFloat(currentYear)
			dto.EPSTrend.NextYear.Days60Ago = parseFloat(nextYear)
		case "90 Days Ago":
			dto.EPSTrend.CurrentQtr.Days90Ago = parseFloat(currentQtr)
			dto.EPSTrend.NextQtr.Days90Ago = parseFloat(nextQtr)
			dto.EPSTrend.CurrentYear.Days90Ago = parseFloat(currentYear)
			dto.EPSTrend.NextYear.Days90Ago = parseFloat(nextYear)
		}
	}

	return nil
}

// extractEPSRevisions extracts EPS revisions data from HTML
func extractEPSRevisions(html string, dto *ComprehensiveAnalysisDTO) error {
	// Find the EPS revisions table section
	re := regexp.MustCompile(analysisRegexConfig.EPSRevisions.SectionPattern)
	match := re.FindString(html)
	if match == "" {
		return fmt.Errorf("EPS revisions section not found")
	}

	// Extract currency from table header
	re = regexp.MustCompile(analysisRegexConfig.EPSRevisions.CurrencyPattern)
	currencyMatch := re.FindStringSubmatch(match)
	if len(currencyMatch) > 1 {
		dto.EPSRevisions.Currency = currencyMatch[1]
	} else {
		dto.EPSRevisions.Currency = "USD" // Default fallback
	}

	// Extract table rows
	re = regexp.MustCompile(analysisRegexConfig.EPSRevisions.TableRowPattern)
	matches := re.FindAllStringSubmatch(match, -1)

	for _, rowMatch := range matches {
		if len(rowMatch) < 6 {
			continue
		}
		
		rowTitle := strings.TrimSpace(rowMatch[1])
		currentQtr := strings.TrimSpace(rowMatch[2])
		nextQtr := strings.TrimSpace(rowMatch[3])
		currentYear := strings.TrimSpace(rowMatch[4])
		nextYear := strings.TrimSpace(rowMatch[5])

		switch rowTitle {
		case "Up Last 7 Days":
			dto.EPSRevisions.CurrentQtr.UpLast7Days = parseInt(currentQtr)
			dto.EPSRevisions.NextQtr.UpLast7Days = parseInt(nextQtr)
			dto.EPSRevisions.CurrentYear.UpLast7Days = parseInt(currentYear)
			dto.EPSRevisions.NextYear.UpLast7Days = parseInt(nextYear)
		case "Up Last 30 Days":
			dto.EPSRevisions.CurrentQtr.UpLast30Days = parseInt(currentQtr)
			dto.EPSRevisions.NextQtr.UpLast30Days = parseInt(nextQtr)
			dto.EPSRevisions.CurrentYear.UpLast30Days = parseInt(currentYear)
			dto.EPSRevisions.NextYear.UpLast30Days = parseInt(nextYear)
		case "Down Last 7 Days":
			dto.EPSRevisions.CurrentQtr.DownLast7Days = parseInt(currentQtr)
			dto.EPSRevisions.NextQtr.DownLast7Days = parseInt(nextQtr)
			dto.EPSRevisions.CurrentYear.DownLast7Days = parseInt(currentYear)
			dto.EPSRevisions.NextYear.DownLast7Days = parseInt(nextYear)
		case "Down Last 30 Days":
			dto.EPSRevisions.CurrentQtr.DownLast30Days = parseInt(currentQtr)
			dto.EPSRevisions.NextQtr.DownLast30Days = parseInt(nextQtr)
			dto.EPSRevisions.CurrentYear.DownLast30Days = parseInt(currentYear)
			dto.EPSRevisions.NextYear.DownLast30Days = parseInt(nextYear)
		}
	}

	return nil
}

// extractGrowthEstimate extracts growth estimate data from HTML (only ticker data, not S&P 500)
func extractGrowthEstimate(html string, dto *ComprehensiveAnalysisDTO) error {
	// Find the growth estimates table section
	re := regexp.MustCompile(analysisRegexConfig.GrowthEstimate.SectionPattern)
	match := re.FindString(html)
	if match == "" {
		return fmt.Errorf("growth estimate section not found")
	}

	// Extract table rows - we only want the first row (ticker data, not S&P 500)
	re = regexp.MustCompile(analysisRegexConfig.GrowthEstimate.TableRowPattern)
	matches := re.FindAllStringSubmatch(match, -1)

	// Only process the first row (ticker data)
	if len(matches) > 0 && len(matches[0]) >= 6 {
		rowMatch := matches[0]
		currentQtr := strings.TrimSpace(rowMatch[2])
		nextQtr := strings.TrimSpace(rowMatch[3])
		currentYear := strings.TrimSpace(rowMatch[4])
		nextYear := strings.TrimSpace(rowMatch[5])

		dto.GrowthEstimate.CurrentQtr = parseString(currentQtr)
		dto.GrowthEstimate.NextQtr = parseString(nextQtr)
		dto.GrowthEstimate.CurrentYear = parseString(currentYear)
		dto.GrowthEstimate.NextYear = parseString(nextYear)
	}

	return nil
}
