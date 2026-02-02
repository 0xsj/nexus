package domain

import (
	"testing"
)

// ============================================================================
// ParseSchemaVersion Tests
// ============================================================================

func TestParseSchemaVersion_Valid(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		major   int
		minor   int
		patch   int
		display string
	}{
		{"simple", "1.0.0", 1, 0, 0, "1.0.0"},
		{"with_minor", "1.2.0", 1, 2, 0, "1.2.0"},
		{"with_patch", "1.2.3", 1, 2, 3, "1.2.3"},
		{"large_numbers", "10.20.30", 10, 20, 30, "10.20.30"},
		{"zero_version", "0.0.0", 0, 0, 0, "0.0.0"},
		{"with_v_prefix", "v1.0.0", 1, 0, 0, "1.0.0"},
		{"with_V_prefix", "V2.1.0", 2, 1, 0, "2.1.0"},
		{"with_spaces", "  1.0.0  ", 1, 0, 0, "1.0.0"},
		{"with_v_and_spaces", "  v1.0.0  ", 1, 0, 0, "1.0.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := ParseSchemaVersion(tt.input)
			if err != nil {
				t.Fatalf("ParseSchemaVersion(%q) unexpected error: %v", tt.input, err)
			}
			if v.Major() != tt.major {
				t.Errorf("Major() = %d, want %d", v.Major(), tt.major)
			}
			if v.Minor() != tt.minor {
				t.Errorf("Minor() = %d, want %d", v.Minor(), tt.minor)
			}
			if v.Patch() != tt.patch {
				t.Errorf("Patch() = %d, want %d", v.Patch(), tt.patch)
			}
			if v.String() != tt.display {
				t.Errorf("String() = %q, want %q", v.String(), tt.display)
			}
		})
	}
}

func TestParseSchemaVersion_Invalid(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty", ""},
		{"single_number", "1"},
		{"two_numbers", "1.0"},
		{"four_numbers", "1.0.0.0"},
		{"non_numeric_major", "a.0.0"},
		{"non_numeric_minor", "1.b.0"},
		{"non_numeric_patch", "1.0.c"},
		{"negative_major", "-1.0.0"},
		{"negative_minor", "1.-1.0"},
		{"negative_patch", "1.0.-1"},
		{"decimal", "1.0.0.5"},
		{"spaces_between", "1. 0. 0"},
		{"letters_mixed", "1.0.0a"},
		{"only_dots", "..."},
		{"missing_parts", "1..0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseSchemaVersion(tt.input)
			if err == nil {
				t.Errorf("ParseSchemaVersion(%q) expected error, got nil", tt.input)
			}
		})
	}
}

func TestMustParseSchemaVersion_Valid(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("MustParseSchemaVersion panicked unexpectedly: %v", r)
		}
	}()

	v := MustParseSchemaVersion("1.2.3")
	if v.String() != "1.2.3" {
		t.Errorf("MustParseSchemaVersion(\"1.2.3\").String() = %q, want %q", v.String(), "1.2.3")
	}
}

func TestMustParseSchemaVersion_Invalid_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustParseSchemaVersion(\"invalid\") expected panic")
		}
	}()

	MustParseSchemaVersion("invalid")
}

// ============================================================================
// NewSchemaVersion Tests
// ============================================================================

func TestNewSchemaVersion_Valid(t *testing.T) {
	tests := []struct {
		name  string
		major int
		minor int
		patch int
	}{
		{"all_zeros", 0, 0, 0},
		{"basic", 1, 0, 0},
		{"with_minor", 1, 2, 0},
		{"with_patch", 1, 2, 3},
		{"large", 100, 200, 300},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := NewSchemaVersion(tt.major, tt.minor, tt.patch)
			if err != nil {
				t.Fatalf("NewSchemaVersion(%d, %d, %d) unexpected error: %v", tt.major, tt.minor, tt.patch, err)
			}
			if v.Major() != tt.major {
				t.Errorf("Major() = %d, want %d", v.Major(), tt.major)
			}
			if v.Minor() != tt.minor {
				t.Errorf("Minor() = %d, want %d", v.Minor(), tt.minor)
			}
			if v.Patch() != tt.patch {
				t.Errorf("Patch() = %d, want %d", v.Patch(), tt.patch)
			}
		})
	}
}

func TestNewSchemaVersion_Invalid(t *testing.T) {
	tests := []struct {
		name  string
		major int
		minor int
		patch int
	}{
		{"negative_major", -1, 0, 0},
		{"negative_minor", 1, -1, 0},
		{"negative_patch", 1, 0, -1},
		{"all_negative", -1, -1, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewSchemaVersion(tt.major, tt.minor, tt.patch)
			if err == nil {
				t.Errorf("NewSchemaVersion(%d, %d, %d) expected error, got nil", tt.major, tt.minor, tt.patch)
			}
		})
	}
}

