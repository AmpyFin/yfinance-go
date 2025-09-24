package bus

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvelopeBuilder_BuildEnvelope(t *testing.T) {
	builder := NewEnvelopeBuilder("test-producer", "test-source")
	
	key := &Key{
		Symbol: "AAPL",
		MIC:    "XNAS",
	}
	
	envelope := builder.BuildEnvelope(
		"ampy.bars.v1.BarBatch",
		"1.0.0",
		key,
		"test-run-id",
		"test-trace-id",
		nil,
	)
	
	// Validate envelope
	require.NotNil(t, envelope)
	assert.NotEmpty(t, envelope.MessageID)
	assert.Equal(t, "ampy.bars.v1.BarBatch", envelope.SchemaFQDN)
	assert.Equal(t, "1.0.0", envelope.SchemaVersion)
	assert.Equal(t, "application/x-protobuf", envelope.ContentType)
	assert.Equal(t, "test-producer", envelope.Producer)
	assert.Equal(t, "test-source", envelope.Source)
	assert.Equal(t, "test-run-id", envelope.RunID)
	assert.Equal(t, "test-trace-id", envelope.TraceID)
	assert.Equal(t, "XNAS.AAPL", envelope.PartitionKey)
	
	// Validate UUID format
	_, err := uuid.Parse(envelope.MessageID)
	assert.NoError(t, err)
	
	// Validate timestamp
	assert.WithinDuration(t, time.Now(), envelope.ProducedAt, time.Second)
}

func TestEnvelopeBuilder_BuildChunkedEnvelope(t *testing.T) {
	builder := NewEnvelopeBuilder("test-producer", "test-source")
	
	key := &Key{
		Symbol: "AAPL",
		MIC:    "XNAS",
	}
	
	envelope := builder.BuildChunkedEnvelope(
		"ampy.bars.v1.BarBatch",
		"1.0.0",
		key,
		"test-run-id",
		"test-trace-id",
		1,
		3,
		nil,
	)
	
	// Validate envelope
	require.NotNil(t, envelope)
	assert.Equal(t, "1", envelope.Extensions["chunk_index"])
	assert.Equal(t, "3", envelope.Extensions["total_chunks"])
}

func TestValidateEnvelope(t *testing.T) {
	tests := []struct {
		name      string
		envelope  *Envelope
		wantError bool
	}{
		{
			name:      "valid envelope",
			envelope:  createValidEnvelope(),
			wantError: false,
		},
		{
			name:      "nil envelope",
			envelope:  nil,
			wantError: true,
		},
		{
			name: "missing message_id",
			envelope: &Envelope{
				SchemaFQDN:    "ampy.bars.v1.BarBatch",
				SchemaVersion: "1.0.0",
				Producer:      "test-producer",
				Source:        "test-source",
				RunID:         "test-run-id",
				PartitionKey:  "XNAS.AAPL",
			},
			wantError: true,
		},
		{
			name: "invalid message_id format",
			envelope: &Envelope{
				MessageID:     "invalid-uuid",
				SchemaFQDN:    "ampy.bars.v1.BarBatch",
				SchemaVersion: "1.0.0",
				Producer:      "test-producer",
				Source:        "test-source",
				RunID:         "test-run-id",
				PartitionKey:  "XNAS.AAPL",
			},
			wantError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEnvelope(tt.envelope)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestComputeSchemaHash(t *testing.T) {
	hash := ComputeSchemaHash("ampy.bars.v1.BarBatch")
	assert.NotEmpty(t, hash)
	assert.Len(t, hash, 16) // 8 bytes = 16 hex chars
}

func createValidEnvelope() *Envelope {
	return &Envelope{
		MessageID:     uuid.Must(uuid.NewV7()).String(),
		SchemaFQDN:    "ampy.bars.v1.BarBatch",
		SchemaVersion: "1.0.0",
		ContentType:   "application/x-protobuf",
		ProducedAt:    time.Now().UTC(),
		Producer:      "test-producer",
		Source:        "test-source",
		RunID:         "test-run-id",
		PartitionKey:  "XNAS.AAPL",
	}
}
