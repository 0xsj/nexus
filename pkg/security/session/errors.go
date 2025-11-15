package session

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrSessionNotFound = errors.New("session not found")
	ErrSessionExpired  = errors.New("session expired")
	ErrInvalidUserID   = errors.New("invalid user ID")
	ErrInvalidToken    = errors.New("invalid refresh token")
	ErrEmptySessionID  = errors.New("session ID cannot be empty")
	ErrNilStore        = errors.New("store cannot be nil")
)

// ErrInvalidTTL is returned when TTL is invalid.
type ErrInvalidTTL struct {
	TTL time.Duration
}

type ErrInvalidMaxSessions struct {
	Max int
}

func (e ErrInvalidTTL) Error() string {
	return fmt.Sprintf("invalid session TTL: %v (must be > 0)", e.TTL)
}

func (e ErrInvalidMaxSessions) Error() string {
	return fmt.Sprintf("invalid max sessions per user: %d (must be >= 0)", e.Max)
}
