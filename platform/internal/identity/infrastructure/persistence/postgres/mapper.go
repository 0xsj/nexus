package postgres

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/0xsj/nexus/platform/internal/identity/domain"
	"github.com/0xsj/nexus/platform/pkg/did"
	"github.com/lib/pq"
)

// ============================================================================
// User Row
// ============================================================================

// UserRow represents a user database row.
type UserRow struct {
	ID              string         `db:"id"`
	DID             string         `db:"did"`
	Status          string         `db:"status"`
	DisplayName     sql.NullString `db:"display_name"`
	AvatarURL       sql.NullString `db:"avatar_url"`
	Bio             sql.NullString `db:"bio"`
	LastLoginMethod sql.NullString `db:"last_login_method"`
	CreatedAt       time.Time      `db:"created_at"`
	UpdatedAt       time.Time      `db:"updated_at"`
	LastLoginAt     sql.NullTime   `db:"last_login_at"`
}

// ToUserRow converts a domain User to a database row.
func ToUserRow(u *domain.User) *UserRow {
	row := &UserRow{
		ID:        u.ID(),
		DID:       u.PrimaryDID().String(),
		Status:    u.Status().String(),
		CreatedAt: u.CreatedAt(),
		UpdatedAt: u.UpdatedAt(),
	}

	if !u.LastLoginAt().IsZero() {
		row.LastLoginAt = sql.NullTime{Time: u.LastLoginAt(), Valid: true}
	}

	if u.LastLoginMethod() != "" {
		row.LastLoginMethod = sql.NullString{String: u.LastLoginMethod().String(), Valid: true}
	}

	return row
}

// ToDomainUser converts a database row to a domain User.
// Requires linked DIDs, identities, and wallets to be loaded separately.
func (r *UserRow) ToDomainUser(
	linkedDIDs domain.LinkedDIDs,
	linkedIdentities []domain.LinkedIdentity,
	wallets []domain.WalletAddress,
) (*domain.User, error) {
	primaryDID, err := did.Parse(r.DID)
	if err != nil {
		return nil, err
	}

	var lastLoginAt time.Time
	if r.LastLoginAt.Valid {
		lastLoginAt = r.LastLoginAt.Time
	}

	var lastLoginMethod domain.AuthMethod
	if r.LastLoginMethod.Valid {
		lastLoginMethod = domain.AuthMethod(r.LastLoginMethod.String)
	}

	return domain.Reconstitute(
		r.ID,
		0, // version - loaded separately if using event sourcing
		primaryDID,
		linkedDIDs,
		domain.UserStatus(r.Status),
		linkedIdentities,
		wallets,
		r.CreatedAt,
		r.UpdatedAt,
		lastLoginAt,
		lastLoginMethod,
	), nil
}

// ============================================================================
// Linked DID Row
// ============================================================================

// LinkedDIDRow represents a linked DID database row.
type LinkedDIDRow struct {
	ID         string         `db:"id"`
	UserID     string         `db:"user_id"`
	DID        string         `db:"did"`
	Source     string         `db:"source"`
	IsPrimary  bool           `db:"is_primary"`
	Label      sql.NullString `db:"label"`
	Metadata   []byte         `db:"metadata"`
	LinkedAt   time.Time      `db:"linked_at"`
	LastUsedAt sql.NullTime   `db:"last_used_at"`
}

// ToLinkedDIDRow converts a domain LinkedDID to a database row.
func ToLinkedDIDRow(userID string, ld domain.LinkedDID) (*LinkedDIDRow, error) {
	row := &LinkedDIDRow{
		ID:        ld.ID,
		UserID:    userID,
		DID:       ld.DID.String(),
		Source:    ld.Source.String(),
		IsPrimary: ld.IsPrimary,
		LinkedAt:  ld.LinkedAt,
	}

	if ld.Label != "" {
		row.Label = sql.NullString{String: ld.Label, Valid: true}
	}

	if len(ld.Metadata) > 0 {
		data, err := json.Marshal(ld.Metadata)
		if err != nil {
			return nil, err
		}
		row.Metadata = data
	}

	if ld.LastUsedAt != nil {
		row.LastUsedAt = sql.NullTime{Time: *ld.LastUsedAt, Valid: true}
	}

	return row, nil
}

// ToDomainLinkedDID converts a database row to a domain LinkedDID.
func (r *LinkedDIDRow) ToDomainLinkedDID() (domain.LinkedDID, error) {
	parsedDID, err := did.Parse(r.DID)
	if err != nil {
		return domain.LinkedDID{}, err
	}

	ld := domain.LinkedDID{
		ID:        r.ID,
		DID:       parsedDID,
		Source:    domain.DIDSource(r.Source),
		IsPrimary: r.IsPrimary,
		LinkedAt:  r.LinkedAt,
	}

	if r.Label.Valid {
		ld.Label = r.Label.String
	}

	if len(r.Metadata) > 0 {
		var metadata map[string]string
		if err := json.Unmarshal(r.Metadata, &metadata); err == nil {
			ld.Metadata = metadata
		}
	}

	if r.LastUsedAt.Valid {
		ld.LastUsedAt = &r.LastUsedAt.Time
	}

	return ld, nil
}

