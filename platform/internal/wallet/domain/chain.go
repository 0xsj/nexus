package domain

import (
	"fmt"
)

// Chain represents a blockchain network with its chain ID, name, and namespace.
// Uses CAIP-2 chain identification (e.g., "eip155:1" for Ethereum mainnet).
type Chain struct {
	id        int
	name      string
	namespace string
}

// NewChain creates a new Chain with the given parameters.
// Validates that the chain ID is positive and the name is non-empty.
func NewChain(id int, name, namespace string) (Chain, error) {
	if id <= 0 {
		return Chain{}, ChainNotSupported("Chain.New", id)
	}

	if name == "" {
		return Chain{}, ChainNotSupported("Chain.New", id)
	}

	return Chain{
		id:        id,
		name:      name,
		namespace: namespace,
	}, nil
}

// ============================================================================
// Predefined Chains
// ============================================================================

// ChainEthereum returns the Ethereum mainnet chain (eip155:1).
func ChainEthereum() Chain {
	return Chain{
		id:        1,
		name:      "Ethereum",
		namespace: "eip155",
	}
}

// ChainPolygon returns the Polygon mainnet chain (eip155:137).
func ChainPolygon() Chain {
	return Chain{
		id:        137,
		name:      "Polygon",
		namespace: "eip155",
	}
}

// ============================================================================
// Getters
// ============================================================================

// ID returns the chain ID.
func (c Chain) ID() int {
	return c.id
}

// Name returns the chain name.
func (c Chain) Name() string {
	return c.name
}

// Namespace returns the chain namespace (e.g., "eip155").
func (c Chain) Namespace() string {
	return c.namespace
}

// ============================================================================
// Query Methods
// ============================================================================

// IsEVM returns true if this chain uses the EVM (eip155 namespace).
func (c Chain) IsEVM() bool {
	return c.namespace == "eip155"
}

// String returns the CAIP-2 representation of the chain (e.g., "eip155:1").
func (c Chain) String() string {
	return fmt.Sprintf("%s:%d", c.namespace, c.id)
}

// IsZero returns true if the chain is the zero value.
func (c Chain) IsZero() bool {
	return c.id == 0
}

// Equals checks if two chains are equal.
func (c Chain) Equals(other Chain) bool {
	return c.id == other.id && c.namespace == other.namespace
}
