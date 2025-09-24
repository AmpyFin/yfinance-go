package bus

import (
	"fmt"
	"math"
)

// ChunkingStrategy defines how to chunk large payloads
type ChunkingStrategy struct {
	MaxPayloadBytes int64
}

// NewChunkingStrategy creates a new chunking strategy
func NewChunkingStrategy(maxPayloadBytes int64) *ChunkingStrategy {
	return &ChunkingStrategy{
		MaxPayloadBytes: maxPayloadBytes,
	}
}

// ChunkResult represents the result of chunking
type ChunkResult struct {
	Chunks    [][]byte
	ChunkInfo []ChunkInfo
}

// ChunkInfo represents information about a chunk
type ChunkInfo struct {
	Index     int
	Size      int
	IsLast    bool
}

// ChunkPayload chunks a payload if it exceeds the size limit
func (cs *ChunkingStrategy) ChunkPayload(payload []byte) (*ChunkResult, error) {
	if len(payload) == 0 {
		return &ChunkResult{
			Chunks:    [][]byte{payload},
			ChunkInfo: []ChunkInfo{{Index: 0, Size: 0, IsLast: true}},
		}, nil
	}
	
	// If payload is within limits, return as single chunk
	if int64(len(payload)) <= cs.MaxPayloadBytes {
		return &ChunkResult{
			Chunks:    [][]byte{payload},
			ChunkInfo: []ChunkInfo{{Index: 0, Size: len(payload), IsLast: true}},
		}, nil
	}
	
	// Calculate number of chunks needed
	numChunks := int(math.Ceil(float64(len(payload)) / float64(cs.MaxPayloadBytes)))
	
	chunks := make([][]byte, numChunks)
	chunkInfo := make([]ChunkInfo, numChunks)
	
	for i := 0; i < numChunks; i++ {
		start := i * int(cs.MaxPayloadBytes)
		end := start + int(cs.MaxPayloadBytes)
		
		// Ensure we don't exceed the payload length
		if end > len(payload) {
			end = len(payload)
		}
		
		chunks[i] = payload[start:end]
		chunkInfo[i] = ChunkInfo{
			Index:  i,
			Size:   len(chunks[i]),
			IsLast: i == numChunks-1,
		}
	}
	
	return &ChunkResult{
		Chunks:    chunks,
		ChunkInfo: chunkInfo,
	}, nil
}

// EstimateChunkCount estimates the number of chunks needed for a payload
func (cs *ChunkingStrategy) EstimateChunkCount(payloadSize int) int {
	if payloadSize == 0 {
		return 1
	}
	
	if int64(payloadSize) <= cs.MaxPayloadBytes {
		return 1
	}
	
	return int(math.Ceil(float64(payloadSize) / float64(cs.MaxPayloadBytes)))
}

// ValidateChunkSize validates that a chunk is within size limits
func (cs *ChunkingStrategy) ValidateChunkSize(chunk []byte) error {
	if int64(len(chunk)) > cs.MaxPayloadBytes {
		return fmt.Errorf("chunk size %d exceeds maximum %d bytes", len(chunk), cs.MaxPayloadBytes)
	}
	return nil
}

// GetChunkingInfo returns chunking information for preview
func (cs *ChunkingStrategy) GetChunkingInfo(payloadSize int) ChunkingInfo {
	chunkCount := cs.EstimateChunkCount(payloadSize)
	
	chunkSizes := make([]int, chunkCount)
	if chunkCount == 1 {
		chunkSizes[0] = payloadSize
	} else {
		// Estimate chunk sizes
		baseSize := payloadSize / chunkCount
		remainder := payloadSize % chunkCount
		
		for i := 0; i < chunkCount; i++ {
			chunkSizes[i] = baseSize
			if i < remainder {
				chunkSizes[i]++
			}
		}
	}
	
	return ChunkingInfo{
		ChunkCount: chunkCount,
		MaxPayload: cs.MaxPayloadBytes,
		ChunkSizes: chunkSizes,
	}
}
