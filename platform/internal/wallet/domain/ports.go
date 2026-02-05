package domain

import (
	"context"
	"time"

	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// ============================================================================
// Wallet Lookup Port
// ============================================================================

// WalletLookup provides read-optimized queries for validation.
// Used by command handlers to check uniqueness constraints
// without loading full aggregates.
type WalletLookup interface {
	// ExistsByAddress checks if a wallet with the given address exists.
	ExistsByAddress(ctx context.Context, address WalletAddress) (bool, error)

	// GetWalletIDByAddress returns the wallet ID for a given address, if it exists.
	GetWalletIDByAddress(ctx context.Context, address WalletAddress) (WalletID, error)
}

// ============================================================================
// Challenge Repository Port
// ============================================================================

// ChallengeRepository manages SIWE authentication challenges.
// Challenges are short-lived and not event-sourced.
type ChallengeRepository interface {
	// SaveChallenge stores a challenge for the given address with an expiration time.
	SaveChallenge(ctx context.Context, address WalletAddress, challenge string, expiresAt time.Time) error

	// GetChallenge retrieves the current challenge for the given address.
	GetChallenge(ctx context.Context, address WalletAddress) (string, error)

	// DeleteChallenge removes a challenge after use or expiration.
	DeleteChallenge(ctx context.Context, address WalletAddress) error
}

// ============================================================================
// Signature Verifier Port
// ============================================================================

// SignatureVerifier verifies cryptographic signatures from wallets.
type SignatureVerifier interface {
	// VerifySignature verifies that the signature was produced by the given address
	// for the given message.
	VerifySignature(ctx context.Context, address WalletAddress, message string, signature string) (bool, error)
}

// ============================================================================
// DID Deriver Port
// ============================================================================

// DIDDeriver derives decentralized identifiers from wallet addresses.
type DIDDeriver interface {
	// DeriveDID derives a did:pkh from a wallet address and chain.
	DeriveDID(ctx context.Context, address WalletAddress, chain Chain) (string, error)
}

// ============================================================================
// Event Publisher Port
// ============================================================================

// EventPublisher publishes domain events to the event bus.
type EventPublisher interface {
	// Publish publishes one or more domain events synchronously.
	Publish(ctx context.Context, events ...eventsourcing.Event) error
}

// ============================================================================
// Null Implementations (for testing)
// ============================================================================

// NullWalletLookup is a no-op implementation of WalletLookup.
type NullWalletLookup struct{}

// NewNullWalletLookup creates a new NullWalletLookup.
func NewNullWalletLookup() *NullWalletLookup {
	return &NullWalletLookup{}
}

// ExistsByAddress always returns false.
func (l *NullWalletLookup) ExistsByAddress(ctx context.Context, address WalletAddress) (bool, error) {
	return false, nil
}

// GetWalletIDByAddress always returns a not found error.
func (l *NullWalletLookup) GetWalletIDByAddress(ctx context.Context, address WalletAddress) (WalletID, error) {
	return WalletID{}, WalletNotFound("NullWalletLookup.GetWalletIDByAddress", address.String())
}

// Ensure NullWalletLookup implements WalletLookup.
var _ WalletLookup = (*NullWalletLookup)(nil)

// NullChallengeRepository is a no-op implementation of ChallengeRepository.
type NullChallengeRepository struct{}

// NewNullChallengeRepository creates a new NullChallengeRepository.
func NewNullChallengeRepository() *NullChallengeRepository {
	return &NullChallengeRepository{}
}

// SaveChallenge does nothing and returns nil.
func (r *NullChallengeRepository) SaveChallenge(ctx context.Context, address WalletAddress, challenge string, expiresAt time.Time) error {
	return nil
}

// GetChallenge always returns a not found error.
func (r *NullChallengeRepository) GetChallenge(ctx context.Context, address WalletAddress) (string, error) {
	return "", ChallengeNotFound("NullChallengeRepository.GetChallenge", address.String())
}

// DeleteChallenge does nothing and returns nil.
func (r *NullChallengeRepository) DeleteChallenge(ctx context.Context, address WalletAddress) error {
	return nil
}

// Ensure NullChallengeRepository implements ChallengeRepository.
var _ ChallengeRepository = (*NullChallengeRepository)(nil)

// NullSignatureVerifier is a no-op implementation of SignatureVerifier.
type NullSignatureVerifier struct{}

// NewNullSignatureVerifier creates a new NullSignatureVerifier.
func NewNullSignatureVerifier() *NullSignatureVerifier {
	return &NullSignatureVerifier{}
}

// VerifySignature always returns false.
func (v *NullSignatureVerifier) VerifySignature(ctx context.Context, address WalletAddress, message string, signature string) (bool, error) {
	return false, nil
}

// Ensure NullSignatureVerifier implements SignatureVerifier.
var _ SignatureVerifier = (*NullSignatureVerifier)(nil)

// NullDIDDeriver is a no-op implementation of DIDDeriver.
type NullDIDDeriver struct{}

// NewNullDIDDeriver creates a new NullDIDDeriver.
func NewNullDIDDeriver() *NullDIDDeriver {
	return &NullDIDDeriver{}
}

// DeriveDID always returns an empty string.
func (d *NullDIDDeriver) DeriveDID(ctx context.Context, address WalletAddress, chain Chain) (string, error) {
	return "", nil
}

// Ensure NullDIDDeriver implements DIDDeriver.
var _ DIDDeriver = (*NullDIDDeriver)(nil)

// NullEventPublisher is a no-op implementation of EventPublisher.
type NullEventPublisher struct{}

// NewNullEventPublisher creates a new NullEventPublisher.
func NewNullEventPublisher() *NullEventPublisher {
	return &NullEventPublisher{}
}

// Publish does nothing and returns nil.
func (p *NullEventPublisher) Publish(ctx context.Context, events ...eventsourcing.Event) error {
	return nil
}

// Ensure NullEventPublisher implements EventPublisher.
var _ EventPublisher = (*NullEventPublisher)(nil)
