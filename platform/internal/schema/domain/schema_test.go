package domain

import (
	"testing"

	"github.com/0xsj/nexus/platform/pkg/types"
)

// ============================================================================
// Test Helpers
// ============================================================================

func createTestClaims(t *testing.T) []ClaimDefinition {
	t.Helper()
	return []ClaimDefinition{
		MustNewClaimDefinition("name", MustNewClaimType(DataTypeString), Required()),
		MustNewClaimDefinition("email", MustNewClaimType(DataTypeEmail), Required()),
		MustNewClaimDefinition("age", MustNewClaimType(DataTypeInteger), Optional()),
	}
}

func createTestSchema(t *testing.T) *Schema {
	t.Helper()
	id := types.NewID()
	claims := createTestClaims(t)

	schema, err := RegisterSchema(
		id,
		"test.credential",
		"Test Credential",
		"A test credential schema",
		InitialVersion(),
		claims,
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create test schema: %v", err)
	}
	return schema
}

// ============================================================================
// SchemaStatus Tests
// ============================================================================

func TestSchemaStatus_String(t *testing.T) {
	tests := []struct {
		status SchemaStatus
		want   string
	}{
		{SchemaStatusActive, "active"},
		{SchemaStatusDeprecated, "deprecated"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.status.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSchemaStatus_IsValid(t *testing.T) {
	tests := []struct {
		status SchemaStatus
		want   bool
	}{
		{SchemaStatusActive, true},
		{SchemaStatusDeprecated, true},
		{SchemaStatus("invalid"), false},
		{SchemaStatus(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if got := tt.status.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ============================================================================
// RegisterSchema Tests
// ============================================================================

func TestRegisterSchema_Valid(t *testing.T) {
	id := types.NewID()
	claims := createTestClaims(t)

	schema, err := RegisterSchema(
		id,
		"github.contribution",
		"GitHub Contribution",
		"Verifies GitHub contributions",
		InitialVersion(),
		claims,
		nil,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if schema.ID() != id {
		t.Errorf("ID() = %v, want %v", schema.ID(), id)
	}
	if schema.SchemaType() != "github.contribution" {
		t.Errorf("SchemaType() = %v, want github.contribution", schema.SchemaType())
	}
	if schema.Name() != "GitHub Contribution" {
		t.Errorf("Name() = %v, want 'GitHub Contribution'", schema.Name())
	}
	if schema.Description() != "Verifies GitHub contributions" {
		t.Errorf("Description() = %v, want 'Verifies GitHub contributions'", schema.Description())
	}
	if !schema.CurrentVersion().Equals(InitialVersion()) {
		t.Errorf("CurrentVersion() = %v, want %v", schema.CurrentVersion(), InitialVersion())
	}
	if schema.Status() != SchemaStatusActive {
		t.Errorf("Status() = %v, want %v", schema.Status(), SchemaStatusActive)
	}
	if !schema.IsActive() {
		t.Error("IsActive() = false, want true")
	}
	if schema.IsDeprecated() {
		t.Error("IsDeprecated() = true, want false")
	}
	if !schema.IsBuiltIn() {
		t.Error("IsBuiltIn() = false, want true")
	}
	if schema.IsCustom() {
		t.Error("IsCustom() = true, want false")
	}
	if schema.ClaimCount() != 3 {
		t.Errorf("ClaimCount() = %d, want 3", schema.ClaimCount())
	}
}

func TestRegisterSchema_WithIssuer(t *testing.T) {
	id := types.NewID()
	issuerID := types.NewID()
	claims := createTestClaims(t)

	schema, err := RegisterSchema(
		id,
		"custom.credential",
		"Custom Credential",
		"",
		InitialVersion(),
		claims,
		&issuerID,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if schema.IssuerID() == nil {
		t.Fatal("IssuerID() = nil, want non-nil")
	}
	if *schema.IssuerID() != issuerID {
		t.Errorf("IssuerID() = %v, want %v", *schema.IssuerID(), issuerID)
	}
	if schema.IsBuiltIn() {
		t.Error("IsBuiltIn() = true, want false")
	}
	if !schema.IsCustom() {
		t.Error("IsCustom() = false, want true")
	}
}

func TestRegisterSchema_ZeroID(t *testing.T) {
	var zeroID types.ID
	claims := createTestClaims(t)

	_, err := RegisterSchema(
		zeroID,
		"test.credential",
		"Test",
		"",
		InitialVersion(),
		claims,
		nil,
	)

	if err == nil {
		t.Error("expected error for zero ID")
	}
}

func TestRegisterSchema_EmptySchemaType(t *testing.T) {
	id := types.NewID()
	claims := createTestClaims(t)

	_, err := RegisterSchema(
		id,
		"",
		"Test",
		"",
		InitialVersion(),
		claims,
		nil,
	)

	if err == nil {
		t.Error("expected error for empty schema type")
	}
}

func TestRegisterSchema_EmptyName(t *testing.T) {
	id := types.NewID()
	claims := createTestClaims(t)

	_, err := RegisterSchema(
		id,
		"test.credential",
		"",
		"",
		InitialVersion(),
		claims,
		nil,
	)

	if err == nil {
		t.Error("expected error for empty name")
	}
}

func TestRegisterSchema_ZeroVersion(t *testing.T) {
	id := types.NewID()
	claims := createTestClaims(t)
	var zeroVersion SchemaVersion

	_, err := RegisterSchema(
		id,
		"test.credential",
		"Test",
		"",
		zeroVersion,
		claims,
		nil,
	)

	if err == nil {
		t.Error("expected error for zero version")
	}
}

func TestRegisterSchema_NoClaims(t *testing.T) {
	id := types.NewID()

	_, err := RegisterSchema(
		id,
		"test.credential",
		"Test",
		"",
		InitialVersion(),
		[]ClaimDefinition{},
		nil,
	)

	if err == nil {
		t.Error("expected error for empty claims")
	}
}

func TestRegisterSchema_NilClaims(t *testing.T) {
	id := types.NewID()

	_, err := RegisterSchema(
		id,
		"test.credential",
		"Test",
		"",
		InitialVersion(),
		nil,
		nil,
	)

	if err == nil {
		t.Error("expected error for nil claims")
	}
}

func TestRegisterSchema_DuplicateClaimKeys(t *testing.T) {
	id := types.NewID()
	claims := []ClaimDefinition{
		MustNewClaimDefinition("name", MustNewClaimType(DataTypeString)),
		MustNewClaimDefinition("name", MustNewClaimType(DataTypeString)), // Duplicate
	}

	_, err := RegisterSchema(
		id,
		"test.credential",
		"Test",
		"",
		InitialVersion(),
		claims,
		nil,
	)

	if err == nil {
		t.Error("expected error for duplicate claim keys")
	}
}

// ============================================================================
// AddVersion Tests
// ============================================================================

func TestSchema_AddVersion_Valid(t *testing.T) {
	schema := createTestSchema(t)
	newVersion := schema.CurrentVersion().NextMinor()
	newClaims := []ClaimDefinition{
		MustNewClaimDefinition("name", MustNewClaimType(DataTypeString), Required()),
		MustNewClaimDefinition("email", MustNewClaimType(DataTypeEmail), Required()),
		MustNewClaimDefinition("website", MustNewClaimType(DataTypeURL), Optional()),
	}

	err := schema.AddVersion(newVersion, newClaims, "Added website claim")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !schema.CurrentVersion().Equals(newVersion) {
		t.Errorf("CurrentVersion() = %v, want %v", schema.CurrentVersion(), newVersion)
	}
	if schema.ClaimCount() != 3 {
		t.Errorf("ClaimCount() = %d, want 3", schema.ClaimCount())
	}
	if !schema.HasClaim("website") {
		t.Error("HasClaim(website) = false, want true")
	}
}

func TestSchema_AddVersion_OlderVersion(t *testing.T) {
	schema := createTestSchema(t)
	olderVersion := MustParseSchemaVersion("0.1.0")
	claims := createTestClaims(t)

	err := schema.AddVersion(olderVersion, claims, "")

	if err == nil {
		t.Error("expected error for older version")
	}
}

func TestSchema_AddVersion_SameVersion(t *testing.T) {
	schema := createTestSchema(t)
	sameVersion := schema.CurrentVersion()
	claims := createTestClaims(t)

	err := schema.AddVersion(sameVersion, claims, "")

	if err == nil {
		t.Error("expected error for same version")
	}
}

func TestSchema_AddVersion_Deprecated(t *testing.T) {
	schema := createTestSchema(t)
	schema.Deprecate("Testing", nil)

	newVersion := schema.CurrentVersion().NextMinor()
	claims := createTestClaims(t)

	err := schema.AddVersion(newVersion, claims, "")

	if err == nil {
		t.Error("expected error for deprecated schema")
	}
}

func TestSchema_AddVersion_NoClaims(t *testing.T) {
	schema := createTestSchema(t)
	newVersion := schema.CurrentVersion().NextMinor()

	err := schema.AddVersion(newVersion, []ClaimDefinition{}, "")

	if err == nil {
		t.Error("expected error for empty claims")
	}
}

func TestSchema_AddVersion_DuplicateClaimKeys(t *testing.T) {
	schema := createTestSchema(t)
	newVersion := schema.CurrentVersion().NextMinor()
	claims := []ClaimDefinition{
		MustNewClaimDefinition("name", MustNewClaimType(DataTypeString)),
		MustNewClaimDefinition("name", MustNewClaimType(DataTypeString)), // Duplicate
	}

	err := schema.AddVersion(newVersion, claims, "")

	if err == nil {
		t.Error("expected error for duplicate claim keys")
	}
}

// ============================================================================
// Deprecate Tests
// ============================================================================

func TestSchema_Deprecate_Valid(t *testing.T) {
	schema := createTestSchema(t)

	err := schema.Deprecate("No longer maintained", nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !schema.IsDeprecated() {
		t.Error("IsDeprecated() = false, want true")
	}
	if schema.IsActive() {
		t.Error("IsActive() = true, want false")
	}
	if schema.Status() != SchemaStatusDeprecated {
		t.Errorf("Status() = %v, want %v", schema.Status(), SchemaStatusDeprecated)
	}
}

func TestSchema_Deprecate_WithReplacement(t *testing.T) {
	schema := createTestSchema(t)
	replacementID := types.NewID()

	err := schema.Deprecate("Replaced by new version", &replacementID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !schema.IsDeprecated() {
		t.Error("IsDeprecated() = false, want true")
	}
}

func TestSchema_Deprecate_AlreadyDeprecated(t *testing.T) {
	schema := createTestSchema(t)
	schema.Deprecate("First deprecation", nil)

	err := schema.Deprecate("Second deprecation", nil)

	if err == nil {
		t.Error("expected error for already deprecated schema")
	}
}

// ============================================================================
// Activate Tests
// ============================================================================

func TestSchema_Activate_Valid(t *testing.T) {
	schema := createTestSchema(t)
	schema.Deprecate("Testing", nil)

	err := schema.Activate()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !schema.IsActive() {
		t.Error("IsActive() = false, want true")
	}
	if schema.IsDeprecated() {
		t.Error("IsDeprecated() = true, want false")
	}
}

func TestSchema_Activate_AlreadyActive(t *testing.T) {
	schema := createTestSchema(t)

	err := schema.Activate()

	if err == nil {
		t.Error("expected error for already active schema")
	}
}

// ============================================================================
// UpdateMetadata Tests
// ============================================================================

func TestSchema_UpdateMetadata_Name(t *testing.T) {
	schema := createTestSchema(t)

	err := schema.UpdateMetadata("New Name", "")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if schema.Name() != "New Name" {
		t.Errorf("Name() = %v, want 'New Name'", schema.Name())
	}
}

func TestSchema_UpdateMetadata_Description(t *testing.T) {
	schema := createTestSchema(t)

	err := schema.UpdateMetadata("", "New description")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if schema.Description() != "New description" {
		t.Errorf("Description() = %v, want 'New description'", schema.Description())
	}
}

func TestSchema_UpdateMetadata_Both(t *testing.T) {
	schema := createTestSchema(t)

	err := schema.UpdateMetadata("New Name", "New description")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if schema.Name() != "New Name" {
		t.Errorf("Name() = %v, want 'New Name'", schema.Name())
	}
	if schema.Description() != "New description" {
		t.Errorf("Description() = %v, want 'New description'", schema.Description())
	}
}

func TestSchema_UpdateMetadata_Empty(t *testing.T) {
	schema := createTestSchema(t)

	err := schema.UpdateMetadata("", "")

	if err == nil {
		t.Error("expected error when both fields are empty")
	}
}

// ============================================================================
// Claims Query Methods Tests
// ============================================================================

func TestSchema_Claims_Immutability(t *testing.T) {
	schema := createTestSchema(t)
	claims := schema.Claims()

	originalCount := schema.ClaimCount()

	// Modify returned slice
	claims = append(claims, MustNewClaimDefinition("new_claim", MustNewClaimType(DataTypeString)))

	// Original should be unchanged
	if schema.ClaimCount() != originalCount {
		t.Error("Claims() returned slice should not affect original")
	}
}

func TestSchema_GetClaim_Found(t *testing.T) {
	schema := createTestSchema(t)

	claim, found := schema.GetClaim("name")

	if !found {
		t.Error("GetClaim(name) found = false, want true")
	}
	if claim.Key() != "name" {
		t.Errorf("GetClaim(name).Key() = %v, want name", claim.Key())
	}
}

func TestSchema_GetClaim_NotFound(t *testing.T) {
	schema := createTestSchema(t)

	claim, found := schema.GetClaim("nonexistent")

	if found {
		t.Error("GetClaim(nonexistent) found = true, want false")
	}
	if !claim.IsZero() {
		t.Error("GetClaim(nonexistent) should return zero value")
	}
}

func TestSchema_HasClaim(t *testing.T) {
	schema := createTestSchema(t)

	if !schema.HasClaim("name") {
		t.Error("HasClaim(name) = false, want true")
	}
	if !schema.HasClaim("email") {
		t.Error("HasClaim(email) = false, want true")
	}
	if schema.HasClaim("nonexistent") {
		t.Error("HasClaim(nonexistent) = true, want false")
	}
}

func TestSchema_RequiredClaims(t *testing.T) {
	schema := createTestSchema(t)

	required := schema.RequiredClaims()

	if len(required) != 2 {
		t.Fatalf("RequiredClaims() length = %d, want 2", len(required))
	}

	keys := make(map[string]bool)
	for _, c := range required {
		keys[c.Key()] = true
		if !c.IsRequired() {
			t.Errorf("RequiredClaims() contains optional claim: %s", c.Key())
		}
	}

	if !keys["name"] || !keys["email"] {
		t.Error("RequiredClaims() missing expected claims")
	}
}

func TestSchema_OptionalClaims(t *testing.T) {
	schema := createTestSchema(t)

	optional := schema.OptionalClaims()

	if len(optional) != 1 {
		t.Fatalf("OptionalClaims() length = %d, want 1", len(optional))
	}

	if optional[0].Key() != "age" {
		t.Errorf("OptionalClaims()[0].Key() = %v, want age", optional[0].Key())
	}
	if !optional[0].IsOptional() {
		t.Error("OptionalClaims() contains required claim")
	}
}

// ============================================================================
// NewSchema Tests
// ============================================================================

func TestNewSchema(t *testing.T) {
	id := types.NewID()
	schema := NewSchema(id)

	if schema.ID() != id {
		t.Errorf("ID() = %v, want %v", schema.ID(), id)
	}
	// Newly created schema should have zero values until events are applied
	if schema.SchemaType() != "" {
		t.Errorf("SchemaType() = %v, want empty", schema.SchemaType())
	}
}

// ============================================================================
// SchemaFactory Tests
// ============================================================================

func TestSchemaFactory_Create(t *testing.T) {
	factory := NewSchemaFactory()
	id := types.NewID()

	aggregate := factory.Create(id.String())

	schema, ok := aggregate.(*Schema)
	if !ok {
		t.Fatal("Create() did not return *Schema")
	}
	if schema.ID().String() != id.String() {
		t.Errorf("ID() = %v, want %v", schema.ID(), id)
	}
}

// ============================================================================
// Event Application Tests
// ============================================================================

func TestSchema_ApplyEvent_SchemaRegistered(t *testing.T) {
	id := types.NewID()
	claims := createTestClaims(t)

	schema, err := RegisterSchema(
		id,
		"test.credential",
		"Test Credential",
		"Description",
		InitialVersion(),
		claims,
		nil,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify state was updated via event application
	if schema.ID() != id {
		t.Errorf("ID() = %v, want %v", schema.ID(), id)
	}
	if schema.SchemaType() != "test.credential" {
		t.Errorf("SchemaType() = %v, want test.credential", schema.SchemaType())
	}
	if schema.Status() != SchemaStatusActive {
		t.Errorf("Status() = %v, want active", schema.Status())
	}
	if schema.CreatedAt().IsZero() {
		t.Error("CreatedAt() should not be zero")
	}
	if schema.UpdatedAt().IsZero() {
		t.Error("UpdatedAt() should not be zero")
	}
}

func TestSchema_ApplyEvent_SchemaVersionAdded(t *testing.T) {
	schema := createTestSchema(t)
	originalVersion := schema.CurrentVersion()

	newVersion := originalVersion.NextMinor()
	newClaims := []ClaimDefinition{
		MustNewClaimDefinition("name", MustNewClaimType(DataTypeString)),
	}

	err := schema.AddVersion(newVersion, newClaims, "Updated")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !schema.CurrentVersion().Equals(newVersion) {
		t.Errorf("CurrentVersion() = %v, want %v", schema.CurrentVersion(), newVersion)
	}
	if schema.ClaimCount() != 1 {
		t.Errorf("ClaimCount() = %d, want 1", schema.ClaimCount())
	}
}

func TestSchema_ApplyEvent_SchemaDeprecated(t *testing.T) {
	schema := createTestSchema(t)

	err := schema.Deprecate("Testing", nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if schema.Status() != SchemaStatusDeprecated {
		t.Errorf("Status() = %v, want deprecated", schema.Status())
	}
}

func TestSchema_ApplyEvent_SchemaActivated(t *testing.T) {
	schema := createTestSchema(t)
	schema.Deprecate("Testing", nil)

	err := schema.Activate()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if schema.Status() != SchemaStatusActive {
		t.Errorf("Status() = %v, want active", schema.Status())
	}
}

func TestSchema_ApplyEvent_SchemaMetadataUpdated(t *testing.T) {
	schema := createTestSchema(t)
	originalUpdatedAt := schema.UpdatedAt()

	err := schema.UpdateMetadata("Updated Name", "Updated Description")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if schema.Name() != "Updated Name" {
		t.Errorf("Name() = %v, want 'Updated Name'", schema.Name())
	}
	if schema.Description() != "Updated Description" {
		t.Errorf("Description() = %v, want 'Updated Description'", schema.Description())
	}
	if !schema.UpdatedAt().After(originalUpdatedAt) && !schema.UpdatedAt().Equal(originalUpdatedAt) {
		t.Error("UpdatedAt() should be updated")
	}
}

// ============================================================================
// Aggregate Root Tests
// ============================================================================

func TestSchema_GetAggregateRoot(t *testing.T) {
	schema := createTestSchema(t)

	root := schema.GetAggregateRoot()

	if root == nil {
		t.Fatal("GetAggregateRoot() = nil")
	}

	// Should have uncommitted events from registration
	changes := root.Changes()
	if len(changes) == 0 {
		t.Error("Should have uncommitted changes")
	}
}

func TestSchema_AggregateInterface(t *testing.T) {
	schema := createTestSchema(t)

	// Test AggregateID
	if schema.AggregateID() != schema.ID().String() {
		t.Errorf("AggregateID() = %v, want %v", schema.AggregateID(), schema.ID().String())
	}

	// Test AggregateType
	if schema.AggregateType() != AggregateTypeSchema {
		t.Errorf("AggregateType() = %v, want %v", schema.AggregateType(), AggregateTypeSchema)
	}

	// Test Version
	if schema.Version() < 1 {
		t.Errorf("Version() = %d, should be >= 1 after registration", schema.Version())
	}

	// Test Changes
	changes := schema.Changes()
	if len(changes) == 0 {
		t.Error("Changes() should not be empty after registration")
	}

	// Test HasChanges
	if !schema.HasChanges() {
		t.Error("HasChanges() = false, want true")
	}

	// Test ClearChanges
	schema.ClearChanges()
	if schema.HasChanges() {
		t.Error("HasChanges() = true after ClearChanges(), want false")
	}
	if len(schema.Changes()) != 0 {
		t.Error("Changes() should be empty after ClearChanges()")
	}
}

// ============================================================================
// Edge Cases
// ============================================================================

func TestSchema_EmptyDescription(t *testing.T) {
	id := types.NewID()
	claims := createTestClaims(t)

	schema, err := RegisterSchema(
		id,
		"test.credential",
		"Test",
		"", // Empty description is valid
		InitialVersion(),
		claims,
		nil,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if schema.Description() != "" {
		t.Errorf("Description() = %v, want empty", schema.Description())
	}
}

func TestSchema_SingleClaim(t *testing.T) {
	id := types.NewID()
	claims := []ClaimDefinition{
		MustNewClaimDefinition("value", MustNewClaimType(DataTypeString)),
	}

	schema, err := RegisterSchema(
		id,
		"simple.credential",
		"Simple",
		"",
		InitialVersion(),
		claims,
		nil,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if schema.ClaimCount() != 1 {
		t.Errorf("ClaimCount() = %d, want 1", schema.ClaimCount())
	}
}
