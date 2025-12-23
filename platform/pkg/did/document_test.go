package did

import (
	"encoding/json"
	"testing"
)

func TestNewDocument(t *testing.T) {
	d, _ := Parse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")

	doc := NewDocument(d)

	if doc.ID.String() != d.String() {
		t.Errorf("ID = %v, want %v", doc.ID.String(), d.String())
	}

	if len(doc.Context) != 2 {
		t.Errorf("Context count = %v, want 2", len(doc.Context))
	}

	if doc.Context[0] != "https://www.w3.org/ns/did/v1" {
		t.Errorf("Context[0] = %v, want https://www.w3.org/ns/did/v1", doc.Context[0])
	}
}

func TestDocument_AddVerificationMethod(t *testing.T) {
	d, _ := Parse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")
	doc := NewDocument(d)

	vm := VerificationMethod{
		ID:                 d.Fragment("key-1"),
		Type:               Ed25519VerificationKey2020,
		Controller:         d,
		PublicKeyMultibase: "z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
	}

	doc.AddVerificationMethod(vm)

	if len(doc.VerificationMethod) != 1 {
		t.Fatalf("VerificationMethod count = %v, want 1", len(doc.VerificationMethod))
	}

	if doc.VerificationMethod[0].ID != vm.ID {
		t.Errorf("VerificationMethod[0].ID = %v, want %v", doc.VerificationMethod[0].ID, vm.ID)
	}

	if doc.VerificationMethod[0].Type != Ed25519VerificationKey2020 {
		t.Errorf("VerificationMethod[0].Type = %v, want %v", doc.VerificationMethod[0].Type, Ed25519VerificationKey2020)
	}
}

func TestDocument_AddAuthentication(t *testing.T) {
	d, _ := Parse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")
	doc := NewDocument(d)

	ref := d.Fragment("key-1")
	doc.AddAuthentication(ref)

	if len(doc.Authentication) != 1 {
		t.Fatalf("Authentication count = %v, want 1", len(doc.Authentication))
	}

	if doc.Authentication[0].Reference != ref {
		t.Errorf("Authentication[0].Reference = %v, want %v", doc.Authentication[0].Reference, ref)
	}
}

func TestDocument_AddAssertionMethod(t *testing.T) {
	d, _ := Parse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")
	doc := NewDocument(d)

	ref := d.Fragment("key-1")
	doc.AddAssertionMethod(ref)

	if len(doc.AssertionMethod) != 1 {
		t.Fatalf("AssertionMethod count = %v, want 1", len(doc.AssertionMethod))
	}

	if doc.AssertionMethod[0].Reference != ref {
		t.Errorf("AssertionMethod[0].Reference = %v, want %v", doc.AssertionMethod[0].Reference, ref)
	}
}

func TestDocument_AddKeyAgreement(t *testing.T) {
	d, _ := Parse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")
	doc := NewDocument(d)

	ref := d.Fragment("key-1")
	doc.AddKeyAgreement(ref)

	if len(doc.KeyAgreement) != 1 {
		t.Fatalf("KeyAgreement count = %v, want 1", len(doc.KeyAgreement))
	}

	if doc.KeyAgreement[0].Reference != ref {
		t.Errorf("KeyAgreement[0].Reference = %v, want %v", doc.KeyAgreement[0].Reference, ref)
	}
}

func TestDocument_AddCapabilityInvocation(t *testing.T) {
	d, _ := Parse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")
	doc := NewDocument(d)

	ref := d.Fragment("key-1")
	doc.AddCapabilityInvocation(ref)

	if len(doc.CapabilityInvocation) != 1 {
		t.Fatalf("CapabilityInvocation count = %v, want 1", len(doc.CapabilityInvocation))
	}

	if doc.CapabilityInvocation[0].Reference != ref {
		t.Errorf("CapabilityInvocation[0].Reference = %v, want %v", doc.CapabilityInvocation[0].Reference, ref)
	}
}

