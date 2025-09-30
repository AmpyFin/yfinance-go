package yahoo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
)

// BarsResponse represents the Yahoo Finance bars API response
type BarsResponse struct {
	Chart Chart `json:"chart"`
}

// Chart contains the chart data
type Chart struct {
	Result []ChartResult `json:"result"`
	Error  *string       `json:"error"`
}

// ChartResult contains the actual chart data for a symbol
type ChartResult struct {
	Meta       ChartMeta       `json:"meta"`
	Timestamp  []int64         `json:"timestamp"`
	Indicators ChartIndicators `json:"indicators"`
}

// ChartMeta contains metadata about the chart
type ChartMeta struct {
	Currency             string                `json:"currency"`
	Symbol               string                `json:"symbol"`
	ExchangeName         string                `json:"exchangeName"`
	FullExchangeName     string                `json:"fullExchangeName"`
	InstrumentType       string                `json:"instrumentType"`
	FirstTradeDate       int64                 `json:"firstTradeDate"`
	RegularMarketTime    int64                 `json:"regularMarketTime"`
	HasPrePostMarketData bool                  `json:"hasPrePostMarketData"`
	GmtOffset            int64                 `json:"gmtoffset"`
	Timezone             string                `json:"timezone"`
	ExchangeTimezoneName string                `json:"exchangeTimezoneName"`
	RegularMarketPrice   *float64              `json:"regularMarketPrice"`
	FiftyTwoWeekHigh     *float64              `json:"fiftyTwoWeekHigh"`
	FiftyTwoWeekLow      *float64              `json:"fiftyTwoWeekLow"`
	RegularMarketDayHigh *float64              `json:"regularMarketDayHigh"`
	RegularMarketDayLow  *float64              `json:"regularMarketDayLow"`
	RegularMarketVolume  *int64                `json:"regularMarketVolume"`
	LongName             string                `json:"longName"`
	ShortName            string                `json:"shortName"`
	ChartPreviousClose   *float64              `json:"chartPreviousClose"`
	PreviousClose        *float64              `json:"previousClose"`
	Scale                int                   `json:"scale"`
	PriceHint            int                   `json:"priceHint"`
	CurrentTradingPeriod *CurrentTradingPeriod `json:"currentTradingPeriod"`
	DataGranularity      string                `json:"dataGranularity"`
	Range                string                `json:"range"`
	ValidRanges          []string              `json:"validRanges"`
}

// CurrentTradingPeriod contains trading period information
type CurrentTradingPeriod struct {
	Pre     *TradingPeriod `json:"pre"`
	Regular *TradingPeriod `json:"regular"`
	Post    *TradingPeriod `json:"post"`
}

// TradingPeriod represents a trading period
type TradingPeriod struct {
	Timezone  string `json:"timezone"`
	Start     int64  `json:"start"`
	End       int64  `json:"end"`
	GmtOffset int64  `json:"gmtoffset"`
}

// ChartIndicators contains the price and volume indicators
type ChartIndicators struct {
	Quote    []QuoteIndicator    `json:"quote"`
	AdjClose []AdjCloseIndicator `json:"adjclose"`
}

// QuoteIndicator contains OHLCV data
type QuoteIndicator struct {
	Open   []*float64 `json:"open"`
	High   []*float64 `json:"high"`
	Low    []*float64 `json:"low"`
	Close  []*float64 `json:"close"`
	Volume []*int64   `json:"volume"`
}

// AdjCloseIndicator contains adjusted close prices
type AdjCloseIndicator struct {
	AdjClose []*float64 `json:"adjclose"`
}

// DecodeBarsResponse decodes a Yahoo Finance bars response with strict validation
func DecodeBarsResponse(data []byte) (*BarsResponse, error) {
	var response BarsResponse

	// Use JSON decoding that allows unknown fields
	// Yahoo Finance frequently adds new fields, so we need to be flexible
	decoder := json.NewDecoder(bytes.NewReader(data))
	// Allow unknown fields to handle Yahoo Finance API evolution
	// decoder.DisallowUnknownFields()

	if err := decoder.Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode bars response: %w", err)
	}

	// Validate response structure
	if err := response.Validate(); err != nil {
		return nil, fmt.Errorf("invalid bars response: %w", err)
	}

	return &response, nil
}

