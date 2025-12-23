package did

import (
	"fmt"
	"strings"
)

// DID represents a Decentralized Identifier.
// Format: did:<method>:<method-specific-id>
//
// Examples:
//   - did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK
//   - did:web:example.com
//   - did:pkh:eip155:1:0xab16a96d359ec26a11e2c2b3d8f8b8942d5bfcdb
type DID struct {
	// method is the DID method (key, web, pkh, etc.)
	method Method

	// methodSpecificID is the method-specific identifier
	methodSpecificID string

	// raw is the original DID string
	raw string
}

// Parse parses a DID string into a DID struct.
func Parse(s string) (DID, error) {
	const op = "did.Parse"

	s = strings.TrimSpace(s)
	if s == "" {
		return DID{}, ErrInvalidDID(op, "DID cannot be empty")
	}

	// Must start with "did:"
	if !strings.HasPrefix(s, "did:") {
		return DID{}, ErrInvalidDID(op, "DID must start with 'did:'")
	}

	// Split into parts: did:<method>:<method-specific-id>
	parts := strings.SplitN(s, ":", 3)
	if len(parts) < 3 {
		return DID{}, ErrInvalidDID(op, "DID must have format 'did:<method>:<method-specific-id>'")
	}

	methodStr := parts[1]
	methodSpecificID := parts[2]

	if methodStr == "" {
		return DID{}, ErrInvalidDID(op, "DID method cannot be empty")
	}

	if methodSpecificID == "" {
		return DID{}, ErrInvalidDID(op, "DID method-specific-id cannot be empty")
	}

	method := Method(methodStr)
	if err := method.Validate(); err != nil {
		return DID{}, ErrInvalidDID(op, err.Error())
	}

	return DID{
		method:           method,
		methodSpecificID: methodSpecificID,
		raw:              s,
	}, nil
}

// MustParse parses a DID string and panics if invalid.
// Only use for constants or tests.
func MustParse(s string) DID {
	d, err := Parse(s)
	if err != nil {
		panic(fmt.Sprintf("invalid DID: %v", err))
	}
	return d
}

// New creates a new DID from method and method-specific ID.
func New(method Method, methodSpecificID string) (DID, error) {
	const op = "did.New"

	if err := method.Validate(); err != nil {
		return DID{}, ErrInvalidDID(op, err.Error())
	}

	if methodSpecificID == "" {
		return DID{}, ErrInvalidDID(op, "method-specific-id cannot be empty")
	}

	raw := fmt.Sprintf("did:%s:%s", method, methodSpecificID)

	return DID{
		method:           method,
		methodSpecificID: methodSpecificID,
		raw:              raw,
	}, nil
}

// String returns the full DID string.
func (d DID) String() string {
	return d.raw
}

// Method returns the DID method.
func (d DID) Method() Method {
	return d.method
}

// MethodSpecificID returns the method-specific identifier.
func (d DID) MethodSpecificID() string {
	return d.methodSpecificID
}

// IsZero returns true if the DID is empty/unset.
func (d DID) IsZero() bool {
	return d.raw == ""
}

// IsValid returns true if the DID is valid (non-empty).
func (d DID) IsValid() bool {
	return d.raw != ""
}

// Equals checks if two DIDs are equal.
func (d DID) Equals(other DID) bool {
	return d.raw == other.raw
}

// URI returns the DID as a URI string.
// This is the same as String() since DIDs are already URIs.
func (d DID) URI() string {
	return d.raw
}

// Fragment returns a DID URL with a fragment.
// Example: did:key:z6Mk...#key-1
func (d DID) Fragment(fragment string) string {
	if fragment == "" {
		return d.raw
	}
	return d.raw + "#" + fragment
}

// Path returns a DID URL with a path.
// Example: did:web:example.com/path/to/resource
func (d DID) Path(path string) string {
	if path == "" {
		return d.raw
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return d.raw + path
}

// ============================================================================
// JSON Marshaling
// ============================================================================

// MarshalJSON implements json.Marshaler.
func (d DID) MarshalJSON() ([]byte, error) {
	if d.IsZero() {
		return []byte("null"), nil
	}
	return []byte(`"` + d.raw + `"`), nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (d *DID) UnmarshalJSON(data []byte) error {
	// Handle null
	if string(data) == "null" {
		*d = DID{}
		return nil
	}

	// Remove quotes
	s := strings.Trim(string(data), `"`)
	if s == "" {
		*d = DID{}
		return nil
	}

	parsed, err := Parse(s)
	if err != nil {
		return err
	}

	*d = parsed
	return nil
}

// ============================================================================
// Text Marshaling
// ============================================================================

// MarshalText implements encoding.TextMarshaler.
func (d DID) MarshalText() ([]byte, error) {
	if d.IsZero() {
		return []byte{}, nil
	}
	return []byte(d.raw), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (d *DID) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		*d = DID{}
		return nil
	}

	parsed, err := Parse(string(text))
	if err != nil {
		return err
	}

	*d = parsed
	return nil
}

// ============================================================================
// SQL Scanning
// ============================================================================

// Scan implements sql.Scanner.
func (d *DID) Scan(value interface{}) error {
	if value == nil {
		*d = DID{}
		return nil
	}

	switch v := value.(type) {
	case string:
		if v == "" {
			*d = DID{}
			return nil
		}
		parsed, err := Parse(v)
		if err != nil {
			return err
		}
		*d = parsed
		return nil

	case []byte:
		if len(v) == 0 {
			*d = DID{}
			return nil
		}
		parsed, err := Parse(string(v))
		if err != nil {
			return err
		}
		*d = parsed
		return nil

	default:
		return fmt.Errorf("cannot scan %T into DID", value)
	}
}

// Value implements driver.Valuer.
func (d DID) Value() (interface{}, error) {
	if d.IsZero() {
		return nil, nil
	}
	return d.raw, nil
}
