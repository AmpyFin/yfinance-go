package httpx

import (
	"errors"
	"fmt"
)

// Error types for HTTP operations
var (
	ErrTooManyRequests   = errors.New("too many requests (429)")
	ErrServerUnavailable = errors.New("server unavailable (5xx)")
	ErrDecode            = errors.New("decode error")
	ErrClientConfig      = errors.New("client configuration error")
	ErrRateLimited       = errors.New("rate limited")
	ErrCircuitOpen       = errors.New("circuit breaker is open")
	ErrTimeout           = errors.New("request timeout")
	ErrContextCanceled   = errors.New("context canceled")
)

// HTTPError wraps HTTP status errors with additional context
type HTTPError struct {
	StatusCode int
	Message    string
	Err        error
}

func (e *HTTPError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("HTTP %d: %s: %v", e.StatusCode, e.Message, e.Err)
	}
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Message)
}

func (e *HTTPError) Unwrap() error {
	return e.Err
}

// NewHTTPError creates a new HTTP error
func NewHTTPError(statusCode int, message string, err error) *HTTPError {
	return &HTTPError{
		StatusCode: statusCode,
		Message:    message,
		Err:        err,
	}
}

// IsRetryableError checks if an error should trigger a retry
func IsRetryableError(err error) bool {
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		switch httpErr.StatusCode {
		case 429, 500, 502, 503, 504:
			return true
		default:
			return false
		}
	}

	// Check for specific error types
	return errors.Is(err, ErrTooManyRequests) ||
		errors.Is(err, ErrServerUnavailable) ||
		errors.Is(err, ErrTimeout)
}

// IsFatalError checks if an error is fatal and should not be retried
func IsFatalError(err error) bool {
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		switch httpErr.StatusCode {
		case 400, 401, 403, 404, 422:
			return true
		default:
			return false
		}
	}

	return errors.Is(err, ErrClientConfig) ||
		errors.Is(err, ErrDecode)
}

// TransportError represents a network transport error
type TransportError struct {
	Err error
}

func (e *TransportError) Error() string {
	return fmt.Sprintf("transport error: %v", e.Err)
}

func (e *TransportError) Unwrap() error {
	return e.Err
}

// NewTransportError creates a new transport error
func NewTransportError(err error) *TransportError {
	return &TransportError{Err: err}
}
