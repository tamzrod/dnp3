package challenge

import (
	"testing"
	"time"

	"dnp3/internal/sa"
)

func TestNewManager(t *testing.T) {
	m := NewManager(DefaultChallengeTimeout)
	if m == nil {
		t.Fatal("NewManager() returned nil")
	}
	
	if m.challengeTimeout != DefaultChallengeTimeout {
		t.Errorf("Manager.timeout = %v, want %v", m.challengeTimeout, DefaultChallengeTimeout)
	}
}

func TestGenerateChallenge(t *testing.T) {
	m := NewManager(DefaultChallengeTimeout)
	
	seq := uint8(10)
	role := sa.RoleLevel2
	
	challenge, err := m.GenerateChallenge(seq, role)
	if err != nil {
		t.Fatalf("GenerateChallenge() error = %v", err)
	}
	
	if challenge.Seq != seq {
		t.Errorf("Challenge.Seq = %d, want %d", challenge.Seq, seq)
	}
	
	if challenge.Role != role {
		t.Errorf("Challenge.Role = %v, want %v", challenge.Role, role)
	}
	
	// Check pending challenge exists
	pending, ok := m.GetChallenge(seq)
	if !ok {
		t.Error("GetChallenge() returned false for just-generated challenge")
	}
	
	if pending.Challenge.Seq != seq {
		t.Errorf("PendingChallenge.Challenge.Seq = %d, want %d", pending.Challenge.Seq, seq)
	}
}

func TestGetChallengeNotFound(t *testing.T) {
	m := NewManager(DefaultChallengeTimeout)
	
	_, ok := m.GetChallenge(99)
	if ok {
		t.Error("GetChallenge(99) returned true for non-existent challenge")
	}
}

