package soak

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// MemoryMonitor tracks memory usage and goroutine counts during soak testing
type MemoryMonitor struct {
	logger            *zap.Logger
	initialMemory     uint64
	peakMemory        uint64
	initialGoroutines int
	peakGoroutines    int
	samples           []MemorySample
	mu                sync.RWMutex
}

// MemorySample represents a memory usage sample at a point in time
type MemorySample struct {
	Timestamp   time.Time
	AllocBytes  uint64
	SysBytes    uint64
	Goroutines  int
	GCCycles    uint32
	HeapObjects uint64
}

// LeakDetectionResult contains the results of leak detection analysis
type LeakDetectionResult struct {
	MemoryLeakDetected    bool
	GoroutineLeakDetected bool
	MemoryGrowthRate      float64 // bytes per second
	GoroutineGrowthRate   float64 // goroutines per second
	Recommendation        string
	Details               []string
}

// NewMemoryMonitor creates a new memory monitor
func NewMemoryMonitor(logger *zap.Logger) *MemoryMonitor {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return &MemoryMonitor{
		logger:            logger,
		initialMemory:     m.Alloc,
		peakMemory:        m.Alloc,
		initialGoroutines: runtime.NumGoroutine(),
		peakGoroutines:    runtime.NumGoroutine(),
		samples:           make([]MemorySample, 0, 1000), // Pre-allocate for efficiency
	}
}

// Monitor starts monitoring memory usage and goroutine counts
func (mm *MemoryMonitor) Monitor(ctx context.Context, wg *sync.WaitGroup, stopCh <-chan struct{}) {
	defer wg.Done()

	ticker := time.NewTicker(10 * time.Second) // Sample every 10 seconds
	defer ticker.Stop()

	mm.logger.Info("Memory monitor started")
	defer mm.logger.Info("Memory monitor stopped")

	for {
		select {
		case <-ticker.C:
			mm.takeSample()
		case <-stopCh:
			mm.takeFinalSample()
			return
		case <-ctx.Done():
			mm.takeFinalSample()
			return
		}
	}
}

// takeSample captures current memory and goroutine state
func (mm *MemoryMonitor) takeSample() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	goroutines := runtime.NumGoroutine()

	sample := MemorySample{
		Timestamp:   time.Now(),
		AllocBytes:  m.Alloc,
		SysBytes:    m.Sys,
		Goroutines:  goroutines,
		GCCycles:    m.NumGC,
		HeapObjects: m.HeapObjects,
	}

	mm.mu.Lock()
	mm.samples = append(mm.samples, sample)

	// Update peak values
	if m.Alloc > mm.peakMemory {
		mm.peakMemory = m.Alloc
	}
	if goroutines > mm.peakGoroutines {
		mm.peakGoroutines = goroutines
	}

	// Limit sample history to prevent unbounded growth
	if len(mm.samples) > 1000 {
		// Keep last 1000 samples
		copy(mm.samples, mm.samples[len(mm.samples)-1000:])
		mm.samples = mm.samples[:1000]
	}
	mm.mu.Unlock()

	// Log periodic updates
	if len(mm.samples)%30 == 0 { // Every 5 minutes (30 * 10 seconds)
		mm.logger.Info("Memory monitor sample",
			zap.Uint64("alloc_mb", m.Alloc/1024/1024),
			zap.Uint64("sys_mb", m.Sys/1024/1024),
			zap.Int("goroutines", goroutines),
			zap.Uint32("gc_cycles", m.NumGC),
			zap.Uint64("heap_objects", m.HeapObjects),
		)
	}
}

// takeFinalSample captures the final state for analysis
func (mm *MemoryMonitor) takeFinalSample() {
	mm.takeSample()
	mm.logger.Info("Final memory state captured", zap.Int("total_samples", len(mm.samples)))
}

