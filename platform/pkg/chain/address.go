package chain

import "strings"

// ============================================================================
// Address
// ============================================================================

// Address represents a validated blockchain address.
type Address struct {
	// raw is the original address as provided.
	raw string

	// normalized is the canonical form of the address.
	// For EVM: checksummed address (EIP-55)
	// For Solana: base58 encoded
	// For Cosmos: bech32 encoded
	normalized string

	// chainID is the chain this address belongs to.
	chainID ChainID

	// family is the blockchain family.
	family Family
}

// NewAddress creates a new Address.
// This is a low-level constructor; prefer using AddressValidator.Validate().
func NewAddress(raw, normalized string, chainID ChainID, family Family) Address {
	return Address{
		raw:        raw,
		normalized: normalized,
		chainID:    chainID,
		family:     family,
	}
}

// Raw returns the original address as provided.
func (a Address) Raw() string {
	return a.raw
}

// String returns the normalized (canonical) form of the address.
func (a Address) String() string {
	return a.normalized
}

// Normalized returns the normalized (canonical) form of the address.
func (a Address) Normalized() string {
	return a.normalized
}

// ChainID returns the chain ID this address belongs to.
func (a Address) ChainID() ChainID {
	return a.chainID
}

// Family returns the blockchain family.
func (a Address) Family() Family {
	return a.family
}

// IsZero returns true if the address is uninitialized.
func (a Address) IsZero() bool {
	return a.normalized == ""
}

// Equals compares two addresses for equality.
// Comparison is case-insensitive for the normalized form.
func (a Address) Equals(other Address) bool {
	if a.family != other.family {
		return false
	}

	// For EVM, compare lowercase since checksum is just encoding
	if a.family == FamilyEVM {
		return strings.EqualFold(a.normalized, other.normalized)
	}

	// For other families, exact match on normalized form
	return a.normalized == other.normalized
}

// EqualsString compares the address to a string.
// Comparison is case-insensitive for EVM addresses.
func (a Address) EqualsString(other string) bool {
	if a.family == FamilyEVM {
		return strings.EqualFold(a.normalized, other)
	}
	return a.normalized == other
}

// ============================================================================
// AddressValidator
// ============================================================================

// AddressValidator validates and normalizes blockchain addresses.
type AddressValidator interface {
	// Validate validates an address for a specific chain.
	// Returns a normalized Address or an error if invalid.
	Validate(address string, chain Chain) (Address, error)

	// ValidateForFamily validates an address for a chain family.
	// Uses default chain ID for the family.
	ValidateForFamily(address string, family Family) (Address, error)

	// IsValid returns true if the address is valid for the chain.
	IsValid(address string, chain Chain) bool

	// Normalize returns the canonical form of an address.
	// Returns empty string if the address is invalid.
	Normalize(address string, family Family) string
}

// ============================================================================
// FamilyAddressValidator
// ============================================================================

// FamilyAddressValidator validates addresses for a specific chain family.
// Implementations exist for each family (EVM, Solana, Cosmos).
type FamilyAddressValidator interface {
	// Family returns the chain family this validator handles.
	Family() Family

	// Validate validates and normalizes an address.
	Validate(address string) (normalized string, err error)

	// IsValid returns true if the address format is valid.
	IsValid(address string) bool

	// Normalize returns the canonical form of an address.
	Normalize(address string) string
}

// ============================================================================
// CompositeAddressValidator
// ============================================================================

// CompositeAddressValidator delegates to family-specific validators.
type CompositeAddressValidator struct {
	validators map[Family]FamilyAddressValidator
}

// NewCompositeAddressValidator creates a new CompositeAddressValidator.
func NewCompositeAddressValidator() *CompositeAddressValidator {
	return &CompositeAddressValidator{
		validators: make(map[Family]FamilyAddressValidator),
	}
}

// Register registers a validator for a chain family.
func (v *CompositeAddressValidator) Register(validator FamilyAddressValidator) {
	v.validators[validator.Family()] = validator
}

// Validate validates an address for a specific chain.
func (v *CompositeAddressValidator) Validate(address string, chain Chain) (Address, error) {
	const op = "CompositeAddressValidator.Validate"

	validator, ok := v.validators[chain.Family]
	if !ok {
		return Address{}, ErrChainNotSupportedFor(op, string(chain.Family))
	}

	normalized, err := validator.Validate(address)
	if err != nil {
		return Address{}, err
	}

	return NewAddress(address, normalized, chain.ID, chain.Family), nil
}

// ValidateForFamily validates an address for a chain family.
func (v *CompositeAddressValidator) ValidateForFamily(address string, family Family) (Address, error) {
	const op = "CompositeAddressValidator.ValidateForFamily"

	validator, ok := v.validators[family]
	if !ok {
		return Address{}, ErrChainNotSupportedFor(op, string(family))
	}

	normalized, err := validator.Validate(address)
	if err != nil {
		return Address{}, err
	}

	// Use empty chain ID for family-only validation
	return NewAddress(address, normalized, "", family), nil
}

// IsValid returns true if the address is valid for the chain.
func (v *CompositeAddressValidator) IsValid(address string, chain Chain) bool {
	validator, ok := v.validators[chain.Family]
	if !ok {
		return false
	}
	return validator.IsValid(address)
}

// Normalize returns the canonical form of an address.
func (v *CompositeAddressValidator) Normalize(address string, family Family) string {
	validator, ok := v.validators[family]
	if !ok {
		return ""
	}
	return validator.Normalize(address)
}

// ============================================================================
// Interface Compliance
// ============================================================================

var _ AddressValidator = (*CompositeAddressValidator)(nil)
