package yfinance

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/AmpyFin/yfinance-go/internal/emit"
	"github.com/AmpyFin/yfinance-go/internal/httpx"
	"github.com/AmpyFin/yfinance-go/internal/norm"
	"github.com/AmpyFin/yfinance-go/internal/scrape"
	"github.com/AmpyFin/yfinance-go/internal/yahoo"
	fundamentalsv1 "github.com/AmpyFin/ampy-proto/v2/gen/go/ampy/fundamentals/v1"
	newsv1 "github.com/AmpyFin/ampy-proto/v2/gen/go/ampy/news/v1"
)

// Client provides a high-level interface for fetching Yahoo Finance data
type Client struct {
	yahooClient *yahoo.Client
	scrapeClient scrape.Client
}

// NewClient creates a new Yahoo Finance client with default configuration
func NewClient() *Client {
	config := httpx.DefaultConfig()
	httpClient := httpx.NewClient(config)
	yahooClient := yahoo.NewClient(httpClient, "")
	scrapeClient := scrape.NewClient(scrape.DefaultConfig(), httpClient)

	return &Client{
		yahooClient: yahooClient,
		scrapeClient: scrapeClient,
	}
}

// NewClientWithConfig creates a new Yahoo Finance client with custom configuration
func NewClientWithConfig(config *httpx.Config) *Client {
	httpClient := httpx.NewClient(config)
	yahooClient := yahoo.NewClient(httpClient, config.BaseURL)
	scrapeClient := scrape.NewClient(scrape.DefaultConfig(), httpClient)

	return &Client{
		yahooClient: yahooClient,
		scrapeClient: scrapeClient,
	}
}

