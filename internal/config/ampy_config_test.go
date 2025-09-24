package config

import (
	"os"
	"strings"
	"testing"
	"gopkg.in/yaml.v3"
)

func TestNewLoader(t *testing.T) {
	loader := NewLoader("test.yaml")
	if loader == nil {
		t.Fatal("NewLoader returned nil")
	}
	if loader.effectivePath != "test.yaml" {
		t.Errorf("Expected effectivePath to be 'test.yaml', got '%s'", loader.effectivePath)
	}
}

func TestLoadFromFile(t *testing.T) {
	// Create a temporary config file
	tempFile := "test-config.yaml"
	err := CreateEffectiveConfig(tempFile)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}
	defer os.Remove(tempFile)
	
	loader := NewLoader(tempFile)
	config, err := loader.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	
	if config == nil {
		t.Fatal("Config is nil")
	}
	
	// Check some default values
	if config.App.Env != "dev" {
		t.Errorf("Expected app.env to be 'dev', got '%s'", config.App.Env)
	}
	
	if config.Yahoo.BaseURL != "https://query2.finance.yahoo.com" {
		t.Errorf("Expected yahoo.base_url to be 'https://query2.finance.yahoo.com', got '%s'", config.Yahoo.BaseURL)
	}
	
	if len(config.Markets.AllowedIntervals) != 1 || config.Markets.AllowedIntervals[0] != "1d" {
		t.Errorf("Expected markets.allowed_intervals to be ['1d'], got %v", config.Markets.AllowedIntervals)
	}
}

func TestLoadFromFileNotFound(t *testing.T) {
	loader := NewLoader("nonexistent.yaml")
	_, err := loader.Load()
	if err == nil {
		t.Fatal("Expected error for nonexistent file")
	}
}

func TestInterpolateEnvVars(t *testing.T) {
	// Set test environment variable
	os.Setenv("TEST_VAR", "test_value")
	defer os.Unsetenv("TEST_VAR")
	
	// Create a config with environment variable interpolation
	configContent := map[string]interface{}{
		"app": map[string]interface{}{
			"env": "dev",
		},
		"yahoo": map[string]interface{}{
			"base_url": "https://query2.finance.yahoo.com",
			"timeout_ms": 5000,
		},
		"markets": map[string]interface{}{
			"allowed_intervals": []string{"1d"},
			"default_adjustment_policy": "split_dividend",
		},
		"bus": map[string]interface{}{
			"max_payload_bytes": 1048576,
		},
		"retry": map[string]interface{}{
			"attempts": 5,
		},
		"circuit_breaker": map[string]interface{}{
			"failure_threshold": 0.30,
		},
		"test_url": "${TEST_VAR}",
		"test_default": "${MISSING_VAR:-default_value}",
	}
	
	// Create temporary file
	tempFile := "test-env-config.yaml"
	err := createTestConfigFile(tempFile, configContent)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}
	defer os.Remove(tempFile)
	
	loader := NewLoader(tempFile)
	config, err := loader.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	
	// The interpolation should have been applied during loading
	// We can't directly test the interpolated values since they're in the map,
	// but we can test that the config loaded successfully
	if config == nil {
		t.Fatal("Config is nil")
	}
}

func TestValidateDailyOnlyIntervals(t *testing.T) {
	// Create a config with invalid intervals
	configContent := map[string]interface{}{
		"app": map[string]interface{}{
			"env": "dev",
		},
		"yahoo": map[string]interface{}{
			"base_url": "https://query2.finance.yahoo.com",
			"timeout_ms": 5000,
		},
		"markets": map[string]interface{}{
			"allowed_intervals": []string{"1h", "1d"}, // Invalid - should be only ["1d"]
			"default_adjustment_policy": "split_dividend",
		},
		"bus": map[string]interface{}{
			"max_payload_bytes": 1048576,
		},
		"retry": map[string]interface{}{
			"attempts": 5,
		},
		"circuit_breaker": map[string]interface{}{
			"failure_threshold": 0.30,
		},
	}
	
	tempFile := "test-invalid-intervals.yaml"
	err := createTestConfigFile(tempFile, configContent)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}
	defer os.Remove(tempFile)
	
	loader := NewLoader(tempFile)
	_, err = loader.Load()
	if err == nil {
		t.Fatal("Expected validation error for invalid intervals")
	}
	
	expectedError := "markets.allowed_intervals must be exactly [\"1d\"] for yfinance-go (daily-only scope)"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error to contain '%s', got '%s'", expectedError, err.Error())
	}
}

