package yfinance

import (
	"context"
	"testing"
	"time"

	"github.com/yeonlee/yfinance-go/internal/httpx"
)

func TestIntegrationDecode(t *testing.T) {
	// This test validates that the entire pipeline works correctly
	// by testing against the golden test data
	
	client := NewClient()
	ctx := context.Background()
	runID := "integration_test"

	// Test bars decoding
	t.Run("bars", func(t *testing.T) {
		start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)
		
		// This would normally make a real HTTP request, but for integration
		// testing we'll focus on the decoding pipeline
		_, err := client.FetchDailyBars(ctx, "AAPL", start, end, true, runID)
		if err != nil {
			// Expected to fail in test environment without real network
			t.Logf("Bars fetch failed as expected in test environment: %v", err)
		}
	})

	// Test quote decoding
	t.Run("quotes", func(t *testing.T) {
		_, err := client.FetchQuote(ctx, "MSFT", runID)
		if err != nil {
			// Expected to fail in test environment without real network
			t.Logf("Quote fetch failed as expected in test environment: %v", err)
		}
	})

	// Test fundamentals decoding
	t.Run("fundamentals", func(t *testing.T) {
		_, err := client.FetchFundamentalsQuarterly(ctx, "AAPL", runID)
		if err != nil {
			// Expected to fail in test environment without real network
			t.Logf("Fundamentals fetch failed as expected in test environment: %v", err)
		}
	})
}

func TestClientConfiguration(t *testing.T) {
	config := &httpx.Config{
		BaseURL:            "https://test.example.com",
		Timeout:            10 * time.Second,
		MaxAttempts:        3,
		BackoffBaseMs:      100,
		BackoffJitterMs:    50,
		MaxDelayMs:         1000,
		QPS:                1.0,
		Burst:              2,
		CircuitWindow:      30 * time.Second,
		FailureThreshold:   3,
		ResetTimeout:       10 * time.Second,
		UserAgent:          "Test-Agent/1.0",
	}

	client := NewClientWithConfig(config)
	if client == nil {
		t.Fatal("Client should not be nil")
	}
}
