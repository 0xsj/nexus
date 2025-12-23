package did

import (
	"encoding/json"
	"time"
)

// Document represents a DID Document.
// Contains the public keys, authentication methods, and service endpoints
// associated with a DID.
//
// Spec: https://www.w3.org/TR/did-core/
type Document struct {
	// Context defines the JSON-LD context(s).
	Context []string `json:"@context"`

	// ID is the DID subject of this document.
	ID DID `json:"id"`

	// Controller is the DID(s) that can make changes to this document.
	// If empty, the subject (ID) is the controller.
	Controller []DID `json:"controller,omitempty"`

	// VerificationMethod lists the public keys associated with the DID.
	VerificationMethod []VerificationMethod `json:"verificationMethod,omitempty"`

	// Authentication lists methods for authenticating as the DID subject.
	// Can be verification method IDs or embedded verification methods.
	Authentication []VerificationRelationship `json:"authentication,omitempty"`

	// AssertionMethod lists methods for issuing verifiable credentials.
	AssertionMethod []VerificationRelationship `json:"assertionMethod,omitempty"`

	// KeyAgreement lists methods for key agreement (encryption).
	KeyAgreement []VerificationRelationship `json:"keyAgreement,omitempty"`

	// CapabilityInvocation lists methods for invoking capabilities.
	CapabilityInvocation []VerificationRelationship `json:"capabilityInvocation,omitempty"`

	// CapabilityDelegation lists methods for delegating capabilities.
	CapabilityDelegation []VerificationRelationship `json:"capabilityDelegation,omitempty"`

	// Service lists service endpoints associated with the DID.
	Service []Service `json:"service,omitempty"`

	// AlsoKnownAs lists other URIs that refer to the same subject.
	AlsoKnownAs []string `json:"alsoKnownAs,omitempty"`

	// Created is when the document was created.
	Created *time.Time `json:"created,omitempty"`

	// Updated is when the document was last updated.
	Updated *time.Time `json:"updated,omitempty"`

	// Deactivated indicates if the DID has been deactivated.
	Deactivated bool `json:"deactivated,omitempty"`
}

// VerificationMethod represents a public key or other verification method.
type VerificationMethod struct {
	// ID is the unique identifier for this method.
	// Format: <did>#<fragment>
	// Example: did:key:z6Mk...#z6Mk...
	ID string `json:"id"`

	// Type specifies the verification method type.
	// Examples: Ed25519VerificationKey2020, JsonWebKey2020
	Type VerificationMethodType `json:"type"`

	// Controller is the DID that controls this verification method.
	Controller DID `json:"controller"`

	// PublicKeyMultibase is the multibase-encoded public key.
	// Used with Ed25519VerificationKey2020, etc.
	PublicKeyMultibase string `json:"publicKeyMultibase,omitempty"`

	// PublicKeyJwk is the public key in JWK format.
	// Used with JsonWebKey2020.
	PublicKeyJwk *JWK `json:"publicKeyJwk,omitempty"`

	// PublicKeyBase58 is the base58-encoded public key (legacy).
	PublicKeyBase58 string `json:"publicKeyBase58,omitempty"`

	// BlockchainAccountId is used for did:pkh and similar methods.
	// Format: <chain-id>:<account-address>
	BlockchainAccountId string `json:"blockchainAccountId,omitempty"`
}

// VerificationMethodType defines the type of verification method.
type VerificationMethodType string

const (
	// Ed25519VerificationKey2020 is an Ed25519 public key.
	Ed25519VerificationKey2020 VerificationMethodType = "Ed25519VerificationKey2020"

	// JsonWebKey2020 is a key in JWK format.
	JsonWebKey2020 VerificationMethodType = "JsonWebKey2020"

	// EcdsaSecp256k1VerificationKey2019 is a secp256k1 public key.
	EcdsaSecp256k1VerificationKey2019 VerificationMethodType = "EcdsaSecp256k1VerificationKey2019"

	// Bls12381G2Key2020 is a BLS12-381 G2 public key for BBS+.
	Bls12381G2Key2020 VerificationMethodType = "Bls12381G2Key2020"
)

// VerificationRelationship can be either a reference (string ID)
// or an embedded verification method.
type VerificationRelationship struct {
	// Reference is a reference to a verification method by ID.
	Reference string

	// Embedded is an embedded verification method.
	Embedded *VerificationMethod
}

// MarshalJSON implements json.Marshaler.
func (vr VerificationRelationship) MarshalJSON() ([]byte, error) {
	if vr.Embedded != nil {
		return json.Marshal(vr.Embedded)
	}
	return json.Marshal(vr.Reference)
}

// UnmarshalJSON implements json.Unmarshaler.
func (vr *VerificationRelationship) UnmarshalJSON(data []byte) error {
	// Try string first (reference)
	var ref string
	if err := json.Unmarshal(data, &ref); err == nil {
		vr.Reference = ref
		vr.Embedded = nil
		return nil
	}

	// Try embedded verification method
	var vm VerificationMethod
	if err := json.Unmarshal(data, &vm); err == nil {
		vr.Embedded = &vm
		vr.Reference = ""
		return nil
	}

	return ErrInvalidDocument("VerificationRelationship.UnmarshalJSON", "invalid verification relationship")
}