func TestValidateResponse(t *testing.T) {
	m := NewManager(DefaultChallengeTimeout)
	
	seq := uint8(5)
	role := sa.RoleLevel1
	sessionKey := [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	authData := []byte{0x01, 0x02}
	
	// Generate challenge
	challenge, err := m.GenerateChallenge(seq, role)
	if err != nil {
		t.Fatalf("GenerateChallenge() error = %v", err)
	}
	
	// Build data to verify
	dataToVerify := make([]byte, 0, sa.ChallengeSize+len(authData))
	dataToVerify = append(dataToVerify, challenge.ChallengeData[:]...)
	dataToVerify = append(dataToVerify, authData...)
	
	// Calculate MAC
	mac, err := sa.CalculateMAC(dataToVerify, sessionKey)
	if err != nil {
		t.Fatalf("CalculateMAC() error = %v", err)
	}
	
	// Validate response
	err = m.ValidateResponse(seq, mac, sessionKey, authData)
	if err != nil {
		t.Errorf("ValidateResponse() error = %v", err)
	}
}

func TestValidateResponseChallengeExpired(t *testing.T) {
	// Very short timeout to trigger expiration
	m := NewManager(1 * time.Millisecond)
	
	seq := uint8(7)
	role := sa.RoleLevel2
	sessionKey := [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	
	// Generate challenge
	challenge, _ := m.GenerateChallenge(seq, role)
	
	// Build data to verify
	dataToVerify := make([]byte, 0, sa.ChallengeSize)
	dataToVerify = append(dataToVerify, challenge.ChallengeData[:]...)
	
	mac, _ := sa.CalculateMAC(dataToVerify, sessionKey)
	
	// Wait for expiration
	time.Sleep(10 * time.Millisecond)
	
	// Validate should fail
	err := m.ValidateResponse(seq, mac, sessionKey, nil)
	if err != ErrChallengeExpired {
		t.Errorf("ValidateResponse() error = %v, want %v", err, ErrChallengeExpired)
	}
}

func TestValidateResponseChallengeNotFound(t *testing.T) {
	m := NewManager(DefaultChallengeTimeout)
	
	sessionKey := [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	
	// Try to validate non-existent challenge
	err := m.ValidateResponse(99, sessionKey, sessionKey, nil)
	if err != ErrChallengeNotFound {
		t.Errorf("ValidateResponse() error = %v, want %v", err, ErrChallengeNotFound)
	}
}

func TestValidateResponseChallengeUsed(t *testing.T) {
	m := NewManager(DefaultChallengeTimeout)
	
	seq := uint8(8)
	role := sa.RoleLevel2
	sessionKey := [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	
	// Generate and use challenge
	challenge, _ := m.GenerateChallenge(seq, role)
	
	dataToVerify := make([]byte, 0, sa.ChallengeSize)
	dataToVerify = append(dataToVerify, challenge.ChallengeData[:]...)
	
	mac, _ := sa.CalculateMAC(dataToVerify, sessionKey)
	
	// First validation should succeed
	err := m.ValidateResponse(seq, mac, sessionKey, nil)
	if err != nil {
		t.Errorf("First ValidateResponse() error = %v", err)
	}
	
	// Second validation should fail (challenge already used)
	err = m.ValidateResponse(seq, mac, sessionKey, nil)
	if err != ErrChallengeUsed {
		t.Errorf("Second ValidateResponse() error = %v, want %v", err, ErrChallengeUsed)
	}
}

func TestClearChallenge(t *testing.T) {
	m := NewManager(DefaultChallengeTimeout)
	
	seq := uint8(15)
	role := sa.RoleRemote
	
	_, _ = m.GenerateChallenge(seq, role)
	
	// Verify it exists
	_, ok := m.GetChallenge(seq)
	if !ok {
		t.Error("Challenge should exist before ClearChallenge()")
	}
	
	// Clear it
	m.ClearChallenge(seq)
	
	// Verify it's gone
	_, ok = m.GetChallenge(seq)
	if ok {
		t.Error("Challenge should not exist after ClearChallenge()")
	}
}

func TestClearAllChallenges(t *testing.T) {
	m := NewManager(DefaultChallengeTimeout)
	
	// Generate multiple challenges
	for i := uint8(1); i <= 5; i++ {
		_, err := m.GenerateChallenge(i, sa.RoleRemote)
		if err != nil {
			t.Fatalf("GenerateChallenge() error = %v", err)
		}
	}
	
	// Clear all
	m.ClearAllChallenges()
	
	// Verify all are gone
	for i := uint8(1); i <= 5; i++ {
		_, ok := m.GetChallenge(i)
		if ok {
			t.Errorf("Challenge %d should not exist after ClearAllChallenges()", i)
		}
	}
}

func TestStats(t *testing.T) {
	m := NewManager(DefaultChallengeTimeout)
	
	// Generate challenges
	for i := uint8(1); i <= 5; i++ {
		_, _ = m.GenerateChallenge(i, sa.RoleRemote)
	}
	
	// Use some challenges
	sessionKey := [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	for i := uint8(1); i <= 3; i++ {
		pending, _ := m.GetChallenge(i)
		dataToVerify := make([]byte, 0, sa.ChallengeSize)
		dataToVerify = append(dataToVerify, pending.Challenge.ChallengeData[:]...)
		mac, _ := sa.CalculateMAC(dataToVerify, sessionKey)
		_ = m.ValidateResponse(i, mac, sessionKey, nil)
	}
	
	pendingCount, usedCount, _ := m.Stats()
	
	if pendingCount != 2 {
		t.Errorf("Stats().pending = %d, want 2", pendingCount)
	}
	
	if usedCount != 3 {
		t.Errorf("Stats().used = %d, want 3", usedCount)
	}
}

func TestCompareChallenges(t *testing.T) {
	var a, b [sa.ChallengeSize]byte
	
	// Same challenges
	for i := range a {
		a[i] = byte(i)
		b[i] = byte(i)
	}
	
	if !CompareChallenges(&a, &b) {
		t.Error("CompareChallenges() returned false for identical challenges")
	}
	
	// Different challenges
	for i := range b {
		b[i] = byte(0xFF - i)
	}
	
	if CompareChallenges(&a, &b) {
		t.Error("CompareChallenges() returned true for different challenges")
	}
}
