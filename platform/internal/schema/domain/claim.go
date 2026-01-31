package domain

import (
	"fmt"
	"regexp"
)

// ClaimDefinition represents a single claim within a schema.
// It defines the claim's key, type, validation rules, and metadata.
type ClaimDefinition struct {
	key         string
	claimType   ClaimType
	required    bool
	displayName string
	description string
	disclosable bool
	order       int
}

// ClaimDefinitionOption is a functional option for configuring a ClaimDefinition.
type ClaimDefinitionOption func(*ClaimDefinition)

// ============================================================================
// Validation Constants
// ============================================================================

const (
	// ClaimKeyMaxLength is the maximum length of a claim key.
	ClaimKeyMaxLength = 64

	// ClaimKeyMinLength is the minimum length of a claim key.
	ClaimKeyMinLength = 1

	// ClaimDisplayNameMaxLength is the maximum length of a display name.
	ClaimDisplayNameMaxLength = 128

	// ClaimDescriptionMaxLength is the maximum length of a description.
	ClaimDescriptionMaxLength = 512
)

// claimKeyPattern defines valid claim key format: lowercase alphanumeric and underscores.
var claimKeyPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// ============================================================================
// Constructors
// ============================================================================

// NewClaimDefinition creates a new ClaimDefinition with the given key, type, and options.
func NewClaimDefinition(key string, claimType ClaimType, opts ...ClaimDefinitionOption) (ClaimDefinition, error) {
	cd := ClaimDefinition{
		key:         key,
		claimType:   claimType,
		required:    false,
		displayName: "",
		description: "",
		disclosable: true, // Default to disclosable
		order:       0,
	}

	for _, opt := range opts {
		opt(&cd)
	}

	if err := cd.Validate(); err != nil {
		return ClaimDefinition{}, err
	}

	return cd, nil
}

// MustNewClaimDefinition creates a new ClaimDefinition and panics if invalid.
// Only use for constants or tests.
func MustNewClaimDefinition(key string, claimType ClaimType, opts ...ClaimDefinitionOption) ClaimDefinition {
	cd, err := NewClaimDefinition(key, claimType, opts...)
	if err != nil {
		panic(fmt.Sprintf("invalid claim definition: %v", err))
	}
	return cd
}

// ============================================================================
// Functional Options
// ============================================================================

// Required marks the claim as required.
func Required() ClaimDefinitionOption {
	return func(cd *ClaimDefinition) {
		cd.required = true
	}
}

// Optional marks the claim as optional (default).
func Optional() ClaimDefinitionOption {
	return func(cd *ClaimDefinition) {
		cd.required = false
	}
}

// WithDisplayName sets the human-readable display name.
func WithDisplayName(name string) ClaimDefinitionOption {
	return func(cd *ClaimDefinition) {
		cd.displayName = name
	}
}

// WithDescription sets the claim description.
func WithDescription(desc string) ClaimDefinitionOption {
	return func(cd *ClaimDefinition) {
		cd.description = desc
	}
}

// Disclosable marks the claim as selectively disclosable (default).
func Disclosable() ClaimDefinitionOption {
	return func(cd *ClaimDefinition) {
		cd.disclosable = true
	}
}

// NonDisclosable marks the claim as not selectively disclosable.
// The claim must be disclosed in full or not at all.
func NonDisclosable() ClaimDefinitionOption {
	return func(cd *ClaimDefinition) {
		cd.disclosable = false
	}
}

// WithOrder sets the display order for the claim.
func WithOrder(order int) ClaimDefinitionOption {
	return func(cd *ClaimDefinition) {
		cd.order = order
	}
}

// ============================================================================
// Accessors
// ============================================================================

// Key returns the claim key.
func (cd ClaimDefinition) Key() string {
	return cd.key
}

// ClaimType returns the claim type configuration.
func (cd ClaimDefinition) ClaimType() ClaimType {
	return cd.claimType
}

// DataType returns the underlying data type (convenience accessor).
func (cd ClaimDefinition) DataType() DataType {
	return cd.claimType.DataType()
}

// IsRequired returns true if the claim is required.
func (cd ClaimDefinition) IsRequired() bool {
	return cd.required
}

// IsOptional returns true if the claim is optional.
func (cd ClaimDefinition) IsOptional() bool {
	return !cd.required
}

// DisplayName returns the human-readable display name.
// If not set, returns the key.
func (cd ClaimDefinition) DisplayName() string {
	if cd.displayName == "" {
		return cd.key
	}
	return cd.displayName
}

// Description returns the claim description.
func (cd ClaimDefinition) Description() string {
	return cd.description
}

// IsDisclosable returns true if the claim can be selectively disclosed.
func (cd ClaimDefinition) IsDisclosable() bool {
	return cd.disclosable
}

// Order returns the display order.
func (cd ClaimDefinition) Order() int {
	return cd.order
}

// ============================================================================
// Validation
// ============================================================================