// Validate validates the bars response structure
func (r *BarsResponse) Validate() error {
	if r.Chart.Error != nil {
		return fmt.Errorf("yahoo finance error: %s", *r.Chart.Error)
	}

	if len(r.Chart.Result) == 0 {
		return fmt.Errorf("no chart results found")
	}

	for i, result := range r.Chart.Result {
		if err := result.Validate(); err != nil {
			return fmt.Errorf("result[%d]: %w", i, err)
		}
	}

	return nil
}

// Validate validates a chart result
func (r *ChartResult) Validate() error {
	if r.Meta.Symbol == "" {
		return fmt.Errorf("missing symbol")
	}

	if r.Meta.Currency == "" {
		return fmt.Errorf("missing currency")
	}

	if len(r.Timestamp) == 0 {
		return fmt.Errorf("no timestamps found")
	}

	if len(r.Indicators.Quote) == 0 {
		return fmt.Errorf("no quote data found")
	}

	quote := r.Indicators.Quote[0]

	// Validate data consistency
	expectedLen := len(r.Timestamp)
	if len(quote.Open) != expectedLen ||
		len(quote.High) != expectedLen ||
		len(quote.Low) != expectedLen ||
		len(quote.Close) != expectedLen ||
		len(quote.Volume) != expectedLen {
		return fmt.Errorf("data length mismatch: timestamps=%d, quote data lengths vary", expectedLen)
	}

	// Validate OHLCV data for each bar - skip bars with missing data
	validBars := 0
	for i := 0; i < expectedLen; i++ {
		// Check if this bar has complete data
		if quote.Open[i] != nil && quote.High[i] != nil && quote.Low[i] != nil && quote.Close[i] != nil && quote.Volume[i] != nil {
			if err := validateBarData(quote.Open[i], quote.High[i], quote.Low[i], quote.Close[i], quote.Volume[i]); err != nil {
				return fmt.Errorf("bar[%d]: %w", i, err)
			}
			validBars++
		}
		// Skip bars with missing data - this is common for international markets
	}

	// Ensure we have at least some valid data
	if validBars == 0 {
		return fmt.Errorf("no valid bars found - all bars have missing OHLCV data")
	}

	// Validate adjusted close if present
	if len(r.Indicators.AdjClose) > 0 {
		adjClose := r.Indicators.AdjClose[0]
		if len(adjClose.AdjClose) != expectedLen {
			return fmt.Errorf("adjusted close length mismatch: expected=%d, got=%d", expectedLen, len(adjClose.AdjClose))
		}

		for i := 0; i < expectedLen; i++ {
			if adjClose.AdjClose[i] != nil {
				if err := validatePrice(*adjClose.AdjClose[i]); err != nil {
					return fmt.Errorf("adjusted close[%d]: %w", i, err)
				}
			}
		}
	}

	return nil
}

// validateBarData validates OHLCV data for a single bar
func validateBarData(open, high, low, closePrice *float64, volume *int64) error {
	// Check for nil values - be more specific about which field is missing
	if open == nil {
		return fmt.Errorf("missing open price")
	}
	if high == nil {
		return fmt.Errorf("missing high price")
	}
	if low == nil {
		return fmt.Errorf("missing low price")
	}
	if closePrice == nil {
		return fmt.Errorf("missing close price")
	}
	if volume == nil {
		return fmt.Errorf("missing volume")
	}

	// Validate prices
	if err := validatePrice(*open); err != nil {
		return fmt.Errorf("invalid open price: %w", err)
	}
	if err := validatePrice(*high); err != nil {
		return fmt.Errorf("invalid high price: %w", err)
	}
	if err := validatePrice(*low); err != nil {
		return fmt.Errorf("invalid low price: %w", err)
	}
	if err := validatePrice(*closePrice); err != nil {
		return fmt.Errorf("invalid close price: %w", err)
	}

	// Validate volume
	if *volume < 0 {
		return fmt.Errorf("negative volume: %d", *volume)
	}

	// Validate OHLC relationships
	if *high < *low {
		return fmt.Errorf("high < low: high=%.4f, low=%.4f", *high, *low)
	}
	if *high < *open {
		return fmt.Errorf("high < open: high=%.4f, open=%.4f", *high, *open)
	}
	if *high < *closePrice {
		return fmt.Errorf("high < close: high=%.4f, close=%.4f", *high, *closePrice)
	}
	if *low > *open {
		return fmt.Errorf("low > open: low=%.4f, open=%.4f", *low, *open)
	}
	if *low > *closePrice {
		return fmt.Errorf("low > close: low=%.4f, close=%.4f", *low, *closePrice)
	}

	return nil
}

