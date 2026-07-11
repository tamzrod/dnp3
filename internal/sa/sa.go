// Package sa implements DNP3 Secure Authentication (IEEE 1815-2012).
//
// Secure Authentication provides cryptographic authentication to prevent
// unauthorized control operations and data manipulation.
//
// Key Components:
//   - Challenge-Response mechanism for authentication
//   - Key management (master keys, session keys)
//   - Session state tracking
//   - MAC (Message Authentication Code) verification
//
// Reference: IEEE 1815-2012 Section 8
package sa

import (
	"crypto/aes"
	"crypto/rand"
	"fmt"
)

// AES-CMAC constants
var zeroBlock [16]byte
var rb = [16]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x87}

// leftShift performs left shift on a 128-bit block
func leftShift(b []byte) {
	c := uint8(0)
	for i := 15; i >= 0; i-- {
		newC := b[i] >> 7
		b[i] = (b[i] << 1) | c
		c = newC
	}
}

// generateSubkeys generates the subkeys K1 and K2 for AES-CMAC
func generateSubkeys(key []byte) ([16]byte, [16]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return [16]byte{}, [16]byte{}, err
	}

	var L [16]byte
	block.Encrypt(L[:], zeroBlock[:])

	var K1 [16]byte
	K1 = L
	leftShift(K1[:])
	if L[0]&0x80 != 0 {
		for i := 0; i < 16; i++ {
			K1[i] ^= rb[i]
		}
	}

	var K2 [16]byte
	K2 = K1
	leftShift(K2[:])
	if K1[0]&0x80 != 0 {
		for i := 0; i < 16; i++ {
			K2[i] ^= rb[i]
		}
	}

	return K1, K2, nil
}

// aesCMAC calculates AES-CMAC
func aesCMAC(key []byte, message []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	K1, K2, err := generateSubkeys(key)
	if err != nil {
		return nil, err
	}

	messageLen := len(message)
	
	var X [16]byte

	// Process all complete blocks
	for i := 0; i <= messageLen-16; i += 16 {
		var Y [16]byte
		for j := 0; j < 16; j++ {
			Y[j] = X[j] ^ message[i+j]
		}
		block.Encrypt(X[:], Y[:])
	}

	// Handle the last block
	if messageLen%16 == 0 && messageLen > 0 {
		// Message length is a multiple of block size, XOR with K1
		var Y [16]byte
		for j := 0; j < 16; j++ {
			Y[j] = X[j] ^ message[messageLen-16+j] ^ K1[j]
		}
		block.Encrypt(X[:], Y[:])
	} else {
		// Need to pad: add 0x80 followed by zeros
		var Y [16]byte
		offset := messageLen % 16
		for j := 0; j < 16; j++ {
			if j < offset {
				Y[j] = X[j] ^ message[messageLen-offset+j]
			} else if j == offset {
				Y[j] = X[j] ^ 0x80
			} else {
				Y[j] = X[j]
			}
		}
		// XOR with K2
		for j := 0; j < 16; j++ {
			Y[j] ^= K2[j]
		}
		block.Encrypt(X[:], Y[:])
	}

	return X[:], nil
}

// Security levels as defined in IEEE 1815-2012
const (
	SecurityLevelNone     = 0 // No authentication
	SecurityLevelChallenge = 1 // Challenge-response only
	SecurityLevelMAC      = 2  // Per-message authentication
)

// Maximum challenge size (128 bits)
const ChallengeSize = 16

// Maximum MAC size (128 bits for AES-CMAC)
const MACSize = 16

// Authentication error codes
var (
	ErrInvalidMAC          = fmt.Errorf("invalid MAC")
	ErrChallengeExpired    = fmt.Errorf("challenge expired")
	ErrInvalidChallenge    = fmt.Errorf("invalid challenge")
	ErrSessionNotActive    = fmt.Errorf("session not active")
	ErrKeyNotFound        = fmt.Errorf("key not found")
	ErrInvalidRole        = fmt.Errorf("invalid role")
	ErrAuthenticationFailed = fmt.Errorf("authentication failed")
)

// Role represents a security role with specific privileges
type Role uint8

