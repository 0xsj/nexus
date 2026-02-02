package domain

import (
	"testing"
)

// ============================================================================
// DataType Validation Tests
// ============================================================================

func TestDataType_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		dataType DataType
		want     bool
	}{
		{"string", DataTypeString, true},
		{"integer", DataTypeInteger, true},
		{"boolean", DataTypeBoolean, true},
		{"date", DataTypeDate, true},
		{"datetime", DataTypeDateTime, true},
		{"duration", DataTypeDuration, true},
		{"string_array", DataTypeStringArray, true},
		{"object", DataTypeObject, true},
		{"enum", DataTypeEnum, true},
		{"email", DataTypeEmail, true},
		{"url", DataTypeURL, true},
		{"did", DataTypeDID, true},
		{"invalid", DataType("invalid"), false},
		{"empty", DataType(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.dataType.IsValid(); got != tt.want {
				t.Errorf("DataType(%q).IsValid() = %v, want %v", tt.dataType, got, tt.want)
			}
		})
	}
}

func TestDataType_IsZero(t *testing.T) {
	tests := []struct {
		name     string
		dataType DataType
		want     bool
	}{
		{"empty", DataType(""), true},
		{"string", DataTypeString, false},
		{"integer", DataTypeInteger, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.dataType.IsZero(); got != tt.want {
				t.Errorf("DataType(%q).IsZero() = %v, want %v", tt.dataType, got, tt.want)
			}
		})
	}
}

// ============================================================================
// DataType Category Tests
// ============================================================================

func TestDataType_IsPrimitive(t *testing.T) {
	primitives := []DataType{
		DataTypeString, DataTypeInteger, DataTypeBoolean,
		DataTypeDate, DataTypeDateTime, DataTypeDuration,
	}

	nonPrimitives := []DataType{
		DataTypeStringArray, DataTypeObject, DataTypeEnum,
		DataTypeEmail, DataTypeURL, DataTypeDID,
	}

	for _, dt := range primitives {
		if !dt.IsPrimitive() {
			t.Errorf("DataType(%q).IsPrimitive() = false, want true", dt)
		}
	}

	for _, dt := range nonPrimitives {
		if dt.IsPrimitive() {
			t.Errorf("DataType(%q).IsPrimitive() = true, want false", dt)
		}
	}
}

func TestDataType_IsComplex(t *testing.T) {
	complex := []DataType{
		DataTypeStringArray, DataTypeObject, DataTypeEnum,
	}

	nonComplex := []DataType{
		DataTypeString, DataTypeInteger, DataTypeBoolean,
		DataTypeEmail, DataTypeURL, DataTypeDID,
	}

	for _, dt := range complex {
		if !dt.IsComplex() {
			t.Errorf("DataType(%q).IsComplex() = false, want true", dt)
		}
	}

	for _, dt := range nonComplex {
		if dt.IsComplex() {
			t.Errorf("DataType(%q).IsComplex() = true, want false", dt)
		}
	}
}

func TestDataType_IsSemantic(t *testing.T) {
	semantic := []DataType{
		DataTypeEmail, DataTypeURL, DataTypeDID,
	}

	nonSemantic := []DataType{
		DataTypeString, DataTypeInteger, DataTypeBoolean,
		DataTypeStringArray, DataTypeObject, DataTypeEnum,
	}

	for _, dt := range semantic {
		if !dt.IsSemantic() {
			t.Errorf("DataType(%q).IsSemantic() = false, want true", dt)
		}
	}

	for _, dt := range nonSemantic {
		if dt.IsSemantic() {
			t.Errorf("DataType(%q).IsSemantic() = true, want false", dt)
		}
	}
}

func TestDataType_IsNumeric(t *testing.T) {
	if !DataTypeInteger.IsNumeric() {
		t.Error("DataTypeInteger.IsNumeric() = false, want true")
	}

	nonNumeric := []DataType{
		DataTypeString, DataTypeBoolean, DataTypeDate,
		DataTypeStringArray, DataTypeEmail,
	}

	for _, dt := range nonNumeric {
		if dt.IsNumeric() {
			t.Errorf("DataType(%q).IsNumeric() = true, want false", dt)
		}
	}
}

func TestDataType_IsTemporal(t *testing.T) {
	temporal := []DataType{
		DataTypeDate, DataTypeDateTime, DataTypeDuration,
	}

	nonTemporal := []DataType{
		DataTypeString, DataTypeInteger, DataTypeBoolean,
		DataTypeEmail, DataTypeEnum,
	}

	for _, dt := range temporal {
		if !dt.IsTemporal() {
			t.Errorf("DataType(%q).IsTemporal() = false, want true", dt)
		}
	}

	for _, dt := range nonTemporal {
		if dt.IsTemporal() {
			t.Errorf("DataType(%q).IsTemporal() = true, want false", dt)
		}
	}
}

