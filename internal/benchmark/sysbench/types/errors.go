package types

import "errors"

// Common errors
var (
	ErrInvalidTableSize      = errors.New("invalid table size")
	ErrInvalidNumTables      = errors.New("invalid number of tables")
	ErrInvalidNumThreads     = errors.New("invalid number of threads")
	ErrInvalidDuration       = errors.New("invalid duration")
	ErrInvalidReportInterval = errors.New("invalid report interval")
	ErrInvalidWeight         = errors.New("invalid weight (must be between 0 and 1)")
	ErrInvalidWeightSum      = errors.New("sum of weights must equal 1.0")
	ErrInvalidConfig         = errors.New("invalid configuration")
	ErrTestNotFound          = errors.New("test not found")
	ErrScenarioNotFound      = errors.New("scenario not found")
	ErrDatabaseNotConnected  = errors.New("database not connected")
	ErrTestAlreadyRunning    = errors.New("test is already running")
	ErrTestNotRunning        = errors.New("test is not running")
	ErrTestFailed            = errors.New("test failed")
)
