package magic

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrInvalidUserID  = errors.New("invalid user ID")
	ErrInvalidEmail   = errors.New("invalid email")
	ErrInvalidBaseURL = errors.New("invalid base URL")
	ErrTokenNotFound  = errors.New("token not found")
)

// ErrInvalidTTL is returned when TTL is invalid.
type ErrInvalidTTL struct {
	TTL time.Duration
}

func (e ErrInvalidTTL) Error() string {
	return fmt.Sprintf("invalid TTL: %v (must be > 0)", e.TTL)
}

// ErrTokenGeneration is returned when token generation fails.
type ErrTokenGeneration struct {
	Err error
}

func (e ErrTokenGeneration) Error() string {
	return fmt.Sprintf("token generation failed: %v", e.Err)
}

func (e ErrTokenGeneration) Unwrap() error {
	return e.Err
}

// ErrInvalidPurpose is returned when purpose is invalid.
type ErrInvalidPurpose struct {
	Purpose string
}

func (e ErrInvalidPurpose) Error() string {
	return fmt.Sprintf("invalid token purpose: %s", e.Purpose)
}

// ErrTokenExpired is returned when token has expired.
type ErrTokenExpired struct {
	TokenID   string
	ExpiresAt time.Time
}

func (e ErrTokenExpired) Error() string {
	return fmt.Sprintf("token %s expired at %v", e.TokenID, e.ExpiresAt)
}

// ErrTokenAlreadyUsed is returned when token has been used.
type ErrTokenAlreadyUsed struct {
	TokenID string
	UsedAt  time.Time
}

func (e ErrTokenAlreadyUsed) Error() string {
	return fmt.Sprintf("token %s already used at %v", e.TokenID, e.UsedAt)
}

// ErrTooManyTokens is returned when rate limit is exceeded.
type ErrTooManyTokens struct {
	UserID  string
	Purpose Purpose
	Count   int
	Limit   int
}

func (e ErrTooManyTokens) Error() string {
	return fmt.Sprintf("too many %s tokens for user %s: %d/%d", e.Purpose, e.UserID, e.Count, e.Limit)
}

// ErrInvalidLength is returned when token length is invalid.
type ErrInvalidLength struct {
	Length int
}

func (e ErrInvalidLength) Error() string {
	return fmt.Sprintf("invalid token length: %d (must be > 0)", e.Length)
}

// ErrInvalidMaxTokens is returned when max tokens per user is invalid.
type ErrInvalidMaxTokens struct {
	Max int
}

func (e ErrInvalidMaxTokens) Error() string {
	return fmt.Sprintf("invalid max tokens per user: %d (must be > 0)", e.Max)
}
