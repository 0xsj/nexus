package domain

import (
	"fmt"
	"strconv"
	"strings"
)

// SchemaVersion represents a semantic version for a schema.
// Follows semver conventions: MAJOR.MINOR.PATCH
//
// Version semantics:
//   - MAJOR: Breaking changes (removed required claims, type changes)
//   - MINOR: Backwards-compatible additions (new optional claims)
//   - PATCH: Bug fixes, documentation (no schema changes)
type SchemaVersion struct {
	major int
	minor int
	patch int
}

// ============================================================================
// Constructors
// ============================================================================

// NewSchemaVersion creates a new SchemaVersion with validation.
// All components must be non-negative.
func NewSchemaVersion(major, minor, patch int) (SchemaVersion, error) {
	if major < 0 {
		return SchemaVersion{}, ErrInvalidVersionFormat
	}
	if minor < 0 {
		return SchemaVersion{}, ErrInvalidVersionFormat
	}
	if patch < 0 {
		return SchemaVersion{}, ErrInvalidVersionFormat
	}

	return SchemaVersion{
		major: major,
		minor: minor,
		patch: patch,
	}, nil
}

// MustNewSchemaVersion creates a new SchemaVersion and panics if invalid.
// Only use for constants or tests.
func MustNewSchemaVersion(major, minor, patch int) SchemaVersion {
	v, err := NewSchemaVersion(major, minor, patch)
	if err != nil {
		panic(fmt.Sprintf("invalid version: %d.%d.%d", major, minor, patch))
	}
	return v
}

// ParseSchemaVersion parses a version string in "MAJOR.MINOR.PATCH" format.
// Examples: "1.0.0", "2.1.3", "0.1.0"
func ParseSchemaVersion(s string) (SchemaVersion, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return SchemaVersion{}, ErrInvalidVersionFormat
	}

	// Remove optional "v" prefix
	s = strings.TrimPrefix(s, "v")
	s = strings.TrimPrefix(s, "V")

	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return SchemaVersion{}, ErrInvalidVersionFormat
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return SchemaVersion{}, ErrInvalidVersionFormat
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return SchemaVersion{}, ErrInvalidVersionFormat
	}

	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return SchemaVersion{}, ErrInvalidVersionFormat
	}

	return NewSchemaVersion(major, minor, patch)
}

// MustParseSchemaVersion parses a version string and panics if invalid.
// Only use for constants or tests.
func MustParseSchemaVersion(s string) SchemaVersion {
	v, err := ParseSchemaVersion(s)
	if err != nil {
		panic("invalid version: " + s)
	}
	return v
}

// ============================================================================
// Accessors
// ============================================================================

// Major returns the major version component.
func (v SchemaVersion) Major() int {
	return v.major
}

// Minor returns the minor version component.
func (v SchemaVersion) Minor() int {
	return v.minor
}

// Patch returns the patch version component.
func (v SchemaVersion) Patch() int {
	return v.patch
}

// ============================================================================
// String Representation
// ============================================================================

// String returns the version in "MAJOR.MINOR.PATCH" format.
func (v SchemaVersion) String() string {
	return fmt.Sprintf("%d.%d.%d", v.major, v.minor, v.patch)
}

// GoString implements fmt.GoStringer for debugging.
func (v SchemaVersion) GoString() string {
	return fmt.Sprintf("SchemaVersion{%d.%d.%d}", v.major, v.minor, v.patch)
}

// ============================================================================
// Comparison
// ============================================================================

// IsZero returns true if the version is the zero value.
func (v SchemaVersion) IsZero() bool {
	return v.major == 0 && v.minor == 0 && v.patch == 0
}

// Equals returns true if two versions are equal.
func (v SchemaVersion) Equals(other SchemaVersion) bool {
	return v.major == other.major &&
		v.minor == other.minor &&
		v.patch == other.patch
}

// Compare compares two versions.
// Returns -1 if v < other, 0 if equal, +1 if v > other.
func (v SchemaVersion) Compare(other SchemaVersion) int {
	if v.major != other.major {
		if v.major < other.major {
			return -1
		}
		return 1
	}

	if v.minor != other.minor {
		if v.minor < other.minor {
			return -1
		}
		return 1
	}

	if v.patch != other.patch {
		if v.patch < other.patch {
			return -1
		}
		return 1
	}

	return 0
}

// IsNewerThan returns true if v is newer than other.
func (v SchemaVersion) IsNewerThan(other SchemaVersion) bool {
	return v.Compare(other) > 0
}

// IsOlderThan returns true if v is older than other.
func (v SchemaVersion) IsOlderThan(other SchemaVersion) bool {
	return v.Compare(other) < 0
}

// ============================================================================
// Compatibility
// ============================================================================

// IsCompatibleWith returns true if v is backwards-compatible with other.
// Versions are compatible if they share the same major version.
// Example: 1.2.0 is compatible with 1.0.0, but 2.0.0 is not.
func (v SchemaVersion) IsCompatibleWith(other SchemaVersion) bool {
	return v.major == other.major
}

// IsPreRelease returns true if this is a pre-release version (0.x.x).
func (v SchemaVersion) IsPreRelease() bool {
	return v.major == 0
}

// ============================================================================
// Version Bumping
// ============================================================================

// NextMajor returns a new version with major incremented and minor/patch reset.
// Example: 1.2.3 → 2.0.0
func (v SchemaVersion) NextMajor() SchemaVersion {
	return SchemaVersion{
		major: v.major + 1,
		minor: 0,
		patch: 0,
	}
}

// NextMinor returns a new version with minor incremented and patch reset.
// Example: 1.2.3 → 1.3.0
func (v SchemaVersion) NextMinor() SchemaVersion {
	return SchemaVersion{
		major: v.major,
		minor: v.minor + 1,
		patch: 0,
	}
}

// NextPatch returns a new version with patch incremented.
// Example: 1.2.3 → 1.2.4
func (v SchemaVersion) NextPatch() SchemaVersion {
	return SchemaVersion{
		major: v.major,
		minor: v.minor,
		patch: v.patch + 1,
	}
}

// ============================================================================
// Initial Versions
// ============================================================================

// InitialVersion returns the initial stable version (1.0.0).
func InitialVersion() SchemaVersion {
	return SchemaVersion{major: 1, minor: 0, patch: 0}
}

// InitialPreReleaseVersion returns the initial pre-release version (0.1.0).
func InitialPreReleaseVersion() SchemaVersion {
	return SchemaVersion{major: 0, minor: 1, patch: 0}
}
