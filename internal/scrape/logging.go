package scrape

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

// Logger handles structured logging for scraping operations
type Logger struct {
	logger *log.Logger
}

// NewLogger creates a new logger instance
func NewLogger() *Logger {
	return &Logger{
		logger: log.New(os.Stderr, "", 0), // No prefix, we'll add our own
	}
}

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Source    string                 `json:"source"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

// LogRequest logs a scraping request
func (l *Logger) LogRequest(url, host string, status, attempt int, duration time.Duration, bytes int, gzip bool, redirects int, errorMsg string) {
	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     "info",
		Source:    "yfinance-go/scrape",
		Message:   "scrape request",
		Fields: map[string]interface{}{
			"url":         url,
			"host":        host,
			"status":      status,
			"attempt":     attempt,
			"duration_ms": duration.Milliseconds(),
			"bytes":       bytes,
			"gzip":        gzip,
			"redirects":   redirects,
		},
	}

	if errorMsg != "" {
		entry.Level = "error"
		entry.Message = "scrape request failed"
		entry.Fields["error"] = errorMsg
	}

	l.logStructured(entry)
}

// LogRetry logs a retry event
func (l *Logger) LogRetry(url, host string, attempt int, reason, errorMsg string) {
	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     "warn",
		Source:    "yfinance-go/scrape",
		Message:   "scrape retry",
		Fields: map[string]interface{}{
			"url":     url,
			"host":    host,
			"attempt": attempt,
			"reason":  reason,
			"error":   errorMsg,
		},
	}

	l.logStructured(entry)
}

// LogBackoff logs a backoff event
func (l *Logger) LogBackoff(url, host string, delay time.Duration) {
	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     "info",
		Source:    "yfinance-go/scrape",
		Message:   "scrape backoff",
		Fields: map[string]interface{}{
			"url":      url,
			"host":     host,
			"sleep_ms": delay.Milliseconds(),
		},
	}

	l.logStructured(entry)
}

// LogRobotsDenied logs a robots.txt denial
func (l *Logger) LogRobotsDenied(url, host, errorMsg string) {
	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     "warn",
		Source:    "yfinance-go/scrape",
		Message:   "robots.txt denied",
		Fields: map[string]interface{}{
			"url":   url,
			"host":  host,
			"error": errorMsg,
		},
	}

	l.logStructured(entry)
}

// LogRateLimit logs a rate limit event
func (l *Logger) LogRateLimit(url, host, errorMsg string) {
	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     "warn",
		Source:    "yfinance-go/scrape",
		Message:   "rate limit exceeded",
		Fields: map[string]interface{}{
			"url":   url,
			"host":  host,
			"error": errorMsg,
		},
	}

	l.logStructured(entry)
}

// LogRobotsFetch logs a robots.txt fetch event
func (l *Logger) LogRobotsFetch(host string, success bool, errorMsg string) {
	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     "info",
		Source:    "yfinance-go/scrape",
		Message:   "robots.txt fetch",
		Fields: map[string]interface{}{
			"host":    host,
			"success": success,
		},
	}

	if !success {
		entry.Level = "warn"
		entry.Fields["error"] = errorMsg
	}

	l.logStructured(entry)
}

// LogConfig logs configuration information
func (l *Logger) LogConfig(config *Config) {
	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     "info",
		Source:    "yfinance-go/scrape",
		Message:   "scrape configuration",
		Fields: map[string]interface{}{
			"enabled":        config.Enabled,
			"user_agent":     config.UserAgent,
			"timeout_ms":     config.TimeoutMs,
			"qps":            config.QPS,
			"burst":          config.Burst,
			"robots_policy":  config.RobotsPolicy,
			"retry_attempts": config.Retry.Attempts,
			"retry_base_ms":  config.Retry.BaseMs,
			"retry_max_ms":   config.Retry.MaxDelayMs,
		},
	}

	l.logStructured(entry)
}

// LogError logs a general error
func (l *Logger) LogError(message string, err error, fields map[string]interface{}) {
	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     "error",
		Source:    "yfinance-go/scrape",
		Message:   message,
		Fields:    fields,
	}

	if err != nil {
		entry.Fields["error"] = err.Error()
	}

	l.logStructured(entry)
}

// LogInfo logs an info message
func (l *Logger) LogInfo(message string, fields map[string]interface{}) {
	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     "info",
		Source:    "yfinance-go/scrape",
		Message:   message,
		Fields:    fields,
	}

	l.logStructured(entry)
}

// LogDebug logs a debug message
func (l *Logger) LogDebug(message string, fields map[string]interface{}) {
	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     "debug",
		Source:    "yfinance-go/scrape",
		Message:   message,
		Fields:    fields,
	}

	l.logStructured(entry)
}

// logStructured logs a structured entry as JSON
func (l *Logger) logStructured(entry LogEntry) {
	data, err := json.Marshal(entry)
	if err != nil {
		// Fallback to simple logging if JSON marshaling fails
		l.logger.Printf("scrape: %s - %s", entry.Level, entry.Message)
		return
	}

	l.logger.Println(string(data))
}

// SetOutput sets the output destination for the logger
func (l *Logger) SetOutput(output interface{ Write([]byte) (int, error) }) {
	l.logger.SetOutput(output)
}

// GetStats returns logger statistics
func (l *Logger) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"logger_type": "structured_json",
		"source":      "yfinance-go/scrape",
	}
}
