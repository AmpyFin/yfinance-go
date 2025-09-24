package httpx

import (
	"testing"
	"time"
)

func TestCalculateBackoff(t *testing.T) {
	config := &Config{
		BackoffBaseMs: 100,
		MaxDelayMs:    5000,
	}
	
	client := &Client{config: config}
	
	// Test first attempt (should return base delay)
	delay := client.calculateBackoff(0)
	expected := 100 * time.Millisecond
	if delay != expected {
		t.Errorf("Expected first attempt delay to be %v, got %v", expected, delay)
	}
	
	// Test subsequent attempts (should have exponential backoff with jitter)
	delay1 := client.calculateBackoff(1)
	delay2 := client.calculateBackoff(2)
	delay3 := client.calculateBackoff(3)
	
	// All delays should be within reasonable bounds
	if delay1 < 100*time.Millisecond || delay1 > 1000*time.Millisecond {
		t.Errorf("Delay 1 out of bounds: %v", delay1)
	}
	if delay2 < 100*time.Millisecond || delay2 > 2000*time.Millisecond {
		t.Errorf("Delay 2 out of bounds: %v", delay2)
	}
	if delay3 < 100*time.Millisecond || delay3 > 5000*time.Millisecond {
		t.Errorf("Delay 3 out of bounds: %v", delay3)
	}
	
	// With jitter, delays can vary, so we just check they're within bounds
	// The important thing is that they're not all the same due to jitter
}

func TestCalculateBackoffMaxDelay(t *testing.T) {
	config := &Config{
		BackoffBaseMs: 1000,
		MaxDelayMs:    2000, // Low max delay
	}
	
	client := &Client{config: config}
	
	// High attempt number should be capped at max delay
	delay := client.calculateBackoff(10)
	maxDelay := 2000 * time.Millisecond
	
	if delay > maxDelay {
		t.Errorf("Expected delay to be capped at %v, got %v", maxDelay, delay)
	}
}

func TestCalculateBackoffJitter(t *testing.T) {
	config := &Config{
		BackoffBaseMs: 100,
		MaxDelayMs:    10000,
	}
	
	client := &Client{config: config}
	
	// Run multiple times to test jitter
	delays := make([]time.Duration, 10)
	for i := 0; i < 10; i++ {
		delays[i] = client.calculateBackoff(2)
	}
	
	// Check that we have some variation (jitter is working)
	allSame := true
	for i := 1; i < len(delays); i++ {
		if delays[i] != delays[0] {
			allSame = false
			break
		}
	}
	
	if allSame {
		t.Error("Expected jitter to produce different delays, but all delays were the same")
	}
}
