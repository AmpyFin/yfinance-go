package emit

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/AmpyFin/yfinance-go/internal/scrape"
	commonv1 "github.com/AmpyFin/ampy-proto/v2/gen/go/ampy/common/v1"
	fundamentalsv1 "github.com/AmpyFin/ampy-proto/v2/gen/go/ampy/fundamentals/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// MapFinancialsDTO converts FinancialsDTO to ampy.fundamentals.v1.FundamentalsSnapshot
func MapFinancialsDTO(dto *scrape.FinancialsDTO, runID, producer string) (*fundamentalsv1.FundamentalsSnapshot, error) {
	if dto == nil {
		return nil, fmt.Errorf("FinancialsDTO cannot be nil")
	}

	// Convert security
	security := &commonv1.SecurityId{
		Symbol: dto.Symbol,
		Mic:    normalizeMIC(dto.Market),
	}

	// Convert line items with validation
	lines := make([]*fundamentalsv1.LineItem, 0, len(dto.Lines))
	for i, line := range dto.Lines {
		// Validate period dates
		if line.PeriodStart.After(line.PeriodEnd) {
			return nil, fmt.Errorf("line %d: period_start (%v) must be before period_end (%v)", 
				i, line.PeriodStart, line.PeriodEnd)
		}

		ampyLine, err := mapFinancialLine(&line)
		if err != nil {
			return nil, fmt.Errorf("failed to map line item %d (%s): %w", i, line.Key, err)
		}
		lines = append(lines, ampyLine)
	}

	// Validate monotonic periods (optional but recommended)
	if err := validateMonotonicPeriods(lines); err != nil {
		// Log warning but don't fail - real data might have overlapping periods
		// TODO: Add proper logging
		_ = err
	}

	// Create metadata
	meta := &commonv1.Meta{
		RunId:         runID,
		Source:        "yfinance-go/scrape",
		Producer:      producer,
		SchemaVersion: "ampy.fundamentals.v1:2.1.0",
	}

	return &fundamentalsv1.FundamentalsSnapshot{
		Security: security,
		Lines:    lines,
		Source:   "yfinance/scrape",
		AsOf:     timestamppb.New(dto.AsOf),
		Meta:     meta,
	}, nil
}

// MapComprehensiveFinancialsDTO converts ComprehensiveFinancialsDTO to multiple FundamentalsSnapshot messages
func MapComprehensiveFinancialsDTO(dto *scrape.ComprehensiveFinancialsDTO, runID, producer string) ([]*fundamentalsv1.FundamentalsSnapshot, error) {
	if dto == nil {
		return nil, fmt.Errorf("ComprehensiveFinancialsDTO cannot be nil")
	}

	var snapshots []*fundamentalsv1.FundamentalsSnapshot

	// Create security
	security := &commonv1.SecurityId{
		Symbol: dto.Symbol,
		Mic:    normalizeMIC(dto.Market),
	}

	// Create metadata
	meta := &commonv1.Meta{
		RunId:         runID,
		Source:        "yfinance-go/scrape",
		Producer:      producer,
		SchemaVersion: "ampy.fundamentals.v1:2.1.0",
	}

	// Map current period data
	currentLines := extractCurrentPeriodLines(dto)
	if len(currentLines) > 0 {
		currentSnapshot := &fundamentalsv1.FundamentalsSnapshot{
			Security: security,
			Lines:    currentLines,
			Source:   "yfinance/scrape",
			AsOf:     timestamppb.New(dto.AsOf),
			Meta:     meta,
		}
		snapshots = append(snapshots, currentSnapshot)
	}

	// Map historical data if available
	// Note: ComprehensiveFinancialsDTO might have historical data that we can extract
	// For now, we focus on the current period

	return snapshots, nil
}

