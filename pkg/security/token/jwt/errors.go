package jwt

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrInvalidToken      = errors.New("invalid token")
	ErrTokenExpired      = errors.New("token expired")
	ErrInvalidSigningKey = errors.New("invalid signing key")
	ErrInvalidAlgorithm  = errors.New("invalid algorithm")
	ErrMissingClaims     = errors.New("missing required claims")
	ErrMissingKeyPaths   = errors.New("missing private or public key paths for asymmetric algorithm")
	ErrMissingIssuer     = errors.New("missing issuer")
	ErrMissingAudience   = errors.New("missing audience")
)

type ErrWeakSecretKey struct {
	Length int
}

// ErrInvalidTokenType is returned when token type doesn't match expected.
type ErrInvalidTokenType struct {
	Expected string
	Actual   string
}

func (e ErrInvalidTokenType) Error() string {
	return fmt.Sprintf("invalid token type: expected %s, got %s", e.Expected, e.Actual)
}

// ErrTokenNotYetValid is returned when token is used before its nbf (not before) time.
type ErrTokenNotYetValid struct {
	NotBefore time.Time
}

func (e ErrTokenNotYetValid) Error() string {
	return fmt.Sprintf("token not yet valid: not before %v", e.NotBefore)
}

// ErrInvalidAudience is returned when token audience doesn't match.
type ErrInvalidAudience struct {
	Expected []string
	Actual   []string
}

func (e ErrInvalidAudience) Error() string {
	return fmt.Sprintf("invalid audience: expected %v, got %v", e.Expected, e.Actual)
}

// ErrInvalidIssuer is returned when token issuer doesn't match.
type ErrInvalidIssuer struct {
	Expected string
	Actual   string
}

func (e ErrInvalidIssuer) Error() string {
	return fmt.Sprintf("invalid issuer: expected %s, got %s", e.Expected, e.Actual)
}

func (e ErrWeakSecretKey) Error() string {
	return fmt.Sprintf("weak secret key: length %d (minimum 32 bytes required)", e.Length)
}

// ErrInvalidTTL is returned when TTL is invalid.
type ErrInvalidTTL struct {
	Name string
	TTL  time.Duration
}

func (e ErrInvalidTTL) Error() string {
	return fmt.Sprintf("invalid TTL for %s: %v (must be > 0)", e.Name, e.TTL)
}

// ErrRefreshTTLTooShort is returned when refresh token TTL is shorter than access token TTL.
type ErrRefreshTTLTooShort struct {
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

func (e ErrRefreshTTLTooShort) Error() string {
	return fmt.Sprintf("refresh token TTL (%v) must be longer than access token TTL (%v)", e.RefreshTTL, e.AccessTTL)
}
