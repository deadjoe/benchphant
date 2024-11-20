package sysbench

import (
	"context"
	"testing"
	"time"

	"github.com/deadjoe/benchphant/internal/benchmark/sysbench/types"
	"go.uber.org/zap/zaptest"
)

func TestOLTPTest(t *testing.T) {
	// Create test configuration
	config := types.NewOLTPTestConfig()
	config.TestType = types.TestTypeOLTPRead
	config.NumThreads = 4
	config.Duration = 10 * time.Second
	config.ReportInterval = 1 * time.Second

	// Create test instance
	test := NewOLTPTest(config, zaptest.NewLogger(t))

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
	err := test.Run(ctx)
	if err != nil {
		t.Errorf("Failed to run test: %v", err)
	}

	// Verify test status
	if test.Status().Status != string(types.TestStatusCompleted) {
		t.Errorf("Expected test status to be completed, got %s", test.Status().Status)
	}

	// Get report
	report := test.GetReport()
	if report == nil {
		t.Error("Expected report to be non-nil")
	}

	// Verify report
	t.Log("Verifying report...")
	if report.TestName != string(types.TestTypeOLTPRead) {
		t.Errorf("Expected test name %s, got %s", types.TestTypeOLTPRead, report.TestName)
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
	testTypes := []types.TestType{
		types.TestTypeOLTPWrite,
		types.TestTypeOLTPReadWrite,
		types.TestTypeOLTPPointSelect,
		types.TestTypeOLTPSimpleSelect,
		types.TestTypeOLTPSumRange,
		types.TestTypeOLTPOrderRange,
		types.TestTypeOLTPDistinctRange,
	}

	for _, testType := range testTypes {
		t.Run(string(testType), func(t *testing.T) {
			config.TestType = testType
			test := NewOLTPTest(config, zaptest.NewLogger(t))

			if err := test.Prepare(ctx); err != nil {
				t.Fatalf("Failed to prepare %s test: %v", testType, err)
			}
			defer test.Cleanup(ctx)

			err := test.Run(ctx)
			if err != nil {
				t.Fatalf("Failed to run %s test: %v", testType, err)
			}

			if test.Status().Status != string(types.TestStatusCompleted) {
				t.Errorf("Expected test status to be completed, got %s", test.Status().Status)
			}

			report := test.GetReport()
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
	config := types.NewOLTPTestConfig()

	// Test default values
	if config.NumThreads != 1 {
		t.Errorf("Expected default threads to be 1, got %d", config.NumThreads)
	}

	if config.Duration != 10*time.Second {
		t.Errorf("Expected default duration to be 10s, got %s", config.Duration)
	}

	if config.ReportInterval != 1*time.Second {
		t.Errorf("Expected default report interval to be 1s, got %s", config.ReportInterval)
	}

	// Test configuration updates
	config.NumThreads = 4
	config.Duration = 5 * time.Second
	config.ReportInterval = 2 * time.Second

	test := NewOLTPTest(config, zaptest.NewLogger(t))
	args := test.buildBaseArgs()

	// Verify command line arguments
	expectedArgs := []string{
		"--threads=4",
		"--time=5",
		"--report-interval=2",
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

func TestOLTPTestValidation(t *testing.T) {
	testCases := []struct {
		name          string
		configModify  func(*types.OLTPTestConfig)
		expectedError bool
	}{
		{
			name: "invalid threads",
			configModify: func(config *types.OLTPTestConfig) {
				config.NumThreads = 0
			},
			expectedError: true,
		},
		{
			name: "invalid duration",
			configModify: func(config *types.OLTPTestConfig) {
				config.Duration = 0
			},
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := types.NewOLTPTestConfig()
			tc.configModify(config)

			test := NewOLTPTest(config, zaptest.NewLogger(t))
			ctx := context.Background()

			err := test.Prepare(ctx)
			if tc.expectedError && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tc.expectedError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if err == nil {
				defer test.Cleanup(ctx)

				err := test.Run(ctx)
				if tc.expectedError && err == nil {
					t.Error("Expected error but got nil")
				}
				if !tc.expectedError && err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
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
			config := types.NewOLTPTestConfig()
			config.DBType = tc.dbType
			config.Host = tc.host
			config.Port = tc.port
			config.Username = tc.username
			config.Password = tc.password
			config.Database = tc.database
			config.TableSize = 100
			config.TablesCount = 1

			test := NewOLTPTest(config, zaptest.NewLogger(t))
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
	config := types.NewOLTPTestConfig()
	config.TestType = types.TestTypeOLTPReadWrite
	config.NumThreads = 32
	config.TableSize = 1000
	config.TablesCount = 4
	config.Duration = 10 * time.Second

	test := NewOLTPTest(config, zaptest.NewLogger(t))
	ctx := context.Background()

	// Prepare test data
	err := test.Prepare(ctx)
	require.NoError(t, err)
	defer test.Cleanup(ctx)

	// Run high concurrency test
	err = test.Run(ctx)
	require.NoError(t, err)

	// Verify high concurrency metrics
	report := test.GetReport()
	require.NotNil(t, report)

	assert.True(t, report.TotalTransactions > 0, "Expected non-zero total transactions")
	assert.True(t, report.TPS > 0, "Expected non-zero TPS")
	assert.True(t, report.LatencyAvg > 0, "Expected non-zero average latency")
	assert.True(t, report.LatencyP95 > 0, "Expected non-zero P95 latency")
	assert.True(t, report.LatencyP99 > 0, "Expected non-zero P99 latency")
}

func TestStressTest(t *testing.T) {
	config := types.NewOLTPTestConfig()
	config.TestType = types.TestTypeOLTPReadWrite
	config.NumThreads = 16
	config.TableSize = 10000
	config.TablesCount = 8
	config.Duration = 30 * time.Second

	test := NewOLTPTest(config, zaptest.NewLogger(t))
	ctx := context.Background()

	// Prepare test data
	err := test.Prepare(ctx)
	require.NoError(t, err)
	defer test.Cleanup(ctx)

	// Run stress test
	err = test.Run(ctx)
	require.NoError(t, err)

	// Verify stress test metrics
	report := test.GetReport()
	require.NotNil(t, report)

	assert.True(t, report.TotalTransactions > 0, "Expected non-zero total transactions")
	assert.True(t, report.TPS > 0, "Expected non-zero TPS")
	assert.True(t, report.LatencyAvg > 0, "Expected non-zero average latency")
	assert.True(t, report.LatencyP95 > report.LatencyAvg, "P95 latency should be higher than average")
	assert.True(t, report.LatencyP99 > report.LatencyP95, "P99 latency should be higher than P95")
}

func TestErrorRecovery(t *testing.T) {
	config := types.NewOLTPTestConfig()
	config.TestType = types.TestTypeOLTPReadWrite
	config.NumThreads = 4
	config.TableSize = 100
	config.TablesCount = 2

	test := NewOLTPTest(config, zaptest.NewLogger(t))
	ctx := context.Background()

	t.Run("DatabaseConnectionError", func(t *testing.T) {
		// Test with invalid database connection
		config.Host = "nonexistent-host"
		test := NewOLTPTest(config, zaptest.NewLogger(t))
		err := test.Prepare(ctx)
		assert.Error(t, err, "Expected error for invalid database connection")
	})

	t.Run("InvalidTableSize", func(t *testing.T) {
		// Test with invalid table size
		config.TableSize = -1
		test := NewOLTPTest(config, zaptest.NewLogger(t))
		err := test.Prepare(ctx)
		assert.Error(t, err, "Expected error for invalid table size")
	})

	t.Run("InvalidThreadCount", func(t *testing.T) {
		// Test with invalid thread count
		config.NumThreads = 0
		test := NewOLTPTest(config, zaptest.NewLogger(t))
		err := test.Prepare(ctx)
		assert.Error(t, err, "Expected error for invalid thread count")
	})

	t.Run("TestInterruption", func(t *testing.T) {
		// Test interruption during test execution
		config := types.NewOLTPTestConfig()
		config.Duration = 10 * time.Second
		test := NewOLTPTest(config, zaptest.NewLogger(t))

		ctx, cancel := context.WithCancel(context.Background())
		err := test.Prepare(ctx)
		require.NoError(t, err)
		defer test.Cleanup(ctx)

		// Start test in goroutine
		errChan := make(chan error)
		go func() {
			err := test.Run(ctx)
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