// mapFinancialLine converts a PeriodLine to ampy.fundamentals.v1.LineItem
func mapFinancialLine(line *scrape.PeriodLine) (*fundamentalsv1.LineItem, error) {
	// Normalize the key to canonical form
	normalizedKey := normalizeFinancialKey(line.Key)

	// Convert scaled decimal with validation
	if line.Value.Scale < 0 || line.Value.Scale > 9 {
		return nil, fmt.Errorf("invalid scale %d, must be between 0 and 9", line.Value.Scale)
	}

	value := &commonv1.Decimal{
		Scaled: line.Value.Scaled,
		Scale:  int32(line.Value.Scale),
	}

	// Convert and validate currency
	currencyCode := string(line.Currency)
	if currencyCode == "" {
		// Don't invent currency - omit it
		currencyCode = ""
	} else {
		// Validate currency format
		if len(currencyCode) != 3 {
			return nil, fmt.Errorf("invalid currency code '%s', must be 3 characters", currencyCode)
		}
		currencyCode = strings.ToUpper(currencyCode)
	}

	return &fundamentalsv1.LineItem{
		Key:          normalizedKey,
		Value:        value,
		CurrencyCode: currencyCode,
		PeriodStart:  timestamppb.New(line.PeriodStart),
		PeriodEnd:    timestamppb.New(line.PeriodEnd),
	}, nil
}

