package idempotency

import (
	"context"
	"time"
)

// Store defines the interface for idempotency key storage.
type Store interface {
	// Exists checks if an idempotency key has been processed.
	Exists(ctx context.Context, key string) (bool, error)

	// Save stores an idempotency key with optional result.
	Save(ctx context.Context, key string, result interface{}, ttl time.Duration) error

	// Get retrieves the stored result for an idempotency key.
	Get(ctx context.Context, key string) (interface{}, error)

	// Delete removes an idempotency key.
	Delete(ctx context.Context, key string) error
}

// Key represents an idempotency key.
type Key string

// NewKey creates a new idempotency key.
func NewKey(value string) Key {
	return Key(value)
}

// String returns the string representation of the key.
func (k Key) String() string {
	return string(k)
}

// IsEmpty checks if the key is empty.
func (k Key) IsEmpty() bool {
	return string(k) == ""
}
