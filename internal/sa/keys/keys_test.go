package keys

import (
	"testing"

	"dnp3/internal/sa"
)

func TestNewKeyTable(t *testing.T) {
	kt := NewKeyTable()
	if kt == nil {
		t.Fatal("NewKeyTable() returned nil")
	}
	
	if kt.UserCount() != 0 {
		t.Errorf("NewKeyTable().UserCount() = %d, want 0", kt.UserCount())
	}
}

func TestAddUser(t *testing.T) {
	kt := NewKeyTable()
	
	userNumber := uint8(5)
	username := "testuser"
	var key [16]byte
	for i := range key {
		key[i] = byte(i)
	}
	role := sa.RoleLevel2
	challengeDuration := uint16(30)
	
	err := kt.AddUser(userNumber, username, key, role, challengeDuration)
	if err != nil {
		t.Fatalf("AddUser() error = %v", err)
	}
	
	if kt.UserCount() != 1 {
		t.Errorf("UserCount() = %d, want 1", kt.UserCount())
	}
	
	entry, ok := kt.GetUser(userNumber)
	if !ok {
		t.Fatal("GetUser() returned false for just-added user")
	}
	
	if entry.Username != username {
		t.Errorf("Entry.Username = %q, want %q", entry.Username, username)
	}
	
	if entry.Role != role {
		t.Errorf("Entry.Role = %v, want %v", entry.Role, role)
	}
}

func TestAddUserInvalidNumber(t *testing.T) {
	kt := NewKeyTable()
	
	var key [16]byte
	
	// User number 0 is invalid
	err := kt.AddUser(0, "user0", key, sa.RoleRemote, 30)
	if err == nil {
		t.Error("AddUser(0) expected error")
	}
	
	// User number > 63 is invalid
	err = kt.AddUser(64, "user64", key, sa.RoleRemote, 30)
	if err == nil {
		t.Error("AddUser(64) expected error")
	}
}

func TestAddUserDuplicate(t *testing.T) {
	kt := NewKeyTable()
	
	var key [16]byte
	
	err := kt.AddUser(5, "user1", key, sa.RoleRemote, 30)
	if err != nil {
		t.Fatalf("First AddUser() error = %v", err)
	}
	
	err = kt.AddUser(5, "user2", key, sa.RoleLevel2, 60)
	if err == nil {
		t.Error("Duplicate AddUser() expected error")
	}
}

func TestGetUserNotFound(t *testing.T) {
	kt := NewKeyTable()
	
	_, ok := kt.GetUser(99)
	if ok {
		t.Error("GetUser(99) returned true for non-existent user")
	}
}

func TestRemoveUser(t *testing.T) {
	kt := NewKeyTable()
	
	var key [16]byte
	kt.AddUser(10, "testuser", key, sa.RoleRemote, 30)
	
	// Remove existing user
	ok := kt.RemoveUser(10)
	if !ok {
		t.Error("RemoveUser(10) returned false for existing user")
	}
	
	if kt.UserCount() != 0 {
		t.Errorf("UserCount() = %d, want 0 after RemoveUser()", kt.UserCount())
	}
	
	// Remove non-existent user
	ok = kt.RemoveUser(99)
	if ok {
		t.Error("RemoveUser(99) returned true for non-existent user")
	}
}

func TestUpdateKey(t *testing.T) {
	kt := NewKeyTable()
	
	var oldKey, newKey [16]byte
	for i := range oldKey {
		oldKey[i] = byte(i)
		newKey[i] = byte(0xFF - i)
	}
	
	kt.AddUser(7, "testuser", oldKey, sa.RoleLevel1, 30)
	
	// Update key
	err := kt.UpdateKey(7, newKey)
	if err != nil {
		t.Fatalf("UpdateKey() error = %v", err)
	}
	
	entry, _ := kt.GetUser(7)
	for i := 0; i < 16; i++ {
		if entry.Value[i] != newKey[i] {
			t.Errorf("Entry.Value[%d] = 0x%02X, want 0x%02X", i, entry.Value[i], newKey[i])
		}
	}
	
	if entry.KeyVersion != 2 {
		t.Errorf("Entry.KeyVersion = %d, want 2", entry.KeyVersion)
	}
}

func TestUpdateKeyNotFound(t *testing.T) {
	kt := NewKeyTable()
	
	var key [16]byte
	err := kt.UpdateKey(99, key)
	if err != sa.ErrKeyNotFound {
		t.Errorf("UpdateKey() error = %v, want %v", err, sa.ErrKeyNotFound)
	}
}

func TestIncrementFailures(t *testing.T) {
	kt := NewKeyTable()
	
	var key [16]byte
	kt.AddUser(8, "testuser", key, sa.RoleLevel2, 30)
	
	// Fail authentication 3 times
	for i := 0; i < 3; i++ {
		err := kt.IncrementFailures(8)
		if err != nil {
			t.Fatalf("IncrementFailures() error = %v", err)
		}
	}
	
	entry, _ := kt.GetUser(8)
	if entry.AuthFailures != 3 {
		t.Errorf("Entry.AuthFailures = %d, want 3", entry.AuthFailures)
	}
	
	if !entry.Locked {
		t.Error("Entry.Locked = false, want true after max failures")
	}
}

