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
	
	// Fetch daily bars for Apple
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC)
	
	bars, err := client.FetchDailyBars(ctx, "AAPL", start, end, true, "test-run")
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Printf("Fetched %d bars for AAPL\n", len(bars.Bars))
	for _, bar := range bars.Bars {
		price := float64(bar.Close.Scaled) / float64(bar.Close.Scale)
		fmt.Printf("Date: %s, Close: %.4f %s\n", 
			bar.EventTime.Format("2006-01-02"),
			price, bar.CurrencyCode)
	}
	
	// Test quote functionality
	quote, err := client.FetchQuote(ctx, "AAPL", "test-quote-run")
	if err != nil {
		log.Printf("Error fetching quote: %v", err)
	} else {
		fmt.Printf("\nQuote for AAPL:\n")
		fmt.Printf("Symbol: %s\n", quote.Security.Symbol)
		if quote.RegularMarketPrice != nil {
			price := float64(quote.RegularMarketPrice.Scaled) / float64(quote.RegularMarketPrice.Scale)
			fmt.Printf("Price: %.4f %s\n", price, quote.CurrencyCode)
		}
		if quote.RegularMarketVolume != nil {
			fmt.Printf("Volume: %d\n", *quote.RegularMarketVolume)
		}
		fmt.Printf("Event Time: %s\n", quote.EventTime.Format("2006-01-02 15:04:05"))
	}
}
