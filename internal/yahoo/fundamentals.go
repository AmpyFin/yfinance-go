package yahoo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
)

// FundamentalsResponse represents the Yahoo Finance fundamentals API response
type FundamentalsResponse struct {
	QuoteSummary QuoteSummary `json:"quoteSummary"`
}

// QuoteSummary contains the fundamentals data
type QuoteSummary struct {
	Result []FundamentalsResult `json:"result"`
	Error  *string              `json:"error"`
}

// FundamentalsResult contains fundamentals data for a single symbol
type FundamentalsResult struct {
	IncomeStatementHistoryQuarterly *IncomeStatementHistory `json:"incomeStatementHistoryQuarterly"`
	BalanceSheetHistoryQuarterly    *BalanceSheetHistory    `json:"balanceSheetHistoryQuarterly"`
	CashflowStatementHistoryQuarterly *CashflowStatementHistory `json:"cashflowStatementHistoryQuarterly"`
}

// IncomeStatementHistory contains quarterly income statement data
type IncomeStatementHistory struct {
	IncomeStatementHistory []IncomeStatement `json:"incomeStatementHistory"`
}

// BalanceSheetHistory contains quarterly balance sheet data
type BalanceSheetHistory struct {
	BalanceSheetHistory []BalanceSheet `json:"balanceSheetHistory"`
}

// CashflowStatementHistory contains quarterly cashflow statement data
type CashflowStatementHistory struct {
	CashflowStatementHistory []CashflowStatement `json:"cashflowStatementHistory"`
}

// IncomeStatement represents a single income statement
type IncomeStatement struct {
	MaxAge                        int64   `json:"maxAge"`
	EndDate                       DateValue `json:"endDate"`
	TotalRevenue                  *Value  `json:"totalRevenue"`
	CostOfRevenue                 *Value  `json:"costOfRevenue"`
	GrossProfit                   *Value  `json:"grossProfit"`
	ResearchDevelopment           *Value  `json:"researchDevelopment"`
	SellingGeneralAdministrative  *Value  `json:"sellingGeneralAdministrative"`
	TotalOperatingExpenses        *Value  `json:"totalOperatingExpenses"`
	OperatingIncome               *Value  `json:"operatingIncome"`
	TotalOtherIncomeExpenseNet    *Value  `json:"totalOtherIncomeExpenseNet"`
	EBIT                          *Value  `json:"ebit"`
	InterestExpense               *Value  `json:"interestExpense"`
	IncomeBeforeTax               *Value  `json:"incomeBeforeTax"`
	IncomeTaxExpense              *Value  `json:"incomeTaxExpense"`
	NetIncome                     *Value  `json:"netIncome"`
	NetIncomeCommonStockholders   *Value  `json:"netIncomeCommonStockholders"`
	EPS                           *Value  `json:"eps"`
	EPSDiluted                    *Value  `json:"epsDiluted"`
	WeightedAverageShares         *Value  `json:"weightedAverageShares"`
	WeightedAverageSharesDiluted  *Value  `json:"weightedAverageSharesDiluted"`
}

