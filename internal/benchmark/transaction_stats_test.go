package benchmark

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTransactionStats_AverageDuration(t *testing.T) {
	stats := &TransactionStats{}

	// Test with no transactions
	assert.Equal(t, time.Duration(0), stats.AverageDuration())

	// Test with one transaction
	stats.TotalDuration = time.Second
	stats.SuccessfulTransactions = 1
	assert.Equal(t, time.Second, stats.AverageDuration())

	// Test with multiple transactions
	stats.TotalDuration = 5 * time.Second
	stats.SuccessfulTransactions = 2
	assert.Equal(t, 2500*time.Millisecond, stats.AverageDuration())
}

func TestTransactionStats_Total(t *testing.T) {
	stats := &TransactionStats{}

	// Test initial state
	assert.Equal(t, int64(0), stats.Total())

	// Test after adding transactions
	stats.TotalTransactions = 5
	assert.Equal(t, int64(5), stats.Total())
}

func TestTransactionStats_SuccessRate(t *testing.T) {
	stats := &TransactionStats{}

	// Test with no transactions
	assert.Equal(t, float64(0), stats.SuccessRate())

	// Test with all successful transactions
	stats.TotalTransactions = 10
	stats.SuccessfulTransactions = 10
	assert.Equal(t, float64(100), stats.SuccessRate())

	// Test with mixed success/failure
	stats.TotalTransactions = 10
	stats.SuccessfulTransactions = 7
	assert.Equal(t, float64(70), stats.SuccessRate())

	// Test with all failed transactions
	stats.TotalTransactions = 10
	stats.SuccessfulTransactions = 0
	assert.Equal(t, float64(0), stats.SuccessRate())
}

func TestTransactionStats_TPS(t *testing.T) {
	stats := &TransactionStats{}

	// Test with zero duration
	assert.Equal(t, float64(0), stats.TPS(0))

	// Test with no transactions
	assert.Equal(t, float64(0), stats.TPS(time.Second))

	// Test with one transaction per second
	stats.SuccessfulTransactions = 1
	assert.Equal(t, float64(1), stats.TPS(time.Second))

	// Test with multiple transactions
	stats.SuccessfulTransactions = 100
	assert.Equal(t, float64(100), stats.TPS(time.Second))

	// Test with longer duration
	stats.SuccessfulTransactions = 100
	assert.Equal(t, float64(10), stats.TPS(10*time.Second))
}

func TestTransactionStats_Concurrency(t *testing.T) {
	stats := &TransactionStats{}
	goroutines := 10
	iterations := 1000
	done := make(chan bool)

	// Concurrently update stats
	for i := 0; i < goroutines; i++ {
		go func() {
			for j := 0; j < iterations; j++ {
				stats.mu.Lock()
				stats.TotalTransactions++
				stats.SuccessfulTransactions++
				stats.TotalDuration += time.Millisecond
				stats.mu.Unlock()
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < goroutines; i++ {
		<-done
	}

	// Verify results
	expectedTotal := int64(goroutines * iterations)
	assert.Equal(t, expectedTotal, stats.TotalTransactions)
	assert.Equal(t, expectedTotal, stats.SuccessfulTransactions)
	assert.Equal(t, time.Duration(expectedTotal)*time.Millisecond, stats.TotalDuration)
}

func TestTransactionStats_LockStats(t *testing.T) {
	stats := &TransactionStats{}

	// Test initial state
	assert.Equal(t, int64(0), stats.LockStats.DeadlockCount)
	assert.Equal(t, int64(0), stats.LockStats.RetryCount)
	assert.Equal(t, time.Duration(0), stats.LockStats.TotalLockTime)
	assert.Equal(t, time.Duration(0), stats.LockStats.AvgLockTime)
	assert.Equal(t, time.Duration(0), stats.LockStats.MaxLockTime)
	assert.Equal(t, int64(0), stats.LockStats.LockCount)

	// Update lock stats
	stats.LockStats.DeadlockCount = 5
	stats.LockStats.RetryCount = 10
	stats.LockStats.TotalLockTime = 5 * time.Second
	stats.LockStats.AvgLockTime = 500 * time.Millisecond
	stats.LockStats.MaxLockTime = time.Second
	stats.LockStats.LockCount = 15

	// Verify updates
	assert.Equal(t, int64(5), stats.LockStats.DeadlockCount)
	assert.Equal(t, int64(10), stats.LockStats.RetryCount)
	assert.Equal(t, 5*time.Second, stats.LockStats.TotalLockTime)
	assert.Equal(t, 500*time.Millisecond, stats.LockStats.AvgLockTime)
	assert.Equal(t, time.Second, stats.LockStats.MaxLockTime)
	assert.Equal(t, int64(15), stats.LockStats.LockCount)
}
