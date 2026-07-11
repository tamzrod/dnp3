package session

import (
	"testing"
	"time"

	"dnp3/internal/sa"
)

func TestNewSession(t *testing.T) {
	userNumber := uint8(5)
	role := sa.RoleLevel2
	var sessionKey, macKey [16]byte
	for i := range sessionKey {
		sessionKey[i] = byte(i)
		macKey[i] = byte(0xFF - i)
	}
	timeout := 5 * time.Minute
	
	session := NewSession(userNumber, role, sessionKey, macKey, timeout)
	
	if session.UserNumber != userNumber {
		t.Errorf("Session.UserNumber = %d, want %d", session.UserNumber, userNumber)
	}
	
	if session.Role != role {
		t.Errorf("Session.Role = %v, want %v", session.Role, role)
	}
	
	if session.State != StateAuthenticated {
		t.Errorf("Session.State = %d, want %d (StateAuthenticated)", session.State, StateAuthenticated)
	}
	
	if session.Seq != 0 {
		t.Errorf("Session.Seq = %d, want 0", session.Seq)
	}
}

func TestSessionIsActive(t *testing.T) {
	var sessionKey, macKey [16]byte
	session := NewSession(1, sa.RoleRemote, sessionKey, macKey, 5*time.Minute)
	
	if !session.IsActive() {
		t.Error("New session should be active")
	}
}

func TestSessionInvalidate(t *testing.T) {
	var sessionKey, macKey [16]byte
	session := NewSession(1, sa.RoleRemote, sessionKey, macKey, 5*time.Minute)
	
	session.Invalidate()
	
	if session.IsActive() {
		t.Error("Invalidated session should not be active")
	}
	
	if session.State != StateInactive {
		t.Errorf("Session.State = %d, want %d (StateInactive)", session.State, StateInactive)
	}
}

func TestSessionExtend(t *testing.T) {
	var sessionKey, macKey [16]byte
	session := NewSession(1, sa.RoleLevel1, sessionKey, macKey, 1*time.Millisecond)
	
	// Wait for session to be near expiration
	time.Sleep(2 * time.Millisecond)
	
	// Extend should fail if expired
	err := session.Extend(5 * time.Minute)
	if err == nil {
		t.Error("Extend() should fail for expired session")
	}
	
	// Create fresh session
	session = NewSession(1, sa.RoleLevel1, sessionKey, macKey, 5*time.Minute)
	
	// Extend should succeed
	err = session.Extend(10 * time.Minute)
	if err != nil {
		t.Errorf("Extend() error = %v", err)
	}
	
	// Remaining time should be longer
	remaining := session.RemainingTime()
	if remaining < 9*time.Minute {
		t.Errorf("RemainingTime() = %v, want >= 9 minutes", remaining)
	}
}

func TestSessionIncrementSeq(t *testing.T) {
	var sessionKey, macKey [16]byte
	session := NewSession(1, sa.RoleRemote, sessionKey, macKey, 5*time.Minute)
	
	// Test sequence wrapping at 0x0F
	seen := make(map[uint8]bool)
	
	// First increment returns 1 (starts at 0)
	firstSeq := session.IncrementSeq()
	if firstSeq != 1 {
		t.Errorf("First IncrementSeq() = %d, want 1", firstSeq)
	}
	seen[firstSeq] = true
	
	// Subsequent increments should increase
	for i := 0; i < 14; i++ {
		seq := session.IncrementSeq()
		if seen[seq] {
			t.Errorf("Sequence number %d was repeated", seq)
		}
		seen[seq] = true
	}
	
	// After 15 increments, we should have seen 1-15
	// Next increment should wrap to 0
	seq := session.IncrementSeq()
	if seq != 0 {
		t.Errorf("After wrap, IncrementSeq() = %d, want 0", seq)
	}
}

func TestSessionGetSeq(t *testing.T) {
	var sessionKey, macKey [16]byte
	session := NewSession(1, sa.RoleRemote, sessionKey, macKey, 5*time.Minute)
	
	session.IncrementSeq()
	session.IncrementSeq()
	
	seq := session.GetSeq()
	if seq != 2 {
		t.Errorf("GetSeq() = %d, want 2", seq)
	}
}

func TestSessionGetRole(t *testing.T) {
	var sessionKey, macKey [16]byte
	session := NewSession(10, sa.RoleLevel2, sessionKey, macKey, 5*time.Minute)
	
	role := session.GetRole()
	if role != sa.RoleLevel2 {
		t.Errorf("GetRole() = %v, want %v", role, sa.RoleLevel2)
	}
}

