package events

import "context"

// Handler processes events.
// Handlers should be idempotent as events may be delivered multiple times.
type Handler interface {
	// Handle processes the event.
	// Returns an error if the event cannot be processed.
	Handle(ctx context.Context, event Event) error

	// HandlerName returns a unique name for this handler.
	// Used for logging and monitoring.
	HandlerName() string
}

// HandlerFunc is a function adapter for Handler interface.
type HandlerFunc func(ctx context.Context, event Event) error

// Handle implements Handler interface.
func (f HandlerFunc) Handle(ctx context.Context, event Event) error {
	return f(ctx, event)
}

// HandlerName returns the function name.
func (f HandlerFunc) HandlerName() string {
	return "HandlerFunc"
}

// NamedHandlerFunc creates a handler with a custom name.
// This is useful when you need multiple handlers with distinct names,
// especially with JetStream durable subscriptions.
//
// Example:
//
//	handler := events.NamedHandlerFunc("user-created-handler", func(ctx context.Context, event events.Event) error {
//	    // Handle user.created event
//	    return nil
//	})
func NamedHandlerFunc(name string, fn func(ctx context.Context, event Event) error) Handler {
	return &namedHandlerFunc{
		name: name,
		fn:   fn,
	}
}

// namedHandlerFunc is a handler with a custom name.
type namedHandlerFunc struct {
	name string
	fn   func(context.Context, Event) error
}

// Handle implements Handler interface.
func (h *namedHandlerFunc) Handle(ctx context.Context, event Event) error {
	return h.fn(ctx, event)
}

// HandlerName returns the custom handler name.
func (h *namedHandlerFunc) HandlerName() string {
	return h.name
}

// Middleware wraps a Handler with additional behavior.
type Middleware func(Handler) Handler

// Chain applies multiple middlewares to a handler.
func Chain(h Handler, middlewares ...Middleware) Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}