// BalanceSheet represents a single balance sheet
type BalanceSheet struct {
	MaxAge                        int64   `json:"maxAge"`
	EndDate                       DateValue `json:"endDate"`
	Cash                          *Value  `json:"cash"`
	ShortTermInvestments          *Value  `json:"shortTermInvestments"`
	NetReceivables                *Value  `json:"netReceivables"`
	Inventory                     *Value  `json:"inventory"`
	OtherCurrentAssets            *Value  `json:"otherCurrentAssets"`
	TotalCurrentAssets            *Value  `json:"totalCurrentAssets"`
	LongTermInvestments           *Value  `json:"longTermInvestments"`
	PropertyPlantEquipment        *Value  `json:"propertyPlantEquipment"`
	OtherAssets                   *Value  `json:"otherAssets"`
	TotalAssets                   *Value  `json:"totalAssets"`
	AccountsPayable               *Value  `json:"accountsPayable"`
	ShortLongTermDebt             *Value  `json:"shortLongTermDebt"`
	OtherCurrentLiab              *Value  `json:"otherCurrentLiab"`
	LongTermDebt                  *Value  `json:"longTermDebt"`
	OtherLiab                     *Value  `json:"otherLiab"`
	TotalCurrentLiabilities       *Value  `json:"totalCurrentLiabilities"`
	TotalLiab                     *Value  `json:"totalLiab"`
	CommonStock                   *Value  `json:"commonStock"`
	RetainedEarnings              *Value  `json:"retainedEarnings"`
	TreasuryStock                 *Value  `json:"treasuryStock"`
	OtherStockholderEquity        *Value  `json:"otherStockholderEquity"`
	TotalStockholderEquity        *Value  `json:"totalStockholderEquity"`
	NetTangibleAssets             *Value  `json:"netTangibleAssets"`
}

// CashflowStatement represents a single cashflow statement
type CashflowStatement struct {
	MaxAge                        int64   `json:"maxAge"`
	EndDate                       DateValue `json:"endDate"`
	Investments                   *Value  `json:"investments"`
	ChangeToLiabilities           *Value  `json:"changeToLiabilities"`
	TotalCashflowsFromInvestingActivities *Value `json:"totalCashflowsFromInvestingActivities"`
	NetBorrowings                 *Value  `json:"netBorrowings"`
	TotalCashFromFinancingActivities *Value `json:"totalCashFromFinancingActivities"`
	ChangeToOperatingActivities   *Value  `json:"changeToOperatingActivities"`
	NetIncome                     *Value  `json:"netIncome"`
	ChangeInCash                  *Value  `json:"changeInCash"`
	BeginPeriodCashFlow           *Value  `json:"beginPeriodCashFlow"`
	EndPeriodCashFlow             *Value  `json:"endPeriodCashFlow"`
	TotalCashFromOperatingActivities *Value `json:"totalCashFromOperatingActivities"`
	Depreciation                  *Value  `json:"depreciation"`
	OtherCashflowsFromInvestingActivities *Value `json:"otherCashflowsFromInvestingActivities"`
	DividendsPaid                 *Value  `json:"dividendsPaid"`
	ChangeToInventory             *Value  `json:"changeToInventory"`
	ChangeToAccountReceivables    *Value  `json:"changeToAccountReceivables"`
	SalePurchaseOfStock           *Value  `json:"salePurchaseOfStock"`
	OtherCashflowsFromFinancingActivities *Value `json:"otherCashflowsFromFinancingActivities"`
	ChangeToNetincome             *Value  `json:"changeToNetincome"`
	CapitalExpenditures           *Value  `json:"capitalExpenditures"`
	ChangeReceivables             *Value  `json:"changeReceivables"`
	CashFlowsOtherOperating       *Value  `json:"cashFlowsOtherOperating"`
	ExchangeRateChanges           *Value  `json:"exchangeRateChanges"`
	CashAndCashEquivalentsChanges *Value  `json:"cashAndCashEquivalentsChanges"`
	ChangeInWorkingCapital        *Value  `json:"changeInWorkingCapital"`
}

// DateValue represents a date with raw timestamp and formatted string
type DateValue struct {
	Raw int64  `json:"raw"`
	Fmt string `json:"fmt"`
}

// Value represents a financial value with raw number and formatted string
type Value struct {
	Raw     *int64  `json:"raw"`
	Fmt     *string `json:"fmt"`
	LongFmt *string `json:"longFmt"`
}

