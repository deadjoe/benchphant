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
	startTime := time.Now()

	// Run the benchmark
	if err := b.runner.Run(ctx); err != nil {
		return nil, fmt.Errorf("run benchmark: %w", err)
	}

	// Get statistics
	stats := b.runner.GetStats()
	duration := time.Since(startTime)

	// Convert metrics to map[string]interface{}
	metrics := make(map[string]interface{})
	for k, v := range stats.Metrics {
		metrics[k] = interface{}(v)
	}

	// Create result
	result := &benchmark.Result{
		Name:      b.Name(),
		StartTime: startTime,
		Duration:  duration,
		Metrics:   metrics,
	}

	b.logger.Info("Benchmark completed",
		zap.Float64("tpmC", stats.TPMc),
		zap.Float64("efficiency", stats.Efficiency),
		zap.Int64("total_transactions", stats.TotalTransactions),
		zap.Int64("total_errors", stats.TotalErrors),
		zap.Float64("overall_latency_avg", stats.OverallLatencyAvg),
	)

	return result, nil
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

	if b.config.Warehouses < 1 {
		return fmt.Errorf("warehouses must be greater than 0")
	}

	if b.config.Terminals < 1 {
		return fmt.Errorf("terminals must be greater than 0")
	}

	if b.config.Duration < time.Second {
		return fmt.Errorf("duration must be at least 1 second")
	}

	if b.config.ReportInterval < time.Second {
		return fmt.Errorf("report interval must be at least 1 second")
	}

	totalPercentage := b.config.NewOrderPercentage +
		b.config.PaymentPercentage +
		b.config.OrderStatusPercentage +
		b.config.DeliveryPercentage +
		b.config.StockLevelPercentage

	if totalPercentage != 100 {
		return fmt.Errorf("transaction percentages must sum to 100")
	}

	return nil
}
