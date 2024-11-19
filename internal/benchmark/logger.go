package benchmark

import (
	"go.uber.org/zap"
)

// Logger is an interface for logging
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

// ZapLogger adapts zap.Logger to our Logger interface
type ZapLogger struct {
	logger *zap.Logger
}

// NewZapLogger creates a new ZapLogger
func NewZapLogger(logger *zap.Logger) *ZapLogger {
	return &ZapLogger{logger: logger}
}

// Debug implements Logger
func (l *ZapLogger) Debug(msg string, args ...interface{}) {
	l.logger.Sugar().Debugf(msg, args...)
}

// Info implements Logger
func (l *ZapLogger) Info(msg string, args ...interface{}) {
	l.logger.Sugar().Infof(msg, args...)
}

// Warn implements Logger
func (l *ZapLogger) Warn(msg string, args ...interface{}) {
	l.logger.Sugar().Warnf(msg, args...)
}

// Error implements Logger
func (l *ZapLogger) Error(msg string, args ...interface{}) {
	l.logger.Sugar().Errorf(msg, args...)
}