func TestDocument_AddCapabilityDelegation(t *testing.T) {
	d, _ := Parse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")
	doc := NewDocument(d)

	ref := d.Fragment("key-1")
	doc.AddCapabilityDelegation(ref)

	if len(doc.CapabilityDelegation) != 1 {
		t.Fatalf("CapabilityDelegation count = %v, want 1", len(doc.CapabilityDelegation))
	}

	if doc.CapabilityDelegation[0].Reference != ref {
		t.Errorf("CapabilityDelegation[0].Reference = %v, want %v", doc.CapabilityDelegation[0].Reference, ref)
	}
}

func TestDocument_AddService(t *testing.T) {
	d, _ := Parse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")
	doc := NewDocument(d)

	svc := Service{
		ID:              d.Fragment("service-1"),
		Type:            "LinkedDomains",
		ServiceEndpoint: ServiceEndpoint{URI: "https://example.com"},
	}

	doc.AddService(svc)

	if len(doc.Service) != 1 {
		t.Fatalf("Service count = %v, want 1", len(doc.Service))
	}

	if doc.Service[0].ID != svc.ID {
		t.Errorf("Service[0].ID = %v, want %v", doc.Service[0].ID, svc.ID)
	}

	if doc.Service[0].Type != "LinkedDomains" {
		t.Errorf("Service[0].Type = %v, want LinkedDomains", doc.Service[0].Type)
	}
}

func TestDocument_SetController(t *testing.T) {
	d, _ := Parse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")
	controller, _ := Parse("did:key:z6MkoDgemAx8aw9xGMVCdoaHiGc1FbimMJLPgRsRtiBQRBFy")

	doc := NewDocument(d)
	doc.SetController(controller)

	if len(doc.Controller) != 1 {
		t.Fatalf("Controller count = %v, want 1", len(doc.Controller))
	}

	if !doc.Controller[0].Equals(controller) {
		t.Errorf("Controller[0] = %v, want %v", doc.Controller[0], controller)
	}
}

func TestDocument_GetVerificationMethod(t *testing.T) {
	d, _ := Parse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")
	doc := NewDocument(d)

	vmID := d.Fragment("key-1")
	vm := VerificationMethod{
		ID:                 vmID,
		Type:               Ed25519VerificationKey2020,
		Controller:         d,
		PublicKeyMultibase: "z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
	}
	doc.AddVerificationMethod(vm)

	// Find existing
	found := doc.GetVerificationMethod(vmID)
	if found == nil {
		t.Fatal("GetVerificationMethod() returned nil for existing ID")
	}
	if found.ID != vmID {
		t.Errorf("GetVerificationMethod().ID = %v, want %v", found.ID, vmID)
	}

	// Find non-existing
	notFound := doc.GetVerificationMethod("did:key:z6Mk...#nonexistent")
	if notFound != nil {
		t.Error("GetVerificationMethod() should return nil for non-existing ID")
	}
}

func TestDocument_GetService(t *testing.T) {
	d, _ := Parse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")
	doc := NewDocument(d)

	svcID := d.Fragment("service-1")
	svc := Service{
		ID:              svcID,
		Type:            "LinkedDomains",
		ServiceEndpoint: ServiceEndpoint{URI: "https://example.com"},
	}
	doc.AddService(svc)

	// Find existing
	found := doc.GetService(svcID)
	if found == nil {
		t.Fatal("GetService() returned nil for existing ID")
	}
	if found.ID != svcID {
		t.Errorf("GetService().ID = %v, want %v", found.ID, svcID)
	}

	// Find non-existing
	notFound := doc.GetService("did:key:z6Mk...#nonexistent")
	if notFound != nil {
		t.Error("GetService() should return nil for non-existing ID")
	}
}

