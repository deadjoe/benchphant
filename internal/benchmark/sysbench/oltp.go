package sysbench

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/deadjoe/benchphant/internal/benchmark"
	"github.com/deadjoe/benchphant/internal/benchmark/sysbench/types"
	"go.uber.org/zap"
)

// OLTPTestConfig represents configuration for OLTP tests
type OLTPTestConfig struct {
	TestType        types.TestType  `json:"test_type"`
	NumThreads      int            `json:"num_threads"`
	Duration        time.Duration  `json:"duration"`
	ReportInterval  time.Duration  `json:"report_interval"`
	TableSize       int            `json:"table_size"`
	TablesCount     int            `json:"tables_count"`
	DistinctRanges  int            `json:"distinct_ranges"`
	SimpleRanges    int            `json:"simple_ranges"`
	SumRanges       int            `json:"sum_ranges"`
	OrderRanges     int            `json:"order_ranges"`
	PointSelects    int            `json:"point_selects"`
	SimpleSelects   int            `json:"simple_selects"`
	IndexUpdates    int            `json:"index_updates"`
	NonIndexUpdates int            `json:"non_index_updates"`
	Inserts         int            `json:"inserts"`
	Deletes         int            `json:"deletes"`
	ReadOnly        bool           `json:"read_only"`
	WriteOnly       bool           `json:"write_only"`
	AutoInc         bool           `json:"auto_inc"`
	SecondaryIndexes int           `json:"secondary_indexes"`
	DeleteInserts    bool          `json:"delete_inserts"`
	RangeSize        int           `json:"range_size"`
	RangeSelects     bool          `json:"range_selects"`
	JoinSelects      bool          `json:"join_selects"`
	SkipTrx          bool          `json:"skip_trx"`
	TrxRate          float64       `json:"trx_rate"`
}

// NewOLTPTestConfig creates a new OLTP test configuration with default values
func NewOLTPTestConfig() *OLTPTestConfig {
	return &OLTPTestConfig{
		TestType:        types.TestTypeOLTPReadWrite,
		NumThreads:      1,
		Duration:        10 * time.Second,
		ReportInterval:  1 * time.Second,
		TableSize:       10000,
		TablesCount:     1,
		DistinctRanges:  100,
		SimpleRanges:    100,
		SumRanges:       100,
		OrderRanges:     100,
		PointSelects:    10,
		SimpleSelects:   1,
		IndexUpdates:    1,
		NonIndexUpdates: 1,
		Inserts:         1,
		Deletes:         1,
		ReadOnly:        false,
		WriteOnly:       false,
		AutoInc:         true,
		SecondaryIndexes: 0,
		DeleteInserts:    false,
		RangeSize:        100,
		RangeSelects:     true,
		JoinSelects:      false,
		SkipTrx:          false,
		TrxRate:          0, // No limit
	}
}

// OLTPTest represents a sysbench OLTP test
type OLTPTest struct {
	db     *sql.DB
	config *types.OLTPTestConfig
	logger *zap.Logger
	stats  *types.TestStats
	status benchmark.BenchmarkStatus
	mu     sync.RWMutex
	done   chan struct{}
}

// NewOLTPTest creates a new OLTP test
func NewOLTPTest(config *types.OLTPTestConfig, logger *zap.Logger) *OLTPTest {
	return &OLTPTest{
		config: config,
		logger: logger,
		stats:  types.NewTestStats(),
		status: benchmark.BenchmarkStatus{
			Status:   string(types.TestStatusPending),
			Progress: 0,
			Metrics:  make(map[string]interface{}),
		},
		done: make(chan struct{}),
	}
}

// SetDB sets the database connection
func (t *OLTPTest) SetDB(db *sql.DB) {
	t.db = db
}

// Start starts the test
func (t *OLTPTest) Start() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.db == nil {
		return fmt.Errorf("database connection not set")
	}

	if t.status.Status == string(types.TestStatusRunning) {
		return fmt.Errorf("test is already running")
	}

	t.status.Status = string(types.TestStatusRunning)
	t.status.Progress = 0
	t.status.Metrics = make(map[string]interface{})

	t.logger.Info("Starting OLTP test",
		zap.String("test_type", string(t.config.TestType)),
		zap.Int("num_threads", t.config.NumThreads),
		zap.Duration("duration", t.config.Duration),
	)

	// Run test in a goroutine
	go func() {
		ctx := context.Background()
		if err := t.Run(ctx); err != nil {
			t.logger.Error("Test failed", zap.Error(err))
			t.mu.Lock()
			t.status.Status = string(types.TestStatusFailed)
			t.mu.Unlock()
			return
		}

		t.mu.Lock()
		t.status.Status = string(types.TestStatusCompleted)
		t.status.Progress = 100
		t.status.Metrics = map[string]interface{}{
			"total_transactions": t.stats.TotalTransactions,
			"tps":               t.stats.TPS,
			"latency_avg":       t.stats.AvgLatency,
			"latency_p95":       t.stats.P95Latency,
			"latency_p99":       t.stats.P99Latency,
			"errors":            t.stats.TotalErrors,
		}
		t.mu.Unlock()
	}()

	return nil
}