func TestMustNewSchemaVersion_Valid(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("MustNewSchemaVersion panicked unexpectedly: %v", r)
		}
	}()

	v := MustNewSchemaVersion(1, 2, 3)
	if v.String() != "1.2.3" {
		t.Errorf("MustNewSchemaVersion(1, 2, 3).String() = %q, want %q", v.String(), "1.2.3")
	}
}

func TestMustNewSchemaVersion_Invalid_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustNewSchemaVersion(-1, 0, 0) expected panic")
		}
	}()

	MustNewSchemaVersion(-1, 0, 0)
}

// ============================================================================
// InitialVersion Tests
// ============================================================================

func TestInitialVersion(t *testing.T) {
	v := InitialVersion()

	if v.Major() != 1 {
		t.Errorf("InitialVersion().Major() = %d, want 1", v.Major())
	}
	if v.Minor() != 0 {
		t.Errorf("InitialVersion().Minor() = %d, want 0", v.Minor())
	}
	if v.Patch() != 0 {
		t.Errorf("InitialVersion().Patch() = %d, want 0", v.Patch())
	}
	if v.String() != "1.0.0" {
		t.Errorf("InitialVersion().String() = %q, want %q", v.String(), "1.0.0")
	}
}

func TestInitialPreReleaseVersion(t *testing.T) {
	v := InitialPreReleaseVersion()

	if v.Major() != 0 {
		t.Errorf("InitialPreReleaseVersion().Major() = %d, want 0", v.Major())
	}
	if v.Minor() != 1 {
		t.Errorf("InitialPreReleaseVersion().Minor() = %d, want 1", v.Minor())
	}
	if v.Patch() != 0 {
		t.Errorf("InitialPreReleaseVersion().Patch() = %d, want 0", v.Patch())
	}
	if v.String() != "0.1.0" {
		t.Errorf("InitialPreReleaseVersion().String() = %q, want %q", v.String(), "0.1.0")
	}
}

// ============================================================================
// IsZero Tests
// ============================================================================

