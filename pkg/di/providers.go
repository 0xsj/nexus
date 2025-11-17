package di

import (
	"github.com/0xsj/nexus/pkg/events"
	"github.com/0xsj/nexus/pkg/events/memory"
	grpcinterceptors "github.com/0xsj/nexus/pkg/grpc/interceptors"
	"github.com/0xsj/nexus/pkg/observability/logger"
	zaplogger "github.com/0xsj/nexus/pkg/observability/logger/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

// ProvideLoggerConfig provides logger configuration.
func ProvideLoggerConfig() logger.Config {
	return logger.NewDevelopmentConfig()
}

// ProvideLogger provides a logger instance.
func ProvideLogger(cfg logger.Config) (logger.Logger, error) {
	return zaplogger.NewZapLogger(cfg)
}

// ProvideEventBus provides an in-memory event bus instance.
func ProvideEventBus(log logger.Logger) events.EventBus {
	log.Info("Using in-memory event bus")
	return memory.New(log)
}

// ProvideGRPCServer provides a configured gRPC server.
func ProvideGRPCServer(log logger.Logger) *grpc.Server {
	// Create server with interceptors
	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			grpcinterceptors.LoggingUnaryInterceptor(log),
			grpcinterceptors.RecoveryUnaryInterceptor(log),
		),
		grpc.ChainStreamInterceptor(
			grpcinterceptors.LoggingStreamInterceptor(log),
			grpcinterceptors.RecoveryStreamInterceptor(log),
		),
	}

	server := grpc.NewServer(opts...)

	// Register health check service
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(server, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	// Register reflection service (for grpcurl)
	reflection.Register(server)

	log.Info("gRPC server created with interceptors")

	return server
}
