package domain

import (
	"github.com/0xsj/nexus/platform/pkg/chain"
)

// ============================================================================
// Chain Family
// ============================================================================

// ChainFamily represents a blockchain family/ecosystem.
// Re-exported from pkg/chain for domain isolation.
type ChainFamily string

const (
	// ChainFamilyEVM represents Ethereum Virtual Machine compatible chains.
	ChainFamilyEVM ChainFamily = ChainFamily(chain.FamilyEVM)

	// ChainFamilySolana represents the Solana ecosystem.
	ChainFamilySolana ChainFamily = ChainFamily(chain.FamilySolana)

	// ChainFamilyCosmos represents Cosmos SDK based chains.
	ChainFamilyCosmos ChainFamily = ChainFamily(chain.FamilyCosmos)
)

// String returns the string representation of the family.
func (f ChainFamily) String() string {
	return string(f)
}

// IsValid returns true if the family is a known family.
func (f ChainFamily) IsValid() bool {
	return chain.Family(f).IsValid()
}

// ToPkg converts to pkg/chain.Family.
func (f ChainFamily) ToPkg() chain.Family {
	return chain.Family(f)
}

// ChainFamilyFromPkg converts from pkg/chain.Family.
func ChainFamilyFromPkg(f chain.Family) ChainFamily {
	return ChainFamily(f)
}

// ============================================================================
// Chain ID
// ============================================================================

// ChainID uniquely identifies a blockchain.
// Re-exported from pkg/chain for domain isolation.
type ChainID string

// Well-known EVM chain IDs.
const (
	ChainIDEthereumMainnet ChainID = ChainID(chain.ChainIDEthereumMainnet)
	ChainIDEthereumSepolia ChainID = ChainID(chain.ChainIDEthereumSepolia)
	ChainIDPolygon         ChainID = ChainID(chain.ChainIDPolygon)
	ChainIDArbitrum        ChainID = ChainID(chain.ChainIDArbitrum)
	ChainIDOptimism        ChainID = ChainID(chain.ChainIDOptimism)
	ChainIDBase            ChainID = ChainID(chain.ChainIDBase)
)

// Well-known Solana chain IDs.
const (
	ChainIDSolanaMainnet ChainID = ChainID(chain.ChainIDSolanaMainnet)
	ChainIDSolanaDevnet  ChainID = ChainID(chain.ChainIDSolanaDevnet)
)

// Well-known Cosmos chain IDs.
const (
	ChainIDCosmosHub ChainID = ChainID(chain.ChainIDCosmosHub)
	ChainIDOsmosis   ChainID = ChainID(chain.ChainIDOsmosis)
)

// String returns the string representation of the chain ID.
func (c ChainID) String() string {
	return string(c)
}

// IsEmpty returns true if the chain ID is empty.
func (c ChainID) IsEmpty() bool {
	return c == ""
}

// ToPkg converts to pkg/chain.ChainID.
func (c ChainID) ToPkg() chain.ChainID {
	return chain.ChainID(c)
}

// ChainIDFromPkg converts from pkg/chain.ChainID.
func ChainIDFromPkg(c chain.ChainID) ChainID {
	return ChainID(c)
}

// ============================================================================
// Chain Info
// ============================================================================

// ChainInfo represents a blockchain network configuration.
type ChainInfo struct {
	// ID is the unique identifier for this chain.
	ID ChainID

	// Family is the blockchain family (evm, solana, cosmos).
	Family ChainFamily

	// Name is the human-readable chain name.
	Name string

	// DisplayName is the user-facing display name.
	DisplayName string

	// IsTestnet indicates if this is a test network.
	IsTestnet bool

	// NativeCurrency is the native currency symbol (e.g., "ETH", "SOL").
	NativeCurrency string

	// ExplorerURL is the block explorer base URL.
	ExplorerURL string
}

// IsZero returns true if the chain info is uninitialized.
func (c ChainInfo) IsZero() bool {
	return c.ID.IsEmpty()
}

