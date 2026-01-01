package postgres

// ============================================================================
// SQL Queries
// ============================================================================

const (
	// -------------------------------------------------------------------------
	// Insert
	// -------------------------------------------------------------------------

	queryInsert = `
		INSERT INTO credentials (
			id,
			credential_type,
			schema_id,
			holder_did,
			issuer_did,
			status,
			claims,
			signed_vc,
			issued_at,
			expires_at,
			revoked_at,
			revoked_by,
			revocation_reason,
			suspended_at,
			suspended_by,
			suspension_reason,
			suspended_until,
			created_at,
			updated_at,
			version
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15, $16, $17, $18, $19, $20
		)`

	// -------------------------------------------------------------------------
	// Update
	// -------------------------------------------------------------------------

	queryUpdate = `
		UPDATE credentials SET
			credential_type = $2,
			schema_id = $3,
			holder_did = $4,
			issuer_did = $5,
			status = $6,
			claims = $7,
			signed_vc = $8,
			issued_at = $9,
			expires_at = $10,
			revoked_at = $11,
			revoked_by = $12,
			revocation_reason = $13,
			suspended_at = $14,
			suspended_by = $15,
			suspension_reason = $16,
			suspended_until = $17,
			version = version + 1
		WHERE id = $1 AND version = $18
		RETURNING version`

	// -------------------------------------------------------------------------
	// Select
	// -------------------------------------------------------------------------

	querySelectByID = `
		SELECT
			id,
			credential_type,
			schema_id,
			holder_did,
			issuer_did,
			status,
			claims,
			signed_vc,
			issued_at,
			expires_at,
			revoked_at,
			revoked_by,
			revocation_reason,
			suspended_at,
			suspended_by,
			suspension_reason,
			suspended_until,
			created_at,
			updated_at,
			version
		FROM credentials
		WHERE id = $1`

	querySelectBase = `
		SELECT
			id,
			credential_type,
			schema_id,
			holder_did,
			issuer_did,
			status,
			claims,
			signed_vc,
			issued_at,
			expires_at,
			revoked_at,
			revoked_by,
			revocation_reason,
			suspended_at,
			suspended_by,
			suspension_reason,
			suspended_until,
			created_at,
			updated_at,
			version
		FROM credentials`

	queryCountBase = `SELECT COUNT(*) FROM credentials`

	// -------------------------------------------------------------------------
	// Delete
	// -------------------------------------------------------------------------

	queryDelete = `DELETE FROM credentials WHERE id = $1`

	// -------------------------------------------------------------------------
	// Existence
	// -------------------------------------------------------------------------

	queryExists = `SELECT EXISTS(SELECT 1 FROM credentials WHERE id = $1)`
)

// ============================================================================
// Query Builder Helpers
// ============================================================================

// OrderByColumn maps sort field names to SQL column names.
var OrderByColumn = map[string]string{
	"created_at":      "created_at",
	"updated_at":      "updated_at",
	"issued_at":       "issued_at",
	"credential_type": "credential_type",
	"status":          "status",
}

// ValidSortOrder validates sort order.
func ValidSortOrder(order string) string {
	if order == "asc" || order == "ASC" {
		return "ASC"
	}
	return "DESC"
}
