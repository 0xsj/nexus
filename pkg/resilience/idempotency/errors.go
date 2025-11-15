package idempotency

import "errors"

var (
	// ErrKeyAlreadyProcessed indicates the idempotency key has already been processed.
	ErrKeyAlreadyProcessed = errors.New("idempotency key already processed")

	// ErrKeyNotFound indicates the idempotency key was not found.
	ErrKeyNotFound = errors.New("idempotency key not found")

	// ErrInvalidKey indicates the idempotency key is invalid.
	ErrInvalidKey = errors.New("invalid idempotency key")

	// ErrStoreUnavailable indicates the underlying store is unavailable.
	ErrStoreUnavailable = errors.New("idempotency store unavailable")
)
