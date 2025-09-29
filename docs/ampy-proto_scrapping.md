# AMPY-PROTO Scraping Integration Guide

## Overview

This guide explains how to use the yfinance-go scraping system with the AmpyFin ampy-proto protocol for standardized financial data communication. The system now supports comprehensive ampy-proto message generation for all 8 scraping endpoints, enabling seamless integration with financial systems, trading algorithms, and data pipelines.

## What is AMPY-PROTO?

AMPY-PROTO is AmpyFin's standardized protocol for financial data communication. It provides:

- **Standardized Message Formats**: Consistent structure across all financial data types
- **Scaled Decimal Precision**: High-precision financial calculations without floating-point errors
- **Rich Metadata**: Run IDs, timestamps, schema versions, and observability data
- **Security Identification**: Symbol and Market Identifier Code (MIC) standardization
- **Period-based Data**: Time-bounded financial periods for accurate analysis

## Supported Endpoints and Message Types

| Scraping Endpoint | AMPY-PROTO Message Type | Description |
|------------------|------------------------|-------------|
| `financials` | `ampy.fundamentals.v1.FundamentalsSnapshot` | Income statement data |
| `balance-sheet` | `ampy.fundamentals.v1.FundamentalsSnapshot` | Balance sheet data |
| `cash-flow` | `ampy.fundamentals.v1.FundamentalsSnapshot` | Cash flow statement data |
| `key-statistics` | `ampy.fundamentals.v1.FundamentalsSnapshot` | Valuation metrics and ratios |
| `analysis` | `ampy.fundamentals.v1.FundamentalsSnapshot` | Analyst estimates and forecasts |
| `analyst-insights` | `ampy.fundamentals.v1.FundamentalsSnapshot` | Analyst recommendations |
| `profile` | `ampy.profile.v1.ProfileSnapshot` | Company profile and information |
| `news` | `ampy.news.v1.NewsSnapshot` | Financial news and articles |

## Quick Start

### Generate AMPY-PROTO Messages

```bash
# Basic usage - generate ampy-proto messages for a ticker
./yfin scrape --preview-proto --ticker AAPL --config configs/effective.yaml

# Specific endpoints
./yfin scrape --preview-proto --ticker AAPL --endpoints financials,balance-sheet --config configs/effective.yaml

# All endpoints
./yfin scrape --preview-proto --ticker AAPL --endpoints financials,balance-sheet,cash-flow,key-statistics,analysis,analyst-insights,profile,news --config configs/effective.yaml
```

### Example Output

```json
🔍 AMPY-PROTO PREVIEW: ticker=AAPL endpoint=financials
{
  "security": {
    "symbol": "AAPL",
    "mic": "XNAS"
  },
  "meta": {
    "runId": "run_2025_01_29_17_23_27",
    "source": "yfinance-go/scrape",
    "producer": "yfinance-go",
    "schemaVersion": "ampy.fundamentals.v1:2.1.0",
    "timestamp": "2025-01-29T17:23:27Z"
  },
  "lines": [
    {
      "name": "total_revenue",
      "value": {
        "scaled": 394328000000,
        "scale": 0
      },
      "currency": "USD",
      "period": {
        "start": "2024-10-01T00:00:00Z",
        "end": "2024-12-31T23:59:59Z"
      }
    }
  ]
}
```

## Message Structure Deep Dive

### Fundamentals Message Structure

All financial data (financials, balance-sheet, cash-flow, key-statistics, analysis, analyst-insights) uses the `FundamentalsSnapshot` message type:

```json
{
  "security": {
    "symbol": "AAPL",           // Stock symbol
    "mic": "XNAS"              // Market Identifier Code
  },
  "meta": {
    "runId": "run_2025_01_29_17_23_27",  // Unique run identifier
    "source": "yfinance-go/scrape",       // Data source
    "producer": "yfinance-go",            // Producer identifier
    "schemaVersion": "ampy.fundamentals.v1:2.1.0",  // Schema version
    "timestamp": "2025-01-29T17:23:27Z"  // Generation timestamp
  },
  "lines": [                    // Financial line items
    {
      "name": "total_revenue",  // Line item name
      "value": {
        "scaled": 394328000000, // Scaled integer value
        "scale": 0              // Decimal scale (0 = no decimals)
      },
      "currency": "USD",        // Currency code
      "period": {               // Time period
        "start": "2024-10-01T00:00:00Z",
        "end": "2024-12-31T23:59:59Z"
      }
    }
  ]
}
```