// DecodeFundamentalsResponse decodes a Yahoo Finance fundamentals response with strict validation
func DecodeFundamentalsResponse(data []byte) (*FundamentalsResponse, error) {
	var response FundamentalsResponse
	
	// Use strict JSON decoding
	decoder := json.NewDecoder(bytes.NewReader(data))
	// Allow unknown fields for fundamentals as the response has many fields we don't use
	// decoder.DisallowUnknownFields()
	
	if err := decoder.Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode fundamentals response: %w", err)
	}
	
	// Validate response structure
	if err := response.Validate(); err != nil {
		return nil, fmt.Errorf("invalid fundamentals response: %w", err)
	}
	
	return &response, nil
}

// Validate validates the fundamentals response structure
func (r *FundamentalsResponse) Validate() error {
	if r.QuoteSummary.Error != nil {
		return fmt.Errorf("yahoo finance error: %s", *r.QuoteSummary.Error)
	}
	
	if len(r.QuoteSummary.Result) == 0 {
		return fmt.Errorf("no fundamentals results found")
	}
	
	for i, result := range r.QuoteSummary.Result {
		if err := result.Validate(); err != nil {
			return fmt.Errorf("result[%d]: %w", i, err)
		}
	}
	
	return nil
}

// Validate validates a fundamentals result
func (r *FundamentalsResult) Validate() error {
	// At least one statement type should be present
	if r.IncomeStatementHistoryQuarterly == nil &&
		r.BalanceSheetHistoryQuarterly == nil &&
		r.CashflowStatementHistoryQuarterly == nil {
		return fmt.Errorf("no financial statements found")
	}
	
	// Validate income statement if present
	if r.IncomeStatementHistoryQuarterly != nil {
		if err := r.IncomeStatementHistoryQuarterly.Validate(); err != nil {
			return fmt.Errorf("income statement: %w", err)
		}
	}
	
	// Validate balance sheet if present
	if r.BalanceSheetHistoryQuarterly != nil {
		if err := r.BalanceSheetHistoryQuarterly.Validate(); err != nil {
			return fmt.Errorf("balance sheet: %w", err)
		}
	}
	
	// Validate cashflow statement if present
	if r.CashflowStatementHistoryQuarterly != nil {
		if err := r.CashflowStatementHistoryQuarterly.Validate(); err != nil {
			return fmt.Errorf("cashflow statement: %w", err)
		}
	}
	
	return nil
}

// Validate validates income statement history
func (h *IncomeStatementHistory) Validate() error {
	if len(h.IncomeStatementHistory) == 0 {
		return fmt.Errorf("no income statements found")
	}
	
	for i, stmt := range h.IncomeStatementHistory {
		if err := stmt.Validate(); err != nil {
			return fmt.Errorf("income statement[%d]: %w", i, err)
		}
	}
	
	return nil
}

// Validate validates balance sheet history
func (h *BalanceSheetHistory) Validate() error {
	if len(h.BalanceSheetHistory) == 0 {
		return fmt.Errorf("no balance sheets found")
	}
	
	for i, sheet := range h.BalanceSheetHistory {
		if err := sheet.Validate(); err != nil {
			return fmt.Errorf("balance sheet[%d]: %w", i, err)
		}
	}
	
	return nil
}

// Validate validates cashflow statement history
func (h *CashflowStatementHistory) Validate() error {
	if len(h.CashflowStatementHistory) == 0 {
		return fmt.Errorf("no cashflow statements found")
	}
	
	for i, stmt := range h.CashflowStatementHistory {
		if err := stmt.Validate(); err != nil {
			return fmt.Errorf("cashflow statement[%d]: %w", i, err)
		}
	}
	
	return nil
}

// Validate validates an income statement
func (s *IncomeStatement) Validate() error {
	if s.EndDate.Raw == 0 {
		return fmt.Errorf("missing end date")
	}
	
	// Validate key financial values if present
	if s.TotalRevenue != nil && s.TotalRevenue.Raw != nil {
		if err := validateFinancialValue(*s.TotalRevenue.Raw); err != nil {
			return fmt.Errorf("total revenue: %w", err)
		}
	}
	
	if s.NetIncome != nil && s.NetIncome.Raw != nil {
		if err := validateFinancialValue(*s.NetIncome.Raw); err != nil {
			return fmt.Errorf("net income: %w", err)
		}
	}
	
	if s.EPS != nil && s.EPS.Raw != nil {
		if err := validateFinancialValue(*s.EPS.Raw); err != nil {
			return fmt.Errorf("EPS: %w", err)
		}
	}
	
	return nil
}

