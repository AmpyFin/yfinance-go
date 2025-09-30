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
	runID := fmt.Sprintf("historical-example-%d", time.Now().Unix())

	fmt.Println("=== HISTORICAL DATA API EXAMPLE ===")
	fmt.Println("Demonstrating programmatic access to historical ampy-proto data")
	fmt.Println()

	// Example 1: Fetch daily bars
	fmt.Println("1. Fetching Daily Bars...")
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)

	dailyBars, err := client.FetchDailyBars(ctx, "AAPL", start, end, true, runID)
	if err != nil {
		log.Printf("Error fetching daily bars: %v", err)
	} else {
		fmt.Printf("✅ Daily Bars: %d bars, schema: %s\n",
			len(dailyBars.Bars), dailyBars.Meta.SchemaVersion)
		fmt.Printf("   Source: %s\n", dailyBars.Meta.Source)
		fmt.Printf("   Producer: %s\n", dailyBars.Meta.Producer)
	}

	// Example 2: Fetch quote
	fmt.Println("\n2. Fetching Quote...")
	quote, err := client.FetchQuote(ctx, "AAPL", runID)
	if err != nil {
		log.Printf("Error fetching quote: %v", err)
	} else {
		fmt.Printf("✅ Quote: schema: %s\n", quote.Meta.SchemaVersion)
		fmt.Printf("   Source: %s\n", quote.Meta.Source)
		fmt.Printf("   Producer: %s\n", quote.Meta.Producer)
	}

	// Example 3: Fetch intraday bars
	fmt.Println("\n3. Fetching Intraday Bars (1m)...")
	intradayBars, err := client.FetchIntradayBars(ctx, "AAPL", start, end, "1m", runID)
	if err != nil {
		log.Printf("Error fetching intraday bars: %v", err)
	} else {
		fmt.Printf("✅ Intraday Bars: %d bars, schema: %s\n",
			len(intradayBars.Bars), intradayBars.Meta.SchemaVersion)
		fmt.Printf("   Source: %s\n", intradayBars.Meta.Source)
		fmt.Printf("   Producer: %s\n", intradayBars.Meta.Producer)
	}

	fmt.Println("\n=== HISTORICAL DATA EXAMPLE COMPLETE ===")
	fmt.Println("All historical data is returned as ampy-proto messages:")
	fmt.Println("- Daily/Weekly/Monthly bars → ampy.bars.v1.BarBatch")
	fmt.Println("- Quotes → ampy.quotes.v1.QuoteSnapshot")
	fmt.Println("- Intraday bars → ampy.bars.v1.BarBatch")
	fmt.Println("- Company info → ampy.company.v1.CompanySnapshot")
	fmt.Println("- Market data → ampy.market.v1.MarketSnapshot")
}