// extractCurrentPeriodLines extracts current period data from ComprehensiveFinancialsDTO
func extractCurrentPeriodLines(dto *scrape.ComprehensiveFinancialsDTO) []*fundamentalsv1.LineItem {
	var lines []*fundamentalsv1.LineItem

	// Use a recent quarter end for period (approximate)
	now := dto.AsOf
	quarterStart := time.Date(now.Year(), ((now.Month()-1)/3)*3+1, 1, 0, 0, 0, 0, time.UTC)
	quarterEnd := quarterStart.AddDate(0, 3, -1)

	// Map current values to line items
	if dto.Current.TotalRevenue != nil {
		line := createLineItem("total_revenue", dto.Current.TotalRevenue, dto.Currency, quarterStart, quarterEnd)
		if line != nil {
			lines = append(lines, line)
		}
	}

	if dto.Current.OperatingIncome != nil {
		line := createLineItem("operating_income", dto.Current.OperatingIncome, dto.Currency, quarterStart, quarterEnd)
		if line != nil {
			lines = append(lines, line)
		}
	}

	if dto.Current.NetIncomeCommonStockholders != nil {
		line := createLineItem("net_income", dto.Current.NetIncomeCommonStockholders, dto.Currency, quarterStart, quarterEnd)
		if line != nil {
			lines = append(lines, line)
		}
	}

	if dto.Current.BasicEPS != nil {
		line := createLineItem("eps_basic", dto.Current.BasicEPS, dto.Currency, quarterStart, quarterEnd)
		if line != nil {
			lines = append(lines, line)
		}
	}

	if dto.Current.DilutedEPS != nil {
		line := createLineItem("eps_diluted", dto.Current.DilutedEPS, dto.Currency, quarterStart, quarterEnd)
		if line != nil {
			lines = append(lines, line)
		}
	}

	// Balance Sheet items
	if dto.Current.TotalAssets != nil {
		line := createLineItem("total_assets", dto.Current.TotalAssets, dto.Currency, quarterStart, quarterEnd)
		if line != nil {
			lines = append(lines, line)
		}
	}

	if dto.Current.CommonStockEquity != nil {
		line := createLineItem("shareholders_equity", dto.Current.CommonStockEquity, dto.Currency, quarterStart, quarterEnd)
		if line != nil {
			lines = append(lines, line)
		}
	}

	if dto.Current.TotalDebt != nil {
		line := createLineItem("total_debt", dto.Current.TotalDebt, dto.Currency, quarterStart, quarterEnd)
		if line != nil {
			lines = append(lines, line)
		}
	}

	if dto.Current.WorkingCapital != nil {
		line := createLineItem("working_capital", dto.Current.WorkingCapital, dto.Currency, quarterStart, quarterEnd)
		if line != nil {
			lines = append(lines, line)
		}
	}

	if dto.Current.TangibleBookValue != nil {
		line := createLineItem("tangible_book_value", dto.Current.TangibleBookValue, dto.Currency, quarterStart, quarterEnd)
		if line != nil {
			lines = append(lines, line)
		}
	}

	// Cash Flow items
	if dto.Current.OperatingCashFlow != nil {
		line := createLineItem("operating_cash_flow", dto.Current.OperatingCashFlow, dto.Currency, quarterStart, quarterEnd)
		if line != nil {
			lines = append(lines, line)
		}
	}

	if dto.Current.InvestingCashFlow != nil {
		line := createLineItem("investing_cash_flow", dto.Current.InvestingCashFlow, dto.Currency, quarterStart, quarterEnd)
		if line != nil {
			lines = append(lines, line)
		}
	}

	if dto.Current.FinancingCashFlow != nil {
		line := createLineItem("financing_cash_flow", dto.Current.FinancingCashFlow, dto.Currency, quarterStart, quarterEnd)
		if line != nil {
			lines = append(lines, line)
		}
	}

	if dto.Current.FreeCashFlow != nil {
		line := createLineItem("free_cash_flow", dto.Current.FreeCashFlow, dto.Currency, quarterStart, quarterEnd)
		if line != nil {
			lines = append(lines, line)
		}
	}

	if dto.Current.CapitalExpenditure != nil {
		line := createLineItem("capital_expenditure", dto.Current.CapitalExpenditure, dto.Currency, quarterStart, quarterEnd)
		if line != nil {
			lines = append(lines, line)
		}
	}

	// Additional Income Statement items
	if dto.Current.CostOfRevenue != nil {
		line := createLineItem("cost_of_revenue", dto.Current.CostOfRevenue, dto.Currency, quarterStart, quarterEnd)
		if line != nil {
			lines = append(lines, line)
		}
	}

	if dto.Current.GrossProfit != nil {
		line := createLineItem("gross_profit", dto.Current.GrossProfit, dto.Currency, quarterStart, quarterEnd)
		if line != nil {
			lines = append(lines, line)
		}
	}

	if dto.Current.EBITDA != nil {
		line := createLineItem("ebitda", dto.Current.EBITDA, dto.Currency, quarterStart, quarterEnd)
		if line != nil {
			lines = append(lines, line)
		}
	}

	if dto.Current.EBIT != nil {
		line := createLineItem("ebit", dto.Current.EBIT, dto.Currency, quarterStart, quarterEnd)
		if line != nil {
			lines = append(lines, line)
		}
	}

	// Share count data (convert to scaled decimals for consistency)
	if dto.Current.BasicAverageShares != nil {
		shareValue := &scrape.Scaled{
			Scaled: *dto.Current.BasicAverageShares,
			Scale:  0, // Shares are whole numbers
		}
		line := createLineItem("shares_outstanding_basic", shareValue, "", quarterStart, quarterEnd)
		if line != nil {
			lines = append(lines, line)
		}
	}

	if dto.Current.DilutedAverageShares != nil {
		shareValue := &scrape.Scaled{
			Scaled: *dto.Current.DilutedAverageShares,
			Scale:  0, // Shares are whole numbers
		}
		line := createLineItem("shares_outstanding_diluted", shareValue, "", quarterStart, quarterEnd)
		if line != nil {
			lines = append(lines, line)
		}
	}

	return lines
}

// createLineItem creates a LineItem from scaled value
func createLineItem(key string, value *scrape.Scaled, currency string, periodStart, periodEnd time.Time) *fundamentalsv1.LineItem {
	if value == nil {
		return nil
	}

	// Validate scale
	if value.Scale < 0 || value.Scale > 9 {
		return nil
	}

	decimal := &commonv1.Decimal{
		Scaled: value.Scaled,
		Scale:  int32(value.Scale),
	}

	currencyCode := strings.ToUpper(currency)
	if len(currencyCode) != 3 {
		currencyCode = "" // Omit invalid currency
	}

	return &fundamentalsv1.LineItem{
		Key:          key,
		Value:        decimal,
		CurrencyCode: currencyCode,
		PeriodStart:  timestamppb.New(periodStart),
		PeriodEnd:    timestamppb.New(periodEnd),
	}
}

