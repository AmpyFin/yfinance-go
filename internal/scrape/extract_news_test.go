package scrape

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

const (
	yahooFinanceBaseURL = "https://finance.yahoo.com"
)

// TestParseNews tests the main ParseNews function with various fixtures
func TestParseNews(t *testing.T) {
	testCases := []struct {
		name             string
		fixture          string
		expectedMinItems int
		expectedMaxItems int
		expectError      bool
		expectedDeduped  int
	}{
		{
			name:             "AAPL news - happy path",
			fixture:          "AAPL_news.html",
			expectedMinItems: 4,
			expectedMaxItems: 6,
			expectError:      false,
			expectedDeduped:  0,
		},
		{
			name:             "MSFT news - happy path",
			fixture:          "MSFT_news.html",
			expectedMinItems: 4,
			expectedMaxItems: 6,
			expectError:      false,
			expectedDeduped:  0,
		},
		{
			name:             "TSM news - happy path",
			fixture:          "TSM_news.html",
			expectedMinItems: 4,
			expectedMaxItems: 6,
			expectError:      false,
			expectedDeduped:  0,
		},
		{
			name:             "Missing fields - some items skipped",
			fixture:          "AAPL_news_missing_fields.html",
			expectedMinItems: 2, // Only items with both title and URL
			expectedMaxItems: 4,
			expectError:      false,
			expectedDeduped:  0,
		},
		{
			name:             "Duplicates - deduplication works",
			fixture:          "AAPL_news_duplicates.html",
			expectedMinItems: 3, // After deduplication
			expectedMaxItems: 5,
			expectError:      false,
			expectedDeduped:  1, // At least 1 duplicate
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Load fixture
			html, err := loadFixture(tc.fixture)
			if err != nil {
				t.Fatalf("Failed to load fixture %s: %v", tc.fixture, err)
			}

			// Parse news
			now := time.Date(2025, 9, 29, 12, 0, 0, 0, time.UTC)
			baseURL := yahooFinanceBaseURL

			articles, stats, err := ParseNews(html, baseURL, now)

			// Check error expectation
			if tc.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if err != nil {
				return // Skip further checks if error occurred
			}

			// Check article count
			if len(articles) < tc.expectedMinItems {
				t.Errorf("Expected at least %d articles, got %d", tc.expectedMinItems, len(articles))
			}
			if len(articles) > tc.expectedMaxItems {
				t.Errorf("Expected at most %d articles, got %d", tc.expectedMaxItems, len(articles))
			}

			// Check stats
			if stats == nil {
				t.Errorf("Expected stats but got nil")
				return
			}

			if stats.TotalReturned != len(articles) {
				t.Errorf("Stats.TotalReturned (%d) doesn't match actual articles (%d)", stats.TotalReturned, len(articles))
			}

			if stats.Deduped < tc.expectedDeduped {
				t.Errorf("Expected at least %d deduped items, got %d", tc.expectedDeduped, stats.Deduped)
			}

			// Validate article structure
			for i, article := range articles {
				if article.Title == "" {
					t.Errorf("Article %d has empty title", i)
				}
				if article.URL == "" {
					t.Errorf("Article %d has empty URL", i)
				}
				if !isValidURL(article.URL) {
					t.Errorf("Article %d has invalid URL: %s", i, article.URL)
				}

				// Check that published time is not in the future
				if article.PublishedAt != nil && article.PublishedAt.After(now) {
					t.Errorf("Article %d has future published time: %v", i, article.PublishedAt)
				}
			}
		})
	}
}

