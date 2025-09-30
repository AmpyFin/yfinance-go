package yahoo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// QuoteResponse represents the Yahoo Finance quotes API response
type QuoteResponse struct {
	QuoteResponse QuoteResponseData `json:"quoteResponse"`
}

// QuoteResponseData contains the actual quote data
type QuoteResponseData struct {
	Result []QuoteResult `json:"result"`
	Error  *string       `json:"error"`
}

// QuoteResult contains quote data for a single symbol
type QuoteResult struct {
	Language                   string   `json:"language"`
	Region                     string   `json:"region"`
	QuoteType                  string   `json:"quoteType"`
	TypeDisp                   string   `json:"typeDisp"`
	QuoteSourceName            string   `json:"quoteSourceName"`
	Triggerable                bool     `json:"triggerable"`
	CustomPriceAlertConfidence string   `json:"customPriceAlertConfidence"`
	Currency                   string   `json:"currency"`
	Exchange                   string   `json:"exchange"`
	ShortName                  string   `json:"shortName"`
	LongName                   string   `json:"longName"`
	MessageBoardId             string   `json:"messageBoardId"`
	ExchangeTimezoneName       string   `json:"exchangeTimezoneName"`
	ExchangeTimezoneShortName  string   `json:"exchangeTimezoneShortName"`
	GmtOffsetMilliseconds      int64    `json:"gmtOffSetMilliseconds"`
	Market                     string   `json:"market"`
	EsgPopulated               bool     `json:"esgPopulated"`
	RegularMarketPrice         *float64 `json:"regularMarketPrice"`
	RegularMarketTime          *int64   `json:"regularMarketTime"`
	RegularMarketChange        *float64 `json:"regularMarketChange"`
	RegularMarketOpen          *float64 `json:"regularMarketOpen"`
	RegularMarketDayHigh       *float64 `json:"regularMarketDayHigh"`
	RegularMarketDayLow        *float64 `json:"regularMarketDayLow"`
	RegularMarketVolume        *int64   `json:"regularMarketVolume"`
	Bid                        *float64 `json:"bid"`
	Ask                        *float64 `json:"ask"`
	BidSize                    *int64   `json:"bidSize"`
	AskSize                    *int64   `json:"askSize"`
	FullExchangeName           string   `json:"fullExchangeName"`
	FinancialCurrency          string   `json:"financialCurrency"`
	RegularMarketChangePercent *float64 `json:"regularMarketChangePercent"`
	MarketState                string   `json:"marketState"`
	Symbol                     string   `json:"symbol"`
}

// DecodeQuoteResponse decodes a Yahoo Finance quote response with strict validation
func DecodeQuoteResponse(data []byte) (*QuoteResponse, error) {
	var response QuoteResponse

	// Use strict JSON decoding
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode quote response: %w", err)
	}

	// Validate response structure
	if err := response.Validate(); err != nil {
		return nil, fmt.Errorf("invalid quote response: %w", err)
	}

	return &response, nil
}

// Validate validates the quote response structure
func (r *QuoteResponse) Validate() error {
	if r.QuoteResponse.Error != nil {
		return fmt.Errorf("yahoo finance error: %s", *r.QuoteResponse.Error)
	}

	if len(r.QuoteResponse.Result) == 0 {
		return fmt.Errorf("no quote results found")
	}

	for i, result := range r.QuoteResponse.Result {
		if err := result.Validate(); err != nil {
			return fmt.Errorf("result[%d]: %w", i, err)
		}
	}

	return nil
}

// Validate validates a quote result
func (r *QuoteResult) Validate() error {
	if r.Symbol == "" {
		return fmt.Errorf("missing symbol")
	}

	if r.Currency == "" {
		return fmt.Errorf("missing currency")
	}

	// Validate bid/ask data if present
	if r.Bid != nil && r.Ask != nil {
		if err := validatePrice(*r.Bid); err != nil {
			return fmt.Errorf("invalid bid price: %w", err)
		}
		if err := validatePrice(*r.Ask); err != nil {
			return fmt.Errorf("invalid ask price: %w", err)
		}
		if *r.Bid > *r.Ask {
			return fmt.Errorf("bid > ask: bid=%.4f, ask=%.4f", *r.Bid, *r.Ask)
		}
	}

	// Validate bid/ask sizes if present
	if r.BidSize != nil && *r.BidSize < 0 {
		return fmt.Errorf("negative bid size: %d", *r.BidSize)
	}
	if r.AskSize != nil && *r.AskSize < 0 {
		return fmt.Errorf("negative ask size: %d", *r.AskSize)
	}

	// Validate regular market price if present
	if r.RegularMarketPrice != nil {
		if err := validatePrice(*r.RegularMarketPrice); err != nil {
			return fmt.Errorf("invalid regular market price: %w", err)
		}
	}

	// Validate other price fields if present
	if r.RegularMarketOpen != nil {
		if err := validatePrice(*r.RegularMarketOpen); err != nil {
			return fmt.Errorf("invalid regular market open: %w", err)
		}
	}
	if r.RegularMarketDayHigh != nil {
		if err := validatePrice(*r.RegularMarketDayHigh); err != nil {
			return fmt.Errorf("invalid regular market day high: %w", err)
		}
	}
	if r.RegularMarketDayLow != nil {
		if err := validatePrice(*r.RegularMarketDayLow); err != nil {
			return fmt.Errorf("invalid regular market day low: %w", err)
		}
	}

	// Validate volume if present
	if r.RegularMarketVolume != nil && *r.RegularMarketVolume < 0 {
		return fmt.Errorf("negative regular market volume: %d", *r.RegularMarketVolume)
	}

	return nil
}