// normalizeFinancialKey normalizes financial statement keys to canonical form
func normalizeFinancialKey(key string) string {
	// Convert to lowercase and replace spaces/hyphens with underscores
	normalized := strings.ToLower(key)
	normalized = strings.ReplaceAll(normalized, " ", "_")
	normalized = strings.ReplaceAll(normalized, "-", "_")
	normalized = strings.ReplaceAll(normalized, ".", "_")

	// Map common variations to canonical names
	keyMappings := map[string]string{
		"total_revenues":                    "total_revenue",
		"revenues":                          "total_revenue",
		"net_revenues":                      "total_revenue",
		"total_operating_revenues":          "total_revenue",
		"operating_revenues":                "total_revenue",
		"net_income_common_stockholders":    "net_income",
		"net_income_applicable_to_common":   "net_income",
		"net_earnings":                      "net_income",
		"earnings":                          "net_income",
		"basic_earnings_per_share":          "eps_basic",
		"basic_eps":                         "eps_basic",
		"diluted_earnings_per_share":        "eps_diluted",
		"diluted_eps":                       "eps_diluted",
		"operating_earnings":                "operating_income",
		"operating_profit":                  "operating_income",
		"gross_revenues":                    "gross_profit",
		"total_stockholders_equity":         "shareholders_equity",
		"stockholders_equity":               "shareholders_equity",
		"shareholders_equity_total":         "shareholders_equity",
		"cash_and_cash_equivalents":         "cash_and_equivalents",
		"cash_equivalents":                  "cash_and_equivalents",
		"total_cash":                        "cash_and_equivalents",
		"free_cash_flows":                   "free_cash_flow",
		"fcf":                               "free_cash_flow",
		"operating_cash_flows":              "operating_cash_flow",
		"cash_from_operations":              "operating_cash_flow",
		"investing_cash_flows":              "investing_cash_flow",
		"cash_from_investing":               "investing_cash_flow",
		"financing_cash_flows":              "financing_cash_flow",
		"cash_from_financing":               "financing_cash_flow",
	}

	if canonical, exists := keyMappings[normalized]; exists {
		return canonical
	}

	return normalized
}

// normalizeMIC converts market identifier to proper MIC format
func normalizeMIC(market string) string {
	if market == "" {
		return ""
	}

	// Convert common market names to MIC codes
	micMappings := map[string]string{
		"NASDAQ":     "XNAS",
		"NYSE":       "XNYS",
		"AMEX":       "XASE",
		"OTC":        "OTCM",
		"TSX":        "XTSE",
		"LSE":        "XLON",
		"TOKYO":      "XJPX",
		"SHANGHAI":   "XSHG",
		"SHENZHEN":   "XSHE",
		"HONG_KONG":  "XHKG",
		"FRANKFURT":  "XFRA",
		"EURONEXT":   "XPAR",
	}

	upperMarket := strings.ToUpper(market)
	if mic, exists := micMappings[upperMarket]; exists {
		return mic
	}

	// If already looks like a MIC (4 uppercase chars), return as is
	if len(market) == 4 && strings.ToUpper(market) == market {
		return market
	}

	// Otherwise, truncate to 4 chars and uppercase
	if len(market) > 4 {
		return strings.ToUpper(market[:4])
	}

	return strings.ToUpper(market)
}

