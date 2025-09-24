package bus

import (
	"fmt"
	"strings"
)

// TopicBuilder builds ampy-bus topics
type TopicBuilder struct {
	env         string
	topicPrefix string
}

// NewTopicBuilder creates a new topic builder
func NewTopicBuilder(env, topicPrefix string) *TopicBuilder {
	return &TopicBuilder{
		env:         env,
		topicPrefix: topicPrefix,
	}
}

// BuildBarsTopic builds a topic for bars data
func (b *TopicBuilder) BuildBarsTopic(key *Key, version string) string {
	subtopic := b.buildSubtopic(key)
	return fmt.Sprintf("%s.%s.bars.%s.%s", b.topicPrefix, b.env, version, subtopic)
}

// BuildQuotesTopic builds a topic for quotes data
func (b *TopicBuilder) BuildQuotesTopic(key *Key, version string) string {
	subtopic := b.buildSubtopic(key)
	return fmt.Sprintf("%s.%s.ticks.%s.%s", b.topicPrefix, b.env, version, subtopic)
}

// BuildFundamentalsTopic builds a topic for fundamentals data
func (b *TopicBuilder) BuildFundamentalsTopic(key *Key, version string) string {
	// For fundamentals, we typically use just the symbol as subtopic
	// since MIC might not be available or relevant
	subtopic := key.Symbol
	return fmt.Sprintf("%s.%s.fundamentals.%s.%s", b.topicPrefix, b.env, version, subtopic)
}

// buildSubtopic builds the subtopic portion of the topic
func (b *TopicBuilder) buildSubtopic(key *Key) string {
	if key.MIC == "" {
		return key.Symbol
	}
	return key.MIC + "." + key.Symbol
}

// ValidateTopic validates a topic format
func ValidateTopic(topic string) error {
	if topic == "" {
		return fmt.Errorf("topic cannot be empty")
	}
	
	parts := strings.Split(topic, ".")
	if len(parts) < 4 {
		return fmt.Errorf("topic must have at least 4 parts: prefix.env.domain.version.subtopic")
	}
	
	// Validate environment
	env := parts[1]
	if env == "" {
		return fmt.Errorf("environment cannot be empty")
	}
	
	// Validate domain
	domain := parts[2]
	validDomains := map[string]bool{
		"bars":         true,
		"ticks":        true,
		"fundamentals": true,
		"news":         true,
		"fx":           true,
		"signals":      true,
		"orders":       true,
		"fills":        true,
		"positions":    true,
		"metrics":      true,
		"dlq":          true,
		"control":      true,
	}
	
	if !validDomains[domain] {
		return fmt.Errorf("invalid domain: %s", domain)
	}
	
	// Validate version
	version := parts[3]
	if !strings.HasPrefix(version, "v") {
		return fmt.Errorf("version must start with 'v': %s", version)
	}
	
	return nil
}

// ParseTopic parses a topic into its components
func ParseTopic(topic string) (*TopicComponents, error) {
	if err := ValidateTopic(topic); err != nil {
		return nil, err
	}
	
	parts := strings.Split(topic, ".")
	
	components := &TopicComponents{
		Prefix:    parts[0],
		Env:       parts[1],
		Domain:    parts[2],
		Version:   parts[3],
		Subtopic:  strings.Join(parts[4:], "."),
	}
	
	return components, nil
}

// TopicComponents represents the parsed components of a topic
type TopicComponents struct {
	Prefix   string
	Env      string
	Domain   string
	Version  string
	Subtopic string
}

// String returns the topic as a string
func (tc *TopicComponents) String() string {
	if tc.Subtopic == "" {
		return fmt.Sprintf("%s.%s.%s.%s", tc.Prefix, tc.Env, tc.Domain, tc.Version)
	}
	return fmt.Sprintf("%s.%s.%s.%s.%s", tc.Prefix, tc.Env, tc.Domain, tc.Version, tc.Subtopic)
}
