package domain

import "strings"

// DataType represents the type of value a claim can hold.
// It determines validation rules and how the claim value is stored/displayed.
type DataType string

// ============================================================================
// Data Type Constants
// ============================================================================

const (
	// Primitive types
	DataTypeString   DataType = "string"
	DataTypeInteger  DataType = "integer"
	DataTypeBoolean  DataType = "boolean"
	DataTypeDate     DataType = "date"
	DataTypeDateTime DataType = "datetime"
	DataTypeDuration DataType = "duration"

	// Complex types
	DataTypeStringArray DataType = "string_array"
	DataTypeObject      DataType = "object"
	DataTypeEnum        DataType = "enum"

	// Semantic types (strings with validation)
	DataTypeEmail DataType = "email"
	DataTypeURL   DataType = "url"
	DataTypeDID   DataType = "did"
)

// ============================================================================
// Data Type Sets
// ============================================================================

// allDataTypes contains all valid data types for iteration and validation.
var allDataTypes = []DataType{
	DataTypeString,
	DataTypeInteger,
	DataTypeBoolean,
	DataTypeDate,
	DataTypeDateTime,
	DataTypeDuration,
	DataTypeStringArray,
	DataTypeObject,
	DataTypeEnum,
	DataTypeEmail,
	DataTypeURL,
	DataTypeDID,
}

// primitiveTypes are basic value types without special structure.
var primitiveTypes = map[DataType]struct{}{
	DataTypeString:   {},
	DataTypeInteger:  {},
	DataTypeBoolean:  {},
	DataTypeDate:     {},
	DataTypeDateTime: {},
	DataTypeDuration: {},
}

// complexTypes require additional structure or configuration.
var complexTypes = map[DataType]struct{}{
	DataTypeStringArray: {},
	DataTypeObject:      {},
	DataTypeEnum:        {},
}

// semanticTypes are strings with specific validation rules.
var semanticTypes = map[DataType]struct{}{
	DataTypeEmail: {},
	DataTypeURL:   {},
	DataTypeDID:   {},
}

// constraintRequiredTypes need additional configuration (e.g., enum values).
var constraintRequiredTypes = map[DataType]struct{}{
	DataTypeEnum: {},
}

// ============================================================================
// Methods
// ============================================================================

// String returns the string representation of the data type.
func (dt DataType) String() string {
	return string(dt)
}

// IsValid returns true if the data type is a known valid type.
func (dt DataType) IsValid() bool {
	for _, valid := range allDataTypes {
		if dt == valid {
			return true
		}
	}
	return false
}

// IsZero returns true if the data type is empty.
func (dt DataType) IsZero() bool {
	return dt == ""
}

// IsPrimitive returns true if this is a primitive type.
func (dt DataType) IsPrimitive() bool {
	_, ok := primitiveTypes[dt]
	return ok
}

// IsComplex returns true if this is a complex type.
func (dt DataType) IsComplex() bool {
	_, ok := complexTypes[dt]
	return ok
}

// IsSemantic returns true if this is a semantic type with validation.
func (dt DataType) IsSemantic() bool {
	_, ok := semanticTypes[dt]
	return ok
}

// RequiresConstraints returns true if this type requires additional
// configuration (e.g., enum requires allowed values).
func (dt DataType) RequiresConstraints() bool {
	_, ok := constraintRequiredTypes[dt]
	return ok
}

// IsNumeric returns true if this type represents a numeric value.
func (dt DataType) IsNumeric() bool {
	return dt == DataTypeInteger
}

// IsTemporal returns true if this type represents a time-based value.
func (dt DataType) IsTemporal() bool {
	return dt == DataTypeDate || dt == DataTypeDateTime || dt == DataTypeDuration
}

// IsArrayType returns true if this type holds multiple values.
func (dt DataType) IsArrayType() bool {
	return dt == DataTypeStringArray
}

// ============================================================================
// Parsing
// ============================================================================

// ParseDataType parses a string into a DataType.
// Returns an error if the string is not a valid data type.
func ParseDataType(s string) (DataType, error) {
	normalized := DataType(strings.ToLower(strings.TrimSpace(s)))

	if !normalized.IsValid() {
		return "", ErrInvalidDataType
	}

	return normalized, nil
}

// MustParseDataType parses a string into a DataType and panics if invalid.
// Only use for constants or tests.
func MustParseDataType(s string) DataType {
	dt, err := ParseDataType(s)
	if err != nil {
		panic("invalid data type: " + s)
	}
	return dt
}

// ============================================================================
// Helpers
// ============================================================================

// AllDataTypes returns a slice of all valid data types.
func AllDataTypes() []DataType {
	result := make([]DataType, len(allDataTypes))
	copy(result, allDataTypes)
	return result
}

// PrimitiveDataTypes returns a slice of all primitive data types.
func PrimitiveDataTypes() []DataType {
	result := make([]DataType, 0, len(primitiveTypes))
	for dt := range primitiveTypes {
		result = append(result, dt)
	}
	return result
}

// SemanticDataTypes returns a slice of all semantic data types.
func SemanticDataTypes() []DataType {
	result := make([]DataType, 0, len(semanticTypes))
	for dt := range semanticTypes {
		result = append(result, dt)
	}
	return result
}
