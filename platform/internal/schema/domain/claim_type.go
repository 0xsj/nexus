package domain

import (
	"fmt"
	"regexp"
)

// ClaimType represents the type configuration for a claim.
// It combines a data type with optional constraints for validation.
type ClaimType struct {
	dataType    DataType
	constraints ClaimConstraints
}

// ClaimConstraints holds optional validation constraints for a claim type.
type ClaimConstraints struct {
	// AllowedValues specifies permitted values for Enum types.
	AllowedValues []string

	// Min is the minimum value for Integer types.
	Min *int64

	// Max is the maximum value for Integer types.
	Max *int64

	// MinLength is the minimum length for String and StringArray types.
	MinLength *int

	// MaxLength is the maximum length for String and StringArray types.
	MaxLength *int

	// Pattern is a regex pattern for String-based types.
	Pattern *string

	// Format specifies the expected format for Date/DateTime types.
	// Example: "2006-01-02" for dates.
	Format *string
}

// ClaimTypeOption is a functional option for configuring a ClaimType.
type ClaimTypeOption func(*ClaimType)

// ============================================================================
// Constructors
// ============================================================================

// NewClaimType creates a new ClaimType with the given data type and options.
func NewClaimType(dataType DataType, opts ...ClaimTypeOption) (ClaimType, error) {
	if !dataType.IsValid() {
		return ClaimType{}, ErrInvalidDataType
	}

	ct := ClaimType{
		dataType:    dataType,
		constraints: ClaimConstraints{},
	}

	for _, opt := range opts {
		opt(&ct)
	}

	if err := ct.Validate(); err != nil {
		return ClaimType{}, err
	}

	return ct, nil
}

// MustNewClaimType creates a new ClaimType and panics if invalid.
// Only use for constants or tests.
func MustNewClaimType(dataType DataType, opts ...ClaimTypeOption) ClaimType {
	ct, err := NewClaimType(dataType, opts...)
	if err != nil {
		panic(fmt.Sprintf("invalid claim type: %v", err))
	}
	return ct
}

// ============================================================================
// Functional Options
// ============================================================================

// WithAllowedValues sets the allowed values for Enum types.
func WithAllowedValues(values ...string) ClaimTypeOption {
	return func(ct *ClaimType) {
		ct.constraints.AllowedValues = values
	}
}

// WithMin sets the minimum value for Integer types.
func WithMin(min int64) ClaimTypeOption {
	return func(ct *ClaimType) {
		ct.constraints.Min = &min
	}
}

// WithMax sets the maximum value for Integer types.
func WithMax(max int64) ClaimTypeOption {
	return func(ct *ClaimType) {
		ct.constraints.Max = &max
	}
}

// WithMinMax sets both minimum and maximum values for Integer types.
func WithMinMax(min, max int64) ClaimTypeOption {
	return func(ct *ClaimType) {
		ct.constraints.Min = &min
		ct.constraints.Max = &max
	}
}

// WithMinLength sets the minimum length for String and StringArray types.
func WithMinLength(min int) ClaimTypeOption {
	return func(ct *ClaimType) {
		ct.constraints.MinLength = &min
	}
}

// WithMaxLength sets the maximum length for String and StringArray types.
func WithMaxLength(max int) ClaimTypeOption {
	return func(ct *ClaimType) {
		ct.constraints.MaxLength = &max
	}
}

// WithLengthRange sets both minimum and maximum length for String and StringArray types.
func WithLengthRange(min, max int) ClaimTypeOption {
	return func(ct *ClaimType) {
		ct.constraints.MinLength = &min
		ct.constraints.MaxLength = &max
	}
}

// WithPattern sets the regex pattern for String-based types.
func WithPattern(pattern string) ClaimTypeOption {
	return func(ct *ClaimType) {
		ct.constraints.Pattern = &pattern
	}
}

// WithFormat sets the format for Date/DateTime types.
func WithFormat(format string) ClaimTypeOption {
	return func(ct *ClaimType) {
		ct.constraints.Format = &format
	}
}

// ============================================================================
// Accessors
// ============================================================================

// DataType returns the underlying data type.
func (ct ClaimType) DataType() DataType {
	return ct.dataType
}

// Constraints returns the claim constraints.
func (ct ClaimType) Constraints() ClaimConstraints {
	return ct.constraints
}

// HasConstraints returns true if any constraints are set.
func (ct ClaimType) HasConstraints() bool {
	c := ct.constraints
	return len(c.AllowedValues) > 0 ||
		c.Min != nil ||
		c.Max != nil ||
		c.MinLength != nil ||
		c.MaxLength != nil ||
		c.Pattern != nil ||
		c.Format != nil
}

// ============================================================================
// Validation
// ============================================================================

