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

// YahooFinanceData represents the JSON structure from Yahoo Finance
type YahooFinanceData struct {
	QuoteSummary struct {
		Result []struct {
			FinancialData struct {
				TrailingTotalRevenue []FinancialDataPoint `json:"trailingTotalRevenue"`
				AnnualTotalRevenue   []FinancialDataPoint `json:"annualTotalRevenue"`
				TrailingOperatingIncome []FinancialDataPoint `json:"trailingOperatingIncome"`
				AnnualOperatingIncome   []FinancialDataPoint `json:"annualOperatingIncome"`
				TrailingNetIncome []FinancialDataPoint `json:"trailingNetIncome"`
				AnnualNetIncome   []FinancialDataPoint `json:"annualNetIncome"`
				TrailingBasicEPS []FinancialDataPoint `json:"trailingBasicEPS"`
				AnnualBasicEPS   []FinancialDataPoint `json:"annualBasicEPS"`
				TrailingDilutedEPS []FinancialDataPoint `json:"trailingDilutedEPS"`
				AnnualDilutedEPS   []FinancialDataPoint `json:"annualDilutedEPS"`
				TrailingEBITDA []FinancialDataPoint `json:"trailingEBITDA"`
				AnnualEBITDA   []FinancialDataPoint `json:"annualEBITDA"`
				TrailingGrossProfit []FinancialDataPoint `json:"trailingGrossProfit"`
				AnnualGrossProfit   []FinancialDataPoint `json:"annualGrossProfit"`
				TrailingCostOfRevenue []FinancialDataPoint `json:"trailingCostOfRevenue"`
				AnnualCostOfRevenue   []FinancialDataPoint `json:"annualCostOfRevenue"`
				TrailingOperatingExpense []FinancialDataPoint `json:"trailingOperatingExpense"`
				AnnualOperatingExpense   []FinancialDataPoint `json:"annualOperatingExpense"`
				TrailingTotalExpenses []FinancialDataPoint `json:"trailingTotalExpenses"`
				AnnualTotalExpenses   []FinancialDataPoint `json:"annualTotalExpenses"`
				TrailingTaxProvision []FinancialDataPoint `json:"trailingTaxProvision"`
				AnnualTaxProvision   []FinancialDataPoint `json:"annualTaxProvision"`
				TrailingPretaxIncome []FinancialDataPoint `json:"trailingPretaxIncome"`
				AnnualPretaxIncome   []FinancialDataPoint `json:"annualPretaxIncome"`
				TrailingOtherIncomeExpense []FinancialDataPoint `json:"trailingOtherIncomeExpense"`
				AnnualOtherIncomeExpense   []FinancialDataPoint `json:"annualOtherIncomeExpense"`
				TrailingNetNonOperatingInterestIncomeExpense []FinancialDataPoint `json:"trailingNetNonOperatingInterestIncomeExpense"`
				AnnualNetNonOperatingInterestIncomeExpense   []FinancialDataPoint `json:"annualNetNonOperatingInterestIncomeExpense"`
				TrailingBasicAverageShares []FinancialDataPoint `json:"trailingBasicAverageShares"`
				AnnualBasicAverageShares   []FinancialDataPoint `json:"annualBasicAverageShares"`
				TrailingDilutedAverageShares []FinancialDataPoint `json:"trailingDilutedAverageShares"`
				AnnualDilutedAverageShares   []FinancialDataPoint `json:"annualDilutedAverageShares"`
				TrailingEBIT []FinancialDataPoint `json:"trailingEBIT"`
				AnnualEBIT   []FinancialDataPoint `json:"annualEBIT"`
				TrailingNormalizedIncome []FinancialDataPoint `json:"trailingNormalizedIncome"`
				AnnualNormalizedIncome   []FinancialDataPoint `json:"annualNormalizedIncome"`
				TrailingNormalizedEBITDA []FinancialDataPoint `json:"trailingNormalizedEBITDA"`
				AnnualNormalizedEBITDA   []FinancialDataPoint `json:"annualNormalizedEBITDA"`
				TrailingReconciledCostOfRevenue []FinancialDataPoint `json:"trailingReconciledCostOfRevenue"`
				AnnualReconciledCostOfRevenue   []FinancialDataPoint `json:"annualReconciledCostOfRevenue"`
				TrailingReconciledDepreciation []FinancialDataPoint `json:"trailingReconciledDepreciation"`
				AnnualReconciledDepreciation   []FinancialDataPoint `json:"annualReconciledDepreciation"`
			} `json:"financialData"`
		} `json:"result"`
	} `json:"quoteSummary"`
}

// FinancialDataPoint represents a single financial data point from Yahoo Finance
type FinancialDataPoint struct {
	DataID       int64  `json:"dataId"`
	AsOfDate     string `json:"asOfDate"`
	PeriodType   string `json:"periodType"`
	CurrencyCode string `json:"currencyCode"`
	ReportedValue struct {
		Raw float64 `json:"raw"`
		Fmt string  `json:"fmt"`
	} `json:"reportedValue"`
}

