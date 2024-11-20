package sysbench

import (
	"context"
	"testing"
	"time"

	"github.com/deadjoe/benchphant/internal/benchmark/sysbench/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOLTPTest(t *testing.T) {
	// Create test configuration
	config := NewOLTPTestConfig()
	config.TestType = TestTypeOLTPRead
	config.Database = "test_db"
	config.Username = "root"
	config.Password = ""
	config.Host = "localhost"
	config.Port = 3306
	config.Threads = 4
	config.TableSize = 1000
	config.TablesCount = 2

	// Create test instance
	test := NewOLTPTest(config)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Prepare test
	t.Log("Preparing test...")
	if err := test.Prepare(ctx); err != nil {
		t.Fatalf("Failed to prepare test: %v", err)
	}
	defer func() {
		if err := test.Cleanup(ctx); err != nil {
			t.Errorf("Failed to cleanup test: %v", err)
		}
	}()

	// Run test
	t.Log("Running test...")
	report, err := test.Run(ctx)
	if err != nil {
		t.Fatalf("Failed to run test: %v", err)
	}

	// Verify report
	t.Log("Verifying report...")
	if report == nil {
		t.Fatal("Report is nil")
	}

	if report.TestName != string(TestTypeOLTPRead) {
		t.Errorf("Expected test name %s, got %s", TestTypeOLTPRead, report.TestName)
	}

	if report.TotalTransactions == 0 {
		t.Error("Expected non-zero total transactions")
	}

	if report.TPS == 0 {
		t.Error("Expected non-zero TPS")
	}

	if report.LatencyAvg == 0 {
		t.Error("Expected non-zero average latency")
	}

	// Test different OLTP test types
	testTypes := []TestType{
		TestTypeOLTPWrite,
		TestTypeOLTPReadWrite,
		TestTypeOLTPPointSelect,
		TestTypeOLTPSimpleSelect,
		TestTypeOLTPSumRange,
		TestTypeOLTPOrderRange,
		TestTypeOLTPDistinctRange,
	}

	for _, testType := range testTypes {
		t.Run(string(testType), func(t *testing.T) {
			config.TestType = testType
			test := NewOLTPTest(config)

			if err := test.Prepare(ctx); err != nil {
				t.Fatalf("Failed to prepare %s test: %v", testType, err)
			}
			defer test.Cleanup(ctx)

			report, err := test.Run(ctx)
			if err != nil {
				t.Fatalf("Failed to run %s test: %v", testType, err)
			}

			if report == nil {
				t.Fatal("Report is nil")
			}

			if report.TestName != string(testType) {
				t.Errorf("Expected test name %s, got %s", testType, report.TestName)
			}
		})
	}
}

func TestOLTPTestConfig(t *testing.T) {
	config := NewOLTPTestConfig()

	// Test default values
	if config.Threads != 1 {
		t.Errorf("Expected default threads to be 1, got %d", config.Threads)
	}

	if config.Database != "sbtest" {
		t.Errorf("Expected default database to be 'sbtest', got %s", config.Database)
	}

	if config.TableSize != 10000 {
		t.Errorf("Expected default table size to be 10000, got %d", config.TableSize)
	}

	if config.TablesCount != 1 {
		t.Errorf("Expected default tables count to be 1, got %d", config.TablesCount)
	}

	// Test configuration updates
	config.Threads = 4
	config.TableSize = 1000
	config.TablesCount = 2
	config.ReadOnly = true

	test := NewOLTPTest(config)
	args := test.buildBaseArgs()

	// Verify command line arguments
	expectedArgs := []string{
		"--threads=4",
		"--tables=2",
		"--table-size=1000",
	}

	for _, expected := range expectedArgs {
		found := false
		for _, arg := range args {
			if arg == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected argument %s not found in args: %v", expected, args)
		}
	}
}

func TestOLTPTestOutput(t *testing.T) {
	test := NewOLTPTest(NewOLTPTestConfig())

	// Test output parsing
	sampleOutput := `
sysbench 1.0.20 (using bundled LuaJIT 2.1.0-beta2)

Running the test with following options:
Number of threads: 4
Initializing random number generator from current time

Initializing worker threads...

Threads started!

SQL statistics:
    queries performed:
        read:                            140
        write:                           40
        other:                           20
        total:                           200
    transactions:                        10     (1.23 per sec.)
    queries:                            200    (24.60 per sec.)
    ignored errors:                      0      (0.00 per sec.)
    reconnects:                         0      (0.00 per sec.)

General statistics:
    total time:                          8.1234s
    total number of events:              10

Latency (ms):
         min:                                    1.23
         avg:                                    4.56
         max:                                   10.89
         95th percentile:                        8.90
         99th percentile:                        9.99
         sum:                                45.60

Threads fairness:
    events (avg/stddev):           2.5000/0.50
    execution time (avg/stddev):   0.0114/0.00
`

	stats := test.parseOutput(sampleOutput)

	if stats.TPS != 1.23 {
		t.Errorf("Expected TPS 1.23, got %f", stats.TPS)
	}

	if stats.LatencyAvg != time.Duration(4.56*float64(time.Millisecond)) {
		t.Errorf("Expected avg latency 4.56ms, got %v", stats.LatencyAvg)
	}

	if stats.LatencyP95 != time.Duration(8.90*float64(time.Millisecond)) {
		t.Errorf("Expected P95 latency 8.90ms, got %v", stats.LatencyP95)
	}

	if stats.LatencyP99 != time.Duration(9.99*float64(time.Millisecond)) {
		t.Errorf("Expected P99 latency 9.99ms, got %v", stats.LatencyP99)
	}

	if stats.TotalTransactions != 10 {
		t.Errorf("Expected 10 total transactions, got %d", stats.TotalTransactions)
	}

	if stats.Errors != 0 {
		t.Errorf("Expected 0 errors, got %d", stats.Errors)
	}
}

