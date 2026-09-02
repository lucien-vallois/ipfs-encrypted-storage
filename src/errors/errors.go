// Package errors provides enhanced error handling utilities
package errors

import (
	"fmt"
	"strings"
)

// ErrorContext represents contextual information about an error
type ErrorContext struct {
	Operation   string
	Resource    string
	UserID      string
	FileName    string
	CID         string
	RetryCount  int
	Suggestions []string
}

// EnhancedError wraps an error with additional context and recovery suggestions
type EnhancedError struct {
	Err     error
	Context *ErrorContext
	Code    ErrorCode
}

func (e *EnhancedError) Error() string {
	var msg strings.Builder

	if e.Err != nil {
		msg.WriteString(e.Err.Error())
	}

	if e.Context != nil {
		if e.Context.Operation != "" {
			msg.WriteString(fmt.Sprintf(" (operation: %s)", e.Context.Operation))
		}
		if e.Context.Resource != "" {
			msg.WriteString(fmt.Sprintf(" (resource: %s)", e.Context.Resource))
		}
		if e.Context.FileName != "" {
			msg.WriteString(fmt.Sprintf(" (file: %s)", e.Context.FileName))
		}
		if e.Context.CID != "" {
			msg.WriteString(fmt.Sprintf(" (CID: %s)", e.Context.CID))
		}
		if e.Context.UserID != "" {
			msg.WriteString(fmt.Sprintf(" (user: %s)", e.Context.UserID))
		}
	}

	return msg.String()
}

// Suggestions returns recovery suggestions for the error
func (e *EnhancedError) Suggestions() []string {
	if e.Context != nil && len(e.Context.Suggestions) > 0 {
		return e.Context.Suggestions
	}

	// Default suggestions based on error code
	switch e.Code {
	case ErrCodeNetworkFailure:
		return []string{
			"Check your internet connection",
			"Verify IPFS daemon is running",
			"Try again in a few moments",
			"Check firewall settings",
		}
	case ErrCodeInvalidInput:
		return []string{
			"Verify input parameters are correct",
			"Check file format and size limits",
			"Ensure all required fields are provided",
		}
	case ErrCodeAuthentication:
		return []string{
			"Verify your credentials",
			"Check password strength requirements",
			"Ensure key pair is valid",
		}
	case ErrCodePermissionDenied:
		return []string{
			"Verify you have access to this resource",
			"Check your DID permissions",
			"Contact resource owner for access",
		}
	case ErrCodeResourceNotFound:
		return []string{
			"Verify the CID or resource ID",
			"Check if the file was deleted",
			"Ensure IPFS node has the content",
		}
	case ErrCodeQuotaExceeded:
		return []string{
			"Reduce file size",
			"Delete unused files",
			"Contact administrator for quota increase",
		}
	default:
		return []string{
			"Check system logs for more details",
			"Try the operation again",
			"Contact support if the problem persists",
		}
	}
}

// ErrorCode represents different types of errors
type ErrorCode int

const (
	ErrCodeUnknown ErrorCode = iota
	ErrCodeNetworkFailure
	ErrCodeInvalidInput
	ErrCodeAuthentication
	ErrCodePermissionDenied
	ErrCodeResourceNotFound
	ErrCodeQuotaExceeded
	ErrCodeInternalError
)

// NewEnhancedError creates a new enhanced error
func NewEnhancedError(err error, code ErrorCode, context *ErrorContext) *EnhancedError {
	return &EnhancedError{
		Err:     err,
		Code:    code,
		Context: context,
	}
}

// ErrorHandler provides centralized error handling
type ErrorHandler struct {
	MaxRetries int
}

// HandleError processes an error and returns enhanced error information
func (eh *ErrorHandler) HandleError(err error, context *ErrorContext) *EnhancedError {
	if err == nil {
		return nil
	}

	// Classify error type
	code := eh.classifyError(err)

	// Add context-specific suggestions
	if context != nil {
		context.Suggestions = eh.generateSuggestions(code, context)
	}

	return NewEnhancedError(err, code, context)
}

// classifyError determines the error code based on error content
func (eh *ErrorHandler) classifyError(err error) ErrorCode {
	errStr := strings.ToLower(err.Error())

	switch {
	case strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "network") ||
		strings.Contains(errStr, "timeout"):
		return ErrCodeNetworkFailure

	case strings.Contains(errStr, "invalid") ||
		strings.Contains(errStr, "malformed") ||
		strings.Contains(errStr, "bad format"):
		return ErrCodeInvalidInput

	case strings.Contains(errStr, "unauthorized") ||
		strings.Contains(errStr, "authentication") ||
		strings.Contains(errStr, "password"):
		return ErrCodeAuthentication

	case strings.Contains(errStr, "permission") ||
		strings.Contains(errStr, "access denied"):
		return ErrCodePermissionDenied

	case strings.Contains(errStr, "not found") ||
		strings.Contains(errStr, "does not exist"):
		return ErrCodeResourceNotFound

	case strings.Contains(errStr, "quota") ||
		strings.Contains(errStr, "limit exceeded") ||
		strings.Contains(errStr, "too large"):
		return ErrCodeQuotaExceeded

	default:
		return ErrCodeUnknown
	}
}

