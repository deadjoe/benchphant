package logger

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// Log is the global logger instance
	Log *zap.Logger
)

// Config represents the logger configuration
type Config struct {
	// Level is the minimum enabled logging level
	Level string `json:"level"`
	// Path is the directory where log files will be stored
	Path string `json:"path"`
	// MaxSize is the maximum size in megabytes of the log file before it gets rotated
	MaxSize int `json:"max_size"`
	// MaxBackups is the maximum number of old log files to retain
	MaxBackups int `json:"max_backups"`
	// MaxAge is the maximum number of days to retain old log files
	MaxAge int `json:"max_age"`
	// Compress determines if the rotated log files should be compressed
	Compress bool `json:"compress"`
}

// DefaultConfig returns the default logger configuration
func DefaultConfig() *Config {
	return &Config{
		Level:      "info",
		Path:       "logs",
		MaxSize:    100,
		MaxBackups: 3,
		MaxAge:     7,
		Compress:   true,
	}
}

// Validate validates the logger configuration
func (c *Config) Validate() error {
	if c.Level == "" {
		return fmt.Errorf("log level is required")
	}
	if _, err := zapcore.ParseLevel(c.Level); err != nil {
		return fmt.Errorf("invalid log level: %s", c.Level)
	}
	if c.Path == "" {
		return fmt.Errorf("log path is required")
	}
	if c.MaxSize <= 0 {
		return fmt.Errorf("max size must be greater than 0")
	}
	if c.MaxBackups < 0 {
		return fmt.Errorf("max backups cannot be negative")
	}
	if c.MaxAge < 0 {
		return fmt.Errorf("max age cannot be negative")
	}
	return nil
}

// Init initializes the global logger with the given configuration
func Init(cfg *Config) error {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid logger config: %w", err)
	}

	// Create log directory if it doesn't exist
	if err := os.MkdirAll(cfg.Path, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Parse log level
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		return fmt.Errorf("failed to parse log level: %w", err)
	}

	// Create encoder config
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Create core
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(os.Stdout),
		level,
	)

	// Create logger
	Log = zap.New(core,
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)

	Log.Info("Logger initialized",
		zap.String("level", cfg.Level),
		zap.String("path", cfg.Path),
		zap.Int("max_size", cfg.MaxSize),
		zap.Int("max_backups", cfg.MaxBackups),
		zap.Int("max_age", cfg.MaxAge),
		zap.Bool("compress", cfg.Compress),
	)

	return nil
}

// Close flushes any buffered log entries
func Close() error {
	if Log != nil {
		err := Log.Sync()
		// Ignore sync errors for stdout and stderr
		if err != nil && err.Error() != "sync /dev/stdout: bad file descriptor" && err.Error() != "sync /dev/stderr: bad file descriptor" {
			return err
		}
	}
	return nil
}

// GetLogFilePath returns the full path to the log file
func GetLogFilePath(cfg *Config) string {
	return filepath.Join(cfg.Path, "benchphant.log")
}
