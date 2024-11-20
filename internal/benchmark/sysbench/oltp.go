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
	"time"

	"github.com/deadjoe/benchphant/internal/benchmark/sysbench/types"
	"go.uber.org/zap"
)

// OLTPTestConfig represents configuration for OLTP tests
type OLTPTestConfig struct {
	types.TestConfig
	TableSize        int     // Number of rows per table
	TablesCount      int     // Number of tables
	DistinctRanges   int     // Number of distinct ranges for range queries
	SimpleRanges     int     // Number of simple ranges for range queries
	SumRanges        int     // Number of sum ranges for range queries
	OrderRanges      int     // Number of order ranges for range queries
	PointSelects     int     // Number of point selects per transaction
	SimpleSelects    int     // Number of simple range selects per transaction
	IndexUpdates     int     // Number of index updates per transaction
	NonIndexUpdates  int     // Number of non-index updates per transaction
	Inserts          int     // Number of inserts per transaction
	Deletes          int     // Number of deletes per transaction
	ReadOnly         bool    // Whether to perform read-only transactions
	WriteOnly        bool    // Whether to perform write-only transactions
	AutoInc          bool    // Whether to use auto-increment column
	SecondaryIndexes int     // Number of secondary indexes
	DeleteInserts    bool    // Whether to delete and insert instead of update
	RangeSize        int     // Size of ranges for range queries
	RangeSelects     bool    // Whether to perform range selects
	JoinSelects      bool    // Whether to perform join selects
	SkipTrx          bool    // Whether to skip BEGIN/COMMIT statements
	TrxRate          float64 // Transaction rate limit in transactions per second
}

// NewOLTPTestConfig creates a new OLTP test configuration with default values
func NewOLTPTestConfig() *OLTPTestConfig {
	return &OLTPTestConfig{
		TestConfig: types.TestConfig{
			Threads:  1,
			Database: "sbtest",
			TestType: types.TestTypeOLTPReadWrite,
		},
		TableSize:        10000,
		TablesCount:      1,
		DistinctRanges:   100,
		SimpleRanges:     100,
		SumRanges:        100,
		OrderRanges:      100,
		PointSelects:     10,
		SimpleSelects:    1,
		IndexUpdates:     1,
		NonIndexUpdates:  1,
		Inserts:          1,
		Deletes:          1,
		ReadOnly:         false,
		WriteOnly:        false,
		AutoInc:          true,
		SecondaryIndexes: 0,
		DeleteInserts:    false,
		RangeSize:        100,
		RangeSelects:     true,
		JoinSelects:      false,
		SkipTrx:          false,
		TrxRate:          0, // No limit
	}
}

// OLTPTest represents an OLTP test
type OLTPTest struct {
	db     *sql.DB
	config *OLTPTestConfig
	logger *zap.Logger
	stats  *types.TestStats
}

// NewOLTPTest creates a new OLTP test
func NewOLTPTest(db *sql.DB, config *OLTPTestConfig, logger *zap.Logger) (*OLTPTest, error) {
	if err := validateConfig(config); err != nil {
		return nil, err
	}

	return &OLTPTest{
		db:     db,
		config: config,
		logger: logger,
		stats:  &types.TestStats{},
	}, nil
}

