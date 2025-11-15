package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenType defines the type of JWT token.
type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
)

// String returns the string representation of the token type.
func (t TokenType) String() string {
	return string(t)
}

// Claims represents custom JWT claims with user information.
type Claims struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	Username  string    `json:"username,omitempty"`
	Roles     []string  `json:"roles,omitempty"`
	TokenType TokenType `json:"token_type"`

	jwt.RegisteredClaims
}

// NewAccessTokenClaims creates claims for an access token.
func NewAccessTokenClaims(userID, email, username string, roles []string, issuer, audience string, expiresIn time.Duration) *Claims {
	now := time.Now()

	return &Claims{
		UserID:    userID,
		Email:     email,
		Username:  username,
		Roles:     roles,
		TokenType: TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Subject:   userID,
			Audience:  jwt.ClaimStrings{audience},
			ExpiresAt: jwt.NewNumericDate(now.Add(expiresIn)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        generateJTI(),
		},
	}
}

// NewRefreshTokenClaims creates claims for a refresh token.
func NewRefreshTokenClaims(userID, issuer, audience string, expiresIn time.Duration) *Claims {
	now := time.Now()

	return &Claims{
		UserID:    userID,
		TokenType: TokenTypeRefresh,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Subject:   userID,
			Audience:  jwt.ClaimStrings{audience},
			ExpiresAt: jwt.NewNumericDate(now.Add(expiresIn)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        generateJTI(),
		},
	}
}

// IsExpired checks if the token has expired.
func (c *Claims) IsExpired() bool {
	if c.ExpiresAt == nil {
		return false
	}
	return time.Now().After(c.ExpiresAt.Time)
}

// IsNotYetValid checks if the token is not yet valid (before nbf).
func (c *Claims) IsNotYetValid() bool {
	if c.NotBefore == nil {
		return false
	}
	return time.Now().Before(c.NotBefore.Time)
}

// TimeUntilExpiry returns the duration until the token expires.
// Returns 0 if already expired.
func (c *Claims) TimeUntilExpiry() time.Duration {
	if c.ExpiresAt == nil || c.IsExpired() {
		return 0
	}
	return time.Until(c.ExpiresAt.Time)
}

// IsAccessToken checks if this is an access token.
func (c *Claims) IsAccessToken() bool {
	return c.TokenType == TokenTypeAccess
}

// IsRefreshToken checks if this is a refresh token.
func (c *Claims) IsRefreshToken() bool {
	return c.TokenType == TokenTypeRefresh
}

// HasRole checks if the user has a specific role.
func (c *Claims) HasRole(role string) bool {
	for _, r := range c.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// HasAnyRole checks if the user has any of the specified roles.
func (c *Claims) HasAnyRole(roles ...string) bool {
	for _, requiredRole := range roles {
		if c.HasRole(requiredRole) {
			return true
		}
	}
	return false
}

// HasAllRoles checks if the user has all of the specified roles.
func (c *Claims) HasAllRoles(roles ...string) bool {
	for _, requiredRole := range roles {
		if !c.HasRole(requiredRole) {
			return false
		}
	}
	return true
}

// generateJTI generates a unique JWT ID.
func generateJTI() string {
	// Use a simple timestamp-based ID for now
	// In production, you might want to use UUID or similar
	return time.Now().Format("20060102150405.000000")
}
