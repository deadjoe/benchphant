package sysbench

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"

	"github.com/deadjoe/benchphant/internal/benchmark"
	"github.com/deadjoe/benchphant/internal/benchmark/sysbench/types"
	"github.com/deadjoe/benchphant/internal/models"
)

// Factory creates sysbench benchmarks
type Factory struct{}

// NewFactory creates a new sysbench benchmark factory
func NewFactory() *Factory {
	return &Factory{}
}

// Name returns the name of the benchmark type
func (f *Factory) Name() string {
	return "sysbench"
}

// Create creates a new sysbench benchmark instance
func (f *Factory) Create(config *models.Benchmark, conn *models.DBConnection, logger *zap.Logger) (benchmark.BenchmarkRunner, error) {
	if config == nil {
		return nil, fmt.Errorf("config is required")
	}

	if conn == nil {
		return nil, fmt.Errorf("connection is required")
	}

	// Parse sysbench specific config
	var oltpConfig types.OLTPTestConfig
	if err := json.Unmarshal(config.Config, &oltpConfig); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	// Create database connection
	db, err := sql.Open(conn.Driver, conn.DSN)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Create benchmark
	b := NewOLTPTest(&oltpConfig, logger)
	b.SetDB(db)
	return b, nil
}

func init() {
	// Register factory
	benchmark.RegisterFactory("sysbench", &Factory{})
}
