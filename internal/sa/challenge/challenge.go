// Package challenge implements DNP3 Secure Authentication challenge handling.
package challenge

import (
	"crypto/subtle"
	"fmt"
	"time"

	"dnp3/internal/sa"
)

// DefaultChallengeTimeout is the default challenge timeout (30 seconds)
const DefaultChallengeTimeout = 30 * time.Second

// MaxChallengeAge is the maximum age for a challenge before it's considered expired
const MaxChallengeAge = 60 * time.Second

// ChallengeError represents challenge-related errors
type ChallengeError struct {
	Code    string
	Message string
}

func (e *ChallengeError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

var (
	ErrChallengeExpired = &ChallengeError{"ERR_EXPIRED", "challenge has expired"}
	ErrChallengeUsed    = &ChallengeError{"ERR_USED", "challenge has already been used"}
	ErrChallengeNotFound = &ChallengeError{"ERR_NOT_FOUND", "challenge not found"}
)

// PendingChallenge represents a challenge that has been sent but not yet responded to
type PendingChallenge struct {
	Challenge    *sa.Challenge
	IssuedAt     time.Time
	ExpiresAt    time.Time
	Used         bool
}

// Manager handles challenge generation and validation
type Manager struct {
	challengeTimeout time.Duration
	challenges      map[uint8]*PendingChallenge
	maxChallenges    int
}

// NewManager creates a new challenge manager
func NewManager(timeout time.Duration) *Manager {
	return &Manager{
		challengeTimeout: timeout,
		challenges:      make(map[uint8]*PendingChallenge),
		maxChallenges:   256, // Max 256 pending challenges
	}
}

// GenerateChallenge creates a new challenge for the given role
func (m *Manager) GenerateChallenge(seq uint8, role sa.Role) (*sa.Challenge, error) {
	challenge, err := sa.NewChallenge(seq, role)
	if err != nil {
		return nil, fmt.Errorf("failed to generate challenge: %w", err)
	}
	
	// Store as pending challenge
	pending := &PendingChallenge{
		Challenge: challenge,
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(m.challengeTimeout),
		Used:      false,
	}
	
	// Clean up old challenges if needed
	m.cleanup()
	
	m.challenges[seq] = pending
	
	return challenge, nil
}

// ValidateResponse validates an authentication response against a pending challenge
func (m *Manager) ValidateResponse(seq uint8, mac [sa.MACSize]byte, sessionKey [16]byte, authData []byte) error {
	pending, ok := m.challenges[seq]
	if !ok {
		return ErrChallengeNotFound
	}
	
	// Check if challenge is expired
	if time.Now().After(pending.ExpiresAt) {
		delete(m.challenges, seq)
		return ErrChallengeExpired
	}
	
	// Check if challenge was already used (anti-replay)
	if pending.Used {
		return ErrChallengeUsed
	}
	
	// Mark as used
	pending.Used = true
	
	// Build data to verify: challenge bytes + auth data
	dataToVerify := make([]byte, 0, sa.ChallengeSize+len(authData))
	dataToVerify = append(dataToVerify, pending.Challenge.ChallengeData[:]...)
	dataToVerify = append(dataToVerify, authData...)
	
	// Verify MAC
	if !sa.VerifyMAC(dataToVerify, sessionKey, mac) {
		return sa.ErrInvalidMAC
	}
	
	return nil
}

// GetChallenge retrieves a pending challenge by sequence number
func (m *Manager) GetChallenge(seq uint8) (*PendingChallenge, bool) {
	pending, ok := m.challenges[seq]
	if !ok {
		return nil, false
	}
	
	// Check if expired
	if time.Now().After(pending.ExpiresAt) {
		delete(m.challenges, seq)
		return nil, false
	}
	
	return pending, true
}

// MarkUsed marks a challenge as used
func (m *Manager) MarkUsed(seq uint8) {
	if pending, ok := m.challenges[seq]; ok {
		pending.Used = true
	}
}

// ClearChallenge removes a challenge from the manager
func (m *Manager) ClearChallenge(seq uint8) {
	delete(m.challenges, seq)
}

// ClearAllChallenges removes all pending challenges
func (m *Manager) ClearAllChallenges() {
	m.challenges = make(map[uint8]*PendingChallenge)
}

// cleanup removes expired and used challenges
func (m *Manager) cleanup() {
	now := time.Now()
	for seq, pending := range m.challenges {
		if pending.Used || now.After(pending.ExpiresAt) {
			delete(m.challenges, seq)
		}
	}
}

// Stats returns statistics about pending challenges
func (m *Manager) Stats() (pending, used, expired int) {
	now := time.Now()
	for _, p := range m.challenges {
		if p.Used {
			used++
		} else if now.After(p.ExpiresAt) {
			expired++
		} else {
			pending++
		}
	}
	return
}

// CompareChallenges performs constant-time comparison of two challenges
func CompareChallenges(a, b *[sa.ChallengeSize]byte) bool {
	return subtle.ConstantTimeCompare(a[:], b[:]) == 1
}
