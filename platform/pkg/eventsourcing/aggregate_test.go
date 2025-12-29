package eventsourcing_test

import (
	"context"
	"testing"

	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing/memory"
)

// ============================================================================
// Test Aggregate
// ============================================================================

type UserAggregate struct {
	eventsourcing.AggregateRoot
	email  string
	name   string
	status string
}

func NewUserAggregate(id string) *UserAggregate {
	u := &UserAggregate{}
	u.InitAggregate("User", id)
	return u
}

func (u *UserAggregate) Create(email, name string) error {
	if u.status != "" {
		return eventsourcing.ErrAggregateValidation("User.Create", "user already exists")
	}

	u.AggregateRoot.Raise(u, NewUserCreatedEvent(u.AggregateID(), u.AggregateID(), email, name))
	return nil
}

func (u *UserAggregate) ChangeEmail(newEmail string) error {
	if u.status == "" {
		return eventsourcing.ErrAggregateValidation("User.ChangeEmail", "user does not exist")
	}

	if u.status == "deleted" {
		return eventsourcing.ErrAggregateValidation("User.ChangeEmail", "user is deleted")
	}

	oldEmail := u.email
	u.AggregateRoot.Raise(u, NewUserEmailChangedEvent(u.AggregateID(), oldEmail, newEmail))
	return nil
}

func (u *UserAggregate) Delete(reason string) error {
	if u.status == "" {
		return eventsourcing.ErrAggregateValidation("User.Delete", "user does not exist")
	}

	if u.status == "deleted" {
		return eventsourcing.ErrAggregateValidation("User.Delete", "user already deleted")
	}

	u.AggregateRoot.Raise(u, NewUserDeletedEvent(u.AggregateID(), reason))
	return nil
}

func (u *UserAggregate) ApplyEvent(event eventsourcing.Event) {
	switch e := event.(type) {
	case *UserCreatedEvent:
		u.email = e.Email
		u.name = e.Name
		u.status = "active"
	case *UserEmailChangedEvent:
		u.email = e.NewEmail
	case *UserDeletedEvent:
		u.status = "deleted"
	}
}

func (u *UserAggregate) GetAggregateRoot() *eventsourcing.AggregateRoot {
	return &u.AggregateRoot
}

func (u *UserAggregate) Email() string {
	return u.email
}

func (u *UserAggregate) Name() string {
	return u.name
}

func (u *UserAggregate) Status() string {
	return u.status
}

// ============================================================================
// AggregateRoot Tests
// ============================================================================

func TestAggregateRoot_InitAggregate(t *testing.T) {
	user := NewUserAggregate("user-123")

	if user.AggregateID() != "user-123" {
		t.Errorf("expected aggregate ID 'user-123', got '%s'", user.AggregateID())
	}

	if user.AggregateType() != "User" {
		t.Errorf("expected aggregate type 'User', got '%s'", user.AggregateType())
	}

	if user.Version() != 0 {
		t.Errorf("expected version 0, got %d", user.Version())
	}

	if user.HasChanges() {
		t.Error("expected no changes")
	}
}

func TestAggregateRoot_Raise(t *testing.T) {
	user := NewUserAggregate("user-123")

	err := user.Create("test@example.com", "Test User")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if user.Version() != 1 {
		t.Errorf("expected version 1, got %d", user.Version())
	}

	if !user.HasChanges() {
		t.Error("expected changes")
	}

	changes := user.Changes()
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}

	if changes[0].EventType() != "User.Created" {
		t.Errorf("expected event type 'User.Created', got '%s'", changes[0].EventType())
	}

	// State should be applied
	if user.Email() != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got '%s'", user.Email())
	}

	if user.Name() != "Test User" {
		t.Errorf("expected name 'Test User', got '%s'", user.Name())
	}

	if user.Status() != "active" {
		t.Errorf("expected status 'active', got '%s'", user.Status())
	}
}

func TestAggregateRoot_MultipleEvents(t *testing.T) {
	user := NewUserAggregate("user-123")

	// Create user
	err := user.Create("test@example.com", "Test User")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Change email
	err = user.ChangeEmail("new@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if user.Version() != 2 {
		t.Errorf("expected version 2, got %d", user.Version())
	}

	changes := user.Changes()
	if len(changes) != 2 {
		t.Fatalf("expected 2 changes, got %d", len(changes))
	}

	if user.Email() != "new@example.com" {
		t.Errorf("expected email 'new@example.com', got '%s'", user.Email())
	}
}

func TestAggregateRoot_ClearChanges(t *testing.T) {
	user := NewUserAggregate("user-123")

	user.Create("test@example.com", "Test User")

	if !user.HasChanges() {
		t.Error("expected changes before clear")
	}

	user.ClearChanges()

	if user.HasChanges() {
		t.Error("expected no changes after clear")
	}

	// Version should remain
	if user.Version() != 1 {
		t.Errorf("expected version 1 after clear, got %d", user.Version())
	}
}

