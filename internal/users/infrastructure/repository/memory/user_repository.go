// internal/users/infrastructure/repository/memory/user_repository.go
package memory

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/0xsj/nexus/internal/users/domain"
	"github.com/0xsj/result"
)

// UserRepository is an in-memory implementation of domain.UserRepository.
// Thread-safe using RWMutex for concurrent access.
type UserRepository struct {
	mu         sync.RWMutex
	users      map[string]*domain.User // id -> user
	byEmail    map[string]string       // email -> id
	byUsername map[string]string       // username -> id
}

// NewUserRepository creates a new in-memory user repository.
func NewUserRepository() *UserRepository {
	return &UserRepository{
		users:      make(map[string]*domain.User),
		byEmail:    make(map[string]string),
		byUsername: make(map[string]string),
	}
}

// Create persists a new user.
func (r *UserRepository) Create(ctx context.Context, user *domain.User) result.Result[*domain.User] {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check for duplicate email
	if _, exists := r.byEmail[user.Email]; exists {
		return result.Err[*domain.User](domain.ErrEmailAlreadyExists(user.Email))
	}

	// Check for duplicate username
	if _, exists := r.byUsername[user.Username]; exists {
		return result.Err[*domain.User](domain.ErrUsernameAlreadyExists(user.Username))
	}

	// Store user
	r.users[user.ID] = user
	r.byEmail[user.Email] = user.ID
	r.byUsername[user.Username] = user.ID

	return result.Ok(user)
}

// FindByID retrieves a user by their ID.
func (r *UserRepository) FindByID(ctx context.Context, id string) result.Result[*domain.User] {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[id]
	if !exists {
		return result.Err[*domain.User](domain.ErrUserNotFound(id))
	}

	return result.Ok(user)
}

// FindByEmail retrieves a user by their email address.
func (r *UserRepository) FindByEmail(ctx context.Context, email domain.Email) result.Result[*domain.User] {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, exists := r.byEmail[email.Value()]
	if !exists {
		return result.Err[*domain.User](domain.ErrUserNotFound(email.Value()))
	}

	user := r.users[id]
	return result.Ok(user)
}

// FindByUsername retrieves a user by their username.
func (r *UserRepository) FindByUsername(ctx context.Context, username domain.Username) result.Result[*domain.User] {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, exists := r.byUsername[username.Value()]
	if !exists {
		return result.Err[*domain.User](domain.ErrUserNotFound(username.Value()))
	}

	user := r.users[id]
	return result.Ok(user)
}

// Update updates an existing user.
func (r *UserRepository) Update(ctx context.Context, user *domain.User) result.Result[*domain.User] {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if user exists
	existing, exists := r.users[user.ID]
	if !exists {
		return result.Err[*domain.User](domain.ErrUserNotFound(user.ID))
	}

	// If email changed, check for conflicts
	if existing.Email != user.Email {
		if existingID, exists := r.byEmail[user.Email]; exists && existingID != user.ID {
			return result.Err[*domain.User](domain.ErrEmailAlreadyExists(user.Email))
		}
		// Update email index
		delete(r.byEmail, existing.Email)
		r.byEmail[user.Email] = user.ID
	}

	// If username changed, check for conflicts
	if existing.Username != user.Username {
		if existingID, exists := r.byUsername[user.Username]; exists && existingID != user.ID {
			return result.Err[*domain.User](domain.ErrUsernameAlreadyExists(user.Username))
		}
		// Update username index
		delete(r.byUsername, existing.Username)
		r.byUsername[user.Username] = user.ID
	}

	// Update timestamp
	user.UpdatedAt = time.Now().UTC()

	// Store updated user
	r.users[user.ID] = user

	return result.Ok(user)
}

// Delete soft-deletes a user by marking them as inactive.
func (r *UserRepository) Delete(ctx context.Context, id string) result.Result[struct{}] {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, exists := r.users[id]
	if !exists {
		return result.Err[struct{}](domain.ErrUserNotFound(id))
	}

	// Soft delete by marking inactive
	user.Deactivate()

	return result.Ok(struct{}{})
}

// List retrieves a paginated list of users.
func (r *UserRepository) List(ctx context.Context, params domain.ListParams) result.Result[*domain.ListResult] {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Collect all users
	allUsers := make([]*domain.User, 0, len(r.users))
	for _, user := range r.users {
		allUsers = append(allUsers, user)
	}

	// Sort by the requested field
	r.sortUsers(allUsers, params.SortBy, params.Order)

	// Calculate pagination
	total := len(allUsers)
	start := params.Offset
	end := params.Offset + params.Limit

	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	paginatedUsers := allUsers[start:end]
	hasMore := end < total

	return result.Ok(&domain.ListResult{
		Users:   paginatedUsers,
		Total:   total,
		Limit:   params.Limit,
		Offset:  params.Offset,
		HasMore: hasMore,
	})
}

// ExistsByEmail checks if a user with the given email exists.
func (r *UserRepository) ExistsByEmail(ctx context.Context, email domain.Email) result.Result[bool] {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.byEmail[email.Value()]
	return result.Ok(exists)
}

// ExistsByUsername checks if a user with the given username exists.
func (r *UserRepository) ExistsByUsername(ctx context.Context, username domain.Username) result.Result[bool] {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.byUsername[username.Value()]
	return result.Ok(exists)
}

// sortUsers sorts users by the specified field and order.
func (r *UserRepository) sortUsers(users []*domain.User, sortBy, order string) {
	// Simple sorting implementation
	// In production, you'd use a more sophisticated sorting algorithm
	if sortBy == "created_at" {
		if order == "asc" {
			// Sort ascending by created_at
			for i := 0; i < len(users); i++ {
				for j := i + 1; j < len(users); j++ {
					if users[i].CreatedAt.After(users[j].CreatedAt) {
						users[i], users[j] = users[j], users[i]
					}
				}
			}
		} else {
			// Sort descending by created_at (default)
			for i := 0; i < len(users); i++ {
				for j := i + 1; j < len(users); j++ {
					if users[i].CreatedAt.Before(users[j].CreatedAt) {
						users[i], users[j] = users[j], users[i]
					}
				}
			}
		}
	} else if sortBy == "username" {
		if order == "asc" {
			for i := 0; i < len(users); i++ {
				for j := i + 1; j < len(users); j++ {
					if strings.ToLower(users[i].Username) > strings.ToLower(users[j].Username) {
						users[i], users[j] = users[j], users[i]
					}
				}
			}
		} else {
			for i := 0; i < len(users); i++ {
				for j := i + 1; j < len(users); j++ {
					if strings.ToLower(users[i].Username) < strings.ToLower(users[j].Username) {
						users[i], users[j] = users[j], users[i]
					}
				}
			}
		}
	}
}
