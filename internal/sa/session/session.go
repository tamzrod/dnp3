// Package session implements DNP3 Secure Authentication session management.
package session

import (
	"sync"
	"time"

	"dnp3/internal/sa"
)

// Session state
const (
	StateInactive = iota
	StatePendingChallenge
	StateAuthenticated
	StateExpired
)

// Session represents an authenticated session
type Session struct {
	mu           sync.RWMutex
	UserNumber   uint8
	Role         sa.Role
	Seq          uint8        // Session sequence number
	SessionKey   [16]byte     // Session encryption key
	MACKey       [16]byte     // MAC verification key
	State        int
	CreatedAt    time.Time
	LastActivity time.Time
	ExpiresAt    time.Time
	AuthSeq      uint8        // Authentication sequence number
}

// NewSession creates a new authenticated session
func NewSession(userNumber uint8, role sa.Role, sessionKey [16]byte, macKey [16]byte, timeout time.Duration) *Session {
	now := time.Now()
	return &Session{
		UserNumber:   userNumber,
		Role:         role,
		Seq:          0,
		SessionKey:   sessionKey,
		MACKey:       macKey,
		State:        StateAuthenticated,
		CreatedAt:    now,
		LastActivity: now,
		ExpiresAt:    now.Add(timeout),
		AuthSeq:      0,
	}
}

// IsActive returns true if the session is active
func (s *Session) IsActive() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	return s.State == StateAuthenticated && time.Now().Before(s.ExpiresAt)
}

// Extend extends the session timeout
func (s *Session) Extend(timeout time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.State != StateAuthenticated {
		return sa.ErrSessionNotActive
	}
	
	// Also check if session has expired
	if time.Now().After(s.ExpiresAt) {
		s.State = StateExpired
		return sa.ErrSessionNotActive
	}
	
	s.LastActivity = time.Now()
	s.ExpiresAt = s.LastActivity.Add(timeout)
	return nil
}

// Invalidate invalidates the session
func (s *Session) Invalidate() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.State = StateInactive
}

// IncrementSeq increments the session sequence number
func (s *Session) IncrementSeq() uint8 {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.Seq = (s.Seq + 1) & 0x0F
	s.LastActivity = time.Now()
	return s.Seq
}

// IncrementAuthSeq increments the authentication sequence number
func (s *Session) IncrementAuthSeq() uint8 {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.AuthSeq = (s.AuthSeq + 1) & 0x3F
	s.LastActivity = time.Now()
	return s.AuthSeq
}

// GetSeq returns the current session sequence number
func (s *Session) GetSeq() uint8 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	return s.Seq
}

// GetAuthSeq returns the current authentication sequence number
func (s *Session) GetAuthSeq() uint8 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	return s.AuthSeq
}

// GetRole returns the session role
func (s *Session) GetRole() sa.Role {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	return s.Role
}

// GetUserNumber returns the session user number
func (s *Session) GetUserNumber() uint8 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	return s.UserNumber
}

// RemainingTime returns the remaining session time
func (s *Session) RemainingTime() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	remaining := time.Until(s.ExpiresAt)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// SessionManager manages active sessions
type SessionManager struct {
	mu       sync.RWMutex
	sessions map[uint8]*Session // UserNumber -> Session
	timeout  time.Duration
}

// NewSessionManager creates a new session manager
func NewSessionManager(timeout time.Duration) *SessionManager {
	return &SessionManager{
		sessions: make(map[uint8]*Session),
		timeout:  timeout,
	}
}

// CreateSession creates a new authenticated session
func (sm *SessionManager) CreateSession(userNumber uint8, role sa.Role, sessionKey [16]byte, macKey [16]byte) (*Session, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	// Check if session already exists
	if _, exists := sm.sessions[userNumber]; exists {
		sm.sessions[userNumber].Invalidate()
	}
	
	session := NewSession(userNumber, role, sessionKey, macKey, sm.timeout)
	sm.sessions[userNumber] = session
	
	return session, nil
}

// GetSession retrieves a session by user number
func (sm *SessionManager) GetSession(userNumber uint8) (*Session, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	session, exists := sm.sessions[userNumber]
	if !exists {
		return nil, false
	}
	
	// Check if session is still active
	if !session.IsActive() {
		delete(sm.sessions, userNumber)
		return nil, false
	}
	
	return session, true
}

// RemoveSession removes a session
func (sm *SessionManager) RemoveSession(userNumber uint8) bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	if _, exists := sm.sessions[userNumber]; !exists {
		return false
	}
	
	delete(sm.sessions, userNumber)
	return true
}

// InvalidateSession invalidates a session
func (sm *SessionManager) InvalidateSession(userNumber uint8) bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	session, exists := sm.sessions[userNumber]
	if !exists {
		return false
	}
	
	session.Invalidate()
	delete(sm.sessions, userNumber)
	return true
}

// Cleanup removes expired sessions
func (sm *SessionManager) Cleanup() int {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	count := 0
	now := time.Now()
	for userNumber, session := range sm.sessions {
		if now.After(session.ExpiresAt) || session.State == StateInactive {
			delete(sm.sessions, userNumber)
			count++
		}
	}
	return count
}

// ActiveSessionCount returns the number of active sessions
func (sm *SessionManager) ActiveSessionCount() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	count := 0
	now := time.Now()
	for _, session := range sm.sessions {
		if session.State == StateAuthenticated && now.Before(session.ExpiresAt) {
			count++
		}
	}
	return count
}

// ListActiveSessions returns a list of active session user numbers
func (sm *SessionManager) ListActiveSessions() []uint8 {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	users := make([]uint8, 0)
	now := time.Now()
	for userNumber, session := range sm.sessions {
		if session.State == StateAuthenticated && now.Before(session.ExpiresAt) {
			users = append(users, userNumber)
		}
	}
	return users
}

// HasRole checks if a user has at least the specified role
func HasRole(userRole, requiredRole sa.Role) bool {
	return userRole >= requiredRole
}

// AuthorizeControl checks if a role is authorized for critical controls
func AuthorizeControl(role sa.Role, isCritical bool) bool {
	if isCritical {
		return role >= sa.RoleLevel2
	}
	return role >= sa.RoleLevel1
}

// AuthorizeRead checks if a role is authorized for read operations
func AuthorizeRead(role sa.Role) bool {
	return role >= sa.RoleRemote
}

// AuthorizeManagement checks if a role is authorized for configuration
func AuthorizeManagement(role sa.Role) bool {
	return role >= sa.RoleManager
}
