package scrape

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// RobotsManager handles robots.txt fetching, caching, and policy enforcement
type RobotsManager struct {
	policy RobotsPolicy
	ttl    time.Duration
	cache  map[string]*RobotsCache
	mu     sync.RWMutex
	client *http.Client
}

// NewRobotsManager creates a new robots manager
func NewRobotsManager(policy string, ttl time.Duration) *RobotsManager {
	if !IsValidRobotsPolicy(policy) {
		policy = string(RobotsEnforce)
	}

	return &RobotsManager{
		policy: RobotsPolicy(policy),
		ttl:    ttl,
		cache:  make(map[string]*RobotsCache),
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// CheckRobots checks if a path is allowed by robots.txt
func (rm *RobotsManager) CheckRobots(ctx context.Context, host, path string) error {
	// Skip check if policy is ignore
	if rm.policy == RobotsIgnore {
		return nil
	}

	// Get robots.txt for the host
	robots, err := rm.getRobots(ctx, host)
	if err != nil {
		// If we can't fetch robots.txt, warn but continue if policy is warn
		if rm.policy == RobotsWarn {
			// Log warning but don't block
			return nil
		}
		// If policy is enforce, block on robots.txt fetch failure
		return &ScrapeError{
			Type:    "robots_fetch_failed",
			Message: fmt.Sprintf("failed to fetch robots.txt: %v", err),
			URL:     fmt.Sprintf("https://%s/robots.txt", host),
		}
	}

	// Check if path is allowed
	if !rm.isPathAllowed(robots, path) {
		err := &ScrapeError{
			Type:    "robots_denied",
			Message: fmt.Sprintf("robots.txt disallows path: %s", path),
			URL:     fmt.Sprintf("https://%s%s", host, path),
		}
		
		if rm.policy == RobotsWarn {
			// Log warning but don't block
			return nil
		}
		
		return err
	}

	return nil
}

// getRobots fetches and caches robots.txt for a host
func (rm *RobotsManager) getRobots(ctx context.Context, host string) (*RobotsCache, error) {
	rm.mu.RLock()
	cached, exists := rm.cache[host]
	rm.mu.RUnlock()

	// Return cached version if not expired
	if exists && !cached.IsExpired() {
		return cached, nil
	}

	// Fetch fresh robots.txt
	robotsURL := fmt.Sprintf("https://%s/robots.txt", host)
	req, err := http.NewRequestWithContext(ctx, "GET", robotsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create robots.txt request: %w", err)
	}

	// Set appropriate headers for robots.txt
	req.Header.Set("User-Agent", "Mozilla/5.0 (Ampy yfinance-go scraper)")
	req.Header.Set("Accept", "text/plain")

	resp, err := rm.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch robots.txt: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("robots.txt returned status %d", resp.StatusCode)
	}

	// Parse robots.txt
	rules, err := rm.parseRobotsTxt(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse robots.txt: %w", err)
	}

	// Cache the result
	robots := &RobotsCache{
		Host:      host,
		Rules:     rules,
		FetchedAt: time.Now(),
		TTL:       rm.ttl,
	}

	rm.mu.Lock()
	rm.cache[host] = robots
	rm.mu.Unlock()

	return robots, nil
}

// parseRobotsTxt parses robots.txt content
func (rm *RobotsManager) parseRobotsTxt(body interface{ Read([]byte) (int, error) }) ([]RobotsRule, error) {
	var rules []RobotsRule
	var currentRule *RobotsRule

	scanner := bufio.NewScanner(body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse directive
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		directive := strings.ToLower(strings.TrimSpace(parts[0]))
		value := strings.TrimSpace(parts[1])

		switch directive {
		case "user-agent":
			// Start new rule
			if currentRule != nil {
				rules = append(rules, *currentRule)
			}
			currentRule = &RobotsRule{
				UserAgent: value,
				Allow:     []string{},
				Disallow:  []string{},
			}
		case "allow":
			if currentRule != nil {
				currentRule.Allow = append(currentRule.Allow, value)
			}
		case "disallow":
			if currentRule != nil {
				currentRule.Disallow = append(currentRule.Disallow, value)
			}
		}
	}

	// Add the last rule
	if currentRule != nil {
		rules = append(rules, *currentRule)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read robots.txt: %w", err)
	}

	return rules, nil
}

// isPathAllowed checks if a path is allowed by robots.txt rules
func (rm *RobotsManager) isPathAllowed(robots *RobotsCache, path string) bool {
	// Find applicable rules (look for wildcard or our user agent)
	var applicableRules []RobotsRule
	
	for _, rule := range robots.Rules {
		if rule.UserAgent == "*" || strings.Contains(strings.ToLower(rule.UserAgent), "ampy") {
			applicableRules = append(applicableRules, rule)
		}
	}

	// If no specific rules found, use wildcard rules
	if len(applicableRules) == 0 {
		for _, rule := range robots.Rules {
			if rule.UserAgent == "*" {
				applicableRules = append(applicableRules, rule)
			}
		}
	}

	// If still no rules, allow by default
	if len(applicableRules) == 0 {
		return true
	}

	// Check each applicable rule
	for _, rule := range applicableRules {
		// Check disallow rules first
		for _, disallow := range rule.Disallow {
			if rm.pathMatches(path, disallow) {
				// Check if there's a more specific allow rule
				allowed := false
				for _, allow := range rule.Allow {
					if rm.pathMatches(path, allow) && len(allow) > len(disallow) {
						allowed = true
						break
					}
				}
				if !allowed {
					return false
				}
			}
		}
	}

	return true
}

// pathMatches checks if a path matches a robots.txt pattern
func (rm *RobotsManager) pathMatches(path, pattern string) bool {
	// Handle empty pattern (disallow nothing)
	if pattern == "" {
		return false
	}

	// Handle wildcard patterns
	if strings.Contains(pattern, "*") {
		// Simple wildcard matching
		pattern = strings.ReplaceAll(pattern, "*", ".*")
		matched := url.PathEscape(path)
		return strings.Contains(matched, pattern)
	}

	// Exact prefix matching
	return strings.HasPrefix(path, pattern)
}

// ClearCache clears the robots.txt cache
func (rm *RobotsManager) ClearCache() {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.cache = make(map[string]*RobotsCache)
}

// GetCacheStats returns cache statistics
func (rm *RobotsManager) GetCacheStats() map[string]interface{} {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	stats := map[string]interface{}{
		"cached_hosts": len(rm.cache),
		"policy":       string(rm.policy),
		"ttl_seconds":  rm.ttl.Seconds(),
	}

	// Count expired entries
	expired := 0
	now := time.Now()
	for _, cached := range rm.cache {
		if now.Sub(cached.FetchedAt) > cached.TTL {
			expired++
		}
	}
	stats["expired_entries"] = expired

	return stats
}