// ComprehensiveFinancialsDTO holds all financials data including historical
type ComprehensiveFinancialsDTO struct {
	Symbol   string    `json:"symbol"`
	Market   string    `json:"market"`
	Currency string    `json:"currency"`
	AsOf     time.Time `json:"as_of"`
	
	// Current values (most recent quarter)
	Current struct {
		TotalRevenue                           *Scaled `json:"total_revenue,omitempty"`
		CostOfRevenue                          *Scaled `json:"cost_of_revenue,omitempty"`
		GrossProfit                            *Scaled `json:"gross_profit,omitempty"`
		OperatingExpense                       *Scaled `json:"operating_expense,omitempty"`
		OperatingIncome                        *Scaled `json:"operating_income,omitempty"`
		NetNonOperatingInterestIncomeExpense   *Scaled `json:"net_non_operating_interest_income_expense,omitempty"`
		OtherIncomeExpense                     *Scaled `json:"other_income_expense,omitempty"`
		PretaxIncome                           *Scaled `json:"pretax_income,omitempty"`
		TaxProvision                           *Scaled `json:"tax_provision,omitempty"`
		NetIncomeCommonStockholders            *Scaled `json:"net_income_common_stockholders,omitempty"`
		BasicEPS                               *Scaled `json:"basic_eps,omitempty"`
		DilutedEPS                             *Scaled `json:"diluted_eps,omitempty"`
		BasicAverageShares                     *int64  `json:"basic_average_shares,omitempty"`
		DilutedAverageShares                   *int64  `json:"diluted_average_shares,omitempty"`
		TotalExpenses                          *Scaled `json:"total_expenses,omitempty"`
		NormalizedIncome                       *Scaled `json:"normalized_income,omitempty"`
		EBIT                                   *Scaled `json:"ebit,omitempty"`
		EBITDA                                 *Scaled `json:"ebitda,omitempty"`
		ReconciledCostOfRevenue                *Scaled `json:"reconciled_cost_of_revenue,omitempty"`
		ReconciledDepreciation                 *Scaled `json:"reconciled_depreciation,omitempty"`
		NormalizedEBITDA                       *Scaled `json:"normalized_ebitda,omitempty"`
		
		// Balance Sheet fields
		TotalAssets                            *Scaled `json:"total_assets,omitempty"`
		TotalCapitalization                    *Scaled `json:"total_capitalization,omitempty"`
		CommonStockEquity                      *Scaled `json:"common_stock_equity,omitempty"`
		CapitalLeaseObligations                *Scaled `json:"capital_lease_obligations,omitempty"`
		NetTangibleAssets                      *Scaled `json:"net_tangible_assets,omitempty"`
		WorkingCapital                         *Scaled `json:"working_capital,omitempty"`
		InvestedCapital                        *Scaled `json:"invested_capital,omitempty"`
		TangibleBookValue                      *Scaled `json:"tangible_book_value,omitempty"`
		TotalDebt                              *Scaled `json:"total_debt,omitempty"`
		ShareIssued                            *int64  `json:"share_issued,omitempty"`
		
		// Cash Flow fields
		OperatingCashFlow                      *Scaled `json:"operating_cash_flow,omitempty"`
		InvestingCashFlow                      *Scaled `json:"investing_cash_flow,omitempty"`
		FinancingCashFlow                      *Scaled `json:"financing_cash_flow,omitempty"`
		EndCashPosition                        *Scaled `json:"end_cash_position,omitempty"`
		CapitalExpenditure                     *Scaled `json:"capital_expenditure,omitempty"`
		IssuanceOfDebt                         *Scaled `json:"issuance_of_debt,omitempty"`
		RepaymentOfDebt                        *Scaled `json:"repayment_of_debt,omitempty"`
		RepurchaseOfCapitalStock               *Scaled `json:"repurchase_of_capital_stock,omitempty"`
		FreeCashFlow                           *Scaled `json:"free_cash_flow,omitempty"`
	} `json:"current"`
	
	// Historical values
	Historical struct {
		Q2_2025 struct {
			Date                                 string  `json:"date"`
			TotalRevenue                         *Scaled `json:"total_revenue,omitempty"`
			CostOfRevenue                        *Scaled `json:"cost_of_revenue,omitempty"`
			GrossProfit                          *Scaled `json:"gross_profit,omitempty"`
			OperatingExpense                     *Scaled `json:"operating_expense,omitempty"`
			OperatingIncome                      *Scaled `json:"operating_income,omitempty"`
			NetNonOperatingInterestIncomeExpense *Scaled `json:"net_non_operating_interest_income_expense,omitempty"`
			OtherIncomeExpense                   *Scaled `json:"other_income_expense,omitempty"`
			PretaxIncome                         *Scaled `json:"pretax_income,omitempty"`
			TaxProvision                         *Scaled `json:"tax_provision,omitempty"`
			NetIncomeCommonStockholders          *Scaled `json:"net_income_common_stockholders,omitempty"`
			BasicEPS                             *Scaled `json:"basic_eps,omitempty"`
			DilutedEPS                           *Scaled `json:"diluted_eps,omitempty"`
			BasicAverageShares                   *int64  `json:"basic_average_shares,omitempty"`
			DilutedAverageShares                 *int64  `json:"diluted_average_shares,omitempty"`
			TotalExpenses                        *Scaled `json:"total_expenses,omitempty"`
			NormalizedIncome                     *Scaled `json:"normalized_income,omitempty"`
			EBIT                                 *Scaled `json:"ebit,omitempty"`
			EBITDA                               *Scaled `json:"ebitda,omitempty"`
			ReconciledCostOfRevenue              *Scaled `json:"reconciled_cost_of_revenue,omitempty"`
			ReconciledDepreciation               *Scaled `json:"reconciled_depreciation,omitempty"`
			NormalizedEBITDA                     *Scaled `json:"normalized_ebitda,omitempty"`
		} `json:"q2_2025"`
		
		Q1_2025 struct {
			Date                                 string  `json:"date"`
			TotalRevenue                         *Scaled `json:"total_revenue,omitempty"`
			CostOfRevenue                        *Scaled `json:"cost_of_revenue,omitempty"`
			GrossProfit                          *Scaled `json:"gross_profit,omitempty"`
			OperatingExpense                     *Scaled `json:"operating_expense,omitempty"`
			OperatingIncome                      *Scaled `json:"operating_income,omitempty"`
			NetNonOperatingInterestIncomeExpense *Scaled `json:"net_non_operating_interest_income_expense,omitempty"`
			OtherIncomeExpense                   *Scaled `json:"other_income_expense,omitempty"`
			PretaxIncome                         *Scaled `json:"pretax_income,omitempty"`
			TaxProvision                         *Scaled `json:"tax_provision,omitempty"`
			NetIncomeCommonStockholders          *Scaled `json:"net_income_common_stockholders,omitempty"`
			BasicEPS                             *Scaled `json:"basic_eps,omitempty"`
			DilutedEPS                           *Scaled `json:"diluted_eps,omitempty"`
			BasicAverageShares                   *int64  `json:"basic_average_shares,omitempty"`
			DilutedAverageShares                 *int64  `json:"diluted_average_shares,omitempty"`
			TotalExpenses                        *Scaled `json:"total_expenses,omitempty"`
			NormalizedIncome                     *Scaled `json:"normalized_income,omitempty"`
			EBIT                                 *Scaled `json:"ebit,omitempty"`
			EBITDA                               *Scaled `json:"ebitda,omitempty"`
			ReconciledCostOfRevenue              *Scaled `json:"reconciled_cost_of_revenue,omitempty"`
			ReconciledDepreciation               *Scaled `json:"reconciled_depreciation,omitempty"`
			NormalizedEBITDA                     *Scaled `json:"normalized_ebitda,omitempty"`
		} `json:"q1_2025"`
		
		Q4_2024 struct {
			Date                                 string  `json:"date"`
			TotalRevenue                         *Scaled `json:"total_revenue,omitempty"`
			CostOfRevenue                        *Scaled `json:"cost_of_revenue,omitempty"`
			GrossProfit                          *Scaled `json:"gross_profit,omitempty"`
			OperatingExpense                     *Scaled `json:"operating_expense,omitempty"`
			OperatingIncome                      *Scaled `json:"operating_income,omitempty"`
			NetNonOperatingInterestIncomeExpense *Scaled `json:"net_non_operating_interest_income_expense,omitempty"`
			OtherIncomeExpense                   *Scaled `json:"other_income_expense,omitempty"`
			PretaxIncome                         *Scaled `json:"pretax_income,omitempty"`
			TaxProvision                         *Scaled `json:"tax_provision,omitempty"`
			NetIncomeCommonStockholders          *Scaled `json:"net_income_common_stockholders,omitempty"`
			BasicEPS                             *Scaled `json:"basic_eps,omitempty"`
			DilutedEPS                           *Scaled `json:"diluted_eps,omitempty"`
			BasicAverageShares                   *int64  `json:"basic_average_shares,omitempty"`
			DilutedAverageShares                 *int64  `json:"diluted_average_shares,omitempty"`
			TotalExpenses                        *Scaled `json:"total_expenses,omitempty"`
			NormalizedIncome                     *Scaled `json:"normalized_income,omitempty"`
			EBIT                                 *Scaled `json:"ebit,omitempty"`
			EBITDA                               *Scaled `json:"ebitda,omitempty"`
			ReconciledCostOfRevenue              *Scaled `json:"reconciled_cost_of_revenue,omitempty"`
			ReconciledDepreciation               *Scaled `json:"reconciled_depreciation,omitempty"`
			NormalizedEBITDA                     *Scaled `json:"normalized_ebitda,omitempty"`
		} `json:"q4_2024"`
		
		Q3_2024 struct {
			Date                                 string  `json:"date"`
			TotalRevenue                         *Scaled `json:"total_revenue,omitempty"`
			CostOfRevenue                        *Scaled `json:"cost_of_revenue,omitempty"`
			GrossProfit                          *Scaled `json:"gross_profit,omitempty"`
			OperatingExpense                     *Scaled `json:"operating_expense,omitempty"`
			OperatingIncome                      *Scaled `json:"operating_income,omitempty"`
			NetNonOperatingInterestIncomeExpense *Scaled `json:"net_non_operating_interest_income_expense,omitempty"`
			OtherIncomeExpense                   *Scaled `json:"other_income_expense,omitempty"`
			PretaxIncome                         *Scaled `json:"pretax_income,omitempty"`
			TaxProvision                         *Scaled `json:"tax_provision,omitempty"`
			NetIncomeCommonStockholders          *Scaled `json:"net_income_common_stockholders,omitempty"`
			BasicEPS                             *Scaled `json:"basic_eps,omitempty"`
			DilutedEPS                           *Scaled `json:"diluted_eps,omitempty"`
			BasicAverageShares                   *int64  `json:"basic_average_shares,omitempty"`
			DilutedAverageShares                 *int64  `json:"diluted_average_shares,omitempty"`
			TotalExpenses                        *Scaled `json:"total_expenses,omitempty"`
			NormalizedIncome                     *Scaled `json:"normalized_income,omitempty"`
			EBIT                                 *Scaled `json:"ebit,omitempty"`
			EBITDA                               *Scaled `json:"ebitda,omitempty"`
			ReconciledCostOfRevenue              *Scaled `json:"reconciled_cost_of_revenue,omitempty"`
			ReconciledDepreciation               *Scaled `json:"reconciled_depreciation,omitempty"`
			NormalizedEBITDA                     *Scaled `json:"normalized_ebitda,omitempty"`
		} `json:"q3_2024"`
		
		Q2_2024 struct {
			Date                                 string  `json:"date"`
			TotalRevenue                         *Scaled `json:"total_revenue,omitempty"`
			CostOfRevenue                        *Scaled `json:"cost_of_revenue,omitempty"`
			GrossProfit                          *Scaled `json:"gross_profit,omitempty"`
			OperatingExpense                     *Scaled `json:"operating_expense,omitempty"`
			OperatingIncome                      *Scaled `json:"operating_income,omitempty"`
			NetNonOperatingInterestIncomeExpense *Scaled `json:"net_non_operating_interest_income_expense,omitempty"`
			OtherIncomeExpense                   *Scaled `json:"other_income_expense,omitempty"`
			PretaxIncome                         *Scaled `json:"pretax_income,omitempty"`
			TaxProvision                         *Scaled `json:"tax_provision,omitempty"`
			NetIncomeCommonStockholders          *Scaled `json:"net_income_common_stockholders,omitempty"`
			BasicEPS                             *Scaled `json:"basic_eps,omitempty"`
			DilutedEPS                           *Scaled `json:"diluted_eps,omitempty"`
			BasicAverageShares                   *int64  `json:"basic_average_shares,omitempty"`
			DilutedAverageShares                 *int64  `json:"diluted_average_shares,omitempty"`
			TotalExpenses                        *Scaled `json:"total_expenses,omitempty"`
			NormalizedIncome                     *Scaled `json:"normalized_income,omitempty"`
			EBIT                                 *Scaled `json:"ebit,omitempty"`
			EBITDA                               *Scaled `json:"ebitda,omitempty"`
			ReconciledCostOfRevenue              *Scaled `json:"reconciled_cost_of_revenue,omitempty"`
			ReconciledDepreciation               *Scaled `json:"reconciled_depreciation,omitempty"`
			NormalizedEBITDA                     *Scaled `json:"normalized_ebitda,omitempty"`
		} `json:"q2_2024"`
	} `json:"historical"`
}

