package signature

import (
	"context"
	"time"

	"github.com/0xsj/nexus/platform/internal/wallet/domain"
)

// ============================================================================
// Composite Signature Verifier
// ============================================================================

// CompositeSignatureVerifier implements domain.SignatureVerificationService.
// It delegates to family-specific verifiers based on address chain family.
type CompositeSignatureVerifier struct {
	evmVerifier *SIWEVerifier
	// solanaVerifier *SolanaVerifier  // Future
	// cosmosVerifier *CosmosVerifier  // Future
}

// NewCompositeSignatureVerifier creates a new composite signature verifier.
func NewCompositeSignatureVerifier(config VerifierConfig) *CompositeSignatureVerifier {
	return &CompositeSignatureVerifier{
		evmVerifier: NewSIWEVerifier(SIWEConfig{
			Domain:             config.Domain,
			AllowedDomains:     config.AllowedDomains,
			SkipTimeValidation: config.SkipTimeValidation,
			MaxMessageAge:      config.MaxMessageAge,
		}),
	}
}

// VerifierConfig contains configuration for the composite verifier.
type VerifierConfig struct {
	// Domain is the expected domain for SIWE verification.
	Domain string

	// AllowedDomains is a list of allowed domains.
	AllowedDomains []string

	// SkipTimeValidation skips time-based validation.
	SkipTimeValidation bool

	// MaxMessageAge is the maximum age of a signed message.
	MaxMessageAge time.Duration
}

// DefaultVerifierConfig returns default verifier configuration.
func DefaultVerifierConfig() VerifierConfig {
	return VerifierConfig{
		SkipTimeValidation: false,
		MaxMessageAge:      10 * time.Minute,
	}
}

// ============================================================================
// SignatureVerificationService Implementation
// ============================================================================

// VerifySignature verifies a signature against a message and address.
func (v *CompositeSignatureVerifier) VerifySignature(ctx context.Context, params domain.VerifySignatureParams) error {
	const op = "CompositeSignatureVerifier.VerifySignature"

	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Determine verifier based on address family
	switch params.Address.Family() {
	case domain.ChainFamilyEVM:
		return v.evmVerifier.VerifySignature(ctx, params)

	case domain.ChainFamilySolana:
		return domain.ErrChainNotSupported(op, "solana signature verification not yet implemented")

	case domain.ChainFamilyCosmos:
		return domain.ErrChainNotSupported(op, "cosmos signature verification not yet implemented")

	default:
		return domain.ErrChainNotSupported(op, params.Address.Family().String())
	}
}

// VerifySIWE verifies a Sign-In With Ethereum (EIP-4361) message.
func (v *CompositeSignatureVerifier) VerifySIWE(ctx context.Context, params domain.VerifySIWEParams) (*domain.SIWEVerificationResult, error) {
	const op = "CompositeSignatureVerifier.VerifySIWE"

	// Check context
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// SIWE is EVM-specific
	return v.evmVerifier.VerifySIWE(ctx, params)
}

// RecoverAddress recovers the signer address from a signature.
func (v *CompositeSignatureVerifier) RecoverAddress(ctx context.Context, params domain.RecoverAddressParams) (domain.Address, error) {
	const op = "CompositeSignatureVerifier.RecoverAddress"

	// Check context
	select {
	case <-ctx.Done():
		return domain.Address{}, ctx.Err()
	default:
	}

	// Determine chain family from chain ID
	chainInfo := domain.GetChainInfo(params.ChainID)
	if chainInfo.IsZero() {
		// Default to EVM if unknown
		return v.evmVerifier.RecoverAddress(ctx, params)
	}

	switch chainInfo.Family {
	case domain.ChainFamilyEVM:
		return v.evmVerifier.RecoverAddress(ctx, params)

	case domain.ChainFamilySolana:
		return domain.Address{}, domain.ErrChainNotSupported(op, "solana address recovery not yet implemented")

	case domain.ChainFamilyCosmos:
		return domain.Address{}, domain.ErrChainNotSupported(op, "cosmos address recovery not yet implemented")

	default:
		return domain.Address{}, domain.ErrChainNotSupported(op, chainInfo.Family.String())
	}
}

// ============================================================================
// Factory Functions
// ============================================================================

// NewDefaultSignatureVerifier creates a verifier with default configuration.
func NewDefaultSignatureVerifier() *CompositeSignatureVerifier {
	return NewCompositeSignatureVerifier(DefaultVerifierConfig())
}

// NewSignatureVerifierWithDomain creates a verifier with a specific domain.
func NewSignatureVerifierWithDomain(domain string) *CompositeSignatureVerifier {
	config := DefaultVerifierConfig()
	config.Domain = domain
	return NewCompositeSignatureVerifier(config)
}

// ============================================================================
// Interface Compliance
// ============================================================================

var _ domain.SignatureVerificationService = (*CompositeSignatureVerifier)(nil)
