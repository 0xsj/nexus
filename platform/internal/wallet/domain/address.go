package domain

import (
	"strings"

	"github.com/0xsj/nexus/platform/pkg/chain"
	"github.com/0xsj/nexus/platform/pkg/did"
	"github.com/0xsj/nexus/platform/pkg/did/pkh"
)

// ============================================================================
// Address
// ============================================================================

// Address represents a validated blockchain wallet address.
type Address struct {
	// raw is the original address string.
	raw string

	// normalized is the normalized/checksummed address.
	normalized string

	// chainID is the chain this address belongs to.
	chainID ChainID

	// family is the chain family (EVM, Solana, Cosmos).
	family ChainFamily
}

// ============================================================================
// Constructors
// ============================================================================

// NewAddress creates a new Address with validation.
func NewAddress(raw string, chainID ChainID) (Address, error) {
	const op = "Address.New"

	if raw == "" {
		return Address{}, ErrInvalidAddress(op, raw, "address cannot be empty")
	}

	if chainID.IsEmpty() {
		return Address{}, ErrInvalidChain(op, "")
	}

	chainInfo := GetChainInfo(chainID)
	if chainInfo.IsZero() {
		return Address{}, ErrChainNotSupported(op, chainID.String())
	}

	// Normalize based on chain family
	normalized, err := normalizeAddress(raw, chainInfo.Family)
	if err != nil {
		return Address{}, err
	}

	return Address{
		raw:        raw,
		normalized: normalized,
		chainID:    chainID,
		family:     chainInfo.Family,
	}, nil
}

// NewAddressUnchecked creates an Address without validation.
// Use only when the address is known to be valid (e.g., from database).
func NewAddressUnchecked(raw, normalized string, chainID ChainID, family ChainFamily) Address {
	return Address{
		raw:        raw,
		normalized: normalized,
		chainID:    chainID,
		family:     family,
	}
}

// ============================================================================
// Getters
// ============================================================================

// Raw returns the original address string.
func (a Address) Raw() string {
	return a.raw
}

// Normalized returns the normalized address string.
// For EVM: checksummed address
// For Solana: base58 address
// For Cosmos: bech32 address
func (a Address) Normalized() string {
	return a.normalized
}

// String returns the normalized address string.
func (a Address) String() string {
	return a.normalized
}

// ChainID returns the chain ID for this address.
func (a Address) ChainID() ChainID {
	return a.chainID
}

// Family returns the chain family for this address.
func (a Address) Family() ChainFamily {
	return a.family
}

// ============================================================================
// Predicates
// ============================================================================

// IsZero returns true if the address is empty/uninitialized.
func (a Address) IsZero() bool {
	return a.normalized == ""
}

// IsEVM returns true if this is an EVM address.
func (a Address) IsEVM() bool {
	return a.family == ChainFamilyEVM
}

// IsSolana returns true if this is a Solana address.
func (a Address) IsSolana() bool {
	return a.family == ChainFamilySolana
}

// IsCosmos returns true if this is a Cosmos address.
func (a Address) IsCosmos() bool {
	return a.family == ChainFamilyCosmos
}

// ============================================================================
// Comparison
// ============================================================================

// Equals checks if two addresses are equal.
// Compares normalized addresses and chain IDs.
func (a Address) Equals(other Address) bool {
	if a.family != other.family {
		return false
	}

	// For EVM, compare case-insensitive
	if a.IsEVM() {
		return strings.EqualFold(a.normalized, other.normalized) && a.chainID == other.chainID
	}

	return a.normalized == other.normalized && a.chainID == other.chainID
}

// EqualsString checks if the address equals a raw string.
func (a Address) EqualsString(raw string) bool {
	if a.IsEVM() {
		return strings.EqualFold(a.normalized, raw)
	}
	return a.normalized == raw
}

// SameAddress checks if two addresses are the same (ignoring chain).
// Useful for checking if the same address is used across chains.
func (a Address) SameAddress(other Address) bool {
	if a.family != other.family {
		return false
	}

	if a.IsEVM() {
		return strings.EqualFold(a.normalized, other.normalized)
	}

	return a.normalized == other.normalized
}

// ============================================================================
// DID Derivation
// ============================================================================

// ToDID derives a did:pkh DID from this address.
func (a Address) ToDID() (did.DID, error) {
	const op = "Address.ToDID"

	if a.IsZero() {
		return did.DID{}, ErrInvalidAddress(op, "", "cannot derive DID from empty address")
	}

	// Convert to pkg/chain.Address for pkh.Generate
	pkgAddress := a.ToPkgAddress()

	return pkh.Generate(pkgAddress)
}

