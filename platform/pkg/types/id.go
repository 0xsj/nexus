package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// ID is a strongly-typed, globally unique identifier.
// Uses UUIDv7 (RFC 9562) which embeds a timestamp for natural ordering
// and efficient database indexing.
//
// Zero value is invalid — use NewID() to create valid IDs.
type ID struct {
	value uuid.UUID
}

// NewID generates a new time-ordered unique ID (UUIDv7).
func NewID() ID {
	return ID{
		value: uuid.Must(uuid.NewV7()),
	}
}

// ParseID parses a string into an ID.
// Accepts standard UUID format (with or without hyphens).
func ParseID(s string) (ID, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return ID{}, fmt.Errorf("id cannot be empty")
	}

	parsed, err := uuid.Parse(s)
	if err != nil {
		return ID{}, fmt.Errorf("invalid id format: %w", err)
	}

	return ID{value: parsed}, nil
}

// MustParseID parses a string into an ID and panics if invalid.
// Only use for constants or tests.
func MustParseID(s string) ID {
	id, err := ParseID(s)
	if err != nil {
		panic(fmt.Sprintf("invalid id: %v", err))
	}
	return id
}

// FromUUID creates an ID from an existing UUID.
func FromUUID(u uuid.UUID) ID {
	return ID{value: u}
}

// String returns the canonical string representation.
func (id ID) String() string {
	return id.value.String()
}

// UUID returns the underlying UUID value.
func (id ID) UUID() uuid.UUID {
	return id.value
}

// IsZero returns true if the ID is the zero value.
func (id ID) IsZero() bool {
	return id.value == uuid.Nil
}

// IsValid returns true if the ID is valid (non-zero).
func (id ID) IsValid() bool {
	return id.value != uuid.Nil
}

// Equals checks if two IDs are equal.
func (id ID) Equals(other ID) bool {
	return id.value == other.value
}

// Compare compares two IDs.
// For UUIDv7, this provides chronological ordering.
// Returns -1 if id < other, 0 if equal, +1 if id > other.
func (id ID) Compare(other ID) int {
	for i := range 16 {
		if id.value[i] < other.value[i] {
			return -1
		}
		if id.value[i] > other.value[i] {
			return 1
		}
	}
	return 0
}

// Time returns the Unix timestamp (milliseconds) embedded in UUIDv7.
// Returns 0 for zero IDs or non-v7 UUIDs.
func (id ID) Time() int64 {
	if id.IsZero() {
		return 0
	}

	// UUIDv7 timestamp is in first 48 bits (6 bytes)
	timestamp := int64(id.value[0])<<40 |
		int64(id.value[1])<<32 |
		int64(id.value[2])<<24 |
		int64(id.value[3])<<16 |
		int64(id.value[4])<<8 |
		int64(id.value[5])

	return timestamp
}

// Version returns the UUID version.
func (id ID) Version() uuid.Version {
	return id.value.Version()
}

// Bytes returns the raw 16-byte representation.
func (id ID) Bytes() []byte {
	return id.value[:]
}

// ============================================================================
// JSON Marshaling
// ============================================================================

// MarshalJSON implements json.Marshaler.
func (id ID) MarshalJSON() ([]byte, error) {
	if id.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(id.String())
}

// UnmarshalJSON implements json.Unmarshaler.
func (id *ID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	if s == "" || s == "null" {
		*id = ID{}
		return nil
	}

	parsed, err := ParseID(s)
	if err != nil {
		return err
	}

	*id = parsed
	return nil
}

// ============================================================================
// SQL Scanning
// ============================================================================

// Scan implements sql.Scanner.
func (id *ID) Scan(value any) error {
	if value == nil {
		*id = ID{}
		return nil
	}

	switch v := value.(type) {
	case string:
		if v == "" {
			*id = ID{}
			return nil
		}
		parsed, err := ParseID(v)
		if err != nil {
			return err
		}
		*id = parsed
		return nil

	case []byte:
		if len(v) == 0 {
			*id = ID{}
			return nil
		}

		// Try parsing as string first (36 with hyphens, 32 without)
		if len(v) == 36 || len(v) == 32 {
			parsed, err := ParseID(string(v))
			if err != nil {
				return err
			}
			*id = parsed
			return nil
		}

		// Try parsing as 16-byte binary UUID
		if len(v) == 16 {
			var u uuid.UUID
			copy(u[:], v)
			*id = ID{value: u}
			return nil
		}

		return fmt.Errorf("invalid byte length for UUID: %d", len(v))

	case uuid.UUID:
		*id = ID{value: v}
		return nil

	default:
		return fmt.Errorf("cannot scan %T into ID", value)
	}
}

// Value implements driver.Valuer.
func (id ID) Value() (driver.Value, error) {
	if id.IsZero() {
		return nil, nil
	}
	return id.value, nil
}

// ============================================================================
// Text Marshaling
// ============================================================================

// MarshalText implements encoding.TextMarshaler.
func (id ID) MarshalText() ([]byte, error) {
	if id.IsZero() {
		return []byte{}, nil
	}
	return []byte(id.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (id *ID) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		*id = ID{}
		return nil
	}

	parsed, err := ParseID(string(text))
	if err != nil {
		return err
	}

	*id = parsed
	return nil
}

// ============================================================================
// Helpers
// ============================================================================

// GoString implements fmt.GoStringer for debugging.
func (id ID) GoString() string {
	if id.IsZero() {
		return "ID{zero}"
	}
	return fmt.Sprintf("ID{%s}", id.String())
}
