package cmd

import (
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"ipfs-encrypted-storage/src/encryption"
	"ipfs-encrypted-storage/src/errors"
	"ipfs-encrypted-storage/src/integrity"
	"ipfs-encrypted-storage/src/ipfs"
	"ipfs-encrypted-storage/src/utils"
)

// NewVerifyCmd creates the verify command
func NewVerifyCmd() *cobra.Command {
	var (
		cid          string
		metadataFile string
		outputJSON   bool
	)

	cmd := &cobra.Command{
		Use:   "verify [cid]",
		Short: "Verify file integrity and authenticity",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cid = args[0]
			return runVerify(cid, metadataFile, outputJSON)
		},
	}

	cmd.Flags().StringVar(&metadataFile, "metadata", "", "Metadata file for encrypted file verification")
	cmd.Flags().BoolVar(&outputJSON, "json", false, "Output report in JSON format")

	return cmd
}

func runVerify(cid, metadataFile string, outputJSON bool) error {
	// Validate CID
	if err := utils.ValidateCID(cid); err != nil {
		return err
	}

	// Validate metadata file if provided
	if metadataFile != "" {
		if err := utils.ValidateFile(metadataFile, nil); err != nil {
			return err
		}
	}

	client, err := ipfs.NewIPFSClient(GetIPFSURL())
	if err != nil {
		return errors.NewEnhancedError(err, errors.ErrCodeNetworkFailure,
			&errors.ErrorContext{
				Operation: "IPFS_client_creation",
				Resource:  GetIPFSURL(),
				CID:       cid,
				Suggestions: []string{
					"Ensure IPFS daemon is running",
					"Check IPFS API endpoint configuration",
					"Verify network connectivity",
				},
			})
	}
	defer client.Close()

	validator := integrity.NewIntegrityValidator(client)

	var report *integrity.IntegrityReport
	var publicKey ed25519.PublicKey

	if metadataFile != "" {
		// Load metadata for encrypted file verification
		metadataJSON, err := os.ReadFile(metadataFile)
		if err != nil {
			return fmt.Errorf("failed to read metadata file: %w", err)
		}

		var fileInfo map[string]interface{}
		if err := json.Unmarshal(metadataJSON, &fileInfo); err != nil {
			return fmt.Errorf("failed to parse metadata: %w", err)
		}

		metaData, ok := fileInfo["encryption_metadata"].(map[string]interface{})
		if !ok {
			return fmt.Errorf("encryption metadata not found in file")
		}

		// Convert metadata back to struct with safe type assertions
		converter := &utils.SafeJSONConverter{}

		salt, err := converter.Bytes(metaData["salt"], "salt")
		if err != nil {
			return fmt.Errorf("failed to parse salt: %w", err)
		}

		publicKeyBytes, err := converter.Bytes(metaData["public_key"], "public_key")
		if err != nil {
			return fmt.Errorf("failed to parse public_key: %w", err)
		}

		metadata := &encryption.EncryptedMetadata{
			Salt:      salt,
			PublicKey: publicKeyBytes,
		}

		if sigVal, ok := metaData["signature"]; ok && sigVal != nil {
			sig, err := converter.Bytes(sigVal, "signature")
			if err == nil {
				metadata.Signature = sig
			}
		}

		if configVal, ok := metaData["config"].(map[string]interface{}); ok {
			timeVal, err := converter.Uint32(configVal["time"], "config.time")
			if err != nil {
				return fmt.Errorf("failed to parse config.time: %w", err)
			}

			memoryVal, err := converter.Uint32(configVal["memory"], "config.memory")
			if err != nil {
				return fmt.Errorf("failed to parse config.memory: %w", err)
			}

			threadsVal, err := converter.Uint8(configVal["threads"], "config.threads")
			if err != nil {
				return fmt.Errorf("failed to parse config.threads: %w", err)
			}

			keyLenVal, err := converter.Uint32(configVal["key_len"], "config.key_len")
			if err != nil {
				return fmt.Errorf("failed to parse config.key_len: %w", err)
			}

			metadata.Config = &encryption.KeyDerivationConfig{
				Time:    timeVal,
				Memory:  memoryVal,
				Threads: threadsVal,
				KeyLen:  keyLenVal,
			}
		}

		publicKey, err = loadPublicKey(metadata.PublicKey)
		if err != nil {
			return fmt.Errorf("failed to load public key: %w", err)
		}

		// Get password if needed for decryption validation
		passwd := GetPassword()
		if passwd == "" {
			passwd = "dummy_password" // This should be improved
		}

		report, err = validator.ValidateEncryptedFileIntegrity(cid, metadata, passwd, publicKey)
		if err != nil {
			return fmt.Errorf("failed to validate encrypted file: %w", err)
		}
	} else {
		// Basic integrity check without encryption metadata
		report = &integrity.IntegrityReport{
			ResourceID:    cid,
			Timestamp:     time.Now(),
			OverallStatus: integrity.StatusPassed,
			Checks:        []integrity.IntegrityCheck{},
		}

		// Check CID format
		if err := ipfs.ValidateCID(cid); err != nil {
			report.AddCheck(integrity.IntegrityCheck{
				Name:        "cid_format",
				Description: "Validate CID format and structure",
				Status:      integrity.StatusFailed,
				Error:       fmt.Sprintf("Invalid CID format: %v", err),
			})
		} else {
			report.AddCheck(integrity.IntegrityCheck{
				Name:        "cid_format",
				Description: "Validate CID format and structure",
				Status:      integrity.StatusPassed,
				Details:     "CID format is valid",
			})
		}

		// Check file retrieval
		data, err := client.GetFile(cid)
		if err != nil {
			report.AddCheck(integrity.IntegrityCheck{
				Name:        "file_retrieval",
				Description: "Verify file can be retrieved from IPFS",
				Status:      integrity.StatusFailed,
				Error:       fmt.Sprintf("Failed to retrieve file: %v", err),
			})
		} else {
			report.AddCheck(integrity.IntegrityCheck{
				Name:        "file_retrieval",
				Description: "Verify file can be retrieved from IPFS",
				Status:      integrity.StatusPassed,
				Details:     fmt.Sprintf("Successfully retrieved %d bytes", len(data)),
			})
		}

		report.CalculateSummary()
	}

	// Output report
	if outputJSON {
		jsonData, err := report.ToJSON()
		if err != nil {
			return fmt.Errorf("failed to marshal report: %w", err)
		}
		fmt.Println(string(jsonData))
	} else {
		// Human-readable output
		fmt.Printf("Integrity Verification Report\n")
		fmt.Printf("=============================\n\n")
		fmt.Printf("Resource ID: %s\n", report.ResourceID)
		fmt.Printf("Timestamp: %s\n", report.Timestamp.Format(time.RFC3339))
		fmt.Printf("Overall Status: %s\n\n", report.OverallStatus)

		fmt.Printf("Checks:\n")
		for i, check := range report.Checks {
			statusIcon := "✓"
			if check.Status == integrity.StatusFailed {
				statusIcon = "✗"
			} else if check.Status == integrity.StatusWarning {
				statusIcon = "⚠"
			}

			fmt.Printf("  %d. %s [%s] %s\n", i+1, check.Name, statusIcon, check.Status)
			if check.Details != "" {
				fmt.Printf("     Details: %s\n", check.Details)
			}
			if check.Error != "" {
				fmt.Printf("     Error: %s\n", check.Error)
			}
		}

		fmt.Printf("\nSummary:\n")
		fmt.Printf("  Total Checks: %d\n", report.Summary.TotalChecks)
		fmt.Printf("  Passed: %d\n", report.Summary.PassedChecks)
		fmt.Printf("  Failed: %d\n", report.Summary.FailedChecks)
		fmt.Printf("  Warnings: %d\n", report.Summary.WarningChecks)
	}

	if report.OverallStatus == integrity.StatusFailed {
		return fmt.Errorf("integrity verification failed")
	}

	return nil
}
