package postgres

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/0xsj/nexus/platform/internal/identity/domain"
	"github.com/0xsj/nexus/platform/internal/identity/infrastructure/persistence/postgres/generated"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
	"github.com/0xsj/nexus/platform/pkg/types"
)

// ============================================================================
// User Mapper Tests
// ============================================================================

func TestUserToUpsertParams(t *testing.T) {
	t.Run("user with email", func(t *testing.T) {
		userID := domain.NewUserID()
		email, _ := types.NewEmail("test@example.com")
		displayName, _ := domain.NewDisplayName("Test User")
		primaryDID := "did:key:z6MkTest123"

		user, err := domain.NewUser(userID, email, displayName, domain.AuthMethodMagicLink, primaryDID)
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		params := UserToUpsertParams(user)

		if params.ID.String() != userID.String() {
			t.Errorf("ID mismatch: got %s, want %s", params.ID, userID)
		}
		if params.Email == nil {
			t.Error("Email should not be nil")
		} else if *params.Email != "test@example.com" {
			t.Errorf("Email mismatch: got %s, want test@example.com", *params.Email)
		}
		if params.DisplayName != "Test User" {
			t.Errorf("DisplayName mismatch: got %s, want Test User", params.DisplayName)
		}
		if params.Status != "pending" {
			t.Errorf("Status mismatch: got %s, want pending", params.Status)
		}
		if params.PrimaryDid != primaryDID {
			t.Errorf("PrimaryDid mismatch: got %s, want %s", params.PrimaryDid, primaryDID)
		}
		if params.Version != 1 {
			t.Errorf("Version mismatch: got %d, want 1", params.Version)
		}
	})

	t.Run("user without email", func(t *testing.T) {
		userID := domain.NewUserID()
		displayName, _ := domain.NewDisplayName("Wallet User")
		primaryDID := "did:pkh:eip155:1:0x1234567890abcdef"

		user, err := domain.NewUser(userID, types.Email{}, displayName, domain.AuthMethodWallet, primaryDID)
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		params := UserToUpsertParams(user)

		if params.Email != nil {
			t.Errorf("Email should be nil for wallet user, got %v", params.Email)
		}
	})
}

func TestUserDIDToInsertParams(t *testing.T) {
	userID := domain.NewUserID()
	did := "did:key:z6MkTest456"
	addedAt := time.Now().UTC()

	t.Run("primary DID", func(t *testing.T) {
		params := UserDIDToInsertParams(userID, did, true, addedAt)

		if params.UserID.String() != userID.String() {
			t.Errorf("UserID mismatch: got %s, want %s", params.UserID, userID)
		}
		if params.Did != did {
			t.Errorf("Did mismatch: got %s, want %s", params.Did, did)
		}
		if !params.IsPrimary {
			t.Error("IsPrimary should be true")
		}
		if params.ID == uuid.Nil {
			t.Error("ID should be generated")
		}
	})

	t.Run("secondary DID", func(t *testing.T) {
		params := UserDIDToInsertParams(userID, did, false, addedAt)

		if params.IsPrimary {
			t.Error("IsPrimary should be false")
		}
	})
}

func TestUserOAuthLinkToInsertParams(t *testing.T) {
	userID := domain.NewUserID()
	subject, _ := domain.NewOAuthSubject(domain.OAuthProviderGoogle, "google-123")
	linkedAt := time.Now().UTC()

	t.Run("with email", func(t *testing.T) {
		params := UserOAuthLinkToInsertParams(userID, subject, "oauth@example.com", linkedAt)

		if params.UserID.String() != userID.String() {
			t.Errorf("UserID mismatch: got %s, want %s", params.UserID, userID)
		}
		if params.Provider != "google" {
			t.Errorf("Provider mismatch: got %s, want google", params.Provider)
		}
		if params.ExternalID != "google-123" {
			t.Errorf("ExternalID mismatch: got %s, want google-123", params.ExternalID)
		}
		if params.Email == nil || *params.Email != "oauth@example.com" {
			t.Errorf("Email mismatch: got %v, want oauth@example.com", params.Email)
		}
	})

	t.Run("without email", func(t *testing.T) {
		params := UserOAuthLinkToInsertParams(userID, subject, "", linkedAt)

		if params.Email != nil {
			t.Errorf("Email should be nil, got %v", params.Email)
		}
	})
}

