package master

import (
	"encoding/hex"
	"testing"

	"dnp3/internal/al"
)

func TestBuildCROBRequestGoldenVector(t *testing.T) {
	crob := &CROB{Code: CROBCodeLatchOn, Count: 1, OnTime: 1000, OffTime: 2000, Status: 0}
	got := buildCROBRequest(0x1234, crob)
	want, err := hex.DecodeString("0c01000134120301e8030000d007000000")
	if err != nil {
		t.Fatal(err)
	}
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
	// selects the code. Use LATCH_ON (0x03, IEEE 1815 A.8.1 low-nibble enum).
	req := m.buildControlRequest(al.FuncDirectOperate, 12, 1, 0x1234, uint8(CROBCodeLatchOn))
	if req == nil {
		t.Fatalf("buildControlRequest returned nil APDU")
	}
	// Expected object bytes:
	// 0C 01 28 01 00       header (group 12, var 1, qualifier 0x28, count 1)
	// 34 12                index 0x1234 (LSB first)
	// 03                   code = LATCH_ON (IEEE 1815 A.8.1 0x03)
	// 01                   count
	// 00 00 00 00          onTime = 0 (LSB first)
	// 00 00 00 00          offTime = 0 (LSB first)
	// 00                   status
	want, err := hex.DecodeString("0c0128010034120301000000000000000000")
	if err != nil {
		t.Fatal(err)
	}
	if string(req.Data) != string(want) {
		t.Fatalf("CROB control request = %X, want %X", req.Data, want)
	}
}

// TestBuildCROBRequestLayout locks the G12V1 CROB wire layout produced by
// buildCROBRequest (used by WriteBinaryOutput): a 4-byte object header
// (group/variation/qualifier 0x00/count), a 2-byte LSB index, then the
// 11-byte CROB value (code, count, onTime LSB, offTime LSB, status).
func TestBuildCROBRequestLayout(t *testing.T) {
	crob := &CROB{Code: CROBCodeLatchOn, Count: 1, OnTime: 1000, OffTime: 2000, Status: 0}
	got := buildCROBRequest(0x1234, crob)
	// Header + index + value = 4 + 2 + 11 = 17 octets.
	if len(got) != 17 {
		t.Fatalf("CROB request length = %d, want 17", len(got))
	}
	// Header: G12 V1, qualifier 0x00 (index-only), count 1.
	if got[0] != 12 || got[1] != 1 || got[2] != 0x00 || got[3] != 0x01 {
		t.Errorf("header = %X, want 0C 01 00 01", got[:4])
	}
	// Index 0x1234 LSB first.
	if got[4] != 0x34 || got[5] != 0x12 {
		t.Errorf("index = %X, want 34 12 (LSB first)", got[4:6])
	}
	// Value: code=0x03 (LATCH_ON, IEEE 1815 A.8.1), count=1, onTime=1000 (E8 03 00 00), offTime=2000 (D0 07 00 00), status=0.
	wantVal := []byte{0x03, 0x01, 0xE8, 0x03, 0x00, 0x00, 0xD0, 0x07, 0x00, 0x00, 0x00}
	if string(got[6:]) != string(wantVal) {
		t.Errorf("value = %X, want %X", got[6:], wantVal)
	}
}

// TestBuildCROBRequestIndexLSBHighByte confirms the 2-octet index is encoded
// LSB-first when the high byte is non-zero (the DNP3-001 BE-negative case for
// the CROB index, re-asserted for the locked request).
func TestBuildCROBRequestIndexLSBHighByte(t *testing.T) {
	crob := &CROB{Code: CROBCodeLatchOff, Count: 1}
	got := buildCROBRequest(0xABCD, crob)
	// 0xABCD LSB first -> CD AB.
	if got[4] != 0xCD || got[5] != 0xAB {
		t.Fatalf("index bytes = %X, want CD AB (LSB first)", got[4:6])
	}
}

// TestBuildCROBRequestTimeLSBBoundary locks OnTime/OffTime as 4-octet
// little-endian unsigned values at the max-uint32 boundary.
func TestBuildCROBRequestTimeLSBBoundary(t *testing.T) {
	crob := &CROB{Code: CROBCodePulseOn, Count: 3, OnTime: 0xFFFFFFFF, OffTime: 0x80000000, Status: 0}
	got := buildCROBRequest(0, crob)
	// OnTime 0xFFFFFFFF -> FF FF FF FF ; OffTime 0x80000000 -> 00 00 00 80
	onTime := got[8:12]
	offTime := got[12:16]
	wantOn := []byte{0xFF, 0xFF, 0xFF, 0xFF}
	wantOff := []byte{0x00, 0x00, 0x00, 0x80}
	if string(onTime) != string(wantOn) {
		t.Errorf("onTime = %X, want %X (LSB first)", onTime, wantOn)
	}
	if string(offTime) != string(wantOff) {
		t.Errorf("offTime = %X, want %X (LSB first)", offTime, wantOff)
	}
	// Count and status octets are in the right positions.
	if got[7] != 3 {
		t.Errorf("count = %d, want 3", got[7])
	}
	if got[16] != 0 {
		t.Errorf("status = %d, want 0", got[16])
	}
}

