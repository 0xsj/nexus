package memory

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/0xsj/nexus/platform/internal/credential/application/query"
	"github.com/0xsj/nexus/platform/internal/credential/domain"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing/memory"
)

// ============================================================================
// Memory Repository
// ============================================================================

// Repository is an in-memory implementation of credential repositories.
// It implements both domain.CredentialRepository (write) and query.CredentialReadRepository (read).
type Repository struct {
	store    *memory.Store
	registry *eventsourcing.EventRegistry

	// Read model cache (in production, this would be a separate read database)
	mu    sync.RWMutex
	views map[string]*query.CredentialView
}

// NewRepository creates a new in-memory credential repository.
func NewRepository() *Repository {
	return &Repository{
		store:    memory.NewStore(),
		registry: domain.NewRegistry(),
		views:    make(map[string]*query.CredentialView),
	}
}

// ============================================================================
// Write Repository Implementation (domain.CredentialRepository)
// ============================================================================

// Load loads a credential aggregate by ID.
func (r *Repository) Load(ctx context.Context, id string) (*domain.Credential, error) {
	envelopes, err := r.store.Load(ctx, id)
	if err != nil {
		return nil, err
	}

	if len(envelopes) == 0 {
		return nil, eventsourcing.ErrAggregateNotFound("Repository.Load", domain.AggregateType, id)
	}

	credential := domain.NewCredential(id)
	if err := eventsourcing.HydrateFromEnvelopes(credential, envelopes, r.registry); err != nil {
		return nil, err
	}

	return credential, nil
}

// Save saves a credential aggregate.
func (r *Repository) Save(ctx context.Context, credential *domain.Credential) error {
	changes := credential.Changes()
	if len(changes) == 0 {
		return nil
	}

	// Convert events to envelopes
	expectedVersion := credential.Version() - len(changes)
	envelopes := make([]*eventsourcing.EventEnvelope, len(changes))

	for i, evt := range changes {
		envelope, err := eventsourcing.NewEventEnvelope(evt, expectedVersion+i+1)
		if err != nil {
			return err
		}
		envelopes[i] = envelope
	}

	// Save to event store
	if err := r.store.Save(ctx, credential.AggregateID(), envelopes, expectedVersion); err != nil {
		return err
	}

	// Update read model
	r.updateReadModel(credential)

	// Clear changes
	credential.ClearChanges()

	return nil
}

// Exists checks if a credential exists.
func (r *Repository) Exists(ctx context.Context, id string) (bool, error) {
	envelopes, err := r.store.Load(ctx, id)
	if err != nil {
		return false, err
	}
	return len(envelopes) > 0, nil
}

// ============================================================================
// Read Model Updates
// ============================================================================

// updateReadModel updates the read model from the aggregate state.
func (r *Repository) updateReadModel(credential *domain.Credential) {
	r.mu.Lock()
	defer r.mu.Unlock()

	view := r.views[credential.AggregateID()]
	if view == nil {
		view = &query.CredentialView{
			ID:        credential.AggregateID(),
			CreatedAt: time.Now().UTC(),
		}
	}

	view.CredentialType = credential.CredentialType()
	view.SchemaID = credential.SchemaID()
	view.HolderDID = credential.HolderDID()
	view.IssuerDID = credential.IssuerDID()
	view.Status = credential.Status().String()
	view.Claims = credential.Claims()
	view.IssuedAt = credential.IssuedAt()
	view.ExpiresAt = credential.ExpiresAt()
	view.UpdatedAt = time.Now().UTC()
	view.Version = credential.Version()

	// Update revocation/suspension timestamps
	if info := credential.RevocationInfo(); info != nil {
		view.RevokedAt = &info.RevokedAt
	}
	if info := credential.SuspensionInfo(); info != nil {
		view.SuspendedAt = &info.SuspendedAt
	} else {
		view.SuspendedAt = nil // Clear if reinstated
	}

	r.views[credential.AggregateID()] = view
}

// ============================================================================
// Read Repository Implementation (query.CredentialReadRepository)
// ============================================================================

// GetByID retrieves a credential view by ID.
func (r *Repository) GetByID(ctx context.Context, id string) (*query.CredentialView, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	view, exists := r.views[id]
	if !exists {
		return nil, eventsourcing.ErrAggregateNotFound("Repository.GetByID", domain.AggregateType, id)
	}

	return r.copyView(view), nil
}

// List retrieves a paginated list of credentials.
func (r *Repository) List(ctx context.Context, opts query.ListOptions) (*query.CredentialListResult, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Collect and filter
	var filtered []*query.CredentialView
	for _, view := range r.views {
		if r.matchesFilters(view, opts) {
			filtered = append(filtered, view)
		}
	}

	// Sort
	r.sortViews(filtered, opts.SortBy, opts.SortOrder)

	// Paginate
	return r.paginate(filtered, opts), nil
}

// ListByHolder retrieves credentials for a specific holder.
func (r *Repository) ListByHolder(ctx context.Context, holderDID string, opts query.ListOptions) (*query.CredentialListResult, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []*query.CredentialView
	for _, view := range r.views {
		if view.HolderDID == holderDID && r.matchesFilters(view, opts) {
			filtered = append(filtered, view)
		}
	}

	r.sortViews(filtered, opts.SortBy, opts.SortOrder)
	return r.paginate(filtered, opts), nil
}

