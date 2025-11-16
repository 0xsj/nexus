package domain

import (
	"context"

	"github.com/0xsj/result"
)

// UserRepository defines the interface for user persistence operations.
type UserRepository interface {
	// Create persists a new user.
	// Returns ErrEmailAlreadyExists or ErrUsernameAlreadyExists if duplicates exist.
	Create(ctx context.Context, user *User) result.Result[*User]

	// FindByID retrieves a user by their ID.
	// Returns ErrUserNotFound if the user doesn't exist.
	FindByID(ctx context.Context, id string) result.Result[*User]

	// FindByEmail retrieves a user by their email address.
	// Returns ErrUserNotFound if the user doesn't exist.
	FindByEmail(ctx context.Context, email Email) result.Result[*User]

	// FindByUsername retrieves a user by their username.
	// Returns ErrUserNotFound if the user doesn't exist.
	FindByUsername(ctx context.Context, username Username) result.Result[*User]

	// Update updates an existing user.
	// Returns ErrUserNotFound if the user doesn't exist.
	// Returns ErrEmailAlreadyExists or ErrUsernameAlreadyExists if attempting to update to existing values.
	Update(ctx context.Context, user *User) result.Result[*User]

	// Delete soft-deletes a user by marking them as inactive.
	// Returns ErrUserNotFound if the user doesn't exist.
	Delete(ctx context.Context, id string) result.Result[struct{}]

	// List retrieves a paginated list of users.
	List(ctx context.Context, params ListParams) result.Result[*ListResult]

	// ExistsByEmail checks if a user with the given email exists.
	ExistsByEmail(ctx context.Context, email Email) result.Result[bool]

	// ExistsByUsername checks if a user with the given username exists.
	ExistsByUsername(ctx context.Context, username Username) result.Result[bool]
}

// ListParams defines parameters for listing users.
type ListParams struct {
	Limit  int
	Offset int
	SortBy string // "created_at", "username", "email"
	Order  string // "asc", "desc"
}

// ListResult contains paginated user results.
type ListResult struct {
	Users   []*User
	Total   int
	Limit   int
	Offset  int
	HasMore bool
}

// NewListParams creates ListParams with defaults.
func NewListParams(limit, offset int) ListParams {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	return ListParams{
		Limit:  limit,
		Offset: offset,
		SortBy: "created_at",
		Order:  "desc",
	}
}

// WithSortBy sets the sort field.
func (p ListParams) WithSortBy(sortBy string) ListParams {
	validSortFields := map[string]bool{
		"created_at": true,
		"username":   true,
		"email":      true,
	}

	if validSortFields[sortBy] {
		p.SortBy = sortBy
	}

	return p
}

// WithOrder sets the sort order.
func (p ListParams) WithOrder(order string) ListParams {
	if order == "asc" || order == "desc" {
		p.Order = order
	}

	return p
}
