package domain

import (
	"testing"
)

// ============================================================================
// NewClaimType Tests
// ============================================================================

func TestNewClaimType_Basic(t *testing.T) {
	ct, err := NewClaimType(DataTypeString)
	if err != nil {
		t.Fatalf("NewClaimType(DataTypeString) unexpected error: %v", err)
	}

	if ct.DataType() != DataTypeString {
		t.Errorf("DataType() = %v, want %v", ct.DataType(), DataTypeString)
	}
	if ct.IsZero() {
		t.Error("IsZero() = true, want false")
	}
}

func TestNewClaimType_AllPrimitiveDataTypes(t *testing.T) {
	dataTypes := []DataType{
		DataTypeString,
		DataTypeInteger,
		DataTypeBoolean,
		DataTypeDate,
		DataTypeDateTime,
		DataTypeDuration,
		DataTypeStringArray,
		DataTypeObject,
		DataTypeEmail,
		DataTypeURL,
		DataTypeDID,
	}

	for _, dt := range dataTypes {
		t.Run(dt.String(), func(t *testing.T) {
			ct, err := NewClaimType(dt)
			if err != nil {
				t.Fatalf("NewClaimType(%v) unexpected error: %v", dt, err)
			}
			if ct.DataType() != dt {
				t.Errorf("DataType() = %v, want %v", ct.DataType(), dt)
			}
		})
	}
}

func TestNewClaimType_InvalidDataType(t *testing.T) {
	_, err := NewClaimType(DataType("invalid"))
	if err == nil {
		t.Error("NewClaimType(invalid) expected error, got nil")
	}
}

func TestNewClaimType_ZeroDataType(t *testing.T) {
	_, err := NewClaimType(DataType(""))
	if err == nil {
		t.Error("NewClaimType(\"\") expected error, got nil")
	}
}

// ============================================================================
// MustNewClaimType Tests
// ============================================================================

func TestMustNewClaimType_Valid(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("MustNewClaimType panicked unexpectedly: %v", r)
		}
	}()

	ct := MustNewClaimType(DataTypeString)
	if ct.DataType() != DataTypeString {
		t.Errorf("DataType() = %v, want %v", ct.DataType(), DataTypeString)
	}
}

func TestMustNewClaimType_Invalid_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustNewClaimType(invalid) expected panic")
		}
	}()

	MustNewClaimType(DataType("invalid"))
}

// ============================================================================
// Enum Type Tests
// ============================================================================