// MapKeyStatisticsDTO converts ComprehensiveKeyStatisticsDTO to ampy.fundamentals.v1.FundamentalsSnapshot
func MapKeyStatisticsDTO(dto *scrape.ComprehensiveKeyStatisticsDTO, runID, producer string) (*fundamentalsv1.FundamentalsSnapshot, error) {
	if dto == nil {
		return nil, fmt.Errorf("ComprehensiveKeyStatisticsDTO cannot be nil")
	}

	// Create security
	security := &commonv1.SecurityId{
		Symbol: dto.Symbol,
		Mic:    normalizeMIC(dto.Market),
	}

	// Create metadata
	meta := &commonv1.Meta{
		RunId:         runID,
		Source:        "yfinance-go/scrape",
		Producer:      producer,
		SchemaVersion: "ampy.fundamentals.v1:2.1.0",
	}

	// Use current time as period (key statistics are point-in-time)
	now := dto.AsOf
	periodStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	periodEnd := periodStart.Add(24 * time.Hour)

	var lines []*fundamentalsv1.LineItem

	// Market valuation metrics
	if dto.Current.MarketCap != nil {
		line := createLineItem("market_cap", dto.Current.MarketCap, dto.Currency, periodStart, periodEnd)
		if line != nil {
			lines = append(lines, line)
		}
	}

	if dto.Current.EnterpriseValue != nil {
		line := createLineItem("enterprise_value", dto.Current.EnterpriseValue, dto.Currency, periodStart, periodEnd)
		if line != nil {
			lines = append(lines, line)
		}
	}

	// Valuation ratios (no currency as they are ratios)
	if dto.Current.TrailingPE != nil {
		line := createLineItem("pe_ratio_trailing", dto.Current.TrailingPE, "", periodStart, periodEnd)
		if line != nil {
			lines = append(lines, line)
		}
	}

	if dto.Current.ForwardPE != nil {
		line := createLineItem("pe_ratio_forward", dto.Current.ForwardPE, "", periodStart, periodEnd)
		if line != nil {
			lines = append(lines, line)
		}
	}

	if dto.Current.PEGRatio != nil {
		line := createLineItem("peg_ratio", dto.Current.PEGRatio, "", periodStart, periodEnd)
		if line != nil {
			lines = append(lines, line)
		}
	}

	if dto.Current.PriceSales != nil {
		line := createLineItem("price_to_sales", dto.Current.PriceSales, "", periodStart, periodEnd)
		if line != nil {
			lines = append(lines, line)
		}
	}

	if dto.Current.PriceBook != nil {
		line := createLineItem("price_to_book", dto.Current.PriceBook, "", periodStart, periodEnd)
		if line != nil {
			lines = append(lines, line)
		}
	}

	if dto.Current.EnterpriseValueRevenue != nil {
		line := createLineItem("ev_to_revenue", dto.Current.EnterpriseValueRevenue, "", periodStart, periodEnd)
		if line != nil {
			lines = append(lines, line)
		}
	}

	if dto.Current.EnterpriseValueEBITDA != nil {
		line := createLineItem("ev_to_ebitda", dto.Current.EnterpriseValueEBITDA, "", periodStart, periodEnd)
		if line != nil {
			lines = append(lines, line)
		}
	}

	// Additional metrics from the Additional section
	if dto.Additional.Beta != nil {
		line := createLineItem("beta", dto.Additional.Beta, "", periodStart, periodEnd)
		if line != nil {
			lines = append(lines, line)
		}
	}

	if dto.Additional.SharesOutstanding != nil {
		shareValue := &scrape.Scaled{
			Scaled: *dto.Additional.SharesOutstanding,
			Scale:  0, // Shares are whole numbers
		}
		line := createLineItem("shares_outstanding", shareValue, "", periodStart, periodEnd)
		if line != nil {
			lines = append(lines, line)
		}
	}

	if dto.Additional.ProfitMargin != nil {
		line := createLineItem("profit_margin", dto.Additional.ProfitMargin, "", periodStart, periodEnd)
		if line != nil {
			lines = append(lines, line)
		}
	}

	if dto.Additional.OperatingMargin != nil {
		line := createLineItem("operating_margin", dto.Additional.OperatingMargin, "", periodStart, periodEnd)
		if line != nil {
			lines = append(lines, line)
		}
	}

	if dto.Additional.ReturnOnAssets != nil {
		line := createLineItem("return_on_assets", dto.Additional.ReturnOnAssets, "", periodStart, periodEnd)
		if line != nil {
			lines = append(lines, line)
		}
	}

	if dto.Additional.ReturnOnEquity != nil {
		line := createLineItem("return_on_equity", dto.Additional.ReturnOnEquity, "", periodStart, periodEnd)
		if line != nil {
			lines = append(lines, line)
		}
	}

	return &fundamentalsv1.FundamentalsSnapshot{
		Security: security,
		Lines:    lines,
		Source:   "yfinance/scrape",
		AsOf:     timestamppb.New(dto.AsOf),
		Meta:     meta,
	}, nil
}

