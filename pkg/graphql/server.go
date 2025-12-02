// pkg/graphql/server.go

package graphql

import (
	"context"
	"net/http"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"

	"github.com/0xsj/nexus-go/pkg/observability/logger"
)

// ============================================================================
// Server
// ============================================================================

// Server wraps the GraphQL handler with configuration.
type Server struct {
	config  Config
	handler *handler.Server
	logger  logger.Logger
}

// ServerOption configures the server.
type ServerOption func(*serverOptions)

type serverOptions struct {
	logger            logger.Logger
	skipIntrospection bool
	enableAPQ         bool
	recoverFunc       graphql.RecoverFunc
}

// WithServerLogger sets the logger.
func WithServerLogger(log logger.Logger) ServerOption {
	return func(o *serverOptions) {
		o.logger = log
	}
}

// WithIntrospection enables or disables introspection.
func WithIntrospection(enabled bool) ServerOption {
	return func(o *serverOptions) {
		o.skipIntrospection = !enabled
	}
}

// WithAPQ enables Automatic Persisted Queries.
func WithAPQ(enabled bool) ServerOption {
	return func(o *serverOptions) {
		o.enableAPQ = enabled
	}
}

// WithRecoverFunc sets the panic recovery function.
func WithRecoverFunc(fn graphql.RecoverFunc) ServerOption {
	return func(o *serverOptions) {
		o.recoverFunc = fn
	}
}

// NewServer creates a new GraphQL server.
func NewServer(es graphql.ExecutableSchema, config Config, opts ...ServerOption) *Server {
	options := &serverOptions{}

	for _, opt := range opts {
		opt(options)
	}

	// Create base handler
	srv := handler.New(es)

	// Configure transports
	configureTransports(srv, config)

	// Configure extensions
	configureExtensions(srv, config, options)

	// Configure recovery
	if options.recoverFunc != nil {
		srv.SetRecoverFunc(options.recoverFunc)
	}

	return &Server{
		config:  config,
		handler: srv,
		logger:  options.logger,
	}
}

// configureTransports sets up HTTP and WebSocket transports.
func configureTransports(srv *handler.Server, config Config) {
	// HTTP POST (primary)
	srv.AddTransport(transport.POST{})

	// HTTP GET (for simple queries)
	srv.AddTransport(transport.GET{})

	// Options (for CORS preflight)
	srv.AddTransport(transport.Options{})

	// Multipart (for file uploads)
	srv.AddTransport(transport.MultipartForm{
		MaxMemory:     config.MaxUploadSize,
		MaxUploadSize: config.MaxUploadSize,
	})

	// WebSocket (for subscriptions)
	srv.AddTransport(transport.Websocket{
		KeepAlivePingInterval: config.WebsocketKeepAliveDuration,
	})
}

// configureExtensions sets up GraphQL extensions.
func configureExtensions(srv *handler.Server, config Config, options *serverOptions) {
	// Introspection
	if config.IntrospectionEnabled {
		srv.Use(extension.Introspection{})
	}

	// Automatic Persisted Queries
	if config.APQEnabled {
		srv.Use(extension.AutomaticPersistedQuery{
			Cache: lru.New[string](1000),
		})
	}

	// Query cache
	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))
}

// ============================================================================
// Handler Access
// ============================================================================

// Handler returns the underlying gqlgen handler.
// Use this to add custom extensions and middleware.
//
// Usage:
//
//	srv := graphql.NewServer(schema, config)
//	srv.Handler().Use(myExtension)
//	srv.Handler().AroundOperations(myMiddleware)
func (s *Server) Handler() *handler.Server {
	return s.handler
}

// ============================================================================
// HTTP Handlers
// ============================================================================

// ServeHTTP implements http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.handler.ServeHTTP(w, r)
}

// PlaygroundHandler returns the GraphQL Playground handler.
func (s *Server) PlaygroundHandler(endpoint string) http.Handler {
	return playground.Handler("GraphQL Playground", endpoint)
}

// ============================================================================
// Router Helpers
// ============================================================================

// Routes returns HTTP handlers for mounting on a router.
//
// Usage with Chi:
//
//	r := chi.NewRouter()
//	routes := server.Routes()
//	r.Handle(routes.GraphQLPath, routes.GraphQLHandler)
//	if routes.PlaygroundHandler != nil {
//	    r.Handle(routes.PlaygroundPath, routes.PlaygroundHandler)
//	}
type Routes struct {
	GraphQLPath       string
	GraphQLHandler    http.Handler
	PlaygroundPath    string
	PlaygroundHandler http.Handler
}

// Routes returns the configured routes.
func (s *Server) Routes() Routes {
	routes := Routes{
		GraphQLPath:    s.config.Path,
		GraphQLHandler: s.handler,
	}

	if s.config.PlaygroundEnabled && s.config.PlaygroundPath != "" {
		routes.PlaygroundPath = s.config.PlaygroundPath
		routes.PlaygroundHandler = s.PlaygroundHandler(s.config.Path)
	}

	return routes
}

// Mount mounts the GraphQL handlers on a router.
//
// Usage:
//
//	r := chi.NewRouter()
//	server.Mount(r)
func (s *Server) Mount(mux interface{ Handle(string, http.Handler) }) {
	routes := s.Routes()

	mux.Handle(routes.GraphQLPath, routes.GraphQLHandler)

	if routes.PlaygroundHandler != nil {
		mux.Handle(routes.PlaygroundPath, routes.PlaygroundHandler)
	}
}

// ============================================================================
// Request ID Middleware
// ============================================================================

// RequestIDMiddleware generates or extracts request IDs.
func RequestIDMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Try to get from header
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = generateRequestID()
			}

			// Add to context
			ctx = WithRequestID(ctx, requestID)

			// Add to response header
			w.Header().Set("X-Request-ID", requestID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// generateRequestID generates a new request ID.
func generateRequestID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString generates a random string of given length.
func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
		time.Sleep(1 * time.Nanosecond)
	}
	return string(b)
}

// ============================================================================
// Error Presenter
// ============================================================================

// SetErrorPresenter sets a custom error presenter.
func (s *Server) SetErrorPresenter(presenter graphql.ErrorPresenterFunc) {
	s.handler.SetErrorPresenter(presenter)
}

// DefaultErrorPresenter returns a default error presenter that converts
// domain errors to GraphQL errors.
func DefaultErrorPresenter(log logger.Logger) graphql.ErrorPresenterFunc {
	return func(ctx context.Context, err error) *gqlerror.Error {
		// Convert to GraphQL error
		gqlErr := ToGraphQLError(err)

		// Log server errors
		if gqlErr.Extensions.Code == CodeInternalError {
			if log != nil {
				log.Error("graphql internal error",
					logger.String("message", gqlErr.Message),
					logger.String("request_id", GetRequestID(ctx)),
				)
			}
		}

		return &gqlerror.Error{
			Message:    gqlErr.Message,
			Path:       graphql.GetPath(ctx),
			Extensions: extensionsToMap(gqlErr.Extensions),
		}
	}
}

// extensionsToMap converts Extensions to map[string]interface{}.
func extensionsToMap(ext Extensions) map[string]interface{} {
	m := map[string]interface{}{
		"code": ext.Code,
	}

	if ext.ErrorCode != "" {
		m["errorCode"] = ext.ErrorCode
	}

	if ext.Field != "" {
		m["field"] = ext.Field
	}

	if len(ext.Fields) > 0 {
		m["fields"] = ext.Fields
	}

	if ext.Retryable {
		m["retryable"] = ext.Retryable
	}

	return m
}