// ============================================================================
// Session Mapper Tests
// ============================================================================

func TestSessionToUpsertParams(t *testing.T) {
	sessionID := domain.NewSessionID()
	userID := domain.NewUserID()
	token, _ := domain.NewToken()

	t.Run("with IP and user agent", func(t *testing.T) {
		session, err := domain.NewSession(
			sessionID,
			userID,
			token,
			domain.AuthMethodMagicLink,
			"192.168.1.1",
			"Mozilla/5.0",
			24*time.Hour,
		)
		if err != nil {
			t.Fatalf("failed to create session: %v", err)
		}

		params := SessionToUpsertParams(session)

		if params.ID.String() != sessionID.String() {
			t.Errorf("ID mismatch: got %s, want %s", params.ID, sessionID)
		}
		if params.UserID.String() != userID.String() {
			t.Errorf("UserID mismatch: got %s, want %s", params.UserID, userID)
		}
		if params.TokenHash != token.HashString() {
			t.Errorf("TokenHash mismatch: got %s, want %s", params.TokenHash, token.HashString())
		}
		if params.AuthMethod != "magic_link" {
			t.Errorf("AuthMethod mismatch: got %s, want magic_link", params.AuthMethod)
		}
		if params.Status != "active" {
			t.Errorf("Status mismatch: got %s, want active", params.Status)
		}
		if params.IpAddress == nil || *params.IpAddress != "192.168.1.1" {
			t.Errorf("IpAddress mismatch: got %v, want 192.168.1.1", params.IpAddress)
		}
		if params.UserAgent == nil || *params.UserAgent != "Mozilla/5.0" {
			t.Errorf("UserAgent mismatch: got %v, want Mozilla/5.0", params.UserAgent)
		}
	})

	t.Run("without IP and user agent", func(t *testing.T) {
		session, err := domain.NewSession(
			sessionID,
			userID,
			token,
			domain.AuthMethodWallet,
			"",
			"",
			24*time.Hour,
		)
		if err != nil {
			t.Fatalf("failed to create session: %v", err)
		}

		params := SessionToUpsertParams(session)

		if params.IpAddress != nil {
			t.Errorf("IpAddress should be nil, got %v", params.IpAddress)
		}
		if params.UserAgent != nil {
			t.Errorf("UserAgent should be nil, got %v", params.UserAgent)
		}
	})
}

func TestSessionRowToData(t *testing.T) {
	sessionUUID := uuid.New()
	userUUID := uuid.New()
	ipAddress := "10.0.0.1"
	userAgent := "TestAgent/1.0"
	now := time.Now().UTC()
	expiresAt := now.Add(24 * time.Hour)

	row := generated.IdentitySession{
		ID:         sessionUUID,
		UserID:     userUUID,
		TokenHash:  "abc123hash",
		AuthMethod: "magic_link",
		Status:     "active",
		IpAddress:  &ipAddress,
		UserAgent:  &userAgent,
		ExpiresAt:  expiresAt,
		CreatedAt:  now,
		UpdatedAt:  now,
		Version:    5,
	}

	data := SessionRowToData(row)

	if data.ID.String() != sessionUUID.String() {
		t.Errorf("ID mismatch: got %s, want %s", data.ID, sessionUUID)
	}
	if data.UserID.String() != userUUID.String() {
		t.Errorf("UserID mismatch: got %s, want %s", data.UserID, userUUID)
	}
	if data.TokenHash != "abc123hash" {
		t.Errorf("TokenHash mismatch: got %s, want abc123hash", data.TokenHash)
	}
	if data.AuthMethod != domain.AuthMethodMagicLink {
		t.Errorf("AuthMethod mismatch: got %v, want %v", data.AuthMethod, domain.AuthMethodMagicLink)
	}
	if data.Status != domain.SessionStatusActive {
		t.Errorf("Status mismatch: got %v, want %v", data.Status, domain.SessionStatusActive)
	}
	if data.IPAddress != "10.0.0.1" {
		t.Errorf("IPAddress mismatch: got %s, want 10.0.0.1", data.IPAddress)
	}
	if data.UserAgent != "TestAgent/1.0" {
		t.Errorf("UserAgent mismatch: got %s, want TestAgent/1.0", data.UserAgent)
	}
	if data.Version != 5 {
		t.Errorf("Version mismatch: got %d, want 5", data.Version)
	}
}

