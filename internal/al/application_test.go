package al

import (
	"bytes"
	"testing"
)

func TestAppControlHeader(t *testing.T) {
	tests := []struct {
		name     string
		a        AppControl
		expected byte
	}{
		{"Single Fragment", AppControl{true, true, false, false, 0}, 0xC0},
		{"With CON", AppControl{true, true, true, false, 0}, 0xE0},
		{"Unsolicited", AppControl{true, true, false, true, 0}, 0xD0},
		{"Both CON and UNS", AppControl{true, true, true, true, 0}, 0xF0},
		{"Max Seq", AppControl{true, true, false, false, 15}, 0xCF},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.Header(); got != tt.expected {
				t.Errorf("Header() = 0x%02X, want 0x%02X", got, tt.expected)
			}
		})
	}
}

func TestAppControlSetHeader(t *testing.T) {
	tests := []struct {
		name    string
		header  byte
		wantFIR bool
		wantFIN bool
		wantCON bool
		wantUNS bool
		wantSeq uint8
	}{
		{"Single Fragment", 0xC0, true, true, false, false, 0},
		{"With CON", 0xE5, true, true, true, false, 5},
		{"Unsolicited", 0xD3, true, true, false, true, 3},
		{"Max Seq", 0xFF, true, true, true, true, 15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var a AppControl
			a.SetHeader(tt.header)
			if a.FIR != tt.wantFIR {
				t.Errorf("FIR = %v, want %v", a.FIR, tt.wantFIR)
			}
			if a.FIN != tt.wantFIN {
				t.Errorf("FIN = %v, want %v", a.FIN, tt.wantFIN)
			}
			if a.CON != tt.wantCON {
				t.Errorf("CON = %v, want %v", a.CON, tt.wantCON)
			}
			if a.UNS != tt.wantUNS {
				t.Errorf("UNS = %v, want %v", a.UNS, tt.wantUNS)
			}
			if a.Seq != tt.wantSeq {
				t.Errorf("Seq = %d, want %d", a.Seq, tt.wantSeq)
			}
		})
	}
}

func TestAPDUEncode(t *testing.T) {
	apdu := &APDU{
		Control: AppControl{true, true, false, false, 5},
		FuncCode: FuncRead,
		Data:     []byte{0x01, 0x02, 0x03},
	}

	encoded := apdu.Encode()

	expected := []byte{0xC5, 0x02, 0x01, 0x02, 0x03}
	if !bytes.Equal(encoded, expected) {
		t.Errorf("Encode() = %v, want %v", encoded, expected)
	}
}

func TestAPDUEncodeEmpty(t *testing.T) {
	apdu := &APDU{
		Control: AppControl{true, true, false, false, 0},
		FuncCode: FuncConfirm,
		Data:     nil,
	}

	encoded := apdu.Encode()

	if len(encoded) != 2 {
		t.Errorf("Encode() len = %d, want 2", len(encoded))
	}
	if encoded[0] != 0xC0 || encoded[1] != 0 {
		t.Errorf("Encode() = %v, want [0xC0, 0x00]", encoded)
	}
}

func TestDecodeAPDU(t *testing.T) {
	data := []byte{0xC5, 0x02, 0x01, 0x02, 0x03}

	apdu, err := Decode(data)
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	if !apdu.Control.FIR || !apdu.Control.FIN {
		t.Error("Expected FIR and FIN to be set")
	}
	if apdu.Control.Seq != 5 {
		t.Errorf("Seq = %d, want 5", apdu.Control.Seq)
	}
	if apdu.FuncCode != FuncRead {
		t.Errorf("FuncCode = %d, want %d", apdu.FuncCode, FuncRead)
	}
	if !bytes.Equal(apdu.Data, []byte{0x01, 0x02, 0x03}) {
		t.Errorf("Data = %v, want [1, 2, 3]", apdu.Data)
	}
}

func TestDecodeAPDUTooShort(t *testing.T) {
	_, err := Decode([]byte{0xC0})
	if err == nil {
		t.Error("Expected error for too short APDU")
	}
}

func TestAPDUIsRequest(t *testing.T) {
	req := NewRequest(0, FuncRead)
	if !req.IsRequest() {
		t.Error("Expected IsRequest() = true")
	}

	resp := NewResponse(0, FuncResponse, nil)
	if resp.IsRequest() {
		t.Error("Expected IsRequest() = false for response")
	}
}

func TestAPDUIsResponse(t *testing.T) {
	resp := NewResponse(0, FuncResponse, nil)
	if !resp.IsResponse() {
		t.Error("Expected IsResponse() = true")
	}

	req := NewRequest(0, FuncRead)
	if req.IsResponse() {
		t.Error("Expected IsResponse() = false for request")
	}
}

