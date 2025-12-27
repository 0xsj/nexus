package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// ============================================================================
// Server Configuration
// ============================================================================

// ServerConfig configures the HTTP server.
type ServerConfig struct {
	// Host is the host to bind to.
	// Default: "" (all interfaces)
	Host string

	// Port is the port to bind to.
	// Default: 8080
	Port int

	// ReadTimeout is the maximum duration for reading the entire request.
	// Default: 15 seconds
	ReadTimeout time.Duration

	// WriteTimeout is the maximum duration for writing the response.
	// Default: 15 seconds
	WriteTimeout time.Duration

	// IdleTimeout is the maximum duration to wait for the next request.
	// Default: 60 seconds
	IdleTimeout time.Duration

	// ReadHeaderTimeout is the maximum duration for reading request headers.
	// Default: 5 seconds
	ReadHeaderTimeout time.Duration

	// MaxHeaderBytes is the maximum size of request headers.
	// Default: 1MB
	MaxHeaderBytes int

	// ShutdownTimeout is the maximum duration to wait for graceful shutdown.
	// Default: 30 seconds
	ShutdownTimeout time.Duration

	// Logger is the logger to use.
	// Default: nil (no logging)
	Logger log.Logger
}

// DefaultServerConfig returns the default server configuration.
func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		Host:              "",
		Port:              8080,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1MB
		ShutdownTimeout:   30 * time.Second,
		Logger:            nil,
	}
}

// Address returns the server address as "host:port".
func (c ServerConfig) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// ============================================================================
// Server
// ============================================================================

// Server wraps http.Server with graceful shutdown support.
type Server struct {
	server *http.Server
	config ServerConfig
	logger log.Logger
}

// NewServer creates a new HTTP server.
func NewServer(handler http.Handler, config ServerConfig) *Server {
	if config.Port == 0 {
		config.Port = 8080
	}
	if config.ReadTimeout == 0 {
		config.ReadTimeout = 15 * time.Second
	}
	if config.WriteTimeout == 0 {
		config.WriteTimeout = 15 * time.Second
	}
	if config.IdleTimeout == 0 {
		config.IdleTimeout = 60 * time.Second
	}
	if config.ReadHeaderTimeout == 0 {
		config.ReadHeaderTimeout = 5 * time.Second
	}
	if config.MaxHeaderBytes == 0 {
		config.MaxHeaderBytes = 1 << 20
	}
	if config.ShutdownTimeout == 0 {
		config.ShutdownTimeout = 30 * time.Second
	}

	server := &http.Server{
		Addr:              config.Address(),
		Handler:           handler,
		ReadTimeout:       config.ReadTimeout,
		WriteTimeout:      config.WriteTimeout,
		IdleTimeout:       config.IdleTimeout,
		ReadHeaderTimeout: config.ReadHeaderTimeout,
		MaxHeaderBytes:    config.MaxHeaderBytes,
	}

	return &Server{
		server: server,
		config: config,
		logger: config.Logger,
	}
}

// NewServerWithDefaults creates a new HTTP server with default configuration.
func NewServerWithDefaults(handler http.Handler, port int) *Server {
	config := DefaultServerConfig()
	config.Port = port
	return NewServer(handler, config)
}

// ============================================================================
// Server Lifecycle
// ============================================================================

// Start starts the server and blocks until it's stopped.
func (s *Server) Start() error {
	s.log("info", "starting HTTP server", "address", s.config.Address())

	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}

// StartTLS starts the server with TLS and blocks until it's stopped.
func (s *Server) StartTLS(certFile, keyFile string) error {
	s.log("info", "starting HTTPS server", "address", s.config.Address())

	if err := s.server.ListenAndServeTLS(certFile, keyFile); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown(ctx context.Context) error {
	s.log("info", "shutting down HTTP server")

	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown error: %w", err)
	}

	s.log("info", "HTTP server stopped")
	return nil
}

// Close immediately closes the server without waiting for connections.
func (s *Server) Close() error {
	return s.server.Close()
}

// ============================================================================
// Graceful Shutdown with Signal Handling
// ============================================================================

// ListenAndServe starts the server and handles graceful shutdown on SIGINT/SIGTERM.
func (s *Server) ListenAndServe() error {
	// Channel to receive shutdown signal
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Channel to receive server errors
	serverErr := make(chan error, 1)

	// Start server in goroutine
	go func() {
		serverErr <- s.Start()
	}()

	// Wait for shutdown signal or server error
	select {
	case err := <-serverErr:
		return err
	case sig := <-shutdown:
		s.log("info", "received shutdown signal", "signal", sig.String())
	}

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), s.config.ShutdownTimeout)
	defer cancel()

	// Attempt graceful shutdown
	if err := s.Shutdown(ctx); err != nil {
		s.log("error", "graceful shutdown failed, forcing close", "error", err.Error())
		return s.Close()
	}

	return nil
}

// ListenAndServeTLS starts the HTTPS server with graceful shutdown.
func (s *Server) ListenAndServeTLS(certFile, keyFile string) error {
	// Channel to receive shutdown signal
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Channel to receive server errors
	serverErr := make(chan error, 1)

	// Start server in goroutine
	go func() {
		serverErr <- s.StartTLS(certFile, keyFile)
	}()

	// Wait for shutdown signal or server error
	select {
	case err := <-serverErr:
		return err
	case sig := <-shutdown:
		s.log("info", "received shutdown signal", "signal", sig.String())
	}

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), s.config.ShutdownTimeout)
	defer cancel()

	// Attempt graceful shutdown
	if err := s.Shutdown(ctx); err != nil {
		s.log("error", "graceful shutdown failed, forcing close", "error", err.Error())
		return s.Close()
	}

	return nil
}

// ============================================================================
// Logging Helper
// ============================================================================

// log logs a message if a logger is configured.
func (s *Server) log(level, msg string, args ...any) {
	if s.logger == nil {
		return
	}

	fields := make([]log.Field, 0, len(args)/2)
	for i := 0; i < len(args)-1; i += 2 {
		key, ok := args[i].(string)
		if !ok {
			continue
		}
		fields = append(fields, log.Any(key, args[i+1]))
	}

	switch level {
	case "info":
		s.logger.Info(msg, fields...)
	case "warn":
		s.logger.Warn(msg, fields...)
	case "error":
		s.logger.Error(msg, fields...)
	case "debug":
		s.logger.Debug(msg, fields...)
	}
}

// ============================================================================
// Accessors
// ============================================================================

// Address returns the server address.
func (s *Server) Address() string {
	return s.config.Address()
}

// Config returns the server configuration.
func (s *Server) Config() ServerConfig {
	return s.config
}
