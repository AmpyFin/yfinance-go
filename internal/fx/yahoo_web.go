package fx

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/AmpyFin/yfinance-go/internal/httpx"
	"github.com/AmpyFin/yfinance-go/internal/norm"
)

// YahooWebProvider implements FX interface using Yahoo Finance web scraping
type YahooWebProvider struct {
	config     *YahooWebConfig
	httpClient *httpx.Client
	cache      *FXCache
	rateScale  int
}

// NewYahooWebProvider creates a new yahoo-web FX provider
func NewYahooWebProvider(config *YahooWebConfig, rateScale int) *YahooWebProvider {
	// Create HTTP client with conservative settings
	httpConfig := httpx.DefaultConfig()
	httpConfig.QPS = config.QPS
	httpConfig.Burst = config.Burst
	httpConfig.Timeout = config.Timeout
	httpConfig.MaxAttempts = config.BackoffAttempts
	httpConfig.BackoffBaseMs = int(config.BackoffBase.Milliseconds())
	httpConfig.MaxDelayMs = int(config.BackoffMaxDelay.Milliseconds())
	httpConfig.ResetTimeout = config.CircuitReset

	httpClient := httpx.NewClient(httpConfig)
	cache := NewFXCache(60 * time.Second) // 60s TTL

	return &YahooWebProvider{
		config:     config,
		httpClient: httpClient,
		cache:      cache,
		rateScale:  rateScale,
	}
}

// Rates fetches FX rates from Yahoo Finance web interface
func (p *YahooWebProvider) Rates(ctx context.Context, base string, symbols []string, at time.Time) (map[string]norm.ScaledDecimal, time.Time, error) {
	// Check cache first
	if rates, asOf, hit := p.cache.Get(base, symbols, at); hit {
		return rates, asOf, nil
	}

	// Fetch rates from Yahoo Finance
	rates, asOf, err := p.fetchRatesFromYahoo(ctx, base, symbols)
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("failed to fetch rates from yahoo-web: %w", err)
	}

	// Cache the results
	p.cache.Set(base, symbols, at, rates, asOf)

	return rates, asOf, nil
}

// fetchRatesFromYahoo fetches rates from Yahoo Finance web interface
func (p *YahooWebProvider) fetchRatesFromYahoo(ctx context.Context, base string, symbols []string) (map[string]norm.ScaledDecimal, time.Time, error) {
	rates := make(map[string]norm.ScaledDecimal)
	asOf := time.Now().UTC()

	// For each symbol, fetch the rate
	for _, symbol := range symbols {
		rate, err := p.fetchSingleRate(ctx, base, symbol)
		if err != nil {
			return nil, time.Time{}, fmt.Errorf("failed to fetch rate for %s/%s: %w", base, symbol, err)
		}
		rates[symbol] = rate
	}

	return rates, asOf, nil
}

// fetchSingleRate fetches a single FX rate from Yahoo Finance
func (p *YahooWebProvider) fetchSingleRate(ctx context.Context, base, target string) (norm.ScaledDecimal, error) {
	// Construct Yahoo Finance URL for FX pair
	// Example: https://query1.finance.yahoo.com/v8/finance/chart/EURUSD=X
	pair := fmt.Sprintf("%s%s=X", base, target)
	url := fmt.Sprintf("https://query1.finance.yahoo.com/v8/finance/chart/%s", pair)

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return norm.ScaledDecimal{}, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers to mimic browser request
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")
	req.Header.Set("Accept", "application/json")

	// Make request
	resp, err := p.httpClient.Do(ctx, req)
	if err != nil {
		return norm.ScaledDecimal{}, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return norm.ScaledDecimal{}, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	// Parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return norm.ScaledDecimal{}, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse JSON response
	var yahooResp YahooChartResponse
	if unmarshalErr := json.Unmarshal(body, &yahooResp); unmarshalErr != nil {
		return norm.ScaledDecimal{}, fmt.Errorf("failed to parse JSON response: %w", unmarshalErr)
	}

	// Extract rate from response
	rate, err := p.extractRateFromResponse(&yahooResp)
	if err != nil {
		return norm.ScaledDecimal{}, fmt.Errorf("failed to extract rate: %w", err)
	}

	// Convert to scaled decimal
	scaledRate, err := norm.ToScaledDecimal(rate, p.rateScale)
	if err != nil {
		return norm.ScaledDecimal{}, fmt.Errorf("failed to convert rate to scaled decimal: %w", err)
	}

	return scaledRate, nil
}

// extractRateFromResponse extracts the FX rate from Yahoo Finance response
func (p *YahooWebProvider) extractRateFromResponse(resp *YahooChartResponse) (float64, error) {
	if resp.Chart == nil || len(resp.Chart.Result) == 0 {
		return 0, fmt.Errorf("no chart data in response")
	}

	result := resp.Chart.Result[0]
	if result.Meta == nil {
		return 0, fmt.Errorf("no meta data in response")
	}

	// Try to get regular market price
	if result.Meta.RegularMarketPrice != nil {
		return *result.Meta.RegularMarketPrice, nil
	}

	// Try to get previous close
	if result.Meta.PreviousClose != nil {
		return *result.Meta.PreviousClose, nil
	}

	// Try to get chart previous close
	if result.Meta.ChartPreviousClose != nil {
		return *result.Meta.ChartPreviousClose, nil
	}

	return 0, fmt.Errorf("no valid price found in response")
}

// YahooChartResponse represents the response from Yahoo Finance chart API
type YahooChartResponse struct {
	Chart *YahooChart `json:"chart"`
}

type YahooChart struct {
	Result []YahooChartResult `json:"result"`
	Error  *YahooChartError   `json:"error"`
}

type YahooChartResult struct {
	Meta *YahooChartMeta `json:"meta"`
}

type YahooChartMeta struct {
	RegularMarketPrice   *float64 `json:"regularMarketPrice"`
	PreviousClose        *float64 `json:"previousClose"`
	ChartPreviousClose   *float64 `json:"chartPreviousClose"`
	Currency             string   `json:"currency"`
	ExchangeName         string   `json:"exchangeName"`
	InstrumentType       string   `json:"instrumentType"`
	FirstTradeDate       *int64   `json:"firstTradeDate"`
	Timezone             string   `json:"timezone"`
	ExchangeTimezone     string   `json:"exchangeTimezone"`
	GMTOffset            int      `json:"gmtoffset"`
	LastUpdateTime       *int64   `json:"lastUpdateTime"`
	RegularMarketTime    *int64   `json:"regularMarketTime"`
	HasPrePostMarketData bool     `json:"hasPrePostMarketData"`
}

type YahooChartError struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}
