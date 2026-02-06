package organization

import (
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/0xsj/nexus/platform/internal/organization/app/command"
	"github.com/0xsj/nexus/platform/internal/organization/app/query"
	"github.com/0xsj/nexus/platform/internal/organization/domain"
	"github.com/0xsj/nexus/platform/internal/organization/infrastructure/persistence/postgres"
	v1 "github.com/0xsj/nexus/platform/internal/organization/interface/http/v1"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// ============================================================================
// Provider
// ============================================================================

// Provider holds all dependencies for the Organization bounded context.
type Provider struct {
	// Infrastructure
	Repository         domain.OrganizationRepository
	OrganizationLookup *postgres.OrganizationLookup

	// Application
	CommandHandlers *command.Handlers
	QueryHandlers   *query.Handlers

	// Interface
	HTTPHandler *v1.Handler
}

// ============================================================================
// Configuration
// ============================================================================

// ProviderConfig holds configuration for the Organization provider.
type ProviderConfig struct {
	Pool                *pgxpool.Pool
	IdentityReader      domain.IdentityReader
	SlugLookup          domain.SlugLookup
	DIDService          domain.DIDService
	NotificationService domain.NotificationService
	Publisher           domain.EventPublisher
	Logger              log.Logger
}

// ============================================================================
// Constructor
// ============================================================================

// NewProvider creates a new Organization provider with all dependencies wired.
func NewProvider(cfg ProviderConfig) *Provider {
	// Infrastructure
	repository := postgres.NewOrganizationRepository(cfg.Pool)
	lookup := postgres.NewOrganizationLookup(cfg.Pool)

	// Use lookup as slug lookup if not provided
	slugLookup := cfg.SlugLookup
	if slugLookup == nil {
		slugLookup = lookup
	}

	// Application - Command
	commandHandlers := command.NewHandlers(
		repository,
		slugLookup,
		cfg.IdentityReader,
		cfg.Publisher,
		cfg.Logger,
	)

	// Application - Query
	queryHandlers := query.NewHandlers(
		lookup,
		cfg.Logger,
	)

	// Interface - HTTP
	httpHandler := v1.NewHandler(commandHandlers, queryHandlers)

	return &Provider{
		Repository:         repository,
		OrganizationLookup: lookup,
		CommandHandlers:    commandHandlers,
		QueryHandlers:      queryHandlers,
		HTTPHandler:        httpHandler,
	}
}

// ============================================================================
// Route Registration
// ============================================================================

// RegisterRoutes registers all Organization HTTP routes on the given router.
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
		Pool:                pool,
		IdentityReader:      domain.NewNullIdentityReader(),
		SlugLookup:          domain.NewNullSlugLookup(),
		DIDService:          domain.NewNullDIDService(),
		NotificationService: domain.NewNullNotificationService(),
		Publisher:           domain.NewNullEventPublisher(),
		Logger:              logger,
	})
}
