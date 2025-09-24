package fx

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/AmpyFin/yfinance-go/internal/norm"
)

// CacheEntry represents a cached FX rate entry
type CacheEntry struct {
	Rates   map[string]norm.ScaledDecimal
	AsOf    time.Time
	Expires time.Time
}

// FXCache provides in-memory caching for FX rates
type FXCache struct {
	mu    sync.RWMutex
	store map[string]*CacheEntry
	ttl   time.Duration
}

// NewFXCache creates a new FX cache with the specified TTL
func NewFXCache(ttl time.Duration) *FXCache {
	return &FXCache{
		store: make(map[string]*CacheEntry),
		ttl:   ttl,
	}
}

// Get retrieves cached FX rates
func (c *FXCache) Get(base string, symbols []string, at time.Time) (map[string]norm.ScaledDecimal, time.Time, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := c.makeKey(base, symbols, at)
	entry, exists := c.store[key]
	if !exists {
		return nil, time.Time{}, false
	}

	// Check if expired
	if time.Now().After(entry.Expires) {
		return nil, time.Time{}, false
	}

	// Return a copy of the rates
	rates := make(map[string]norm.ScaledDecimal)
	for k, v := range entry.Rates {
		rates[k] = v
	}

	return rates, entry.AsOf, true
}

// Set stores FX rates in the cache
func (c *FXCache) Set(base string, symbols []string, at time.Time, rates map[string]norm.ScaledDecimal, asOf time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := c.makeKey(base, symbols, at)
	
	// Create a copy of the rates
	ratesCopy := make(map[string]norm.ScaledDecimal)
	for k, v := range rates {
		ratesCopy[k] = v
	}

	c.store[key] = &CacheEntry{
		Rates:   ratesCopy,
		AsOf:    asOf,
		Expires: time.Now().Add(c.ttl),
	}
}

// makeKey creates a cache key from the parameters
func (c *FXCache) makeKey(base string, symbols []string, at time.Time) string {
	// Sort symbols for consistent key generation
	sortedSymbols := make([]string, len(symbols))
	copy(sortedSymbols, symbols)
	sort.Strings(sortedSymbols)

	// Bucket the time to the minute to improve cache hit rate
	bucketedAt := at.Truncate(time.Minute)

	return fmt.Sprintf("%s:%v:%d", base, sortedSymbols, bucketedAt.Unix())
}

// Clear removes all entries from the cache
func (c *FXCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store = make(map[string]*CacheEntry)
}

// Size returns the number of entries in the cache
func (c *FXCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.store)
}
