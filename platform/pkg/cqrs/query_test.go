package cqrs_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/0xsj/nexus/platform/pkg/cqrs"
)

// ============================================================================
// Test Queries
// ============================================================================

type GetUserQuery struct {
	cqrs.BaseQuery
	UserID string
}

func NewGetUserQuery(userID string) *GetUserQuery {
	return &GetUserQuery{
		BaseQuery: cqrs.NewBaseQuery("GetUser"),
		UserID:    userID,
	}
}

func (q *GetUserQuery) QueryName() string {
	return "GetUser"
}

type ListUsersQuery struct {
	cqrs.BaseQuery
	Limit  int
	Offset int
}

func NewListUsersQuery(limit, offset int) *ListUsersQuery {
	return &ListUsersQuery{
		BaseQuery: cqrs.NewBaseQuery("ListUsers"),
		Limit:     limit,
		Offset:    offset,
	}
}

func (q *ListUsersQuery) QueryName() string {
	return "ListUsers"
}

type CacheableUserQuery struct {
	cqrs.BaseQuery
	UserID string
	ttl    time.Duration
}

func NewCacheableUserQuery(userID string, ttl time.Duration) *CacheableUserQuery {
	return &CacheableUserQuery{
		BaseQuery: cqrs.NewBaseQuery("CacheableUser"),
		UserID:    userID,
		ttl:       ttl,
	}
}

func (q *CacheableUserQuery) QueryName() string {
	return "CacheableUser"
}

func (q *CacheableUserQuery) CacheKey() string {
	return "user:" + q.UserID
}

func (q *CacheableUserQuery) CacheTTL() time.Duration {
	return q.ttl
}

// ============================================================================
// Test Results
// ============================================================================

type UserResult struct {
	ID    string
	Email string
	Name  string
}

type UsersListResult struct {
	Users []UserResult
	Total int
}

// ============================================================================
// Test Handlers
// ============================================================================

type GetUserHandler struct {
	called    bool
	lastQuery *GetUserQuery
	result    *UserResult
	returnErr error
}

func (h *GetUserHandler) Handle(ctx context.Context, query *GetUserQuery) (*UserResult, error) {
	h.called = true
	h.lastQuery = query

	if h.returnErr != nil {
		return nil, h.returnErr
	}

	return h.result, nil
}

type ListUsersHandler struct {
	called    bool
	lastQuery *ListUsersQuery
	result    *UsersListResult
}

func (h *ListUsersHandler) Handle(ctx context.Context, query *ListUsersQuery) (*UsersListResult, error) {
	h.called = true
	h.lastQuery = query

	return h.result, nil
}

type CacheableUserHandler struct {
	callCount int
	result    *UserResult
}

func (h *CacheableUserHandler) Handle(ctx context.Context, query *CacheableUserQuery) (*UserResult, error) {
	h.callCount++
	return h.result, nil
}

type PanicQueryHandler struct{}

func (h *PanicQueryHandler) Handle(ctx context.Context, query *GetUserQuery) (*UserResult, error) {
	panic("query handler panic!")
}

type SlowQueryHandler struct {
	delay time.Duration
}

