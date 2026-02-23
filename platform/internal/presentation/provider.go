package presentation

import (
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/0xsj/nexus/platform/internal/presentation/app/command"
	"github.com/0xsj/nexus/platform/internal/presentation/app/query"
	"github.com/0xsj/nexus/platform/internal/presentation/domain"
	"github.com/0xsj/nexus/platform/internal/presentation/infrastructure/persistence/postgres"
	"github.com/0xsj/nexus/platform/internal/presentation/infrastructure/persistence/postgres/generated"
	"github.com/0xsj/nexus/platform/internal/presentation/infrastructure/projections"
	v1 "github.com/0xsj/nexus/platform/internal/presentation/interface/http/v1"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// ============================================================================
// Provider
// ============================================================================

// Provider holds all dependencies for the Presentation bounded context.
type Provider struct {
	// Infrastructure
	PresentationRepo   domain.PresentationRepository
	ShareLinkRepo      domain.ShareLinkRepository
	PresentationLookup *postgres.PresentationLookup
	ShareLinkLookup    *postgres.ShareLinkLookup

	// Projections
	IdentityProjector   *projections.IdentityProjector
	CredentialProjector *projections.CredentialProjector

	// Application
	CommandHandlers *command.Handlers
	QueryHandlers   *query.Handlers

	// Interface
	HTTPHandler *v1.Handler
}

// ============================================================================
// Configuration
// ============================================================================

// ProviderConfig holds configuration for the Presentation provider.
type ProviderConfig struct {
	Pool             *pgxpool.Pool
	CredentialReader domain.CredentialReader
	IdentityReader   domain.IdentityReader
	Publisher        domain.EventPublisher
	Logger           log.Logger
}

// ============================================================================
// Constructor
// ============================================================================

// NewProvider creates a new Presentation provider with all dependencies wired.
func NewProvider(cfg ProviderConfig) *Provider {
	// Infrastructure
	presentationRepo := postgres.NewPresentationRepository(cfg.Pool)
	shareLinkRepo := postgres.NewShareLinkRepository(cfg.Pool)
	presentationLookup := postgres.NewPresentationLookup(cfg.Pool)
	shareLinkLookup := postgres.NewShareLinkLookup(cfg.Pool)

	// Projections
	projQueries := generated.New(cfg.Pool)
	identityProjector := projections.NewIdentityProjector(projQueries)
	credentialProjector := projections.NewCredentialProjector(projQueries)

	// Use projection-backed readers if none provided
	identityReader := cfg.IdentityReader
	if identityReader == nil {
		identityReader = projections.NewIdentityReader(projQueries)
	}
	credentialReader := cfg.CredentialReader
	if credentialReader == nil {
		credentialReader = projections.NewCredentialReader(projQueries)
	}

	// Application - Command
	commandHandlers := command.NewHandlers(
		presentationRepo,
		shareLinkRepo,
		shareLinkLookup,
		identityReader,
		credentialReader,
		cfg.Publisher,
		cfg.Logger,
	)

	// Application - Query
	queryHandlers := query.NewHandlers(
		presentationLookup,
		shareLinkLookup,
		cfg.Logger,
	)

	// Interface - HTTP
	httpHandler := v1.NewHandler(commandHandlers, queryHandlers)

	return &Provider{
		PresentationRepo:    presentationRepo,
		ShareLinkRepo:       shareLinkRepo,
		PresentationLookup:  presentationLookup,
		ShareLinkLookup:     shareLinkLookup,
		IdentityProjector:   identityProjector,
		CredentialProjector: credentialProjector,
		CommandHandlers:     commandHandlers,
		QueryHandlers:       queryHandlers,
		HTTPHandler:         httpHandler,
	}
}

// ============================================================================
// Route Registration
// ============================================================================

// RegisterRoutes registers all Presentation HTTP routes on the given router.
func (p *Provider) RegisterRoutes(r chi.Router) {
	p.HTTPHandler.RegisterRoutes(r)
}

// ============================================================================
// Optional Dependencies
// ============================================================================

// NewProviderWithDefaults creates a Provider with default/null implementations
// for optional dependencies. Useful for testing or standalone operation.
func NewProviderWithDefaults(pool *pgxpool.Pool, logger log.Logger) *Provider {
	return NewProvider(ProviderConfig{
		Pool:             pool,
		CredentialReader: &domain.NullCredentialReader{},
		IdentityReader:   &domain.NullIdentityReader{},
		Publisher:        &domain.NullEventPublisher{},
		Logger:           logger,
	})
}
