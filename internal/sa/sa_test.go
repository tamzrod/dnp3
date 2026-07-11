package sa

import (
	"testing"
)

func TestRoleString(t *testing.T) {
	tests := []struct {
		role     Role
		expected string
	}{
		{RoleRemote, "Remote"},
		{RoleLevel1, "Level1"},
		{RoleLevel2, "Level2"},
		{RoleManager, "Manager"},
		{Role(99), "Unknown"},
	}
	
	for _, tt := range tests {
		if got := tt.role.String(); got != tt.expected {
			t.Errorf("Role(%d).String() = %q, want %q", tt.role, got, tt.expected)
		}
	}
}

func TestRoleFromUint8(t *testing.T) {
	tests := []struct {
		value    uint8
		expected Role
		hasError bool
	}{
		{0, RoleRemote, false},
		{1, RoleLevel1, false},
		{2, RoleLevel2, false},
		{3, RoleManager, false},
		{4, 0, true},
		{255, 0, true},
	}
	
	for _, tt := range tests {
		role, err := RoleFromUint8(tt.value)
		if tt.hasError {
			if err == nil {
				t.Errorf("RoleFromUint8(%d) expected error, got nil", tt.value)
			}
		} else {
			if err != nil {
				t.Errorf("RoleFromUint8(%d) unexpected error: %v", tt.value, err)
			}
			if role != tt.expected {
				t.Errorf("RoleFromUint8(%d) = %v, want %v", tt.value, role, tt.expected)
			}
		}
	}
}

func TestNewChallenge(t *testing.T) {
	seq := uint8(5)
	role := RoleLevel2
	
	challenge, err := NewChallenge(seq, role)
	if err != nil {
		t.Fatalf("NewChallenge() error = %v", err)
	}
	
	if challenge.Seq != seq {
		t.Errorf("Challenge.Seq = %d, want %d", challenge.Seq, seq)
	}
	
	if challenge.Role != role {
		t.Errorf("Challenge.Role = %v, want %v", challenge.Role, role)
	}
	
	// Check that challenge data is non-zero (random)
	allZero := true
	for _, b := range challenge.ChallengeData {
		if b != 0 {
			allZero = false
			break
		}
	}
	if allZero {
		t.Error("Challenge.ChallengeData is all zeros (not random)")
	}
}

func TestChallengeBytes(t *testing.T) {
	seq := uint8(10)
	role := RoleManager
	
	challenge, err := NewChallenge(seq, role)
	if err != nil {
		t.Fatalf("NewChallenge() error = %v", err)
	}
	
	bytes := challenge.Bytes()
	if len(bytes) != ChallengeSize+2 {
		t.Errorf("Bytes() len = %d, want %d", len(bytes), ChallengeSize+2)
	}
	
	// Parse back
	parsed, err := ChallengeFromBytes(bytes)
	if err != nil {
		t.Fatalf("ChallengeFromBytes() error = %v", err)
	}
	
	if parsed.Seq != challenge.Seq {
		t.Errorf("Parsed.Seq = %d, want %d", parsed.Seq, challenge.Seq)
	}
	
	if parsed.Role != challenge.Role {
		t.Errorf("Parsed.Role = %v, want %v", parsed.Role, challenge.Role)
	}
	
	// Compare challenge data
	for i := 0; i < ChallengeSize; i++ {
		if parsed.ChallengeData[i] != challenge.ChallengeData[i] {
			t.Errorf("Parsed.ChallengeData[%d] = 0x%02X, want 0x%02X", 
				i, parsed.ChallengeData[i], challenge.ChallengeData[i])
		}
	}
}

func TestChallengeFromBytesTooShort(t *testing.T) {
	_, err := ChallengeFromBytes([]byte{0x01, 0x02})
	if err == nil {
		t.Error("ChallengeFromBytes() expected error for short input")
	}
}

