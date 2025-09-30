package soak

import (
	"context"
	"encoding/json"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// FailureServer simulates various failure scenarios for soak testing
type FailureServer struct {
	failureRate float64
	logger      *zap.Logger
	server      *http.Server
	scenarios   []FailureScenario
	mu          sync.RWMutex
	active      bool
}

// FailureScenario represents a specific failure pattern
type FailureScenario struct {
	Name        string
	Probability float64
	StatusCode  int
	Delay       time.Duration
	Response    string
	Headers     map[string]string
}

// FailureStats tracks failure injection statistics
type FailureStats struct {
	TotalRequests   int64
	FailuresSent    int64
	ScenarioStats   map[string]int64
	LastFailureTime time.Time
	mu              sync.RWMutex
}

// NewFailureServer creates a new failure injection server
func NewFailureServer(failureRate float64, logger *zap.Logger) *FailureServer {
	scenarios := []FailureScenario{
		{
			Name:        "rate_limit",
			Probability: 0.3,
			StatusCode:  429,
			Delay:       100 * time.Millisecond,
			Response:    `{"error": "Rate limit exceeded", "retry_after": 60}`,
			Headers:     map[string]string{"Retry-After": "60"},
		},
		{
			Name:        "server_error",
			Probability: 0.2,
			StatusCode:  500,
			Delay:       50 * time.Millisecond,
			Response:    `{"error": "Internal server error"}`,
		},
		{
			Name:        "bad_gateway",
			Probability: 0.15,
			StatusCode:  502,
			Delay:       200 * time.Millisecond,
			Response:    `{"error": "Bad gateway"}`,
		},
		{
			Name:        "service_unavailable",
			Probability: 0.15,
			StatusCode:  503,
			Delay:       300 * time.Millisecond,
			Response:    `{"error": "Service temporarily unavailable"}`,
			Headers:     map[string]string{"Retry-After": "30"},
		},
		{
			Name:        "timeout",
			Probability: 0.1,
			StatusCode:  0, // Special case: connection timeout
			Delay:       30 * time.Second,
			Response:    "",
		},
		{
			Name:        "auth_required",
			Probability: 0.1,
			StatusCode:  401,
			Delay:       50 * time.Millisecond,
			Response:    `{"error": "Authentication required"}`,
			Headers:     map[string]string{"WWW-Authenticate": "Bearer"},
		},
	}

	return &FailureServer{
		failureRate: failureRate,
		logger:      logger,
		scenarios:   scenarios,
	}
}

// Start starts the failure injection server
func (fs *FailureServer) Start() error {
	if fs.failureRate <= 0 {
		fs.logger.Info("Failure injection disabled (rate = 0)")
		return nil
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", fs.handleRequest)
	mux.HandleFunc("/health", fs.handleHealth)
	mux.HandleFunc("/stats", fs.handleStats)
	mux.HandleFunc("/scenarios", fs.handleScenarios)

	fs.server = &http.Server{
		Addr:         ":8080", // Failure injection server port
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	fs.mu.Lock()
	fs.active = true
	fs.mu.Unlock()

	go func() {
		fs.logger.Info("Failure injection server starting",
			zap.String("addr", fs.server.Addr),
			zap.Float64("failure_rate", fs.failureRate))

		if err := fs.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fs.logger.Error("Failure server error", zap.Error(err))
		}
	}()

	// Wait a moment for server to start
	time.Sleep(100 * time.Millisecond)

	return nil
}

// Stop stops the failure injection server
func (fs *FailureServer) Stop() error {
	if fs.server == nil {
		return nil
	}

	fs.mu.Lock()
	fs.active = false
	fs.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fs.logger.Info("Stopping failure injection server")
	return fs.server.Shutdown(ctx)
}

// handleRequest handles incoming requests and potentially injects failures
func (fs *FailureServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	fs.mu.RLock()
	active := fs.active
	fs.mu.RUnlock()

	if !active {
		http.Error(w, "Failure server not active", http.StatusServiceUnavailable)
		return
	}

	// Decide whether to inject a failure
	if rand.Float64() < fs.failureRate {
		fs.injectFailure(w, r)
		return
	}

	// Return success response
	fs.handleSuccess(w, r)
}

// injectFailure injects a random failure scenario
func (fs *FailureServer) injectFailure(w http.ResponseWriter, r *http.Request) {
	scenario := fs.selectFailureScenario()

	fs.logger.Debug("Injecting failure",
		zap.String("scenario", scenario.Name),
		zap.String("path", r.URL.Path),
		zap.Int("status_code", scenario.StatusCode))

	// Add delay if specified
	if scenario.Delay > 0 {
		time.Sleep(scenario.Delay)
	}

	// Handle timeout scenario (close connection)
	if scenario.StatusCode == 0 {
		// Simulate connection timeout by closing without response
		if hijacker, ok := w.(http.Hijacker); ok {
			conn, _, err := hijacker.Hijack()
			if err == nil {
				conn.Close()
				return
			}
		}
		// Fallback to 408 if hijacking fails
		scenario.StatusCode = 408
		scenario.Response = `{"error": "Request timeout"}`
	}

	// Set custom headers
	for key, value := range scenario.Headers {
		w.Header().Set(key, value)
	}

	// Set content type
	w.Header().Set("Content-Type", "application/json")

	// Write status code and response
	w.WriteHeader(scenario.StatusCode)
	if scenario.Response != "" {
		_, _ = w.Write([]byte(scenario.Response))
	}
}

// selectFailureScenario selects a failure scenario based on probabilities
func (fs *FailureServer) selectFailureScenario() FailureScenario {
	totalProb := 0.0
	for _, scenario := range fs.scenarios {
		totalProb += scenario.Probability
	}

	r := rand.Float64() * totalProb
	cumulative := 0.0

	for _, scenario := range fs.scenarios {
		cumulative += scenario.Probability
		if r <= cumulative {
			return scenario
		}
	}

	// Fallback to first scenario
	return fs.scenarios[0]
}

// handleSuccess returns a successful response
func (fs *FailureServer) handleSuccess(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "success",
		"timestamp": time.Now().Unix(),
		"path":      r.URL.Path,
		"method":    r.Method,
		"message":   "Request processed successfully",
	}

	// Simulate different response types based on path
	if strings.Contains(r.URL.Path, "quote") {
		response["data"] = map[string]interface{}{
			"symbol": "TEST",
			"price":  100.50,
			"change": 1.25,
		}
	} else if strings.Contains(r.URL.Path, "financials") {
		response["data"] = map[string]interface{}{
			"revenue":    1000000000,
			"net_income": 200000000,
			"currency":   "USD",
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		fs.logger.Error("Failed to encode success response", zap.Error(err))
	}
}

// handleHealth returns server health status
func (fs *FailureServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	fs.mu.RLock()
	active := fs.active
	fs.mu.RUnlock()

	health := map[string]interface{}{
		"status":       "healthy",
		"active":       active,
		"failure_rate": fs.failureRate,
		"timestamp":    time.Now().Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(health)
}

// handleStats returns failure injection statistics
func (fs *FailureServer) handleStats(w http.ResponseWriter, r *http.Request) {
	// This would return actual statistics if we were tracking them
	stats := map[string]interface{}{
		"total_requests":    0,
		"failures_injected": 0,
		"scenarios":         fs.scenarios,
		"timestamp":         time.Now().Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(stats)
}

// handleScenarios returns available failure scenarios
func (fs *FailureServer) handleScenarios(w http.ResponseWriter, r *http.Request) {
	scenarios := map[string]interface{}{
		"scenarios":    fs.scenarios,
		"failure_rate": fs.failureRate,
		"timestamp":    time.Now().Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(scenarios)
}

// UpdateFailureRate updates the failure injection rate
func (fs *FailureServer) UpdateFailureRate(rate float64) {
	fs.mu.Lock()
	fs.failureRate = rate
	fs.mu.Unlock()

	fs.logger.Info("Updated failure rate", zap.Float64("new_rate", rate))
}

// AddScenario adds a new failure scenario
func (fs *FailureServer) AddScenario(scenario FailureScenario) {
	fs.mu.Lock()
	fs.scenarios = append(fs.scenarios, scenario)
	fs.mu.Unlock()

	fs.logger.Info("Added failure scenario", zap.String("name", scenario.Name))
}

// IsActive returns whether the failure server is active
func (fs *FailureServer) IsActive() bool {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	return fs.active
}
