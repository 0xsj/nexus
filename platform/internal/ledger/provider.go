package ledger

import (
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/0xsj/nexus/platform/internal/ledger/app/projection"
	"github.com/0xsj/nexus/platform/internal/ledger/app/query"
	"github.com/0xsj/nexus/platform/internal/ledger/domain"
	"github.com/0xsj/nexus/platform/internal/ledger/infrastructure/persistence/postgres"
	generated "github.com/0xsj/nexus/platform/internal/ledger/infrastructure/persistence/postgres/generated"
	v1 "github.com/0xsj/nexus/platform/internal/ledger/interface/http/v1"
)

// ============================================================================
// Provider
// ============================================================================

// Provider holds all dependencies for the Ledger bounded context.
type Provider struct {
	// Infrastructure
	Writer domain.Writer
	Reader query.Reader

	// Application
	QueryHandlers     *query.Handlers
	ProjectionHandler *projection.Handler

	// Interface
	HTTPHandler *v1.Handler
}

// ============================================================================
// Constructor
// ============================================================================

// NewProvider creates a new Ledger provider with all dependencies wired.
func NewProvider(pool *pgxpool.Pool, extractor domain.MetadataExtractor) *Provider {
	// Infrastructure
	queries := generated.New(pool)
	writer := postgres.NewWriter(queries)
	reader := postgres.NewReader(queries)

	// Application - Query
	queryHandlers := query.NewHandlers(reader)

	// Application - Projection
	mapper := projection.NewMapper()
	projectionHandler := projection.NewHandler(writer, mapper, extractor)

	// Interface - HTTP
	httpHandler := v1.NewHandler(queryHandlers)

	return &Provider{
		Writer:            writer,
		Reader:            reader,
		QueryHandlers:     queryHandlers,
		ProjectionHandler: projectionHandler,
		HTTPHandler:       httpHandler,
	}
}

// ============================================================================
// Route Registration
// ============================================================================

// RegisterRoutes registers all Ledger HTTP routes on the given router.
func (p *Provider) RegisterRoutes(r chi.Router) {
	p.HTTPHandler.RegisterRoutes(r)
}

// ============================================================================
// Event Subscription
// ============================================================================

// SubscribeToEvents sets up the projection handler to receive domain events.
func (p *Provider) SubscribeToEvents(subscriber domain.EventSubscriber) error {
	return p.ProjectionHandler.Subscribe(nil, subscriber)
}
