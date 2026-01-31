package domain

import (
	"context"

	"github.com/0xsj/nexus/platform/pkg/types"
)

// ============================================================================
// Wallet Context Port
// ============================================================================

// WalletAddress represents a blockchain wallet address.
type WalletAddress struct {
	Address string
	ChainID string
}

// WalletVerificationRequest contains data for verifying a wallet signature.
type WalletVerificationRequest struct {
	Address   string
	Message   string
	Signature string
	ChainID   string
}

// WalletVerificationResult contains the result of a wallet verification.
type WalletVerificationResult struct {
	Valid   bool
	Address string
	ChainID string
	DID     string // did:pkh derived from wallet
}

// WalletReader provides read access to the Wallet context.
type WalletReader interface {
	// VerifySignature verifies a SIWE signature and returns the derived DID.
	VerifySignature(ctx context.Context, req WalletVerificationRequest) (*WalletVerificationResult, error)

	// GetDIDForWallet returns the did:pkh for a wallet address.
	GetDIDForWallet(ctx context.Context, address WalletAddress) (string, error)
}

// ============================================================================
// DID Generator Port
// ============================================================================

// DIDGeneratorRequest contains data for generating a DID.
type DIDGeneratorRequest struct {
	// Method is the DID method (key, pkh, web).
	Method string

	// For did:key - generated from a new keypair
	// For did:pkh - derived from wallet address
	WalletAddress string
	ChainID       string
}

// DIDGenerator generates DIDs for users.
type DIDGenerator interface {
	// GenerateDIDKey generates a new did:key for custodial users.
	GenerateDIDKey(ctx context.Context) (string, error)

	// DeriveDIDPKH derives a did:pkh from a wallet address.
	DeriveDIDPKH(ctx context.Context, address string, chainID string) (string, error)
}

// ============================================================================
// Email Service Port
// ============================================================================

// EmailRequest contains data for sending an email.
type EmailRequest struct {
	To       types.Email
	Subject  string
	Template string
	Data     map[string]any
}

// EmailService sends emails.
type EmailService interface {
	// SendMagicLink sends a magic link email for authentication.
	SendMagicLink(ctx context.Context, to types.Email, token string, expiresIn int) error

	// SendWelcome sends a welcome email to a new user.
	SendWelcome(ctx context.Context, to types.Email, displayName string) error

	// SendPasswordReset sends a password reset email (if password auth is added later).
	// SendPasswordReset(ctx context.Context, to types.Email, token string, expiresIn int) error
}

// ============================================================================
// Event Publisher Port
// ============================================================================

// DomainEvent represents a domain event for publishing.
type DomainEvent interface {
	EventType() string
	AggregateID() string
	AggregateType() string
}

// EventPublisher publishes domain events to the event bus.
type EventPublisher interface {
	// Publish publishes one or more domain events.
	Publish(ctx context.Context, events ...DomainEvent) error
}

// ============================================================================
// User Lookup Port (for uniqueness checks)
// ============================================================================

// UserLookup provides read access to user data for validation.
// This is used by command handlers to check uniqueness constraints
// without loading full aggregates.
type UserLookup interface {
	// ExistsByEmail checks if a user with the given email exists.
	ExistsByEmail(ctx context.Context, email types.Email) (bool, error)

	// ExistsByDID checks if a user with the given DID exists.
	ExistsByDID(ctx context.Context, did string) (bool, error)

	// ExistsByOAuthSubject checks if a user with the given OAuth subject exists.
	ExistsByOAuthSubject(ctx context.Context, subject OAuthSubject) (bool, error)

	// GetUserIDByEmail returns the user ID for a given email, if it exists.
	GetUserIDByEmail(ctx context.Context, email types.Email) (UserID, error)

	// GetUserIDByDID returns the user ID for a given DID, if it exists.
	GetUserIDByDID(ctx context.Context, did string) (UserID, error)

	// GetUserIDByOAuthSubject returns the user ID for a given OAuth subject, if it exists.
	GetUserIDByOAuthSubject(ctx context.Context, subject OAuthSubject) (UserID, error)
}

// ============================================================================
// Session Lookup Port
// ============================================================================

// SessionLookup provides read access to session data for validation.
type SessionLookup interface {
	// GetSessionByTokenHash retrieves a session by its token hash.
	GetSessionByTokenHash(ctx context.Context, tokenHash string) (*Session, error)

	// GetActiveSessionsForUser retrieves all active sessions for a user.
	GetActiveSessionsForUser(ctx context.Context, userID UserID) ([]*Session, error)

	// CountActiveSessionsForUser counts active sessions for a user.
	CountActiveSessionsForUser(ctx context.Context, userID UserID) (int, error)
}
