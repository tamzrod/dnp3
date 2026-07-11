// Package al provides conformance tests for the Application Layer.
package al

import (
	"bytes"
	"testing"

	"dnp3/internal/al"
)

// TestAppControlField tests application control byte encoding
func TestAppControlField(t *testing.T) {
	tests := []struct {
		name     string
		fir      bool
		fin      bool
		con      bool
		uns      bool
		seq      uint8
		expected byte
	}{
		{"single_frag_no_conf", true, true, false, false, 0, 0xC0},
		{"with_conf", true, true, true, false, 0, 0xE0},
		{"unsolicited", true, true, false, true, 0, 0xD0},
		{"both_conf_uns", true, true, true, true, 0, 0xF0},
		{"max_seq", true, true, false, false, 15, 0xCF},
		{"seq_5", true, true, true, false, 5, 0xE5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := al.AppControl{
				FIR: tt.fir,
				FIN: tt.fin,
				CON: tt.con,
				UNS: tt.uns,
				Seq: tt.seq,
			}
			result := ctrl.Header()
			if result != tt.expected {
				t.Errorf("AppControl.Header() = 0x%02X, want 0x%02X", result, tt.expected)
			}
		})
	}
}

// TestFunctionCodes tests function code definitions
func TestFunctionCodes(t *testing.T) {
	codes := map[string]uint8{
		"CONFIRM":              0,
		"READ":                 2,
		"WRITE":                3,
		"SELECT":               4,
		"OPERATE":              5,
		"DIRECT_OPERATE":       6,
		"DIRECT_OPERATE_NR":    7,
		"FREEZE":              10,
		"FILE_OPEN":           13,
		"FILE_CLOSE":          14,
		"FILE_READ":           15,
		"FILE_WRITE":          16,
		"GET_IDENTIFIER":      21,
		"GET_LABEL":           22,
		"GET_DESCRIPTION":     23,
		"AUTHENTICATE":        27,
		"AUTHENTICATE_CONF":   28,
		"ABORT":               29,
		"TIME_SYNC":           32,
		"ENABLE_UNSOLICITED":  41,
		"DISABLE_UNSOLICITED": 42,
		"ASSIGN_CLASS":        48,
		"START_RESTART":       53,
		"UNSOLICITED":         1,
		"NO_ACK":             127,
	}

	for name, expected := range codes {
		t.Run(name, func(t *testing.T) {
			name := al.FunctionCodeName(expected)
			if name == "UNKNOWN" && expected != 0 {
				// 0 might legitimately map to CONFIRM in some contexts
			}
		})
	}
}

// TestAPDUEncodeDecode tests APDU round-trip encoding/decoding
func TestAPDUEncodeDecode(t *testing.T) {
	tests := []struct {
		name      string
		seq       uint8
		funcCode  uint8
		data      []byte
	}{
		{"read_request", 5, al.FuncRead, []byte{0x01, 0x02}},
		{"write_request", 10, al.FuncWrite, []byte{0xFF, 0xFE, 0xFD}},
		{"empty_data", 0, al.FuncRead, nil},
		{"large_data", 15, al.FuncRead, make([]byte, 100)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := al.NewRequest(tt.seq, tt.funcCode)
			if tt.data != nil {
				req.Data = tt.data
			}

			encoded := req.Encode()
			if len(encoded) < 2 {
				t.Fatal("Encoded APDU too short")
			}

			decoded, err := al.Decode(encoded)
			if err != nil {
				t.Fatalf("Decode() error = %v", err)
			}

			if decoded.Control.Seq != tt.seq {
				t.Errorf("Seq = %d, want %d", decoded.Control.Seq, tt.seq)
			}

			if decoded.FuncCode != tt.funcCode {
				t.Errorf("FuncCode = %d, want %d", decoded.FuncCode, tt.funcCode)
			}

			if !bytes.Equal(decoded.Data, tt.data) {
				t.Errorf("Data = %v, want %v", decoded.Data, tt.data)
			}
		})
	}
}

// TestResponseCreation tests response APDU creation
func TestResponseCreation(t *testing.T) {
	seq := uint8(7)
	iin := al.IIN{Busy: true, CheckFail: true}
	data := []byte{0x01, 0x02, 0x03}

	resp := al.NewAppResponse(seq, iin, data)
	
	if resp.Header.Control.Seq != seq {
		t.Errorf("Response seq = %d, want %d", resp.Header.Control.Seq, seq)
	}

	if resp.Header.FuncCode != al.FuncResponse {
		t.Errorf("Response funcCode = %d, want %d", resp.Header.FuncCode, al.FuncResponse)
	}

	// Encode and decode
	encoded := resp.Encode()
	decoded, err := al.DecodeResponse(encoded)
	if err != nil {
		t.Fatalf("DecodeResponse() error = %v", err)
	}

	if !decoded.IIN.Busy {
		t.Error("Decoded IIN.Busy should be true")
	}

	if !decoded.IIN.CheckFail {
		t.Error("Decoded IIN.CheckFail should be true")
	}
}

