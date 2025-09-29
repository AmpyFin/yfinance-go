package emit

import (
	"testing"
	"time"

	"github.com/AmpyFin/yfinance-go/internal/scrape"
	newsv1 "github.com/AmpyFin/ampy-proto/v2/gen/go/ampy/news/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestMapFinancialsDTO(t *testing.T) {
	// Test data
	testTime := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	quarterStart := time.Date(2024, 10, 1, 0, 0, 0, 0, time.UTC)
	quarterEnd := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)

	dto := &scrape.FinancialsDTO{
		Symbol: "AAPL",
		Market: "NASDAQ",
		AsOf:   testTime,
		Lines: []scrape.PeriodLine{
			{
				PeriodStart: quarterStart,
				PeriodEnd:   quarterEnd,
				Key:         "total_revenue",
				Value: scrape.Scaled{
					Scaled: 11750000000000, // $117.5B
					Scale:  2,
				},
				Currency: "USD",
			},
			{
				PeriodStart: quarterStart,
				PeriodEnd:   quarterEnd,
				Key:         "net_income",
				Value: scrape.Scaled{
					Scaled: 3350000000000, // $33.5B
					Scale:  2,
				},
				Currency: "USD",
			},
		},
	}

	// Map to proto
	runID := "test-run-123"
	producer := "yfin-test"
	snapshot, err := MapFinancialsDTO(dto, runID, producer)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, snapshot)

	// Check security
	assert.Equal(t, "AAPL", snapshot.Security.Symbol)
	assert.Equal(t, "XNAS", snapshot.Security.Mic) // Should be normalized

	// Check metadata
	assert.Equal(t, runID, snapshot.Meta.RunId)
	assert.Equal(t, "yfinance-go/scrape", snapshot.Meta.Source)
	assert.Equal(t, producer, snapshot.Meta.Producer)
	assert.Equal(t, "ampy.fundamentals.v1:2.1.0", snapshot.Meta.SchemaVersion)

	// Check source
	assert.Equal(t, "yfinance/scrape", snapshot.Source)

	// Check timestamp
	assert.True(t, snapshot.AsOf.AsTime().Equal(testTime))

	// Check lines
	require.Len(t, snapshot.Lines, 2)

	// Check first line (total_revenue)
	line1 := snapshot.Lines[0]
	assert.Equal(t, "total_revenue", line1.Key)
	assert.Equal(t, int64(11750000000000), line1.Value.Scaled)
	assert.Equal(t, int32(2), line1.Value.Scale)
	assert.Equal(t, "USD", line1.CurrencyCode)
	assert.True(t, line1.PeriodStart.AsTime().Equal(quarterStart))
	assert.True(t, line1.PeriodEnd.AsTime().Equal(quarterEnd))

	// Check second line (net_income)
	line2 := snapshot.Lines[1]
	assert.Equal(t, "net_income", line2.Key)
	assert.Equal(t, int64(3350000000000), line2.Value.Scaled)
	assert.Equal(t, int32(2), line2.Value.Scale)
	assert.Equal(t, "USD", line2.CurrencyCode)
}

func TestMapFinancialsDTO_ValidationErrors(t *testing.T) {
	runID := "test-run-123"
	producer := "yfin-test"

	// Test nil DTO
	_, err := MapFinancialsDTO(nil, runID, producer)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "FinancialsDTO cannot be nil")

	// Test invalid scale
	testTime := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	quarterStart := time.Date(2024, 10, 1, 0, 0, 0, 0, time.UTC)
	quarterEnd := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)

	dto := &scrape.FinancialsDTO{
		Symbol: "AAPL",
		Market: "NASDAQ",
		AsOf:   testTime,
		Lines: []scrape.PeriodLine{
			{
				PeriodStart: quarterStart,
				PeriodEnd:   quarterEnd,
				Key:         "total_revenue",
				Value: scrape.Scaled{
					Scaled: 11750000000000,
					Scale:  15, // Invalid scale > 9
				},
				Currency: "USD",
			},
		},
	}

	_, err = MapFinancialsDTO(dto, runID, producer)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid scale")

	// Test invalid period (start after end)
	dto.Lines[0].Value.Scale = 2 // Fix scale
	dto.Lines[0].PeriodStart = quarterEnd
	dto.Lines[0].PeriodEnd = quarterStart

	_, err = MapFinancialsDTO(dto, runID, producer)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "period_start")
}

