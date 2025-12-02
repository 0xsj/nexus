// pkg/graphql/middleware/metrics.go

package middleware

import (
	"context"
	"time"

	"github.com/99designs/gqlgen/graphql"

	pkggraphql "github.com/0xsj/nexus-go/pkg/graphql"
	"github.com/0xsj/nexus-go/pkg/observability/metrics"
)

// ============================================================================
// GraphQL Metrics
// ============================================================================

// GraphQLMetrics holds the metrics for GraphQL operations.
type GraphQLMetrics struct {
	operationsTotal    metrics.Counter
	operationDuration  metrics.Histogram
	operationsInFlight metrics.Gauge
	errorsTotal        metrics.Counter
	complexity         metrics.Histogram
	fieldResolutions   metrics.Counter
	fieldDuration      metrics.Histogram
}

// NewGraphQLMetrics creates GraphQL metrics using the provided metrics provider.
func NewGraphQLMetrics(provider metrics.Provider) *GraphQLMetrics {
	return &GraphQLMetrics{
		operationsTotal: provider.Counter(
			"graphql_operations_total",
			"Total number of GraphQL operations",
			"operation_type", "operation_name", "status",
		),
		operationDuration: provider.Histogram(
			"graphql_operation_duration_seconds",
			"GraphQL operation duration in seconds",
			metrics.HTTPLatencyBuckets(),
			"operation_type", "operation_name",
		),
		operationsInFlight: provider.Gauge(
			"graphql_operations_in_flight",
			"Number of GraphQL operations currently being processed",
			"operation_type",
		),
		errorsTotal: provider.Counter(
			"graphql_errors_total",
			"Total number of GraphQL errors",
			"operation_type", "operation_name", "error_code",
		),
		complexity: provider.Histogram(
			"graphql_query_complexity",
			"GraphQL query complexity",
			[]float64{10, 25, 50, 100, 150, 200, 300, 500},
			"operation_type", "operation_name",
		),
		fieldResolutions: provider.Counter(
			"graphql_field_resolutions_total",
			"Total number of GraphQL field resolutions",
			"object", "field",
		),
		fieldDuration: provider.Histogram(
			"graphql_field_duration_seconds",
			"GraphQL field resolution duration in seconds",
			metrics.HTTPLatencyBuckets(),
			"object", "field",
		),
	}
}

// ============================================================================
// Metrics Middleware
// ============================================================================

// MetricsMiddleware collects metrics for GraphQL operations.
type MetricsMiddleware struct {
	metrics *GraphQLMetrics
	opts    metricsOptions
}

// metricsOptions holds configuration for the metrics middleware.
type metricsOptions struct {
	// skipIntrospection skips metrics for introspection queries.
	skipIntrospection bool

	// trackFields enables field-level metrics.
	trackFields bool
}

// MetricsOption configures the metrics middleware.
type MetricsOption func(*metricsOptions)

// WithSkipIntrospectionMetrics skips metrics for introspection queries.
func WithSkipIntrospectionMetrics(skip bool) MetricsOption {
	return func(o *metricsOptions) {
		o.skipIntrospection = skip
	}
}

// WithFieldMetrics enables field-level metrics collection.
func WithFieldMetrics(enabled bool) MetricsOption {
	return func(o *metricsOptions) {
		o.trackFields = enabled
	}
}

// NewMetricsMiddleware creates a new metrics middleware.
func NewMetricsMiddleware(m *GraphQLMetrics, opts ...MetricsOption) *MetricsMiddleware {
	options := metricsOptions{
		skipIntrospection: true,
		trackFields:       false,
	}

	for _, opt := range opts {
		opt(&options)
	}

	return &MetricsMiddleware{
		metrics: m,
		opts:    options,
	}
}

// ExtensionName returns the name of the extension.
func (m *MetricsMiddleware) ExtensionName() string {
	return "MetricsMiddleware"
}

// Validate is called when adding the extension to the server.
func (m *MetricsMiddleware) Validate(_ graphql.ExecutableSchema) error {
	return nil
}

// InterceptResponse collects metrics for the operation.
func (m *MetricsMiddleware) InterceptResponse(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
	opCtx := graphql.GetOperationContext(ctx)
	if opCtx == nil {
		return next(ctx)
	}

	// Skip introspection queries
	if m.opts.skipIntrospection && isIntrospectionQuery(opCtx) {
		return next(ctx)
	}

	// Extract operation info
	opType := string(opCtx.Operation.Operation)
	opName := opCtx.OperationName
	if opName == "" {
		opName = "anonymous"
	}

	// Track in-flight operations
	m.metrics.operationsInFlight.Inc(opType)
	defer m.metrics.operationsInFlight.Dec(opType)

	// Record start time
	start := time.Now()

	// Execute operation
	resp := next(ctx)

	// Calculate duration
	duration := time.Since(start).Seconds()

	// Determine status
	status := "success"
	if resp != nil && len(resp.Errors) > 0 {
		status = "error"
	}

	// Record metrics
	m.metrics.operationsTotal.Inc(opType, opName, status)
	m.metrics.operationDuration.Observe(duration, opType, opName)

	// Record complexity if available
	if complexity := pkggraphql.GetComplexity(ctx); complexity > 0 {
		m.metrics.complexity.Observe(float64(complexity), opType, opName)
	}

	// Record errors
	if resp != nil && len(resp.Errors) > 0 {
		for _, err := range resp.Errors {
			errorCode := "UNKNOWN"
			if err.Extensions != nil {
				if code, ok := err.Extensions["code"].(string); ok {
					errorCode = code
				}
			}
			m.metrics.errorsTotal.Inc(opType, opName, errorCode)
		}
	}

	return resp
}

