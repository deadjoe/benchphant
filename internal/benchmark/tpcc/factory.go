package tpcc

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"

	"github.com/deadjoe/benchphant/internal/benchmark"
	"github.com/deadjoe/benchphant/internal/models"
)

// Factory creates TPC-C benchmarks
type Factory struct{}

// NewFactory creates a new TPC-C benchmark factory
func NewFactory() *Factory {
	return &Factory{}
}

// Name returns the name of the benchmark type
func (f *Factory) Name() string {
	return "tpcc"
}

// Create creates a new TPC-C benchmark instance
func (f *Factory) Create(config *models.Benchmark, conn *models.DBConnection, logger *zap.Logger) (benchmark.Benchmark, error) {
	if config == nil {
		return nil, fmt.Errorf("config is required")
	}

	if conn == nil {
		return nil, fmt.Errorf("connection is required")
	}

	// Parse TPC-C specific config
	var tpccConfig Config
	if err := json.Unmarshal(config.Config, &tpccConfig); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	// Create database connection
	db, err := sql.Open(conn.Driver, conn.DSN)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Create benchmark
	b := NewTPCCBenchmark(&tpccConfig, db, logger)
	return b, nil
}

func init() {
	// Register factory
	benchmark.RegisterFactory("tpcc", &Factory{})
}
