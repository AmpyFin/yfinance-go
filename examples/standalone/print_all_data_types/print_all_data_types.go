package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/AmpyFin/yfinance-go"
)

func main() {
	// Create a new client
	client := yfinance.NewClient()
	ctx := context.Background()
	runID := fmt.Sprintf("all-types-example-%d", time.Now().Unix())

	fmt.Println("=== ALL AMPY-PROTO DATA TYPES CONTENTS ===")
	fmt.Println("Showing contents of all available data types")
	fmt.Println()

	// Example 1: Analysis data contents
	fmt.Println("1. ANALYSIS DATA CONTENTS")
	fmt.Println("=========================")
	analysis, err := client.ScrapeAnalysis(ctx, "AAPL", runID)
	if err != nil {
		log.Printf("Error fetching analysis: %v", err)
	} else {
		fmt.Printf("Analysis for %s (%d line items):\n", analysis.Security.Symbol, len(analysis.Lines))
		fmt.Printf("  Schema: %s\n", analysis.Meta.SchemaVersion)
		fmt.Printf("  Source: %s\n", analysis.Source)
		fmt.Printf("  Producer: %s\n", analysis.Meta.Producer)
		fmt.Printf("  As Of: %s\n", analysis.AsOf.AsTime().Format(time.RFC3339))
		
		// Show all line items (analysis usually has fewer items)
		for i, line := range analysis.Lines {
			fmt.Printf("  Line %d:\n", i+1)
			fmt.Printf("    Key: %s\n", line.Key)
			fmt.Printf("    Value: %d (scale: %d)\n", line.Value.Scaled, line.Value.Scale)
			fmt.Printf("    Currency: %s\n", line.CurrencyCode)
			fmt.Printf("    Period: %s to %s\n", 
				line.PeriodStart.AsTime().Format("2006-01-02"),
				line.PeriodEnd.AsTime().Format("2006-01-02"))
		}
	}

	// Example 2: Analyst Insights data contents
	fmt.Println("\n2. ANALYST INSIGHTS DATA CONTENTS")
	fmt.Println("==================================")
	analystInsights, err := client.ScrapeAnalystInsights(ctx, "AAPL", runID)
	if err != nil {
		log.Printf("Error fetching analyst insights: %v", err)
	} else {
		fmt.Printf("Analyst Insights for %s (%d line items):\n", analystInsights.Security.Symbol, len(analystInsights.Lines))
		fmt.Printf("  Schema: %s\n", analystInsights.Meta.SchemaVersion)
		fmt.Printf("  Source: %s\n", analystInsights.Source)
		fmt.Printf("  Producer: %s\n", analystInsights.Meta.Producer)
		fmt.Printf("  As Of: %s\n", analystInsights.AsOf.AsTime().Format(time.RFC3339))
		
		// Show all line items (analyst insights usually has fewer items)
		for i, line := range analystInsights.Lines {
			fmt.Printf("  Line %d:\n", i+1)
			fmt.Printf("    Key: %s\n", line.Key)
			fmt.Printf("    Value: %d (scale: %d)\n", line.Value.Scaled, line.Value.Scale)
			fmt.Printf("    Currency: %s\n", line.CurrencyCode)
			fmt.Printf("    Period: %s to %s\n", 
				line.PeriodStart.AsTime().Format("2006-01-02"),
				line.PeriodEnd.AsTime().Format("2006-01-02"))
		}
	}

	// Example 3: Balance Sheet data contents
	fmt.Println("\n3. BALANCE SHEET DATA CONTENTS")
	fmt.Println("==============================")
	balanceSheet, err := client.ScrapeBalanceSheet(ctx, "AAPL", runID)
	if err != nil {
		log.Printf("Error fetching balance sheet: %v", err)
	} else {
		fmt.Printf("Balance Sheet for %s (%d line items):\n", balanceSheet.Security.Symbol, len(balanceSheet.Lines))
		fmt.Printf("  Schema: %s\n", balanceSheet.Meta.SchemaVersion)
		fmt.Printf("  Source: %s\n", balanceSheet.Source)
		fmt.Printf("  Producer: %s\n", balanceSheet.Meta.Producer)
		fmt.Printf("  As Of: %s\n", balanceSheet.AsOf.AsTime().Format(time.RFC3339))
		
		// Show first few line items
		for i, line := range balanceSheet.Lines {
			if i >= 5 { // Show only first 5 line items
				fmt.Printf("  ... and %d more line items\n", len(balanceSheet.Lines)-5)
				break
			}
			fmt.Printf("  Line %d:\n", i+1)
			fmt.Printf("    Key: %s\n", line.Key)
			fmt.Printf("    Value: %d (scale: %d)\n", line.Value.Scaled, line.Value.Scale)
			fmt.Printf("    Currency: %s\n", line.CurrencyCode)
			fmt.Printf("    Period: %s to %s\n", 
				line.PeriodStart.AsTime().Format("2006-01-02"),
				line.PeriodEnd.AsTime().Format("2006-01-02"))
		}
	}

	// Example 4: Cash Flow data contents
	fmt.Println("\n4. CASH FLOW DATA CONTENTS")
	fmt.Println("===========================")
	cashFlow, err := client.ScrapeCashFlow(ctx, "AAPL", runID)
	if err != nil {
		log.Printf("Error fetching cash flow: %v", err)
	} else {
		fmt.Printf("Cash Flow for %s (%d line items):\n", cashFlow.Security.Symbol, len(cashFlow.Lines))
		fmt.Printf("  Schema: %s\n", cashFlow.Meta.SchemaVersion)
		fmt.Printf("  Source: %s\n", cashFlow.Source)
		fmt.Printf("  Producer: %s\n", cashFlow.Meta.Producer)
		fmt.Printf("  As Of: %s\n", cashFlow.AsOf.AsTime().Format(time.RFC3339))
		
		// Show first few line items
		for i, line := range cashFlow.Lines {
			if i >= 5 { // Show only first 5 line items
				fmt.Printf("  ... and %d more line items\n", len(cashFlow.Lines)-5)
				break
			}
			fmt.Printf("  Line %d:\n", i+1)
			fmt.Printf("    Key: %s\n", line.Key)
			fmt.Printf("    Value: %d (scale: %d)\n", line.Value.Scaled, line.Value.Scale)
			fmt.Printf("    Currency: %s\n", line.CurrencyCode)
			fmt.Printf("    Period: %s to %s\n", 
				line.PeriodStart.AsTime().Format("2006-01-02"),
				line.PeriodEnd.AsTime().Format("2006-01-02"))
		}
	}

	// Example 5: Show JSON representation of fundamentals data
	fmt.Println("\n5. JSON REPRESENTATION EXAMPLE (ANALYSIS)")
	fmt.Println("==========================================")
	if analysis != nil {
		fmt.Println("Analysis data as JSON:")
		jsonData, err := json.MarshalIndent(analysis, "", "  ")
		if err != nil {
			log.Printf("Error marshaling to JSON: %v", err)
		} else {
			fmt.Println(string(jsonData))
		}
	}

	// Example 6: Show all fundamentals at once with type differentiation
	fmt.Println("\n6. ALL FUNDAMENTALS WITH TYPE DIFFERENTIATION")
	fmt.Println("==============================================")
	allFundamentals, err := client.ScrapeAllFundamentals(ctx, "AAPL", runID)
	if err != nil {
		log.Printf("Error fetching all fundamentals: %v", err)
	} else {
		fmt.Printf("All Fundamentals for AAPL (%d snapshots):\n", len(allFundamentals))
		for i, snapshot := range allFundamentals {
			fmt.Printf("  %d. %s\n", i+1, snapshot.Source)
			fmt.Printf("     Schema: %s\n", snapshot.Meta.SchemaVersion)
			fmt.Printf("     Line Items: %d\n", len(snapshot.Lines))
			fmt.Printf("     As Of: %s\n", snapshot.AsOf.AsTime().Format(time.RFC3339))
			
			// Show first line item as example
			if len(snapshot.Lines) > 0 {
				firstLine := snapshot.Lines[0]
				fmt.Printf("     Example Line: %s = %d (scale: %d) %s\n", 
					firstLine.Key, firstLine.Value.Scaled, firstLine.Value.Scale, firstLine.CurrencyCode)
			}
			fmt.Println()
		}
	}

	fmt.Println("=== ALL DATA TYPES COMPLETE ===")
	fmt.Println("Summary of ampy-proto data types available:")
	fmt.Println("1. Quotes → ampy.ticks.v1.QuoteSnapshot")
	fmt.Println("2. Historical Bars → ampy.bars.v1.BarBatch") 
	fmt.Println("3. News → ampy.news.v1.NewsItem[]")
	fmt.Println("4. Financials → ampy.fundamentals.v1.FundamentalsSnapshot")
	fmt.Println("5. Balance Sheet → ampy.fundamentals.v1.FundamentalsSnapshot")
	fmt.Println("6. Cash Flow → ampy.fundamentals.v1.FundamentalsSnapshot")
	fmt.Println("7. Key Statistics → ampy.fundamentals.v1.FundamentalsSnapshot")
	fmt.Println("8. Analysis → ampy.fundamentals.v1.FundamentalsSnapshot")
	fmt.Println("9. Analyst Insights → ampy.fundamentals.v1.FundamentalsSnapshot")
	fmt.Println()
	fmt.Println("Type differentiation via Source field:")
	fmt.Println("- yfinance/scrape/financials")
	fmt.Println("- yfinance/scrape/balance-sheet")
	fmt.Println("- yfinance/scrape/cash-flow")
	fmt.Println("- yfinance/scrape/key-statistics")
	fmt.Println("- yfinance/scrape/analysis")
	fmt.Println("- yfinance/scrape/analyst-insights")
}
