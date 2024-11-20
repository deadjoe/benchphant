package oltp

import (
	"time"

	"github.com/deadjoe/benchphant/internal/benchmark/sysbench/types"
)

// Config represents the configuration for OLTP tests
type Config struct {
	// Basic settings
	TableSize      int           `json:"table_size"`
	NumTables      int           `json:"num_tables"`
	NumThreads     int           `json:"num_threads"`
	Duration       time.Duration `json:"duration"`
	ReportInterval time.Duration `json:"report_interval"`

	// Test specific settings
	ReadOnly        bool    `json:"read_only"`
	PointSelects    int     `json:"point_selects"`
	SimpleRanges    int     `json:"simple_ranges"`
	SumRanges       int     `json:"sum_ranges"`
	OrderRanges     int     `json:"order_ranges"`
	DistinctRanges  int     `json:"distinct_ranges"`
	IndexUpdates    int     `json:"index_updates"`
	NonIndexUpdates int     `json:"non_index_updates"`
	DeleteInserts   int     `json:"delete_inserts"`
	WriteWeight     float64 `json:"write_weight"`
	ReadWeight      float64 `json:"read_weight"`

	// Database specific settings
	AutoInc       bool   `json:"auto_inc"`
	SecondaryKeys bool   `json:"secondary_keys"`
	Engine        string `json:"engine"`
}

// NewDefaultConfig returns a new Config with default values
func NewDefaultConfig() *Config {
	return &Config{
		TableSize:       10000,
		NumTables:       1,
		NumThreads:      4,
		Duration:        10 * time.Minute,
		ReportInterval:  10 * time.Second,
		ReadOnly:        false,
		PointSelects:    10,
		SimpleRanges:    1,
		SumRanges:       1,
		OrderRanges:     1,
		DistinctRanges:  1,
		IndexUpdates:    1,
		NonIndexUpdates: 1,
		DeleteInserts:   1,
		WriteWeight:     0.5,
		ReadWeight:      0.5,
		AutoInc:         true,
		SecondaryKeys:   true,
		Engine:          "InnoDB",
	}
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.TableSize <= 0 {
		return types.ErrInvalidTableSize
	}
	if c.NumTables <= 0 {
		return types.ErrInvalidNumTables
	}
	if c.NumThreads <= 0 {
		return types.ErrInvalidNumThreads
	}
	if c.Duration <= 0 {
		return types.ErrInvalidDuration
	}
	if c.ReportInterval <= 0 {
		return types.ErrInvalidReportInterval
	}
	if c.WriteWeight < 0 || c.WriteWeight > 1 {
		return types.ErrInvalidWeight
	}
	if c.ReadWeight < 0 || c.ReadWeight > 1 {
		return types.ErrInvalidWeight
	}
	if c.WriteWeight+c.ReadWeight != 1.0 {
		return types.ErrInvalidWeightSum
	}
	return nil
}
