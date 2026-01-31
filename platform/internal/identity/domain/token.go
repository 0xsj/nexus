package domain

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"

	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
)

// Token byte lengths
const (
	// TokenLength is the length of a session token in bytes.
	// 32 bytes = 256 bits of entropy.
	TokenLength = 32

	// TokenEncodedLength is the length of a base64url-encoded token.
	// 32 bytes encodes to 43 characters in base64url (no padding).
	TokenEncodedLength = 43
)

// Token represents an opaque session token.
// Used for session authentication via cookies or Authorization header.
// Stored as a secure random value, compared in constant time.
type Token struct {
	value []byte
}

// NewToken generates a new cryptographically secure random token.
func NewToken() (Token, error) {
	bytes := make([]byte, TokenLength)
	if _, err := rand.Read(bytes); err != nil {
		return Token{}, pkgerrors.Internal("Token.New", err).
			WithMessage("failed to generate secure random token")
	}
	return Token{value: bytes}, nil
}

// MustNewToken generates a new token and panics on error.
// Only use in contexts where failure is unrecoverable.
func MustNewToken() Token {
	t, err := NewToken()
	if err != nil {
		panic(err)
	}
	return t
}

// ParseToken parses a base64url-encoded string into a Token.
func ParseToken(s string) (Token, error) {
	if s == "" {
		return Token{}, pkgerrors.Validation("Token.Parse", "token cannot be empty")
	}

	bytes, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return Token{}, pkgerrors.Validation("Token.Parse", "invalid token encoding").
			WithMeta("error", err.Error())
	}

	if len(bytes) != TokenLength {
		return Token{}, pkgerrors.Validation("Token.Parse", "invalid token length").
			WithMeta("expected", TokenLength).
			WithMeta("actual", len(bytes))
	}

	return Token{value: bytes}, nil
}

// String returns the base64url-encoded representation of the token.
// Safe for use in cookies, headers, and URLs.
func (t Token) String() string {
	if t.IsZero() {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(t.value)
}

// IsZero returns true if the token is the zero value.
func (t Token) IsZero() bool {
	return len(t.value) == 0
}

// IsValid returns true if the token is valid (non-zero, correct length).
func (t Token) IsValid() bool {
	return len(t.value) == TokenLength
}

// Equals compares two tokens in constant time to prevent timing attacks.
func (t Token) Equals(other Token) bool {
	if len(t.value) != len(other.value) {
		return false
	}
	return subtle.ConstantTimeCompare(t.value, other.value) == 1
}

// Bytes returns a copy of the token bytes.
// Returns a copy to prevent mutation of the internal value.
func (t Token) Bytes() []byte {
	if t.IsZero() {
		return nil
	}
	cp := make([]byte, len(t.value))
	copy(cp, t.value)
	return cp
}

// Hash returns a SHA-256 hash of the token for storage.
// The actual token should never be stored; only its hash.
func (t Token) Hash() []byte {
	if t.IsZero() {
		return nil
	}
	hash := sha256.Sum256(t.value)
	return hash[:]
}

// HashString returns the base64url-encoded hash of the token.
func (t Token) HashString() string {
	hash := t.Hash()
	if hash == nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(hash)
}

// VerifyHash verifies that the token matches the given hash.
// Uses constant-time comparison.
func (t Token) VerifyHash(hash []byte) bool {
	if t.IsZero() || len(hash) == 0 {
		return false
	}
	computed := t.Hash()
	return subtle.ConstantTimeCompare(computed, hash) == 1
}
