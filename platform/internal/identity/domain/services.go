package domain

import (
	"context"
	"time"
)

// ============================================================================
// Signature Verifier
// ============================================================================

// SignatureVerifier verifies cryptographic signatures.
// Implementation handles different signature types (SIWE, Ed25519, etc.).
type SignatureVerifier interface {
	// VerifyWalletSignature verifies a wallet signature (SIWE/EIP-4361).
	VerifyWalletSignature(ctx context.Context, params WalletSignatureParams) error

	// VerifyDIDSignature verifies a DID-based signature.
	VerifyDIDSignature(ctx context.Context, params DIDSignatureParams) error
}

// WalletSignatureParams contains parameters for wallet signature verification.
type WalletSignatureParams struct {
	Address   string
	Chain     Chain
	Message   string
	Signature string
	Nonce     string
}

// DIDSignatureParams contains parameters for DID signature verification.
type DIDSignatureParams struct {
	DID       string
	Message   string
	Signature string
	KeyID     string // Optional: specific key from DID document
}

// ============================================================================
// Token Service
// ============================================================================

// TokenService handles JWT token generation and validation.
type TokenService interface {
	// GenerateAccessToken generates a short-lived access token.
	GenerateAccessToken(ctx context.Context, params AccessTokenParams) (string, error)

	// GenerateRefreshToken generates a long-lived refresh token.
	GenerateRefreshToken(ctx context.Context, params RefreshTokenParams) (string, error)

	// ValidateAccessToken validates an access token and returns claims.
	ValidateAccessToken(ctx context.Context, token string) (*TokenClaims, error)

	// ValidateRefreshToken validates a refresh token and returns claims.
	ValidateRefreshToken(ctx context.Context, token string) (*TokenClaims, error)

	// RevokeToken revokes a token (adds to blacklist).
	RevokeToken(ctx context.Context, token string) error

	// IsRevoked checks if a token is revoked.
	IsRevoked(ctx context.Context, tokenID string) (bool, error)
}

// AccessTokenParams contains parameters for access token generation.
type AccessTokenParams struct {
	UserID    string
	DID       string
	SessionID string
	Scopes    []string
	ExpiresIn time.Duration
}

// RefreshTokenParams contains parameters for refresh token generation.
type RefreshTokenParams struct {
	UserID    string
	SessionID string
	ExpiresIn time.Duration
}

// TokenClaims contains validated token claims.
type TokenClaims struct {
	TokenID   string
	UserID    string
	DID       string
	SessionID string
	Scopes    []string
	IssuedAt  time.Time
	ExpiresAt time.Time
	TokenType TokenType
}

// IsExpired returns true if the token has expired.
func (c *TokenClaims) IsExpired() bool {
	return time.Now().After(c.ExpiresAt)
}

// HasScope checks if the token has a specific scope.
func (c *TokenClaims) HasScope(scope string) bool {
	for _, s := range c.Scopes {
		if s == scope {
			return true
		}
	}
	return false
}

// ============================================================================
// Challenge Service
// ============================================================================

// ChallengeService manages authentication challenges.
type ChallengeService interface {
	// CreateChallenge creates a new authentication challenge.
	CreateChallenge(ctx context.Context, params CreateChallengeParams) (*Challenge, error)

	// ValidateChallenge validates and consumes a challenge.
	// Returns error if challenge is invalid, expired, or already used.
	ValidateChallenge(ctx context.Context, nonce string) (*Challenge, error)

	// InvalidateChallenge invalidates a challenge without validation.
	InvalidateChallenge(ctx context.Context, nonce string) error
}

// CreateChallengeParams contains parameters for challenge creation.
type CreateChallengeParams struct {
	Address string
	Chain   Chain
	Domain  string
	URI     string
	TTL     time.Duration
}

// ============================================================================
// DID Service
// ============================================================================

// DIDService handles DID operations.
type DIDService interface {
	// Generate generates a new DID.
	Generate(ctx context.Context, method DIDMethod) (string, error)

	// Resolve resolves a DID to its document.
	Resolve(ctx context.Context, did string) (*DIDDocument, error)

	// DeriveFromWallet derives a did:pkh from a wallet address.
	DeriveFromWallet(address string, chain Chain) string
}

// DIDMethod represents supported DID methods.
type DIDMethod string

const (
	DIDMethodKey DIDMethod = "key"
	DIDMethodPKH DIDMethod = "pkh"
	DIDMethodWeb DIDMethod = "web"
)

// String returns the string representation.
func (m DIDMethod) String() string {
	return string(m)
}

// DIDDocument represents a resolved DID document (simplified).
type DIDDocument struct {
	ID                   string
	Controller           string
	VerificationMethods  []VerificationMethod
	AuthenticationKeyIDs []string
}

// VerificationMethod represents a verification method in a DID document.
type VerificationMethod struct {
	ID              string
	Type            string
	Controller      string
	PublicKeyBase58 string
	PublicKeyHex    string
	PublicKeyJWK    map[string]any
}

// ============================================================================
// Password Service (Optional - for Web2 bridge)
// ============================================================================

// PasswordService handles password hashing and verification.
// Optional — only needed if supporting password-based auth as a bridge.
type PasswordService interface {
	// Hash hashes a password.
	Hash(ctx context.Context, password string) (string, error)

	// Verify verifies a password against a hash.
	Verify(ctx context.Context, password, hash string) error

	// NeedsRehash checks if a hash needs to be rehashed (algorithm upgrade).
	NeedsRehash(hash string) bool
}