// FinancialsRegexConfig holds the regex patterns for financials extraction
type FinancialsRegexConfig struct {
	Currency struct {
		Pattern string `yaml:"pattern"`
	} `yaml:"currency"`
	
	IncomeStatement struct {
		TotalRevenue        string `yaml:"total_revenue"`
		CostOfRevenue       string `yaml:"cost_of_revenue"`
		OperatingIncome     string `yaml:"operating_income"`
		NetIncome          string `yaml:"net_income"`
		BasicEPS           string `yaml:"basic_eps"`
		DilutedEPS         string `yaml:"diluted_eps"`
		EBITDA             string `yaml:"ebitda"`
		EBIT               string `yaml:"ebit"`
		TotalExpenses      string `yaml:"total_expenses"`
		NormalizedEBITDA   string `yaml:"normalized_ebitda"`
	} `yaml:"income_statement"`
	
	Shares struct {
		BasicAverageShares   string `yaml:"basic_average_shares"`
		DilutedAverageShares string `yaml:"diluted_average_shares"`
	} `yaml:"shares"`
	
	BalanceSheet struct {
		TotalAssets             string `yaml:"total_assets"`
		TotalCapitalization     string `yaml:"total_capitalization"`
		CommonStockEquity       string `yaml:"common_stock_equity"`
		CapitalLeaseObligations string `yaml:"capital_lease_obligations"`
		NetTangibleAssets       string `yaml:"net_tangible_assets"`
		WorkingCapital          string `yaml:"working_capital"`
		InvestedCapital         string `yaml:"invested_capital"`
		TangibleBookValue       string `yaml:"tangible_book_value"`
		TotalDebt              string `yaml:"total_debt"`
		ShareIssued            string `yaml:"share_issued"`
	} `yaml:"balance_sheet"`
	
	CashFlow struct {
		OperatingCashFlow            string `yaml:"operating_cash_flow"`
		InvestingCashFlow            string `yaml:"investing_cash_flow"`
		FinancingCashFlow            string `yaml:"financing_cash_flow"`
		EndCashPosition              string `yaml:"end_cash_position"`
		CapitalExpenditure           string `yaml:"capital_expenditure"`
		IssuanceOfDebt               string `yaml:"issuance_of_debt"`
		RepaymentOfDebt              string `yaml:"repayment_of_debt"`
		RepurchaseOfCapitalStock     string `yaml:"repurchase_of_capital_stock"`
		FreeCashFlow                 string `yaml:"free_cash_flow"`
	} `yaml:"cash_flow"`
}

var financialsRegexConfig *FinancialsRegexConfig

// LoadFinancialsRegexConfig loads the regex patterns from YAML file
func LoadFinancialsRegexConfig() error {
	if financialsRegexConfig != nil {
		return nil // Already loaded
	}
	
	// Get the directory of the current file
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return fmt.Errorf("unable to get current file path")
	}
	
	configPath := filepath.Join(filepath.Dir(filename), "regex", "financials.yaml")
	
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read financials regex config file: %w", err)
	}
	
	financialsRegexConfig = &FinancialsRegexConfig{}
	if err := yaml.Unmarshal(data, financialsRegexConfig); err != nil {
		return fmt.Errorf("failed to parse financials regex config YAML: %w", err)
	}
	
	return nil
}

// ParseComprehensiveFinancials extracts comprehensive financials data from HTML using JSON parsing
func ParseComprehensiveFinancials(html []byte, symbol, market string) (*ComprehensiveFinancialsDTO, error) {
	if err := LoadFinancialsRegexConfig(); err != nil {
		return nil, fmt.Errorf("failed to load financials regex config: %w", err)
	}
	
	dto := &ComprehensiveFinancialsDTO{
		Symbol:   symbol,
		Market:   market,
		Currency: "USD", // Default, will be updated from actual data
		AsOf:     time.Now().UTC(),
	}
	
	htmlStr := string(html)
	
	// Extract financial data from HTML table
	financialData, err := extractFinancialDataFromHTML(htmlStr)
	if err != nil {
		return nil, fmt.Errorf("failed to extract financial data from HTML: %w", err)
	}
	
	// Populate the DTO with extracted data
	populateDTOFromHTMLData(financialData, dto)
	
	return dto, nil
}

// ParseComprehensiveFinancialsWithCurrency parses financial data from one HTML source and currency from financials HTML
func ParseComprehensiveFinancialsWithCurrency(html, financialsHTML []byte, symbol, market string) (*ComprehensiveFinancialsDTO, error) {
	if err := LoadFinancialsRegexConfig(); err != nil {
		return nil, fmt.Errorf("failed to load financials regex config: %w", err)
	}
	
	dto := &ComprehensiveFinancialsDTO{
		Symbol: symbol,
		Market: market,
		AsOf:   time.Now(),
	}

	// Extract currency from financials HTML
	financialsStr := string(financialsHTML)
	re := regexp.MustCompile(financialsRegexConfig.Currency.Pattern)
	matches := re.FindStringSubmatch(financialsStr)
	if len(matches) > 1 {
		dto.Currency = matches[1]
	} else {
		dto.Currency = "USD" // Default fallback
	}

	// Extract financial data from the main HTML (balance sheet or cash flow)
	htmlStr := string(html)
	
	// Extract financial data from HTML table
	financialData, err := extractFinancialDataFromHTML(htmlStr)
	if err != nil {
		return nil, fmt.Errorf("failed to extract financial data from HTML: %w", err)
	}
	
	// Override the currency with the one from financials page
	financialData["Currency"] = dto.Currency

	// Populate the DTO with extracted data
	populateDTOFromHTMLData(financialData, dto)

	return dto, nil
}

