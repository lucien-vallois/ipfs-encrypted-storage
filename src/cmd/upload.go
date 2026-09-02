package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"ipfs-encrypted-storage/src/encryption"
	"ipfs-encrypted-storage/src/errors"
	"ipfs-encrypted-storage/src/ipfs"
	"ipfs-encrypted-storage/src/utils"
)

// NewUploadCmd creates the upload command
func NewUploadCmd() *cobra.Command {
	var (
		filePath    string
		encrypt     bool
		outputFile  string
		description string
	)

	cmd := &cobra.Command{
		Use:   "upload [file]",
		Short: "Upload and encrypt a file to IPFS",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath = args[0]
			return runUpload(filePath, outputFile, description, encrypt)
		},
	}

	cmd.Flags().BoolVar(&encrypt, "encrypt", true, "Encrypt file before uploading")
	cmd.Flags().StringVar(&outputFile, "output", "", "Output file for metadata (default: <filename>.meta.json)")
	cmd.Flags().StringVar(&description, "description", "", "File description")

	return cmd
}

func runUpload(filePath, outputFile, description string, encrypt bool) error {
	// Validate file before processing
	if err := utils.ValidateFile(filePath, nil); err != nil {
		return err
	}

	client, err := ipfs.NewIPFSClient(GetIPFSURL())
	if err != nil {
		return errors.NewEnhancedError(err, errors.ErrCodeNetworkFailure,
			&errors.ErrorContext{
				Operation: "IPFS_client_creation",
				Resource:  GetIPFSURL(),
				Suggestions: []string{
					"Ensure IPFS daemon is running",
					"Check IPFS API endpoint configuration",
					"Verify network connectivity",
				},
			})
	}
	defer client.Close()

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return errors.NewEnhancedError(err, errors.ErrCodeInvalidInput,
			&errors.ErrorContext{
				Operation: "file_read",
				FileName:  filePath,
				Suggestions: []string{
					"Check file permissions",
					"Ensure file is not corrupted",
					"Verify file path is correct",
				},
			})
	}

	var cid string
	var metadata *encryption.EncryptedMetadata

	if encrypt {
		// Generate key pair for signing
		_, privateKey, err := encryption.GenerateKeyPair()
		if err != nil {
			return errors.NewEnhancedError(err, errors.ErrCodeInternalError,
				&errors.ErrorContext{
					Operation: "key_generation",
					FileName:  filePath,
					Suggestions: []string{
						"Check system entropy source",
						"Try again in a moment",
						"Contact support if problem persists",
					},
				})
		}

		// Encrypt and upload
		ciphertext, meta, err := encryption.EncryptWithMetadata(data, GetPassword(), privateKey)
		if err != nil {
			return errors.NewEnhancedError(err, errors.ErrCodeInternalError,
				&errors.ErrorContext{
					Operation: "file_encryption",
					FileName:  filePath,
					Suggestions: []string{
						"Verify password is correct",
						"Check file is not corrupted",
						"Ensure sufficient memory available",
					},
				})
		}

		cid, err = client.AddFile(ciphertext, filepath.Base(filePath))
		if err != nil {
			return errors.NewEnhancedError(err, errors.ErrCodeNetworkFailure,
				&errors.ErrorContext{
					Operation: "IPFS_upload",
					FileName:  filePath,
					Suggestions: []string{
						"Check IPFS daemon is running",
						"Verify network connectivity",
						"Try again in a few moments",
					},
				})
		}

		metadata = meta
	} else {
		// Upload without encryption
		cid, err = client.AddFile(data, filepath.Base(filePath))
		if err != nil {
			return errors.NewEnhancedError(err, errors.ErrCodeNetworkFailure,
				&errors.ErrorContext{
					Operation: "IPFS_upload",
					FileName:  filePath,
					Suggestions: []string{
						"Check IPFS daemon is running",
						"Verify network connectivity",
						"Ensure sufficient disk space on IPFS node",
					},
				})
		}
	}

	// Prepare metadata
	fileInfo := map[string]interface{}{
		"cid":         cid,
		"filename":    filepath.Base(filePath),
		"size":        len(data),
		"uploaded_at": time.Now().Unix(),
		"description": description,
		"encrypted":   encrypt,
	}

	if metadata != nil {
		fileInfo["encryption_metadata"] = metadata
	}

	// Save metadata
	if outputFile == "" {
		outputFile = strings.TrimSuffix(filePath, filepath.Ext(filePath)) + ".meta.json"
	}

	metadataJSON, err := json.MarshalIndent(fileInfo, "", "  ")
	if err != nil {
		return errors.NewEnhancedError(err, errors.ErrCodeInternalError,
			&errors.ErrorContext{
				Operation: "metadata_marshal",
				FileName:  filePath,
				Suggestions: []string{
					"Check file metadata is valid",
					"Ensure sufficient memory for marshaling",
				},
			})
	}

	err = os.WriteFile(outputFile, metadataJSON, 0644)
	if err != nil {
		return errors.NewEnhancedError(err, errors.ErrCodeInvalidInput,
			&errors.ErrorContext{
				Operation: "metadata_save",
				FileName:  outputFile,
				Suggestions: []string{
					"Check write permissions for output directory",
					"Ensure sufficient disk space",
					"Verify output path is valid",
				},
			})
	}

	fmt.Printf("File uploaded successfully!\n")
	fmt.Printf("CID: %s\n", cid)
	fmt.Printf("Metadata saved to: %s\n", outputFile)

	return nil
}
