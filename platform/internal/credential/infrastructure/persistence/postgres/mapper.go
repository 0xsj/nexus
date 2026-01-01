package postgres

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/0xsj/nexus/platform/internal/credential/application/query"
	"github.com/0xsj/nexus/platform/internal/credential/domain"
)

// ============================================================================
// Database Row
// ============================================================================

// CredentialRow represents a credential row in the database.
type CredentialRow struct {
	ID               sql.NullString
	CredentialType   sql.NullString
	SchemaID         sql.NullString
	HolderDID        sql.NullString
	IssuerDID        sql.NullString
	Status           sql.NullString
	Claims           []byte
	SignedVC         sql.NullString
	IssuedAt         sql.NullTime
	ExpiresAt        sql.NullTime
	RevokedAt        sql.NullTime
	RevokedBy        sql.NullString
	RevocationReason sql.NullString
	SuspendedAt      sql.NullTime
	SuspendedBy      sql.NullString
	SuspensionReason sql.NullString
	SuspendedUntil   sql.NullTime
	CreatedAt        sql.NullTime
	UpdatedAt        sql.NullTime
	Version          sql.NullInt32
}

// ScanFields returns the fields for scanning a row.
func (r *CredentialRow) ScanFields() []any {
	return []any{
		&r.ID,
		&r.CredentialType,
		&r.SchemaID,
		&r.HolderDID,
		&r.IssuerDID,
		&r.Status,
		&r.Claims,
		&r.SignedVC,
		&r.IssuedAt,
		&r.ExpiresAt,
		&r.RevokedAt,
		&r.RevokedBy,
		&r.RevocationReason,
		&r.SuspendedAt,
		&r.SuspendedBy,
		&r.SuspensionReason,
		&r.SuspendedUntil,
		&r.CreatedAt,
		&r.UpdatedAt,
		&r.Version,
	}
}

// ============================================================================
// Domain -> Row Mapping
// ============================================================================

// CredentialToRow converts a domain credential to database row values for insert.
func CredentialToRow(c *domain.Credential, createdAt, updatedAt time.Time) ([]any, error) {
	// Serialize claims to JSON
	claimsJSON, err := json.Marshal(c.Claims())
	if err != nil {
		return nil, err
	}

	// Extract revocation info
	var revokedAt sql.NullTime
	var revokedBy sql.NullString
	var revocationReason sql.NullString
	if info := c.RevocationInfo(); info != nil {
		revokedAt = sql.NullTime{Time: info.RevokedAt, Valid: true}
		revokedBy = sql.NullString{String: info.RevokedBy, Valid: true}
		revocationReason = sql.NullString{String: info.Reason, Valid: true}
	}

	// Extract suspension info
	var suspendedAt sql.NullTime
	var suspendedBy sql.NullString
	var suspensionReason sql.NullString
	var suspendedUntil sql.NullTime
	if info := c.SuspensionInfo(); info != nil {
		suspendedAt = sql.NullTime{Time: info.SuspendedAt, Valid: true}
		suspendedBy = sql.NullString{String: info.SuspendedBy, Valid: true}
		suspensionReason = sql.NullString{String: info.Reason, Valid: true}
		if info.Until != nil {
			suspendedUntil = sql.NullTime{Time: *info.Until, Valid: true}
		}
	}

	return []any{
		c.AggregateID(),            // $1  id
		c.CredentialType(),         // $2  credential_type
		nullString(c.SchemaID()),   // $3  schema_id
		c.HolderDID(),              // $4  holder_did
		c.IssuerDID(),              // $5  issuer_did
		c.Status().String(),        // $6  status
		claimsJSON,                 // $7  claims
		nullString(c.SignedVC()),   // $8  signed_vc
		nullTimePtr(c.IssuedAt()),  // $9  issued_at
		nullTimePtr(c.ExpiresAt()), // $10 expires_at
		revokedAt,                  // $11 revoked_at
		revokedBy,                  // $12 revoked_by
		revocationReason,           // $13 revocation_reason
		suspendedAt,                // $14 suspended_at
		suspendedBy,                // $15 suspended_by
		suspensionReason,           // $16 suspension_reason
		suspendedUntil,             // $17 suspended_until
		createdAt,                  // $18 created_at
		updatedAt,                  // $19 updated_at
		c.Version(),                // $20 version
	}, nil
}

