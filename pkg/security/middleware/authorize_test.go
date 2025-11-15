package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/0xsj/nexus/pkg/security/token/jwt"
	"github.com/stretchr/testify/assert"
)

func createContextWithClaims(roles []string) *http.Request {
	claims := &jwt.Claims{
		UserID:   "user-123",
		Email:    "test@example.com",
		Username: "testuser",
		Roles:    roles,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := WithClaims(req.Context(), claims)
	ctx = WithUserID(ctx, claims.UserID) // ✅ Add UserID to context
	return req.WithContext(ctx)
}

func TestRequireRole_Success(t *testing.T) {
	req := createContextWithClaims([]string{"admin", "user"})
	rec := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequireRole("admin")
	middleware(handler).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRequireRole_MissingRole(t *testing.T) {
	req := createContextWithClaims([]string{"user"})
	rec := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not reach handler")
	})

	middleware := RequireRole("admin")
	middleware(handler).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestRequireRole_NoClaims(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not reach handler")
	})

	middleware := RequireRole("admin")
	middleware(handler).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestRequireAnyRole_Success(t *testing.T) {
	req := createContextWithClaims([]string{"editor"})
	rec := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequireAnyRole("admin", "editor")
	middleware(handler).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRequireAnyRole_MissingAllRoles(t *testing.T) {
	req := createContextWithClaims([]string{"viewer"})
	rec := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not reach handler")
	})

	middleware := RequireAnyRole("admin", "editor")
	middleware(handler).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestRequireAllRoles_Success(t *testing.T) {
	req := createContextWithClaims([]string{"admin", "editor", "viewer"})
	rec := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequireAllRoles("admin", "editor")
	middleware(handler).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRequireAllRoles_MissingOneRole(t *testing.T) {
	req := createContextWithClaims([]string{"admin"})
	rec := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not reach handler")
	})

	middleware := RequireAllRoles("admin", "editor")
	middleware(handler).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestRequireAuthenticated_Success(t *testing.T) {
	req := createContextWithClaims([]string{"user"})
	rec := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequireAuthenticated()
	middleware(handler).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRequireAuthenticated_NotAuthenticated(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not reach handler")
	})

	middleware := RequireAuthenticated()
	middleware(handler).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestRequireOwnership_Success(t *testing.T) {
	req := createContextWithClaims([]string{"user"})
	rec := httptest.NewRecorder()

	ownerExtractor := func(r *http.Request) string {
		return "user-123" // Same as authenticated user
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequireOwnership(ownerExtractor)
	middleware(handler).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRequireOwnership_NotOwner(t *testing.T) {
	req := createContextWithClaims([]string{"user"})
	rec := httptest.NewRecorder()

	ownerExtractor := func(r *http.Request) string {
		return "user-456" // Different user
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not reach handler")
	})

	middleware := RequireOwnership(ownerExtractor)
	middleware(handler).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestRequireOwnershipOrRole_OwnerSuccess(t *testing.T) {
	req := createContextWithClaims([]string{"user"})
	rec := httptest.NewRecorder()

	ownerExtractor := func(r *http.Request) string {
		return "user-123" // Same as authenticated user
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequireOwnershipOrRole(ownerExtractor, "admin")
	middleware(handler).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRequireOwnershipOrRole_AdminSuccess(t *testing.T) {
	req := createContextWithClaims([]string{"admin"})
	rec := httptest.NewRecorder()

	ownerExtractor := func(r *http.Request) string {
		return "user-456" // Different user, but has admin role
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequireOwnershipOrRole(ownerExtractor, "admin")
	middleware(handler).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRequireOwnershipOrRole_Forbidden(t *testing.T) {
	req := createContextWithClaims([]string{"user"})
	rec := httptest.NewRecorder()

	ownerExtractor := func(r *http.Request) string {
		return "user-456" // Different user, no admin role
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not reach handler")
	})

	middleware := RequireOwnershipOrRole(ownerExtractor, "admin")
	middleware(handler).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestCustomAuthorize_Success(t *testing.T) {
	req := createContextWithClaims([]string{"user"})
	rec := httptest.NewRecorder()

	authorizer := func(r *http.Request) bool {
		return true // Always allow
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := CustomAuthorize(authorizer)
	middleware(handler).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestCustomAuthorize_Forbidden(t *testing.T) {
	req := createContextWithClaims([]string{"user"})
	rec := httptest.NewRecorder()

	authorizer := func(r *http.Request) bool {
		return false // Always deny
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not reach handler")
	})

	middleware := CustomAuthorize(authorizer)
	middleware(handler).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}
