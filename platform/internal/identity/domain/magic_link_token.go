package domain

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"time"

	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
)

// Magic link constants
const (
	// MagicLinkTokenLength is the length of a magic link token in bytes.
	// 32 bytes = 256 bits of entropy.
	MagicLinkTokenLength = 32

	// MagicLinkDefaultTTL is the default time-to-live for magic links.
	MagicLinkDefaultTTL = 15 * time.Minute
)

// MagicLinkToken represents a single-use token for email-based authentication.
// Contains the token value, associated email, and expiration time.
type MagicLinkToken struct {
	value     []byte
	expiresAt time.Time
}

// NewMagicLinkToken generates a new magic link token with the default TTL.
func NewMagicLinkToken() (MagicLinkToken, error) {
	return NewMagicLinkTokenWithTTL(MagicLinkDefaultTTL)
}

// NewMagicLinkTokenWithTTL generates a new magic link token with a custom TTL.
func NewMagicLinkTokenWithTTL(ttl time.Duration) (MagicLinkToken, error) {
	bytes := make([]byte, MagicLinkTokenLength)
	if _, err := rand.Read(bytes); err != nil {
		return MagicLinkToken{}, pkgerrors.Internal("MagicLinkToken.New", err).
			WithMessage("failed to generate secure random token")
	}

	return MagicLinkToken{
		value:     bytes,
		expiresAt: time.Now().UTC().Add(ttl),
	}, nil
}

// MustNewMagicLinkToken generates a new magic link token and panics on error.
// Only use in contexts where failure is unrecoverable.
func MustNewMagicLinkToken() MagicLinkToken {
	t, err := NewMagicLinkToken()
	if err != nil {
		panic(err)
	}
	return t
}

// ParseMagicLinkToken parses a base64url-encoded string into a MagicLinkToken.
// Note: This creates a token without expiration info; use for verification only.
func ParseMagicLinkToken(s string) (MagicLinkToken, error) {
	if s == "" {
		return MagicLinkToken{}, pkgerrors.Validation("MagicLinkToken.Parse", "token cannot be empty")
	}

	bytes, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return MagicLinkToken{}, pkgerrors.Validation("MagicLinkToken.Parse", "invalid token encoding").
			WithMeta("error", err.Error())
	}

	if len(bytes) != MagicLinkTokenLength {
		return MagicLinkToken{}, pkgerrors.Validation("MagicLinkToken.Parse", "invalid token length").
			WithMeta("expected", MagicLinkTokenLength).
			WithMeta("actual", len(bytes))
	}

	return MagicLinkToken{value: bytes}, nil
}

// String returns the base64url-encoded representation of the token.
// Safe for use in URLs.
func (t MagicLinkToken) String() string {
	if t.IsZero() {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(t.value)
}

// IsZero returns true if the token is the zero value.
func (t MagicLinkToken) IsZero() bool {
	return len(t.value) == 0
}

// IsValid returns true if the token is valid (non-zero, correct length).
func (t MagicLinkToken) IsValid() bool {
	return len(t.value) == MagicLinkTokenLength
}

// IsExpired returns true if the token has expired.
func (t MagicLinkToken) IsExpired() bool {
	if t.expiresAt.IsZero() {
		// If no expiration set (e.g., parsed token), consider it not expired
		// Expiration should be checked against stored value
		return false
	}
	return time.Now().UTC().After(t.expiresAt)
}

// ExpiresAt returns the expiration time of the token.
func (t MagicLinkToken) ExpiresAt() time.Time {
	return t.expiresAt
}

// TTL returns the remaining time-to-live of the token.
// Returns 0 if the token has expired.
func (t MagicLinkToken) TTL() time.Duration {
	if t.expiresAt.IsZero() {
		return 0
	}
	remaining := time.Until(t.expiresAt)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// Equals compares two tokens in constant time to prevent timing attacks.
func (t MagicLinkToken) Equals(other MagicLinkToken) bool {
	if len(t.value) != len(other.value) {
		return false
	}
	return subtle.ConstantTimeCompare(t.value, other.value) == 1
}

// Hash returns a SHA-256 hash of the token for storage.
// The actual token should never be stored; only its hash.
func (t MagicLinkToken) Hash() []byte {
	if t.IsZero() {
		return nil
	}
	hash := sha256.Sum256(t.value)
	return hash[:]
}

// HashString returns the base64url-encoded hash of the token.
func (t MagicLinkToken) HashString() string {
	hash := t.Hash()
	if hash == nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(hash)
}

// VerifyHash verifies that the token matches the given hash.
// Uses constant-time comparison.
func (t MagicLinkToken) VerifyHash(hash []byte) bool {
	if t.IsZero() || len(hash) == 0 {
		return false
	}
	computed := t.Hash()
	return subtle.ConstantTimeCompare(computed, hash) == 1
}

// Bytes returns a copy of the token bytes.
// Returns a copy to prevent mutation of the internal value.
func (t MagicLinkToken) Bytes() []byte {
	if t.IsZero() {
		return nil
	}
	cp := make([]byte, len(t.value))
	copy(cp, t.value)
	return cp
}
