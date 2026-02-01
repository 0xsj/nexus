package domain

import (
	"encoding/json"
	"testing"
)

// ============================================================================
// NewMetadata Tests
// ============================================================================

func TestNewMetadata_Empty(t *testing.T) {
	m := NewMetadata()

	if m.Len() != 0 {
		t.Errorf("expected empty Metadata, got %d entries", m.Len())
	}
	if !m.IsEmpty() {
		t.Error("expected IsEmpty() to be true")
	}
}

// ============================================================================
// MetadataFromMap Tests
// ============================================================================

func TestMetadataFromMap_Valid(t *testing.T) {
	input := map[string]any{
		"key1": "value1",
		"key2": 42,
		"key3": true,
	}

	m := MetadataFromMap(input)

	if m.Get("key1") != "value1" {
		t.Errorf("Get(key1) = %v, want %v", m.Get("key1"), "value1")
	}
	if m.Get("key2") != 42 {
		t.Errorf("Get(key2) = %v, want %v", m.Get("key2"), 42)
	}
	if m.Get("key3") != true {
		t.Errorf("Get(key3) = %v, want %v", m.Get("key3"), true)
	}
}

func TestMetadataFromMap_Nil(t *testing.T) {
	m := MetadataFromMap(nil)

	if m.Len() != 0 {
		t.Errorf("expected empty Metadata, got %d entries", m.Len())
	}
}

func TestMetadataFromMap_DoesNotMutateOriginal(t *testing.T) {
	input := map[string]any{"key": "value"}
	m := MetadataFromMap(input)

	// Modify via Set (returns new Metadata, original unchanged)
	_ = m.Set("new_key", "new_value")

	// Original input should be unchanged
	if _, exists := input["new_key"]; exists {
		t.Error("original map was mutated")
	}
}

// ============================================================================
// MetadataFromJSON Tests
// ============================================================================

func TestMetadataFromJSON_Valid(t *testing.T) {
	jsonData := []byte(`{"key1": "value1", "key2": 42, "nested": {"a": "b"}}`)

	m, err := MetadataFromJSON(jsonData)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if m.Get("key1") != "value1" {
		t.Errorf("Get(key1) = %v, want %v", m.Get("key1"), "value1")
	}
	// JSON numbers unmarshal as float64
	if m.Get("key2") != float64(42) {
		t.Errorf("Get(key2) = %v, want %v", m.Get("key2"), float64(42))
	}

	nested, ok := m.Get("nested").(map[string]any)
	if !ok {
		t.Fatal("expected nested to be map[string]any")
	}
	if nested["a"] != "b" {
		t.Errorf("nested[a] = %v, want %v", nested["a"], "b")
	}
}

func TestMetadataFromJSON_Empty(t *testing.T) {
	jsonData := []byte(`{}`)

	m, err := MetadataFromJSON(jsonData)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if m.Len() != 0 {
		t.Errorf("expected empty Metadata, got %d entries", m.Len())
	}
}

