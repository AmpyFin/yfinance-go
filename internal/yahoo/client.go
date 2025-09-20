package yahoo

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/yeonlee/yfinance-go/internal/httpx"
)

// Client provides methods to fetch and normalize Yahoo Finance data
type Client struct {
	httpClient *httpx.Client
	baseURL    string
}

// NewClient creates a new Yahoo Finance client
func NewClient(httpClient *httpx.Client, baseURL string) *Client {
	if baseURL == "" {
		baseURL = "https://query1.finance.yahoo.com"
	}
	
	return &Client{
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

// FetchDailyBars fetches daily bars for a symbol
func (c *Client) FetchDailyBars(ctx context.Context, symbol string, start, end time.Time, adjusted bool) (*BarsResponse, error) {
	// Build URL for daily bars
	u, err := c.buildBarsURL(symbol, start, end, adjusted)
	if err != nil {
		return nil, fmt.Errorf("failed to build bars URL: %w", err)
	}
	
	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	// Execute request
	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch bars: %w", err)
	}
	defer resp.Body.Close()
	
	// Decode response
	barsResp, err := DecodeBarsResponseFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to decode bars response: %w", err)
	}
	
	// Validate response has data
	meta := barsResp.GetMetadata()
	if meta == nil {
		return nil, fmt.Errorf("missing metadata")
	}
	
	return barsResp, nil
}

// FetchQuote fetches a quote for a symbol using the chart endpoint (since v7 quote endpoint is restricted)
func (c *Client) FetchQuote(ctx context.Context, symbol string) (*QuoteResponse, error) {
	// Use chart endpoint to get quote data from metadata
	// This is more reliable since v7 quote endpoint returns 401
	end := time.Now()
	start := end.AddDate(0, 0, -1) // Just get 1 day of data to minimize response size
	
	barsResp, err := c.FetchDailyBars(ctx, symbol, start, end, true)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch quote via chart endpoint: %w", err)
	}
	
	// Convert chart metadata to quote response
	quoteResp, err := c.convertChartToQuote(barsResp, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to convert chart to quote: %w", err)
	}
	
	return quoteResp, nil
}

// FetchFundamentalsQuarterly fetches quarterly fundamentals for a symbol
func (c *Client) FetchFundamentalsQuarterly(ctx context.Context, symbol string) (*FundamentalsResponse, error) {
	// Build URL for fundamentals
	u, err := c.buildFundamentalsURL(symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to build fundamentals URL: %w", err)
	}
	
	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	// Execute request
	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch fundamentals: %w", err)
	}
	defer resp.Body.Close()
	
	// Decode response
	fundResp, err := DecodeFundamentalsResponseFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to decode fundamentals response: %w", err)
	}
	
	// Validate response has data
	if len(fundResp.QuoteSummary.Result) == 0 {
		return nil, fmt.Errorf("no fundamentals results found")
	}
	
	return fundResp, nil
}

// buildBarsURL builds the URL for fetching daily bars
func (c *Client) buildBarsURL(symbol string, start, end time.Time, adjusted bool) (string, error) {
	u, err := url.Parse(c.baseURL + "/v8/finance/chart/" + symbol)
	if err != nil {
		return "", err
	}
	
	// Add query parameters
	params := url.Values{}
	params.Set("period1", strconv.FormatInt(start.Unix(), 10))
	params.Set("period2", strconv.FormatInt(end.Unix(), 10))
	params.Set("interval", "1d")
	params.Set("includePrePost", "false")
	params.Set("events", "div,split")
	
	u.RawQuery = params.Encode()
	return u.String(), nil
}

// buildQuoteURL builds the URL for fetching quotes
func (c *Client) buildQuoteURL(symbol string) (string, error) {
	u, err := url.Parse(c.baseURL + "/v7/finance/quote")
	if err != nil {
		return "", err
	}
	
	// Add query parameters
	params := url.Values{}
	params.Set("symbols", symbol)
	
	u.RawQuery = params.Encode()
	return u.String(), nil
}

// buildFundamentalsURL builds the URL for fetching fundamentals
func (c *Client) buildFundamentalsURL(symbol string) (string, error) {
	u, err := url.Parse(c.baseURL + "/v10/finance/quoteSummary/" + symbol)
	if err != nil {
		return "", err
	}
	
	// Add query parameters
	params := url.Values{}
	params.Set("modules", "incomeStatementHistoryQuarterly,balanceSheetHistoryQuarterly,cashflowStatementHistoryQuarterly")
	
	u.RawQuery = params.Encode()
	return u.String(), nil
}

// convertChartToQuote converts chart metadata to a quote response
func (c *Client) convertChartToQuote(barsResp *BarsResponse, symbol string) (*QuoteResponse, error) {
	if len(barsResp.Chart.Result) == 0 {
		return nil, fmt.Errorf("no chart results found")
	}
	
	result := barsResp.Chart.Result[0]
	meta := result.Meta
	
	// Create a quote result from chart metadata
	quoteResult := QuoteResult{
		Symbol:                        meta.Symbol,
		Currency:                      meta.Currency,
		Exchange:                      meta.ExchangeName,
		FullExchangeName:              meta.FullExchangeName,
		ShortName:                     meta.ShortName,
		LongName:                      meta.LongName,
		RegularMarketPrice:            meta.RegularMarketPrice,
		RegularMarketDayHigh:          meta.RegularMarketDayHigh,
		RegularMarketDayLow:           meta.RegularMarketDayLow,
		RegularMarketVolume:           meta.RegularMarketVolume,
		RegularMarketTime:             &meta.RegularMarketTime,
		ExchangeTimezoneName:          meta.ExchangeTimezoneName,
		GmtOffsetMilliseconds:         meta.GmtOffset,
		MarketState:                   "REGULAR", // Default to regular market
		QuoteType:                     "EQUITY",  // Default to equity
		Language:                      "en-US",   // Default language
		Region:                        "US",      // Default region
	}
	
	// Create quote response
	quoteResponse := &QuoteResponse{
		QuoteResponse: QuoteResponseData{
			Result: []QuoteResult{quoteResult},
			Error:  nil,
		},
	}
	
	return quoteResponse, nil
}

