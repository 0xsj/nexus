package did

import (
	"context"

	"github.com/0xsj/nexus/platform/internal/wallet/domain"
	"github.com/0xsj/nexus/platform/pkg/did"
	"github.com/0xsj/nexus/platform/pkg/did/pkh"
)

// ============================================================================
// DID Derivation Service
// ============================================================================

// DerivationService implements domain.DIDDerivationService.
// It derives did:pkh DIDs from wallet addresses.
type DerivationService struct{}

// NewDerivationService creates a new DID derivation service.
func NewDerivationService() *DerivationService {
	return &DerivationService{}
}

// DeriveDID derives a did:pkh DID from a wallet address.
func (s *DerivationService) DeriveDID(ctx context.Context, address domain.Address) (did.DID, error) {
	const op = "DerivationService.DeriveDID"

	// Check context
	select {
	case <-ctx.Done():
		return did.DID{}, ctx.Err()
	default:
	}

	// Validate address
	if address.IsZero() {
		return did.DID{}, domain.ErrInvalidAddress(op, "", "address is required")
	}

	// Use the address's built-in ToDID method which uses pkh.Generate
	return address.ToDID()
}

// DeriveDIDFromRaw derives a did:pkh DID from raw address components.
func (s *DerivationService) DeriveDIDFromRaw(ctx context.Context, address string, chainID domain.ChainID) (did.DID, error) {
	const op = "DerivationService.DeriveDIDFromRaw"

	// Check context
	select {
	case <-ctx.Done():
		return did.DID{}, ctx.Err()
	default:
	}

	// Create a domain address first
	addr, err := domain.NewAddress(address, chainID)
	if err != nil {
		return did.DID{}, err
	}

	// Derive DID from the address
	return addr.ToDID()
}

// ParseDID extracts address information from a did:pkh DID.
func (s *DerivationService) ParseDID(ctx context.Context, d did.DID) (domain.Address, error) {
	const op = "DerivationService.ParseDID"

	// Check context
	select {
	case <-ctx.Done():
		return domain.Address{}, ctx.Err()
	default:
	}

	// Parse the did:pkh to get account info
	account, err := pkh.ParseDID(d)
	if err != nil {
		return domain.Address{}, domain.ErrInvalidAddress(op, d.String(), err.Error())
	}

	// Map namespace to chain family
	family := s.namespaceToChainFamily(account.Namespace)

	// Map reference to chain ID
	chainID := s.referenceToChainID(account.Namespace, account.Reference)

	// Create domain address (unchecked since we trust the DID format)
	return domain.NewAddressUnchecked(
		account.Address,
		account.Address,
		chainID,
		family,
	), nil
}

// ============================================================================
// Chain Mapping Helpers
// ============================================================================

// namespaceToChainFamily maps CAIP-2 namespace to domain.ChainFamily.
func (s *DerivationService) namespaceToChainFamily(namespace string) domain.ChainFamily {
	switch namespace {
	case pkh.NamespaceEIP155:
		return domain.ChainFamilyEVM
	case pkh.NamespaceSolana:
		return domain.ChainFamilySolana
	case pkh.NamespaceCosmos:
		return domain.ChainFamilyCosmos
	default:
		return domain.ChainFamily("")
	}
}

// referenceToChainID maps CAIP-2 namespace and reference to domain.ChainID.
func (s *DerivationService) referenceToChainID(namespace, reference string) domain.ChainID {
	switch namespace {
	case pkh.NamespaceEIP155:
		return s.evmReferenceToChainID(reference)
	case pkh.NamespaceSolana:
		return s.solanaReferenceToChainID(reference)
	case pkh.NamespaceCosmos:
		return s.cosmosReferenceToChainID(reference)
	default:
		return domain.ChainID(namespace + ":" + reference)
	}
}

// evmReferenceToChainID maps EVM chain reference to domain.ChainID.
func (s *DerivationService) evmReferenceToChainID(reference string) domain.ChainID {
	switch reference {
	case "1":
		return domain.ChainIDEthereumMainnet
	case "11155111":
		return domain.ChainIDEthereumSepolia
	case "137":
		return domain.ChainIDPolygon
	case "42161":
		return domain.ChainIDArbitrum
	case "10":
		return domain.ChainIDOptimism
	case "8453":
		return domain.ChainIDBase
	default:
		return domain.ChainID(reference)
	}
}

// solanaReferenceToChainID maps Solana chain reference to domain.ChainID.
func (s *DerivationService) solanaReferenceToChainID(reference string) domain.ChainID {
	switch reference {
	case "mainnet":
		return domain.ChainIDSolanaMainnet
	case "devnet":
		return domain.ChainIDSolanaDevnet
	default:
		return domain.ChainID(reference)
	}
}

// cosmosReferenceToChainID maps Cosmos chain reference to domain.ChainID.
func (s *DerivationService) cosmosReferenceToChainID(reference string) domain.ChainID {
	switch reference {
	case "cosmoshub-4":
		return domain.ChainIDCosmosHub
	case "osmosis-1":
		return domain.ChainIDOsmosis
	default:
		return domain.ChainID(reference)
	}
}

// ============================================================================
// Interface Compliance
// ============================================================================

var _ domain.DIDDerivationService = (*DerivationService)(nil)
