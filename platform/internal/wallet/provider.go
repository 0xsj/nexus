package wallet

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/wire"

	"github.com/0xsj/nexus/platform/internal/wallet/application/command"
	"github.com/0xsj/nexus/platform/internal/wallet/application/query"
	"github.com/0xsj/nexus/platform/internal/wallet/domain"
	"github.com/0xsj/nexus/platform/internal/wallet/infrastructure/challenge"
	infraDID "github.com/0xsj/nexus/platform/internal/wallet/infrastructure/did"
	"github.com/0xsj/nexus/platform/internal/wallet/infrastructure/nonce"
	"github.com/0xsj/nexus/platform/internal/wallet/infrastructure/persistence/postgres"
	"github.com/0xsj/nexus/platform/internal/wallet/infrastructure/signature"
	"github.com/0xsj/nexus/platform/internal/wallet/infrastructure/validation"
	httpv1 "github.com/0xsj/nexus/platform/internal/wallet/interface/http/v1"
	"github.com/0xsj/nexus/platform/pkg/cqrs"
	pgadapter "github.com/0xsj/nexus/platform/pkg/database/postgres"
	"github.com/0xsj/nexus/platform/pkg/id"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// ============================================================================
// Wire Provider Set
// ============================================================================

// ProviderSet is the Wire provider set for the wallet module.
var ProviderSet = wire.NewSet(
	NewModule,
)

// ============================================================================
// Module Configuration
// ============================================================================

// ModuleConfig holds configuration for the wallet module.
type ModuleConfig struct {
	// Database is the PostgreSQL database connection.
	// Required - wallet module requires persistence.
	Database *pgadapter.DB

	// Domain is the domain for SIWE challenges (e.g., "nexus.io")
	Domain string

	// URI is the URI for SIWE challenges (e.g., "https://nexus.io")
	URI string

	// ChallengeTTL is the time-to-live for challenges in seconds.
	// Default: 600 (10 minutes)
	ChallengeTTL int

	// NonceTTL is the time-to-live for nonces in seconds.
	// Default: 600 (10 minutes)
	NonceTTL int

	// AuthMiddleware is the authentication middleware for protected routes.
	// If nil, protected routes will not require authentication (development only).
	AuthMiddleware func(http.Handler) http.Handler
}

// DefaultModuleConfig returns default configuration.
func DefaultModuleConfig() ModuleConfig {
	return ModuleConfig{
		Domain:       "nexus.io",
		URI:          "https://nexus.io",
		ChallengeTTL: 600,
		NonceTTL:     600,
	}
}

// ============================================================================
// Module
// ============================================================================

// Module is the wallet module that provides all wallet functionality.
type Module struct {
	routes *httpv1.Routes

	// Repositories
	walletRepo    domain.WalletRepository
	challengeRepo domain.ChallengeRepository

	// Services
	challengeService  domain.ChallengeService
	nonceService      domain.NonceService
	signatureVerifier domain.SignatureVerificationService
	addressValidator  domain.AddressValidationService
	didDerivation     domain.DIDDerivationService

	// CQRS
	commandBus *cqrs.InMemoryCommandBus
	queryBus   *cqrs.InMemoryQueryBus

	logger log.Logger
}

// NewModule creates a new wallet module with default configuration.
func NewModule(logger log.Logger) (*Module, error) {
	return NewModuleWithConfig(logger, DefaultModuleConfig())
}

