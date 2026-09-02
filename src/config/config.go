// Package config provides configuration management for IPFS Encrypted Storage
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Config represents the application configuration
type Config struct {
	IPFS       IPFSConfig       `json:"ipfs"`
	Encryption EncryptionConfig `json:"encryption"`
	P2P        P2PConfig        `json:"p2p"`
	Logging    LoggingConfig    `json:"logging"`
}

// Duration is a custom type for JSON serialization of time.Duration
type Duration time.Duration

// MarshalJSON implements json.Marshaler
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

// UnmarshalJSON implements json.Unmarshaler
func (d *Duration) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	dur, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d = Duration(dur)
	return nil
}

// IPFSConfig holds IPFS-related configuration
type IPFSConfig struct {
	URL     string   `json:"url"`
	Timeout Duration `json:"timeout"` // Serialized as duration string in JSON
}

// EncryptionConfig holds encryption-related configuration
type EncryptionConfig struct {
	KeyDerivation KeyDerivationConfig `json:"key_derivation"`
}

// KeyDerivationConfig holds Argon2 key derivation parameters
type KeyDerivationConfig struct {
	Time    uint32 `json:"time"`
	Memory  uint32 `json:"memory"`
	Threads uint8  `json:"threads"`
	KeyLen  uint32 `json:"key_len"`
}

// P2PConfig holds configuration for the local P2P stub.
type P2PConfig struct {
	ListenAddr     string   `json:"listen_addr"`
	BootstrapPeers []string `json:"bootstrap_peers"`
	Bootstrap      bool     `json:"bootstrap"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level      string `json:"level"`       // debug, info, warn, error
	Format     string `json:"format"`      // text, json
	OutputFile string `json:"output_file"` // empty for stdout
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		IPFS: IPFSConfig{
			URL:     "localhost:5001",
			Timeout: Duration(30 * time.Second),
		},
		Encryption: EncryptionConfig{
			KeyDerivation: KeyDerivationConfig{
				Time:    1,
				Memory:  64 * 1024, // 64 MiB
				Threads: 4,
				KeyLen:  32, // 256 bits
			},
		},
		P2P: P2PConfig{
			ListenAddr:     "/ip4/0.0.0.0/tcp/0",
			BootstrapPeers: []string{},
			Bootstrap:      true,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "text",
		},
	}
}

// GetConfigDir returns the configuration directory path
func GetConfigDir() (string, error) {
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		// Fallback for Windows
		homeDir = os.Getenv("USERPROFILE")
		if homeDir == "" {
			return "", fmt.Errorf("unable to determine home directory")
		}
	}
	return filepath.Join(homeDir, ".ipfs-encrypted-storage"), nil
}

// GetConfigPath returns the full path to the config file
func GetConfigPath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "config.json"), nil
}

// LoadConfig loads configuration from file, or returns default if file doesn't exist
func LoadConfig(configPath string) (*Config, error) {
	// If no path provided, use default
	if configPath == "" {
		var err error
		configPath, err = GetConfigPath()
		if err != nil {
			return nil, fmt.Errorf("failed to get config path: %w", err)
		}
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Return default config if file doesn't exist
		return DefaultConfig(), nil
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse JSON
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate and set defaults for missing fields
	config = *validateAndMergeDefaults(&config)

	return &config, nil
}

// SaveConfig saves configuration to file
func SaveConfig(config *Config, configPath string) error {
	// If no path provided, use default
	if configPath == "" {
		var err error
		configPath, err = GetConfigPath()
		if err != nil {
			return fmt.Errorf("failed to get config path: %w", err)
		}
	}

	// Ensure config directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// validateAndMergeDefaults validates config and merges with defaults
func validateAndMergeDefaults(config *Config) *Config {
	defaults := DefaultConfig()

	// Merge IPFS config
	if config.IPFS.URL == "" {
		config.IPFS.URL = defaults.IPFS.URL
	}
	if config.IPFS.Timeout == 0 {
		config.IPFS.Timeout = defaults.IPFS.Timeout
	}

	// Convert Duration to time.Duration for use in code
	_ = time.Duration(config.IPFS.Timeout)

	// Merge encryption config
	if config.Encryption.KeyDerivation.Time == 0 {
		config.Encryption.KeyDerivation.Time = defaults.Encryption.KeyDerivation.Time
	}
	if config.Encryption.KeyDerivation.Memory == 0 {
		config.Encryption.KeyDerivation.Memory = defaults.Encryption.KeyDerivation.Memory
	}
	if config.Encryption.KeyDerivation.Threads == 0 {
		config.Encryption.KeyDerivation.Threads = defaults.Encryption.KeyDerivation.Threads
	}
	if config.Encryption.KeyDerivation.KeyLen == 0 {
		config.Encryption.KeyDerivation.KeyLen = defaults.Encryption.KeyDerivation.KeyLen
	}

	// Merge P2P config
	if config.P2P.ListenAddr == "" {
		config.P2P.ListenAddr = defaults.P2P.ListenAddr
	}
	if config.P2P.BootstrapPeers == nil {
		config.P2P.BootstrapPeers = defaults.P2P.BootstrapPeers
	}

	// Merge logging config
	if config.Logging.Level == "" {
		config.Logging.Level = defaults.Logging.Level
	}
	if config.Logging.Format == "" {
		config.Logging.Format = defaults.Logging.Format
	}

	return config
}

// ValidateConfig validates the configuration
func ValidateConfig(config *Config) error {
	if config.IPFS.URL == "" {
		return fmt.Errorf("IPFS URL cannot be empty")
	}

	if config.Encryption.KeyDerivation.Time < 1 || config.Encryption.KeyDerivation.Time > 10 {
		return fmt.Errorf("key derivation time must be between 1 and 10")
	}

	if config.Encryption.KeyDerivation.Memory < 1024 {
		return fmt.Errorf("key derivation memory must be at least 1024 KiB")
	}

	if config.Encryption.KeyDerivation.Threads < 1 || config.Encryption.KeyDerivation.Threads > 8 {
		return fmt.Errorf("key derivation threads must be between 1 and 8")
	}

	if config.Encryption.KeyDerivation.KeyLen != 32 {
		return fmt.Errorf("key length must be 32 bytes for AES-256")
	}

	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLogLevels[config.Logging.Level] {
		return fmt.Errorf("invalid log level: %s", config.Logging.Level)
	}

	validLogFormats := map[string]bool{
		"text": true,
		"json": true,
	}
	if !validLogFormats[config.Logging.Format] {
		return fmt.Errorf("invalid log format: %s", config.Logging.Format)
	}

	return nil
}
