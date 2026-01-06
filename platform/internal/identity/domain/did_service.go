package domain

import (
	"context"

	"github.com/0xsj/nexus/platform/pkg/did"
)

// ============================================================================
// DID Generation Service
// ============================================================================

// DIDGenerationService handles DID generation for users.
// This extends the basic DIDService with custodial key management.
type DIDGenerationService interface {
	// GenerateCustodialDID generates a new custodial did:key for a user.
	// The private key is generated and stored securely by Nexus.
	// Returns the DID and the LinkedDID ID.
	GenerateCustodialDID(ctx context.Context, userID string) (did.DID, string, error)

	// DeriveWalletDID derives a did:pkh from a wallet address.
	// Uses the domain WalletAddress which contains address and chain info.
	DeriveWalletDID(ctx context.Context, wallet WalletAddress) (did.DID, error)

	// DeriveWalletDIDFromRaw derives a did:pkh from raw address and chain.
	// Convenience method when you don't have a WalletAddress.
	DeriveWalletDIDFromRaw(ctx context.Context, address string, chain Chain) (did.DID, error)
}

// ============================================================================
// DID Resolution Service
// ============================================================================

// DIDResolutionService resolves DIDs to DID Documents.
type DIDResolutionService interface {
	// Resolve resolves a DID to its DID Document.
	Resolve(ctx context.Context, d did.DID) (*did.Document, error)

	// ResolveString resolves a DID string to its DID Document.
	ResolveString(ctx context.Context, didString string) (*did.Document, error)

	// CanResolve returns true if this service can resolve the given DID method.
	CanResolve(method did.Method) bool
}

// ============================================================================
// Custodial Key Service
// ============================================================================

// CustodialKeyService manages private keys for custodial DIDs.
// Keys are encrypted at rest and never exposed directly.
type CustodialKeyService interface {
	// GenerateKey generates a new key pair and stores it securely.
	// Returns the key ID (used to reference the key) and the public key.
	GenerateKey(ctx context.Context, userID string, algorithm KeyAlgorithm) (keyID string, publicKey []byte, err error)

	// Sign signs data using a stored key.
	// The private key never leaves the secure storage.
	Sign(ctx context.Context, keyID string, data []byte) (signature []byte, err error)

	// GetPublicKey retrieves the public key for a key ID.
	GetPublicKey(ctx context.Context, keyID string) (publicKey []byte, algorithm KeyAlgorithm, err error)

	// DeleteKey permanently deletes a key.
	// Use with caution — credentials signed with this key become unverifiable.
	DeleteKey(ctx context.Context, keyID string) error

	// KeyExists checks if a key exists.
	KeyExists(ctx context.Context, keyID string) (bool, error)
}

// KeyAlgorithm represents supported key algorithms for custodial DIDs.
type KeyAlgorithm string

const (
	// KeyAlgorithmEd25519 is the Ed25519 signature algorithm.
	// Default for did:key generation.
	KeyAlgorithmEd25519 KeyAlgorithm = "Ed25519"

	// KeyAlgorithmSecp256k1 is the secp256k1 signature algorithm.
	// Used for Ethereum-compatible signatures.
	KeyAlgorithmSecp256k1 KeyAlgorithm = "secp256k1"
)

// String returns the string representation.
func (a KeyAlgorithm) String() string {
	return string(a)
}

// IsValid returns true if the algorithm is supported.
func (a KeyAlgorithm) IsValid() bool {
	switch a {
	case KeyAlgorithmEd25519, KeyAlgorithmSecp256k1:
		return true
	default:
		return false
	}
}

// ============================================================================
// Combined DID Service
// ============================================================================

// FullDIDService combines all DID-related operations.
// Use this interface when you need complete DID functionality.
type FullDIDService interface {
	DIDGenerationService
	DIDResolutionService
}

// ============================================================================
// DID Linking Params
// ============================================================================

// LinkCustodialDIDParams contains parameters for linking a custodial DID.
type LinkCustodialDIDParams struct {
	UserID      string
	Label       string
	Algorithm   KeyAlgorithm
	MakePrimary bool
}

// LinkWalletDIDParams contains parameters for linking a wallet DID.
type LinkWalletDIDParams struct {
	UserID      string
	Address     string
	Chain       Chain
	Label       string
	MakePrimary bool
}

// ============================================================================
// DID Operation Results
// ============================================================================

// GenerateDIDResult contains the result of DID generation.
type GenerateDIDResult struct {
	DID         did.DID
	LinkedDIDID string
	KeyID       string // For custodial DIDs, the key storage ID
	Source      DIDSource
}
