package domain

import (
	"time"

	"github.com/0xsj/nexus/pkg/events"
)

const (
	// Event types
	EventTypeUserAuthenticated = "auth.user.authenticated"
	EventTypeSessionCreated    = "auth.session.created"
	EventTypeSessionDeleted    = "auth.session.deleted"
	EventTypeUserLoggedOut     = "auth.user.logged_out"
	EventTypeMagicLinkSent     = "auth.magic_link.sent"
	EventTypeMagicLinkVerified = "auth.magic_link.verified"
	EventTypeOAuthConnected    = "auth.oauth.connected"
	EventTypeOAuthVerified     = "auth.oauth.verified"

	// Aggregate type
	AggregateTypeAuth = "auth"
)

// UserAuthenticatedPayload is the payload for user authenticated event.
type UserAuthenticatedPayload struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	Method    string `json:"method"` // "password", "magic_link", "oauth"
	IPAddress string `json:"ip_address"`
	UserAgent string `json:"user_agent"`
}

// NewUserAuthenticatedEvent creates a new user authenticated event.
func NewUserAuthenticatedEvent(userID, email, method, ipAddress, userAgent string) events.Event {
	payload := UserAuthenticatedPayload{
		UserID:    userID,
		Email:     email,
		Method:    method,
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}

	event := events.NewBaseEvent(
		EventTypeUserAuthenticated,
		AggregateTypeAuth,
		userID,
		payload,
	)

	event.WithMetadata("auth_method", method)
	event.WithMetadata("ip_address", ipAddress)

	return event
}

// SessionCreatedPayload is the payload for session created event.
type SessionCreatedPayload struct {
	SessionID string    `json:"session_id"`
	UserID    string    `json:"user_id"`
	DeviceID  *string   `json:"device_id,omitempty"`
	ExpiresAt time.Time `json:"expires_at"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
}

// NewSessionCreatedEvent creates a new session created event.
func NewSessionCreatedEvent(
	sessionID, userID string,
	deviceID *string,
	expiresAt time.Time,
	ipAddress, userAgent string,
) events.Event {
	payload := SessionCreatedPayload{
		SessionID: sessionID,
		UserID:    userID,
		DeviceID:  deviceID,
		ExpiresAt: expiresAt,
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}

	event := events.NewBaseEvent(
		EventTypeSessionCreated,
		AggregateTypeAuth,
		userID,
		payload,
	)

	event.WithMetadata("session_id", sessionID)

	return event
}

// SessionDeletedPayload is the payload for session deleted event.
type SessionDeletedPayload struct {
	SessionID string `json:"session_id"`
	UserID    string `json:"user_id"`
	Reason    string `json:"reason"` // "logout", "expired", "revoked"
}

// NewSessionDeletedEvent creates a new session deleted event.
func NewSessionDeletedEvent(sessionID, userID, reason string) events.Event {
	payload := SessionDeletedPayload{
		SessionID: sessionID,
		UserID:    userID,
		Reason:    reason,
	}

	event := events.NewBaseEvent(
		EventTypeSessionDeleted,
		AggregateTypeAuth,
		userID,
		payload,
	)

	event.WithMetadata("session_id", sessionID)
	event.WithMetadata("reason", reason)

	return event
}

// UserLoggedOutPayload is the payload for user logged out event.
type UserLoggedOutPayload struct {
	UserID     string `json:"user_id"`
	SessionID  string `json:"session_id,omitempty"`
	LogoutType string `json:"logout_type"` // "single", "all_devices"
}

// NewUserLoggedOutEvent creates a new user logged out event.
func NewUserLoggedOutEvent(userID, sessionID, logoutType string) events.Event {
	payload := UserLoggedOutPayload{
		UserID:     userID,
		SessionID:  sessionID,
		LogoutType: logoutType,
	}

	event := events.NewBaseEvent(
		EventTypeUserLoggedOut,
		AggregateTypeAuth,
		userID,
		payload,
	)

	event.WithMetadata("logout_type", logoutType)

	return event
}

// MagicLinkSentPayload is the payload for magic link sent event.
type MagicLinkSentPayload struct {
	Email     string    `json:"email"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
}

