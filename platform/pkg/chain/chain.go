package chain

import "fmt"

// ============================================================================
// Family
// ============================================================================

// Family represents a blockchain family/ecosystem.
type Family string

const (
	// FamilyEVM represents Ethereum Virtual Machine compatible chains.
	FamilyEVM Family = "evm"

	// FamilySolana represents the Solana ecosystem.
	FamilySolana Family = "solana"

	// FamilyCosmos represents Cosmos SDK based chains.
	FamilyCosmos Family = "cosmos"
)

// String returns the string representation of the family.
func (f Family) String() string {
	return string(f)
}

// IsValid returns true if the family is a known family.
func (f Family) IsValid() bool {
	switch f {
	case FamilyEVM, FamilySolana, FamilyCosmos:
		return true
	default:
		return false
	}
}

// ============================================================================
// ChainID
// ============================================================================

// ChainID uniquely identifies a blockchain.
// Format varies by family:
//   - EVM: numeric string (e.g., "1" for Ethereum mainnet, "137" for Polygon)
//   - Solana: cluster name (e.g., "mainnet-beta", "devnet")
//   - Cosmos: chain-id string (e.g., "cosmoshub-4", "osmosis-1")
type ChainID string

// String returns the string representation of the chain ID.
func (c ChainID) String() string {
	return string(c)
}

// IsEmpty returns true if the chain ID is empty.
func (c ChainID) IsEmpty() bool {
	return c == ""
}

// ============================================================================
// Well-Known Chain IDs
// ============================================================================

// EVM chain IDs.
const (
	ChainIDEthereumMainnet ChainID = "1"
	ChainIDEthereumGoerli  ChainID = "5"
	ChainIDEthereumSepolia ChainID = "11155111"
	ChainIDPolygon         ChainID = "137"
	ChainIDPolygonMumbai   ChainID = "80001"
	ChainIDArbitrum        ChainID = "42161"
	ChainIDOptimism        ChainID = "10"
	ChainIDBase            ChainID = "8453"
	ChainIDAvalanche       ChainID = "43114"
	ChainIDBSC             ChainID = "56"
)

// Solana chain IDs.
const (
	ChainIDSolanaMainnet ChainID = "mainnet-beta"
	ChainIDSolanaDevnet  ChainID = "devnet"
	ChainIDSolanaTestnet ChainID = "testnet"
)

// Cosmos chain IDs.
const (
	ChainIDCosmosHub ChainID = "cosmoshub-4"
	ChainIDOsmosis   ChainID = "osmosis-1"
)

// ============================================================================
// Chain
// ============================================================================

// Chain represents a blockchain network configuration.
type Chain struct {
	// ID is the unique identifier for this chain.
	ID ChainID

	// Family is the blockchain family (evm, solana, cosmos).
	Family Family

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

	// RPCURL is the default RPC endpoint (optional).
	RPCURL string
}

// String returns a string representation of the chain.
func (c Chain) String() string {
	return fmt.Sprintf("%s (%s)", c.Name, c.ID)
}

// IsZero returns true if the chain is uninitialized.
func (c Chain) IsZero() bool {
	return c.ID.IsEmpty()
}

// Validate validates the chain configuration.
func (c Chain) Validate() error {
	const op = "Chain.Validate"

	if c.ID.IsEmpty() {
		return ErrInvalidChainIDFor(op, "")
	}

	if !c.Family.IsValid() {
		return ErrChainNotSupportedFor(op, string(c.Family))
	}

	if c.Name == "" {
		return ErrInvalidChainIDFor(op, string(c.ID))
	}

	return nil
}

// ============================================================================
// Well-Known Chains
// ============================================================================

// Pre-defined chain configurations for common networks.
var (
	EthereumMainnet = Chain{
		ID:             ChainIDEthereumMainnet,
		Family:         FamilyEVM,
		Name:           "ethereum",
		DisplayName:    "Ethereum",
		IsTestnet:      false,
		NativeCurrency: "ETH",
		ExplorerURL:    "https://etherscan.io",
	}

	EthereumSepolia = Chain{
		ID:             ChainIDEthereumSepolia,
		Family:         FamilyEVM,
		Name:           "ethereum-sepolia",
		DisplayName:    "Ethereum Sepolia",
		IsTestnet:      true,
		NativeCurrency: "ETH",
		ExplorerURL:    "https://sepolia.etherscan.io",
	}

	Polygon = Chain{
		ID:             ChainIDPolygon,
		Family:         FamilyEVM,
		Name:           "polygon",
		DisplayName:    "Polygon",
		IsTestnet:      false,
		NativeCurrency: "MATIC",
		ExplorerURL:    "https://polygonscan.com",
	}

	Arbitrum = Chain{
		ID:             ChainIDArbitrum,
		Family:         FamilyEVM,
		Name:           "arbitrum",
		DisplayName:    "Arbitrum One",
		IsTestnet:      false,
		NativeCurrency: "ETH",
		ExplorerURL:    "https://arbiscan.io",
	}

	Optimism = Chain{
		ID:             ChainIDOptimism,
		Family:         FamilyEVM,
		Name:           "optimism",
		DisplayName:    "Optimism",
		IsTestnet:      false,
		NativeCurrency: "ETH",
		ExplorerURL:    "https://optimistic.etherscan.io",
	}

	Base = Chain{
		ID:             ChainIDBase,
		Family:         FamilyEVM,
		Name:           "base",
		DisplayName:    "Base",
		IsTestnet:      false,
		NativeCurrency: "ETH",
		ExplorerURL:    "https://basescan.org",
	}

	// SolanaMainnet is pre-defined for future use.
	SolanaMainnet = Chain{
		ID:             ChainIDSolanaMainnet,
		Family:         FamilySolana,
		Name:           "solana",
		DisplayName:    "Solana",
		IsTestnet:      false,
		NativeCurrency: "SOL",
		ExplorerURL:    "https://explorer.solana.com",
	}

	// CosmosHub is pre-defined for future use.
	CosmosHub = Chain{
		ID:             ChainIDCosmosHub,
		Family:         FamilyCosmos,
		Name:           "cosmoshub",
		DisplayName:    "Cosmos Hub",
		IsTestnet:      false,
		NativeCurrency: "ATOM",
		ExplorerURL:    "https://www.mintscan.io/cosmos",
	}
)

// DefaultEVMChains returns the default set of supported EVM chains.
func DefaultEVMChains() []Chain {
	return []Chain{
		EthereumMainnet,
		EthereumSepolia,
		Polygon,
		Arbitrum,
		Optimism,
		Base,
	}
}

// AllDefaultChains returns all default supported chains.
func AllDefaultChains() []Chain {
	chains := DefaultEVMChains()
	// Add Solana and Cosmos when supported
	// chains = append(chains, SolanaMainnet, CosmosHub)
	return chains
}