// MapAnalysisDTO converts ComprehensiveAnalysisDTO to ampy.fundamentals.v1.FundamentalsSnapshot
// Note: Analysis data contains mostly forward-looking estimates, so we map the most relevant quantitative data
func MapAnalysisDTO(dto *scrape.ComprehensiveAnalysisDTO, runID, producer string) (*fundamentalsv1.FundamentalsSnapshot, error) {
	if dto == nil {
		return nil, fmt.Errorf("ComprehensiveAnalysisDTO cannot be nil")
	}

	// Create security
	security := &commonv1.SecurityId{
		Symbol: dto.Symbol,
		Mic:    normalizeMIC(dto.Market),
	}

	// Create metadata
	meta := &commonv1.Meta{
		RunId:         runID,
		Source:        "yfinance-go/scrape",
		Producer:      producer,
		SchemaVersion: "ampy.fundamentals.v1:2.1.0",
	}

	// Use current time as period (analysis data is forward-looking estimates)
	now := dto.AsOf
	periodStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	periodEnd := periodStart.Add(24 * time.Hour)

	var lines []*fundamentalsv1.LineItem

	// Map current quarter earnings estimates
	if dto.EarningsEstimate.CurrentQtr.AvgEstimate != nil {
		epsValue := &scrape.Scaled{
			Scaled: int64(*dto.EarningsEstimate.CurrentQtr.AvgEstimate * 10000), // 4 decimal places
			Scale:  4,
		}
		line := createLineItem("eps_estimate_current_quarter", epsValue, dto.EarningsEstimate.Currency, periodStart, periodEnd)
		if line != nil {
			lines = append(lines, line)
		}
	}

	if dto.EarningsEstimate.NextQtr.AvgEstimate != nil {
		epsValue := &scrape.Scaled{
			Scaled: int64(*dto.EarningsEstimate.NextQtr.AvgEstimate * 10000), // 4 decimal places
			Scale:  4,
		}
		line := createLineItem("eps_estimate_next_quarter", epsValue, dto.EarningsEstimate.Currency, periodStart, periodEnd)
		if line != nil {
			lines = append(lines, line)
		}
	}

	// Map current year earnings estimates
	if dto.EarningsEstimate.CurrentYear.AvgEstimate != nil {
		epsValue := &scrape.Scaled{
			Scaled: int64(*dto.EarningsEstimate.CurrentYear.AvgEstimate * 10000), // 4 decimal places
			Scale:  4,
		}
		line := createLineItem("eps_estimate_current_year", epsValue, dto.EarningsEstimate.Currency, periodStart, periodEnd)
		if line != nil {
			lines = append(lines, line)
		}
	}

	if dto.EarningsEstimate.NextYear.AvgEstimate != nil {
		epsValue := &scrape.Scaled{
			Scaled: int64(*dto.EarningsEstimate.NextYear.AvgEstimate * 10000), // 4 decimal places
			Scale:  4,
		}
		line := createLineItem("eps_estimate_next_year", epsValue, dto.EarningsEstimate.Currency, periodStart, periodEnd)
		if line != nil {
			lines = append(lines, line)
		}
	}

	// Map analyst counts as metrics (no currency)
	if dto.EarningsEstimate.CurrentQtr.NoOfAnalysts != nil {
		analystValue := &scrape.Scaled{
			Scaled: int64(*dto.EarningsEstimate.CurrentQtr.NoOfAnalysts),
			Scale:  0, // Whole numbers
		}
		line := createLineItem("analyst_count_current_quarter", analystValue, "", periodStart, periodEnd)
		if line != nil {
			lines = append(lines, line)
		}
	}

	// Map earnings history (recent actual EPS)
	if len(dto.EarningsHistory.Data) > 0 {
		// Get the most recent earnings data
		recent := dto.EarningsHistory.Data[0]
		if recent.EPSActual != nil {
			epsValue := &scrape.Scaled{
				Scaled: int64(*recent.EPSActual * 10000), // 4 decimal places
				Scale:  4,
			}
			line := createLineItem("eps_actual_recent", epsValue, dto.EarningsHistory.Currency, periodStart, periodEnd)
			if line != nil {
				lines = append(lines, line)
			}
		}
	}

	// Map growth estimates (if available)
	if dto.GrowthEstimate.CurrentYear != nil {
		// Parse growth rate if it's a percentage string
		growthStr := *dto.GrowthEstimate.CurrentYear
		if strings.HasSuffix(growthStr, "%") {
			growthStr = strings.TrimSuffix(growthStr, "%")
			if growthVal, err := strconv.ParseFloat(growthStr, 64); err == nil {
				growthValue := &scrape.Scaled{
					Scaled: int64(growthVal * 100), // Store as basis points
					Scale:  2,
				}
				line := createLineItem("growth_estimate_current_year", growthValue, "", periodStart, periodEnd)
				if line != nil {
					lines = append(lines, line)
				}
			}
		}
	}

	return &fundamentalsv1.FundamentalsSnapshot{
		Security: security,
		Lines:    lines,
		Source:   "yfinance/scrape",
		AsOf:     timestamppb.New(dto.AsOf),
		Meta:     meta,
	}, nil
}

