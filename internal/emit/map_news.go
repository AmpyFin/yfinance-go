package emit

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/AmpyFin/yfinance-go/internal/scrape"
	commonv1 "github.com/AmpyFin/ampy-proto/v2/gen/go/ampy/common/v1"
	newsv1 "github.com/AmpyFin/ampy-proto/v2/gen/go/ampy/news/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// MapNewsItems converts slice of NewsItem to slice of ampy.news.v1.NewsItem
func MapNewsItems(items []scrape.NewsItem, symbol string, runID, producer string) ([]*newsv1.NewsItem, error) {
	if len(items) == 0 {
		return nil, nil
	}

	articles := make([]*newsv1.NewsItem, 0, len(items))
	
	for i, item := range items {
		article, err := mapSingleNewsItem(&item, symbol, runID, producer)
		if err != nil {
			return nil, fmt.Errorf("failed to map news item %d (%s): %w", i, item.Title, err)
		}
		
		if article != nil { // Skip nil articles (filtered out)
			articles = append(articles, article)
		}
	}

	return articles, nil
}

// mapSingleNewsItem converts a single NewsItem to ampy.news.v1.NewsItem
func mapSingleNewsItem(item *scrape.NewsItem, symbol, runID, producer string) (*newsv1.NewsItem, error) {
	if item == nil {
		return nil, fmt.Errorf("NewsItem cannot be nil")
	}

	// Validate required fields
	if item.Title == "" {
		return nil, fmt.Errorf("news title cannot be empty")
	}
	
	if item.URL == "" {
		return nil, fmt.Errorf("news URL cannot be empty")
	}

	// Validate and normalize URL
	normalizedURL, err := normalizeNewsURL(item.URL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL '%s': %w", item.URL, err)
	}

	// Note: Security field not available in ampy-proto v2.1.0 NewsItem
	// Primary ticker information is stored in the Tickers field

	// Convert published time (optional)
	var publishedAt *timestamppb.Timestamp
	if item.PublishedAt != nil {
		// Validate timestamp is not in the future
		now := time.Now().UTC()
		pubTime := item.PublishedAt.UTC()
		
		if pubTime.After(now.Add(5 * time.Minute)) { // Allow 5 minute clock skew
			// Log warning and clamp to now
			pubTime = now
		}
		
		publishedAt = timestamppb.New(pubTime)
	}

	// Clean and validate source
	source := cleanNewsSource(item.Source)

	// Note: ImageUrl field not available in ampy-proto v2.1.0 NewsItem
	// Image information would need to be stored elsewhere if needed

	// Validate and clean related tickers
	relatedTickers := cleanRelatedTickers(item.RelatedTickers)

	// Create metadata
	meta := &commonv1.Meta{
		RunId:         runID,
		Source:        "yfinance-go/scrape",
		Producer:      producer,
		SchemaVersion: "ampy.news.v1:2.1.0",
	}

	return &newsv1.NewsItem{
		Headline:    strings.TrimSpace(item.Title),
		Url:         normalizedURL,
		Source:      source,
		PublishedAt: publishedAt,
		Tickers:     relatedTickers,
		Meta:        meta,
		// Note: ImageUrl and Security fields not available in ampy-proto v2.1.0 NewsItem
		// Body field could be populated if we had article content
	}, nil
}

// normalizeNewsURL validates and normalizes news URLs
func normalizeNewsURL(rawURL string) (string, error) {
	if rawURL == "" {
		return "", fmt.Errorf("URL cannot be empty")
	}

	// Parse URL
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %w", err)
	}

	// Ensure URL has scheme
	if parsedURL.Scheme == "" {
		parsedURL.Scheme = "https"
	}

	// Validate scheme
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return "", fmt.Errorf("invalid URL scheme '%s', must be http or https", parsedURL.Scheme)
	}

	// Ensure URL has host
	if parsedURL.Host == "" {
		return "", fmt.Errorf("URL must have a host")
	}

	// Remove tracking parameters
	cleanURL := removeTrackingParams(parsedURL)

	return cleanURL.String(), nil
}

// removeTrackingParams removes common tracking parameters from URLs
func removeTrackingParams(u *url.URL) *url.URL {
	// List of tracking parameters to remove
	trackingParams := map[string]bool{
		"utm_source":    true,
		"utm_medium":    true,
		"utm_campaign":  true,
		"utm_term":      true,
		"utm_content":   true,
		"fbclid":        true,
		"gclid":         true,
		"msclkid":       true,
		"mc_cid":        true,
		"mc_eid":        true,
		"_ga":           true,
		"_gid":          true,
		"ref":           true,
		"referer":       true,
		"referrer":      true,
		"source":        true,
	}

	// Create new URL to avoid modifying original
	cleanURL := *u
	
	// Filter query parameters
	if u.RawQuery != "" {
		values := u.Query()
		cleanValues := url.Values{}
		
		for key, vals := range values {
			if !trackingParams[strings.ToLower(key)] {
				cleanValues[key] = vals
			}
		}
		
		cleanURL.RawQuery = cleanValues.Encode()
	}

	// Remove fragment (hash)
	cleanURL.Fragment = ""

	return &cleanURL
}

