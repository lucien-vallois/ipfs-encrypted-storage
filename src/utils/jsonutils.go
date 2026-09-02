// Package utils provides shared utilities for the IPFS encrypted storage system
package utils

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
)

// SafeJSONConverter provides safe type conversions from JSON interface{}
type SafeJSONConverter struct{}

// Bytes converts interface{} to []byte with multiple format support
func (c *SafeJSONConverter) Bytes(v interface{}, fieldName string) ([]byte, error) {
	if v == nil {
		return nil, fmt.Errorf("%s is nil", fieldName)
	}

	switch val := v.(type) {
	case []byte:
		return val, nil
	case []interface{}:
		// Convert []interface{} to []byte
		result := make([]byte, len(val))
		for i, item := range val {
			if num, ok := item.(float64); ok {
				result[i] = byte(num)
			} else {
				return nil, fmt.Errorf("%s contains invalid byte value at index %d", fieldName, i)
			}
		}
		return result, nil
	case string:
		// Try base64 first
		if data, err := base64.StdEncoding.DecodeString(val); err == nil {
			return data, nil
		}
		// Try hex
		if data, err := hex.DecodeString(val); err == nil {
			return data, nil
		}
		// Fall back to raw bytes
		return []byte(val), nil
	default:
		return nil, fmt.Errorf("%s has invalid type: expected []byte, got %T", fieldName, v)
	}
}

// Uint32 converts interface{} to uint32
func (c *SafeJSONConverter) Uint32(v interface{}, fieldName string) (uint32, error) {
	if v == nil {
		return 0, fmt.Errorf("%s is nil", fieldName)
	}

	switch val := v.(type) {
	case uint32:
		return val, nil
	case float64:
		return uint32(val), nil
	case int:
		return uint32(val), nil
	case int64:
		return uint32(val), nil
	case string:
		// Try parsing as number
		if num, err := strconv.ParseUint(val, 10, 32); err == nil {
			return uint32(num), nil
		}
		return 0, fmt.Errorf("%s has invalid string format for uint32: %s", fieldName, val)
	default:
		return 0, fmt.Errorf("%s has invalid type: expected number, got %T", fieldName, v)
	}
}

// Uint8 converts interface{} to uint8
func (c *SafeJSONConverter) Uint8(v interface{}, fieldName string) (uint8, error) {
	if v == nil {
		return 0, fmt.Errorf("%s is nil", fieldName)
	}

	switch val := v.(type) {
	case uint8:
		return val, nil
	case float64:
		return uint8(val), nil
	case int:
		return uint8(val), nil
	case string:
		// Try parsing as number
		if num, err := strconv.ParseUint(val, 10, 8); err == nil {
			return uint8(num), nil
		}
		return 0, fmt.Errorf("%s has invalid string format for uint8: %s", fieldName, val)
	default:
		return 0, fmt.Errorf("%s has invalid type: expected number, got %T", fieldName, v)
	}
}

// String converts interface{} to string
func (c *SafeJSONConverter) String(v interface{}, fieldName string) (string, error) {
	if v == nil {
		return "", fmt.Errorf("%s is nil", fieldName)
	}

	switch val := v.(type) {
	case string:
		return val, nil
	default:
		return fmt.Sprintf("%v", val), nil
	}
}

// Bool converts interface{} to bool
func (c *SafeJSONConverter) Bool(v interface{}, fieldName string) (bool, error) {
	if v == nil {
		return false, fmt.Errorf("%s is nil", fieldName)
	}

	switch val := v.(type) {
	case bool:
		return val, nil
	case string:
		// Try parsing common boolean strings
		switch strings.ToLower(val) {
		case "true", "yes", "1", "on":
			return true, nil
		case "false", "no", "0", "off":
			return false, nil
		default:
			return false, fmt.Errorf("%s has invalid boolean string: %s", fieldName, val)
		}
	case float64:
		return val != 0, nil
	case int:
		return val != 0, nil
	default:
		return false, fmt.Errorf("%s has invalid type: expected bool, got %T", fieldName, v)
	}
}
