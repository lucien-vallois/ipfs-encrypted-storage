// Package models provides error response structures for the REST API
package models

import (
	"net/http"
	"time"

	"ipfs-encrypted-storage/src/errors"
)

// APIError represents a standardized API error response
type APIError struct {
	Error       string   `json:"error"`
	Code        string   `json:"code"`
	Message     string   `json:"message,omitempty"`
	Suggestions []string `json:"suggestions,omitempty"`
	Operation   string   `json:"operation,omitempty"`
	Resource    string   `json:"resource,omitempty"`
	Timestamp   int64    `json:"timestamp"`
}

// NewAPIError creates a new API error from an enhanced error
func NewAPIError(err *errors.EnhancedError) *APIError {
	apiErr := &APIError{
		Error:       err.Error(),
		Code:        errorCodeToString(err.Code),
		Suggestions: err.Suggestions(),
		Timestamp:   time.Now().Unix(),
	}

	// Add context information if available
	if err.Context != nil {
		apiErr.Operation = err.Context.Operation
		apiErr.Resource = err.Context.Resource
		if err.Context.FileName != "" {
			apiErr.Resource = err.Context.FileName
		} else if err.Context.CID != "" {
			apiErr.Resource = err.Context.CID
		}
	}

	return apiErr
}

// NewAPIErrorFromError creates a new API error from a regular error
func NewAPIErrorFromError(err error) *APIError {
	return &APIError{
		Error:     err.Error(),
		Code:      "INTERNAL_ERROR",
		Timestamp: time.Now().Unix(),
		Suggestions: []string{
			"Check system logs for more details",
			"Try the operation again",
			"Contact support if the problem persists",
		},
	}
}

// errorCodeToString converts error codes to string representations
func errorCodeToString(code errors.ErrorCode) string {
	switch code {
	case errors.ErrCodeUnknown:
		return "UNKNOWN_ERROR"
	case errors.ErrCodeNetworkFailure:
		return "NETWORK_ERROR"
	case errors.ErrCodeInvalidInput:
		return "INVALID_INPUT"
	case errors.ErrCodeAuthentication:
		return "AUTHENTICATION_FAILED"
	case errors.ErrCodePermissionDenied:
		return "PERMISSION_DENIED"
	case errors.ErrCodeResourceNotFound:
		return "RESOURCE_NOT_FOUND"
	case errors.ErrCodeQuotaExceeded:
		return "QUOTA_EXCEEDED"
	case errors.ErrCodeInternalError:
		return "INTERNAL_ERROR"
	default:
		return "UNKNOWN_ERROR"
	}
}

// GetHTTPStatusCode returns the appropriate HTTP status code for an error
func (e *APIError) GetHTTPStatusCode() int {
	switch e.Code {
	case "INVALID_INPUT":
		return http.StatusBadRequest
	case "AUTHENTICATION_FAILED":
		return http.StatusUnauthorized
	case "PERMISSION_DENIED":
		return http.StatusForbidden
	case "RESOURCE_NOT_FOUND":
		return http.StatusNotFound
	case "QUOTA_EXCEEDED":
		return http.StatusRequestEntityTooLarge
	case "NETWORK_ERROR":
		return http.StatusBadGateway
	case "INTERNAL_ERROR":
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// ValidationError represents validation-specific errors
type ValidationError struct {
	Field       string   `json:"field"`
	Value       string   `json:"value,omitempty"`
	Message     string   `json:"message"`
	Suggestions []string `json:"suggestions,omitempty"`
}

// ValidationErrors represents multiple validation errors
type ValidationErrors struct {
	Errors  []ValidationError `json:"errors"`
	Message string            `json:"message"`
	Code    string            `json:"code"`
}

// NewValidationErrors creates validation errors from a list of field errors
func NewValidationErrors(fieldErrors map[string]string) *ValidationErrors {
	errors := make([]ValidationError, 0, len(fieldErrors))

	for field, message := range fieldErrors {
		errors = append(errors, ValidationError{
			Field:   field,
			Message: message,
			Suggestions: []string{
				"Check the field format and requirements",
				"Refer to API documentation for valid values",
			},
		})
	}

	return &ValidationErrors{
		Errors:  errors,
		Message: "Validation failed",
		Code:    "VALIDATION_ERROR",
	}
}

// GetHTTPStatusCode returns HTTP status code for validation errors
func (ve *ValidationErrors) GetHTTPStatusCode() int {
	return http.StatusBadRequest
}