// cleanNewsSource normalizes news source names
func cleanNewsSource(source string) string {
	if source == "" {
		return "unknown"
	}

	cleaned := strings.TrimSpace(source)
	if cleaned == "" {
		return "unknown"
	}

	// Normalize common source names
	sourceMappings := map[string]string{
		"yahoo finance":        "Yahoo Finance",
		"yahoo! finance":       "Yahoo Finance",
		"yahoo":                "Yahoo Finance",
		"bloomberg":            "Bloomberg",
		"bloomberg.com":        "Bloomberg",
		"reuters":              "Reuters",
		"reuters.com":          "Reuters",
		"marketwatch":          "MarketWatch",
		"marketwatch.com":      "MarketWatch",
		"cnbc":                 "CNBC",
		"cnbc.com":             "CNBC",
		"cnn business":         "CNN Business",
		"cnn":                  "CNN Business",
		"fox business":         "Fox Business",
		"foxbusiness":          "Fox Business",
		"wall street journal":  "Wall Street Journal",
		"wsj":                  "Wall Street Journal",
		"wsj.com":              "Wall Street Journal",
		"financial times":      "Financial Times",
		"ft.com":               "Financial Times",
		"investing.com":        "Investing.com",
		"seeking alpha":        "Seeking Alpha",
		"seekingalpha":         "Seeking Alpha",
		"seekingalpha.com":     "Seeking Alpha",
		"barron's":             "Barron's",
		"barrons":              "Barron's",
		"barrons.com":          "Barron's",
		"thestreet":            "TheStreet",
		"thestreet.com":        "TheStreet",
		"motley fool":          "The Motley Fool",
		"fool.com":             "The Motley Fool",
	}

	lowerSource := strings.ToLower(cleaned)
	if normalized, exists := sourceMappings[lowerSource]; exists {
		return normalized
	}

	// Capitalize first letter of each word
	words := strings.Fields(cleaned)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
		}
	}

	return strings.Join(words, " ")
}

// cleanRelatedTickers validates and cleans ticker symbols
func cleanRelatedTickers(tickers []string) []string {
	if len(tickers) == 0 {
		return nil
	}

	var cleaned []string
	seen := make(map[string]bool)

	for _, ticker := range tickers {
		cleanTicker := strings.TrimSpace(strings.ToUpper(ticker))
		
		// Basic validation
		if cleanTicker == "" {
			continue
		}
		
		// Remove duplicates
		if seen[cleanTicker] {
			continue
		}
		seen[cleanTicker] = true

		// Basic ticker format validation (1-10 alphanumeric chars, possibly with dots)
		if isValidTicker(cleanTicker) {
			cleaned = append(cleaned, cleanTicker)
		}
	}

	return cleaned
}

// isValidTicker validates ticker symbol format
func isValidTicker(ticker string) bool {
	if len(ticker) == 0 || len(ticker) > 10 {
		return false
	}

	// Allow alphanumeric characters, dots, and hyphens
	for _, char := range ticker {
		if !((char >= 'A' && char <= 'Z') || 
			 (char >= '0' && char <= '9') || 
			 char == '.' || char == '-') {
			return false
		}
	}

	return true
}

// NewsSummary provides a concise summary of news data for preview
type NewsSummary struct {
	TotalArticles    int       `json:"total_articles"`
	UniqueArticles   int       `json:"unique_articles"`
	UniqueSources    int       `json:"unique_sources"`
	EarliestTime     *time.Time `json:"earliest_time,omitempty"`
	LatestTime       *time.Time `json:"latest_time,omitempty"`
	TopSources       []string  `json:"top_sources"`
	RelatedTickers   []string  `json:"related_tickers"`
	HasImages        int       `json:"has_images"`
	AverageURLLength int       `json:"average_url_length"`
}

// CreateNewsSummary creates a summary of news articles for preview
func CreateNewsSummary(articles []*newsv1.NewsItem) *NewsSummary {
	if len(articles) == 0 {
		return &NewsSummary{}
	}

	summary := &NewsSummary{
		TotalArticles:  len(articles),
		UniqueArticles: len(articles), // Assuming deduplication already happened
	}

	// Track sources and tickers
	sources := make(map[string]int)
	tickers := make(map[string]bool)
	urlLengths := 0
	imageCount := 0

	var earliest, latest *time.Time

	for _, article := range articles {
		// Count sources
		if article.Source != "" {
			sources[article.Source]++
		}

	// Collect tickers
	for _, ticker := range article.Tickers {
		tickers[ticker] = true
	}

		// Track URL lengths
		urlLengths += len(article.Url)

		// Note: ImageUrl not available in ampy-proto v2.1.0 NewsItem
		// imageCount remains 0

		// Track time range
		if article.PublishedAt != nil {
			pubTime := article.PublishedAt.AsTime()
			if earliest == nil || pubTime.Before(*earliest) {
				earliest = &pubTime
			}
			if latest == nil || pubTime.After(*latest) {
				latest = &pubTime
			}
		}
	}

	summary.UniqueSources = len(sources)
	summary.HasImages = imageCount
	summary.AverageURLLength = urlLengths / len(articles)
	summary.EarliestTime = earliest
	summary.LatestTime = latest

	// Get top sources (up to 5)
	type sourceCount struct {
		name  string
		count int
	}
	var sortedSources []sourceCount
	for source, count := range sources {
		sortedSources = append(sortedSources, sourceCount{source, count})
	}
	
	// Simple bubble sort by count (descending)
	for i := 0; i < len(sortedSources)-1; i++ {
		for j := i + 1; j < len(sortedSources); j++ {
			if sortedSources[i].count < sortedSources[j].count {
				sortedSources[i], sortedSources[j] = sortedSources[j], sortedSources[i]
			}
		}
	}

	for i, source := range sortedSources {
		if i >= 5 { // Limit to top 5
			break
		}
		summary.TopSources = append(summary.TopSources, source.name)
	}

	// Get related tickers (up to 10)
	tickerCount := 0
	for ticker := range tickers {
		if tickerCount >= 10 {
			break
		}
		summary.RelatedTickers = append(summary.RelatedTickers, ticker)
		tickerCount++
	}

	return summary
}
