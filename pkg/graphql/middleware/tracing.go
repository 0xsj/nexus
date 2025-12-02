// pkg/graphql/middleware/tracing.go

package middleware

import (
	"context"
	"fmt"

	"github.com/99designs/gqlgen/graphql"

	pkggraphql "github.com/0xsj/nexus-go/pkg/graphql"
	"github.com/0xsj/nexus-go/pkg/observability/tracing"
)

// ============================================================================
// Tracing Middleware
// ============================================================================

// TracingMiddleware provides distributed tracing for GraphQL operations.
type TracingMiddleware struct {
	tracer tracing.Tracer
	opts   tracingOptions
}

// tracingOptions holds configuration for the tracing middleware.
type tracingOptions struct {
	// skipIntrospection skips tracing for introspection queries.
	skipIntrospection bool

	// traceFields enables tracing of individual field resolutions.
	traceFields bool
}

// TracingOption configures the tracing middleware.
type TracingOption func(*tracingOptions)

// WithSkipIntrospectionTracing skips tracing for introspection queries.
func WithSkipIntrospectionTracing(skip bool) TracingOption {
	return func(o *tracingOptions) {
		o.skipIntrospection = skip
	}
}

// WithFieldTracing enables tracing of individual field resolutions.
func WithFieldTracing(enabled bool) TracingOption {
	return func(o *tracingOptions) {
		o.traceFields = enabled
	}
}

// NewTracingMiddleware creates a new tracing middleware.
func NewTracingMiddleware(tracer tracing.Tracer, opts ...TracingOption) *TracingMiddleware {
	options := tracingOptions{
		skipIntrospection: true,
		traceFields:       false,
	}

	for _, opt := range opts {
		opt(&options)
	}

	return &TracingMiddleware{
		tracer: tracer,
		opts:   options,
	}
}

// ExtensionName returns the name of the extension.
func (m *TracingMiddleware) ExtensionName() string {
	return "TracingMiddleware"
}

// Validate is called when adding the extension to the server.
func (m *TracingMiddleware) Validate(_ graphql.ExecutableSchema) error {
	return nil
}

// InterceptResponse wraps the entire operation with a span.
func (m *TracingMiddleware) InterceptResponse(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
	opCtx := graphql.GetOperationContext(ctx)
	if opCtx == nil {
		return next(ctx)
	}

	// Skip introspection queries
	if m.opts.skipIntrospection && isIntrospectionQuery(opCtx) {
		return next(ctx)
	}

	// Create span name
	spanName := m.buildSpanName(opCtx)

	// Start span
	ctx, span := m.tracer.Start(ctx, spanName,
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithAttributes(map[string]any{
			"graphql.operation.type": string(opCtx.Operation.Operation),
			"graphql.operation.name": opCtx.OperationName,
			"graphql.document":       opCtx.RawQuery,
		}),
	)
	defer span.End()

	// Add context attributes
	m.addContextAttributes(ctx, span)

	// Execute operation
	resp := next(ctx)

	// Record result
	m.recordResult(span, resp)

	return resp
}

// InterceptField traces individual field resolutions if enabled.
func (m *TracingMiddleware) InterceptField(ctx context.Context, next graphql.Resolver) (interface{}, error) {
	if !m.opts.traceFields {
		return next(ctx)
	}

	fc := graphql.GetFieldContext(ctx)
	if fc == nil {
		return next(ctx)
	}

	// Skip trivial fields
	if fc.IsResolver == false {
		return next(ctx)
	}

	// Create span
	spanName := fmt.Sprintf("field:%s.%s", fc.Object, fc.Field.Name)
	ctx, span := m.tracer.Start(ctx, spanName,
		tracing.WithSpanKind(tracing.SpanKindInternal),
		tracing.WithAttributes(map[string]any{
			"graphql.field.name":   fc.Field.Name,
			"graphql.field.object": fc.Object,
			"graphql.field.path":   fc.Path().String(),
		}),
	)
	defer span.End()

	// Execute resolver
	result, err := next(ctx)

	// Record result
	if err != nil {
		span.RecordError(err)
		span.SetStatus(tracing.StatusError, err.Error())
	} else {
		span.SetStatus(tracing.StatusOK, "")
	}

	return result, err
}

// buildSpanName creates a span name for the operation.
func (m *TracingMiddleware) buildSpanName(opCtx *graphql.OperationContext) string {
	opType := string(opCtx.Operation.Operation)

	if opCtx.OperationName != "" {
		return fmt.Sprintf("graphql.%s.%s", opType, opCtx.OperationName)
	}

	return fmt.Sprintf("graphql.%s", opType)
}