// extractFinancialDataFromHTML extracts financial data from Yahoo Finance HTML table
func extractFinancialDataFromHTML(html string) (map[string]string, error) {
	// The financial data is in HTML table format, not JSON
	// Look for the financial table structure
	financialData := make(map[string]string)
	
	// Extract currency information
	re := regexp.MustCompile(financialsRegexConfig.Currency.Pattern)
	matches := re.FindStringSubmatch(html)
	if len(matches) > 1 {
		financialData["Currency"] = matches[1]
	} else {
		financialData["Currency"] = "USD" // Default fallback
	}
	
	// Extract Total Revenue data
	re = regexp.MustCompile(financialsRegexConfig.IncomeStatement.TotalRevenue)
	matches = re.FindStringSubmatch(html)
	if len(matches) > 2 {
		financialData["TTM_TotalRevenue"] = strings.TrimSpace(strings.ReplaceAll(matches[1], ",", ""))
		financialData["2024_TotalRevenue"] = strings.TrimSpace(strings.ReplaceAll(matches[2], ",", ""))
	}
	
	// Extract Operating Income data
	re = regexp.MustCompile(financialsRegexConfig.IncomeStatement.OperatingIncome)
	matches = re.FindStringSubmatch(html)
	if len(matches) > 2 {
		financialData["TTM_OperatingIncome"] = strings.TrimSpace(strings.ReplaceAll(matches[1], ",", ""))
		financialData["2024_OperatingIncome"] = strings.TrimSpace(strings.ReplaceAll(matches[2], ",", ""))
	}
	
	// Extract Net Income data
	re = regexp.MustCompile(financialsRegexConfig.IncomeStatement.NetIncome)
	matches = re.FindStringSubmatch(html)
	if len(matches) > 2 {
		financialData["TTM_NetIncome"] = strings.TrimSpace(strings.ReplaceAll(matches[1], ",", ""))
		financialData["2024_NetIncome"] = strings.TrimSpace(strings.ReplaceAll(matches[2], ",", ""))
	}
	
	// Extract Basic EPS data
	re = regexp.MustCompile(financialsRegexConfig.IncomeStatement.BasicEPS)
	matches = re.FindStringSubmatch(html)
	if len(matches) > 2 {
		financialData["TTM_BasicEPS"] = strings.TrimSpace(matches[1])
		financialData["2024_BasicEPS"] = strings.TrimSpace(matches[2])
	}
	
	// Extract EBITDA data
	re = regexp.MustCompile(financialsRegexConfig.IncomeStatement.EBITDA)
	matches = re.FindStringSubmatch(html)
	if len(matches) > 2 {
		financialData["TTM_EBITDA"] = strings.TrimSpace(strings.ReplaceAll(matches[1], ",", ""))
		financialData["2024_EBITDA"] = strings.TrimSpace(strings.ReplaceAll(matches[2], ",", ""))
	}
	
	// Extract Cost of Revenue data
	re = regexp.MustCompile(financialsRegexConfig.IncomeStatement.CostOfRevenue)
	matches = re.FindStringSubmatch(html)
	if len(matches) > 2 {
		financialData["TTM_CostOfRevenue"] = strings.TrimSpace(strings.ReplaceAll(matches[1], ",", ""))
		financialData["2024_CostOfRevenue"] = strings.TrimSpace(strings.ReplaceAll(matches[2], ",", ""))
	}
	
	// Extract Diluted EPS data - flexible pattern for different HTML structures
	re = regexp.MustCompile(financialsRegexConfig.IncomeStatement.DilutedEPS)
	matches = re.FindStringSubmatch(html)
	if len(matches) > 2 {
		financialData["TTM_DilutedEPS"] = strings.TrimSpace(strings.ReplaceAll(matches[1], ",", ""))
		financialData["2024_DilutedEPS"] = strings.TrimSpace(strings.ReplaceAll(matches[2], ",", ""))
	}
	
	// Extract Basic Average Shares data - flexible pattern for different HTML structures
	re = regexp.MustCompile(financialsRegexConfig.Shares.BasicAverageShares)
	matches = re.FindStringSubmatch(html)
	if len(matches) > 2 {
		financialData["TTM_BasicAverageShares"] = strings.TrimSpace(strings.ReplaceAll(matches[1], ",", ""))
		financialData["2024_BasicAverageShares"] = strings.TrimSpace(strings.ReplaceAll(matches[2], ",", ""))
	}
	
	// Extract Diluted Average Shares data - flexible pattern for different HTML structures
	re = regexp.MustCompile(financialsRegexConfig.Shares.DilutedAverageShares)
	matches = re.FindStringSubmatch(html)
	if len(matches) > 2 {
		financialData["TTM_DilutedAverageShares"] = strings.TrimSpace(strings.ReplaceAll(matches[1], ",", ""))
		financialData["2024_DilutedAverageShares"] = strings.TrimSpace(strings.ReplaceAll(matches[2], ",", ""))
	}
	
	// Extract Total Expenses data
	re = regexp.MustCompile(financialsRegexConfig.IncomeStatement.TotalExpenses)
	matches = re.FindStringSubmatch(html)
	if len(matches) > 2 {
		financialData["TTM_TotalExpenses"] = strings.TrimSpace(strings.ReplaceAll(matches[1], ",", ""))
		financialData["2024_TotalExpenses"] = strings.TrimSpace(strings.ReplaceAll(matches[2], ",", ""))
	}
	
	// Extract EBIT data
	re = regexp.MustCompile(financialsRegexConfig.IncomeStatement.EBIT)
	matches = re.FindStringSubmatch(html)
	if len(matches) > 2 {
		financialData["TTM_EBIT"] = strings.TrimSpace(strings.ReplaceAll(matches[1], ",", ""))
		financialData["2024_EBIT"] = strings.TrimSpace(strings.ReplaceAll(matches[2], ",", ""))
	}
	
	// Extract Normalized EBITDA data
	re = regexp.MustCompile(financialsRegexConfig.IncomeStatement.NormalizedEBITDA)
	matches = re.FindStringSubmatch(html)
	if len(matches) > 2 {
		financialData["TTM_NormalizedEBITDA"] = strings.TrimSpace(strings.ReplaceAll(matches[1], ",", ""))
		financialData["2024_NormalizedEBITDA"] = strings.TrimSpace(strings.ReplaceAll(matches[2], ",", ""))
	}
	
	// Balance Sheet extraction patterns
	// Extract Total Assets data
	re = regexp.MustCompile(financialsRegexConfig.BalanceSheet.TotalAssets)
	matches = re.FindStringSubmatch(html)
	if len(matches) > 2 {
		financialData["Current_TotalAssets"] = strings.TrimSpace(strings.ReplaceAll(matches[1], ",", ""))
		financialData["2024_TotalAssets"] = strings.TrimSpace(strings.ReplaceAll(matches[2], ",", ""))
	}
	
	// Extract Total Capitalization data
	re = regexp.MustCompile(financialsRegexConfig.BalanceSheet.TotalCapitalization)
	matches = re.FindStringSubmatch(html)
	if len(matches) > 2 {
		financialData["Current_TotalCapitalization"] = strings.TrimSpace(strings.ReplaceAll(matches[1], ",", ""))
		financialData["2024_TotalCapitalization"] = strings.TrimSpace(strings.ReplaceAll(matches[2], ",", ""))
	}
	
	// Extract Common Stock Equity data
	re = regexp.MustCompile(financialsRegexConfig.BalanceSheet.CommonStockEquity)
	matches = re.FindStringSubmatch(html)
	if len(matches) > 2 {
		financialData["Current_CommonStockEquity"] = strings.TrimSpace(strings.ReplaceAll(matches[1], ",", ""))
		financialData["2024_CommonStockEquity"] = strings.TrimSpace(strings.ReplaceAll(matches[2], ",", ""))
	}
	
	// Extract Capital Lease Obligations data
	re = regexp.MustCompile(financialsRegexConfig.BalanceSheet.CapitalLeaseObligations)
	matches = re.FindStringSubmatch(html)
	if len(matches) > 2 {
		financialData["Current_CapitalLeaseObligations"] = strings.TrimSpace(strings.ReplaceAll(matches[1], ",", ""))
		financialData["2024_CapitalLeaseObligations"] = strings.TrimSpace(strings.ReplaceAll(matches[2], ",", ""))
	}
	
	// Extract Net Tangible Assets data
	re = regexp.MustCompile(financialsRegexConfig.BalanceSheet.NetTangibleAssets)
	matches = re.FindStringSubmatch(html)
	if len(matches) > 2 {
		financialData["Current_NetTangibleAssets"] = strings.TrimSpace(strings.ReplaceAll(matches[1], ",", ""))
		financialData["2024_NetTangibleAssets"] = strings.TrimSpace(strings.ReplaceAll(matches[2], ",", ""))
	}
	
	// Extract Working Capital data
	re = regexp.MustCompile(financialsRegexConfig.BalanceSheet.WorkingCapital)
	matches = re.FindStringSubmatch(html)
	if len(matches) > 2 {
		financialData["Current_WorkingCapital"] = strings.TrimSpace(strings.ReplaceAll(matches[1], ",", ""))
		financialData["2024_WorkingCapital"] = strings.TrimSpace(strings.ReplaceAll(matches[2], ",", ""))
	}
	
	// Extract Invested Capital data
	re = regexp.MustCompile(financialsRegexConfig.BalanceSheet.InvestedCapital)
	matches = re.FindStringSubmatch(html)
	if len(matches) > 2 {
		financialData["Current_InvestedCapital"] = strings.TrimSpace(strings.ReplaceAll(matches[1], ",", ""))
		financialData["2024_InvestedCapital"] = strings.TrimSpace(strings.ReplaceAll(matches[2], ",", ""))
	}
	
	// Extract Tangible Book Value data
	re = regexp.MustCompile(financialsRegexConfig.BalanceSheet.TangibleBookValue)
	matches = re.FindStringSubmatch(html)
	if len(matches) > 2 {
		financialData["Current_TangibleBookValue"] = strings.TrimSpace(strings.ReplaceAll(matches[1], ",", ""))
		financialData["2024_TangibleBookValue"] = strings.TrimSpace(strings.ReplaceAll(matches[2], ",", ""))
	}
	
	// Extract Total Debt data
	re = regexp.MustCompile(financialsRegexConfig.BalanceSheet.TotalDebt)
	matches = re.FindStringSubmatch(html)
	if len(matches) > 2 {
		financialData["Current_TotalDebt"] = strings.TrimSpace(strings.ReplaceAll(matches[1], ",", ""))
		financialData["2024_TotalDebt"] = strings.TrimSpace(strings.ReplaceAll(matches[2], ",", ""))
	}
	
	// Extract Share Issued data
	re = regexp.MustCompile(financialsRegexConfig.BalanceSheet.ShareIssued)
	matches = re.FindStringSubmatch(html)
	if len(matches) > 2 {
		financialData["Current_ShareIssued"] = strings.TrimSpace(strings.ReplaceAll(matches[1], ",", ""))
		financialData["2024_ShareIssued"] = strings.TrimSpace(strings.ReplaceAll(matches[2], ",", ""))
	}
	
	// Cash Flow extraction patterns
	// Extract Operating Cash Flow data
	re = regexp.MustCompile(financialsRegexConfig.CashFlow.OperatingCashFlow)
	matches = re.FindStringSubmatch(html)
	if len(matches) > 2 {
		financialData["Current_OperatingCashFlow"] = strings.TrimSpace(strings.ReplaceAll(matches[1], ",", ""))
		financialData["2024_OperatingCashFlow"] = strings.TrimSpace(strings.ReplaceAll(matches[2], ",", ""))
	}
	
	// Extract Investing Cash Flow data
	re = regexp.MustCompile(financialsRegexConfig.CashFlow.InvestingCashFlow)
	matches = re.FindStringSubmatch(html)
	if len(matches) > 2 {
		financialData["Current_InvestingCashFlow"] = strings.TrimSpace(strings.ReplaceAll(matches[1], ",", ""))
		financialData["2024_InvestingCashFlow"] = strings.TrimSpace(strings.ReplaceAll(matches[2], ",", ""))
	}
	
	// Extract Financing Cash Flow data
	re = regexp.MustCompile(financialsRegexConfig.CashFlow.FinancingCashFlow)
	matches = re.FindStringSubmatch(html)
	if len(matches) > 2 {
		financialData["Current_FinancingCashFlow"] = strings.TrimSpace(strings.ReplaceAll(matches[1], ",", ""))
		financialData["2024_FinancingCashFlow"] = strings.TrimSpace(strings.ReplaceAll(matches[2], ",", ""))
	}
	
	// Extract End Cash Position data
	re = regexp.MustCompile(financialsRegexConfig.CashFlow.EndCashPosition)
	matches = re.FindStringSubmatch(html)
	if len(matches) > 2 {
		financialData["Current_EndCashPosition"] = strings.TrimSpace(strings.ReplaceAll(matches[1], ",", ""))
		financialData["2024_EndCashPosition"] = strings.TrimSpace(strings.ReplaceAll(matches[2], ",", ""))
	}
	
	// Extract Capital Expenditure data
	re = regexp.MustCompile(financialsRegexConfig.CashFlow.CapitalExpenditure)
	matches = re.FindStringSubmatch(html)
	if len(matches) > 2 {
		financialData["Current_CapitalExpenditure"] = strings.TrimSpace(strings.ReplaceAll(matches[1], ",", ""))
		financialData["2024_CapitalExpenditure"] = strings.TrimSpace(strings.ReplaceAll(matches[2], ",", ""))
	}
	
	// Extract Issuance of Debt data
	re = regexp.MustCompile(financialsRegexConfig.CashFlow.IssuanceOfDebt)
	matches = re.FindStringSubmatch(html)
	if len(matches) > 2 {
		financialData["Current_IssuanceOfDebt"] = strings.TrimSpace(strings.ReplaceAll(matches[1], ",", ""))
		financialData["2024_IssuanceOfDebt"] = strings.TrimSpace(strings.ReplaceAll(matches[2], ",", ""))
	}
	
	// Extract Repayment of Debt data
	re = regexp.MustCompile(financialsRegexConfig.CashFlow.RepaymentOfDebt)
	matches = re.FindStringSubmatch(html)
	if len(matches) > 2 {
		financialData["Current_RepaymentOfDebt"] = strings.TrimSpace(strings.ReplaceAll(matches[1], ",", ""))
		financialData["2024_RepaymentOfDebt"] = strings.TrimSpace(strings.ReplaceAll(matches[2], ",", ""))
	}
	
	// Extract Repurchase of Capital Stock data
	re = regexp.MustCompile(financialsRegexConfig.CashFlow.RepurchaseOfCapitalStock)
	matches = re.FindStringSubmatch(html)
	if len(matches) > 2 {
		financialData["Current_RepurchaseOfCapitalStock"] = strings.TrimSpace(strings.ReplaceAll(matches[1], ",", ""))
		financialData["2024_RepurchaseOfCapitalStock"] = strings.TrimSpace(strings.ReplaceAll(matches[2], ",", ""))
	}
	
	// Extract Free Cash Flow data
	re = regexp.MustCompile(financialsRegexConfig.CashFlow.FreeCashFlow)
	matches = re.FindStringSubmatch(html)
	if len(matches) > 2 {
		financialData["Current_FreeCashFlow"] = strings.TrimSpace(strings.ReplaceAll(matches[1], ",", ""))
		financialData["2024_FreeCashFlow"] = strings.TrimSpace(strings.ReplaceAll(matches[2], ",", ""))
	}
	
	if len(financialData) == 0 {
		return nil, fmt.Errorf("could not find financial data in HTML table")
	}
	
	return financialData, nil
}

