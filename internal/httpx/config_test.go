package httpx

import (
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	
	// Test that all required fields are set
	if config.BaseURL == "" {
		t.Error("Expected BaseURL to be set")
	}
	
	if config.Timeout <= 0 {
		t.Error("Expected Timeout to be positive")
	}
	
	if config.IdleTimeout <= 0 {
		t.Error("Expected IdleTimeout to be positive")
	}
	
	if config.MaxConnsPerHost <= 0 {
		t.Error("Expected MaxConnsPerHost to be positive")
	}
	
	if config.MaxAttempts <= 0 {
		t.Error("Expected MaxAttempts to be positive")
	}
	
	if config.BackoffBaseMs <= 0 {
		t.Error("Expected BackoffBaseMs to be positive")
	}
	
	if config.BackoffJitterMs <= 0 {
		t.Error("Expected BackoffJitterMs to be positive")
	}
	
	if config.MaxDelayMs <= 0 {
		t.Error("Expected MaxDelayMs to be positive")
	}
	
	if config.QPS <= 0 {
		t.Error("Expected QPS to be positive")
	}
	
	if config.Burst <= 0 {
		t.Error("Expected Burst to be positive")
	}
	
	if config.CircuitWindow <= 0 {
		t.Error("Expected CircuitWindow to be positive")
	}
	
	if config.FailureThreshold <= 0 {
		t.Error("Expected FailureThreshold to be positive")
	}
	
	if config.ResetTimeout <= 0 {
		t.Error("Expected ResetTimeout to be positive")
	}
	
	if config.UserAgent == "" {
		t.Error("Expected UserAgent to be set")
	}
}

func TestConfigFields(t *testing.T) {
	// Test that we can create a config with all fields
	config := &Config{
		BaseURL:            "https://example.com",
		Timeout:            30 * time.Second,
		IdleTimeout:        90 * time.Second,
		MaxConnsPerHost:    10,
		MaxAttempts:        3,
		BackoffBaseMs:      200,
		BackoffJitterMs:    100,
		MaxDelayMs:         10000,
		QPS:                1.0,
		Burst:              3,
		CircuitWindow:      60 * time.Second,
		FailureThreshold:   3,
		ResetTimeout:       30 * time.Second,
		UserAgent:          "test-agent",
		EnableSessionRotation: false,
		NumSessions:        1,
	}
	
	// Basic validation that fields are set
	if config.BaseURL == "" {
		t.Error("BaseURL should not be empty")
	}
	if config.Timeout <= 0 {
		t.Error("Timeout should be positive")
	}
	if config.UserAgent == "" {
		t.Error("UserAgent should not be empty")
	}
}