// Stop stops the test
func (t *OLTPTest) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()

	close(t.done)
	t.status.Status = string(types.TestStatusCancelled)
}

// Status returns the current test status
func (t *OLTPTest) Status() benchmark.BenchmarkStatus {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.status
}

// GetReport returns the test report
func (t *OLTPTest) GetReport() *types.Report {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return &types.Report{
		Name:     string(t.config.TestType),
		Duration: t.config.Duration,
		Stats:    t.stats,
	}
}

// Run executes the test
func (t *OLTPTest) Run(ctx context.Context) error {
	if t.db == nil {
		return fmt.Errorf("database connection not set")
	}

	if t.logger == nil {
		return fmt.Errorf("logger not set")
	}

	t.logger.Info("Starting OLTP test",
		zap.String("test_type", string(t.config.TestType)),
		zap.Int("num_threads", t.config.NumThreads),
		zap.Duration("duration", t.config.Duration),
	)

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < t.config.NumThreads; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			t.worker(ctx, workerID)
		}(i)
	}

	// Wait for test completion or context cancellation
	select {
	case <-ctx.Done():
		close(t.done)
		wg.Wait()
		return ctx.Err()
	case <-time.After(t.config.Duration):
		close(t.done)
		wg.Wait()
		return nil
	}
}

// worker represents a test worker
func (t *OLTPTest) worker(ctx context.Context, id int) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.done:
			return
		default:
			start := time.Now()
			err := t.executeTransaction(ctx)
			elapsed := time.Since(start)

			if err != nil {
				t.stats.AddError()
				t.logger.Error("Transaction failed",
					zap.Int("worker", id),
					zap.Error(err))
				continue
			}

			t.stats.AddTransaction(elapsed)
		}
	}
}

// executeTransaction executes a single transaction based on test type
func (t *OLTPTest) executeTransaction(ctx context.Context) error {
	switch t.config.TestType {
	case types.TestTypeOLTPRead:
		return t.executeReadOnly(ctx)
	case types.TestTypeOLTPWrite:
		return t.executeWriteOnly(ctx)
	case types.TestTypeOLTPReadWrite:
		return t.executeReadWrite(ctx)
	case types.TestTypeOLTPPointSelect:
		return t.executePointSelect(ctx)
	case types.TestTypeOLTPSimpleSelect:
		return t.executeSimpleSelect(ctx)
	case types.TestTypeOLTPSumRange:
		return t.executeSumRange(ctx)
	default:
		return fmt.Errorf("unsupported test type: %s", t.config.TestType)
	}
}

// executeReadOnly executes a read-only transaction
func (t *OLTPTest) executeReadOnly(ctx context.Context) error {
	tx, err := t.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return fmt.Errorf("begin transaction failed: %w", err)
	}
	defer tx.Rollback()

	// Execute read operations
	if err := t.doReads(ctx, tx); err != nil {
		return err
	}

	return tx.Commit()
}

// executeWriteOnly executes a write-only transaction
func (t *OLTPTest) executeWriteOnly(ctx context.Context) error {
	tx, err := t.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction failed: %w", err)
	}
	defer tx.Rollback()

	// Execute write operations
	if err := t.doWrites(ctx, tx); err != nil {
		return err
	}

	return tx.Commit()
}

// executeReadWrite executes a mixed read-write transaction
func (t *OLTPTest) executeReadWrite(ctx context.Context) error {
	tx, err := t.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction failed: %w", err)
	}
	defer tx.Rollback()

	// Execute read operations
	if err := t.doReads(ctx, tx); err != nil {
		return err
	}

	// Execute write operations
	if err := t.doWrites(ctx, tx); err != nil {
		return err
	}

	return tx.Commit()
}

// executePointSelect executes point select queries
func (t *OLTPTest) executePointSelect(ctx context.Context) error {
	tx, err := t.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return fmt.Errorf("begin transaction failed: %w", err)
	}
	defer tx.Rollback()

	// Execute point selects
	if err := t.doPointSelects(ctx, tx); err != nil {
		return err
	}

	return tx.Commit()
}

// executeSimpleSelect executes simple range select queries
func (t *OLTPTest) executeSimpleSelect(ctx context.Context) error {
	tx, err := t.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return fmt.Errorf("begin transaction failed: %w", err)
	}
	defer tx.Rollback()

	// Execute simple range selects
	if err := t.doSimpleRangeSelects(ctx, tx); err != nil {
		return err
	}

	return tx.Commit()
}

