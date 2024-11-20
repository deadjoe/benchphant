package benchmark

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/deadjoe/benchphant/internal/models"
	"go.uber.org/zap"
)

// BenchmarkStatus represents the current status of a benchmark
type BenchmarkStatus struct {
	Status   string                 `json:"status"`
	Progress float64                `json:"progress"`
	Metrics  map[string]interface{} `json:"metrics"`
}

// BenchmarkRunner represents a database benchmark
type BenchmarkRunner interface {
	// Start starts the benchmark
	Start() error
	// Stop stops the benchmark
	Stop()
	// Status returns the current benchmark status
	Status() BenchmarkStatus
}

// Benchmark represents a database benchmark implementation
type Benchmark struct {
	config     *models.Benchmark
	connection *models.DBConnection
	db         *sql.DB
	logger     *zap.Logger
	status     BenchmarkStatus
	startTime  time.Time
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	mu         sync.RWMutex
	done       chan struct{}
	ctx        context.Context
}

// NewBenchmark creates a new benchmark
func NewBenchmark(config *models.Benchmark, conn *models.DBConnection, logger *zap.Logger) *Benchmark {
	metrics := make(map[string]interface{})
	metrics["queries"] = float64(0)
	metrics["errors"] = float64(0)
	metrics["latency_sum"] = float64(0)
	metrics["latency_min"] = float64(0)
	metrics["latency_max"] = float64(0)
	metrics["latency_avg"] = float64(0)
	metrics["latency_p95"] = float64(0)
	metrics["latency_p99"] = float64(0)
	metrics["qps"] = float64(0)
	metrics["latencies"] = make([]float64, 0, 1000)

	return &Benchmark{
		config:     config,
		connection: conn,
		db:         conn.DB,
		logger:     logger,
		done:       make(chan struct{}),
		status: BenchmarkStatus{
			Status:  string(models.BenchmarkStatusPending),
			Metrics: metrics,
		},
	}
}

// Start starts the benchmark
func (b *Benchmark) Start() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Validate configuration
	if b.config.NumThreads <= 0 {
		return fmt.Errorf("number of threads must be greater than 0")
	}
	if b.config.Duration <= 0 {
		return fmt.Errorf("duration must be greater than 0")
	}
	if b.config.QueryTemplate == "" {
		return fmt.Errorf("query template cannot be empty")
	}

	// Check if already running
	if b.status.Status == string(models.BenchmarkStatusRunning) {
		return fmt.Errorf("benchmark is already running")
	}

	// Reset metrics
	b.status.Metrics = map[string]interface{}{
		"qps":         float64(0),
		"latency_avg": float64(0),
		"latency_p95": float64(0),
		"latency_p99": float64(0),
		"errors":      float64(0),
	}

	// Initialize benchmark
	if b.connection == nil {
		b.status.Status = string(models.BenchmarkStatusFailed)
		return fmt.Errorf("database connection is nil")
	}

	// Reset done channel
	b.done = make(chan struct{})

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), b.config.Duration)
	b.ctx = ctx
	b.cancel = cancel

	// Prepare statement
	stmt, err := b.db.PrepareContext(b.ctx, b.config.QueryTemplate)
	if err != nil {
		b.status.Status = string(models.BenchmarkStatusFailed)
		return fmt.Errorf("failed to prepare statement: %w", err)
	}

	// Start benchmark
	b.startTime = time.Now()
	b.status.Status = string(models.BenchmarkStatusRunning)
	b.status.Progress = 0

	// Start workers
	b.wg.Add(b.config.NumThreads)
	for i := 0; i < b.config.NumThreads; i++ {
		go b.worker(b.ctx, stmt, i)
	}

	// Start progress updater
	go b.updateProgress(b.ctx, stmt)

	return nil
}

// Stop stops the benchmark
func (b *Benchmark) Stop() {
	b.mu.Lock()
	if b.status.Status == string(models.BenchmarkStatusRunning) {
		b.cancel()
	}
	b.mu.Unlock()
}

// Status returns the current benchmark status
func (b *Benchmark) Status() BenchmarkStatus {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.status
}

