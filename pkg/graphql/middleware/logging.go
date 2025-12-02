// pkg/graphql/middleware/logging.go

package middleware

import (
	"context"
	"time"

	"github.com/99designs/gqlgen/graphql"

	pkggraphql "github.com/0xsj/nexus-go/pkg/graphql"
	"github.com/0xsj/nexus-go/pkg/observability/logger"
)

// ============================================================================
// Logging Middleware
// ============================================================================

// LoggingMiddleware logs GraphQL operations.
type LoggingMiddleware struct {
	logger logger.Logger
	opts   loggingOptions
}

// loggingOptions holds configuration for the logging middleware.
type loggingOptions struct {
	// logVariables enables logging of operation variables.
	// Use with caution - may log sensitive data.
	logVariables bool

	// logResponses enables logging of response data.
	// Use with caution - may log large payloads.
	logResponses bool

	// slowThreshold is the duration above which operations are logged as slow.
	slowThreshold time.Duration

	// skipIntrospection skips logging for introspection queries.
	skipIntrospection bool
}

// LoggingOption configures the logging middleware.
type LoggingOption func(*loggingOptions)

// WithVariableLogging enables logging of operation variables.
func WithVariableLogging(enabled bool) LoggingOption {
	return func(o *loggingOptions) {
		o.logVariables = enabled
	}
}

// WithResponseLogging enables logging of response data.
func WithResponseLogging(enabled bool) LoggingOption {
	return func(o *loggingOptions) {
		o.logResponses = enabled
	}
}

// WithSlowThreshold sets the duration threshold for slow operation logging.
func WithSlowThreshold(d time.Duration) LoggingOption {
	return func(o *loggingOptions) {
		o.slowThreshold = d
	}
}

// WithSkipIntrospection skips logging for introspection queries.
func WithSkipIntrospection(skip bool) LoggingOption {
	return func(o *loggingOptions) {
		o.skipIntrospection = skip
	}
}

// NewLoggingMiddleware creates a new logging middleware.
func NewLoggingMiddleware(log logger.Logger, opts ...LoggingOption) *LoggingMiddleware {
	options := loggingOptions{
		slowThreshold:     3 * time.Second,
		skipIntrospection: true,
	}

	for _, opt := range opts {
		opt(&options)
	}

	return &LoggingMiddleware{
		logger: log,
		opts:   options,
	}
}

// ExtensionName returns the name of the extension.
func (m *LoggingMiddleware) ExtensionName() string {
	return "LoggingMiddleware"
}

// Validate is called when adding the extension to the server.
func (m *LoggingMiddleware) Validate(_ graphql.ExecutableSchema) error {
	return nil
}

// InterceptResponse logs the operation after completion.
func (m *LoggingMiddleware) InterceptResponse(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
	// Get operation context
	opCtx := graphql.GetOperationContext(ctx)
	if opCtx == nil {
		return next(ctx)
	}

	// Skip introspection queries
	if m.opts.skipIntrospection && isIntrospectionQuery(opCtx) {
		return next(ctx)
	}

	// Record start time
	start := time.Now()

	// Execute operation
	resp := next(ctx)

	// Calculate duration
	duration := time.Since(start)

	// Log the operation
	m.logOperation(ctx, opCtx, resp, duration)

	return resp
}

// logOperation logs details about the GraphQL operation.
func (m *LoggingMiddleware) logOperation(
	ctx context.Context,
	opCtx *graphql.OperationContext,
	resp *graphql.Response,
	duration time.Duration,
) {
	// Build base fields
	fields := []logger.Field{
		logger.String("operation_type", string(opCtx.Operation.Operation)),
		logger.Duration("duration", duration),
	}

	// Add operation name if present
	if opCtx.OperationName != "" {
		fields = append(fields, logger.String("operation_name", opCtx.OperationName))
	}

	// Add context info
	if requestID := pkggraphql.GetRequestID(ctx); requestID != "" {
		fields = append(fields, logger.String("request_id", requestID))
	}

	if userID := pkggraphql.GetUserID(ctx); userID != "" {
		fields = append(fields, logger.String("user_id", userID))
	}

	if tenantID := pkggraphql.GetTenantID(ctx); tenantID != "" {
		fields = append(fields, logger.String("tenant_id", tenantID))
	}

	// Add complexity if available
	if complexity := pkggraphql.GetComplexity(ctx); complexity > 0 {
		fields = append(fields, logger.Int("complexity", complexity))
	}

	// Add variables if enabled
	if m.opts.logVariables && len(opCtx.Variables) > 0 {
		fields = append(fields, logger.Any("variables", sanitizeVariables(opCtx.Variables)))
	}

	// Add error info
	hasErrors := resp != nil && len(resp.Errors) > 0
	if hasErrors {
		fields = append(fields, logger.Int("error_count", len(resp.Errors)))
		fields = append(fields, logger.Any("errors", formatErrors(resp.Errors)))
	}

	// Add response data if enabled
	if m.opts.logResponses && resp != nil && resp.Data != nil {
		fields = append(fields, logger.Any("response_data", resp.Data))
	}

	// Determine log level
	switch {
	case hasErrors:
		m.logger.Error("graphql operation completed with errors", fields...)
	case duration > m.opts.slowThreshold:
		m.logger.Warn("graphql operation completed (slow)", fields...)
	default:
		m.logger.Info("graphql operation completed", fields...)
	}
}

