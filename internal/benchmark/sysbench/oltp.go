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
	TestConfig
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
		TestConfig: TestConfig{
			Threads:  1,
			Database: "sbtest",
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

	executor *oltp.Executor
}

// NewOLTPTest creates a new OLTP test
func NewOLTPTest(db *sql.DB, config *OLTPTestConfig, logger *zap.Logger) (*OLTPTest, error) {
	if err := validateConfig(config); err != nil {
		return nil, err
	}

	executor, err := oltp.NewExecutor(db, config, logger)
	if err != nil {
		return nil, err
	}

	return &OLTPTest{
		db:       db,
		config:   config,
		logger:   logger,
		executor: executor,
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

	if err := t.executor.Start(ctx); err != nil {
		return fmt.Errorf("failed to run test: %w", err)
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

// Stats returns the test statistics
func (t *OLTPTest) Stats() *types.TestStats {
	return t.executor.Stats()
}

func (t *OLTPTest) worker(ctx context.Context, workCh <-chan struct{}, errCh chan<- error) {
	for range workCh {
		start := time.Now()
		if err := t.executeTransaction(ctx); err != nil {
			errCh <- err
			return
		}
		t.executor.AddTransaction(time.Since(start))
	}
}

func (t *OLTPTest) executeTransaction(ctx context.Context) error {
	switch t.config.TestType {
	case types.ReadOnly:
		return t.executeReadOnlyTransaction(ctx)
	case types.ReadWrite:
		return t.executeReadWriteTransaction(ctx)
	case types.WriteOnly:
		return t.executeWriteOnlyTransaction(ctx)
	case types.PointSelect:
		return t.executePointSelectTransaction(ctx)
	case types.SimpleRanges:
		return t.executeSimpleRangesTransaction(ctx)
	case types.SumRanges:
		return t.executeSumRangesTransaction(ctx)
	case types.OrderRanges:
		return t.executeOrderRangesTransaction(ctx)
	case types.DistinctRanges:
		return t.executeDistinctRangesTransaction(ctx)
	case types.IndexUpdates:
		return t.executeIndexUpdatesTransaction(ctx)
	case types.NonIndexUpdates:
		return t.executeNonIndexUpdatesTransaction(ctx)
	default:
		return fmt.Errorf("unsupported OLTP test type: %s", t.config.TestType)
	}
}

func (t *OLTPTest) report() {
	stats := t.executor.Stats()
	t.logger.Info("Test progress",
		zap.Float64("tps", stats.TPS),
		zap.Duration("latency_avg", stats.LatencyAvg),
		zap.Duration("latency_p95", stats.LatencyP95),
		zap.Duration("latency_p99", stats.LatencyP99),
		zap.Int64("total_transactions", stats.TotalTransactions),
		zap.Int64("errors", stats.Errors),
	)
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

// Transaction implementations
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
	for i := 0; i < t.config.NumPoints; i++ {
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
			return err
		}
		defer tx.Rollback()
	}

	// Point selects
	for i := 0; i < t.config.NumPoints; i++ {
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

	// Update index
	tableNum = rand.Intn(t.config.TablesCount) + 1
	id = rand.Intn(t.config.TableSize) + 1
	k := rand.Intn(t.config.TableSize) + 1
	query = fmt.Sprintf("UPDATE sbtest%d SET k=k+1 WHERE id=?", tableNum)
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

// Other transaction type implementations...
// (Point Select, Simple Ranges, Sum Ranges, Order Ranges, Distinct Ranges, Index Updates, Non-Index Updates)

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// RunSysbench runs the sysbench command
func (t *OLTPTest) RunSysbench(ctx context.Context) (*types.Report, error) {
	return t.Run(ctx)
}

// PrepareSysbench prepares the database for the test
func (t *OLTPTest) PrepareSysbench(ctx context.Context) error {
	// Implementation for database preparation
	return nil
}

// CleanupSysbench cleans up the database after the test
func (t *OLTPTest) CleanupSysbench(ctx context.Context) error {
	// Implementation for database cleanup
	return nil
}

// buildBaseArgs builds the base sysbench arguments for OLTP tests
func (t *OLTPTest) buildBaseArgs() []string {
	args := []string{
		"--db-driver=mysql",
		"--mysql-db=" + t.config.Database,
		"--mysql-user=" + t.config.Username,
		"--mysql-password=" + t.config.Password,
		"--mysql-host=" + t.config.Host,
		"--mysql-port=" + strconv.Itoa(t.config.Port),
		"--threads=" + strconv.Itoa(t.config.Threads),
		"--tables=" + strconv.Itoa(t.config.TablesCount),
		"--table-size=" + strconv.Itoa(t.config.TableSize),
	}

	if t.config.TrxRate > 0 {
		args = append(args, "--rate="+strconv.FormatFloat(t.config.TrxRate, 'f', 2, 64))
	}

	if t.config.SkipTrx {
		args = append(args, "--skip-trx=on")
	}

	if t.config.AutoInc {
		args = append(args, "--auto-inc=on")
	}

	if t.config.SecondaryIndexes > 0 {
		args = append(args, "--secondary="+strconv.Itoa(t.config.SecondaryIndexes))
	}

	return args
}

// parseOutput parses sysbench output and returns statistics
func (t *OLTPTest) parseOutput(output string) types.Stats {
	stats := types.Stats{}

	// Regular expressions for parsing output
	tpsRegex := regexp.MustCompile(`transactions:\s+\d+\s+\((\d+\.\d+)\s+per sec\.\)`)
	latencyRegex := regexp.MustCompile(`avg:\s+(\d+\.\d+)ms`)
	p95Regex := regexp.MustCompile(`95th percentile:\s+(\d+\.\d+)ms`)
	p99Regex := regexp.MustCompile(`99th percentile:\s+(\d+\.\d+)ms`)
	totalTxRegex := regexp.MustCompile(`transactions:\s+(\d+)\s+`)
	errorsRegex := regexp.MustCompile(`errors:\s+(\d+)\s+`)

	// Extract values using regex
	if matches := tpsRegex.FindStringSubmatch(output); len(matches) > 1 {
		if tps, err := strconv.ParseFloat(matches[1], 64); err == nil {
			stats.TPS = tps
		}
	}

	if matches := latencyRegex.FindStringSubmatch(output); len(matches) > 1 {
		if latency, err := strconv.ParseFloat(matches[1], 64); err == nil {
			stats.LatencyAvg = time.Duration(latency * float64(time.Millisecond))
		}
	}

	if matches := p95Regex.FindStringSubmatch(output); len(matches) > 1 {
		if latency, err := strconv.ParseFloat(matches[1], 64); err == nil {
			stats.LatencyP95 = time.Duration(latency * float64(time.Millisecond))
		}
	}

	if matches := p99Regex.FindStringSubmatch(output); len(matches) > 1 {
		if latency, err := strconv.ParseFloat(matches[1], 64); err == nil {
			stats.LatencyP99 = time.Duration(latency * float64(time.Millisecond))
		}
	}

	if matches := totalTxRegex.FindStringSubmatch(output); len(matches) > 1 {
		if total, err := strconv.ParseInt(matches[1], 10, 64); err == nil {
			stats.TotalTransactions = total
		}
	}

	if matches := errorsRegex.FindStringSubmatch(output); len(matches) > 1 {
		if errors, err := strconv.ParseInt(matches[1], 10, 64); err == nil {
			stats.Errors = errors
		}
	}

	return stats
}

// GetStats returns the current test statistics
func (t *OLTPTest) GetStats() types.Stats {
	return t.executor.Stats()
}

// Reset resets the test statistics
func (t *OLTPTest) Reset() {
	t.executor.Reset()
}