// IsEVM returns true if this is an EVM-compatible chain.
func (c ChainInfo) IsEVM() bool {
	return c.Family == ChainFamilyEVM
}

// IsSolana returns true if this is a Solana chain.
func (c ChainInfo) IsSolana() bool {
	return c.Family == ChainFamilySolana
}

// IsCosmos returns true if this is a Cosmos chain.
func (c ChainInfo) IsCosmos() bool {
	return c.Family == ChainFamilyCosmos
}

// ChainInfoFromPkg converts from pkg/chain.Chain.
func ChainInfoFromPkg(c chain.Chain) ChainInfo {
	return ChainInfo{
		ID:             ChainIDFromPkg(c.ID),
		Family:         ChainFamilyFromPkg(c.Family),
		Name:           c.Name,
		DisplayName:    c.DisplayName,
		IsTestnet:      c.IsTestnet,
		NativeCurrency: c.NativeCurrency,
		ExplorerURL:    c.ExplorerURL,
	}
}

// ============================================================================
// Well-Known Chains
// ============================================================================

var (
	// EthereumMainnet is the Ethereum mainnet chain.
	EthereumMainnet = ChainInfoFromPkg(chain.EthereumMainnet)

	// EthereumSepolia is the Ethereum Sepolia testnet.
	EthereumSepolia = ChainInfoFromPkg(chain.EthereumSepolia)

	// Polygon is the Polygon mainnet chain.
	Polygon = ChainInfoFromPkg(chain.Polygon)

	// Arbitrum is the Arbitrum One chain.
	Arbitrum = ChainInfoFromPkg(chain.Arbitrum)

	// Optimism is the Optimism mainnet chain.
	Optimism = ChainInfoFromPkg(chain.Optimism)

	// Base is the Base mainnet chain.
	Base = ChainInfoFromPkg(chain.Base)

	// SolanaMainnet is the Solana mainnet chain.
	SolanaMainnet = ChainInfoFromPkg(chain.SolanaMainnet)

	// CosmosHub is the Cosmos Hub chain.
	CosmosHub = ChainInfoFromPkg(chain.CosmosHub)
)

// ============================================================================
// Chain Lookup
// ============================================================================

// GetChainInfo returns chain info for a given chain ID.
// Returns zero value if chain is not found.
func GetChainInfo(id ChainID) ChainInfo {
	switch id {
	case ChainIDEthereumMainnet:
		return EthereumMainnet
	case ChainIDEthereumSepolia:
		return EthereumSepolia
	case ChainIDPolygon:
		return Polygon
	case ChainIDArbitrum:
		return Arbitrum
	case ChainIDOptimism:
		return Optimism
	case ChainIDBase:
		return Base
	case ChainIDSolanaMainnet:
		return SolanaMainnet
	case ChainIDCosmosHub:
		return CosmosHub
	default:
		return ChainInfo{}
	}
}

// GetChainFamily returns the family for a given chain ID.
func GetChainFamily(id ChainID) ChainFamily {
	info := GetChainInfo(id)
	if info.IsZero() {
		return ""
	}
	return info.Family
}

// IsChainSupported returns true if the chain ID is supported.
func IsChainSupported(id ChainID) bool {
	return !GetChainInfo(id).IsZero()
}

// SupportedChainIDs returns all supported chain IDs.
func SupportedChainIDs() []ChainID {
	return []ChainID{
		ChainIDEthereumMainnet,
		ChainIDEthereumSepolia,
		ChainIDPolygon,
		ChainIDArbitrum,
		ChainIDOptimism,
		ChainIDBase,
		ChainIDSolanaMainnet,
		ChainIDCosmosHub,
	}
}

// SupportedEVMChainIDs returns all supported EVM chain IDs.
func SupportedEVMChainIDs() []ChainID {
	return []ChainID{
		ChainIDEthereumMainnet,
		ChainIDEthereumSepolia,
		ChainIDPolygon,
		ChainIDArbitrum,
		ChainIDOptimism,
		ChainIDBase,
	}
}
