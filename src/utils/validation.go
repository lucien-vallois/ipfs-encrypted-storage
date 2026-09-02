// Package utils provides comprehensive input validation utilities
package utils

import (
	"fmt"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"ipfs-encrypted-storage/src/errors"
)

// ValidationConfig holds configuration for validation rules
type ValidationConfig struct {
	MaxFileSize       int64
	MinPasswordLength int
	RequireMixedCase  bool
	RequireNumbers    bool
	RequireSymbols    bool
	CommonPasswords   []string
}

// DefaultValidationConfig returns default validation configuration
func DefaultValidationConfig() *ValidationConfig {
	return &ValidationConfig{
		MaxFileSize:       100 * 1024 * 1024, // 100MB
		MinPasswordLength: 8,
		RequireMixedCase:  true,
		RequireNumbers:    true,
		RequireSymbols:    false,
		CommonPasswords: []string{
			"password", "123456", "123456789", "qwerty", "abc123",
			"password123", "admin", "letmein", "welcome", "monkey",
			"1234567890", "iloveyou", "princess", "rockyou", "1234567",
			"12345678", "password1", "123123", "football", "baseball",
		},
	}
}

// ValidateFile performs comprehensive file validation
func ValidateFile(filePath string, config *ValidationConfig) *errors.EnhancedError {
	if config == nil {
		config = DefaultValidationConfig()
	}

	// Check if file exists and is accessible
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.NewEnhancedError(err, errors.ErrCodeInvalidInput,
				&errors.ErrorContext{
					Operation: "file_validation",
					FileName:  filePath,
					Suggestions: []string{
						"Check if the file path is correct",
						"Ensure the file exists",
						"Verify read permissions",
					},
				})
		}
		return errors.NewEnhancedError(err, errors.ErrCodeInvalidInput,
			&errors.ErrorContext{
				Operation: "file_validation",
				FileName:  filePath,
				Suggestions: []string{
					"Check file permissions",
					"Ensure the path is accessible",
				},
			})
	}

	// Check if it's actually a file (not a directory)
	if info.IsDir() {
		return errors.NewEnhancedError(
			fmt.Errorf("path is a directory, not a file"),
			errors.ErrCodeInvalidInput,
			&errors.ErrorContext{
				Operation: "file_validation",
				FileName:  filePath,
				Suggestions: []string{
					"Provide a file path instead of a directory",
					"Use a specific file within the directory",
				},
			})
	}

	// Validate file size
	if info.Size() > config.MaxFileSize {
		return errors.NewEnhancedError(
			fmt.Errorf("file too large: %d bytes (max: %d bytes)", info.Size(), config.MaxFileSize),
			errors.ErrCodeQuotaExceeded,
			&errors.ErrorContext{
				Operation: "file_validation",
				FileName:  filePath,
				Suggestions: []string{
					fmt.Sprintf("Maximum file size is %d MB", config.MaxFileSize/(1024*1024)),
					"Consider splitting large files",
					"Compress the file if possible",
				},
			})
	}

	// Validate filename
	if err := validateFilename(filepath.Base(filePath)); err != nil {
		return errors.NewEnhancedError(err, errors.ErrCodeInvalidInput,
			&errors.ErrorContext{
				Operation: "file_validation",
				FileName:  filePath,
				Suggestions: []string{
					"Use only alphanumeric characters, dots, hyphens, and underscores in filename",
					"Avoid special characters and spaces",
					"Keep filename length under 255 characters",
				},
			})
	}

	return nil
}

