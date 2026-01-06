package validation

import (
	"context"

	"github.com/0xsj/nexus/platform/internal/wallet/domain"
	"github.com/0xsj/nexus/platform/pkg/chain"
)

// ============================================================================
// Composite Address Validator
// ============================================================================

// CompositeAddressValidator implements domain.AddressValidationService.
// It delegates to family-specific validators based on chain.
type CompositeAddressValidator struct {
	validators map[domain.ChainFamily]chain.FamilyAddressValidator
}

// NewCompositeAddressValidator creates a new composite address validator.
func NewCompositeAddressValidator() *CompositeAddressValidator {
	return &CompositeAddressValidator{
		validators: make(map[domain.ChainFamily]chain.FamilyAddressValidator),
	}
}

// Register registers a validator for a chain family.
func (v *CompositeAddressValidator) Register(validator chain.FamilyAddressValidator) {
	family := domain.ChainFamilyFromPkg(validator.Family())
	v.validators[family] = validator
}

// Validate validates an address for a specific chain.
func (v *CompositeAddressValidator) Validate(ctx context.Context, address string, chainID domain.ChainID) (domain.Address, error) {
	const op = "CompositeAddressValidator.Validate"

	// Get chain info
	chainInfo := domain.GetChainInfo(chainID)
	if chainInfo.IsZero() {
		return domain.Address{}, domain.ErrChainNotSupported(op, chainID.String())
	}

	// Get validator for family
	validator, ok := v.validators[chainInfo.Family]
	if !ok {
		return domain.Address{}, domain.ErrChainNotSupported(op, chainInfo.Family.String())
	}

	// Check context
	select {
	case <-ctx.Done():
		return domain.Address{}, ctx.Err()
	default:
	}

	// Validate and normalize
	normalized, err := validator.Validate(address)
	if err != nil {
		return domain.Address{}, domain.ErrInvalidAddress(op, address, err.Error())
	}

	// Create domain address
	return domain.NewAddressUnchecked(address, normalized, chainID, chainInfo.Family), nil
}

// ValidateForFamily validates an address for a chain family.
func (v *CompositeAddressValidator) ValidateForFamily(ctx context.Context, address string, family domain.ChainFamily) (domain.Address, error) {
	const op = "CompositeAddressValidator.ValidateForFamily"

	// Get validator for family
	validator, ok := v.validators[family]
	if !ok {
		return domain.Address{}, domain.ErrChainNotSupported(op, family.String())
	}

	// Check context
	select {
	case <-ctx.Done():
		return domain.Address{}, ctx.Err()
	default:
	}

	// Validate and normalize
	normalized, err := validator.Validate(address)
	if err != nil {
		return domain.Address{}, domain.ErrInvalidAddress(op, address, err.Error())
	}

	// Use default chain ID for family
	chainID := defaultChainIDForFamily(family)

	// Create domain address
	return domain.NewAddressUnchecked(address, normalized, chainID, family), nil
}

// IsValid returns true if the address is valid for the chain.
func (v *CompositeAddressValidator) IsValid(ctx context.Context, address string, chainID domain.ChainID) bool {
	_, err := v.Validate(ctx, address, chainID)
	return err == nil
}

// Normalize returns the normalized form of an address.
func (v *CompositeAddressValidator) Normalize(ctx context.Context, address string, family domain.ChainFamily) string {
	validator, ok := v.validators[family]
	if !ok {
		return ""
	}
	return validator.Normalize(address)
}

// ============================================================================
// Default Validator Factory
// ============================================================================

// NewDefaultAddressValidator creates a validator with all supported chain families.
func NewDefaultAddressValidator() *CompositeAddressValidator {
	v := NewCompositeAddressValidator()

	// Register EVM validator
	v.Register(NewEVMAddressValidator())

	// Register Solana validator (when implemented)
	// v.Register(NewSolanaAddressValidator())

	// Register Cosmos validator (when implemented)
	// v.Register(NewCosmosAddressValidator())

	return v
}

// ============================================================================
// Helpers
// ============================================================================

// defaultChainIDForFamily returns the default chain ID for a family.
func defaultChainIDForFamily(family domain.ChainFamily) domain.ChainID {
	switch family {
	case domain.ChainFamilyEVM:
		return domain.ChainIDEthereumMainnet
	case domain.ChainFamilySolana:
		return domain.ChainIDSolanaMainnet
	case domain.ChainFamilyCosmos:
		return domain.ChainIDCosmosHub
	default:
		return domain.ChainIDEthereumMainnet
	}
}

// ============================================================================
// Interface Compliance
// ============================================================================

var _ domain.AddressValidationService = (*CompositeAddressValidator)(nil)