// ============================================================================
// Operation Middleware
// ============================================================================

// LoggingOperationMiddleware returns an operation middleware for logging.
//
// Usage:
//
//	srv := handler.NewDefaultServer(schema)
//	srv.AroundOperations(middleware.LoggingOperationMiddleware(log))
func LoggingOperationMiddleware(log logger.Logger, opts ...LoggingOption) graphql.OperationMiddleware {
	m := NewLoggingMiddleware(log, opts...)

	return func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
		opCtx := graphql.GetOperationContext(ctx)

		// Log operation start
		if opCtx != nil && !m.opts.skipIntrospection || !isIntrospectionQuery(opCtx) {
			startFields := []logger.Field{
				logger.String("operation_type", string(opCtx.Operation.Operation)),
			}

			if opCtx.OperationName != "" {
				startFields = append(startFields, logger.String("operation_name", opCtx.OperationName))
			}

			if requestID := pkggraphql.GetRequestID(ctx); requestID != "" {
				startFields = append(startFields, logger.String("request_id", requestID))
			}

			log.Info("graphql operation started", startFields...)
		}

		return next(ctx)
	}
}

// ============================================================================
// Field Logging (for debugging)
// ============================================================================

// FieldLoggingMiddleware logs individual field resolutions.
// Use only for debugging - very verbose.
type FieldLoggingMiddleware struct {
	logger logger.Logger
}

// NewFieldLoggingMiddleware creates a new field logging middleware.
func NewFieldLoggingMiddleware(log logger.Logger) *FieldLoggingMiddleware {
	return &FieldLoggingMiddleware{
		logger: log,
	}
}

// ExtensionName returns the name of the extension.
func (m *FieldLoggingMiddleware) ExtensionName() string {
	return "FieldLoggingMiddleware"
}

// Validate is called when adding the extension to the server.
func (m *FieldLoggingMiddleware) Validate(_ graphql.ExecutableSchema) error {
	return nil
}

// InterceptField logs field resolution.
func (m *FieldLoggingMiddleware) InterceptField(ctx context.Context, next graphql.Resolver) (interface{}, error) {
	fc := graphql.GetFieldContext(ctx)
	if fc == nil {
		return next(ctx)
	}

	start := time.Now()
	result, err := next(ctx)
	duration := time.Since(start)

	fields := []logger.Field{
		logger.String("field", fc.Field.Name),
		logger.String("path", fc.Path().String()),
		logger.Duration("duration", duration),
	}

	if err != nil {
		fields = append(fields, logger.Err(err))
		m.logger.Debug("graphql field resolved with error", fields...)
	} else {
		m.logger.Debug("graphql field resolved", fields...)
	}

	return result, err
}

// ============================================================================
// Helper Functions
// ============================================================================

// isIntrospectionQuery checks if the operation is an introspection query.
func isIntrospectionQuery(opCtx *graphql.OperationContext) bool {
	if opCtx == nil || opCtx.Operation == nil {
		return false
	}

	// Check operation name
	if opCtx.OperationName == "IntrospectionQuery" {
		return true
	}

	// Check for __schema or __type in selection set
	for _, selection := range opCtx.Operation.SelectionSet {
		if field, ok := selection.(*graphql.CollectedField); ok {
			if field.Name == "__schema" || field.Name == "__type" {
				return true
			}
		}
	}

	return false
}

// sanitizeVariables removes sensitive fields from variables.
func sanitizeVariables(variables map[string]interface{}) map[string]interface{} {
	sensitiveFields := map[string]bool{
		"password":        true,
		"currentPassword": true,
		"newPassword":     true,
		"confirmPassword": true,
		"token":           true,
		"refreshToken":    true,
		"accessToken":     true,
		"secret":          true,
		"apiKey":          true,
		"creditCard":      true,
		"ssn":             true,
	}

	sanitized := make(map[string]interface{})
	for k, v := range variables {
		if sensitiveFields[k] {
			sanitized[k] = "[REDACTED]"
		} else {
			sanitized[k] = v
		}
	}

	return sanitized
}

// formatErrors formats GraphQL errors for logging.
func formatErrors(errors interface{}) []map[string]interface{} {
	// Type assertion for gqlerror.List
	if errList, ok := errors.(interface{ Error() string }); ok {
		return []map[string]interface{}{
			{"message": errList.Error()},
		}
	}

	return nil
}
