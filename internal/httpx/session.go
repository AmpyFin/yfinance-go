package httpx

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sync"
	"time"
)

// SessionManager manages multiple HTTP sessions with cookie rotation
// This helps avoid rate limiting by rotating between different sessions
type SessionManager struct {
	sessions []*http.Client
	current  int
	mu       sync.RWMutex
	baseURL  string
}

// NewSessionManager creates a new session manager with multiple sessions
func NewSessionManager(baseURL string, numSessions int) *SessionManager {
	if numSessions <= 0 {
		numSessions = 5 // Default to 5 sessions
	}
	
	sessions := make([]*http.Client, numSessions)
	
	for i := 0; i < numSessions; i++ {
		// Create a cookie jar for each session
		jar, err := cookiejar.New(nil)
		if err != nil {
			// Fallback to no cookies if jar creation fails
			jar = nil
		}
		
		// Create HTTP client with cookie jar
		client := &http.Client{
			Jar: jar,
			Timeout: 30 * time.Second,
		}
		
		sessions[i] = client
	}
	
	return &SessionManager{
		sessions: sessions,
		current:  0,
		baseURL:  baseURL,
	}
}

// GetNextSession returns the next session in rotation
func (sm *SessionManager) GetNextSession() *http.Client {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	session := sm.sessions[sm.current]
	sm.current = (sm.current + 1) % len(sm.sessions)
	
	return session
}

// InitializeSessions initializes all sessions by making a request to get initial cookies
func (sm *SessionManager) InitializeSessions() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	for i, session := range sm.sessions {
		// Make a simple request to initialize the session and get cookies
		req, err := http.NewRequest("GET", sm.baseURL, nil)
		if err != nil {
			return fmt.Errorf("failed to create request for session %d: %w", i, err)
		}
		
		// Set a realistic User-Agent
		req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
		
		// Make the request (we don't care about the response, just want cookies)
		resp, err := session.Do(req)
		if err != nil {
			// Don't fail completely, just log and continue
			continue
		}
		resp.Body.Close()
	}
	
	return nil
}

// GetSessionCount returns the number of sessions
func (sm *SessionManager) GetSessionCount() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.sessions)
}

// RotateSession manually rotates to the next session
func (sm *SessionManager) RotateSession() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.current = (sm.current + 1) % len(sm.sessions)
}

// GetCurrentSessionIndex returns the current session index
func (sm *SessionManager) GetCurrentSessionIndex() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.current
}

// AddCustomCookie adds a custom cookie to all sessions
func (sm *SessionManager) AddCustomCookie(name, value string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	baseURL, err := url.Parse(sm.baseURL)
	if err != nil {
		return fmt.Errorf("failed to parse base URL: %w", err)
	}
	
	for _, session := range sm.sessions {
		if session.Jar != nil {
			cookie := &http.Cookie{
				Name:   name,
				Value:  value,
				Domain: baseURL.Host,
				Path:   "/",
			}
			session.Jar.SetCookies(baseURL, []*http.Cookie{cookie})
		}
	}
	
	return nil
}

// ClearAllCookies clears all cookies from all sessions
func (sm *SessionManager) ClearAllCookies() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	baseURL, err := url.Parse(sm.baseURL)
	if err != nil {
		return fmt.Errorf("failed to parse base URL: %w", err)
	}
	
	for _, session := range sm.sessions {
		if session.Jar != nil {
			// Get all cookies and clear them
			cookies := session.Jar.Cookies(baseURL)
			for _, cookie := range cookies {
				cookie.MaxAge = -1 // Expire immediately
			}
			session.Jar.SetCookies(baseURL, cookies)
		}
	}
	
	return nil
}

// GetSessionStats returns statistics about session usage
func (sm *SessionManager) GetSessionStats() map[string]interface{} {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	return map[string]interface{}{
		"total_sessions": len(sm.sessions),
		"current_session": sm.current,
		"base_url": sm.baseURL,
	}
}