func TestSessionGetUserNumber(t *testing.T) {
	var sessionKey, macKey [16]byte
	session := NewSession(42, sa.RoleManager, sessionKey, macKey, 5*time.Minute)
	
	userNumber := session.GetUserNumber()
	if userNumber != 42 {
		t.Errorf("GetUserNumber() = %d, want 42", userNumber)
	}
}

func TestNewSessionManager(t *testing.T) {
	sm := NewSessionManager(5 * time.Minute)
	if sm == nil {
		t.Fatal("NewSessionManager() returned nil")
	}
	
	if sm.ActiveSessionCount() != 0 {
		t.Errorf("NewSessionManager().ActiveSessionCount() = %d, want 0", sm.ActiveSessionCount())
	}
}

func TestCreateSession(t *testing.T) {
	sm := NewSessionManager(5 * time.Minute)
	
	userNumber := uint8(10)
	role := sa.RoleLevel1
	var sessionKey, macKey [16]byte
	for i := range sessionKey {
		sessionKey[i] = byte(i)
		macKey[i] = byte(0xFF - i)
	}
	
	session, err := sm.CreateSession(userNumber, role, sessionKey, macKey)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	
	if session.UserNumber != userNumber {
		t.Errorf("Session.UserNumber = %d, want %d", session.UserNumber, userNumber)
	}
	
	if sm.ActiveSessionCount() != 1 {
		t.Errorf("ActiveSessionCount() = %d, want 1", sm.ActiveSessionCount())
	}
}

func TestCreateSessionDuplicate(t *testing.T) {
	sm := NewSessionManager(5 * time.Minute)
	
	var sessionKey, macKey [16]byte
	_, _ = sm.CreateSession(5, sa.RoleRemote, sessionKey, macKey)
	session2, _ := sm.CreateSession(5, sa.RoleLevel2, sessionKey, macKey)
	
	// Should replace old session, not create duplicate
	if sm.ActiveSessionCount() != 1 {
		t.Errorf("ActiveSessionCount() = %d, want 1", sm.ActiveSessionCount())
	}
	
	// New session should have updated role
	if session2.Role != sa.RoleLevel2 {
		t.Errorf("Replaced session.Role = %v, want %v", session2.Role, sa.RoleLevel2)
	}
}

func TestGetSession(t *testing.T) {
	sm := NewSessionManager(5 * time.Minute)
	
	var sessionKey, macKey [16]byte
	created, _ := sm.CreateSession(15, sa.RoleLevel1, sessionKey, macKey)
	
	retrieved, ok := sm.GetSession(15)
	if !ok {
		t.Fatal("GetSession(15) returned false for existing session")
	}
	
	if retrieved != created {
		t.Error("GetSession() returned different session")
	}
}

func TestGetSessionNotFound(t *testing.T) {
	sm := NewSessionManager(5 * time.Minute)
	
	_, ok := sm.GetSession(99)
	if ok {
		t.Error("GetSession(99) returned true for non-existent session")
	}
}

func TestRemoveSession(t *testing.T) {
	sm := NewSessionManager(5 * time.Minute)
	
	var sessionKey, macKey [16]byte
	sm.CreateSession(20, sa.RoleLevel2, sessionKey, macKey)
	
	ok := sm.RemoveSession(20)
	if !ok {
		t.Error("RemoveSession(20) returned false for existing session")
	}
	
	if sm.ActiveSessionCount() != 0 {
		t.Errorf("ActiveSessionCount() = %d, want 0", sm.ActiveSessionCount())
	}
	
	// Remove non-existent
	ok = sm.RemoveSession(99)
	if ok {
		t.Error("RemoveSession(99) returned true for non-existent session")
	}
}

func TestInvalidateSession(t *testing.T) {
	sm := NewSessionManager(5 * time.Minute)
	
	var sessionKey, macKey [16]byte
	sm.CreateSession(25, sa.RoleManager, sessionKey, macKey)
	
	ok := sm.InvalidateSession(25)
	if !ok {
		t.Error("InvalidateSession(25) returned false for existing session")
	}
	
	// Session should no longer be retrievable
	_, ok = sm.GetSession(25)
	if ok {
		t.Error("GetSession(25) returned true for invalidated session")
	}
}