// ListByIssuer retrieves credentials issued by a specific issuer.
func (r *Repository) ListByIssuer(ctx context.Context, issuerDID string, opts query.ListOptions) (*query.CredentialListResult, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []*query.CredentialView
	for _, view := range r.views {
		if view.IssuerDID == issuerDID && r.matchesFilters(view, opts) {
			filtered = append(filtered, view)
		}
	}

	r.sortViews(filtered, opts.SortBy, opts.SortOrder)
	return r.paginate(filtered, opts), nil
}

// Count returns the total number of credentials.
func (r *Repository) Count(ctx context.Context) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.views), nil
}

// CountByStatus returns the number of credentials with a specific status.
func (r *Repository) CountByStatus(ctx context.Context, status string) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	count := 0
	for _, view := range r.views {
		if view.Status == status {
			count++
		}
	}

	return count, nil
}

// ============================================================================
// Helpers
// ============================================================================

// matchesFilters checks if a view matches the filter options.
func (r *Repository) matchesFilters(view *query.CredentialView, opts query.ListOptions) bool {
	if opts.Status != "" && view.Status != opts.Status {
		return false
	}
	if opts.CredentialType != "" && view.CredentialType != opts.CredentialType {
		return false
	}
	return true
}

// sortViews sorts the views by the specified field and order.
func (r *Repository) sortViews(views []*query.CredentialView, sortBy, sortOrder string) {
	if sortBy == "" {
		sortBy = "created_at"
	}
	if sortOrder == "" {
		sortOrder = "desc"
	}

	sort.Slice(views, func(i, j int) bool {
		var less bool

		switch sortBy {
		case "created_at":
			less = views[i].CreatedAt.Before(views[j].CreatedAt)
		case "updated_at":
			less = views[i].UpdatedAt.Before(views[j].UpdatedAt)
		case "issued_at":
			if views[i].IssuedAt == nil || views[j].IssuedAt == nil {
				less = views[i].IssuedAt != nil
			} else {
				less = views[i].IssuedAt.Before(*views[j].IssuedAt)
			}
		case "credential_type":
			less = views[i].CredentialType < views[j].CredentialType
		case "status":
			less = views[i].Status < views[j].Status
		default:
			less = views[i].CreatedAt.Before(views[j].CreatedAt)
		}

		if sortOrder == "desc" {
			return !less
		}
		return less
	})
}

// paginate applies pagination to the views.
func (r *Repository) paginate(views []*query.CredentialView, opts query.ListOptions) *query.CredentialListResult {
	total := len(views)

	// Apply offset
	offset := opts.Offset
	if offset > total {
		offset = total
	}
	views = views[offset:]

	// Apply limit
	limit := opts.Limit
	if limit <= 0 {
		limit = 20
	}

	hasMore := len(views) > limit
	if len(views) > limit {
		views = views[:limit]
	}

	// Copy views for result
	results := make([]*query.CredentialView, len(views))
	for i, v := range views {
		results[i] = r.copyView(v)
	}

	return &query.CredentialListResult{
		Credentials: results,
		Total:       total,
		Limit:       limit,
		Offset:      opts.Offset,
		HasMore:     hasMore,
	}
}

// copyView creates a deep copy of a CredentialView.
func (r *Repository) copyView(v *query.CredentialView) *query.CredentialView {
	cp := &query.CredentialView{
		ID:             v.ID,
		CredentialType: v.CredentialType,
		SchemaID:       v.SchemaID,
		HolderDID:      v.HolderDID,
		IssuerDID:      v.IssuerDID,
		Status:         v.Status,
		CreatedAt:      v.CreatedAt,
		UpdatedAt:      v.UpdatedAt,
		Version:        v.Version,
	}

	if v.Claims != nil {
		cp.Claims = make(map[string]any, len(v.Claims))
		for k, val := range v.Claims {
			cp.Claims[k] = val
		}
	}

	if v.IssuedAt != nil {
		t := *v.IssuedAt
		cp.IssuedAt = &t
	}
	if v.ExpiresAt != nil {
		t := *v.ExpiresAt
		cp.ExpiresAt = &t
	}
	if v.RevokedAt != nil {
		t := *v.RevokedAt
		cp.RevokedAt = &t
	}
	if v.SuspendedAt != nil {
		t := *v.SuspendedAt
		cp.SuspendedAt = &t
	}

	return cp
}

// ============================================================================
// Testing Helpers
// ============================================================================

// Clear removes all data from the repository.
func (r *Repository) Clear() {
	r.store.Clear()

	r.mu.Lock()
	r.views = make(map[string]*query.CredentialView)
	r.mu.Unlock()
}

// EventCount returns the total number of events in the store.
func (r *Repository) EventCount() int {
	return r.store.EventCount()
}

// ViewCount returns the total number of views in the read model.
func (r *Repository) ViewCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.views)
}

// ============================================================================
// Interface Compliance
// ============================================================================

var (
	_ domain.CredentialRepository    = (*Repository)(nil)
	_ query.CredentialReadRepository = (*Repository)(nil)
)
