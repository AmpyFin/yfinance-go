package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/AmpyFin/yfinance-go"
)

func main() {
	// Create a new client
	client := yfinance.NewClient()
	ctx := context.Background()
	runID := fmt.Sprintf("example-%d", time.Now().Unix())

	fmt.Println("=== YFINANCE-GO API EXAMPLE ===")
	fmt.Println("Demonstrating programmatic access to ampy-proto data")
	fmt.Println()

	// Example 1: Fetch financials data
	fmt.Println("1. Fetching Financials Data...")
	financials, err := client.ScrapeFinancials(ctx, "AAPL", runID)
	if err != nil {
		log.Printf("Error fetching financials: %v", err)
	} else {
		fmt.Printf("✅ Financials: %d line items, source: %s\n",
			len(financials.Lines), financials.Source)
	}

	// Example 2: Fetch key statistics
	fmt.Println("\n2. Fetching Key Statistics...")
	keyStats, err := client.ScrapeKeyStatistics(ctx, "AAPL", runID)
	if err != nil {
		log.Printf("Error fetching key statistics: %v", err)
	} else {
		fmt.Printf("✅ Key Statistics: %d line items, source: %s\n",
			len(keyStats.Lines), keyStats.Source)
	}

	// Example 3: Fetch analysis data
	fmt.Println("\n3. Fetching Analysis Data...")
	analysis, err := client.ScrapeAnalysis(ctx, "AAPL", runID)
	if err != nil {
		log.Printf("Error fetching analysis: %v", err)
	} else {
		fmt.Printf("✅ Analysis: %d line items, source: %s\n",
			len(analysis.Lines), analysis.Source)
	}

	// Example 4: Fetch news data
	fmt.Println("\n4. Fetching News Data...")
	news, err := client.ScrapeNews(ctx, "AAPL", runID)
	if err != nil {
		log.Printf("Error fetching news: %v", err)
	} else {
		fmt.Printf("✅ News: %d articles\n", len(news))
	}

	// Example 5: Fetch all fundamentals at once
	fmt.Println("\n5. Fetching All Fundamentals...")
	allFundamentals, err := client.ScrapeAllFundamentals(ctx, "AAPL", runID)
	if err != nil {
		log.Printf("Error fetching all fundamentals: %v", err)
	} else {
		fmt.Printf("✅ All Fundamentals: %d snapshots\n", len(allFundamentals))
		for i, snapshot := range allFundamentals {
			fmt.Printf("   %d. %s (%d line items)\n",
				i+1, snapshot.Source, len(snapshot.Lines))
		}
	}

	fmt.Println("\n=== API EXAMPLE COMPLETE ===")
	fmt.Println("All data is returned as ampy-proto messages that can be:")
	fmt.Println("- Serialized to protobuf")
	fmt.Println("- Published to message queues")
	fmt.Println("- Stored in databases")
	fmt.Println("- Processed by other ampy-proto compatible systems")
}
