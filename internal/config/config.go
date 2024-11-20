package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds the application configuration
type Config struct {
	Port              int    `json:"port"`
	ServerOnly        bool   `json:"server_only"`
	LogLevel          string `json:"log_level"`
	StoragePath       string `json:"storage_path"`
	EncryptionKeyFile string `json:"encryption_key_file"`
	WebDir            string `json:"web_dir"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Port:              8080,
		ServerOnly:        false,
		LogLevel:          "info",
		StoragePath:       "./data/benchphant.db",
		EncryptionKeyFile: "./data/key.txt",
		WebDir:            "./web/dist",
	}
}

// LoadConfig loads the configuration from the specified file
func LoadConfig(path string) (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// Load loads the configuration from default locations and environment variables
func Load() (*Config, error) {
	cfg := DefaultConfig()

	// Get user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	// Create .benchphant directory if it doesn't exist
	configDir := filepath.Join(homeDir, ".benchphant")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	// Try to load from config file
	configFile := filepath.Join(configDir, "config.json")
	if _, err := os.Stat(configFile); err == nil {
		if cfg, err = LoadConfig(configFile); err != nil {
			return nil, fmt.Errorf("failed to load config file: %w", err)
		}
	}

	// Override with environment variables if set
	if port := os.Getenv("BENCHPHANT_PORT"); port != "" {
		var p int
		if _, err := fmt.Sscanf(port, "%d", &p); err == nil {
			cfg.Port = p
		}
	}
	if logLevel := os.Getenv("BENCHPHANT_LOG_LEVEL"); logLevel != "" {
		cfg.LogLevel = logLevel
	}
	if storagePath := os.Getenv("BENCHPHANT_STORAGE_PATH"); storagePath != "" {
		cfg.StoragePath = storagePath
	}
	if keyFile := os.Getenv("BENCHPHANT_KEY_FILE"); keyFile != "" {
		cfg.EncryptionKeyFile = keyFile
	}
	if webDir := os.Getenv("BENCHPHANT_WEB_DIR"); webDir != "" {
		cfg.WebDir = webDir
	}

	return cfg, nil
}

// Save saves the configuration to file
func (c *Config) Save() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".benchphant")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configFile := filepath.Join(configDir, "config.json")
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("invalid port: %d", c.Port)
	}

	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
		"fatal": true,
	}
	if !validLevels[c.LogLevel] {
		return fmt.Errorf("invalid log level: %s", c.LogLevel)
	}

	if c.StoragePath == "" {
		return fmt.Errorf("storage path is required")
	}
	if c.EncryptionKeyFile == "" {
		return fmt.Errorf("encryption key file is required")
	}
	if c.WebDir == "" {
		return fmt.Errorf("web directory is required")
	}
	return nil
}
