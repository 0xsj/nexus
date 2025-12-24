package database

// QueryOptions configures database queries.
type QueryOptions struct {
	// Pagination
	Limit  int
	Offset int

	// Sorting
	SortBy    string
	SortOrder SortOrder

	// Filtering
	Filters []Filter

	// Include soft-deleted records
	IncludeDeleted bool
}

// SortOrder represents the sort direction.
type SortOrder string

const (
	SortAsc  SortOrder = "ASC"
	SortDesc SortOrder = "DESC"
)

// Filter represents a query filter.
type Filter struct {
	Field    string
	Operator FilterOperator
	Value    any
}

// FilterOperator represents comparison operators.
type FilterOperator string

const (
	OpEqual        FilterOperator = "eq"
	OpNotEqual     FilterOperator = "neq"
	OpGreaterThan  FilterOperator = "gt"
	OpGreaterEqual FilterOperator = "gte"
	OpLessThan     FilterOperator = "lt"
	OpLessEqual    FilterOperator = "lte"
	OpIn           FilterOperator = "in"
	OpNotIn        FilterOperator = "not_in"
	OpLike         FilterOperator = "like"
	OpIsNull       FilterOperator = "is_null"
	OpIsNotNull    FilterOperator = "is_not_null"
)

// DefaultQueryOptions returns default query options.
func DefaultQueryOptions() QueryOptions {
	return QueryOptions{
		Limit:     20,
		Offset:    0,
		SortOrder: SortDesc,
	}
}

// NewQueryOptions creates new query options with limit.
func NewQueryOptions(limit int) QueryOptions {
	opts := DefaultQueryOptions()
	opts.Limit = limit
	return opts
}

// WithLimit sets the limit.
func (o QueryOptions) WithLimit(limit int) QueryOptions {
	o.Limit = limit
	return o
}

// WithOffset sets the offset.
func (o QueryOptions) WithOffset(offset int) QueryOptions {
	o.Offset = offset
	return o
}

// WithPage sets pagination by page number (1-based).
func (o QueryOptions) WithPage(page int) QueryOptions {
	if page < 1 {
		page = 1
	}
	o.Offset = (page - 1) * o.Limit
	return o
}

// WithSort sets the sort field and order.
func (o QueryOptions) WithSort(field string, order SortOrder) QueryOptions {
	o.SortBy = field
	o.SortOrder = order
	return o
}

// WithSortAsc sets ascending sort on a field.
func (o QueryOptions) WithSortAsc(field string) QueryOptions {
	o.SortBy = field
	o.SortOrder = SortAsc
	return o
}

// WithSortDesc sets descending sort on a field.
func (o QueryOptions) WithSortDesc(field string) QueryOptions {
	o.SortBy = field
	o.SortOrder = SortDesc
	return o
}

// WithFilter adds a filter.
func (o QueryOptions) WithFilter(field string, op FilterOperator, value any) QueryOptions {
	o.Filters = append(o.Filters, Filter{
		Field:    field,
		Operator: op,
		Value:    value,
	})
	return o
}

// WithEqual adds an equality filter.
func (o QueryOptions) WithEqual(field string, value any) QueryOptions {
	return o.WithFilter(field, OpEqual, value)
}

// WithNotEqual adds a not-equal filter.
func (o QueryOptions) WithNotEqual(field string, value any) QueryOptions {
	return o.WithFilter(field, OpNotEqual, value)
}

// WithIn adds an IN filter.
func (o QueryOptions) WithIn(field string, values any) QueryOptions {
	return o.WithFilter(field, OpIn, values)
}

// WithLike adds a LIKE filter.
func (o QueryOptions) WithLike(field string, pattern string) QueryOptions {
	return o.WithFilter(field, OpLike, pattern)
}

// WithIncludeDeleted includes soft-deleted records.
func (o QueryOptions) WithIncludeDeleted() QueryOptions {
	o.IncludeDeleted = true
	return o
}

// Page returns the current page number (1-based).
func (o QueryOptions) Page() int {
	if o.Limit == 0 {
		return 1
	}
	return (o.Offset / o.Limit) + 1
}

// ============================================================================
// Paginated Result
// ============================================================================

// PaginatedResult wraps a paginated query result.
type PaginatedResult[T any] struct {
	// Items is the result set.
	Items []T

	// Total is the total count of matching records.
	Total int64

	// Limit is the page size.
	Limit int

	// Offset is the current offset.
	Offset int
}

// NewPaginatedResult creates a new paginated result.
func NewPaginatedResult[T any](items []T, total int64, opts QueryOptions) PaginatedResult[T] {
	return PaginatedResult[T]{
		Items:  items,
		Total:  total,
		Limit:  opts.Limit,
		Offset: opts.Offset,
	}
}

// Page returns the current page number (1-based).
func (r PaginatedResult[T]) Page() int {
	if r.Limit == 0 {
		return 1
	}
	return (r.Offset / r.Limit) + 1
}

// TotalPages returns the total number of pages.
func (r PaginatedResult[T]) TotalPages() int {
	if r.Limit == 0 {
		return 1
	}
	pages := int(r.Total) / r.Limit
	if int(r.Total)%r.Limit > 0 {
		pages++
	}
	return pages
}

// HasNext returns true if there is a next page.
func (r PaginatedResult[T]) HasNext() bool {
	return r.Page() < r.TotalPages()
}

// HasPrev returns true if there is a previous page.
func (r PaginatedResult[T]) HasPrev() bool {
	return r.Page() > 1
}

// IsEmpty returns true if the result set is empty.
func (r PaginatedResult[T]) IsEmpty() bool {
	return len(r.Items) == 0
}

// Count returns the number of items in this page.
func (r PaginatedResult[T]) Count() int {
	return len(r.Items)
}
