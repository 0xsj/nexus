package session

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManager_Success(t *testing.T) {
	config := DefaultConfig()
	store := NewMemoryStore()

	result := NewManager(config, store)

	require.True(t, result.IsOk())
	manager := result.Unwrap()
	assert.NotNil(t, manager)
}

func TestNewManager_InvalidConfig(t *testing.T) {
	t.Run("invalid TTL", func(t *testing.T) {
		config := DefaultConfig().WithSessionTTL(0)
		store := NewMemoryStore()

		result := NewManager(config, store)

		assert.True(t, result.IsErr())
	})

	t.Run("nil store", func(t *testing.T) {
		config := DefaultConfig()

		result := NewManager(config, nil)

		assert.True(t, result.IsErr())
		assert.Equal(t, ErrNilStore, result.UnwrapErr())
	})
}

func TestManager_Create_Success(t *testing.T) {
	config := DefaultConfig()
	store := NewMemoryStore()
	managerResult := NewManager(config, store)
	require.True(t, managerResult.IsOk())
	manager := managerResult.Unwrap()

	ctx := context.Background()
	userID := "user-123"
	refreshToken := "refresh-token-abc"
	ipAddress := "192.168.1.1"
	userAgent := "Mozilla/5.0"

	result := manager.Create(ctx, userID, refreshToken, ipAddress, userAgent)

	require.True(t, result.IsOk())
	session := result.Unwrap()

	assert.NotEmpty(t, session.ID)
	assert.Equal(t, userID, session.UserID)
	assert.Equal(t, refreshToken, session.RefreshToken)
	assert.Equal(t, ipAddress, session.IPAddress)
	assert.Equal(t, userAgent, session.UserAgent)
	assert.False(t, session.IsExpired())
}

func TestManager_Create_InvalidInputs(t *testing.T) {
	config := DefaultConfig()
	store := NewMemoryStore()
	managerResult := NewManager(config, store)
	require.True(t, managerResult.IsOk())
	manager := managerResult.Unwrap()

	ctx := context.Background()

	t.Run("empty user ID", func(t *testing.T) {
		result := manager.Create(ctx, "", "token", "ip", "ua")

		assert.True(t, result.IsErr())
		assert.Equal(t, ErrInvalidUserID, result.UnwrapErr())
	})

	t.Run("empty refresh token", func(t *testing.T) {
		result := manager.Create(ctx, "user-123", "", "ip", "ua")

		assert.True(t, result.IsErr())
		assert.Equal(t, ErrInvalidToken, result.UnwrapErr())
	})
}

func TestManager_Create_MaxSessionsLimit(t *testing.T) {
	config := DefaultConfig().WithMaxSessionsPerUser(2)
	store := NewMemoryStore()
	managerResult := NewManager(config, store)
	require.True(t, managerResult.IsOk())
	manager := managerResult.Unwrap()

	ctx := context.Background()
	userID := "user-123"

	// Create 2 sessions (at limit)
	for i := 0; i < 2; i++ {
		result := manager.Create(ctx, userID, "token-"+string(rune(i)), "ip", "ua")
		require.True(t, result.IsOk())
	}

	// Verify we have 2 sessions
	countResult := store.CountByUser(ctx, userID)
	require.True(t, countResult.IsOk())
	assert.Equal(t, 2, countResult.Unwrap())

	// Create 3rd session - should delete oldest and create new one
	result := manager.Create(ctx, userID, "token-3", "ip", "ua")
	require.True(t, result.IsOk())

	// Should still have 2 sessions
	countResult = store.CountByUser(ctx, userID)
	require.True(t, countResult.IsOk())
	assert.Equal(t, 2, countResult.Unwrap())
}

func TestManager_Get_Success(t *testing.T) {
	config := DefaultConfig()
	store := NewMemoryStore()
	managerResult := NewManager(config, store)
	require.True(t, managerResult.IsOk())
	manager := managerResult.Unwrap()

	ctx := context.Background()

	// Create session
	createResult := manager.Create(ctx, "user-123", "token", "ip", "ua")
	require.True(t, createResult.IsOk())
	created := createResult.Unwrap()

	// Get session
	result := manager.Get(ctx, created.ID)

	require.True(t, result.IsOk())
	retrieved := result.Unwrap()
	assert.Equal(t, created.ID, retrieved.ID)
	assert.Equal(t, created.UserID, retrieved.UserID)
}