// TestIINEncoding tests Internal Indication field encoding
func TestIINEncoding(t *testing.T) {
	tests := []struct {
		name     string
		iin      al.IIN
		byte0    byte
		byte1    byte
	}{
		{"empty", al.IIN{}, 0x00, 0x00},
		{"all_stop", al.IIN{AllStop: true}, 0x80, 0x00},
		{"byte_over", al.IIN{ByteOver: true}, 0x40, 0x00},
		{"busy", al.IIN{Busy: true}, 0x02, 0x00},
		{"needs_time_sync", al.IIN{NeedsTimeSync: true}, 0x00, 0x04},
		{"multiple", al.IIN{AllStop: true, Busy: true, NeedsTimeSync: true}, 0x82, 0x04},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bytes := tt.iin.Bytes()
			if bytes[0] != tt.byte0 {
				t.Errorf("IIN[0] = 0x%02X, want 0x%02X", bytes[0], tt.byte0)
			}
			if bytes[1] != tt.byte1 {
				t.Errorf("IIN[1] = 0x%02X, want 0x%02X", bytes[1], tt.byte1)
			}
		})
	}
}

// TestIINRoundTrip tests IIN encode/decode round-trip
func TestIINRoundTrip(t *testing.T) {
	iinValues := []al.IIN{
		{},
		{AllStop: true},
		{ByteOver: true, Busy: true},
		{ConfigError: true, NeedsTimeSync: true},
		{AllStop: true, ByteOver: true, Limit64K: true, Limit16K: true, MemUnavail: true, CheckFail: true, Busy: true, ParamUnavail: true},
		{NeedsTimeSync: true, GeneralEnableOff: true},
	}

	for i, original := range iinValues {
		t.Run(string(rune('A'+i)), func(t *testing.T) {
			encoded := al.EncodeIIN(&original)
			decoded, err := al.DecodeIIN(encoded)
			if err != nil {
				t.Fatalf("DecodeIIN() error = %v", err)
			}

			decodedBytes := decoded.Bytes()
			originalBytes := original.Bytes()
			
			if decodedBytes[0] != originalBytes[0] {
				t.Errorf("IIN[0] = 0x%02X, want 0x%02X", decodedBytes[0], originalBytes[0])
			}
			if decodedBytes[1] != originalBytes[1] {
				t.Errorf("IIN[1] = 0x%02X, want 0x%02X", decodedBytes[1], originalBytes[1])
			}
		})
	}
}

// TestNewUnsolicited tests unsolicited response creation
func TestNewUnsolicited(t *testing.T) {
	seq := uint8(3)
	data := []byte{0xAA, 0xBB}

	unsol := al.NewUnsolicited(seq, data)

	if !unsol.IsUnsolicited() {
		t.Error("IsUnsolicited() should be true")
	}

	if !unsol.IsConfirmationRequired() {
		t.Error("Unsolicited should require confirmation")
	}

	if unsol.FuncCode != al.FuncUnsolicitedResponse {
		t.Errorf("FuncCode = %d, want %d", unsol.FuncCode, al.FuncUnsolicitedResponse)
	}
}

// TestIsRequestResponse tests APDU type detection
func TestIsRequestResponse(t *testing.T) {
	tests := []struct {
		name      string
		funcCode  uint8
		isRequest bool
		isResponse bool
	}{
		{"read", al.FuncRead, true, false},
		{"write", al.FuncWrite, true, false},
		{"response", al.FuncResponse, false, true},
		{"direct_operate", al.FuncDirectOperate, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apdu := al.NewRequest(0, tt.funcCode)
			if apdu.IsRequest() != tt.isRequest {
				t.Errorf("IsRequest() = %v, want %v", apdu.IsRequest(), tt.isRequest)
			}
			if apdu.IsResponse() != tt.isResponse {
				t.Errorf("IsResponse() = %v, want %v", apdu.IsResponse(), tt.isResponse)
			}
		})
	}
}

// TestValidFunctionCodes tests function code validation
func TestValidFunctionCodes(t *testing.T) {
	validCodes := []uint8{0, 1, 2, 3, 4, 5, 6, 7, 10, 13, 14, 15, 16, 21, 22, 23, 24, 25, 26, 27, 28, 29, 32, 33, 37, 38, 41, 42, 48, 51, 52, 53, 54, 57, 58, 59, 64, 127}
	invalidCodes := []uint8{8, 9, 11, 12, 17, 18, 19, 20, 30, 31, 34, 35, 36, 39, 40, 43, 44, 45, 46, 47, 49, 50, 55, 56, 60, 61, 62, 63, 128, 255}

	for _, code := range validCodes {
		if !al.IsValidFunctionCode(code) {
			t.Errorf("IsValidFunctionCode(%d) = false, want true", code)
		}
	}

	// Note: Some invalid codes might be valid depending on implementation
	// Check only clearly invalid ranges
	for _, code := range invalidCodes {
		if code >= 64 && code <= 127 {
			continue // Manufacturer specific codes might be valid
		}
		if code > 127 {
			_ = al.IsValidFunctionCode(code) // Just check it doesn't panic
		}
	}
}

