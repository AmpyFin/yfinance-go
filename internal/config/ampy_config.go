package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/AmpyFin/ampy-config/go/ampyconfig"
	"gopkg.in/yaml.v3"
)

// Config represents the complete configuration for yfinance-go
type Config struct {
	App           AppConfig           `yaml:"app"`
	Yahoo         YahooConfig         `yaml:"yahoo"`
	Concurrency   ConcurrencyConfig   `yaml:"concurrency"`
	RateLimit     RateLimitConfig     `yaml:"rate_limit"`
	Sessions      SessionsConfig      `yaml:"sessions"`
	Retry         RetryConfig         `yaml:"retry"`
	CircuitBreaker CircuitBreakerConfig `yaml:"circuit_breaker"`
	Markets       MarketsConfig       `yaml:"markets"`
	FX            FXConfig            `yaml:"fx"`
	Bus           BusConfig           `yaml:"bus"`
	Scrape        ScrapeConfig        `yaml:"scrape"`
	Observability ObservabilityConfig `yaml:"observability"`
	Secrets       []SecretConfig      `yaml:"secrets"`
}

// AppConfig represents application-level configuration
type AppConfig struct {
	Env   string `yaml:"env"`
	RunID string `yaml:"run_id"`
}

// YahooConfig represents Yahoo Finance API configuration
type YahooConfig struct {
	BaseURL         string        `yaml:"base_url"`
	TimeoutMs       int           `yaml:"timeout_ms"`
	IdleTimeoutMs   int           `yaml:"idle_timeout_ms"`
	MaxConnsPerHost int           `yaml:"max_conns_per_host"`
	UserAgent       string        `yaml:"user_agent"`
}

// ConcurrencyConfig represents concurrency configuration
type ConcurrencyConfig struct {
	GlobalWorkers   int `yaml:"global_workers"`
	PerHostWorkers  int `yaml:"per_host_workers"`
}

// RateLimitConfig represents rate limiting configuration
type RateLimitConfig struct {
	PerHostQPS      float64 `yaml:"per_host_qps"`
	PerHostBurst    int     `yaml:"per_host_burst"`
	PerSessionQPS   float64 `yaml:"per_session_qps"`
	PerSessionBurst int     `yaml:"per_session_burst"`
}

// SessionsConfig represents session rotation configuration
type SessionsConfig struct {
	N                   int `yaml:"n"`
	EjectAfter          int `yaml:"eject_after"`
	RecreateCooldownMs  int `yaml:"recreate_cooldown_ms"`
}

// RetryConfig represents retry configuration
type RetryConfig struct {
	Attempts     int `yaml:"attempts"`
	BaseMs       int `yaml:"base_ms"`
	MaxDelayMs   int `yaml:"max_delay_ms"`
}

// CircuitBreakerConfig represents circuit breaker configuration
type CircuitBreakerConfig struct {
	Window            int     `yaml:"window"`
	FailureThreshold  float64 `yaml:"failure_threshold"`
	ResetTimeoutMs    int     `yaml:"reset_timeout_ms"`
	HalfOpenProbes    int     `yaml:"half_open_probes"`
}

// MarketsConfig represents market configuration
type MarketsConfig struct {
	AllowedIntervals         []string `yaml:"allowed_intervals"`
	AllowedMics              []string `yaml:"allowed_mics"`
	DefaultAdjustmentPolicy  string   `yaml:"default_adjustment_policy"`
}

// FXConfig represents FX configuration
type FXConfig struct {
	Provider  string         `yaml:"provider"`
	Target    string         `yaml:"target"`
	CacheTTLMs int           `yaml:"cache_ttl_ms"`
	RateScale int            `yaml:"rate_scale"`
	Rounding  string         `yaml:"rounding"`
	YahooWeb  YahooWebConfig `yaml:"yahoo_web"`
}

