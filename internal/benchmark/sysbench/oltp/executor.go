package oltp

import (
	"context"
	"database/sql"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/deadjoe/benchphant/internal/benchmark/sysbench/types"
	"go.uber.org/zap"
)

// Executor handles the execution of OLTP tests
type Executor struct {
	db     *sql.DB
	config *Config
	logger *zap.Logger

	mu       sync.RWMutex
	running  bool
	stopChan chan struct{}
	results  chan *types.Result
}

// NewExecutor creates a new OLTP test executor
func NewExecutor(db *sql.DB, config *Config, logger *zap.Logger) (*Executor, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &Executor{
		db:       db,
		config:   config,
		logger:   logger,
		stopChan: make(chan struct{}),
		results:  make(chan *types.Result, 1000),
	}, nil
}

// Start begins the test execution
func (e *Executor) Start(ctx context.Context) error {
	e.mu.Lock()
	if e.running {
		e.mu.Unlock()
		return types.ErrTestAlreadyRunning
	}
	e.running = true
	e.mu.Unlock()

	// Create worker pool
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < e.config.NumThreads; i++ {
		wg.Add(1)
		go e.worker(ctx, i, &wg)
	}

	// Start result collector
	go e.collectResults()

	// Wait for completion or context cancellation
	select {
	case <-ctx.Done():
		close(e.stopChan)
		wg.Wait()
		return ctx.Err()
	case <-time.After(e.config.Duration):
		close(e.stopChan)
		wg.Wait()
		return nil
	}
}

// Stop stops the test execution
func (e *Executor) Stop() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.running {
		return types.ErrTestNotRunning
	}

	close(e.stopChan)
	e.running = false
	return nil
}

// worker represents a single test worker
func (e *Executor) worker(ctx context.Context, id int, wg *sync.WaitGroup) {
	defer wg.Done()

	e.logger.Info("Starting worker", zap.Int("worker_id", id))

	for {
		select {
		case <-ctx.Done():
			return
		case <-e.stopChan:
			return
		default:
			// Execute test operations based on weights
			if e.config.ReadOnly || e.shouldRead() {
				if err := e.executeRead(ctx); err != nil {
					e.logger.Error("Read operation failed", zap.Error(err))
				}
			} else {
				if err := e.executeWrite(ctx); err != nil {
					e.logger.Error("Write operation failed", zap.Error(err))
				}
			}
		}
	}
}

// shouldRead determines if the next operation should be a read based on weights
func (e *Executor) shouldRead() bool {
	return e.config.ReadWeight > 0 && (e.config.WriteWeight == 0 || rand.Float64() <= e.config.ReadWeight)
}

// executeRead performs a read operation
func (e *Executor) executeRead(ctx context.Context) error {
	start := time.Now()
	var err error

	switch rand.Intn(5) {
	case 0:
		err = e.executePointSelect(ctx)
	case 1:
		err = e.executeSimpleRange(ctx)
	case 2:
		err = e.executeSumRange(ctx)
	case 3:
		err = e.executeOrderRange(ctx)
	case 4:
		err = e.executeDistinctRange(ctx)
	}

	duration := time.Since(start)
	e.recordResult("read", duration, err)
	return err
}

// executeWrite performs a write operation
func (e *Executor) executeWrite(ctx context.Context) error {
	start := time.Now()
	var err error

	switch rand.Intn(3) {
	case 0:
		err = e.executeIndexUpdate(ctx)
	case 1:
		err = e.executeNonIndexUpdate(ctx)
	case 2:
		err = e.executeDeleteInsert(ctx)
	}

	duration := time.Since(start)
	e.recordResult("write", duration, err)
	return err
}

// recordResult records a single operation result
func (e *Executor) recordResult(opType string, duration time.Duration, err error) {
	result := &types.Result{
		Type:      opType,
		Duration:  duration,
		Success:   err == nil,
		Timestamp: time.Now(),
	}
	e.results <- result
}

