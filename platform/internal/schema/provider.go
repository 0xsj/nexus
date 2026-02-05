package schema

import (
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/0xsj/nexus/platform/internal/schema/app/command"
	"github.com/0xsj/nexus/platform/internal/schema/app/query"
	"github.com/0xsj/nexus/platform/internal/schema/domain"
	"github.com/0xsj/nexus/platform/internal/schema/infrastructure/persistence/postgres"
	v1 "github.com/0xsj/nexus/platform/internal/schema/interface/http/v1"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// ============================================================================
// Provider
// ============================================================================

// Provider holds all dependencies for the Schema bounded context.
type Provider struct {
	// Infrastructure
	Repository   domain.SchemaRepository
	SchemaLookup domain.SchemaLookup

	// Application
	CommandHandlers *command.Handlers
	QueryHandlers   *query.Handlers

	// Interface
	HTTPHandler *v1.Handler
}

// ============================================================================
// Constructor
// ============================================================================

// NewProvider creates a new Schema provider with all dependencies wired.
func NewProvider(
	pool *pgxpool.Pool,
	issuerReader domain.IssuerReader,
	eventPublisher domain.EventPublisher,
	logger log.Logger,
) *Provider {
	// Infrastructure
	repository := postgres.NewRepository(pool)

	// The repository implements both SchemaRepository and SchemaLookup
	schemaLookup := repository

	// Application - Command
	commandHandlers := command.NewHandlers(
		repository,
		schemaLookup,
		issuerReader,
		eventPublisher,
		logger,
	)

	// Application - Query
	queryHandlers := query.NewHandlers(
		repository,
		schemaLookup,
		logger,
	)

	// Interface - HTTP
	httpHandler := v1.NewHandler(commandHandlers, queryHandlers)

	return &Provider{
		Repository:      repository,
		SchemaLookup:    schemaLookup,
		CommandHandlers: commandHandlers,
		QueryHandlers:   queryHandlers,
		HTTPHandler:     httpHandler,
	}
}

// ============================================================================
// Route Registration
// ============================================================================

// RegisterRoutes registers all Schema HTTP routes on the given router.
func (p *Provider) RegisterRoutes(r chi.Router) {
	p.HTTPHandler.RegisterRoutes(r)
}

// ============================================================================
// Optional Dependencies
// ============================================================================

// NewProviderWithDefaults creates a Provider with default/null implementations
// for optional dependencies. Useful for testing or standalone operation.
func NewProviderWithDefaults(pool *pgxpool.Pool, logger log.Logger) *Provider {
	return NewProvider(
		pool,
		domain.NewNullIssuerReader(),
		domain.NewNullEventPublisher(),
		logger,
	)
}
