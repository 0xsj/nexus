package main

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib" // Register pgx driver for database/sql

	"github.com/0xsj/nexus/platform/internal/credential"
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
	// buildTime = "unknown"
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

	// Initialize database (optional based on environment)
	var stdDB *sql.DB

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
		stdDB, err = openStdlibDB(dbConfig)
		if err != nil {
			logger.Fatal("failed to connect to database", log.Err(err))
		}
		defer stdDB.Close()

		// Register database health check
		checker.Register("postgres", health.DatabaseWithStats(stdDB))

		logger.Info("connected to PostgreSQL",
			log.String("host", dbConfig.Host),
			log.Int("port", dbConfig.Port),
			log.String("database", dbConfig.Database),
		)
	} else {
		logger.Info("running without database (in-memory mode)")
	}

	// Initialize credential module
	credentialConfig := credential.ModuleConfig{
		EnableSigning: true,
		Database:      stdDB, // nil = in-memory, non-nil = PostgreSQL
	}

	credentialModule, err := credential.NewModuleWithConfig(logger, credentialConfig)
	if err != nil {
		logger.Fatal("failed to initialize credential module", log.Err(err))
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
		middleware.CORSAllowAll(), // TODO: Configure for production
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
			r.Mount("/", credentialModule.Routes())
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

	storageMode := "in-memory"
	if stdDB != nil {
		storageMode = "PostgreSQL"
	}

	logger.Info("application ready",
		log.Int("port", port),
		log.String("version", version),
		log.String("storage", storageMode),
		log.String("health", "/health"),
		log.String("liveness", "/healthz"),
		log.String("readiness", "/readyz"),
		log.String("credentials", "/api/v1/credentials"),
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

// shouldConnectDatabase determines if we should connect to the database.
func shouldConnectDatabase() bool {
	return config.GetEnv("DATABASE_HOST", "") != ""
}

// buildDatabaseConfig builds database configuration from environment variables.
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

// openStdlibDB opens a database/sql connection using pgx stdlib driver.
func openStdlibDB(cfg postgres.Config) (*sql.DB, error) {
	db, err := sql.Open("pgx", cfg.DSN())
	if err != nil {
		return nil, err
	}

	// Configure pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime())
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime())

	// Verify connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

// ============================================================================
// Handlers
// ============================================================================

// handleRoot handles the root endpoint.
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