func TestResetFailures(t *testing.T) {
	kt := NewKeyTable()
	
	var key [16]byte
	kt.AddUser(9, "testuser", key, sa.RoleLevel2, 30)
	
	// Fail authentication
	kt.IncrementFailures(9)
	kt.IncrementFailures(9)
	
	// Reset
	err := kt.ResetFailures(9)
	if err != nil {
		t.Fatalf("ResetFailures() error = %v", err)
	}
	
	entry, _ := kt.GetUser(9)
	if entry.AuthFailures != 0 {
		t.Errorf("Entry.AuthFailures = %d, want 0 after ResetFailures()", entry.AuthFailures)
	}
	
	if entry.Locked {
		t.Error("Entry.Locked = true, want false after ResetFailures()")
	}
}

func TestIsLocked(t *testing.T) {
	kt := NewKeyTable()
	
	var key [16]byte
	kt.AddUser(10, "testuser", key, sa.RoleLevel2, 30)
	
	// Should not be locked initially
	if kt.IsLocked(10) {
		t.Error("IsLocked(10) = true, want false")
	}
	
	// Lock user
	kt.IncrementFailures(10)
	kt.IncrementFailures(10)
	kt.IncrementFailures(10)
	
	// Should be locked now
	if !kt.IsLocked(10) {
		t.Error("IsLocked(10) = false, want true after max failures")
	}
	
	// Non-existent user should be treated as locked
	if !kt.IsLocked(99) {
		t.Error("IsLocked(99) = false, want true for non-existent user")
	}
}

func TestGenerateRandomKey(t *testing.T) {
	key1, err := GenerateRandomKey()
	if err != nil {
		t.Fatalf("GenerateRandomKey() error = %v", err)
	}
	
	key2, err := GenerateRandomKey()
	if err != nil {
		t.Fatalf("GenerateRandomKey() second call error = %v", err)
	}
	
	// Keys should be different
	same := true
	for i := 0; i < 16; i++ {
		if key1[i] != key2[i] {
			same = false
			break
		}
	}
	if same {
		t.Error("Two calls to GenerateRandomKey() produced same key")
	}
}

func TestDeriveSessionKey(t *testing.T) {
	var masterKey [16]byte
	for i := range masterKey {
		masterKey[i] = byte(i)
	}
	
	var challenge [sa.ChallengeSize]byte
	for i := range challenge {
		challenge[i] = byte(0xFF - i)
	}
	
	userNumber := uint8(5)
	
	sessionKey, err := DeriveSessionKey(masterKey, challenge, userNumber)
	if err != nil {
		t.Fatalf("DeriveSessionKey() error = %v", err)
	}
	
	// Session key should be non-zero
	allZero := true
	for _, b := range sessionKey {
		if b != 0 {
			allZero = false
			break
		}
	}
	if allZero {
		t.Error("DeriveSessionKey() returned all-zero key")
	}
	
	// Different inputs should produce different keys
	sessionKey2, _ := DeriveSessionKey(masterKey, challenge, userNumber+1)
	
	different := false
	for i := 0; i < 16; i++ {
		if sessionKey[i] != sessionKey2[i] {
			different = true
			break
		}
	}
	if !different {
		t.Error("Different user numbers produced same session key")
	}
}

func TestDeriveMACKey(t *testing.T) {
	var sessionKey [16]byte
	for i := range sessionKey {
		sessionKey[i] = byte(i)
	}
	
	macKey, err := DeriveMACKey(sessionKey)
	if err != nil {
		t.Fatalf("DeriveMACKey() error = %v", err)
	}
	
	// MAC key should be non-zero
	allZero := true
	for _, b := range macKey {
		if b != 0 {
			allZero = false
			break
		}
	}
	if allZero {
		t.Error("DeriveMACKey() returned all-zero key")
	}
}

func TestListUsers(t *testing.T) {
	kt := NewKeyTable()
	
	var key [16]byte
	kt.AddUser(5, "user5", key, sa.RoleRemote, 30)
	kt.AddUser(10, "user10", key, sa.RoleLevel1, 30)
	kt.AddUser(15, "user15", key, sa.RoleLevel2, 30)
	
	users := kt.ListUsers()
	
	if len(users) != 3 {
		t.Errorf("ListUsers() len = %d, want 3", len(users))
	}
	
	// Check all expected users are present
	expected := map[uint8]bool{5: true, 10: true, 15: true}
	for _, u := range users {
		if !expected[u] {
			t.Errorf("Unexpected user %d in ListUsers()", u)
		}
	}
}

func TestSetMasterKeyGlobal(t *testing.T) {
	kt := NewKeyTable()
	
	var masterKey [16]byte
	for i := range masterKey {
		masterKey[i] = byte(0xAA)
	}
	
	kt.SetMasterKeyGlobal(masterKey)
	
	retrieved, ok := kt.GetMasterKey()
	if !ok {
		t.Fatal("GetMasterKey() returned false after SetMasterKeyGlobal()")
	}
	
	for i := 0; i < 16; i++ {
		if retrieved[i] != masterKey[i] {
			t.Errorf("Retrieved master key[%d] = 0x%02X, want 0x%02X", i, retrieved[i], masterKey[i])
		}
	}
}
