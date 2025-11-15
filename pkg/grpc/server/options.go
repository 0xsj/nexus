package server

import (
	"github.com/0xsj/nexus/pkg/grpc/interceptors"
	"github.com/0xsj/nexus/pkg/observability/logger"
	"google.golang.org/grpc"
)

// Options builds common gRPC server options with interceptors
func Options(log logger.Logger) []grpc.ServerOption {
	return []grpc.ServerOption{
		// Unary interceptors (executed in order)
		grpc.ChainUnaryInterceptor(
			interceptors.RecoveryUnaryInterceptor(log),
			interceptors.LoggingUnaryInterceptor(log),
			// Add auth interceptor here later
		),
		// Stream interceptors (executed in order)
		grpc.ChainStreamInterceptor(
			interceptors.RecoveryStreamInterceptor(log),
			interceptors.LoggingStreamInterceptor(log),
			// Add auth interceptor here later
		),
	}
}