// YahooWebConfig represents Yahoo Web FX provider configuration
type YahooWebConfig struct {
	QPS                float64 `yaml:"qps"`
	Burst              int     `yaml:"burst"`
	TimeoutMs          int     `yaml:"timeout_ms"`
	BackoffAttempts    int     `yaml:"backoff_attempts"`
	BackoffBaseMs      int     `yaml:"backoff_base_ms"`
	BackoffMaxDelayMs  int     `yaml:"backoff_max_delay_ms"`
	CircuitResetMs     int     `yaml:"circuit_reset_ms"`
}

// BusConfig represents bus configuration
type BusConfig struct {
	Enabled         bool              `yaml:"enabled"`
	Env             string            `yaml:"env"`
	TopicPrefix     string            `yaml:"topic_prefix"`
	MaxPayloadBytes int64             `yaml:"max_payload_bytes"`
	Publisher       PublisherConfig   `yaml:"publisher"`
	Retry           RetryConfig       `yaml:"retry"`
	CircuitBreaker  CircuitBreakerConfig `yaml:"circuit_breaker"`
}

// ScrapeConfig represents scraping configuration
type ScrapeConfig struct {
	Enabled      bool                    `yaml:"enabled"`
	UserAgent    string                  `yaml:"user_agent"`
	TimeoutMs    int                     `yaml:"timeout_ms"`
	QPS          float64                 `yaml:"qps"`
	Burst        int                     `yaml:"burst"`
	Retry        ScrapeRetryConfig       `yaml:"retry"`
	RobotsPolicy string                  `yaml:"robots_policy"`
	CacheTTLMs   int                     `yaml:"cache_ttl_ms"`
	Endpoints    ScrapeEndpointConfig    `yaml:"endpoints"`
}

// ScrapeRetryConfig represents scraping retry configuration
type ScrapeRetryConfig struct {
	Attempts   int `yaml:"attempts"`
	BaseMs     int `yaml:"base_ms"`
	MaxDelayMs int `yaml:"max_delay_ms"`
}

// ScrapeEndpointConfig represents endpoint-specific scraping configuration
type ScrapeEndpointConfig struct {
	KeyStatistics bool `yaml:"key_statistics"`
	Financials    bool `yaml:"financials"`
	Analysis      bool `yaml:"analysis"`
	Profile       bool `yaml:"profile"`
	News          bool `yaml:"news"`
}

// PublisherConfig represents publisher configuration
type PublisherConfig struct {
	Backend string      `yaml:"backend"`
	NATS    NATSConfig  `yaml:"nats"`
	Kafka   KafkaConfig `yaml:"kafka"`
}

// NATSConfig represents NATS configuration
type NATSConfig struct {
	URL          string `yaml:"url"`
	SubjectStyle string `yaml:"subject_style"`
	AckWaitMs    int    `yaml:"ack_wait_ms"`
}

// KafkaConfig represents Kafka configuration
type KafkaConfig struct {
	Brokers     []string `yaml:"brokers"`
	Acks        string   `yaml:"acks"`
	Compression string   `yaml:"compression"`
}

// ObservabilityConfig represents observability configuration
type ObservabilityConfig struct {
	Logs      LogsConfig      `yaml:"logs"`
	Metrics   MetricsConfig   `yaml:"metrics"`
	Tracing   TracingConfig   `yaml:"tracing"`
}

// LogsConfig represents logging configuration
type LogsConfig struct {
	Level string `yaml:"level"`
}

// MetricsConfig represents metrics configuration
type MetricsConfig struct {
	Prometheus PrometheusConfig `yaml:"prometheus"`
}

// PrometheusConfig represents Prometheus configuration
type PrometheusConfig struct {
	Enabled bool   `yaml:"enabled"`
	Addr    string `yaml:"addr"`
}

// TracingConfig represents tracing configuration
type TracingConfig struct {
	OTLP OTLPConfig `yaml:"otlp"`
}

// OTLPConfig represents OTLP configuration
type OTLPConfig struct {
	Enabled      bool    `yaml:"enabled"`
	Endpoint     string  `yaml:"endpoint"`
	SampleRatio  float64 `yaml:"sample_ratio"`
}

