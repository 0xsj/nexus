package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/0xsj/nexus/platform/internal/credential/infrastructure/persistence/postgres/generated"
)

// CredentialLookup provides read-side queries on the credentials projection table.
type CredentialLookup struct {
	pool    *pgxpool.Pool
	queries *generated.Queries
}

// NewCredentialLookup creates a new CredentialLookup.
func NewCredentialLookup(pool *pgxpool.Pool) *CredentialLookup {
	return &CredentialLookup{
		pool:    pool,
		queries: generated.New(pool),
	}
}

// GetByID returns a credential projection by ID.
func (l *CredentialLookup) GetByID(ctx context.Context, id string) (*CredentialProjection, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	row, err := l.queries.GetCredentialByID(ctx, uid)
	if err != nil {
		return nil, err
	}

	proj := CredentialRowToProjection(row)
	return &proj, nil
}

// ExistsByID checks if a credential exists by ID.
func (l *CredentialLookup) ExistsByID(ctx context.Context, id string) (bool, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return false, err
	}

	return l.queries.CredentialExistsByID(ctx, uid)
}

// ListBySubjectDID returns credentials for a subject DID with pagination.
func (l *CredentialLookup) ListBySubjectDID(ctx context.Context, subjectDID string, limit, offset int) ([]CredentialProjection, error) {
	rows, err := l.queries.ListCredentialsBySubjectDID(ctx, generated.ListCredentialsBySubjectDIDParams{
		SubjectDid: subjectDID,
		Limit:      int32(limit),
		Offset:     int32(offset),
	})
	if err != nil {
		return nil, err
	}

	projections := make([]CredentialProjection, len(rows))
	for i, row := range rows {
		projections[i] = CredentialRowToProjection(row)
	}
	return projections, nil
}

// CountBySubjectDID returns the count of credentials for a subject DID.
func (l *CredentialLookup) CountBySubjectDID(ctx context.Context, subjectDID string) (int, error) {
	count, err := l.queries.CountCredentialsBySubjectDID(ctx, subjectDID)
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

// ListByIssuerDID returns credentials for an issuer DID with pagination.
func (l *CredentialLookup) ListByIssuerDID(ctx context.Context, issuerDID string, limit, offset int) ([]CredentialProjection, error) {
	rows, err := l.queries.ListCredentialsByIssuerDID(ctx, generated.ListCredentialsByIssuerDIDParams{
		IssuerDid: issuerDID,
		Limit:     int32(limit),
		Offset:    int32(offset),
	})
	if err != nil {
		return nil, err
	}

	projections := make([]CredentialProjection, len(rows))
	for i, row := range rows {
		projections[i] = CredentialRowToProjection(row)
	}
	return projections, nil
}

// CountByIssuerDID returns the count of credentials for an issuer DID.
func (l *CredentialLookup) CountByIssuerDID(ctx context.Context, issuerDID string) (int, error) {
	count, err := l.queries.CountCredentialsByIssuerDID(ctx, issuerDID)
	if err != nil {
		return 0, err
	}
	return int(count), nil
}
