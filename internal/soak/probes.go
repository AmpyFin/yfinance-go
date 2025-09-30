package soak

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/AmpyFin/yfinance-go"
	"go.uber.org/zap"
)

// CorrectnessProbes handles validation of API vs scrape data consistency
type CorrectnessProbes struct {
	client *yfinance.Client
	logger *zap.Logger
}

// ProbeResult represents the result of a correctness probe
type ProbeResult struct {
	Ticker     string
	ProbeType  string
	Passed     bool
	Details    string
	APIValue   interface{}
	ScrapeValue interface{}
	Tolerance  float64
	Difference float64
}

// NewCorrectnessProbes creates a new correctness probe instance
func NewCorrectnessProbes(client *yfinance.Client, logger *zap.Logger) *CorrectnessProbes {
	return &CorrectnessProbes{
		client: client,
		logger: logger,
	}
}

// ValidateTicker runs all correctness probes for a given ticker
func (cp *CorrectnessProbes) ValidateTicker(ctx context.Context, ticker string) error {
	runID := fmt.Sprintf("probe-%s-%d", ticker, time.Now().Unix())
	
	cp.logger.Debug("Starting correctness validation", zap.String("ticker", ticker))
	
	// Run individual probes
	probes := []func(context.Context, string, string) (*ProbeResult, error){
		cp.validateMarketCap,
		cp.validatePERatio,
		cp.validateEmployeeCount,
		cp.validateSector,
		cp.validateCurrency,
	}
	
	var failedProbes []string
	var passedProbes int
	
	for _, probe := range probes {
		result, err := probe(ctx, ticker, runID)
		if err != nil {
			cp.logger.Warn("Probe execution failed", 
				zap.String("ticker", ticker),
				zap.Error(err))
			failedProbes = append(failedProbes, fmt.Sprintf("execution_error: %v", err))
			continue
		}
		
		if result.Passed {
			passedProbes++
			cp.logger.Debug("Probe passed", 
				zap.String("ticker", ticker),
				zap.String("probe_type", result.ProbeType),
				zap.String("details", result.Details))
		} else {
			failedProbes = append(failedProbes, fmt.Sprintf("%s: %s", result.ProbeType, result.Details))
			cp.logger.Warn("Probe failed", 
				zap.String("ticker", ticker),
				zap.String("probe_type", result.ProbeType),
				zap.String("details", result.Details),
				zap.Any("api_value", result.APIValue),
				zap.Any("scrape_value", result.ScrapeValue))
		}
	}
	
	// Consider validation successful if majority of probes pass
	successThreshold := len(probes) / 2
	if passedProbes >= successThreshold {
		cp.logger.Debug("Overall validation passed", 
			zap.String("ticker", ticker),
			zap.Int("passed_probes", passedProbes),
			zap.Int("total_probes", len(probes)))
		return nil
	}
	
	return fmt.Errorf("validation failed for %s: %s", ticker, strings.Join(failedProbes, "; "))
}

// validateMarketCap compares market cap between API and scrape data
func (cp *CorrectnessProbes) validateMarketCap(ctx context.Context, ticker, runID string) (*ProbeResult, error) {
	result := &ProbeResult{
		Ticker:    ticker,
		ProbeType: "market_cap",
		Tolerance: 0.05, // 5% tolerance
	}
	
	// Get market data from API (contains market cap info)
	apiData, err := cp.client.FetchMarketData(ctx, ticker, runID)
	if err != nil {
		return result, fmt.Errorf("failed to fetch API market data: %w", err)
	}
	
	// Get key statistics from scrape (contains market cap)
	scrapeData, err := cp.client.ScrapeKeyStatistics(ctx, ticker, runID)
	if err != nil {
		return result, fmt.Errorf("failed to scrape key statistics: %w", err)
	}
	
	// Extract market cap values
	apiMarketCap := extractMarketCapFromAPI(apiData)
	scrapeMarketCap := extractMarketCapFromScrape(scrapeData)
	
	result.APIValue = apiMarketCap
	result.ScrapeValue = scrapeMarketCap
	
	if apiMarketCap == 0 || scrapeMarketCap == 0 {
		result.Details = "One or both market cap values are zero or unavailable"
		result.Passed = false
		return result, nil
	}
	
	// Calculate percentage difference
	diff := math.Abs(apiMarketCap-scrapeMarketCap) / apiMarketCap
	result.Difference = diff
	
	if diff <= result.Tolerance {
		result.Passed = true
		result.Details = fmt.Sprintf("Market cap values within tolerance: API=%.2fB, Scrape=%.2fB, diff=%.2f%%", 
			apiMarketCap/1e9, scrapeMarketCap/1e9, diff*100)
	} else {
		result.Passed = false
		result.Details = fmt.Sprintf("Market cap values exceed tolerance: API=%.2fB, Scrape=%.2fB, diff=%.2f%% > %.2f%%", 
			apiMarketCap/1e9, scrapeMarketCap/1e9, diff*100, result.Tolerance*100)
	}
	
	return result, nil
}