// populateDTOFromHTMLData populates the DTO with data extracted from HTML table
func populateDTOFromHTMLData(financialData map[string]string, dto *ComprehensiveFinancialsDTO) {
	// Set currency from extracted data
	if currency, exists := financialData["Currency"]; exists {
		dto.Currency = currency
	}
	
	// Helper function to convert string to Scaled (multiply by 1000 for thousands)
	convertToScaled := func(value string) *Scaled {
		if value == "" || value == "--" {
			return nil
		}
		// Remove commas and convert to int64, then multiply by 1000 for thousands
		cleanValue := strings.ReplaceAll(value, ",", "")
		if val, err := strconv.ParseInt(cleanValue, 10, 64); err == nil {
			return &Scaled{Scaled: val * 1000, Scale: 0}
		}
		return nil
	}
	
	// Helper function to convert EPS string to Scaled
	convertEPSToScaled := func(value string) *Scaled {
		if value == "" || value == "--" {
			return nil
		}
		// Handle Korean Won values with 'k' suffix (thousands)
		if strings.HasSuffix(value, "k") {
			cleanValue := strings.TrimSuffix(value, "k")
			if val, err := strconv.ParseFloat(cleanValue, 64); err == nil {
				// Convert to actual value (multiply by 1000, then by 100 for cents)
				return &Scaled{Scaled: int64(val * 1000 * 100), Scale: 2}
			}
		} else if val, err := strconv.ParseFloat(value, 64); err == nil {
			// Convert to cents (multiply by 100)
			return &Scaled{Scaled: int64(val * 100), Scale: 2}
		}
		return nil
	}
	
	// Helper function to convert shares string to int64
	convertSharesToInt64 := func(value string) *int64 {
		if value == "" || value == "--" {
			return nil
		}
		// Remove commas for parsing
		cleanValue := strings.ReplaceAll(value, ",", "")
		// Try parsing as float first (to handle decimals), then convert to int64
		if val, err := strconv.ParseFloat(cleanValue, 64); err == nil {
			result := int64(val)
			return &result
		}
		return nil
	}
	
	// Populate current (TTM) data
	if val, exists := financialData["TTM_TotalRevenue"]; exists {
		dto.Current.TotalRevenue = convertToScaled(val)
	}
	if val, exists := financialData["TTM_CostOfRevenue"]; exists {
		dto.Current.CostOfRevenue = convertToScaled(val)
	}
	if val, exists := financialData["TTM_OperatingIncome"]; exists {
		dto.Current.OperatingIncome = convertToScaled(val)
	}
	if val, exists := financialData["TTM_NetIncome"]; exists {
		dto.Current.NetIncomeCommonStockholders = convertToScaled(val)
	}
	if val, exists := financialData["TTM_BasicEPS"]; exists {
		dto.Current.BasicEPS = convertEPSToScaled(val)
	}
	if val, exists := financialData["TTM_DilutedEPS"]; exists {
		dto.Current.DilutedEPS = convertEPSToScaled(val)
	}
	if val, exists := financialData["TTM_BasicAverageShares"]; exists {
		dto.Current.BasicAverageShares = convertSharesToInt64(val)
	}
	if val, exists := financialData["TTM_DilutedAverageShares"]; exists {
		dto.Current.DilutedAverageShares = convertSharesToInt64(val)
	}
	if val, exists := financialData["TTM_TotalExpenses"]; exists {
		dto.Current.TotalExpenses = convertToScaled(val)
	}
	if val, exists := financialData["TTM_EBIT"]; exists {
		dto.Current.EBIT = convertToScaled(val)
	}
	if val, exists := financialData["TTM_EBITDA"]; exists {
		dto.Current.EBITDA = convertToScaled(val)
	}
	if val, exists := financialData["TTM_NormalizedEBITDA"]; exists {
		dto.Current.NormalizedEBITDA = convertToScaled(val)
	}
	
	// Balance Sheet data population
	if val, exists := financialData["Current_TotalAssets"]; exists {
		dto.Current.TotalAssets = convertToScaled(val)
	}
	if val, exists := financialData["Current_TotalCapitalization"]; exists {
		dto.Current.TotalCapitalization = convertToScaled(val)
	}
	if val, exists := financialData["Current_CommonStockEquity"]; exists {
		dto.Current.CommonStockEquity = convertToScaled(val)
	}
	if val, exists := financialData["Current_CapitalLeaseObligations"]; exists {
		dto.Current.CapitalLeaseObligations = convertToScaled(val)
	}
	if val, exists := financialData["Current_NetTangibleAssets"]; exists {
		dto.Current.NetTangibleAssets = convertToScaled(val)
	}
	if val, exists := financialData["Current_WorkingCapital"]; exists {
		dto.Current.WorkingCapital = convertToScaled(val)
	}
	if val, exists := financialData["Current_InvestedCapital"]; exists {
		dto.Current.InvestedCapital = convertToScaled(val)
	}
	if val, exists := financialData["Current_TangibleBookValue"]; exists {
		dto.Current.TangibleBookValue = convertToScaled(val)
	}
	if val, exists := financialData["Current_TotalDebt"]; exists {
		dto.Current.TotalDebt = convertToScaled(val)
	}
	if val, exists := financialData["Current_ShareIssued"]; exists {
		dto.Current.ShareIssued = convertSharesToInt64(val)
	}
	
	// Cash Flow data population
	if val, exists := financialData["Current_OperatingCashFlow"]; exists {
		dto.Current.OperatingCashFlow = convertToScaled(val)
	}
	if val, exists := financialData["Current_InvestingCashFlow"]; exists {
		dto.Current.InvestingCashFlow = convertToScaled(val)
	}
	if val, exists := financialData["Current_FinancingCashFlow"]; exists {
		dto.Current.FinancingCashFlow = convertToScaled(val)
	}
	if val, exists := financialData["Current_EndCashPosition"]; exists {
		dto.Current.EndCashPosition = convertToScaled(val)
	}
	if val, exists := financialData["Current_CapitalExpenditure"]; exists {
		dto.Current.CapitalExpenditure = convertToScaled(val)
	}
	if val, exists := financialData["Current_IssuanceOfDebt"]; exists {
		dto.Current.IssuanceOfDebt = convertToScaled(val)
	}
	if val, exists := financialData["Current_RepaymentOfDebt"]; exists {
		dto.Current.RepaymentOfDebt = convertToScaled(val)
	}
	if val, exists := financialData["Current_RepurchaseOfCapitalStock"]; exists {
		dto.Current.RepurchaseOfCapitalStock = convertToScaled(val)
	}
	if val, exists := financialData["Current_FreeCashFlow"]; exists {
		dto.Current.FreeCashFlow = convertToScaled(val)
	}
	
	// Populate historical (2024) data
	if val, exists := financialData["2024_TotalRevenue"]; exists {
		dto.Historical.Q4_2024.TotalRevenue = convertToScaled(val)
	}
	if val, exists := financialData["2024_CostOfRevenue"]; exists {
		dto.Historical.Q4_2024.CostOfRevenue = convertToScaled(val)
	}
	if val, exists := financialData["2024_OperatingIncome"]; exists {
		dto.Historical.Q4_2024.OperatingIncome = convertToScaled(val)
	}
	if val, exists := financialData["2024_NetIncome"]; exists {
		dto.Historical.Q4_2024.NetIncomeCommonStockholders = convertToScaled(val)
	}
	if val, exists := financialData["2024_BasicEPS"]; exists {
		dto.Historical.Q4_2024.BasicEPS = convertEPSToScaled(val)
	}
	if val, exists := financialData["2024_DilutedEPS"]; exists {
		dto.Historical.Q4_2024.DilutedEPS = convertEPSToScaled(val)
	}
	if val, exists := financialData["2024_BasicAverageShares"]; exists {
		dto.Historical.Q4_2024.BasicAverageShares = convertSharesToInt64(val)
	}
	if val, exists := financialData["2024_DilutedAverageShares"]; exists {
		dto.Historical.Q4_2024.DilutedAverageShares = convertSharesToInt64(val)
	}
	if val, exists := financialData["2024_TotalExpenses"]; exists {
		dto.Historical.Q4_2024.TotalExpenses = convertToScaled(val)
	}
	if val, exists := financialData["2024_EBIT"]; exists {
		dto.Historical.Q4_2024.EBIT = convertToScaled(val)
	}
	if val, exists := financialData["2024_EBITDA"]; exists {
		dto.Historical.Q4_2024.EBITDA = convertToScaled(val)
	}
	if val, exists := financialData["2024_NormalizedEBITDA"]; exists {
		dto.Historical.Q4_2024.NormalizedEBITDA = convertToScaled(val)
	}
}

