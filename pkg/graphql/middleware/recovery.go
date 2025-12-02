// pkg/graphql/middleware/recovery.go

package middleware

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/gqlerror"

	pkggraphql "github.com/0xsj/nexus-go/pkg/graphql"
	"github.com/0xsj/nexus-go/pkg/observability/logger"
)

// ============================================================================
// Recovery Middleware
// ============================================================================

// RecoveryMiddleware catches panics in GraphQL resolvers and converts them to errors.
type RecoveryMiddleware struct {
	logger logger.Logger
}

// NewRecoveryMiddleware creates a new recovery middleware.
func NewRecoveryMiddleware(log logger.Logger) *RecoveryMiddleware {
	return &RecoveryMiddleware{
		logger: log,
	}
}

// ExtensionName returns the name of the extension.
func (m *RecoveryMiddleware) ExtensionName() string {
	return "RecoveryMiddleware"
}

// Validate is called when adding the extension to the server.
func (m *RecoveryMiddleware) Validate(_ graphql.ExecutableSchema) error {
	return nil
}

// InterceptField wraps field resolution with panic recovery.
func (m *RecoveryMiddleware) InterceptField(ctx context.Context, next graphql.Resolver) (interface{}, error) {
	defer func() {
		if r := recover(); r != nil {
			m.handlePanic(ctx, r)
		}
	}()

	return next(ctx)
}

// InterceptResponse wraps response handling with panic recovery.
func (m *RecoveryMiddleware) InterceptResponse(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
	defer func() {
		if r := recover(); r != nil {
			m.handlePanic(ctx, r)
		}
	}()

	return next(ctx)
}

// handlePanic logs the panic and adds an error to the response.
func (m *RecoveryMiddleware) handlePanic(ctx context.Context, recovered interface{}) {
	// Capture stack trace
	stack := debug.Stack()

	// Build log fields
	fields := []logger.Field{
		logger.String("error", fmt.Sprintf("%v", recovered)),
		logger.String("stack_trace", string(stack)),
	}

	// Add context info
	if requestID := pkggraphql.GetRequestID(ctx); requestID != "" {
		fields = append(fields, logger.String("request_id", requestID))
	}

	if operationName := pkggraphql.GetOperationName(ctx); operationName != "" {
		fields = append(fields, logger.String("operation", operationName))
	}

	// Log the panic
	m.logger.Error("graphql panic recovered", fields...)

	// Add error to GraphQL response
	graphql.AddError(ctx, pkggraphql.InternalError())
}

// ============================================================================
// Functional API
// ============================================================================

// RecoveryHandler returns a recovery handler function for gqlgen.
//
// Usage:
//
//	srv := handler.NewDefaultServer(schema)
//	srv.SetRecoverFunc(middleware.RecoveryHandler(log))
func RecoveryHandler(log logger.Logger) graphql.RecoverFunc {
	return func(ctx context.Context, err interface{}) error {
		// Capture stack trace
		stack := debug.Stack()

		// Build log fields
		fields := []logger.Field{
			logger.String("error", fmt.Sprintf("%v", err)),
			logger.String("stack_trace", string(stack)),
		}

		// Add context info
		if requestID := pkggraphql.GetRequestID(ctx); requestID != "" {
			fields = append(fields, logger.String("request_id", requestID))
		}

		if operationName := pkggraphql.GetOperationName(ctx); operationName != "" {
			fields = append(fields, logger.String("operation", operationName))
		}

		// Log the panic
		log.Error("graphql panic recovered", fields...)

		// Return safe error (don't expose panic details)
		return pkggraphql.InternalError()
	}
}

// ============================================================================
// AroundOperations Recovery (Alternative)
// ============================================================================

// RecoveryOperationMiddleware wraps entire operations with panic recovery.
func RecoveryOperationMiddleware(log logger.Logger) graphql.OperationMiddleware {
	return func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
		return func(ctx context.Context) (resp *graphql.Response) {
			defer func() {
				if r := recover(); r != nil {
					// Capture stack trace
					stack := debug.Stack()

					// Build log fields
					fields := []logger.Field{
						logger.String("error", fmt.Sprintf("%v", r)),
						logger.String("stack_trace", string(stack)),
					}

					if requestID := pkggraphql.GetRequestID(ctx); requestID != "" {
						fields = append(fields, logger.String("request_id", requestID))
					}

					if operationName := pkggraphql.GetOperationName(ctx); operationName != "" {
						fields = append(fields, logger.String("operation", operationName))
					}

					log.Error("graphql operation panic recovered", fields...)

					// Return error response
					resp = &graphql.Response{
						Errors: gqlerror.List{
							{
								Message: "an internal error occurred",
								Extensions: map[string]interface{}{
									"code": "INTERNAL_SERVER_ERROR",
								},
							},
						},
					}
				}
			}()

			return next(ctx)(ctx)
		}
	}
}
