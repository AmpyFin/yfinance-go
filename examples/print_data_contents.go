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
	runID := fmt.Sprintf("contents-example-%d", time.Now().Unix())

	fmt.Println("=== AMPY-PROTO DATA CONTENTS ===")
	fmt.Println("Showing actual contents of returned ampy-proto messages")
	fmt.Println()

	// Example 1: Quote data contents
	fmt.Println("1. QUOTE DATA CONTENTS")
	fmt.Println("======================")
	quote, err := client.FetchQuote(ctx, "AAPL", runID)
	if err != nil {
		log.Printf("Error fetching quote: %v", err)
	} else {
		fmt.Printf("Quote for %s:\n", quote.Security.Symbol)
		fmt.Printf("  Schema: %s\n", quote.Meta.SchemaVersion)
		fmt.Printf("  Source: %s\n", quote.Meta.Source)
		fmt.Printf("  Producer: %s\n", quote.Meta.Producer)
		fmt.Printf("  Currency: %s\n", quote.CurrencyCode)
		fmt.Printf("  Type: %s\n", quote.Type)
		
		if quote.RegularMarketPrice != nil {
			fmt.Printf("  Regular Market Price: %d (scale: %d)\n", 
				quote.RegularMarketPrice.Scaled, quote.RegularMarketPrice.Scale)
		}
		if quote.Bid != nil {
			fmt.Printf("  Bid: %d (scale: %d)\n", quote.Bid.Scaled, quote.Bid.Scale)
		}
		if quote.Ask != nil {
			fmt.Printf("  Ask: %d (scale: %d)\n", quote.Ask.Scaled, quote.Ask.Scale)
		}
		if quote.RegularMarketVolume != nil {
			fmt.Printf("  Volume: %d\n", *quote.RegularMarketVolume)
		}
		
		fmt.Printf("  Event Time: %s\n", quote.EventTime.Format(time.RFC3339))
		fmt.Printf("  Ingest Time: %s\n", quote.IngestTime.Format(time.RFC3339))
	}

	// Example 2: Historical bars data contents
	fmt.Println("\n2. HISTORICAL BARS DATA CONTENTS")
	fmt.Println("================================")
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC) // Just 5 days for brevity
	
	dailyBars, err := client.FetchDailyBars(ctx, "AAPL", start, end, true, runID)
	if err != nil {
		log.Printf("Error fetching daily bars: %v", err)
	} else {
		fmt.Printf("Daily Bars for %s (%d bars):\n", dailyBars.Security.Symbol, len(dailyBars.Bars))
		fmt.Printf("  Schema: %s\n", dailyBars.Meta.SchemaVersion)
		fmt.Printf("  Source: %s\n", dailyBars.Meta.Source)
		fmt.Printf("  Producer: %s\n", dailyBars.Meta.Producer)
		
		// Show first few bars
		for i, bar := range dailyBars.Bars {
			if i >= 3 { // Show only first 3 bars
				fmt.Printf("  ... and %d more bars\n", len(dailyBars.Bars)-3)
				break
			}
			fmt.Printf("  Bar %d:\n", i+1)
			fmt.Printf("    Start: %s\n", bar.Start.Format("2006-01-02"))
			fmt.Printf("    End: %s\n", bar.End.Format("2006-01-02"))
			fmt.Printf("    Open: %d (scale: %d)\n", bar.Open.Scaled, bar.Open.Scale)
			fmt.Printf("    High: %d (scale: %d)\n", bar.High.Scaled, bar.High.Scale)
			fmt.Printf("    Low: %d (scale: %d)\n", bar.Low.Scaled, bar.Low.Scale)
			fmt.Printf("    Close: %d (scale: %d)\n", bar.Close.Scaled, bar.Close.Scale)
			fmt.Printf("    Volume: %d\n", bar.Volume)
			fmt.Printf("    Adjusted: %t\n", bar.Adjusted)
			fmt.Printf("    Currency: %s\n", bar.CurrencyCode)
		}
	}

	// Example 3: News data contents
	fmt.Println("\n3. NEWS DATA CONTENTS")
	fmt.Println("=====================")
	news, err := client.ScrapeNews(ctx, "AAPL", runID)
	if err != nil {
		log.Printf("Error fetching news: %v", err)
	} else {
		fmt.Printf("News for AAPL (%d articles):\n", len(news))
		
		// Show first few news articles
		for i, article := range news {
			if i >= 3 { // Show only first 3 articles
				fmt.Printf("  ... and %d more articles\n", len(news)-3)
				break
			}
			fmt.Printf("  Article %d:\n", i+1)
			fmt.Printf("    Headline: %s\n", article.Headline)
			fmt.Printf("    Source: %s\n", article.Source)
			if article.PublishedAt != nil {
				fmt.Printf("    Published: %s\n", article.PublishedAt.AsTime().Format(time.RFC3339))
			} else {
				fmt.Printf("    Published: (not available)\n")
			}
			if len(article.Tickers) > 0 {
				fmt.Printf("    Related Tickers: %v\n", article.Tickers)
			}
			fmt.Printf("    URL: %s\n", article.Url)
			fmt.Printf("    Schema: %s\n", article.Meta.SchemaVersion)
		}
	}

	// Example 4: Financials data contents
	fmt.Println("\n4. FINANCIALS DATA CONTENTS")
	fmt.Println("===========================")
	financials, err := client.ScrapeFinancials(ctx, "AAPL", runID)
	if err != nil {
		log.Printf("Error fetching financials: %v", err)
	} else {
		fmt.Printf("Financials for %s (%d line items):\n", financials.Security.Symbol, len(financials.Lines))
		fmt.Printf("  Schema: %s\n", financials.Meta.SchemaVersion)
		fmt.Printf("  Source: %s\n", financials.Source)
		fmt.Printf("  Producer: %s\n", financials.Meta.Producer)
		fmt.Printf("  As Of: %s\n", financials.AsOf.AsTime().Format(time.RFC3339))
		
		// Show first few line items
		for i, line := range financials.Lines {
			if i >= 5 { // Show only first 5 line items
				fmt.Printf("  ... and %d more line items\n", len(financials.Lines)-5)
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

	// Example 5: Key Statistics data contents
	fmt.Println("\n5. KEY STATISTICS DATA CONTENTS")
	fmt.Println("===============================")
	keyStats, err := client.ScrapeKeyStatistics(ctx, "AAPL", runID)
	if err != nil {
		log.Printf("Error fetching key statistics: %v", err)
	} else {
		fmt.Printf("Key Statistics for %s (%d line items):\n", keyStats.Security.Symbol, len(keyStats.Lines))
		fmt.Printf("  Schema: %s\n", keyStats.Meta.SchemaVersion)
		fmt.Printf("  Source: %s\n", keyStats.Source)
		fmt.Printf("  Producer: %s\n", keyStats.Meta.Producer)
		fmt.Printf("  As Of: %s\n", keyStats.AsOf.AsTime().Format(time.RFC3339))
		
		// Show first few line items
		for i, line := range keyStats.Lines {
			if i >= 5 { // Show only first 5 line items
				fmt.Printf("  ... and %d more line items\n", len(keyStats.Lines)-5)
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

	// Example 6: Show JSON representation of one message
	fmt.Println("\n6. JSON REPRESENTATION EXAMPLE")
	fmt.Println("==============================")
	if quote != nil {
		fmt.Println("Quote data as JSON:")
		jsonData, err := json.MarshalIndent(quote, "", "  ")
		if err != nil {
			log.Printf("Error marshaling to JSON: %v", err)
		} else {
			fmt.Println(string(jsonData))
		}
	}

	fmt.Println("\n=== DATA CONTENTS COMPLETE ===")
	fmt.Println("All data is structured ampy-proto messages with:")
	fmt.Println("- Proper schema versions")
	fmt.Println("- Scaled decimal values for precision")
	fmt.Println("- UTC timestamps")
	fmt.Println("- Type-specific source identifiers")
	fmt.Println("- Complete metadata for tracing and provenance")
}