// extractCurrentFromJSON extracts current (trailing) financial values from JSON data
func extractCurrentFromJSON(financialData interface{}, dto *ComprehensiveFinancialsDTO) {
	// Convert interface{} to the actual financial data structure
	fd, ok := financialData.(struct {
		TrailingTotalRevenue []FinancialDataPoint `json:"trailingTotalRevenue"`
		AnnualTotalRevenue   []FinancialDataPoint `json:"annualTotalRevenue"`
		TrailingOperatingIncome []FinancialDataPoint `json:"trailingOperatingIncome"`
		AnnualOperatingIncome   []FinancialDataPoint `json:"annualOperatingIncome"`
		TrailingNetIncome []FinancialDataPoint `json:"trailingNetIncome"`
		AnnualNetIncome   []FinancialDataPoint `json:"annualNetIncome"`
		TrailingBasicEPS []FinancialDataPoint `json:"trailingBasicEPS"`
		AnnualBasicEPS   []FinancialDataPoint `json:"annualBasicEPS"`
		TrailingDilutedEPS []FinancialDataPoint `json:"trailingDilutedEPS"`
		AnnualDilutedEPS   []FinancialDataPoint `json:"annualDilutedEPS"`
		TrailingEBITDA []FinancialDataPoint `json:"trailingEBITDA"`
		AnnualEBITDA   []FinancialDataPoint `json:"annualEBITDA"`
		TrailingGrossProfit []FinancialDataPoint `json:"trailingGrossProfit"`
		AnnualGrossProfit   []FinancialDataPoint `json:"annualGrossProfit"`
		TrailingCostOfRevenue []FinancialDataPoint `json:"trailingCostOfRevenue"`
		AnnualCostOfRevenue   []FinancialDataPoint `json:"annualCostOfRevenue"`
		TrailingOperatingExpense []FinancialDataPoint `json:"trailingOperatingExpense"`
		AnnualOperatingExpense   []FinancialDataPoint `json:"annualOperatingExpense"`
		TrailingTotalExpenses []FinancialDataPoint `json:"trailingTotalExpenses"`
		AnnualTotalExpenses   []FinancialDataPoint `json:"annualTotalExpenses"`
		TrailingTaxProvision []FinancialDataPoint `json:"trailingTaxProvision"`
		AnnualTaxProvision   []FinancialDataPoint `json:"annualTaxProvision"`
		TrailingPretaxIncome []FinancialDataPoint `json:"trailingPretaxIncome"`
		AnnualPretaxIncome   []FinancialDataPoint `json:"annualPretaxIncome"`
		TrailingOtherIncomeExpense []FinancialDataPoint `json:"trailingOtherIncomeExpense"`
		AnnualOtherIncomeExpense   []FinancialDataPoint `json:"annualOtherIncomeExpense"`
		TrailingNetNonOperatingInterestIncomeExpense []FinancialDataPoint `json:"trailingNetNonOperatingInterestIncomeExpense"`
		AnnualNetNonOperatingInterestIncomeExpense   []FinancialDataPoint `json:"annualNetNonOperatingInterestIncomeExpense"`
		TrailingBasicAverageShares []FinancialDataPoint `json:"trailingBasicAverageShares"`
		AnnualBasicAverageShares   []FinancialDataPoint `json:"annualBasicAverageShares"`
		TrailingDilutedAverageShares []FinancialDataPoint `json:"trailingDilutedAverageShares"`
		AnnualDilutedAverageShares   []FinancialDataPoint `json:"annualDilutedAverageShares"`
		TrailingEBIT []FinancialDataPoint `json:"trailingEBIT"`
		AnnualEBIT   []FinancialDataPoint `json:"annualEBIT"`
		TrailingNormalizedIncome []FinancialDataPoint `json:"trailingNormalizedIncome"`
		AnnualNormalizedIncome   []FinancialDataPoint `json:"annualNormalizedIncome"`
		TrailingNormalizedEBITDA []FinancialDataPoint `json:"trailingNormalizedEBITDA"`
		AnnualNormalizedEBITDA   []FinancialDataPoint `json:"annualNormalizedEBITDA"`
		TrailingReconciledCostOfRevenue []FinancialDataPoint `json:"trailingReconciledCostOfRevenue"`
		AnnualReconciledCostOfRevenue   []FinancialDataPoint `json:"annualReconciledCostOfRevenue"`
		TrailingReconciledDepreciation []FinancialDataPoint `json:"trailingReconciledDepreciation"`
		AnnualReconciledDepreciation   []FinancialDataPoint `json:"annualReconciledDepreciation"`
	})
	
	if !ok {
		return
	}
	
	// Extract the most recent trailing values (current values)
	if len(fd.TrailingTotalRevenue) > 0 {
		dto.Current.TotalRevenue = convertToScaled(fd.TrailingTotalRevenue[0])
	}
	if len(fd.TrailingOperatingIncome) > 0 {
		dto.Current.OperatingIncome = convertToScaled(fd.TrailingOperatingIncome[0])
	}
	if len(fd.TrailingNetIncome) > 0 {
		dto.Current.NetIncomeCommonStockholders = convertToScaled(fd.TrailingNetIncome[0])
	}
	if len(fd.TrailingBasicEPS) > 0 {
		dto.Current.BasicEPS = convertToScaled(fd.TrailingBasicEPS[0])
	}
	if len(fd.TrailingDilutedEPS) > 0 {
		dto.Current.DilutedEPS = convertToScaled(fd.TrailingDilutedEPS[0])
	}
	if len(fd.TrailingEBITDA) > 0 {
		dto.Current.EBITDA = convertToScaled(fd.TrailingEBITDA[0])
	}
	if len(fd.TrailingGrossProfit) > 0 {
		dto.Current.GrossProfit = convertToScaled(fd.TrailingGrossProfit[0])
	}
	if len(fd.TrailingCostOfRevenue) > 0 {
		dto.Current.CostOfRevenue = convertToScaled(fd.TrailingCostOfRevenue[0])
	}
	if len(fd.TrailingOperatingExpense) > 0 {
		dto.Current.OperatingExpense = convertToScaled(fd.TrailingOperatingExpense[0])
	}
	if len(fd.TrailingTotalExpenses) > 0 {
		dto.Current.TotalExpenses = convertToScaled(fd.TrailingTotalExpenses[0])
	}
	if len(fd.TrailingTaxProvision) > 0 {
		dto.Current.TaxProvision = convertToScaled(fd.TrailingTaxProvision[0])
	}
	if len(fd.TrailingPretaxIncome) > 0 {
		dto.Current.PretaxIncome = convertToScaled(fd.TrailingPretaxIncome[0])
	}
	if len(fd.TrailingOtherIncomeExpense) > 0 {
		dto.Current.OtherIncomeExpense = convertToScaled(fd.TrailingOtherIncomeExpense[0])
	}
	if len(fd.TrailingNetNonOperatingInterestIncomeExpense) > 0 {
		dto.Current.NetNonOperatingInterestIncomeExpense = convertToScaled(fd.TrailingNetNonOperatingInterestIncomeExpense[0])
	}
	if len(fd.TrailingBasicAverageShares) > 0 {
		dto.Current.BasicAverageShares = convertToInt64(fd.TrailingBasicAverageShares[0])
	}
	if len(fd.TrailingDilutedAverageShares) > 0 {
		dto.Current.DilutedAverageShares = convertToInt64(fd.TrailingDilutedAverageShares[0])
	}
	if len(fd.TrailingEBIT) > 0 {
		dto.Current.EBIT = convertToScaled(fd.TrailingEBIT[0])
	}
	if len(fd.TrailingNormalizedIncome) > 0 {
		dto.Current.NormalizedIncome = convertToScaled(fd.TrailingNormalizedIncome[0])
	}
	if len(fd.TrailingNormalizedEBITDA) > 0 {
		dto.Current.NormalizedEBITDA = convertToScaled(fd.TrailingNormalizedEBITDA[0])
	}
	if len(fd.TrailingReconciledCostOfRevenue) > 0 {
		dto.Current.ReconciledCostOfRevenue = convertToScaled(fd.TrailingReconciledCostOfRevenue[0])
	}
	if len(fd.TrailingReconciledDepreciation) > 0 {
		dto.Current.ReconciledDepreciation = convertToScaled(fd.TrailingReconciledDepreciation[0])
	}
}