func (h *SlowQueryHandler) Handle(ctx context.Context, query *GetUserQuery) (*UserResult, error) {
	select {
	case <-time.After(h.delay):
		return &UserResult{ID: "slow"}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// ============================================================================
// Query Bus Tests
// ============================================================================

func TestNewQueryBus(t *testing.T) {
	bus := cqrs.NewQueryBus()

	if bus == nil {
		t.Fatal("expected non-nil query bus")
	}
}

func TestQueryBus_Register(t *testing.T) {
	bus := cqrs.NewQueryBus()
	handler := &GetUserHandler{}

	err := bus.Register("GetUser", handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestQueryBus_Register_EmptyType(t *testing.T) {
	bus := cqrs.NewQueryBus()
	handler := &GetUserHandler{}

	err := bus.Register("", handler)
	if err == nil {
		t.Fatal("expected error for empty query type")
	}
}

func TestQueryBus_Register_NilHandler(t *testing.T) {
	bus := cqrs.NewQueryBus()

	err := bus.Register("GetUser", nil)
	if err == nil {
		t.Fatal("expected error for nil handler")
	}
}

func TestQueryBus_Dispatch(t *testing.T) {
	bus := cqrs.NewQueryBus()
	handler := &GetUserHandler{
		result: &UserResult{
			ID:    "user-123",
			Email: "test@example.com",
			Name:  "Test User",
		},
	}

	err := bus.Register("GetUser", handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	query := NewGetUserQuery("user-123")
	result, err := bus.Dispatch(context.Background(), query)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !handler.called {
		t.Error("expected handler to be called")
	}

	if handler.lastQuery.UserID != "user-123" {
		t.Errorf("expected UserID 'user-123', got '%s'", handler.lastQuery.UserID)
	}

	userResult, ok := result.(*UserResult)
	if !ok {
		t.Fatalf("expected *UserResult, got %T", result)
	}

	if userResult.ID != "user-123" {
		t.Errorf("expected result ID 'user-123', got '%s'", userResult.ID)
	}

	if userResult.Email != "test@example.com" {
		t.Errorf("expected result Email 'test@example.com', got '%s'", userResult.Email)
	}
}

func TestQueryBus_Dispatch_NilQuery(t *testing.T) {
	bus := cqrs.NewQueryBus()

	_, err := bus.Dispatch(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil query")
	}
}

func TestQueryBus_Dispatch_HandlerNotFound(t *testing.T) {
	bus := cqrs.NewQueryBus()

	query := NewGetUserQuery("user-123")
	_, err := bus.Dispatch(context.Background(), query)

	if err == nil {
		t.Fatal("expected error for missing handler")
	}

	if !cqrs.IsQueryNotFound(err) {
		t.Errorf("expected query not found error, got: %v", err)
	}
}

func TestQueryBus_Dispatch_HandlerReturnsError(t *testing.T) {
	bus := cqrs.NewQueryBus()
	expectedErr := errors.New("handler error")
	handler := &GetUserHandler{returnErr: expectedErr}

	err := bus.Register("GetUser", handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	query := NewGetUserQuery("user-123")
	_, err = bus.Dispatch(context.Background(), query)

	if err == nil {
		t.Fatal("expected error from handler")
	}

	if err.Error() != expectedErr.Error() {
		t.Errorf("expected error '%v', got '%v'", expectedErr, err)
	}
}

func TestQueryBus_Dispatch_HandlerPanics(t *testing.T) {
	bus := cqrs.NewQueryBus()
	handler := &PanicQueryHandler{}

	err := bus.Register("GetUser", handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	query := NewGetUserQuery("user-123")
	_, err = bus.Dispatch(context.Background(), query)

	if err == nil {
		t.Fatal("expected error from panic")
	}
}

func TestQueryBus_Dispatch_MultipleQueries(t *testing.T) {
	bus := cqrs.NewQueryBus()

	getUserHandler := &GetUserHandler{
		result: &UserResult{ID: "user-123", Email: "test@example.com"},
	}
	listUsersHandler := &ListUsersHandler{
		result: &UsersListResult{
			Users: []UserResult{{ID: "user-1"}, {ID: "user-2"}},
			Total: 2,
		},
	}

	bus.Register("GetUser", getUserHandler)
	bus.Register("ListUsers", listUsersHandler)

	// Dispatch GetUser
	getUserQuery := NewGetUserQuery("user-123")
	result1, err := bus.Dispatch(context.Background(), getUserQuery)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	userResult, ok := result1.(*UserResult)
	if !ok {
		t.Fatalf("expected *UserResult, got %T", result1)
	}
	if userResult.ID != "user-123" {
		t.Errorf("expected ID 'user-123', got '%s'", userResult.ID)
	}

	// Dispatch ListUsers
	listUsersQuery := NewListUsersQuery(10, 0)
	result2, err := bus.Dispatch(context.Background(), listUsersQuery)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	listResult, ok := result2.(*UsersListResult)
	if !ok {
		t.Fatalf("expected *UsersListResult, got %T", result2)
	}
	if listResult.Total != 2 {
		t.Errorf("expected Total 2, got %d", listResult.Total)
	}
}

// ============================================================================
// RegisterQuery Helper Tests
// ============================================================================

func TestRegisterQuery(t *testing.T) {
	bus := cqrs.NewQueryBus()
	handler := &GetUserHandler{
		result: &UserResult{ID: "user-456", Email: "test@example.com"},
	}

	err := cqrs.RegisterQuery[*GetUserQuery, *UserResult](bus, handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	query := NewGetUserQuery("user-456")
	result, err := bus.Dispatch(context.Background(), query)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	userResult, ok := result.(*UserResult)
	if !ok {
		t.Fatalf("expected *UserResult, got %T", result)
	}

	if userResult.ID != "user-456" {
		t.Errorf("expected ID 'user-456', got '%s'", userResult.ID)
	}
}

// ============================================================================
// DispatchQuery Helper Tests
// ============================================================================

func TestDispatchQuery(t *testing.T) {
	bus := cqrs.NewQueryBus()
	handler := &GetUserHandler{
		result: &UserResult{ID: "user-789", Email: "typed@example.com"},
	}

	bus.Register("GetUser", handler)

	query := NewGetUserQuery("user-789")
	result, err := cqrs.DispatchQuery[*UserResult](context.Background(), bus, query)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != "user-789" {
		t.Errorf("expected ID 'user-789', got '%s'", result.ID)
	}

	if result.Email != "typed@example.com" {
		t.Errorf("expected Email 'typed@example.com', got '%s'", result.Email)
	}
}

func TestDispatchQuery_NilResult(t *testing.T) {
	bus := cqrs.NewQueryBus()
	handler := &GetUserHandler{result: nil}

	bus.Register("GetUser", handler)

	query := NewGetUserQuery("user-789")
	result, err := cqrs.DispatchQuery[*UserResult](context.Background(), bus, query)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

// ============================================================================
// Middleware Tests
// ============================================================================

func TestMiddlewareQueryBus(t *testing.T) {
	baseBus := cqrs.NewQueryBus()
	handler := &GetUserHandler{
		result: &UserResult{ID: "user-123"},
	}
	baseBus.Register("GetUser", handler)

	var middlewareCalls []string

	middleware1 := func(next cqrs.QueryHandlerFunc[cqrs.Query, any]) cqrs.QueryHandlerFunc[cqrs.Query, any] {
		return func(ctx context.Context, query cqrs.Query) (any, error) {
			middlewareCalls = append(middlewareCalls, "middleware1-before")
			result, err := next(ctx, query)
			middlewareCalls = append(middlewareCalls, "middleware1-after")
			return result, err
		}
	}

	middleware2 := func(next cqrs.QueryHandlerFunc[cqrs.Query, any]) cqrs.QueryHandlerFunc[cqrs.Query, any] {
		return func(ctx context.Context, query cqrs.Query) (any, error) {
			middlewareCalls = append(middlewareCalls, "middleware2-before")
			result, err := next(ctx, query)
			middlewareCalls = append(middlewareCalls, "middleware2-after")
			return result, err
		}
	}

	bus := cqrs.NewMiddlewareQueryBus(baseBus, middleware1, middleware2)

	query := NewGetUserQuery("user-123")
	_, err := bus.Dispatch(context.Background(), query)

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

func TestQueryLoggingMiddleware(t *testing.T) {
	baseBus := cqrs.NewQueryBus()
	handler := &GetUserHandler{
		result: &UserResult{ID: "user-123"},
	}
	baseBus.Register("GetUser", handler)

	var loggedQuery string
	var loggedDuration time.Duration
	var loggedErr error

	logFn := func(ctx context.Context, queryName string, duration time.Duration, err error) {
		loggedQuery = queryName
		loggedDuration = duration
		loggedErr = err
	}

	bus := cqrs.NewMiddlewareQueryBus(baseBus, cqrs.QueryLoggingMiddleware(logFn))

	query := NewGetUserQuery("user-123")
	_, err := bus.Dispatch(context.Background(), query)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if loggedQuery != "GetUser" {
		t.Errorf("expected logged query 'GetUser', got '%s'", loggedQuery)
	}

	if loggedDuration <= 0 {
		t.Error("expected positive duration")
	}

	if loggedErr != nil {
		t.Errorf("expected no logged error, got: %v", loggedErr)
	}
}

func TestQueryRecoveryMiddleware(t *testing.T) {
	baseBus := cqrs.NewQueryBus()
	handler := &PanicQueryHandler{}
	baseBus.Register("GetUser", handler)

	bus := cqrs.NewMiddlewareQueryBus(baseBus, cqrs.QueryRecoveryMiddleware())

	query := NewGetUserQuery("user-123")
	_, err := bus.Dispatch(context.Background(), query)

	if err == nil {
		t.Fatal("expected error from recovered panic")
	}
}

func TestQueryTimeoutMiddleware(t *testing.T) {
	baseBus := cqrs.NewQueryBus()
	handler := &SlowQueryHandler{delay: 100 * time.Millisecond}
	baseBus.Register("GetUser", handler)

	bus := cqrs.NewMiddlewareQueryBus(baseBus, cqrs.QueryTimeoutMiddleware(10*time.Millisecond))

	query := NewGetUserQuery("user-123")
	_, err := bus.Dispatch(context.Background(), query)

	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestQueryTimeoutMiddleware_Completes(t *testing.T) {
	baseBus := cqrs.NewQueryBus()
	handler := &SlowQueryHandler{delay: 10 * time.Millisecond}
	baseBus.Register("GetUser", handler)

	bus := cqrs.NewMiddlewareQueryBus(baseBus, cqrs.QueryTimeoutMiddleware(100*time.Millisecond))

	query := NewGetUserQuery("user-123")
	result, err := bus.Dispatch(context.Background(), query)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	userResult, ok := result.(*UserResult)
	if !ok {
		t.Fatalf("expected *UserResult, got %T", result)
	}

	if userResult.ID != "slow" {
		t.Errorf("expected ID 'slow', got '%s'", userResult.ID)
	}
}

// ============================================================================
// Caching Middleware Tests
// ============================================================================

type MockQueryCache struct {
	data     map[string]any
	getCount int
	setCount int
}

func NewMockQueryCache() *MockQueryCache {
	return &MockQueryCache{
		data: make(map[string]any),
	}
}

func (c *MockQueryCache) Get(ctx context.Context, key string) (any, bool) {
	c.getCount++
	val, exists := c.data[key]
	return val, exists
}

func (c *MockQueryCache) Set(ctx context.Context, key string, value any, ttl time.Duration) {
	c.setCount++
	c.data[key] = value
}

func (c *MockQueryCache) Delete(ctx context.Context, key string) {
	delete(c.data, key)
}

func TestQueryCachingMiddleware(t *testing.T) {
	baseBus := cqrs.NewQueryBus()
	handler := &CacheableUserHandler{
		result: &UserResult{ID: "cached-user", Email: "cached@example.com"},
	}
	baseBus.Register("CacheableUser", handler)

	cache := NewMockQueryCache()
	bus := cqrs.NewMiddlewareQueryBus(baseBus, cqrs.QueryCachingMiddleware(cache))

	query := NewCacheableUserQuery("user-123", time.Minute)

	// First call - should hit handler
	result1, err := bus.Dispatch(context.Background(), query)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if handler.callCount != 1 {
		t.Errorf("expected handler call count 1, got %d", handler.callCount)
	}

	userResult1, ok := result1.(*UserResult)
	if !ok {
		t.Fatalf("expected *UserResult, got %T", result1)
	}
	if userResult1.ID != "cached-user" {
		t.Errorf("expected ID 'cached-user', got '%s'", userResult1.ID)
	}

	// Second call - should hit cache
	result2, err := bus.Dispatch(context.Background(), query)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if handler.callCount != 1 {
		t.Errorf("expected handler call count still 1, got %d", handler.callCount)
	}

	userResult2, ok := result2.(*UserResult)
	if !ok {
		t.Fatalf("expected *UserResult, got %T", result2)
	}
	if userResult2.ID != "cached-user" {
		t.Errorf("expected ID 'cached-user', got '%s'", userResult2.ID)
	}

	if cache.getCount != 2 {
		t.Errorf("expected cache get count 2, got %d", cache.getCount)
	}

	if cache.setCount != 1 {
		t.Errorf("expected cache set count 1, got %d", cache.setCount)
	}
}

func TestQueryCachingMiddleware_NonCacheableQuery(t *testing.T) {
	baseBus := cqrs.NewQueryBus()
	handler := &GetUserHandler{
		result: &UserResult{ID: "user-123"},
	}
	baseBus.Register("GetUser", handler)

	cache := NewMockQueryCache()
	bus := cqrs.NewMiddlewareQueryBus(baseBus, cqrs.QueryCachingMiddleware(cache))

	// GetUserQuery does NOT implement CacheableQuery
	query := NewGetUserQuery("user-123")

	_, err := bus.Dispatch(context.Background(), query)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Cache should not be used
	if cache.getCount != 0 {
		t.Errorf("expected cache get count 0, got %d", cache.getCount)
	}

	if cache.setCount != 0 {
		t.Errorf("expected cache set count 0, got %d", cache.setCount)
	}
}

// ============================================================================
// QueryHandlerFunc Tests
// ============================================================================

func TestQueryHandlerFunc(t *testing.T) {
	handlerFn := cqrs.QueryHandlerFunc[*GetUserQuery, *UserResult](func(ctx context.Context, query *GetUserQuery) (*UserResult, error) {
		return &UserResult{
			ID:    query.UserID,
			Email: "func@example.com",
		}, nil
	})

	query := NewGetUserQuery("user-func")
	result, err := handlerFn.Handle(context.Background(), query)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != "user-func" {
		t.Errorf("expected ID 'user-func', got '%s'", result.ID)
	}
}

// ============================================================================
// BaseQuery Tests
// ============================================================================

func TestBaseQuery(t *testing.T) {
	base := cqrs.NewBaseQuery("TestQuery")

	if base.QueryName() != "TestQuery" {
		t.Errorf("expected query name 'TestQuery', got '%s'", base.QueryName())
	}
}

// ============================================================================
// Error Checker Tests
// ============================================================================

func TestIsQueryNotFound(t *testing.T) {
	err := cqrs.ErrQueryNotFound("test", "UnknownQuery")

	if !cqrs.IsQueryNotFound(err) {
		t.Error("expected IsQueryNotFound to return true")
	}

	if cqrs.IsQueryNotFound(nil) {
		t.Error("expected IsQueryNotFound(nil) to return false")
	}

	if cqrs.IsQueryNotFound(errors.New("other error")) {
		t.Error("expected IsQueryNotFound to return false for other errors")
	}
}
