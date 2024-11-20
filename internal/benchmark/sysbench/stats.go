package sysbench

import (
	"sort"
	"sync"
	"time"
)

// TestStats represents test statistics
type TestStats struct {
	mu                sync.RWMutex
	startTime         time.Time
	totalTransactions int64
	errors            int64
	latencies         []time.Duration
}

// Stats represents a snapshot of test statistics
type Stats struct {
	TPS               float64
	LatencyAvg        time.Duration
	LatencyP95        time.Duration
	LatencyP99        time.Duration
	TotalTransactions int64
	Errors            int64
}

// NewTestStats creates a new TestStats
func NewTestStats() *TestStats {
	return &TestStats{
		startTime: time.Now(),
		latencies: make([]time.Duration, 0, 1000),
	}
}

// AddTransaction adds a transaction to the statistics
func (s *TestStats) AddTransaction(latency time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.totalTransactions++
	s.latencies = append(s.latencies, latency)
}

// AddError adds an error to the statistics
func (s *TestStats) AddError() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.errors++
}

// GetStats returns a snapshot of the current statistics
func (s *TestStats) GetStats() Stats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	elapsed := time.Since(s.startTime).Seconds()
	if elapsed == 0 {
		elapsed = 1 // Avoid division by zero
	}

	stats := Stats{
		TPS:               float64(s.totalTransactions) / elapsed,
		TotalTransactions: s.totalTransactions,
		Errors:            s.errors,
	}

	if len(s.latencies) > 0 {
		// Calculate average latency
		var total time.Duration
		for _, lat := range s.latencies {
			total += lat
		}
		stats.LatencyAvg = total / time.Duration(len(s.latencies))

		// Sort latencies for percentiles
		sorted := make([]time.Duration, len(s.latencies))
		copy(sorted, s.latencies)
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i] < sorted[j]
		})

		// Calculate percentiles
		p95 := int(float64(len(sorted)) * 0.95)
		p99 := int(float64(len(sorted)) * 0.99)
		if p95 < len(sorted) {
			stats.LatencyP95 = sorted[p95]
		}
		if p99 < len(sorted) {
			stats.LatencyP99 = sorted[p99]
		}
	}

	return stats
}

// Reset resets the statistics
func (s *TestStats) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.startTime = time.Now()
	s.totalTransactions = 0
	s.errors = 0
	s.latencies = make([]time.Duration, 0, 1000)
}