func TestNewUnsolicited(t *testing.T) {
	unsol := NewUnsolicited(0, []byte{0x01})

	if !unsol.IsUnsolicited() {
		t.Error("Expected IsUnsolicited() = true")
	}
	if !unsol.IsConfirmationRequired() {
		t.Error("Expected IsConfirmationRequired() = true for unsolicited")
	}
	if unsol.FuncCode != FuncUnsolicitedResponse {
		t.Errorf("FuncCode = %d, want %d", unsol.FuncCode, FuncUnsolicitedResponse)
	}
}

func TestIINBytes(t *testing.T) {
	iin := &IIN{
		AllStop:      true,
		ByteOver:     true,
		NeedsTimeSync: true,
		Busy:         true,
	}

	bytes := iin.Bytes()

	if bytes[0] != 0xC2 { // AllStop + ByteOver + Busy = 0x80 + 0x40 + 0x02
		t.Errorf("IIN[0] = 0x%02X, want 0xC2", bytes[0])
	}
	if bytes[1] != 0x04 { // NeedsTimeSync
		t.Errorf("IIN[1] = 0x%02X, want 0x04", bytes[1])
	}
}

func TestIINSetIIN(t *testing.T) {
	var iin IIN
	iin.SetIIN([2]byte{0x82, 0x04})

	if !iin.AllStop {
		t.Error("Expected AllStop = true")
	}
	if !iin.Busy {
		t.Error("Expected Busy = true")
	}
	if !iin.NeedsTimeSync {
		t.Error("Expected NeedsTimeSync = true")
	}
	if iin.CheckFail {
		t.Error("Expected CheckFail = false")
	}
}

func TestEncodeDecodeIIN(t *testing.T) {
	original := &IIN{
		AllStop:        true,
		ByteOver:       true,
		Limit64K:       true,
		ConfigError:    true,
		NeedsTimeSync:  true,
	}

	encoded := EncodeIIN(original)
	decoded, err := DecodeIIN(encoded)
	if err != nil {
		t.Fatalf("DecodeIIN() error = %v", err)
	}

	if decoded.AllStop != original.AllStop {
		t.Error("AllStop mismatch")
	}
	if decoded.ByteOver != original.ByteOver {
		t.Error("ByteOver mismatch")
	}
	if decoded.Limit64K != original.Limit64K {
		t.Error("Limit64K mismatch")
	}
	if decoded.ConfigError != original.ConfigError {
		t.Error("ConfigError mismatch")
	}
	if decoded.NeedsTimeSync != original.NeedsTimeSync {
		t.Error("NeedsTimeSync mismatch")
	}
}

func TestDecodeIINTooShort(t *testing.T) {
	_, err := DecodeIIN([]byte{0x00})
	if err == nil {
		t.Error("Expected error for too short IIN")
	}
}

func TestResponseEncodeDecode(t *testing.T) {
	resp := NewAppResponse(5, IIN{Busy: true, CheckFail: true}, []byte{0x01, 0x02})

	encoded := resp.Encode()
	if len(encoded) < 4 {
		t.Fatalf("Response too short: %d bytes", len(encoded))
	}

	decoded, err := DecodeResponse(encoded)
	if err != nil {
		t.Fatalf("DecodeResponse() error = %v", err)
	}

	if decoded.Header.Control.Seq != 5 {
		t.Errorf("Seq = %d, want 5", decoded.Header.Control.Seq)
	}
	if decoded.Header.FuncCode != FuncResponse {
		t.Errorf("FuncCode = %d, want %d", decoded.Header.FuncCode, FuncResponse)
	}
	if !decoded.IIN.Busy {
		t.Error("Expected IIN.Busy = true")
	}
	if !decoded.IIN.CheckFail {
		t.Error("Expected IIN.CheckFail = true")
	}
	if !bytes.Equal(decoded.Data, []byte{0x01, 0x02}) {
		t.Errorf("Data = %v, want [1, 2]", decoded.Data)
	}
}

func TestResponseEncodeDecodeEmpty(t *testing.T) {
	resp := NewAppResponse(0, IIN{}, nil)

	encoded := resp.Encode()
	decoded, err := DecodeResponse(encoded)
	if err != nil {
		t.Fatalf("DecodeResponse() error = %v", err)
	}

	if len(decoded.Data) != 0 {
		t.Errorf("Data len = %d, want 0", len(decoded.Data))
	}
}

func TestResponseTooShort(t *testing.T) {
	_, err := DecodeResponse([]byte{0xC0, 0x00, 0x00})
	if err == nil {
		t.Error("Expected error for too short response")
	}
}

func TestAPDUString(t *testing.T) {
	apdu := NewRequest(5, FuncRead)
	s := apdu.String()
	
	if s == "" {
		t.Error("String() returned empty string")
	}
}