// TestRelativeTimeConversion tests relative time parsing
func TestRelativeTimeConversion(t *testing.T) {
	now := time.Date(2025, 9, 29, 12, 0, 0, 0, time.UTC)

	testCases := []struct {
		input    string
		expected *time.Time
	}{
		{
			input:    "16m ago",
			expected: timePtr(now.Add(-16 * time.Minute)),
		},
		{
			input:    "2h ago",
			expected: timePtr(now.Add(-2 * time.Hour)),
		},
		{
			input:    "1 day ago",
			expected: timePtr(now.Add(-24 * time.Hour)),
		},
		{
			input:    "3 days ago",
			expected: timePtr(now.Add(-3 * 24 * time.Hour)),
		},
		{
			input:    "1 week ago",
			expected: timePtr(now.Add(-7 * 24 * time.Hour)),
		},
		{
			input:    "2 weeks ago",
			expected: timePtr(now.Add(-2 * 7 * 24 * time.Hour)),
		},
		{
			input:    "yesterday",
			expected: timePtr(time.Date(2025, 9, 28, 0, 0, 0, 0, time.UTC)),
		},
		{
			input:    "invalid time",
			expected: nil,
		},
	}

	// Load regex config
	if err := LoadNewsRegexConfig(); err != nil {
		t.Fatalf("Failed to load news regex config: %v", err)
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := parseRelativeTime(tc.input, now)

			if tc.expected == nil && result != nil {
				t.Errorf("Expected nil but got %v", result)
			}
			if tc.expected != nil && result == nil {
				t.Errorf("Expected %v but got nil", tc.expected)
			}
			if tc.expected != nil && result != nil {
				// Allow small time differences due to processing
				diff := result.Sub(*tc.expected)
				if diff < -time.Second || diff > time.Second {
					t.Errorf("Expected %v but got %v (diff: %v)", tc.expected, result, diff)
				}
			}
		})
	}
}

// TestURLNormalization tests URL cleaning and normalization
func TestURLNormalization(t *testing.T) {
	baseURL := yahooFinanceBaseURL

	testCases := []struct {
		input    string
		expected string
	}{
		{
			input:    "https://finance.yahoo.com/news/article-123.html",
			expected: "https://finance.yahoo.com/news/article-123.html",
		},
		{
			input:    "https://finance.yahoo.com/news/article-123.html?utm_source=feed&utm_medium=rss",
			expected: "https://finance.yahoo.com/news/article-123.html",
		},
		{
			input:    "https://finance.yahoo.com/news/article-123.html?guccounter=1&guce_referrer=test",
			expected: "https://finance.yahoo.com/news/article-123.html",
		},
		{
			input:    "/news/relative-article-456.html",
			expected: "https://finance.yahoo.com/news/relative-article-456.html",
		},
	}

	// Load regex config
	if err := LoadNewsRegexConfig(); err != nil {
		t.Fatalf("Failed to load news regex config: %v", err)
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := normalizeURL(tc.input, baseURL)
			if result != tc.expected {
				t.Errorf("Expected %s but got %s", tc.expected, result)
			}
		})
	}
}

// TestTickerValidation tests ticker symbol validation
func TestTickerValidation(t *testing.T) {
	testCases := []struct {
		input    string
		expected bool
	}{
		{"AAPL", true},
		{"MSFT", true},
		{"TSM", true},
		{"BRK.A", true},
		{"BRK-B", true},
		{"GOOGL", true},
		{"", false},
		{"toolong123", false},
		{"INVALID!", false},
		{"test@symbol", false},
		{"123", true},
		{"A1B2C3", true},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := isValidTicker(tc.input)
			if result != tc.expected {
				t.Errorf("Expected %t for %s but got %t", tc.expected, tc.input, result)
			}
		})
	}
}