// FetchIntradayBars fetches intraday bars for a symbol
func (c *Client) FetchIntradayBars(ctx context.Context, symbol string, start, end time.Time, interval string) (*BarsResponse, error) {
	// Build URL for intraday bars
	u, err := c.buildIntradayBarsURL(symbol, start, end, interval)
	if err != nil {
		return nil, fmt.Errorf("failed to build intraday bars URL: %w", err)
	}
	
	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	// Execute request
	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch intraday bars: %w", err)
	}
	defer resp.Body.Close()
	
	// Decode response
	barsResp, err := DecodeBarsResponseFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to decode intraday bars response: %w", err)
	}
	
	// Validate response has data
	meta := barsResp.GetMetadata()
	if meta == nil {
		return nil, fmt.Errorf("missing metadata")
	}
	
	return barsResp, nil
}

// FetchWeeklyBars fetches weekly bars for a symbol
func (c *Client) FetchWeeklyBars(ctx context.Context, symbol string, start, end time.Time, adjusted bool) (*BarsResponse, error) {
	// Build URL for weekly bars
	u, err := c.buildWeeklyBarsURL(symbol, start, end, adjusted)
	if err != nil {
		return nil, fmt.Errorf("failed to build weekly bars URL: %w", err)
	}
	
	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	// Execute request
	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch weekly bars: %w", err)
	}
	defer resp.Body.Close()
	
	// Decode response
	barsResp, err := DecodeBarsResponseFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to decode weekly bars response: %w", err)
	}
	
	// Validate response has data
	meta := barsResp.GetMetadata()
	if meta == nil {
		return nil, fmt.Errorf("missing metadata")
	}
	
	return barsResp, nil
}

// FetchMonthlyBars fetches monthly bars for a symbol
func (c *Client) FetchMonthlyBars(ctx context.Context, symbol string, start, end time.Time, adjusted bool) (*BarsResponse, error) {
	// Build URL for monthly bars
	u, err := c.buildMonthlyBarsURL(symbol, start, end, adjusted)
	if err != nil {
		return nil, fmt.Errorf("failed to build monthly bars URL: %w", err)
	}
	
	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	// Execute request
	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch monthly bars: %w", err)
	}
	defer resp.Body.Close()
	
	// Decode response
	barsResp, err := DecodeBarsResponseFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to decode monthly bars response: %w", err)
	}
	
	// Validate response has data
	meta := barsResp.GetMetadata()
	if meta == nil {
		return nil, fmt.Errorf("missing metadata")
	}
	
	return barsResp, nil
}

// buildIntradayBarsURL builds the URL for fetching intraday bars
func (c *Client) buildIntradayBarsURL(symbol string, start, end time.Time, interval string) (string, error) {
	u, err := url.Parse(c.baseURL + "/v8/finance/chart/" + symbol)
	if err != nil {
		return "", err
	}
	
	// Add query parameters
	params := url.Values{}
	params.Set("period1", strconv.FormatInt(start.Unix(), 10))
	params.Set("period2", strconv.FormatInt(end.Unix(), 10))
	params.Set("interval", interval)
	params.Set("includePrePost", "false")
	params.Set("events", "div,split")
	
	u.RawQuery = params.Encode()
	return u.String(), nil
}

// buildWeeklyBarsURL builds the URL for fetching weekly bars
func (c *Client) buildWeeklyBarsURL(symbol string, start, end time.Time, adjusted bool) (string, error) {
	u, err := url.Parse(c.baseURL + "/v8/finance/chart/" + symbol)
	if err != nil {
		return "", err
	}
	
	// Add query parameters
	params := url.Values{}
	params.Set("period1", strconv.FormatInt(start.Unix(), 10))
	params.Set("period2", strconv.FormatInt(end.Unix(), 10))
	params.Set("interval", "1wk")
	params.Set("includePrePost", "false")
	params.Set("events", "div,split")
	
	u.RawQuery = params.Encode()
	return u.String(), nil
}

// buildMonthlyBarsURL builds the URL for fetching monthly bars
func (c *Client) buildMonthlyBarsURL(symbol string, start, end time.Time, adjusted bool) (string, error) {
	u, err := url.Parse(c.baseURL + "/v8/finance/chart/" + symbol)
	if err != nil {
		return "", err
	}
	
	// Add query parameters
	params := url.Values{}
	params.Set("period1", strconv.FormatInt(start.Unix(), 10))
	params.Set("period2", strconv.FormatInt(end.Unix(), 10))
	params.Set("interval", "1mo")
	params.Set("includePrePost", "false")
	params.Set("events", "div,split")
	
	u.RawQuery = params.Encode()
	return u.String(), nil
}