### Profile Message Structure

Company profile data uses the `ProfileSnapshot` message type:

```json
{
  "security": {
    "symbol": "AAPL",
    "mic": "XNAS"
  },
  "meta": {
    "runId": "run_2025_01_29_17_23_27",
    "source": "yfinance-go/scrape",
    "producer": "yfinance-go",
    "schemaVersion": "ampy.profile.v1:2.1.0",
    "timestamp": "2025-01-29T17:23:27Z"
  },
  "company": {
    "name": "Apple Inc.",
    "description": "Apple Inc. designs, manufactures, and markets smartphones...",
    "sector": "Technology",
    "industry": "Consumer Electronics",
    "website": "https://www.apple.com",
    "employees": 164000,
    "headquarters": {
      "address": "One Apple Park Way",
      "city": "Cupertino",
      "state": "CA",
      "country": "United States"
    }
  }
}
```

### News Message Structure

News data uses the `NewsSnapshot` message type:

```json
{
  "security": {
    "symbol": "AAPL",
    "mic": "XNAS"
  },
  "meta": {
    "runId": "run_2025_01_29_17_23_27",
    "source": "yfinance-go/scrape",
    "producer": "yfinance-go",
    "schemaVersion": "ampy.news.v1:2.1.0",
    "timestamp": "2025-01-29T17:23:27Z"
  },
  "articles": [
    {
      "title": "Apple Just Unveiled the iPhone 17: Here's What's New",
      "url": "https://finance.yahoo.com/news/apple-iphone-17-unveiled-whats-new-140000123.html",
      "source": "Yahoo Finance",
      "publishedAt": "2025-01-29T15:23:27Z",
      "imageUrl": "https://media.zenfs.com/en/yahoo_finance_350/iphone-17-launch.webp",
      "relatedTickers": ["AAPL"]
    }
  ],
  "stats": {
    "totalFound": 18,
    "totalReturned": 18,
    "deduped": 0
  }
}
```

## Financial Data Line Items

### Available Line Items by Endpoint

#### Financials Endpoint
- `total_revenue` - Total revenue
- `operating_income` - Operating income
- `net_income` - Net income
- `ebitda` - Earnings before interest, taxes, depreciation, amortization
- `basic_eps` - Basic earnings per share
- `diluted_eps` - Diluted earnings per share

#### Balance Sheet Endpoint
- `total_assets` - Total assets
- `total_debt` - Total debt
- `shareholders_equity` - Shareholders' equity
- `working_capital` - Working capital
- `tangible_book_value` - Tangible book value
- `cash_and_equivalents` - Cash and cash equivalents

#### Cash Flow Endpoint
- `operating_cash_flow` - Operating cash flow
- `investing_cash_flow` - Investing cash flow
- `financing_cash_flow` - Financing cash flow
- `free_cash_flow` - Free cash flow
- `capital_expenditure` - Capital expenditure

#### Key Statistics Endpoint
- `market_cap` - Market capitalization
- `enterprise_value` - Enterprise value
- `forward_pe` - Forward price-to-earnings ratio
- `trailing_pe` - Trailing price-to-earnings ratio
- `peg_ratio` - Price/earnings-to-growth ratio
- `price_sales` - Price-to-sales ratio
- `price_book` - Price-to-book ratio
- `beta` - Beta (volatility measure)
- `shares_outstanding` - Shares outstanding
- `profit_margin` - Profit margin
- `operating_margin` - Operating margin
- `return_on_assets` - Return on assets
- `return_on_equity` - Return on equity

#### Analysis Endpoint
- `eps_estimate_current_year` - Current year EPS estimate
- `eps_estimate_next_year` - Next year EPS estimate
- `revenue_estimate_current_year` - Current year revenue estimate
- `revenue_estimate_next_year` - Next year revenue estimate
- `growth_estimate_current_year` - Current year growth estimate
- `growth_estimate_next_year` - Next year growth estimate

