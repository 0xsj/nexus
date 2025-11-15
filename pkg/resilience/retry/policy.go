package retry

import (
	"errors"
	"time"
)

// Policy defines retry behavior.
type Policy struct {
	MaxAttempts     int
	Backoff         Backoff
	RetryableErrors []error
	OnRetry         func(attempt int, err error, delay time.Duration)
}

// Config holds retry policy configuration.
type Config struct {
	MaxAttempts     int
	BaseDelay       time.Duration
	MaxDelay        time.Duration
	BackoffType     BackoffType
	RetryableErrors []error
}

// BackoffType defines the backoff strategy type.
type BackoffType string

const (
	BackoffConstant    BackoffType = "constant"
	BackoffExponential BackoffType = "exponential"
	BackoffLinear      BackoffType = "linear"
)

// DefaultConfig returns a default retry configuration.
func DefaultConfig() Config {
	return Config{
		MaxAttempts:     3,
		BaseDelay:       100 * time.Millisecond,
		MaxDelay:        10 * time.Second,
		BackoffType:     BackoffExponential,
		RetryableErrors: nil, // Retry all errors
	}
}

// NewPolicy creates a new retry policy from config.
func NewPolicy(cfg Config) *Policy {
	var backoff Backoff

	switch cfg.BackoffType {
	case BackoffConstant:
		backoff = NewConstantBackoff(cfg.BaseDelay)
	case BackoffLinear:
		backoff = NewLinearBackoff(cfg.BaseDelay, cfg.MaxDelay, 100*time.Millisecond)
	case BackoffExponential:
		fallthrough
	default:
		backoff = NewExponentialBackoff(cfg.BaseDelay, cfg.MaxDelay)
	}

	return &Policy{
		MaxAttempts:     cfg.MaxAttempts,
		Backoff:         backoff,
		RetryableErrors: cfg.RetryableErrors,
	}
}

// WithOnRetry sets the retry callback.
func (p *Policy) WithOnRetry(fn func(attempt int, err error, delay time.Duration)) *Policy {
	p.OnRetry = fn
	return p
}

// isRetryable checks if an error is retryable.
func (p *Policy) isRetryable(err error) bool {
	// If no specific errors configured, retry all
	if len(p.RetryableErrors) == 0 {
		return true
	}

	// Check if error matches any retryable error
	for _, retryableErr := range p.RetryableErrors {
		if errors.Is(err, retryableErr) {
			return true
		}
	}

	return false
}