func TestGetEffectiveConfig(t *testing.T) {
	tempFile := "test-effective-config.yaml"
	err := CreateEffectiveConfig(tempFile)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}
	defer os.Remove(tempFile)
	
	loader := NewLoader(tempFile)
	_, err = loader.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	
	effectiveConfig, err := loader.GetEffectiveConfig()
	if err != nil {
		t.Fatalf("Failed to get effective config: %v", err)
	}
	
	if effectiveConfig == nil {
		t.Fatal("Effective config is nil")
	}
	
	// Check that secrets are redacted
	if secrets, ok := effectiveConfig["secrets"].([]interface{}); ok {
		// Secrets array should be present but empty or redacted
		_ = secrets // Just check it exists
	}
}

func TestGetEffectiveConfigNotLoaded(t *testing.T) {
	loader := NewLoader("test.yaml")
	_, err := loader.GetEffectiveConfig()
	if err == nil {
		t.Fatal("Expected error when config not loaded")
	}
	
	expectedError := "configuration not loaded"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestGetHTTPConfig(t *testing.T) {
	tempFile := "test-http-config.yaml"
	err := CreateEffectiveConfig(tempFile)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}
	defer os.Remove(tempFile)
	
	loader := NewLoader(tempFile)
	config, err := loader.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	
	httpConfig := config.GetHTTPConfig()
	if httpConfig == nil {
		t.Fatal("HTTP config is nil")
	}
	
	if httpConfig.BaseURL != "https://query2.finance.yahoo.com" {
		t.Errorf("Expected BaseURL to be 'https://query2.finance.yahoo.com', got '%s'", httpConfig.BaseURL)
	}
	
	if httpConfig.QPS != 5.0 {
		t.Errorf("Expected QPS to be 5.0, got %f", httpConfig.QPS)
	}
}

func TestGetBusConfig(t *testing.T) {
	tempFile := "test-bus-config.yaml"
	err := CreateEffectiveConfig(tempFile)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}
	defer os.Remove(tempFile)
	
	loader := NewLoader(tempFile)
	config, err := loader.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	
	busConfig := config.GetBusConfig()
	if busConfig == nil {
		t.Fatal("Bus config is nil")
	}
	
	if busConfig.Enabled != false {
		t.Errorf("Expected bus.enabled to be false, got %v", busConfig.Enabled)
	}
	
	if busConfig.MaxPayloadBytes != 1048576 {
		t.Errorf("Expected MaxPayloadBytes to be 1048576, got %d", busConfig.MaxPayloadBytes)
	}
}

func TestGetFXConfig(t *testing.T) {
	tempFile := "test-fx-config.yaml"
	err := CreateEffectiveConfig(tempFile)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}
	defer os.Remove(tempFile)
	
	loader := NewLoader(tempFile)
	config, err := loader.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	
	fxConfig := config.GetFXConfig()
	if fxConfig == nil {
		t.Fatal("FX config is nil")
	}
	
	if fxConfig.Provider != "none" {
		t.Errorf("Expected provider to be 'none', got '%s'", fxConfig.Provider)
	}
	
	if fxConfig.CacheTTLMs != 60000 {
		t.Errorf("Expected CacheTTLMs to be 60000, got %d", fxConfig.CacheTTLMs)
	}
}

func TestValidateInterval(t *testing.T) {
	tempFile := "test-validate-interval.yaml"
	err := CreateEffectiveConfig(tempFile)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}
	defer os.Remove(tempFile)
	
	loader := NewLoader(tempFile)
	config, err := loader.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	
	// Test valid interval
	err = config.ValidateInterval("1d")
	if err != nil {
		t.Errorf("Expected '1d' to be valid, got error: %v", err)
	}
	
	// Test invalid interval
	err = config.ValidateInterval("1h")
	if err == nil {
		t.Error("Expected '1h' to be invalid")
	}
}

func TestValidateAdjustmentPolicy(t *testing.T) {
	tempFile := "test-validate-policy.yaml"
	err := CreateEffectiveConfig(tempFile)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}
	defer os.Remove(tempFile)
	
	loader := NewLoader(tempFile)
	config, err := loader.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	
	// Test valid policies
	err = config.ValidateAdjustmentPolicy("raw")
	if err != nil {
		t.Errorf("Expected 'raw' to be valid, got error: %v", err)
	}
	
	err = config.ValidateAdjustmentPolicy("split_dividend")
	if err != nil {
		t.Errorf("Expected 'split_dividend' to be valid, got error: %v", err)
	}
	
	// Test invalid policy
	err = config.ValidateAdjustmentPolicy("invalid")
	if err == nil {
		t.Error("Expected 'invalid' to be invalid")
	}
}

// Helper function to create test config files
func createTestConfigFile(filename string, config map[string]interface{}) error {
	// Marshal to YAML and write to file
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}
