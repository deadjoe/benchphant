package benchmark

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/deadjoe/benchphant/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setupTestBenchmark(t *testing.T) (*Benchmark, *sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	logger := zap.NewNop()
	conn := &models.DBConnection{
		Name:        "test_db",
		Type:        models.MySQL,
		Host:        "localhost",
		Port:        3306,
		Database:    "test",
		MaxIdleConn: 10,
		MaxOpenConn: 100,
	}
	conn.SetDB(db)

	b := NewBenchmark(&models.Benchmark{
		Name:          "Test Benchmark",
		Description:   "Test benchmark description",
		QueryTemplate: "SELECT 1",
		NumThreads:    1,
		Duration:      time.Second,
		Status:        models.BenchmarkStatusPending,
	}, conn, logger)

	require.NotNil(t, b)

	return b, db, mock
}

func TestNewBenchmark(t *testing.T) {
	b, _, _ := setupTestBenchmark(t)

	assert.Equal(t, string(models.BenchmarkStatusPending), b.Status().Status)
	assert.NotNil(t, b.Status().Metrics)
}

func TestBenchmarkStart(t *testing.T) {
	b, _, mock := setupTestBenchmark(t)

	// Set up mock expectations for multiple queries
	mock.ExpectPrepare("SELECT 1").WillBeClosed()

	// Expect multiple executions based on duration and threads
	duration := time.Second
	threads := 1
	expectedQueries := int(duration.Seconds() * float64(threads) * 500) // Expect more queries since we sleep only 1ms
	for i := 0; i < expectedQueries; i++ {
		mock.ExpectExec("SELECT 1").WillReturnResult(sqlmock.NewResult(0, 0))
	}

	// Start the benchmark
	err := b.Start()
	assert.NoError(t, err)
	assert.Equal(t, string(models.BenchmarkStatusRunning), b.Status().Status)

	// Wait for completion
	<-b.done

	// Verify final state
	status := b.Status()
	assert.Equal(t, string(models.BenchmarkStatusCompleted), status.Status)
	assert.Equal(t, float64(100), status.Progress)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBenchmarkStop(t *testing.T) {
	b, _, mock := setupTestBenchmark(t)

	// Set up mock expectations
	mock.ExpectPrepare("SELECT 1").WillBeClosed()

	// Expect multiple executions
	for i := 0; i < 500; i++ { // Expect more queries to ensure we don't run out during the test
		mock.ExpectExec("SELECT 1").WillReturnResult(sqlmock.NewResult(0, 0))
	}

	// Start the benchmark
	err := b.Start()
	assert.NoError(t, err)

	// Wait briefly to allow some queries to execute
	time.Sleep(100 * time.Millisecond)

	// Stop the benchmark
	b.Stop()

	// Wait for completion
	<-b.done

	// Verify final state
	status := b.Status()
	assert.Equal(t, string(models.BenchmarkStatusCancelled), status.Status)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBenchmarkStatus(t *testing.T) {
	b, _, mock := setupTestBenchmark(t)

	// Set up mock expectations
	mock.ExpectPrepare("SELECT 1").WillBeClosed()

	// Expect multiple executions
	for i := 0; i < 500; i++ { // Expect more queries to ensure we don't run out
		mock.ExpectExec("SELECT 1").WillReturnResult(sqlmock.NewResult(0, 0))
	}

	// Start the benchmark
	err := b.Start()
	assert.NoError(t, err)
	assert.Equal(t, string(models.BenchmarkStatusRunning), b.Status().Status)

	// Wait for completion
	<-b.done

	// Verify final state
	status := b.Status()
	assert.Equal(t, string(models.BenchmarkStatusCompleted), status.Status)
	assert.Equal(t, float64(100), status.Progress)

	// Verify metrics
	assert.NotZero(t, status.Metrics["qps"])
	assert.NotZero(t, status.Metrics["latency_avg"])
	assert.NotZero(t, status.Metrics["latency_p95"])
	assert.NotZero(t, status.Metrics["latency_p99"])
	assert.Zero(t, status.Metrics["errors"])

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBenchmarkErrorHandling(t *testing.T) {
	t.Run("Query Error", func(t *testing.T) {
		b, _, mock := setupTestBenchmark(t)

		// Set up mock expectations
		mock.ExpectPrepare("SELECT 1").WillBeClosed()
		mock.ExpectExec("SELECT 1").WillReturnError(fmt.Errorf("query error"))

		// Start the benchmark
		err := b.Start()
		assert.NoError(t, err)

		// Wait for completion
		<-b.done

		// Verify error state
		status := b.Status()
		assert.Equal(t, string(models.BenchmarkStatusFailed), status.Status)
		assert.NotZero(t, status.Metrics["errors"])
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Prepare Error", func(t *testing.T) {
		b, _, mock := setupTestBenchmark(t)

		// Set up mock expectations
		mock.ExpectPrepare("SELECT 1").WillReturnError(fmt.Errorf("prepare error"))

		// Start the benchmark
		err := b.Start()
		assert.Error(t, err)
		assert.Equal(t, string(models.BenchmarkStatusFailed), b.Status().Status)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestBenchmarkEdgeCases(t *testing.T) {
	t.Run("Invalid Configuration", func(t *testing.T) {
		logger := zap.NewNop()

		// Test invalid thread count
		b := NewBenchmark(&models.Benchmark{
			Name:          "Test Benchmark",
			Description:   "Test benchmark description",
			QueryTemplate: "SELECT 1",
			NumThreads:    0,
			Duration:      time.Second,
			Status:        models.BenchmarkStatusPending,
		}, nil, logger)

		err := b.Start()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "number of threads must be greater than 0")

		// Test invalid duration
		b = NewBenchmark(&models.Benchmark{
			Name:          "Test Benchmark",
			Description:   "Test benchmark description",
			QueryTemplate: "SELECT 1",
			NumThreads:    1,
			Duration:      0,
			Status:        models.BenchmarkStatusPending,
		}, nil, logger)

		err = b.Start()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "duration must be greater than 0")

		// Test empty query
		b = NewBenchmark(&models.Benchmark{
			Name:          "Test Benchmark",
			Description:   "Test benchmark description",
			QueryTemplate: "",
			NumThreads:    1,
			Duration:      time.Second,
			Status:        models.BenchmarkStatusPending,
		}, nil, logger)

		err = b.Start()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "query template cannot be empty")
	})

	t.Run("Nil Database", func(t *testing.T) {
		logger := zap.NewNop()
		b := NewBenchmark(&models.Benchmark{
			Name:          "Test Benchmark",
			Description:   "Test benchmark description",
			QueryTemplate: "SELECT 1",
			NumThreads:    1,
			Duration:      time.Second,
			Status:        models.BenchmarkStatusPending,
		}, nil, logger)

		err := b.Start()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database connection is nil")
	})

	t.Run("Double Start", func(t *testing.T) {
		b, _, mock := setupTestBenchmark(t)

		// Set up mock expectations for the first start
		mock.ExpectPrepare("SELECT 1").WillBeClosed()
		for i := 0; i < 100; i++ {
			mock.ExpectExec("SELECT 1").WillReturnResult(sqlmock.NewResult(0, 0))
		}

		// First start should succeed
		err := b.Start()
		assert.NoError(t, err)

		// Second start should fail
		err = b.Start()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "benchmark is already running")

		// Stop the benchmark
		b.Stop()
		<-b.done
	})
}

func TestProgressCalculation(t *testing.T) {
	b, _, mock := setupTestBenchmark(t)

	// Set up mock expectations
	mock.ExpectPrepare("SELECT 1").WillBeClosed()

	// Expect multiple executions
	for i := 0; i < 1000; i++ { // Expect more queries to ensure we don't run out
		mock.ExpectExec("SELECT 1").WillReturnResult(sqlmock.NewResult(0, 0))
	}

	// Define progress checkpoints
	type checkpoint struct {
		minProg float64
		maxProg float64
	}

	checkpoints := []checkpoint{
		{0, 30},   // Early stage
		{30, 55},  // Quarter way
		{55, 80},  // Half way
		{80, 100}, // Late stage
	}

	// Start the benchmark
	err := b.Start()
	assert.NoError(t, err)

	// Check progress at each checkpoint
	for _, cp := range checkpoints {
		progress := b.Status().Progress
		assert.Greater(t, progress, cp.minProg)
		assert.Less(t, progress, cp.maxProg)
		time.Sleep(250 * time.Millisecond)
	}

	// Wait for completion
	<-b.done

	// Verify final state
	status := b.Status()
	assert.Equal(t, string(models.BenchmarkStatusCompleted), status.Status)
	assert.Equal(t, float64(100), status.Progress)

	// Verify metrics
	assert.NotZero(t, status.Metrics["qps"])
	assert.NotZero(t, status.Metrics["latency_avg"])
	assert.NotZero(t, status.Metrics["latency_p95"])
	assert.NotZero(t, status.Metrics["latency_p99"])
	assert.Zero(t, status.Metrics["errors"])

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}
