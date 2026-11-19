package master

import (
	"encoding/hex"
	"testing"

	"dnp3/internal/al"
)

// TestBuildControlRequestPulseCROBWire locks the wire bytes emitted for a
// complete G12V1 pulse CROB. The public NewPulseControl constructor encodes
// code=0x01, count=2, onTime=1000, offTime=500 into an 11-octet payload which
// buildControlRequest must preserve verbatim (LSB-first times).
func TestBuildControlRequestPulseCROBWire(t *testing.T) {
	payload := []byte{
		0x01, 0x02, // Pulse On, two cycles
		0xE8, 0x03, 0x00, 0x00, // OnTime 1000 ms LSB-first
		0xF4, 0x01, 0x00, 0x00, // OffTime 500 ms LSB-first
		0x00, // request status
	}
	m := &Master{}
	req := m.buildControlRequest(al.FuncDirectOperate, 12, 1, 0x0001, payload)
	if req == nil {
		t.Fatal("buildControlRequest rejected a valid pulse CROB payload")
	}
	want, err := hex.DecodeString("0c0128010001000102e8030000f401000000")
	if err != nil {
		t.Fatal(err)
	}
	if string(req.Data) != string(want) {
		t.Fatalf("pulse CROB request = %X, want %X", req.Data, want)
	}
}

func TestBuildControlRequestRejectsZeroCountPulse(t *testing.T) {
	payload := []byte{0x01, 0x00, 0xE8, 0x03, 0, 0, 0xF4, 0x01, 0, 0, 0}
	m := &Master{}
	if req := m.buildControlRequest(al.FuncDirectOperate, 12, 1, 0, payload); req != nil {
		t.Fatal("zero-cycle pulse CROB was encoded onto the wire")
	}
}
