package retry

import (
	"math"
	"math/rand"
	"time"
)

// Backoff defines a backoff strategy.
type Backoff interface {
	// Next returns the next backoff duration.
	Next(attempt int) time.Duration

	// Reset resets the backoff state.
	Reset()
}

// ConstantBackoff implements a constant backoff strategy.
type ConstantBackoff struct {
	Duration time.Duration
}

// NewConstantBackoff creates a constant backoff strategy.
func NewConstantBackoff(duration time.Duration) *ConstantBackoff {
	return &ConstantBackoff{Duration: duration}
}

// Next returns the constant duration.
func (b *ConstantBackoff) Next(attempt int) time.Duration {
	return b.Duration
}

// Reset does nothing for constant backoff.
func (b *ConstantBackoff) Reset() {}

// ExponentialBackoff implements exponential backoff with jitter.
type ExponentialBackoff struct {
	BaseDelay  time.Duration
	MaxDelay   time.Duration
	Multiplier float64
	Jitter     float64 // 0.0 to 1.0
}

// NewExponentialBackoff creates an exponential backoff strategy.
func NewExponentialBackoff(baseDelay, maxDelay time.Duration) *ExponentialBackoff {
	return &ExponentialBackoff{
		BaseDelay:  baseDelay,
		MaxDelay:   maxDelay,
		Multiplier: 2.0,
		Jitter:     0.1, // 10% jitter
	}
}

// Next calculates the next backoff duration with exponential growth and jitter.
func (b *ExponentialBackoff) Next(attempt int) time.Duration {
	if attempt <= 0 {
		return b.BaseDelay
	}

	// Calculate exponential delay: baseDelay * (multiplier ^ attempt)
	delay := float64(b.BaseDelay) * math.Pow(b.Multiplier, float64(attempt))

	// Cap at max delay
	if delay > float64(b.MaxDelay) {
		delay = float64(b.MaxDelay)
	}

	// Add jitter: delay ± (delay * jitter * random)
	jitter := delay * b.Jitter * (rand.Float64()*2 - 1)
	delay += jitter

	// Ensure positive
	if delay < 0 {
		delay = float64(b.BaseDelay)
	}

	return time.Duration(delay)
}

// Reset does nothing for exponential backoff.
func (b *ExponentialBackoff) Reset() {}

// LinearBackoff implements linear backoff.
type LinearBackoff struct {
	BaseDelay time.Duration
	MaxDelay  time.Duration
	Increment time.Duration
}

// NewLinearBackoff creates a linear backoff strategy.
func NewLinearBackoff(baseDelay, maxDelay, increment time.Duration) *LinearBackoff {
	return &LinearBackoff{
		BaseDelay: baseDelay,
		MaxDelay:  maxDelay,
		Increment: increment,
	}
}

// Next calculates linear backoff: baseDelay + (increment * attempt).
func (b *LinearBackoff) Next(attempt int) time.Duration {
	delay := b.BaseDelay + (b.Increment * time.Duration(attempt))
	if delay > b.MaxDelay {
		delay = b.MaxDelay
	}
	return delay
}

// Reset does nothing for linear backoff.
func (b *LinearBackoff) Reset() {}
