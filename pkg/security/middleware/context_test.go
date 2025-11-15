package middleware

import (
	"context"
	"testing"

	"github.com/0xsj/nexus/pkg/security/token/jwt"
	"github.com/stretchr/testify/assert"
)

func TestWithClaims_ClaimsFromContext(t *testing.T) {
	claims := &jwt.Claims{
		UserID:   "user-123",
		Email:    "test@example.com",
		Username: "testuser",
		Roles:    []string{"admin", "user"},
	}

	ctx := WithClaims(context.Background(), claims)

	retrieved, ok := ClaimsFromContext(ctx)

	assert.True(t, ok)
	assert.Equal(t, claims.UserID, retrieved.UserID)
	assert.Equal(t, claims.Email, retrieved.Email)
	assert.Equal(t, claims.Username, retrieved.Username)
	assert.Equal(t, claims.Roles, retrieved.Roles)
}

func TestClaimsFromContext_NotFound(t *testing.T) {
	ctx := context.Background()

	claims, ok := ClaimsFromContext(ctx)

	assert.False(t, ok)
	assert.Nil(t, claims)
}

func TestMustClaimsFromContext_Success(t *testing.T) {
	claims := &jwt.Claims{
		UserID: "user-123",
	}

	ctx := WithClaims(context.Background(), claims)

	retrieved := MustClaimsFromContext(ctx)

	assert.Equal(t, claims.UserID, retrieved.UserID)
}

func TestMustClaimsFromContext_Panic(t *testing.T) {
	ctx := context.Background()

	assert.Panics(t, func() {
		MustClaimsFromContext(ctx)
	})
}

func TestWithUserID_UserIDFromContext(t *testing.T) {
	userID := "user-123"

	ctx := WithUserID(context.Background(), userID)

	retrieved, ok := UserIDFromContext(ctx)

	assert.True(t, ok)
	assert.Equal(t, userID, retrieved)
}

func TestUserIDFromContext_NotFound(t *testing.T) {
	ctx := context.Background()

	userID, ok := UserIDFromContext(ctx)

	assert.False(t, ok)
	assert.Empty(t, userID)
}

func TestMustUserIDFromContext_Success(t *testing.T) {
	userID := "user-123"

	ctx := WithUserID(context.Background(), userID)

	retrieved := MustUserIDFromContext(ctx)

	assert.Equal(t, userID, retrieved)
}

func TestMustUserIDFromContext_Panic(t *testing.T) {
	ctx := context.Background()

	assert.Panics(t, func() {
		MustUserIDFromContext(ctx)
	})
}

func TestUserInfoFromContext_Success(t *testing.T) {
	claims := &jwt.Claims{
		UserID:   "user-123",
		Email:    "test@example.com",
		Username: "testuser",
		Roles:    []string{"admin", "user"},
	}

	ctx := WithClaims(context.Background(), claims)

	userInfo, ok := UserInfoFromContext(ctx)

	assert.True(t, ok)
	assert.Equal(t, claims.UserID, userInfo.UserID)
	assert.Equal(t, claims.Email, userInfo.Email)
	assert.Equal(t, claims.Username, userInfo.Username)
	assert.Equal(t, claims.Roles, userInfo.Roles)
}

func TestUserInfoFromContext_NotFound(t *testing.T) {
	ctx := context.Background()

	userInfo, ok := UserInfoFromContext(ctx)

	assert.False(t, ok)
	assert.Nil(t, userInfo)
}

func TestMustUserInfoFromContext_Success(t *testing.T) {
	claims := &jwt.Claims{
		UserID:   "user-123",
		Email:    "test@example.com",
		Username: "testuser",
		Roles:    []string{"admin"},
	}

	ctx := WithClaims(context.Background(), claims)

	userInfo := MustUserInfoFromContext(ctx)

	assert.Equal(t, claims.UserID, userInfo.UserID)
}

func TestMustUserInfoFromContext_Panic(t *testing.T) {
	ctx := context.Background()

	assert.Panics(t, func() {
		MustUserInfoFromContext(ctx)
	})
}

func TestIsAuthenticated(t *testing.T) {
	t.Run("authenticated", func(t *testing.T) {
		claims := &jwt.Claims{UserID: "user-123"}
		ctx := WithClaims(context.Background(), claims)

		assert.True(t, IsAuthenticated(ctx))
	})

	t.Run("not authenticated", func(t *testing.T) {
		ctx := context.Background()

		assert.False(t, IsAuthenticated(ctx))
	})
}

func TestHasRole(t *testing.T) {
	claims := &jwt.Claims{
		UserID: "user-123",
		Roles:  []string{"admin", "editor"},
	}

	ctx := WithClaims(context.Background(), claims)

	t.Run("has role", func(t *testing.T) {
		assert.True(t, HasRole(ctx, "admin"))
		assert.True(t, HasRole(ctx, "editor"))
	})

	t.Run("does not have role", func(t *testing.T) {
		assert.False(t, HasRole(ctx, "superadmin"))
	})

	t.Run("no claims in context", func(t *testing.T) {
		ctx := context.Background()
		assert.False(t, HasRole(ctx, "admin"))
	})
}

func TestHasAnyRole(t *testing.T) {
	claims := &jwt.Claims{
		UserID: "user-123",
		Roles:  []string{"editor"},
	}

	ctx := WithClaims(context.Background(), claims)

	t.Run("has one of the roles", func(t *testing.T) {
		assert.True(t, HasAnyRole(ctx, "admin", "editor"))
	})

	t.Run("does not have any role", func(t *testing.T) {
		assert.False(t, HasAnyRole(ctx, "admin", "superadmin"))
	})

	t.Run("no claims in context", func(t *testing.T) {
		ctx := context.Background()
		assert.False(t, HasAnyRole(ctx, "admin"))
	})
}

func TestHasAllRoles(t *testing.T) {
	claims := &jwt.Claims{
		UserID: "user-123",
		Roles:  []string{"admin", "editor", "viewer"},
	}

	ctx := WithClaims(context.Background(), claims)

	t.Run("has all roles", func(t *testing.T) {
		assert.True(t, HasAllRoles(ctx, "admin", "editor"))
	})

	t.Run("does not have all roles", func(t *testing.T) {
		assert.False(t, HasAllRoles(ctx, "admin", "superadmin"))
	})

	t.Run("no claims in context", func(t *testing.T) {
		ctx := context.Background()
		assert.False(t, HasAllRoles(ctx, "admin"))
	})
}
