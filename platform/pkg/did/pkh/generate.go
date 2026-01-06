package pkh

import (
	"fmt"
	"strings"

	"github.com/0xsj/nexus/platform/pkg/chain"
	"github.com/0xsj/nexus/platform/pkg/did"
)

// ============================================================================
// CAIP-10 Account ID
// ============================================================================

// AccountID represents a CAIP-10 account identifier.
// Format: <chain_namespace>:<chain_reference>:<address>
//
// Examples:
//   - eip155:1:0xab16a96d359ec26a11e2c2b3d8f8b8942d5bfcdb (Ethereum mainnet)
//   - eip155:137:0xab16a96d359ec26a11e2c2b3d8f8b8942d5bfcdb (Polygon)
//   - solana:mainnet:7S3P4HxJpyyigGzodYwHtCxZyUQe9JiBMHyRWXArAaKv (Solana)
//   - cosmos:cosmoshub-4:cosmos1... (Cosmos Hub)
type AccountID struct {
	// Namespace is the chain namespace (e.g., "eip155", "solana", "cosmos").
	Namespace string

	// Reference is the chain reference (e.g., "1" for Ethereum mainnet).
	Reference string

	// Address is the account address.
	Address string
}

// String returns the CAIP-10 account ID string.
func (a AccountID) String() string {
	return fmt.Sprintf("%s:%s:%s", a.Namespace, a.Reference, a.Address)
}

// Parse parses a CAIP-10 account ID string.
func Parse(s string) (AccountID, error) {
	const op = "pkh.Parse"

	parts := strings.SplitN(s, ":", 3)
	if len(parts) != 3 {
		return AccountID{}, did.ErrInvalidDID(op, "invalid CAIP-10 format: expected namespace:reference:address")
	}

	if parts[0] == "" {
		return AccountID{}, did.ErrInvalidDID(op, "namespace cannot be empty")
	}
	if parts[1] == "" {
		return AccountID{}, did.ErrInvalidDID(op, "reference cannot be empty")
	}
	if parts[2] == "" {
		return AccountID{}, did.ErrInvalidDID(op, "address cannot be empty")
	}

	return AccountID{
		Namespace: parts[0],
		Reference: parts[1],
		Address:   parts[2],
	}, nil
}

// ============================================================================
// Chain Namespaces
// ============================================================================

// CAIP-2 chain namespaces.
const (
	// NamespaceEIP155 is the namespace for EVM-compatible chains.
	NamespaceEIP155 = "eip155"

	// NamespaceSolana is the namespace for Solana.
	NamespaceSolana = "solana"

	// NamespaceCosmos is the namespace for Cosmos SDK chains.
	NamespaceCosmos = "cosmos"
)

// NamespaceForFamily returns the CAIP-2 namespace for a chain family.
func NamespaceForFamily(family chain.Family) string {
	switch family {
	case chain.FamilyEVM:
		return NamespaceEIP155
	case chain.FamilySolana:
		return NamespaceSolana
	case chain.FamilyCosmos:
		return NamespaceCosmos
	default:
		return ""
	}
}

// ============================================================================
// DID Generation
// ============================================================================

// Generate creates a did:pkh DID from an address and chain.
//
// The did:pkh format follows CAIP-10:
//   - did:pkh:<namespace>:<chain_reference>:<address>
//
// Examples:
//   - did:pkh:eip155:1:0xAb5801a7D398351b8bE11C439e05C5B3259aeC9B
//   - did:pkh:eip155:137:0xAb5801a7D398351b8bE11C439e05C5B3259aeC9B
//   - did:pkh:solana:mainnet:7S3P4HxJpyyigGzodYwHtCxZyUQe9JiBMHyRWXArAaKv
func Generate(address chain.Address) (did.DID, error) {
	const op = "pkh.Generate"

	if address.IsZero() {
		return did.DID{}, did.ErrInvalidDID(op, "address is empty")
	}

	// Build CAIP-10 account ID
	namespace := NamespaceForFamily(address.Family())
	if namespace == "" {
		return did.DID{}, did.ErrUnsupportedMethod(op, "pkh").
			WithMeta("reason", "unsupported chain family: "+string(address.Family()))
	}

	// Use chain ID as reference, defaulting to mainnet if not specified
	reference := string(address.ChainID())
	if reference == "" {
		reference = defaultReferenceForFamily(address.Family())
	}

	// Build method-specific ID: namespace:reference:address
	methodSpecificID := fmt.Sprintf("%s:%s:%s", namespace, reference, address.Normalized())

	return did.New(did.MethodPKH, methodSpecificID)
}