// ToLinkedDIDRows converts a slice of domain LinkedDIDs to database rows.
func ToLinkedDIDRows(userID string, linkedDIDs domain.LinkedDIDs) ([]*LinkedDIDRow, error) {
	rows := make([]*LinkedDIDRow, 0, len(linkedDIDs))
	for _, ld := range linkedDIDs {
		row, err := ToLinkedDIDRow(userID, ld)
		if err != nil {
			return nil, err
		}
		rows = append(rows, row)
	}
	return rows, nil
}

// ToDomainLinkedDIDs converts a slice of database rows to domain LinkedDIDs.
func ToDomainLinkedDIDs(rows []*LinkedDIDRow) (domain.LinkedDIDs, error) {
	linkedDIDs := make(domain.LinkedDIDs, 0, len(rows))
	for _, row := range rows {
		ld, err := row.ToDomainLinkedDID()
		if err != nil {
			return nil, err
		}
		linkedDIDs = append(linkedDIDs, ld)
	}
	return linkedDIDs, nil
}

// ============================================================================
// User Wallet Row
// ============================================================================

// UserWalletRow represents a user wallet database row.
type UserWalletRow struct {
	ID         string       `db:"id"`
	UserID     string       `db:"user_id"`
	Address    string       `db:"address"`
	Chain      string       `db:"chain"`
	IsPrimary  bool         `db:"is_primary"`
	Verified   bool         `db:"verified"`
	VerifiedAt sql.NullTime `db:"verified_at"`
	LinkedAt   time.Time    `db:"linked_at"`
}

// ToDomainWalletAddress converts a database row to domain WalletAddress.
func (r *UserWalletRow) ToDomainWalletAddress() domain.WalletAddress {
	return domain.NewWalletAddress(r.Address, domain.Chain(r.Chain))
}

// ============================================================================
// User Email Row
// ============================================================================

// UserEmailRow represents a user email database row.
type UserEmailRow struct {
	ID         string       `db:"id"`
	UserID     string       `db:"user_id"`
	Email      string       `db:"email"`
	IsPrimary  bool         `db:"is_primary"`
	Verified   bool         `db:"verified"`
	VerifiedAt sql.NullTime `db:"verified_at"`
	LinkedAt   time.Time    `db:"linked_at"`
}

// ToDomainLinkedIdentity converts a database row to domain LinkedIdentity.
func (r *UserEmailRow) ToDomainLinkedIdentity() domain.LinkedIdentity {
	var verifiedAt time.Time
	if r.VerifiedAt.Valid {
		verifiedAt = r.VerifiedAt.Time
	}

	return domain.LinkedIdentity{
		Type:       domain.IdentityTypeEmail,
		Value:      r.Email,
		Verified:   r.Verified,
		VerifiedAt: verifiedAt,
		LinkedAt:   r.LinkedAt,
	}
}

// ============================================================================
// Session Row
// ============================================================================

// SessionRow represents a session database row.
type SessionRow struct {
	ID           string         `db:"id"`
	UserID       string         `db:"user_id"`
	TokenHash    string         `db:"token_hash"`
	AuthMethod   string         `db:"auth_method"`
	IPAddress    sql.NullString `db:"ip_address"`
	UserAgent    sql.NullString `db:"user_agent"`
	DeviceID     sql.NullString `db:"device_id"`
	Status       string         `db:"status"`
	CreatedAt    time.Time      `db:"created_at"`
	ExpiresAt    time.Time      `db:"expires_at"`
	LastActiveAt time.Time      `db:"last_active_at"`
	RevokedAt    sql.NullTime   `db:"revoked_at"`
}

// ToSessionRow converts a domain Session to a database row.
func ToSessionRow(s *domain.Session) *SessionRow {
	row := &SessionRow{
		ID:           s.ID(),
		UserID:       s.UserID(),
		TokenHash:    s.TokenHash(),
		AuthMethod:   s.Method().String(),
		Status:       s.Status().String(),
		CreatedAt:    s.CreatedAt(),
		ExpiresAt:    s.ExpiresAt(),
		LastActiveAt: s.LastSeenAt(),
	}

	if s.IPAddress() != "" {
		row.IPAddress = sql.NullString{String: s.IPAddress(), Valid: true}
	}
	if s.UserAgent() != "" {
		row.UserAgent = sql.NullString{String: s.UserAgent(), Valid: true}
	}
	if s.Device() != "" {
		row.DeviceID = sql.NullString{String: s.Device(), Valid: true}
	}
	if !s.RevokedAt().IsZero() {
		row.RevokedAt = sql.NullTime{Time: s.RevokedAt(), Valid: true}
	}

	return row
}