// ValidatePassword performs comprehensive password validation
func ValidatePassword(password string, config *ValidationConfig) *errors.EnhancedError {
	if config == nil {
		config = DefaultValidationConfig()
	}

	// Check minimum length
	if len(password) < config.MinPasswordLength {
		return errors.NewEnhancedError(
			fmt.Errorf("password too short: minimum %d characters", config.MinPasswordLength),
			errors.ErrCodeInvalidInput,
			&errors.ErrorContext{
				Operation: "password_validation",
				Suggestions: []string{
					fmt.Sprintf("Use at least %d characters", config.MinPasswordLength),
					"Include uppercase and lowercase letters",
					"Add numbers for better security",
					"Consider using special characters",
				},
			})
	}

	// Check for common passwords
	for _, common := range config.CommonPasswords {
		if strings.EqualFold(password, common) {
			return errors.NewEnhancedError(
				fmt.Errorf("password is too common and easily guessable"),
				errors.ErrCodeInvalidInput,
				&errors.ErrorContext{
					Operation: "password_validation",
					Suggestions: []string{
						"Avoid common passwords like 'password', '123456', etc.",
						"Use a unique passphrase",
						"Consider using password manager generated passwords",
					},
				})
		}
	}

	// Check character requirements
	var hasUpper, hasLower, hasNumber, hasSymbol bool
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSymbol = true
		}
	}

	var missingReqs []string

	if config.RequireMixedCase && (!hasUpper || !hasLower) {
		missingReqs = append(missingReqs, "both uppercase and lowercase letters")
	} else if config.RequireMixedCase && !hasUpper {
		missingReqs = append(missingReqs, "uppercase letters")
	} else if config.RequireMixedCase && !hasLower {
		missingReqs = append(missingReqs, "lowercase letters")
	}

	if config.RequireNumbers && !hasNumber {
		missingReqs = append(missingReqs, "numbers")
	}

	if config.RequireSymbols && !hasSymbol {
		missingReqs = append(missingReqs, "special characters")
	}

	if len(missingReqs) > 0 {
		return errors.NewEnhancedError(
			fmt.Errorf("password does not meet complexity requirements"),
			errors.ErrCodeInvalidInput,
			&errors.ErrorContext{
				Operation: "password_validation",
				Suggestions: []string{
					fmt.Sprintf("Password must include: %s", strings.Join(missingReqs, ", ")),
					"Use a mix of character types for better security",
				},
			})
	}

	// Calculate password entropy (basic estimation)
	entropy := calculatePasswordEntropy(password)
	if entropy < 50 {
		return errors.NewEnhancedError(
			fmt.Errorf("password entropy too low (%.1f bits)", entropy),
			errors.ErrCodeInvalidInput,
			&errors.ErrorContext{
				Operation: "password_validation",
				Suggestions: []string{
					"Use longer passwords for better security",
					"Incorporate more character variety",
					"Avoid predictable patterns",
				},
			})
	}

	return nil
}

// ValidateCID validates IPFS Content Identifier format
func ValidateCID(cid string) *errors.EnhancedError {
	if cid == "" {
		return errors.NewEnhancedError(
			fmt.Errorf("CID cannot be empty"),
			errors.ErrCodeInvalidInput,
			&errors.ErrorContext{
				Operation: "cid_validation",
				Suggestions: []string{
					"Provide a valid IPFS Content Identifier",
					"CIDs should start with 'Qm', 'bafy', or similar prefixes",
				},
			})
	}

	// Basic CID format validation (simplified)
	// In a real implementation, you'd use go-cid library for proper validation
	validPrefixes := []string{"Qm", "bafy", "bafz", "baga"}
	isValid := false

	for _, prefix := range validPrefixes {
		if strings.HasPrefix(cid, prefix) {
			isValid = true
			break
		}
	}

	if !isValid {
		return errors.NewEnhancedError(
			fmt.Errorf("invalid CID format"),
			errors.ErrCodeInvalidInput,
			&errors.ErrorContext{
				Operation: "cid_validation",
				Resource:  cid,
				Suggestions: []string{
					"CID should start with valid prefix (Qm, bafy, bafz, baga, etc.)",
					"Verify the CID was copied correctly",
					"Ensure the content is properly pinned",
				},
			})
	}

	// Check length (basic validation)
	if len(cid) < 10 {
		return errors.NewEnhancedError(
			fmt.Errorf("CID too short"),
			errors.ErrCodeInvalidInput,
			&errors.ErrorContext{
				Operation: "cid_validation",
				Resource:  cid,
				Suggestions: []string{
					"Valid CIDs are typically longer than 10 characters",
					"Check if the CID is complete",
				},
			})
	}

	return nil
}

// ValidatePeerAddress validates libp2p peer multiaddress format
func ValidatePeerAddress(peerAddr string) *errors.EnhancedError {
	if peerAddr == "" {
		return errors.NewEnhancedError(
			fmt.Errorf("peer address cannot be empty"),
			errors.ErrCodeInvalidInput,
			&errors.ErrorContext{
				Operation: "peer_validation",
				Suggestions: []string{
					"Provide a valid peer multiaddress",
					"Example: /ip4/192.168.1.100/tcp/4001/p2p/peerID",
				},
			})
	}

	// Check if it starts with valid IP protocol
	validStarts := []string{"/ip4/", "/ip6/"}
	isValidStart := false

	for _, start := range validStarts {
		if strings.HasPrefix(peerAddr, start) {
			isValidStart = true
			break
		}
	}

	if !isValidStart {
		return errors.NewEnhancedError(
			fmt.Errorf("invalid peer address format"),
			errors.ErrCodeInvalidInput,
			&errors.ErrorContext{
				Operation: "peer_validation",
				Resource:  peerAddr,
				Suggestions: []string{
					"Peer address must start with /ip4/ or /ip6/",
					"Example: /ip4/192.168.1.100/tcp/4001/p2p/peerID",
					"Include protocol, IP, port, and peer ID",
				},
			})
	}

	// Basic structure validation
	parts := strings.Split(peerAddr, "/")
	if len(parts) < 6 {
		return errors.NewEnhancedError(
			fmt.Errorf("incomplete peer address"),
			errors.ErrCodeInvalidInput,
			&errors.ErrorContext{
				Operation: "peer_validation",
				Resource:  peerAddr,
				Suggestions: []string{
					"Peer address should include: /ip4|ip6/IP/tcp/PORT/p2p/PEER_ID",
					"Check that all required components are present",
				},
			})
	}

	return nil
}