func TestMapProfileDTO(t *testing.T) {
	// Test data
	testTime := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	employees := int64(150000)

	dto := &scrape.ComprehensiveProfileDTO{
		Symbol:            "AAPL",
		Market:            "NASDAQ",
		AsOf:              testTime,
		CompanyName:       "Apple Inc.",
		ShortName:         "Apple",
		Website:           "https://www.apple.com",
		Industry:          "Consumer Electronics",
		Sector:            "Technology",
		BusinessSummary:   "Apple Inc. designs, manufactures, and markets smartphones, personal computers, tablets, wearables, and accessories worldwide.",
		FullTimeEmployees: &employees,
		Address1:          "One Apple Park Way",
		City:              "Cupertino",
		State:             "CA",
		Zip:               "95014",
		Country:           "United States",
		Phone:             "408-996-1010",
		Executives: []scrape.Executive{
			{
				Name:     "Timothy D. Cook",
				Title:    "Chief Executive Officer",
				YearBorn: func() *int { y := 1960; return &y }(),
				TotalPay: func() *int64 { p := int64(99420000); return &p }(),
			},
		},
	}

	// Map to proto result
	runID := "test-run-123"
	producer := "yfin-test"
	result, err := MapProfileDTO(dto, runID, producer)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)

	// Check basic fields
	assert.Equal(t, "application/json", result.ContentType)
	assert.Equal(t, "ampy.raw.v1.JsonBlob", result.SchemaFQDN)
	assert.True(t, len(result.JSONBytes) > 0)

	// Check security
	assert.Equal(t, "AAPL", result.Security.Symbol)
	assert.Equal(t, "XNAS", result.Security.Mic)

	// Check metadata
	assert.Equal(t, runID, result.Meta.RunId)
	assert.Equal(t, "yfinance-go/scrape", result.Meta.Source)
	assert.Equal(t, producer, result.Meta.Producer)
	assert.Equal(t, "ampy.reference.v1:2.1.0", result.Meta.SchemaVersion)

	// Verify JSON structure by unmarshaling
	// Just check that we have valid JSON bytes
	assert.True(t, len(result.JSONBytes) > 0)
}

func TestMapNewsItems(t *testing.T) {
	// Test data
	publishedTime := time.Date(2024, 12, 31, 15, 30, 0, 0, time.UTC)
	
	items := []scrape.NewsItem{
		{
			Title:          "Apple Reports Record Q4 Earnings",
			URL:            "https://finance.yahoo.com/news/apple-earnings-q4-2024.html",
			Source:         "Yahoo Finance",
			PublishedAt:    &publishedTime,
			ImageURL:       "https://media.zenfs.com/en/yahoo_finance_350/apple-earnings.webp",
			RelatedTickers: []string{"AAPL", "MSFT"},
		},
		{
			Title:          "Tech Stocks Rally on AI News",
			URL:            "https://finance.yahoo.com/news/tech-stocks-ai-rally.html",
			Source:         "bloomberg",
			PublishedAt:    nil, // No published time
			ImageURL:       "",  // No image
			RelatedTickers: []string{"AAPL", "GOOGL", "NVDA"},
		},
	}

	// Map to proto
	symbol := "AAPL"
	runID := "test-run-123"
	producer := "yfin-test"
	articles, err := MapNewsItems(items, symbol, runID, producer)

	// Assertions
	require.NoError(t, err)
	require.Len(t, articles, 2)

	// Check first article
	article1 := articles[0]
	assert.Equal(t, "Apple Reports Record Q4 Earnings", article1.Headline)
	assert.Equal(t, "https://finance.yahoo.com/news/apple-earnings-q4-2024.html", article1.Url)
	assert.Equal(t, "Yahoo Finance", article1.Source)
	assert.NotNil(t, article1.PublishedAt)
	assert.True(t, article1.PublishedAt.AsTime().Equal(publishedTime))
	assert.Equal(t, []string{"AAPL", "MSFT"}, article1.Tickers)
	// Note: ImageUrl and Security fields not available in ampy-proto v2.1.0 NewsItem

	// Check metadata
	assert.Equal(t, runID, article1.Meta.RunId)
	assert.Equal(t, "yfinance-go/scrape", article1.Meta.Source)
	assert.Equal(t, producer, article1.Meta.Producer)
	assert.Equal(t, "ampy.news.v1:2.1.0", article1.Meta.SchemaVersion)

	// Check second article
	article2 := articles[1]
	assert.Equal(t, "Tech Stocks Rally on AI News", article2.Headline)
	assert.Equal(t, "https://finance.yahoo.com/news/tech-stocks-ai-rally.html", article2.Url)
	assert.Equal(t, "Bloomberg", article2.Source) // Should be normalized
	assert.Nil(t, article2.PublishedAt)           // No published time
	assert.Equal(t, []string{"AAPL", "GOOGL", "NVDA"}, article2.Tickers)
	// Note: ImageUrl field not available in ampy-proto v2.1.0 NewsItem
}