// CredentialToUpdateRow converts a domain credential to update parameters.
func CredentialToUpdateRow(c *domain.Credential) ([]any, error) {
	// Serialize claims to JSON
	claimsJSON, err := json.Marshal(c.Claims())
	if err != nil {
		return nil, err
	}

	// Extract revocation info
	var revokedAt sql.NullTime
	var revokedBy sql.NullString
	var revocationReason sql.NullString
	if info := c.RevocationInfo(); info != nil {
		revokedAt = sql.NullTime{Time: info.RevokedAt, Valid: true}
		revokedBy = sql.NullString{String: info.RevokedBy, Valid: true}
		revocationReason = sql.NullString{String: info.Reason, Valid: true}
	}

	// Extract suspension info
	var suspendedAt sql.NullTime
	var suspendedBy sql.NullString
	var suspensionReason sql.NullString
	var suspendedUntil sql.NullTime
	if info := c.SuspensionInfo(); info != nil {
		suspendedAt = sql.NullTime{Time: info.SuspendedAt, Valid: true}
		suspendedBy = sql.NullString{String: info.SuspendedBy, Valid: true}
		suspensionReason = sql.NullString{String: info.Reason, Valid: true}
		if info.Until != nil {
			suspendedUntil = sql.NullTime{Time: *info.Until, Valid: true}
		}
	}

	return []any{
		c.AggregateID(),            // $1  id (WHERE)
		c.CredentialType(),         // $2  credential_type
		nullString(c.SchemaID()),   // $3  schema_id
		c.HolderDID(),              // $4  holder_did
		c.IssuerDID(),              // $5  issuer_did
		c.Status().String(),        // $6  status
		claimsJSON,                 // $7  claims
		nullString(c.SignedVC()),   // $8  signed_vc
		nullTimePtr(c.IssuedAt()),  // $9  issued_at
		nullTimePtr(c.ExpiresAt()), // $10 expires_at
		revokedAt,                  // $11 revoked_at
		revokedBy,                  // $12 revoked_by
		revocationReason,           // $13 revocation_reason
		suspendedAt,                // $14 suspended_at
		suspendedBy,                // $15 suspended_by
		suspensionReason,           // $16 suspension_reason
		suspendedUntil,             // $17 suspended_until
		c.Version(),                // $18 version (WHERE)
	}, nil
}

// ============================================================================
// Row -> Domain Mapping
// ============================================================================

// RowToCredential converts a database row to a domain credential.
func RowToCredential(row *CredentialRow) (*domain.Credential, error) {
	// Parse claims
	var claims map[string]any
	if len(row.Claims) > 0 {
		if err := json.Unmarshal(row.Claims, &claims); err != nil {
			return nil, err
		}
	}

	// Parse status
	status := domain.ParseCredentialStatus(row.Status.String)

	// Build revocation info if present
	var revocationInfo *domain.RevocationInfo
	if row.RevokedAt.Valid {
		revocationInfo = &domain.RevocationInfo{
			RevokedBy: stringFromNull(row.RevokedBy),
			Reason:    stringFromNull(row.RevocationReason),
			RevokedAt: row.RevokedAt.Time,
		}
	}

	// Build suspension info if present
	var suspensionInfo *domain.SuspensionInfo
	if row.SuspendedAt.Valid {
		suspensionInfo = &domain.SuspensionInfo{
			SuspendedBy: stringFromNull(row.SuspendedBy),
			Reason:      stringFromNull(row.SuspensionReason),
			SuspendedAt: row.SuspendedAt.Time,
			Until:       timePtrFromNull(row.SuspendedUntil),
		}
	}

	return domain.ReconstructCredential(
		row.ID.String,
		row.CredentialType.String,
		stringFromNull(row.SchemaID),
		row.HolderDID.String,
		row.IssuerDID.String,
		status,
		claims,
		stringFromNull(row.SignedVC),
		timePtrFromNull(row.IssuedAt),
		timePtrFromNull(row.ExpiresAt),
		revocationInfo,
		suspensionInfo,
		int(row.Version.Int32),
	), nil
}

// RowToCredentialView converts a database row to a query view.
func RowToCredentialView(row *CredentialRow) *query.CredentialView {
	// Parse claims
	var claims map[string]any
	if len(row.Claims) > 0 {
		json.Unmarshal(row.Claims, &claims)
	}

	return &query.CredentialView{
		ID:               row.ID.String,
		CredentialType:   row.CredentialType.String,
		SchemaID:         stringFromNull(row.SchemaID),
		HolderDID:        row.HolderDID.String,
		IssuerDID:        row.IssuerDID.String,
		Status:           row.Status.String,
		Claims:           claims,
		SignedVC:         stringFromNull(row.SignedVC),
		IssuedAt:         timePtrFromNull(row.IssuedAt),
		ExpiresAt:        timePtrFromNull(row.ExpiresAt),
		RevokedAt:        timePtrFromNull(row.RevokedAt),
		RevokedBy:        stringFromNull(row.RevokedBy),
		RevocationReason: stringFromNull(row.RevocationReason),
		SuspendedAt:      timePtrFromNull(row.SuspendedAt),
		SuspendedBy:      stringFromNull(row.SuspendedBy),
		SuspensionReason: stringFromNull(row.SuspensionReason),
		SuspendedUntil:   timePtrFromNull(row.SuspendedUntil),
		CreatedAt:        timeFromNull(row.CreatedAt),
		UpdatedAt:        timeFromNull(row.UpdatedAt),
		Version:          int(row.Version.Int32),
	}
}

// ============================================================================
// Null Helpers
// ============================================================================

// nullString converts a string to sql.NullString.
func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

// nullTimePtr converts a *time.Time to sql.NullTime.
func nullTimePtr(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *t, Valid: true}
}

// stringFromNull extracts a string from sql.NullString.
func stringFromNull(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

// timeFromNull extracts a time.Time from sql.NullTime.
func timeFromNull(nt sql.NullTime) time.Time {
	if nt.Valid {
		return nt.Time
	}
	return time.Time{}
}

// timePtrFromNull extracts a *time.Time from sql.NullTime.
func timePtrFromNull(nt sql.NullTime) *time.Time {
	if nt.Valid {
		return &nt.Time
	}
	return nil
}