// generateSuggestions creates context-specific recovery suggestions
func (eh *ErrorHandler) generateSuggestions(code ErrorCode, context *ErrorContext) []string {
	baseSuggestions := []string{}

	// Add context-specific suggestions
	switch code {
	case ErrCodeNetworkFailure:
		if context != nil && context.Operation == "IPFS Upload" {
			baseSuggestions = append(baseSuggestions,
				"Ensure IPFS daemon is running: 'ipfs daemon'",
				"Check IPFS API endpoint configuration",
				"Verify network connectivity to IPFS node")
		}
	case ErrCodeInvalidInput:
		if context != nil && context.FileName != "" {
			baseSuggestions = append(baseSuggestions,
				fmt.Sprintf("Check file '%s' format and encoding", context.FileName),
				"Ensure file size is within limits (max 100MB)",
				"Verify file is not corrupted")
		}
		if context != nil && context.CID != "" {
			baseSuggestions = append(baseSuggestions,
				fmt.Sprintf("Verify CID '%s' format", context.CID),
				"Ensure CID was copied correctly",
				"Check if content is available on the network")
		}

	case ErrCodeAuthentication:
		if context != nil && context.Operation == "Encryption" {
			baseSuggestions = append(baseSuggestions,
				"Use a password with at least 8 characters",
				"Include uppercase, lowercase, and numeric characters",
				"Ensure you're using the correct password for decryption")
		}

	case ErrCodePermissionDenied:
		if context != nil && context.Resource != "" {
			baseSuggestions = append(baseSuggestions,
				fmt.Sprintf("Request access to resource '%s' from owner", context.Resource),
				"Verify your DID has necessary permissions",
				"Check ZKP proofs are valid and current")
		}

	case ErrCodeResourceNotFound:
		if context != nil && context.CID != "" {
			baseSuggestions = append(baseSuggestions,
				fmt.Sprintf("Content with CID '%s' may not be available", context.CID),
				"Try pinning the content to ensure availability",
				"Check if the content was garbage collected")
		}
	}

	// Add retry suggestions if applicable
	if context != nil && context.RetryCount < eh.MaxRetries {
		baseSuggestions = append(baseSuggestions,
			fmt.Sprintf("Retry the operation (attempt %d/%d)", context.RetryCount+1, eh.MaxRetries))
	}

	return baseSuggestions
}

// FormatErrorMessage creates a user-friendly error message
func FormatErrorMessage(err *EnhancedError) string {
	var msg strings.Builder

	msg.WriteString("Error: ")
	msg.WriteString(err.Error())
	msg.WriteString("\n\n")

	if suggestions := err.Suggestions(); len(suggestions) > 0 {
		msg.WriteString("Suggestions:\n")
		for i, suggestion := range suggestions {
			msg.WriteString(fmt.Sprintf("   %d. %s\n", i+1, suggestion))
		}
		msg.WriteString("\n")
	}

	msg.WriteString("For more help, check the logs or documentation.")

	return msg.String()
}

// RetryableError checks if an error is retryable
func RetryableError(err *EnhancedError) bool {
	switch err.Code {
	case ErrCodeNetworkFailure:
		return true
	case ErrCodeResourceNotFound:
		// Some not found errors might be retryable
		return strings.Contains(err.Err.Error(), "temporary")
	default:
		return false
	}
}

// LogError logs an error with appropriate level
func LogError(err *EnhancedError, logger interface{}) {
	// This would integrate with actual logging system
	// For now, just print to console
	fmt.Printf("ERROR [%d]: %s\n", err.Code, err.Error())

	if suggestions := err.Suggestions(); len(suggestions) > 0 {
		fmt.Println("Suggestions:")
		for _, suggestion := range suggestions {
			fmt.Printf("  - %s\n", suggestion)
		}
	}
}

// ValidateOperation pre-validates operation parameters
func ValidateOperation(operation string, params map[string]interface{}) *EnhancedError {
	switch operation {
	case "encrypt":
		if password, ok := params["password"].(string); ok {
			if len(password) < 8 {
				return NewEnhancedError(
					fmt.Errorf("password too short"),
					ErrCodeInvalidInput,
					&ErrorContext{
						Operation: operation,
						Suggestions: []string{
							"Use a password with at least 8 characters",
							"Include uppercase, lowercase, and digits",
						},
					},
				)
			}
		}

	case "upload":
		if fileSize, ok := params["file_size"].(int64); ok && fileSize > 100*1024*1024 {
			return NewEnhancedError(
				fmt.Errorf("file too large: %d bytes", fileSize),
				ErrCodeQuotaExceeded,
				&ErrorContext{
					Operation: operation,
					Suggestions: []string{
						"Split large files into smaller chunks",
						"Compress the file before uploading",
						"Maximum file size is 100MB",
					},
				},
			)
		}

	case "connect":
		if peerAddr, ok := params["peer_address"].(string); ok {
			if !strings.HasPrefix(peerAddr, "/ip4/") && !strings.HasPrefix(peerAddr, "/ip6/") {
				return NewEnhancedError(
					fmt.Errorf("invalid peer address format"),
					ErrCodeInvalidInput,
					&ErrorContext{
						Operation: operation,
						Suggestions: []string{
							"Peer address must start with /ip4/ or /ip6/",
							"Example: /ip4/192.168.1.100/tcp/4001/p2p/peerID",
						},
					},
				)
			}
		}
	}

	return nil
}