// ValidateIPFSEndpoint validates IPFS API endpoint URL
func ValidateIPFSEndpoint(endpoint string) *errors.EnhancedError {
	if endpoint == "" {
		return errors.NewEnhancedError(
			fmt.Errorf("IPFS endpoint cannot be empty"),
			errors.ErrCodeInvalidInput,
			&errors.ErrorContext{
				Operation: "endpoint_validation",
				Suggestions: []string{
					"Provide IPFS API endpoint URL",
					"Default is usually localhost:5001",
					"Format: host:port or http://host:port",
				},
			})
	}

	// Parse as URL
	parsedURL, err := url.Parse(endpoint)
	if err != nil {
		// Try adding http:// prefix if not present
		if !strings.Contains(endpoint, "://") {
			parsedURL, err = url.Parse("http://" + endpoint)
			if err != nil {
				return errors.NewEnhancedError(
					fmt.Errorf("invalid URL format: %w", err),
					errors.ErrCodeInvalidInput,
					&errors.ErrorContext{
						Operation: "endpoint_validation",
						Resource:  endpoint,
						Suggestions: []string{
							"Use format: host:port or http://host:port",
							"Example: localhost:5001 or http://localhost:5001",
						},
					})
			}
		} else {
			return errors.NewEnhancedError(
				fmt.Errorf("invalid URL format: %w", err),
				errors.ErrCodeInvalidInput,
				&errors.ErrorContext{
					Operation: "endpoint_validation",
					Resource:  endpoint,
					Suggestions: []string{
						"Check URL syntax and format",
						"Ensure protocol (http/https) is specified",
					},
				})
		}
	}

	// Validate scheme
	if parsedURL.Scheme != "" && parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return errors.NewEnhancedError(
			fmt.Errorf("unsupported URL scheme: %s", parsedURL.Scheme),
			errors.ErrCodeInvalidInput,
			&errors.ErrorContext{
				Operation: "endpoint_validation",
				Resource:  endpoint,
				Suggestions: []string{
					"Use http or https protocol",
					"For local IPFS daemon, use http://localhost:5001",
				},
			})
	}

	// Validate host and port
	if parsedURL.Hostname() == "" {
		return errors.NewEnhancedError(
			fmt.Errorf("missing hostname in URL"),
			errors.ErrCodeInvalidInput,
			&errors.ErrorContext{
				Operation: "endpoint_validation",
				Resource:  endpoint,
				Suggestions: []string{
					"Include hostname (e.g., localhost, 127.0.0.1)",
					"For local development: localhost:5001",
				},
			})
	}

	return nil
}

// Helper functions

func validateFilename(filename string) error {
	if filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}

	if len(filename) > 255 {
		return fmt.Errorf("filename too long: %d characters (max 255)", len(filename))
	}

	// Check for invalid characters
	invalidChars := regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f]`)
	if invalidChars.MatchString(filename) {
		return fmt.Errorf("filename contains invalid characters")
	}

	// Check for reserved names (Windows)
	reservedNames := []string{
		"CON", "PRN", "AUX", "NUL",
		"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
		"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9",
	}

	upperName := strings.ToUpper(filename)
	for _, reserved := range reservedNames {
		if upperName == reserved {
			return fmt.Errorf("filename is a reserved name")
		}
	}

	return nil
}

func calculatePasswordEntropy(password string) float64 {
	charsetSize := 0

	var hasLower, hasUpper, hasDigit, hasSymbol bool
	for _, char := range password {
		switch {
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSymbol = true
		}
	}

	if hasLower {
		charsetSize += 26
	}
	if hasUpper {
		charsetSize += 26
	}
	if hasDigit {
		charsetSize += 10
	}
	if hasSymbol {
		charsetSize += 32 // Approximate for common symbols
	}

	if charsetSize == 0 {
		return 0
	}

	// Basic entropy calculation: length * log2(charsetSize)
	// This is a simplified estimation
	return float64(len(password)) * math.Log2(float64(charsetSize))
}