// validatePrice validates a price value
func validatePrice(price float64) error {
	if math.IsNaN(price) {
		return fmt.Errorf("NaN price")
	}
	if math.IsInf(price, 0) {
		return fmt.Errorf("infinite price")
	}
	if price < 0 {
		return fmt.Errorf("negative price: %.4f", price)
	}
	return nil
}

// GetBars extracts bar data from the response
func (r *BarsResponse) GetBars() ([]Bar, error) {
	if len(r.Chart.Result) == 0 {
		return nil, fmt.Errorf("no chart results")
	}

	result := r.Chart.Result[0]
	quote := result.Indicators.Quote[0]

	bars := make([]Bar, 0, len(result.Timestamp))

	for i, timestamp := range result.Timestamp {
		// Skip bars with missing OHLCV data
		if quote.Open[i] == nil || quote.High[i] == nil || quote.Low[i] == nil || quote.Close[i] == nil || quote.Volume[i] == nil {
			continue
		}

		bar := Bar{
			Timestamp: timestamp,
			Open:      *quote.Open[i],
			High:      *quote.High[i],
			Low:       *quote.Low[i],
			Close:     *quote.Close[i],
			Volume:    *quote.Volume[i],
		}

		// Add adjusted close if available
		if len(result.Indicators.AdjClose) > 0 &&
			result.Indicators.AdjClose[0].AdjClose[i] != nil {
			bar.AdjClose = result.Indicators.AdjClose[0].AdjClose[i]
		}

		bars = append(bars, bar)
	}

	return bars, nil
}

// Bar represents a single bar of OHLCV data
type Bar struct {
	Timestamp int64    `json:"timestamp"`
	Open      float64  `json:"open"`
	High      float64  `json:"high"`
	Low       float64  `json:"low"`
	Close     float64  `json:"close"`
	Volume    int64    `json:"volume"`
	AdjClose  *float64 `json:"adjclose,omitempty"`
}

// GetMetadata returns the chart metadata
func (r *BarsResponse) GetMetadata() *ChartMeta {
	if len(r.Chart.Result) == 0 {
		return nil
	}
	return &r.Chart.Result[0].Meta
}

// IsAdjusted returns true if adjusted close data is available
func (r *BarsResponse) IsAdjusted() bool {
	if len(r.Chart.Result) == 0 {
		return false
	}
	return len(r.Chart.Result[0].Indicators.AdjClose) > 0
}

// DecodeBarsResponseFromReader decodes a Yahoo Finance bars response from an io.Reader
func DecodeBarsResponseFromReader(reader io.Reader) (*BarsResponse, error) {
	var response BarsResponse

	// Use JSON decoding that allows unknown fields
	// Yahoo Finance frequently adds new fields, so we need to be flexible
	decoder := json.NewDecoder(reader)
	// Allow unknown fields to handle Yahoo Finance API evolution
	// decoder.DisallowUnknownFields()

	if err := decoder.Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode bars response: %w", err)
	}

	// Validate response structure
	if err := response.Validate(); err != nil {
		return nil, fmt.Errorf("invalid bars response: %w", err)
	}

	return &response, nil
}
