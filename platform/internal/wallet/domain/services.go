package domain

import (
	"context"
	"time"

	"github.com/0xsj/nexus/platform/pkg/did"
)

// ============================================================================
// Address Validation Service
// ============================================================================

// AddressValidationService validates wallet addresses.
type AddressValidationService interface {
	// Validate validates an address for a specific chain.
	Validate(ctx context.Context, address string, chainID ChainID) (Address, error)

	// ValidateForFamily validates an address for a chain family.
	ValidateForFamily(ctx context.Context, address string, family ChainFamily) (Address, error)

	// IsValid returns true if the address is valid for the chain.
	IsValid(ctx context.Context, address string, chainID ChainID) bool

	// Normalize returns the normalized form of an address.
	Normalize(ctx context.Context, address string, family ChainFamily) string
}

// ============================================================================
// Signature Verification Service
// ============================================================================

// SignatureVerificationService verifies wallet signatures.
type SignatureVerificationService interface {
	// VerifySignature verifies a signature against a message and address.
	VerifySignature(ctx context.Context, params VerifySignatureParams) error

	// VerifySIWE verifies a Sign-In With Ethereum (EIP-4361) message.
	VerifySIWE(ctx context.Context, params VerifySIWEParams) (*SIWEVerificationResult, error)

	// RecoverAddress recovers the signer address from a signature.
	// Only supported for EVM chains.
	RecoverAddress(ctx context.Context, params RecoverAddressParams) (Address, error)
}

// VerifySignatureParams contains parameters for signature verification.
type VerifySignatureParams struct {
	// Address is the expected signer address.
	Address Address

	// Message is the message that was signed.
	Message string

	// Signature is the cryptographic signature.
	Signature Signature
}

// VerifySIWEParams contains parameters for SIWE verification.
type VerifySIWEParams struct {
	// Message is the raw SIWE message string.
	Message string

	// Signature is the cryptographic signature.
	Signature Signature

	// ExpectedAddress is the expected signer address (optional).
	// If provided, verification fails if recovered address doesn't match.
	ExpectedAddress *Address

	// ExpectedDomain is the expected domain (optional).
	// If provided, verification fails if domain doesn't match.
	ExpectedDomain string

	// ExpectedNonce is the expected nonce (optional).
	// If provided, verification fails if nonce doesn't match.
	ExpectedNonce string

	// ExpectedChainID is the expected chain ID (optional).
	// If provided, verification fails if chain ID doesn't match.
	ExpectedChainID *ChainID
}

// SIWEVerificationResult contains the result of SIWE verification.
type SIWEVerificationResult struct {
	// Valid indicates if the signature is valid.
	Valid bool

	// Address is the recovered signer address.
	Address Address

	// Message is the parsed SIWE message.
	Message SIWEMessage
}

// RecoverAddressParams contains parameters for address recovery.
type RecoverAddressParams struct {
	// Message is the message that was signed.
	Message string

	// Signature is the cryptographic signature.
	Signature Signature

	// ChainID is the chain ID for the recovered address.
	ChainID ChainID
}

// ============================================================================
// DID Derivation Service
// ============================================================================

// DIDDerivationService derives DIDs from wallet addresses.
type DIDDerivationService interface {
	// DeriveDID derives a did:pkh DID from an address.
	DeriveDID(ctx context.Context, address Address) (did.DID, error)

	// DeriveDIDFromRaw derives a did:pkh DID from raw address components.
	DeriveDIDFromRaw(ctx context.Context, address string, chainID ChainID) (did.DID, error)

	// ParseDID extracts address information from a did:pkh DID.
	ParseDID(ctx context.Context, d did.DID) (Address, error)
}

// ============================================================================
// Challenge Service
// ============================================================================

// ChallengeService manages authentication challenges for wallets.
type ChallengeService interface {
	// CreateChallenge creates a new authentication challenge.
	CreateChallenge(ctx context.Context, params CreateChallengeParams) (*Challenge, error)

	// ValidateChallenge validates and consumes a challenge.
	ValidateChallenge(ctx context.Context, nonce string) (*Challenge, error)

	// InvalidateChallenge invalidates a challenge without consuming it.
	InvalidateChallenge(ctx context.Context, nonce string) error
}

// CreateChallengeParams contains parameters for challenge creation.
type CreateChallengeParams struct {
	// Address is the wallet address requesting the challenge.
	Address Address

	// Domain is the domain requesting authentication.
	Domain string

	// URI is the URI requesting authentication.
	URI string

	// Statement is an optional human-readable statement.
	Statement string

	// Resources is a list of resources being requested.
	Resources []string
}

