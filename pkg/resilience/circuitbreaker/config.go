package circuitbreaker

import (
	"errors"
	"time"
)

// Config holds circuit breaker configuration.
type Config struct {
	// MaxFailures is the number of consecutive failures before opening.
	MaxFailures int

	// Timeout is how long to wait in Open state before transitioning to HalfOpen.
	Timeout time.Duration

	// MaxHalfOpenRequests is the number of test requests allowed in HalfOpen state.
	MaxHalfOpenRequests int

	// SuccessThreshold is the number of consecutive successes needed to close from HalfOpen.
	SuccessThreshold int

	// OnStateChange is called when the state changes.
	OnStateChange func(from, to State)
}

// DefaultConfig returns a default circuit breaker configuration.
func DefaultConfig() Config {
	return Config{
		MaxFailures:         5,
		Timeout:             60 * time.Second,
		MaxHalfOpenRequests: 3,
		SuccessThreshold:    2,
	}
}

// Validate checks if the configuration is valid.
func (c Config) Validate() error {
	if c.MaxFailures < 1 {
		return errors.New("max failures must be at least 1")
	}
	if c.Timeout < 1*time.Second {
		return errors.New("timeout must be at least 1 second")
	}
	if c.MaxHalfOpenRequests < 1 {
		return errors.New("max half-open requests must be at least 1")
	}
	if c.SuccessThreshold < 1 {
		return errors.New("success threshold must be at least 1")
	}
	return nil
}
