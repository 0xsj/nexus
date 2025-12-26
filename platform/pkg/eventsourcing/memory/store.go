package memory

import (
	"context"
	"sync"

	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
)

// ============================================================================
// In-Memory Event Store
// ============================================================================

// Store is an in-memory event store implementation.
// Useful for testing and development.
type Store struct {
	mu     sync.RWMutex
	events map[string][]*eventsourcing.EventEnvelope // aggregateID -> events
}

// NewStore creates a new in-memory event store.
func NewStore() *Store {
	return &Store{
		events: make(map[string][]*eventsourcing.EventEnvelope),
	}
}

// Save persists events for an aggregate.
func (s *Store) Save(ctx context.Context, aggregateID string, events []*eventsourcing.EventEnvelope, expectedVersion int) error {
	const op = "memory.Store.Save"

	if len(events) == 0 {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check for context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Get current events
	current := s.events[aggregateID]
	currentVersion := len(current)

	// Optimistic concurrency check
	if currentVersion != expectedVersion {
		return eventsourcing.ErrConcurrencyConflict(op, aggregateID, expectedVersion, currentVersion)
	}

	// Append events
	s.events[aggregateID] = append(current, events...)

	return nil
}

// Load retrieves all events for an aggregate.
func (s *Store) Load(ctx context.Context, aggregateID string) ([]*eventsourcing.EventEnvelope, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check for context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	events := s.events[aggregateID]
	if len(events) == 0 {
		return []*eventsourcing.EventEnvelope{}, nil
	}

	// Return deep copies to prevent external mutation
	result := make([]*eventsourcing.EventEnvelope, len(events))
	for i, e := range events {
		result[i] = copyEnvelope(e)
	}

	return result, nil
}

// LoadFrom retrieves events starting from a specific version.
func (s *Store) LoadFrom(ctx context.Context, aggregateID string, fromVersion int) ([]*eventsourcing.EventEnvelope, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check for context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	events := s.events[aggregateID]
	if len(events) == 0 || fromVersion >= len(events) {
		return []*eventsourcing.EventEnvelope{}, nil
	}

	// Filter events from version
	result := make([]*eventsourcing.EventEnvelope, 0)
	for _, event := range events {
		if event.Version >= fromVersion {
			result = append(result, copyEnvelope(event))
		}
	}

	return result, nil
}

// LoadByType retrieves all events for an aggregate type.
func (s *Store) LoadByType(ctx context.Context, aggregateType string, opts eventsourcing.LoadOptions) ([]*eventsourcing.EventEnvelope, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check for context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	result := make([]*eventsourcing.EventEnvelope, 0)

	for _, events := range s.events {
		for _, event := range events {
			if event.AggregateType != aggregateType {
				continue
			}

			// Apply filters
			if !s.matchesOptions(event, opts) {
				continue
			}

			result = append(result, copyEnvelope(event))

			// Check limit
			if opts.Limit > 0 && len(result) >= opts.Limit {
				return result, nil
			}
		}
	}

	return result, nil
}

// matchesOptions checks if an event matches the load options.
func (s *Store) matchesOptions(event *eventsourcing.EventEnvelope, opts eventsourcing.LoadOptions) bool {
	// Version filter
	if opts.FromVersion > 0 && event.Version < opts.FromVersion {
		return false
	}
	if opts.ToVersion > 0 && event.Version > opts.ToVersion {
		return false
	}

	// Time filter
	if opts.FromTime != nil && event.Timestamp.Before(*opts.FromTime) {
		return false
	}
	if opts.ToTime != nil && event.Timestamp.After(*opts.ToTime) {
		return false
	}

	// Event type filter
	if len(opts.EventTypes) > 0 {
		found := false
		for _, t := range opts.EventTypes {
			if event.Type == t {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// ============================================================================
// Extended Operations
// ============================================================================

// LoadAll retrieves all events across all aggregates.
func (s *Store) LoadAll(ctx context.Context, opts eventsourcing.LoadOptions) ([]*eventsourcing.EventEnvelope, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check for context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	result := make([]*eventsourcing.EventEnvelope, 0)

	for _, events := range s.events {
		for _, event := range events {
			if !s.matchesOptions(event, opts) {
				continue
			}

			result = append(result, copyEnvelope(event))

			if opts.Limit > 0 && len(result) >= opts.Limit {
				return result, nil
			}
		}
	}

	return result, nil
}

// Count returns the total number of events for an aggregate.
func (s *Store) Count(ctx context.Context, aggregateID string) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.events[aggregateID]), nil
}

// CountByType returns the total number of events for an aggregate type.
func (s *Store) CountByType(ctx context.Context, aggregateType string) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := 0
	for _, events := range s.events {
		for _, event := range events {
			if event.AggregateType == aggregateType {
				count++
			}
		}
	}

	return count, nil
}

// Delete removes all events for an aggregate.
func (s *Store) Delete(ctx context.Context, aggregateID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.events, aggregateID)
	return nil
}

// ============================================================================
// Testing Helpers
// ============================================================================

// Clear removes all events from the store.
func (s *Store) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.events = make(map[string][]*eventsourcing.EventEnvelope)
}

