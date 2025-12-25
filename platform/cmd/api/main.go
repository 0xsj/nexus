package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/0xsj/nexus/platform/pkg/observability"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

func main() {
	// Initialize observability provider
	obs := observability.NewBuilder().
		WithLogLevel("debug").
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

	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handleHealth(logger))
	mux.HandleFunc("GET /ready", handleReady(logger))

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info("starting server",
			log.String("port", port),
			log.String("addr", server.Addr),
		)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server error", log.Err(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	logger.Info("received shutdown signal",
		log.String("signal", sig.String()),
	)

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	logger.Info("shutting down server")

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("server shutdown error", log.Err(err))
		os.Exit(1)
	}

	// Stop observability
	if err := obs.Stop(shutdownCtx); err != nil {
		logger.Error("observability shutdown error", log.Err(err))
	}

	logger.Info("server stopped gracefully")
}

func handleHealth(logger log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("health check requested",
			log.String("remote_addr", r.RemoteAddr),
			log.String("user_agent", r.UserAgent()),
		)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status":"healthy"}`)
	}
}

func handleReady(logger log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("readiness check requested",
			log.String("remote_addr", r.RemoteAddr),
		)

		// TODO: Check dependencies (database, etc.)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status":"ready"}`)
	}
}
