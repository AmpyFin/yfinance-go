package httpx

import (
	"context"
	"testing"
	"time"
)

func TestTokenBucketLimiter(t *testing.T) {
	// Test basic token bucket functionality
	limiter := NewRateLimiter(2, 5) // 2 QPS, burst of 5
	
	ctx := context.Background()
	
	// Should be able to consume 5 tokens immediately (burst)
	for i := 0; i < 5; i++ {
		err := limiter.Wait(ctx)
		if err != nil {
			t.Errorf("Expected no error for burst token %d, got %v", i, err)
		}
	}
	
	// Next token should require waiting
	start := time.Now()
	err := limiter.Wait(ctx)
	duration := time.Since(start)
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	// Should have waited approximately 500ms (1/2 QPS)
	expectedMin := 400 * time.Millisecond
	expectedMax := 600 * time.Millisecond
	
	if duration < expectedMin || duration > expectedMax {
		t.Errorf("Expected wait time between %v and %v, got %v", expectedMin, expectedMax, duration)
	}
}

func TestTokenBucketLimiterContextCancellation(t *testing.T) {
	limiter := NewRateLimiter(1, 1) // Very slow rate
	
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	
	// First token should work (burst)
	err := limiter.Wait(ctx)
	if err != nil {
		t.Errorf("Expected no error for burst token, got %v", err)
	}
	
	// Second token should timeout
	err = limiter.Wait(ctx)
	if err != context.DeadlineExceeded {
		t.Errorf("Expected context.DeadlineExceeded, got %v", err)
	}
}

func TestTokenBucketLimiterConcurrency(t *testing.T) {
	limiter := NewRateLimiter(10, 5) // 10 QPS, burst of 5
	
	ctx := context.Background()
	done := make(chan bool, 10)
	
	// Start 10 goroutines trying to get tokens
	for i := 0; i < 10; i++ {
		go func() {
			err := limiter.Wait(ctx)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
			done <- true
		}()
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Fatal("Timeout waiting for goroutines to complete")
		}
	}
}