func TestSessionRowToData_NilOptionalFields(t *testing.T) {
	row := generated.IdentitySession{
		ID:         uuid.New(),
		UserID:     uuid.New(),
		TokenHash:  "hash",
		AuthMethod: "wallet",
		Status:     "revoked",
		IpAddress:  nil,
		UserAgent:  nil,
		ExpiresAt:  time.Now(),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Version:    1,
	}

	data := SessionRowToData(row)

	if data.IPAddress != "" {
		t.Errorf("IPAddress should be empty, got %s", data.IPAddress)
	}
	if data.UserAgent != "" {
		t.Errorf("UserAgent should be empty, got %s", data.UserAgent)
	}
}

// ============================================================================
// Magic Link Mapper Tests
// ============================================================================

func TestMagicLinkRecordToInsertParams(t *testing.T) {
	now := time.Now().UTC()
	record := &domain.MagicLinkRecord{
		TokenHash: "magiclinkhash123",
		Email:     "magic@example.com",
		ExpiresAt: now.Add(15 * time.Minute).Unix(),
		Used:      false,
		CreatedAt: now.Unix(),
	}

	params := MagicLinkRecordToInsertParams(record)

	if params.TokenHash != "magiclinkhash123" {
		t.Errorf("TokenHash mismatch: got %s, want magiclinkhash123", params.TokenHash)
	}
	if params.Email != "magic@example.com" {
		t.Errorf("Email mismatch: got %s, want magic@example.com", params.Email)
	}
	if params.Used {
		t.Error("Used should be false")
	}
	if params.ID == uuid.Nil {
		t.Error("ID should be generated")
	}
}

func TestMagicLinkRowToRecord(t *testing.T) {
	now := time.Now().UTC()
	expiresAt := now.Add(15 * time.Minute)

	// Create a valid pgtype.Timestamptz for UsedAt
	usedAt := pgtype.Timestamptz{
		Time:  now,
		Valid: true,
	}

	row := generated.IdentityMagicLink{
		ID:        uuid.New(),
		TokenHash: "rowlinkhash",
		Email:     "row@example.com",
		ExpiresAt: expiresAt,
		Used:      true,
		UsedAt:    usedAt,
		CreatedAt: now,
	}

	record := MagicLinkRowToRecord(row)

	if record.TokenHash != "rowlinkhash" {
		t.Errorf("TokenHash mismatch: got %s, want rowlinkhash", record.TokenHash)
	}
	if record.Email != "row@example.com" {
		t.Errorf("Email mismatch: got %s, want row@example.com", record.Email)
	}
	if record.ExpiresAt != expiresAt.Unix() {
		t.Errorf("ExpiresAt mismatch: got %d, want %d", record.ExpiresAt, expiresAt.Unix())
	}
	if !record.Used {
		t.Error("Used should be true")
	}
	if record.CreatedAt != now.Unix() {
		t.Errorf("CreatedAt mismatch: got %d, want %d", record.CreatedAt, now.Unix())
	}
}

func TestMagicLinkRowToRecord_UnusedLink(t *testing.T) {
	now := time.Now().UTC()

	// UsedAt is invalid (not set)
	row := generated.IdentityMagicLink{
		ID:        uuid.New(),
		TokenHash: "unusedhash",
		Email:     "unused@example.com",
		ExpiresAt: now.Add(15 * time.Minute),
		Used:      false,
		UsedAt:    pgtype.Timestamptz{Valid: false},
		CreatedAt: now,
	}

	record := MagicLinkRowToRecord(row)

	if record.Used {
		t.Error("Used should be false")
	}
}

// ============================================================================
// OAuth State Mapper Tests
// ============================================================================

