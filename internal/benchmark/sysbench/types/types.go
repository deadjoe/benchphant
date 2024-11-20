package types

import (
	"time"
)

// ScenarioType represents different predefined test scenarios
type ScenarioType string

const (
	// ScenarioTypeReadOnly represents a read-only test scenario
	ScenarioTypeReadOnly ScenarioType = "read_only"
	// ScenarioTypeWriteOnly represents a write-only test scenario
	ScenarioTypeWriteOnly ScenarioType = "write_only"
	// ScenarioTypeReadWrite represents a mixed read-write test scenario
	ScenarioTypeReadWrite       ScenarioType = "read_write"
	ScenarioTypeBasicOLTP       ScenarioType = "basic_oltp"
	ScenarioTypeReadIntensive   ScenarioType = "read_intensive"
	ScenarioTypeWriteHeavy      ScenarioType = "write_heavy"
	ScenarioTypeMixedLoad       ScenarioType = "mixed_load"
	ScenarioTypeHighConcurrency ScenarioType = "high_concurrency"
	ScenarioTypeLongRunning     ScenarioType = "long_running"
	ScenarioTypeStressTest      ScenarioType = "stress_test"
)

// TestType represents different types of OLTP tests
type TestType string

const (
	// TestTypeOLTPRead represents a read-only OLTP test
	TestTypeOLTPRead TestType = "oltp_read_only"
	// TestTypeOLTPWrite represents a write-only OLTP test
	TestTypeOLTPWrite TestType = "oltp_write_only"
	// TestTypeOLTPReadWrite represents a mixed read-write OLTP test
	TestTypeOLTPReadWrite TestType = "oltp_read_write"
	// TestTypeOLTPPointSelect represents a point select OLTP test
	TestTypeOLTPPointSelect TestType = "oltp_point_select"
	// TestTypeOLTPSimpleSelect represents a simple select OLTP test
	TestTypeOLTPSimpleSelect TestType = "oltp_simple_select"
	// TestTypeOLTPSumRange represents a sum range OLTP test
	TestTypeOLTPSumRange TestType = "oltp_sum_range"
	// TestTypeOLTPOrderRange represents an order range OLTP test
	TestTypeOLTPOrderRange TestType = "oltp_order_range"
	// TestTypeOLTPDistinctRange represents a distinct range OLTP test
	TestTypeOLTPDistinctRange TestType = "oltp_distinct_range"
	// TestTypeOLTPIndexScan represents an index scan OLTP test
	TestTypeOLTPIndexScan TestType = "oltp_index_scan"
	// TestTypeOLTPNonIndexScan represents a non-index scan OLTP test
	TestTypeOLTPNonIndexScan TestType = "oltp_non_index_scan"
)

// OLTPTestConfig represents the configuration for OLTP tests
type OLTPTestConfig struct {
	TestType        TestType      `json:"test_type"`
	TableSize       int           `json:"table_size"`
	NumTables       int           `json:"num_tables"`
	NumThreads      int           `json:"num_threads"`
	Duration        time.Duration `json:"duration"`
	ReportInterval  time.Duration `json:"report_interval"`
	ReadOnly        bool          `json:"read_only"`
	PointSelects    int           `json:"point_selects"`
	SimpleRanges    int           `json:"simple_ranges"`
	SumRanges       int           `json:"sum_ranges"`
	OrderRanges     int           `json:"order_ranges"`
	DistinctRanges  int           `json:"distinct_ranges"`
	IndexUpdates    int           `json:"index_updates"`
	NonIndexUpdates int           `json:"non_index_updates"`
	DeleteInserts   int           `json:"delete_inserts"`
	WriteWeight     float64       `json:"write_weight"`
	ReadWeight      float64       `json:"read_weight"`
}

// TestConfig represents the configuration for a sysbench OLTP test
type TestConfig struct {
	// TestType is the type of OLTP test to run
	TestType TestType
	// DBType is the type of database to test (mysql, pgsql)
	DBType string
	// Host is the database host
	Host string
	// Port is the database port
	Port int
	// Username is the database username
	Username string
	// Password is the database password
	Password string
	// Database is the database name
	Database string
	// TableSize is the size of each test table
	TableSize int
	// TablesCount is the number of test tables
	TablesCount int
	// Threads is the number of test threads
	Threads int
	// Duration is the test duration
	Duration string
	// ReportInterval is the interval between progress reports
	ReportInterval string
	// Debug enables debug output
	Debug bool
}

// TestStats represents test statistics
type TestStats struct {
	TotalTransactions int64
	TPS               float64
	LatencyAvg        time.Duration
	LatencyP95        time.Duration
	LatencyP99        time.Duration
	Errors            int64
}

// AddTransaction adds a transaction to the statistics
func (s *TestStats) AddTransaction(duration time.Duration) {
	s.TotalTransactions++
	s.LatencyAvg = (s.LatencyAvg*time.Duration(s.TotalTransactions-1) + duration) / time.Duration(s.TotalTransactions)
	// Note: P95 and P99 calculations would require storing all durations and sorting them
	// For simplicity, we're not implementing them here
}

// Report represents a test report
type Report struct {
	TestName          string
	Duration          time.Duration
	TotalTransactions int64
	TPS               float64
	LatencyAvg        time.Duration
	LatencyP95        time.Duration
	LatencyP99        time.Duration
	Errors            int64
	StartTime         time.Time
	EndTime           time.Time
	Metrics           map[string]interface{}
}

// TestReport represents the results of a sysbench OLTP test
type TestReport struct {
	// TestName is the name of the test
	TestName string
	// StartTime is when the test started
	StartTime string
	// EndTime is when the test ended
	EndTime string
	// TotalTransactions is the total number of transactions executed
	TotalTransactions int64
	// TPS is the transactions per second
	TPS float64
	// LatencyAvg is the average transaction latency in milliseconds
	LatencyAvg float64
	// LatencyP95 is the 95th percentile transaction latency in milliseconds
	LatencyP95 float64
	// LatencyP99 is the 99th percentile transaction latency in milliseconds
	LatencyP99 float64
	// Errors is the number of errors encountered
	Errors int64
}

// Scenario represents a test scenario
type Scenario struct {
	Type        ScenarioType  `json:"type"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Duration    time.Duration `json:"duration"`
	Tests       []*Test       `json:"tests"`
}

// Test represents a single test within a scenario
type Test struct {
	Name     string        `json:"name"`
	Type     TestType      `json:"type"`
	Config   *TestConfig   `json:"config"`
	Weight   float64       `json:"weight"`
	Duration time.Duration `json:"duration"`
}

// Result represents a single operation result
type Result struct {
	Type      string        `json:"type"`
	Duration  time.Duration `json:"duration"`
	Success   bool          `json:"success"`
	Timestamp time.Time     `json:"timestamp"`
}

// IsReadOnly checks if the test type is read-only
func (t TestType) IsReadOnly() bool {
	switch t {
	case TestTypeOLTPRead, TestTypeOLTPPointSelect, TestTypeOLTPSimpleSelect,
		TestTypeOLTPSumRange, TestTypeOLTPOrderRange, TestTypeOLTPDistinctRange,
		TestTypeOLTPIndexScan, TestTypeOLTPNonIndexScan:
		return true
	default:
		return false
	}
}

// IsWriteOnly checks if the test type is write-only
func (t TestType) IsWriteOnly() bool {
	return t == TestTypeOLTPWrite
}

// IsReadWrite checks if the test type involves both read and write operations
func (t TestType) IsReadWrite() bool {
	return t == TestTypeOLTPReadWrite
}
