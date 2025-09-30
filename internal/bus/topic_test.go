package bus

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTopicBuilder_BuildBarsTopic(t *testing.T) {
	builder := NewTopicBuilder("prod", "ampy")

	key := &Key{
		Symbol: "AAPL",
		MIC:    "XNAS",
	}

	topic := builder.BuildBarsTopic(key, "v1")
	assert.Equal(t, "ampy.prod.bars.v1.XNAS.AAPL", topic)
}

func TestTopicBuilder_BuildQuotesTopic(t *testing.T) {
	builder := NewTopicBuilder("dev", "ampy")

	key := &Key{
		Symbol: "MSFT",
		MIC:    "XNAS",
	}

	topic := builder.BuildQuotesTopic(key, "v1")
	assert.Equal(t, "ampy.dev.ticks.v1.XNAS.MSFT", topic)
}

func TestTopicBuilder_BuildFundamentalsTopic(t *testing.T) {
	builder := NewTopicBuilder("staging", "ampy")

	key := &Key{
		Symbol: "GOOGL",
		MIC:    "XNAS",
	}

	topic := builder.BuildFundamentalsTopic(key, "v1")
	assert.Equal(t, "ampy.staging.fundamentals.v1.GOOGL", topic)
}

func TestTopicBuilder_BuildSubtopic(t *testing.T) {
	builder := NewTopicBuilder("prod", "ampy")

	tests := []struct {
		name     string
		key      *Key
		expected string
	}{
		{
			name: "with MIC",
			key: &Key{
				Symbol: "AAPL",
				MIC:    "XNAS",
			},
			expected: "XNAS.AAPL",
		},
		{
			name: "without MIC",
			key: &Key{
				Symbol: "AAPL",
				MIC:    "",
			},
			expected: "AAPL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subtopic := builder.buildSubtopic(tt.key)
			assert.Equal(t, tt.expected, subtopic)
		})
	}
}

func TestValidateTopic(t *testing.T) {
	tests := []struct {
		name      string
		topic     string
		wantError bool
	}{
		{
			name:      "valid bars topic",
			topic:     "ampy.prod.bars.v1.XNAS.AAPL",
			wantError: false,
		},
		{
			name:      "valid quotes topic",
			topic:     "ampy.dev.ticks.v1.XNAS.MSFT",
			wantError: false,
		},
		{
			name:      "valid fundamentals topic",
			topic:     "ampy.staging.fundamentals.v1.GOOGL",
			wantError: false,
		},
		{
			name:      "empty topic",
			topic:     "",
			wantError: true,
		},
		{
			name:      "too few parts",
			topic:     "ampy.prod.bars",
			wantError: true,
		},
		{
			name:      "invalid domain",
			topic:     "ampy.prod.invalid.v1.XNAS.AAPL",
			wantError: true,
		},
		{
			name:      "invalid version",
			topic:     "ampy.prod.bars.invalid.XNAS.AAPL",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTopic(tt.topic)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParseTopic(t *testing.T) {
	tests := []struct {
		name      string
		topic     string
		expected  *TopicComponents
		wantError bool
	}{
		{
			name:  "valid bars topic",
			topic: "ampy.prod.bars.v1.XNAS.AAPL",
			expected: &TopicComponents{
				Prefix:   "ampy",
				Env:      "prod",
				Domain:   "bars",
				Version:  "v1",
				Subtopic: "XNAS.AAPL",
			},
			wantError: false,
		},
		{
			name:  "valid fundamentals topic",
			topic: "ampy.dev.fundamentals.v1.GOOGL",
			expected: &TopicComponents{
				Prefix:   "ampy",
				Env:      "dev",
				Domain:   "fundamentals",
				Version:  "v1",
				Subtopic: "GOOGL",
			},
			wantError: false,
		},
		{
			name:      "invalid topic",
			topic:     "invalid.topic",
			expected:  nil,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			components, err := ParseTopic(tt.topic)
			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, components)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, components)
				assert.Equal(t, tt.expected, components)
			}
		})
	}
}

func TestTopicComponents_String(t *testing.T) {
	tests := []struct {
		name       string
		components *TopicComponents
		expected   string
	}{
		{
			name: "with subtopic",
			components: &TopicComponents{
				Prefix:   "ampy",
				Env:      "prod",
				Domain:   "bars",
				Version:  "v1",
				Subtopic: "XNAS.AAPL",
			},
			expected: "ampy.prod.bars.v1.XNAS.AAPL",
		},
		{
			name: "without subtopic",
			components: &TopicComponents{
				Prefix:  "ampy",
				Env:     "dev",
				Domain:  "fundamentals",
				Version: "v1",
			},
			expected: "ampy.dev.fundamentals.v1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.components.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}
