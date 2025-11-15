package server

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/0xsj/nexus/pkg/observability/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

// Server wraps a gRPC server with graceful shutdown
type Server struct {
	grpcServer *grpc.Server
	listener   net.Listener
	config     *Config
	logger     logger.Logger
}

// New creates a new gRPC server with the given options
func New(cfg *Config, log logger.Logger, opts ...grpc.ServerOption) (*Server, error) {
	// Add default server options
	defaultOpts := []grpc.ServerOption{
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     cfg.MaxConnectionIdle,
			MaxConnectionAge:      cfg.MaxConnectionAge,
			MaxConnectionAgeGrace: cfg.MaxConnectionAgeGrace,
			Time:                  cfg.KeepAliveTime,
			Timeout:               cfg.KeepAliveTimeout,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             5 * time.Second,
			PermitWithoutStream: true,
		}),
		grpc.MaxRecvMsgSize(cfg.MaxRecvMsgSize),
		grpc.MaxSendMsgSize(cfg.MaxSendMsgSize),
	}

	// Combine with provided options
	allOpts := append(defaultOpts, opts...)

	// Create gRPC server
	grpcServer := grpc.NewServer(allOpts...)

	// Enable reflection if configured
	if cfg.EnableReflection {
		reflection.Register(grpcServer)
		log.Info("gRPC reflection enabled")
	}

	return &Server{
		grpcServer: grpcServer,
		config:     cfg,
		logger:     log,
	}, nil
}

// Server returns the underlying gRPC server for service registration
func (s *Server) Server() *grpc.Server {
	return s.grpcServer
}

// Start starts the gRPC server
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.config.Port)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	s.listener = listener

	s.logger.Info("gRPC server starting",
		logger.String("address", addr),
		logger.Int("port", s.config.Port),
	)

	if err := s.grpcServer.Serve(listener); err != nil {
		return fmt.Errorf("failed to serve gRPC: %w", err)
	}

	return nil
}

// Stop gracefully stops the gRPC server
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Shutting down gRPC server...")

	// Channel to signal when graceful stop is complete
	stopped := make(chan struct{})

	go func() {
		s.grpcServer.GracefulStop()
		close(stopped)
	}()

	// Wait for graceful stop or context timeout
	select {
	case <-stopped:
		s.logger.Info("gRPC server stopped gracefully")
		return nil
	case <-ctx.Done():
		s.logger.Warn("Forcing gRPC server stop due to timeout")
		s.grpcServer.Stop()
		return ctx.Err()
	}
}

// Port returns the port the server is listening on
func (s *Server) Port() int {
	if s.listener != nil {
		return s.listener.Addr().(*net.TCPAddr).Port
	}
	return s.config.Port
}
