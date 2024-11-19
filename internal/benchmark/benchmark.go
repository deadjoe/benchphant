package benchmark

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/deadjoe/benchphant/internal/models"
	"go.uber.org/zap"
)

// Status represents the benchmark status
type Status string

const (
	// StatusIdle means the benchmark is not running
	StatusIdle Status = "idle"
	// StatusRunning means the benchmark is currently running
	StatusRunning Status = "running"
	// StatusFinished means the benchmark has finished
	StatusFinished Status = "finished"
)

// Config represents benchmark configuration
type Config struct {
	// Duration is the duration of the benchmark
	Duration time.Duration `json:"duration"`

	// Concurrency is the number of concurrent workers
	Concurrency int `json:"concurrency"`

	// QueryRate is the target number of queries per second
	QueryRate int `json:"query_rate"`

	// Queries is the list of SQL queries to execute
	Queries []string `json:"queries"`

	// QueryDistribution is the type of query distribution to use
	QueryDistribution QueryDistributionType `json:"query_distribution"`

	// QueryWeights is the list of weights for each query when using weighted distribution
	QueryWeights []float64 `json:"query_weights,omitempty"`

	// WarmupTime is the duration to warm up before starting measurements
	WarmupTime time.Duration `json:"warmup_time"`

	// Transactions is the list of transactions to execute
	Transactions []*Transaction `json:"transactions"`

	// TransactionRate is the target number of transactions per second
	TransactionRate int `json:"transaction_rate"`

	// TransactionDistribution is the type of transaction distribution to use
	TransactionDistribution string `json:"transaction_distribution"`

	// Distribution is the TPC-C transaction distribution configuration
	Distribution *TPCCDistribution `json:"distribution"`
}

// Result represents benchmark results
type Result struct {
	StartTime       time.Time                `json:"start_time"`
	EndTime         time.Time                `json:"end_time"`
	Duration        time.Duration            `json:"duration"`
	WarmupTime      time.Duration            `json:"warmup_time"`
	Stats           *TransactionStats        `json:"stats"`
	QueryStats      map[string]QueryStats   `json:"query_stats"`
}

// Benchmark represents a database benchmark
type Benchmark struct {
	config        *Config
	connection    *models.DBConnection
	pool          *models.ConnectionManager
	status        Status
	result        *Result
	logger        *zap.Logger
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	mu            sync.RWMutex
	txExecutor    *TransactionExecutor
}

// New creates a new benchmark
func New(config *Config, conn *models.DBConnection, logger *zap.Logger) (*Benchmark, error) {
	if err := validateConfig(config); err != nil {
		return nil, err
	}

	// Get connection pool for the database
	connManager, err := models.NewConnectionManager(conn)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection manager: %w", err)
	}

	stats := &TransactionStats{}
	result := &Result{
		Stats: stats,
		QueryStats: make(map[string]QueryStats),
	}

	zapLogger := NewZapLogger(logger)
	txExecutor := NewTransactionExecutor(conn.DB(), stats, zapLogger)

	b := &Benchmark{
		config:     config,
		connection: conn,
		pool:      connManager,
		status:    StatusIdle,
		result:    result,
		logger:    logger,
		txExecutor: txExecutor,
	}

	return b, nil
}

// Start starts the benchmark
func (b *Benchmark) Start(ctx context.Context) error {
	b.mu.Lock()
	if b.status == StatusRunning {
		b.mu.Unlock()
		return fmt.Errorf("benchmark is already running")
	}

	b.status = StatusRunning
	b.result.StartTime = time.Now()
	b.mu.Unlock()

	ctx, b.cancel = context.WithTimeout(ctx, b.config.Duration)
	defer b.cancel()

	// Start worker goroutines
	for i := 0; i < b.config.Concurrency; i++ {
		b.wg.Add(1)
		go b.worker(ctx)
	}

	// Wait for all workers to finish
	b.wg.Wait()

	b.mu.Lock()
	b.status = StatusFinished
	b.result.EndTime = time.Now()
	b.result.Duration = b.result.EndTime.Sub(b.result.StartTime)
	b.mu.Unlock()

	return nil
}

// Stop stops the benchmark
func (b *Benchmark) Stop() {
	if b.cancel != nil {
		b.cancel()
	}
}

// Status returns the current benchmark status
func (b *Benchmark) Status() Status {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.status
}

// Result returns the benchmark results
func (b *Benchmark) Result() *Result {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.result
}

// worker runs transactions in a loop
func (b *Benchmark) worker(ctx context.Context) {
	defer b.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			tx := b.selectTransaction()
			if tx == nil {
				continue
			}

			// Add think time
			if tx.ThinkTime > 0 {
				time.Sleep(tx.ThinkTime)
			}

			// Add keying time
			if tx.KeyingTime > 0 {
				time.Sleep(tx.KeyingTime)
			}

			if err := b.txExecutor.Execute(ctx, tx); err != nil {
				b.logger.Error("Failed to execute transaction",
					zap.String("type", tx.Type),
					zap.Error(err))
			}
		}
	}
}

// selectTransaction selects a transaction based on the configured distribution
func (b *Benchmark) selectTransaction() *Transaction {
	if b.config.Distribution == nil {
		b.config.Distribution = NewTPCCDistribution()
	}

	txType := b.config.Distribution.SelectTransactionType()
	for _, tx := range b.config.Transactions {
		if tx.Type == string(txType) {
			return tx
		}
	}
	return nil
}

// validateConfig validates the benchmark configuration
func validateConfig(config *Config) error {
	if config == nil {
		return fmt.Errorf("config is required")
	}
	if config.Duration <= 0 {
		return fmt.Errorf("duration must be greater than 0")
	}
	if config.Concurrency <= 0 {
		return fmt.Errorf("concurrency must be greater than 0")
	}
	if config.QueryRate <= 0 {
		return fmt.Errorf("query rate must be greater than 0")
	}
	if len(config.Transactions) == 0 {
		return fmt.Errorf("at least one transaction is required")
	}
	return nil
}
