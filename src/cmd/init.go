package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"ipfs-encrypted-storage/src/encryption"
)

// NewInitCmd creates the init command
func NewInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize encrypted storage system",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit()
		},
	}

	return cmd
}

func runInit() error {
	// Create config directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to determine user home directory: %w", err)
	}
	configDir := filepath.Join(homeDir, ".ipfs-encrypted-storage")
	err = os.MkdirAll(configDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Generate key pair
	publicKey, privateKey, err := encryption.GenerateKeyPair()
	if err != nil {
		return fmt.Errorf("failed to generate key pair: %w", err)
	}

	// Save keys
	keyFile := filepath.Join(configDir, "keys.json")
	keys := map[string]interface{}{
		"public_key":  publicKey,
		"private_key": privateKey,
		"created_at":  time.Now().Unix(),
	}

	keyJSON, err := json.MarshalIndent(keys, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal keys: %w", err)
	}

	err = os.WriteFile(keyFile, keyJSON, 0600)
	if err != nil {
		return fmt.Errorf("failed to save keys: %w", err)
	}

	fmt.Printf("System initialized successfully!\n")
	fmt.Printf("Keys saved to: %s\n", keyFile)
	fmt.Println("Keep this file secure - it contains your private key!")

	return nil
}
