package postgres

import (
	"context"
	"fmt"

	"github.com/0xsj/nexus/platform/internal/verification/domain"
	"github.com/0xsj/nexus/platform/pkg/database"
	pgadapter "github.com/0xsj/nexus/platform/pkg/database/postgres"
)

// ============================================================================
// Verification Reader
// ============================================================================

// VerificationReader implements domain.VerificationReadRepository using PostgreSQL.
type VerificationReader struct {
	adapter *pgadapter.BaseAdapter
}

// NewVerificationReader creates a new VerificationReader.
func NewVerificationReader(db *pgadapter.DB) *VerificationReader {
	return &VerificationReader{
		adapter: pgadapter.NewBaseAdapter(db),
	}
}

// ============================================================================
// Single Record Queries
// ============================================================================

// GetByID retrieves a verification view by ID.
func (r *VerificationReader) GetByID(ctx context.Context, id string) (*domain.VerificationView, error) {
	row := &verificationViewRow{}
	dbRow := r.adapter.Executor().QueryRow(ctx, queryGetVerificationViewByID, id)
	err := dbRow.Scan(
		&row.ID,
		&row.UserID,
		&row.Provider,
		&row.CredentialType,
		&row.Status,
		&row.ProviderUserID,
		&row.Username,
		&row.CredentialID,
		&row.FailureReason,
		&row.FailureCode,
		&row.InitiatedAt,
		&row.AuthorizedAt,
		&row.CompletedAt,
		&row.FailedAt,
		&row.ExpiresAt,
	)
	if err != nil {
		if isNoRows(err) {
			return nil, domain.ErrVerificationNotFound("VerificationReader.GetByID", id)
		}
		return nil, fmt.Errorf("failed to get verification: %w", err)
	}

	return row.toView(), nil
}

// GetLatestByUserAndProvider retrieves the most recent verification.
func (r *VerificationReader) GetLatestByUserAndProvider(ctx context.Context, userID string, provider domain.Provider) (*domain.VerificationView, error) {
	row := &verificationViewRow{}
	dbRow := r.adapter.Executor().QueryRow(ctx, queryGetLatestVerificationByUserAndProvider, userID, string(provider))
	err := dbRow.Scan(
		&row.ID,
		&row.UserID,
		&row.Provider,
		&row.CredentialType,
		&row.Status,
		&row.ProviderUserID,
		&row.Username,
		&row.CredentialID,
		&row.FailureReason,
		&row.FailureCode,
		&row.InitiatedAt,
		&row.AuthorizedAt,
		&row.CompletedAt,
		&row.FailedAt,
		&row.ExpiresAt,
	)
	if err != nil {
		if isNoRows(err) {
			return nil, domain.ErrVerificationNotFound("VerificationReader.GetLatestByUserAndProvider", userID)
		}
		return nil, fmt.Errorf("failed to get latest verification: %w", err)
	}

	return row.toView(), nil
}

// ============================================================================
// List Queries
// ============================================================================

// GetByUser retrieves all verifications for a user with pagination.
func (r *VerificationReader) GetByUser(ctx context.Context, userID string, opts domain.VerificationListOptions) ([]*domain.VerificationSummary, int, error) {
	// Get total count
	var total int
	countRow := r.adapter.Executor().QueryRow(ctx, queryCountVerificationsByUser, userID)
	if err := countRow.Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count verifications: %w", err)
	}

	// Get paginated results
	rows, err := r.adapter.Select(ctx, queryGetVerificationsByUser, userID, opts.Limit, opts.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get verifications: %w", err)
	}
	defer rows.Close()

	summaries, err := r.scanVerificationSummaries(rows)
	if err != nil {
		return nil, 0, err
	}

	return summaries, total, nil
}