#### Analyst Insights Endpoint
- `price_target_high` - High price target
- `price_target_low` - Low price target
- `price_target_median` - Median price target
- `price_target_average` - Average price target
- `recommendation_score` - Recommendation score (1=Strong Buy, 5=Strong Sell)
- `number_of_analysts` - Number of analysts covering

## Scaled Decimal Values

AMPY-PROTO uses scaled decimal values to avoid floating-point precision issues:

```go
// Example: $394.328 billion revenue
{
  "scaled": 394328000000,  // Integer value
  "scale": 0               // No decimal places
}

// Example: 31.75 P/E ratio
{
  "scaled": 3175,          // Integer value (3175)
  "scale": 2               // 2 decimal places (31.75)
}

// Example: 24.30% profit margin
{
  "scaled": 2430,          // Integer value (2430)
  "scale": 2               // 2 decimal places (24.30%)
}
```

### Converting Scaled Values

```go
func convertScaledValue(scaled int64, scale int32) float64 {
    multiplier := math.Pow10(int(scale))
    return float64(scaled) / multiplier
}

// Examples:
// convertScaledValue(394328000000, 0) = 394328000000.0
// convertScaledValue(3175, 2) = 31.75
// convertScaledValue(2430, 2) = 24.30
```

## System Integration Examples

### 1. Debt Analysis System

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"
    fundamentalsv1 "github.com/AmpyFin/ampy-proto/v2/gen/go/ampy/fundamentals/v1"
)

type DebtAnalysis struct {
    TotalDebt        int64   `json:"total_debt"`
    TotalAssets      int64   `json:"total_assets"`
    ShareholdersEquity int64 `json:"shareholders_equity"`
    DebtToEquity     float64 `json:"debt_to_equity"`
    DebtToAssets     float64 `json:"debt_to_assets"`
    RiskLevel        string  `json:"risk_level"`
}

func analyzeDebtFromProto(protoData []byte) (*DebtAnalysis, error) {
    var snapshot fundamentalsv1.FundamentalsSnapshot
    if err := json.Unmarshal(protoData, &snapshot); err != nil {
        return nil, err
    }
    
    var totalDebt, totalAssets, shareholdersEquity int64
    
    for _, line := range snapshot.Lines {
        switch line.Name {
        case "total_debt":
            totalDebt = line.Value.Scaled
        case "total_assets":
            totalAssets = line.Value.Scaled
        case "shareholders_equity":
            shareholdersEquity = line.Value.Scaled
        }
    }
    
    // Calculate ratios
    debtToEquity := float64(totalDebt) / float64(shareholdersEquity)
    debtToAssets := float64(totalDebt) / float64(totalAssets)
    
    // Assess risk level
    riskLevel := "LOW"
    if debtToEquity > 0.5 {
        riskLevel = "HIGH"
    } else if debtToEquity > 0.3 {
        riskLevel = "MEDIUM"
    }
    
    return &DebtAnalysis{
        TotalDebt: totalDebt,
        TotalAssets: totalAssets,
        ShareholdersEquity: shareholdersEquity,
        DebtToEquity: debtToEquity,
        DebtToAssets: debtToAssets,
        RiskLevel: riskLevel,
    }, nil
}

func main() {
    // This would be called with actual proto data from yfinance-go
    fmt.Println("Debt Analysis System Ready")
    fmt.Println("Use: ./yfin scrape --preview-proto --ticker AAPL --endpoints balance-sheet")
}
```

### 2. Revenue Analysis System

```go
package main

import (
    "encoding/json"
    "fmt"
    fundamentalsv1 "github.com/AmpyFin/ampy-proto/v2/gen/go/ampy/fundamentals/v1"
)

type RevenueAnalysis struct {
    TotalRevenue     int64   `json:"total_revenue"`
    OperatingIncome  int64   `json:"operating_income"`
    NetIncome        int64   `json:"net_income"`
    OperatingMargin  float64 `json:"operating_margin"`
    NetMargin        float64 `json:"net_margin"`
    GrowthTrend      string  `json:"growth_trend"`
}

