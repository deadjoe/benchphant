package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestLogger(t *testing.T) {
	t.Run("DefaultConfig", func(t *testing.T) {
		cfg := DefaultConfig()
		assert.NotNil(t, cfg)
		assert.Equal(t, "info", cfg.Level)
		assert.Equal(t, "logs", cfg.Path)
		assert.Equal(t, 100, cfg.MaxSize)
		assert.Equal(t, 3, cfg.MaxBackups)
		assert.Equal(t, 7, cfg.MaxAge)
		assert.True(t, cfg.Compress)
	})

	t.Run("ValidateConfig", func(t *testing.T) {
		tests := []struct {
			name    string
			cfg     *Config
			wantErr bool
		}{
			{
				name: "ValidConfig",
				cfg: &Config{
					Level:      "info",
					Path:       "logs",
					MaxSize:    100,
					MaxBackups: 3,
					MaxAge:     7,
					Compress:   true,
				},
				wantErr: false,
			},
			{
				name: "EmptyLevel",
				cfg: &Config{
					Level:      "",
					Path:       "logs",
					MaxSize:    100,
					MaxBackups: 3,
					MaxAge:     7,
					Compress:   true,
				},
				wantErr: true,
			},
			{
				name: "InvalidLevel",
				cfg: &Config{
					Level:      "invalid",
					Path:       "logs",
					MaxSize:    100,
					MaxBackups: 3,
					MaxAge:     7,
					Compress:   true,
				},
				wantErr: true,
			},
			{
				name: "EmptyPath",
				cfg: &Config{
					Level:      "info",
					Path:       "",
					MaxSize:    100,
					MaxBackups: 3,
					MaxAge:     7,
					Compress:   true,
				},
				wantErr: true,
			},
			{
				name: "InvalidMaxSize",
				cfg: &Config{
					Level:      "info",
					Path:       "logs",
					MaxSize:    0,
					MaxBackups: 3,
					MaxAge:     7,
					Compress:   true,
				},
				wantErr: true,
			},
			{
				name: "InvalidMaxBackups",
				cfg: &Config{
					Level:      "info",
					Path:       "logs",
					MaxSize:    100,
					MaxBackups: -1,
					MaxAge:     7,
					Compress:   true,
				},
				wantErr: true,
			},
			{
				name: "InvalidMaxAge",
				cfg: &Config{
					Level:      "info",
					Path:       "logs",
					MaxSize:    100,
					MaxBackups: 3,
					MaxAge:     -1,
					Compress:   true,
				},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := tt.cfg.Validate()
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})

	t.Run("Init", func(t *testing.T) {
		// Test with nil config
		err := Init(nil)
		assert.NoError(t, err)
		assert.NotNil(t, Log)
		Close()

		// Test with valid config
		tempDir := t.TempDir()
		cfg := &Config{
			Level:      "debug",
			Path:       tempDir,
			MaxSize:    100,
			MaxBackups: 3,
			MaxAge:     7,
			Compress:   true,
		}
		err = Init(cfg)
		assert.NoError(t, err)
		assert.NotNil(t, Log)
		Close()

		// Test with invalid config
		cfg.Level = "invalid"
		err = Init(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid logger config")

		// Test with invalid directory permissions
		if os.Getuid() == 0 {
			t.Skip("Skipping directory permission test when running as root")
		}
		cfg = DefaultConfig()
		cfg.Path = "/root/test-logs" // Should fail due to permissions
		err = Init(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create log directory")
	})

	t.Run("Close", func(t *testing.T) {
		// Test closing nil logger
		Log = nil
		assert.NoError(t, Close())

		// Test closing initialized logger
		cfg := DefaultConfig()
		err := Init(cfg)
		assert.NoError(t, err)
		assert.NoError(t, Close())

		// Test closing logger with custom sync error
		Log = zap.New(zapcore.NewCore(
			zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
			zapcore.AddSync(&mockSyncer{err: fmt.Errorf("custom sync error")}),
			zap.InfoLevel,
		))
		err = Close()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "custom sync error")
	})

	t.Run("LogLevels", func(t *testing.T) {
		tempDir := t.TempDir()
		cfg := &Config{
			Level:      "debug",
			Path:       tempDir,
			MaxSize:    100,
			MaxBackups: 3,
			MaxAge:     7,
			Compress:   true,
		}
		err := Init(cfg)
		assert.NoError(t, err)
		defer Close()

		// Test all log levels
		Log.Debug("Debug message")
		Log.Info("Info message")
		Log.Warn("Warning message")
		Log.Error("Error message")
	})

	t.Run("GetLogFilePath", func(t *testing.T) {
		cfg := DefaultConfig()
		expected := filepath.Join(cfg.Path, "benchphant.log")
		assert.Equal(t, expected, GetLogFilePath(cfg))
	})
}

func TestLogLevels(t *testing.T) {
	levels := []string{
		"debug",
		"info",
		"warn",
		"error",
		"dpanic",
		"panic",
		"fatal",
	}

	for _, level := range levels {
		t.Run(level, func(t *testing.T) {
			_, err := zapcore.ParseLevel(level)
			assert.NoError(t, err)
		})
	}
}

type mockSyncer struct {
	err error
}

func (s *mockSyncer) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (s *mockSyncer) Sync() error {
	return s.err
}
