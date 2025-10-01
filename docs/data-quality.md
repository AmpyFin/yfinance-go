# Data Quality & Validation Guidelines

This guide provides comprehensive information about data quality expectations, validation strategies, and best practices for ensuring reliable data when using yfinance-go.

## Table of Contents

1. [Data Quality Expectations](#data-quality-expectations)
2. [Validation Strategies](#validation-strategies)
3. [Data Quality Checks](#data-quality-checks)
4. [Handling Missing Data](#handling-missing-data)
5. [Data Consistency](#data-consistency)
6. [Quality Monitoring](#quality-monitoring)
7. [Best Practices](#best-practices)

## Data Quality Expectations

### Expected Data Availability

| Data Type | Availability | Quality Level | Notes |
|-----------|--------------|---------------|-------|
| **Quotes** | High (95%+) | Excellent | Generally available for all active stocks |
| **Historical Data** | High (90%+) | Excellent | Available for most stocks, limited by listing date |
| **Company Info** | High (95%+) | Good | Basic info only, detailed profiles not available |
| **Financials** | Medium (80%+) | Good | Available for most public companies |
| **Key Statistics** | Medium (75%+) | Good | May be limited for smaller companies |
| **Analysis** | Medium (70%+) | Variable | May be limited for smaller companies |
| **News** | Low (50%+) | Variable | Highly variable, may be empty for many stocks |

### Data Quality Indicators

#### High Quality Indicators
- ✅ Data fields are populated
- ✅ Timestamps are recent and valid
- ✅ Price data is reasonable (not zero or negative)
- ✅ Volume data is positive
- ✅ Currency codes are valid ISO-4217 codes
- ✅ Scaled decimal values have appropriate scales

#### Low Quality Indicators
- ❌ Missing required fields
- ❌ Zero or negative prices
- ❌ Negative volume
- ❌ Invalid timestamps
- ❌ Missing currency codes
- ❌ Inconsistent data formats

## Validation Strategies

### 1. Basic Data Validation

```go
func validateBarData(bar *yfinance.NormalizedBar) error {
    // Check for required fields
    if bar.EventTime.IsZero() {
        return fmt.Errorf("missing event time")
    }
    
    if bar.CurrencyCode == "" {
        return fmt.Errorf("missing currency code")
    }
    
    // Validate price data
    if err := validateScaledDecimal(bar.Open); err != nil {
        return fmt.Errorf("invalid open price: %w", err)
    }
    
    if err := validateScaledDecimal(bar.High); err != nil {
        return fmt.Errorf("invalid high price: %w", err)
    }
    
    if err := validateScaledDecimal(bar.Low); err != nil {
        return fmt.Errorf("invalid low price: %w", err)
    }
    
    if err := validateScaledDecimal(bar.Close); err != nil {
        return fmt.Errorf("invalid close price: %w", err)
    }
    
    // Validate volume
    if bar.Volume < 0 {
        return fmt.Errorf("negative volume: %d", bar.Volume)
    }
    
    // Validate price relationships
    if err := validatePriceRelationships(bar); err != nil {
        return fmt.Errorf("invalid price relationships: %w", err)
    }
    
    return nil
}

func validateScaledDecimal(sd yfinance.ScaledDecimal) error {
    if sd.Scale < 0 {
        return fmt.Errorf("negative scale: %d", sd.Scale)
    }
    
    if sd.Scale > 10 {
        return fmt.Errorf("scale too large: %d", sd.Scale)
    }
    
    if sd.Scaled == 0 && sd.Scale > 0 {
        return fmt.Errorf("zero value with non-zero scale")
    }
    
    return nil
}

func validatePriceRelationships(bar *yfinance.NormalizedBar) error {
    open := float64(bar.Open.Scaled) / float64(bar.Open.Scale)
    high := float64(bar.High.Scaled) / float64(bar.High.Scale)
    low := float64(bar.Low.Scaled) / float64(bar.Low.Scale)
    close := float64(bar.Close.Scaled) / float64(bar.Close.Scale)
    
    // High should be >= all other prices
    if high < open || high < low || high < close {
        return fmt.Errorf("high price is not the highest")
    }
    
    // Low should be <= all other prices
    if low > open || low > high || low > close {
        return fmt.Errorf("low price is not the lowest")
    }
    
    // Prices should be positive
    if open <= 0 || high <= 0 || low <= 0 || close <= 0 {
        return fmt.Errorf("non-positive prices detected")
    }
    
    return nil
}
```

### 2. Quote Data Validation

```go
func validateQuoteData(quote *yfinance.NormalizedQuote) error {
    // Check for required fields
    if quote.Security.Symbol == "" {
        return fmt.Errorf("missing symbol")
    }
    
    if quote.CurrencyCode == "" {
        return fmt.Errorf("missing currency code")
    }
    
    if quote.EventTime.IsZero() {
        return fmt.Errorf("missing event time")
    }
    
    // Validate market price if present
    if quote.RegularMarketPrice != nil {
        if err := validateScaledDecimal(*quote.RegularMarketPrice); err != nil {
            return fmt.Errorf("invalid market price: %w", err)
        }
        
        price := float64(quote.RegularMarketPrice.Scaled) / float64(quote.RegularMarketPrice.Scale)
        if price <= 0 {
            return fmt.Errorf("non-positive market price: %.2f", price)
        }
    }
    
    // Validate volume if present
    if quote.RegularMarketVolume != nil {
        if *quote.RegularMarketVolume < 0 {
            return fmt.Errorf("negative volume: %d", *quote.RegularMarketVolume)
        }
    }
    
    // Validate bid/ask if present
    if quote.Bid != nil && quote.Ask != nil {
        bid := float64(quote.Bid.Scaled) / float64(quote.Bid.Scale)
        ask := float64(quote.Ask.Scaled) / float64(quote.Ask.Scale)
        
        if bid > ask {
            return fmt.Errorf("bid price higher than ask price")
        }
    }
    
    return nil
}
```

### 3. Financial Data Validation

```go
func validateFinancialsData(financials *yfinance.FundamentalsSnapshot) error {
    if len(financials.Lines) == 0 {
        return fmt.Errorf("no financial lines found")
    }
    
    // Check for required metadata
    if financials.Meta.SchemaVersion == "" {
        return fmt.Errorf("missing schema version")
    }
    
    if financials.Meta.RunId == "" {
        return fmt.Errorf("missing run ID")
    }
    
    // Validate each line item
    for i, line := range financials.Lines {
        if err := validateFinancialLine(line); err != nil {
            return fmt.Errorf("invalid line %d: %w", i, err)
        }
    }
    
    return nil
}

func validateFinancialLine(line *yfinance.FundamentalsLine) error {
    if line.Key == "" {
        return fmt.Errorf("missing key")
    }
    
    if line.CurrencyCode == "" {
        return fmt.Errorf("missing currency code")
    }
    
    // Validate scaled decimal
    if err := validateScaledDecimal(line.Value); err != nil {
        return fmt.Errorf("invalid value: %w", err)
    }
    
    // Check for reasonable values
    value := float64(line.Value.Scaled) / float64(line.Value.Scale)
    
    // Some financial metrics should be positive
    positiveMetrics := []string{"revenue", "net_income", "total_assets", "market_cap"}
    for _, metric := range positiveMetrics {
        if strings.Contains(strings.ToLower(line.Key), metric) && value < 0 {
            return fmt.Errorf("negative value for positive metric %s: %.2f", line.Key, value)
        }
    }
    
    return nil
}
```

## Data Quality Checks

### 1. Completeness Checks

```go
func checkDataCompleteness(data *StockData) []string {
    var issues []string
    
    // Check quote data
    if data.Quote == nil {
        issues = append(issues, "Missing quote data")
    } else {
        if data.Quote.RegularMarketPrice == nil {
            issues = append(issues, "Missing market price")
        }
        if data.Quote.RegularMarketVolume == nil {
            issues = append(issues, "Missing volume data")
        }
    }
    
    // Check company info
    if data.CompanyInfo == nil {
        issues = append(issues, "Missing company information")
    } else {
        if data.CompanyInfo.LongName == "" {
            issues = append(issues, "Missing company name")
        }
        if data.CompanyInfo.Exchange == "" {
            issues = append(issues, "Missing exchange information")
        }
    }
    
    // Check financials
    if data.Financials == nil {
        issues = append(issues, "Missing financial data")
    } else if len(data.Financials.Lines) == 0 {
        issues = append(issues, "Empty financial data")
    }
    
    // Check news
    if data.News == nil || len(data.News) == 0 {
        issues = append(issues, "No news articles available")
    }
    
    return issues
}
```

### 2. Consistency Checks

```go
func checkDataConsistency(data *StockData) []string {
    var issues []string
    
    // Check currency consistency
    if data.Quote != nil && data.CompanyInfo != nil {
        if data.Quote.CurrencyCode != data.CompanyInfo.Currency {
            issues = append(issues, fmt.Sprintf("Currency mismatch: quote=%s, company=%s", 
                data.Quote.CurrencyCode, data.CompanyInfo.Currency))
        }
    }
    
    // Check symbol consistency
    if data.Quote != nil && data.CompanyInfo != nil {
        if data.Quote.Security.Symbol != data.CompanyInfo.Security.Symbol {
            issues = append(issues, fmt.Sprintf("Symbol mismatch: quote=%s, company=%s", 
                data.Quote.Security.Symbol, data.CompanyInfo.Security.Symbol))
        }
    }
    
    // Check timestamp consistency
    if data.Quote != nil && data.CompanyInfo != nil {
        timeDiff := data.Quote.EventTime.Sub(data.CompanyInfo.EventTime)
        if timeDiff > 24*time.Hour {
            issues = append(issues, fmt.Sprintf("Large time difference: %.2f hours", 
                timeDiff.Hours()))
        }
    }
    
    return issues
}
```

### 3. Reasonableness Checks

```go
func checkDataReasonableness(data *StockData) []string {
    var issues []string
    
    // Check price reasonableness
    if data.Quote != nil && data.Quote.RegularMarketPrice != nil {
        price := float64(data.Quote.RegularMarketPrice.Scaled) / float64(data.Quote.RegularMarketPrice.Scale)
        
        if price <= 0 {
            issues = append(issues, "Non-positive price")
        } else if price > 10000 {
            issues = append(issues, fmt.Sprintf("Unusually high price: %.2f", price))
        } else if price < 0.01 {
            issues = append(issues, fmt.Sprintf("Unusually low price: %.2f", price))
        }
    }
    
    // Check volume reasonableness
    if data.Quote != nil && data.Quote.RegularMarketVolume != nil {
        volume := *data.Quote.RegularMarketVolume
        
        if volume < 0 {
            issues = append(issues, "Negative volume")
        } else if volume > 1000000000 { // 1 billion
            issues = append(issues, fmt.Sprintf("Unusually high volume: %d", volume))
        }
    }
    
    // Check financial data reasonableness
    if data.Financials != nil {
        for _, line := range data.Financials.Lines {
            value := float64(line.Value.Scaled) / float64(line.Value.Scale)
            
            // Check for extreme values
            if strings.Contains(strings.ToLower(line.Key), "revenue") {
                if value > 1000000000000 { // 1 trillion
                    issues = append(issues, fmt.Sprintf("Extremely high revenue: %.2f", value))
                }
            }
            
            if strings.Contains(strings.ToLower(line.Key), "market_cap") {
                if value > 10000000000000 { // 10 trillion
                    issues = append(issues, fmt.Sprintf("Extremely high market cap: %.2f", value))
                }
            }
        }
    }
    
    return issues
}
```

## Handling Missing Data

### 1. Graceful Degradation

```go
func processStockDataWithFallback(client *yfinance.Client, symbol string) *StockData {
    data := &StockData{Symbol: symbol}
    
    // Try to fetch quote (required)
    quote, err := client.FetchQuote(ctx, symbol, runID)
    if err != nil {
        data.Errors = append(data.Errors, fmt.Sprintf("Quote failed: %v", err))
        return data // Cannot continue without quote
    }
    data.Quote = quote
    
    // Try to fetch company info (optional)
    companyInfo, err := client.FetchCompanyInfo(ctx, symbol, runID)
    if err != nil {
        data.Warnings = append(data.Warnings, fmt.Sprintf("Company info failed: %v", err))
        // Continue without company info
    } else {
        data.CompanyInfo = companyInfo
    }
    
    // Try to fetch financials (optional)
    financials, err := client.ScrapeFinancials(ctx, symbol, runID)
    if err != nil {
        data.Warnings = append(data.Warnings, fmt.Sprintf("Financials failed: %v", err))
        // Continue without financials
    } else {
        data.Financials = financials
    }
    
    // Try to fetch news (optional)
    news, err := client.ScrapeNews(ctx, symbol, runID)
    if err != nil {
        data.Warnings = append(data.Warnings, fmt.Sprintf("News failed: %v", err))
        // Continue without news
    } else {
        data.News = news
    }
    
    return data
}
```

### 2. Data Imputation

```go
func imputeMissingData(data *StockData) {
    // Impute missing volume with average
    if data.Quote != nil && data.Quote.RegularMarketVolume == nil {
        // Use historical average or industry average
        avgVolume := getAverageVolume(data.Symbol)
        data.Quote.RegularMarketVolume = &avgVolume
        data.Warnings = append(data.Warnings, "Volume imputed from historical average")
    }
    
    // Impute missing company info
    if data.CompanyInfo == nil {
        data.CompanyInfo = &yfinance.NormalizedCompanyInfo{
            Security: yfinance.Security{Symbol: data.Symbol},
            LongName: data.Symbol, // Use symbol as fallback
            Currency: "USD",       // Default currency
        }
        data.Warnings = append(data.Warnings, "Company info imputed with defaults")
    }
}

func getAverageVolume(symbol string) int64 {
    // This would typically query a database or cache
    // For now, return a reasonable default
    return 1000000 // 1 million shares
}
```

### 3. Data Quality Scoring

```go
type DataQualityScore struct {
    Overall    float64
    Completeness float64
    Consistency  float64
    Reasonableness float64
    Issues     []string
}

func calculateDataQualityScore(data *StockData) DataQualityScore {
    score := DataQualityScore{
        Issues: make([]string, 0),
    }
    
    // Calculate completeness score
    completenessIssues := checkDataCompleteness(data)
    score.Completeness = 1.0 - float64(len(completenessIssues))/4.0 // 4 main data types
    score.Issues = append(score.Issues, completenessIssues...)
    
    // Calculate consistency score
    consistencyIssues := checkDataConsistency(data)
    score.Consistency = 1.0 - float64(len(consistencyIssues))/3.0 // 3 consistency checks
    score.Issues = append(score.Issues, consistencyIssues...)
    
    // Calculate reasonableness score
    reasonablenessIssues := checkDataReasonableness(data)
    score.Reasonableness = 1.0 - float64(len(reasonablenessIssues))/5.0 // 5 reasonableness checks
    score.Issues = append(score.Issues, reasonablenessIssues...)
    
    // Calculate overall score
    score.Overall = (score.Completeness + score.Consistency + score.Reasonableness) / 3.0
    
    return score
}
```

## Data Consistency

### 1. Cross-Reference Validation

```go
func validateCrossReferences(data *StockData) []string {
    var issues []string
    
    // Validate symbol consistency across all data types
    symbols := make(map[string]bool)
    
    if data.Quote != nil {
        symbols[data.Quote.Security.Symbol] = true
    }
    
    if data.CompanyInfo != nil {
        symbols[data.CompanyInfo.Security.Symbol] = true
    }
    
    if data.Financials != nil {
        symbols[data.Financials.Security.Symbol] = true
    }
    
    if len(symbols) > 1 {
        issues = append(issues, "Symbol mismatch across data types")
    }
    
    // Validate currency consistency
    currencies := make(map[string]bool)
    
    if data.Quote != nil {
        currencies[data.Quote.CurrencyCode] = true
    }
    
    if data.CompanyInfo != nil {
        currencies[data.CompanyInfo.Currency] = true
    }
    
    if len(currencies) > 1 {
        issues = append(issues, "Currency mismatch across data types")
    }
    
    return issues
}
```

### 2. Temporal Consistency

```go
func validateTemporalConsistency(data *StockData) []string {
    var issues []string
    
    // Check if timestamps are reasonable
    now := time.Now()
    
    if data.Quote != nil {
        if data.Quote.EventTime.After(now) {
            issues = append(issues, "Quote timestamp is in the future")
        }
        
        if now.Sub(data.Quote.EventTime) > 7*24*time.Hour {
            issues = append(issues, "Quote data is older than 7 days")
        }
    }
    
    if data.CompanyInfo != nil {
        if data.CompanyInfo.EventTime.After(now) {
            issues = append(issues, "Company info timestamp is in the future")
        }
    }
    
    return issues
}
```

## Quality Monitoring

### 1. Quality Metrics

```go
type QualityMetrics struct {
    TotalRequests    int64
    SuccessfulRequests int64
    FailedRequests   int64
    DataQualityIssues int64
    AverageQualityScore float64
}

func (qm *QualityMetrics) RecordRequest(success bool, qualityScore float64) {
    qm.TotalRequests++
    
    if success {
        qm.SuccessfulRequests++
    } else {
        qm.FailedRequests++
    }
    
    if qualityScore < 0.8 {
        qm.DataQualityIssues++
    }
    
    // Update average quality score
    qm.AverageQualityScore = (qm.AverageQualityScore*float64(qm.TotalRequests-1) + qualityScore) / float64(qm.TotalRequests)
}
```

### 2. Quality Alerts

```go
func checkQualityAlerts(metrics QualityMetrics) []string {
    var alerts []string
    
    // Check success rate
    successRate := float64(metrics.SuccessfulRequests) / float64(metrics.TotalRequests)
    if successRate < 0.9 {
        alerts = append(alerts, fmt.Sprintf("Low success rate: %.2f%%", successRate*100))
    }
    
    // Check data quality
    if metrics.AverageQualityScore < 0.8 {
        alerts = append(alerts, fmt.Sprintf("Low data quality score: %.2f", metrics.AverageQualityScore))
    }
    
    // Check error rate
    errorRate := float64(metrics.FailedRequests) / float64(metrics.TotalRequests)
    if errorRate > 0.1 {
        alerts = append(alerts, fmt.Sprintf("High error rate: %.2f%%", errorRate*100))
    }
    
    return alerts
}
```

## Best Practices

### 1. Always Validate Data

```go
func processStockData(client *yfinance.Client, symbol string) (*StockData, error) {
    // Fetch data
    data := fetchStockData(client, symbol)
    
    // Validate data
    if err := validateStockData(data); err != nil {
        return nil, fmt.Errorf("data validation failed: %w", err)
    }
    
    // Check quality
    qualityScore := calculateDataQualityScore(data)
    if qualityScore.Overall < 0.7 {
        log.Printf("Warning: Low data quality score for %s: %.2f", symbol, qualityScore.Overall)
    }
    
    return data, nil
}
```

### 2. Implement Quality Gates

```go
func qualityGate(data *StockData) error {
    // Must have quote data
    if data.Quote == nil {
        return fmt.Errorf("quote data is required")
    }
    
    // Must have valid price
    if data.Quote.RegularMarketPrice == nil {
        return fmt.Errorf("market price is required")
    }
    
    // Must have valid currency
    if data.Quote.CurrencyCode == "" {
        return fmt.Errorf("currency code is required")
    }
    
    // Must have recent data
    if time.Since(data.Quote.EventTime) > 24*time.Hour {
        return fmt.Errorf("data is too old")
    }
    
    return nil
}
```

### 3. Monitor Data Quality

```go
func monitorDataQuality(data *StockData) {
    qualityScore := calculateDataQualityScore(data)
    
    // Log quality issues
    if len(qualityScore.Issues) > 0 {
        log.Printf("Data quality issues for %s: %v", data.Symbol, qualityScore.Issues)
    }
    
    // Send metrics
    metrics.RecordDataQuality(data.Symbol, qualityScore.Overall)
    
    // Alert if quality is too low
    if qualityScore.Overall < 0.5 {
        alerting.SendAlert(fmt.Sprintf("Critical data quality issue for %s: %.2f", 
            data.Symbol, qualityScore.Overall))
    }
}
```

### 4. Handle Edge Cases

```go
func handleEdgeCases(data *StockData) {
    // Handle zero prices
    if data.Quote != nil && data.Quote.RegularMarketPrice != nil {
        price := float64(data.Quote.RegularMarketPrice.Scaled) / float64(data.Quote.RegularMarketPrice.Scale)
        if price == 0 {
            log.Printf("Warning: Zero price detected for %s", data.Symbol)
            // Mark for manual review
        }
    }
    
    // Handle missing volume
    if data.Quote != nil && data.Quote.RegularMarketVolume == nil {
        log.Printf("Warning: Missing volume for %s", data.Symbol)
        // Use historical average or mark for review
    }
    
    // Handle stale data
    if data.Quote != nil {
        age := time.Since(data.Quote.EventTime)
        if age > 7*24*time.Hour {
            log.Printf("Warning: Stale data for %s: %v old", data.Symbol, age)
        }
    }
}
```

## Next Steps

- [API Reference](api-reference.md) - Complete API documentation
- [Data Structures](data-structures.md) - Detailed data structure guide
- [Complete Examples](examples.md) - Working code examples
- [Error Handling Guide](error-handling.md) - Comprehensive error handling
- [Method Comparison](method-comparison.md) - Method comparison and use cases
