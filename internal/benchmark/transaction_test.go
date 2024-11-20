package benchmark

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

// MockLogger implements the Logger interface for testing
type MockLogger struct{}

func (m *MockLogger) Info(msg string, keysAndValues ...interface{})  {}
func (m *MockLogger) Error(msg string, keysAndValues ...interface{}) {}
func (m *MockLogger) Warn(msg string, keysAndValues ...interface{})  {}
func (m *MockLogger) Debug(msg string, keysAndValues ...interface{}) {}

func TestNewTransactionExecutor(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	stats := &TransactionStats{}
	logger := &MockLogger{}

	executor := NewTransactionExecutor(db, stats, logger)
	assert.NotNil(t, executor)
	assert.Equal(t, db, executor.db)
	assert.Equal(t, stats, executor.stats)
	assert.Equal(t, logger, executor.logger)
}

func TestExecute_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	stats := &TransactionStats{}
	logger := &MockLogger{}
	executor := NewTransactionExecutor(db, stats, logger)

	transaction := &Transaction{
		Type: "test",
		Statements: []string{
			"INSERT INTO test VALUES (?)",
			"UPDATE test SET value = ?",
		},
	}

	// Expect transaction begin
	mock.ExpectBegin()

	// Expect statements execution
	mock.ExpectExec("INSERT INTO test VALUES").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE test SET value").WillReturnResult(sqlmock.NewResult(1, 1))

	// Expect transaction commit
	mock.ExpectCommit()

	err = executor.Execute(context.Background(), transaction)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), stats.SuccessfulTransactions)
	assert.Equal(t, int64(1), stats.TotalTransactions)
	assert.Equal(t, int64(0), stats.FailedTransactions)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestExecute_BeginError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	stats := &TransactionStats{}
	logger := &MockLogger{}
	executor := NewTransactionExecutor(db, stats, logger)

	transaction := &Transaction{
		Type:       "test",
		Statements: []string{"INSERT INTO test VALUES (?)"},
	}

	expectedError := errors.New("begin error")
	mock.ExpectBegin().WillReturnError(expectedError)

	err = executor.Execute(context.Background(), transaction)
	assert.Error(t, err)
	assert.Equal(t, int64(0), stats.SuccessfulTransactions)
	assert.Equal(t, int64(0), stats.TotalTransactions)
	assert.Equal(t, int64(1), stats.FailedTransactions)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestExecute_StatementError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	stats := &TransactionStats{}
	logger := &MockLogger{}
	executor := NewTransactionExecutor(db, stats, logger)

	transaction := &Transaction{
		Type:       "test",
		Statements: []string{"INSERT INTO test VALUES (?)"},
	}

	mock.ExpectBegin()
	expectedError := errors.New("statement error")
	mock.ExpectExec("INSERT INTO test VALUES").WillReturnError(expectedError)
	mock.ExpectRollback()

	err = executor.Execute(context.Background(), transaction)
	assert.Error(t, err)
	assert.Equal(t, int64(0), stats.SuccessfulTransactions)
	assert.Equal(t, int64(0), stats.TotalTransactions)
	assert.Equal(t, int64(1), stats.FailedTransactions)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestExecute_CommitError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	stats := &TransactionStats{}
	logger := &MockLogger{}
	executor := NewTransactionExecutor(db, stats, logger)

	transaction := &Transaction{
		Type:       "test",
		Statements: []string{"INSERT INTO test VALUES (?)"},
	}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO test VALUES").WillReturnResult(sqlmock.NewResult(1, 1))
	expectedError := errors.New("commit error")
	mock.ExpectCommit().WillReturnError(expectedError)

	err = executor.Execute(context.Background(), transaction)
	assert.Error(t, err)
	assert.Equal(t, int64(0), stats.SuccessfulTransactions)
	assert.Equal(t, int64(0), stats.TotalTransactions)
	assert.Equal(t, int64(1), stats.FailedTransactions)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestExecute_Deadlock(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	stats := &TransactionStats{}
	logger := &MockLogger{}
	executor := NewTransactionExecutor(db, stats, logger)

	transaction := &Transaction{
		Type:       "test",
		Statements: []string{"INSERT INTO test VALUES (?)"},
	}

	mock.ExpectBegin()
	deadlockError := errors.New("deadlock error")
	mock.ExpectExec("INSERT INTO test VALUES").WillReturnError(deadlockError)
	mock.ExpectRollback()

	err = executor.Execute(context.Background(), transaction)
	assert.Error(t, err)
	assert.Equal(t, int64(0), stats.SuccessfulTransactions)
	assert.Equal(t, int64(0), stats.TotalTransactions)
	assert.Equal(t, int64(1), stats.FailedTransactions)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestUpdateTransactionTime(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	stats := &TransactionStats{}
	logger := &MockLogger{}
	executor := NewTransactionExecutor(db, stats, logger)

	// Test first update
	duration1 := time.Second
	executor.updateTransactionTime(duration1)
	assert.Equal(t, duration1, stats.TotalDuration)
	assert.Equal(t, duration1, stats.MinDuration)
	assert.Equal(t, duration1, stats.MaxDuration)

	// Test update with shorter duration
	duration2 := 500 * time.Millisecond
	executor.updateTransactionTime(duration2)
	assert.Equal(t, duration1+duration2, stats.TotalDuration)
	assert.Equal(t, duration2, stats.MinDuration)
	assert.Equal(t, duration1, stats.MaxDuration)

	// Test update with longer duration
	duration3 := 2 * time.Second
	executor.updateTransactionTime(duration3)
	assert.Equal(t, duration1+duration2+duration3, stats.TotalDuration)
	assert.Equal(t, duration2, stats.MinDuration)
	assert.Equal(t, duration3, stats.MaxDuration)
}

func TestIsDeadlock(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	stats := &TransactionStats{}
	logger := &MockLogger{}
	executor := NewTransactionExecutor(db, stats, logger)

	assert.False(t, executor.IsDeadlock(nil))
	assert.False(t, executor.IsDeadlock(errors.New("random error")))
	assert.True(t, executor.IsDeadlock(errors.New("deadlock detected")))
}
