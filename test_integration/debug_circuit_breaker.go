package main

import (
	"context"
	"fmt"
	"time"

	"github.com/yeonlee/yfinance-go"
)

func main() {
	fmt.Println("=== CIRCUIT BREAKER DEBUG TEST ===")
	fmt.Println("Testing circuit breaker behavior with session rotation")
	fmt.Println()

	// Create client with session rotation enabled
	client := yfinance.NewClientWithSessionRotation()
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	runID := fmt.Sprintf("debug-test-%d", time.Now().Unix())

	// Test symbols in order: AAPL, MSFT, TSLA, 005930.KS, BABA, TSM, NVAX
	symbols := []string{"AAPL", "MSFT", "TSLA", "005930.KS", "BABA", "TSM", "NVAX"}
	
	fmt.Printf("Run ID: %s\n", runID)
	fmt.Printf("Test Time: %s\n", time.Now().UTC().Format(time.RFC3339))
	fmt.Printf("Testing %d symbols to debug circuit breaker behavior\n", len(symbols))
	fmt.Println()

	for i, symbol := range symbols {
		fmt.Printf("=== TESTING %s (Symbol %d/7) ===\n", symbol, i+1)
		
		// Small delay between symbols
		if i > 0 {
			fmt.Println("   Waiting 500ms between symbols...")
			time.Sleep(500 * time.Millisecond)
		}
		
		// Test just the quote to see when circuit breaker opens
		fmt.Print("Current Quote: ")
		quote, err := client.FetchQuote(ctx, symbol, runID)
		if err != nil {
			fmt.Printf("❌ FAILED: %v\n", err)
			
			// Check if it's a circuit breaker error
			if err.Error() == "failed to fetch quote via chart endpoint: failed to fetch bars: circuit breaker is open" {
				fmt.Printf("🔴 CIRCUIT BREAKER OPENED at symbol %d (%s)\n", i+1, symbol)
				fmt.Printf("   This means we accumulated too many failures from previous symbols\n")
				fmt.Printf("   The circuit breaker is GLOBAL, not per-symbol\n")
				break
			}
		} else {
			fmt.Printf("✅ Price=%v, Venue=%s\n", quote.RegularMarketPrice, quote.Venue)
		}

		// Test fundamentals (which should fail with 401)
		fmt.Print("Fundamentals (should fail): ")
		_, err = client.FetchFundamentalsQuarterly(ctx, symbol, runID)
		if err != nil {
			if err.Error() == "fundamentals data requires Yahoo Finance paid subscription: failed to fetch fundamentals: HTTP 401" {
				fmt.Printf("✅ PROPER ERROR (401 - expected)\n")
			} else {
				fmt.Printf("❌ UNEXPECTED ERROR: %v\n", err)
			}
		} else {
			fmt.Printf("✅ SUCCESS (unexpected)\n")
		}

		fmt.Println()
	}

	fmt.Println("=== CIRCUIT BREAKER ANALYSIS ===")
	fmt.Println("The circuit breaker is GLOBAL and accumulates failures from ALL requests.")
	fmt.Println("Even though session rotation helps, some requests still fail (like fundamentals 401s).")
	fmt.Println("When we reach the failure threshold (5), the circuit opens and blocks ALL requests.")
	fmt.Println()
	fmt.Println("SOLUTIONS:")
	fmt.Println("1. Increase FailureThreshold (e.g., from 5 to 15)")
	fmt.Println("2. Reduce ResetTimeout (e.g., from 30s to 10s)")
	fmt.Println("3. Make circuit breaker per-symbol (more complex)")
	fmt.Println("4. Don't count 401 errors as circuit breaker failures")
}