// extractHistoricalFromJSON extracts historical (annual) financial values from JSON data
func extractHistoricalFromJSON(financialData interface{}, dto *ComprehensiveFinancialsDTO) {
	// For now, we'll implement a simplified version that extracts the most recent annual data
	// This can be expanded to extract multiple years of historical data
	
	// Convert interface{} to the actual financial data structure
	fd, ok := financialData.(struct {
		TrailingTotalRevenue []FinancialDataPoint `json:"trailingTotalRevenue"`
		AnnualTotalRevenue   []FinancialDataPoint `json:"annualTotalRevenue"`
		TrailingOperatingIncome []FinancialDataPoint `json:"trailingOperatingIncome"`
		AnnualOperatingIncome   []FinancialDataPoint `json:"annualOperatingIncome"`
		TrailingNetIncome []FinancialDataPoint `json:"trailingNetIncome"`
		AnnualNetIncome   []FinancialDataPoint `json:"annualNetIncome"`
		TrailingBasicEPS []FinancialDataPoint `json:"trailingBasicEPS"`
		AnnualBasicEPS   []FinancialDataPoint `json:"annualBasicEPS"`
		TrailingDilutedEPS []FinancialDataPoint `json:"trailingDilutedEPS"`
		AnnualDilutedEPS   []FinancialDataPoint `json:"annualDilutedEPS"`
		TrailingEBITDA []FinancialDataPoint `json:"trailingEBITDA"`
		AnnualEBITDA   []FinancialDataPoint `json:"annualEBITDA"`
		TrailingGrossProfit []FinancialDataPoint `json:"trailingGrossProfit"`
		AnnualGrossProfit   []FinancialDataPoint `json:"annualGrossProfit"`
		TrailingCostOfRevenue []FinancialDataPoint `json:"trailingCostOfRevenue"`
		AnnualCostOfRevenue   []FinancialDataPoint `json:"annualCostOfRevenue"`
		TrailingOperatingExpense []FinancialDataPoint `json:"trailingOperatingExpense"`
		AnnualOperatingExpense   []FinancialDataPoint `json:"annualOperatingExpense"`
		TrailingTotalExpenses []FinancialDataPoint `json:"trailingTotalExpenses"`
		AnnualTotalExpenses   []FinancialDataPoint `json:"annualTotalExpenses"`
		TrailingTaxProvision []FinancialDataPoint `json:"trailingTaxProvision"`
		AnnualTaxProvision   []FinancialDataPoint `json:"annualTaxProvision"`
		TrailingPretaxIncome []FinancialDataPoint `json:"trailingPretaxIncome"`
		AnnualPretaxIncome   []FinancialDataPoint `json:"annualPretaxIncome"`
		TrailingOtherIncomeExpense []FinancialDataPoint `json:"trailingOtherIncomeExpense"`
		AnnualOtherIncomeExpense   []FinancialDataPoint `json:"annualOtherIncomeExpense"`
		TrailingNetNonOperatingInterestIncomeExpense []FinancialDataPoint `json:"trailingNetNonOperatingInterestIncomeExpense"`
		AnnualNetNonOperatingInterestIncomeExpense   []FinancialDataPoint `json:"annualNetNonOperatingInterestIncomeExpense"`
		TrailingBasicAverageShares []FinancialDataPoint `json:"trailingBasicAverageShares"`
		AnnualBasicAverageShares   []FinancialDataPoint `json:"annualBasicAverageShares"`
		TrailingDilutedAverageShares []FinancialDataPoint `json:"trailingDilutedAverageShares"`
		AnnualDilutedAverageShares   []FinancialDataPoint `json:"annualDilutedAverageShares"`
		TrailingEBIT []FinancialDataPoint `json:"trailingEBIT"`
		AnnualEBIT   []FinancialDataPoint `json:"annualEBIT"`
		TrailingNormalizedIncome []FinancialDataPoint `json:"trailingNormalizedIncome"`
		AnnualNormalizedIncome   []FinancialDataPoint `json:"annualNormalizedIncome"`
		TrailingNormalizedEBITDA []FinancialDataPoint `json:"trailingNormalizedEBITDA"`
		AnnualNormalizedEBITDA   []FinancialDataPoint `json:"annualNormalizedEBITDA"`
		TrailingReconciledCostOfRevenue []FinancialDataPoint `json:"trailingReconciledCostOfRevenue"`
		AnnualReconciledCostOfRevenue   []FinancialDataPoint `json:"annualReconciledCostOfRevenue"`
		TrailingReconciledDepreciation []FinancialDataPoint `json:"trailingReconciledDepreciation"`
		AnnualReconciledDepreciation   []FinancialDataPoint `json:"annualReconciledDepreciation"`
	})
	
	if !ok {
		return
	}
	
	// Extract the most recent annual data for Q4 2024 (most recent full year)
	if len(fd.AnnualTotalRevenue) > 0 {
		dto.Historical.Q4_2024.TotalRevenue = convertToScaled(fd.AnnualTotalRevenue[0])
	}
	if len(fd.AnnualOperatingIncome) > 0 {
		dto.Historical.Q4_2024.OperatingIncome = convertToScaled(fd.AnnualOperatingIncome[0])
	}
	if len(fd.AnnualNetIncome) > 0 {
		dto.Historical.Q4_2024.NetIncomeCommonStockholders = convertToScaled(fd.AnnualNetIncome[0])
	}
	if len(fd.AnnualBasicEPS) > 0 {
		dto.Historical.Q4_2024.BasicEPS = convertToScaled(fd.AnnualBasicEPS[0])
	}
	if len(fd.AnnualDilutedEPS) > 0 {
		dto.Historical.Q4_2024.DilutedEPS = convertToScaled(fd.AnnualDilutedEPS[0])
	}
	if len(fd.AnnualEBITDA) > 0 {
		dto.Historical.Q4_2024.EBITDA = convertToScaled(fd.AnnualEBITDA[0])
	}
}

// Helper function to convert FinancialDataPoint to Scaled value
func convertToScaled(dataPoint FinancialDataPoint) *Scaled {
	if dataPoint.ReportedValue.Raw == 0 {
		return nil
	}
	
	// For financial data, we typically use scale 0 (no decimal places)
	// since the raw values are already in the appropriate units
	scaled := int64(dataPoint.ReportedValue.Raw)
	
	return &Scaled{
		Scaled: scaled,
		Scale:  0,
	}
}

// Helper function to convert FinancialDataPoint to int64 for shares
func convertToInt64(dataPoint FinancialDataPoint) *int64 {
	if dataPoint.ReportedValue.Raw == 0 {
		return nil
	}
	
	value := int64(dataPoint.ReportedValue.Raw)
	return &value
}