// AnalyzeLeaks analyzes the collected samples for memory and goroutine leaks
func (mm *MemoryMonitor) AnalyzeLeaks() *LeakDetectionResult {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	if len(mm.samples) < 10 {
		return &LeakDetectionResult{
			Recommendation: "Insufficient samples for leak analysis",
			Details:        []string{"Need at least 10 samples for meaningful analysis"},
		}
	}

	result := &LeakDetectionResult{
		Details: make([]string, 0),
	}

	// Analyze memory growth
	memoryGrowth := mm.analyzeMemoryGrowth()
	result.MemoryGrowthRate = memoryGrowth.growthRate
	result.MemoryLeakDetected = memoryGrowth.leakDetected
	result.Details = append(result.Details, memoryGrowth.details...)

	// Analyze goroutine growth
	goroutineGrowth := mm.analyzeGoroutineGrowth()
	result.GoroutineGrowthRate = goroutineGrowth.growthRate
	result.GoroutineLeakDetected = goroutineGrowth.leakDetected
	result.Details = append(result.Details, goroutineGrowth.details...)

	// Generate recommendation
	result.Recommendation = mm.generateRecommendation(result)

	return result
}

// memoryGrowthAnalysis holds memory growth analysis results
type memoryGrowthAnalysis struct {
	growthRate   float64
	leakDetected bool
	details      []string
}

// analyzeMemoryGrowth analyzes memory usage patterns
func (mm *MemoryMonitor) analyzeMemoryGrowth() memoryGrowthAnalysis {
	if len(mm.samples) < 2 {
		return memoryGrowthAnalysis{
			details: []string{"Insufficient samples for memory analysis"},
		}
	}

	// Calculate linear regression to detect sustained growth
	n := len(mm.samples)
	startSample := mm.samples[0]
	endSample := mm.samples[n-1]

	duration := endSample.Timestamp.Sub(startSample.Timestamp).Seconds()
	if duration <= 0 {
		return memoryGrowthAnalysis{
			details: []string{"Invalid time duration for analysis"},
		}
	}

	// Calculate growth rate (bytes per second)
	memoryDiff := int64(endSample.AllocBytes) - int64(startSample.AllocBytes)
	growthRate := float64(memoryDiff) / duration

	analysis := memoryGrowthAnalysis{
		growthRate: growthRate,
		details:    make([]string, 0),
	}

	// Detect memory leak patterns
	const leakThreshold = 1024 * 1024 // 1MB per second growth is concerning

	if growthRate > leakThreshold {
		analysis.leakDetected = true
		analysis.details = append(analysis.details,
			fmt.Sprintf("Sustained memory growth detected: %.2f MB/sec", growthRate/1024/1024))
	}

	// Check for memory spikes
	maxSpike := mm.detectMemorySpikes()
	if maxSpike > 100*1024*1024 { // 100MB spike
		analysis.details = append(analysis.details,
			fmt.Sprintf("Large memory spike detected: %.2f MB", float64(maxSpike)/1024/1024))
	}

	// Analyze GC behavior
	gcAnalysis := mm.analyzeGCBehavior()
	analysis.details = append(analysis.details, gcAnalysis...)

	return analysis
}

// goroutineGrowthAnalysis holds goroutine growth analysis results
type goroutineGrowthAnalysis struct {
	growthRate   float64
	leakDetected bool
	details      []string
}

// analyzeGoroutineGrowth analyzes goroutine count patterns
func (mm *MemoryMonitor) analyzeGoroutineGrowth() goroutineGrowthAnalysis {
	if len(mm.samples) < 2 {
		return goroutineGrowthAnalysis{
			details: []string{"Insufficient samples for goroutine analysis"},
		}
	}

	n := len(mm.samples)
	startSample := mm.samples[0]
	endSample := mm.samples[n-1]

	duration := endSample.Timestamp.Sub(startSample.Timestamp).Seconds()
	if duration <= 0 {
		return goroutineGrowthAnalysis{
			details: []string{"Invalid time duration for analysis"},
		}
	}

	// Calculate growth rate (goroutines per second)
	goroutineDiff := endSample.Goroutines - startSample.Goroutines
	growthRate := float64(goroutineDiff) / duration

	analysis := goroutineGrowthAnalysis{
		growthRate: growthRate,
		details:    make([]string, 0),
	}

	// Detect goroutine leak patterns
	const leakThreshold = 0.1 // 0.1 goroutines per second sustained growth

	if growthRate > leakThreshold {
		analysis.leakDetected = true
		analysis.details = append(analysis.details,
			fmt.Sprintf("Sustained goroutine growth detected: %.3f goroutines/sec", growthRate))
	}

	// Check for goroutine spikes
	maxGoroutines := mm.peakGoroutines
	initialGoroutines := mm.initialGoroutines

	if maxGoroutines > initialGoroutines*2 { // More than 2x initial count
		analysis.details = append(analysis.details,
			fmt.Sprintf("Goroutine spike detected: %d -> %d (%.1fx increase)",
				initialGoroutines, maxGoroutines, float64(maxGoroutines)/float64(initialGoroutines)))
	}

	return analysis
}

