package session

import (
	"context"
	"sync"
	"time"

	"github.com/0xsj/result"
)

// MemoryStore is an in-memory implementation of Store for testing.
type MemoryStore struct {
	mu              sync.RWMutex
	sessions        map[string]*Session // sessionID -> Session
	sessionsByToken map[string]string   // refreshToken -> sessionID
	sessionsByUser  map[string][]string // userID -> []sessionID
}

// NewMemoryStore creates a new in-memory session store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		sessions:        make(map[string]*Session),
		sessionsByToken: make(map[string]string),
		sessionsByUser:  make(map[string][]string),
	}
}

// Create creates a new session.
func (s *MemoryStore) Create(ctx context.Context, session *Session) result.Result[*Session] {
	if session == nil {
		return result.Err[*Session](ErrSessionNotFound)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Store session
	s.sessions[session.ID] = session
	s.sessionsByToken[session.RefreshToken] = session.ID

	// Add to user's sessions
	s.sessionsByUser[session.UserID] = append(s.sessionsByUser[session.UserID], session.ID)

	return result.Ok(session)
}

// GetByID retrieves a session by its ID.
func (s *MemoryStore) GetByID(ctx context.Context, sessionID string) result.Result[*Session] {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return result.Err[*Session](ErrSessionNotFound)
	}

	return result.Ok(session)
}

// GetByRefreshToken retrieves a session by its refresh token.
func (s *MemoryStore) GetByRefreshToken(ctx context.Context, refreshToken string) result.Result[*Session] {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sessionID, exists := s.sessionsByToken[refreshToken]
	if !exists {
		return result.Err[*Session](ErrSessionNotFound)
	}

	session, exists := s.sessions[sessionID]
	if !exists {
		return result.Err[*Session](ErrSessionNotFound)
	}

	return result.Ok(session)
}

// Update updates an existing session.
func (s *MemoryStore) Update(ctx context.Context, session *Session) result.Result[*Session] {
	if session == nil {
		return result.Err[*Session](ErrSessionNotFound)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	existing, exists := s.sessions[session.ID]
	if !exists {
		return result.Err[*Session](ErrSessionNotFound)
	}

	// Update token mapping if refresh token changed
	if existing.RefreshToken != session.RefreshToken {
		delete(s.sessionsByToken, existing.RefreshToken)
		s.sessionsByToken[session.RefreshToken] = session.ID
	}

	// Update session
	s.sessions[session.ID] = session

	return result.Ok(session)
}

// Delete deletes a session by ID.
func (s *MemoryStore) Delete(ctx context.Context, sessionID string) result.Result[struct{}] {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return result.Ok(struct{}{}) // Idempotent
	}

	// Remove from all mappings
	delete(s.sessions, sessionID)
	delete(s.sessionsByToken, session.RefreshToken)

	// Remove from user's sessions
	userSessions := s.sessionsByUser[session.UserID]
	for i, id := range userSessions {
		if id == sessionID {
			s.sessionsByUser[session.UserID] = append(userSessions[:i], userSessions[i+1:]...)
			break
		}
	}

	return result.Ok(struct{}{})
}

// DeleteByRefreshToken deletes a session by its refresh token.
func (s *MemoryStore) DeleteByRefreshToken(ctx context.Context, refreshToken string) result.Result[struct{}] {
	s.mu.Lock()
	defer s.mu.Unlock()

	sessionID, exists := s.sessionsByToken[refreshToken]
	if !exists {
		return result.Ok(struct{}{}) // Idempotent
	}

	session, exists := s.sessions[sessionID]
	if !exists {
		return result.Ok(struct{}{})
	}

	// Remove from all mappings
	delete(s.sessions, sessionID)
	delete(s.sessionsByToken, refreshToken)

	// Remove from user's sessions
	userSessions := s.sessionsByUser[session.UserID]
	for i, id := range userSessions {
		if id == sessionID {
			s.sessionsByUser[session.UserID] = append(userSessions[:i], userSessions[i+1:]...)
			break
		}
	}

	return result.Ok(struct{}{})
}

// DeleteAllForUser deletes all sessions for a specific user.
func (s *MemoryStore) DeleteAllForUser(ctx context.Context, userID string) result.Result[int] {
	s.mu.Lock()
	defer s.mu.Unlock()

	sessionIDs, exists := s.sessionsByUser[userID]
	if !exists {
		return result.Ok(0)
	}

	count := 0
	for _, sessionID := range sessionIDs {
		session, exists := s.sessions[sessionID]
		if !exists {
			continue
		}

		delete(s.sessions, sessionID)
		delete(s.sessionsByToken, session.RefreshToken)
		count++
	}

	delete(s.sessionsByUser, userID)

	return result.Ok(count)
}

// ListByUser retrieves all active sessions for a user.
func (s *MemoryStore) ListByUser(ctx context.Context, userID string) result.Result[[]*Session] {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sessionIDs, exists := s.sessionsByUser[userID]
	if !exists {
		return result.Ok([]*Session{})
	}

	sessions := make([]*Session, 0, len(sessionIDs))
	for _, sessionID := range sessionIDs {
		if session, exists := s.sessions[sessionID]; exists {
			sessions = append(sessions, session)
		}
	}

	return result.Ok(sessions)
}

// CountByUser counts active sessions for a user.
func (s *MemoryStore) CountByUser(ctx context.Context, userID string) result.Result[int] {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sessionIDs, exists := s.sessionsByUser[userID]
	if !exists {
		return result.Ok(0)
	}

	return result.Ok(len(sessionIDs))
}

// DeleteExpired removes all expired sessions.
func (s *MemoryStore) DeleteExpired(ctx context.Context) result.Result[int] {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	count := 0

	// Find expired sessions
	expiredSessions := make([]*Session, 0)
	for _, session := range s.sessions {
		if now.After(session.ExpiresAt) {
			expiredSessions = append(expiredSessions, session)
		}
	}

	// Delete expired sessions
	for _, session := range expiredSessions {
		delete(s.sessions, session.ID)
		delete(s.sessionsByToken, session.RefreshToken)

		// Remove from user's sessions
		userSessions := s.sessionsByUser[session.UserID]
		for i, id := range userSessions {
			if id == session.ID {
				s.sessionsByUser[session.UserID] = append(userSessions[:i], userSessions[i+1:]...)
				break
			}
		}

		count++
	}

	return result.Ok(count)
}

// ExtendExpiry extends a session's expiry time.
func (s *MemoryStore) ExtendExpiry(ctx context.Context, sessionID string, duration time.Duration) result.Result[*Session] {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return result.Err[*Session](ErrSessionNotFound)
	}

	session.Extend(duration)
	s.sessions[sessionID] = session

	return result.Ok(session)
}
