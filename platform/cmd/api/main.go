package main

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/0xsj/nexus/platform/internal/credential"
	"github.com/0xsj/nexus/platform/internal/identity"
	"github.com/0xsj/nexus/platform/pkg/config"
	"github.com/0xsj/nexus/platform/pkg/database"
	"github.com/0xsj/nexus/platform/pkg/database/postgres"
	pkghttp "github.com/0xsj/nexus/platform/pkg/http"
	"github.com/0xsj/nexus/platform/pkg/http/middleware"
	"github.com/0xsj/nexus/platform/pkg/http/response"
	"github.com/0xsj/nexus/platform/pkg/observability"
	"github.com/0xsj/nexus/platform/pkg/observability/health"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

var (
	version = "0.1.0"
)

func main() {
	// Load .env file if exists
	_ = config.LoadEnvFileIfExists(".env")

	// Initialize observability provider
	obs := observability.NewBuilder().
		WithLogLevel(config.GetEnv("LOG_LEVEL", "debug")).
		WithColorize(true).
		WithCaller(true).
		WithComponent("api-server").
		Build()

	// Get logger
	logger := obs.Logger()

	// Start observability
	ctx := context.Background()
	if err := obs.Start(ctx); err != nil {
		logger.Fatal("failed to start observability", log.Err(err))
	}

	// Initialize health checker
	checker := health.NewChecker().
		WithVersion(version).
		WithDefaultTimeout(5 * time.Second)

	// Register health checks
	checker.Register("self", health.AlwaysUp())

	// Initialize database
	var db *postgres.DB

	if shouldConnectDatabase() {
		logger.Info("connecting to PostgreSQL...")

		dbConfig := buildDatabaseConfig()
		logger.Debug("database config",
			log.String("host", dbConfig.Host),
			log.Int("port", dbConfig.Port),
			log.String("database", dbConfig.Database),
			log.String("user", dbConfig.Username),
		)

		var err error
		db, err = postgres.New(ctx, dbConfig)
		if err != nil {
			logger.Fatal("failed to connect to database", log.Err(err))
		}
		defer db.Close()

		// Register database health check
		checker.Register("postgres", postgresHealthCheck(db))

		logger.Info("connected to PostgreSQL",
			log.String("host", dbConfig.Host),
			log.Int("port", dbConfig.Port),
			log.String("database", dbConfig.Database),
		)
	} else {
		logger.Info("running without database")
	}

	// Initialize credential module
	credentialConfig := credential.ModuleConfig{
		EnableSigning: true,
		Database:      db,
	}

	credentialModule, err := credential.NewModuleWithConfig(logger, credentialConfig)
	if err != nil {
		logger.Fatal("failed to initialize credential module", log.Err(err))
	}

	// Initialize identity module (requires database)
	var identityModule *identity.Module
	if db != nil {
		identityConfig := identity.ModuleConfig{
			Database:      db,
			Domain:        config.GetEnv("AUTH_DOMAIN", "proof.io"),
			URI:           config.GetEnv("AUTH_URI", "https://proof.io"),
			TokenIssuer:   config.GetEnv("TOKEN_ISSUER", "https://proof.io"),
			TokenAudience: []string{config.GetEnv("TOKEN_AUDIENCE", "https://proof.io")},
		}

		identityModule, err = identity.NewModuleWithConfig(logger, identityConfig)
		if err != nil {
			logger.Fatal("failed to initialize identity module", log.Err(err))
		}
	}

	// Create health handler
	healthHandler := health.NewHandler(checker)

	// Get port from environment
	port := config.GetEnvInt("PORT", 8090)

	// Create router
	router := chi.NewRouter()

	// Apply global middleware
	router.Use(
		middleware.RequestID(),
		middleware.Recovery(),
		middleware.Logger(logger),
		middleware.CORSAllowAll(),
		middleware.Timeout(30*time.Second),
	)

	// Health routes
	router.Get("/health", healthHandler.Health)
	router.Get("/healthz", healthHandler.Liveness)
	router.Get("/readyz", healthHandler.Readiness)
	router.Get("/startupz", healthHandler.Startup)

	// API routes
	router.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			r.Get("/", handleRoot(logger))

			// Mount credential routes
			r.Mount("/credentials", credentialModule.Routes())

			// Mount identity routes (if database available)
			if identityModule != nil {
				r.Mount("/", identityModule.Routes())
			}
		})
	})

	// Create server config
	serverConfig := pkghttp.ServerConfig{
		Port:            port,
		ReadTimeout:     config.GetEnvDuration("READ_TIMEOUT", 15*time.Second),
		WriteTimeout:    config.GetEnvDuration("WRITE_TIMEOUT", 15*time.Second),
		IdleTimeout:     config.GetEnvDuration("IDLE_TIMEOUT", 60*time.Second),
		ShutdownTimeout: config.GetEnvDuration("SHUTDOWN_TIMEOUT", 30*time.Second),
		Logger:          logger,
	}

	// Create server
	server := pkghttp.NewServer(router, serverConfig)

	// Mark as started
	checker.MarkStarted()

	storageMode := "none"
	if db != nil {
		storageMode = "PostgreSQL"
	}

	logger.Info("application ready",
		log.Int("port", port),
		log.String("version", version),
		log.String("storage", storageMode),
		log.String("health", "/health"),
		log.String("credentials", "/api/v1/credentials"),
		log.String("auth", "/api/v1/auth"),
	)

	// Start server with graceful shutdown
	if err := server.ListenAndServe(); err != nil {
		logger.Error("server error", log.Err(err))
	}

	// Stop observability
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := obs.Stop(shutdownCtx); err != nil {
		logger.Error("observability shutdown error", log.Err(err))
	}

	logger.Info("server stopped gracefully")
}