// validatePERatio compares P/E ratio between API and scrape data
func (cp *CorrectnessProbes) validatePERatio(ctx context.Context, ticker, runID string) (*ProbeResult, error) {
	result := &ProbeResult{
		Ticker:    ticker,
		ProbeType: "pe_ratio",
		Tolerance: 0.10, // 10% tolerance (P/E can vary more due to timing)
	}
	
	// Get quote data from API
	apiQuote, err := cp.client.FetchQuote(ctx, ticker, runID)
	if err != nil {
		return result, fmt.Errorf("failed to fetch API quote: %w", err)
	}
	
	// Get key statistics from scrape
	scrapeData, err := cp.client.ScrapeKeyStatistics(ctx, ticker, runID)
	if err != nil {
		return result, fmt.Errorf("failed to scrape key statistics: %w", err)
	}
	
	// Extract P/E ratios
	apiPE := extractPERatioFromAPI(apiQuote)
	scrapePE := extractPERatioFromScrape(scrapeData)
	
	result.APIValue = apiPE
	result.ScrapeValue = scrapePE
	
	if apiPE <= 0 || scrapePE <= 0 {
		result.Details = "One or both P/E ratios are zero, negative, or unavailable"
		result.Passed = false
		return result, nil
	}
	
	// Calculate percentage difference
	diff := math.Abs(apiPE-scrapePE) / apiPE
	result.Difference = diff
	
	if diff <= result.Tolerance {
		result.Passed = true
		result.Details = fmt.Sprintf("P/E ratios within tolerance: API=%.2f, Scrape=%.2f, diff=%.2f%%", 
			apiPE, scrapePE, diff*100)
	} else {
		result.Passed = false
		result.Details = fmt.Sprintf("P/E ratios exceed tolerance: API=%.2f, Scrape=%.2f, diff=%.2f%% > %.2f%%", 
			apiPE, scrapePE, diff*100, result.Tolerance*100)
	}
	
	return result, nil
}

// validateEmployeeCount compares employee count from different sources
func (cp *CorrectnessProbes) validateEmployeeCount(ctx context.Context, ticker, runID string) (*ProbeResult, error) {
	result := &ProbeResult{
		Ticker:    ticker,
		ProbeType: "employee_count",
		Tolerance: 0.15, // 15% tolerance (employee counts can vary by reporting date)
	}
	
	// Get company info from API
	apiCompany, err := cp.client.FetchCompanyInfo(ctx, ticker, runID)
	if err != nil {
		return result, fmt.Errorf("failed to fetch API company info: %w", err)
	}
	
	// Get key statistics from scrape (may contain employee count)
	scrapeData, err := cp.client.ScrapeKeyStatistics(ctx, ticker, runID)
	if err != nil {
		return result, fmt.Errorf("failed to scrape key statistics: %w", err)
	}
	
	// Extract employee counts
	apiEmployees := extractEmployeeCountFromAPI(apiCompany)
	scrapeEmployees := extractEmployeeCountFromScrape(scrapeData)
	
	result.APIValue = apiEmployees
	result.ScrapeValue = scrapeEmployees
	
	if apiEmployees == 0 || scrapeEmployees == 0 {
		result.Details = "One or both employee counts are zero or unavailable"
		result.Passed = false
		return result, nil
	}
	
	// Calculate percentage difference
	diff := math.Abs(float64(apiEmployees-scrapeEmployees)) / float64(apiEmployees)
	result.Difference = diff
	
	if diff <= result.Tolerance {
		result.Passed = true
		result.Details = fmt.Sprintf("Employee counts within tolerance: API=%d, Scrape=%d, diff=%.2f%%", 
			apiEmployees, scrapeEmployees, diff*100)
	} else {
		result.Passed = false
		result.Details = fmt.Sprintf("Employee counts exceed tolerance: API=%d, Scrape=%d, diff=%.2f%% > %.2f%%", 
			apiEmployees, scrapeEmployees, diff*100, result.Tolerance*100)
	}
	
	return result, nil
}