// validateConfig validates the test configuration
func validateConfig(config *OLTPTestConfig) error {
	if config == nil {
		return types.ErrInvalidConfig
	}
	if config.TableSize <= 0 {
		return types.ErrInvalidTableSize
	}
	if config.TablesCount <= 0 {
		return types.ErrInvalidNumTables
	}
	if config.Threads <= 0 {
		return types.ErrInvalidNumThreads
	}
	if config.Duration <= 0 {
		return types.ErrInvalidDuration
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

// Run runs the test
func (t *OLTPTest) Run(ctx context.Context) error {
	t.logger.Info("Starting OLTP test",
		zap.String("test_type", string(t.config.TestType)),
		zap.Int("num_threads", t.config.Threads),
		zap.Duration("duration", t.config.Duration),
	)

	// Create worker channels
	workCh := make(chan struct{}, t.config.Threads)
	errCh := make(chan error, t.config.Threads)

	// Start workers
	for i := 0; i < t.config.Threads; i++ {
		go t.worker(ctx, workCh, errCh)
	}

	// Run test for the specified duration
	timer := time.NewTimer(t.config.Duration)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-errCh:
			return err
		case <-timer.C:
			close(workCh)
			return nil
		default:
			select {
			case workCh <- struct{}{}:
			default:
				// Channel is full, skip
			}
		}
	}
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

// Stats returns the test statistics
func (t *OLTPTest) Stats() *types.TestStats {
	return t.stats
}

func (t *OLTPTest) worker(ctx context.Context, workCh <-chan struct{}, errCh chan<- error) {
	for range workCh {
		start := time.Now()
		if err := t.executeTransaction(ctx); err != nil {
			errCh <- err
			return
		}
		t.stats.AddTransaction(time.Since(start))
	}
}

func (t *OLTPTest) executeTransaction(ctx context.Context) error {
	switch t.config.TestType {
	case types.TestTypeOLTPRead:
		return t.executeReadOnlyTransaction(ctx)
	case types.TestTypeOLTPWrite:
		return t.executeWriteOnlyTransaction(ctx)
	case types.TestTypeOLTPReadWrite:
		return t.executeReadWriteTransaction(ctx)
	case types.TestTypeOLTPPointSelect:
		return t.executePointSelectTransaction(ctx)
	case types.TestTypeOLTPSimpleSelect:
		return t.executeSimpleSelectTransaction(ctx)
	case types.TestTypeOLTPSumRange:
		return t.executeSumRangeTransaction(ctx)
	case types.TestTypeOLTPOrderRange:
		return t.executeOrderRangeTransaction(ctx)
	case types.TestTypeOLTPDistinctRange:
		return t.executeDistinctRangeTransaction(ctx)
	case types.TestTypeOLTPIndexScan:
		return t.executeIndexScanTransaction(ctx)
	case types.TestTypeOLTPNonIndexScan:
		return t.executeNonIndexScanTransaction(ctx)
	default:
		return fmt.Errorf("unsupported OLTP test type: %s", t.config.TestType)
	}
}

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

	if t.config.UseSecondaryIndex {
		query = fmt.Sprintf("CREATE INDEX k_%d ON sbtest%d(k)", tableNum, tableNum)
		if _, err := t.db.ExecContext(ctx, query); err != nil {
			return err
		}
	}

	return nil
}

func (t *OLTPTest) dropTable(ctx context.Context, tableNum int) error {
	query := fmt.Sprintf("DROP TABLE IF EXISTS sbtest%d", tableNum)
	_, err := t.db.ExecContext(ctx, query)
	return err
}

func (t *OLTPTest) executeReadOnlyTransaction(ctx context.Context) error {
	// Implementation follows sysbench's oltp_read_only.lua
	if !t.config.SkipTrx {
		tx, err := t.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()
	}

	// Point selects
	for i := 0; i < t.config.PointSelects; i++ {
		tableNum := rand.Intn(t.config.TablesCount) + 1
		id := rand.Intn(t.config.TableSize) + 1
		query := fmt.Sprintf("SELECT c FROM sbtest%d WHERE id = ?", tableNum)
		if _, err := t.db.QueryContext(ctx, query, id); err != nil {
			return err
		}
	}

	// Range sum
	tableNum := rand.Intn(t.config.TablesCount) + 1
	id := rand.Intn(t.config.TableSize-t.config.RangeSize) + 1
	query := fmt.Sprintf(`
		SELECT COUNT(*) FROM sbtest%d 
		WHERE id BETWEEN ? AND ?`,
		tableNum)
	if _, err := t.db.QueryContext(ctx, query, id, id+t.config.RangeSize); err != nil {
		return err
	}

	if !t.config.SkipTrx {
		return nil
	}
	return nil
}

func (t *OLTPTest) executeReadWriteTransaction(ctx context.Context) error {
	// Implementation follows sysbench's oltp_read_write.lua
	if !t.config.SkipTrx {
		tx, err := t.db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("begin transaction failed: %w", err)
		}
		defer tx.Rollback()
	}

	// Execute different types of queries
	if err := t.executePointSelects(ctx); err != nil {
		return err
	}

	if err := t.executeRangeSum(ctx); err != nil {
		return err
	}

	if err := t.executeIndexUpdate(ctx); err != nil {
		return err
	}

	if err := t.executeNonIndexUpdate(ctx); err != nil {
		return err
	}

	if !t.config.SkipTrx {
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit transaction failed: %w", err)
		}
	}

	return nil
}

