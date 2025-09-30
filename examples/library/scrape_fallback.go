package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/AmpyFin/yfinance-go"
	"github.com/AmpyFin/yfinance-go/internal/httpx"
)

// Example 1: Basic Scrape Fallback Usage
// This example demonstrates the simplest way to use the scrape fallback system
func basicScrapeExample() {
	fmt.Println("=== Example 1: Basic Scrape Fallback ===")
	
	// Create client with default configuration (automatic fallback enabled)
	client := yfinance.NewClient()
	ctx := context.Background()
	runID := fmt.Sprintf("example-basic-%d", time.Now().Unix())
	
	// Scrape key statistics (only available through scraping)
	keyStats, err := client.ScrapeKeyStatistics(ctx, "AAPL", runID)
	if err != nil {
		log.Printf("Error scraping key statistics: %v", err)
		return
	}
	
	fmt.Printf("Successfully scraped key statistics for AAPL\n")
	fmt.Printf("Schema version: %s\n", keyStats.Meta.SchemaVersion)
	fmt.Printf("Run ID: %s\n", keyStats.Meta.RunId)
	fmt.Printf("Source: %s\n", keyStats.Meta.Source)
	
	// Print some key data if available
	if len(keyStats.Lines) > 0 {
		fmt.Printf("Found %d data lines:\n", len(keyStats.Lines))
		for i, line := range keyStats.Lines {
			if i >= 3 { // Limit output
				fmt.Printf("... and %d more lines\n", len(keyStats.Lines)-3)
				break
			}
			fmt.Printf("  - %s: %v\n", line.Key, line.Value)
		}
	}
}

// Example 2: Comprehensive Data Collection
// This example shows how to collect multiple types of data with error handling
func comprehensiveDataExample() {
	fmt.Println("\n=== Example 2: Comprehensive Data Collection ===")
	
	client := yfinance.NewClientWithSessionRotation() // Use session rotation for better reliability
	ctx := context.Background()
	ticker := "MSFT"
	runID := fmt.Sprintf("example-comprehensive-%d", time.Now().Unix())
	
	// Collect different types of data
	dataTypes := []struct {
		name string
		fn   func() error
	}{
		{
			name: "Key Statistics",
			fn: func() error {
				data, err := client.ScrapeKeyStatistics(ctx, ticker, runID)
				if err != nil {
					return err
				}
				fmt.Printf("  ✓ Key Statistics: %d lines collected\n", len(data.Lines))
				return nil
			},
		},
		{
			name: "Financials",
			fn: func() error {
				data, err := client.ScrapeFinancials(ctx, ticker, runID)
				if err != nil {
					return err
				}
				fmt.Printf("  ✓ Financials: %d line items collected\n", len(data.Lines))
				return nil
			},
		},
		{
			name: "Analysis",
			fn: func() error {
				data, err := client.ScrapeAnalysis(ctx, ticker, runID)
				if err != nil {
					return err
				}
				fmt.Printf("  ✓ Analysis: %d lines collected\n", len(data.Lines))
				return nil
			},
		},
		{
			name: "News",
			fn: func() error {
				articles, err := client.ScrapeNews(ctx, ticker, runID)
				if err != nil {
					return err
				}
				fmt.Printf("  ✓ News: %d articles collected\n", len(articles))
				return nil
			},
		},
	}
	
	fmt.Printf("Collecting data for %s...\n", ticker)
	successCount := 0
	
	for _, dataType := range dataTypes {
		if err := dataType.fn(); err != nil {
			fmt.Printf("  ✗ %s: %v\n", dataType.name, err)
		} else {
			successCount++
		}
		
		// Rate limiting - wait between requests
		time.Sleep(500 * time.Millisecond)
	}
	
	fmt.Printf("Completed: %d/%d data types collected successfully\n", successCount, len(dataTypes))
}