// MapAnalystInsightsDTO converts AnalystInsightsDTO to ampy.fundamentals.v1.FundamentalsSnapshot
func MapAnalystInsightsDTO(dto *scrape.AnalystInsightsDTO, runID, producer string) (*fundamentalsv1.FundamentalsSnapshot, error) {
	if dto == nil {
		return nil, fmt.Errorf("AnalystInsightsDTO cannot be nil")
	}

	// Create security
	security := &commonv1.SecurityId{
		Symbol: dto.Symbol,
		Mic:    normalizeMIC(dto.Market),
	}

	// Create metadata
	meta := &commonv1.Meta{
		RunId:         runID,
		Source:        "yfinance-go/scrape",
		Producer:      producer,
		SchemaVersion: "ampy.fundamentals.v1:2.1.0",
	}

	// Use current time as period (analyst insights are point-in-time)
	now := dto.AsOf
	periodStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	periodEnd := periodStart.Add(24 * time.Hour)

	var lines []*fundamentalsv1.LineItem

	// Map price targets (assuming USD currency since it's not provided in DTO)
	if dto.CurrentPrice != nil {
		priceValue := &scrape.Scaled{
			Scaled: int64(*dto.CurrentPrice * 10000), // 4 decimal places
			Scale:  4,
		}
		line := createLineItem("current_price", priceValue, "USD", periodStart, periodEnd)
		if line != nil {
			lines = append(lines, line)
		}
	}

	if dto.TargetMeanPrice != nil {
		priceValue := &scrape.Scaled{
			Scaled: int64(*dto.TargetMeanPrice * 10000), // 4 decimal places
			Scale:  4,
		}
		line := createLineItem("target_price_mean", priceValue, "USD", periodStart, periodEnd)
		if line != nil {
			lines = append(lines, line)
		}
	}

	if dto.TargetMedianPrice != nil {
		priceValue := &scrape.Scaled{
			Scaled: int64(*dto.TargetMedianPrice * 10000), // 4 decimal places
			Scale:  4,
		}
		line := createLineItem("target_price_median", priceValue, "USD", periodStart, periodEnd)
		if line != nil {
			lines = append(lines, line)
		}
	}

	if dto.TargetHighPrice != nil {
		priceValue := &scrape.Scaled{
			Scaled: int64(*dto.TargetHighPrice * 10000), // 4 decimal places
			Scale:  4,
		}
		line := createLineItem("target_price_high", priceValue, "USD", periodStart, periodEnd)
		if line != nil {
			lines = append(lines, line)
		}
	}

	if dto.TargetLowPrice != nil {
		priceValue := &scrape.Scaled{
			Scaled: int64(*dto.TargetLowPrice * 10000), // 4 decimal places
			Scale:  4,
		}
		line := createLineItem("target_price_low", priceValue, "USD", periodStart, periodEnd)
		if line != nil {
			lines = append(lines, line)
		}
	}

	// Map analyst metrics (no currency)
	if dto.NumberOfAnalysts != nil {
		analystValue := &scrape.Scaled{
			Scaled: int64(*dto.NumberOfAnalysts),
			Scale:  0, // Whole numbers
		}
		line := createLineItem("analyst_count", analystValue, "", periodStart, periodEnd)
		if line != nil {
			lines = append(lines, line)
		}
	}

	if dto.RecommendationMean != nil {
		recValue := &scrape.Scaled{
			Scaled: int64(*dto.RecommendationMean * 100), // 2 decimal places
			Scale:  2,
		}
		line := createLineItem("recommendation_score", recValue, "", periodStart, periodEnd)
		if line != nil {
			lines = append(lines, line)
		}
	}

	// Calculate upside potential if both current and target prices are available
	if dto.CurrentPrice != nil && dto.TargetMeanPrice != nil {
		upside := ((*dto.TargetMeanPrice - *dto.CurrentPrice) / *dto.CurrentPrice) * 100
		upsideValue := &scrape.Scaled{
			Scaled: int64(upside * 100), // Store as basis points
			Scale:  2,
		}
		line := createLineItem("upside_potential_percent", upsideValue, "", periodStart, periodEnd)
		if line != nil {
			lines = append(lines, line)
		}
	}

	return &fundamentalsv1.FundamentalsSnapshot{
		Security: security,
		Lines:    lines,
		Source:   "yfinance/scrape",
		AsOf:     timestamppb.New(dto.AsOf),
		Meta:     meta,
	}, nil
}

