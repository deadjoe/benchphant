package benchmark

import (
	"context"
	"database/sql"
	"strings"
	"time"
)

// Transaction represents a database transaction with multiple SQL statements
type Transaction struct {
	Type       string
	Statements []string
	StartTime  time.Time
	EndTime    time.Time
}

// TransactionExecutor executes database transactions
type TransactionExecutor struct {
	db     *sql.DB
	stats  *TransactionStats
	logger Logger
}

// NewTransactionExecutor creates a new TransactionExecutor
func NewTransactionExecutor(db *sql.DB, stats *TransactionStats, logger Logger) *TransactionExecutor {
	return &TransactionExecutor{
		db:     db,
		stats:  stats,
		logger: logger,
	}
}

// Execute executes a transaction
func (e *TransactionExecutor) Execute(ctx context.Context, tx *Transaction) error {
	tx.StartTime = time.Now()
	defer func() {
		tx.EndTime = time.Now()
		e.updateTransactionTime(tx.EndTime.Sub(tx.StartTime))
	}()

	// Begin transaction
	txn, err := e.db.BeginTx(ctx, nil)
	if err != nil {
		e.stats.mu.Lock()
		e.stats.FailedTransactions++
		e.stats.mu.Unlock()
		e.logger.Error("Failed to begin transaction", err)
		return err
	}

	// Execute statements
	for _, stmt := range tx.Statements {
		_, err := txn.ExecContext(ctx, stmt)
		if err != nil {
			rollbackErr := txn.Rollback()
			if rollbackErr != nil {
				e.logger.Error("Failed to rollback transaction", rollbackErr)
			}
			e.stats.mu.Lock()
			e.stats.FailedTransactions++
			if e.IsDeadlock(err) {
				e.stats.LockStats.DeadlockCount++
			}
			e.stats.mu.Unlock()
			e.logger.Error("Failed to execute statement", err)
			return err
		}
	}

	// Commit transaction
	err = txn.Commit()
	if err != nil {
		e.stats.mu.Lock()
		e.stats.FailedTransactions++
		e.stats.mu.Unlock()
		e.logger.Error("Failed to commit transaction", err)
		return err
	}

	e.stats.mu.Lock()
	e.stats.SuccessfulTransactions++
	e.stats.TotalTransactions++
	e.stats.mu.Unlock()
	e.logger.Info("Transaction executed successfully")
	return nil
}

// updateTransactionTime updates transaction time statistics
func (e *TransactionExecutor) updateTransactionTime(duration time.Duration) {
	e.stats.mu.Lock()
	defer e.stats.mu.Unlock()

	e.stats.TotalDuration += duration

	if duration < e.stats.MinDuration || e.stats.MinDuration == 0 {
		e.stats.MinDuration = duration
	}
	if duration > e.stats.MaxDuration {
		e.stats.MaxDuration = duration
	}
}

// IsDeadlock checks if an error is a deadlock error
func (e *TransactionExecutor) IsDeadlock(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), "deadlock")
}
