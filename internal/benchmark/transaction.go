package benchmark

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Transaction represents a TPC-C transaction
type Transaction struct {
	Type           string         `json:"type"`
	Weight         float64        `json:"weight"`
	Statements     []string       `json:"statements"`
	ThinkTime      time.Duration  `json:"think_time"`
	KeyingTime     time.Duration  `json:"keying_time"`
}

// TransactionExecutor handles transaction execution and retries
type TransactionExecutor struct {
	tx     *sql.Tx
	db     *sql.DB
	stats  *TransactionStats
	logger Logger
}

// NewTransactionExecutor creates a new transaction executor
func NewTransactionExecutor(db *sql.DB, stats *TransactionStats, logger Logger) *TransactionExecutor {
	return &TransactionExecutor{
		db:     db,
		stats:  stats,
		logger: logger,
	}
}

// Execute executes a transaction with all its statements
func (e *TransactionExecutor) Execute(ctx context.Context, transaction *Transaction) error {
	start := time.Now()
	var err error

	// Start transaction
	e.tx, err = e.db.BeginTx(ctx, nil)
	if err != nil {
		e.stats.FailedTransactions++
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Execute each statement
	for _, statement := range transaction.Statements {
		retries := 0
		for {
			err = e.executeStatement(ctx, statement)
			if err == nil {
				break
			}

			// Check if it's a deadlock and we should retry
			if isDeadlock(err) && retries < 3 {
				e.stats.LockStats.DeadlockCount++
				e.stats.LockStats.RetryCount++
				retries++
				e.logger.Warn("Deadlock detected, retrying transaction statement",
					"transaction", transaction.Type,
					"statement", statement,
					"retry", retries)
				continue
			}

			// Rollback on error
			if rbErr := e.tx.Rollback(); rbErr != nil {
				e.logger.Error("Failed to rollback transaction",
					"error", rbErr,
					"original_error", err)
			}
			e.stats.FailedTransactions++
			return fmt.Errorf("failed to execute transaction statement: %w", err)
		}
	}

	// Commit transaction
	if err = e.tx.Commit(); err != nil {
		e.stats.FailedTransactions++
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Update statistics
	duration := time.Since(start)
	e.stats.SuccessfulTransactions++
	e.stats.TotalTransactions++
	e.updateTransactionTime(duration)

	return nil
}

// executeStatement executes a single transaction statement
func (e *TransactionExecutor) executeStatement(ctx context.Context, statement string) error {
	_, err := e.tx.ExecContext(ctx, statement)
	return err
}

// updateTransactionTime updates transaction timing statistics
func (e *TransactionExecutor) updateTransactionTime(duration time.Duration) {
	e.stats.mu.Lock()
	defer e.stats.mu.Unlock()

	e.stats.TotalDuration += duration
	if duration > e.stats.MaxDuration {
		e.stats.MaxDuration = duration
	}
	if e.stats.MinDuration == 0 || duration < e.stats.MinDuration {
		e.stats.MinDuration = duration
	}
}

// isDeadlock checks if an error is a deadlock error
func isDeadlock(err error) bool {
	if err == nil {
		return false
	}
	// Add specific database deadlock error checks here
	// This will depend on the specific database being used
	return false
}
