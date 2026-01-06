package chain

import "sync"

// ============================================================================
// ChainRegistry Interface
// ============================================================================

// ChainRegistry provides chain configuration and lookup.
type ChainRegistry interface {
	// Get retrieves a chain by ID.
	// Returns ErrChainNotFound if the chain is not registered.
	Get(chainID ChainID) (Chain, error)

	// GetByName retrieves a chain by name.
	// Returns ErrChainNotFound if the chain is not registered.
	GetByName(name string) (Chain, error)

	// List returns all registered chains.
	List() []Chain

	// ListByFamily returns all chains for a specific family.
	ListByFamily(family Family) []Chain

	// Has returns true if the chain is registered.
	Has(chainID ChainID) bool

	// Register adds a chain to the registry.
	// Returns ErrConflict if a chain with the same ID already exists.
	Register(chain Chain) error

	// Unregister removes a chain from the registry.
	Unregister(chainID ChainID) error
}

// ============================================================================
// InMemoryChainRegistry
// ============================================================================

// InMemoryChainRegistry is a thread-safe in-memory chain registry.
type InMemoryChainRegistry struct {
	mu       sync.RWMutex
	chains   map[ChainID]Chain
	byName   map[string]ChainID
	byFamily map[Family][]ChainID
}

// NewChainRegistry creates a new empty chain registry.
func NewChainRegistry() *InMemoryChainRegistry {
	return &InMemoryChainRegistry{
		chains:   make(map[ChainID]Chain),
		byName:   make(map[string]ChainID),
		byFamily: make(map[Family][]ChainID),
	}
}

// NewChainRegistryWithDefaults creates a chain registry with default EVM chains.
func NewChainRegistryWithDefaults() *InMemoryChainRegistry {
	r := NewChainRegistry()
	for _, c := range DefaultEVMChains() {
		_ = r.Register(c) // Safe to ignore error for known-good chains
	}
	return r
}

// Get retrieves a chain by ID.
func (r *InMemoryChainRegistry) Get(chainID ChainID) (Chain, error) {
	const op = "ChainRegistry.Get"

	r.mu.RLock()
	defer r.mu.RUnlock()

	chain, ok := r.chains[chainID]
	if !ok {
		return Chain{}, ErrChainNotFoundFor(op, string(chainID))
	}

	return chain, nil
}

// GetByName retrieves a chain by name.
func (r *InMemoryChainRegistry) GetByName(name string) (Chain, error) {
	const op = "ChainRegistry.GetByName"

	r.mu.RLock()
	defer r.mu.RUnlock()

	chainID, ok := r.byName[name]
	if !ok {
		return Chain{}, ErrChainNotFoundFor(op, name)
	}

	return r.chains[chainID], nil
}

// List returns all registered chains.
func (r *InMemoryChainRegistry) List() []Chain {
	r.mu.RLock()
	defer r.mu.RUnlock()

	chains := make([]Chain, 0, len(r.chains))
	for _, c := range r.chains {
		chains = append(chains, c)
	}

	return chains
}

// ListByFamily returns all chains for a specific family.
func (r *InMemoryChainRegistry) ListByFamily(family Family) []Chain {
	r.mu.RLock()
	defer r.mu.RUnlock()

	chainIDs, ok := r.byFamily[family]
	if !ok {
		return nil
	}

	chains := make([]Chain, 0, len(chainIDs))
	for _, id := range chainIDs {
		if c, exists := r.chains[id]; exists {
			chains = append(chains, c)
		}
	}

	return chains
}

// Has returns true if the chain is registered.
func (r *InMemoryChainRegistry) Has(chainID ChainID) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.chains[chainID]
	return ok
}

// Register adds a chain to the registry.
func (r *InMemoryChainRegistry) Register(chain Chain) error {
	const op = "ChainRegistry.Register"

	if err := chain.Validate(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Check for existing chain
	if _, exists := r.chains[chain.ID]; exists {
		return ErrChainNotSupportedFor(op, string(chain.ID)).
			WithMessage("chain already registered: " + string(chain.ID))
	}

	// Register chain
	r.chains[chain.ID] = chain
	r.byName[chain.Name] = chain.ID

	// Update family index
	r.byFamily[chain.Family] = append(r.byFamily[chain.Family], chain.ID)

	return nil
}

// Unregister removes a chain from the registry.
func (r *InMemoryChainRegistry) Unregister(chainID ChainID) error {
	const op = "ChainRegistry.Unregister"

	r.mu.Lock()
	defer r.mu.Unlock()

	chain, exists := r.chains[chainID]
	if !exists {
		return ErrChainNotFoundFor(op, string(chainID))
	}

	// Remove from main map
	delete(r.chains, chainID)
	delete(r.byName, chain.Name)

	// Remove from family index
	familyChains := r.byFamily[chain.Family]
	for i, id := range familyChains {
		if id == chainID {
			r.byFamily[chain.Family] = append(familyChains[:i], familyChains[i+1:]...)
			break
		}
	}

	return nil
}

// ============================================================================
// Interface Compliance
// ============================================================================

var _ ChainRegistry = (*InMemoryChainRegistry)(nil)