func analyzeRevenueFromProto(protoData []byte) (*RevenueAnalysis, error) {
    var snapshot fundamentalsv1.FundamentalsSnapshot
    if err := json.Unmarshal(protoData, &snapshot); err != nil {
        return nil, err
    }
    
    var totalRevenue, operatingIncome, netIncome int64
    
    for _, line := range snapshot.Lines {
        switch line.Name {
        case "total_revenue":
            totalRevenue = line.Value.Scaled
        case "operating_income":
            operatingIncome = line.Value.Scaled
        case "net_income":
            netIncome = line.Value.Scaled
        }
    }
    
    // Calculate margins
    operatingMargin := float64(operatingIncome) / float64(totalRevenue)
    netMargin := float64(netIncome) / float64(totalRevenue)
    
    // Determine growth trend (simplified)
    growthTrend := "STABLE"
    if netMargin > 0.2 {
        growthTrend = "STRONG"
    } else if netMargin < 0.1 {
        growthTrend = "WEAK"
    }
    
    return &RevenueAnalysis{
        TotalRevenue: totalRevenue,
        OperatingIncome: operatingIncome,
        NetIncome: netIncome,
        OperatingMargin: operatingMargin,
        NetMargin: netMargin,
        GrowthTrend: growthTrend,
    }, nil
}
```

### 3. News Sentiment System

```go
package main

import (
    "encoding/json"
    "fmt"
    "strings"
    "time"
    newsv1 "github.com/AmpyFin/ampy-proto/v2/gen/go/ampy/news/v1"
)

type NewsAnalysis struct {
    RecentArticles   int     `json:"recent_articles"`
    AverageSentiment float64 `json:"average_sentiment"`
    SentimentTrend   string  `json:"sentiment_trend"`
    KeyTopics        []string `json:"key_topics"`
}

func analyzeNewsFromProto(protoData []byte) (*NewsAnalysis, error) {
    var snapshot newsv1.NewsSnapshot
    if err := json.Unmarshal(protoData, &snapshot); err != nil {
        return nil, err
    }
    
    var sentiment float64
    var recentNews int
    var topics []string
    
    cutoff := time.Now().Add(-24 * time.Hour)
    
    for _, article := range snapshot.Articles {
        // Check if article is recent
        if article.PublishedAt.AsTime().After(cutoff) {
            recentNews++
            sentiment += calculateArticleSentiment(article.Title)
        }
        
        // Extract topics (simplified)
        topics = append(topics, extractTopics(article.Title)...)
    }
    
    avgSentiment := sentiment / float64(len(snapshot.Articles))
    
    // Determine sentiment trend
    sentimentTrend := "NEUTRAL"
    if avgSentiment > 0.1 {
        sentimentTrend = "POSITIVE"
    } else if avgSentiment < -0.1 {
        sentimentTrend = "NEGATIVE"
    }
    
    return &NewsAnalysis{
        RecentArticles: recentNews,
        AverageSentiment: avgSentiment,
        SentimentTrend: sentimentTrend,
        KeyTopics: uniqueTopics(topics),
    }, nil
}

func calculateArticleSentiment(title string) float64 {
    positiveWords := []string{"beat", "exceed", "growth", "strong", "bullish", "up", "rise"}
    negativeWords := []string{"miss", "decline", "weak", "bearish", "down", "fall", "drop"}
    
    title = strings.ToLower(title)
    sentiment := 0.0
    
    for _, word := range positiveWords {
        if strings.Contains(title, word) {
            sentiment += 0.1
        }
    }
    
    for _, word := range negativeWords {
        if strings.Contains(title, word) {
            sentiment -= 0.1
        }
    }
    
    return sentiment
}

func extractTopics(title string) []string {
    // Simplified topic extraction
    topics := []string{}
    if strings.Contains(strings.ToLower(title), "iphone") {
        topics = append(topics, "iPhone")
    }
    if strings.Contains(strings.ToLower(title), "earnings") {
        topics = append(topics, "Earnings")
    }
    if strings.Contains(strings.ToLower(title), "revenue") {
        topics = append(topics, "Revenue")
    }
    return topics
}

