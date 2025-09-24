package fx

import (
	"testing"
	"time"

	"github.com/yeonlee/yfinance-go/internal/norm"
)

func TestFXCache(t *testing.T) {
	cache := NewFXCache(100 * time.Millisecond)

	base := "EUR"
	symbols := []string{"USD", "GBP"}
	at := time.Now().UTC()

	rates := map[string]norm.ScaledDecimal{
		"USD": {Scaled: 110000000, Scale: 8}, // 1.10000000
		"GBP": {Scaled: 85000000, Scale: 8},  // 0.85000000
	}
	asOf := time.Now().UTC()

	// Test cache miss
	_, _, hit := cache.Get(base, symbols, at)
	if hit {
		t.Error("Expected cache miss")
	}

	// Test cache set
	cache.Set(base, symbols, at, rates, asOf)

	// Test cache hit
	cachedRates, cachedAsOf, hit := cache.Get(base, symbols, at)
	if !hit {
		t.Error("Expected cache hit")
	}

	// Verify cached data
	if len(cachedRates) != len(rates) {
		t.Errorf("Expected %d rates, got %d", len(rates), len(cachedRates))
	}

	for symbol, expectedRate := range rates {
		if cachedRate, exists := cachedRates[symbol]; !exists {
			t.Errorf("Missing rate for %s", symbol)
		} else if cachedRate.Scaled != expectedRate.Scaled || cachedRate.Scale != expectedRate.Scale {
			t.Errorf("Rate for %s: expected %v, got %v", symbol, expectedRate, cachedRate)
		}
	}

	if !cachedAsOf.Equal(asOf) {
		t.Errorf("Expected asOf %v, got %v", asOf, cachedAsOf)
	}

	// Test cache expiration
	time.Sleep(150 * time.Millisecond)
	_, _, hit = cache.Get(base, symbols, at)
	if hit {
		t.Error("Expected cache miss after expiration")
	}
}

func TestFXCacheKeyGeneration(t *testing.T) {
	cache := NewFXCache(time.Minute)

	base := "EUR"
	at := time.Now().UTC()

	// Test that different symbol orders produce the same key
	symbols1 := []string{"USD", "GBP"}
	symbols2 := []string{"GBP", "USD"}

	rates := map[string]norm.ScaledDecimal{
		"USD": {Scaled: 110000000, Scale: 8},
		"GBP": {Scaled: 85000000, Scale: 8},
	}
	asOf := time.Now().UTC()

	// Set with first order
	cache.Set(base, symbols1, at, rates, asOf)

	// Get with second order - should hit cache
	_, _, hit := cache.Get(base, symbols2, at)
	if !hit {
		t.Error("Expected cache hit with different symbol order")
	}
}

func TestFXCacheTimeBucketing(t *testing.T) {
	cache := NewFXCache(time.Minute)

	base := "EUR"
	symbols := []string{"USD"}

	rates := map[string]norm.ScaledDecimal{
		"USD": {Scaled: 110000000, Scale: 8},
	}
	asOf := time.Now().UTC()

	// Set with one time
	at1 := time.Date(2024, 1, 1, 12, 30, 45, 0, time.UTC)
	cache.Set(base, symbols, at1, rates, asOf)

	// Get with slightly different time in same minute - should hit cache
	at2 := time.Date(2024, 1, 1, 12, 30, 50, 0, time.UTC)
	_, _, hit := cache.Get(base, symbols, at2)
	if !hit {
		t.Error("Expected cache hit with time in same minute")
	}

	// Get with time in different minute - should miss cache
	at3 := time.Date(2024, 1, 1, 12, 31, 0, 0, time.UTC)
	_, _, hit = cache.Get(base, symbols, at3)
	if hit {
		t.Error("Expected cache miss with time in different minute")
	}
}

func TestFXCacheClear(t *testing.T) {
	cache := NewFXCache(time.Minute)

	base := "EUR"
	symbols := []string{"USD"}
	at := time.Now().UTC()

	rates := map[string]norm.ScaledDecimal{
		"USD": {Scaled: 110000000, Scale: 8},
	}
	asOf := time.Now().UTC()

	// Set data
	cache.Set(base, symbols, at, rates, asOf)

	// Verify it's cached
	_, _, hit := cache.Get(base, symbols, at)
	if !hit {
		t.Error("Expected cache hit before clear")
	}

	// Clear cache
	cache.Clear()

	// Verify it's no longer cached
	_, _, hit = cache.Get(base, symbols, at)
	if hit {
		t.Error("Expected cache miss after clear")
	}

	// Verify size is 0
	if cache.Size() != 0 {
		t.Errorf("Expected size 0 after clear, got %d", cache.Size())
	}
}

func TestFXCacheSize(t *testing.T) {
	cache := NewFXCache(time.Minute)

	// Initially empty
	if cache.Size() != 0 {
		t.Errorf("Expected initial size 0, got %d", cache.Size())
	}

	// Add one entry
	rates := map[string]norm.ScaledDecimal{
		"USD": {Scaled: 110000000, Scale: 8},
	}
	cache.Set("EUR", []string{"USD"}, time.Now().UTC(), rates, time.Now().UTC())

	if cache.Size() != 1 {
		t.Errorf("Expected size 1 after adding entry, got %d", cache.Size())
	}

	// Add another entry
	cache.Set("GBP", []string{"USD"}, time.Now().UTC(), rates, time.Now().UTC())

	if cache.Size() != 2 {
		t.Errorf("Expected size 2 after adding second entry, got %d", cache.Size())
	}
}
