//go:build ignore

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/AmpyFin/yfinance-go"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type NVDAAllData struct {
	Timestamp    string      `json:"timestamp"`
	Ticker       string      `json:"ticker"`
	Quote        interface{} `json:"quote,omitempty"`
	Historical   interface{} `json:"historical_bars,omitempty"`
	KeyStats     interface{} `json:"key_statistics,omitempty"`
	Financials   interface{} `json:"financials,omitempty"`
	BalanceSheet interface{} `json:"balance_sheet,omitempty"`
	CashFlow     interface{} `json:"cash_flow,omitempty"`
	Analysis     interface{} `json:"analysis,omitempty"`
	AnalystInsights interface{} `json:"analyst_insights,omitempty"`
	Profile      interface{} `json:"profile,omitempty"`
	News         interface{} `json:"news,omitempty"`
	CompanyInfo interface{} `json:"company_info,omitempty"`
}

func main() {
	ctx := context.Background()
	client := yfinance.NewClient()
	ticker := "NVDA"
	runID := fmt.Sprintf("nvda-full-data-%d", time.Now().Unix())

	// Redirect all logs to stderr so stdout is clean JSON only
	fmt.Fprintf(os.Stderr, "Fetching all data for %s...\n", ticker)

	allData := NVDAAllData{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Ticker:    ticker,
	}

	// 1. Fetch Quote
	fmt.Fprintf(os.Stderr, "Fetching quote...\n")
	quote, err := client.FetchQuote(ctx, ticker, runID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching quote: %v\n", err)
	} else {
		allData.Quote = quote
	}

	// 2. Fetch Historical Data (1 year)
	fmt.Fprintf(os.Stderr, "Fetching historical data...\n")
	end := time.Now().UTC()
	start := end.AddDate(-1, 0, 0)
	bars, err := client.FetchDailyBars(ctx, ticker, start, end, true, runID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching historical data: %v\n", err)
	} else {
		allData.Historical = bars
	}

	// 3. Fetch Company Info
	fmt.Fprintf(os.Stderr, "Fetching company info...\n")
	companyInfo, err := client.FetchCompanyInfo(ctx, ticker, runID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching company info: %v\n", err)
	} else {
		allData.CompanyInfo = companyInfo
	}

	// Helper function to convert protobuf to JSON
	protoToJSON := func(pb interface{}) interface{} {
		if pb == nil {
			return nil
		}
		// Use protojson to marshal protobuf message
		// All protobuf messages implement proto.Message interface
		jsonBytes, err := protojson.MarshalOptions{
			Multiline:       false,
			EmitUnpopulated: true,
			UseProtoNames:   true,
		}.Marshal(pb.(proto.Message))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Error marshaling protobuf: %v\n", err)
			return nil
		}
		var result interface{}
		if err := json.Unmarshal(jsonBytes, &result); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Error unmarshaling JSON: %v\n", err)
			return nil
		}
		return result
	}

	// 4. Scrape Key Statistics
	fmt.Fprintf(os.Stderr, "Scraping key statistics...\n")
	keyStats, err := client.ScrapeKeyStatistics(ctx, ticker, runID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scraping key statistics: %v\n", err)
	} else {
		allData.KeyStats = protoToJSON(keyStats)
	}

	// 5. Scrape Financials
	fmt.Fprintf(os.Stderr, "Scraping financials...\n")
	financials, err := client.ScrapeFinancials(ctx, ticker, runID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scraping financials: %v\n", err)
	} else {
		allData.Financials = protoToJSON(financials)
	}

	// 6. Scrape Balance Sheet
	fmt.Fprintf(os.Stderr, "Scraping balance sheet...\n")
	balanceSheet, err := client.ScrapeBalanceSheet(ctx, ticker, runID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scraping balance sheet: %v\n", err)
	} else {
		allData.BalanceSheet = protoToJSON(balanceSheet)
	}

	// 7. Scrape Cash Flow
	fmt.Fprintf(os.Stderr, "Scraping cash flow...\n")
	cashFlow, err := client.ScrapeCashFlow(ctx, ticker, runID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scraping cash flow: %v\n", err)
	} else {
		allData.CashFlow = protoToJSON(cashFlow)
	}

	// 8. Scrape Analysis
	fmt.Fprintf(os.Stderr, "Scraping analysis...\n")
	analysis, err := client.ScrapeAnalysis(ctx, ticker, runID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scraping analysis: %v\n", err)
	} else {
		allData.Analysis = protoToJSON(analysis)
	}

	// 9. Scrape Analyst Insights
	fmt.Fprintf(os.Stderr, "Scraping analyst insights...\n")
	analystInsights, err := client.ScrapeAnalystInsights(ctx, ticker, runID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scraping analyst insights: %v\n", err)
	} else {
		allData.AnalystInsights = protoToJSON(analystInsights)
	}

	// 10. Scrape News
	fmt.Fprintf(os.Stderr, "Scraping news...\n")
	news, err := client.ScrapeNews(ctx, ticker, runID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scraping news: %v\n", err)
	} else {
		// Convert news slice to JSON
		newsJSON := make([]interface{}, 0, len(news))
		for _, item := range news {
			newsJSON = append(newsJSON, protoToJSON(item))
		}
		allData.News = newsJSON
	}

	// Convert to JSON
	jsonData, err := json.MarshalIndent(allData, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling to JSON: %v\n", err)
		os.Exit(1)
	}

	// Output to stdout
	fmt.Println(string(jsonData))
}