// SecretConfig represents secret configuration
type SecretConfig struct {
	Name     string `yaml:"name"`
	Ref      string `yaml:"ref"`
	Required bool   `yaml:"required"`
}

// Loader handles configuration loading using ampy-config
type Loader struct {
	effectivePath string
	config        *Config
}

// NewLoader creates a new configuration loader using ampy-config
func NewLoader(effectivePath string) *Loader {
	return &Loader{
		effectivePath: effectivePath,
	}
}

// Load loads and validates configuration from the effective YAML file
func (l *Loader) Load() (*Config, error) {
	// Use ampy-config Loader to read the effective YAML
	ampyLoader := ampyconfig.NewLoader(l.effectivePath)
	configMap, err := ampyLoader.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load effective config: %w", err)
	}
	
	// Interpolate environment variables
	l.interpolateEnvVars(configMap)
	
	// Convert map to our Config struct
	config, err := l.mapToConfig(configMap)
	if err != nil {
		return nil, fmt.Errorf("failed to convert config: %w", err)
	}
	
	// Validate configuration
	if err := l.validate(config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}
	
	l.config = config
	return config, nil
}

// interpolateEnvVars interpolates environment variables in the configuration map
func (l *Loader) interpolateEnvVars(configMap map[string]interface{}) {
	for key, value := range configMap {
		if str, ok := value.(string); ok {
			// Handle ${VAR} and ${VAR:-default} syntax
			configMap[key] = l.interpolateString(str)
		} else if nestedMap, ok := value.(map[string]interface{}); ok {
			// Recursively process nested maps
			l.interpolateEnvVars(nestedMap)
		} else if slice, ok := value.([]interface{}); ok {
			// Process slices
			for i, item := range slice {
				if str, ok := item.(string); ok {
					slice[i] = l.interpolateString(str)
				} else if nestedMap, ok := item.(map[string]interface{}); ok {
					l.interpolateEnvVars(nestedMap)
				}
			}
		}
	}
}

// interpolateString interpolates environment variables in a string
func (l *Loader) interpolateString(str string) string {
	// Handle ${VAR} and ${VAR:-default} syntax
	result := str
	for {
		start := strings.Index(result, "${")
		if start == -1 {
			break
		}
		
		end := strings.Index(result[start:], "}")
		if end == -1 {
			break
		}
		
		end += start
		varExpr := result[start+2 : end]
		
		var value string
		if strings.Contains(varExpr, ":-") {
			// Handle ${VAR:-default} syntax
			parts := strings.SplitN(varExpr, ":-", 2)
			envVar := parts[0]
			defaultValue := parts[1]
			value = os.Getenv(envVar)
			if value == "" {
				value = defaultValue
			}
		} else {
			// Handle ${VAR} syntax
			value = os.Getenv(varExpr)
		}
		
		result = result[:start] + value + result[end+1:]
	}
	
	return result
}

// mapToConfig converts a map to our Config struct
func (l *Loader) mapToConfig(configMap map[string]interface{}) (*Config, error) {
	// Marshal to YAML and unmarshal to struct
	data, err := yaml.Marshal(configMap)
	if err != nil {
		return nil, err
	}
	
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	
	return &config, nil
}