// InterceptField collects field-level metrics if enabled.
func (m *MetricsMiddleware) InterceptField(ctx context.Context, next graphql.Resolver) (interface{}, error) {
	if !m.opts.trackFields {
		return next(ctx)
	}

	fc := graphql.GetFieldContext(ctx)
	if fc == nil {
		return next(ctx)
	}

	// Skip non-resolver fields
	if !fc.IsResolver {
		return next(ctx)
	}

	// Record start time
	start := time.Now()

	// Execute resolver
	result, err := next(ctx)

	// Calculate duration
	duration := time.Since(start).Seconds()

	// Record metrics
	m.metrics.fieldResolutions.Inc(fc.Object, fc.Field.Name)
	m.metrics.fieldDuration.Observe(duration, fc.Object, fc.Field.Name)

	return result, err
}

// ============================================================================
// Operation Middleware
// ============================================================================

// MetricsOperationMiddleware returns an operation middleware for metrics.
//
// Usage:
//
//	gqlMetrics := middleware.NewGraphQLMetrics(metricsProvider)
//	srv := handler.NewDefaultServer(schema)
//	srv.AroundOperations(middleware.MetricsOperationMiddleware(gqlMetrics))
func MetricsOperationMiddleware(m *GraphQLMetrics, opts ...MetricsOption) graphql.OperationMiddleware {
	middleware := NewMetricsMiddleware(m, opts...)

	return func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
		opCtx := graphql.GetOperationContext(ctx)
		if opCtx == nil {
			return next(ctx)
		}

		// Skip introspection queries
		if middleware.opts.skipIntrospection && isIntrospectionQuery(opCtx) {
			return next(ctx)
		}

		// Extract operation info
		opType := string(opCtx.Operation.Operation)
		opName := opCtx.OperationName
		if opName == "" {
			opName = "anonymous"
		}

		// Track in-flight operations
		m.operationsInFlight.Inc(opType)

		return func(ctx context.Context) *graphql.Response {
			defer m.operationsInFlight.Dec(opType)

			// Record start time
			start := time.Now()

			// Execute operation
			resp := next(ctx)(ctx)

			// Calculate duration
			duration := time.Since(start).Seconds()

			// Determine status
			status := "success"
			if resp != nil && len(resp.Errors) > 0 {
				status = "error"
			}

			// Record metrics
			m.operationsTotal.Inc(opType, opName, status)
			m.operationDuration.Observe(duration, opType, opName)

			// Record complexity if available
			if complexity := pkggraphql.GetComplexity(ctx); complexity > 0 {
				m.complexity.Observe(float64(complexity), opType, opName)
			}

			// Record errors
			if resp != nil && len(resp.Errors) > 0 {
				for _, err := range resp.Errors {
					errorCode := "UNKNOWN"
					if err.Extensions != nil {
						if code, ok := err.Extensions["code"].(string); ok {
							errorCode = code
						}
					}
					m.errorsTotal.Inc(opType, opName, errorCode)
				}
			}

			return resp
		}
	}
}

// ============================================================================
// Field Metrics Middleware
// ============================================================================

// FieldMetricsMiddleware collects field-level metrics.
//
// Usage:
//
//	gqlMetrics := middleware.NewGraphQLMetrics(metricsProvider)
//	srv := handler.NewDefaultServer(schema)
//	srv.AroundFields(middleware.FieldMetricsMiddleware(gqlMetrics))
func FieldMetricsMiddleware(m *GraphQLMetrics) graphql.FieldMiddleware {
	return func(ctx context.Context, next graphql.Resolver) (interface{}, error) {
		fc := graphql.GetFieldContext(ctx)
		if fc == nil {
			return next(ctx)
		}

		// Skip non-resolver fields
		if !fc.IsResolver {
			return next(ctx)
		}

		// Record start time
		start := time.Now()

		// Execute resolver
		result, err := next(ctx)

		// Calculate duration
		duration := time.Since(start).Seconds()

		// Record metrics
		m.fieldResolutions.Inc(fc.Object, fc.Field.Name)
		m.fieldDuration.Observe(duration, fc.Object, fc.Field.Name)

		return result, err
	}
}
