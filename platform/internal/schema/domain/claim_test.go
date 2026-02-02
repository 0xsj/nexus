package domain

import (
	"strings"
	"testing"
)

// ============================================================================
// NewClaimDefinition Tests
// ============================================================================

func TestNewClaimDefinition_Basic(t *testing.T) {
	ct := MustNewClaimType(DataTypeString)
	cd, err := NewClaimDefinition("first_name", ct)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cd.Key() != "first_name" {
		t.Errorf("Key() = %v, want first_name", cd.Key())
	}
	if cd.DataType() != DataTypeString {
		t.Errorf("DataType() = %v, want %v", cd.DataType(), DataTypeString)
	}
	if cd.IsRequired() {
		t.Error("IsRequired() = true, want false (default)")
	}
	if !cd.IsOptional() {
		t.Error("IsOptional() = false, want true (default)")
	}
	if !cd.IsDisclosable() {
		t.Error("IsDisclosable() = false, want true (default)")
	}
	if cd.Order() != 0 {
		t.Errorf("Order() = %d, want 0 (default)", cd.Order())
	}
}

func TestNewClaimDefinition_WithAllOptions(t *testing.T) {
	ct := MustNewClaimType(DataTypeString, WithMinLength(1), WithMaxLength(100))
	cd, err := NewClaimDefinition(
		"email_address",
		ct,
		Required(),
		WithDisplayName("Email Address"),
		WithDescription("The user's primary email address"),
		NonDisclosable(),
		WithOrder(5),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cd.Key() != "email_address" {
		t.Errorf("Key() = %v, want email_address", cd.Key())
	}
	if !cd.IsRequired() {
		t.Error("IsRequired() = false, want true")
	}
	if cd.DisplayName() != "Email Address" {
		t.Errorf("DisplayName() = %v, want Email Address", cd.DisplayName())
	}
	if cd.Description() != "The user's primary email address" {
		t.Errorf("Description() = %v, want 'The user's primary email address'", cd.Description())
	}
	if cd.IsDisclosable() {
		t.Error("IsDisclosable() = true, want false")
	}
	if cd.Order() != 5 {
		t.Errorf("Order() = %d, want 5", cd.Order())
	}
}

// ============================================================================
// Key Validation Tests
// ============================================================================

func TestNewClaimDefinition_ValidKeys(t *testing.T) {
	ct := MustNewClaimType(DataTypeString)

	validKeys := []string{
		"a",
		"name",
		"first_name",
		"user_id",
		"address1",
		"claim_123",
		"x",
		"a1b2c3",
	}

	for _, key := range validKeys {
		t.Run(key, func(t *testing.T) {
			cd, err := NewClaimDefinition(key, ct)
			if err != nil {
				t.Errorf("NewClaimDefinition(%q) unexpected error: %v", key, err)
			}
			if cd.Key() != key {
				t.Errorf("Key() = %v, want %v", cd.Key(), key)
			}
		})
	}
}

func TestNewClaimDefinition_InvalidKeys(t *testing.T) {
	ct := MustNewClaimType(DataTypeString)

	tests := []struct {
		name string
		key  string
	}{
		{"empty", ""},
		{"starts_with_number", "1name"},
		{"starts_with_underscore", "_name"},
		{"contains_uppercase", "firstName"},
		{"contains_hyphen", "first-name"},
		{"contains_space", "first name"},
		{"contains_special", "first@name"},
		{"too_long", strings.Repeat("a", ClaimKeyMaxLength+1)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewClaimDefinition(tt.key, ct)
			if err == nil {
				t.Errorf("NewClaimDefinition(%q) expected error, got nil", tt.key)
			}
		})
	}
}

func TestNewClaimDefinition_MaxKeyLength(t *testing.T) {
	ct := MustNewClaimType(DataTypeString)
	maxKey := "a" + strings.Repeat("b", ClaimKeyMaxLength-1)

	cd, err := NewClaimDefinition(maxKey, ct)
	if err != nil {
		t.Fatalf("unexpected error for max length key: %v", err)
	}
	if len(cd.Key()) != ClaimKeyMaxLength {
		t.Errorf("Key length = %d, want %d", len(cd.Key()), ClaimKeyMaxLength)
	}
}