func TestAggregateRoot_ValidationErrors(t *testing.T) {
	user := NewUserAggregate("user-123")

	// Try to change email before creating
	err := user.ChangeEmail("new@example.com")
	if err == nil {
		t.Fatal("expected error for changing email before create")
	}

	// Create user
	user.Create("test@example.com", "Test User")

	// Try to create again
	err = user.Create("another@example.com", "Another User")
	if err == nil {
		t.Fatal("expected error for creating again")
	}

	// Delete user
	user.Delete("requested")

	// Try to change email after delete
	err = user.ChangeEmail("after-delete@example.com")
	if err == nil {
		t.Fatal("expected error for changing email after delete")
	}

	// Try to delete again
	err = user.Delete("again")
	if err == nil {
		t.Fatal("expected error for deleting again")
	}
}

// ============================================================================
// Hydration Tests
// ============================================================================

func TestHydrate(t *testing.T) {
	// Create events
	events := []eventsourcing.Event{
		NewUserCreatedEvent("user-123", "user-123", "test@example.com", "Test User"),
		NewUserEmailChangedEvent("user-123", "test@example.com", "new@example.com"),
	}

	// Hydrate new aggregate
	user := NewUserAggregate("user-123")
	eventsourcing.Hydrate(user, events)

	if user.Version() != 2 {
		t.Errorf("expected version 2, got %d", user.Version())
	}

	if user.Email() != "new@example.com" {
		t.Errorf("expected email 'new@example.com', got '%s'", user.Email())
	}

	if user.Name() != "Test User" {
		t.Errorf("expected name 'Test User', got '%s'", user.Name())
	}

	if user.Status() != "active" {
		t.Errorf("expected status 'active', got '%s'", user.Status())
	}

	// Should have no uncommitted changes
	if user.HasChanges() {
		t.Error("expected no uncommitted changes after hydration")
	}
}

