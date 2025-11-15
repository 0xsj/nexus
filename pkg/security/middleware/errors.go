package middleware

import (
	"errors"
	"fmt"
)

var (
	ErrMissingAuthHeader  = errors.New("missing authorization header")
	ErrInvalidAuthHeader  = errors.New("invalid authorization header format")
	ErrMissingBearerToken = errors.New("missing bearer token")
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token expired")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden: insufficient permissions")
	ErrMissingUserInCtx   = errors.New("user not found in context")
)

// ErrMissingRole is returned when user doesn't have required role.
type ErrMissingRole struct {
	Required []string
	Actual   []string
}

func (e ErrMissingRole) Error() string {
	return fmt.Sprintf("missing required role: need one of %v, have %v", e.Required, e.Actual)
}

// ErrMissingPermission is returned when user doesn't have required permission.
type ErrMissingPermission struct {
	Required string
}

func (e ErrMissingPermission) Error() string {
	return fmt.Sprintf("missing required permission: %s", e.Required)
}