func uniqueTopics(topics []string) []string {
    seen := make(map[string]bool)
    result := []string{}
    for _, topic := range topics {
        if !seen[topic] {
            seen[topic] = true
            result = append(result, topic)
        }
    }
    return result
}
```

### 4. Multi-System Integration

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "os/exec"
    fundamentalsv1 "github.com/AmpyFin/ampy-proto/v2/gen/go/ampy/fundamentals/v1"
    newsv1 "github.com/AmpyFin/ampy-proto/v2/gen/go/ampy/news/v1"
    profilev1 "github.com/AmpyFin/ampy-proto/v2/gen/go/ampy/profile/v1"
)

type ComprehensiveAnalysis struct {
    Ticker     string         `json:"ticker"`
    Debt       *DebtAnalysis  `json:"debt"`
    Revenue    *RevenueAnalysis `json:"revenue"`
    News       *NewsAnalysis  `json:"news"`
    Profile    *ProfileData   `json:"profile"`
    RiskScore  float64        `json:"risk_score"`
    Timestamp  string         `json:"timestamp"`
}

type ProfileData struct {
    Name        string `json:"name"`
    Sector      string `json:"sector"`
    Industry    string `json:"industry"`
    Employees   int64  `json:"employees"`
    Website     string `json:"website"`
}

type FinancialSystem struct {
    debtAnalyzer    *DebtAnalyzer
    revenueAnalyzer *RevenueAnalyzer
    newsAnalyzer    *NewsAnalyzer
}

func (fs *FinancialSystem) ProcessTicker(ticker string) (*ComprehensiveAnalysis, error) {
    // Get fundamentals data
    fundamentals, err := fs.getFundamentals(ticker)
    if err != nil {
        return nil, err
    }
    
    // Get news data
    news, err := fs.getNews(ticker)
    if err != nil {
        return nil, err
    }
    
    // Get profile data
    profile, err := fs.getProfile(ticker)
    if err != nil {
        return nil, err
    }
    
    // Analyze each component
    debtAnalysis, err := fs.debtAnalyzer.Analyze(fundamentals)
    if err != nil {
        return nil, err
    }
    
    revenueAnalysis, err := fs.revenueAnalyzer.Analyze(fundamentals)
    if err != nil {
        return nil, err
    }
    
    newsAnalysis, err := fs.newsAnalyzer.Analyze(news)
    if err != nil {
        return nil, err
    }
    
    // Calculate risk score
    riskScore := fs.calculateRiskScore(debtAnalysis, revenueAnalysis, newsAnalysis)
    
    return &ComprehensiveAnalysis{
        Ticker: ticker,
        Debt: debtAnalysis,
        Revenue: revenueAnalysis,
        News: newsAnalysis,
        Profile: profile,
        RiskScore: riskScore,
        Timestamp: time.Now().Format(time.RFC3339),
    }, nil
}

func (fs *FinancialSystem) getFundamentals(ticker string) ([]byte, error) {
    cmd := exec.Command("./yfin", "scrape", "--preview-proto", 
        "--ticker", ticker, 
        "--endpoints", "financials,balance-sheet",
        "--config", "configs/effective.yaml")
    
    output, err := cmd.Output()
    if err != nil {
        return nil, err
    }
    
    return output, nil
}

func (fs *FinancialSystem) getNews(ticker string) ([]byte, error) {
    cmd := exec.Command("./yfin", "scrape", "--preview-proto", 
        "--ticker", ticker, 
        "--endpoints", "news",
        "--config", "configs/effective.yaml")
    
    output, err := cmd.Output()
    if err != nil {
        return nil, err
    }
    
    return output, nil
}

func (fs *FinancialSystem) getProfile(ticker string) (*ProfileData, error) {
    cmd := exec.Command("./yfin", "scrape", "--preview-proto", 
        "--ticker", ticker, 
        "--endpoints", "profile",
        "--config", "configs/effective.yaml")
    
    output, err := cmd.Output()
    if err != nil {
        return nil, err
    }
    
    var snapshot profilev1.ProfileSnapshot
    if err := json.Unmarshal(output, &snapshot); err != nil {
        return nil, err
    }
    
    return &ProfileData{
        Name: snapshot.Company.Name,
        Sector: snapshot.Company.Sector,
        Industry: snapshot.Company.Industry,
        Employees: snapshot.Company.Employees,
        Website: snapshot.Company.Website,
    }, nil
}

func (fs *FinancialSystem) calculateRiskScore(debt *DebtAnalysis, revenue *RevenueAnalysis, news *NewsAnalysis) float64 {
    score := 0.0
    
    // Debt risk (0-40 points)
    if debt.DebtToEquity > 0.5 {
        score += 40
    } else if debt.DebtToEquity > 0.3 {
        score += 20
    }
    
    // Revenue risk (0-30 points)
    if revenue.NetMargin < 0.1 {
        score += 30
    } else if revenue.NetMargin < 0.15 {
        score += 15
    }
    
    // News sentiment risk (0-30 points)
    if news.SentimentTrend == "NEGATIVE" {
        score += 30
    } else if news.SentimentTrend == "NEUTRAL" {
        score += 15
    }
    
    return score
}

func main() {
    system := &FinancialSystem{
        debtAnalyzer: &DebtAnalyzer{},
        revenueAnalyzer: &RevenueAnalyzer{},
        newsAnalyzer: &NewsAnalyzer{},
    }
    
    analysis, err := system.ProcessTicker("AAPL")
    if err != nil {
        log.Fatal(err)
    }
    
    output, err := json.MarshalIndent(analysis, "", "  ")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println(string(output))
}
```

