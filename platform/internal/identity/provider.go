package identity

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/0xsj/nexus/platform/internal/identity/app/command"
	"github.com/0xsj/nexus/platform/internal/identity/app/query"
	"github.com/0xsj/nexus/platform/internal/identity/domain"
	"github.com/0xsj/nexus/platform/internal/identity/infrastructure/persistence/postgres"
	v1 "github.com/0xsj/nexus/platform/internal/identity/interface/http/v1"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
	"github.com/0xsj/nexus/platform/pkg/types"
)

// ============================================================================
// Provider
// ============================================================================

// Provider holds all dependencies for the Identity bounded context.
type Provider struct {
	// Infrastructure - Repositories
	UserRepository       domain.UserRepository
	SessionRepository    domain.SessionRepository
	MagicLinkRepository  domain.MagicLinkRepository
	OAuthStateRepository domain.OAuthStateRepository

	// Infrastructure - Lookups
	UserLookup    domain.UserLookup
	SessionLookup domain.SessionLookup

	// Application
	CommandHandlers *command.Handlers
	QueryHandlers   *query.Handlers

	// Interface
	HTTPHandler *v1.Handler
}

// ============================================================================
// Configuration
// ============================================================================

// ProviderConfig holds external dependencies required by the Identity provider.
type ProviderConfig struct {
	Pool         *pgxpool.Pool
	DIDGenerator domain.DIDGenerator
	WalletReader domain.WalletReader
	OAuthService domain.OAuthService
	EmailService domain.EmailService
	Publisher    domain.EventPublisher
	Logger       log.Logger
}

// ============================================================================
// Constructor
// ============================================================================

// NewProvider creates a new Identity provider with all dependencies wired.
func NewProvider(cfg ProviderConfig) *Provider {
	// Infrastructure - Repositories
	userRepo := postgres.NewUserRepository(cfg.Pool)
	sessionRepo := postgres.NewSessionRepository(cfg.Pool)
	magicLinkRepo := postgres.NewMagicLinkRepository(cfg.Pool)
	oauthStateRepo := postgres.NewOAuthStateRepository(cfg.Pool)

	// Infrastructure - Lookups (with adapters to match domain interfaces)
	postgresUserLookup := postgres.NewUserLookup(cfg.Pool)
	postgresSessionLookup := postgres.NewSessionLookup(cfg.Pool)

	userLookup := &userLookupAdapter{lookup: postgresUserLookup}
	sessionLookup := &sessionLookupAdapter{lookup: postgresSessionLookup, sessionRepo: sessionRepo}

	// Application - Command Handlers
	commandHandlers := command.NewHandlers(
		userRepo,
		sessionRepo,
		magicLinkRepo,
		oauthStateRepo,
		userLookup,
		sessionLookup,
		cfg.DIDGenerator,
		cfg.WalletReader,
		cfg.OAuthService,
		cfg.EmailService,
		cfg.Publisher,
		cfg.Logger,
	)

	// Application - Query Handlers
	queryHandlers := query.NewHandlers(
		userRepo,
		sessionRepo,
		oauthStateRepo,
		userLookup,
		sessionLookup,
		cfg.Logger,
	)

	// Interface - HTTP
	httpHandler := v1.NewHandler(commandHandlers, queryHandlers)

	return &Provider{
		UserRepository:       userRepo,
		SessionRepository:    sessionRepo,
		MagicLinkRepository:  magicLinkRepo,
		OAuthStateRepository: oauthStateRepo,
		UserLookup:           userLookup,
		SessionLookup:        sessionLookup,
		CommandHandlers:      commandHandlers,
		QueryHandlers:        queryHandlers,
		HTTPHandler:          httpHandler,
	}
}

// ============================================================================
// Route Registration
// ============================================================================

// RegisterRoutes registers all Identity HTTP routes on the given router.
func (p *Provider) RegisterRoutes(r chi.Router, authMiddleware func(next http.Handler) http.Handler) {
	v1.RegisterRoutes(r, p.HTTPHandler, authMiddleware)
}

// ============================================================================
// Session Validation (for auth middleware)
// ============================================================================

// ValidateSession validates a session and returns user/session info if valid.
// This is exposed for use by auth middleware in other contexts.
func (p *Provider) ValidateSession(ctx context.Context, sessionID string, token string) (*query.SessionValidationView, error) {
	parsedID, err := types.ParseID(sessionID)
	if err != nil {
		return nil, err
	}

	return p.QueryHandlers.HandleValidateSession(ctx, query.ValidateSession{
		SessionID: parsedID,
		Token:     token,
	})
}

// ============================================================================
// Lookup Adapters
// ============================================================================

// userLookupAdapter adapts postgres.UserLookup to domain.UserLookup interface.
type userLookupAdapter struct {
	lookup *postgres.UserLookup
}

func (a *userLookupAdapter) ExistsByEmail(ctx context.Context, email types.Email) (bool, error) {
	return a.lookup.ExistsByEmail(ctx, email.String())
}

func (a *userLookupAdapter) ExistsByDID(ctx context.Context, did string) (bool, error) {
	return a.lookup.ExistsByDID(ctx, did)
}

func (a *userLookupAdapter) ExistsByOAuthSubject(ctx context.Context, subject domain.OAuthSubject) (bool, error) {
	return a.lookup.ExistsByOAuthSubject(ctx, subject.Provider(), subject.ExternalID())
}

func (a *userLookupAdapter) GetUserIDByEmail(ctx context.Context, email types.Email) (domain.UserID, error) {
	projection, err := a.lookup.FindByEmail(ctx, email.String())
	if err != nil {
		return domain.UserID{}, err
	}
	return projection.ID, nil
}

func (a *userLookupAdapter) GetUserIDByDID(ctx context.Context, did string) (domain.UserID, error) {
	projection, err := a.lookup.FindByDID(ctx, did)
	if err != nil {
		return domain.UserID{}, err
	}
	return projection.ID, nil
}

func (a *userLookupAdapter) GetUserIDByOAuthSubject(ctx context.Context, subject domain.OAuthSubject) (domain.UserID, error) {
	projection, err := a.lookup.FindByOAuthSubject(ctx, subject.Provider(), subject.ExternalID())
	if err != nil {
		return domain.UserID{}, err
	}
	return projection.ID, nil
}

// sessionLookupAdapter adapts postgres.SessionLookup to domain.SessionLookup interface.
type sessionLookupAdapter struct {
	lookup      *postgres.SessionLookup
	sessionRepo domain.SessionRepository
}

func (a *sessionLookupAdapter) GetSessionByTokenHash(ctx context.Context, tokenHash string) (*domain.Session, error) {
	projection, err := a.lookup.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, err
	}
	// Load full aggregate from repository
	return a.sessionRepo.Get(ctx, projection.ID)
}

func (a *sessionLookupAdapter) GetActiveSessionsForUser(ctx context.Context, userID domain.UserID) ([]*domain.Session, error) {
	projections, err := a.lookup.FindActiveByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	sessions := make([]*domain.Session, 0, len(projections))
	for _, proj := range projections {
		session, err := a.sessionRepo.Get(ctx, proj.ID)
		if err != nil {
			// Skip sessions that can't be loaded (shouldn't happen)
			continue
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

func (a *sessionLookupAdapter) CountActiveSessionsForUser(ctx context.Context, userID domain.UserID) (int, error) {
	return a.lookup.CountActiveByUserID(ctx, userID)
}
