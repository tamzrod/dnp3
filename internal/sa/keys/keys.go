// Package keys implements DNP3 Secure Authentication key management.
package keys

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"sync"

	"dnp3/internal/sa"
)

// Key types
const (
	KeyTypeMaster = iota // Master key used to derive session keys
	KeyTypeSession       // Session key used for authentication
	KeyTypeUpdate        // Update key used to change other keys
)

// Key represents an encryption key with metadata
type Key struct {
	Type       int           // Key type (master, session, update)
	Role       sa.Role       // Role this key authorizes
	Value      [16]byte      // 128-bit AES key
	KeyVersion uint16        // Key version for tracking
	UpdatedAt  uint32        // Last update timestamp
}

// KeyEntry represents an entry in the key table
type KeyEntry struct {
	Key
	UserNumber  uint8    // User number (1-63)
	Username    string   // Username for audit
	ChallengeDuration uint16 // Max challenge age in seconds
	MaxAuthFailures uint8  // Max authentication failures before lockout
	AuthFailures     uint8  // Current authentication failure count
	Locked           bool   // Whether the user is locked
}

// KeyTable manages encryption keys for secure authentication
type KeyTable struct {
	mu      sync.RWMutex
	entries map[uint8]*KeyEntry // User number -> Key entry
	keys    map[int][16]byte     // Type -> Key value (for master/update keys)
}

// NewKeyTable creates a new key table
func NewKeyTable() *KeyTable {
	return &KeyTable{
		entries: make(map[uint8]*KeyEntry),
		keys:    make(map[int][16]byte),
	}
}

// AddUser adds a new user to the key table
func (kt *KeyTable) AddUser(userNumber uint8, username string, key [16]byte, role sa.Role, challengeDuration uint16) error {
	kt.mu.Lock()
	defer kt.mu.Unlock()
	
	if userNumber == 0 || userNumber > 63 {
		return fmt.Errorf("invalid user number: must be 1-63")
	}
	
	if _, exists := kt.entries[userNumber]; exists {
		return fmt.Errorf("user %d already exists", userNumber)
	}
	
	entry := &KeyEntry{
		Key: Key{
			Type:        KeyTypeSession,
			Role:        role,
			Value:       key,
			KeyVersion:  1,
		},
		UserNumber:        userNumber,
		Username:         username,
		ChallengeDuration: challengeDuration,
		MaxAuthFailures:  3,
	}
	
	kt.entries[userNumber] = entry
	return nil
}

// GetUser retrieves a user from the key table
func (kt *KeyTable) GetUser(userNumber uint8) (*KeyEntry, bool) {
	kt.mu.RLock()
	defer kt.mu.RUnlock()
	
	entry, ok := kt.entries[userNumber]
	return entry, ok
}

// RemoveUser removes a user from the key table
func (kt *KeyTable) RemoveUser(userNumber uint8) bool {
	kt.mu.Lock()
	defer kt.mu.Unlock()
	
	if _, exists := kt.entries[userNumber]; !exists {
		return false
	}
	
	delete(kt.entries, userNumber)
	return true
}

// UpdateKey updates a user's session key
func (kt *KeyTable) UpdateKey(userNumber uint8, newKey [16]byte) error {
	kt.mu.Lock()
	defer kt.mu.Unlock()
	
	entry, ok := kt.entries[userNumber]
	if !ok {
		return sa.ErrKeyNotFound
	}
	
	entry.Value = newKey
	entry.KeyVersion++
	
	return nil
}

// SetMasterKey sets the master key for a user
func (kt *KeyTable) SetMasterKey(userNumber uint8, masterKey [16]byte) error {
	kt.mu.Lock()
	defer kt.mu.Unlock()
	
	entry, ok := kt.entries[userNumber]
	if !ok {
		return sa.ErrKeyNotFound
	}
	
	entry.Key.Type = KeyTypeMaster
	entry.Value = masterKey
	
	return nil
}