func TestMultiDatabaseSupport(t *testing.T) {
	testCases := []struct {
		name     string
		dbType   string
		host     string
		port     int
		username string
		password string
		database string
	}{
		{
			name:     "MySQL",
			dbType:   "mysql",
			host:     "localhost",
			port:     3306,
			username: "root",
			password: "",
			database: "sbtest",
		},
		{
			name:     "PostgreSQL",
			dbType:   "pgsql",
			host:     "localhost",
			port:     5432,
			username: "postgres",
			password: "",
			database: "sbtest",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := NewOLTPTestConfig()
			config.DBType = tc.dbType
			config.Host = tc.host
			config.Port = tc.port
			config.Username = tc.username
			config.Password = tc.password
			config.Database = tc.database
			config.TableSize = 100
			config.TablesCount = 1

			test := NewOLTPTest(config)
			args := test.buildBaseArgs()

			// Verify database-specific arguments
			expectedArgs := []string{
				"--db-driver=" + tc.dbType,
				"--mysql-host=" + tc.host,
				"--mysql-port=" + string(tc.port),
				"--mysql-user=" + tc.username,
				"--mysql-password=" + tc.password,
				"--mysql-db=" + tc.database,
			}

			for _, expected := range expectedArgs {
				found := false
				for _, arg := range args {
					if arg == expected {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected argument %s not found in args: %v", expected, args)
			}
		})
	}
}

func TestHighConcurrency(t *testing.T) {
	config := NewOLTPTestConfig()
	config.TestType = TestTypeOLTPReadWrite
	config.Threads = 32
	config.TableSize = 1000
	config.TablesCount = 4
	config.Duration = 10 * time.Second

	test := NewOLTPTest(config)
	ctx := context.Background()

	// Prepare test data
	err := test.Prepare(ctx)
	require.NoError(t, err)
	defer test.Cleanup(ctx)

	// Run high concurrency test
	report, err := test.Run(ctx)
	require.NoError(t, err)
	require.NotNil(t, report)

	// Verify high concurrency metrics
	assert.True(t, report.TotalTransactions > 0, "Expected non-zero total transactions")
	assert.True(t, report.TPS > 0, "Expected non-zero TPS")
	assert.True(t, report.LatencyAvg > 0, "Expected non-zero average latency")
	assert.True(t, report.LatencyP95 > 0, "Expected non-zero P95 latency")
	assert.True(t, report.LatencyP99 > 0, "Expected non-zero P99 latency")
}

func TestStressTest(t *testing.T) {
	config := NewOLTPTestConfig()
	config.TestType = TestTypeOLTPReadWrite
	config.Threads = 16
	config.TableSize = 10000
	config.TablesCount = 8
	config.Duration = 30 * time.Second

	test := NewOLTPTest(config)
	ctx := context.Background()

	// Prepare test data
	err := test.Prepare(ctx)
	require.NoError(t, err)
	defer test.Cleanup(ctx)

	// Run stress test
	report, err := test.Run(ctx)
	require.NoError(t, err)
	require.NotNil(t, report)

	// Verify stress test metrics
	assert.True(t, report.TotalTransactions > 0, "Expected non-zero total transactions")
	assert.True(t, report.TPS > 0, "Expected non-zero TPS")
	assert.True(t, report.LatencyAvg > 0, "Expected non-zero average latency")
	assert.True(t, report.LatencyP95 > report.LatencyAvg, "P95 latency should be higher than average")
	assert.True(t, report.LatencyP99 > report.LatencyP95, "P99 latency should be higher than P95")
}

func TestErrorRecovery(t *testing.T) {
	config := NewOLTPTestConfig()
	config.TestType = TestTypeOLTPReadWrite
	config.Threads = 4
	config.TableSize = 100
	config.TablesCount = 2

	test := NewOLTPTest(config)
	ctx := context.Background()

	t.Run("DatabaseConnectionError", func(t *testing.T) {
		// Test with invalid database connection
		config.Host = "nonexistent-host"
		test := NewOLTPTest(config)
		err := test.Prepare(ctx)
		assert.Error(t, err, "Expected error for invalid database connection")
	})

	t.Run("InvalidTableSize", func(t *testing.T) {
		// Test with invalid table size
		config.TableSize = -1
		test := NewOLTPTest(config)
		err := test.Prepare(ctx)
		assert.Error(t, err, "Expected error for invalid table size")
	})

	t.Run("InvalidThreadCount", func(t *testing.T) {
		// Test with invalid thread count
		config.Threads = 0
		test := NewOLTPTest(config)
		err := test.Prepare(ctx)
		assert.Error(t, err, "Expected error for invalid thread count")
	})

	t.Run("TestInterruption", func(t *testing.T) {
		// Test interruption during test execution
		config := NewOLTPTestConfig()
		config.Duration = 10 * time.Second
		test := NewOLTPTest(config)

		ctx, cancel := context.WithCancel(context.Background())
		err := test.Prepare(ctx)
		require.NoError(t, err)
		defer test.Cleanup(ctx)

		// Start test in goroutine
		errChan := make(chan error)
		go func() {
			_, err := test.Run(ctx)
			errChan <- err
		}()

		// Cancel context after short delay
		time.Sleep(100 * time.Millisecond)
		cancel()

		// Verify test was interrupted
		err = <-errChan
		assert.Error(t, err, "Expected error due to context cancellation")
	})
}