func TestManager_Get_NotFound(t *testing.T) {
	config := DefaultConfig()
	store := NewMemoryStore()
	managerResult := NewManager(config, store)
	require.True(t, managerResult.IsOk())
	manager := managerResult.Unwrap()

	ctx := context.Background()

	result := manager.Get(ctx, "non-existent-id")

	assert.True(t, result.IsErr())
	assert.Equal(t, ErrSessionNotFound, result.UnwrapErr())
}

func TestManager_GetByRefreshToken_Success(t *testing.T) {
	config := DefaultConfig()
	store := NewMemoryStore()
	managerResult := NewManager(config, store)
	require.True(t, managerResult.IsOk())
	manager := managerResult.Unwrap()

	ctx := context.Background()
	refreshToken := "unique-refresh-token"

	// Create session
	createResult := manager.Create(ctx, "user-123", refreshToken, "ip", "ua")
	require.True(t, createResult.IsOk())
	created := createResult.Unwrap()

	// Get by refresh token
	result := manager.GetByRefreshToken(ctx, refreshToken)

	require.True(t, result.IsOk())
	retrieved := result.Unwrap()
	assert.Equal(t, created.ID, retrieved.ID)
	assert.Equal(t, refreshToken, retrieved.RefreshToken)
}

func TestManager_Refresh_Success(t *testing.T) {
	config := DefaultConfig()
	store := NewMemoryStore()
	managerResult := NewManager(config, store)
	require.True(t, managerResult.IsOk())
	manager := managerResult.Unwrap()

	ctx := context.Background()

	// Create session
	createResult := manager.Create(ctx, "user-123", "old-token", "ip", "ua")
	require.True(t, createResult.IsOk())
	created := createResult.Unwrap()

	oldToken := created.RefreshToken
	oldExpiresAt := created.ExpiresAt

	// Refresh session
	time.Sleep(10 * time.Millisecond)
	newToken := "new-token"
	result := manager.Refresh(ctx, created.ID, newToken)

	require.True(t, result.IsOk())
	refreshed := result.Unwrap()

	assert.NotEqual(t, oldToken, refreshed.RefreshToken)
	assert.Equal(t, newToken, refreshed.RefreshToken)
	assert.True(t, refreshed.ExpiresAt.After(oldExpiresAt))
}

func TestManager_Refresh_ExpiredSession(t *testing.T) {
	config := DefaultConfig().WithSessionTTL(10 * time.Millisecond)
	store := NewMemoryStore()
	managerResult := NewManager(config, store)
	require.True(t, managerResult.IsOk())
	manager := managerResult.Unwrap()

	ctx := context.Background()

	// Create session with short TTL
	createResult := manager.Create(ctx, "user-123", "token", "ip", "ua")
	require.True(t, createResult.IsOk())
	created := createResult.Unwrap()

	// Wait for expiration
	time.Sleep(20 * time.Millisecond)

	// Try to refresh
	result := manager.Refresh(ctx, created.ID, "new-token")

	assert.True(t, result.IsErr())
	assert.Equal(t, ErrSessionExpired, result.UnwrapErr())
}

func TestManager_Delete_Success(t *testing.T) {
	config := DefaultConfig()
	store := NewMemoryStore()
	managerResult := NewManager(config, store)
	require.True(t, managerResult.IsOk())
	manager := managerResult.Unwrap()

	ctx := context.Background()

	// Create session
	createResult := manager.Create(ctx, "user-123", "token", "ip", "ua")
	require.True(t, createResult.IsOk())
	created := createResult.Unwrap()

	// Delete session
	deleteResult := manager.Delete(ctx, created.ID)
	require.True(t, deleteResult.IsOk())

	// Verify deleted
	getResult := manager.Get(ctx, created.ID)
	assert.True(t, getResult.IsErr())
	assert.Equal(t, ErrSessionNotFound, getResult.UnwrapErr())
}

