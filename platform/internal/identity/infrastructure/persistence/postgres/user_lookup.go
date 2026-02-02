package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/0xsj/nexus/platform/internal/identity/domain"
	"github.com/0xsj/nexus/platform/internal/identity/infrastructure/persistence/postgres/generated"
)

// UserLookup provides query-side operations for users.
// It reads from the projection tables (not event store).
type UserLookup struct {
	pool    *pgxpool.Pool
	queries *generated.Queries
}

// NewUserLookup creates a new UserLookup.
func NewUserLookup(pool *pgxpool.Pool) *UserLookup {
	return &UserLookup{
		pool:    pool,
		queries: generated.New(pool),
	}
}

// FindByID finds a user by ID from the projection.
func (l *UserLookup) FindByID(ctx context.Context, id domain.UserID) (*UserProjection, error) {
	row, err := l.queries.GetUserByID(ctx, uuidFromUserID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.UserNotFound("UserLookup.FindByID", id.String())
		}
		return nil, err
	}

	return userRowToProjection(row), nil
}

// FindByEmail finds a user by email from the projection.
func (l *UserLookup) FindByEmail(ctx context.Context, email string) (*UserProjection, error) {
	row, err := l.queries.GetUserByEmail(ctx, &email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.UserNotFound("UserLookup.FindByEmail", email)
		}
		return nil, err
	}

	return userRowToProjection(row), nil
}

// FindByDID finds a user by DID.
func (l *UserLookup) FindByDID(ctx context.Context, did string) (*UserProjection, error) {
	userUUID, err := l.queries.GetUserIDByDID(ctx, did)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.DIDNotFound("UserLookup.FindByDID", did)
		}
		return nil, err
	}

	row, err := l.queries.GetUserByID(ctx, userUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.UserNotFound("UserLookup.FindByDID", userUUID.String())
		}
		return nil, err
	}

	return userRowToProjection(row), nil
}

// FindByOAuthSubject finds a user by OAuth provider and external ID.
func (l *UserLookup) FindByOAuthSubject(ctx context.Context, provider domain.OAuthProvider, externalID string) (*UserProjection, error) {
	params := generated.GetUserIDByOAuthSubjectParams{
		Provider:   provider.String(),
		ExternalID: externalID,
	}

	userUUID, err := l.queries.GetUserIDByOAuthSubject(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.UserNotFound("UserLookup.FindByOAuthSubject", provider.String()+":"+externalID)
		}
		return nil, err
	}

	row, err := l.queries.GetUserByID(ctx, userUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.UserNotFound("UserLookup.FindByOAuthSubject", userUUID.String())
		}
		return nil, err
	}

	return userRowToProjection(row), nil
}

// ExistsByID checks if a user exists by ID.
func (l *UserLookup) ExistsByID(ctx context.Context, id domain.UserID) (bool, error) {
	return l.queries.UserExistsByID(ctx, uuidFromUserID(id))
}

// ExistsByEmail checks if a user exists by email.
func (l *UserLookup) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	return l.queries.UserExistsByEmail(ctx, &email)
}

// ExistsByDID checks if a user exists by DID.
func (l *UserLookup) ExistsByDID(ctx context.Context, did string) (bool, error) {
	return l.queries.UserExistsByDID(ctx, did)
}

// ExistsByOAuthSubject checks if a user exists by OAuth subject.
func (l *UserLookup) ExistsByOAuthSubject(ctx context.Context, provider domain.OAuthProvider, externalID string) (bool, error) {
	params := generated.UserExistsByOAuthSubjectParams{
		Provider:   provider.String(),
		ExternalID: externalID,
	}
	return l.queries.UserExistsByOAuthSubject(ctx, params)
}

// GetUserDIDs returns all DIDs for a user.
func (l *UserLookup) GetUserDIDs(ctx context.Context, userID domain.UserID) ([]DIDProjection, error) {
	rows, err := l.queries.GetUserDIDs(ctx, uuidFromUserID(userID))
	if err != nil {
		return nil, err
	}

	dids := make([]DIDProjection, len(rows))
	for i, row := range rows {
		dids[i] = DIDProjection{
			DID:       row.Did,
			IsPrimary: row.IsPrimary,
			AddedAt:   row.AddedAt,
		}
	}

	return dids, nil
}

// GetUserOAuthLinks returns all OAuth links for a user.
func (l *UserLookup) GetUserOAuthLinks(ctx context.Context, userID domain.UserID) ([]OAuthLinkProjection, error) {
	rows, err := l.queries.GetUserOAuthLinks(ctx, uuidFromUserID(userID))
	if err != nil {
		return nil, err
	}

	links := make([]OAuthLinkProjection, len(rows))
	for i, row := range rows {
		var email string
		if row.Email != nil {
			email = *row.Email
		}
		links[i] = OAuthLinkProjection{
			Provider:   row.Provider,
			ExternalID: row.ExternalID,
			Email:      email,
			LinkedAt:   row.LinkedAt,
		}
	}

	return links, nil
}

// ============================================================================
// Projection Types
// ============================================================================

// UserProjection represents a user from the read model.
type UserProjection struct {
	ID          domain.UserID
	Email       string
	DisplayName string
	Status      domain.UserStatus
	PrimaryDID  string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Version     int
}

// DIDProjection represents a DID from the read model.
type DIDProjection struct {
	DID       string
	IsPrimary bool
	AddedAt   time.Time
}

// OAuthLinkProjection represents an OAuth link from the read model.
type OAuthLinkProjection struct {
	Provider   string
	ExternalID string
	Email      string
	LinkedAt   time.Time
}

// ============================================================================
// Helpers
// ============================================================================

// userRowToProjection converts a database row to a UserProjection.
func userRowToProjection(row generated.IdentityUser) *UserProjection {
	userID := userIDFromUUID(row.ID)
	status, _ := domain.ParseUserStatus(row.Status)

	var email string
	if row.Email != nil {
		email = *row.Email
	}

	return &UserProjection{
		ID:          userID,
		Email:       email,
		DisplayName: row.DisplayName,
		Status:      status,
		PrimaryDID:  row.PrimaryDid,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
		Version:     int(row.Version),
	}
}