// ============================================================================
// OAuth Service
// ============================================================================

// OAuthService handles OAuth provider interactions.
// Implementation details are in infrastructure layer.
type OAuthService interface {
	// GetAuthURL returns the OAuth authorization URL.
	GetAuthURL(ctx context.Context, provider OAuthProvider, state string, scopes []string) (string, error)

	// ExchangeCode exchanges an authorization code for tokens.
	ExchangeCode(ctx context.Context, provider OAuthProvider, code string) (*OAuthTokens, error)

	// RefreshTokens refreshes OAuth tokens.
	RefreshTokens(ctx context.Context, provider OAuthProvider, refreshToken string) (*OAuthTokens, error)

	// FetchProfile fetches the user profile from the provider.
	FetchProfile(ctx context.Context, provider OAuthProvider, accessToken string) (*ConnectionProfile, error)

	// RevokeAccess revokes access to the provider.
	RevokeAccess(ctx context.Context, provider OAuthProvider, accessToken string) error
}

// OAuthTokens contains OAuth tokens.
type OAuthTokens struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	ExpiresAt    time.Time
	Scopes       []string
}

// IsExpired returns true if the access token has expired.
func (t *OAuthTokens) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// NeedsRefresh returns true if the token should be refreshed soon.
func (t *OAuthTokens) NeedsRefresh() bool {
	return time.Until(t.ExpiresAt) < 5*time.Minute
}

// ============================================================================
// Notification Service
// ============================================================================

// IdentityNotificationService handles identity-related notifications.
type IdentityNotificationService interface {
	// SendEmailVerification sends an email verification link.
	SendEmailVerification(ctx context.Context, userID, email, token string) error

	// SendLoginAlert sends a login alert notification.
	SendLoginAlert(ctx context.Context, userID, method, ipAddress, userAgent string) error

	// SendNewDeviceAlert sends a new device login alert.
	SendNewDeviceAlert(ctx context.Context, userID, device, ipAddress string) error

	// SendAPIKeyCreated sends an API key created notification.
	SendAPIKeyCreated(ctx context.Context, userID, keyName string) error

	// SendSecurityAlert sends a security alert.
	SendSecurityAlert(ctx context.Context, userID, alertType, message string) error
}

// ============================================================================
// Rate Limiter
// ============================================================================

// RateLimiter handles rate limiting for auth operations.
type RateLimiter interface {
	// Allow checks if an operation is allowed.
	Allow(ctx context.Context, key string, limit RateLimit) (bool, error)

	// Reset resets the rate limit for a key.
	Reset(ctx context.Context, key string) error

	// Remaining returns the remaining attempts.
	Remaining(ctx context.Context, key string, limit RateLimit) (int, error)
}

// RateLimit defines rate limit parameters.
type RateLimit struct {
	Requests int
	Window   time.Duration
}

// Common rate limits.
var (
	RateLimitLogin       = RateLimit{Requests: 5, Window: 15 * time.Minute}
	RateLimitRegister    = RateLimit{Requests: 3, Window: time.Hour}
	RateLimitEmailVerify = RateLimit{Requests: 5, Window: time.Hour}
	RateLimitAPIKey      = RateLimit{Requests: 10, Window: time.Hour}
	RateLimitChallenge   = RateLimit{Requests: 10, Window: time.Minute}
)

// ============================================================================
// Audit Logger
// ============================================================================

// AuditLogger logs security-relevant events.
type AuditLogger interface {
	// LogAuthAttempt logs an authentication attempt.
	LogAuthAttempt(ctx context.Context, event AuthAttemptEvent)

	// LogAuthSuccess logs a successful authentication.
	LogAuthSuccess(ctx context.Context, event AuthSuccessEvent)

	// LogAuthFailure logs a failed authentication.
	LogAuthFailure(ctx context.Context, event AuthFailureEvent)

	// LogSecurityEvent logs a security event.
	LogSecurityEvent(ctx context.Context, event SecurityEvent)
}

// AuthAttemptEvent represents an authentication attempt.
type AuthAttemptEvent struct {
	Timestamp time.Time
	Method    AuthMethod
	Identity  string // DID, wallet address, or email
	IPAddress string
	UserAgent string
}

// AuthSuccessEvent represents a successful authentication.
type AuthSuccessEvent struct {
	AuthAttemptEvent
	UserID    string
	SessionID string
}

// AuthFailureEvent represents a failed authentication.
type AuthFailureEvent struct {
	AuthAttemptEvent
	Reason string
}

// SecurityEvent represents a security-relevant event.
type SecurityEvent struct {
	Timestamp time.Time
	UserID    string
	EventType string
	Details   map[string]any
	IPAddress string
	UserAgent string
}

// Security event types.
const (
	SecurityEventSessionRevoked     = "session.revoked"
	SecurityEventAllSessionsRevoked = "sessions.all_revoked"
	SecurityEventAPIKeyCreated      = "apikey.created"
	SecurityEventAPIKeyRevoked      = "apikey.revoked"
	SecurityEventWalletLinked       = "wallet.linked"
	SecurityEventWalletUnlinked     = "wallet.unlinked"
	SecurityEventOAuthLinked        = "oauth.linked"
	SecurityEventOAuthUnlinked      = "oauth.unlinked"
	SecurityEventUserSuspended      = "user.suspended"
	SecurityEventUserActivated      = "user.activated"
	SecurityEventSuspiciousActivity = "suspicious.activity"
)