## Real-time Data Streaming

### Stream Processing with AMPY-PROTO

```go
package main

import (
    "bufio"
    "encoding/json"
    "fmt"
    "log"
    "os/exec"
    "strings"
)

type MessageRouter struct {
    riskSystem      chan []byte
    analyticsSystem chan []byte
    sentimentSystem chan []byte
    alertSystem     chan []byte
}

func (mr *MessageRouter) StreamTicker(ticker string) error {
    cmd := exec.Command("./yfin", "scrape", "--preview-proto", 
        "--ticker", ticker, 
        "--endpoints", "financials,balance-sheet,cash-flow,news",
        "--config", "configs/effective.yaml")
    
    stdout, err := cmd.StdoutPipe()
    if err != nil {
        return err
    }
    
    if err := cmd.Start(); err != nil {
        return err
    }
    
    scanner := bufio.NewScanner(stdout)
    for scanner.Scan() {
        line := scanner.Text()
        
        // Parse ampy-proto message
        if strings.Contains(line, "🔍 AMPY-PROTO PREVIEW") {
            continue // Skip header lines
        }
        
        if strings.HasPrefix(line, "{") {
            // This is a JSON message
            var message map[string]interface{}
            if err := json.Unmarshal([]byte(line), &message); err != nil {
                log.Printf("Error parsing message: %v", err)
                continue
            }
            
            // Route message based on schema version
            schemaVersion, ok := message["meta"].(map[string]interface{})["schemaVersion"].(string)
            if !ok {
                continue
            }
            
            switch {
            case strings.Contains(schemaVersion, "fundamentals"):
                mr.routeToRiskSystem([]byte(line))
                mr.routeToAnalyticsSystem([]byte(line))
            case strings.Contains(schemaVersion, "news"):
                mr.routeToSentimentSystem([]byte(line))
                mr.routeToAlertSystem([]byte(line))
            }
        }
    }
    
    return cmd.Wait()
}

func (mr *MessageRouter) routeToRiskSystem(data []byte) {
    select {
    case mr.riskSystem <- data:
    default:
        log.Println("Risk system channel full, dropping message")
    }
}

func (mr *MessageRouter) routeToAnalyticsSystem(data []byte) {
    select {
    case mr.analyticsSystem <- data:
    default:
        log.Println("Analytics system channel full, dropping message")
    }
}

func (mr *MessageRouter) routeToSentimentSystem(data []byte) {
    select {
    case mr.sentimentSystem <- data:
    default:
        log.Println("Sentiment system channel full, dropping message")
    }
}

func (mr *MessageRouter) routeToAlertSystem(data []byte) {
    select {
    case mr.alertSystem <- data:
    default:
        log.Println("Alert system channel full, dropping message")
    }
}

func main() {
    router := &MessageRouter{
        riskSystem:      make(chan []byte, 100),
        analyticsSystem: make(chan []byte, 100),
        sentimentSystem: make(chan []byte, 100),
        alertSystem:     make(chan []byte, 100),
    }
    
    // Start processing systems
    go processRiskSystem(router.riskSystem)
    go processAnalyticsSystem(router.analyticsSystem)
    go processSentimentSystem(router.sentimentSystem)
    go processAlertSystem(router.alertSystem)
    
    // Stream data for multiple tickers
    tickers := []string{"AAPL", "MSFT", "GOOGL", "TSLA"}
    
    for _, ticker := range tickers {
        if err := router.StreamTicker(ticker); err != nil {
            log.Printf("Error streaming %s: %v", ticker, err)
        }
    }
}

func processRiskSystem(ch <-chan []byte) {
    for data := range ch {
        // Process risk analysis
        fmt.Printf("Risk System: Processing %d bytes\n", len(data))
    }
}

func processAnalyticsSystem(ch <-chan []byte) {
    for data := range ch {
        // Process analytics
        fmt.Printf("Analytics System: Processing %d bytes\n", len(data))
    }
}

func processSentimentSystem(ch <-chan []byte) {
    for data := range ch {
        // Process sentiment analysis
        fmt.Printf("Sentiment System: Processing %d bytes\n", len(data))
    }
}

func processAlertSystem(ch <-chan []byte) {
    for data := range ch {
        // Process alerts
        fmt.Printf("Alert System: Processing %d bytes\n", len(data))
    }
}
```