// worker runs queries in a loop
func (b *Benchmark) worker(ctx context.Context, stmt *sql.Stmt, id int) {
	b.logger.Debug("worker started", zap.Int("worker_id", id))
	defer func() {
		b.logger.Debug("worker stopped", zap.Int("worker_id", id))
		b.wg.Done()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			if err := b.runQuery(ctx, stmt); err != nil {
				if err == context.Canceled || err == context.DeadlineExceeded {
					return
				}
				b.logger.Error("query failed", zap.Error(err), zap.Int("worker", id))
				b.mu.Lock()
				b.status.Metrics["errors"] = b.status.Metrics["errors"].(float64) + 1
				b.status.Status = string(models.BenchmarkStatusFailed)
				b.mu.Unlock()
				b.cancel() // Cancel other workers when one fails
				return
			}
			// Update metrics for successful query
			b.mu.Lock()
			b.status.Metrics["qps"] = b.status.Metrics["qps"].(float64) + 1
			b.mu.Unlock()

			// Sleep a tiny bit to avoid overwhelming the mock
			time.Sleep(time.Millisecond)
		}
	}
}

// runQuery executes a single query and updates metrics
func (b *Benchmark) runQuery(ctx context.Context, stmt *sql.Stmt) error {
	start := time.Now()
	_, err := stmt.ExecContext(ctx)
	duration := time.Since(start)

	if err != nil {
		if err == context.Canceled || err == context.DeadlineExceeded {
			return err
		}
		return fmt.Errorf("query execution failed: %w", err)
	}

	// Update latency metrics
	b.mu.Lock()
	latencyAvg := b.status.Metrics["latency_avg"].(float64)
	qps := b.status.Metrics["qps"].(float64)
	b.status.Metrics["latency_avg"] = (latencyAvg*qps + duration.Seconds()) / (qps + 1)
	b.status.Metrics["latency_p95"] = duration.Seconds() // Simplified for now
	b.status.Metrics["latency_p99"] = duration.Seconds() // Simplified for now
	b.mu.Unlock()

	return nil
}

// updateProgress updates the benchmark progress
func (b *Benchmark) updateProgress(ctx context.Context, stmt *sql.Stmt) {
	defer func() {
		stmt.Close()
		b.wg.Wait() // Wait for all workers to finish before updating final status
		b.mu.Lock()
		if b.status.Status != string(models.BenchmarkStatusFailed) {
			if ctx.Err() == context.Canceled {
				b.status.Status = string(models.BenchmarkStatusCancelled)
			} else {
				b.status.Status = string(models.BenchmarkStatusCompleted)
			}
		}
		b.status.Progress = 100
		b.mu.Unlock()
		close(b.done)
	}()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			b.mu.Lock()
			elapsed := time.Since(b.startTime)
			progress := (elapsed.Seconds() / b.config.Duration.Seconds()) * 100
			b.status.Progress = math.Min(100, progress)
			b.mu.Unlock()
		}
	}
}

// Factory represents a benchmark factory
type Factory interface {
	// Name returns the name of the benchmark type
	Name() string
	// Create creates a new benchmark instance
	Create(config *models.Benchmark, conn *models.DBConnection, logger *zap.Logger) (BenchmarkRunner, error)
}

var factories = make(map[string]Factory)

// RegisterFactory registers a benchmark factory
func RegisterFactory(name string, factory Factory) {
	factories[name] = factory
}

// Result represents the result of a benchmark run
type Result struct {
	Name              string                 `json:"name"`
	Duration          time.Duration          `json:"duration"`
	TotalTransactions int64                  `json:"total_transactions"`
	TPS               float64                `json:"tps"`
	LatencyAvg        time.Duration          `json:"latency_avg"`
	LatencyP95        time.Duration          `json:"latency_p95"`
	LatencyP99        time.Duration          `json:"latency_p99"`
	Errors            int64                  `json:"errors"`
	StartTime         time.Time              `json:"start_time"`
	EndTime           time.Time              `json:"end_time"`
	Metrics           map[string]interface{} `json:"metrics"`
}
