package fx

import (
	"context"
	"time"

	"github.com/AmpyFin/yfinance-go/internal/norm"
)

// FX interface for currency conversion providers
type FX interface {
	Rates(ctx context.Context, base string, symbols []string, at time.Time) (map[string]norm.ScaledDecimal, time.Time, error)
}

// FXMeta contains metadata about FX conversion
type FXMeta struct {
	Provider       string    `json:"provider"`        // "none" | "yahoo-web"
	Base           string    `json:"base"`            // e.g., "EUR"
	Symbols        []string  `json:"symbols"`         // e.g., ["USD"]
	AsOf           time.Time `json:"as_of"`           // timestamp of FX rates
	RateScale      int       `json:"rate_scale"`      // scale of the rate decimals (e.g., 8)
	CacheHit       bool      `json:"cache_hit"`       // whether this was a cache hit
	Attempts       int       `json:"attempts"`        // number of attempts made
	BackoffProfile string    `json:"backoff_profile"` // backoff profile used
	Stale          bool      `json:"stale"`           // whether rates are stale
}

// Config contains FX configuration
type Config struct {
	Provider  string         `yaml:"provider"`     // "none" (default) or "yahoo-web"
	Target    string         `yaml:"target"`       // e.g., "USD" (optional for CLI previews)
	CacheTTL  time.Duration  `yaml:"cache_ttl_ms"` // cache TTL in milliseconds
	RateScale int            `yaml:"rate_scale"`   // scale for FX rates (default 8)
	Rounding  string         `yaml:"rounding"`     // rounding mode (fixed to "half_up")
	YahooWeb  YahooWebConfig `yaml:"yahoo_web"`    // yahoo-web provider config
}

// YahooWebConfig contains configuration for the yahoo-web provider
type YahooWebConfig struct {
	QPS             float64       `yaml:"qps"`                  // queries per second
	Burst           int           `yaml:"burst"`                // burst limit
	Timeout         time.Duration `yaml:"timeout_ms"`           // timeout in milliseconds
	BackoffAttempts int           `yaml:"backoff_attempts"`     // number of backoff attempts
	BackoffBase     time.Duration `yaml:"backoff_base_ms"`      // base backoff delay in milliseconds
	BackoffMaxDelay time.Duration `yaml:"backoff_max_delay_ms"` // max backoff delay in milliseconds
	CircuitReset    time.Duration `yaml:"circuit_reset_ms"`     // circuit breaker reset time in milliseconds
}

// DefaultConfig returns the default FX configuration
func DefaultConfig() *Config {
	return &Config{
		Provider:  "none",
		Target:    "",
		CacheTTL:  60 * time.Second,
		RateScale: 8,
		Rounding:  "half_up",
		YahooWeb: YahooWebConfig{
			QPS:             0.5,
			Burst:           1,
			Timeout:         5 * time.Second,
			BackoffAttempts: 4,
			BackoffBase:     250 * time.Millisecond,
			BackoffMaxDelay: 4 * time.Second,
			CircuitReset:    30 * time.Second,
		},
	}
}
