package retry

import (
	"context"
	"fmt"
	"time"
)

// Do executes the function with retry logic.
func Do(ctx context.Context, policy *Policy, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt < policy.MaxAttempts; attempt++ {
		// Execute the function
		err := fn()

		// Success!
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if we should retry
		if !policy.isRetryable(err) {
			return fmt.Errorf("non-retryable error: %w", err)
		}

		// Last attempt - don't wait
		if attempt == policy.MaxAttempts-1 {
			break
		}

		// Calculate backoff
		delay := policy.Backoff.Next(attempt)

		// Call retry callback if set
		if policy.OnRetry != nil {
			policy.OnRetry(attempt+1, err, delay)
		}

		// Wait with context cancellation support
		select {
		case <-time.After(delay):
			// Continue to next attempt
		case <-ctx.Done():
			return fmt.Errorf("retry cancelled: %w", ctx.Err())
		}
	}

	return fmt.Errorf("max retry attempts (%d) exceeded: %w", policy.MaxAttempts, lastErr)
}

// DoWithData executes a function returning data with retry logic.
func DoWithData[T any](ctx context.Context, policy *Policy, fn func() (T, error)) (T, error) {
	var result T
	var lastErr error

	for attempt := 0; attempt < policy.MaxAttempts; attempt++ {
		// Execute the function
		res, err := fn()

		// Success!
		if err == nil {
			return res, nil
		}

		lastErr = err

		// Check if we should retry
		if !policy.isRetryable(err) {
			return result, fmt.Errorf("non-retryable error: %w", err)
		}

		// Last attempt - don't wait
		if attempt == policy.MaxAttempts-1 {
			break
		}

		// Calculate backoff
		delay := policy.Backoff.Next(attempt)

		// Call retry callback if set
		if policy.OnRetry != nil {
			policy.OnRetry(attempt+1, err, delay)
		}

		// Wait with context cancellation support
		select {
		case <-time.After(delay):
			// Continue to next attempt
		case <-ctx.Done():
			return result, fmt.Errorf("retry cancelled: %w", ctx.Err())
		}
	}

	return result, fmt.Errorf("max retry attempts (%d) exceeded: %w", policy.MaxAttempts, lastErr)
}