const (
	RoleRemote Role = iota // Basic read operations
	RoleLevel1             // Non-critical controls
	RoleLevel2             // Critical controls
	RoleManager            // Configuration, key management
)

// String returns the role name
func (r Role) String() string {
	switch r {
	case RoleRemote:
		return "Remote"
	case RoleLevel1:
		return "Level1"
	case RoleLevel2:
		return "Level2"
	case RoleManager:
		return "Manager"
	default:
		return "Unknown"
	}
}

// RoleFromUint8 converts a uint8 to Role
func RoleFromUint8(v uint8) (Role, error) {
	switch v {
	case 0:
		return RoleRemote, nil
	case 1:
		return RoleLevel1, nil
	case 2:
		return RoleLevel2, nil
	case 3:
		return RoleManager, nil
	default:
		return 0, ErrInvalidRole
	}
}

// Challenge represents an authentication challenge
type Challenge struct {
	ChallengeData [ChallengeSize]byte // 128-bit random challenge
	Seq          uint8               // Challenge sequence number
	Role         Role                // Role being authenticated
}

// NewChallenge generates a new random challenge
func NewChallenge(seq uint8, role Role) (*Challenge, error) {
	c := &Challenge{
		Seq:  seq,
		Role: role,
	}
	if _, err := rand.Read(c.ChallengeData[:]); err != nil {
		return nil, fmt.Errorf("failed to generate challenge: %w", err)
	}
	return c, nil
}

// ChallengeFromBytes parses a challenge from bytes
func ChallengeFromBytes(data []byte) (*Challenge, error) {
	if len(data) < ChallengeSize+2 {
		return nil, fmt.Errorf("challenge data too short: %d bytes", len(data))
	}
	c := &Challenge{}
	copy(c.ChallengeData[:], data[:ChallengeSize])
	c.Seq = data[ChallengeSize]
	c.Role = Role(data[ChallengeSize+1])
	return c, nil
}

// Bytes returns the challenge as bytes
func (c *Challenge) Bytes() []byte {
	result := make([]byte, ChallengeSize+2)
	copy(result[:ChallengeSize], c.ChallengeData[:])
	result[ChallengeSize] = c.Seq
	result[ChallengeSize+1] = uint8(c.Role)
	return result
}

// AuthenticationRequest represents an AUTHENTICATE request
type AuthRequest struct {
	Seq       uint8   // Authentication sequence number
	UserNumber uint8   // User number
	MAC       [MACSize]byte // Message Authentication Code
}

// NewAuthRequest creates a new authentication request
func NewAuthRequest(seq, userNumber uint8, mac [MACSize]byte) *AuthRequest {
	return &AuthRequest{
		Seq:        seq,
		UserNumber: userNumber,
		MAC:        mac,
	}
}

// AuthenticationConfirm represents an AUTHENTICATE_CONFIRM
type AuthConfirm struct {
	Seq       uint8   // Authentication sequence number
	UserNumber uint8   // User number
	MAC       [MACSize]byte // Message Authentication Code
}

// NewAuthConfirm creates a new authentication confirm
func NewAuthConfirm(seq, userNumber uint8, mac [MACSize]byte) *AuthConfirm {
	return &AuthConfirm{
		Seq:        seq,
		UserNumber: userNumber,
		MAC:        mac,
	}
}

// CalculateMAC calculates AES-CMAC for the given data and key
func CalculateMAC(data []byte, key [16]byte) ([16]byte, error) {
	mac, err := aesCMAC(key[:], data)
	if err != nil {
		return [16]byte{}, fmt.Errorf("failed to calculate CMAC: %w", err)
	}
	
	var result [16]byte
	copy(result[:], mac)
	return result, nil
}

// VerifyMAC verifies that the provided MAC matches the calculated MAC
func VerifyMAC(data []byte, key [16]byte, expectedMAC [16]byte) bool {
	calculatedMAC, err := CalculateMAC(data, key)
	if err != nil {
		return false
	}
	
	// Constant-time comparison to prevent timing attacks
	var diff byte
	for i := 0; i < 16; i++ {
		diff |= calculatedMAC[i] ^ expectedMAC[i]
	}
	return diff == 0
}
