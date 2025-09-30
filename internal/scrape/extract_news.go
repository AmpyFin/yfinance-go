package scrape

import (
	"encoding/json"
	"fmt"
	"html"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	bodyPatternRegex   = `"body":"(\{.*?\})"`
	scriptPatternRegex = `<script[^>]*>([^<]*tickerStream[^<]*)</script>`
)

// NewsRegexConfig holds the regex patterns for news extraction
type NewsRegexConfig struct {
	ArticleContainer string `yaml:"article_container"`
	Title            string `yaml:"title"`
	ArticleLink      string `yaml:"article_link"`
	PublishingInfo   string `yaml:"publishing_info"`
	ImageURL         string `yaml:"image_url"`
	RelatedTickers   string `yaml:"related_tickers"`
	NextPageHint     string `yaml:"next_page_hint"`

	RelativeTime struct {
		Minutes   string `yaml:"minutes"`
		Hours     string `yaml:"hours"`
		Days      string `yaml:"days"`
		Weeks     string `yaml:"weeks"`
		Yesterday string `yaml:"yesterday"`
	} `yaml:"relative_time"`

	URLCleanup struct {
		UTMParams      string `yaml:"utm_params"`
		TrackingParams string `yaml:"tracking_params"`
		Fragment       string `yaml:"fragment"`
		QuerySeparator string `yaml:"query_separator"`
	} `yaml:"url_cleanup"`
}

var newsRegexConfig *NewsRegexConfig

// LoadNewsRegexConfig loads the news regex patterns from YAML file
func LoadNewsRegexConfig() error {
	if newsRegexConfig != nil {
		return nil // Already loaded
	}

	// Get the directory of the current file
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return fmt.Errorf("unable to get current file path")
	}

	configPath := filepath.Join(filepath.Dir(filename), "regex", "news.yaml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read news regex config file: %w", err)
	}

	newsRegexConfig = &NewsRegexConfig{}
	if err := yaml.Unmarshal(data, newsRegexConfig); err != nil {
		return fmt.Errorf("failed to parse news regex config YAML: %w", err)
	}

	return nil
}

// ParseNews extracts news articles from HTML with robust error handling and deduplication
func ParseNews(html []byte, baseURL string, now time.Time) ([]NewsItem, *NewsStats, error) {
	start := time.Now()

	// Initialize metrics
	metrics := NewMetrics()
	defer func() {
		metrics.RecordNewsParseLatency(time.Since(start))
	}()

	htmlStr := string(html)

	// Try JSON-based extraction first (for real Yahoo Finance pages)
	articles, err := extractNewsFromJSON(htmlStr, baseURL, now)
	if err == nil && len(articles) > 0 {
		// JSON extraction successful
		originalCount := len(articles)
		articles = deduplicateArticles(articles)
		deduped := originalCount - len(articles)

		// Limit results (default 25 articles)
		const maxArticles = 25
		if len(articles) > maxArticles {
			articles = articles[:maxArticles]
		}

		// Extract pagination hint
		nextPageHint := extractNextPageHint(htmlStr)

		// Create statistics
		stats := &NewsStats{
			TotalFound:    originalCount,
			TotalReturned: len(articles),
			Deduped:       deduped,
			NextPageHint:  nextPageHint,
			AsOf:          now.UTC(),
		}

		metrics.RecordNews("success")
		return articles, stats, nil
	}

	// Fall back to HTML-based extraction (for test fixtures or other formats)
	return parseNewsFromHTML(htmlStr, baseURL, now, metrics)
}

// extractArticleContainers finds all article containers in the HTML
func extractArticleContainers(html string) ([]string, error) {
	re, err := regexp.Compile(newsRegexConfig.ArticleContainer)
	if err != nil {
		return nil, fmt.Errorf("invalid article container regex: %w", err)
	}

	matches := re.FindAllStringSubmatch(html, -1)
	var containers []string

	for _, match := range matches {
		if len(match) > 1 {
			containers = append(containers, match[1])
		}
	}

	return containers, nil
}