func TestDataType_IsArrayType(t *testing.T) {
	if !DataTypeStringArray.IsArrayType() {
		t.Error("DataTypeStringArray.IsArrayType() = false, want true")
	}

	nonArray := []DataType{
		DataTypeString, DataTypeInteger, DataTypeObject, DataTypeEnum,
	}

	for _, dt := range nonArray {
		if dt.IsArrayType() {
			t.Errorf("DataType(%q).IsArrayType() = true, want false", dt)
		}
	}
}

func TestDataType_RequiresConstraints(t *testing.T) {
	if !DataTypeEnum.RequiresConstraints() {
		t.Error("DataTypeEnum.RequiresConstraints() = false, want true")
	}

	noConstraintsRequired := []DataType{
		DataTypeString, DataTypeInteger, DataTypeBoolean,
		DataTypeEmail, DataTypeStringArray,
	}

	for _, dt := range noConstraintsRequired {
		if dt.RequiresConstraints() {
			t.Errorf("DataType(%q).RequiresConstraints() = true, want false", dt)
		}
	}
}

// ============================================================================
// ParseDataType Tests
// ============================================================================

func TestParseDataType_Valid(t *testing.T) {
	tests := []struct {
		input string
		want  DataType
	}{
		{"string", DataTypeString},
		{"STRING", DataTypeString},
		{"  string  ", DataTypeString},
		{"integer", DataTypeInteger},
		{"boolean", DataTypeBoolean},
		{"date", DataTypeDate},
		{"datetime", DataTypeDateTime},
		{"duration", DataTypeDuration},
		{"string_array", DataTypeStringArray},
		{"object", DataTypeObject},
		{"enum", DataTypeEnum},
		{"email", DataTypeEmail},
		{"url", DataTypeURL},
		{"did", DataTypeDID},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseDataType(tt.input)
			if err != nil {
				t.Fatalf("ParseDataType(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("ParseDataType(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseDataType_Invalid(t *testing.T) {
	invalidInputs := []string{
		"",
		"invalid",
		"str",
		"int",
		"bool",
		"array",
	}

	for _, input := range invalidInputs {
		t.Run(input, func(t *testing.T) {
			_, err := ParseDataType(input)
			if err == nil {
				t.Errorf("ParseDataType(%q) expected error, got nil", input)
			}
		})
	}
}

func TestMustParseDataType_Valid(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("MustParseDataType panicked unexpectedly: %v", r)
		}
	}()

	dt := MustParseDataType("string")
	if dt != DataTypeString {
		t.Errorf("MustParseDataType(\"string\") = %v, want %v", dt, DataTypeString)
	}
}

func TestMustParseDataType_Invalid_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustParseDataType(\"invalid\") expected panic")
		}
	}()

	MustParseDataType("invalid")
}

// ============================================================================
// String Tests
// ============================================================================

func TestDataType_String(t *testing.T) {
	tests := []struct {
		dataType DataType
		want     string
	}{
		{DataTypeString, "string"},
		{DataTypeInteger, "integer"},
		{DataTypeBoolean, "boolean"},
		{DataTypeEnum, "enum"},
		{DataTypeEmail, "email"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.dataType.String(); got != tt.want {
				t.Errorf("DataType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ============================================================================
// Helper Function Tests
// ============================================================================

func TestAllDataTypes(t *testing.T) {
	all := AllDataTypes()

	if len(all) != 12 {
		t.Errorf("AllDataTypes() returned %d types, want 12", len(all))
	}

	// Verify all returned types are valid
	for _, dt := range all {
		if !dt.IsValid() {
			t.Errorf("AllDataTypes() returned invalid type: %v", dt)
		}
	}
}

func TestPrimitiveDataTypes(t *testing.T) {
	primitives := PrimitiveDataTypes()

	for _, dt := range primitives {
		if !dt.IsPrimitive() {
			t.Errorf("PrimitiveDataTypes() returned non-primitive: %v", dt)
		}
	}
}

func TestSemanticDataTypes(t *testing.T) {
	semantics := SemanticDataTypes()

	for _, dt := range semantics {
		if !dt.IsSemantic() {
			t.Errorf("SemanticDataTypes() returned non-semantic: %v", dt)
		}
	}
}
