package identity

import (
	"github.com/go-chi/chi/v5"
	"github.com/google/wire"

	"github.com/0xsj/nexus/platform/internal/identity/application/command"
	"github.com/0xsj/nexus/platform/internal/identity/application/query"
	"github.com/0xsj/nexus/platform/internal/identity/domain"
	"github.com/0xsj/nexus/platform/internal/identity/infrastructure/challenge"
	"github.com/0xsj/nexus/platform/internal/identity/infrastructure/persistence/postgres"
	"github.com/0xsj/nexus/platform/internal/identity/infrastructure/signature"
	"github.com/0xsj/nexus/platform/internal/identity/infrastructure/token"
	httpv1 "github.com/0xsj/nexus/platform/internal/identity/interface/http/v1"
	"github.com/0xsj/nexus/platform/pkg/cqrs"
	"github.com/0xsj/nexus/platform/pkg/crypto/ed25519"
	pgadapter "github.com/0xsj/nexus/platform/pkg/database/postgres"
	"github.com/0xsj/nexus/platform/pkg/id"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// ============================================================================
// Wire Provider Set
// ============================================================================

// ProviderSet is the Wire provider set for the identity module.
var ProviderSet = wire.NewSet(
	NewModule,
)

// ============================================================================
// Module Configuration
// ============================================================================

// ModuleConfig holds configuration for the identity module.
type ModuleConfig struct {
	// Database is the PostgreSQL database connection.
	// Required - identity module requires persistence.
	Database *pgadapter.DB

	// Domain is the domain for SIWE challenges (e.g., "proof.io")
	Domain string

	// URI is the URI for SIWE challenges (e.g., "https://proof.io")
	URI string

	// TokenIssuer is the JWT token issuer
	TokenIssuer string

	// TokenAudience is the JWT token audience
	TokenAudience []string

	// PrivateKey is the Ed25519 private key bytes for signing tokens.
	// If empty, a new key pair will be generated.
	PrivateKey []byte
}

// DefaultModuleConfig returns default configuration.
func DefaultModuleConfig() ModuleConfig {
	return ModuleConfig{
		Domain:        "proof.io",
		URI:           "https://proof.io",
		TokenIssuer:   "https://proof.io",
		TokenAudience: []string{"https://proof.io"},
	}
}

// ============================================================================
// Module
// ============================================================================

// Module is the identity module that provides all identity functionality.
type Module struct {
	router chi.Router

	// Repositories
	userRepo       domain.UserRepository
	sessionRepo    domain.SessionRepository
	apiKeyRepo     domain.APIKeyRepository
	connectionRepo domain.ConnectionRepository

	// Services
	tokenService      domain.TokenService
	challengeService  domain.ChallengeService
	signatureVerifier domain.SignatureVerifier

	// CQRS
	commandBus *cqrs.InMemoryCommandBus
	queryBus   *cqrs.InMemoryQueryBus

	logger log.Logger
}

// NewModule creates a new identity module with default configuration.
func NewModule(logger log.Logger) (*Module, error) {
	return NewModuleWithConfig(logger, DefaultModuleConfig())
}

// NewModuleWithConfig creates a new identity module with custom configuration.
func NewModuleWithConfig(logger log.Logger, cfg ModuleConfig) (*Module, error) {
	if cfg.Database == nil {
		return nil, ErrMissingDependency("Database")
	}

	// Create database adapter
	adapter := pgadapter.NewBaseAdapter(cfg.Database)

	// Create repositories
	userRepo := postgres.NewUserRepository(adapter)
	sessionRepo := postgres.NewSessionRepository(adapter)
	apiKeyRepo := postgres.NewAPIKeyRepository(adapter)
	connectionRepo := postgres.NewConnectionRepository(adapter)

	// Create key pair for token signing
	keyPair, err := createKeyPair(cfg.PrivateKey, logger)
	if err != nil {
		return nil, err
	}

	// Create signer and verifier
	signer, err := ed25519.NewSigner(keyPair)
	if err != nil {
		return nil, err
	}

	verifier, err := ed25519.NewVerifier(keyPair.PublicKey())
	if err != nil {
		return nil, err
	}

	// Create token service
	tokenConfig := token.Config{
		Issuer:   cfg.TokenIssuer,
		Audience: cfg.TokenAudience,
	}
	tokenService := token.NewService(signer, verifier, tokenConfig)

	// Create challenge service
	challengeConfig := challenge.Config{
		Domain: cfg.Domain,
		URI:    cfg.URI,
	}
	challengeService := challenge.NewService(challengeConfig)

	// Create signature verifier
	signatureConfig := signature.Config{
		Domain: cfg.Domain,
	}
	signatureVerifier := signature.NewVerifier(signatureConfig)

	// Create reader (for queries)
	reader := postgres.NewReader(adapter, tokenService)

	// Create ID generator
	idGenerator := id.NewGenerator()

	// Create CQRS buses
	commandBus := cqrs.NewCommandBus()
	queryBus := cqrs.NewQueryBus()

	// Register command handlers
	cmdDeps := command.HandlerDependencies{
		UserRepo:          userRepo,
		SessionRepo:       sessionRepo,
		APIKeyRepo:        apiKeyRepo,
		ChallengeService:  challengeService,
		TokenService:      tokenService,
		SignatureVerifier: signatureVerifier,
		IDGenerator:       idGenerator,
	}

	if err := command.RegisterHandlers(commandBus, cmdDeps); err != nil {
		return nil, err
	}

	// Register query handlers
	queryDeps := query.HandlerDependencies{
		UserReader:    reader,
		SessionReader: reader,
		APIKeyReader:  reader,
		TokenReader:   reader,
	}

	if err := query.RegisterHandlers(queryBus, queryDeps); err != nil {
		return nil, err
	}

	// Create HTTP router
	router := httpv1.NewRouter(httpv1.RouterConfig{
		CommandBus:   commandBus,
		QueryBus:     queryBus,
		TokenService: tokenService,
		Logger:       logger,
	})

	logger.Info("identity module initialized",
		log.String("storage", "PostgreSQL"),
		log.String("domain", cfg.Domain),
	)

	return &Module{
		router:            router,
		userRepo:          userRepo,
		sessionRepo:       sessionRepo,
		apiKeyRepo:        apiKeyRepo,
		connectionRepo:    connectionRepo,
		tokenService:      tokenService,
		challengeService:  challengeService,
		signatureVerifier: signatureVerifier,
		commandBus:        commandBus,
		queryBus:          queryBus,
		logger:            logger,
	}, nil
}

// createKeyPair creates or loads an Ed25519 key pair.
func createKeyPair(privateKey []byte, logger log.Logger) (*ed25519.KeyPair, error) {
	if len(privateKey) > 0 {
		return ed25519.FromPrivateKey(privateKey)
	}

	keyPair, err := ed25519.Generate()
	if err != nil {
		return nil, err
	}

	logger.Info("generated new Ed25519 key pair for token signing")
	return keyPair, nil
}

// ============================================================================
// Accessors
// ============================================================================

// Routes returns the HTTP routes for mounting.
func (m *Module) Routes() chi.Router {
	return m.router
}

// CommandBus returns the command bus for direct access.
func (m *Module) CommandBus() cqrs.CommandBus {
	return m.commandBus
}

// QueryBus returns the query bus for direct access.
func (m *Module) QueryBus() cqrs.QueryBus {
	return m.queryBus
}

// TokenService returns the token service.
func (m *Module) TokenService() domain.TokenService {
	return m.tokenService
}

// ChallengeService returns the challenge service.
func (m *Module) ChallengeService() domain.ChallengeService {
	return m.challengeService
}

// ============================================================================
// Errors
// ============================================================================

// ErrMissingDependency creates an error for missing dependencies.
func ErrMissingDependency(name string) error {
	return &MissingDependencyError{Name: name}
}

// MissingDependencyError indicates a required dependency is missing.
type MissingDependencyError struct {
	Name string
}

// Error implements error.
func (e *MissingDependencyError) Error() string {
	return "missing required dependency: " + e.Name
}
