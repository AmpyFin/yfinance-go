package bus

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChunkingStrategy_ChunkPayload(t *testing.T) {
	strategy := NewChunkingStrategy(1000) // 1KB max
	
	tests := []struct {
		name           string
		payload        []byte
		expectedChunks int
	}{
		{
			name:           "empty payload",
			payload:        []byte{},
			expectedChunks: 1,
		},
		{
			name:           "small payload",
			payload:        make([]byte, 500),
			expectedChunks: 1,
		},
		{
			name:           "exact size payload",
			payload:        make([]byte, 1000),
			expectedChunks: 1,
		},
		{
			name:           "large payload",
			payload:        make([]byte, 2500),
			expectedChunks: 3,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := strategy.ChunkPayload(tt.payload)
			require.NoError(t, err)
			require.NotNil(t, result)
			
			assert.Equal(t, tt.expectedChunks, len(result.Chunks))
			assert.Equal(t, tt.expectedChunks, len(result.ChunkInfo))
			
			// Verify chunk sizes
			for i, chunk := range result.Chunks {
				assert.LessOrEqual(t, int64(len(chunk)), strategy.MaxPayloadBytes)
				assert.Equal(t, i, result.ChunkInfo[i].Index)
				assert.Equal(t, len(chunk), result.ChunkInfo[i].Size)
				assert.Equal(t, i == len(result.Chunks)-1, result.ChunkInfo[i].IsLast)
			}
		})
	}
}

func TestChunkingStrategy_EstimateChunkCount(t *testing.T) {
	strategy := NewChunkingStrategy(1000) // 1KB max
	
	tests := []struct {
		name           string
		payloadSize    int
		expectedChunks int
	}{
		{
			name:           "empty payload",
			payloadSize:    0,
			expectedChunks: 1,
		},
		{
			name:           "small payload",
			payloadSize:    500,
			expectedChunks: 1,
		},
		{
			name:           "exact size payload",
			payloadSize:    1000,
			expectedChunks: 1,
		},
		{
			name:           "large payload",
			payloadSize:    2500,
			expectedChunks: 3,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunkCount := strategy.EstimateChunkCount(tt.payloadSize)
			assert.Equal(t, tt.expectedChunks, chunkCount)
		})
	}
}

func TestChunkingStrategy_ValidateChunkSize(t *testing.T) {
	strategy := NewChunkingStrategy(1000) // 1KB max
	
	tests := []struct {
		name      string
		chunk     []byte
		wantError bool
	}{
		{
			name:      "valid chunk size",
			chunk:     make([]byte, 500),
			wantError: false,
		},
		{
			name:      "exact max size",
			chunk:     make([]byte, 1000),
			wantError: false,
		},
		{
			name:      "oversized chunk",
			chunk:     make([]byte, 1001),
			wantError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := strategy.ValidateChunkSize(tt.chunk)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestChunkingStrategy_GetChunkingInfo(t *testing.T) {
	strategy := NewChunkingStrategy(1000) // 1KB max
	
	tests := []struct {
		name           string
		payloadSize    int
		expectedChunks int
	}{
		{
			name:           "small payload",
			payloadSize:    500,
			expectedChunks: 1,
		},
		{
			name:           "large payload",
			payloadSize:    2500,
			expectedChunks: 3,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := strategy.GetChunkingInfo(tt.payloadSize)
			
			assert.Equal(t, tt.expectedChunks, info.ChunkCount)
			assert.Equal(t, strategy.MaxPayloadBytes, info.MaxPayload)
			assert.Equal(t, tt.expectedChunks, len(info.ChunkSizes))
			
			// Verify chunk sizes sum to payload size
			totalSize := 0
			for _, size := range info.ChunkSizes {
				totalSize += size
			}
			assert.Equal(t, tt.payloadSize, totalSize)
		})
	}
}
