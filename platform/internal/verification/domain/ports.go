package domain

import (
	"context"

	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// ============================================================================
// Data Fetcher Port
// ============================================================================

// DataFetcher fetches verified data from an external provider.
type DataFetcher interface {
	// FetchData retrieves user data from the specified provider using the access token.
	FetchData(ctx context.Context, provider ProviderType, accessToken string) (map[string]any, error)
}

// ============================================================================
// Credential Issuer Port
// ============================================================================

// CredentialIssuer issues a verifiable credential based on fetched data.
type CredentialIssuer interface {
	// IssueCredential creates a new credential from the verified provider data.
	IssueCredential(ctx context.Context, userID string, provider ProviderType, data map[string]any) (credentialID string, err error)
}

// ============================================================================
// OAuth State Repository Port
// ============================================================================

// OAuthStateRepository manages OAuth state tokens for CSRF protection.
type OAuthStateRepository interface {
	// Save stores an OAuth state mapping to a verification ID.
	Save(ctx context.Context, state string, verificationID VerificationID) error

	// GetVerificationID retrieves the verification ID for a given OAuth state.
	GetVerificationID(ctx context.Context, state string) (VerificationID, error)

	// Delete removes an OAuth state after use or expiration.
	Delete(ctx context.Context, state string) error
}

// ============================================================================
// Provider Token Repository Port
// ============================================================================

// ProviderTokenRepository manages provider access tokens for data fetching.
type ProviderTokenRepository interface {
	// SaveToken stores an access token for a verification.
	SaveToken(ctx context.Context, verificationID VerificationID, accessToken string) error

	// GetToken retrieves the access token for a verification.
	GetToken(ctx context.Context, verificationID VerificationID) (string, error)

	// DeleteToken removes the access token for a verification.
	DeleteToken(ctx context.Context, verificationID VerificationID) error
}

// ============================================================================
// Event Publisher Port
// ============================================================================

// EventPublisher publishes domain events to the event bus.
type EventPublisher interface {
	// Publish publishes one or more domain events.
	Publish(ctx context.Context, events ...eventsourcing.Event) error
}

// ============================================================================
// Null Implementations
// ============================================================================

// NullDataFetcher is a no-op implementation of DataFetcher.
type NullDataFetcher struct{}

// Compile-time check.
var _ DataFetcher = NullDataFetcher{}

// FetchData is a no-op.
func (NullDataFetcher) FetchData(_ context.Context, _ ProviderType, _ string) (map[string]any, error) {
	return nil, nil
}

// NullCredentialIssuer is a no-op implementation of CredentialIssuer.
type NullCredentialIssuer struct{}

// Compile-time check.
var _ CredentialIssuer = NullCredentialIssuer{}

// IssueCredential is a no-op.
func (NullCredentialIssuer) IssueCredential(_ context.Context, _ string, _ ProviderType, _ map[string]any) (string, error) {
	return "", nil
}

// NullOAuthStateRepository is a no-op implementation of OAuthStateRepository.
type NullOAuthStateRepository struct{}

// Compile-time check.
var _ OAuthStateRepository = NullOAuthStateRepository{}

// Save is a no-op.
func (NullOAuthStateRepository) Save(_ context.Context, _ string, _ VerificationID) error {
	return nil
}

// GetVerificationID is a no-op.
func (NullOAuthStateRepository) GetVerificationID(_ context.Context, _ string) (VerificationID, error) {
	return VerificationID{}, nil
}

// Delete is a no-op.
func (NullOAuthStateRepository) Delete(_ context.Context, _ string) error {
	return nil
}

// NullProviderTokenRepository is a no-op implementation of ProviderTokenRepository.
type NullProviderTokenRepository struct{}

// Compile-time check.
var _ ProviderTokenRepository = NullProviderTokenRepository{}

// SaveToken is a no-op.
func (NullProviderTokenRepository) SaveToken(_ context.Context, _ VerificationID, _ string) error {
	return nil
}

// GetToken is a no-op.
func (NullProviderTokenRepository) GetToken(_ context.Context, _ VerificationID) (string, error) {
	return "", nil
}

// DeleteToken is a no-op.
func (NullProviderTokenRepository) DeleteToken(_ context.Context, _ VerificationID) error {
	return nil
}

// NullEventPublisher is a no-op implementation of EventPublisher.
type NullEventPublisher struct{}

// Compile-time check.
var _ EventPublisher = NullEventPublisher{}

// Publish is a no-op.
func (NullEventPublisher) Publish(_ context.Context, _ ...eventsourcing.Event) error {
	return nil
}
