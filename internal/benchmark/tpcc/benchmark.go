package tpcc

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/deadjoe/benchphant/internal/benchmark"
	"github.com/deadjoe/benchphant/internal/models"
	"go.uber.org/zap"
)

// TPCCBenchmark implements the BenchmarkRunner interface for TPC-C
type TPCCBenchmark struct {
	config *Config
	db     *sql.DB
	logger *zap.Logger
	runner *Runner
	status benchmark.BenchmarkStatus
	mu     sync.RWMutex
}

// NewTPCCBenchmark creates a new TPC-C benchmark instance
func NewTPCCBenchmark(config *Config, db *sql.DB, logger *zap.Logger) *TPCCBenchmark {
	return &TPCCBenchmark{
		config: config,
		db:     db,
		logger: logger,
		status: benchmark.BenchmarkStatus{
			Status:   string(models.BenchmarkStatusPending),
			Progress: 0,
			Metrics:  make(map[string]interface{}),
		},
	}
}

// Name returns the name of the benchmark
func (b *TPCCBenchmark) Name() string {
	return "tpcc"
}

// Setup prepares the benchmark environment
func (b *TPCCBenchmark) Setup(ctx context.Context) error {
	b.logger.Info("Setting up TPC-C benchmark",
		zap.Int("warehouses", b.config.Warehouses),
		zap.Int("terminals", b.config.Terminals),
		zap.Duration("duration", b.config.Duration),
	)

	// Create schema
	if err := CreateSchema(ctx, b.db); err != nil {
		return fmt.Errorf("create schema: %w", err)
	}

	// Load initial data
	loader := NewLoader(b.db, b.config)
	if err := loader.Load(ctx); err != nil {
		return fmt.Errorf("load data: %w", err)
	}

	// Create runner
	b.runner = NewRunner(b.db, b.config, b.logger)

	return nil
}

// Run executes the benchmark
func (b *TPCCBenchmark) Run(ctx context.Context) (*benchmark.Result, error) {
	if err := b.Setup(ctx); err != nil {
		return nil, fmt.Errorf("setup: %w", err)
	}

	b.logger.Info("Starting TPC-C benchmark",
		zap.Int("warehouses", b.config.Warehouses),
		zap.Int("terminals", b.config.Terminals),
		zap.Duration("duration", b.config.Duration),
	)

	startTime := time.Now()
	stats, err := b.runner.Run(ctx)
	if err != nil {
		return nil, fmt.Errorf("run: %w", err)
	}
	endTime := time.Now()

	result := &benchmark.Result{
		Name:              "TPC-C",
		Duration:          endTime.Sub(startTime),
		TotalTransactions: stats.TotalTransactions,
		TPS:              float64(stats.TotalTransactions) / endTime.Sub(startTime).Seconds(),
		LatencyAvg:       stats.LatencyAvg,
		LatencyP95:       stats.LatencyP95,
		LatencyP99:       stats.LatencyP99,
		Errors:           stats.Errors,
		StartTime:        startTime,
		EndTime:          endTime,
		Metrics:          make(map[string]interface{}),
	}

	// Convert metrics to interface{} map
	for k, v := range stats.Metrics {
		result.Metrics[k] = interface{}(v)
	}

	return result, nil
}

// GetStats returns the benchmark statistics
func (b *TPCCBenchmark) GetStats() *benchmark.Result {
	if b.runner == nil {
		return &benchmark.Result{
			Name:     "TPC-C",
			Duration: time.Duration(0),
			Metrics:  make(map[string]interface{}),
		}
	}

	stats := b.runner.GetStats()
	result := &benchmark.Result{
		Name:              "TPC-C",
		Duration:          time.Since(b.runner.startTime),
		TotalTransactions: stats.TotalTransactions,
		TPS:              float64(stats.TotalTransactions) / time.Since(b.runner.startTime).Seconds(),
		LatencyAvg:       stats.LatencyAvg,
		LatencyP95:       stats.LatencyP95,
		LatencyP99:       stats.LatencyP99,
		Errors:           stats.Errors,
		StartTime:        b.runner.startTime,
		EndTime:          time.Now(),
		Metrics:          make(map[string]interface{}),
	}

	// Convert metrics to interface{} map
	for k, v := range stats.Metrics {
		result.Metrics[k] = interface{}(v)
	}

	return result
}

// Cleanup performs necessary cleanup after the benchmark
func (b *TPCCBenchmark) Cleanup(ctx context.Context) error {
	b.logger.Info("Cleaning up TPC-C benchmark")
	return DropSchema(ctx, b.db)
}

// Validate checks if the benchmark configuration is valid
func (b *TPCCBenchmark) Validate() error {
	if b.config == nil {
		return fmt.Errorf("config is nil")
	}
	if b.db == nil {
		return fmt.Errorf("database connection is nil")
	}
	if b.logger == nil {
		return fmt.Errorf("logger is nil")
	}
	return b.config.Validate()
}

// Start starts the benchmark
func (b *TPCCBenchmark) Start() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.status.Status == string(models.BenchmarkStatusRunning) {
		return fmt.Errorf("benchmark is already running")
	}

	b.status.Status = string(models.BenchmarkStatusRunning)
	b.status.Progress = 0
	b.status.Metrics = make(map[string]interface{})

	// Create runner
	b.runner = NewRunner(b.db, b.config, b.logger)

	// Run benchmark in a goroutine
	go func() {
		ctx := context.Background()
		if err := b.Setup(ctx); err != nil {
			b.logger.Error("Failed to setup benchmark", zap.Error(err))
			b.status.Status = string(models.BenchmarkStatusFailed)
			return
		}

		stats, err := b.runner.Run(ctx)
		if err != nil {
			b.logger.Error("Failed to run benchmark", zap.Error(err))
			b.status.Status = string(models.BenchmarkStatusFailed)
			return
		}

		b.mu.Lock()
		b.status.Status = string(models.BenchmarkStatusCompleted)
		b.status.Progress = 100
		b.status.Metrics = map[string]interface{}{
			"total_transactions": stats.TotalTransactions,
			"tps":               stats.TPS,
			"latency_avg":       stats.LatencyAvg,
			"latency_p95":       stats.LatencyP95,
			"latency_p99":       stats.LatencyP99,
			"errors":            stats.Errors,
		}
		b.mu.Unlock()
	}()

	return nil
}

// Stop stops the benchmark
func (b *TPCCBenchmark) Stop() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.runner != nil {
		b.runner.Stop()
	}
	b.status.Status = string(models.BenchmarkStatusCancelled)
}

// Status returns the current benchmark status
func (b *TPCCBenchmark) Status() benchmark.BenchmarkStatus {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.status
}