// GetQuotes extracts quote data from the response
func (r *QuoteResponse) GetQuotes() []Quote {
	quotes := make([]Quote, 0, len(r.QuoteResponse.Result))

	for _, result := range r.QuoteResponse.Result {
		quote := Quote{
			Symbol:                    result.Symbol,
			Currency:                  result.Currency,
			Exchange:                  result.Exchange,
			FullExchangeName:          result.FullExchangeName,
			ShortName:                 result.ShortName,
			LongName:                  result.LongName,
			QuoteType:                 result.QuoteType,
			MarketState:               result.MarketState,
			ExchangeTimezoneName:      result.ExchangeTimezoneName,
			ExchangeTimezoneShortName: result.ExchangeTimezoneShortName,
			GmtOffsetMilliseconds:     result.GmtOffsetMilliseconds,
		}

		// Copy price data if present
		if result.Bid != nil {
			quote.Bid = result.Bid
		}
		if result.Ask != nil {
			quote.Ask = result.Ask
		}
		if result.BidSize != nil {
			quote.BidSize = result.BidSize
		}
		if result.AskSize != nil {
			quote.AskSize = result.AskSize
		}
		if result.RegularMarketPrice != nil {
			quote.RegularMarketPrice = result.RegularMarketPrice
		}
		if result.RegularMarketTime != nil {
			quote.RegularMarketTime = result.RegularMarketTime
		}
		if result.RegularMarketChange != nil {
			quote.RegularMarketChange = result.RegularMarketChange
		}
		if result.RegularMarketOpen != nil {
			quote.RegularMarketOpen = result.RegularMarketOpen
		}
		if result.RegularMarketDayHigh != nil {
			quote.RegularMarketDayHigh = result.RegularMarketDayHigh
		}
		if result.RegularMarketDayLow != nil {
			quote.RegularMarketDayLow = result.RegularMarketDayLow
		}
		if result.RegularMarketVolume != nil {
			quote.RegularMarketVolume = result.RegularMarketVolume
		}
		if result.RegularMarketChangePercent != nil {
			quote.RegularMarketChangePercent = result.RegularMarketChangePercent
		}

		quotes = append(quotes, quote)
	}

	return quotes
}

// Quote represents a single quote
type Quote struct {
	Symbol                     string   `json:"symbol"`
	Currency                   string   `json:"currency"`
	Exchange                   string   `json:"exchange"`
	FullExchangeName           string   `json:"fullExchangeName"`
	ShortName                  string   `json:"shortName"`
	LongName                   string   `json:"longName"`
	QuoteType                  string   `json:"quoteType"`
	MarketState                string   `json:"marketState"`
	ExchangeTimezoneName       string   `json:"exchangeTimezoneName"`
	ExchangeTimezoneShortName  string   `json:"exchangeTimezoneShortName"`
	GmtOffsetMilliseconds      int64    `json:"gmtOffsetMilliseconds"`
	Bid                        *float64 `json:"bid,omitempty"`
	Ask                        *float64 `json:"ask,omitempty"`
	BidSize                    *int64   `json:"bidSize,omitempty"`
	AskSize                    *int64   `json:"askSize,omitempty"`
	RegularMarketPrice         *float64 `json:"regularMarketPrice,omitempty"`
	RegularMarketTime          *int64   `json:"regularMarketTime,omitempty"`
	RegularMarketChange        *float64 `json:"regularMarketChange,omitempty"`
	RegularMarketOpen          *float64 `json:"regularMarketOpen,omitempty"`
	RegularMarketDayHigh       *float64 `json:"regularMarketDayHigh,omitempty"`
	RegularMarketDayLow        *float64 `json:"regularMarketDayLow,omitempty"`
	RegularMarketVolume        *int64   `json:"regularMarketVolume,omitempty"`
	RegularMarketChangePercent *float64 `json:"regularMarketChangePercent,omitempty"`
}

// DecodeQuoteResponseFromReader decodes a Yahoo Finance quote response from an io.Reader
func DecodeQuoteResponseFromReader(reader io.Reader) (*QuoteResponse, error) {
	var response QuoteResponse

	// Use strict JSON decoding
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode quote response: %w", err)
	}

	// Validate response structure
	if err := response.Validate(); err != nil {
		return nil, fmt.Errorf("invalid quote response: %w", err)
	}

	return &response, nil
}