func (t *OLTPTest) executePointSelects(ctx context.Context) error {
	for i := 0; i < t.config.PointSelects; i++ {
		tableNum := rand.Intn(t.config.TablesCount) + 1
		id := rand.Intn(t.config.TableSize) + 1
		query := fmt.Sprintf("SELECT c FROM sbtest%d WHERE id = ?", tableNum)
		if _, err := t.db.QueryContext(ctx, query, id); err != nil {
			return fmt.Errorf("point select failed: %w", err)
		}
	}
	return nil
}

func (t *OLTPTest) executeRangeSum(ctx context.Context) error {
	tableNum := rand.Intn(t.config.TablesCount) + 1
	id := rand.Intn(t.config.TableSize-t.config.RangeSize) + 1
	query := fmt.Sprintf(`
		SELECT COUNT(*) FROM sbtest%d 
		WHERE id BETWEEN ? AND ?`,
		tableNum)
	if _, err := t.db.QueryContext(ctx, query, id, id+t.config.RangeSize); err != nil {
		return fmt.Errorf("range sum failed: %w", err)
	}
	return nil
}

func (t *OLTPTest) executeIndexUpdate(ctx context.Context) error {
	tableNum := rand.Intn(t.config.TablesCount) + 1
	id := rand.Intn(t.config.TableSize) + 1
	query := fmt.Sprintf("UPDATE sbtest%d SET k=k+1 WHERE id=?", tableNum)
	if _, err := t.db.ExecContext(ctx, query, id); err != nil {
		return fmt.Errorf("index update failed: %w", err)
	}
	return nil
}

func (t *OLTPTest) executeNonIndexUpdate(ctx context.Context) error {
	tableNum := rand.Intn(t.config.TablesCount) + 1
	id := rand.Intn(t.config.TableSize) + 1
	c := randomString(120)
	query := fmt.Sprintf("UPDATE sbtest%d SET c=? WHERE id=?", tableNum)
	if _, err := t.db.ExecContext(ctx, query, c, id); err != nil {
		return fmt.Errorf("non-index update failed: %w", err)
	}
	return nil
}

func (t *OLTPTest) executeWriteOnlyTransaction(ctx context.Context) error {
	// Implementation follows sysbench's oltp_write_only.lua
	if !t.config.SkipTrx {
		tx, err := t.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()
	}

	// Update index
	tableNum := rand.Intn(t.config.TablesCount) + 1
	id := rand.Intn(t.config.TableSize) + 1
	query := fmt.Sprintf("UPDATE sbtest%d SET k=k+1 WHERE id=?", tableNum)
	if _, err := t.db.ExecContext(ctx, query, id); err != nil {
		return err
	}

	// Update non-index
	tableNum = rand.Intn(t.config.TablesCount) + 1
	id = rand.Intn(t.config.TableSize) + 1
	c := randomString(120)
	query = fmt.Sprintf("UPDATE sbtest%d SET c=? WHERE id=?", tableNum)
	if _, err := t.db.ExecContext(ctx, query, c, id); err != nil {
		return err
	}

	if !t.config.SkipTrx {
		return nil
	}
	return nil
}

func (t *OLTPTest) executePointSelectTransaction(ctx context.Context) error {
	// Implementation follows sysbench's oltp_point_select.lua
	if !t.config.SkipTrx {
		tx, err := t.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()
	}

	// Point selects
	for i := 0; i < t.config.PointSelects; i++ {
		tableNum := rand.Intn(t.config.TablesCount) + 1
		id := rand.Intn(t.config.TableSize) + 1
		query := fmt.Sprintf("SELECT c FROM sbtest%d WHERE id = ?", tableNum)
		if _, err := t.db.QueryContext(ctx, query, id); err != nil {
			return err
		}
	}

	if !t.config.SkipTrx {
		return nil
	}
	return nil
}

