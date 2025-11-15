package interceptors

import (
	"context"
	"time"

	grpcmetadata "github.com/0xsj/nexus/pkg/grpc/metadata"
	"github.com/0xsj/nexus/pkg/observability/logger"
	"google.golang.org/grpc"
)

// LoggingUnaryInterceptor logs all unary RPC calls with metadata
func LoggingUnaryInterceptor(log logger.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		// Extract metadata from gRPC headers
		requestID := grpcmetadata.ExtractRequestID(ctx)
		traceID := grpcmetadata.ExtractTraceID(ctx)

		// Log request
		log.Debug("gRPC request started",
			logger.String("method", info.FullMethod),
			logger.String("request_id", requestID),
			logger.String("trace_id", traceID),
		)

		// Call handler
		resp, err := handler(ctx, req)

		// Log response
		duration := time.Since(start)
		if err != nil {
			log.Error("gRPC request failed",
				logger.String("method", info.FullMethod),
				logger.String("request_id", requestID),
				logger.String("trace_id", traceID),
				logger.Duration("duration", duration),
				logger.Err(err),
			)
		} else {
			log.Info("gRPC request completed",
				logger.String("method", info.FullMethod),
				logger.String("request_id", requestID),
				logger.String("trace_id", traceID),
				logger.Duration("duration", duration),
			)
		}

		return resp, err
	}
}

// LoggingStreamInterceptor logs all streaming RPC calls with metadata
func LoggingStreamInterceptor(log logger.Logger) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		start := time.Now()

		ctx := ss.Context()
		requestID := grpcmetadata.ExtractRequestID(ctx)
		traceID := grpcmetadata.ExtractTraceID(ctx)

		log.Debug("gRPC stream started",
			logger.String("method", info.FullMethod),
			logger.String("request_id", requestID),
			logger.String("trace_id", traceID),
			logger.Bool("is_client_stream", info.IsClientStream),
			logger.Bool("is_server_stream", info.IsServerStream),
		)

		err := handler(srv, ss)

		duration := time.Since(start)
		if err != nil {
			log.Error("gRPC stream failed",
				logger.String("method", info.FullMethod),
				logger.String("request_id", requestID),
				logger.String("trace_id", traceID),
				logger.Duration("duration", duration),
				logger.Err(err),
			)
		} else {
			log.Info("gRPC stream completed",
				logger.String("method", info.FullMethod),
				logger.String("request_id", requestID),
				logger.String("trace_id", traceID),
				logger.Duration("duration", duration),
			)
		}

		return err
	}
}