// NewClientWithSessionRotation creates a new Yahoo Finance client with session rotation enabled
func NewClientWithSessionRotation() *Client {
	config := httpx.SessionRotationConfig()
	httpClient := httpx.NewClient(config)
	yahooClient := yahoo.NewClient(httpClient, config.BaseURL)
	scrapeClient := scrape.NewClient(scrape.DefaultConfig(), httpClient)

	return &Client{
		yahooClient: yahooClient,
		scrapeClient: scrapeClient,
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

// Scraping Functions - Return AMPY-PROTO Data

// ScrapeFinancials fetches financials data and returns ampy-proto FundamentalsSnapshot
func (c *Client) ScrapeFinancials(ctx context.Context, symbol string, runID string) (*fundamentalsv1.FundamentalsSnapshot, error) {
	url := fmt.Sprintf("https://finance.yahoo.com/quote/%s/financials", symbol)
	body, _, err := c.scrapeClient.Fetch(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch financials: %w", err)
	}

	dto, err := scrape.ParseComprehensiveFinancials(body, symbol, "XNAS")
	if err != nil {
		return nil, fmt.Errorf("failed to parse financials: %w", err)
	}

	snapshots, err := emit.MapComprehensiveFinancialsDTO(dto, runID, "yfinance-go")
	if err != nil {
		return nil, fmt.Errorf("failed to map financials: %w", err)
	}

	if len(snapshots) == 0 {
		return nil, fmt.Errorf("no financials data found")
	}

	return snapshots[0], nil
}

// ScrapeBalanceSheet fetches balance sheet data and returns ampy-proto FundamentalsSnapshot
func (c *Client) ScrapeBalanceSheet(ctx context.Context, symbol string, runID string) (*fundamentalsv1.FundamentalsSnapshot, error) {
	url := fmt.Sprintf("https://finance.yahoo.com/quote/%s/balance-sheet", symbol)
	body, _, err := c.scrapeClient.Fetch(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch balance sheet: %w", err)
	}

	dto, err := scrape.ParseComprehensiveFinancials(body, symbol, "XNAS")
	if err != nil {
		return nil, fmt.Errorf("failed to parse balance sheet: %w", err)
	}

	return emit.MapBalanceSheetDTO(dto, runID, "yfinance-go")
}

// ScrapeCashFlow fetches cash flow data and returns ampy-proto FundamentalsSnapshot
func (c *Client) ScrapeCashFlow(ctx context.Context, symbol string, runID string) (*fundamentalsv1.FundamentalsSnapshot, error) {
	url := fmt.Sprintf("https://finance.yahoo.com/quote/%s/cash-flow", symbol)
	body, _, err := c.scrapeClient.Fetch(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch cash flow: %w", err)
	}

	dto, err := scrape.ParseComprehensiveFinancials(body, symbol, "XNAS")
	if err != nil {
		return nil, fmt.Errorf("failed to parse cash flow: %w", err)
	}

	return emit.MapCashFlowDTO(dto, runID, "yfinance-go")
}

// ScrapeKeyStatistics fetches key statistics data and returns ampy-proto FundamentalsSnapshot
func (c *Client) ScrapeKeyStatistics(ctx context.Context, symbol string, runID string) (*fundamentalsv1.FundamentalsSnapshot, error) {
	url := fmt.Sprintf("https://finance.yahoo.com/quote/%s/key-statistics", symbol)
	body, _, err := c.scrapeClient.Fetch(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch key statistics: %w", err)
	}

	dto, err := scrape.ParseComprehensiveKeyStatistics(body, symbol, "XNAS")
	if err != nil {
		return nil, fmt.Errorf("failed to parse key statistics: %w", err)
	}

	return emit.MapKeyStatisticsDTO(dto, runID, "yfinance-go")
}

// ScrapeAnalysis fetches analysis data and returns ampy-proto FundamentalsSnapshot
func (c *Client) ScrapeAnalysis(ctx context.Context, symbol string, runID string) (*fundamentalsv1.FundamentalsSnapshot, error) {
	url := fmt.Sprintf("https://finance.yahoo.com/quote/%s/analysis", symbol)
	body, _, err := c.scrapeClient.Fetch(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch analysis: %w", err)
	}

	dto, err := scrape.ParseAnalysis(body, symbol, "XNAS")
	if err != nil {
		return nil, fmt.Errorf("failed to parse analysis: %w", err)
	}

	return emit.MapAnalysisDTO(dto, runID, "yfinance-go")
}

// ScrapeAnalystInsights fetches analyst insights data and returns ampy-proto FundamentalsSnapshot
func (c *Client) ScrapeAnalystInsights(ctx context.Context, symbol string, runID string) (*fundamentalsv1.FundamentalsSnapshot, error) {
	url := fmt.Sprintf("https://finance.yahoo.com/quote/%s/analyst-insights", symbol)
	body, _, err := c.scrapeClient.Fetch(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch analyst insights: %w", err)
	}

	dto, err := scrape.ParseAnalystInsights(body, symbol, "XNAS")
	if err != nil {
		return nil, fmt.Errorf("failed to parse analyst insights: %w", err)
	}

	return emit.MapAnalystInsightsDTO(dto, runID, "yfinance-go")
}

// ScrapeNews fetches news data and returns ampy-proto NewsItem slice
func (c *Client) ScrapeNews(ctx context.Context, symbol string, runID string) ([]*newsv1.NewsItem, error) {
	url := fmt.Sprintf("https://finance.yahoo.com/quote/%s/news", symbol)
	body, _, err := c.scrapeClient.Fetch(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch news: %w", err)
	}

	articles, _, err := scrape.ParseNews(body, "https://finance.yahoo.com", time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to parse news: %w", err)
	}

	protoArticles, err := emit.MapNewsItems(articles, symbol, runID, "yfinance-go")
	if err != nil {
		return nil, fmt.Errorf("failed to map news: %w", err)
	}

	return protoArticles, nil
}

// ScrapeAllFundamentals fetches all fundamentals data and returns multiple ampy-proto FundamentalsSnapshot messages
func (c *Client) ScrapeAllFundamentals(ctx context.Context, symbol string, runID string) ([]*fundamentalsv1.FundamentalsSnapshot, error) {
	var snapshots []*fundamentalsv1.FundamentalsSnapshot

	// Fetch all fundamentals data types
	financials, err := c.ScrapeFinancials(ctx, symbol, runID)
	if err != nil {
		return nil, fmt.Errorf("failed to scrape financials: %w", err)
	}
	snapshots = append(snapshots, financials)

	balanceSheet, err := c.ScrapeBalanceSheet(ctx, symbol, runID)
	if err != nil {
		return nil, fmt.Errorf("failed to scrape balance sheet: %w", err)
	}
	snapshots = append(snapshots, balanceSheet)

	cashFlow, err := c.ScrapeCashFlow(ctx, symbol, runID)
	if err != nil {
		return nil, fmt.Errorf("failed to scrape cash flow: %w", err)
	}
	snapshots = append(snapshots, cashFlow)

	keyStats, err := c.ScrapeKeyStatistics(ctx, symbol, runID)
	if err != nil {
		return nil, fmt.Errorf("failed to scrape key statistics: %w", err)
	}
	snapshots = append(snapshots, keyStats)

	analysis, err := c.ScrapeAnalysis(ctx, symbol, runID)
	if err != nil {
		return nil, fmt.Errorf("failed to scrape analysis: %w", err)
	}
	snapshots = append(snapshots, analysis)

	analystInsights, err := c.ScrapeAnalystInsights(ctx, symbol, runID)
	if err != nil {
		return nil, fmt.Errorf("failed to scrape analyst insights: %w", err)
	}
	snapshots = append(snapshots, analystInsights)

	return snapshots, nil
}

// isAuthenticationError checks if an error indicates authentication is required
func isAuthenticationError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "401") || strings.Contains(errStr, "Unauthorized") || strings.Contains(errStr, "authentication")
}
