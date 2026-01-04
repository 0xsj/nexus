package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/0xsj/nexus/platform/internal/credential/application/query"
	"github.com/0xsj/nexus/platform/internal/credential/domain"
	"github.com/0xsj/nexus/platform/pkg/database"
	pgadapter "github.com/0xsj/nexus/platform/pkg/database/postgres"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// ============================================================================
// Repository
// ============================================================================

// Repository implements credential persistence using PostgreSQL.
type Repository struct {
	adapter *pgadapter.BaseAdapter
}

// NewRepository creates a new PostgreSQL repository.
func NewRepository(db *pgadapter.DB) *Repository {
	return &Repository{
		adapter: pgadapter.NewBaseAdapter(db),
	}
}

// ============================================================================
// Write Operations (domain.CredentialRepository)
// ============================================================================

// Save persists a credential aggregate.
func (r *Repository) Save(ctx context.Context, credential *domain.Credential) error {
	// Check if credential exists
	exists, err := r.Exists(ctx, credential.AggregateID())
	if err != nil {
		return err
	}

	if exists {
		return r.update(ctx, credential)
	}
	return r.insert(ctx, credential)
}

// insert creates a new credential record.
func (r *Repository) insert(ctx context.Context, credential *domain.Credential) error {
	now := time.Now().UTC()
	args, err := CredentialToRow(credential, now, now)
	if err != nil {
		return fmt.Errorf("failed to map credential: %w", err)
	}

	_, err = r.adapter.Exec(ctx, queryInsert, args...)
	if err != nil {
		return fmt.Errorf("failed to insert credential: %w", err)
	}

	return nil
}

// update updates an existing credential record with optimistic locking.
func (r *Repository) update(ctx context.Context, credential *domain.Credential) error {
	args, err := CredentialToUpdateRow(credential)
	if err != nil {
		return fmt.Errorf("failed to map credential: %w", err)
	}

	result, err := r.adapter.Exec(ctx, queryUpdate, args...)
	if err != nil {
		return fmt.Errorf("failed to update credential: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		// Fetch actual version for error message
		var actualVersion int
		row := r.adapter.Executor().QueryRow(ctx, "SELECT version FROM credentials WHERE id = $1", credential.AggregateID())
		_ = row.Scan(&actualVersion)

		return eventsourcing.ErrConcurrencyConflict(
			"Repository.update",
			credential.AggregateID(),
			credential.Version(),
			actualVersion,
		)
	}

	return nil
}

// Load retrieves a credential aggregate by ID.
func (r *Repository) Load(ctx context.Context, id string) (*domain.Credential, error) {
	row := &CredentialRow{}
	dbRow := r.adapter.Executor().QueryRow(ctx, querySelectByID, id)
	err := dbRow.Scan(row.ScanFields()...)
	if err != nil {
		if isNoRows(err) {
			return nil, eventsourcing.ErrAggregateNotFound(
				"Repository.Load",
				domain.AggregateType,
				id,
			)
		}
		return nil, fmt.Errorf("failed to load credential: %w", err)
	}

	return RowToCredential(row)
}

// Exists checks if a credential exists.
func (r *Repository) Exists(ctx context.Context, id string) (bool, error) {
	var exists bool
	row := r.adapter.Executor().QueryRow(ctx, queryExists, id)
	err := row.Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check credential existence: %w", err)
	}
	return exists, nil
}

// Delete removes a credential by ID.
func (r *Repository) Delete(ctx context.Context, id string) error {
	result, err := r.adapter.Exec(ctx, queryDelete, id)
	if err != nil {
		return fmt.Errorf("failed to delete credential: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return eventsourcing.ErrAggregateNotFound(
			"Repository.Delete",
			domain.AggregateType,
			id,
		)
	}

	return nil
}

// ============================================================================
// Read Operations (query.CredentialReadRepository)
// ============================================================================

// GetByID retrieves a credential view by ID.
func (r *Repository) GetByID(ctx context.Context, id string) (*query.CredentialView, error) {
	row := &CredentialRow{}
	dbRow := r.adapter.Executor().QueryRow(ctx, querySelectByID, id)
	err := dbRow.Scan(row.ScanFields()...)
	if err != nil {
		if isNoRows(err) {
			return nil, eventsourcing.ErrAggregateNotFound(
				"Repository.GetByID",
				domain.AggregateType,
				id,
			)
		}
		return nil, fmt.Errorf("failed to get credential: %w", err)
	}

	return RowToCredentialView(row), nil
}

// List retrieves a paginated list of credentials.
func (r *Repository) List(ctx context.Context, opts query.ListOptions) (*query.CredentialListResult, error) {
	// Build WHERE clause
	where, args := buildWhereClause(opts, 1)

	// Count total
	countQuery := queryCountBase + where
	var total int
	row := r.adapter.Executor().QueryRow(ctx, countQuery, args...)
	if err := row.Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count credentials: %w", err)
	}

	// Build ORDER BY
	orderBy := buildOrderBy(opts)

	// Build full query with pagination
	selectQuery := querySelectBase + where + orderBy + fmt.Sprintf(" LIMIT %d OFFSET %d", opts.Limit, opts.Offset)

	rows, err := r.adapter.Select(ctx, selectQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list credentials: %w", err)
	}
	defer rows.Close()

	credentials, err := scanCredentialViews(rows)
	if err != nil {
		return nil, err
	}

	return &query.CredentialListResult{
		Credentials: credentials,
		Total:       total,
		Limit:       opts.Limit,
		Offset:      opts.Offset,
		HasMore:     opts.Offset+len(credentials) < total,
	}, nil
}