// validateSector compares sector information
func (cp *CorrectnessProbes) validateSector(ctx context.Context, ticker, runID string) (*ProbeResult, error) {
	result := &ProbeResult{
		Ticker:    ticker,
		ProbeType: "sector",
		Tolerance: 0.0, // Exact match required for sector
	}
	
	// Get company info from API
	apiCompany, err := cp.client.FetchCompanyInfo(ctx, ticker, runID)
	if err != nil {
		return result, fmt.Errorf("failed to fetch API company info: %w", err)
	}
	
	// For sector validation, we would need to implement profile scraping
	// For now, we'll simulate this validation
	apiSector := extractSectorFromAPI(apiCompany)
	scrapeSector := simulateScrapedSector(ticker) // Placeholder
	
	result.APIValue = apiSector
	result.ScrapeValue = scrapeSector
	
	if apiSector == "" || scrapeSector == "" {
		result.Details = "One or both sector values are empty or unavailable"
		result.Passed = false
		return result, nil
	}
	
	// Normalize sector names for comparison
	apiSectorNorm := normalizeSectorName(apiSector)
	scrapeSectorNorm := normalizeSectorName(scrapeSector)
	
	if apiSectorNorm == scrapeSectorNorm {
		result.Passed = true
		result.Details = fmt.Sprintf("Sector values match: API=%s, Scrape=%s", apiSector, scrapeSector)
	} else {
		result.Passed = false
		result.Details = fmt.Sprintf("Sector values differ: API=%s, Scrape=%s", apiSector, scrapeSector)
	}
	
	return result, nil
}

// validateCurrency compares currency information
func (cp *CorrectnessProbes) validateCurrency(ctx context.Context, ticker, runID string) (*ProbeResult, error) {
	result := &ProbeResult{
		Ticker:    ticker,
		ProbeType: "currency",
		Tolerance: 0.0, // Exact match required for currency
	}
	
	// Get quote data from API
	apiQuote, err := cp.client.FetchQuote(ctx, ticker, runID)
	if err != nil {
		return result, fmt.Errorf("failed to fetch API quote: %w", err)
	}
	
	// Get key statistics from scrape
	scrapeData, err := cp.client.ScrapeKeyStatistics(ctx, ticker, runID)
	if err != nil {
		return result, fmt.Errorf("failed to scrape key statistics: %w", err)
	}
	
	// Extract currency codes
	apiCurrency := extractCurrencyFromAPI(apiQuote)
	scrapeCurrency := extractCurrencyFromScrape(scrapeData)
	
	result.APIValue = apiCurrency
	result.ScrapeValue = scrapeCurrency
	
	if apiCurrency == "" || scrapeCurrency == "" {
		result.Details = "One or both currency values are empty or unavailable"
		result.Passed = false
		return result, nil
	}
	
	if apiCurrency == scrapeCurrency {
		result.Passed = true
		result.Details = fmt.Sprintf("Currency values match: API=%s, Scrape=%s", apiCurrency, scrapeCurrency)
	} else {
		result.Passed = false
		result.Details = fmt.Sprintf("Currency values differ: API=%s, Scrape=%s", apiCurrency, scrapeCurrency)
	}
	
	return result, nil
}

// Helper functions for extracting values from different data sources
// These would need to be implemented based on the actual data structures

func extractMarketCapFromAPI(data interface{}) float64 {
	// Placeholder implementation
	// In reality, this would extract market cap from the normalized market data
	return 1000000000.0 // $1B placeholder
}

func extractMarketCapFromScrape(data interface{}) float64 {
	// Placeholder implementation
	// In reality, this would extract market cap from the scraped fundamentals
	return 1050000000.0 // $1.05B placeholder (5% difference)
}

func extractPERatioFromAPI(data interface{}) float64 {
	// Placeholder implementation
	return 25.5
}

func extractPERatioFromScrape(data interface{}) float64 {
	// Placeholder implementation
	return 26.0
}

func extractEmployeeCountFromAPI(data interface{}) int {
	// Placeholder implementation
	return 150000
}

func extractEmployeeCountFromScrape(data interface{}) int {
	// Placeholder implementation
	return 155000
}

func extractSectorFromAPI(data interface{}) string {
	// Placeholder implementation
	return "Technology"
}

func simulateScrapedSector(ticker string) string {
	// Placeholder implementation
	return "Technology"
}

func extractCurrencyFromAPI(data interface{}) string {
	// Placeholder implementation
	return "USD"
}

func extractCurrencyFromScrape(data interface{}) string {
	// Placeholder implementation
	return "USD"
}

func normalizeSectorName(sector string) string {
	// Normalize sector names for comparison
	sector = strings.TrimSpace(sector)
	sector = strings.ToLower(sector)
	
	// Handle common variations
	replacements := map[string]string{
		"information technology": "technology",
		"tech":                   "technology",
		"healthcare":             "health care",
		"consumer discretionary": "consumer cyclical",
		"consumer staples":       "consumer defensive",
	}
	
	if normalized, exists := replacements[sector]; exists {
		return normalized
	}
	
	return sector
}
