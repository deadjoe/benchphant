package sysbench

import (
	"encoding/json"
	"time"
)

// TestType represents different types of sysbench tests
type TestType string

const (
	// OLTP test types
	TestTypeOLTPRead          TestType = "oltp_read_only"
	TestTypeOLTPWrite         TestType = "oltp_write_only"
	TestTypeOLTPReadWrite     TestType = "oltp_read_write"
	TestTypeOLTPPointSelect   TestType = "oltp_point_select"
	TestTypeOLTPSimpleSelect  TestType = "oltp_simple_select"
	TestTypeOLTPSumRange      TestType = "oltp_sum_range"
	TestTypeOLTPOrderRange    TestType = "oltp_order_range"
	TestTypeOLTPDistinctRange TestType = "oltp_distinct_range"
	TestTypeOLTPIndexScan     TestType = "oltp_index_scan"
	TestTypeOLTPNonIndexScan  TestType = "oltp_non_index_scan"
)

// TestMode represents the mode of operation for a test
type TestMode string

const (
	TestModePrepare TestMode = "prepare"
	TestModeRun     TestMode = "run"
	TestModeCleanup TestMode = "cleanup"
)

// OLTPTestType represents different types of OLTP tests
type OLTPTestType string

const (
	// ReadOnly represents read-only OLTP test
	ReadOnly OLTPTestType = "read_only"
	// ReadWrite represents read-write OLTP test
	ReadWrite OLTPTestType = "read_write"
	// WriteOnly represents write-only OLTP test
	WriteOnly OLTPTestType = "write_only"
	// PointSelect represents point select OLTP test
	PointSelect OLTPTestType = "point_select"
	// SimpleRanges represents simple ranges OLTP test
	SimpleRanges OLTPTestType = "simple_ranges"
	// SumRanges represents sum ranges OLTP test
	SumRanges OLTPTestType = "sum_ranges"
	// OrderRanges represents order ranges OLTP test
	OrderRanges OLTPTestType = "order_ranges"
	// DistinctRanges represents distinct ranges OLTP test
	DistinctRanges OLTPTestType = "distinct_ranges"
	// IndexUpdates represents index updates OLTP test
	IndexUpdates OLTPTestType = "index_updates"
	// NonIndexUpdates represents non-index updates OLTP test
	NonIndexUpdates OLTPTestType = "non_index_updates"
)

// Config represents sysbench test configuration
type Config struct {
	// Test type
	Type TestType `json:"type"`
	// OLTP specific configuration
	OLTP *OLTPConfig `json:"oltp,omitempty"`
	// FileIO specific configuration
	FileIO *FileIOConfig `json:"fileio,omitempty"`
	// CPU specific configuration
	CPU *CPUConfig `json:"cpu,omitempty"`
	// Memory specific configuration
	Memory *MemoryConfig `json:"memory,omitempty"`
	// Threads specific configuration
	Threads *ThreadsConfig `json:"threads,omitempty"`
	// Mutex specific configuration
	Mutex *MutexConfig `json:"mutex,omitempty"`
	// Common configuration
	Common CommonConfig `json:"common"`
}

// CommonConfig represents common configuration for all test types
type CommonConfig struct {
	// Number of threads to use
	NumThreads int `json:"num_threads"`
	// Test duration in seconds
	Duration int `json:"duration"`
	// Report interval in seconds
	ReportInterval int `json:"report_interval"`
}

// OLTPConfig represents OLTP test configuration
type OLTPConfig struct {
	// Test type
	TestType OLTPTestType `json:"test_type"`
	// Table size
	TableSize int `json:"table_size"`
	// Number of tables
	NumTables int `json:"num_tables"`
	// Number of ranges for range tests
	NumRanges int `json:"num_ranges"`
	// Number of points for point select tests
	NumPoints int `json:"num_points"`
	// Whether to skip trx begin/commit
	SkipTrx bool `json:"skip_trx"`
	// Range size for range tests
	RangeSize int `json:"range_size"`
	// Whether to use secondary index
	UseSecondaryIndex bool `json:"use_secondary_index"`
}

// FileIOConfig represents file I/O test configuration
type FileIOConfig struct {
	// File total size
	FileSize int64 `json:"file_size"`
	// Block size
	BlockSize int `json:"block_size"`
	// Number of files
	NumFiles int `json:"num_files"`
	// File test mode (seqwr, seqrewr, seqrd, rndrd, rndwr, rndrw)
	Mode string `json:"mode"`
}

// CPUConfig represents CPU test configuration
type CPUConfig struct {
	// Maximum prime number to calculate
	MaxPrime int `json:"max_prime"`
	// CPU operation mode (simple, double, triple)
	Mode string `json:"mode"`
}

// MemoryConfig represents memory test configuration
type MemoryConfig struct {
	// Block size
	BlockSize int `json:"block_size"`
	// Total size
	TotalSize int64 `json:"total_size"`
	// Memory operation (read, write)
	Operation string `json:"operation"`
	// Memory access mode (seq, rnd)
	Mode string `json:"mode"`
}

// ThreadsConfig represents threads test configuration
type ThreadsConfig struct {
	// Number of mutex locks per thread
	NumMutexes int `json:"num_mutexes"`
	// Lock time in milliseconds
	LockTime int `json:"lock_time"`
}

// MutexConfig represents mutex test configuration
type MutexConfig struct {
	// Number of mutex locks per thread
	NumMutexes int `json:"num_mutexes"`
	// Lock time in milliseconds
	LockTime int `json:"lock_time"`
}

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

// GetTestTypes returns all available database test types
func GetTestTypes() []TestType {
	return []TestType{
		TestTypeOLTPRead,
		TestTypeOLTPWrite,
		TestTypeOLTPReadWrite,
		TestTypeOLTPPointSelect,
		TestTypeOLTPSimpleSelect,
		TestTypeOLTPSumRange,
		TestTypeOLTPOrderRange,
		TestTypeOLTPDistinctRange,
		TestTypeOLTPIndexScan,
		TestTypeOLTPNonIndexScan,
	}
}
