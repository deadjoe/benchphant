package models

import (
	"time"
)

// TestConfig represents the configuration for a sysbench test
type TestConfig struct {
	// Common configuration
	Threads        int           `json:"threads"`
	Tables         int           `json:"tables"`
	TableSize      int           `json:"table_size"`
	Database       string        `json:"database"`
	Duration       time.Duration `json:"duration"`
	ReportInterval int           `json:"report_interval"`
}

// Stats represents test statistics
type Stats struct {
	TPS               float64
	LatencyAvg        time.Duration
	LatencyP95        time.Duration
	LatencyP99        time.Duration
	TotalTransactions int64
	Errors            int64
}

// Report represents a test report
type Report struct {
	TestName          string
	TestConfig        TestConfig
	Duration          time.Duration
	TPS               float64
	LatencyAvg        time.Duration
	LatencyP95        time.Duration
	LatencyP99        time.Duration
	TotalTransactions int64
	Errors            int64
	StartTime         time.Time
	EndTime           time.Time
}

// ScenarioType represents different types of test scenarios
type ScenarioType string

const (
	ScenarioTypeBasicOLTP       ScenarioType = "basic_oltp"
	ScenarioTypeReadIntensive   ScenarioType = "read_intensive"
	ScenarioTypeWriteHeavy      ScenarioType = "write_heavy"
	ScenarioTypeMixedLoad       ScenarioType = "mixed_load"
	ScenarioTypeHighConcurrency ScenarioType = "high_concurrency"
	ScenarioTypeLongRunning     ScenarioType = "long_running"
	ScenarioTypeStressTest      ScenarioType = "stress_test"
)

// Scenario represents a test scenario
type Scenario struct {
	Type        ScenarioType  `json:"type"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Duration    time.Duration `json:"duration"`
}