func TestNewClaimType_Enum_WithAllowedValues(t *testing.T) {
	ct, err := NewClaimType(DataTypeEnum, WithAllowedValues("red", "green", "blue"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ct.DataType() != DataTypeEnum {
		t.Errorf("DataType() = %v, want %v", ct.DataType(), DataTypeEnum)
	}

	allowed := ct.Constraints().AllowedValues
	if len(allowed) != 3 {
		t.Fatalf("AllowedValues length = %d, want 3", len(allowed))
	}
	if allowed[0] != "red" || allowed[1] != "green" || allowed[2] != "blue" {
		t.Errorf("AllowedValues = %v, want [red green blue]", allowed)
	}
}

func TestNewClaimType_Enum_WithoutAllowedValues(t *testing.T) {
	_, err := NewClaimType(DataTypeEnum)
	if err == nil {
		t.Error("NewClaimType(DataTypeEnum) without allowed values expected error")
	}
}

func TestNewClaimType_Enum_EmptyAllowedValues(t *testing.T) {
	_, err := NewClaimType(DataTypeEnum, WithAllowedValues())
	if err == nil {
		t.Error("NewClaimType(DataTypeEnum) with empty allowed values expected error")
	}
}

func TestNewClaimType_NonEnum_WithAllowedValues(t *testing.T) {
	_, err := NewClaimType(DataTypeString, WithAllowedValues("a", "b"))
	if err == nil {
		t.Error("NewClaimType(DataTypeString) with allowed values expected error")
	}
}

// ============================================================================
// String Constraint Tests
// ============================================================================

func TestNewClaimType_String_WithMinLength(t *testing.T) {
	ct, err := NewClaimType(DataTypeString, WithMinLength(5))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	c := ct.Constraints()
	if c.MinLength == nil || *c.MinLength != 5 {
		t.Errorf("MinLength = %v, want 5", c.MinLength)
	}
}

func TestNewClaimType_String_WithMaxLength(t *testing.T) {
	ct, err := NewClaimType(DataTypeString, WithMaxLength(100))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	c := ct.Constraints()
	if c.MaxLength == nil || *c.MaxLength != 100 {
		t.Errorf("MaxLength = %v, want 100", c.MaxLength)
	}
}

func TestNewClaimType_String_WithLengthRange(t *testing.T) {
	ct, err := NewClaimType(DataTypeString, WithLengthRange(5, 100))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	c := ct.Constraints()
	if c.MinLength == nil || *c.MinLength != 5 {
		t.Errorf("MinLength = %v, want 5", c.MinLength)
	}
	if c.MaxLength == nil || *c.MaxLength != 100 {
		t.Errorf("MaxLength = %v, want 100", c.MaxLength)
	}
}

func TestNewClaimType_String_WithPattern(t *testing.T) {
	ct, err := NewClaimType(DataTypeString, WithPattern(`^[A-Z]{2,3}$`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	c := ct.Constraints()
	if c.Pattern == nil || *c.Pattern != `^[A-Z]{2,3}$` {
		t.Errorf("Pattern = %v, want ^[A-Z]{2,3}$", c.Pattern)
	}
}

func TestNewClaimType_String_WithInvalidPattern(t *testing.T) {
	_, err := NewClaimType(DataTypeString, WithPattern(`[invalid(`))
	if err == nil {
		t.Error("invalid regex pattern expected error")
	}
}

func TestNewClaimType_String_InvalidMinLength(t *testing.T) {
	_, err := NewClaimType(DataTypeString, WithMinLength(-1))
	if err == nil {
		t.Error("WithMinLength(-1) expected error")
	}
}

func TestNewClaimType_String_InvalidMaxLength(t *testing.T) {
	_, err := NewClaimType(DataTypeString, WithMaxLength(-1))
	if err == nil {
		t.Error("WithMaxLength(-1) expected error")
	}
}

func TestNewClaimType_String_MinLengthGreaterThanMaxLength(t *testing.T) {
	_, err := NewClaimType(DataTypeString, WithMinLength(100), WithMaxLength(10))
	if err == nil {
		t.Error("MinLength > MaxLength expected error")
	}
}

// ============================================================================
// Integer Constraint Tests
// ============================================================================

func TestNewClaimType_Integer_WithMin(t *testing.T) {
	ct, err := NewClaimType(DataTypeInteger, WithMin(0))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	c := ct.Constraints()
	if c.Min == nil || *c.Min != 0 {
		t.Errorf("Min = %v, want 0", c.Min)
	}
}

func TestNewClaimType_Integer_WithMax(t *testing.T) {
	ct, err := NewClaimType(DataTypeInteger, WithMax(100))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	c := ct.Constraints()
	if c.Max == nil || *c.Max != 100 {
		t.Errorf("Max = %v, want 100", c.Max)
	}
}

func TestNewClaimType_Integer_WithMinMax(t *testing.T) {
	ct, err := NewClaimType(DataTypeInteger, WithMinMax(1, 10))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	c := ct.Constraints()
	if c.Min == nil || *c.Min != 1 {
		t.Errorf("Min = %v, want 1", c.Min)
	}
	if c.Max == nil || *c.Max != 10 {
		t.Errorf("Max = %v, want 10", c.Max)
	}
}

func TestNewClaimType_Integer_NegativeRange(t *testing.T) {
	ct, err := NewClaimType(DataTypeInteger, WithMinMax(-100, -1))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	c := ct.Constraints()
	if c.Min == nil || *c.Min != -100 {
		t.Errorf("Min = %v, want -100", c.Min)
	}
	if c.Max == nil || *c.Max != -1 {
		t.Errorf("Max = %v, want -1", c.Max)
	}
}

func TestNewClaimType_Integer_MinGreaterThanMax(t *testing.T) {
	_, err := NewClaimType(DataTypeInteger, WithMinMax(100, 10))
	if err == nil {
		t.Error("Min > Max expected error")
	}
}

func TestNewClaimType_NonInteger_WithMin(t *testing.T) {
	_, err := NewClaimType(DataTypeString, WithMin(0))
	if err == nil {
		t.Error("String with Min constraint expected error")
	}
}

func TestNewClaimType_NonInteger_WithMax(t *testing.T) {
	_, err := NewClaimType(DataTypeString, WithMax(100))
	if err == nil {
		t.Error("String with Max constraint expected error")
	}
}

// ============================================================================
// Temporal Constraint Tests
// ============================================================================

func TestNewClaimType_Date_WithFormat(t *testing.T) {
	ct, err := NewClaimType(DataTypeDate, WithFormat("2006-01-02"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	c := ct.Constraints()
	if c.Format == nil || *c.Format != "2006-01-02" {
		t.Errorf("Format = %v, want 2006-01-02", c.Format)
	}
}

func TestNewClaimType_DateTime_WithFormat(t *testing.T) {
	ct, err := NewClaimType(DataTypeDateTime, WithFormat("2006-01-02T15:04:05Z07:00"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	c := ct.Constraints()
	if c.Format == nil || *c.Format != "2006-01-02T15:04:05Z07:00" {
		t.Errorf("Format = %v, want RFC3339 format", c.Format)
	}
}

func TestNewClaimType_NonTemporal_WithFormat(t *testing.T) {
	_, err := NewClaimType(DataTypeString, WithFormat("2006-01-02"))
	if err == nil {
		t.Error("String with Format constraint expected error")
	}
}

// ============================================================================
// Length Constraint Applicability Tests
// ============================================================================

func TestNewClaimType_LengthConstraints_AllowedTypes(t *testing.T) {
	allowedTypes := []DataType{
		DataTypeString,
		DataTypeStringArray,
		DataTypeEmail,
		DataTypeURL,
		DataTypeDID,
	}

	for _, dt := range allowedTypes {
		t.Run(dt.String(), func(t *testing.T) {
			ct, err := NewClaimType(dt, WithMinLength(1), WithMaxLength(100))
			if err != nil {
				t.Fatalf("NewClaimType(%v) with length constraints unexpected error: %v", dt, err)
			}
			c := ct.Constraints()
			if c.MinLength == nil || *c.MinLength != 1 {
				t.Errorf("MinLength = %v, want 1", c.MinLength)
			}
		})
	}
}

func TestNewClaimType_LengthConstraints_DisallowedTypes(t *testing.T) {
	disallowedTypes := []DataType{
		DataTypeInteger,
		DataTypeBoolean,
		DataTypeDate,
		DataTypeDateTime,
		DataTypeDuration,
		DataTypeObject,
	}

	for _, dt := range disallowedTypes {
		t.Run(dt.String(), func(t *testing.T) {
			_, err := NewClaimType(dt, WithMinLength(1))
			if err == nil {
				t.Errorf("NewClaimType(%v) with length constraint expected error", dt)
			}
		})
	}
}

// ============================================================================
// Pattern Constraint Applicability Tests
// ============================================================================

func TestNewClaimType_PatternConstraints_AllowedTypes(t *testing.T) {
	allowedTypes := []DataType{
		DataTypeString,
		DataTypeEmail,
		DataTypeURL,
		DataTypeDID,
	}

	for _, dt := range allowedTypes {
		t.Run(dt.String(), func(t *testing.T) {
			ct, err := NewClaimType(dt, WithPattern(`^[a-z]+$`))
			if err != nil {
				t.Fatalf("NewClaimType(%v) with pattern constraint unexpected error: %v", dt, err)
			}
			c := ct.Constraints()
			if c.Pattern == nil || *c.Pattern != `^[a-z]+$` {
				t.Errorf("Pattern = %v, want ^[a-z]+$", c.Pattern)
			}
		})
	}
}

func TestNewClaimType_PatternConstraints_DisallowedTypes(t *testing.T) {
	disallowedTypes := []DataType{
		DataTypeInteger,
		DataTypeBoolean,
		DataTypeStringArray,
		DataTypeObject,
	}

	for _, dt := range disallowedTypes {
		t.Run(dt.String(), func(t *testing.T) {
			_, err := NewClaimType(dt, WithPattern(`^[a-z]+$`))
			if err == nil {
				t.Errorf("NewClaimType(%v) with pattern constraint expected error", dt)
			}
		})
	}
}

// ============================================================================
// IsZero Tests
// ============================================================================

func TestClaimType_IsZero(t *testing.T) {
	var zero ClaimType
	if !zero.IsZero() {
		t.Error("zero ClaimType.IsZero() = false, want true")
	}

	ct, _ := NewClaimType(DataTypeString)
	if ct.IsZero() {
		t.Error("initialized ClaimType.IsZero() = true, want false")
	}
}

// ============================================================================
// HasConstraints Tests
// ============================================================================

func TestClaimType_HasConstraints(t *testing.T) {
	tests := []struct {
		name string
		ct   ClaimType
		want bool
	}{
		{
			name: "no_constraints",
			ct:   MustNewClaimType(DataTypeString),
			want: false,
		},
		{
			name: "with_min",
			ct:   MustNewClaimType(DataTypeInteger, WithMin(0)),
			want: true,
		},
		{
			name: "with_max",
			ct:   MustNewClaimType(DataTypeInteger, WithMax(100)),
			want: true,
		},
		{
			name: "with_min_length",
			ct:   MustNewClaimType(DataTypeString, WithMinLength(1)),
			want: true,
		},
		{
			name: "with_max_length",
			ct:   MustNewClaimType(DataTypeString, WithMaxLength(50)),
			want: true,
		},
		{
			name: "with_pattern",
			ct:   MustNewClaimType(DataTypeString, WithPattern(`^\d+$`)),
			want: true,
		},
		{
			name: "with_format",
			ct:   MustNewClaimType(DataTypeDate, WithFormat("2006-01-02")),
			want: true,
		},
		{
			name: "with_allowed_values",
			ct:   MustNewClaimType(DataTypeEnum, WithAllowedValues("a", "b")),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ct.HasConstraints(); got != tt.want {
				t.Errorf("HasConstraints() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ============================================================================
// Equals Tests
// ============================================================================

func TestClaimType_Equals(t *testing.T) {
	tests := []struct {
		name string
		a    ClaimType
		b    ClaimType
		want bool
	}{
		{
			name: "same_type_no_constraints",
			a:    MustNewClaimType(DataTypeString),
			b:    MustNewClaimType(DataTypeString),
			want: true,
		},
		{
			name: "different_types",
			a:    MustNewClaimType(DataTypeString),
			b:    MustNewClaimType(DataTypeInteger),
			want: false,
		},
		{
			name: "same_with_constraints",
			a:    MustNewClaimType(DataTypeInteger, WithMinMax(0, 100)),
			b:    MustNewClaimType(DataTypeInteger, WithMinMax(0, 100)),
			want: true,
		},
		{
			name: "different_min",
			a:    MustNewClaimType(DataTypeInteger, WithMin(0)),
			b:    MustNewClaimType(DataTypeInteger, WithMin(1)),
			want: false,
		},
		{
			name: "different_max",
			a:    MustNewClaimType(DataTypeInteger, WithMax(100)),
			b:    MustNewClaimType(DataTypeInteger, WithMax(200)),
			want: false,
		},
		{
			name: "same_allowed_values",
			a:    MustNewClaimType(DataTypeEnum, WithAllowedValues("a", "b")),
			b:    MustNewClaimType(DataTypeEnum, WithAllowedValues("a", "b")),
			want: true,
		},
		{
			name: "different_allowed_values",
			a:    MustNewClaimType(DataTypeEnum, WithAllowedValues("a", "b")),
			b:    MustNewClaimType(DataTypeEnum, WithAllowedValues("a", "c")),
			want: false,
		},
		{
			name: "different_allowed_values_order",
			a:    MustNewClaimType(DataTypeEnum, WithAllowedValues("a", "b")),
			b:    MustNewClaimType(DataTypeEnum, WithAllowedValues("b", "a")),
			want: false,
		},
		{
			name: "same_pattern",
			a:    MustNewClaimType(DataTypeString, WithPattern(`^\d+$`)),
			b:    MustNewClaimType(DataTypeString, WithPattern(`^\d+$`)),
			want: true,
		},
		{
			name: "different_pattern",
			a:    MustNewClaimType(DataTypeString, WithPattern(`^\d+$`)),
			b:    MustNewClaimType(DataTypeString, WithPattern(`^[a-z]+$`)),
			want: false,
		},
		{
			name: "one_with_constraint_one_without",
			a:    MustNewClaimType(DataTypeString),
			b:    MustNewClaimType(DataTypeString, WithMinLength(1)),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.Equals(tt.b); got != tt.want {
				t.Errorf("Equals() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ============================================================================
// String Representation Tests
// ============================================================================

func TestClaimType_String(t *testing.T) {
	tests := []struct {
		name string
		ct   ClaimType
		want string
	}{
		{
			name: "simple_type",
			ct:   MustNewClaimType(DataTypeString),
			want: "string",
		},
		{
			name: "integer",
			ct:   MustNewClaimType(DataTypeInteger),
			want: "integer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ct.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClaimType_String_WithConstraints(t *testing.T) {
	ct := MustNewClaimType(DataTypeInteger, WithMinMax(0, 100))
	s := ct.String()

	// Should contain the data type and constraints info
	if s == "integer" {
		t.Error("String() should include constraints info")
	}
}

func TestClaimType_GoString(t *testing.T) {
	ct := MustNewClaimType(DataTypeString)
	s := ct.GoString()

	if s != "ClaimType{string}" {
		t.Errorf("GoString() = %v, want ClaimType{string}", s)
	}
}

// ============================================================================
// Edge Cases
// ============================================================================

func TestNewClaimType_ZeroMinLength(t *testing.T) {
	ct, err := NewClaimType(DataTypeString, WithMinLength(0))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	c := ct.Constraints()
	if c.MinLength == nil || *c.MinLength != 0 {
		t.Errorf("MinLength = %v, want 0", c.MinLength)
	}
}

func TestNewClaimType_ZeroMaxLength(t *testing.T) {
	ct, err := NewClaimType(DataTypeString, WithMaxLength(0))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	c := ct.Constraints()
	if c.MaxLength == nil || *c.MaxLength != 0 {
		t.Errorf("MaxLength = %v, want 0", c.MaxLength)
	}
}

func TestNewClaimType_EqualMinMax(t *testing.T) {
	ct, err := NewClaimType(DataTypeInteger, WithMinMax(5, 5))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	c := ct.Constraints()
	if c.Min == nil || *c.Min != 5 {
		t.Errorf("Min = %v, want 5", c.Min)
	}
	if c.Max == nil || *c.Max != 5 {
		t.Errorf("Max = %v, want 5", c.Max)
	}
}

func TestNewClaimType_EqualMinMaxLength(t *testing.T) {
	ct, err := NewClaimType(DataTypeString, WithLengthRange(10, 10))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	c := ct.Constraints()
	if c.MinLength == nil || *c.MinLength != 10 {
		t.Errorf("MinLength = %v, want 10", c.MinLength)
	}
	if c.MaxLength == nil || *c.MaxLength != 10 {
		t.Errorf("MaxLength = %v, want 10", c.MaxLength)
	}
}

func TestNewClaimType_SingleAllowedValue(t *testing.T) {
	ct, err := NewClaimType(DataTypeEnum, WithAllowedValues("only"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	allowed := ct.Constraints().AllowedValues
	if len(allowed) != 1 || allowed[0] != "only" {
		t.Errorf("AllowedValues = %v, want [only]", allowed)
	}
}

func TestNewClaimType_MultipleOptions(t *testing.T) {
	ct, err := NewClaimType(
		DataTypeString,
		WithMinLength(1),
		WithMaxLength(100),
		WithPattern(`^[a-z]+$`),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	c := ct.Constraints()
	if c.MinLength == nil || *c.MinLength != 1 {
		t.Errorf("MinLength = %v, want 1", c.MinLength)
	}
	if c.MaxLength == nil || *c.MaxLength != 100 {
		t.Errorf("MaxLength = %v, want 100", c.MaxLength)
	}
	if c.Pattern == nil || *c.Pattern != `^[a-z]+$` {
		t.Errorf("Pattern = %v, want ^[a-z]+$", c.Pattern)
	}
}

// ============================================================================
// Constraints Accessor Tests
// ============================================================================

func TestClaimType_Constraints_NoConstraints(t *testing.T) {
	ct, _ := NewClaimType(DataTypeString)
	c := ct.Constraints()

	if c.Min != nil {
		t.Errorf("Min = %v, want nil", c.Min)
	}
	if c.Max != nil {
		t.Errorf("Max = %v, want nil", c.Max)
	}
	if c.MinLength != nil {
		t.Errorf("MinLength = %v, want nil", c.MinLength)
	}
	if c.MaxLength != nil {
		t.Errorf("MaxLength = %v, want nil", c.MaxLength)
	}
	if c.Pattern != nil {
		t.Errorf("Pattern = %v, want nil", c.Pattern)
	}
	if c.Format != nil {
		t.Errorf("Format = %v, want nil", c.Format)
	}
	if len(c.AllowedValues) != 0 {
		t.Errorf("AllowedValues = %v, want empty", c.AllowedValues)
	}
}
