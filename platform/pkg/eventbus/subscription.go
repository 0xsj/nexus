package eventbus

import "time"

// SubscribeOption configures subscription behavior.
type SubscribeOption func(*SubscribeOptions)

// SubscribeOptions holds the resolved subscription configuration.
type SubscribeOptions struct {
	// QueueGroup enables competing consumer mode.
	// Only one subscriber in the group receives each event.
	QueueGroup string

	// BufferSize is the channel buffer size for async delivery.
	BufferSize int

	// MaxRetries is the maximum number of retry attempts for failed handlers.
	MaxRetries int

	// RetryDelay is the base delay between retry attempts.
	RetryDelay time.Duration

	// AckTimeout is the deadline for acknowledging an event.
	AckTimeout time.Duration
}

// DefaultSubscribeOptions returns sensible defaults.
func DefaultSubscribeOptions() SubscribeOptions {
	return SubscribeOptions{
		BufferSize: 64,
		MaxRetries: 0,
		RetryDelay: time.Second,
		AckTimeout: 30 * time.Second,
	}
}

// ApplyOptions resolves a set of SubscribeOption functions into SubscribeOptions.
func ApplyOptions(opts ...SubscribeOption) SubscribeOptions {
	o := DefaultSubscribeOptions()
	for _, fn := range opts {
		fn(&o)
	}
	return o
}

// WithQueueGroup sets the queue group for competing consumers.
func WithQueueGroup(name string) SubscribeOption {
	return func(o *SubscribeOptions) {
		o.QueueGroup = name
	}
}

// WithBufferSize sets the channel buffer size.
func WithBufferSize(n int) SubscribeOption {
	return func(o *SubscribeOptions) {
		o.BufferSize = n
	}
}

// WithMaxRetries sets the maximum retry attempts.
func WithMaxRetries(n int) SubscribeOption {
	return func(o *SubscribeOptions) {
		o.MaxRetries = n
	}
}

// WithRetryDelay sets the base delay between retries.
func WithRetryDelay(d time.Duration) SubscribeOption {
	return func(o *SubscribeOptions) {
		o.RetryDelay = d
	}
}

// WithAckTimeout sets the acknowledgment deadline.
func WithAckTimeout(d time.Duration) SubscribeOption {
	return func(o *SubscribeOptions) {
		o.AckTimeout = d
	}
}
