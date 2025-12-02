// pkg/graphql/scalars/uuid.go

package scalars

import (
	"fmt"
	"io"
	"strconv"

	"github.com/99designs/gqlgen/graphql"
	"github.com/google/uuid"
)

// ============================================================================
// UUID Scalar
// ============================================================================

// MarshalUUID marshals a uuid.UUID to a GraphQL scalar.
func MarshalUUID(id uuid.UUID) graphql.Marshaler {
	if id == uuid.Nil {
		return graphql.Null
	}

	return graphql.WriterFunc(func(w io.Writer) {
		io.WriteString(w, strconv.Quote(id.String()))
	})
}

// UnmarshalUUID unmarshals a GraphQL scalar to uuid.UUID.
func UnmarshalUUID(v interface{}) (uuid.UUID, error) {
	switch v := v.(type) {
	case string:
		return parseUUID(v)
	case []byte:
		return parseUUID(string(v))
	case nil:
		return uuid.Nil, nil
	default:
		return uuid.Nil, fmt.Errorf("uuid must be a string, got %T", v)
	}
}

// parseUUID parses a string as a UUID.
func parseUUID(s string) (uuid.UUID, error) {
	if s == "" {
		return uuid.Nil, nil
	}

	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid uuid: %s", s)
	}

	return id, nil
}

// ============================================================================
// ID Scalar (string-based, compatible with pkg/types.ID)
// ============================================================================

// MarshalID marshals a string ID to a GraphQL scalar.
// This is for compatibility with the built-in ID type but with validation.
func MarshalID(id string) graphql.Marshaler {
	if id == "" {
		return graphql.Null
	}

	return graphql.WriterFunc(func(w io.Writer) {
		io.WriteString(w, strconv.Quote(id))
	})
}

// UnmarshalID unmarshals a GraphQL scalar to a string ID.
func UnmarshalID(v interface{}) (string, error) {
	switch v := v.(type) {
	case string:
		if v == "" {
			return "", nil
		}
		// Validate it's a valid UUID format
		if _, err := uuid.Parse(v); err != nil {
			return "", fmt.Errorf("invalid id format: %s", v)
		}
		return v, nil
	case []byte:
		return UnmarshalID(string(v))
	case nil:
		return "", nil
	default:
		return "", fmt.Errorf("id must be a string, got %T", v)
	}
}

// ============================================================================
// NullableUUID (for optional UUID fields)
// ============================================================================

// NullableUUID represents a UUID that can be null.
type NullableUUID struct {
	UUID  uuid.UUID
	Valid bool
}

// MarshalNullableUUID marshals a NullableUUID to a GraphQL scalar.
func MarshalNullableUUID(n NullableUUID) graphql.Marshaler {
	if !n.Valid || n.UUID == uuid.Nil {
		return graphql.Null
	}

	return graphql.WriterFunc(func(w io.Writer) {
		io.WriteString(w, strconv.Quote(n.UUID.String()))
	})
}

// UnmarshalNullableUUID unmarshals a GraphQL scalar to NullableUUID.
func UnmarshalNullableUUID(v interface{}) (NullableUUID, error) {
	if v == nil {
		return NullableUUID{Valid: false}, nil
	}

	id, err := UnmarshalUUID(v)
	if err != nil {
		return NullableUUID{}, err
	}

	return NullableUUID{
		UUID:  id,
		Valid: id != uuid.Nil,
	}, nil
}

// ============================================================================
// Helper Functions
// ============================================================================

// NewUUID generates a new random UUID.
func NewUUID() uuid.UUID {
	return uuid.New()
}

// MustParseUUID parses a UUID string and panics if invalid.
// Use only for known-valid UUIDs (e.g., constants).
func MustParseUUID(s string) uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		panic(fmt.Sprintf("invalid uuid: %s", s))
	}
	return id
}

// IsValidUUID checks if a string is a valid UUID.
func IsValidUUID(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}