func TestCleanup(t *testing.T) {
	sm := NewSessionManager(1 * time.Millisecond)
	
	var sessionKey, macKey [16]byte
	sm.CreateSession(1, sa.RoleRemote, sessionKey, macKey)
	sm.CreateSession(2, sa.RoleLevel1, sessionKey, macKey)
	
	// Wait for sessions to expire
	time.Sleep(5 * time.Millisecond)
	
	count := sm.Cleanup()
	if count != 2 {
		t.Errorf("Cleanup() returned %d, want 2", count)
	}
	
	if sm.ActiveSessionCount() != 0 {
		t.Errorf("ActiveSessionCount() = %d, want 0 after cleanup", sm.ActiveSessionCount())
	}
}

func TestListActiveSessions(t *testing.T) {
	sm := NewSessionManager(5 * time.Minute)
	
	var sessionKey, macKey [16]byte
	sm.CreateSession(1, sa.RoleRemote, sessionKey, macKey)
	sm.CreateSession(5, sa.RoleLevel1, sessionKey, macKey)
	sm.CreateSession(10, sa.RoleLevel2, sessionKey, macKey)
	
	users := sm.ListActiveSessions()
	
	if len(users) != 3 {
		t.Errorf("ListActiveSessions() len = %d, want 3", len(users))
	}
	
	// Check all expected users are present
	expected := map[uint8]bool{1: true, 5: true, 10: true}
	for _, u := range users {
		if !expected[u] {
			t.Errorf("Unexpected user %d in ListActiveSessions()", u)
		}
	}
}

func TestHasRole(t *testing.T) {
	tests := []struct {
		userRole     sa.Role
		requiredRole sa.Role
		expected     bool
	}{
		{sa.RoleRemote, sa.RoleRemote, true},
		{sa.RoleLevel1, sa.RoleRemote, true},
		{sa.RoleLevel2, sa.RoleLevel1, true},
		{sa.RoleManager, sa.RoleLevel2, true},
		{sa.RoleRemote, sa.RoleLevel1, false},
		{sa.RoleLevel1, sa.RoleLevel2, false},
	}
	
	for _, tt := range tests {
		result := HasRole(tt.userRole, tt.requiredRole)
		if result != tt.expected {
			t.Errorf("HasRole(%v, %v) = %v, want %v", tt.userRole, tt.requiredRole, result, tt.expected)
		}
	}
}

func TestAuthorizeControl(t *testing.T) {
	// Non-critical control requires Level1 or higher
	if AuthorizeControl(sa.RoleRemote, false) {
		t.Error("AuthorizeControl(RoleRemote, false) = true, want false")
	}
	if !AuthorizeControl(sa.RoleLevel1, false) {
		t.Error("AuthorizeControl(RoleLevel1, false) = false, want true")
	}
	if !AuthorizeControl(sa.RoleLevel2, false) {
		t.Error("AuthorizeControl(RoleLevel2, false) = false, want true")
	}
	
	// Critical control requires Level2 or higher
	if AuthorizeControl(sa.RoleRemote, true) {
		t.Error("AuthorizeControl(RoleRemote, true) = true, want false")
	}
	if AuthorizeControl(sa.RoleLevel1, true) {
		t.Error("AuthorizeControl(RoleLevel1, true) = true, want false")
	}
	if !AuthorizeControl(sa.RoleLevel2, true) {
		t.Error("AuthorizeControl(RoleLevel2, true) = false, want true")
	}
	if !AuthorizeControl(sa.RoleManager, true) {
		t.Error("AuthorizeControl(RoleManager, true) = false, want true")
	}
}

func TestAuthorizeRead(t *testing.T) {
	if !AuthorizeRead(sa.RoleRemote) {
		t.Error("AuthorizeRead(RoleRemote) = false, want true")
	}
	if !AuthorizeRead(sa.RoleLevel1) {
		t.Error("AuthorizeRead(RoleLevel1) = false, want true")
	}
	if !AuthorizeRead(sa.RoleManager) {
		t.Error("AuthorizeRead(RoleManager) = false, want true")
	}
}

func TestAuthorizeManagement(t *testing.T) {
	if AuthorizeManagement(sa.RoleRemote) {
		t.Error("AuthorizeManagement(RoleRemote) = true, want false")
	}
	if AuthorizeManagement(sa.RoleLevel1) {
		t.Error("AuthorizeManagement(RoleLevel1) = true, want false")
	}
	if AuthorizeManagement(sa.RoleLevel2) {
		t.Error("AuthorizeManagement(RoleLevel2) = true, want false")
	}
	if !AuthorizeManagement(sa.RoleManager) {
		t.Error("AuthorizeManagement(RoleManager) = false, want true")
	}
}
