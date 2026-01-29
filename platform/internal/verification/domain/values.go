package domain

import "time"

// ============================================================================
// Provider
// ============================================================================

// Provider represents a verification source provider.
type Provider string

const (
	ProviderGitHub   Provider = "github"
	ProviderLinkedIn Provider = "linkedin"
	ProviderGoogle   Provider = "google"
	ProviderTwitter  Provider = "twitter"
	ProviderDiscord  Provider = "discord"
)

// String returns the string representation.
func (p Provider) String() string {
	return string(p)
}

// IsValid checks if the provider is supported.
func (p Provider) IsValid() bool {
	switch p {
	case ProviderGitHub, ProviderLinkedIn, ProviderGoogle, ProviderTwitter, ProviderDiscord:
		return true
	default:
		return false
	}
}

// DisplayName returns the human-readable display name.
func (p Provider) DisplayName() string {
	switch p {
	case ProviderGitHub:
		return "GitHub"
	case ProviderLinkedIn:
		return "LinkedIn"
	case ProviderGoogle:
		return "Google"
	case ProviderTwitter:
		return "Twitter/X"
	case ProviderDiscord:
		return "Discord"
	default:
		return string(p)
	}
}

// CredentialTypes returns the credential types this provider can issue.
func (p Provider) CredentialTypes() []CredentialType {
	switch p {
	case ProviderGitHub:
		return []CredentialType{
			CredentialTypeGitHubContributor,
			CredentialTypeGitHubAccount,
		}
	case ProviderLinkedIn:
		return []CredentialType{
			CredentialTypeLinkedInEmployment,
			CredentialTypeLinkedInAccount,
		}
	case ProviderGoogle:
		return []CredentialType{CredentialTypeGoogleAccount}
	case ProviderTwitter:
		return []CredentialType{CredentialTypeTwitterAccount}
	case ProviderDiscord:
		return []CredentialType{CredentialTypeDiscordAccount}
	default:
		return nil
	}
}

// ParseProvider parses a string into a Provider.
func ParseProvider(s string) (Provider, bool) {
	p := Provider(s)
	return p, p.IsValid()
}

// AllProviders returns all supported providers.
func AllProviders() []Provider {
	return []Provider{
		ProviderGitHub,
		ProviderLinkedIn,
		ProviderGoogle,
		ProviderTwitter,
		ProviderDiscord,
	}
}

// ============================================================================
// Verification Status
// ============================================================================

// VerificationStatus represents the status of a verification flow.
type VerificationStatus string

const (
	// StatusPending indicates the verification has been initiated.
	StatusPending VerificationStatus = "pending"

	// StatusAuthorized indicates OAuth authorization was successful.
	StatusAuthorized VerificationStatus = "authorized"

	// StatusFetching indicates provider data is being fetched.
	StatusFetching VerificationStatus = "fetching"

	// StatusCompleted indicates verification completed successfully.
	StatusCompleted VerificationStatus = "completed"

	// StatusFailed indicates verification failed.
	StatusFailed VerificationStatus = "failed"

	// StatusExpired indicates the verification flow expired.
	StatusExpired VerificationStatus = "expired"
)

// String returns the string representation.
func (s VerificationStatus) String() string {
	return string(s)
}

// IsValid checks if the status is valid.
func (s VerificationStatus) IsValid() bool {
	switch s {
	case StatusPending, StatusAuthorized, StatusFetching, StatusCompleted, StatusFailed, StatusExpired:
		return true
	default:
		return false
	}
}

// IsTerminal checks if the status is a terminal state.
func (s VerificationStatus) IsTerminal() bool {
	return s == StatusCompleted || s == StatusFailed || s == StatusExpired
}

// IsActive checks if the verification is still in progress.
func (s VerificationStatus) IsActive() bool {
	return s == StatusPending || s == StatusAuthorized || s == StatusFetching
}

// CanTransitionTo checks if a transition to the target status is valid.
func (s VerificationStatus) CanTransitionTo(target VerificationStatus) bool {
	switch s {
	case StatusPending:
		return target == StatusAuthorized || target == StatusFailed || target == StatusExpired
	case StatusAuthorized:
		return target == StatusFetching || target == StatusFailed || target == StatusExpired
	case StatusFetching:
		return target == StatusCompleted || target == StatusFailed
	case StatusCompleted, StatusFailed, StatusExpired:
		return false // Terminal states
	default:
		return false
	}
}

// ParseVerificationStatus parses a string into a VerificationStatus.
func ParseVerificationStatus(s string) (VerificationStatus, bool) {
	status := VerificationStatus(s)
	return status, status.IsValid()
}

// ============================================================================
// Credential Type
// ============================================================================

// CredentialType represents the type of credential to be issued.
type CredentialType string

const (
	// GitHub credential types
	CredentialTypeGitHubContributor CredentialType = "GitHubContributorCredential"
	CredentialTypeGitHubAccount     CredentialType = "GitHubAccountCredential"

	// LinkedIn credential types
	CredentialTypeLinkedInEmployment CredentialType = "LinkedInEmploymentCredential"
	CredentialTypeLinkedInAccount    CredentialType = "LinkedInAccountCredential"

	// Other provider credential types
	CredentialTypeGoogleAccount  CredentialType = "GoogleAccountCredential"
	CredentialTypeTwitterAccount CredentialType = "TwitterAccountCredential"
	CredentialTypeDiscordAccount CredentialType = "DiscordAccountCredential"
)

// String returns the string representation.
func (c CredentialType) String() string {
	return string(c)
}

// IsValid checks if the credential type is valid.
func (c CredentialType) IsValid() bool {
	switch c {
	case CredentialTypeGitHubContributor, CredentialTypeGitHubAccount,
		CredentialTypeLinkedInEmployment, CredentialTypeLinkedInAccount,
		CredentialTypeGoogleAccount, CredentialTypeTwitterAccount,
		CredentialTypeDiscordAccount:
		return true
	default:
		return false
	}
}

