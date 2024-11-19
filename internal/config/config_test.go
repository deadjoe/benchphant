package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	configContent := `{
		"port": 8080,
		"log_level": "debug",
		"storage_path": "/tmp/benchphant.db",
		"encryption_key_file": "/tmp/key.txt",
		"web_dir": "./web/dist"
	}`

	tmpfile, err := os.CreateTemp("", "config_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write([]byte(configContent))
	require.NoError(t, err)
	require.NoError(t, tmpfile.Close())

	// Test loading config
	t.Run("LoadValidConfig", func(t *testing.T) {
		cfg, err := LoadConfig(tmpfile.Name())
		require.NoError(t, err)
		assert.Equal(t, 8080, cfg.Port)
		assert.Equal(t, "debug", cfg.LogLevel)
		assert.Equal(t, "/tmp/benchphant.db", cfg.StoragePath)
		assert.Equal(t, "/tmp/key.txt", cfg.EncryptionKeyFile)
		assert.Equal(t, "./web/dist", cfg.WebDir)
	})

	// Test loading non-existent config
	t.Run("LoadNonExistentConfig", func(t *testing.T) {
		_, err := LoadConfig("/non/existent/path")
		assert.Error(t, err)
	})

	// Test loading invalid config
	t.Run("LoadInvalidConfig", func(t *testing.T) {
		invalidContent := `{invalid json`
		tmpfile, err := os.CreateTemp("", "invalid_config_*.json")
		require.NoError(t, err)
		defer os.Remove(tmpfile.Name())

		_, err = tmpfile.Write([]byte(invalidContent))
		require.NoError(t, err)
		require.NoError(t, tmpfile.Close())

		_, err = LoadConfig(tmpfile.Name())
		assert.Error(t, err)
	})
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	assert.Equal(t, 8080, cfg.Port)
	assert.Equal(t, "info", cfg.LogLevel)
	assert.Equal(t, "./data/benchphant.db", cfg.StoragePath)
	assert.Equal(t, "./data/key.txt", cfg.EncryptionKeyFile)
	assert.Equal(t, "./web/dist", cfg.WebDir)
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name: "ValidConfig",
			cfg: &Config{
				Port:             8080,
				LogLevel:         "info",
				StoragePath:      "/tmp/db.sqlite",
				EncryptionKeyFile: "/tmp/key.txt",
				WebDir:          "./web/dist",
			},
			wantErr: false,
		},
		{
			name: "InvalidPort",
			cfg: &Config{
				Port:             -1,
				LogLevel:         "info",
				StoragePath:      "/tmp/db.sqlite",
				EncryptionKeyFile: "/tmp/key.txt",
				WebDir:          "./web/dist",
			},
			wantErr: true,
		},
		{
			name: "InvalidLogLevel",
			cfg: &Config{
				Port:             8080,
				LogLevel:         "invalid",
				StoragePath:      "/tmp/db.sqlite",
				EncryptionKeyFile: "/tmp/key.txt",
				WebDir:          "./web/dist",
			},
			wantErr: true,
		},
		{
			name: "EmptyStoragePath",
			cfg: &Config{
				Port:             8080,
				LogLevel:         "info",
				StoragePath:      "",
				EncryptionKeyFile: "/tmp/key.txt",
				WebDir:          "./web/dist",
			},
			wantErr: true,
		},
		{
			name: "EmptyEncryptionKeyFile",
			cfg: &Config{
				Port:             8080,
				LogLevel:         "info",
				StoragePath:      "/tmp/db.sqlite",
				EncryptionKeyFile: "",
				WebDir:          "./web/dist",
			},
			wantErr: true,
		},
		{
			name: "EmptyWebDir",
			cfg: &Config{
				Port:             8080,
				LogLevel:         "info",
				StoragePath:      "/tmp/db.sqlite",
				EncryptionKeyFile: "/tmp/key.txt",
				WebDir:          "",
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
}
