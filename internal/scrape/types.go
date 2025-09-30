package scrape

import (
	"time"
)

// FetchMeta contains metadata about a fetch operation
type FetchMeta struct {
	URL          string        `json:"url"`
	Host         string        `json:"host"`
	Status       int           `json:"status"`
	Attempt      int           `json:"attempt"`
	Bytes        int           `json:"bytes"`
	Gzip         bool          `json:"gzip"`
	Redirects    int           `json:"redirects"`
	Duration     time.Duration `json:"duration"`
	FromCache    bool          `json:"from_cache"` // reserved for optional HTML in-run cache
	RobotsPolicy string        `json:"robots_policy"`
}

// Config represents the scraping configuration
type Config struct {
	Enabled      bool           `yaml:"enabled"`
	UserAgent    string         `yaml:"user_agent"`
	TimeoutMs    int            `yaml:"timeout_ms"`
	QPS          float64        `yaml:"qps"`
	Burst        int            `yaml:"burst"`
	Retry        RetryConfig    `yaml:"retry"`
	RobotsPolicy string         `yaml:"robots_policy"`
	CacheTTLMs   int            `yaml:"cache_ttl_ms"`
	Endpoints    EndpointConfig `yaml:"endpoints"`
}

// RetryConfig represents retry configuration
type RetryConfig struct {
	Attempts   int `yaml:"attempts"`
	BaseMs     int `yaml:"base_ms"`
	MaxDelayMs int `yaml:"max_delay_ms"`
}

// EndpointConfig represents endpoint-specific configuration
type EndpointConfig struct {
	KeyStatistics bool `yaml:"key_statistics"`
	Financials    bool `yaml:"financials"`
	Analysis      bool `yaml:"analysis"`
	Profile       bool `yaml:"profile"`
	News          bool `yaml:"news"`
}

// DefaultConfig returns a sensible default configuration
func DefaultConfig() *Config {
	return &Config{
		Enabled:   true,
		UserAgent: "Mozilla/5.0 (Ampy yfinance-go scraper)",
		TimeoutMs: 10000,
		QPS:       0.7,
		Burst:     1,
		Retry: RetryConfig{
			Attempts:   4,
			BaseMs:     300,
			MaxDelayMs: 4000,
		},
		RobotsPolicy: "enforce",
		CacheTTLMs:   60000,
		Endpoints: EndpointConfig{
			KeyStatistics: true,
			Financials:    true,
			Analysis:      true,
			Profile:       true,
			News:          true,
		},
	}
}

// RobotsPolicy represents the robots.txt policy
type RobotsPolicy string

const (
	RobotsEnforce RobotsPolicy = "enforce"
	RobotsWarn    RobotsPolicy = "warn"
	RobotsIgnore  RobotsPolicy = "ignore"
)

// IsValidRobotsPolicy checks if a robots policy is valid
func IsValidRobotsPolicy(policy string) bool {
	return policy == string(RobotsEnforce) ||
		policy == string(RobotsWarn) ||
		policy == string(RobotsIgnore)
}

// RobotsRule represents a robots.txt rule
type RobotsRule struct {
	UserAgent string
	Allow     []string
	Disallow  []string
}

// RobotsCache represents cached robots.txt data
type RobotsCache struct {
	Host      string
	Rules     []RobotsRule
	FetchedAt time.Time
	TTL       time.Duration
}

// IsExpired checks if the robots cache is expired
func (rc *RobotsCache) IsExpired() bool {
	return time.Since(rc.FetchedAt) > rc.TTL
}

// BackoffPolicyConfig represents backoff configuration
type BackoffPolicyConfig struct {
	BaseDelay    time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
	JitterFactor float64
}

// DefaultBackoffPolicyConfig returns a sensible default backoff policy
func DefaultBackoffPolicyConfig() *BackoffPolicyConfig {
	return &BackoffPolicyConfig{
		BaseDelay:    300 * time.Millisecond,
		MaxDelay:     4 * time.Second,
		Multiplier:   2.0,
		JitterFactor: 0.2, // ±20% jitter
	}
}

// RateLimitConfig represents rate limiting configuration
type RateLimitConfig struct {
	QPS            float64
	Burst          int
	PerHostWorkers int
}

// DefaultRateLimitConfig returns a sensible default rate limit configuration
func DefaultRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		QPS:            0.7,
		Burst:          1,
		PerHostWorkers: 4,
	}
}

// NewsItem represents a single news article extracted from Yahoo Finance
type NewsItem struct {
	Title          string     `json:"title"`
	URL            string     `json:"url"` // absolute; normalized
	Source         string     `json:"source"`
	PublishedAt    *time.Time `json:"published_at"` // UTC if resolvable
	ImageURL       string     `json:"image_url"`
	RelatedTickers []string   `json:"related_tickers"`
}

// NewsStats represents statistics about news extraction
type NewsStats struct {
	TotalFound    int       `json:"total_found"`
	TotalReturned int       `json:"total_returned"`
	Deduped       int       `json:"deduped"`
	NextPageHint  string    `json:"next_page_hint"` // e.g., a data-cursor or bool flag if detected
	AsOf          time.Time `json:"as_of"`
}