func TestOAuthStateRecordToInsertParams(t *testing.T) {
	now := time.Now().UTC()
	record := &domain.OAuthStateRecord{
		State:       "random-state-123",
		Provider:    "github",
		RedirectURL: "https://example.com/callback",
		ExpiresAt:   now.Add(10 * time.Minute).Unix(),
		CreatedAt:   now.Unix(),
	}

	params := OAuthStateRecordToInsertParams(record)

	if params.State != "random-state-123" {
		t.Errorf("State mismatch: got %s, want random-state-123", params.State)
	}
	if params.Provider != "github" {
		t.Errorf("Provider mismatch: got %s, want github", params.Provider)
	}
	if params.RedirectUrl != "https://example.com/callback" {
		t.Errorf("RedirectUrl mismatch: got %s, want https://example.com/callback", params.RedirectUrl)
	}
	if params.ID == uuid.Nil {
		t.Error("ID should be generated")
	}
}

func TestOAuthStateRowToRecord(t *testing.T) {
	now := time.Now().UTC()
	expiresAt := now.Add(10 * time.Minute)

	row := generated.IdentityOauthState{
		ID:          uuid.New(),
		State:       "state-from-row",
		Provider:    "google",
		RedirectUrl: "https://app.com/oauth/callback",
		ExpiresAt:   expiresAt,
		CreatedAt:   now,
	}

	record := OAuthStateRowToRecord(row)

	if record.State != "state-from-row" {
		t.Errorf("State mismatch: got %s, want state-from-row", record.State)
	}
	if record.Provider != "google" {
		t.Errorf("Provider mismatch: got %s, want google", record.Provider)
	}
	if record.RedirectURL != "https://app.com/oauth/callback" {
		t.Errorf("RedirectURL mismatch: got %s, want https://app.com/oauth/callback", record.RedirectURL)
	}
	if record.ExpiresAt != expiresAt.Unix() {
		t.Errorf("ExpiresAt mismatch: got %d, want %d", record.ExpiresAt, expiresAt.Unix())
	}
}

// ============================================================================
// Event Mapper Tests
// ============================================================================

func TestUserEventToInsertParams(t *testing.T) {
	userID := domain.NewUserID()
	now := time.Now().UTC()

	event := &domain.UserRegisteredEvent{
		BaseEvent:    eventsourcing.NewBaseEvent(domain.AggregateTypeUser, userID.String()),
		UserID:       userID.String(),
		Email:        "event@example.com",
		DisplayName:  "Event User",
		AuthMethod:   "magic_link",
		PrimaryDID:   "did:key:z6MkEvent",
		RegisteredAt: now,
	}

	params, err := UserEventToInsertParams(userID.String(), event, 1)
	if err != nil {
		t.Fatalf("failed to create params: %v", err)
	}

	if params.AggregateID.String() != userID.String() {
		t.Errorf("AggregateID mismatch: got %s, want %s", params.AggregateID, userID)
	}
	if params.AggregateType != domain.AggregateTypeUser {
		t.Errorf("AggregateType mismatch: got %s, want %s", params.AggregateType, domain.AggregateTypeUser)
	}
	if params.EventType != domain.EventTypeUserRegistered {
		t.Errorf("EventType mismatch: got %s, want %s", params.EventType, domain.EventTypeUserRegistered)
	}
	if params.Version != 1 {
		t.Errorf("Version mismatch: got %d, want 1", params.Version)
	}
	if params.ID == uuid.Nil {
		t.Error("ID should be generated")
	}

	// Verify event data is valid JSON
	var decoded domain.UserRegisteredEvent
	if err := json.Unmarshal(params.EventData, &decoded); err != nil {
		t.Errorf("failed to unmarshal event data: %v", err)
	}
	if decoded.Email != "event@example.com" {
		t.Errorf("decoded Email mismatch: got %s, want event@example.com", decoded.Email)
	}
}

func TestUserEventToInsertParams_InvalidAggregateID(t *testing.T) {
	event := &domain.UserActivatedEvent{
		UserID:      "some-id",
		ActivatedAt: time.Now(),
	}

	_, err := UserEventToInsertParams("not-a-uuid", event, 1)
	if err == nil {
		t.Error("expected error for invalid aggregate ID")
	}
}