// GenerateFromRaw creates a did:pkh DID from raw address components.
func GenerateFromRaw(namespace, reference, address string) (did.DID, error) {
	const op = "pkh.GenerateFromRaw"

	if namespace == "" {
		return did.DID{}, did.ErrInvalidDID(op, "namespace is empty")
	}
	if reference == "" {
		return did.DID{}, did.ErrInvalidDID(op, "reference is empty")
	}
	if address == "" {
		return did.DID{}, did.ErrInvalidDID(op, "address is empty")
	}

	methodSpecificID := fmt.Sprintf("%s:%s:%s", namespace, reference, address)
	return did.New(did.MethodPKH, methodSpecificID)
}

// GenerateEVM creates a did:pkh DID for an EVM address.
// This is a convenience function for the common EVM case.
func GenerateEVM(address string, chainID chain.ChainID) (did.DID, error) {
	const op = "pkh.GenerateEVM"

	if address == "" {
		return did.DID{}, did.ErrInvalidDID(op, "address is empty")
	}

	reference := string(chainID)
	if reference == "" {
		reference = string(chain.ChainIDEthereumMainnet) // Default to mainnet
	}

	methodSpecificID := fmt.Sprintf("%s:%s:%s", NamespaceEIP155, reference, address)
	return did.New(did.MethodPKH, methodSpecificID)
}

// GenerateEthereumMainnet creates a did:pkh DID for Ethereum mainnet.
func GenerateEthereumMainnet(address string) (did.DID, error) {
	return GenerateEVM(address, chain.ChainIDEthereumMainnet)
}

// ============================================================================
// DID Parsing
// ============================================================================

// ParseDID extracts the AccountID from a did:pkh DID.
func ParseDID(d did.DID) (AccountID, error) {
	const op = "pkh.ParseDID"

	if d.Method() != did.MethodPKH {
		return AccountID{}, did.ErrInvalidDID(op, "not a did:pkh DID")
	}

	return Parse(d.MethodSpecificID())
}

// ExtractAddress extracts just the address from a did:pkh DID.
func ExtractAddress(d did.DID) (string, error) {
	account, err := ParseDID(d)
	if err != nil {
		return "", err
	}
	return account.Address, nil
}

// ExtractChainID extracts the chain reference from a did:pkh DID.
func ExtractChainID(d did.DID) (chain.ChainID, error) {
	account, err := ParseDID(d)
	if err != nil {
		return "", err
	}
	return chain.ChainID(account.Reference), nil
}

// ============================================================================
// Helpers
// ============================================================================

// defaultReferenceForFamily returns the default chain reference for a family.
func defaultReferenceForFamily(family chain.Family) string {
	switch family {
	case chain.FamilyEVM:
		return string(chain.ChainIDEthereumMainnet)
	case chain.FamilySolana:
		return string(chain.ChainIDSolanaMainnet)
	case chain.FamilyCosmos:
		return string(chain.ChainIDCosmosHub)
	default:
		return ""
	}
}

// IsPKH returns true if the DID is a did:pkh DID.
func IsPKH(d did.DID) bool {
	return d.Method() == did.MethodPKH
}

// IsEVM returns true if the did:pkh DID is for an EVM chain.
func IsEVM(d did.DID) bool {
	if d.Method() != did.MethodPKH {
		return false
	}

	account, err := ParseDID(d)
	if err != nil {
		return false
	}

	return account.Namespace == NamespaceEIP155
}
