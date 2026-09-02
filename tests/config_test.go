package tests

import (
	"os"
	"path/filepath"
	"testing"

	"ipfs-encrypted-storage/src/config"
)

func TestDefaultConfig(t *testing.T) {
	cfg := config.DefaultConfig()
	if cfg == nil {
		t.Fatal("DefaultConfig returned nil")
	}

	if cfg.IPFS.URL == "" {
		t.Error("Default IPFS URL should not be empty")
	}

	if cfg.Encryption.KeyDerivation.KeyLen != 32 {
		t.Error("Default key length should be 32 bytes")
	}

	if cfg.Encryption.KeyDerivation.Time == 0 {
		t.Error("Default key derivation time should be set")
	}

	if cfg.Encryption.KeyDerivation.Memory == 0 {
		t.Error("Default key derivation memory should be set")
	}

	if cfg.P2P.ListenAddr == "" {
		t.Error("Default P2P listen address should not be empty")
	}

	if cfg.Logging.Level == "" {
		t.Error("Default log level should not be empty")
	}
}

func TestConfigValidation(t *testing.T) {
	cfg := config.DefaultConfig()
	err := config.ValidateConfig(cfg)
	if err != nil {
		t.Errorf("Default config should be valid: %v", err)
	}

	// Test invalid key length
	invalidCfg := config.DefaultConfig()
	invalidCfg.Encryption.KeyDerivation.KeyLen = 16
	err = config.ValidateConfig(invalidCfg)
	if err == nil {
		t.Error("Config with invalid key length should fail validation")
	}

	// Test invalid log level
	invalidCfg = config.DefaultConfig()
	invalidCfg.Logging.Level = "invalid"
	err = config.ValidateConfig(invalidCfg)
	if err == nil {
		t.Error("Config with invalid log level should fail validation")
	}

	// Test invalid log format
	invalidCfg = config.DefaultConfig()
	invalidCfg.Logging.Format = "invalid"
	err = config.ValidateConfig(invalidCfg)
	if err == nil {
		t.Error("Config with invalid log format should fail validation")
	}
}

func TestConfigSaveLoad(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.json")

	// Create and save config
	cfg := config.DefaultConfig()
	cfg.IPFS.URL = "test-url:5001"
	cfg.Logging.Level = "debug"

	err := config.SaveConfig(cfg, configPath)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("Config file was not created")
	}

	// Load config
	loadedCfg, err := config.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if loadedCfg.IPFS.URL != cfg.IPFS.URL {
		t.Errorf("Loaded config URL mismatch: expected %s, got %s", cfg.IPFS.URL, loadedCfg.IPFS.URL)
	}

	if loadedCfg.Logging.Level != cfg.Logging.Level {
		t.Errorf("Loaded config log level mismatch: expected %s, got %s", cfg.Logging.Level, loadedCfg.Logging.Level)
	}
}

func TestConfigLoadNonExistent(t *testing.T) {
	// Load non-existent config should return default
	nonExistentPath := "/tmp/non-existent-config-12345.json"
	cfg, err := config.LoadConfig(nonExistentPath)
	if err != nil {
		t.Fatalf("LoadConfig should not fail for non-existent file: %v", err)
	}

	// Should return default config
	if cfg == nil {
		t.Fatal("LoadConfig should return default config for non-existent file")
	}

	// Verify it's a default config
	defaultCfg := config.DefaultConfig()
	if cfg.IPFS.URL != defaultCfg.IPFS.URL {
		t.Error("Non-existent config should return default values")
	}
}

func TestConfigGetConfigDir(t *testing.T) {
	configDir, err := config.GetConfigDir()
	if err != nil {
		t.Fatalf("GetConfigDir failed: %v", err)
	}

	if configDir == "" {
		t.Error("Config directory should not be empty")
	}
}

func TestConfigGetConfigPath(t *testing.T) {
	configPath, err := config.GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath failed: %v", err)
	}

	if configPath == "" {
		t.Error("Config path should not be empty")
	}

	// Should end with config.json
	if filepath.Base(configPath) != "config.json" {
		t.Errorf("Config path should end with config.json, got: %s", configPath)
	}
}