func TestHydrateFromEnvelopes(t *testing.T) {
	registry := eventsourcing.NewEventRegistry()
	registry.Register("User.Created", func() eventsourcing.Event {
		return &UserCreatedEvent{}
	})
	registry.Register("User.EmailChanged", func() eventsourcing.Event {
		return &UserEmailChangedEvent{}
	})

	// Create envelopes
	event1 := NewUserCreatedEvent("user-123", "user-123", "test@example.com", "Test User")
	envelope1, _ := eventsourcing.NewEventEnvelope(event1, 1)

	event2 := NewUserEmailChangedEvent("user-123", "test@example.com", "new@example.com")
	envelope2, _ := eventsourcing.NewEventEnvelope(event2, 2)

	envelopes := []*eventsourcing.EventEnvelope{envelope1, envelope2}

	// Hydrate
	user := NewUserAggregate("user-123")
	err := eventsourcing.HydrateFromEnvelopes(user, envelopes, registry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if user.Version() != 2 {
		t.Errorf("expected version 2, got %d", user.Version())
	}

	if user.Email() != "new@example.com" {
		t.Errorf("expected email 'new@example.com', got '%s'", user.Email())
	}
}

// ============================================================================
// AggregateFactory Tests
// ============================================================================

func TestAggregateFactoryFunc(t *testing.T) {
	factory := eventsourcing.AggregateFactoryFunc(func(id string) eventsourcing.Aggregate {
		return NewUserAggregate(id)
	})

	aggregate := factory.Create("user-456")

	if aggregate.AggregateID() != "user-456" {
		t.Errorf("expected ID 'user-456', got '%s'", aggregate.AggregateID())
	}

	if aggregate.AggregateType() != "User" {
		t.Errorf("expected type 'User', got '%s'", aggregate.AggregateType())
	}
}

// ============================================================================
// GenericRepository Tests
// ============================================================================

func TestGenericRepository_SaveAndLoad(t *testing.T) {
	store := memory.NewStore()
	registry := eventsourcing.NewEventRegistry()
	registry.Register("User.Created", func() eventsourcing.Event {
		return &UserCreatedEvent{}
	})
	registry.Register("User.EmailChanged", func() eventsourcing.Event {
		return &UserEmailChangedEvent{}
	})

	factory := eventsourcing.AggregateFactoryFunc(func(id string) eventsourcing.Aggregate {
		return NewUserAggregate(id)
	})

	repo := eventsourcing.NewGenericRepository[*UserAggregate](store, factory, registry)

	// Create and save
	user := NewUserAggregate("user-123")
	user.Create("test@example.com", "Test User")
	user.ChangeEmail("new@example.com")

	err := repo.Save(context.Background(), user)
	if err != nil {
		t.Fatalf("unexpected error saving: %v", err)
	}

	// Changes should be cleared
	if user.HasChanges() {
		t.Error("expected no changes after save")
	}

	// Load
	loaded, err := repo.Load(context.Background(), "user-123")
	if err != nil {
		t.Fatalf("unexpected error loading: %v", err)
	}

	if loaded.AggregateID() != "user-123" {
		t.Errorf("expected ID 'user-123', got '%s'", loaded.AggregateID())
	}

	if loaded.Version() != 2 {
		t.Errorf("expected version 2, got %d", loaded.Version())
	}

	if loaded.Email() != "new@example.com" {
		t.Errorf("expected email 'new@example.com', got '%s'", loaded.Email())
	}

	if loaded.Name() != "Test User" {
		t.Errorf("expected name 'Test User', got '%s'", loaded.Name())
	}
}

func TestGenericRepository_Load_NotFound(t *testing.T) {
	store := memory.NewStore()
	registry := eventsourcing.NewEventRegistry()
	factory := eventsourcing.AggregateFactoryFunc(func(id string) eventsourcing.Aggregate {
		return NewUserAggregate(id)
	})

	repo := eventsourcing.NewGenericRepository[*UserAggregate](store, factory, registry)

	_, err := repo.Load(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent aggregate")
	}

	if !eventsourcing.IsAggregateNotFound(err) {
		t.Errorf("expected aggregate not found error, got: %v", err)
	}
}

func TestGenericRepository_Exists(t *testing.T) {
	store := memory.NewStore()
	registry := eventsourcing.NewEventRegistry()
	registry.Register("User.Created", func() eventsourcing.Event {
		return &UserCreatedEvent{}
	})

	factory := eventsourcing.AggregateFactoryFunc(func(id string) eventsourcing.Aggregate {
		return NewUserAggregate(id)
	})

	repo := eventsourcing.NewGenericRepository[*UserAggregate](store, factory, registry)

	// Check nonexistent
	exists, err := repo.Exists(context.Background(), "user-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Error("expected not exists")
	}

	// Create and save
	user := NewUserAggregate("user-123")
	user.Create("test@example.com", "Test User")
	repo.Save(context.Background(), user)

	// Check exists
	exists, err = repo.Exists(context.Background(), "user-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Error("expected exists")
	}
}

func TestGenericRepository_Save_NoChanges(t *testing.T) {
	store := memory.NewStore()
	registry := eventsourcing.NewEventRegistry()
	factory := eventsourcing.AggregateFactoryFunc(func(id string) eventsourcing.Aggregate {
		return NewUserAggregate(id)
	})

	repo := eventsourcing.NewGenericRepository[*UserAggregate](store, factory, registry)

	// Save with no changes
	user := NewUserAggregate("user-123")
	err := repo.Save(context.Background(), user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Store should be empty
	if store.EventCount() != 0 {
		t.Errorf("expected 0 events, got %d", store.EventCount())
	}
}

func TestGenericRepository_ConcurrentModification(t *testing.T) {
	store := memory.NewStore()
	registry := eventsourcing.NewEventRegistry()
	registry.Register("User.Created", func() eventsourcing.Event {
		return &UserCreatedEvent{}
	})
	registry.Register("User.EmailChanged", func() eventsourcing.Event {
		return &UserEmailChangedEvent{}
	})

	factory := eventsourcing.AggregateFactoryFunc(func(id string) eventsourcing.Aggregate {
		return NewUserAggregate(id)
	})

	repo := eventsourcing.NewGenericRepository[*UserAggregate](store, factory, registry)

	// Create initial
	user1 := NewUserAggregate("user-123")
	user1.Create("test@example.com", "Test User")
	repo.Save(context.Background(), user1)

	// Load two copies
	userA, _ := repo.Load(context.Background(), "user-123")
	userB, _ := repo.Load(context.Background(), "user-123")

	// Modify and save A
	userA.ChangeEmail("a@example.com")
	err := repo.Save(context.Background(), userA)
	if err != nil {
		t.Fatalf("unexpected error saving A: %v", err)
	}

	// Modify and save B - should fail due to concurrency conflict
	userB.ChangeEmail("b@example.com")
	err = repo.Save(context.Background(), userB)
	if err == nil {
		t.Fatal("expected concurrency conflict error")
	}

	if !eventsourcing.IsConcurrencyConflict(err) {
		t.Errorf("expected concurrency conflict error, got: %v", err)
	}
}

// ============================================================================
// ValidatableAggregate Tests
// ============================================================================

type ValidatedUserAggregate struct {
	UserAggregate
}

func NewValidatedUserAggregate(id string) *ValidatedUserAggregate {
	v := &ValidatedUserAggregate{}
	v.InitAggregate("ValidatedUser", id)
	return v
}

func (v *ValidatedUserAggregate) Validate() error {
	if v.email == "" {
		return eventsourcing.ErrAggregateValidation("ValidatedUser.Validate", "email is required")
	}
	return nil
}

func TestValidateAggregate(t *testing.T) {
	// Valid aggregate
	user := NewValidatedUserAggregate("user-123")
	user.Create("test@example.com", "Test User")

	err := eventsourcing.ValidateAggregate(user)
	if err != nil {
		t.Errorf("unexpected validation error: %v", err)
	}

	// Invalid aggregate (no email)
	invalidUser := NewValidatedUserAggregate("user-456")
	err = eventsourcing.ValidateAggregate(invalidUser)
	if err == nil {
		t.Error("expected validation error for invalid aggregate")
	}
}

func TestValidateAggregate_NonValidatable(t *testing.T) {
	// Regular aggregate doesn't implement ValidatableAggregate
	user := NewUserAggregate("user-123")

	err := eventsourcing.ValidateAggregate(user)
	if err != nil {
		t.Errorf("expected no error for non-validatable aggregate, got: %v", err)
	}
}