// ID returns the ID of the verification relationship.
func (vr VerificationRelationship) ID() string {
	if vr.Embedded != nil {
		return vr.Embedded.ID
	}
	return vr.Reference
}

// Service represents a service endpoint.
type Service struct {
	// ID is the unique identifier for this service.
	ID string `json:"id"`

	// Type specifies the service type.
	// Examples: LinkedDomains, CredentialRegistry
	Type string `json:"type"`

	// ServiceEndpoint is the URL or object for the service.
	ServiceEndpoint ServiceEndpoint `json:"serviceEndpoint"`
}

// ServiceEndpoint can be a string URL or a complex object.
type ServiceEndpoint struct {
	// URI is a single endpoint URI.
	URI string

	// URIs is multiple endpoint URIs.
	URIs []string

	// Object is a complex endpoint object.
	Object map[string]interface{}
}

// MarshalJSON implements json.Marshaler.
func (se ServiceEndpoint) MarshalJSON() ([]byte, error) {
	if se.URI != "" {
		return json.Marshal(se.URI)
	}
	if len(se.URIs) > 0 {
		return json.Marshal(se.URIs)
	}
	if se.Object != nil {
		return json.Marshal(se.Object)
	}
	return []byte("null"), nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (se *ServiceEndpoint) UnmarshalJSON(data []byte) error {
	// Try string first
	var uri string
	if err := json.Unmarshal(data, &uri); err == nil {
		se.URI = uri
		return nil
	}

	// Try array of strings
	var uris []string
	if err := json.Unmarshal(data, &uris); err == nil {
		se.URIs = uris
		return nil
	}

	// Try object
	var obj map[string]interface{}
	if err := json.Unmarshal(data, &obj); err == nil {
		se.Object = obj
		return nil
	}

	return ErrInvalidDocument("ServiceEndpoint.UnmarshalJSON", "invalid service endpoint")
}

// JWK represents a JSON Web Key.
type JWK struct {
	// Kty is the key type (EC, OKP, RSA).
	Kty string `json:"kty"`

	// Crv is the curve (Ed25519, secp256k1, P-256).
	Crv string `json:"crv,omitempty"`

	// X is the public key x-coordinate (base64url).
	X string `json:"x,omitempty"`

	// Y is the public key y-coordinate (base64url, EC keys only).
	Y string `json:"y,omitempty"`

	// Use is the intended use (sig, enc).
	Use string `json:"use,omitempty"`

	// Kid is the key ID.
	Kid string `json:"kid,omitempty"`
}

// ============================================================================
// Document Builder
// ============================================================================

// NewDocument creates a new DID document with default context.
func NewDocument(id DID) *Document {
	return &Document{
		Context: []string{
			"https://www.w3.org/ns/did/v1",
			"https://w3id.org/security/suites/ed25519-2020/v1",
		},
		ID: id,
	}
}

// AddVerificationMethod adds a verification method to the document.
func (d *Document) AddVerificationMethod(vm VerificationMethod) *Document {
	d.VerificationMethod = append(d.VerificationMethod, vm)
	return d
}

// AddAuthentication adds an authentication relationship.
func (d *Document) AddAuthentication(ref string) *Document {
	d.Authentication = append(d.Authentication, VerificationRelationship{Reference: ref})
	return d
}

// AddAssertionMethod adds an assertion method relationship.
func (d *Document) AddAssertionMethod(ref string) *Document {
	d.AssertionMethod = append(d.AssertionMethod, VerificationRelationship{Reference: ref})
	return d
}

// AddKeyAgreement adds a key agreement relationship.
func (d *Document) AddKeyAgreement(ref string) *Document {
	d.KeyAgreement = append(d.KeyAgreement, VerificationRelationship{Reference: ref})
	return d
}

// AddService adds a service endpoint.
func (d *Document) AddService(svc Service) *Document {
	d.Service = append(d.Service, svc)
	return d
}

// SetController sets the controller(s) of the document.
func (d *Document) SetController(controllers ...DID) *Document {
	d.Controller = controllers
	return d
}

// ============================================================================
// Document Queries
// ============================================================================

// GetVerificationMethod finds a verification method by ID.
func (d *Document) GetVerificationMethod(id string) *VerificationMethod {
	for i := range d.VerificationMethod {
		if d.VerificationMethod[i].ID == id {
			return &d.VerificationMethod[i]
		}
	}
	return nil
}

// GetService finds a service by ID.
func (d *Document) GetService(id string) *Service {
	for i := range d.Service {
		if d.Service[i].ID == id {
			return &d.Service[i]
		}
	}
	return nil
}

// GetServiceByType finds services by type.
func (d *Document) GetServiceByType(serviceType string) []Service {
	var services []Service
	for _, svc := range d.Service {
		if svc.Type == serviceType {
			services = append(services, svc)
		}
	}
	return services
}

// IsDeactivated returns true if the DID document is deactivated.
func (d *Document) IsDeactivated() bool {
	return d.Deactivated
}

// PrimaryVerificationMethod returns the first verification method.
// Returns nil if no verification methods exist.
func (d *Document) PrimaryVerificationMethod() *VerificationMethod {
	if len(d.VerificationMethod) == 0 {
		return nil
	}
	return &d.VerificationMethod[0]
}