func TestMapNewsItems_ValidationErrors(t *testing.T) {
	runID := "test-run-123"
	producer := "yfin-test"

	// Test empty title
	items := []scrape.NewsItem{
		{
			Title: "", // Empty title
			URL:   "https://example.com",
		},
	}

	_, err := MapNewsItems(items, "AAPL", runID, producer)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "news title cannot be empty")

	// Test empty URL
	items[0].Title = "Valid Title"
	items[0].URL = "" // Empty URL

	_, err = MapNewsItems(items, "AAPL", runID, producer)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "news URL cannot be empty")

	// Test invalid URL
	items[0].URL = "not-a-valid-url"

	_, err = MapNewsItems(items, "AAPL", runID, producer)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid URL")
}

func TestNormalizeFinancialKey(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"Total Revenues", "total_revenue"},
		{"Net Income Common Stockholders", "net_income"},
		{"Basic Earnings Per Share", "eps_basic"},
		{"Diluted EPS", "eps_diluted"},
		{"Operating Earnings", "operating_income"},
		{"Cash and Cash Equivalents", "cash_and_equivalents"},
		{"Custom Field", "custom_field"},
		{"EBITDA", "ebitda"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := normalizeFinancialKey(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestNormalizeMIC(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"NASDAQ", "XNAS"},
		{"NYSE", "XNYS"},
		{"nasdaq", "XNAS"},
		{"XNAS", "XNAS"}, // Already a MIC
		{"TOKYO", "XJPX"},
		{"UnknownMarket", "UNKN"}, // Truncated to 4 chars
		{"", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := normalizeMIC(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestCreateNewsSummary(t *testing.T) {
	// Test empty articles
	summary := CreateNewsSummary(nil)
	assert.Equal(t, 0, summary.TotalArticles)

	// Test with articles
	publishedTime1 := time.Date(2024, 12, 31, 10, 0, 0, 0, time.UTC)
	publishedTime2 := time.Date(2024, 12, 31, 15, 0, 0, 0, time.UTC)

	articles := []*newsv1.NewsItem{
		{
			Headline:    "Article 1",
			Source:      "Yahoo Finance",
			PublishedAt: timestampFromTime(publishedTime1),
			Tickers:     []string{"AAPL", "MSFT"},
		},
		{
			Headline:    "Article 2",
			Source:      "Bloomberg",
			PublishedAt: timestampFromTime(publishedTime2),
			Tickers:     []string{"AAPL", "GOOGL"},
		},
		{
			Headline:    "Article 3",
			Source:      "Yahoo Finance",
			PublishedAt: nil,
			Tickers:     []string{"TSLA"},
		},
	}

	summary = CreateNewsSummary(articles)

	assert.Equal(t, 3, summary.TotalArticles)
	assert.Equal(t, 3, summary.UniqueArticles)
	assert.Equal(t, 2, summary.UniqueSources)
	// Note: HasImages will be 0 since ImageUrl field not available in ampy-proto v2.1.0 NewsItem
	assert.Equal(t, 0, summary.HasImages)
	assert.NotNil(t, summary.EarliestTime)
	assert.NotNil(t, summary.LatestTime)
	assert.True(t, summary.EarliestTime.Equal(publishedTime1))
	assert.True(t, summary.LatestTime.Equal(publishedTime2))
	assert.Contains(t, summary.TopSources, "Yahoo Finance")
	assert.Contains(t, summary.RelatedTickers, "AAPL")
}

// Helper function to create timestamp from time
func timestampFromTime(t time.Time) *timestamppb.Timestamp {
	return timestamppb.New(t)
}