## Best Practices

### 1. Error Handling

```go
func processProtoMessage(data []byte) error {
    var message map[string]interface{}
    if err := json.Unmarshal(data, &message); err != nil {
        return fmt.Errorf("failed to unmarshal proto message: %w", err)
    }
    
    // Validate required fields
    if _, ok := message["security"]; !ok {
        return fmt.Errorf("missing security field")
    }
    
    if _, ok := message["meta"]; !ok {
        return fmt.Errorf("missing meta field")
    }
    
    return nil
}
```

### 2. Performance Optimization

```go
// Use connection pooling for multiple requests
type ProtoClient struct {
    pool chan *exec.Cmd
}

func (pc *ProtoClient) GetFundamentals(ticker string) ([]byte, error) {
    cmd := <-pc.pool
    defer func() { pc.pool <- cmd }()
    
    cmd.Args = []string{"./yfin", "scrape", "--preview-proto", 
        "--ticker", ticker, 
        "--endpoints", "financials",
        "--config", "configs/effective.yaml"}
    
    return cmd.Output()
}
```

### 3. Monitoring and Observability

```go
type ProtoMetrics struct {
    MessagesProcessed int64
    ProcessingTime    time.Duration
    Errors           int64
}

func (pm *ProtoMetrics) RecordMessage(processingTime time.Duration, err error) {
    pm.MessagesProcessed++
    pm.ProcessingTime += processingTime
    
    if err != nil {
        pm.Errors++
    }
}
```

## Troubleshooting

### Common Issues

1. **Schema Version Mismatch**
   ```
   Error: unsupported schema version ampy.fundamentals.v1:1.0.0
   ```
   **Solution**: Update your ampy-proto dependency to match the schema version.

2. **Missing Line Items**
   ```
   Warning: line item 'total_revenue' not found
   ```
   **Solution**: Check if the endpoint supports the requested line item.

3. **Invalid Scaled Values**
   ```
   Error: invalid scaled value for 'market_cap'
   ```
   **Solution**: Validate scaled values before processing.

### Debug Mode

```bash
# Enable debug logging
./yfin scrape --preview-proto --ticker AAPL --endpoints financials --config configs/effective.yaml --debug

# Check specific endpoint
./yfin scrape --preview-proto --ticker AAPL --endpoints balance-sheet --config configs/effective.yaml --verbose
```

## Summary

The yfinance-go scraping system now provides comprehensive ampy-proto integration, enabling:

- **Standardized Communication**: All 8 endpoints generate ampy-proto messages
- **High Precision**: Scaled decimal values for accurate financial calculations
- **Rich Metadata**: Complete observability and traceability
- **System Integration**: Easy integration with financial systems and trading algorithms
- **Real-time Processing**: Stream processing capabilities for live data feeds

This integration makes yfinance-go a powerful tool for building financial data pipelines and trading systems that communicate using the AmpyFin protocol standard.
