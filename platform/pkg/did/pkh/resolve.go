package pkh

import (
	"context"

	"github.com/0xsj/nexus/platform/pkg/chain"
	"github.com/0xsj/nexus/platform/pkg/did"
)

// ============================================================================
// Context URIs
// ============================================================================

const (
	// ContextDIDCore is the DID Core context.
	ContextDIDCore = "https://www.w3.org/ns/did/v1"

	// ContextSecp256k1Recovery is the secp256k1 recovery context.
	ContextSecp256k1Recovery = "https://w3id.org/security/suites/secp256k1recovery-2020/v2"

	// ContextEd25519 is the Ed25519 context.
	ContextEd25519 = "https://w3id.org/security/suites/ed25519-2020/v1"

	// ContextSecp256k1 is the secp256k1 context.
	ContextSecp256k1 = "https://w3id.org/security/suites/secp256k1-2019/v1"
)

// ============================================================================
// Verification Method Types
// ============================================================================

const (
	// VMTypeEcdsaSecp256k1Recovery is the verification method type for EVM chains.
	VMTypeEcdsaSecp256k1Recovery did.VerificationMethodType = "EcdsaSecp256k1RecoveryMethod2020"

	// VMTypeEd25519 is the verification method type for Solana.
	VMTypeEd25519 did.VerificationMethodType = "Ed25519VerificationKey2020"

	// VMTypeEcdsaSecp256k1 is the verification method type for Cosmos.
	VMTypeEcdsaSecp256k1 did.VerificationMethodType = "EcdsaSecp256k1VerificationKey2019"
)

// ============================================================================
// Resolver
// ============================================================================

// Resolver resolves did:pkh DIDs to DID Documents.
type Resolver struct {
	chainRegistry chain.ChainRegistry
}

// NewResolver creates a new did:pkh resolver.
func NewResolver(chainRegistry chain.ChainRegistry) *Resolver {
	return &Resolver{
		chainRegistry: chainRegistry,
	}
}

// Method returns the DID method this resolver handles.
func (r *Resolver) Method() did.Method {
	return did.MethodPKH
}

// Resolve resolves a did:pkh DID to a DID Document.
//
// The did:pkh method is "self-resolving" - the DID Document is deterministically
// derived from the DID itself without requiring external resolution.
func (r *Resolver) Resolve(ctx context.Context, d did.DID) (*did.Document, error) {
	const op = "pkh.Resolver.Resolve"

	// Validate method
	if d.Method() != did.MethodPKH {
		return nil, did.ErrUnsupportedMethod(op, string(d.Method()))
	}

	// Parse the account ID
	account, err := ParseDID(d)
	if err != nil {
		return nil, err
	}

	// Build the DID Document
	doc, err := r.buildDocument(d, account)
	if err != nil {
		return nil, did.ErrResolutionFailed(op, d.String(), err)
	}

	return doc, nil
}

// ============================================================================
// Document Building
// ============================================================================

// buildDocument constructs a DID Document for a did:pkh DID.
func (r *Resolver) buildDocument(d did.DID, account AccountID) (*did.Document, error) {
	// Determine verification method type based on namespace
	vmType := verificationMethodTypeForNamespace(account.Namespace)
	blockchainAccountID := account.String()

	// Verification method ID
	vmID := d.String() + "#blockchainAccountId"

	// Create the verification method
	vm := did.VerificationMethod{
		ID:                  vmID,
		Type:                vmType,
		Controller:          d,
		BlockchainAccountId: blockchainAccountID,
	}

	// Build context list
	contexts := []string{ContextDIDCore}
	if nsContext := contextForNamespace(account.Namespace); nsContext != "" {
		contexts = append(contexts, nsContext)
	}

	// Build the document
	doc := &did.Document{
		Context: contexts,
		ID:      d,
		VerificationMethod: []did.VerificationMethod{
			vm,
		},
		Authentication: []did.VerificationRelationship{
			{Reference: vmID},
		},
		AssertionMethod: []did.VerificationRelationship{
			{Reference: vmID},
		},
	}

	return doc, nil
}

// ============================================================================
// Namespace Helpers
// ============================================================================

// verificationMethodTypeForNamespace returns the appropriate verification method type.
func verificationMethodTypeForNamespace(namespace string) did.VerificationMethodType {
	switch namespace {
	case NamespaceEIP155:
		return VMTypeEcdsaSecp256k1Recovery
	case NamespaceSolana:
		return VMTypeEd25519
	case NamespaceCosmos:
		return VMTypeEcdsaSecp256k1
	default:
		return VMTypeEcdsaSecp256k1Recovery // Default to EVM
	}
}

// contextForNamespace returns additional context URIs for a namespace.
func contextForNamespace(namespace string) string {
	switch namespace {
	case NamespaceEIP155:
		return ContextSecp256k1Recovery
	case NamespaceSolana:
		return ContextEd25519
	case NamespaceCosmos:
		return ContextSecp256k1
	default:
		return ""
	}
}

// ============================================================================
// Convenience Functions
// ============================================================================

// ResolveString resolves a did:pkh DID string to a DID Document.
func (r *Resolver) ResolveString(ctx context.Context, didString string) (*did.Document, error) {
	d, err := did.Parse(didString)
	if err != nil {
		return nil, err
	}

	return r.Resolve(ctx, d)
}

// CanResolve returns true if this resolver can resolve the given DID.
func (r *Resolver) CanResolve(d did.DID) bool {
	return d.Method() == did.MethodPKH
}

// CanResolveString returns true if this resolver can resolve the given DID string.
func (r *Resolver) CanResolveString(didString string) bool {
	d, err := did.Parse(didString)
	if err != nil {
		return false
	}
	return r.CanResolve(d)
}

// ============================================================================
// Interface Compliance
// ============================================================================

var _ did.Resolver = (*Resolver)(nil)