// ============================================================================
// ClaimType Validation Tests
// ============================================================================

func TestNewClaimDefinition_ZeroClaimType(t *testing.T) {
	var zeroCT ClaimType
	_, err := NewClaimDefinition("name", zeroCT)
	if err == nil {
		t.Error("expected error for zero ClaimType")
	}
}

// ============================================================================
// DisplayName Tests
// ============================================================================

func TestClaimDefinition_DisplayName_Default(t *testing.T) {
	ct := MustNewClaimType(DataTypeString)
	cd, _ := NewClaimDefinition("first_name", ct)

	// Should return key when displayName is not set
	if cd.DisplayName() != "first_name" {
		t.Errorf("DisplayName() = %v, want first_name (key)", cd.DisplayName())
	}
}

func TestClaimDefinition_DisplayName_Set(t *testing.T) {
	ct := MustNewClaimType(DataTypeString)
	cd, _ := NewClaimDefinition("first_name", ct, WithDisplayName("First Name"))

	if cd.DisplayName() != "First Name" {
		t.Errorf("DisplayName() = %v, want 'First Name'", cd.DisplayName())
	}
}

func TestNewClaimDefinition_DisplayName_TooLong(t *testing.T) {
	ct := MustNewClaimType(DataTypeString)
	longName := strings.Repeat("a", ClaimDisplayNameMaxLength+1)

	_, err := NewClaimDefinition("name", ct, WithDisplayName(longName))
	if err == nil {
		t.Error("expected error for display name too long")
	}
}

func TestNewClaimDefinition_DisplayName_MaxLength(t *testing.T) {
	ct := MustNewClaimType(DataTypeString)
	maxName := strings.Repeat("a", ClaimDisplayNameMaxLength)

	cd, err := NewClaimDefinition("name", ct, WithDisplayName(maxName))
	if err != nil {
		t.Fatalf("unexpected error for max length display name: %v", err)
	}
	if len(cd.DisplayName()) != ClaimDisplayNameMaxLength {
		t.Errorf("DisplayName length = %d, want %d", len(cd.DisplayName()), ClaimDisplayNameMaxLength)
	}
}

// ============================================================================
// Description Tests
// ============================================================================

func TestNewClaimDefinition_Description_TooLong(t *testing.T) {
	ct := MustNewClaimType(DataTypeString)
	longDesc := strings.Repeat("a", ClaimDescriptionMaxLength+1)

	_, err := NewClaimDefinition("name", ct, WithDescription(longDesc))
	if err == nil {
		t.Error("expected error for description too long")
	}
}

func TestNewClaimDefinition_Description_MaxLength(t *testing.T) {
	ct := MustNewClaimType(DataTypeString)
	maxDesc := strings.Repeat("a", ClaimDescriptionMaxLength)

	cd, err := NewClaimDefinition("name", ct, WithDescription(maxDesc))
	if err != nil {
		t.Fatalf("unexpected error for max length description: %v", err)
	}
	if len(cd.Description()) != ClaimDescriptionMaxLength {
		t.Errorf("Description length = %d, want %d", len(cd.Description()), ClaimDescriptionMaxLength)
	}
}

// ============================================================================
// Order Tests
// ============================================================================

func TestNewClaimDefinition_Order_Valid(t *testing.T) {
	ct := MustNewClaimType(DataTypeString)

	tests := []int{0, 1, 10, 100, 1000}
	for _, order := range tests {
		cd, err := NewClaimDefinition("name", ct, WithOrder(order))
		if err != nil {
			t.Fatalf("unexpected error for order %d: %v", order, err)
		}
		if cd.Order() != order {
			t.Errorf("Order() = %d, want %d", cd.Order(), order)
		}
	}
}

func TestNewClaimDefinition_Order_Negative(t *testing.T) {
	ct := MustNewClaimType(DataTypeString)

	_, err := NewClaimDefinition("name", ct, WithOrder(-1))
	if err == nil {
		t.Error("expected error for negative order")
	}
}