// validateMonotonicPeriods checks if periods are monotonic (non-overlapping)
func validateMonotonicPeriods(lines []*fundamentalsv1.LineItem) error {
	if len(lines) <= 1 {
		return nil
	}

	// Group by key and check periods within each group
	keyGroups := make(map[string][]*fundamentalsv1.LineItem)
	for _, line := range lines {
		keyGroups[line.Key] = append(keyGroups[line.Key], line)
	}

	for key, keyLines := range keyGroups {
		if len(keyLines) <= 1 {
			continue
		}

		// Sort by period start time
		for i := 0; i < len(keyLines)-1; i++ {
			for j := i + 1; j < len(keyLines); j++ {
				if keyLines[i].PeriodStart.AsTime().After(keyLines[j].PeriodStart.AsTime()) {
					keyLines[i], keyLines[j] = keyLines[j], keyLines[i]
				}
			}
		}

		// Check for overlaps
		for i := 0; i < len(keyLines)-1; i++ {
			currentEnd := keyLines[i].PeriodEnd.AsTime()
			nextStart := keyLines[i+1].PeriodStart.AsTime()

			if currentEnd.After(nextStart) {
				return fmt.Errorf("overlapping periods for key '%s': period %d ends %v, period %d starts %v",
					key, i, currentEnd, i+1, nextStart)
			}
		}
	}

	return nil
}