// collectResults collects and aggregates test results
func (e *Executor) collectResults() {
	var (
		totalOps     int64
		totalErrors  int64
		totalLatency time.Duration
		latencies    []time.Duration
		mu           sync.Mutex
	)

	ticker := time.NewTicker(e.config.ReportInterval)
	defer ticker.Stop()

	for {
		select {
		case <-e.stopChan:
			return
		case result := <-e.results:
			mu.Lock()
			totalOps++
			if !result.Success {
				totalErrors++
			}
			totalLatency += result.Duration
			latencies = append(latencies, result.Duration)
			mu.Unlock()
		case <-ticker.C:
			mu.Lock()
			if totalOps > 0 {
				avgLatency := totalLatency / time.Duration(totalOps)
				p99 := calculateP99(latencies)
				e.logger.Info("Test progress",
					zap.Int64("total_ops", totalOps),
					zap.Int64("errors", totalErrors),
					zap.Duration("avg_latency", avgLatency),
					zap.Duration("p99_latency", p99),
				)
			}
			mu.Unlock()
		}
	}
}

// calculateP99 calculates the 99th percentile latency
func calculateP99(latencies []time.Duration) time.Duration {
	if len(latencies) == 0 {
		return 0
	}
	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})
	idx := int(float64(len(latencies)) * 0.99)
	return latencies[idx]
}

// executePointSelect performs a point select query
func (e *Executor) executePointSelect(ctx context.Context) error {
	id := rand.Int63n(int64(e.config.TableSize)) + 1
	query := "SELECT id, k, c, pad FROM sbtest1 WHERE id = ?"
	_, err := e.db.ExecContext(ctx, query, id)
	return err
}

// executeSimpleRange performs a simple range query
func (e *Executor) executeSimpleRange(ctx context.Context) error {
	id := rand.Int63n(int64(e.config.TableSize-100)) + 1
	query := "SELECT id, k, c, pad FROM sbtest1 WHERE id BETWEEN ? AND ?"
	_, err := e.db.ExecContext(ctx, query, id, id+100)
	return err
}

// executeSumRange performs a sum range query
func (e *Executor) executeSumRange(ctx context.Context) error {
	id := rand.Int63n(int64(e.config.TableSize-100)) + 1
	query := "SELECT SUM(k) FROM sbtest1 WHERE id BETWEEN ? AND ?"
	_, err := e.db.ExecContext(ctx, query, id, id+100)
	return err
}

// executeOrderRange performs an ordered range query
func (e *Executor) executeOrderRange(ctx context.Context) error {
	id := rand.Int63n(int64(e.config.TableSize-100)) + 1
	query := "SELECT id, k, c, pad FROM sbtest1 WHERE id BETWEEN ? AND ? ORDER BY id"
	_, err := e.db.ExecContext(ctx, query, id, id+100)
	return err
}

// executeDistinctRange performs a distinct range query
func (e *Executor) executeDistinctRange(ctx context.Context) error {
	id := rand.Int63n(int64(e.config.TableSize-100)) + 1
	query := "SELECT DISTINCT k FROM sbtest1 WHERE id BETWEEN ? AND ?"
	_, err := e.db.ExecContext(ctx, query, id, id+100)
	return err
}

// executeIndexUpdate performs an indexed update
func (e *Executor) executeIndexUpdate(ctx context.Context) error {
	id := rand.Int63n(int64(e.config.TableSize)) + 1
	k := rand.Int31()
	query := "UPDATE sbtest1 SET k = ? WHERE id = ?"
	_, err := e.db.ExecContext(ctx, query, k, id)
	return err
}

// executeNonIndexUpdate performs a non-indexed update
func (e *Executor) executeNonIndexUpdate(ctx context.Context) error {
	id := rand.Int63n(int64(e.config.TableSize)) + 1
	c := generateRandomString(120)
	query := "UPDATE sbtest1 SET c = ? WHERE id = ?"
	_, err := e.db.ExecContext(ctx, query, c, id)
	return err
}

// executeDeleteInsert performs a delete followed by an insert
func (e *Executor) executeDeleteInsert(ctx context.Context) error {
	tx, err := e.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	id := rand.Int63n(int64(e.config.TableSize)) + 1
	deleteQuery := "DELETE FROM sbtest1 WHERE id = ?"
	if _, err := tx.ExecContext(ctx, deleteQuery, id); err != nil {
		return err
	}

	k := rand.Int31()
	c := generateRandomString(120)
	pad := generateRandomString(60)
	insertQuery := "INSERT INTO sbtest1 (id, k, c, pad) VALUES (?, ?, ?, ?)"
	if _, err := tx.ExecContext(ctx, insertQuery, id, k, c, pad); err != nil {
		return err
	}

	return tx.Commit()
}

// generateRandomString generates a random string of specified length
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