// validate validates the configuration
func (l *Loader) validate(config *Config) error {
	// Validate app.env
	if config.App.Env != "dev" && config.App.Env != "staging" && config.App.Env != "prod" {
		// Allow custom env but warn
		// In a real implementation, you might want to log a warning
		_ = config.App.Env // Suppress unused variable warning
	}
	
	// Validate concurrency constraints
	if config.Concurrency.GlobalWorkers < config.Concurrency.PerHostWorkers {
		return fmt.Errorf("concurrency.global_workers (%d) must be >= per_host_workers (%d)", 
			config.Concurrency.GlobalWorkers, config.Concurrency.PerHostWorkers)
	}
	
	if config.Concurrency.PerHostWorkers < config.Sessions.N {
		return fmt.Errorf("concurrency.per_host_workers (%d) must be >= sessions.n (%d)", 
			config.Concurrency.PerHostWorkers, config.Sessions.N)
	}
	
	// Validate rate limit constraints
	if config.RateLimit.PerSessionQPS*float64(config.Sessions.N) > config.RateLimit.PerHostQPS {
		// This is a warning, not an error
		// In a real implementation, you might want to log a warning
		_ = config.RateLimit.PerSessionQPS // Suppress unused variable warning
	}
	
	// Validate markets.allowed_intervals (daily-only enforcement)
	if len(config.Markets.AllowedIntervals) != 1 || config.Markets.AllowedIntervals[0] != "1d" {
		return fmt.Errorf("markets.allowed_intervals must be exactly [\"1d\"] for yfinance-go (daily-only scope)")
	}
	
	// Validate markets.default_adjustment_policy
	if config.Markets.DefaultAdjustmentPolicy != "raw" && config.Markets.DefaultAdjustmentPolicy != "split_dividend" {
		return fmt.Errorf("markets.default_adjustment_policy must be 'raw' or 'split_dividend'")
	}
	
	// Validate bus.max_payload_bytes
	if config.Bus.MaxPayloadBytes < 262144 || config.Bus.MaxPayloadBytes > 10485760 {
		return fmt.Errorf("bus.max_payload_bytes must be between 262144 and 10485760")
	}
	
	// Validate retry.attempts
	if config.Retry.Attempts < 1 {
		return fmt.Errorf("retry.attempts must be >= 1")
	}
	
	// Validate circuit breaker thresholds
	if config.CircuitBreaker.FailureThreshold <= 0 || config.CircuitBreaker.FailureThreshold > 1 {
		return fmt.Errorf("circuit_breaker.failure_threshold must be between 0 and 1")
	}
	
	// Validate bus configuration if enabled
	if config.Bus.Enabled {
		if config.Bus.Publisher.Backend == "nats" && config.Bus.Publisher.NATS.URL == "" {
			return fmt.Errorf("bus.publisher.nats.url is required when bus.enabled=true and backend=nats")
		}
	}
	
	// Validate observability configuration
	if config.Observability.Metrics.Prometheus.Enabled && config.Observability.Metrics.Prometheus.Addr == "" {
		return fmt.Errorf("observability.metrics.prometheus.addr is required when prometheus is enabled")
	}
	
	if config.Observability.Tracing.OTLP.Enabled && config.Observability.Tracing.OTLP.Endpoint == "" {
		return fmt.Errorf("observability.tracing.otlp.endpoint is required when OTLP tracing is enabled")
	}
	
	return nil
}

// GetEffectiveConfig returns the effective configuration as a map for printing
func (l *Loader) GetEffectiveConfig() (map[string]interface{}, error) {
	if l.config == nil {
		return nil, fmt.Errorf("configuration not loaded")
	}
	
	// Use ampy-config Loader to get the raw effective config
	ampyLoader := ampyconfig.NewLoader(l.effectivePath)
	configMap, err := ampyLoader.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load effective config: %w", err)
	}
	
	// Interpolate environment variables
	l.interpolateEnvVars(configMap)
	
	// Redact secrets
	l.redactSecrets(configMap)
	
	return configMap, nil
}

// redactSecrets redacts secret values in the configuration map
func (l *Loader) redactSecrets(configMap map[string]interface{}) {
	// Redact secrets section
	if secrets, ok := configMap["secrets"].([]interface{}); ok {
		for i := range secrets {
			if secretMap, ok := secrets[i].(map[string]interface{}); ok {
				if _, ok := secretMap["ref"].(string); ok {
					secretMap["ref"] = "[REDACTED]"
				}
			}
		}
	}
	
	// Redact known secret patterns
	l.redactSecretPatterns(configMap)
}

