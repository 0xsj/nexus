package signature

import (
	"context"

	"github.com/0xsj/nexus/platform/internal/identity/domain"
	"github.com/0xsj/nexus/platform/pkg/errors"
	"github.com/0xsj/nexus/platform/pkg/siwe"
)

// ============================================================================
// Verifier
// ============================================================================

// Verifier implements domain.SignatureVerifier using SIWE for wallet signatures.
type Verifier struct {
	// Default domain for SIWE verification
	domain string

	// Supported chain IDs
	supportedChains map[domain.Chain]int
}

// Ensure Verifier implements domain.SignatureVerifier.
var _ domain.SignatureVerifier = (*Verifier)(nil)

// Config contains verifier configuration.
type Config struct {
	// Domain is the expected SIWE domain (e.g., "proof.io")
	Domain string

	// SupportedChains maps chain identifiers to chain IDs
	SupportedChains map[domain.Chain]int
}

// DefaultConfig returns default configuration.
func DefaultConfig() Config {
	return Config{
		Domain: "proof.io",
		SupportedChains: map[domain.Chain]int{
			domain.ChainEthereum: siwe.ChainIDEthereum,
			domain.ChainPolygon:  siwe.ChainIDPolygon,
			domain.ChainArbitrum: siwe.ChainIDArbitrum,
			domain.ChainOptimism: siwe.ChainIDOptimism,
			domain.ChainBase:     siwe.ChainIDBase,
		},
	}
}

// NewVerifier creates a new signature verifier.
func NewVerifier(config Config) *Verifier {
	return &Verifier{
		domain:          config.Domain,
		supportedChains: config.SupportedChains,
	}
}

// ============================================================================
// Wallet Signature Verification
// ============================================================================

// VerifyWalletSignature verifies a wallet signature (SIWE/EIP-4361).
func (v *Verifier) VerifyWalletSignature(ctx context.Context, params domain.WalletSignatureParams) error {
	const op = "signature.Verifier.VerifyWalletSignature"

	// Get expected chain ID
	chainID, ok := v.supportedChains[params.Chain]
	if !ok {
		return errors.Validation(op, "unsupported chain: "+params.Chain.String())
	}

	// Verify using SIWE
	msg, err := siwe.VerifyWithOptions(params.Message, params.Signature, siwe.VerifyOptions{
		ExpectedDomain:  v.domain,
		ExpectedNonce:   params.Nonce,
		ExpectedChainID: chainID,
	})
	if err != nil {
		// Map SIWE errors to domain errors
		if siwe.IsInvalidSignature(err) {
			return domain.ErrInvalidCredentials(op)
		}
		if siwe.IsMessageExpired(err) {
			return domain.ErrChallengeExpired(op, params.Nonce)
		}
		if siwe.IsAddressMismatch(err) {
			return domain.ErrInvalidCredentials(op)
		}
		return errors.Wrap(err, op)
	}

	// Verify the address matches (case-insensitive)
	if !siwe.AddressesEqual(msg.Address, params.Address) {
		return domain.ErrInvalidCredentials(op)
	}

	return nil
}

// VerifyDIDSignature verifies a DID-based signature.
func (v *Verifier) VerifyDIDSignature(ctx context.Context, params domain.DIDSignatureParams) error {
	const op = "signature.Verifier.VerifyDIDSignature"

	// TODO: Implement DID signature verification
	// This would involve:
	// 1. Resolving the DID document
	// 2. Finding the verification method (by KeyID or default)
	// 3. Verifying the signature with the public key

	return errors.Validation(op, "DID signature verification not yet implemented")
}

// ============================================================================
// Address Utilities
// ============================================================================

// RecoverAddress recovers the Ethereum address from a signed message.
func (v *Verifier) RecoverAddress(message, signature string) (string, error) {
	const op = "signature.Verifier.RecoverAddress"

	address, err := siwe.RecoverAddress(message, signature)
	if err != nil {
		return "", errors.Wrap(err, op)
	}

	return address, nil
}

// ValidateAddress checks if an address is valid for the given chain.
func (v *Verifier) ValidateAddress(address string, chain domain.Chain) error {
	const op = "signature.Verifier.ValidateAddress"

	// For EVM chains, validate Ethereum address format
	switch chain {
	case domain.ChainEthereum, domain.ChainPolygon, domain.ChainArbitrum, domain.ChainOptimism, domain.ChainBase:
		if !siwe.IsValidAddress(address) {
			return errors.Validation(op, "invalid Ethereum address: "+address)
		}
	default:
		return errors.Validation(op, "unsupported chain: "+chain.String())
	}

	return nil
}

// NormalizeAddress normalizes an address for consistent storage.
func (v *Verifier) NormalizeAddress(address string) string {
	return siwe.NormalizeAddress(address)
}

// AddressesEqual compares two addresses (case-insensitive for EVM).
func (v *Verifier) AddressesEqual(a, b string) bool {
	return siwe.AddressesEqual(a, b)
}