// parseArticleFromContainer extracts article data from a single container
func parseArticleFromContainer(container, baseURL string, now time.Time) *NewsItem {
	article := &NewsItem{}

	// Extract title
	title := extractStringFromContainer(container, newsRegexConfig.Title)
	if title == "" {
		return nil // Skip articles without title
	}
	article.Title = html.UnescapeString(strings.TrimSpace(title))

	// Extract URL
	articleURL := extractStringFromContainer(container, newsRegexConfig.ArticleLink)
	if articleURL == "" {
		return nil // Skip articles without URL
	}
	article.URL = normalizeURL(articleURL, baseURL)

	// Extract publishing info (source and time)
	publishingInfo := extractStringFromContainer(container, newsRegexConfig.PublishingInfo)
	if publishingInfo != "" {
		source, publishedAt := parsePublishingInfo(publishingInfo, now)
		article.Source = source
		article.PublishedAt = publishedAt
	}

	// Extract image URL (optional)
	imageURL := extractStringFromContainer(container, newsRegexConfig.ImageURL)
	if imageURL != "" {
		article.ImageURL = imageURL
	}

	// Extract related tickers
	article.RelatedTickers = extractRelatedTickers(container)

	return article
}

// extractStringFromContainer extracts a string using regex from a container
func extractStringFromContainer(container, pattern string) string {
	if pattern == "" {
		return ""
	}

	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(container)

	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	return ""
}

// normalizeURL converts relative URLs to absolute and cleans tracking parameters
func normalizeURL(articleURL, baseURL string) string {
	// Make URL absolute
	if !strings.HasPrefix(articleURL, "http") {
		if u, err := url.Parse(baseURL); err == nil {
			if parsed, err := url.Parse(articleURL); err == nil {
				articleURL = u.ResolveReference(parsed).String()
			}
		}
	}

	// Clean tracking parameters
	articleURL = cleanTrackingParams(articleURL)

	return articleURL
}

// cleanTrackingParams removes UTM and other tracking parameters
func cleanTrackingParams(urlStr string) string {
	patterns := []string{
		newsRegexConfig.URLCleanup.UTMParams,
		newsRegexConfig.URLCleanup.TrackingParams,
	}

	for _, pattern := range patterns {
		if pattern != "" {
			re := regexp.MustCompile(pattern)
			urlStr = re.ReplaceAllString(urlStr, "")
		}
	}

	// Clean up any remaining & at the end or beginning of query string
	urlStr = regexp.MustCompile(`[?&]+$`).ReplaceAllString(urlStr, "")
	urlStr = regexp.MustCompile(`\?&`).ReplaceAllString(urlStr, "?")

	return urlStr
}

// parsePublishingInfo extracts source and published time from publishing info
func parsePublishingInfo(info string, now time.Time) (string, *time.Time) {
	// Split on bullet point or similar separators
	parts := regexp.MustCompile(`\s*[•·|]\s*`).Split(info, -1)

	var source string
	var publishedAt *time.Time

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Try to parse as relative time
		if parsedTime := parseRelativeTime(part, now); parsedTime != nil {
			publishedAt = parsedTime
		} else {
			// Assume it's the source
			source = part
		}
	}

	return source, publishedAt
}

// parseRelativeTime converts relative time strings to absolute time
func parseRelativeTime(timeStr string, now time.Time) *time.Time {
	timeStr = strings.ToLower(strings.TrimSpace(timeStr))

	// Minutes ago
	if re := regexp.MustCompile(newsRegexConfig.RelativeTime.Minutes); re != nil {
		if matches := re.FindStringSubmatch(timeStr); len(matches) > 1 {
			if minutes, err := strconv.Atoi(matches[1]); err == nil {
				result := now.Add(-time.Duration(minutes) * time.Minute).UTC()
				return &result
			}
		}
	}

	// Hours ago
	if re := regexp.MustCompile(newsRegexConfig.RelativeTime.Hours); re != nil {
		if matches := re.FindStringSubmatch(timeStr); len(matches) > 1 {
			if hours, err := strconv.Atoi(matches[1]); err == nil {
				result := now.Add(-time.Duration(hours) * time.Hour).UTC()
				return &result
			}
		}
	}

	// Days ago
	if re := regexp.MustCompile(newsRegexConfig.RelativeTime.Days); re != nil {
		if matches := re.FindStringSubmatch(timeStr); len(matches) > 1 {
			if days, err := strconv.Atoi(matches[1]); err == nil {
				result := now.Add(-time.Duration(days) * 24 * time.Hour).UTC()
				return &result
			}
		}
	}

	// Weeks ago
	if re := regexp.MustCompile(newsRegexConfig.RelativeTime.Weeks); re != nil {
		if matches := re.FindStringSubmatch(timeStr); len(matches) > 1 {
			if weeks, err := strconv.Atoi(matches[1]); err == nil {
				result := now.Add(-time.Duration(weeks) * 7 * 24 * time.Hour).UTC()
				return &result
			}
		}
	}

	// Yesterday
	if re := regexp.MustCompile(newsRegexConfig.RelativeTime.Yesterday); re != nil {
		if re.MatchString(timeStr) {
			// Set to start of yesterday (conservative approach)
			yesterday := now.Add(-24 * time.Hour)
			result := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, time.UTC)
			return &result
		}
	}

	// Ensure no future times
	return nil
}