// executeSumRange executes sum range queries
func (t *OLTPTest) executeSumRange(ctx context.Context) error {
	tx, err := t.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return fmt.Errorf("begin transaction failed: %w", err)
	}
	defer tx.Rollback()

	// Execute sum range queries
	if err := t.doSumRangeQueries(ctx, tx); err != nil {
		return err
	}

	return tx.Commit()
}

// doReads performs read operations
func (t *OLTPTest) doReads(ctx context.Context, tx *sql.Tx) error {
	// Implement read operations
	return nil
}

// doWrites performs write operations
func (t *OLTPTest) doWrites(ctx context.Context, tx *sql.Tx) error {
	// Implement write operations
	return nil
}

// doPointSelects performs point select operations
func (t *OLTPTest) doPointSelects(ctx context.Context, tx *sql.Tx) error {
	// Implement point select operations
	return nil
}

// doSimpleRangeSelects performs simple range select operations
func (t *OLTPTest) doSimpleRangeSelects(ctx context.Context, tx *sql.Tx) error {
	// Implement simple range select operations
	return nil
}

// doSumRangeQueries performs sum range queries
func (t *OLTPTest) doSumRangeQueries(ctx context.Context, tx *sql.Tx) error {
	// Implement sum range queries
	return nil
}

// validateConfig validates the test configuration
func validateConfig(config *OLTPTestConfig) error {
	if config == nil {
		return types.ErrInvalidConfig
	}

	// Validate test type
	switch config.TestType {
	case types.TestTypeOLTPRead,
		types.TestTypeOLTPWrite,
		types.TestTypeOLTPReadWrite,
		types.TestTypeOLTPPointSelect,
		types.TestTypeOLTPSimpleSelect,
		types.TestTypeOLTPSumRange,
		types.TestTypeOLTPOrderRange,
		types.TestTypeOLTPDistinctRange,
		types.TestTypeOLTPIndexScan,
		types.TestTypeOLTPNonIndexScan:
		// Valid test type
	default:
		return fmt.Errorf("invalid test type: %s", config.TestType)
	}

	// Validate basic parameters
	if config.TableSize <= 0 {
		return types.ErrInvalidTableSize
	}
	if config.TablesCount <= 0 {
		return types.ErrInvalidNumTables
	}
	if config.NumThreads <= 0 {
		return types.ErrInvalidNumThreads
	}

	// Validate test-specific parameters
	switch config.TestType {
	case types.TestTypeOLTPPointSelect:
		if config.PointSelects <= 0 {
			return fmt.Errorf("point selects must be positive for point select test")
		}
	case types.TestTypeOLTPSimpleSelect:
		if config.SimpleRanges <= 0 {
			return fmt.Errorf("simple ranges must be positive for simple select test")
		}
	case types.TestTypeOLTPSumRange:
		if config.SumRanges <= 0 {
			return fmt.Errorf("sum ranges must be positive for sum range test")
		}
	case types.TestTypeOLTPOrderRange:
		if config.OrderRanges <= 0 {
			return fmt.Errorf("order ranges must be positive for order range test")
		}
	case types.TestTypeOLTPDistinctRange:
		if config.DistinctRanges <= 0 {
			return fmt.Errorf("distinct ranges must be positive for distinct range test")
		}
	}

	return nil
}

// Prepare prepares the test database
func (t *OLTPTest) Prepare(ctx context.Context) error {
	// Create tables
	for i := 1; i <= t.config.TablesCount; i++ {
		if err := t.createTable(ctx, i); err != nil {
			return fmt.Errorf("failed to create table %d: %w", i, err)
		}
	}
	return nil
}

// Cleanup cleans up the test database
func (t *OLTPTest) Cleanup(ctx context.Context) error {
	// Drop tables
	for i := 1; i <= t.config.TablesCount; i++ {
		if err := t.dropTable(ctx, i); err != nil {
			return fmt.Errorf("failed to drop table %d: %w", i, err)
		}
	}
	return nil
}

// createTable creates a table
func (t *OLTPTest) createTable(ctx context.Context, tableNum int) error {
	query := fmt.Sprintf(`
		CREATE TABLE sbtest%d (
			id INTEGER NOT NULL,
			k INTEGER DEFAULT '0' NOT NULL,
			c CHAR(120) DEFAULT '' NOT NULL,
			pad CHAR(60) DEFAULT '' NOT NULL,
			PRIMARY KEY (id)
		)`, tableNum)

	if _, err := t.db.ExecContext(ctx, query); err != nil {
		return err
	}

	if t.config.SecondaryIndexes > 0 {
		query = fmt.Sprintf("CREATE INDEX k_%d ON sbtest%d(k)", tableNum, tableNum)
		if _, err := t.db.ExecContext(ctx, query); err != nil {
			return err
		}
	}

	return nil
}

// dropTable drops a table
func (t *OLTPTest) dropTable(ctx context.Context, tableNum int) error {
	query := fmt.Sprintf("DROP TABLE IF EXISTS sbtest%d", tableNum)
	_, err := t.db.ExecContext(ctx, query)
	return err
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