func TestMetadataFromJSON_Nil(t *testing.T) {
	m, err := MetadataFromJSON(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if m.Len() != 0 {
		t.Errorf("expected empty Metadata, got %d entries", m.Len())
	}
}

func TestMetadataFromJSON_Null(t *testing.T) {
	m, err := MetadataFromJSON([]byte("null"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if m.Len() != 0 {
		t.Errorf("expected empty Metadata, got %d entries", m.Len())
	}
}

func TestMetadataFromJSON_Invalid(t *testing.T) {
	invalidJSON := []byte(`{invalid json}`)

	_, err := MetadataFromJSON(invalidJSON)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

// ============================================================================
// Get / Set Tests
// ============================================================================

func TestMetadata_Set_ReturnsNewMetadata(t *testing.T) {
	m1 := NewMetadata()
	m2 := m1.Set("key", "value")

	// m1 should be unchanged (immutable)
	if m1.Has("key") {
		t.Error("original Metadata was mutated")
	}

	// m2 should have the value
	if m2.Get("key") != "value" {
		t.Errorf("Get(key) = %v, want %v", m2.Get("key"), "value")
	}
}

func TestMetadata_Set_Chaining(t *testing.T) {
	m := NewMetadata().
		Set("key1", "value1").
		Set("key2", "value2").
		Set("key3", "value3")

	if m.Get("key1") != "value1" {
		t.Errorf("Get(key1) = %v, want %v", m.Get("key1"), "value1")
	}
	if m.Get("key2") != "value2" {
		t.Errorf("Get(key2) = %v, want %v", m.Get("key2"), "value2")
	}
	if m.Get("key3") != "value3" {
		t.Errorf("Get(key3) = %v, want %v", m.Get("key3"), "value3")
	}
}

func TestMetadata_Set_Overwrite(t *testing.T) {
	m := NewMetadata().
		Set("key", "value1").
		Set("key", "value2")

	if m.Get("key") != "value2" {
		t.Errorf("Get(key) = %v, want %v", m.Get("key"), "value2")
	}
}

func TestMetadata_Get_NotFound(t *testing.T) {
	m := NewMetadata()

	if m.Get("nonexistent") != nil {
		t.Errorf("expected nil for nonexistent key, got %v", m.Get("nonexistent"))
	}
}

// ============================================================================
// GetString / GetInt / GetBool Tests
// ============================================================================

func TestMetadata_GetString(t *testing.T) {
	m := NewMetadata().
		Set("string", "value").
		Set("int", 42)

	if m.GetString("string") != "value" {
		t.Errorf("GetString(string) = %v, want %v", m.GetString("string"), "value")
	}
	if m.GetString("int") != "" {
		t.Errorf("GetString(int) = %v, want empty string", m.GetString("int"))
	}
	if m.GetString("nonexistent") != "" {
		t.Errorf("GetString(nonexistent) = %v, want empty string", m.GetString("nonexistent"))
	}
}

func TestMetadata_GetInt(t *testing.T) {
	m := NewMetadata().
		Set("int", 42).
		Set("int64", int64(100)).
		Set("float64", float64(3.14)).
		Set("string", "not a number")

	if m.GetInt("int") != 42 {
		t.Errorf("GetInt(int) = %v, want %v", m.GetInt("int"), 42)
	}
	if m.GetInt("int64") != 100 {
		t.Errorf("GetInt(int64) = %v, want %v", m.GetInt("int64"), 100)
	}
	if m.GetInt("float64") != 3 {
		t.Errorf("GetInt(float64) = %v, want %v", m.GetInt("float64"), 3)
	}
	if m.GetInt("string") != 0 {
		t.Errorf("GetInt(string) = %v, want 0", m.GetInt("string"))
	}
	if m.GetInt("nonexistent") != 0 {
		t.Errorf("GetInt(nonexistent) = %v, want 0", m.GetInt("nonexistent"))
	}
}

func TestMetadata_GetBool(t *testing.T) {
	m := NewMetadata().
		Set("true", true).
		Set("false", false).
		Set("string", "true")

	if m.GetBool("true") != true {
		t.Errorf("GetBool(true) = %v, want %v", m.GetBool("true"), true)
	}
	if m.GetBool("false") != false {
		t.Errorf("GetBool(false) = %v, want %v", m.GetBool("false"), false)
	}
	if m.GetBool("string") != false {
		t.Errorf("GetBool(string) = %v, want false", m.GetBool("string"))
	}
	if m.GetBool("nonexistent") != false {
		t.Errorf("GetBool(nonexistent) = %v, want false", m.GetBool("nonexistent"))
	}
}

// ============================================================================
// Has Tests
// ============================================================================

func TestMetadata_Has(t *testing.T) {
	m := NewMetadata().Set("exists", "value")

	if !m.Has("exists") {
		t.Error("expected Has(exists) to be true")
	}
	if m.Has("nonexistent") {
		t.Error("expected Has(nonexistent) to be false")
	}
}

func TestMetadata_Has_NilValue(t *testing.T) {
	m := NewMetadata().Set("nil_value", nil)

	// Key exists even if value is nil
	if !m.Has("nil_value") {
		t.Error("expected Has(nil_value) to be true")
	}
}

// ============================================================================
// Keys / Len / IsEmpty Tests
// ============================================================================

func TestMetadata_Keys(t *testing.T) {
	m := NewMetadata().
		Set("key1", "value1").
		Set("key2", "value2")

	keys := m.Keys()

	if len(keys) != 2 {
		t.Errorf("expected 2 keys, got %d", len(keys))
	}

	keySet := make(map[string]bool)
	for _, k := range keys {
		keySet[k] = true
	}

	if !keySet["key1"] || !keySet["key2"] {
		t.Errorf("expected keys [key1, key2], got %v", keys)
	}
}

func TestMetadata_Keys_Empty(t *testing.T) {
	m := NewMetadata()

	keys := m.Keys()

	if keys != nil && len(keys) != 0 {
		t.Errorf("expected nil or empty keys, got %v", keys)
	}
}

func TestMetadata_Len(t *testing.T) {
	m := NewMetadata().
		Set("key1", "value1").
		Set("key2", "value2")

	if m.Len() != 2 {
		t.Errorf("Len() = %v, want %v", m.Len(), 2)
	}
}

func TestMetadata_IsEmpty(t *testing.T) {
	empty := NewMetadata()
	nonEmpty := NewMetadata().Set("key", "value")

	if !empty.IsEmpty() {
		t.Error("expected empty Metadata to be empty")
	}
	if nonEmpty.IsEmpty() {
		t.Error("expected non-empty Metadata to not be empty")
	}
}

// ============================================================================
// With* Fluent Methods Tests
// ============================================================================

func TestMetadata_WithCorrelationID(t *testing.T) {
	m := NewMetadata().WithCorrelationID("corr-123")

	if m.CorrelationID() != "corr-123" {
		t.Errorf("CorrelationID() = %v, want %v", m.CorrelationID(), "corr-123")
	}
}

func TestMetadata_WithCausationID(t *testing.T) {
	m := NewMetadata().WithCausationID("cause-456")

	if m.CausationID() != "cause-456" {
		t.Errorf("CausationID() = %v, want %v", m.CausationID(), "cause-456")
	}
}

func TestMetadata_WithUserAgent(t *testing.T) {
	m := NewMetadata().WithUserAgent("Mozilla/5.0")

	if m.GetString(MetaKeyUserAgent) != "Mozilla/5.0" {
		t.Errorf("GetString(user_agent) = %v, want %v", m.GetString(MetaKeyUserAgent), "Mozilla/5.0")
	}
}

func TestMetadata_WithIPAddress(t *testing.T) {
	m := NewMetadata().WithIPAddress("192.168.1.1")

	if m.GetString(MetaKeyIPAddress) != "192.168.1.1" {
		t.Errorf("GetString(ip_address) = %v, want %v", m.GetString(MetaKeyIPAddress), "192.168.1.1")
	}
}

func TestMetadata_WithRequestID(t *testing.T) {
	m := NewMetadata().WithRequestID("req-789")

	if m.GetString(MetaKeyRequestID) != "req-789" {
		t.Errorf("GetString(request_id) = %v, want %v", m.GetString(MetaKeyRequestID), "req-789")
	}
}

func TestMetadata_WithReason(t *testing.T) {
	m := NewMetadata().WithReason("user requested deletion")

	if m.GetString(MetaKeyReason) != "user requested deletion" {
		t.Errorf("GetString(reason) = %v, want %v", m.GetString(MetaKeyReason), "user requested deletion")
	}
}

func TestMetadata_With_Chaining(t *testing.T) {
	m := NewMetadata().
		WithCorrelationID("corr-123").
		WithCausationID("cause-456").
		WithIPAddress("192.168.1.1").
		WithUserAgent("Mozilla/5.0")

	if m.CorrelationID() != "corr-123" {
		t.Errorf("CorrelationID() = %v, want %v", m.CorrelationID(), "corr-123")
	}
	if m.CausationID() != "cause-456" {
		t.Errorf("CausationID() = %v, want %v", m.CausationID(), "cause-456")
	}
	if m.GetString(MetaKeyIPAddress) != "192.168.1.1" {
		t.Errorf("GetString(ip_address) mismatch")
	}
	if m.GetString(MetaKeyUserAgent) != "Mozilla/5.0" {
		t.Errorf("GetString(user_agent) mismatch")
	}
}

// ============================================================================
// CorrelationID / CausationID Tests
// ============================================================================

func TestMetadata_CorrelationID_NotSet(t *testing.T) {
	m := NewMetadata()

	if m.CorrelationID() != "" {
		t.Errorf("expected empty CorrelationID, got %v", m.CorrelationID())
	}
}

func TestMetadata_CausationID_NotSet(t *testing.T) {
	m := NewMetadata()

	if m.CausationID() != "" {
		t.Errorf("expected empty CausationID, got %v", m.CausationID())
	}
}

func TestMetadata_CorrelationID_NonString(t *testing.T) {
	m := NewMetadata().Set(MetaKeyCorrelationID, 12345) // wrong type

	if m.CorrelationID() != "" {
		t.Errorf("expected empty CorrelationID for non-string value, got %v", m.CorrelationID())
	}
}

// ============================================================================
// Merge Tests
// ============================================================================

func TestMetadata_Merge(t *testing.T) {
	m1 := NewMetadata().Set("key1", "value1").Set("shared", "m1")
	m2 := NewMetadata().Set("key2", "value2").Set("shared", "m2")

	merged := m1.Merge(m2)

	if merged.Get("key1") != "value1" {
		t.Errorf("Get(key1) = %v, want %v", merged.Get("key1"), "value1")
	}
	if merged.Get("key2") != "value2" {
		t.Errorf("Get(key2) = %v, want %v", merged.Get("key2"), "value2")
	}
	// m2 overwrites m1 for shared keys
	if merged.Get("shared") != "m2" {
		t.Errorf("Get(shared) = %v, want %v", merged.Get("shared"), "m2")
	}
}

func TestMetadata_Merge_Empty(t *testing.T) {
	m := NewMetadata().Set("key", "value")
	empty := NewMetadata()

	merged := m.Merge(empty)

	if merged.Get("key") != "value" {
		t.Errorf("Get(key) = %v, want %v", merged.Get("key"), "value")
	}
}

func TestMetadata_Merge_DoesNotMutateOriginal(t *testing.T) {
	m1 := NewMetadata().Set("key1", "value1")
	m2 := NewMetadata().Set("key2", "value2")

	merged := m1.Merge(m2)

	// Verify originals are unchanged
	if m1.Has("key2") {
		t.Error("m1 was mutated by merge")
	}
	if m2.Has("key1") {
		t.Error("m2 was mutated by merge")
	}

	// Verify merged has both
	if !merged.Has("key1") || !merged.Has("key2") {
		t.Error("merged should have both keys")
	}
}

// ============================================================================
// ToMap Tests
// ============================================================================

func TestMetadata_ToMap(t *testing.T) {
	m := NewMetadata().Set("key", "value")

	result := m.ToMap()

	if result["key"] != "value" {
		t.Errorf("result[key] = %v, want %v", result["key"], "value")
	}
}

func TestMetadata_ToMap_Empty(t *testing.T) {
	m := NewMetadata()

	result := m.ToMap()

	if result == nil {
		t.Error("expected non-nil map")
	}
	if len(result) != 0 {
		t.Errorf("expected empty map, got %v", result)
	}
}

func TestMetadata_ToMap_DoesNotMutateOriginal(t *testing.T) {
	m := NewMetadata().Set("key", "value")

	result := m.ToMap()
	result["new_key"] = "new_value"

	if m.Has("new_key") {
		t.Error("original was mutated")
	}
}

// ============================================================================
// ToJSON Tests
// ============================================================================

func TestMetadata_ToJSON(t *testing.T) {
	m := NewMetadata().
		Set("string", "value").
		Set("number", 42).
		Set("bool", true)

	jsonData, err := m.ToJSON()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Parse back to verify
	var parsed map[string]any
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if parsed["string"] != "value" {
		t.Errorf("parsed[string] = %v, want %v", parsed["string"], "value")
	}
	if parsed["number"] != float64(42) {
		t.Errorf("parsed[number] = %v, want %v", parsed["number"], float64(42))
	}
	if parsed["bool"] != true {
		t.Errorf("parsed[bool] = %v, want %v", parsed["bool"], true)
	}
}

func TestMetadata_ToJSON_Empty(t *testing.T) {
	m := NewMetadata()

	jsonData, err := m.ToJSON()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(jsonData) != "{}" {
		t.Errorf("expected empty JSON object, got %s", string(jsonData))
	}
}

// ============================================================================
// MarshalJSON / UnmarshalJSON Tests
// ============================================================================

func TestMetadata_MarshalJSON(t *testing.T) {
	m := NewMetadata().Set("key", "value")

	jsonData, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := `{"key":"value"}`
	if string(jsonData) != expected {
		t.Errorf("MarshalJSON() = %s, want %s", string(jsonData), expected)
	}
}

func TestMetadata_UnmarshalJSON(t *testing.T) {
	jsonData := []byte(`{"key":"value","count":42}`)

	var m Metadata
	if err := json.Unmarshal(jsonData, &m); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if m.GetString("key") != "value" {
		t.Errorf("GetString(key) = %v, want %v", m.GetString("key"), "value")
	}
	if m.Get("count") != float64(42) {
		t.Errorf("Get(count) = %v, want %v", m.Get("count"), float64(42))
	}
}

// ============================================================================
// Roundtrip Tests
// ============================================================================

func TestMetadata_JSONRoundtrip(t *testing.T) {
	original := NewMetadata().
		Set("string", "value").
		Set("number", float64(42)).
		Set("bool", true).
		WithCorrelationID("corr-123").
		WithCausationID("cause-456")

	jsonData, err := original.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON error: %v", err)
	}

	parsed, err := MetadataFromJSON(jsonData)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}

	if parsed.Get("string") != original.Get("string") {
		t.Error("string mismatch after roundtrip")
	}
	if parsed.Get("number") != original.Get("number") {
		t.Error("number mismatch after roundtrip")
	}
	if parsed.Get("bool") != original.Get("bool") {
		t.Error("bool mismatch after roundtrip")
	}
	if parsed.CorrelationID() != original.CorrelationID() {
		t.Error("correlation_id mismatch after roundtrip")
	}
	if parsed.CausationID() != original.CausationID() {
		t.Error("causation_id mismatch after roundtrip")
	}
}
