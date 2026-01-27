package postgres

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/0xsj/nexus/platform/internal/verification/domain"
)

// ============================================================================
// Verification Row
// ============================================================================

// verificationRow represents a database row for verifications.
type verificationRow struct {
	ID             string
	UserID         string
	Provider       string
	CredentialType string
	Status         string
	OAuthState     sql.NullString
	ProviderUserID sql.NullString
	Username       sql.NullString
	Email          sql.NullString
	DisplayName    sql.NullString
	AvatarURL      sql.NullString
	ProfileURL     sql.NullString
	RawData        []byte
	CredentialID   sql.NullString
	FailureReason  sql.NullString
	FailureCode    sql.NullString
	RedirectURL    sql.NullString
	InitiatedAt    time.Time
	AuthorizedAt   sql.NullTime
	CompletedAt    sql.NullTime
	FailedAt       sql.NullTime
	ExpiresAt      time.Time
	Version        int
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// toDomain converts a verificationRow to a domain.Verification.
func (r *verificationRow) toDomain() (*domain.Verification, error) {
	// Parse claims from raw data if present
	var claims map[string]any
	if len(r.RawData) > 0 {
		if err := json.Unmarshal(r.RawData, &claims); err != nil {
			claims = nil
		}
	}

	// Convert timestamps
	var authorizedAt *time.Time
	if r.AuthorizedAt.Valid {
		authorizedAt = &r.AuthorizedAt.Time
	}

	var completedAt *time.Time
	if r.CompletedAt.Valid {
		completedAt = &r.CompletedAt.Time
	}

	var failedAt *time.Time
	if r.FailedAt.Valid {
		failedAt = &r.FailedAt.Time
	}

	// Reconstruct the verification
	return domain.ReconstructVerification(
		r.ID,
		r.UserID,
		domain.Provider(r.Provider),
		domain.CredentialType(r.CredentialType),
		r.OAuthState.String,
		r.RedirectURL.String,
		domain.VerificationStatus(r.Status),
		r.ExpiresAt,
		r.ProviderUserID.String,
		r.Username.String,
		r.CredentialID.String,
		claims,
		r.FailureReason.String,
		r.FailureCode.String,
		r.InitiatedAt,
		authorizedAt,
		completedAt,
		failedAt,
		r.Version,
	), nil
}

// fromVerification converts a domain.Verification to row parameters.
func fromVerification(v *domain.Verification) *verificationRow {
	row := &verificationRow{
		ID:             v.AggregateID(),
		UserID:         v.UserID(),
		Provider:       string(v.Provider()),
		CredentialType: string(v.CredentialType()),
		Status:         string(v.Status()),
		OAuthState:     toNullString(v.OAuthState()),
		ProviderUserID: toNullString(v.ProviderUserID()),
		Username:       toNullString(v.Username()),
		CredentialID:   toNullString(v.CredentialID()),
		FailureReason:  toNullString(v.FailureReason()),
		FailureCode:    toNullString(v.FailureCode()),
		RedirectURL:    toNullString(v.RedirectURL()),
		InitiatedAt:    v.InitiatedAt(),
		ExpiresAt:      v.ExpiresAt(),
		Version:        v.Version(),
	}

	// Profile data
	if profile := v.Profile(); profile != nil {
		row.Email = toNullString(profile.Email)
		row.DisplayName = toNullString(profile.DisplayName)
		row.AvatarURL = toNullString(profile.AvatarURL)
		row.ProfileURL = toNullString(profile.ProfileURL)
		if profile.RawData != nil {
			row.RawData, _ = json.Marshal(profile.RawData)
		}
	}

	// Claims as raw data (if no profile)
	if row.RawData == nil && v.Claims() != nil {
		row.RawData, _ = json.Marshal(v.Claims())
	}

	// Timestamps
	if at := v.AuthorizedAt(); at != nil {
		row.AuthorizedAt = sql.NullTime{Time: *at, Valid: true}
	}
	if ct := v.CompletedAt(); ct != nil {
		row.CompletedAt = sql.NullTime{Time: *ct, Valid: true}
	}
	if ft := v.FailedAt(); ft != nil {
		row.FailedAt = sql.NullTime{Time: *ft, Valid: true}
	}

	return row
}

// ============================================================================
// OAuth State Row
// ============================================================================

// oauthStateRow represents a database row for oauth_states.
type oauthStateRow struct {
	State          string
	VerificationID string
	UserID         string
	Provider       string
	CodeVerifier   sql.NullString
	Nonce          sql.NullString
	RedirectURL    sql.NullString
	CreatedAt      time.Time
	ExpiresAt      time.Time
	UsedAt         sql.NullTime
}

// toDomain converts an oauthStateRow to a domain.OAuthState.
func (r *oauthStateRow) toDomain() *domain.OAuthState {
	return &domain.OAuthState{
		Value:       r.State,
		UserID:      r.UserID,
		Provider:    domain.Provider(r.Provider),
		RedirectURL: r.RedirectURL.String,
		CreatedAt:   r.CreatedAt,
		ExpiresAt:   r.ExpiresAt,
	}
}

// fromOAuthState converts a domain.OAuthState to row parameters.
func fromOAuthState(verificationID string, s *domain.OAuthState, codeVerifier, nonce *string) *oauthStateRow {
	row := &oauthStateRow{
		State:          s.Value,
		VerificationID: verificationID,
		UserID:         s.UserID,
		Provider:       string(s.Provider),
		RedirectURL:    toNullString(s.RedirectURL),
		CreatedAt:      s.CreatedAt,
		ExpiresAt:      s.ExpiresAt,
	}

	if codeVerifier != nil {
		row.CodeVerifier = toNullString(*codeVerifier)
	}
	if nonce != nil {
		row.Nonce = toNullString(*nonce)
	}

	return row
}

// ============================================================================
// Provider Tokens Row
// ============================================================================

// providerTokensRow represents a database row for provider_tokens.
type providerTokensRow struct {
	UserID         string
	Provider       string
	AccessToken    string
	RefreshToken   sql.NullString
	TokenType      string
	Scopes         []string
	ExpiresAt      sql.NullTime
	ProviderUserID sql.NullString
	CreatedAt      time.Time
	UpdatedAt      time.Time
	LastUsedAt     sql.NullTime
}

// toDomain converts a providerTokensRow to a domain.OAuthTokens.
func (r *providerTokensRow) toDomain() *domain.OAuthTokens {
	var expiresAt time.Time
	if r.ExpiresAt.Valid {
		expiresAt = r.ExpiresAt.Time
	}

	return &domain.OAuthTokens{
		AccessToken:  r.AccessToken,
		RefreshToken: r.RefreshToken.String,
		TokenType:    r.TokenType,
		Scopes:       r.Scopes,
		ExpiresAt:    expiresAt,
	}
}

// fromOAuthTokens converts domain.OAuthTokens to row parameters.
func fromOAuthTokens(userID string, provider domain.Provider, tokens *domain.OAuthTokens, providerUserID string) *providerTokensRow {
	row := &providerTokensRow{
		UserID:       userID,
		Provider:     string(provider),
		AccessToken:  tokens.AccessToken,
		RefreshToken: toNullString(tokens.RefreshToken),
		TokenType:    tokens.TokenType,
		Scopes:       tokens.Scopes,
	}

	if !tokens.ExpiresAt.IsZero() {
		row.ExpiresAt = sql.NullTime{Time: tokens.ExpiresAt, Valid: true}
	}

	if providerUserID != "" {
		row.ProviderUserID = toNullString(providerUserID)
	}

	return row
}

// ============================================================================
// Read Model Rows
// ============================================================================

// verificationViewRow represents a row for the VerificationView read model.
type verificationViewRow struct {
	ID             string
	UserID         string
	Provider       string
	CredentialType string
	Status         string
	ProviderUserID sql.NullString
	Username       sql.NullString
	CredentialID   sql.NullString
	FailureReason  sql.NullString
	FailureCode    sql.NullString
	InitiatedAt    time.Time
	AuthorizedAt   sql.NullTime
	CompletedAt    sql.NullTime
	FailedAt       sql.NullTime
	ExpiresAt      time.Time
}

// toView converts a verificationViewRow to a domain.VerificationView.
func (r *verificationViewRow) toView() *domain.VerificationView {
	return &domain.VerificationView{
		ID:             r.ID,
		UserID:         r.UserID,
		Provider:       r.Provider,
		ProviderName:   domain.Provider(r.Provider).DisplayName(),
		CredentialType: r.CredentialType,
		Status:         r.Status,
		ProviderUserID: r.ProviderUserID.String,
		Username:       r.Username.String,
		CredentialID:   r.CredentialID.String,
		FailureReason:  r.FailureReason.String,
		FailureCode:    r.FailureCode.String,
		InitiatedAt:    r.InitiatedAt,
		AuthorizedAt:   nullTimeToPtr(r.AuthorizedAt),
		CompletedAt:    nullTimeToPtr(r.CompletedAt),
		FailedAt:       nullTimeToPtr(r.FailedAt),
		ExpiresAt:      r.ExpiresAt,
	}
}

// verificationSummaryRow represents a row for the VerificationSummary read model.
type verificationSummaryRow struct {
	ID          string
	Provider    string
	Status      string
	InitiatedAt time.Time
	CompletedAt sql.NullTime
}

// toSummary converts a verificationSummaryRow to a domain.VerificationSummary.
func (r *verificationSummaryRow) toSummary() *domain.VerificationSummary {
	return &domain.VerificationSummary{
		ID:           r.ID,
		Provider:     r.Provider,
		ProviderName: domain.Provider(r.Provider).DisplayName(),
		Status:       r.Status,
		InitiatedAt:  r.InitiatedAt,
		CompletedAt:  nullTimeToPtr(r.CompletedAt),
	}
}

// providerConnectionRow represents a row for provider connection status.
type providerConnectionRow struct {
	Provider       string
	Connected      bool
	ProviderUserID sql.NullString
	Username       sql.NullString
	CredentialID   sql.NullString
	VerifiedAt     sql.NullTime
}

// toView converts a providerConnectionRow to a domain.ProviderConnectionView.
func (r *providerConnectionRow) toView() *domain.ProviderConnectionView {
	return &domain.ProviderConnectionView{
		Provider:       r.Provider,
		ProviderName:   domain.Provider(r.Provider).DisplayName(),
		Connected:      r.Connected,
		ProviderUserID: r.ProviderUserID.String,
		Username:       r.Username.String,
		CredentialID:   r.CredentialID.String,
		VerifiedAt:     nullTimeToPtr(r.VerifiedAt),
	}
}

// ============================================================================
// Helpers
// ============================================================================

// toNullString converts string to sql.NullString.
func toNullString(s string) sql.NullString {
	if s != "" {
		return sql.NullString{String: s, Valid: true}
	}
	return sql.NullString{}
}

// nullTimeToPtr converts sql.NullTime to *time.Time.
func nullTimeToPtr(nt sql.NullTime) *time.Time {
	if nt.Valid {
		return &nt.Time
	}
	return nil
}