// GetByUserAndProvider retrieves verifications for a user and provider.
func (r *VerificationReader) GetByUserAndProvider(ctx context.Context, userID string, provider domain.Provider) ([]*domain.VerificationView, error) {
	rows, err := r.adapter.Select(ctx, queryGetVerificationsByUserAndProvider, userID, string(provider))
	if err != nil {
		return nil, fmt.Errorf("failed to get verifications: %w", err)
	}
	defer rows.Close()

	return r.scanVerificationViews(rows)
}

// List retrieves verifications with filters and pagination.
func (r *VerificationReader) List(ctx context.Context, filter domain.VerificationFilter, opts domain.VerificationListOptions) ([]*domain.VerificationView, int, error) {
	// Build dynamic query with filters
	query, args := r.buildListQuery(filter, opts)
	countQuery, countArgs := r.buildCountQuery(filter)

	// Get total count
	var total int
	countRow := r.adapter.Executor().QueryRow(ctx, countQuery, countArgs...)
	if err := countRow.Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count verifications: %w", err)
	}

	// Get paginated results
	rows, err := r.adapter.Select(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list verifications: %w", err)
	}
	defer rows.Close()

	views, err := r.scanVerificationViews(rows)
	if err != nil {
		return nil, 0, err
	}

	return views, total, nil
}

// ============================================================================
// Provider Connections
// ============================================================================

// GetProviderConnections retrieves all provider connection statuses for a user.
func (r *VerificationReader) GetProviderConnections(ctx context.Context, userID string) ([]*domain.ProviderConnectionView, error) {
	rows, err := r.adapter.Select(ctx, queryGetProviderConnections, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider connections: %w", err)
	}
	defer rows.Close()

	var connections []*domain.ProviderConnectionView
	for rows.Next() {
		row := &providerConnectionRow{}
		err := rows.Scan(
			&row.Provider,
			&row.Connected,
			&row.ProviderUserID,
			&row.Username,
			&row.CredentialID,
			&row.VerifiedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan provider connection: %w", err)
		}
		connections = append(connections, row.toView())
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating provider connections: %w", err)
	}

	return connections, nil
}

// CheckProviderConnected checks if a provider is connected for a user.
func (r *VerificationReader) CheckProviderConnected(ctx context.Context, userID string, provider domain.Provider) (bool, *string, *string, error) {
	var connected bool
	var credentialID, username *string

	dbRow := r.adapter.Executor().QueryRow(ctx, queryCheckProviderConnected, userID, string(provider))
	err := dbRow.Scan(&connected, &credentialID, &username)
	if err != nil {
		if isNoRows(err) {
			return false, nil, nil, nil
		}
		return false, nil, nil, fmt.Errorf("failed to check provider connected: %w", err)
	}

	return connected, credentialID, username, nil
}

// ============================================================================
// Count Queries
// ============================================================================

// CountByUser returns the count of verifications for a user.
func (r *VerificationReader) CountByUser(ctx context.Context, userID string) (int, error) {
	var count int
	dbRow := r.adapter.Executor().QueryRow(ctx, queryCountVerificationsByUser, userID)
	if err := dbRow.Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count verifications: %w", err)
	}
	return count, nil
}

// CountByStatus returns the count of verifications by status.
func (r *VerificationReader) CountByStatus(ctx context.Context, status domain.VerificationStatus) (int, error) {
	var count int
	dbRow := r.adapter.Executor().QueryRow(ctx, queryCountVerificationsByStatus, string(status))
	if err := dbRow.Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count verifications by status: %w", err)
	}
	return count, nil
}

// ============================================================================
// Helper Methods
// ============================================================================

// scanVerificationSummaries scans rows into verification summaries.
func (r *VerificationReader) scanVerificationSummaries(rows database.Rows) ([]*domain.VerificationSummary, error) {
	var summaries []*domain.VerificationSummary

	for rows.Next() {
		row := &verificationSummaryRow{}
		err := rows.Scan(
			&row.ID,
			&row.Provider,
			&row.Status,
			&row.InitiatedAt,
			&row.CompletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan verification summary: %w", err)
		}
		summaries = append(summaries, row.toSummary())
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating verification summaries: %w", err)
	}

	return summaries, nil
}