// ToDomainSession converts a database row to a domain Session.
func (r *SessionRow) ToDomainSession() *domain.Session {
	var userAgent, ipAddress, device string
	var revokedAt time.Time

	if r.UserAgent.Valid {
		userAgent = r.UserAgent.String
	}
	if r.IPAddress.Valid {
		ipAddress = r.IPAddress.String
	}
	if r.DeviceID.Valid {
		device = r.DeviceID.String
	}
	if r.RevokedAt.Valid {
		revokedAt = r.RevokedAt.Time
	}

	return domain.ReconstituteSession(
		r.ID,
		r.UserID,
		domain.SessionStatus(r.Status),
		domain.AuthMethod(r.AuthMethod),
		r.TokenHash,
		userAgent,
		ipAddress,
		device,
		r.CreatedAt,
		r.ExpiresAt,
		revokedAt,
		r.LastActiveAt,
	)
}

// ============================================================================
// Connection Row
// ============================================================================

// ConnectionRow represents a connection database row.
type ConnectionRow struct {
	ID               string         `db:"id"`
	UserID           string         `db:"user_id"`
	Provider         string         `db:"provider"`
	ProviderUserID   string         `db:"provider_user_id"`
	ProviderUsername sql.NullString `db:"provider_username"`
	ProviderEmail    sql.NullString `db:"provider_email"`
	ProviderAvatar   sql.NullString `db:"provider_avatar"`
	ProfileData      []byte         `db:"profile_data"`
	AccessToken      sql.NullString `db:"access_token"`
	RefreshToken     sql.NullString `db:"refresh_token"`
	TokenExpiresAt   sql.NullTime   `db:"token_expires_at"`
	Scopes           pq.StringArray `db:"scopes"`
	Status           string         `db:"status"`
	LastSyncedAt     sql.NullTime   `db:"last_synced_at"`
	SyncError        sql.NullString `db:"sync_error"`
	ConnectedAt      time.Time      `db:"connected_at"`
	UpdatedAt        time.Time      `db:"updated_at"`
	DisconnectedAt   sql.NullTime   `db:"disconnected_at"`
}

// ToConnectionRow converts a domain Connection to a database row.
func ToConnectionRow(c *domain.Connection) (*ConnectionRow, error) {
	row := &ConnectionRow{
		ID:             c.ID(),
		UserID:         c.UserID(),
		Provider:       c.Provider().String(),
		ProviderUserID: c.ProviderUserID(),
		Status:         c.Status().String(),
		ConnectedAt:    c.ConnectedAt(),
		UpdatedAt:      c.UpdatedAt(),
	}

	if c.Username() != "" {
		row.ProviderUsername = sql.NullString{String: c.Username(), Valid: true}
	}
	if c.Email() != "" {
		row.ProviderEmail = sql.NullString{String: c.Email(), Valid: true}
	}
	if c.AvatarURL() != "" {
		row.ProviderAvatar = sql.NullString{String: c.AvatarURL(), Valid: true}
	}
	if metadata := c.Metadata(); len(metadata) > 0 {
		data, err := json.Marshal(metadata)
		if err != nil {
			return nil, err
		}
		row.ProfileData = data
	}
	if !c.LastSyncedAt().IsZero() {
		row.LastSyncedAt = sql.NullTime{Time: c.LastSyncedAt(), Valid: true}
	}

	return row, nil
}

// ToDomainConnection converts a database row to a domain Connection.
func (r *ConnectionRow) ToDomainConnection() *domain.Connection {
	var displayName, email, username, avatarURL, profileURL string
	var metadata map[string]any
	var lastSyncedAt time.Time

	if r.ProviderUsername.Valid {
		username = r.ProviderUsername.String
	}
	if r.ProviderEmail.Valid {
		email = r.ProviderEmail.String
	}
	if r.ProviderAvatar.Valid {
		avatarURL = r.ProviderAvatar.String
	}
	if len(r.ProfileData) > 0 {
		_ = json.Unmarshal(r.ProfileData, &metadata)
	}
	if r.LastSyncedAt.Valid {
		lastSyncedAt = r.LastSyncedAt.Time
	}

	return domain.ReconstituteConnection(
		r.ID,
		r.UserID,
		domain.OAuthProvider(r.Provider),
		r.ProviderUserID,
		domain.ConnectionStatus(r.Status),
		displayName,
		email,
		username,
		avatarURL,
		profileURL,
		metadata,
		r.ConnectedAt,
		r.UpdatedAt,
		lastSyncedAt,
	)
}

// ============================================================================
// API Key Row
// ============================================================================