// addContextAttributes adds context values as span attributes.
func (m *TracingMiddleware) addContextAttributes(ctx context.Context, span tracing.Span) {
	if requestID := pkggraphql.GetRequestID(ctx); requestID != "" {
		span.SetAttribute("request.id", requestID)
	}

	if userID := pkggraphql.GetUserID(ctx); userID != "" {
		span.SetAttribute("user.id", userID)
	}

	if tenantID := pkggraphql.GetTenantID(ctx); tenantID != "" {
		span.SetAttribute("tenant.id", tenantID)
	}

	if complexity := pkggraphql.GetComplexity(ctx); complexity > 0 {
		span.SetAttribute("graphql.complexity", complexity)
	}
}

// recordResult records the operation result on the span.
func (m *TracingMiddleware) recordResult(span tracing.Span, resp *graphql.Response) {
	if resp == nil {
		return
	}

	hasErrors := len(resp.Errors) > 0

	span.SetAttribute("graphql.has_errors", hasErrors)
	span.SetAttribute("graphql.error_count", len(resp.Errors))

	if hasErrors {
		// Record first error
		span.SetStatus(tracing.StatusError, resp.Errors[0].Message)

		// Add error event
		span.AddEvent("graphql.errors", map[string]any{
			"count": len(resp.Errors),
		})
	} else {
		span.SetStatus(tracing.StatusOK, "")
	}
}

// ============================================================================
// Operation Middleware
// ============================================================================

// TracingOperationMiddleware returns an operation middleware for tracing.
//
// Usage:
//
//	srv := handler.NewDefaultServer(schema)
//	srv.AroundOperations(middleware.TracingOperationMiddleware(tracer))
func TracingOperationMiddleware(tracer tracing.Tracer, opts ...TracingOption) graphql.OperationMiddleware {
	m := NewTracingMiddleware(tracer, opts...)

	return func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
		opCtx := graphql.GetOperationContext(ctx)
		if opCtx == nil {
			return next(ctx)
		}

		// Skip introspection queries
		if m.opts.skipIntrospection && isIntrospectionQuery(opCtx) {
			return next(ctx)
		}

		// Create span name
		spanName := m.buildSpanName(opCtx)

		// Start span
		ctx, span := tracer.Start(ctx, spanName,
			tracing.WithSpanKind(tracing.SpanKindServer),
			tracing.WithAttributes(map[string]any{
				"graphql.operation.type": string(opCtx.Operation.Operation),
				"graphql.operation.name": opCtx.OperationName,
			}),
		)

		// Add context attributes
		m.addContextAttributes(ctx, span)

		// Return response handler that closes span
		return func(ctx context.Context) *graphql.Response {
			defer span.End()

			resp := next(ctx)(ctx)
			m.recordResult(span, resp)

			return resp
		}
	}
}

// ============================================================================
// Field Tracing Middleware
// ============================================================================

// FieldTracingMiddleware traces individual field resolutions.
// Use for debugging or detailed performance analysis.
//
// Usage:
//
//	srv := handler.NewDefaultServer(schema)
//	srv.AroundFields(middleware.FieldTracingMiddleware(tracer))
func FieldTracingMiddleware(tracer tracing.Tracer) graphql.FieldMiddleware {
	return func(ctx context.Context, next graphql.Resolver) (interface{}, error) {
		fc := graphql.GetFieldContext(ctx)
		if fc == nil {
			return next(ctx)
		}

		// Skip non-resolver fields
		if !fc.IsResolver {
			return next(ctx)
		}

		// Create span
		spanName := fmt.Sprintf("field:%s.%s", fc.Object, fc.Field.Name)
		ctx, span := tracer.Start(ctx, spanName,
			tracing.WithSpanKind(tracing.SpanKindInternal),
			tracing.WithAttributes(map[string]any{
				"graphql.field.name":   fc.Field.Name,
				"graphql.field.object": fc.Object,
				"graphql.field.path":   fc.Path().String(),
			}),
		)
		defer span.End()

		// Execute resolver
		result, err := next(ctx)

		// Record result
		if err != nil {
			span.RecordError(err)
			span.SetStatus(tracing.StatusError, err.Error())
		} else {
			span.SetStatus(tracing.StatusOK, "")
		}

		return result, err
	}
}