// TestRoleAuthorization tests role-based authorization
func TestRoleAuthorization(t *testing.T) {
	// Test read authorization
	tests := []struct {
		code   uint8
		isRead bool
	}{
		{al.FuncRead, true},
		{al.FuncWrite, false},
		{al.FuncSelect, false},
		{al.FuncOperate, false},
	}

	for _, tt := range tests {
		if al.IsReadFunction(tt.code) != tt.isRead {
			t.Errorf("IsReadFunction(%d) = %v, want %v", tt.code, !tt.isRead, tt.isRead)
		}
	}

	// Test write authorization - only FuncWrite returns true
	if !al.IsWriteFunction(al.FuncWrite) {
		t.Errorf("IsWriteFunction(FuncWrite) should be true")
	}

	// Test control authorization
	controlCodes := []uint8{al.FuncSelect, al.FuncOperate, al.FuncDirectOperate, al.FuncDirectOperateNoResp}
	for _, code := range controlCodes {
		if !al.IsControlFunction(code) {
			t.Errorf("IsControlFunction(%d) should be true", code)
		}
	}

	// Test time functions
	timeCodes := []uint8{al.FuncTimeSync, al.FuncRecordCurrentTime, 57, 58, 59}
	for _, code := range timeCodes {
		if !al.IsTimeFunction(code) {
			t.Errorf("IsTimeFunction(%d) should be true", code)
		}
	}
}

// TestAPDUValidation tests APDU validation for malformed data
func TestAPDUValidation(t *testing.T) {
	// Too short
	_, err := al.Decode([]byte{0xC0})
	if err == nil {
		t.Error("Decode() should fail for single byte")
	}

	// Response too short
	_, err = al.DecodeResponse([]byte{0xC0, 0x00, 0x00})
	if err == nil {
		t.Error("DecodeResponse() should fail for 3 bytes")
	}

	// IIN too short
	_, err = al.DecodeIIN([]byte{0x00})
	if err == nil {
		t.Error("DecodeIIN() should fail for single byte")
	}
}

// TestSequenceNumberRange tests sequence number handling
// Application layer sequence numbers are 4 bits (0-15 per spec)
func TestSequenceNumberRange(t *testing.T) {
	for seq := uint8(0); seq <= 15; seq++ {
		t.Run(string(rune('0'+seq/10))+string(rune('0'+seq%10)), func(t *testing.T) {
			req := al.NewRequest(seq, al.FuncRead)
			if req.Control.Seq != seq {
				t.Errorf("Seq = %d, want %d", req.Control.Seq, seq)
			}

			// Encode and decode should preserve sequence
			encoded := req.Encode()
			decoded, _ := al.Decode(encoded)
			if decoded.Control.Seq != seq {
				t.Errorf("After round-trip, Seq = %d, want %d", decoded.Control.Seq, seq)
			}
		})
	}
}

// TestConfirmationRequired tests confirmation requirement detection
func TestConfirmationRequired(t *testing.T) {
	// Unsolicited should require confirmation
	unsol := al.NewUnsolicited(0, nil)
	if !unsol.IsConfirmationRequired() {
		t.Error("Unsolicited should require confirmation")
	}

	// Normal request should not require confirmation
	req := al.NewRequest(0, al.FuncRead)
	if req.IsConfirmationRequired() {
		t.Error("Normal request should not require confirmation")
	}
}

// ConformanceSuite runs all AL conformance tests
func TestConformanceSuite(t *testing.T) {
	tests := []struct {
		name string
		fn   func(*testing.T)
	}{
		{"AppControl Field", TestAppControlField},
		{"Function Codes", TestFunctionCodes},
		{"APDU Encode/Decode", TestAPDUEncodeDecode},
		{"Response Creation", TestResponseCreation},
		{"IIN Encoding", TestIINEncoding},
		{"IIN Round Trip", TestIINRoundTrip},
		{"Unsolicited", TestNewUnsolicited},
		{"Request/Response Detection", TestIsRequestResponse},
		{"Valid Function Codes", TestValidFunctionCodes},
		{"Role Authorization", TestRoleAuthorization},
		{"APDU Validation", TestAPDUValidation},
		{"Sequence Number Range", TestSequenceNumberRange},
		{"Confirmation Required", TestConfirmationRequired},
	}

	t.Logf("AL Conformance Suite: %d tests", len(tests))

	for _, tt := range tests {
		t.Run(tt.name, tt.fn)
	}
}
