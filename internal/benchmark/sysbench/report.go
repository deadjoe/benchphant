package sysbench

import (
	"fmt"
	"time"
)

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

// NewReport creates a new Report from test statistics
func NewReport(testName string, config TestConfig, stats Stats, startTime time.Time, endTime time.Time) *Report {
	return &Report{
		TestName:          testName,
		TestConfig:        config,
		Duration:          endTime.Sub(startTime),
		TPS:               stats.TPS,
		LatencyAvg:        stats.LatencyAvg,
		LatencyP95:        stats.LatencyP95,
		LatencyP99:        stats.LatencyP99,
		TotalTransactions: stats.TotalTransactions,
		Errors:            stats.Errors,
		StartTime:         startTime,
		EndTime:           endTime,
	}
}

// String returns a string representation of the report
func (r *Report) String() string {
	return fmt.Sprintf(`
Test Report
==========
Test Name: %s
Duration: %v
Start Time: %v
End Time: %v

Test Configuration
-----------------
Threads: %d
Tables: %d
Table Size: %d
Database: %s

Performance Metrics
------------------
Total Transactions: %d
Transactions per Second: %.2f
Average Latency: %v
95th Percentile Latency: %v
99th Percentile Latency: %v
Total Errors: %d
Error Rate: %.2f%%
`,
		r.TestName,
		r.Duration,
		r.StartTime.Format(time.RFC3339),
		r.EndTime.Format(time.RFC3339),
		r.TestConfig.Threads,
		r.TestConfig.Tables,
		r.TestConfig.TableSize,
		r.TestConfig.Database,
		r.TotalTransactions,
		r.TPS,
		r.LatencyAvg,
		r.LatencyP95,
		r.LatencyP99,
		r.Errors,
		float64(r.Errors)/float64(r.TotalTransactions)*100,
	)
}

// JSON returns a JSON representation of the report
func (r *Report) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"test_name":        r.TestName,
		"duration_seconds": r.Duration.Seconds(),
		"start_time":       r.StartTime,
		"end_time":         r.EndTime,
		"config": map[string]interface{}{
			"threads":    r.TestConfig.Threads,
			"tables":     r.TestConfig.Tables,
			"table_size": r.TestConfig.TableSize,
			"database":   r.TestConfig.Database,
		},
		"metrics": map[string]interface{}{
			"total_transactions": r.TotalTransactions,
			"tps":                r.TPS,
			"latency_avg_ms":     r.LatencyAvg.Milliseconds(),
			"latency_p95_ms":     r.LatencyP95.Milliseconds(),
			"latency_p99_ms":     r.LatencyP99.Milliseconds(),
			"errors":             r.Errors,
			"error_rate":         float64(r.Errors) / float64(r.TotalTransactions) * 100,
		},
	}
}