// Example 3: Advanced Configuration and Error Handling
// This example demonstrates advanced configuration options and robust error handling
func advancedConfigExample() {
	fmt.Println("\n=== Example 3: Advanced Configuration ===")
	
	// Create client with custom configuration
	config := &httpx.Config{
		Timeout:     45 * time.Second,  // Longer timeout for scraping
		MaxAttempts: 3,                 // Retry failed requests
		QPS:         1.0,               // Conservative rate limiting
		UserAgent:   "MyApp/1.0 (contact@mycompany.com)",
	}
	
	client := yfinance.NewClientWithConfig(config)
	ctx := context.Background()
	runID := fmt.Sprintf("example-advanced-%d", time.Now().Unix())
	
	// Test multiple tickers with different characteristics
	tickers := []struct {
		symbol      string
		description string
	}{
		{"AAPL", "US Large Cap"},
		{"TSLA", "US Growth Stock"},
		{"0700.HK", "International (Hong Kong)"},
		{"SAP", "European Stock"},
	}
	
	fmt.Println("Testing scraping across different market types...")
	
	for _, ticker := range tickers {
		fmt.Printf("\nTesting %s (%s):\n", ticker.symbol, ticker.description)
		
		// Test key statistics with comprehensive error handling
		startTime := time.Now()
		keyStats, err := client.ScrapeKeyStatistics(ctx, ticker.symbol, runID)
		duration := time.Since(startTime)
		
		if err != nil {
			fmt.Printf("  ✗ Failed after %v: %v\n", duration, err)
			
			// Analyze error type for better handling
			switch {
			case isRateLimitError(err):
				fmt.Printf("    → Rate limit detected, should implement backoff\n")
			case isParseError(err):
				fmt.Printf("    → Parse error, may indicate schema drift\n")
			case isNetworkError(err):
				fmt.Printf("    → Network error, should retry\n")
			default:
				fmt.Printf("    → Unknown error type\n")
			}
			continue
		}
		
		fmt.Printf("  ✓ Success in %v\n", duration)
		fmt.Printf("    Schema: %s\n", keyStats.Meta.SchemaVersion)
		fmt.Printf("    Lines: %d\n", len(keyStats.Lines))
		fmt.Printf("    Source: %s\n", keyStats.Meta.Source)
		
		// Rate limiting between tickers
		time.Sleep(2 * time.Second)
	}
}

// Example 4: Batch Processing with Orchestration
// This example shows how to process multiple tickers efficiently
func batchProcessingExample() {
	fmt.Println("\n=== Example 4: Batch Processing ===")
	
	client := yfinance.NewClientWithSessionRotation()
	ctx := context.Background()
	runID := fmt.Sprintf("example-batch-%d", time.Now().Unix())
	
	// Define a universe of tickers to process
	universe := []string{"AAPL", "MSFT", "GOOGL", "AMZN", "TSLA"}
	
	// Results collection
	type Result struct {
		Ticker   string
		Success  bool
		Duration time.Duration
		Error    error
		Metrics  int
	}
	
	results := make([]Result, 0, len(universe))
	
	fmt.Printf("Processing %d tickers in batch...\n", len(universe))
	
	for i, ticker := range universe {
		fmt.Printf("[%d/%d] Processing %s...", i+1, len(universe), ticker)
		
		startTime := time.Now()
		keyStats, err := client.ScrapeKeyStatistics(ctx, ticker, runID)
		duration := time.Since(startTime)
		
		result := Result{
			Ticker:   ticker,
			Duration: duration,
		}
		
		if err != nil {
			result.Success = false
			result.Error = err
			fmt.Printf(" ✗ (%.2fs)\n", duration.Seconds())
		} else {
			result.Success = true
			result.Metrics = len(keyStats.Lines)
			fmt.Printf(" ✓ (%.2fs, %d metrics)\n", duration.Seconds(), result.Metrics)
		}
		
		results = append(results, result)
		
		// Rate limiting between requests
		if i < len(universe)-1 {
			time.Sleep(1 * time.Second)
		}
	}
	
	// Print summary
	fmt.Println("\n=== Batch Processing Summary ===")
	successCount := 0
	totalDuration := time.Duration(0)
	totalMetrics := 0
	
	for _, result := range results {
		totalDuration += result.Duration
		if result.Success {
			successCount++
			totalMetrics += result.Metrics
		}
	}
	
	fmt.Printf("Success Rate: %d/%d (%.1f%%)\n", successCount, len(results), float64(successCount)/float64(len(results))*100)
	fmt.Printf("Average Duration: %.2fs\n", totalDuration.Seconds()/float64(len(results)))
	fmt.Printf("Total Metrics Collected: %d\n", totalMetrics)
	
	// Print failures for debugging
	if successCount < len(results) {
		fmt.Println("\nFailures:")
		for _, result := range results {
			if !result.Success {
				fmt.Printf("  %s: %v\n", result.Ticker, result.Error)
			}
		}
	}
}