// detectMemorySpikes finds the largest memory allocation spike
func (mm *MemoryMonitor) detectMemorySpikes() uint64 {
	if len(mm.samples) < 2 {
		return 0
	}

	var maxSpike uint64
	for i := 1; i < len(mm.samples); i++ {
		if mm.samples[i].AllocBytes > mm.samples[i-1].AllocBytes {
			spike := mm.samples[i].AllocBytes - mm.samples[i-1].AllocBytes
			if spike > maxSpike {
				maxSpike = spike
			}
		}
	}

	return maxSpike
}

// analyzeGCBehavior analyzes garbage collection patterns
func (mm *MemoryMonitor) analyzeGCBehavior() []string {
	if len(mm.samples) < 2 {
		return []string{"Insufficient samples for GC analysis"}
	}

	var details []string

	startGC := mm.samples[0].GCCycles
	endGC := mm.samples[len(mm.samples)-1].GCCycles
	duration := mm.samples[len(mm.samples)-1].Timestamp.Sub(mm.samples[0].Timestamp)

	gcRate := float64(endGC-startGC) / duration.Seconds()

	if gcRate > 1.0 { // More than 1 GC per second on average
		details = append(details, fmt.Sprintf("High GC frequency: %.2f cycles/sec", gcRate))
	}

	// Check for GC pressure indicators
	if len(mm.samples) > 10 {
		recentSamples := mm.samples[len(mm.samples)-10:]
		avgHeapObjects := uint64(0)
		for _, sample := range recentSamples {
			avgHeapObjects += sample.HeapObjects
		}
		avgHeapObjects /= uint64(len(recentSamples))

		if avgHeapObjects > 1000000 { // More than 1M heap objects
			details = append(details, fmt.Sprintf("High heap object count: %d objects", avgHeapObjects))
		}
	}

	return details
}

// generateRecommendation generates actionable recommendations based on analysis
func (mm *MemoryMonitor) generateRecommendation(result *LeakDetectionResult) string {
	if !result.MemoryLeakDetected && !result.GoroutineLeakDetected {
		return "✅ No significant leaks detected. Memory and goroutine usage appear stable."
	}

	recommendations := []string{}

	if result.MemoryLeakDetected {
		recommendations = append(recommendations,
			"🔍 Memory leak detected - investigate object retention and ensure proper cleanup")
	}

	if result.GoroutineLeakDetected {
		recommendations = append(recommendations,
			"🔍 Goroutine leak detected - check for unclosed channels and missing context cancellation")
	}

	if result.MemoryGrowthRate > 10*1024*1024 { // > 10MB/sec
		recommendations = append(recommendations,
			"⚠️  High memory growth rate - consider implementing memory pooling or reducing allocations")
	}

	if result.GoroutineGrowthRate > 1.0 { // > 1 goroutine/sec
		recommendations = append(recommendations,
			"⚠️  High goroutine growth rate - review goroutine lifecycle management")
	}

	if len(recommendations) == 0 {
		return "⚠️  Potential issues detected - review detailed analysis"
	}

	return strings.Join(recommendations, "; ")
}

// GetCurrentStats returns current memory and goroutine statistics
func (mm *MemoryMonitor) GetCurrentStats() MemorySample {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return MemorySample{
		Timestamp:   time.Now(),
		AllocBytes:  m.Alloc,
		SysBytes:    m.Sys,
		Goroutines:  runtime.NumGoroutine(),
		GCCycles:    m.NumGC,
		HeapObjects: m.HeapObjects,
	}
}

// GetSampleHistory returns a copy of the sample history
func (mm *MemoryMonitor) GetSampleHistory() []MemorySample {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	samples := make([]MemorySample, len(mm.samples))
	copy(samples, mm.samples)
	return samples
}