// Validate checks that the claim definition is valid.
func (cd ClaimDefinition) Validate() error {
	// Validate key
	if err := cd.validateKey(); err != nil {
		return err
	}

	// Validate claim type
	if cd.claimType.IsZero() {
		return ErrInvalidClaimDefinition
	}

	// Validate display name length
	if len(cd.displayName) > ClaimDisplayNameMaxLength {
		return ErrInvalidClaimDefinition
	}

	// Validate description length
	if len(cd.description) > ClaimDescriptionMaxLength {
		return ErrInvalidClaimDefinition
	}

	// Validate order is non-negative
	if cd.order < 0 {
		return ErrInvalidClaimDefinition
	}

	return nil
}

// validateKey checks that the key is valid.
func (cd ClaimDefinition) validateKey() error {
	if cd.key == "" {
		return ErrInvalidClaimDefinition
	}

	if len(cd.key) < ClaimKeyMinLength || len(cd.key) > ClaimKeyMaxLength {
		return ErrInvalidClaimDefinition
	}

	if !claimKeyPattern.MatchString(cd.key) {
		return ErrInvalidClaimDefinition
	}

	return nil
}

// ============================================================================
// Comparison
// ============================================================================

// IsZero returns true if the claim definition is the zero value.
func (cd ClaimDefinition) IsZero() bool {
	return cd.key == ""
}

// Equals returns true if two claim definitions are equal.
func (cd ClaimDefinition) Equals(other ClaimDefinition) bool {
	return cd.key == other.key &&
		cd.claimType.Equals(other.claimType) &&
		cd.required == other.required &&
		cd.displayName == other.displayName &&
		cd.description == other.description &&
		cd.disclosable == other.disclosable &&
		cd.order == other.order
}

// KeyEquals returns true if the keys are equal.
func (cd ClaimDefinition) KeyEquals(other ClaimDefinition) bool {
	return cd.key == other.key
}

// ============================================================================
// String Representation
// ============================================================================

// String returns a string representation of the claim definition.
func (cd ClaimDefinition) String() string {
	requiredStr := "optional"
	if cd.required {
		requiredStr = "required"
	}
	return fmt.Sprintf("%s: %s (%s)", cd.key, cd.claimType, requiredStr)
}

// GoString implements fmt.GoStringer for debugging.
func (cd ClaimDefinition) GoString() string {
	return fmt.Sprintf("ClaimDefinition{key=%q, type=%s, required=%t, disclosable=%t}",
		cd.key, cd.claimType, cd.required, cd.disclosable)
}

// ============================================================================
// Builder Pattern (Alternative)
// ============================================================================

// ClaimDefinitionBuilder provides a fluent builder for ClaimDefinition.
type ClaimDefinitionBuilder struct {
	key         string
	claimType   ClaimType
	required    bool
	displayName string
	description string
	disclosable bool
	order       int
}

// NewClaimDefinitionBuilder creates a new builder with required fields.
func NewClaimDefinitionBuilder(key string, claimType ClaimType) *ClaimDefinitionBuilder {
	return &ClaimDefinitionBuilder{
		key:         key,
		claimType:   claimType,
		required:    false,
		disclosable: true,
		order:       0,
	}
}

// Required marks the claim as required.
func (b *ClaimDefinitionBuilder) Required() *ClaimDefinitionBuilder {
	b.required = true
	return b
}

// Optional marks the claim as optional.
func (b *ClaimDefinitionBuilder) Optional() *ClaimDefinitionBuilder {
	b.required = false
	return b
}

// DisplayName sets the display name.
func (b *ClaimDefinitionBuilder) DisplayName(name string) *ClaimDefinitionBuilder {
	b.displayName = name
	return b
}

// Description sets the description.
func (b *ClaimDefinitionBuilder) Description(desc string) *ClaimDefinitionBuilder {
	b.description = desc
	return b
}

// Disclosable marks the claim as disclosable.
func (b *ClaimDefinitionBuilder) Disclosable() *ClaimDefinitionBuilder {
	b.disclosable = true
	return b
}

// NonDisclosable marks the claim as non-disclosable.
func (b *ClaimDefinitionBuilder) NonDisclosable() *ClaimDefinitionBuilder {
	b.disclosable = false
	return b
}

// Order sets the display order.
func (b *ClaimDefinitionBuilder) Order(order int) *ClaimDefinitionBuilder {
	b.order = order
	return b
}

// Build creates the ClaimDefinition, returning an error if invalid.
func (b *ClaimDefinitionBuilder) Build() (ClaimDefinition, error) {
	return NewClaimDefinition(
		b.key,
		b.claimType,
		func(cd *ClaimDefinition) {
			cd.required = b.required
			cd.displayName = b.displayName
			cd.description = b.description
			cd.disclosable = b.disclosable
			cd.order = b.order
		},
	)
}

// MustBuild creates the ClaimDefinition and panics if invalid.
func (b *ClaimDefinitionBuilder) MustBuild() ClaimDefinition {
	cd, err := b.Build()
	if err != nil {
		panic(fmt.Sprintf("invalid claim definition: %v", err))
	}
	return cd
}
