package credential

import (
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/0xsj/nexus/platform/internal/credential/app/command"
	"github.com/0xsj/nexus/platform/internal/credential/app/query"
	"github.com/0xsj/nexus/platform/internal/credential/domain"
	"github.com/0xsj/nexus/platform/internal/credential/infrastructure/persistence/postgres"
	v1 "github.com/0xsj/nexus/platform/internal/credential/interface/http/v1"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// ============================================================================
// Provider
// ============================================================================

// Provider holds all dependencies for the Credential bounded context.
type Provider struct {
	// Infrastructure
	Repository       domain.CredentialRepository
	CredentialLookup *postgres.CredentialLookup

	// Application
	CommandHandlers *command.Handlers
	QueryHandlers   *query.Handlers

	// Interface
	HTTPHandler *v1.Handler
}

// ============================================================================
// Configuration
// ============================================================================

// ProviderConfig holds configuration for the Credential provider.
type ProviderConfig struct {
	Pool           *pgxpool.Pool
	Signer         domain.CredentialSigner
	SchemaResolver domain.SchemaResolver
	Publisher      domain.EventPublisher
	Logger         log.Logger
}

// ============================================================================
// Constructor
// ============================================================================

// NewProvider creates a new Credential provider with all dependencies wired.
func NewProvider(cfg ProviderConfig) *Provider {
	// Infrastructure
	repository := postgres.NewCredentialRepository(cfg.Pool)
	lookup := postgres.NewCredentialLookup(cfg.Pool)

	// Application - Command
	commandHandlers := command.NewHandlers(
		repository,
		cfg.Signer,
		cfg.SchemaResolver,
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
		Repository:       repository,
		CredentialLookup: lookup,
		CommandHandlers:  commandHandlers,
		QueryHandlers:    queryHandlers,
		HTTPHandler:      httpHandler,
	}
}

// ============================================================================
// Route Registration
// ============================================================================

// RegisterRoutes registers all Credential HTTP routes on the given router.
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
		Pool:           pool,
		Signer:         &domain.NullCredentialSigner{},
		SchemaResolver: &domain.NullSchemaResolver{},
		Publisher:      &domain.NullEventPublisher{},
		Logger:         logger,
	})
}