func TestManager_DeleteByRefreshToken_Success(t *testing.T) {
	config := DefaultConfig()
	store := NewMemoryStore()
	managerResult := NewManager(config, store)
	require.True(t, managerResult.IsOk())
	manager := managerResult.Unwrap()

	ctx := context.Background()
	refreshToken := "token-to-delete"

	// Create session
	createResult := manager.Create(ctx, "user-123", refreshToken, "ip", "ua")
	require.True(t, createResult.IsOk())

	// Delete by token
	deleteResult := manager.DeleteByRefreshToken(ctx, refreshToken)
	require.True(t, deleteResult.IsOk())

	// Verify deleted
	getResult := manager.GetByRefreshToken(ctx, refreshToken)
	assert.True(t, getResult.IsErr())
}

func TestManager_DeleteAllForUser_Success(t *testing.T) {
	config := DefaultConfig()
	store := NewMemoryStore()
	managerResult := NewManager(config, store)
	require.True(t, managerResult.IsOk())
	manager := managerResult.Unwrap()

	ctx := context.Background()
	userID := "user-123"

	// Create multiple sessions
	for i := 0; i < 3; i++ {
		result := manager.Create(ctx, userID, "token-"+string(rune(i)), "ip", "ua")
		require.True(t, result.IsOk())
	}

	// Delete all
	deleteResult := manager.DeleteAllForUser(ctx, userID)
	require.True(t, deleteResult.IsOk())
	count := deleteResult.Unwrap()
	assert.Equal(t, 3, count)

	// Verify all deleted
	listResult := manager.ListUserSessions(ctx, userID)
	require.True(t, listResult.IsOk())
	assert.Empty(t, listResult.Unwrap())
}

func TestManager_ListUserSessions_Success(t *testing.T) {
	config := DefaultConfig()
	store := NewMemoryStore()
	managerResult := NewManager(config, store)
	require.True(t, managerResult.IsOk())
	manager := managerResult.Unwrap()

	ctx := context.Background()
	userID := "user-123"

	// Create multiple sessions
	for i := 0; i < 3; i++ {
		result := manager.Create(ctx, userID, "token-"+string(rune(i)), "ip", "ua")
		require.True(t, result.IsOk())
	}

	// List sessions
	listResult := manager.ListUserSessions(ctx, userID)

	require.True(t, listResult.IsOk())
	sessions := listResult.Unwrap()
	assert.Len(t, sessions, 3)
}

func TestManager_Validate_Success(t *testing.T) {
	config := DefaultConfig()
	store := NewMemoryStore()
	managerResult := NewManager(config, store)
	require.True(t, managerResult.IsOk())
	manager := managerResult.Unwrap()

	ctx := context.Background()

	// Create session
	createResult := manager.Create(ctx, "user-123", "token", "ip", "ua")
	require.True(t, createResult.IsOk())
	created := createResult.Unwrap()

	// Validate
	result := manager.Validate(ctx, created.ID)

	require.True(t, result.IsOk())
	validated := result.Unwrap()
	assert.Equal(t, created.ID, validated.ID)
}

func TestManager_Validate_ExpiredSession(t *testing.T) {
	config := DefaultConfig().WithSessionTTL(10 * time.Millisecond)
	store := NewMemoryStore()
	managerResult := NewManager(config, store)
	require.True(t, managerResult.IsOk())
	manager := managerResult.Unwrap()

	ctx := context.Background()

	// Create session with short TTL
	createResult := manager.Create(ctx, "user-123", "token", "ip", "ua")
	require.True(t, createResult.IsOk())
	created := createResult.Unwrap()

	// Wait for expiration
	time.Sleep(20 * time.Millisecond)

	// Validate
	result := manager.Validate(ctx, created.ID)

	assert.True(t, result.IsErr())
	assert.Equal(t, ErrSessionExpired, result.UnwrapErr())
}