// scanVerificationViews scans rows into verification views.
func (r *VerificationReader) scanVerificationViews(rows database.Rows) ([]*domain.VerificationView, error) {
	var views []*domain.VerificationView

	for rows.Next() {
		row := &verificationViewRow{}
		err := rows.Scan(
			&row.ID,
			&row.UserID,
			&row.Provider,
			&row.CredentialType,
			&row.Status,
			&row.ProviderUserID,
			&row.Username,
			&row.CredentialID,
			&row.FailureReason,
			&row.FailureCode,
			&row.InitiatedAt,
			&row.AuthorizedAt,
			&row.CompletedAt,
			&row.FailedAt,
			&row.ExpiresAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan verification view: %w", err)
		}
		views = append(views, row.toView())
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating verification views: %w", err)
	}

	return views, nil
}

// buildListQuery builds a dynamic list query with filters.
func (r *VerificationReader) buildListQuery(filter domain.VerificationFilter, opts domain.VerificationListOptions) (string, []interface{}) {
	query := queryListVerifications
	args := []interface{}{}
	argIndex := 1

	if filter.UserID != "" {
		query += fmt.Sprintf(" AND user_id = $%d", argIndex)
		args = append(args, filter.UserID)
		argIndex++
	}

	if filter.Provider != "" {
		query += fmt.Sprintf(" AND provider = $%d", argIndex)
		args = append(args, string(filter.Provider))
		argIndex++
	}

	if filter.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, string(filter.Status))
		argIndex++
	}

	if filter.CredentialType != "" {
		query += fmt.Sprintf(" AND credential_type = $%d", argIndex)
		args = append(args, string(filter.CredentialType))
		argIndex++
	}

	if filter.FromDate != nil {
		query += fmt.Sprintf(" AND initiated_at >= $%d", argIndex)
		args = append(args, *filter.FromDate)
		argIndex++
	}

	if filter.ToDate != nil {
		query += fmt.Sprintf(" AND initiated_at <= $%d", argIndex)
		args = append(args, *filter.ToDate)
		argIndex++
	}

	// Add ORDER BY
	orderDir := "DESC"
	if !opts.SortDesc {
		orderDir = "ASC"
	}
	sortBy := "initiated_at"
	if opts.SortBy != "" {
		sortBy = opts.SortBy
	}
	query += fmt.Sprintf(" ORDER BY %s %s", sortBy, orderDir)

	// Add LIMIT and OFFSET
	query += fmt.Sprintf(" LIMIT %d OFFSET %d", opts.Limit, opts.Offset)

	return query, args
}

// buildCountQuery builds a dynamic count query with filters.
func (r *VerificationReader) buildCountQuery(filter domain.VerificationFilter) (string, []interface{}) {
	query := "SELECT COUNT(*) FROM verifications WHERE 1=1"
	args := []interface{}{}
	argIndex := 1

	if filter.UserID != "" {
		query += fmt.Sprintf(" AND user_id = $%d", argIndex)
		args = append(args, filter.UserID)
		argIndex++
	}

	if filter.Provider != "" {
		query += fmt.Sprintf(" AND provider = $%d", argIndex)
		args = append(args, string(filter.Provider))
		argIndex++
	}

	if filter.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, string(filter.Status))
		argIndex++
	}

	if filter.CredentialType != "" {
		query += fmt.Sprintf(" AND credential_type = $%d", argIndex)
		args = append(args, string(filter.CredentialType))
		argIndex++
	}

	if filter.FromDate != nil {
		query += fmt.Sprintf(" AND initiated_at >= $%d", argIndex)
		args = append(args, *filter.FromDate)
		argIndex++
	}

	if filter.ToDate != nil {
		query += fmt.Sprintf(" AND initiated_at <= $%d", argIndex)
		args = append(args, *filter.ToDate)
	}

	return query, args
}

// ============================================================================
// Interface Compliance
// ============================================================================

var _ domain.VerificationReadRepository = (*VerificationReader)(nil)
