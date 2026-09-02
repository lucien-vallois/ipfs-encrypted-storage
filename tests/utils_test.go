package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"ipfs-encrypted-storage/src/utils"
)

func TestRetryWithBackoff(t *testing.T) {
	ctx := context.Background()
	attempts := 0
	maxAttempts := 3

	fn := func() error {
		attempts++
		if attempts < maxAttempts {
			return errors.New("temporary error")
		}
		return nil
	}

	config := utils.DefaultRetryConfig()
	config.MaxRetries = maxAttempts

	err := utils.RetryWithBackoff(ctx, fn, config)
	if err != nil {
		t.Errorf("Retry should succeed after %d attempts: %v", maxAttempts, err)
	}

	if attempts != maxAttempts {
		t.Errorf("Expected %d attempts, got %d", maxAttempts, attempts)
	}
}

func TestRetryMaxAttempts(t *testing.T) {
	ctx := context.Background()
	attempts := 0

	fn := func() error {
		attempts++
		return errors.New("persistent error")
	}

	config := utils.DefaultRetryConfig()
	config.MaxRetries = 2

	err := utils.RetryWithBackoff(ctx, fn, config)
	if err == nil {
		t.Error("Retry should fail after max attempts")
	}

	expectedAttempts := config.MaxRetries + 1
	if attempts != expectedAttempts {
		t.Errorf("Expected %d attempts, got %d", expectedAttempts, attempts)
	}
}

func TestRetryContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	attempts := 0

	fn := func() error {
		attempts++
		if attempts == 2 {
			cancel() // Cancel context after second attempt
		}
		return errors.New("error")
	}

	config := utils.DefaultRetryConfig()
	config.MaxRetries = 5

	err := utils.RetryWithBackoff(ctx, fn, config)
	if err == nil {
		t.Error("Retry should fail when context is cancelled")
	}

	if attempts > 3 {
		t.Errorf("Should stop retrying after context cancellation, got %d attempts", attempts)
	}
}

func TestIsRetryableError(t *testing.T) {
	testCases := []struct {
		err       error
		retryable bool
	}{
		{errors.New("connection refused"), true},
		{errors.New("timeout"), true},
		{errors.New("network error"), true},
		{errors.New("temporary failure"), true},
		{errors.New("unavailable"), true},
		{errors.New("EOF"), true},
		{errors.New("invalid input"), false},
		{errors.New("authentication failed"), false},
		{errors.New("permission denied"), false},
		{nil, false},
	}

	for _, tc := range testCases {
		result := utils.IsRetryableError(tc.err)
		if result != tc.retryable {
			t.Errorf("IsRetryableError(%v) = %v, expected %v", tc.err, result, tc.retryable)
		}
	}
}

func TestRetryWithExponentialBackoff(t *testing.T) {
	ctx := context.Background()
	attempts := 0

	fn := func() error {
		attempts++
		if attempts < 2 {
			return errors.New("temporary error")
		}
		return nil
	}

	err := utils.RetryWithExponentialBackoff(ctx, fn, 3)
	if err != nil {
		t.Errorf("Retry should succeed: %v", err)
	}

	if attempts != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempts)
	}
}

func TestRetryWithLinearBackoff(t *testing.T) {
	ctx := context.Background()
	attempts := 0

	fn := func() error {
		attempts++
		if attempts < 2 {
			return errors.New("temporary error")
		}
		return nil
	}

	start := time.Now()
	err := utils.RetryWithLinearBackoff(ctx, fn, 3, 100*time.Millisecond)
	elapsed := time.Since(start)

	if err != nil {
		t.Errorf("Retry should succeed: %v", err)
	}

	// Should have waited at least 100ms between attempts
	if elapsed < 100*time.Millisecond {
		t.Errorf("Linear backoff should wait at least 100ms, elapsed: %v", elapsed)
	}
}

func TestDefaultRetryConfig(t *testing.T) {
	config := utils.DefaultRetryConfig()
	if config == nil {
		t.Fatal("DefaultRetryConfig returned nil")
	}

	if config.MaxRetries <= 0 {
		t.Error("Default max retries should be positive")
	}

	if config.InitialDelay <= 0 {
		t.Error("Default initial delay should be positive")
	}

	if config.MaxDelay <= 0 {
		t.Error("Default max delay should be positive")
	}

	if config.Multiplier <= 1.0 {
		t.Error("Default multiplier should be greater than 1.0")
	}
}
