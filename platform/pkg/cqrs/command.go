package cqrs

import (
	"context"
	"reflect"
	"sync"
)

// ============================================================================
// Command Interface
// ============================================================================

// Command represents an intent to change the system state.
// Commands are named in imperative form: IssueCredential, RevokeCredential.
type Command interface {
	// CommandName returns the unique name of the command.
	CommandName() string
}

// CommandResult represents the result of a command execution.
type CommandResult struct {
	// ID is the identifier of the created/affected resource.
	ID string

	// Version is the new version after the command.
	Version int

	// Data contains any additional result data.
	Data any
}

// ============================================================================
// Command Handler
// ============================================================================

// CommandHandler handles a specific command type.
type CommandHandler[C Command] interface {
	Handle(ctx context.Context, cmd C) (*CommandResult, error)
}

// CommandHandlerFunc is a function adapter for CommandHandler.
type CommandHandlerFunc[C Command] func(ctx context.Context, cmd C) (*CommandResult, error)

// Handle implements CommandHandler.
func (f CommandHandlerFunc[C]) Handle(ctx context.Context, cmd C) (*CommandResult, error) {
	return f(ctx, cmd)
}

// ============================================================================
// Command Bus
// ============================================================================

// CommandBus dispatches commands to their handlers.
type CommandBus interface {
	// Dispatch sends a command to its handler.
	Dispatch(ctx context.Context, cmd Command) (*CommandResult, error)

	// Register registers a handler for a command type.
	Register(cmdType string, handler any) error
}

// ============================================================================
// In-Memory Command Bus Implementation
// ============================================================================

// InMemoryCommandBus is a simple in-memory command bus.
type InMemoryCommandBus struct {
	mu       sync.RWMutex
	handlers map[string]any
}

// NewCommandBus creates a new in-memory command bus.
func NewCommandBus() *InMemoryCommandBus {
	return &InMemoryCommandBus{
		handlers: make(map[string]any),
	}
}

// Register registers a handler for a command type.
func (b *InMemoryCommandBus) Register(cmdType string, handler any) error {
	const op = "CommandBus.Register"

	if cmdType == "" {
		return ErrCommandValidation(op, "command type cannot be empty")
	}
	if handler == nil {
		return ErrCommandValidation(op, "handler cannot be nil")
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	b.handlers[cmdType] = handler
	return nil
}

// RegisterCommand registers a typed command handler.
func RegisterCommand[C Command](bus *InMemoryCommandBus, handler CommandHandler[C]) error {
	var zero C
	return bus.Register(zero.CommandName(), handler)
}

// Dispatch sends a command to its handler.
func (b *InMemoryCommandBus) Dispatch(ctx context.Context, cmd Command) (*CommandResult, error) {
	const op = "CommandBus.Dispatch"

	if cmd == nil {
		return nil, ErrCommandValidation(op, "command cannot be nil")
	}

	cmdType := cmd.CommandName()

	b.mu.RLock()
	handler, exists := b.handlers[cmdType]
	b.mu.RUnlock()

	if !exists {
		return nil, ErrCommandNotFound(op, cmdType)
	}

	// Use reflection to call the handler
	return b.invokeHandler(ctx, handler, cmd)
}

// invokeHandler calls the handler using reflection.
func (b *InMemoryCommandBus) invokeHandler(ctx context.Context, handler any, cmd Command) (result *CommandResult, err error) {
	const op = "CommandBus.invokeHandler"

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

	// Call Handle(ctx, cmd)
	results := handleMethod.Call([]reflect.Value{
		reflect.ValueOf(ctx),
		reflect.ValueOf(cmd),
	})

	// Parse results
	if len(results) != 2 {
		return nil, ErrCommandFailed(op, nil)
	}

	// First result: *CommandResult
	if !results[0].IsNil() {
		result = results[0].Interface().(*CommandResult)
	}

	// Second result: error
	if !results[1].IsNil() {
		err = results[1].Interface().(error)
	}

	return result, err
}

// ============================================================================
// Command Middleware
// ============================================================================

// CommandMiddleware wraps command handling with additional behavior.
type CommandMiddleware func(next CommandHandlerFunc[Command]) CommandHandlerFunc[Command]

// MiddlewareCommandBus wraps a command bus with middleware.
type MiddlewareCommandBus struct {
	bus        CommandBus
	middleware []CommandMiddleware
}

// NewMiddlewareCommandBus creates a command bus with middleware support.
func NewMiddlewareCommandBus(bus CommandBus, middleware ...CommandMiddleware) *MiddlewareCommandBus {
	return &MiddlewareCommandBus{
		bus:        bus,
		middleware: middleware,
	}
}

// Dispatch sends a command through middleware chain then to handler.
func (b *MiddlewareCommandBus) Dispatch(ctx context.Context, cmd Command) (*CommandResult, error) {
	// Build the handler chain
	handler := CommandHandlerFunc[Command](func(ctx context.Context, c Command) (*CommandResult, error) {
		return b.bus.Dispatch(ctx, c)
	})

	// Apply middleware in reverse order
	for i := len(b.middleware) - 1; i >= 0; i-- {
		handler = b.middleware[i](handler)
	}

	return handler(ctx, cmd)
}

// Register delegates to the underlying bus.
func (b *MiddlewareCommandBus) Register(cmdType string, handler any) error {
	return b.bus.Register(cmdType, handler)
}

// ============================================================================
// Common Middleware
// ============================================================================

// LoggingMiddleware logs command execution.
func LoggingMiddleware(logFn func(ctx context.Context, cmdName string, err error)) CommandMiddleware {
	return func(next CommandHandlerFunc[Command]) CommandHandlerFunc[Command] {
		return func(ctx context.Context, cmd Command) (*CommandResult, error) {
			result, err := next(ctx, cmd)
			logFn(ctx, cmd.CommandName(), err)
			return result, err
		}
	}
}

// RecoveryMiddleware recovers from panics.
func RecoveryMiddleware() CommandMiddleware {
	return func(next CommandHandlerFunc[Command]) CommandHandlerFunc[Command] {
		return func(ctx context.Context, cmd Command) (result *CommandResult, err error) {
			defer func() {
				if r := recover(); r != nil {
					err = ErrHandlerPanic("CommandMiddleware", r)
				}
			}()
			return next(ctx, cmd)
		}
	}
}

// ValidationMiddleware validates commands before execution.
func ValidationMiddleware() CommandMiddleware {
	return func(next CommandHandlerFunc[Command]) CommandHandlerFunc[Command] {
		return func(ctx context.Context, cmd Command) (*CommandResult, error) {
			// Check if command implements Validatable
			if v, ok := cmd.(Validatable); ok {
				if err := v.Validate(); err != nil {
					return nil, ErrCommandValidation("CommandMiddleware", err.Error())
				}
			}
			return next(ctx, cmd)
		}
	}
}

// Validatable is implemented by commands that can validate themselves.
type Validatable interface {
	Validate() error
}

// ============================================================================
// Helper Types
// ============================================================================

// BaseCommand provides common command functionality.
type BaseCommand struct {
	name string
}

// NewBaseCommand creates a new base command.
func NewBaseCommand(name string) BaseCommand {
	return BaseCommand{name: name}
}

// CommandName returns the command name.
func (c BaseCommand) CommandName() string {
	return c.name
}
