package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/0xsj/nexus/platform/internal/credential"
	"github.com/0xsj/nexus/platform/internal/identity"
	identityeventbus "github.com/0xsj/nexus/platform/internal/identity/infrastructure/eventbus"
	"github.com/0xsj/nexus/platform/internal/identity/infrastructure/stubs"
	"github.com/0xsj/nexus/platform/internal/integration"
	"github.com/0xsj/nexus/platform/internal/issuer"
	"github.com/0xsj/nexus/platform/internal/ledger"
	ledgereventbus "github.com/0xsj/nexus/platform/internal/ledger/infrastructure/eventbus"
	"github.com/0xsj/nexus/platform/internal/notification"
	"github.com/0xsj/nexus/platform/internal/organization"
	"github.com/0xsj/nexus/platform/internal/presentation"
	"github.com/0xsj/nexus/platform/internal/profile"
	"github.com/0xsj/nexus/platform/internal/schema"
	schemadomain "github.com/0xsj/nexus/platform/internal/schema/domain"
	schemaeventbus "github.com/0xsj/nexus/platform/internal/schema/infrastructure/eventbus"
	"github.com/0xsj/nexus/platform/internal/trust"
	"github.com/0xsj/nexus/platform/internal/verification"
	"github.com/0xsj/nexus/platform/internal/wallet"
	"github.com/0xsj/nexus/platform/pkg/database"
	"github.com/0xsj/nexus/platform/pkg/database/postgres"
	"github.com/0xsj/nexus/platform/pkg/eventbus"
	"github.com/0xsj/nexus/platform/pkg/eventbus/memory"
	pkghttp "github.com/0xsj/nexus/platform/pkg/http"
	"github.com/0xsj/nexus/platform/pkg/http/middleware"
	"github.com/0xsj/nexus/platform/pkg/observability"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()

	// ========================================================================
	// Configuration
	// ========================================================================

	port := envInt("PORT", 8080)
	dbHost := envStr("DATABASE_HOST", "localhost")
	dbPort := envInt("DATABASE_PORT", 5432)
	dbName := envStr("DATABASE_NAME", "nexus")
	dbUser := envStr("DATABASE_USER", "nexus")
	dbPassword := envStr("DATABASE_PASSWORD", "nexus")
	logLevel := envStr("LOG_LEVEL", "debug")
	appEnv := envStr("APP_ENV", "development")

	// ========================================================================
	// Observability
	// ========================================================================

	var obs *observability.Provider
	if appEnv == "production" {
		obs = observability.NewProduction()
	} else {
		obs = observability.NewDevelopment()
	}
	_ = logLevel // Used by observability config when customized

	logger := obs.Logger()
	logger.Info("starting nexus platform",
		log.String("env", appEnv),
		log.Int("port", port),
	)

	// ========================================================================
	// Database
	// ========================================================================

	pgConfig := postgres.NewConfig(database.Config{
		Host:         dbHost,
		Port:         dbPort,
		Database:     dbName,
		Username:     dbUser,
		Password:     dbPassword,
		MaxOpenConns: 25,
		MaxIdleConns: 5,
	})

	db, err := postgres.New(ctx, pgConfig)
	if err != nil {
		return fmt.Errorf("connecting to database: %w", err)
	}
	defer db.Close()

	pool := db.Pool()
	logger.Info("database connected",
		log.String("host", dbHost),
		log.Int("port", dbPort),
		log.String("database", dbName),
	)

	// ========================================================================
	// Event Bus
	// ========================================================================

	bus := memory.New(eventbus.Config{
		Logger: obs.ComponentLogger("eventbus"),
	})
	if err := bus.Start(ctx); err != nil {
		return fmt.Errorf("starting event bus: %w", err)
	}
	defer bus.Close()

	// ========================================================================
	// Bounded Contexts
	// ========================================================================

	// Identity
	identityPublisher := identityeventbus.NewAdapter(bus)
	identityProvider := identity.NewProvider(identity.ProviderConfig{
		Pool:         pool,
		DIDGenerator: stubs.NewNullDIDGenerator(),
		WalletReader: stubs.NewNullWalletReader(),
		OAuthService: stubs.NewNullOAuthService(),
		EmailService: stubs.NewNullEmailService(obs.ComponentLogger("email")),
		Publisher:    identityPublisher,
		Logger:       obs.ComponentLogger("identity"),
	})

	// Schema
	schemaPublisher := schemaeventbus.NewAdapter(bus)
	schemaProvider := schema.NewProvider(
		pool,
		schemadomain.NewNullIssuerReader(),
		schemaPublisher,
		obs.ComponentLogger("schema"),
	)

	// Credential
	credentialProvider := credential.NewProviderWithDefaults(pool, obs.ComponentLogger("credential"))

	// Verification
	verificationProvider := verification.NewProviderWithDefaults(pool, obs.ComponentLogger("verification"))

	// Wallet
	walletProvider := wallet.NewProviderWithDefaults(pool, obs.ComponentLogger("wallet"))

	// Organization
	organizationProvider := organization.NewProviderWithDefaults(pool, obs.ComponentLogger("organization"))

	// Profile
	profileProvider := profile.NewProviderWithDefaults(pool, obs.ComponentLogger("profile"))

	// Notification
	notificationProvider := notification.NewProviderWithDefaults(pool, obs.ComponentLogger("notification"))

	// Presentation
	presentationProvider := presentation.NewProviderWithDefaults(pool, obs.ComponentLogger("presentation"))

	// Trust
	trustProvider := trust.NewProviderWithDefaults(pool, obs.ComponentLogger("trust"))

	// Integration
	integrationProvider := integration.NewProviderWithDefaults(pool, obs.ComponentLogger("integration"))

	// Issuer
	issuerProvider := issuer.NewProviderWithDefaults(pool, obs.ComponentLogger("issuer"))

	// Ledger
	metadataExtractor := ledgereventbus.NewMetadataExtractor()
	ledgerProvider := ledger.NewProvider(pool, metadataExtractor)

	// Subscribe ledger to domain events via eventbus adapter
	ledgerSubscriber := ledgereventbus.NewAdapter(bus)
	if err := ledgerProvider.SubscribeToEvents(ledgerSubscriber); err != nil {
		return fmt.Errorf("subscribing ledger to events: %w", err)
	}

	// ========================================================================
	// HTTP Router
	// ========================================================================

	router := chi.NewRouter()

	// Global middleware
	router.Use(middleware.RequestID())
	router.Use(middleware.Logger(obs.ComponentLogger("http")))
	router.Use(middleware.Recovery())

	// Auth middleware (validates session tokens via Identity context)
	authMiddleware := newAuthMiddleware(identityProvider.SessionLookup, obs.ComponentLogger("auth"))

	// Routes
	router.Route("/api/v1", func(r chi.Router) {
		identityProvider.RegisterRoutes(r, authMiddleware)
	})
	schemaProvider.RegisterRoutes(router)
	credentialProvider.RegisterRoutes(router)
	verificationProvider.RegisterRoutes(router)
	walletProvider.RegisterRoutes(router)
	organizationProvider.RegisterRoutes(router)
	profileProvider.RegisterRoutes(router)
	notificationProvider.RegisterRoutes(router)
	presentationProvider.RegisterRoutes(router)
	trustProvider.RegisterRoutes(router)
	integrationProvider.RegisterRoutes(router)
	issuerProvider.RegisterRoutes(router)
	ledgerProvider.RegisterRoutes(router)

	logger.Info("routes registered")

	// ========================================================================
	// HTTP Server
	// ========================================================================

	serverConfig := pkghttp.DefaultServerConfig()
	serverConfig.Port = port
	serverConfig.Logger = logger

	server := pkghttp.NewServer(router, serverConfig)

	logger.Info("server starting", log.String("address", server.Address()))
	return server.ListenAndServe()
}

// ============================================================================
// Environment Helpers
// ============================================================================

func envStr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