func TestNewAuthRequest(t *testing.T) {
	seq := uint8(7)
	userNumber := uint8(3)
	var mac [MACSize]byte
	for i := range mac {
		mac[i] = byte(i)
	}
	
	req := NewAuthRequest(seq, userNumber, mac)
	
	if req.Seq != seq {
		t.Errorf("AuthRequest.Seq = %d, want %d", req.Seq, seq)
	}
	
	if req.UserNumber != userNumber {
		t.Errorf("AuthRequest.UserNumber = %d, want %d", req.UserNumber, userNumber)
	}
	
	for i := 0; i < MACSize; i++ {
		if req.MAC[i] != mac[i] {
			t.Errorf("AuthRequest.MAC[%d] = 0x%02X, want 0x%02X", i, req.MAC[i], mac[i])
		}
	}
}

func TestNewAuthConfirm(t *testing.T) {
	seq := uint8(15)
	userNumber := uint8(5)
	var mac [MACSize]byte
	for i := range mac {
		mac[i] = byte(0xFF - i)
	}
	
	conf := NewAuthConfirm(seq, userNumber, mac)
	
	if conf.Seq != seq {
		t.Errorf("AuthConfirm.Seq = %d, want %d", conf.Seq, seq)
	}
	
	if conf.UserNumber != userNumber {
		t.Errorf("AuthConfirm.UserNumber = %d, want %d", conf.UserNumber, userNumber)
	}
}

func TestCalculateMAC(t *testing.T) {
	var key [16]byte
	for i := range key {
		key[i] = byte(i)
	}
	
	data := []byte("test data for MAC calculation")
	
	mac1, err := CalculateMAC(data, key)
	if err != nil {
		t.Fatalf("CalculateMAC() error = %v", err)
	}
	
	// Same input should produce same output
	mac2, err := CalculateMAC(data, key)
	if err != nil {
		t.Fatalf("CalculateMAC() second call error = %v", err)
	}
	
	for i := 0; i < MACSize; i++ {
		if mac1[i] != mac2[i] {
			t.Errorf("MAC calculation not deterministic: mac1[%d] = 0x%02X, mac2[%d] = 0x%02X",
				i, mac1[i], i, mac2[i])
		}
	}
	
	// Different key should produce different MAC
	var differentKey [16]byte
	for i := range differentKey {
		differentKey[i] = byte(0xFF - i)
	}
	
	mac3, err := CalculateMAC(data, differentKey)
	if err != nil {
		t.Fatalf("CalculateMAC() with different key error = %v", err)
	}
	
	different := false
	for i := 0; i < MACSize; i++ {
		if mac1[i] != mac3[i] {
			different = true
			break
		}
	}
	if !different {
		t.Error("Different keys produced same MAC")
	}
}

func TestVerifyMAC(t *testing.T) {
	var key [16]byte
	for i := range key {
		key[i] = byte(i)
	}
	
	data := []byte("test data for MAC verification")
	
	// Calculate MAC
	mac, err := CalculateMAC(data, key)
	if err != nil {
		t.Fatalf("CalculateMAC() error = %v", err)
	}
	
	// Verify should succeed with correct MAC
	if !VerifyMAC(data, key, mac) {
		t.Error("VerifyMAC() returned false for correct MAC")
	}
	
	// Verify should fail with incorrect MAC
	var wrongMAC [16]byte
	for i := range wrongMAC {
		wrongMAC[i] = mac[i] ^ 0xFF
	}
	
	if VerifyMAC(data, key, wrongMAC) {
		t.Error("VerifyMAC() returned true for incorrect MAC")
	}
	
	// Verify should fail with wrong key
	var wrongKey [16]byte
	for i := range wrongKey {
		wrongKey[i] = byte(0xFF - i)
	}
	
	if VerifyMAC(data, wrongKey, mac) {
		t.Error("VerifyMAC() returned true for wrong key")
	}
	
	// Verify should fail with wrong data
	wrongData := []byte("wrong data")
	if VerifyMAC(wrongData, key, mac) {
		t.Error("VerifyMAC() returned true for wrong data")
	}
}

func TestVerifyMACConstantTime(t *testing.T) {
	var key [16]byte
	data := []byte("constant time test")
	
	mac, _ := CalculateMAC(data, key)
	
	// This test just ensures VerifyMAC doesn't panic on any input
	_ = VerifyMAC(data, key, mac)
	_ = VerifyMAC(nil, key, mac)
	_ = VerifyMAC(data, key, [16]byte{})
}
