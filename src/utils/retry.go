// Package utils provides utility functions including retry logic
package utils

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"
)

// RetryConfig holds configuration for retry operations
type RetryConfig struct {
	MaxRetries   int
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
}

// DefaultRetryConfig returns a retry configuration with sensible defaults
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:   3,
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
	}
}

// RetryWithBackoff executes a function with exponential backoff retry logic
func RetryWithBackoff(ctx context.Context, fn func() error, config *RetryConfig) error {
	if config == nil {
		config = DefaultRetryConfig()
	}

	var lastErr error
	delay := config.InitialDelay

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled: %w", ctx.Err())
		default:
		}

		// Execute the function
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Don't wait after the last attempt
		if attempt < config.MaxRetries {
			// Wait with exponential backoff
			select {
			case <-ctx.Done():
				return fmt.Errorf("context cancelled: %w", ctx.Err())
			case <-time.After(delay):
				// Calculate next delay
				delay = time.Duration(float64(delay) * config.Multiplier)
				if delay > config.MaxDelay {
					delay = config.MaxDelay
				}
			}
		}
	}

	return fmt.Errorf("max retries exceeded: %w", lastErr)
}

// RetryWithExponentialBackoff is a convenience function with exponential backoff
func RetryWithExponentialBackoff(ctx context.Context, fn func() error, maxRetries int) error {
	config := &RetryConfig{
		MaxRetries:   maxRetries,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     5 * time.Second,
		Multiplier:   2.0,
	}
	return RetryWithBackoff(ctx, fn, config)
}

// RetryWithLinearBackoff executes a function with linear backoff
func RetryWithLinearBackoff(ctx context.Context, fn func() error, maxRetries int, delay time.Duration) error {
	config := &RetryConfig{
		MaxRetries:   maxRetries,
		InitialDelay: delay,
		MaxDelay:     delay,
		Multiplier:   1.0,
	}
	return RetryWithBackoff(ctx, fn, config)
}

// RetryWithJitter adds jitter to the delay to prevent thundering herd
func RetryWithJitter(ctx context.Context, fn func() error, config *RetryConfig) error {
	if config == nil {
		config = DefaultRetryConfig()
	}

	var lastErr error
	baseDelay := config.InitialDelay

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled: %w", ctx.Err())
		default:
		}

		// Execute the function
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Don't wait after the last attempt
		if attempt < config.MaxRetries {
			// Calculate delay with exponential backoff and jitter
			delay := time.Duration(float64(baseDelay) * math.Pow(config.Multiplier, float64(attempt)))
			if delay > config.MaxDelay {
				delay = config.MaxDelay
			}

			// Add jitter (random variation up to 25% of delay)
			jitter := time.Duration(float64(delay) * 0.25 * (float64(attempt%10) / 10.0))
			delay = delay + jitter

			select {
			case <-ctx.Done():
				return fmt.Errorf("context cancelled: %w", ctx.Err())
			case <-time.After(delay):
			}
		}
	}

	return fmt.Errorf("max retries exceeded: %w", lastErr)
}

// IsRetryableError checks if an error is retryable
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	// Network-related errors are typically retryable
	retryablePatterns := []string{
		"connection refused",
		"timeout",
		"network",
		"temporary",
		"unavailable",
		"EOF",
	}

	for _, pattern := range retryablePatterns {
		if strings.Contains(strings.ToLower(errStr), strings.ToLower(pattern)) {
			return true
		}
	}

	return false
}