// ============================================================================
// Database Configuration
// ============================================================================

func shouldConnectDatabase() bool {
	return config.GetEnv("DATABASE_HOST", "") != ""
}

func buildDatabaseConfig() postgres.Config {
	baseConfig := database.DefaultConfig().
		WithHost(config.GetEnv("DATABASE_HOST", "localhost")).
		WithPort(config.GetEnvInt("DATABASE_PORT", 5432)).
		WithDatabase(config.GetEnv("DATABASE_NAME", "nexus")).
		WithCredentials(
			config.GetEnv("DATABASE_USER", "nexus"),
			config.GetEnv("DATABASE_PASSWORD", "nexus"),
		).
		WithSSLMode(config.GetEnv("DATABASE_SSLMODE", "disable"))

	return postgres.NewConfig(baseConfig).
		WithApplicationName("nexus-api")
}

// postgresHealthCheck creates a health check for postgres.DB.
func postgresHealthCheck(db *postgres.DB) health.Check {
	return func(ctx context.Context) *health.Result {
		start := time.Now()

		status := db.HealthCheck(ctx)
		duration := time.Since(start)

		if status.Healthy {
			return health.Up().
				WithDuration(duration).
				WithMessage(status.Message).
				WithDetail("latency_ms", status.LatencyMs).
				WithDetail("open_connections", status.Stats.OpenConnections).
				WithDetail("in_use", status.Stats.InUse).
				WithDetail("idle", status.Stats.Idle)
		}

		return health.Down(nil).
			WithDuration(duration).
			WithMessage(status.Message)
	}
}

// ============================================================================
// Handlers
// ============================================================================

func handleRoot(logger log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := middleware.GetRequestID(r.Context())
		logger.Debug("root endpoint",
			log.String("request_id", requestID),
			log.String("remote_addr", r.RemoteAddr),
		)

		response.OK(w, map[string]string{
			"service": "nexus",
			"status":  "running",
			"version": version,
		})
	}
}