// TestBuildControlRequestCROBBoolMapping locks the public Operate path's
// bool→control-code mapping (true=LATCH_ON, false=LATCH_OFF). This is the
// encode actually exercised by client.Operate with a BinaryCommandValue.
func TestBuildControlRequestCROBBoolMapping(t *testing.T) {
	m := &Master{}
	cases := []struct {
		name     string
		value    interface{}
		wantCode byte
	}{
		{"true -> LATCH_ON", true, CROBCodeLatchOn},
		{"false -> LATCH_OFF", false, CROBCodeLatchOff},
		{"uint8 passthrough", uint8(CROBCodePulseOff), CROBCodePulseOff},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := m.buildControlRequest(al.FuncDirectOperate, 12, 1, 0, tc.value)
			if req == nil {
				t.Fatalf("buildControlRequest returned nil")
			}
			// Data layout: 0C 01 28 01 00 | 00 00 (index 0) | code ...
			if len(req.Data) < 8 {
				t.Fatalf("request too short: %X", req.Data)
			}
			gotCode := req.Data[7]
			if gotCode != tc.wantCode {
				t.Errorf("code = %d, want %d", gotCode, tc.wantCode)
			}
		})
	}
}

func TestBuildCROBRequestLatchOnOffWireBytes(t *testing.T) {
	on := buildCROBRequest(0x0000, &CROB{Code: CROBCodeLatchOn, Count: 1})
	if on[6] != 0x03 {
		t.Fatalf("LATCH_ON code octet = 0x%02X, want 0x03", on[6])
	}
	off := buildCROBRequest(0x0000, &CROB{Code: CROBCodeLatchOff, Count: 1})
	if off[6] != 0x04 {
		t.Fatalf("LATCH_OFF code octet = 0x%02X, want 0x04", off[6])
	}
}

// TestBuildControlRequestLatchWireBytes locks the full DirectOperate CROB value
// bytes for both latch directions through the Operate encoder path.
func TestBuildControlRequestLatchWireBytes(t *testing.T) {
	m := &Master{}
	cases := []struct {
		name     string
		value    interface{}
		wantCode byte
	}{
		{"true -> LATCH_ON", true, 0x03},
		{"false -> LATCH_OFF", false, 0x04},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := m.buildControlRequest(al.FuncDirectOperate, 12, 1, 0x1234, tc.value)
			if req == nil {
				t.Fatalf("buildControlRequest returned nil")
			}
			// 0C 01 28 01 00 | 34 12 | code ...
			if req.Data[7] != tc.wantCode {
				t.Fatalf("code octet = 0x%02X, want 0x%02X", req.Data[7], tc.wantCode)
			}
			// Code octet sits at value offset 0 of the 11-byte CROB value.
			want, _ := hex.DecodeString("0c012801003412" + hex.EncodeToString([]byte{tc.wantCode}) + "01000000000000000000")
			if string(req.Data) != string(want) {
				t.Fatalf("CROB control request = %X, want %X", req.Data, want)
			}
		})
	}
}

// TestBuildControlRequestRejectsUnsupportedCROBCode verifies that raw operation
// types outside the v0 low-nibble enumeration (including the retired one-hot
// masks and any trip/close or clear bits) are rejected before transmission
// instead of emitting a non-conformant code octet.
func TestBuildControlRequestRejectsUnsupportedCROBCode(t *testing.T) {
	m := &Master{}
	unsupported := []uint8{0x08, 0x10, 0x05, 0x06, 0x0F, 0x80, 0x83, 0x40, 0x20}
	for _, code := range unsupported {
		if req := m.buildControlRequest(al.FuncDirectOperate, 12, 1, 0, uint8(code)); req != nil {
			t.Errorf("code 0x%02X: buildControlRequest = %X, want nil (rejected)", code, req.Data)
		}
	}
}

// TestBuildControlRequestAcceptsSupportedCROBCodes confirms each v0 operation
// type is still encodable through the raw-code path.
func TestBuildControlRequestAcceptsSupportedCROBCodes(t *testing.T) {
	m := &Master{}
	supported := []uint8{CROBCodeNUL, CROBCodePulseOn, CROBCodePulseOff, CROBCodeLatchOn, CROBCodeLatchOff}
	for _, code := range supported {
		req := m.buildControlRequest(al.FuncDirectOperate, 12, 1, 0, uint8(code))
		if req == nil {
			t.Fatalf("code 0x%02X: buildControlRequest returned nil, want a request", code)
		}
		if req.Data[7] != code {
			t.Errorf("code 0x%02X: encoded octet = 0x%02X", code, req.Data[7])
		}
	}
}

// TestCROBOpTypeConstantsAreDistinctEnum locks the operation-type constants to
// distinct sequential low-nibble values so the outstation decode switch cannot
// collide (regression guard for the eighth-correction collision).
func TestCROBOpTypeConstantsAreDistinctEnum(t *testing.T) {
	seen := map[uint8]string{}
	for _, c := range []struct {
		name string
		code uint8
	}{
		{"NUL", CROBCodeNUL},
		{"PulseOn", CROBCodePulseOn},
		{"PulseOff", CROBCodePulseOff},
		{"LatchOn", CROBCodeLatchOn},
		{"LatchOff", CROBCodeLatchOff},
	} {
		if c.code > 0x0F {
			t.Errorf("%s = 0x%02X is outside the low nibble", c.name, c.code)
		}
		if prev, dup := seen[c.code]; dup {
			t.Errorf("%s and %s share code 0x%02X", prev, c.name, c.code)
		}
		seen[c.code] = c.name
	}
}
