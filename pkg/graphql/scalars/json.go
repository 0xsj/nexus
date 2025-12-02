// pkg/graphql/scalars/json.go

package scalars

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/99designs/gqlgen/graphql"
)

// ============================================================================
// JSON Scalar
// ============================================================================

// JSON represents arbitrary JSON data.
type JSON map[string]interface{}

// MarshalJSON marshals a JSON map to a GraphQL scalar.
func MarshalJSON(j JSON) graphql.Marshaler {
	if j == nil {
		return graphql.Null
	}

	return graphql.WriterFunc(func(w io.Writer) {
		enc := json.NewEncoder(w)
		enc.Encode(j)
	})
}

// UnmarshalJSON unmarshals a GraphQL scalar to a JSON map.
func UnmarshalJSON(v interface{}) (JSON, error) {
	switch v := v.(type) {
	case map[string]interface{}:
		return JSON(v), nil
	case JSON:
		return v, nil
	case string:
		var result JSON
		if err := json.Unmarshal([]byte(v), &result); err != nil {
			return nil, fmt.Errorf("invalid json string: %w", err)
		}
		return result, nil
	case []byte:
		var result JSON
		if err := json.Unmarshal(v, &result); err != nil {
			return nil, fmt.Errorf("invalid json bytes: %w", err)
		}
		return result, nil
	case nil:
		return nil, nil
	default:
		return nil, fmt.Errorf("json must be a map or string, got %T", v)
	}
}

// ============================================================================
// Any Scalar (arbitrary JSON value)
// ============================================================================

// Any represents any JSON value (object, array, string, number, boolean, null).
type Any interface{}

// MarshalAny marshals any value to a GraphQL scalar.
func MarshalAny(v Any) graphql.Marshaler {
	if v == nil {
		return graphql.Null
	}

	return graphql.WriterFunc(func(w io.Writer) {
		enc := json.NewEncoder(w)
		enc.Encode(v)
	})
}

// UnmarshalAny unmarshals a GraphQL scalar to any value.
func UnmarshalAny(v interface{}) (Any, error) {
	return v, nil
}

// ============================================================================
// JSONArray Scalar
// ============================================================================

// JSONArray represents a JSON array.
type JSONArray []interface{}

// MarshalJSONArray marshals a JSON array to a GraphQL scalar.
func MarshalJSONArray(a JSONArray) graphql.Marshaler {
	if a == nil {
		return graphql.Null
	}

	return graphql.WriterFunc(func(w io.Writer) {
		enc := json.NewEncoder(w)
		enc.Encode(a)
	})
}

// UnmarshalJSONArray unmarshals a GraphQL scalar to a JSON array.
func UnmarshalJSONArray(v interface{}) (JSONArray, error) {
	switch v := v.(type) {
	case []interface{}:
		return JSONArray(v), nil
	case JSONArray:
		return v, nil
	case string:
		var result JSONArray
		if err := json.Unmarshal([]byte(v), &result); err != nil {
			return nil, fmt.Errorf("invalid json array string: %w", err)
		}
		return result, nil
	case []byte:
		var result JSONArray
		if err := json.Unmarshal(v, &result); err != nil {
			return nil, fmt.Errorf("invalid json array bytes: %w", err)
		}
		return result, nil
	case nil:
		return nil, nil
	default:
		return nil, fmt.Errorf("json array must be an array or string, got %T", v)
	}
}

// ============================================================================
// Map Scalar (string keys, string values)
// ============================================================================

// Map represents a simple string-to-string map.
type Map map[string]string

// MarshalMap marshals a Map to a GraphQL scalar.
func MarshalMap(m Map) graphql.Marshaler {
	if m == nil {
		return graphql.Null
	}

	return graphql.WriterFunc(func(w io.Writer) {
		enc := json.NewEncoder(w)
		enc.Encode(m)
	})
}

// UnmarshalMap unmarshals a GraphQL scalar to a Map.
func UnmarshalMap(v interface{}) (Map, error) {
	switch v := v.(type) {
	case map[string]interface{}:
		result := make(Map)
		for k, val := range v {
			if str, ok := val.(string); ok {
				result[k] = str
			} else {
				result[k] = fmt.Sprintf("%v", val)
			}
		}
		return result, nil
	case map[string]string:
		return Map(v), nil
	case Map:
		return v, nil
	case string:
		var result Map
		if err := json.Unmarshal([]byte(v), &result); err != nil {
			return nil, fmt.Errorf("invalid map string: %w", err)
		}
		return result, nil
	case nil:
		return nil, nil
	default:
		return nil, fmt.Errorf("map must be an object or string, got %T", v)
	}
}

// ============================================================================
// RawJSON Scalar (preserves raw JSON)
// ============================================================================

// RawJSON represents raw JSON that should not be parsed.
type RawJSON json.RawMessage

// MarshalRawJSON marshals RawJSON to a GraphQL scalar.
func MarshalRawJSON(r RawJSON) graphql.Marshaler {
	if r == nil {
		return graphql.Null
	}

	return graphql.WriterFunc(func(w io.Writer) {
		w.Write(r)
	})
}

// UnmarshalRawJSON unmarshals a GraphQL scalar to RawJSON.
func UnmarshalRawJSON(v interface{}) (RawJSON, error) {
	switch v := v.(type) {
	case []byte:
		return RawJSON(v), nil
	case string:
		return RawJSON(v), nil
	case json.RawMessage:
		return RawJSON(v), nil
	case RawJSON:
		return v, nil
	case nil:
		return nil, nil
	default:
		// Marshal other types to JSON
		b, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("cannot marshal to raw json: %w", err)
		}
		return RawJSON(b), nil
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

// ToJSON converts any value to JSON type.
func ToJSON(v interface{}) (JSON, error) {
	if v == nil {
		return nil, nil
	}

	// Already JSON
	if j, ok := v.(JSON); ok {
		return j, nil
	}

	// Map
	if m, ok := v.(map[string]interface{}); ok {
		return JSON(m), nil
	}

	// Marshal and unmarshal for other types
	b, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("cannot convert to json: %w", err)
	}

	var result JSON
	if err := json.Unmarshal(b, &result); err != nil {
		return nil, fmt.Errorf("cannot convert to json: %w", err)
	}

	return result, nil
}

// MustToJSON converts any value to JSON or panics.
func MustToJSON(v interface{}) JSON {
	j, err := ToJSON(v)
	if err != nil {
		panic(err)
	}
	return j
}
