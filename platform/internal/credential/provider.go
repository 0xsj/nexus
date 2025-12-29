package credential

import (
	"github.com/go-chi/chi/v5"
	"github.com/google/wire"

	"github.com/0xsj/nexus/platform/internal/credential/application/command"
	"github.com/0xsj/nexus/platform/internal/credential/application/query"
	"github.com/0xsj/nexus/platform/internal/credential/domain"
	"github.com/0xsj/nexus/platform/internal/credential/infrastructure/persistence/memory"
	httpv1 "github.com/0xsj/nexus/platform/internal/credential/interface/http/v1"
	"github.com/0xsj/nexus/platform/pkg/cqrs"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// ============================================================================
// Wire Provider Set
// ============================================================================

// ProviderSet is the Wire provider set for the credential module.
var ProviderSet = wire.NewSet(
	NewModule,
	wire.Bind(new(domain.CredentialRepository), new(*memory.Repository)),
	wire.Bind(new(query.CredentialReadRepository), new(*memory.Repository)),
)

// ============================================================================
// Module
// ============================================================================

// Module is the credential module that provides all credential functionality.
type Module struct {
	router     *httpv1.Router
	repository *memory.Repository
	commandBus *cqrs.InMemoryCommandBus
	queryBus   *cqrs.InMemoryQueryBus
	logger     log.Logger
}

// NewModule creates a new credential module.
func NewModule(logger log.Logger) (*Module, error) {
	// Create repository (implements both write and read interfaces)
	repository := memory.NewRepository()

	// Create CQRS buses
	commandBus := cqrs.NewCommandBus()
	queryBus := cqrs.NewQueryBus()

	// Register command handlers
	if err := command.RegisterHandlers(commandBus, repository); err != nil {
		return nil, err
	}

	// Register query handlers
	if err := query.RegisterHandlers(queryBus, repository); err != nil {
		return nil, err
	}

	// Create HTTP router
	router := httpv1.NewRouter(commandBus, queryBus, logger)

	return &Module{
		router:     router,
		repository: repository,
		commandBus: commandBus,
		queryBus:   queryBus,
		logger:     logger,
	}, nil
}

// ============================================================================
// Accessors
// ============================================================================

// Routes returns the HTTP routes for mounting.
func (m *Module) Routes() chi.Router {
	return m.router.Routes()
}

// CommandBus returns the command bus for direct access if needed.
func (m *Module) CommandBus() cqrs.CommandBus {
	return m.commandBus
}

// QueryBus returns the query bus for direct access if needed.
func (m *Module) QueryBus() cqrs.QueryBus {
	return m.queryBus
}

// Repository returns the repository for testing purposes.
func (m *Module) Repository() *memory.Repository {
	return m.repository
}