// Provider returns the provider for this credential type.
func (c CredentialType) Provider() Provider {
	switch c {
	case CredentialTypeGitHubContributor, CredentialTypeGitHubAccount:
		return ProviderGitHub
	case CredentialTypeLinkedInEmployment, CredentialTypeLinkedInAccount:
		return ProviderLinkedIn
	case CredentialTypeGoogleAccount:
		return ProviderGoogle
	case CredentialTypeTwitterAccount:
		return ProviderTwitter
	case CredentialTypeDiscordAccount:
		return ProviderDiscord
	default:
		return ""
	}
}

// DisplayName returns the human-readable display name.
func (c CredentialType) DisplayName() string {
	switch c {
	case CredentialTypeGitHubContributor:
		return "GitHub Contributor"
	case CredentialTypeGitHubAccount:
		return "GitHub Account"
	case CredentialTypeLinkedInEmployment:
		return "LinkedIn Employment"
	case CredentialTypeLinkedInAccount:
		return "LinkedIn Account"
	case CredentialTypeGoogleAccount:
		return "Google Account"
	case CredentialTypeTwitterAccount:
		return "Twitter/X Account"
	case CredentialTypeDiscordAccount:
		return "Discord Account"
	default:
		return string(c)
	}
}

// ParseCredentialType parses a string into a CredentialType.
func ParseCredentialType(s string) (CredentialType, bool) {
	ct := CredentialType(s)
	return ct, ct.IsValid()
}

// ============================================================================
// OAuth State
// ============================================================================

// OAuthState represents the state parameter for OAuth flows.
type OAuthState struct {
	// Value is the random state string.
	Value string

	// VerificationID is the ID of the verification this state belongs to.
	VerificationID string

	// UserID is the user initiating the verification.
	UserID string

	// Provider is the OAuth provider.
	Provider Provider

	// CredentialType is the credential type being requested.
	CredentialType CredentialType

	// RedirectURL is where to redirect after OAuth completes.
	RedirectURL string

	// CreatedAt is when the state was created.
	CreatedAt time.Time

	// ExpiresAt is when the state expires.
	ExpiresAt time.Time
}

// NewOAuthState creates a new OAuth state.
func NewOAuthState(value, verificationID, userID string, provider Provider, credentialType CredentialType, redirectURL string, ttl time.Duration) OAuthState {
	now := time.Now()
	return OAuthState{
		Value:          value,
		VerificationID: verificationID,
		UserID:         userID,
		Provider:       provider,
		CredentialType: credentialType,
		RedirectURL:    redirectURL,
		CreatedAt:      now,
		ExpiresAt:      now.Add(ttl),
	}
}

// IsExpired returns true if the state has expired.
func (s OAuthState) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// IsValid returns true if the state is valid (not expired).
func (s OAuthState) IsValid() bool {
	return !s.IsExpired() && s.Value != ""
}

// ============================================================================
// OAuth Tokens
// ============================================================================

// OAuthTokens contains tokens received from an OAuth provider.
type OAuthTokens struct {
	// AccessToken is the token for API access.
	AccessToken string

	// RefreshToken is the token for refreshing access (optional).
	RefreshToken string

	// TokenType is the type of token (usually "Bearer").
	TokenType string

	// ExpiresAt is when the access token expires.
	ExpiresAt time.Time

	// Scopes are the granted scopes.
	Scopes []string
}

// IsExpired returns true if the access token has expired.
func (t OAuthTokens) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// NeedsRefresh returns true if the token should be refreshed soon.
func (t OAuthTokens) NeedsRefresh() bool {
	return time.Until(t.ExpiresAt) < 5*time.Minute
}

// HasRefreshToken returns true if a refresh token is available.
func (t OAuthTokens) HasRefreshToken() bool {
	return t.RefreshToken != ""
}

// ============================================================================
// Provider Profile
// ============================================================================

// ProviderProfile contains profile data fetched from a provider.
type ProviderProfile struct {
	// Provider is the source provider.
	Provider Provider

	// ProviderUserID is the user's ID on the provider.
	ProviderUserID string

	// Username is the user's username on the provider.
	Username string

	// DisplayName is the user's display name.
	DisplayName string

	// Email is the user's email (if available).
	Email string

	// AvatarURL is the URL to the user's avatar.
	AvatarURL string

	// ProfileURL is the URL to the user's profile.
	ProfileURL string

	// RawData contains the full provider response.
	RawData map[string]any

	// FetchedAt is when the profile was fetched.
	FetchedAt time.Time
}

// NewProviderProfile creates a new provider profile.
func NewProviderProfile(provider Provider, providerUserID, username string) ProviderProfile {
	return ProviderProfile{
		Provider:       provider,
		ProviderUserID: providerUserID,
		Username:       username,
		FetchedAt:      time.Now(),
	}
}

// WithDisplayName sets the display name.
func (p ProviderProfile) WithDisplayName(name string) ProviderProfile {
	p.DisplayName = name
	return p
}

// WithEmail sets the email.
func (p ProviderProfile) WithEmail(email string) ProviderProfile {
	p.Email = email
	return p
}

// WithAvatarURL sets the avatar URL.
func (p ProviderProfile) WithAvatarURL(url string) ProviderProfile {
	p.AvatarURL = url
	return p
}

// WithProfileURL sets the profile URL.
func (p ProviderProfile) WithProfileURL(url string) ProviderProfile {
	p.ProfileURL = url
	return p
}

// WithRawData sets the raw data.
func (p ProviderProfile) WithRawData(data map[string]any) ProviderProfile {
	p.RawData = data
	return p
}
