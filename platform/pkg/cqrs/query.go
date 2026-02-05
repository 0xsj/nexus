package cqrs

import (
	"context"
	"reflect"
	"sync"
	"time"
)

// ============================================================================
// Query Interface
// ============================================================================

// Query represents a request for data.
// Queries are named as questions: GetCredential, ListCredentials.
type Query interface {
	// QueryName returns the unique name of the query.
	QueryName() string
}

// ============================================================================
// Query Handler
// ============================================================================

// QueryHandler handles a specific query type.
type QueryHandler[Q Query, R any] interface {
	Handle(ctx context.Context, query Q) (R, error)
}

// QueryHandlerFunc is a function adapter for QueryHandler.
type QueryHandlerFunc[Q Query, R any] func(ctx context.Context, query Q) (R, error)

// Handle implements QueryHandler.
func (f QueryHandlerFunc[Q, R]) Handle(ctx context.Context, query Q) (R, error) {
	return f(ctx, query)
}

// ============================================================================
// Query Bus
// ============================================================================

// QueryBus dispatches queries to their handlers.
type QueryBus interface {
	// Dispatch sends a query to its handler and returns the result.
	Dispatch(ctx context.Context, query Query) (any, error)

	// Register registers a handler for a query type.
	Register(queryType string, handler any) error
}

// ============================================================================
// In-Memory Query Bus Implementation
// ============================================================================

// InMemoryQueryBus is a simple in-memory query bus.
type InMemoryQueryBus struct {
	mu       sync.RWMutex
	handlers map[string]any
}

// NewQueryBus creates a new in-memory query bus.
func NewQueryBus() *InMemoryQueryBus {
	return &InMemoryQueryBus{
		handlers: make(map[string]any),
	}
}