// ============================================================================
// Required/Optional Tests
// ============================================================================

func TestClaimDefinition_Required(t *testing.T) {
	ct := MustNewClaimType(DataTypeString)

	cd, _ := NewClaimDefinition("name", ct, Required())
	if !cd.IsRequired() {
		t.Error("IsRequired() = false, want true")
	}
	if cd.IsOptional() {
		t.Error("IsOptional() = true, want false")
	}
}

func TestClaimDefinition_Optional(t *testing.T) {
	ct := MustNewClaimType(DataTypeString)

	cd, _ := NewClaimDefinition("name", ct, Optional())
	if cd.IsRequired() {
		t.Error("IsRequired() = true, want false")
	}
	if !cd.IsOptional() {
		t.Error("IsOptional() = false, want true")
	}
}

func TestClaimDefinition_Required_Then_Optional(t *testing.T) {
	ct := MustNewClaimType(DataTypeString)

	// Last option wins
	cd, _ := NewClaimDefinition("name", ct, Required(), Optional())
	if cd.IsRequired() {
		t.Error("IsRequired() = true, want false (Optional called last)")
	}
}

// ============================================================================
// Disclosable Tests
// ============================================================================

func TestClaimDefinition_Disclosable(t *testing.T) {
	ct := MustNewClaimType(DataTypeString)

	cd, _ := NewClaimDefinition("name", ct, Disclosable())
	if !cd.IsDisclosable() {
		t.Error("IsDisclosable() = false, want true")
	}
}

func TestClaimDefinition_NonDisclosable(t *testing.T) {
	ct := MustNewClaimType(DataTypeString)

	cd, _ := NewClaimDefinition("name", ct, NonDisclosable())
	if cd.IsDisclosable() {
		t.Error("IsDisclosable() = true, want false")
	}
}

// ============================================================================
// MustNewClaimDefinition Tests
// ============================================================================

func TestMustNewClaimDefinition_Valid(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("MustNewClaimDefinition panicked unexpectedly: %v", r)
		}
	}()

	ct := MustNewClaimType(DataTypeString)
	cd := MustNewClaimDefinition("name", ct)
	if cd.Key() != "name" {
		t.Errorf("Key() = %v, want name", cd.Key())
	}
}

func TestMustNewClaimDefinition_Invalid_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustNewClaimDefinition expected panic")
		}
	}()

	ct := MustNewClaimType(DataTypeString)
	MustNewClaimDefinition("", ct) // Empty key should panic
}

// ============================================================================
// IsZero Tests
// ============================================================================

func TestClaimDefinition_IsZero(t *testing.T) {
	var zero ClaimDefinition
	if !zero.IsZero() {
		t.Error("zero ClaimDefinition.IsZero() = false, want true")
	}

	ct := MustNewClaimType(DataTypeString)
	cd, _ := NewClaimDefinition("name", ct)
	if cd.IsZero() {
		t.Error("initialized ClaimDefinition.IsZero() = true, want false")
	}
}

// ============================================================================
// Equals Tests
// ============================================================================