// NewModuleWithConfig creates a new wallet module with custom configuration.
func NewModuleWithConfig(logger log.Logger, cfg ModuleConfig) (*Module, error) {
	if cfg.Database == nil {
		return nil, ErrMissingDependency("Database")
	}

	// Create database adapter
	adapter := pgadapter.NewBaseAdapter(cfg.Database)

	// Create repositories
	walletRepo := postgres.NewWalletRepository(adapter)
	challengeRepo := postgres.NewChallengeRepository(adapter)

	// Create wallet reader (for queries)
	walletReader := postgres.NewWalletReader(adapter)
	challengeReader := postgres.NewChallengeReader(adapter)

	// Create ID generator
	idGenerator := id.NewGenerator()

	// Create nonce service (in-memory for now, can switch to postgres-backed)
	nonceConfig := nonce.DefaultInMemoryConfig()
	if cfg.NonceTTL > 0 {
		nonceConfig.TTL = time.Duration(cfg.NonceTTL) * time.Second
	}
	nonceService := nonce.NewInMemoryService(nonceConfig)

	// Create challenge service
	challengeConfig := challenge.DefaultConfig()
	challengeConfig.DefaultStatement = "Sign in with your wallet to Nexus"
	if cfg.ChallengeTTL > 0 {
		challengeConfig.TTL = time.Duration(cfg.ChallengeTTL) * time.Second
	}
	challengeService := challenge.NewService(nonceService, challengeConfig)

	// Create signature verifier
	signatureConfig := signature.DefaultVerifierConfig()
	signatureConfig.Domain = cfg.Domain
	signatureVerifier := signature.NewCompositeSignatureVerifier(signatureConfig)

	// Create address validator
	addressValidator := validation.NewDefaultAddressValidator()

	// Create DID derivation service
	didDerivation := infraDID.NewDerivationService()

	// Create CQRS buses
	commandBus := cqrs.NewCommandBus()
	queryBus := cqrs.NewQueryBus()

	// Register command handlers
	cmdDeps := command.HandlerDependencies{
		WalletRepo:        walletRepo,
		ChallengeRepo:     challengeRepo,
		ChallengeService:  challengeService,
		SignatureVerifier: signatureVerifier,
		AddressValidator:  addressValidator,
		DIDDerivation:     didDerivation,
		IDGenerator:       idGenerator,
	}

	if err := command.RegisterHandlers(commandBus, cmdDeps); err != nil {
		return nil, err
	}

	// Register query handlers
	queryDeps := query.HandlerDependencies{
		WalletReader:    walletReader,
		ChallengeReader: challengeReader,
	}

	if err := query.RegisterHandlers(queryBus, queryDeps); err != nil {
		return nil, err
	}

	// Create HTTP routes
	routerCfg := httpv1.RouterConfig{
		CommandBus:       commandBus,
		QueryBus:         queryBus,
		ChallengeService: challengeService,
		AuthMiddleware:   cfg.AuthMiddleware,
		Logger:           logger,
	}
	routes := httpv1.NewRoutes(routerCfg)

	logger.Info("wallet module initialized",
		log.String("storage", "PostgreSQL"),
		log.String("domain", cfg.Domain),
		log.String("nonce_storage", "in-memory"),
	)

	return &Module{
		routes:            routes,
		walletRepo:        walletRepo,
		challengeRepo:     challengeRepo,
		challengeService:  challengeService,
		nonceService:      nonceService,
		signatureVerifier: signatureVerifier,
		addressValidator:  addressValidator,
		didDerivation:     didDerivation,
		commandBus:        commandBus,
		queryBus:          queryBus,
		logger:            logger,
	}, nil
}

// ============================================================================
// Route Accessors
// ============================================================================

// Routes returns the HTTP routes container for mounting.
func (m *Module) Routes() *httpv1.Routes {
	return m.routes
}

// PublicRoutes returns the public wallet routes.
// Mount at: /api/v1/wallet
func (m *Module) PublicRoutes() chi.Router {
	return m.routes.Public
}

// ProtectedRoutes returns the protected wallet routes.
// Mount at: /api/v1/wallets
func (m *Module) ProtectedRoutes() chi.Router {
	return m.routes.Protected
}

// ============================================================================
// Service Accessors
// ============================================================================

// CommandBus returns the command bus for direct access.
func (m *Module) CommandBus() cqrs.CommandBus {
	return m.commandBus
}

// QueryBus returns the query bus for direct access.
func (m *Module) QueryBus() cqrs.QueryBus {
	return m.queryBus
}

// ChallengeService returns the challenge service.
func (m *Module) ChallengeService() domain.ChallengeService {
	return m.challengeService
}

// NonceService returns the nonce service.
func (m *Module) NonceService() domain.NonceService {
	return m.nonceService
}

// SignatureVerifier returns the signature verifier.
func (m *Module) SignatureVerifier() domain.SignatureVerificationService {
	return m.signatureVerifier
}

// AddressValidator returns the address validator.
func (m *Module) AddressValidator() domain.AddressValidationService {
	return m.addressValidator
}

// DIDDerivation returns the DID derivation service.
func (m *Module) DIDDerivation() domain.DIDDerivationService {
	return m.didDerivation
}

// WalletRepository returns the wallet repository for direct access.
func (m *Module) WalletRepository() domain.WalletRepository {
	return m.walletRepo
}

// ============================================================================
// Lifecycle
// ============================================================================

// Stop gracefully stops the module and releases resources.
func (m *Module) Stop() {
	// Stop the in-memory nonce service cleanup goroutine
	if inMemoryNonce, ok := m.nonceService.(*nonce.InMemoryService); ok {
		inMemoryNonce.Stop()
	}

	m.logger.Info("wallet module stopped")
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
