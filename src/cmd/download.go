package cmd

import (
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"ipfs-encrypted-storage/src/encryption"
	"ipfs-encrypted-storage/src/errors"
	"ipfs-encrypted-storage/src/ipfs"
	"ipfs-encrypted-storage/src/utils"
)

// NewDownloadCmd creates the download command
func NewDownloadCmd() *cobra.Command {
	var (
		cid          string
		outputPath   string
		metadataFile string
	)

	cmd := &cobra.Command{
		Use:   "download [cid]",
		Short: "Download and decrypt a file from IPFS",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cid = args[0]
			return runDownload(cid, outputPath, metadataFile)
		},
	}

	cmd.Flags().StringVar(&outputPath, "output", "", "Output file path (default: original filename)")
	cmd.Flags().StringVar(&metadataFile, "metadata", "", "Metadata file for decryption")

	return cmd
}

func runDownload(cid, outputPath, metadataFile string) error {
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

	// Get file from IPFS
	data, err := client.GetFile(cid)
	if err != nil {
		return errors.NewEnhancedError(err, errors.ErrCodeResourceNotFound,
			&errors.ErrorContext{
				Operation: "IPFS_download",
				Resource:  cid,
				CID:       cid,
				Suggestions: []string{
					"Verify CID is correct",
					"Check if content is available on the network",
					"Try pinning the content first",
					"Ensure IPFS daemon has access to the content",
				},
			})
	}

	var outputData []byte
	filename := cid // Default filename

	if metadataFile != "" {
		// Load metadata for decryption
		metadataJSON, err := os.ReadFile(metadataFile)
		if err != nil {
			return errors.NewEnhancedError(err, errors.ErrCodeInvalidInput,
				&errors.ErrorContext{
					Operation: "metadata_read",
					FileName:  metadataFile,
					CID:       cid,
					Suggestions: []string{
						"Check metadata file exists",
						"Verify read permissions",
						"Ensure file is not corrupted",
					},
				})
		}

		var fileInfo map[string]interface{}
		err = json.Unmarshal(metadataJSON, &fileInfo)
		if err != nil {
			return errors.NewEnhancedError(err, errors.ErrCodeInvalidInput,
				&errors.ErrorContext{
					Operation: "metadata_parse",
					FileName:  metadataFile,
					CID:       cid,
					Suggestions: []string{
						"Verify metadata file format",
						"Check file is valid JSON",
						"Ensure metadata was created by this tool",
					},
				})
		}

		if encrypted, ok := fileInfo["encrypted"].(bool); ok && encrypted {
			// Decrypt file
			metaData, ok := fileInfo["encryption_metadata"].(map[string]interface{})
			if !ok {
				return errors.NewEnhancedError(
					fmt.Errorf("encryption metadata not found"),
					errors.ErrCodeInvalidInput,
					&errors.ErrorContext{
						Operation: "metadata_parsing",
						FileName:  metadataFile,
						CID:       cid,
						Suggestions: []string{
							"Verify metadata file contains encryption information",
							"Check metadata file format",
							"Ensure file was encrypted with this tool",
						},
					})
			}

			// Convert metadata back to struct with safe type assertions
			converter := &utils.SafeJSONConverter{}

			salt, err := converter.Bytes(metaData["salt"], "salt")
			if err != nil {
				return errors.NewEnhancedError(err, errors.ErrCodeInvalidInput,
					&errors.ErrorContext{
						Operation: "metadata_parsing",
						FileName:  metadataFile,
						CID:       cid,
						Suggestions: []string{
							"Check metadata file integrity",
							"Verify salt format in metadata",
						},
					})
			}

			publicKeyBytes, err := converter.Bytes(metaData["public_key"], "public_key")
			if err != nil {
				return errors.NewEnhancedError(err, errors.ErrCodeInvalidInput,
					&errors.ErrorContext{
						Operation: "metadata_parsing",
						FileName:  metadataFile,
						CID:       cid,
						Suggestions: []string{
							"Check metadata file integrity",
							"Verify public key format in metadata",
						},
					})
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

			// Load public key
			publicKey, err := loadPublicKey(metadata.PublicKey)
			if err != nil {
				return fmt.Errorf("failed to load public key: %w", err)
			}

			outputData, err = encryption.DecryptWithMetadata(data, metadata, GetPassword(), publicKey)
			if err != nil {
				return fmt.Errorf("failed to decrypt file: %w", err)
			}
		} else {
			outputData = data
		}

		if fname, ok := fileInfo["filename"].(string); ok {
			filename = fname
		}
	} else {
		outputData = data
	}

	// Determine output path
	if outputPath == "" {
		outputPath = filename
	}

	// Write file
	err = os.WriteFile(outputPath, outputData, 0644)
	if err != nil {
		return errors.NewEnhancedError(err, errors.ErrCodeInvalidInput,
			&errors.ErrorContext{
				Operation: "file_write",
				FileName:  outputPath,
				CID:       cid,
				Suggestions: []string{
					"Check write permissions for output directory",
					"Ensure sufficient disk space",
					"Verify output path is valid",
				},
			})
	}

	fmt.Printf("File downloaded successfully!\n")
	fmt.Printf("Saved to: %s\n", outputPath)
	fmt.Printf("Size: %d bytes\n", len(outputData))

	return nil
}

// loadPublicKey loads a public key from bytes (helper function)
func loadPublicKey(keyBytes []byte) (ed25519.PublicKey, error) {
	// This would need to be implemented based on your key format
	return ed25519.PublicKey(keyBytes), nil
}