// GetMasterKey retrieves the master key for key derivation
func (kt *KeyTable) GetMasterKey() ([16]byte, bool) {
	kt.mu.RLock()
	defer kt.mu.RUnlock()
	
	key, ok := kt.keys[KeyTypeMaster]
	return key, ok
}

// SetMasterKeyGlobal sets the global master key
func (kt *KeyTable) SetMasterKeyGlobal(key [16]byte) {
	kt.mu.Lock()
	defer kt.mu.Unlock()
	
	kt.keys[KeyTypeMaster] = key
}

// IncrementFailures increments the authentication failure count
func (kt *KeyTable) IncrementFailures(userNumber uint8) error {
	kt.mu.Lock()
	defer kt.mu.Unlock()
	
	entry, ok := kt.entries[userNumber]
	if !ok {
		return sa.ErrKeyNotFound
	}
	
	entry.AuthFailures++
	if entry.AuthFailures >= entry.MaxAuthFailures {
		entry.Locked = true
	}
	
	return nil
}

// ResetFailures resets the authentication failure count
func (kt *KeyTable) ResetFailures(userNumber uint8) error {
	kt.mu.Lock()
	defer kt.mu.Unlock()
	
	entry, ok := kt.entries[userNumber]
	if !ok {
		return sa.ErrKeyNotFound
	}
	
	entry.AuthFailures = 0
	entry.Locked = false
	
	return nil
}

// IsLocked checks if a user is locked
func (kt *KeyTable) IsLocked(userNumber uint8) bool {
	kt.mu.RLock()
	defer kt.mu.RUnlock()
	
	entry, ok := kt.entries[userNumber]
	if !ok {
		return true // Non-existent users are treated as locked
	}
	
	return entry.Locked
}

// DeriveSessionKey derives a session key from the master key and challenge
func DeriveSessionKey(masterKey [16]byte, challenge [sa.ChallengeSize]byte, userNumber uint8) ([16]byte, error) {
	// Derivation: AES-128 with challenge and user number
	block, err := aes.NewCipher(masterKey[:])
	if err != nil {
		return [16]byte{}, fmt.Errorf("failed to create cipher: %w", err)
	}
	
	// Use CTR-like mode for key derivation
	iv := make([]byte, aes.BlockSize)
	iv[0] = byte(userNumber)
	copy(iv[1:13], challenge[:12])
	
	encrypted := make([]byte, 16)
	cipher.NewCTR(block, iv).XORKeyStream(encrypted, encrypted)
	
	var sessionKey [16]byte
	copy(sessionKey[:], encrypted)
	return sessionKey, nil
}

// GenerateRandomKey generates a random 128-bit AES key
func GenerateRandomKey() ([16]byte, error) {
	var key [16]byte
	if _, err := rand.Read(key[:]); err != nil {
		return [16]byte{}, fmt.Errorf("failed to generate random key: %w", err)
	}
	return key, nil
}

// DeriveMACKey derives a MAC key from the session key
func DeriveMACKey(sessionKey [16]byte) ([16]byte, error) {
	// For AES-CMAC, we use a fixed derivation
	block, err := aes.NewCipher(sessionKey[:])
	if err != nil {
		return [16]byte{}, fmt.Errorf("failed to create cipher: %w", err)
	}
	
	// Use cipher's export function (not standard but works for derivation)
	var macKey [16]byte
	// Simple derivation: encrypt a fixed value
	plaintext := [16]byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	block.Encrypt(macKey[:], plaintext[:])
	
	return macKey, nil
}

// UserCount returns the number of users in the key table
func (kt *KeyTable) UserCount() int {
	kt.mu.RLock()
	defer kt.mu.RUnlock()
	
	return len(kt.entries)
}

// ListUsers returns a list of all user numbers
func (kt *KeyTable) ListUsers() []uint8 {
	kt.mu.RLock()
	defer kt.mu.RUnlock()
	
	users := make([]uint8, 0, len(kt.entries))
	for userNum := range kt.entries {
		users = append(users, userNum)
	}
	return users
}