func TestDocument_GetServiceByType(t *testing.T) {
	d, _ := Parse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")
	doc := NewDocument(d)

	doc.AddService(Service{
		ID:              d.Fragment("service-1"),
		Type:            "LinkedDomains",
		ServiceEndpoint: ServiceEndpoint{URI: "https://example.com"},
	})
	doc.AddService(Service{
		ID:              d.Fragment("service-2"),
		Type:            "CredentialRegistry",
		ServiceEndpoint: ServiceEndpoint{URI: "https://credentials.example.com"},
	})
	doc.AddService(Service{
		ID:              d.Fragment("service-3"),
		Type:            "LinkedDomains",
		ServiceEndpoint: ServiceEndpoint{URI: "https://example.org"},
	})

	// Find by type
	linkedDomains := doc.GetServiceByType("LinkedDomains")
	if len(linkedDomains) != 2 {
		t.Errorf("GetServiceByType(LinkedDomains) count = %v, want 2", len(linkedDomains))
	}

	credRegistry := doc.GetServiceByType("CredentialRegistry")
	if len(credRegistry) != 1 {
		t.Errorf("GetServiceByType(CredentialRegistry) count = %v, want 1", len(credRegistry))
	}

	// Find non-existing type
	notFound := doc.GetServiceByType("NonExistent")
	if len(notFound) != 0 {
		t.Errorf("GetServiceByType(NonExistent) count = %v, want 0", len(notFound))
	}
}

func TestDocument_PrimaryVerificationMethod(t *testing.T) {
	d, _ := Parse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")

	// Empty document
	emptyDoc := NewDocument(d)
	if emptyDoc.PrimaryVerificationMethod() != nil {
		t.Error("PrimaryVerificationMethod() should return nil for empty document")
	}

	// Document with verification method
	doc := NewDocument(d)
	vm := VerificationMethod{
		ID:                 d.Fragment("key-1"),
		Type:               Ed25519VerificationKey2020,
		Controller:         d,
		PublicKeyMultibase: "z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
	}
	doc.AddVerificationMethod(vm)

	primary := doc.PrimaryVerificationMethod()
	if primary == nil {
		t.Fatal("PrimaryVerificationMethod() returned nil")
	}
	if primary.ID != vm.ID {
		t.Errorf("PrimaryVerificationMethod().ID = %v, want %v", primary.ID, vm.ID)
	}
}

func TestDocument_IsDeactivated(t *testing.T) {
	d, _ := Parse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")

	doc := NewDocument(d)
	if doc.IsDeactivated() {
		t.Error("new document should not be deactivated")
	}

	doc.Deactivated = true
	if !doc.IsDeactivated() {
		t.Error("document with Deactivated=true should be deactivated")
	}
}

func TestDocument_JSONRoundTrip(t *testing.T) {
	d, _ := Parse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")

	doc := NewDocument(d)
	doc.AddVerificationMethod(VerificationMethod{
		ID:                 d.Fragment("key-1"),
		Type:               Ed25519VerificationKey2020,
		Controller:         d,
		PublicKeyMultibase: "z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
	})
	doc.AddAuthentication(d.Fragment("key-1"))
	doc.AddAssertionMethod(d.Fragment("key-1"))
	doc.AddService(Service{
		ID:              d.Fragment("service-1"),
		Type:            "LinkedDomains",
		ServiceEndpoint: ServiceEndpoint{URI: "https://example.com"},
	})

	// Marshal
	data, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v", err)
	}

	// Unmarshal
	var decoded Document
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}

	// Verify
	if !decoded.ID.Equals(doc.ID) {
		t.Errorf("ID mismatch: got %v, want %v", decoded.ID, doc.ID)
	}

	if len(decoded.VerificationMethod) != len(doc.VerificationMethod) {
		t.Errorf("VerificationMethod count mismatch: got %v, want %v",
			len(decoded.VerificationMethod), len(doc.VerificationMethod))
	}

	if len(decoded.Authentication) != len(doc.Authentication) {
		t.Errorf("Authentication count mismatch: got %v, want %v",
			len(decoded.Authentication), len(doc.Authentication))
	}

	if len(decoded.Service) != len(doc.Service) {
		t.Errorf("Service count mismatch: got %v, want %v",
			len(decoded.Service), len(doc.Service))
	}
}