// Validate validates a balance sheet
func (s *BalanceSheet) Validate() error {
	if s.EndDate.Raw == 0 {
		return fmt.Errorf("missing end date")
	}
	
	// Validate key financial values if present
	if s.TotalAssets != nil && s.TotalAssets.Raw != nil {
		if err := validateFinancialValue(*s.TotalAssets.Raw); err != nil {
			return fmt.Errorf("total assets: %w", err)
		}
	}
	
	if s.TotalLiab != nil && s.TotalLiab.Raw != nil {
		if err := validateFinancialValue(*s.TotalLiab.Raw); err != nil {
			return fmt.Errorf("total liabilities: %w", err)
		}
	}
	
	return nil
}

// Validate validates a cashflow statement
func (s *CashflowStatement) Validate() error {
	if s.EndDate.Raw == 0 {
		return fmt.Errorf("missing end date")
	}
	
	// Validate key financial values if present
	if s.NetIncome != nil && s.NetIncome.Raw != nil {
		if err := validateFinancialValue(*s.NetIncome.Raw); err != nil {
			return fmt.Errorf("net income: %w", err)
		}
	}
	
	return nil
}

// validateFinancialValue validates a financial value
func validateFinancialValue(value int64) error {
	// int64 values cannot be NaN or infinite, so this is always valid
	// This function is kept for future extensibility if we switch to float64
	_ = value // Suppress unused parameter warning
	return nil
}

// GetFundamentals extracts fundamentals data from the response
func (r *FundamentalsResponse) GetFundamentals() (*Fundamentals, error) {
	if len(r.QuoteSummary.Result) == 0 {
		return nil, fmt.Errorf("no fundamentals results")
	}
	
	result := r.QuoteSummary.Result[0]
	fundamentals := &Fundamentals{}
	
	// Extract income statement data
	if result.IncomeStatementHistoryQuarterly != nil {
		fundamentals.IncomeStatements = result.IncomeStatementHistoryQuarterly.IncomeStatementHistory
	}
	
	// Extract balance sheet data
	if result.BalanceSheetHistoryQuarterly != nil {
		fundamentals.BalanceSheets = result.BalanceSheetHistoryQuarterly.BalanceSheetHistory
	}
	
	// Extract cashflow statement data
	if result.CashflowStatementHistoryQuarterly != nil {
		fundamentals.CashflowStatements = result.CashflowStatementHistoryQuarterly.CashflowStatementHistory
	}
	
	return fundamentals, nil
}

// Fundamentals contains all financial statement data
type Fundamentals struct {
	IncomeStatements  []IncomeStatement  `json:"incomeStatements,omitempty"`
	BalanceSheets     []BalanceSheet     `json:"balanceSheets,omitempty"`
	CashflowStatements []CashflowStatement `json:"cashflowStatements,omitempty"`
}

// DecodeFundamentalsResponseFromReader decodes a Yahoo Finance fundamentals response from an io.Reader
func DecodeFundamentalsResponseFromReader(reader io.Reader) (*FundamentalsResponse, error) {
	var response FundamentalsResponse
	
	// Use strict JSON decoding
	decoder := json.NewDecoder(reader)
	// Allow unknown fields for fundamentals as the response has many fields we don't use
	// decoder.DisallowUnknownFields()
	
	if err := decoder.Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode fundamentals response: %w", err)
	}
	
	// Validate response structure
	if err := response.Validate(); err != nil {
		return nil, fmt.Errorf("invalid fundamentals response: %w", err)
	}
	
	return &response, nil
}
