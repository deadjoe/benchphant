package benchmark

import (
	"sync"
	"time"
)

// TransactionStats holds statistics about transactions
type TransactionStats struct {
	mu                     sync.Mutex
	TotalTransactions      int64         `json:"total_transactions"`
	SuccessfulTransactions int64         `json:"successful_transactions"`
	FailedTransactions     int64         `json:"failed_transactions"`
	TotalDuration          time.Duration `json:"total_duration"`
	MinDuration            time.Duration `json:"min_duration"`
	MaxDuration            time.Duration `json:"max_duration"`
	P95Duration            time.Duration `json:"p95_duration"`
	LockStats              LockStats     `json:"lock_stats"`
}

// LockStats holds statistics about database locks and deadlocks
type LockStats struct {
	DeadlockCount int64         `json:"deadlock_count"`
	RetryCount    int64         `json:"retry_count"`
	TotalLockTime time.Duration `json:"total_lock_time"`
	AvgLockTime   time.Duration `json:"avg_lock_time"`
	MaxLockTime   time.Duration `json:"max_lock_time"`
	LockCount     int64         `json:"lock_count"`
}

// QueryStats holds statistics about SQL queries
type QueryStats struct {
	Count         int64         `json:"count"`
	TotalDuration time.Duration `json:"total_duration"`
	MinDuration   time.Duration `json:"min_duration"`
	MaxDuration   time.Duration `json:"max_duration"`
	P95Duration   time.Duration `json:"p95_duration"`
	Errors        int64         `json:"errors"`
}

// AverageDuration returns the average duration of successful transactions
func (s *TransactionStats) AverageDuration() time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.SuccessfulTransactions == 0 {
		return 0
	}
	return time.Duration(int64(s.TotalDuration) / s.SuccessfulTransactions)
}

// Total returns the total number of transactions
func (s *TransactionStats) Total() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.TotalTransactions
}

// SuccessRate returns the percentage of successful transactions
func (s *TransactionStats) SuccessRate() float64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.TotalTransactions == 0 {
		return 0
	}
	return float64(s.SuccessfulTransactions) / float64(s.TotalTransactions) * 100
}

// TPS returns the transactions per second
func (s *TransactionStats) TPS(duration time.Duration) float64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	if duration <= 0 {
		return 0
	}
	return float64(s.SuccessfulTransactions) / duration.Seconds()
}