// NewMagicLinkSentEvent creates a new magic link sent event.
func NewMagicLinkSentEvent(email, token string, expiresAt time.Time, ipAddress, userAgent string) events.Event {
	payload := MagicLinkSentPayload{
		Email:     email,
		Token:     token,
		ExpiresAt: expiresAt,
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}

	event := events.NewBaseEvent(
		EventTypeMagicLinkSent,
		AggregateTypeAuth,
		email,
		payload,
	)

	event.WithMetadata("ip_address", ipAddress)

	return event
}

// MagicLinkVerifiedPayload is the payload for magic link verified event.
type MagicLinkVerifiedPayload struct {
	Email       string  `json:"email"`
	UserID      *string `json:"user_id,omitempty"` // nil if new user signup
	IPAddress   string  `json:"ip_address"`
	UserAgent   string  `json:"user_agent"`
	MagicLinkID string  `json:"magic_link_id"`
}

// NewMagicLinkVerifiedEvent creates a magic link verified event.
func NewMagicLinkVerifiedEvent(
	email string,
	userID *string,
	ipAddress, userAgent, magicLinkID string,
) events.Event {
	payload := MagicLinkVerifiedPayload{
		Email:       email,
		UserID:      userID,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		MagicLinkID: magicLinkID,
	}

	aggregateID := email
	if userID != nil {
		aggregateID = *userID
	}

	event := events.NewBaseEvent(
		EventTypeMagicLinkVerified,
		AggregateTypeAuth,
		aggregateID,
		payload,
	)

	event.WithMetadata("ip_address", ipAddress)

	return event
}

// OAuthConnectedPayload is the payload for OAuth connected event.
type OAuthConnectedPayload struct {
	UserID         string        `json:"user_id"`
	Provider       OAuthProvider `json:"provider"`
	ProviderUserID string        `json:"provider_user_id"`
	Email          string        `json:"email"`
}

// NewOAuthConnectedEvent creates a new OAuth connected event.
func NewOAuthConnectedEvent(userID string, provider OAuthProvider, providerUserID, email string) events.Event {
	payload := OAuthConnectedPayload{
		UserID:         userID,
		Provider:       provider,
		ProviderUserID: providerUserID,
		Email:          email,
	}

	event := events.NewBaseEvent(
		EventTypeOAuthConnected,
		AggregateTypeAuth,
		userID,
		payload,
	)

	event.WithMetadata("provider", string(provider))

	return event
}

// OAuthVerifiedPayload is the payload for OAuth verified event.
type OAuthVerifiedPayload struct {
	Provider       OAuthProvider `json:"provider"`
	ProviderUserID string        `json:"provider_user_id"`
	Email          string        `json:"email"`
	Name           string        `json:"name"`
	Picture        string        `json:"picture"`
	AccessToken    string        `json:"access_token"`
	RefreshToken   string        `json:"refresh_token,omitempty"`
	TokenExpiresAt *time.Time    `json:"token_expires_at,omitempty"`
	IPAddress      string        `json:"ip_address"`
	UserAgent      string        `json:"user_agent"`
}

// NewOAuthVerifiedEvent creates an OAuth verified event.
func NewOAuthVerifiedEvent(
	provider OAuthProvider,
	providerUserID, email, name, picture string,
	accessToken, refreshToken string,
	tokenExpiresAt *time.Time,
	ipAddress, userAgent string,
) events.Event {
	payload := OAuthVerifiedPayload{
		Provider:       provider,
		ProviderUserID: providerUserID,
		Email:          email,
		Name:           name,
		Picture:        picture,
		AccessToken:    accessToken,
		RefreshToken:   refreshToken,
		TokenExpiresAt: tokenExpiresAt,
		IPAddress:      ipAddress,
		UserAgent:      userAgent,
	}

	event := events.NewBaseEvent(
		EventTypeOAuthVerified,
		AggregateTypeAuth,
		email,
		payload,
	)

	event.WithMetadata("provider", string(provider))
	event.WithMetadata("ip_address", ipAddress)

	return event
}