// Validate checks that the constraints are appropriate for the data type.
func (ct ClaimType) Validate() error {
	c := ct.constraints

	// Enum requires allowed values
	if ct.dataType == DataTypeEnum && len(c.AllowedValues) == 0 {
		return ErrInvalidClaimDefinition
	}

	// AllowedValues only valid for Enum
	if len(c.AllowedValues) > 0 && ct.dataType != DataTypeEnum {
		return ErrInvalidClaimDefinition
	}

	// Min/Max only valid for Integer
	if (c.Min != nil || c.Max != nil) && ct.dataType != DataTypeInteger {
		return ErrInvalidClaimDefinition
	}

	// Min must be <= Max
	if c.Min != nil && c.Max != nil && *c.Min > *c.Max {
		return ErrInvalidClaimDefinition
	}

	// MinLength/MaxLength only valid for String-based types and StringArray
	if c.MinLength != nil || c.MaxLength != nil {
		if !ct.isLengthConstraintAllowed() {
			return ErrInvalidClaimDefinition
		}
	}

	// MinLength must be >= 0
	if c.MinLength != nil && *c.MinLength < 0 {
		return ErrInvalidClaimDefinition
	}

	// MaxLength must be >= 0
	if c.MaxLength != nil && *c.MaxLength < 0 {
		return ErrInvalidClaimDefinition
	}

	// MinLength must be <= MaxLength
	if c.MinLength != nil && c.MaxLength != nil && *c.MinLength > *c.MaxLength {
		return ErrInvalidClaimDefinition
	}

	// Pattern only valid for String-based types
	if c.Pattern != nil {
		if !ct.isPatternConstraintAllowed() {
			return ErrInvalidClaimDefinition
		}
		// Validate regex compiles
		if _, err := regexp.Compile(*c.Pattern); err != nil {
			return ErrInvalidClaimDefinition
		}
	}

	// Format only valid for Date/DateTime
	if c.Format != nil && !ct.dataType.IsTemporal() {
		return ErrInvalidClaimDefinition
	}

	return nil
}

// isLengthConstraintAllowed returns true if length constraints are valid for this type.
func (ct ClaimType) isLengthConstraintAllowed() bool {
	switch ct.dataType {
	case DataTypeString, DataTypeStringArray, DataTypeEmail, DataTypeURL, DataTypeDID:
		return true
	default:
		return false
	}
}

// isPatternConstraintAllowed returns true if pattern constraints are valid for this type.
func (ct ClaimType) isPatternConstraintAllowed() bool {
	switch ct.dataType {
	case DataTypeString, DataTypeEmail, DataTypeURL, DataTypeDID:
		return true
	default:
		return false
	}
}

// ============================================================================
// Comparison
// ============================================================================

// IsZero returns true if the claim type is the zero value.
func (ct ClaimType) IsZero() bool {
	return ct.dataType.IsZero()
}

// Equals returns true if two claim types are equal.
func (ct ClaimType) Equals(other ClaimType) bool {
	if ct.dataType != other.dataType {
		return false
	}

	c1 := ct.constraints
	c2 := other.constraints

	// Compare AllowedValues
	if !stringSliceEquals(c1.AllowedValues, c2.AllowedValues) {
		return false
	}

	// Compare Min
	if !int64PtrEquals(c1.Min, c2.Min) {
		return false
	}

	// Compare Max
	if !int64PtrEquals(c1.Max, c2.Max) {
		return false
	}

	// Compare MinLength
	if !intPtrEquals(c1.MinLength, c2.MinLength) {
		return false
	}

	// Compare MaxLength
	if !intPtrEquals(c1.MaxLength, c2.MaxLength) {
		return false
	}

	// Compare Pattern
	if !stringPtrEquals(c1.Pattern, c2.Pattern) {
		return false
	}

	// Compare Format
	if !stringPtrEquals(c1.Format, c2.Format) {
		return false
	}

	return true
}

// ============================================================================
// String Representation
// ============================================================================

// String returns a string representation of the claim type.
func (ct ClaimType) String() string {
	if !ct.HasConstraints() {
		return ct.dataType.String()
	}

	return fmt.Sprintf("%s(%s)", ct.dataType, ct.constraintsString())
}

// GoString implements fmt.GoStringer for debugging.
func (ct ClaimType) GoString() string {
	return fmt.Sprintf("ClaimType{%s}", ct.String())
}

// constraintsString returns a string representation of constraints.
func (ct ClaimType) constraintsString() string {
	c := ct.constraints
	parts := make([]string, 0)

	if len(c.AllowedValues) > 0 {
		parts = append(parts, fmt.Sprintf("values=%v", c.AllowedValues))
	}
	if c.Min != nil {
		parts = append(parts, fmt.Sprintf("min=%d", *c.Min))
	}
	if c.Max != nil {
		parts = append(parts, fmt.Sprintf("max=%d", *c.Max))
	}
	if c.MinLength != nil {
		parts = append(parts, fmt.Sprintf("minLen=%d", *c.MinLength))
	}
	if c.MaxLength != nil {
		parts = append(parts, fmt.Sprintf("maxLen=%d", *c.MaxLength))
	}
	if c.Pattern != nil {
		parts = append(parts, fmt.Sprintf("pattern=%s", *c.Pattern))
	}
	if c.Format != nil {
		parts = append(parts, fmt.Sprintf("format=%s", *c.Format))
	}

	result := ""
	for i, p := range parts {
		if i > 0 {
			result += ", "
		}
		result += p
	}
	return result
}

// ============================================================================
// Helpers
// ============================================================================

func stringSliceEquals(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func int64PtrEquals(a, b *int64) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func intPtrEquals(a, b *int) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func stringPtrEquals(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