// Register registers a handler for a query type.
func (b *InMemoryQueryBus) Register(queryType string, handler any) error {
	const op = "QueryBus.Register"

	if queryType == "" {
		return ErrQueryValidation(op, "query type cannot be empty")
	}
	if handler == nil {
		return ErrQueryValidation(op, "handler cannot be nil")
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	b.handlers[queryType] = handler
	return nil
}

// RegisterQuery registers a typed query handler.
func RegisterQuery[Q Query, R any](bus *InMemoryQueryBus, handler QueryHandler[Q, R]) error {
	var zero Q
	return bus.Register(zero.QueryName(), handler)
}

// Dispatch sends a query to its handler.
func (b *InMemoryQueryBus) Dispatch(ctx context.Context, query Query) (any, error) {
	const op = "QueryBus.Dispatch"

	if query == nil {
		return nil, ErrQueryValidation(op, "query cannot be nil")
	}

	queryType := query.QueryName()

	b.mu.RLock()
	handler, exists := b.handlers[queryType]
	b.mu.RUnlock()

	if !exists {
		return nil, ErrQueryNotFound(op, queryType)
	}

	// Use reflection to call the handler
	return b.invokeHandler(ctx, handler, query)
}

// invokeHandler calls the handler using reflection.
func (b *InMemoryQueryBus) invokeHandler(ctx context.Context, handler any, query Query) (result any, err error) {
	const op = "QueryBus.invokeHandler"

	// Recover from panics
	defer func() {
		if r := recover(); r != nil {
			err = ErrHandlerPanic(op, r)
		}
	}()

	handlerValue := reflect.ValueOf(handler)
	handleMethod := handlerValue.MethodByName("Handle")

	if !handleMethod.IsValid() {
		return nil, ErrHandlerNotFound(op, "Handle method")
	}

	// Call Handle(ctx, query)
	results := handleMethod.Call([]reflect.Value{
		reflect.ValueOf(ctx),
		reflect.ValueOf(query),
	})

	// Parse results
	if len(results) != 2 {
		return nil, ErrQueryFailed(op, nil)
	}

	// First result: R (any type)
	if !results[0].IsNil() {
		result = results[0].Interface()
	} else if results[0].CanInterface() {
		// Handle non-pointer types
		result = results[0].Interface()
	}

	// Second result: error
	if !results[1].IsNil() {
		err = results[1].Interface().(error)
	}

	return result, err
}

// ============================================================================
// Query Middleware
// ============================================================================

// QueryMiddleware wraps query handling with additional behavior.
type QueryMiddleware func(next QueryHandlerFunc[Query, any]) QueryHandlerFunc[Query, any]

// MiddlewareQueryBus wraps a query bus with middleware.
type MiddlewareQueryBus struct {
	bus        QueryBus
	middleware []QueryMiddleware
}

// NewMiddlewareQueryBus creates a query bus with middleware support.
func NewMiddlewareQueryBus(bus QueryBus, middleware ...QueryMiddleware) *MiddlewareQueryBus {
	return &MiddlewareQueryBus{
		bus:        bus,
		middleware: middleware,
	}
}

// Dispatch sends a query through middleware chain then to handler.
func (b *MiddlewareQueryBus) Dispatch(ctx context.Context, query Query) (any, error) {
	// Build the handler chain
	handler := QueryHandlerFunc[Query, any](func(ctx context.Context, q Query) (any, error) {
		return b.bus.Dispatch(ctx, q)
	})

	// Apply middleware in reverse order
	for i := len(b.middleware) - 1; i >= 0; i-- {
		handler = b.middleware[i](handler)
	}

	return handler(ctx, query)
}

// Register delegates to the underlying bus.
func (b *MiddlewareQueryBus) Register(queryType string, handler any) error {
	return b.bus.Register(queryType, handler)
}

// ============================================================================
// Common Middleware
// ============================================================================

// QueryLoggingMiddleware logs query execution.
func QueryLoggingMiddleware(logFn func(ctx context.Context, queryName string, duration time.Duration, err error)) QueryMiddleware {
	return func(next QueryHandlerFunc[Query, any]) QueryHandlerFunc[Query, any] {
		return func(ctx context.Context, query Query) (any, error) {
			start := time.Now()
			result, err := next(ctx, query)
			logFn(ctx, query.QueryName(), time.Since(start), err)
			return result, err
		}
	}
}

// QueryRecoveryMiddleware recovers from panics.
func QueryRecoveryMiddleware() QueryMiddleware {
	return func(next QueryHandlerFunc[Query, any]) QueryHandlerFunc[Query, any] {
		return func(ctx context.Context, query Query) (result any, err error) {
			defer func() {
				if r := recover(); r != nil {
					err = ErrHandlerPanic("QueryMiddleware", r)
				}
			}()
			return next(ctx, query)
		}
	}
}

// QueryTimeoutMiddleware adds timeout to queries.
func QueryTimeoutMiddleware(timeout time.Duration) QueryMiddleware {
	return func(next QueryHandlerFunc[Query, any]) QueryHandlerFunc[Query, any] {
		return func(ctx context.Context, query Query) (any, error) {
			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			resultCh := make(chan any, 1)
			errCh := make(chan error, 1)

			go func() {
				result, err := next(ctx, query)
				if err != nil {
					errCh <- err
					return
				}
				resultCh <- result
			}()

			select {
			case result := <-resultCh:
				return result, nil
			case err := <-errCh:
				return nil, err
			case <-ctx.Done():
				return nil, ErrQueryTimeout("QueryMiddleware", query.QueryName())
			}
		}
	}
}

// QueryCachingMiddleware caches query results.
func QueryCachingMiddleware(cache QueryCache) QueryMiddleware {
	return func(next QueryHandlerFunc[Query, any]) QueryHandlerFunc[Query, any] {
		return func(ctx context.Context, query Query) (any, error) {
			// Check if query is cacheable
			cacheable, ok := query.(CacheableQuery)
			if !ok {
				return next(ctx, query)
			}

			// Try to get from cache
			key := cacheable.CacheKey()
			if cached, found := cache.Get(ctx, key); found {
				return cached, nil
			}

			// Execute query
			result, err := next(ctx, query)
			if err != nil {
				return nil, err
			}

			// Store in cache
			cache.Set(ctx, key, result, cacheable.CacheTTL())

			return result, nil
		}
	}
}

// ============================================================================
// Caching Support
// ============================================================================

// CacheableQuery is implemented by queries that can be cached.
type CacheableQuery interface {
	Query
	CacheKey() string
	CacheTTL() time.Duration
}

// QueryCache is the interface for query result caching.
type QueryCache interface {
	Get(ctx context.Context, key string) (any, bool)
	Set(ctx context.Context, key string, value any, ttl time.Duration)
	Delete(ctx context.Context, key string)
}

// ============================================================================
// Helper Types
// ============================================================================

// BaseQuery provides common query functionality.
type BaseQuery struct {
	name string
}

// NewBaseQuery creates a new base query.
func NewBaseQuery(name string) BaseQuery {
	return BaseQuery{name: name}
}

// QueryName returns the query name.
func (q BaseQuery) QueryName() string {
	return q.name
}

// ============================================================================
// Typed Dispatch Helpers
// ============================================================================

// DispatchQuery is a typed helper for dispatching queries.
func DispatchQuery[R any](ctx context.Context, bus QueryBus, query Query) (R, error) {
	var zero R

	result, err := bus.Dispatch(ctx, query)
	if err != nil {
		return zero, err
	}

	if result == nil {
		return zero, nil
	}

	typed, ok := result.(R)
	if !ok {
		return zero, ErrQueryFailed("DispatchQuery", nil)
	}

	return typed, nil
}