// extractRelatedTickers finds ticker symbols in the container
func extractRelatedTickers(container string) []string {
	if newsRegexConfig.RelatedTickers == "" {
		return nil
	}

	re := regexp.MustCompile(newsRegexConfig.RelatedTickers)
	matches := re.FindAllStringSubmatch(container, -1)

	var tickers []string
	tickerSet := make(map[string]bool) // For deduplication

	for _, match := range matches {
		if len(match) > 1 {
			ticker := strings.ToUpper(strings.TrimSpace(match[1]))
			// Validate ticker format (A-Z, 0-9, ., -)
			if isValidTicker(ticker) && !tickerSet[ticker] {
				tickers = append(tickers, ticker)
				tickerSet[ticker] = true
			}
		}
	}

	return tickers
}

// isValidTicker validates ticker symbol format
func isValidTicker(ticker string) bool {
	if len(ticker) == 0 || len(ticker) > 10 {
		return false
	}

	for _, char := range ticker {
		if !((char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '.' || char == '-') {
			return false
		}
	}

	return true
}

// extractNextPageHint looks for pagination controls
func extractNextPageHint(html string) string {
	if newsRegexConfig.NextPageHint == "" {
		return ""
	}

	re := regexp.MustCompile(newsRegexConfig.NextPageHint)
	matches := re.FindStringSubmatch(html)

	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	return ""
}

// deduplicateArticles removes duplicate articles using URL and content heuristics
func deduplicateArticles(articles []NewsItem) []NewsItem {
	seen := make(map[string]bool)
	var result []NewsItem

	for _, article := range articles {
		// Primary dedup key: normalized URL
		normalizedURL := normalizeURLForDedup(article.URL)
		if seen[normalizedURL] {
			continue
		}

		// Secondary dedup: check for similar articles by title, source, and time
		isDuplicate := false
		titleNorm := strings.ToLower(strings.TrimSpace(article.Title))
		sourceNorm := strings.ToLower(strings.TrimSpace(article.Source))

		for _, existing := range result {
			existingTitleNorm := strings.ToLower(strings.TrimSpace(existing.Title))
			existingSourceNorm := strings.ToLower(strings.TrimSpace(existing.Source))

			// Check if title and source match
			if titleNorm == existingTitleNorm && sourceNorm == existingSourceNorm {
				// Check if times are within 2 minutes of each other
				if article.PublishedAt != nil && existing.PublishedAt != nil {
					timeDiff := article.PublishedAt.Sub(*existing.PublishedAt)
					if timeDiff < 0 {
						timeDiff = -timeDiff
					}
					if timeDiff <= 2*time.Minute {
						isDuplicate = true
						break
					}
				} else if article.PublishedAt == nil && existing.PublishedAt == nil {
					// Both have no timestamp, consider duplicate
					isDuplicate = true
					break
				}
			}
		}

		if isDuplicate {
			continue
		}

		seen[normalizedURL] = true
		result = append(result, article)
	}

	// Sort by published time (newest first)
	sort.Slice(result, func(i, j int) bool {
		if result[i].PublishedAt == nil && result[j].PublishedAt == nil {
			return false
		}
		if result[i].PublishedAt == nil {
			return false
		}
		if result[j].PublishedAt == nil {
			return true
		}
		return result[i].PublishedAt.After(*result[j].PublishedAt)
	})

	return result
}

// normalizeURLForDedup normalizes URL for deduplication
func normalizeURLForDedup(urlStr string) string {
	if u, err := url.Parse(urlStr); err == nil {
		// Lowercase host, remove query and fragment
		u.Host = strings.ToLower(u.Host)
		u.RawQuery = ""
		u.Fragment = ""
		return u.String()
	}
	return strings.ToLower(urlStr)
}

// extractNewsFromJSON extracts news from JSON data embedded in script tags
func extractNewsFromJSON(html, baseURL string, now time.Time) ([]NewsItem, error) {
	// Look for the script tag containing the tickerStream data
	scriptPattern := scriptPatternRegex
	scriptRe := regexp.MustCompile(scriptPattern)
	scriptMatches := scriptRe.FindStringSubmatch(html)

	if len(scriptMatches) < 2 {
		return nil, fmt.Errorf("no tickerStream script found")
	}

	jsonContent := scriptMatches[1]

	// The JSON is nested - extract the body content which contains the actual news data
	bodyPattern := bodyPatternRegex
	bodyRe := regexp.MustCompile(bodyPattern)
	bodyMatches := bodyRe.FindStringSubmatch(jsonContent)

	if len(bodyMatches) < 2 {
		return nil, fmt.Errorf("no body content found in script")
	}

	// The body content is escaped JSON, so we need to unescape it using JSON decoder
	raw := bodyMatches[1]
	var unescaped string
	if err := json.Unmarshal([]byte("\""+raw+"\""), &unescaped); err != nil {
		// Fallback simple unescape
		unescaped = strings.ReplaceAll(raw, `\\`, `\`)
		unescaped = strings.ReplaceAll(unescaped, `\"`, `"`)
	}

	// Now extract individual articles from the content arrays
	return extractArticlesFromNewsJSON(unescaped, baseURL, now)
}

// extractNewsFromFullJSON attempts to extract from the full JSON structure
func extractNewsFromFullJSON(html, baseURL string, now time.Time) ([]NewsItem, error) {
	// Look for the full data structure
	dataPattern := `"data":\s*\{[^}]*"tickerStream"[^}]*\{[^}]*"content":\s*\[([^\]]+)\]`
	re := regexp.MustCompile(dataPattern)
	matches := re.FindStringSubmatch(html)

	if len(matches) < 2 {
		return nil, fmt.Errorf("no full JSON structure found")
	}

	// This is a simplified approach - in practice we'd need more sophisticated JSON parsing
	// For now, let's use a different pattern to find individual articles
	articlePattern := `\{[^{}]*"title":"([^"]*)"[^{}]*"pubDate":"([^"]*)"[^{}]*"provider":\s*\{[^{}]*"displayName":"([^"]*)"[^{}]*\}[^{}]*"canonicalUrl":\s*\{[^{}]*"url":"([^"]*)"[^{}]*\}[^{}]*(?:"thumbnail":\s*\{[^{}]*"originalUrl":"([^"]*)"[^{}]*\}[^{}]*)?(?:"stockTickers":\s*\[([^\]]*)\])?[^{}]*\}`

	articleRe := regexp.MustCompile(articlePattern)
	articleMatches := articleRe.FindAllStringSubmatch(html, -1)

	var articles []NewsItem
	for _, match := range articleMatches {
		if len(match) > 4 {
			article := &NewsItem{
				Title:  strings.TrimSpace(match[1]),
				URL:    strings.TrimSpace(match[4]),
				Source: strings.TrimSpace(match[3]),
			}

			// Parse publish date
			if match[2] != "" {
				if pubTime, err := time.Parse(time.RFC3339, match[2]); err == nil {
					utcTime := pubTime.UTC()
					article.PublishedAt = &utcTime
				}
			}

			// Parse image URL
			if len(match) > 5 && match[5] != "" {
				article.ImageURL = strings.TrimSpace(match[5])
			}

			// Parse tickers
			if len(match) > 6 && match[6] != "" {
				article.RelatedTickers = parseTickersFromJSON(match[6])
			}

			articles = append(articles, *article)
		}
	}

	return articles, nil
}

// parseJSONArticle parses a single JSON article object
func parseJSONArticle(jsonStr, baseURL string, now time.Time) *NewsItem {
	// This is a simplified JSON parser for the article structure
	// In practice, we'd use proper JSON unmarshaling

	titlePattern := `"title":"([^"]*)"`
	urlPattern := `"url":"([^"]*)"`
	sourcePattern := `"displayName":"([^"]*)"`
	datePattern := `"pubDate":"([^"]*)"`
	imagePattern := `"originalUrl":"([^"]*)"`

	article := &NewsItem{}

	// Extract title
	if re := regexp.MustCompile(titlePattern); re != nil {
		if matches := re.FindStringSubmatch(jsonStr); len(matches) > 1 {
			article.Title = strings.TrimSpace(matches[1])
		}
	}

	// Extract URL
	if re := regexp.MustCompile(urlPattern); re != nil {
		if matches := re.FindStringSubmatch(jsonStr); len(matches) > 1 {
			article.URL = strings.TrimSpace(matches[1])
		}
	}

	// Extract source
	if re := regexp.MustCompile(sourcePattern); re != nil {
		if matches := re.FindStringSubmatch(jsonStr); len(matches) > 1 {
			article.Source = strings.TrimSpace(matches[1])
		}
	}

	// Extract publish date
	if re := regexp.MustCompile(datePattern); re != nil {
		if matches := re.FindStringSubmatch(jsonStr); len(matches) > 1 {
			if pubTime, err := time.Parse(time.RFC3339, matches[1]); err == nil {
				utcTime := pubTime.UTC()
				article.PublishedAt = &utcTime
			}
		}
	}

	// Extract image
	if re := regexp.MustCompile(imagePattern); re != nil {
		if matches := re.FindStringSubmatch(jsonStr); len(matches) > 1 {
			article.ImageURL = strings.TrimSpace(matches[1])
		}
	}

	// Skip articles without title or URL
	if article.Title == "" || article.URL == "" {
		return nil
	}

	return article
}

// parseTickersFromJSON extracts ticker symbols from JSON ticker array
func parseTickersFromJSON(tickersJSON string) []string {
	tickerPattern := `"symbol":"([^"]*)"`
	re := regexp.MustCompile(tickerPattern)
	matches := re.FindAllStringSubmatch(tickersJSON, -1)

	var tickers []string
	tickerSet := make(map[string]bool)

	for _, match := range matches {
		if len(match) > 1 {
			ticker := strings.ToUpper(strings.TrimSpace(match[1]))
			if isValidTicker(ticker) && !tickerSet[ticker] {
				tickers = append(tickers, ticker)
				tickerSet[ticker] = true
			}
		}
	}

	return tickers
}

// extractArticlesFromNewsJSON extracts articles from the news JSON structure
func extractArticlesFromNewsJSON(bodyJSON, baseURL string, now time.Time) ([]NewsItem, error) {
	// Find blocks with contentType STORY directly to avoid brittle array parsing
	storyBlock := regexp.MustCompile(`\{"id":"[^"]*","content":\{[^}]*"contentType":"STORY"[^}]*\}`)
	blocks := storyBlock.FindAllString(bodyJSON, -1)

	var allArticles []NewsItem
	for _, blk := range blocks {
		// Extract core fields
		title := extractFirstGroup(blk, `"title":"([^"]*)"`)
		url := extractFirstGroup(blk, `"canonicalUrl":\{[^}]*"url":"([^"]*)"`)
		source := extractFirstGroup(blk, `"provider":\{[^}]*"displayName":"([^"]*)"`)
		pub := extractFirstGroup(blk, `"pubDate":"([^"]*)"`)
		img := extractFirstGroup(blk, `"originalUrl":"([^"]*)"`)
		tickRaw := extractFirstGroup(blk, `"stockTickers":\[([^\]]*)\]`)

		if title == "" || url == "" {
			continue
		}

		item := NewsItem{Title: strings.TrimSpace(title), URL: strings.TrimSpace(url), Source: strings.TrimSpace(source), ImageURL: strings.TrimSpace(img)}
		if pub != "" {
			if t, err := time.Parse(time.RFC3339, pub); err == nil {
				tt := t.UTC()
				item.PublishedAt = &tt
			}
		}
		if tickRaw != "" {
			item.RelatedTickers = parseTickersFromJSON(tickRaw)
		}
		allArticles = append(allArticles, item)
	}
	return allArticles, nil
}

// extractArticlesFromArray extracts articles from a single array
func extractArticlesFromArray(arrayContent, baseURL string, now time.Time) []NewsItem {
	var articles []NewsItem

	// Split the array content by objects (look for },"id": or },{"id":)
	objectSeparator := regexp.MustCompile(`\},\s*\{`)
	objectStrings := objectSeparator.Split(arrayContent, -1)

	for _, objStr := range objectStrings {
		// Clean up the object string
		objStr = strings.TrimSpace(objStr)
		if !strings.HasPrefix(objStr, "{") {
			objStr = "{" + objStr
		}
		if !strings.HasSuffix(objStr, "}") {
			objStr = objStr + "}"
		}

		// Skip if this doesn't look like a content object
		if !strings.Contains(objStr, `"content":{`) || !strings.Contains(objStr, `"contentType":"STORY"`) {
			continue
		}

		// Extract individual fields using simple patterns
		article := NewsItem{}

		// Extract title
		if titleRe := regexp.MustCompile(`"title":"([^"]*)"`); titleRe != nil {
			if matches := titleRe.FindStringSubmatch(objStr); len(matches) > 1 {
				article.Title = strings.TrimSpace(matches[1])
			}
		}

		// Extract pubDate
		if dateRe := regexp.MustCompile(`"pubDate":"([^"]*)"`); dateRe != nil {
			if matches := dateRe.FindStringSubmatch(objStr); len(matches) > 1 {
				if pubTime, err := time.Parse(time.RFC3339, matches[1]); err == nil {
					utcTime := pubTime.UTC()
					article.PublishedAt = &utcTime
				}
			}
		}

		// Extract provider displayName
		if providerRe := regexp.MustCompile(`"provider":\{[^}]*"displayName":"([^"]*)"`); providerRe != nil {
			if matches := providerRe.FindStringSubmatch(objStr); len(matches) > 1 {
				article.Source = strings.TrimSpace(matches[1])
			}
		}

		// Extract canonicalUrl
		if urlRe := regexp.MustCompile(`"canonicalUrl":\{[^}]*"url":"([^"]*)"`); urlRe != nil {
			if matches := urlRe.FindStringSubmatch(objStr); len(matches) > 1 {
				article.URL = strings.TrimSpace(matches[1])
			}
		}

		// Extract image originalUrl
		if imageRe := regexp.MustCompile(`"originalUrl":"([^"]*)"`); imageRe != nil {
			if matches := imageRe.FindStringSubmatch(objStr); len(matches) > 1 {
				article.ImageURL = strings.TrimSpace(matches[1])
			}
		}

		// Extract stockTickers
		if tickerRe := regexp.MustCompile(`"stockTickers":\[([^\]]*)\]`); tickerRe != nil {
			if matches := tickerRe.FindStringSubmatch(objStr); len(matches) > 1 {
				article.RelatedTickers = parseTickersFromJSON(matches[1])
			}
		}

		// Only add articles with required fields
		if article.Title != "" && article.URL != "" {
			articles = append(articles, article)
		}
	}

	return articles
}

// findObjectEnd finds the end of a JSON object starting from a position

// extractFirstGroup is a tiny helper to extract first capturing group
func extractFirstGroup(s, pattern string) string {
	re := regexp.MustCompile(pattern)
	m := re.FindStringSubmatch(s)
	if len(m) > 1 {
		return strings.TrimSpace(m[1])
	}
	return ""
}

// extractArticlesSimplePattern uses simpler regex patterns to extract articles
func extractArticlesSimplePattern(contentJSON, baseURL string, now time.Time) ([]NewsItem, error) {
	// Find all article content blocks
	contentBlockPattern := `"content":\{[^{}]*"contentType":"STORY"[^{}]*"title":"([^"]*)"[^{}]*"pubDate":"([^"]*)"[^{}]*"provider":\{[^{}]*"displayName":"([^"]*)"[^{}]*\}[^{}]*"canonicalUrl":\{[^{}]*"url":"([^"]*)"[^{}]*\}`

	blockRe := regexp.MustCompile(contentBlockPattern)
	blockMatches := blockRe.FindAllStringSubmatch(contentJSON, -1)

	var articles []NewsItem
	for _, match := range blockMatches {
		if len(match) > 4 {
			article := &NewsItem{
				Title:  strings.TrimSpace(match[1]),
				Source: strings.TrimSpace(match[3]),
				URL:    strings.TrimSpace(match[4]),
			}

			// Parse publish date
			if match[2] != "" {
				if pubTime, err := time.Parse(time.RFC3339, match[2]); err == nil {
					utcTime := pubTime.UTC()
					article.PublishedAt = &utcTime
				}
			}

			// Only add articles with required fields
			if article.Title != "" && article.URL != "" {
				articles = append(articles, *article)
			}
		}
	}

	return articles, nil
}

// parseNewsFromHTML falls back to HTML-based extraction for test fixtures
func parseNewsFromHTML(htmlStr, baseURL string, now time.Time, metrics *Metrics) ([]NewsItem, *NewsStats, error) {
	// Load regex configuration
	if err := LoadNewsRegexConfig(); err != nil {
		return nil, nil, fmt.Errorf("failed to load news regex config: %w", err)
	}

	// Extract article containers
	containers, err := extractArticleContainers(htmlStr)
	if err != nil {
		metrics.RecordNews("error")
		return nil, nil, fmt.Errorf("%w: %v", ErrNewsParse, err)
	}

	if len(containers) == 0 {
		metrics.RecordNews("no_articles")
		return nil, nil, ErrNewsNoArticles
	}

	// Parse articles from containers
	var articles []NewsItem
	for _, container := range containers {
		article := parseArticleFromContainer(container, baseURL, now)
		if article != nil {
			articles = append(articles, *article)
		}
	}

	// Enrich articles with source and published time from embedded JSON (if available)
	enrichArticlesWithJSONMeta(htmlStr, articles)

	if len(articles) == 0 {
		metrics.RecordNews("no_valid_articles")
		return nil, nil, ErrNewsNoArticles
	}

	// Deduplicate articles
	originalCount := len(articles)
	articles = deduplicateArticles(articles)
	deduped := originalCount - len(articles)

	// Limit results (default 25 articles)
	const maxArticles = 25
	if len(articles) > maxArticles {
		articles = articles[:maxArticles]
	}

	// Extract pagination hint
	nextPageHint := extractNextPageHint(htmlStr)

	// Create statistics
	stats := &NewsStats{
		TotalFound:    originalCount,
		TotalReturned: len(articles),
		Deduped:       deduped,
		NextPageHint:  nextPageHint,
		AsOf:          now.UTC(),
	}

	metrics.RecordNews("success")
	return articles, stats, nil
}

// enrichArticlesWithJSONMeta builds a title->(source,time) map once and enriches articles in place
func enrichArticlesWithJSONMeta(fullHTML string, articles []NewsItem) {
	if len(articles) == 0 {
		return
	}
	scriptPattern := `<script[^>]*>([^<]*tickerStream[^<]*)</script>`
	scriptRe := regexp.MustCompile(scriptPattern)
	scriptMatches := scriptRe.FindStringSubmatch(fullHTML)
	if len(scriptMatches) < 2 {
		return
	}
	// Unescape JSON body
	bodyPattern := `"body":"(\{.*?\})"`
	bodyRe := regexp.MustCompile(bodyPattern)
	bodyMatches := bodyRe.FindStringSubmatch(scriptMatches[1])
	if len(bodyMatches) < 2 {
		return
	}
	var jsonBody string
	if err := json.Unmarshal([]byte("\""+bodyMatches[1]+"\""), &jsonBody); err != nil {
		jsonBody = bodyMatches[1]
	}

	// Build a map from normalized title to (source, pubDate)
	meta := make(map[string]struct {
		src string
		t   *time.Time
	})

	storyBlock := regexp.MustCompile(`\{"id":"[^"]*","content":\{[^}]*"contentType":"STORY"[^}]*\}`)
	blocks := storyBlock.FindAllString(jsonBody, -1)
	for _, blk := range blocks {
		title := extractFirstGroup(blk, `"title":"([^"]*)"`)
		if title == "" {
			continue
		}
		src := extractFirstGroup(blk, `"provider":\{[^}]*"displayName":"([^"]*)"`)
		pub := extractFirstGroup(blk, `"pubDate":"([^"]*)"`)
		var pt *time.Time
		if pub != "" {
			if t, err := time.Parse(time.RFC3339, pub); err == nil {
				tt := t.UTC()
				pt = &tt
			}
		}
		key := strings.ToLower(strings.TrimSpace(title))
		meta[key] = struct {
			src string
			t   *time.Time
		}{src: strings.TrimSpace(src), t: pt}
	}

	// Enrich
	for i := range articles {
		key := strings.ToLower(strings.TrimSpace(articles[i].Title))
		if m, ok := meta[key]; ok {
			if articles[i].Source == "" && m.src != "" {
				articles[i].Source = m.src
			}
			if articles[i].PublishedAt == nil && m.t != nil {
				articles[i].PublishedAt = m.t
			}
		}
	}
}

// enhanceArticleWithJSON attempts to fill missing fields from JSON data in the full HTML