// redactSecretPatterns redacts values that match secret patterns
func (l *Loader) redactSecretPatterns(configMap map[string]interface{}) {
	secretPatterns := []string{"password", "token", "api_key", "secret", "key"}
	
	for key, value := range configMap {
		keyLower := strings.ToLower(key)
		
		// Skip the secrets array itself - it's handled separately
		if key == "secrets" {
			continue
		}
		
		// Check if key matches secret patterns
		for _, pattern := range secretPatterns {
			if strings.Contains(keyLower, pattern) {
				configMap[key] = "[REDACTED]"
				continue
			}
		}
		
		// Recursively process nested maps
		if nestedMap, ok := value.(map[string]interface{}); ok {
			l.redactSecretPatterns(nestedMap)
		}
	}
}

// GetHTTPConfig converts the configuration to httpx.Config
func (c *Config) GetHTTPConfig() *HTTPConfig {
	return &HTTPConfig{
		BaseURL:            c.Yahoo.BaseURL,
		Timeout:            time.Duration(c.Yahoo.TimeoutMs) * time.Millisecond,
		IdleTimeout:        time.Duration(c.Yahoo.IdleTimeoutMs) * time.Millisecond,
		MaxConnsPerHost:    c.Yahoo.MaxConnsPerHost,
		UserAgent:          c.Yahoo.UserAgent,
		MaxAttempts:        c.Retry.Attempts,
		BackoffBaseMs:      c.Retry.BaseMs,
		BackoffJitterMs:    c.Retry.BaseMs / 2, // Default jitter
		MaxDelayMs:         c.Retry.MaxDelayMs,
		QPS:                c.RateLimit.PerHostQPS,
		Burst:              c.RateLimit.PerHostBurst,
		CircuitWindow:      time.Duration(c.CircuitBreaker.Window) * time.Second,
		FailureThreshold:   c.CircuitBreaker.FailureThreshold,
		ResetTimeout:       time.Duration(c.CircuitBreaker.ResetTimeoutMs) * time.Millisecond,
		EnableSessionRotation: true,
		NumSessions:        c.Sessions.N,
	}
}

// HTTPConfig represents HTTP client configuration (compatible with httpx.Config)
type HTTPConfig struct {
	BaseURL            string
	Timeout            time.Duration
	IdleTimeout        time.Duration
	MaxConnsPerHost    int
	UserAgent          string
	MaxAttempts        int
	BackoffBaseMs      int
	BackoffJitterMs    int
	MaxDelayMs         int
	QPS                float64
	Burst              int
	CircuitWindow      time.Duration
	FailureThreshold   float64
	ResetTimeout       time.Duration
	EnableSessionRotation bool
	NumSessions        int
}

// GetBusConfig converts the configuration to bus.Config
func (c *Config) GetBusConfig() *BusConfig {
	return &c.Bus
}

// GetFXConfig converts the configuration to fx.Config
func (c *Config) GetFXConfig() *FXConfig {
	return &c.FX
}

// GetScrapeConfig converts the configuration to scrape.Config
func (c *Config) GetScrapeConfig() *ScrapeConfig {
	return &c.Scrape
}

// ValidateInterval validates that the interval is allowed
func (c *Config) ValidateInterval(interval string) error {
	for _, allowed := range c.Markets.AllowedIntervals {
		if interval == allowed {
			return nil
		}
	}
	return fmt.Errorf("interval '%s' is not allowed. Allowed intervals: %v", interval, c.Markets.AllowedIntervals)
}

// ValidateAdjustmentPolicy validates that the adjustment policy is allowed
func (c *Config) ValidateAdjustmentPolicy(policy string) error {
	if policy == "raw" || policy == "split_dividend" {
		return nil
	}
	return fmt.Errorf("adjustment policy '%s' is not allowed. Allowed policies: raw, split_dividend", policy)
}

