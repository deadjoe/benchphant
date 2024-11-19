package models

import (
	"errors"
	"time"
)

// BenchmarkStatus represents the status of a benchmark
type BenchmarkStatus string

const (
	// BenchmarkStatusPending indicates the benchmark is pending
	BenchmarkStatusPending BenchmarkStatus = "pending"
	// BenchmarkStatusRunning indicates the benchmark is running
	BenchmarkStatusRunning BenchmarkStatus = "running"
	// BenchmarkStatusCompleted indicates the benchmark has completed successfully
	BenchmarkStatusCompleted BenchmarkStatus = "completed"
	// BenchmarkStatusFailed indicates the benchmark has failed
	BenchmarkStatusFailed BenchmarkStatus = "failed"
	// BenchmarkStatusCancelled indicates the benchmark was cancelled
	BenchmarkStatusCancelled BenchmarkStatus = "cancelled"
)

// Benchmark represents a database benchmark configuration
type Benchmark struct {
	ID            int64          `json:"id"`
	Name          string         `json:"name"`
	Description   string         `json:"description"`
	ConnectionID  int64          `json:"connection_id"`
	QueryTemplate string         `json:"query_template"`
	NumThreads    int           `json:"num_threads"`
	Duration      time.Duration  `json:"duration"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	Status        BenchmarkStatus `json:"status"`
}

// BenchmarkResult represents the result of a benchmark run
type BenchmarkResult struct {
	ID             int64         `json:"id"`
	BenchmarkID    int64         `json:"benchmark_id"`
	StartTime      time.Time     `json:"start_time"`
	EndTime        time.Time     `json:"end_time"`
	TotalQueries   int64         `json:"total_queries"`
	SuccessCount   int64         `json:"success_count"`
	FailureCount   int64         `json:"failure_count"`
	AverageLatency time.Duration `json:"average_latency"`
	MinLatency     time.Duration `json:"min_latency"`
	MaxLatency     time.Duration `json:"max_latency"`
	QPS            float64       `json:"qps"`
	Error          string        `json:"error,omitempty"`
}

// Validate validates the benchmark configuration
func (b *Benchmark) Validate() error {
	if b.Name == "" {
		return errors.New("name is required")
	}
	if b.ConnectionID <= 0 {
		return errors.New("invalid connection ID")
	}
	if b.QueryTemplate == "" {
		return errors.New("query template is required")
	}
	if b.NumThreads <= 0 {
		return errors.New("number of threads must be greater than 0")
	}
	if b.Duration <= 0 {
		return errors.New("duration must be greater than 0")
	}

	switch b.Status {
	case BenchmarkStatusPending, BenchmarkStatusRunning, BenchmarkStatusCompleted,
		BenchmarkStatusFailed, BenchmarkStatusCancelled:
		return nil
	default:
		return errors.New("invalid benchmark status")
	}
}