func TestUserEventRowToEvent(t *testing.T) {
	now := time.Now().UTC()

	tests := []struct {
		name      string
		eventType string
		eventData map[string]any
		validate  func(t *testing.T, event interface{})
	}{
		{
			name:      "UserRegistered",
			eventType: domain.EventTypeUserRegistered,
			eventData: map[string]any{
				"user_id":       "user-123",
				"email":         "test@example.com",
				"display_name":  "Test User",
				"auth_method":   "magic_link",
				"primary_did":   "did:key:z6Mk123",
				"registered_at": now.Format(time.RFC3339Nano),
			},
			validate: func(t *testing.T, event interface{}) {
				e, ok := event.(*domain.UserRegisteredEvent)
				if !ok {
					t.Fatalf("expected *UserRegisteredEvent, got %T", event)
				}
				if e.Email != "test@example.com" {
					t.Errorf("Email mismatch: got %s", e.Email)
				}
			},
		},
		{
			name:      "UserActivated",
			eventType: domain.EventTypeUserActivated,
			eventData: map[string]any{
				"user_id":      "user-123",
				"activated_at": now.Format(time.RFC3339Nano),
			},
			validate: func(t *testing.T, event interface{}) {
				e, ok := event.(*domain.UserActivatedEvent)
				if !ok {
					t.Fatalf("expected *UserActivatedEvent, got %T", event)
				}
				if e.UserID != "user-123" {
					t.Errorf("UserID mismatch: got %s", e.UserID)
				}
			},
		},
		{
			name:      "UserSuspended",
			eventType: domain.EventTypeUserSuspended,
			eventData: map[string]any{
				"user_id":      "user-123",
				"reason":       "policy violation",
				"suspended_at": now.Format(time.RFC3339Nano),
			},
			validate: func(t *testing.T, event interface{}) {
				e, ok := event.(*domain.UserSuspendedEvent)
				if !ok {
					t.Fatalf("expected *UserSuspendedEvent, got %T", event)
				}
				if e.Reason != "policy violation" {
					t.Errorf("Reason mismatch: got %s", e.Reason)
				}
			},
		},
		{
			name:      "UserDIDAdded",
			eventType: domain.EventTypeUserDIDAdded,
			eventData: map[string]any{
				"user_id":  "user-123",
				"did":      "did:pkh:eip155:1:0xabc",
				"added_at": now.Format(time.RFC3339Nano),
			},
			validate: func(t *testing.T, event interface{}) {
				e, ok := event.(*domain.UserDIDAddedEvent)
				if !ok {
					t.Fatalf("expected *UserDIDAddedEvent, got %T", event)
				}
				if e.DID != "did:pkh:eip155:1:0xabc" {
					t.Errorf("DID mismatch: got %s", e.DID)
				}
			},
		},
		{
			name:      "UserOAuthLinked",
			eventType: domain.EventTypeUserOAuthLinked,
			eventData: map[string]any{
				"user_id":     "user-123",
				"provider":    "github",
				"external_id": "gh-456",
				"email":       "github@example.com",
				"linked_at":   now.Format(time.RFC3339Nano),
			},
			validate: func(t *testing.T, event interface{}) {
				e, ok := event.(*domain.UserOAuthLinkedEvent)
				if !ok {
					t.Fatalf("expected *UserOAuthLinkedEvent, got %T", event)
				}
				if e.Provider != "github" {
					t.Errorf("Provider mismatch: got %s", e.Provider)
				}
				if e.ExternalID != "gh-456" {
					t.Errorf("ExternalID mismatch: got %s", e.ExternalID)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, _ := json.Marshal(tt.eventData)
			row := generated.IdentityUserEvent{
				ID:            uuid.New(),
				AggregateID:   uuid.New(),
				AggregateType: domain.AggregateTypeUser,
				EventType:     tt.eventType,
				EventData:     data,
				Version:       1,
				OccurredAt:    now,
			}

			event, err := UserEventRowToEvent(row)
			if err != nil {
				t.Fatalf("failed to unmarshal event: %v", err)
			}

			tt.validate(t, event)
		})
	}
}

func TestUserEventRowToEvent_UnknownEventType(t *testing.T) {
	row := generated.IdentityUserEvent{
		ID:            uuid.New(),
		AggregateID:   uuid.New(),
		AggregateType: domain.AggregateTypeUser,
		EventType:     "User.UnknownEvent",
		EventData:     []byte("{}"),
		Version:       1,
		OccurredAt:    time.Now(),
	}

	_, err := UserEventRowToEvent(row)
	if err == nil {
		t.Error("expected error for unknown event type")
	}

	unknownErr, ok := err.(*UnknownEventTypeError)
	if !ok {
		t.Errorf("expected *UnknownEventTypeError, got %T", err)
	}
	if unknownErr.EventType != "User.UnknownEvent" {
		t.Errorf("EventType mismatch: got %s", unknownErr.EventType)
	}
}

func TestSessionEventRowToEvent(t *testing.T) {
	now := time.Now().UTC()

	tests := []struct {
		name      string
		eventType string
		eventData map[string]any
		validate  func(t *testing.T, event interface{})
	}{
		{
			name:      "SessionStarted",
			eventType: domain.EventTypeSessionStarted,
			eventData: map[string]any{
				"session_id":  "session-123",
				"user_id":     "user-123",
				"token_hash":  "hash123",
				"auth_method": "wallet",
				"ip_address":  "192.168.1.100",
				"user_agent":  "Chrome/100",
				"expires_at":  now.Add(24 * time.Hour).Format(time.RFC3339Nano),
				"started_at":  now.Format(time.RFC3339Nano),
			},
			validate: func(t *testing.T, event interface{}) {
				e, ok := event.(*domain.SessionStartedEvent)
				if !ok {
					t.Fatalf("expected *SessionStartedEvent, got %T", event)
				}
				if e.AuthMethod != "wallet" {
					t.Errorf("AuthMethod mismatch: got %s", e.AuthMethod)
				}
			},
		},
		{
			name:      "SessionRefreshed",
			eventType: domain.EventTypeSessionRefreshed,
			eventData: map[string]any{
				"session_id":     "session-123",
				"old_token_hash": "oldhash",
				"new_token_hash": "newhash",
				"expires_at":     now.Add(48 * time.Hour).Format(time.RFC3339Nano),
				"refreshed_at":   now.Format(time.RFC3339Nano),
			},
			validate: func(t *testing.T, event interface{}) {
				e, ok := event.(*domain.SessionRefreshedEvent)
				if !ok {
					t.Fatalf("expected *SessionRefreshedEvent, got %T", event)
				}
				if e.OldTokenHash != "oldhash" {
					t.Errorf("OldTokenHash mismatch: got %s", e.OldTokenHash)
				}
				if e.NewTokenHash != "newhash" {
					t.Errorf("NewTokenHash mismatch: got %s", e.NewTokenHash)
				}
			},
		},
		{
			name:      "SessionRevoked",
			eventType: domain.EventTypeSessionRevoked,
			eventData: map[string]any{
				"session_id": "session-123",
				"reason":     "user logout",
				"revoked_at": now.Format(time.RFC3339Nano),
			},
			validate: func(t *testing.T, event interface{}) {
				e, ok := event.(*domain.SessionRevokedEvent)
				if !ok {
					t.Fatalf("expected *SessionRevokedEvent, got %T", event)
				}
				if e.Reason != "user logout" {
					t.Errorf("Reason mismatch: got %s", e.Reason)
				}
			},
		},
		{
			name:      "SessionExpired",
			eventType: domain.EventTypeSessionExpired,
			eventData: map[string]any{
				"session_id": "session-123",
				"expired_at": now.Format(time.RFC3339Nano),
			},
			validate: func(t *testing.T, event interface{}) {
				_, ok := event.(*domain.SessionExpiredEvent)
				if !ok {
					t.Fatalf("expected *SessionExpiredEvent, got %T", event)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, _ := json.Marshal(tt.eventData)
			row := generated.IdentitySessionEvent{
				ID:            uuid.New(),
				AggregateID:   uuid.New(),
				AggregateType: domain.AggregateTypeSession,
				EventType:     tt.eventType,
				EventData:     data,
				Version:       1,
				OccurredAt:    now,
			}

			event, err := SessionEventRowToEvent(row)
			if err != nil {
				t.Fatalf("failed to unmarshal event: %v", err)
			}

			tt.validate(t, event)
		})
	}
}

func TestUserEventRowsToEvents(t *testing.T) {
	now := time.Now().UTC()

	registeredData, _ := json.Marshal(map[string]any{
		"user_id":       "user-123",
		"email":         "test@example.com",
		"display_name":  "Test",
		"auth_method":   "magic_link",
		"primary_did":   "did:key:123",
		"registered_at": now.Format(time.RFC3339Nano),
	})

	activatedData, _ := json.Marshal(map[string]any{
		"user_id":      "user-123",
		"activated_at": now.Format(time.RFC3339Nano),
	})

	rows := []generated.IdentityUserEvent{
		{
			ID:            uuid.New(),
			AggregateID:   uuid.New(),
			AggregateType: domain.AggregateTypeUser,
			EventType:     domain.EventTypeUserRegistered,
			EventData:     registeredData,
			Version:       1,
			OccurredAt:    now,
		},
		{
			ID:            uuid.New(),
			AggregateID:   uuid.New(),
			AggregateType: domain.AggregateTypeUser,
			EventType:     domain.EventTypeUserActivated,
			EventData:     activatedData,
			Version:       2,
			OccurredAt:    now,
		},
	}

	events, err := UserEventRowsToEvents(rows)
	if err != nil {
		t.Fatalf("failed to convert rows: %v", err)
	}

	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}

	if _, ok := events[0].(*domain.UserRegisteredEvent); !ok {
		t.Errorf("first event should be UserRegisteredEvent, got %T", events[0])
	}
	if _, ok := events[1].(*domain.UserActivatedEvent); !ok {
		t.Errorf("second event should be UserActivatedEvent, got %T", events[1])
	}
}

func TestSessionEventRowsToEvents(t *testing.T) {
	now := time.Now().UTC()

	startedData, _ := json.Marshal(map[string]any{
		"session_id":  "session-123",
		"user_id":     "user-123",
		"token_hash":  "hash",
		"auth_method": "magic_link",
		"expires_at":  now.Add(24 * time.Hour).Format(time.RFC3339Nano),
		"started_at":  now.Format(time.RFC3339Nano),
	})

	revokedData, _ := json.Marshal(map[string]any{
		"session_id": "session-123",
		"reason":     "logout",
		"revoked_at": now.Format(time.RFC3339Nano),
	})

	rows := []generated.IdentitySessionEvent{
		{
			ID:            uuid.New(),
			AggregateID:   uuid.New(),
			AggregateType: domain.AggregateTypeSession,
			EventType:     domain.EventTypeSessionStarted,
			EventData:     startedData,
			Version:       1,
			OccurredAt:    now,
		},
		{
			ID:            uuid.New(),
			AggregateID:   uuid.New(),
			AggregateType: domain.AggregateTypeSession,
			EventType:     domain.EventTypeSessionRevoked,
			EventData:     revokedData,
			Version:       2,
			OccurredAt:    now,
		},
	}

	events, err := SessionEventRowsToEvents(rows)
	if err != nil {
		t.Fatalf("failed to convert rows: %v", err)
	}

	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}

	if _, ok := events[0].(*domain.SessionStartedEvent); !ok {
		t.Errorf("first event should be SessionStartedEvent, got %T", events[0])
	}
	if _, ok := events[1].(*domain.SessionRevokedEvent); !ok {
		t.Errorf("second event should be SessionRevokedEvent, got %T", events[1])
	}
}

// ============================================================================
// ID Conversion Tests
// ============================================================================

func TestUUIDConversions(t *testing.T) {
	t.Run("UserID roundtrip", func(t *testing.T) {
		original := domain.NewUserID()
		asUUID := uuidFromUserID(original)
		back := userIDFromUUID(asUUID)

		if !original.Equals(back) {
			t.Errorf("roundtrip failed: %s != %s", original, back)
		}
	})

	t.Run("SessionID roundtrip", func(t *testing.T) {
		original := domain.NewSessionID()
		asUUID := uuidFromSessionID(original)
		back := sessionIDFromUUID(asUUID)

		if !original.Equals(back) {
			t.Errorf("roundtrip failed: %s != %s", original, back)
		}
	})
}

// ============================================================================
// Error Tests
// ============================================================================

func TestUnknownEventTypeError(t *testing.T) {
	err := &UnknownEventTypeError{EventType: "Unknown.Event"}

	if err.Error() != "unknown event type: Unknown.Event" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}
