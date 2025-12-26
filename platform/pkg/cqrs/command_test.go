package cqrs_test

import (
	"context"
	"errors"
	"testing"

	"github.com/0xsj/nexus/platform/pkg/cqrs"
)

// ============================================================================
// Test Commands
// ============================================================================

type CreateUserCommand struct {
	cqrs.BaseCommand
	UserID string
	Email  string
}

func NewCreateUserCommand(userID, email string) *CreateUserCommand {
	return &CreateUserCommand{
		BaseCommand: cqrs.NewBaseCommand("CreateUser"),
		UserID:      userID,
		Email:       email,
	}
}

func (c *CreateUserCommand) CommandName() string {
	return "CreateUser"
}

type DeleteUserCommand struct {
	cqrs.BaseCommand
	UserID string
}

func NewDeleteUserCommand(userID string) *DeleteUserCommand {
	return &DeleteUserCommand{
		BaseCommand: cqrs.NewBaseCommand("DeleteUser"),
		UserID:      userID,
	}
}

func (c *DeleteUserCommand) CommandName() string {
	return "DeleteUser"
}

type ValidatedCommand struct {
	cqrs.BaseCommand
	Value string
}

func NewValidatedCommand(value string) *ValidatedCommand {
	return &ValidatedCommand{
		BaseCommand: cqrs.NewBaseCommand("ValidatedCommand"),
		Value:       value,
	}
}

func (c *ValidatedCommand) CommandName() string {
	return "ValidatedCommand"
}

func (c *ValidatedCommand) Validate() error {
	if c.Value == "" {
		return errors.New("value is required")
	}
	return nil
}

// ============================================================================
// Test Handlers
// ============================================================================

type CreateUserHandler struct {
	called    bool
	lastCmd   *CreateUserCommand
	returnID  string
	returnErr error
}

func (h *CreateUserHandler) Handle(ctx context.Context, cmd *CreateUserCommand) (*cqrs.CommandResult, error) {
	h.called = true
	h.lastCmd = cmd

	if h.returnErr != nil {
		return nil, h.returnErr
	}

	return &cqrs.CommandResult{
		ID:      h.returnID,
		Version: 1,
	}, nil
}

type DeleteUserHandler struct {
	called  bool
	lastCmd *DeleteUserCommand
}

func (h *DeleteUserHandler) Handle(ctx context.Context, cmd *DeleteUserCommand) (*cqrs.CommandResult, error) {
	h.called = true
	h.lastCmd = cmd

	return &cqrs.CommandResult{
		ID:      cmd.UserID,
		Version: 1,
	}, nil
}

type ValidatedCommandHandler struct {
	called bool
}

func (h *ValidatedCommandHandler) Handle(ctx context.Context, cmd *ValidatedCommand) (*cqrs.CommandResult, error) {
	h.called = true
	return &cqrs.CommandResult{ID: "ok"}, nil
}

type PanicHandler struct{}

func (h *PanicHandler) Handle(ctx context.Context, cmd *CreateUserCommand) (*cqrs.CommandResult, error) {
	panic("handler panic!")
}

// ============================================================================
// Command Bus Tests
// ============================================================================

func TestNewCommandBus(t *testing.T) {
	bus := cqrs.NewCommandBus()

	if bus == nil {
		t.Fatal("expected non-nil command bus")
	}
}