// TestDeduplication tests the deduplication logic
func TestDeduplication(t *testing.T) {
	now := time.Date(2025, 9, 29, 12, 0, 0, 0, time.UTC)

	articles := []NewsItem{
		{
			Title:       "Test Article 1",
			URL:         "https://finance.yahoo.com/news/test-1.html",
			Source:      "Test Source",
			PublishedAt: &now,
		},
		{
			Title:       "Test Article 1", // Same URL (after normalization)
			URL:         "https://finance.yahoo.com/news/test-1.html?utm_source=test",
			Source:      "Test Source",
			PublishedAt: &now,
		},
		{
			Title:       "Test Article 2",
			URL:         "https://finance.yahoo.com/news/test-2.html",
			Source:      "Test Source",
			PublishedAt: timePtr(now.Add(-1 * time.Minute)), // Within 2-minute window
		},
		{
			Title:       "Test Article 2", // Same title, source, and time window
			URL:         "https://finance.yahoo.com/news/test-2-different.html",
			Source:      "Test Source",
			PublishedAt: &now,
		},
		{
			Title:       "Test Article 3",
			URL:         "https://finance.yahoo.com/news/test-3.html",
			Source:      "Different Source",
			PublishedAt: &now,
		},
	}

	result := deduplicateArticles(articles)

	// Should have 3 unique articles after deduplication
	if len(result) != 3 {
		t.Errorf("Expected 3 articles after deduplication, got %d", len(result))
	}

	// Check that articles are sorted by published time (newest first)
	for i := 1; i < len(result); i++ {
		if result[i-1].PublishedAt != nil && result[i].PublishedAt != nil {
			if result[i-1].PublishedAt.Before(*result[i].PublishedAt) {
				t.Errorf("Articles not sorted by published time")
			}
		}
	}
}

// TestErrorCases tests various error conditions
func TestErrorCases(t *testing.T) {
	now := time.Date(2025, 9, 29, 12, 0, 0, 0, time.UTC)
	baseURL := yahooFinanceBaseURL

	testCases := []struct {
		name        string
		html        string
		expectError bool
		errorType   string
	}{
		{
			name:        "Empty HTML",
			html:        "",
			expectError: true,
			errorType:   "news_no_articles",
		},
		{
			name:        "No article containers",
			html:        "<html><body><div>No news here</div></body></html>",
			expectError: true,
			errorType:   "news_no_articles",
		},
		{
			name: "Articles without required fields",
			html: `<section class="container" data-testid="storyitem" role="article">
				<div>No title or URL</div>
			</section>`,
			expectError: true,
			errorType:   "news_no_articles",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := ParseNews([]byte(tc.html), baseURL, now)

			if tc.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if err != nil && tc.errorType != "" {
				if scrapeErr, ok := err.(*ScrapeError); ok {
					if scrapeErr.Type != tc.errorType {
						t.Errorf("Expected error type %s but got %s", tc.errorType, scrapeErr.Type)
					}
				} else {
					t.Errorf("Expected ScrapeError but got %T", err)
				}
			}
		})
	}
}

// Helper functions

func loadFixture(filename string) ([]byte, error) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return nil, fmt.Errorf("unable to get current file path")
	}

	// Navigate to project root and then to fixtures
	// From internal/scrape -> internal -> root
	projectRoot := filepath.Dir(filepath.Dir(filepath.Dir(currentFile)))
	fixturePath := filepath.Join(projectRoot, "testdata", "fixtures", "yahoo", "news", filename)

	return os.ReadFile(fixturePath)
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func isValidURL(url string) bool {
	return url != "" && (len(url) > 10) && (url[:4] == "http")
}

// Benchmark tests

func BenchmarkParseNews(b *testing.B) {
	html, err := loadFixture("AAPL_news.html")
	if err != nil {
		b.Fatalf("Failed to load fixture: %v", err)
	}

	now := time.Date(2025, 9, 29, 12, 0, 0, 0, time.UTC)
	baseURL := yahooFinanceBaseURL

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := ParseNews(html, baseURL, now)
		if err != nil {
			b.Fatalf("ParseNews failed: %v", err)
		}
	}
}

func BenchmarkDeduplication(b *testing.B) {
	now := time.Date(2025, 9, 29, 12, 0, 0, 0, time.UTC)

	// Create a large set of articles with some duplicates
	articles := make([]NewsItem, 100)
	for i := 0; i < 100; i++ {
		articles[i] = NewsItem{
			Title:       fmt.Sprintf("Test Article %d", i%20), // 20 unique titles (5 duplicates each)
			URL:         fmt.Sprintf("https://finance.yahoo.com/news/test-%d.html", i),
			Source:      "Test Source",
			PublishedAt: &now,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		deduplicateArticles(articles)
	}
}
