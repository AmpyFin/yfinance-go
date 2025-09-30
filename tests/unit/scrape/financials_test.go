package scrape_test

import (
	"testing"

	"github.com/AmpyFin/yfinance-go/internal/scrape"
)

func TestLoadFinancialsRegexConfig(t *testing.T) {
	err := scrape.LoadFinancialsRegexConfig()
	if err != nil {
		t.Fatalf("LoadFinancialsRegexConfig failed: %v", err)
	}

	// Note: financialsRegexConfig is internal and not accessible from external tests
	// The fact that LoadFinancialsRegexConfig() doesn't return an error indicates success
}

// Helper function to create int64 pointer
func int64Ptr(i int64) *int64 {
	return &i
}