func TestClaimDefinition_Equals(t *testing.T) {
	ct := MustNewClaimType(DataTypeString)

	tests := []struct {
		name string
		a    ClaimDefinition
		b    ClaimDefinition
		want bool
	}{
		{
			name: "identical",
			a:    MustNewClaimDefinition("name", ct),
			b:    MustNewClaimDefinition("name", ct),
			want: true,
		},
		{
			name: "different_keys",
			a:    MustNewClaimDefinition("name", ct),
			b:    MustNewClaimDefinition("email", ct),
			want: false,
		},
		{
			name: "different_types",
			a:    MustNewClaimDefinition("value", MustNewClaimType(DataTypeString)),
			b:    MustNewClaimDefinition("value", MustNewClaimType(DataTypeInteger)),
			want: false,
		},
		{
			name: "different_required",
			a:    MustNewClaimDefinition("name", ct, Required()),
			b:    MustNewClaimDefinition("name", ct, Optional()),
			want: false,
		},
		{
			name: "different_display_name",
			a:    MustNewClaimDefinition("name", ct, WithDisplayName("Name")),
			b:    MustNewClaimDefinition("name", ct, WithDisplayName("Full Name")),
			want: false,
		},
		{
			name: "different_description",
			a:    MustNewClaimDefinition("name", ct, WithDescription("Desc A")),
			b:    MustNewClaimDefinition("name", ct, WithDescription("Desc B")),
			want: false,
		},
		{
			name: "different_disclosable",
			a:    MustNewClaimDefinition("name", ct, Disclosable()),
			b:    MustNewClaimDefinition("name", ct, NonDisclosable()),
			want: false,
		},
		{
			name: "different_order",
			a:    MustNewClaimDefinition("name", ct, WithOrder(1)),
			b:    MustNewClaimDefinition("name", ct, WithOrder(2)),
			want: false,
		},
		{
			name: "all_options_same",
			a: MustNewClaimDefinition("name", ct,
				Required(), WithDisplayName("Name"), WithDescription("Desc"), WithOrder(1)),
			b: MustNewClaimDefinition("name", ct,
				Required(), WithDisplayName("Name"), WithDescription("Desc"), WithOrder(1)),
			want: true,
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

func TestClaimDefinition_KeyEquals(t *testing.T) {
	ct := MustNewClaimType(DataTypeString)

	a := MustNewClaimDefinition("name", ct, Required())
	b := MustNewClaimDefinition("name", ct, Optional())
	c := MustNewClaimDefinition("email", ct)

	if !a.KeyEquals(b) {
		t.Error("KeyEquals() = false for same keys")
	}
	if a.KeyEquals(c) {
		t.Error("KeyEquals() = true for different keys")
	}
}

// ============================================================================
// ClaimType Accessor Tests
// ============================================================================

func TestClaimDefinition_ClaimType(t *testing.T) {
	ct := MustNewClaimType(DataTypeInteger, WithMinMax(0, 100))
	cd, _ := NewClaimDefinition("age", ct)

	gotCT := cd.ClaimType()
	if !gotCT.Equals(ct) {
		t.Error("ClaimType() does not match original")
	}

	// Verify constraints are preserved
	c := gotCT.Constraints()
	if c.Min == nil || *c.Min != 0 {
		t.Errorf("Min = %v, want 0", c.Min)
	}
	if c.Max == nil || *c.Max != 100 {
		t.Errorf("Max = %v, want 100", c.Max)
	}
}

// ============================================================================
// String Representation Tests
// ============================================================================

func TestClaimDefinition_String(t *testing.T) {
	ct := MustNewClaimType(DataTypeString)

	cd := MustNewClaimDefinition("name", ct, Required())
	s := cd.String()

	if !strings.Contains(s, "name") {
		t.Errorf("String() should contain key, got %v", s)
	}
	if !strings.Contains(s, "required") {
		t.Errorf("String() should contain 'required', got %v", s)
	}
}

func TestClaimDefinition_String_Optional(t *testing.T) {
	ct := MustNewClaimType(DataTypeString)

	cd := MustNewClaimDefinition("name", ct, Optional())
	s := cd.String()

	if !strings.Contains(s, "optional") {
		t.Errorf("String() should contain 'optional', got %v", s)
	}
}

func TestClaimDefinition_GoString(t *testing.T) {
	ct := MustNewClaimType(DataTypeString)
	cd := MustNewClaimDefinition("name", ct, Required(), NonDisclosable())

	s := cd.GoString()

	if !strings.Contains(s, "ClaimDefinition") {
		t.Errorf("GoString() should contain 'ClaimDefinition', got %v", s)
	}
	if !strings.Contains(s, "name") {
		t.Errorf("GoString() should contain key, got %v", s)
	}
	if !strings.Contains(s, "required=true") {
		t.Errorf("GoString() should contain 'required=true', got %v", s)
	}
	if !strings.Contains(s, "disclosable=false") {
		t.Errorf("GoString() should contain 'disclosable=false', got %v", s)
	}
}

// ============================================================================
// Builder Tests
// ============================================================================

func TestClaimDefinitionBuilder_Build_Valid(t *testing.T) {
	ct := MustNewClaimType(DataTypeString)

	cd, err := NewClaimDefinitionBuilder("name", ct).
		Required().
		DisplayName("Full Name").
		Description("User's full name").
		NonDisclosable().
		Order(10).
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cd.Key() != "name" {
		t.Errorf("Key() = %v, want name", cd.Key())
	}
	if !cd.IsRequired() {
		t.Error("IsRequired() = false, want true")
	}
	if cd.DisplayName() != "Full Name" {
		t.Errorf("DisplayName() = %v, want 'Full Name'", cd.DisplayName())
	}
	if cd.Description() != "User's full name" {
		t.Errorf("Description() = %v, want 'User's full name'", cd.Description())
	}
	if cd.IsDisclosable() {
		t.Error("IsDisclosable() = true, want false")
	}
	if cd.Order() != 10 {
		t.Errorf("Order() = %d, want 10", cd.Order())
	}
}

func TestClaimDefinitionBuilder_Build_Invalid(t *testing.T) {
	ct := MustNewClaimType(DataTypeString)

	_, err := NewClaimDefinitionBuilder("", ct).Build() // Empty key
	if err == nil {
		t.Error("expected error for empty key")
	}
}

func TestClaimDefinitionBuilder_MustBuild_Valid(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("MustBuild panicked unexpectedly: %v", r)
		}
	}()

	ct := MustNewClaimType(DataTypeString)
	cd := NewClaimDefinitionBuilder("name", ct).MustBuild()

	if cd.Key() != "name" {
		t.Errorf("Key() = %v, want name", cd.Key())
	}
}

func TestClaimDefinitionBuilder_MustBuild_Invalid_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustBuild expected panic")
		}
	}()

	ct := MustNewClaimType(DataTypeString)
	NewClaimDefinitionBuilder("", ct).MustBuild() // Empty key
}