func TestCommandBus_Register(t *testing.T) {
	bus := cqrs.NewCommandBus()
	handler := &CreateUserHandler{}

	err := bus.Register("CreateUser", handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCommandBus_Register_EmptyType(t *testing.T) {
	bus := cqrs.NewCommandBus()
	handler := &CreateUserHandler{}

	err := bus.Register("", handler)
	if err == nil {
		t.Fatal("expected error for empty command type")
	}

	if !cqrs.IsCommandValidation(err) {
		t.Errorf("expected command validation error, got: %v", err)
	}
}

func TestCommandBus_Register_NilHandler(t *testing.T) {
	bus := cqrs.NewCommandBus()

	err := bus.Register("CreateUser", nil)
	if err == nil {
		t.Fatal("expected error for nil handler")
	}

	if !cqrs.IsCommandValidation(err) {
		t.Errorf("expected command validation error, got: %v", err)
	}
}

func TestCommandBus_Dispatch(t *testing.T) {
	bus := cqrs.NewCommandBus()
	handler := &CreateUserHandler{returnID: "user-123"}

	err := bus.Register("CreateUser", handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cmd := NewCreateUserCommand("user-123", "test@example.com")
	result, err := bus.Dispatch(context.Background(), cmd)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !handler.called {
		t.Error("expected handler to be called")
	}

	if handler.lastCmd.UserID != "user-123" {
		t.Errorf("expected UserID 'user-123', got '%s'", handler.lastCmd.UserID)
	}

	if handler.lastCmd.Email != "test@example.com" {
		t.Errorf("expected Email 'test@example.com', got '%s'", handler.lastCmd.Email)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if result.ID != "user-123" {
		t.Errorf("expected result ID 'user-123', got '%s'", result.ID)
	}

	if result.Version != 1 {
		t.Errorf("expected result Version 1, got %d", result.Version)
	}
}

func TestCommandBus_Dispatch_NilCommand(t *testing.T) {
	bus := cqrs.NewCommandBus()

	_, err := bus.Dispatch(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil command")
	}

	if !cqrs.IsCommandValidation(err) {
		t.Errorf("expected command validation error, got: %v", err)
	}
}

func TestCommandBus_Dispatch_HandlerNotFound(t *testing.T) {
	bus := cqrs.NewCommandBus()

	cmd := NewCreateUserCommand("user-123", "test@example.com")
	_, err := bus.Dispatch(context.Background(), cmd)

	if err == nil {
		t.Fatal("expected error for missing handler")
	}

	if !cqrs.IsCommandNotFound(err) {
		t.Errorf("expected command not found error, got: %v", err)
	}
}

func TestCommandBus_Dispatch_HandlerReturnsError(t *testing.T) {
	bus := cqrs.NewCommandBus()
	expectedErr := errors.New("handler error")
	handler := &CreateUserHandler{returnErr: expectedErr}

	err := bus.Register("CreateUser", handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cmd := NewCreateUserCommand("user-123", "test@example.com")
	_, err = bus.Dispatch(context.Background(), cmd)

	if err == nil {
		t.Fatal("expected error from handler")
	}

	if err.Error() != expectedErr.Error() {
		t.Errorf("expected error '%v', got '%v'", expectedErr, err)
	}
}

func TestCommandBus_Dispatch_HandlerPanics(t *testing.T) {
	bus := cqrs.NewCommandBus()
	handler := &PanicHandler{}

	err := bus.Register("CreateUser", handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cmd := NewCreateUserCommand("user-123", "test@example.com")
	_, err = bus.Dispatch(context.Background(), cmd)

	if err == nil {
		t.Fatal("expected error from panic")
	}

	if !cqrs.IsHandlerNotFound(err) && !errors.Is(err, err) {
		// Panic should be recovered and converted to error
		t.Logf("panic recovered as error: %v", err)
	}
}

func TestCommandBus_Dispatch_MultipleCommands(t *testing.T) {
	bus := cqrs.NewCommandBus()

	createHandler := &CreateUserHandler{returnID: "user-123"}
	deleteHandler := &DeleteUserHandler{}

	bus.Register("CreateUser", createHandler)
	bus.Register("DeleteUser", deleteHandler)

	// Dispatch CreateUser
	createCmd := NewCreateUserCommand("user-123", "test@example.com")
	_, err := bus.Dispatch(context.Background(), createCmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !createHandler.called {
		t.Error("expected create handler to be called")
	}

	// Dispatch DeleteUser
	deleteCmd := NewDeleteUserCommand("user-123")
	_, err = bus.Dispatch(context.Background(), deleteCmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !deleteHandler.called {
		t.Error("expected delete handler to be called")
	}
}

// ============================================================================
// RegisterCommand Helper Tests
// ============================================================================

func TestRegisterCommand(t *testing.T) {
	bus := cqrs.NewCommandBus()
	handler := &CreateUserHandler{returnID: "user-123"}

	err := cqrs.RegisterCommand[*CreateUserCommand](bus, handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cmd := NewCreateUserCommand("user-123", "test@example.com")
	result, err := bus.Dispatch(context.Background(), cmd)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != "user-123" {
		t.Errorf("expected result ID 'user-123', got '%s'", result.ID)
	}
}

// ============================================================================
// Middleware Tests
// ============================================================================

func TestMiddlewareCommandBus(t *testing.T) {
	baseBus := cqrs.NewCommandBus()
	handler := &CreateUserHandler{returnID: "user-123"}
	baseBus.Register("CreateUser", handler)

	// Track middleware calls
	var middlewareCalls []string

	middleware1 := func(next cqrs.CommandHandlerFunc[cqrs.Command]) cqrs.CommandHandlerFunc[cqrs.Command] {
		return func(ctx context.Context, cmd cqrs.Command) (*cqrs.CommandResult, error) {
			middlewareCalls = append(middlewareCalls, "middleware1-before")
			result, err := next(ctx, cmd)
			middlewareCalls = append(middlewareCalls, "middleware1-after")
			return result, err
		}
	}

	middleware2 := func(next cqrs.CommandHandlerFunc[cqrs.Command]) cqrs.CommandHandlerFunc[cqrs.Command] {
		return func(ctx context.Context, cmd cqrs.Command) (*cqrs.CommandResult, error) {
			middlewareCalls = append(middlewareCalls, "middleware2-before")
			result, err := next(ctx, cmd)
			middlewareCalls = append(middlewareCalls, "middleware2-after")
			return result, err
		}
	}

	bus := cqrs.NewMiddlewareCommandBus(baseBus, middleware1, middleware2)

	cmd := NewCreateUserCommand("user-123", "test@example.com")
	_, err := bus.Dispatch(context.Background(), cmd)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedCalls := []string{
		"middleware1-before",
		"middleware2-before",
		"middleware2-after",
		"middleware1-after",
	}

	if len(middlewareCalls) != len(expectedCalls) {
		t.Fatalf("expected %d middleware calls, got %d", len(expectedCalls), len(middlewareCalls))
	}

	for i, expected := range expectedCalls {
		if middlewareCalls[i] != expected {
			t.Errorf("middleware call %d: expected '%s', got '%s'", i, expected, middlewareCalls[i])
		}
	}
}

func TestLoggingMiddleware(t *testing.T) {
	baseBus := cqrs.NewCommandBus()
	handler := &CreateUserHandler{returnID: "user-123"}
	baseBus.Register("CreateUser", handler)

	var loggedCmd string
	var loggedErr error

	logFn := func(ctx context.Context, cmdName string, err error) {
		loggedCmd = cmdName
		loggedErr = err
	}

	bus := cqrs.NewMiddlewareCommandBus(baseBus, cqrs.LoggingMiddleware(logFn))

	cmd := NewCreateUserCommand("user-123", "test@example.com")
	_, err := bus.Dispatch(context.Background(), cmd)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if loggedCmd != "CreateUser" {
		t.Errorf("expected logged command 'CreateUser', got '%s'", loggedCmd)
	}

	if loggedErr != nil {
		t.Errorf("expected no logged error, got: %v", loggedErr)
	}
}

func TestRecoveryMiddleware(t *testing.T) {
	baseBus := cqrs.NewCommandBus()
	handler := &PanicHandler{}
	baseBus.Register("CreateUser", handler)

	bus := cqrs.NewMiddlewareCommandBus(baseBus, cqrs.RecoveryMiddleware())

	cmd := NewCreateUserCommand("user-123", "test@example.com")
	_, err := bus.Dispatch(context.Background(), cmd)

	if err == nil {
		t.Fatal("expected error from recovered panic")
	}
}

func TestValidationMiddleware(t *testing.T) {
	baseBus := cqrs.NewCommandBus()
	handler := &ValidatedCommandHandler{}
	baseBus.Register("ValidatedCommand", handler)

	bus := cqrs.NewMiddlewareCommandBus(baseBus, cqrs.ValidationMiddleware())

	// Test with invalid command
	invalidCmd := NewValidatedCommand("")
	_, err := bus.Dispatch(context.Background(), invalidCmd)

	if err == nil {
		t.Fatal("expected validation error")
	}

	if handler.called {
		t.Error("handler should not be called for invalid command")
	}

	// Test with valid command
	handler.called = false
	validCmd := NewValidatedCommand("valid-value")
	_, err = bus.Dispatch(context.Background(), validCmd)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !handler.called {
		t.Error("handler should be called for valid command")
	}
}

// ============================================================================
// CommandHandlerFunc Tests
// ============================================================================

func TestCommandHandlerFunc(t *testing.T) {
	handlerFn := cqrs.CommandHandlerFunc[*CreateUserCommand](func(ctx context.Context, cmd *CreateUserCommand) (*cqrs.CommandResult, error) {
		return &cqrs.CommandResult{
			ID:      cmd.UserID,
			Version: 1,
		}, nil
	})

	cmd := NewCreateUserCommand("user-456", "func@example.com")
	result, err := handlerFn.Handle(context.Background(), cmd)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != "user-456" {
		t.Errorf("expected ID 'user-456', got '%s'", result.ID)
	}
}

// ============================================================================
// BaseCommand Tests
// ============================================================================

func TestBaseCommand(t *testing.T) {
	base := cqrs.NewBaseCommand("TestCommand")

	if base.CommandName() != "TestCommand" {
		t.Errorf("expected command name 'TestCommand', got '%s'", base.CommandName())
	}
}

// ============================================================================
// Error Checker Tests
// ============================================================================

func TestIsCommandValidation(t *testing.T) {
	err := cqrs.ErrCommandValidation("test", "invalid")

	if !cqrs.IsCommandValidation(err) {
		t.Error("expected IsCommandValidation to return true")
	}

	if cqrs.IsCommandValidation(nil) {
		t.Error("expected IsCommandValidation(nil) to return false")
	}

	if cqrs.IsCommandValidation(errors.New("other error")) {
		t.Error("expected IsCommandValidation to return false for other errors")
	}
}

func TestIsCommandNotFound(t *testing.T) {
	err := cqrs.ErrCommandNotFound("test", "UnknownCommand")

	if !cqrs.IsCommandNotFound(err) {
		t.Error("expected IsCommandNotFound to return true")
	}

	if cqrs.IsCommandNotFound(nil) {
		t.Error("expected IsCommandNotFound(nil) to return false")
	}
}

func TestIsHandlerNotFound(t *testing.T) {
	err := cqrs.ErrHandlerNotFound("test", "UnknownHandler")

	if !cqrs.IsHandlerNotFound(err) {
		t.Error("expected IsHandlerNotFound to return true")
	}

	if cqrs.IsHandlerNotFound(nil) {
		t.Error("expected IsHandlerNotFound(nil) to return false")
	}
}
