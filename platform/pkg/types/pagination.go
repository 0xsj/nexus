package types

import "fmt"

// Default pagination values.
const (
	DefaultPageLimit = 20
	MaxPageLimit     = 100
	DefaultOffset    = 0
)

// Page represents pagination parameters.
type Page struct {
	Limit  int
	Offset int
}

// DefaultPage returns a Page with default values.
func DefaultPage() Page {
	return Page{
		Limit:  DefaultPageLimit,
		Offset: DefaultOffset,
	}
}

// NewPage creates a Page with the given limit and offset.
// Applies validation and defaults.
func NewPage(limit, offset int) Page {
	p := Page{
		Limit:  limit,
		Offset: offset,
	}
	return p.Normalize()
}

// Normalize ensures the page has valid values.
func (p Page) Normalize() Page {
	if p.Limit <= 0 {
		p.Limit = DefaultPageLimit
	}
	if p.Limit > MaxPageLimit {
		p.Limit = MaxPageLimit
	}
	if p.Offset < 0 {
		p.Offset = DefaultOffset
	}
	return p
}

// Validate checks if the page parameters are valid.
func (p Page) Validate() error {
	if p.Limit <= 0 {
		return fmt.Errorf("limit must be positive")
	}
	if p.Limit > MaxPageLimit {
		return fmt.Errorf("limit cannot exceed %d", MaxPageLimit)
	}
	if p.Offset < 0 {
		return fmt.Errorf("offset cannot be negative")
	}
	return nil
}

// WithLimit returns a new Page with the given limit.
func (p Page) WithLimit(limit int) Page {
	p.Limit = limit
	return p
}

// WithOffset returns a new Page with the given offset.
func (p Page) WithOffset(offset int) Page {
	p.Offset = offset
	return p
}

// Next returns the next page.
func (p Page) Next() Page {
	return Page{
		Limit:  p.Limit,
		Offset: p.Offset + p.Limit,
	}
}

// Prev returns the previous page.
// Returns the same page if already at the beginning.
func (p Page) Prev() Page {
	newOffset := p.Offset - p.Limit
	if newOffset < 0 {
		newOffset = 0
	}
	return Page{
		Limit:  p.Limit,
		Offset: newOffset,
	}
}

// PageNumber returns the current page number (1-indexed).
func (p Page) PageNumber() int {
	if p.Limit <= 0 {
		return 1
	}
	return (p.Offset / p.Limit) + 1
}

// ============================================================================
// PageResult
// ============================================================================

// PageResult contains paginated results metadata.
type PageResult[T any] struct {
	Items      []T  `json:"items"`
	TotalCount int  `json:"total_count"`
	Limit      int  `json:"limit"`
	Offset     int  `json:"offset"`
	HasMore    bool `json:"has_more"`
}

// NewPageResult creates a PageResult from items and pagination info.
func NewPageResult[T any](items []T, totalCount int, page Page) PageResult[T] {
	hasMore := page.Offset+len(items) < totalCount
	return PageResult[T]{
		Items:      items,
		TotalCount: totalCount,
		Limit:      page.Limit,
		Offset:     page.Offset,
		HasMore:    hasMore,
	}
}

// IsEmpty returns true if there are no items.
func (r PageResult[T]) IsEmpty() bool {
	return len(r.Items) == 0
}

// Count returns the number of items in the current page.
func (r PageResult[T]) Count() int {
	return len(r.Items)
}

// TotalPages returns the total number of pages.
func (r PageResult[T]) TotalPages() int {
	if r.Limit <= 0 {
		return 0
	}
	pages := r.TotalCount / r.Limit
	if r.TotalCount%r.Limit > 0 {
		pages++
	}
	return pages
}

// CurrentPage returns the current page number (1-indexed).
func (r PageResult[T]) CurrentPage() int {
	if r.Limit <= 0 {
		return 1
	}
	return (r.Offset / r.Limit) + 1
}