// APIKeyRow represents an API key database row.
type APIKeyRow struct {
	ID         string         `db:"id"`
	UserID     string         `db:"user_id"`
	Name       string         `db:"name"`
	KeyHash    string         `db:"key_hash"`
	KeyPrefix  string         `db:"key_prefix"`
	Scopes     pq.StringArray `db:"scopes"`
	RateLimit  sql.NullInt32  `db:"rate_limit"`
	Status     string         `db:"status"`
	LastUsedAt sql.NullTime   `db:"last_used_at"`
	UsageCount int64          `db:"usage_count"`
	CreatedAt  time.Time      `db:"created_at"`
	ExpiresAt  sql.NullTime   `db:"expires_at"`
	RevokedAt  sql.NullTime   `db:"revoked_at"`
}

// ToAPIKeyRow converts a domain APIKey to a database row.
func ToAPIKeyRow(k *domain.APIKey) *APIKeyRow {
	row := &APIKeyRow{
		ID:         k.ID(),
		UserID:     k.UserID(),
		Name:       k.Name(),
		KeyHash:    k.KeyHash(),
		KeyPrefix:  k.KeyPrefix(),
		Status:     k.Status().String(),
		UsageCount: k.UsageCount(),
		CreatedAt:  k.CreatedAt(),
	}

	scopes := k.Scopes()
	if len(scopes) > 0 {
		row.Scopes = scopes.Strings()
	}
	if !k.LastUsedAt().IsZero() {
		row.LastUsedAt = sql.NullTime{Time: k.LastUsedAt(), Valid: true}
	}
	if !k.ExpiresAt().IsZero() {
		row.ExpiresAt = sql.NullTime{Time: k.ExpiresAt(), Valid: true}
	}
	if !k.RevokedAt().IsZero() {
		row.RevokedAt = sql.NullTime{Time: k.RevokedAt(), Valid: true}
	}

	return row
}

// ToDomainAPIKey converts a database row to a domain APIKey.
func (r *APIKeyRow) ToDomainAPIKey() *domain.APIKey {
	var expiresAt, revokedAt, lastUsedAt time.Time

	if r.ExpiresAt.Valid {
		expiresAt = r.ExpiresAt.Time
	}
	if r.RevokedAt.Valid {
		revokedAt = r.RevokedAt.Time
	}
	if r.LastUsedAt.Valid {
		lastUsedAt = r.LastUsedAt.Time
	}

	scopes := domain.ParseAPIKeyScopes(r.Scopes)

	return domain.ReconstituteAPIKey(
		r.ID,
		r.UserID,
		r.Name,
		r.KeyHash,
		r.KeyPrefix,
		scopes,
		domain.APIKeyStatus(r.Status),
		expiresAt,
		revokedAt,
		"", // revokedBy - not stored in this schema
		"", // revokeReason - not stored in this schema
		lastUsedAt,
		r.UsageCount,
		"", // description - not stored in this schema
		r.CreatedAt,
		r.CreatedAt, // updatedAt - using createdAt as fallback
	)
}

// ============================================================================
// Refresh Token Row
// ============================================================================

// RefreshTokenRow represents a refresh token database row.
type RefreshTokenRow struct {
	ID        string       `db:"id"`
	UserID    string       `db:"user_id"`
	SessionID string       `db:"session_id"`
	TokenHash string       `db:"token_hash"`
	Revoked   bool         `db:"revoked"`
	CreatedAt time.Time    `db:"created_at"`
	ExpiresAt time.Time    `db:"expires_at"`
	RevokedAt sql.NullTime `db:"revoked_at"`
}

// ToRefreshTokenRow converts a domain RefreshToken to a database row.
func ToRefreshTokenRow(t *domain.RefreshToken) *RefreshTokenRow {
	row := &RefreshTokenRow{
		ID:        t.ID,
		UserID:    t.UserID,
		SessionID: t.SessionID,
		TokenHash: t.TokenHash,
		Revoked:   t.Revoked,
		CreatedAt: t.CreatedAt,
		ExpiresAt: t.ExpiresAt,
	}

	if t.Revoked && !t.RevokedAt.IsZero() {
		row.RevokedAt = sql.NullTime{Time: t.RevokedAt, Valid: true}
	}

	return row
}

// ToDomainRefreshToken converts a database row to a domain RefreshToken.
func (r *RefreshTokenRow) ToDomainRefreshToken() *domain.RefreshToken {
	t := &domain.RefreshToken{
		ID:        r.ID,
		UserID:    r.UserID,
		SessionID: r.SessionID,
		TokenHash: r.TokenHash,
		Revoked:   r.Revoked,
		CreatedAt: r.CreatedAt,
		ExpiresAt: r.ExpiresAt,
	}

	if r.RevokedAt.Valid {
		t.RevokedAt = r.RevokedAt.Time
	}

	return t
}
