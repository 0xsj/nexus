package interceptors

import (
	"context"
	"runtime/debug"

	"github.com/0xsj/nexus/pkg/observability/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RecoveryUnaryInterceptor recovers from panics in unary RPCs
func RecoveryUnaryInterceptor(log logger.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				log.Error("Panic recovered in gRPC handler",
					logger.String("method", info.FullMethod),
					logger.Any("panic", r),
					logger.String("stack", string(debug.Stack())),
				)
				err = status.Errorf(codes.Internal, "internal server error: %v", r)
			}
		}()

		return handler(ctx, req)
	}
}

// RecoveryStreamInterceptor recovers from panics in streaming RPCs
func RecoveryStreamInterceptor(log logger.Logger) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) (err error) {
		defer func() {
			if r := recover(); r != nil {
				log.Error("Panic recovered in gRPC stream handler",
					logger.String("method", info.FullMethod),
					logger.Any("panic", r),
					logger.String("stack", string(debug.Stack())),
				)
				err = status.Errorf(codes.Internal, "internal server error: %v", r)
			}
		}()

		return handler(srv, ss)
	}
}
