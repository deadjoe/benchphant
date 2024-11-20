package tpcc

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/deadjoe/benchphant/internal/benchmark"
	"go.uber.org/zap"
)

// TPCCBenchmark implements the Benchmark interface for TPC-C
type TPCCBenchmark struct {
	config *Config
	db     *sql.DB
	logger *zap.Logger
	runner *Runner
}

// NewTPCCBenchmark creates a new TPC-C benchmark instance
func NewTPCCBenchmark(config *Config, db *sql.DB, logger *zap.Logger) *TPCCBenchmark {
	return &TPCCBenchmark{
		config: config,
		db:     db,
		logger: logger,
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
	b.runner = NewRunner(b.db, b.config)
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
