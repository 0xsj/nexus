package did

import "context"

// Resolver resolves DIDs to DID Documents.
// Each DID method has its own resolver implementation.
type Resolver interface {
	// Resolve resolves a DID to a DID Document.
	// Returns ErrDIDNotFound if the DID cannot be resolved.
	// Returns ErrDeactivated if the DID has been deactivated.
	Resolve(ctx context.Context, did DID) (*Document, error)

	// Method returns the DID method this resolver handles.
	Method() Method
}

// MultiResolver resolves DIDs using multiple method-specific resolvers.
type MultiResolver struct {
	resolvers map[Method]Resolver
}

// NewMultiResolver creates a new multi-resolver.
func NewMultiResolver() *MultiResolver {
	return &MultiResolver{
		resolvers: make(map[Method]Resolver),
	}
}

// Register registers a resolver for a DID method.
func (r *MultiResolver) Register(resolver Resolver) {
	r.resolvers[resolver.Method()] = resolver
}

// Resolve resolves a DID using the appropriate method resolver.
func (r *MultiResolver) Resolve(ctx context.Context, did DID) (*Document, error) {
	const op = "did.MultiResolver.Resolve"

	resolver, ok := r.resolvers[did.Method()]
	if !ok {
		return nil, ErrUnsupportedMethod(op, did.Method().String())
	}

	return resolver.Resolve(ctx, did)
}

// SupportsMethod returns true if the resolver supports the given method.
func (r *MultiResolver) SupportsMethod(method Method) bool {
	_, ok := r.resolvers[method]
	return ok
}

// SupportedMethods returns all methods supported by this resolver.
func (r *MultiResolver) SupportedMethods() []Method {
	methods := make([]Method, 0, len(r.resolvers))
	for method := range r.resolvers {
		methods = append(methods, method)
	}
	return methods
}

// ResolutionResult contains the result of a DID resolution.
// Includes metadata about the resolution process.
type ResolutionResult struct {
	// Document is the resolved DID Document.
	Document *Document

	// Metadata contains resolution metadata.
	Metadata ResolutionMetadata
}

// ResolutionMetadata contains metadata about a DID resolution.
type ResolutionMetadata struct {
	// ContentType is the media type of the document.
	ContentType string `json:"contentType,omitempty"`

	// Retrieved is when the document was retrieved.
	Retrieved string `json:"retrieved,omitempty"`

	// Duration is how long resolution took in milliseconds.
	Duration int64 `json:"duration,omitempty"`

	// Error is set if resolution failed.
	Error string `json:"error,omitempty"`

	// ErrorMessage provides additional error details.
	ErrorMessage string `json:"errorMessage,omitempty"`
}

// DocumentMetadata contains metadata about a DID Document.
type DocumentMetadata struct {
	// Created is when the document was created.
	Created string `json:"created,omitempty"`

	// Updated is when the document was last updated.
	Updated string `json:"updated,omitempty"`

	// Deactivated indicates if the DID has been deactivated.
	Deactivated bool `json:"deactivated,omitempty"`

	// VersionId is the version of the document.
	VersionId string `json:"versionId,omitempty"`

	// NextVersionId is the next version of the document.
	NextVersionId string `json:"nextVersionId,omitempty"`

	// NextUpdate is when the document will next be updated.
	NextUpdate string `json:"nextUpdate,omitempty"`
}

// ResolverOptions configures resolution behavior.
type ResolverOptions struct {
	// Accept specifies preferred media types for the document.
	Accept []string

	// NoCache disables caching of resolved documents.
	NoCache bool

	// VersionId requests a specific version of the document.
	VersionId string

	// VersionTime requests the document as of a specific time.
	VersionTime string
}

// DefaultResolverOptions returns default resolver options.
func DefaultResolverOptions() ResolverOptions {
	return ResolverOptions{
		Accept:  []string{"application/did+json", "application/did+ld+json"},
		NoCache: false,
	}
}
