package master

import "testing"

// AO-004 pre-PR corrective F1 — command-status octet 6 semantics.
//
// The pinned OpenDNP3 3.1.2 reference (cpp/lib/include/opendnp3/gen/CommandStatus.h,
// commit 26b4c01e4839bbbda8866655e086471c4917ee53) names command status 6
// HARDWARE_ERROR; BLOCKED is a distinct code 15. These tests pin the corrected
// naming and confirm the wire value 6 is a non-success, fail-closed outcome.

// TestCommandStatusHardwareErrorWireValue pins numeric code 6 to the canonical
// HardwareError name while retaining the deprecated Blocked alias at the same
// wire value.
func TestCommandStatusHardwareErrorWireValue(t *testing.T) {
	if CommandStatusHardwareError != 6 {
		t.Fatalf("CommandStatusHardwareError = %d, want 6", CommandStatusHardwareError)
	}
	if CommandStatusBlocked != CommandStatusHardwareError {
		t.Fatalf("deprecated alias CommandStatusBlocked = %d, want %d (HardwareError)", CommandStatusBlocked, CommandStatusHardwareError)
	}
	if CommandStatusHardwareError == CommandStatusSuccess {
		t.Fatalf("HardwareError must not alias Success")
	}
}

// TestParseCommandStatusHardwareError confirms a response carrying status octet
// 0x06 parses to HardwareError (never success).
func TestParseCommandStatusHardwareError(t *testing.T) {
	// G12V1 header + index + CROB fields with status byte 0x06.
	data := []byte{0x0C, 0x01, 0x00, 0x01, 0x34, 0x12, 0x07, 0x01,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x06}
	got := parseCommandStatus(data)
	if got != CommandStatusHardwareError {
		t.Fatalf("status byte 0x06 -> %d, want %d (HardwareError)", got, CommandStatusHardwareError)
	}
	if got == CommandStatusSuccess {
		t.Fatalf("status byte 0x06 mapped to success")
	}
}

// TestOperateWithStatusHardwareError drives the master against a canned
// transport returning status 0x06 and asserts the real, non-success status
// surfaces (HardwareError), never success.
func TestOperateWithStatusHardwareError(t *testing.T) {
	m := NewMaster(DefaultConfig())
	m.SetTransport(&statusEchoTransport{commandStatus: 0x06})
	m.SetState(StateInitialized)

	status, err := m.OperateWithStatus(1, false, 12, 1, 0x1234, uint8(CROBCodeLatchOn))
	if err != nil {
		t.Fatalf("OperateWithStatus returned error: %v", err)
	}
	if status != CommandStatusHardwareError {
		t.Fatalf("status = %d, want %d (HardwareError)", status, CommandStatusHardwareError)
	}
	if status == CommandStatusSuccess {
		t.Fatalf("wire status 6 surfaced as success")
	}
}