func TestSchemaVersion_IsZero(t *testing.T) {
	tests := []struct {
		name string
		v    SchemaVersion
		want bool
	}{
		{"zero_value", SchemaVersion{}, true},
		{"explicit_zeros", MustNewSchemaVersion(0, 0, 0), true},
		{"initial_version", InitialVersion(), false},
		{"pre_release", InitialPreReleaseVersion(), false},
		{"non_zero", MustParseSchemaVersion("1.2.3"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.v.IsZero(); got != tt.want {
				t.Errorf("IsZero() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ============================================================================
// IsPreRelease Tests
// ============================================================================

func TestSchemaVersion_IsPreRelease(t *testing.T) {
	tests := []struct {
		name string
		v    string
		want bool
	}{
		{"zero_major", "0.1.0", true},
		{"zero_all", "0.0.0", true},
		{"zero_major_high_minor", "0.99.99", true},
		{"stable", "1.0.0", false},
		{"higher_major", "2.0.0", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := MustParseSchemaVersion(tt.v)
			if got := v.IsPreRelease(); got != tt.want {
				t.Errorf("%s.IsPreRelease() = %v, want %v", tt.v, got, tt.want)
			}
		})
	}
}

// ============================================================================
// Comparison Tests
// ============================================================================

func TestSchemaVersion_Equals(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
		want bool
	}{
		{"equal", "1.2.3", "1.2.3", true},
		{"different_major", "1.2.3", "2.2.3", false},
		{"different_minor", "1.2.3", "1.3.3", false},
		{"different_patch", "1.2.3", "1.2.4", false},
		{"all_different", "1.2.3", "4.5.6", false},
		{"initial_versions", "1.0.0", "1.0.0", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := MustParseSchemaVersion(tt.a)
			b := MustParseSchemaVersion(tt.b)
			if got := a.Equals(b); got != tt.want {
				t.Errorf("%s.Equals(%s) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestSchemaVersion_Compare(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
		want int
	}{
		{"equal", "1.2.3", "1.2.3", 0},
		{"a_greater_major", "2.0.0", "1.9.9", 1},
		{"b_greater_major", "1.9.9", "2.0.0", -1},
		{"a_greater_minor", "1.3.0", "1.2.9", 1},
		{"b_greater_minor", "1.2.9", "1.3.0", -1},
		{"a_greater_patch", "1.2.4", "1.2.3", 1},
		{"b_greater_patch", "1.2.3", "1.2.4", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := MustParseSchemaVersion(tt.a)
			b := MustParseSchemaVersion(tt.b)
			if got := a.Compare(b); got != tt.want {
				t.Errorf("%s.Compare(%s) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestSchemaVersion_IsNewerThan(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
		want bool
	}{
		{"newer_major", "2.0.0", "1.0.0", true},
		{"newer_minor", "1.2.0", "1.1.0", true},
		{"newer_patch", "1.1.2", "1.1.1", true},
		{"equal", "1.2.3", "1.2.3", false},
		{"older", "1.0.0", "2.0.0", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := MustParseSchemaVersion(tt.a)
			b := MustParseSchemaVersion(tt.b)
			if got := a.IsNewerThan(b); got != tt.want {
				t.Errorf("%s.IsNewerThan(%s) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestSchemaVersion_IsOlderThan(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
		want bool
	}{
		{"older_major", "1.0.0", "2.0.0", true},
		{"older_minor", "1.1.0", "1.2.0", true},
		{"older_patch", "1.1.1", "1.1.2", true},
		{"equal", "1.2.3", "1.2.3", false},
		{"newer", "2.0.0", "1.0.0", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := MustParseSchemaVersion(tt.a)
			b := MustParseSchemaVersion(tt.b)
			if got := a.IsOlderThan(b); got != tt.want {
				t.Errorf("%s.IsOlderThan(%s) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

// ============================================================================
// Compatibility Tests
// ============================================================================

func TestSchemaVersion_IsCompatibleWith(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
		want bool
	}{
		{"same_version", "1.2.3", "1.2.3", true},
		{"same_major_minor", "1.2.0", "1.2.5", true},
		{"same_major_different_minor", "1.2.0", "1.3.0", true},
		{"different_major", "1.0.0", "2.0.0", false},
		{"zero_major_same", "0.1.0", "0.1.0", true},
		{"zero_major_different_minor", "0.1.0", "0.2.0", true}, // Same major (0)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := MustParseSchemaVersion(tt.a)
			b := MustParseSchemaVersion(tt.b)
			if got := a.IsCompatibleWith(b); got != tt.want {
				t.Errorf("%s.IsCompatibleWith(%s) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

// ============================================================================
// Next Version Tests
// ============================================================================

func TestSchemaVersion_NextMajor(t *testing.T) {
	v := MustParseSchemaVersion("1.2.3")
	next := v.NextMajor()

	if next.Major() != 2 {
		t.Errorf("NextMajor().Major() = %d, want 2", next.Major())
	}
	if next.Minor() != 0 {
		t.Errorf("NextMajor().Minor() = %d, want 0", next.Minor())
	}
	if next.Patch() != 0 {
		t.Errorf("NextMajor().Patch() = %d, want 0", next.Patch())
	}
	if next.String() != "2.0.0" {
		t.Errorf("NextMajor().String() = %q, want %q", next.String(), "2.0.0")
	}

	// Original should be unchanged
	if v.String() != "1.2.3" {
		t.Errorf("Original version mutated: %q", v.String())
	}
}

func TestSchemaVersion_NextMinor(t *testing.T) {
	v := MustParseSchemaVersion("1.2.3")
	next := v.NextMinor()

	if next.Major() != 1 {
		t.Errorf("NextMinor().Major() = %d, want 1", next.Major())
	}
	if next.Minor() != 3 {
		t.Errorf("NextMinor().Minor() = %d, want 3", next.Minor())
	}
	if next.Patch() != 0 {
		t.Errorf("NextMinor().Patch() = %d, want 0", next.Patch())
	}
	if next.String() != "1.3.0" {
		t.Errorf("NextMinor().String() = %q, want %q", next.String(), "1.3.0")
	}

	// Original should be unchanged
	if v.String() != "1.2.3" {
		t.Errorf("Original version mutated: %q", v.String())
	}
}

func TestSchemaVersion_NextPatch(t *testing.T) {
	v := MustParseSchemaVersion("1.2.3")
	next := v.NextPatch()

	if next.Major() != 1 {
		t.Errorf("NextPatch().Major() = %d, want 1", next.Major())
	}
	if next.Minor() != 2 {
		t.Errorf("NextPatch().Minor() = %d, want 2", next.Minor())
	}
	if next.Patch() != 4 {
		t.Errorf("NextPatch().Patch() = %d, want 4", next.Patch())
	}
	if next.String() != "1.2.4" {
		t.Errorf("NextPatch().String() = %q, want %q", next.String(), "1.2.4")
	}

	// Original should be unchanged
	if v.String() != "1.2.3" {
		t.Errorf("Original version mutated: %q", v.String())
	}
}

// ============================================================================
// GoString Tests
// ============================================================================

func TestSchemaVersion_GoString(t *testing.T) {
	v := MustParseSchemaVersion("1.2.3")
	expected := "SchemaVersion{1.2.3}"

	if got := v.GoString(); got != expected {
		t.Errorf("GoString() = %q, want %q", got, expected)
	}
}
