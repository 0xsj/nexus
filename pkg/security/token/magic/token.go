package magic

import (
	"time"
)

// Purpose defines what the magic link token is for.
type Purpose string

const (
	PurposeLogin         Purpose = "login"          // Passwordless login
	PurposeEmailVerify   Purpose = "email_verify"   // Email verification
	PurposePasswordReset Purpose = "password_reset" // Password reset
)

// String returns the string representation of the purpose.
func (p Purpose) String() string {
	return string(p)
}

// IsValid checks if the purpose is a valid value.
func (p Purpose) IsValid() bool {
	switch p {
	case PurposeLogin, PurposeEmailVerify, PurposePasswordReset:
		return true
	default:
		return false
	}
}

// Token represents a magic link token.
type Token struct {
	ID        string     // UUID
	Value     string     // Actual token value (cryptographically secure)
	UserID    string     // User this token belongs to
	Purpose   Purpose    // What the token is for
	Email     string     // Email address (for verification purposes)
	ExpiresAt time.Time  // When the token expires
	UsedAt    *time.Time // When the token was used (nil if not used)
	CreatedAt time.Time  // When the token was created
}

// IsExpired checks if the token has expired.
func (t *Token) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// IsUsed checks if the token has been used.
func (t *Token) IsUsed() bool {
	return t.UsedAt != nil
}

// IsValid checks if the token is valid (not expired, not used).
func (t *Token) IsValid() bool {
	return !t.IsExpired() && !t.IsUsed()
}

// MarkAsUsed marks the token as used.
func (t *Token) MarkAsUsed() {
	now := time.Now()
	t.UsedAt = &now
}

// TimeUntilExpiry returns the duration until the token expires.
// Returns 0 if already expired.
func (t *Token) TimeUntilExpiry() time.Duration {
	if t.IsExpired() {
		return 0
	}
	return time.Until(t.ExpiresAt)
}