// Example 5: Real-time Data Pipeline Integration
// This example demonstrates integration with a data pipeline
func pipelineIntegrationExample() {
	fmt.Println("\n=== Example 5: Pipeline Integration ===")
	
	client := yfinance.NewClientWithSessionRotation()
	ctx := context.Background()
	
	// Simulate a data pipeline that processes financial data
	type DataPipeline struct {
		client *yfinance.Client
		runID  string
	}
	
	pipeline := &DataPipeline{
		client: client,
		runID:  fmt.Sprintf("pipeline-%d", time.Now().Unix()),
	}
	
	// Pipeline method to process a ticker
	processTicker := func(ticker string) error {
		fmt.Printf("Pipeline processing %s...\n", ticker)
		
		// Step 1: Collect key statistics
		keyStats, err := pipeline.client.ScrapeKeyStatistics(ctx, ticker, pipeline.runID)
		if err != nil {
			return fmt.Errorf("key statistics failed: %w", err)
		}
		
		// Step 2: Collect financial data
		financials, err := pipeline.client.ScrapeFinancials(ctx, ticker, pipeline.runID)
		if err != nil {
			return fmt.Errorf("financials failed: %w", err)
		}
		
		// Step 3: Collect news (optional, continue on failure)
		news, err := pipeline.client.ScrapeNews(ctx, ticker, pipeline.runID)
		if err != nil {
			fmt.Printf("  Warning: News collection failed: %v\n", err)
			news = nil // Continue without news
		}
		
		// Step 4: Process and validate data
		fmt.Printf("  ✓ Key Statistics: %d lines\n", len(keyStats.Lines))
		fmt.Printf("  ✓ Financials: %d line items\n", len(financials.Lines))
		if news != nil {
			fmt.Printf("  ✓ News: %d articles\n", len(news))
		}
		
		// Step 5: Data quality checks
		if len(keyStats.Lines) == 0 {
			return fmt.Errorf("no key statistics found")
		}
		
		if len(financials.Lines) == 0 {
			return fmt.Errorf("no financial data found")
		}
		
		// Step 6: Simulate data storage/publishing
		fmt.Printf("  ✓ Data validated and ready for storage\n")
		
		return nil
	}
	
	// Process a few tickers through the pipeline
	testTickers := []string{"AAPL", "MSFT", "GOOGL"}
	
	for _, ticker := range testTickers {
		if err := processTicker(ticker); err != nil {
			fmt.Printf("  ✗ Pipeline failed for %s: %v\n", ticker, err)
		} else {
			fmt.Printf("  ✓ Pipeline completed successfully for %s\n", ticker)
		}
		
		// Rate limiting between pipeline runs
		time.Sleep(2 * time.Second)
	}
}

// Helper functions for error classification
func isRateLimitError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return contains(errStr, "429") || contains(errStr, "rate limit") || contains(errStr, "too many requests")
}

func isParseError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return contains(errStr, "parse") || contains(errStr, "schema") || contains(errStr, "extract")
}

func isNetworkError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return contains(errStr, "timeout") || contains(errStr, "connection") || contains(errStr, "network")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		len(s) > len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		 indexOfSubstring(s, substr) >= 0))
}

func indexOfSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func main() {
	fmt.Println("yfinance-go Scrape Fallback Examples")
	fmt.Println("====================================")
	
	// Run all examples
	basicScrapeExample()
	comprehensiveDataExample()
	advancedConfigExample()
	batchProcessingExample()
	pipelineIntegrationExample()
	
	fmt.Println("\n=== All Examples Completed ===")
	fmt.Println("For more information, see:")
	fmt.Println("- Documentation: docs/scrape/")
	fmt.Println("- Configuration: docs/scrape/config.md")
	fmt.Println("- Troubleshooting: docs/scrape/troubleshooting.md")
}