// ToDIDString derives a did:pkh DID string from this address.
func (a Address) ToDIDString() string {
	d, err := a.ToDID()
	if err != nil {
		return ""
	}
	return d.String()
}

// ============================================================================
// Conversion
// ============================================================================

// ToPkgAddress converts to pkg/chain.Address.
func (a Address) ToPkgAddress() chain.Address {
	return chain.NewAddress(a.raw, a.normalized, a.chainID.ToPkg(), a.family.ToPkg())
}

// AddressFromPkg converts from pkg/chain.Address.
func AddressFromPkg(addr chain.Address) Address {
	return Address{
		raw:        addr.Raw(),
		normalized: addr.Normalized(),
		chainID:    ChainIDFromPkg(addr.ChainID()),
		family:     ChainFamilyFromPkg(addr.Family()),
	}
}

// ============================================================================
// Normalization
// ============================================================================

// normalizeAddress normalizes an address based on chain family.
func normalizeAddress(raw string, family ChainFamily) (string, error) {
	switch family {
	case ChainFamilyEVM:
		return normalizeEVMAddress(raw)
	case ChainFamilySolana:
		return normalizeSolanaAddress(raw)
	case ChainFamilyCosmos:
		return normalizeCosmosAddress(raw)
	default:
		return "", ErrChainNotSupported("normalizeAddress", string(family))
	}
}

// normalizeEVMAddress normalizes an EVM address.
// Validates format and returns checksummed address.
func normalizeEVMAddress(raw string) (string, error) {
	const op = "normalizeEVMAddress"

	// Remove whitespace
	raw = strings.TrimSpace(raw)

	// Must start with 0x
	if !strings.HasPrefix(raw, "0x") && !strings.HasPrefix(raw, "0X") {
		return "", ErrInvalidAddress(op, raw, "must start with 0x")
	}

	// Must be 42 characters (0x + 40 hex chars)
	if len(raw) != 42 {
		return "", ErrInvalidAddress(op, raw, "must be 42 characters")
	}

	// Validate hex characters
	for _, c := range raw[2:] {
		if !isHexChar(c) {
			return "", ErrInvalidAddress(op, raw, "invalid hex character")
		}
	}

	// Return lowercase for now (checksumming can be added later)
	return strings.ToLower(raw), nil
}

// normalizeSolanaAddress normalizes a Solana address.
// Validates base58 format.
func normalizeSolanaAddress(raw string) (string, error) {
	const op = "normalizeSolanaAddress"

	// Remove whitespace
	raw = strings.TrimSpace(raw)

	// Solana addresses are 32-44 characters base58
	if len(raw) < 32 || len(raw) > 44 {
		return "", ErrInvalidAddress(op, raw, "invalid length")
	}

	// Validate base58 characters
	for _, c := range raw {
		if !isBase58Char(c) {
			return "", ErrInvalidAddress(op, raw, "invalid base58 character")
		}
	}

	return raw, nil
}

// normalizeCosmosAddress normalizes a Cosmos address.
// Validates bech32 format.
func normalizeCosmosAddress(raw string) (string, error) {
	const op = "normalizeCosmosAddress"

	// Remove whitespace
	raw = strings.TrimSpace(raw)

	// Cosmos addresses are bech32 encoded
	// Must contain separator '1'
	if !strings.Contains(raw, "1") {
		return "", ErrInvalidAddress(op, raw, "invalid bech32 format")
	}

	// Basic length check (prefix + 1 + data)
	if len(raw) < 10 {
		return "", ErrInvalidAddress(op, raw, "too short")
	}

	// Bech32 is lowercase
	return strings.ToLower(raw), nil
}

// ============================================================================
// Helpers
// ============================================================================

// isHexChar returns true if the rune is a valid hex character.
func isHexChar(c rune) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
}

// isBase58Char returns true if the rune is a valid base58 character.
// Base58 excludes 0, O, I, l to avoid ambiguity.
func isBase58Char(c rune) bool {
	return (c >= '1' && c <= '9') ||
		(c >= 'A' && c <= 'H') ||
		(c >= 'J' && c <= 'N') ||
		(c >= 'P' && c <= 'Z') ||
		(c >= 'a' && c <= 'k') ||
		(c >= 'm' && c <= 'z')
}
