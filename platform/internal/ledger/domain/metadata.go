package domain

import (
	"encoding/json"
	"errors"

	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
)

// ============================================================================
// Metadata Error Codes
// ============================================================================

const (
	CodeInvalidMetadata pkgerrors.Code = "LEDGER_INVALID_METADATA"
)

// ============================================================================
// Metadata Sentinel Errors
// ============================================================================

var (
	ErrInvalidMetadata = errors.New("invalid metadata")
)

// ============================================================================
// Metadata Error Constructors
// ============================================================================

// InvalidMetadata creates an invalid metadata error.
func InvalidMetadata(operation string, reason string) *pkgerrors.Error {
	return pkgerrors.Validation(operation, "invalid metadata: "+reason).
		WithCode(CodeInvalidMetadata)
}

// ============================================================================
// Metadata Value Object
// ============================================================================

// Metadata contains additional context about an audit entry.
// It is a flexible key-value store for event-specific details.
type Metadata struct {
	values map[string]any
}

// NewMetadata creates a new empty Metadata.
func NewMetadata() Metadata {
	return Metadata{
		values: make(map[string]any),
	}
}

// MetadataFromMap creates Metadata from an existing map.
func MetadataFromMap(m map[string]any) Metadata {
	if m == nil {
		return NewMetadata()
	}

	// Create a copy to prevent external mutation
	values := make(map[string]any, len(m))
	for k, v := range m {
		values[k] = v
	}

	return Metadata{values: values}
}

// MetadataFromJSON creates Metadata from a JSON byte slice.
func MetadataFromJSON(data []byte) (Metadata, error) {
	if len(data) == 0 || string(data) == "null" {
		return NewMetadata(), nil
	}

	var values map[string]any
	if err := json.Unmarshal(data, &values); err != nil {
		return Metadata{}, InvalidMetadata("domain.MetadataFromJSON", err.Error())
	}

	return Metadata{values: values}, nil
}

// Set returns a new Metadata with the given key-value pair added.
// Metadata is immutable; this returns a copy.
func (m Metadata) Set(key string, value any) Metadata {
	newValues := make(map[string]any, len(m.values)+1)
	for k, v := range m.values {
		newValues[k] = v
	}
	newValues[key] = value
	return Metadata{values: newValues}
}

// Get retrieves a value by key.
// Returns nil if the key doesn't exist.
func (m Metadata) Get(key string) any {
	if m.values == nil {
		return nil
	}
	return m.values[key]
}

// GetString retrieves a string value by key.
// Returns empty string if the key doesn't exist or isn't a string.
func (m Metadata) GetString(key string) string {
	v := m.Get(key)
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// GetInt retrieves an int value by key.
// Returns 0 if the key doesn't exist or isn't numeric.
func (m Metadata) GetInt(key string) int {
	v := m.Get(key)
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	default:
		return 0
	}
}

// GetBool retrieves a bool value by key.
// Returns false if the key doesn't exist or isn't a bool.
func (m Metadata) GetBool(key string) bool {
	v := m.Get(key)
	if b, ok := v.(bool); ok {
		return b
	}
	return false
}

// Has checks if a key exists in the metadata.
func (m Metadata) Has(key string) bool {
	if m.values == nil {
		return false
	}
	_, exists := m.values[key]
	return exists
}

// Keys returns all keys in the metadata.
func (m Metadata) Keys() []string {
	if m.values == nil {
		return nil
	}
	keys := make([]string, 0, len(m.values))
	for k := range m.values {
		keys = append(keys, k)
	}
	return keys
}

// Len returns the number of key-value pairs.
func (m Metadata) Len() int {
	if m.values == nil {
		return 0
	}
	return len(m.values)
}

// IsEmpty returns true if the metadata has no values.
func (m Metadata) IsEmpty() bool {
	return m.Len() == 0
}

// ToMap returns a copy of the underlying map.
func (m Metadata) ToMap() map[string]any {
	if m.values == nil {
		return make(map[string]any)
	}

	result := make(map[string]any, len(m.values))
	for k, v := range m.values {
		result[k] = v
	}
	return result
}

// ToJSON serializes the metadata to JSON.
func (m Metadata) ToJSON() ([]byte, error) {
	if m.values == nil || len(m.values) == 0 {
		return []byte("{}"), nil
	}
	return json.Marshal(m.values)
}

// MarshalJSON implements json.Marshaler.
func (m Metadata) MarshalJSON() ([]byte, error) {
	return m.ToJSON()
}

// UnmarshalJSON implements json.Unmarshaler.
func (m *Metadata) UnmarshalJSON(data []byte) error {
	meta, err := MetadataFromJSON(data)
	if err != nil {
		return err
	}
	*m = meta
	return nil
}

// Merge returns a new Metadata with values from both.
// Values from other override values from m if keys conflict.
func (m Metadata) Merge(other Metadata) Metadata {
	newValues := make(map[string]any, len(m.values)+len(other.values))

	for k, v := range m.values {
		newValues[k] = v
	}
	for k, v := range other.values {
		newValues[k] = v
	}

	return Metadata{values: newValues}
}

// ============================================================================
// Common Metadata Keys
// ============================================================================

// Common metadata keys used across audit entries.
const (
	// MetaKeyCorrelationID links related operations.
	MetaKeyCorrelationID = "correlation_id"

	// MetaKeyCausationID is the ID of the event/command that caused this event.
	MetaKeyCausationID = "causation_id"

	// MetaKeyIPAddress is the client IP address.
	MetaKeyIPAddress = "ip_address"

	// MetaKeyUserAgent is the client user agent string.
	MetaKeyUserAgent = "user_agent"

	// MetaKeyRequestID is the HTTP request ID.
	MetaKeyRequestID = "request_id"

	// MetaKeyReason provides context for why an action occurred.
	MetaKeyReason = "reason"

	// MetaKeyPreviousValue holds the previous value before a change.
	MetaKeyPreviousValue = "previous_value"

	// MetaKeyNewValue holds the new value after a change.
	MetaKeyNewValue = "new_value"
)

// WithCorrelationID returns Metadata with correlation ID set.
func (m Metadata) WithCorrelationID(id string) Metadata {
	return m.Set(MetaKeyCorrelationID, id)
}

// WithCausationID returns Metadata with causation ID set.
func (m Metadata) WithCausationID(id string) Metadata {
	return m.Set(MetaKeyCausationID, id)
}

// WithIPAddress returns Metadata with IP address set.
func (m Metadata) WithIPAddress(ip string) Metadata {
	return m.Set(MetaKeyIPAddress, ip)
}

// WithUserAgent returns Metadata with user agent set.
func (m Metadata) WithUserAgent(ua string) Metadata {
	return m.Set(MetaKeyUserAgent, ua)
}

// WithRequestID returns Metadata with request ID set.
func (m Metadata) WithRequestID(id string) Metadata {
	return m.Set(MetaKeyRequestID, id)
}

// WithReason returns Metadata with reason set.
func (m Metadata) WithReason(reason string) Metadata {
	return m.Set(MetaKeyReason, reason)
}

// CorrelationID returns the correlation ID if present.
func (m Metadata) CorrelationID() string {
	return m.GetString(MetaKeyCorrelationID)
}

// CausationID returns the causation ID if present.
func (m Metadata) CausationID() string {
	return m.GetString(MetaKeyCausationID)
}
