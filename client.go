package yfinance

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/AmpyFin/yfinance-go/internal/httpx"
	"github.com/AmpyFin/yfinance-go/internal/norm"
	"github.com/AmpyFin/yfinance-go/internal/yahoo"
)

// Client provides a high-level interface for fetching Yahoo Finance data
type Client struct {
	yahooClient *yahoo.Client
}

// NewClient creates a new Yahoo Finance client with default configuration
func NewClient() *Client {
	config := httpx.DefaultConfig()
	httpClient := httpx.NewClient(config)
	yahooClient := yahoo.NewClient(httpClient, "")
	
	return &Client{
		yahooClient: yahooClient,
	}
}

// NewClientWithConfig creates a new Yahoo Finance client with custom configuration
func NewClientWithConfig(config *httpx.Config) *Client {
	httpClient := httpx.NewClient(config)
	yahooClient := yahoo.NewClient(httpClient, config.BaseURL)
	
	return &Client{
		yahooClient: yahooClient,
	}
}

// NewClientWithSessionRotation creates a new Yahoo Finance client with session rotation enabled
func NewClientWithSessionRotation() *Client {
	config := httpx.SessionRotationConfig()
	httpClient := httpx.NewClient(config)
	yahooClient := yahoo.NewClient(httpClient, config.BaseURL)
	
	return &Client{
		yahooClient: yahooClient,
	}
}

// FetchDailyBars fetches daily bars for a symbol and returns normalized data
func (c *Client) FetchDailyBars(ctx context.Context, symbol string, start, end time.Time, adjusted bool, runID string) (*norm.NormalizedBarBatch, error) {
	// Fetch raw data
	barsResp, err := c.yahooClient.FetchDailyBars(ctx, symbol, start, end, adjusted)
	if err != nil {
		return nil, err
	}
	
	// Extract bars and metadata
	bars, err := barsResp.GetBars()
	if err != nil {
		return nil, err
	}
	
	meta := barsResp.GetMetadata()
	if meta == nil {
		return nil, fmt.Errorf("missing metadata")
	}
	
	// Normalize bars
	return norm.NormalizeBars(bars, meta, runID)
}

// FetchQuote fetches a quote for a symbol and returns normalized data
func (c *Client) FetchQuote(ctx context.Context, symbol string, runID string) (*norm.NormalizedQuote, error) {
	// Fetch raw data
	quoteResp, err := c.yahooClient.FetchQuote(ctx, symbol)
	if err != nil {
		return nil, err
	}
	
	// Extract quotes
	quotes := quoteResp.GetQuotes()
	if len(quotes) == 0 {
		return nil, fmt.Errorf("no quotes found")
	}
	
	// Normalize first quote
	return norm.NormalizeQuote(quotes[0], runID)
}

// FetchFundamentalsQuarterly fetches quarterly fundamentals for a symbol and returns normalized data
// Note: This endpoint requires Yahoo Finance paid subscription
func (c *Client) FetchFundamentalsQuarterly(ctx context.Context, symbol string, runID string) (*norm.NormalizedFundamentalsSnapshot, error) {
	// Fetch raw data
	fundResp, err := c.yahooClient.FetchFundamentalsQuarterly(ctx, symbol)
	if err != nil {
		// Check if it's a 401 error (authentication required)
		if isAuthenticationError(err) {
			return nil, fmt.Errorf("fundamentals data requires Yahoo Finance paid subscription: %w", err)
		}
		return nil, err
	}
	
	// Extract fundamentals
	fundamentals, err := fundResp.GetFundamentals()
	if err != nil {
		return nil, err
	}
	
	// Normalize fundamentals
	return norm.NormalizeFundamentals(fundamentals, symbol, runID)
}

// FetchIntradayBars fetches intraday bars for a symbol (1m, 5m, 15m, 30m, 60m intervals)
func (c *Client) FetchIntradayBars(ctx context.Context, symbol string, start, end time.Time, interval string, runID string) (*norm.NormalizedBarBatch, error) {
	// Fetch raw data
	barsResp, err := c.yahooClient.FetchIntradayBars(ctx, symbol, start, end, interval)
	if err != nil {
		return nil, err
	}
	
	// Extract bars and metadata
	bars, err := barsResp.GetBars()
	if err != nil {
		return nil, err
	}
	
	meta := barsResp.GetMetadata()
	if meta == nil {
		return nil, fmt.Errorf("missing metadata")
	}
	
	// Normalize bars
	return norm.NormalizeBars(bars, meta, runID)
}

// FetchWeeklyBars fetches weekly bars for a symbol
func (c *Client) FetchWeeklyBars(ctx context.Context, symbol string, start, end time.Time, adjusted bool, runID string) (*norm.NormalizedBarBatch, error) {
	// Fetch raw data
	barsResp, err := c.yahooClient.FetchWeeklyBars(ctx, symbol, start, end, adjusted)
	if err != nil {
		return nil, err
	}
	
	// Extract bars and metadata
	bars, err := barsResp.GetBars()
	if err != nil {
		return nil, err
	}
	
	meta := barsResp.GetMetadata()
	if meta == nil {
		return nil, fmt.Errorf("missing metadata")
	}
	
	// Normalize bars
	return norm.NormalizeBars(bars, meta, runID)
}

// FetchMonthlyBars fetches monthly bars for a symbol
func (c *Client) FetchMonthlyBars(ctx context.Context, symbol string, start, end time.Time, adjusted bool, runID string) (*norm.NormalizedBarBatch, error) {
	// Fetch raw data
	barsResp, err := c.yahooClient.FetchMonthlyBars(ctx, symbol, start, end, adjusted)
	if err != nil {
		return nil, err
	}
	
	// Extract bars and metadata
	bars, err := barsResp.GetBars()
	if err != nil {
		return nil, err
	}
	
	meta := barsResp.GetMetadata()
	if meta == nil {
		return nil, fmt.Errorf("missing metadata")
	}
	
	// Normalize bars
	return norm.NormalizeBars(bars, meta, runID)
}

// FetchCompanyInfo fetches basic company information from chart metadata
func (c *Client) FetchCompanyInfo(ctx context.Context, symbol string, runID string) (*norm.NormalizedCompanyInfo, error) {
	// Use chart endpoint to get company info from metadata
	end := time.Now()
	start := end.AddDate(0, 0, -1)
	
	barsResp, err := c.yahooClient.FetchDailyBars(ctx, symbol, start, end, true)
	if err != nil {
		return nil, err
	}
	
	// Extract metadata
	meta := barsResp.GetMetadata()
	if meta == nil {
		return nil, fmt.Errorf("missing metadata")
	}
	
	// Normalize company info
	return norm.NormalizeCompanyInfo(meta, runID)
}

// FetchMarketData fetches comprehensive market data (price, volume, 52-week range, etc.)
func (c *Client) FetchMarketData(ctx context.Context, symbol string, runID string) (*norm.NormalizedMarketData, error) {
	// Use chart endpoint to get comprehensive market data
	end := time.Now()
	start := end.AddDate(0, 0, -1)
	
	barsResp, err := c.yahooClient.FetchDailyBars(ctx, symbol, start, end, true)
	if err != nil {
		return nil, err
	}
	
	// Extract metadata
	meta := barsResp.GetMetadata()
	if meta == nil {
		return nil, fmt.Errorf("missing metadata")
	}
	
	// Normalize market data
	return norm.NormalizeMarketData(meta, runID)
}

// isAuthenticationError checks if an error indicates authentication is required
func isAuthenticationError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "401") || strings.Contains(errStr, "Unauthorized") || strings.Contains(errStr, "authentication")
}
