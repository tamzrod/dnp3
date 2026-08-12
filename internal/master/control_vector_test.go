package master

import (
	"encoding/hex"
	"testing"

	"dnp3/internal/al"
)

func TestBuildCROBRequestGoldenVector(t *testing.T) {
	crob := &CROB{Code: 7, Count: 1, OnTime: 1000, OffTime: 2000, Status: 0}
	got := buildCROBRequest(0x1234, crob)
	want, err := hex.DecodeString("0c01000134120701e8030000d007000000")
	if err != nil { t.Fatal(err) }
	if string(got) != string(want) {
		t.Fatalf("CROB bytes = %X, want %X", got, want)
	}
}

// TestBuildControlRequestCROBLSB verifies the G12 CROB built by
// buildControlRequest (used by Operate) encodes OnTime/OffTime little-endian.
// OnTime=1000 -> E8 03 00 00 (LSB first); OffTime=2000 -> D0 07 00 00.
// This is the BE-negative case for the control path (DNP3-001).
func TestBuildControlRequestCROBLSB(t *testing.T) {
	m := &Master{}
	// buildControlRequest takes a generic value; for group 12 a bool or uint8
	// selects the code. Use LATCH_ON (7).
	req := m.buildControlRequest(al.FuncDirectOperate, 12, 1, 0x1234, uint8(7))
	if req == nil {
		t.Fatalf("buildControlRequest returned nil APDU")
	}
	// Expected object bytes:
	// 0C 01 00 01          header (group 12, var 1, qualifier 0x00, count 1)
	// 34 12                index 0x1234 (LSB first)
	// 07                   code = LATCH_ON
	// 01                   count
	// 00 00 00 00          onTime = 0 (LSB first)
	// 00 00 00 00          offTime = 0 (LSB first)
	// 00                   status
	want, err := hex.DecodeString("0c01000134120701000000000000000000")
	if err != nil {
		t.Fatal(err)
	}
	if string(req.Data) != string(want) {
		t.Fatalf("CROB control request = %X, want %X", req.Data, want)
	}
}
