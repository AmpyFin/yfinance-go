package emit

import (
	"context"
	"fmt"

	commonv1 "github.com/AmpyFin/ampy-proto/v2/gen/go/ampy/common/v1"
	fundamentalsv1 "github.com/AmpyFin/ampy-proto/v2/gen/go/ampy/fundamentals/v1"
	newsv1 "github.com/AmpyFin/ampy-proto/v2/gen/go/ampy/news/v1"
	"github.com/AmpyFin/yfinance-go/internal/scrape"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ScrapeMapperConfig holds configuration for scrape mapping
type ScrapeMapperConfig struct {
	RunID    string
	Producer string
	Source   string
	TraceID  string
}

// ScrapeMapper converts scrape DTOs to ampy-proto messages
type ScrapeMapper struct {
	config ScrapeMapperConfig
}

// NewScrapeMapper creates a new scrape mapper with the given configuration
func NewScrapeMapper(config ScrapeMapperConfig) *ScrapeMapper {
	if config.Source == "" {
		config.Source = "yfinance-go/scrape"
	}
	return &ScrapeMapper{
		config: config,
	}
}

// MapFinancials converts FinancialsDTO to ampy.fundamentals.v1.FundamentalsSnapshot
func (m *ScrapeMapper) MapFinancials(ctx context.Context, dto *scrape.FinancialsDTO) (*fundamentalsv1.FundamentalsSnapshot, error) {
	if dto == nil {
		return nil, fmt.Errorf("FinancialsDTO cannot be nil")
	}

	// Convert security
	security := &commonv1.SecurityId{
		Symbol: dto.Symbol,
		Mic:    dto.Market, // Use market as MIC for now
	}

	// Convert line items
	lines := make([]*fundamentalsv1.LineItem, 0, len(dto.Lines))
	for i, line := range dto.Lines {
		ampyLine, err := m.mapPeriodLine(&line)
		if err != nil {
			return nil, fmt.Errorf("failed to map line item %d: %w", i, err)
		}
		lines = append(lines, ampyLine)
	}

	// Create metadata
	meta := m.createMeta("ampy.fundamentals.v1:2.1.0")

	return &fundamentalsv1.FundamentalsSnapshot{
		Security: security,
		Lines:    lines,
		Source:   m.config.Source + "/financials",
		AsOf:     timestamppb.New(dto.AsOf),
		Meta:     meta,
	}, nil
}

// MapProfile converts ProfileDTO to JSON bytes (fallback since reference schema may not be available)
func (m *ScrapeMapper) MapProfile(ctx context.Context, dto *scrape.ComprehensiveProfileDTO) ([]byte, error) {
	if dto == nil {
		return nil, fmt.Errorf("ProfileDTO cannot be nil")
	}

	// For now, use JSON fallback as reference.Company may not be available in ampy-proto v2
	// This follows the instruction to emit JSON export when schema isn't available
	canonicalJSON, err := CanonicalMarshaler.Marshal(dto)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal profile to JSON: %w", err)
	}

	return canonicalJSON, nil
}

// MapNews converts slice of NewsItem to slice of ampy.news.v1.NewsItem
func (m *ScrapeMapper) MapNews(ctx context.Context, items []scrape.NewsItem, symbol string) ([]*newsv1.NewsItem, error) {
	if len(items) == 0 {
		return nil, nil
	}

	articles := make([]*newsv1.NewsItem, 0, len(items))

	for i, item := range items {
		article, err := m.mapNewsItem(&item, symbol)
		if err != nil {
			return nil, fmt.Errorf("failed to map news item %d: %w", i, err)
		}
		articles = append(articles, article)
	}

	return articles, nil
}

// mapPeriodLine converts a PeriodLine to ampy.fundamentals.v1.LineItem
func (m *ScrapeMapper) mapPeriodLine(line *scrape.PeriodLine) (*fundamentalsv1.LineItem, error) {
	// Convert scaled decimal
	value := &commonv1.Decimal{
		Scaled: line.Value.Scaled,
		Scale:  int32(line.Value.Scale),
	}

	// Validate decimal
	if line.Value.Scale < 0 || line.Value.Scale > 9 {
		return nil, fmt.Errorf("invalid scale %d, must be between 0 and 9", line.Value.Scale)
	}

	// Convert currency
	currencyCode := string(line.Currency)
	if currencyCode == "" {
		currencyCode = "USD" // Default fallback
	}

	return &fundamentalsv1.LineItem{
		Key:          line.Key,
		Value:        value,
		CurrencyCode: currencyCode,
		PeriodStart:  timestamppb.New(line.PeriodStart),
		PeriodEnd:    timestamppb.New(line.PeriodEnd),
	}, nil
}

// mapNewsItem converts a NewsItem to ampy.news.v1.NewsItem
func (m *ScrapeMapper) mapNewsItem(item *scrape.NewsItem, symbol string) (*newsv1.NewsItem, error) {
	// Note: Security field not available in ampy-proto v2.1.0 NewsItem
	// Primary ticker information is stored in the Tickers field

	// Convert published time (optional)
	var publishedAt *timestamppb.Timestamp
	if item.PublishedAt != nil {
		publishedAt = timestamppb.New(*item.PublishedAt)
	}

	// Create metadata
	meta := m.createMeta("ampy.news.v1:2.1.0")

	return &newsv1.NewsItem{
		Headline:    item.Title,
		Url:         item.URL,
		Source:      item.Source,
		PublishedAt: publishedAt,
		Tickers:     item.RelatedTickers,
		Meta:        meta,
		// Note: ImageUrl and Security fields not available in ampy-proto v2.1.0 NewsItem
	}, nil
}

// createMeta creates metadata for ampy-proto messages
func (m *ScrapeMapper) createMeta(schemaVersion string) *commonv1.Meta {
	return &commonv1.Meta{
		RunId:         m.config.RunID,
		Source:        m.config.Source,
		Producer:      m.config.Producer,
		SchemaVersion: schemaVersion,
		// Note: Checksum and ProducedAt can be added if needed
	}
}

// ScaledFromFloat converts a float64 to a scaled decimal with the given scale
func ScaledFromFloat(value float64, scale int) *commonv1.Decimal {
	if scale < 0 || scale > 9 {
		scale = 2 // Default to 2 decimal places for currency
	}

	multiplier := int64(1)
	for i := 0; i < scale; i++ {
		multiplier *= 10
	}

	scaled := int64(value * float64(multiplier))

	return &commonv1.Decimal{
		Scaled: scaled,
		Scale:  int32(scale),
	}
}