func (t *OLTPTest) executeSimpleSelectTransaction(ctx context.Context) error {
	// Implementation follows sysbench's oltp_simple_select.lua
	if !t.config.SkipTrx {
		tx, err := t.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()
	}

	// Simple selects
	for i := 0; i < t.config.SimpleSelects; i++ {
		tableNum := rand.Intn(t.config.TablesCount) + 1
		id := rand.Intn(t.config.TableSize) + 1
		query := fmt.Sprintf("SELECT c FROM sbtest%d WHERE id BETWEEN ? AND ?", tableNum)
		if _, err := t.db.QueryContext(ctx, query, id, id+t.config.RangeSize); err != nil {
			return err
		}
	}

	if !t.config.SkipTrx {
		return nil
	}
	return nil
}

func (t *OLTPTest) executeSumRangeTransaction(ctx context.Context) error {
	// Implementation follows sysbench's oltp_sum_range.lua
	if !t.config.SkipTrx {
		tx, err := t.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()
	}

	// Sum range
	tableNum := rand.Intn(t.config.TablesCount) + 1
	id := rand.Intn(t.config.TableSize-t.config.RangeSize) + 1
	query := fmt.Sprintf(`
		SELECT SUM(k) FROM sbtest%d 
		WHERE id BETWEEN ? AND ?`,
		tableNum)
	if _, err := t.db.QueryContext(ctx, query, id, id+t.config.RangeSize); err != nil {
		return err
	}

	if !t.config.SkipTrx {
		return nil
	}
	return nil
}

func (t *OLTPTest) executeOrderRangeTransaction(ctx context.Context) error {
	// Implementation follows sysbench's oltp_order_range.lua
	if !t.config.SkipTrx {
		tx, err := t.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()
	}

	// Order range
	tableNum := rand.Intn(t.config.TablesCount) + 1
	id := rand.Intn(t.config.TableSize-t.config.RangeSize) + 1
	query := fmt.Sprintf(`
		SELECT c FROM sbtest%d 
		WHERE id BETWEEN ? AND ? ORDER BY c`,
		tableNum)
	if _, err := t.db.QueryContext(ctx, query, id, id+t.config.RangeSize); err != nil {
		return err
	}

	if !t.config.SkipTrx {
		return nil
	}
	return nil
}

func (t *OLTPTest) executeDistinctRangeTransaction(ctx context.Context) error {
	// Implementation follows sysbench's oltp_distinct_range.lua
	if !t.config.SkipTrx {
		tx, err := t.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()
	}

	// Distinct range
	tableNum := rand.Intn(t.config.TablesCount) + 1
	id := rand.Intn(t.config.TableSize-t.config.RangeSize) + 1
	query := fmt.Sprintf(`
		SELECT DISTINCT c FROM sbtest%d 
		WHERE id BETWEEN ? AND ?`,
		tableNum)
	if _, err := t.db.QueryContext(ctx, query, id, id+t.config.RangeSize); err != nil {
		return err
	}

	if !t.config.SkipTrx {
		return nil
	}
	return nil
}

func (t *OLTPTest) executeIndexScanTransaction(ctx context.Context) error {
	// Implementation follows sysbench's oltp_index_scan.lua
	if !t.config.SkipTrx {
		tx, err := t.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()
	}

	// Index scan
	tableNum := rand.Intn(t.config.TablesCount) + 1
	id := rand.Intn(t.config.TableSize-t.config.RangeSize) + 1
	query := fmt.Sprintf(`
		SELECT c FROM sbtest%d 
		WHERE k BETWEEN ? AND ?`,
		tableNum)
	if _, err := t.db.QueryContext(ctx, query, id, id+t.config.RangeSize); err != nil {
		return err
	}

	if !t.config.SkipTrx {
		return nil
	}
	return nil
}

func (t *OLTPTest) executeNonIndexScanTransaction(ctx context.Context) error {
	// Implementation follows sysbench's oltp_non_index_scan.lua
	if !t.config.SkipTrx {
		tx, err := t.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()
	}

	// Non-index scan
	tableNum := rand.Intn(t.config.TablesCount) + 1
	id := rand.Intn(t.config.TableSize-t.config.RangeSize) + 1
	query := fmt.Sprintf(`
		SELECT c FROM sbtest%d 
		WHERE c BETWEEN ? AND ?`,
		tableNum)
	if _, err := t.db.QueryContext(ctx, query, id, id+t.config.RangeSize); err != nil {
		return err
	}

	if !t.config.SkipTrx {
		return nil
	}
	return nil
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