// CreateEffectiveConfig creates a default effective config file for testing
func CreateEffectiveConfig(path string) error {
	// Create a default effective config
	defaultConfig := map[string]interface{}{
		"app": map[string]interface{}{
			"env":    "dev",
			"run_id": "",
		},
		"yahoo": map[string]interface{}{
			"base_url":          "https://query2.finance.yahoo.com",
			"timeout_ms":        6000,
			"idle_timeout_ms":   30000,
			"max_conns_per_host": 64,
			"user_agent":        "AmpyFin-yfinance-go/1.x",
		},
		"concurrency": map[string]interface{}{
			"global_workers":  64,
			"per_host_workers": 32,
		},
		"rate_limit": map[string]interface{}{
			"per_host_qps":      5.0,
			"per_host_burst":    5,
			"per_session_qps":   1.0,
			"per_session_burst": 1,
		},
		"sessions": map[string]interface{}{
			"n":                   7,
			"eject_after":         5,
			"recreate_cooldown_ms": 15000,
		},
		"retry": map[string]interface{}{
			"attempts":     5,
			"base_ms":      250,
			"max_delay_ms": 8000,
		},
		"circuit_breaker": map[string]interface{}{
			"window":            50,
			"failure_threshold": 0.30,
			"reset_timeout_ms":  30000,
			"half_open_probes":  3,
		},
		"markets": map[string]interface{}{
			"allowed_intervals":        []string{"1d"},
			"allowed_mics":             []string{"XNAS", "XNYS", "XNMS", "NYQ", "KSC", "XETR", "XTKS"},
			"default_adjustment_policy": "split_dividend",
		},
		"fx": map[string]interface{}{
			"provider":     "none",
			"target":       "",
			"cache_ttl_ms": 60000,
			"rate_scale":   8,
			"rounding":     "half_up",
			"yahoo_web": map[string]interface{}{
				"qps":                0.5,
				"burst":              1,
				"timeout_ms":         5000,
				"backoff_attempts":   4,
				"backoff_base_ms":    250,
				"backoff_max_delay_ms": 4000,
				"circuit_reset_ms":   30000,
			},
		},
		"bus": map[string]interface{}{
			"enabled":           false,
			"env":               "dev",
			"topic_prefix":      "ampy",
			"max_payload_bytes": 1048576,
			"publisher": map[string]interface{}{
				"backend": "nats",
				"nats": map[string]interface{}{
					"url":           "nats://localhost:4222",
					"subject_style": "topic",
					"ack_wait_ms":   5000,
				},
				"kafka": map[string]interface{}{
					"brokers":     []string{},
					"acks":        "all",
					"compression": "snappy",
				},
			},
			"retry": map[string]interface{}{
				"attempts":     5,
				"base_ms":      250,
				"max_delay_ms": 8000,
			},
			"circuit_breaker": map[string]interface{}{
				"window":            50,
				"failure_threshold": 0.30,
				"reset_timeout_ms":  30000,
				"half_open_probes":  3,
			},
		},
		"scrape": map[string]interface{}{
			"enabled":       true,
			"user_agent":    "Mozilla/5.0 (Ampy yfinance-go scraper)",
			"timeout_ms":    10000,
			"qps":           0.7,
			"burst":         1,
			"retry": map[string]interface{}{
				"attempts":     4,
				"base_ms":      300,
				"max_delay_ms": 4000,
			},
			"robots_policy": "enforce",
			"cache_ttl_ms":  60000,
			"endpoints": map[string]interface{}{
				"key_statistics": true,
				"financials":     true,
				"analysis":       true,
				"profile":        true,
				"news":           true,
			},
		},
		"observability": map[string]interface{}{
			"logs": map[string]interface{}{
				"level": "info",
			},
			"metrics": map[string]interface{}{
				"prometheus": map[string]interface{}{
					"enabled": true,
					"addr":    ":9090",
				},
			},
			"tracing": map[string]interface{}{
				"otlp": map[string]interface{}{
					"enabled":      true,
					"endpoint":     "http://localhost:4317",
					"sample_ratio": 0.05,
				},
			},
		},
		"secrets": []interface{}{},
	}
	
	// Marshal to YAML
	data, err := yaml.Marshal(defaultConfig)
	if err != nil {
		return err
	}
	
	// Write to file
	return os.WriteFile(path, data, 0644)
}
