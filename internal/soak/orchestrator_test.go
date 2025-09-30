package soak

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/AmpyFin/yfinance-go/internal/config"
)

func TestNewOrchestrator(t *testing.T) {
	// Create a temporary universe file
	tempDir := t.TempDir()
	universeFile := filepath.Join(tempDir, "test_universe.txt")
	
	universeContent := `# Test universe
AAPL
MSFT
GOOGL`
	
	if err := os.WriteFile(universeFile, []byte(universeContent), 0644); err != nil {
		t.Fatalf("Failed to create test universe file: %v", err)
	}
	
	// Create test configuration
	cfg := &config.Config{
		// Add minimal required config fields
	}
	
	soakConfig := &SoakConfig{
		UniverseFile:  universeFile,
		Endpoints:     "key-statistics,news",
		Fallback:      "scrape-only",
		Duration:      10 * time.Second,
		Concurrency:   2,
		QPS:           1.0,
		Preview:       true,
		Publish:       false,
		ProbeInterval: 1 * time.Hour,
		FailureRate:   0.0, // Disable failure injection for test
		MemoryCheck:   true,
	}
	
	orchestrator, err := NewOrchestrator(cfg, soakConfig)
	if err != nil {
		t.Fatalf("Failed to create orchestrator: %v", err)
	}
	defer orchestrator.Close()
	
	// Verify orchestrator was created correctly
	if orchestrator == nil {
		t.Fatal("Orchestrator is nil")
	}
	
	if len(orchestrator.tickers) != 3 {
		t.Errorf("Expected 3 tickers, got %d", len(orchestrator.tickers))
	}
	
	expectedTickers := []string{"AAPL", "MSFT", "GOOGL"}
	for i, expected := range expectedTickers {
		if i >= len(orchestrator.tickers) || orchestrator.tickers[i] != expected {
			t.Errorf("Expected ticker %s at index %d, got %s", expected, i, orchestrator.tickers[i])
		}
	}
	
	if len(orchestrator.endpoints) != 2 {
		t.Errorf("Expected 2 endpoints, got %d", len(orchestrator.endpoints))
	}
}

func TestLoadTickerUniverse(t *testing.T) {
	tempDir := t.TempDir()
	universeFile := filepath.Join(tempDir, "test_universe.txt")
	
	universeContent := `# Test universe with comments
AAPL
# This is a comment
MSFT

GOOGL
# Another comment
TSLA`
	
	if err := os.WriteFile(universeFile, []byte(universeContent), 0644); err != nil {
		t.Fatalf("Failed to create test universe file: %v", err)
	}
	
	tickers, err := loadTickerUniverse(universeFile)
	if err != nil {
		t.Fatalf("Failed to load ticker universe: %v", err)
	}
	
	expected := []string{"AAPL", "MSFT", "GOOGL", "TSLA"}
	if len(tickers) != len(expected) {
		t.Errorf("Expected %d tickers, got %d", len(expected), len(tickers))
	}
	
	for i, expectedTicker := range expected {
		if i >= len(tickers) || tickers[i] != expectedTicker {
			t.Errorf("Expected ticker %s at index %d, got %s", expectedTicker, i, tickers[i])
		}
	}
}

func TestParseEndpoints(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{
			input:    "key-statistics,financials,news",
			expected: []string{"key-statistics", "financials", "news"},
		},
		{
			input:    "key-statistics, financials , news ",
			expected: []string{"key-statistics", "financials", "news"},
		},
		{
			input:    "single",
			expected: []string{"single"},
		},
	}
	
	for _, test := range tests {
		result := parseEndpoints(test.input)
		if len(result) != len(test.expected) {
			t.Errorf("For input %q, expected %d endpoints, got %d", test.input, len(test.expected), len(result))
			continue
		}
		
		for i, expected := range test.expected {
			if result[i] != expected {
				t.Errorf("For input %q, expected endpoint %s at index %d, got %s", test.input, expected, i, result[i])
			}
		}
	}
}

func TestOrchestratorShortRun(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping orchestrator integration test in short mode")
	}
	
	// Create a temporary universe file with a single ticker
	tempDir := t.TempDir()
	universeFile := filepath.Join(tempDir, "test_universe.txt")
	
	universeContent := `AAPL`
	
	if err := os.WriteFile(universeFile, []byte(universeContent), 0644); err != nil {
		t.Fatalf("Failed to create test universe file: %v", err)
	}
	
	// Create test configuration
	cfg := &config.Config{
		// Add minimal required config fields
	}
	
	soakConfig := &SoakConfig{
		UniverseFile:  universeFile,
		Endpoints:     "key-statistics",
		Fallback:      "scrape-only",
		Duration:      5 * time.Second, // Very short test
		Concurrency:   1,
		QPS:           0.5, // Very low QPS
		Preview:       true,
		Publish:       false,
		ProbeInterval: 1 * time.Hour, // No probes during short test
		FailureRate:   0.0, // No failure injection
		MemoryCheck:   true,
	}
	
	orchestrator, err := NewOrchestrator(cfg, soakConfig)
	if err != nil {
		t.Fatalf("Failed to create orchestrator: %v", err)
	}
	defer orchestrator.Close()
	
	// Run the soak test
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	err = orchestrator.Run(ctx)
	if err != nil {
		t.Fatalf("Soak test failed: %v", err)
	}
	
	// Verify some basic stats were collected
	if orchestrator.stats.TotalRequests == 0 {
		t.Error("Expected some requests to be made, but TotalRequests is 0")
	}
}
