package domain

import (
	"context"
)

// ============================================================================
// Provider Service
// ============================================================================

// ProviderService interacts with external OAuth providers to fetch user data.
type ProviderService interface {
	// GetAuthorizationURL returns the OAuth authorization URL for a provider.
	GetAuthorizationURL(ctx context.Context, params AuthorizationURLParams) (string, error)

	// ExchangeCode exchanges an authorization code for tokens.
	ExchangeCode(ctx context.Context, params ExchangeCodeParams) (*OAuthTokens, error)

	// RefreshTokens refreshes expired access tokens.
	RefreshTokens(ctx context.Context, params RefreshTokensParams) (*OAuthTokens, error)

	// FetchProfile fetches the user's profile from the provider.
	FetchProfile(ctx context.Context, params FetchProfileParams) (*ProviderProfile, error)

	// FetchCredentialData fetches data needed for credential issuance.
	FetchCredentialData(ctx context.Context, params FetchCredentialDataParams) (*CredentialData, error)

	// RevokeTokens revokes the user's tokens with the provider.
	RevokeTokens(ctx context.Context, provider Provider, accessToken string) error

	// SupportsProvider returns true if the provider is supported.
	SupportsProvider(provider Provider) bool
}

// AuthorizationURLParams contains parameters for generating an OAuth authorization URL.
type AuthorizationURLParams struct {
	Provider    Provider
	State       string
	RedirectURI string
	Scopes      []string
}

// ExchangeCodeParams contains parameters for exchanging an authorization code.
type ExchangeCodeParams struct {
	Provider    Provider
	Code        string
	RedirectURI string
}

// RefreshTokensParams contains parameters for refreshing tokens.
type RefreshTokensParams struct {
	Provider     Provider
	RefreshToken string
}

// FetchProfileParams contains parameters for fetching a user profile.
type FetchProfileParams struct {
	Provider    Provider
	AccessToken string
}

// FetchCredentialDataParams contains parameters for fetching credential data.
type FetchCredentialDataParams struct {
	Provider       Provider
	CredentialType CredentialType
	AccessToken    string
	Profile        *ProviderProfile
}

// CredentialData contains data fetched from a provider for credential issuance.
type CredentialData struct {
	// Provider is the source provider.
	Provider Provider

	// CredentialType is the type of credential this data supports.
	CredentialType CredentialType

	// Claims contains the verified claims for the credential.
	Claims map[string]any

	// Evidence contains supporting evidence for the claims.
	Evidence []CredentialEvidence

	// Profile is the user's profile from the provider.
	Profile *ProviderProfile
}

// CredentialEvidence contains evidence supporting a credential claim.
type CredentialEvidence struct {
	// Type is the evidence type.
	Type string

	// Source is where the evidence came from.
	Source string

	// VerifiedAt is when the evidence was verified.
	VerifiedAt string

	// Data contains additional evidence data.
	Data map[string]any
}

// ============================================================================
// Credential Issuer Service
// ============================================================================

// CredentialIssuerService issues verifiable credentials based on verified data.
type CredentialIssuerService interface {
	// IssueCredential issues a credential based on verified provider data.
	IssueCredential(ctx context.Context, params IssueCredentialParams) (*IssuedCredential, error)

	// GetIssuerDID returns the DID used for issuing credentials.
	GetIssuerDID(ctx context.Context) (string, error)
}

// IssueCredentialParams contains parameters for issuing a credential.
type IssueCredentialParams struct {
	// HolderDID is the DID of the credential holder.
	HolderDID string

	// CredentialType is the type of credential to issue.
	CredentialType CredentialType

	// Claims contains the verified claims.
	Claims map[string]any

	// Evidence contains supporting evidence.
	Evidence []CredentialEvidence

	// ExpiresIn is how long until the credential expires (optional).
	ExpiresIn *int64
}

// IssuedCredential contains the result of credential issuance.
type IssuedCredential struct {
	// ID is the credential ID.
	ID string

	// CredentialType is the type of credential issued.
	CredentialType CredentialType

	// SignedVC is the signed verifiable credential (JWT or JSON-LD).
	SignedVC string

	// IssuedAt is when the credential was issued.
	IssuedAt string

	// ExpiresAt is when the credential expires (optional).
	ExpiresAt *string
}

// ============================================================================
// OAuth State Service
// ============================================================================

// OAuthStateService manages OAuth state for CSRF protection.
type OAuthStateService interface {
	// GenerateState generates a new OAuth state.
	GenerateState(ctx context.Context, params GenerateStateParams) (*OAuthState, error)

	// ValidateState validates an OAuth state and returns the associated data.
	ValidateState(ctx context.Context, value string) (*OAuthState, error)

	// ConsumeState validates and consumes an OAuth state (one-time use).
	ConsumeState(ctx context.Context, value string) (*OAuthState, error)

	// InvalidateState invalidates an OAuth state without consuming it.
	InvalidateState(ctx context.Context, value string) error
}

// GenerateStateParams contains parameters for generating OAuth state.
type GenerateStateParams struct {
	UserID         string
	Provider       Provider
	CredentialType CredentialType
	RedirectURL    string
	TTL            int64 // seconds
}

// ============================================================================
// User DID Resolver
// ============================================================================

// UserDIDResolver resolves a user's DID for credential issuance.
type UserDIDResolver interface {
	// ResolvePrimaryDID resolves the primary DID for a user.
	ResolvePrimaryDID(ctx context.Context, userID string) (string, error)

	// ResolveOrCreateDID resolves or creates a DID for a user.
	ResolveOrCreateDID(ctx context.Context, userID string) (string, error)
}

// ============================================================================
// Provider Registry
// ============================================================================

// ProviderRegistry manages available provider implementations.
type ProviderRegistry interface {
	// GetProvider returns the provider implementation for a provider type.
	GetProvider(provider Provider) (ProviderAdapter, error)

	// ListProviders returns all available providers.
	ListProviders() []Provider

	// IsProviderEnabled returns true if the provider is enabled.
	IsProviderEnabled(provider Provider) bool
}

// ProviderAdapter is the interface that provider implementations must satisfy.
type ProviderAdapter interface {
	// Provider returns the provider type.
	Provider() Provider

	// GetAuthorizationURL returns the OAuth authorization URL.
	GetAuthorizationURL(state string, redirectURI string, scopes []string) (string, error)

	// ExchangeCode exchanges an authorization code for tokens.
	ExchangeCode(ctx context.Context, code string, redirectURI string) (*OAuthTokens, error)

	// RefreshTokens refreshes expired tokens.
	RefreshTokens(ctx context.Context, refreshToken string) (*OAuthTokens, error)

	// FetchProfile fetches the user's profile.
	FetchProfile(ctx context.Context, accessToken string) (*ProviderProfile, error)

	// FetchCredentialData fetches data for credential issuance.
	FetchCredentialData(ctx context.Context, credentialType CredentialType, accessToken string, profile *ProviderProfile) (*CredentialData, error)

	// RevokeTokens revokes tokens with the provider.
	RevokeTokens(ctx context.Context, accessToken string) error

	// SupportedCredentialTypes returns the credential types this provider supports.
	SupportedCredentialTypes() []CredentialType

	// RequiredScopes returns the OAuth scopes required for a credential type.
	RequiredScopes(credentialType CredentialType) []string
}