func TestClaimDefinitionBuilder_Optional(t *testing.T) {
	ct := MustNewClaimType(DataTypeString)

	cd := NewClaimDefinitionBuilder("name", ct).
		Required().
		Optional(). // Should override Required
		MustBuild()

	if cd.IsRequired() {
		t.Error("IsRequired() = true, want false")
	}
}

func TestClaimDefinitionBuilder_Disclosable(t *testing.T) {
	ct := MustNewClaimType(DataTypeString)

	cd := NewClaimDefinitionBuilder("name", ct).
		NonDisclosable().
		Disclosable(). // Should override NonDisclosable
		MustBuild()

	if !cd.IsDisclosable() {
		t.Error("IsDisclosable() = false, want true")
	}
}

// ============================================================================
// Real-World Examples
// ============================================================================

func TestClaimDefinition_RealWorldExamples(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		dataType DataType
		options  []ClaimDefinitionOption
	}{
		{
			name:     "email",
			key:      "email",
			dataType: DataTypeEmail,
			options:  []ClaimDefinitionOption{Required(), WithDisplayName("Email Address")},
		},
		{
			name:     "age",
			key:      "age",
			dataType: DataTypeInteger,
			options:  []ClaimDefinitionOption{Optional(), WithDescription("User's age in years")},
		},
		{
			name:     "is_verified",
			key:      "is_verified",
			dataType: DataTypeBoolean,
			options:  []ClaimDefinitionOption{Required()},
		},
		{
			name:     "birth_date",
			key:      "birth_date",
			dataType: DataTypeDate,
			options:  []ClaimDefinitionOption{Optional(), NonDisclosable()},
		},
		{
			name:     "skills",
			key:      "skills",
			dataType: DataTypeStringArray,
			options:  []ClaimDefinitionOption{Optional(), WithDisplayName("Technical Skills")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ct := MustNewClaimType(tt.dataType)
			cd, err := NewClaimDefinition(tt.key, ct, tt.options...)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cd.Key() != tt.key {
				t.Errorf("Key() = %v, want %v", cd.Key(), tt.key)
			}
			if cd.DataType() != tt.dataType {
				t.Errorf("DataType() = %v, want %v", cd.DataType(), tt.dataType)
			}
		})
	}
}