func TestVerificationRelationship_JSONMarshal(t *testing.T) {
	// Reference
	ref := VerificationRelationship{Reference: "did:key:z6Mk...#key-1"}
	data, err := json.Marshal(ref)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v", err)
	}
	if string(data) != `"did:key:z6Mk...#key-1"` {
		t.Errorf("Marshal reference = %v, want %v", string(data), `"did:key:z6Mk...#key-1"`)
	}

	// Embedded
	d, _ := Parse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")
	embedded := VerificationRelationship{
		Embedded: &VerificationMethod{
			ID:         d.Fragment("key-1"),
			Type:       Ed25519VerificationKey2020,
			Controller: d,
		},
	}
	data, err = json.Marshal(embedded)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v", err)
	}
	// Should be an object, not a string
	if data[0] != '{' {
		t.Errorf("Marshal embedded should be object, got: %v", string(data))
	}
}

func TestVerificationRelationship_JSONUnmarshal(t *testing.T) {
	// Reference
	var ref VerificationRelationship
	if err := json.Unmarshal([]byte(`"did:key:z6Mk...#key-1"`), &ref); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}
	if ref.Reference != "did:key:z6Mk...#key-1" {
		t.Errorf("Reference = %v, want did:key:z6Mk...#key-1", ref.Reference)
	}
	if ref.Embedded != nil {
		t.Error("Embedded should be nil for reference")
	}

	// Embedded
	var embedded VerificationRelationship
	embeddedJSON := `{"id":"did:key:z6Mk...#key-1","type":"Ed25519VerificationKey2020","controller":"did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK"}`
	if err := json.Unmarshal([]byte(embeddedJSON), &embedded); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}
	if embedded.Embedded == nil {
		t.Fatal("Embedded should not be nil")
	}
	if embedded.Embedded.ID != "did:key:z6Mk...#key-1" {
		t.Errorf("Embedded.ID = %v, want did:key:z6Mk...#key-1", embedded.Embedded.ID)
	}
}

func TestVerificationRelationship_ID(t *testing.T) {
	// Reference
	ref := VerificationRelationship{Reference: "did:key:z6Mk...#key-1"}
	if ref.ID() != "did:key:z6Mk...#key-1" {
		t.Errorf("ID() = %v, want did:key:z6Mk...#key-1", ref.ID())
	}

	// Embedded
	embedded := VerificationRelationship{
		Embedded: &VerificationMethod{ID: "did:key:z6Mk...#key-2"},
	}
	if embedded.ID() != "did:key:z6Mk...#key-2" {
		t.Errorf("ID() = %v, want did:key:z6Mk...#key-2", embedded.ID())
	}
}

func TestServiceEndpoint_JSONMarshal(t *testing.T) {
	tests := []struct {
		name string
		se   ServiceEndpoint
		want string
	}{
		{
			name: "single URI",
			se:   ServiceEndpoint{URI: "https://example.com"},
			want: `"https://example.com"`,
		},
		{
			name: "multiple URIs",
			se:   ServiceEndpoint{URIs: []string{"https://example.com", "https://example.org"}},
			want: `["https://example.com","https://example.org"]`,
		},
		{
			name: "object",
			se:   ServiceEndpoint{Object: map[string]interface{}{"origins": []interface{}{"https://example.com"}}},
			want: `{"origins":["https://example.com"]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.se)
			if err != nil {
				t.Fatalf("json.Marshal() error: %v", err)
			}
			if string(data) != tt.want {
				t.Errorf("Marshal() = %v, want %v", string(data), tt.want)
			}
		})
	}
}

func TestServiceEndpoint_JSONUnmarshal(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantURI string
		wantLen int
	}{
		{
			name:    "single URI",
			input:   `"https://example.com"`,
			wantURI: "https://example.com",
		},
		{
			name:    "multiple URIs",
			input:   `["https://example.com","https://example.org"]`,
			wantLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var se ServiceEndpoint
			if err := json.Unmarshal([]byte(tt.input), &se); err != nil {
				t.Fatalf("json.Unmarshal() error: %v", err)
			}

			if tt.wantURI != "" && se.URI != tt.wantURI {
				t.Errorf("URI = %v, want %v", se.URI, tt.wantURI)
			}

			if tt.wantLen > 0 && len(se.URIs) != tt.wantLen {
				t.Errorf("URIs count = %v, want %v", len(se.URIs), tt.wantLen)
			}
		})
	}
}
