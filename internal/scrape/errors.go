package scrape

import (
	"fmt"
	"net/http"
)

// ScrapeError represents a scraping-specific error
type ScrapeError struct {
	Type    string
	Message string
	URL     string
	Status  int
}

func (e *ScrapeError) Error() string {
	if e.Status > 0 {
		return fmt.Sprintf("%s: %s (URL: %s, Status: %d)", e.Type, e.Message, e.URL, e.Status)
	}
	return fmt.Sprintf("%s: %s (URL: %s)", e.Type, e.Message, e.URL)
}

// Predefined error types
var (
	ErrRobotsDenied     = &ScrapeError{Type: "robots_denied", Message: "robots.txt disallows this path"}
	ErrTimeout          = &ScrapeError{Type: "timeout", Message: "request timeout"}
	ErrTooManyRedirects = &ScrapeError{Type: "too_many_redirects", Message: "exceeded maximum redirect limit"}
	ErrRetryExhausted   = &ScrapeError{Type: "retry_exhausted", Message: "maximum retry attempts exceeded"}
	ErrRateLimited      = &ScrapeError{Type: "rate_limited", Message: "rate limit exceeded"}
	ErrCircuitOpen      = &ScrapeError{Type: "circuit_open", Message: "circuit breaker is open"}
	ErrInvalidURL       = &ScrapeError{Type: "invalid_url", Message: "invalid URL format"}
	ErrContentTooLarge  = &ScrapeError{Type: "content_too_large", Message: "response content exceeds size limit"}
	
	// Parse-specific errors
	ErrNoQuoteSummary   = &ScrapeError{Type: "no_quote_summary", Message: "could not locate quoteSummary script payload"}
	ErrJSONUnescape     = &ScrapeError{Type: "json_unescape", Message: "failed to unescape JSON from envelope body"}
	ErrJSONDecode       = &ScrapeError{Type: "json_decode", Message: "failed to decode JSON structure"}
	ErrMissingFieldBase = &ScrapeError{Type: "missing_field", Message: "required field is missing"}
	ErrSchemaDriftBase  = &ScrapeError{Type: "schema_drift", Message: "unexpected schema change detected"}
	
	// News-specific errors
	ErrNewsNoArticles   = &ScrapeError{Type: "news_no_articles", Message: "no news articles found"}
	ErrNewsParse        = &ScrapeError{Type: "news_parse", Message: "failed to parse news HTML"}
)

// ErrHTTP creates an HTTP status error
func ErrHTTP(status int, url string) *ScrapeError {
	return &ScrapeError{
		Type:    "http_error",
		Message: http.StatusText(status),
		URL:     url,
		Status:  status,
	}
}

// ErrMissingField creates a missing field error
func ErrMissingField(field string) *ScrapeError {
	return &ScrapeError{
		Type:    "missing_field",
		Message: fmt.Sprintf("required field '%s' is missing", field),
	}
}

// ErrSchemaDrift creates a schema drift error
func ErrSchemaDrift(field string) *ScrapeError {
	return &ScrapeError{
		Type:    "schema_drift",
		Message: fmt.Sprintf("unexpected schema change in field '%s'", field),
	}
}

// IsRetryableError determines if an error should trigger a retry
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check for specific retryable error types
	if scrapeErr, ok := err.(*ScrapeError); ok {
		switch scrapeErr.Type {
		case "timeout", "rate_limited":
			return true
		case "http_error":
			// Retry on 429, 5xx errors
			return scrapeErr.Status == 429 || (scrapeErr.Status >= 500 && scrapeErr.Status < 600)
		}
		return false
	}

	// Check for network errors that should be retried
	return isNetworkError(err)
}

// isNetworkError checks if an error is a network-related error that should be retried
func isNetworkError(err error) bool {
	// This would typically check for net.Error types, connection refused, etc.
	// For now, we'll implement a basic check
	errStr := err.Error()
	networkErrors := []string{
		"connection refused",
		"connection reset",
		"timeout",
		"no such host",
		"network is unreachable",
	}
	
	for _, netErr := range networkErrors {
		if contains(errStr, netErr) {
			return true
		}
	}
	return false
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || 
		    (len(s) > len(substr) && 
		     (s[:len(substr)] == substr || 
		      s[len(s)-len(substr):] == substr || 
		      indexOf(s, substr) >= 0)))
}

// indexOf finds the index of a substring in a string
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