// Challenge represents an authentication challenge.
type Challenge struct {
	// Nonce is the unique challenge identifier.
	Nonce string

	// Message is the full message to be signed.
	Message string

	// Address is the wallet address the challenge is for.
	Address Address

	// Domain is the domain that issued the challenge.
	Domain string

	// URI is the URI that issued the challenge.
	URI string

	// IssuedAt is when the challenge was issued.
	IssuedAt time.Time

	// ExpiresAt is when the challenge expires.
	ExpiresAt time.Time

	// Used indicates if the challenge has been consumed.
	Used bool

	// UsedAt is when the challenge was consumed.
	UsedAt *time.Time
}

// IsExpired returns true if the challenge has expired.
func (c Challenge) IsExpired() bool {
	return time.Now().After(c.ExpiresAt)
}

// IsValid returns true if the challenge is valid (not expired, not used).
func (c Challenge) IsValid() bool {
	return !c.IsExpired() && !c.Used
}

// ============================================================================
// Nonce Service
// ============================================================================

// NonceService manages nonces for replay protection.
type NonceService interface {
	// Generate generates a new unique nonce.
	Generate(ctx context.Context) (string, error)

	// Validate validates that a nonce is valid and unused.
	Validate(ctx context.Context, nonce string) error

	// Consume marks a nonce as used.
	Consume(ctx context.Context, nonce string) error

	// IsUsed returns true if the nonce has been used.
	IsUsed(ctx context.Context, nonce string) (bool, error)
}

// ============================================================================
// Chain Registry Service
// ============================================================================

// ChainRegistryService provides chain information.
type ChainRegistryService interface {
	// GetChain returns chain info for a chain ID.
	GetChain(ctx context.Context, chainID ChainID) (ChainInfo, error)

	// GetChainByName returns chain info for a chain name.
	GetChainByName(ctx context.Context, name string) (ChainInfo, error)

	// ListChains returns all supported chains.
	ListChains(ctx context.Context) ([]ChainInfo, error)

	// ListChainsByFamily returns chains for a specific family.
	ListChainsByFamily(ctx context.Context, family ChainFamily) ([]ChainInfo, error)

	// IsSupported returns true if the chain is supported.
	IsSupported(ctx context.Context, chainID ChainID) bool
}

// ============================================================================
// Wallet Linking Service
// ============================================================================

// WalletLinkingService handles linking wallets to users.
type WalletLinkingService interface {
	// LinkWallet links a wallet to a user after signature verification.
	LinkWallet(ctx context.Context, params LinkWalletParams) (*LinkWalletResult, error)

	// UnlinkWallet unlinks a wallet from a user.
	UnlinkWallet(ctx context.Context, params UnlinkWalletParams) error

	// GetLinkedWallets returns all wallets linked to a user.
	GetLinkedWallets(ctx context.Context, userID string) ([]LinkedWallet, error)
}

// LinkWalletParams contains parameters for linking a wallet.
type LinkWalletParams struct {
	// UserID is the user to link the wallet to.
	UserID string

	// Address is the wallet address to link.
	Address Address

	// Signature is the signature proving ownership.
	Signature Signature

	// Message is the signed message.
	Message string

	// Nonce is the challenge nonce.
	Nonce string

	// Label is an optional user-friendly name for the wallet.
	Label string

	// MakePrimary indicates if this should be the primary wallet.
	MakePrimary bool
}

// LinkWalletResult contains the result of linking a wallet.
type LinkWalletResult struct {
	// WalletID is the unique ID of the linked wallet.
	WalletID string

	// Address is the linked wallet address.
	Address Address

	// DID is the derived did:pkh for the wallet.
	DID did.DID

	// IsPrimary indicates if this is the primary wallet.
	IsPrimary bool
}

// UnlinkWalletParams contains parameters for unlinking a wallet.
type UnlinkWalletParams struct {
	// UserID is the user to unlink the wallet from.
	UserID string

	// WalletID is the wallet ID to unlink.
	WalletID string
}

// LinkedWallet represents a wallet linked to a user.
type LinkedWallet struct {
	// ID is the unique wallet ID.
	ID string

	// UserID is the user this wallet belongs to.
	UserID string

	// Address is the wallet address.
	Address Address

	// DID is the derived did:pkh for the wallet.
	DID did.DID

	// Label is the user-friendly name for the wallet.
	Label string

	// IsPrimary indicates if this is the primary wallet.
	IsPrimary bool

	// LinkedAt is when the wallet was linked.
	LinkedAt time.Time

	// LastUsedAt is when the wallet was last used.
	LastUsedAt *time.Time
}