// ListByHolder retrieves credentials for a specific holder.
func (r *Repository) ListByHolder(ctx context.Context, holderDID string, opts query.ListOptions) (*query.CredentialListResult, error) {
	// Build WHERE clause starting with holder_did
	where := " WHERE holder_did = $1"
	args := []any{holderDID}
	argIndex := 2

	if opts.Status != "" {
		where += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, opts.Status)
		argIndex++
	}
	if opts.CredentialType != "" {
		where += fmt.Sprintf(" AND credential_type = $%d", argIndex)
		args = append(args, opts.CredentialType)
	}

	// Count total
	countQuery := queryCountBase + where
	var total int
	row := r.adapter.Executor().QueryRow(ctx, countQuery, args...)
	if err := row.Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count credentials: %w", err)
	}

	// Build ORDER BY
	orderBy := buildOrderBy(opts)

	// Build full query
	selectQuery := querySelectBase + where + orderBy + fmt.Sprintf(" LIMIT %d OFFSET %d", opts.Limit, opts.Offset)

	rows, err := r.adapter.Select(ctx, selectQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list credentials by holder: %w", err)
	}
	defer rows.Close()

	credentials, err := scanCredentialViews(rows)
	if err != nil {
		return nil, err
	}

	return &query.CredentialListResult{
		Credentials: credentials,
		Total:       total,
		Limit:       opts.Limit,
		Offset:      opts.Offset,
		HasMore:     opts.Offset+len(credentials) < total,
	}, nil
}

// ListByIssuer retrieves credentials issued by a specific issuer.
func (r *Repository) ListByIssuer(ctx context.Context, issuerDID string, opts query.ListOptions) (*query.CredentialListResult, error) {
	// Build WHERE clause starting with issuer_did
	where := " WHERE issuer_did = $1"
	args := []any{issuerDID}
	argIndex := 2

	if opts.Status != "" {
		where += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, opts.Status)
		argIndex++
	}
	if opts.CredentialType != "" {
		where += fmt.Sprintf(" AND credential_type = $%d", argIndex)
		args = append(args, opts.CredentialType)
	}

	// Count total
	countQuery := queryCountBase + where
	var total int
	row := r.adapter.Executor().QueryRow(ctx, countQuery, args...)
	if err := row.Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count credentials: %w", err)
	}

	// Build ORDER BY
	orderBy := buildOrderBy(opts)

	// Build full query
	selectQuery := querySelectBase + where + orderBy + fmt.Sprintf(" LIMIT %d OFFSET %d", opts.Limit, opts.Offset)

	rows, err := r.adapter.Select(ctx, selectQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list credentials by issuer: %w", err)
	}
	defer rows.Close()

	credentials, err := scanCredentialViews(rows)
	if err != nil {
		return nil, err
	}

	return &query.CredentialListResult{
		Credentials: credentials,
		Total:       total,
		Limit:       opts.Limit,
		Offset:      opts.Offset,
		HasMore:     opts.Offset+len(credentials) < total,
	}, nil
}

// Count returns the total number of credentials.
func (r *Repository) Count(ctx context.Context) (int, error) {
	var count int
	row := r.adapter.Executor().QueryRow(ctx, queryCountBase)
	err := row.Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count credentials: %w", err)
	}
	return count, nil
}

// CountByStatus returns the number of credentials with a specific status.
func (r *Repository) CountByStatus(ctx context.Context, status string) (int, error) {
	var count int
	row := r.adapter.Executor().QueryRow(ctx, queryCountBase+" WHERE status = $1", status)
	err := row.Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count credentials by status: %w", err)
	}
	return count, nil
}

// ============================================================================
// Query Helpers
// ============================================================================

// buildWhereClause builds a WHERE clause from list options.
func buildWhereClause(opts query.ListOptions, startArg int) (string, []any) {
	var conditions []string
	var args []any
	argIndex := startArg

	if opts.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, opts.Status)
		argIndex++
	}

	if opts.CredentialType != "" {
		conditions = append(conditions, fmt.Sprintf("credential_type = $%d", argIndex))
		args = append(args, opts.CredentialType)
	}

	if len(conditions) == 0 {
		return "", nil
	}

	return " WHERE " + strings.Join(conditions, " AND "), args
}

// buildOrderBy builds an ORDER BY clause from list options.
func buildOrderBy(opts query.ListOptions) string {
	column := "created_at"
	if col, ok := OrderByColumn[opts.SortBy]; ok {
		column = col
	}

	order := ValidSortOrder(opts.SortOrder)

	return fmt.Sprintf(" ORDER BY %s %s", column, order)
}

// scanCredentialViews scans rows into credential views.
func scanCredentialViews(rows database.Rows) ([]*query.CredentialView, error) {
	var credentials []*query.CredentialView

	for rows.Next() {
		row := &CredentialRow{}
		if err := rows.Scan(row.ScanFields()...); err != nil {
			return nil, fmt.Errorf("failed to scan credential row: %w", err)
		}
		credentials = append(credentials, RowToCredentialView(row))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating credential rows: %w", err)
	}

	return credentials, nil
}

// isNoRows checks if the error is a "no rows" error.
func isNoRows(err error) bool {
	return err != nil && err.Error() == "no rows in result set"
}

// ============================================================================
// Interface Compliance
// ============================================================================

var (
	_ domain.CredentialRepository    = (*Repository)(nil)
	_ query.CredentialReadRepository = (*Repository)(nil)
)