// AggregateIDs returns all aggregate IDs in the store.
func (s *Store) AggregateIDs() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ids := make([]string, 0, len(s.events))
	for id := range s.events {
		ids = append(ids, id)
	}

	return ids
}

// EventCount returns the total number of events in the store.
func (s *Store) EventCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := 0
	for _, events := range s.events {
		count += len(events)
	}

	return count
}

// ============================================================================
// In-Memory Snapshot Store
// ============================================================================

// SnapshotStore is an in-memory snapshot store implementation.
type SnapshotStore struct {
	mu        sync.RWMutex
	snapshots map[string]*eventsourcing.Snapshot
}

// NewSnapshotStore creates a new in-memory snapshot store.
func NewSnapshotStore() *SnapshotStore {
	return &SnapshotStore{
		snapshots: make(map[string]*eventsourcing.Snapshot),
	}
}

// Save saves a snapshot.
func (s *SnapshotStore) Save(ctx context.Context, snapshot *eventsourcing.Snapshot) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Store a copy
	stored := &eventsourcing.Snapshot{
		AggregateID:   snapshot.AggregateID,
		AggregateType: snapshot.AggregateType,
		Version:       snapshot.Version,
		Data:          make([]byte, len(snapshot.Data)),
		Timestamp:     snapshot.Timestamp,
	}
	copy(stored.Data, snapshot.Data)

	s.snapshots[snapshot.AggregateID] = stored
	return nil
}

// Load loads the latest snapshot for an aggregate.
func (s *SnapshotStore) Load(ctx context.Context, aggregateID string) (*eventsourcing.Snapshot, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	snapshot, exists := s.snapshots[aggregateID]
	if !exists {
		return nil, eventsourcing.ErrSnapshotNotFound("memory.SnapshotStore.Load", aggregateID)
	}

	// Return a copy
	result := &eventsourcing.Snapshot{
		AggregateID:   snapshot.AggregateID,
		AggregateType: snapshot.AggregateType,
		Version:       snapshot.Version,
		Data:          make([]byte, len(snapshot.Data)),
		Timestamp:     snapshot.Timestamp,
	}
	copy(result.Data, snapshot.Data)

	return result, nil
}

// Delete removes a snapshot.
func (s *SnapshotStore) Delete(ctx context.Context, aggregateID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.snapshots, aggregateID)
	return nil
}

// Clear removes all snapshots from the store.
func (s *SnapshotStore) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.snapshots = make(map[string]*eventsourcing.Snapshot)
}

// ============================================================================
// Helpers
// ============================================================================

// copyEnvelope creates a deep copy of an EventEnvelope.
func copyEnvelope(e *eventsourcing.EventEnvelope) *eventsourcing.EventEnvelope {
	if e == nil {
		return nil
	}

	// Copy data slice
	dataCopy := make([]byte, len(e.Data))
	copy(dataCopy, e.Data)

	// Copy metadata custom map
	var customCopy map[string]any
	if e.Metadata.Custom != nil {
		customCopy = make(map[string]any, len(e.Metadata.Custom))
		for k, v := range e.Metadata.Custom {
			customCopy[k] = v
		}
	}

	return &eventsourcing.EventEnvelope{
		ID:            e.ID,
		Type:          e.Type,
		AggregateID:   e.AggregateID,
		AggregateType: e.AggregateType,
		Version:       e.Version,
		Timestamp:     e.Timestamp,
		Data:          dataCopy,
		Metadata: eventsourcing.EventMetadata{
			CorrelationID: e.Metadata.CorrelationID,
			CausationID:   e.Metadata.CausationID,
			UserID:        e.Metadata.UserID,
			TenantID:      e.Metadata.TenantID,
			TraceID:       e.Metadata.TraceID,
			SpanID:        e.Metadata.SpanID,
			Custom:        customCopy,
		},
	}
}

// ============================================================================
// Verify Interface Compliance
// ============================================================================

var (
	_ eventsourcing.EventStore         = (*Store)(nil)
	_ eventsourcing.ExtendedEventStore = (*Store)(nil)
	_ eventsourcing.SnapshotStore      = (*SnapshotStore)(nil)
)
