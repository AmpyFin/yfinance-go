package httpx

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSessionManager(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()
	
	// Create session manager
	sm := NewSessionManager(server.URL, 3, 2, 100) // 3 sessions, eject after 2 failures, 100ms cooldown
	
	// Test session count
	if sm.GetSessionCount() != 3 {
		t.Errorf("Expected 3 sessions, got %d", sm.GetSessionCount())
	}
	
	// Test session rotation
	session1 := sm.GetNextSession()
	_ = sm.GetNextSession()
	_ = sm.GetNextSession()
	session4 := sm.GetNextSession()
	
	// Should rotate back to first session
	if session4 != session1 {
		t.Error("Expected session rotation to wrap around")
	}
	
	// Test session ID tracking
	sessionID := sm.GetCurrentSessionID()
	if sessionID == "" {
		t.Error("Expected non-empty session ID")
	}
}

func TestSessionManagerHealthTracking(t *testing.T) {
	// Create a test server that fails
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()
	
	// Create session manager with low failure threshold
	sm := NewSessionManager(server.URL, 2, 1, 50) // Eject after 1 failure, 50ms cooldown
	
	// Get a session
	_ = sm.GetNextSession()
	sessionID := sm.GetCurrentSessionID()
	
	// Record a failure
	sm.RecordSessionFailure(sessionID)
	
	// Session should be marked as unhealthy
	stats := sm.GetSessionStats()
	if stats["unhealthy_sessions"].(int) != 1 {
		t.Errorf("Expected 1 unhealthy session, got %d", stats["unhealthy_sessions"])
	}
	
	// Wait for cooldown and recreation
	time.Sleep(100 * time.Millisecond)
	
	// Session should be recreated and healthy again
	stats = sm.GetSessionStats()
	if stats["unhealthy_sessions"].(int) != 0 {
		t.Errorf("Expected 0 unhealthy sessions after recreation, got %d", stats["unhealthy_sessions"])
	}
}

func TestSessionManagerSuccessTracking(t *testing.T) {
	sm := NewSessionManager("http://example.com", 2, 3, 100)
	
	sessionID := "session_0"
	
	// Record some failures
	sm.RecordSessionFailure(sessionID)
	sm.RecordSessionFailure(sessionID)
	
	// Record a success
	sm.RecordSessionSuccess(sessionID)
	
	// Session should be healthy again
	stats := sm.GetSessionStats()
	if stats["unhealthy_sessions"].(int) != 0 {
		t.Errorf("Expected 0 unhealthy sessions after success, got %d", stats["unhealthy_sessions"])
	}
}

func TestSessionManagerCookieManagement(t *testing.T) {
	sm := NewSessionManager("http://example.com", 2, 3, 100)
	
	// Test adding custom cookie
	err := sm.AddCustomCookie("test", "value")
	if err != nil {
		t.Errorf("Expected no error adding cookie, got %v", err)
	}
	
	// Test clearing cookies
	err = sm.ClearAllCookies()
	if err != nil {
		t.Errorf("Expected no error clearing cookies, got %v", err)
	}
}

func TestSessionManagerStats(t *testing.T) {
	sm := NewSessionManager("http://example.com", 5, 3, 150)
	
	stats := sm.GetSessionStats()
	
	expectedFields := []string{
		"total_sessions",
		"healthy_sessions", 
		"unhealthy_sessions",
		"current_session",
		"base_url",
		"eject_after",
		"recreate_cooldown_ms",
	}
	
	for _, field := range expectedFields {
		if _, exists := stats[field]; !exists {
			t.Errorf("Expected stats to contain field %s", field)
		}
	}
	
	if stats["total_sessions"].(int) != 5 {
		t.Errorf("Expected 5 total sessions, got %d", stats["total_sessions"])
	}
	
	if stats["eject_after"].(int) != 3 {
		t.Errorf("Expected eject_after to be 3, got %d", stats["eject_after"])
	}
}
