package postgres

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/0xsj/nexus/platform/internal/presentation/infrastructure/persistence/postgres/generated"
)

// ShareLinkLookup provides read-side queries on the share_links and access_grants projection tables.
type ShareLinkLookup struct {
	pool    *pgxpool.Pool
	queries *generated.Queries
}

// NewShareLinkLookup creates a new ShareLinkLookup.
func NewShareLinkLookup(pool *pgxpool.Pool) *ShareLinkLookup {
	return &ShareLinkLookup{
		pool:    pool,
		queries: generated.New(pool),
	}
}

// GetByID returns a share link projection by ID.
func (l *ShareLinkLookup) GetByID(ctx context.Context, id string) (*ShareLinkProjection, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	row, err := l.queries.GetShareLinkByID(ctx, uid)
	if err != nil {
		return nil, err
	}

	proj := ShareLinkRowToProjection(row)
	return &proj, nil
}

// GetByToken returns a share link projection by its unique token.
func (l *ShareLinkLookup) GetByToken(ctx context.Context, token string) (*ShareLinkProjection, error) {
	row, err := l.queries.GetShareLinkByToken(ctx, token)
	if err != nil {
		return nil, err
	}

	proj := ShareLinkRowToProjection(row)
	return &proj, nil
}

// ListByPresentationID returns share links for a presentation ID with pagination.
func (l *ShareLinkLookup) ListByPresentationID(ctx context.Context, presentationID string, limit, offset int) ([]ShareLinkProjection, error) {
	uid, err := uuid.Parse(presentationID)
	if err != nil {
		return nil, err
	}

	rows, err := l.queries.ListShareLinksByPresentationID(ctx, generated.ListShareLinksByPresentationIDParams{
		PresentationID: uid,
		Limit:          int32(limit),
		Offset:         int32(offset),
	})
	if err != nil {
		return nil, err
	}

	projections := make([]ShareLinkProjection, len(rows))
	for i, row := range rows {
		projections[i] = ShareLinkRowToProjection(row)
	}
	return projections, nil
}

// CountByPresentationID returns the count of share links for a presentation ID.
func (l *ShareLinkLookup) CountByPresentationID(ctx context.Context, presentationID string) (int, error) {
	uid, err := uuid.Parse(presentationID)
	if err != nil {
		return 0, err
	}

	count, err := l.queries.CountShareLinksByPresentationID(ctx, uid)
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

// ListAccessGrants returns access grants for a share link with pagination.
func (l *ShareLinkLookup) ListAccessGrants(ctx context.Context, shareLinkID string, limit, offset int) ([]AccessGrantProjection, error) {
	uid, err := uuid.Parse(shareLinkID)
	if err != nil {
		return nil, err
	}

	rows, err := l.queries.ListAccessGrantsByShareLinkID(ctx, generated.ListAccessGrantsByShareLinkIDParams{
		ShareLinkID: uid,
		Limit:       int32(limit),
		Offset:      int32(offset),
	})
	if err != nil {
		return nil, err
	}

	projections := make([]AccessGrantProjection, len(rows))
	for i, row := range rows {
		projections[i] = AccessGrantRowToProjection(row)
	}
	return projections, nil
}

// CountAccessGrants returns the count of access grants for a share link.
func (l *ShareLinkLookup) CountAccessGrants(ctx context.Context, shareLinkID string) (int, error) {
	uid, err := uuid.Parse(shareLinkID)
	if err != nil {
		return 0, err
	}

	count, err := l.queries.CountAccessGrantsByShareLinkID(ctx, uid)
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

// InsertAccessGrant inserts a new access grant record.
func (l *ShareLinkLookup) InsertAccessGrant(ctx context.Context, shareLinkID, verifierDID, ipAddress string, accessedAt time.Time, disclosedClaims map[string]any) error {
	slUID, err := uuid.Parse(shareLinkID)
	if err != nil {
		return err
	}

	var verifierDIDPtr *string
	if verifierDID != "" {
		verifierDIDPtr = &verifierDID
	}

	var ipAddressPtr *string
	if ipAddress != "" {
		ipAddressPtr = &ipAddress
	}

	var claimsJSON []byte
	if disclosedClaims != nil {
		claimsJSON, err = json.Marshal(disclosedClaims)
		if err != nil {
			return err
		}
	}

	return l.queries.InsertAccessGrant(ctx, generated.InsertAccessGrantParams{
		ID:              uuid.New(),
		ShareLinkID:     slUID,
		VerifierDid:     verifierDIDPtr,
		AccessedAt:      accessedAt,
		IpAddress:       ipAddressPtr,
		DisclosedClaims: claimsJSON,
	})
}
