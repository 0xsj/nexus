package did

import (
	"context"

	"github.com/0xsj/nexus/platform/internal/identity/domain"
	"github.com/0xsj/nexus/platform/pkg/chain"
	"github.com/0xsj/nexus/platform/pkg/did"
	"github.com/0xsj/nexus/platform/pkg/did/key"
	"github.com/0xsj/nexus/platform/pkg/did/pkh"
	"github.com/0xsj/nexus/platform/pkg/id"
)

// ============================================================================
// Generation Service
// ============================================================================

// GenerationService implements domain.DIDGenerationService.
// It wraps the pkg/did/key and pkg/did/pkh packages to provide
// DID generation for the identity module.
type GenerationService struct {
	idGenerator id.Generator
}

// Ensure GenerationService implements domain.DIDGenerationService.
var _ domain.DIDGenerationService = (*GenerationService)(nil)

// NewGenerationService creates a new DID generation service.
func NewGenerationService(idGenerator id.Generator) *GenerationService {
	return &GenerationService{
		idGenerator: idGenerator,
	}
}

// GenerateCustodialDID generates a new custodial did:key for a user.
// This creates a new Ed25519 key pair and derives the did:key from it.
//
// Note: In production, the private key should be stored securely (e.g., in a
// key management service). Currently, the key pair is generated but not persisted.
// A future CustodialKeyService will handle secure key storage.
func (s *GenerationService) GenerateCustodialDID(ctx context.Context, userID string) (did.DID, string, error) {
	// Generate a unique ID for this linked DID entry
	linkedDIDID := s.idGenerator.Generate().String()

	// Generate a new Ed25519 key pair and derive did:key
	generatedDID, _, err := key.Generate()
	if err != nil {
		return did.DID{}, "", err
	}

	// TODO: Store the key pair securely using CustodialKeyService
	// For MVP, the key is generated but not persisted, which means
	// the user won't be able to sign credentials with this DID.
	// This is acceptable for initial user registration flow.

	return generatedDID, linkedDIDID, nil
}

// DeriveWalletDID derives a did:pkh from a wallet address.
// Uses the pkg/did/pkh package for proper CAIP-10 formatting.
func (s *GenerationService) DeriveWalletDID(ctx context.Context, wallet domain.WalletAddress) (did.DID, error) {
	// Map domain types to chain.Address
	chainID := domainChainToChainID(wallet.Chain)

	// Use the EVM-specific generator for EVM chains
	if wallet.Chain.IsEVM() {
		return pkh.GenerateEVM(wallet.Address, chainID)
	}

	// For non-EVM chains, use the raw generator with appropriate namespace
	namespace := domainChainToNamespace(wallet.Chain)
	reference := string(chainID)

	return pkh.GenerateFromRaw(namespace, reference, wallet.Address)
}

// DeriveWalletDIDFromRaw derives a did:pkh from raw address and chain info.
// This is a convenience method for when you have domain types instead of WalletAddress.
func (s *GenerationService) DeriveWalletDIDFromRaw(ctx context.Context, address string, chainType domain.Chain) (did.DID, error) {
	wallet := domain.NewWalletAddress(address, chainType)
	return s.DeriveWalletDID(ctx, wallet)
}

// ============================================================================
// Chain Mapping Helpers
// ============================================================================

// domainChainToChainID maps domain.Chain to chain.ChainID.
func domainChainToChainID(c domain.Chain) chain.ChainID {
	switch c {
	case domain.ChainEthereum:
		return chain.ChainIDEthereumMainnet
	case domain.ChainPolygon:
		return chain.ChainIDPolygon
	case domain.ChainArbitrum:
		return chain.ChainIDArbitrum
	case domain.ChainOptimism:
		return chain.ChainIDOptimism
	case domain.ChainBase:
		return chain.ChainIDBase
	case domain.ChainSolana:
		return chain.ChainIDSolanaMainnet
	case domain.ChainCosmos:
		return chain.ChainIDCosmosHub
	default:
		return chain.ChainIDEthereumMainnet
	}
}

// domainChainToNamespace maps domain.Chain to CAIP-2 namespace.
func domainChainToNamespace(c domain.Chain) string {
	switch c {
	case domain.ChainEthereum, domain.ChainPolygon, domain.ChainArbitrum,
		domain.ChainOptimism, domain.ChainBase:
		return pkh.NamespaceEIP155
	case domain.ChainSolana:
		return pkh.NamespaceSolana
	case domain.ChainCosmos:
		return pkh.NamespaceCosmos
	default:
		return pkh.NamespaceEIP155
	}
}
