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
		Control:  AppControl{true, true, false, false, 5},
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
		Control:  AppControl{true, true, false, false, 0},
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
	// Verified mapping (IEEE 1815-2012): IIN1.0=0x80 (AllStations),
	// IIN1.1=0x40 (Class1Events), IIN1.6=0x02 (DeviceTrouble),
	// IIN1.4=0x08 (NeedTime); IIN2.5=0x04 (BadConfig).
	iin := &IIN{
		AllStations:   true,
		Class1Events:  true,
		DeviceTrouble: true,
		NeedTime:      true,
		BadConfig:     true,
	}

	bytes := iin.Bytes()

	// AllStations(0x80) + Class1Events(0x40) + NeedTime(0x08) + DeviceTrouble(0x02) = 0xCA
	if bytes[0] != 0xCA {
		t.Errorf("IIN[0] = 0x%02X, want 0xCA", bytes[0])
	}
	// BadConfig = 0x04
	if bytes[1] != 0x04 {
		t.Errorf("IIN[1] = 0x%02X, want 0x04", bytes[1])
	}
}

func TestIINSetIIN(t *testing.T) {
	var iin IIN
	iin.SetIIN([2]byte{0xCA, 0x04})

	if !iin.AllStations {
		t.Error("Expected AllStations = true")
	}
	if !iin.Class1Events {
		t.Error("Expected Class1Events = true")
	}
	if !iin.NeedTime {
		t.Error("Expected NeedTime = true")
	}
	if !iin.DeviceTrouble {
		t.Error("Expected DeviceTrouble = true")
	}
	if !iin.BadConfig {
		t.Error("Expected BadConfig = true")
	}
	if iin.LocalControl {
		t.Error("Expected LocalControl = false")
	}
}

func TestEncodeDecodeIIN(t *testing.T) {
	original := &IIN{
		AllStations:    true,
		Class1Events:   true,
		Class2Events:   true,
		BadConfig:      true,
		BufferOverflow: true,
	}

	encoded := EncodeIIN(original)
	decoded, err := DecodeIIN(encoded)
	if err != nil {
		t.Fatalf("DecodeIIN() error = %v", err)
	}

	if decoded.AllStations != original.AllStations {
		t.Error("AllStations mismatch")
	}
	if decoded.Class1Events != original.Class1Events {
		t.Error("Class1Events mismatch")
	}
	if decoded.Class2Events != original.Class2Events {
		t.Error("Class2Events mismatch")
	}
	if decoded.BadConfig != original.BadConfig {
		t.Error("BadConfig mismatch")
	}
	if decoded.BufferOverflow != original.BufferOverflow {
		t.Error("BufferOverflow mismatch")
	}
}

// TestIINBitPositions verifies each IIN bit maps to the verified octet/position.
func TestIINBitPositions(t *testing.T) {
	tests := []struct {
		name  string
		set   func(i *IIN)
		wantB byte
		want  byte
	}{
		{"IIN1.0 AllStations", func(i *IIN) { i.AllStations = true }, 0, 0x80},
		{"IIN1.1 Class1Events", func(i *IIN) { i.Class1Events = true }, 0, 0x40},
		{"IIN1.2 Class2Events", func(i *IIN) { i.Class2Events = true }, 0, 0x20},
		{"IIN1.3 Class3Events", func(i *IIN) { i.Class3Events = true }, 0, 0x10},
		{"IIN1.4 NeedTime", func(i *IIN) { i.NeedTime = true }, 0, 0x08},
		{"IIN1.5 LocalControl", func(i *IIN) { i.LocalControl = true }, 0, 0x04},
		{"IIN1.6 DeviceTrouble", func(i *IIN) { i.DeviceTrouble = true }, 0, 0x02},
		{"IIN1.7 DeviceRestart", func(i *IIN) { i.DeviceRestart = true }, 0, 0x01},
		{"IIN2.0 FuncUnknown", func(i *IIN) { i.FuncUnknown = true }, 1, 0x80},
		{"IIN2.1 ObjectUnknown", func(i *IIN) { i.ObjectUnknown = true }, 1, 0x40},
		{"IIN2.2 ParameterError", func(i *IIN) { i.ParameterError = true }, 1, 0x20},
		{"IIN2.3 BufferOverflow", func(i *IIN) { i.BufferOverflow = true }, 1, 0x10},
		{"IIN2.4 AlreadyExecuting", func(i *IIN) { i.AlreadyExecuting = true }, 1, 0x08},
		{"IIN2.5 BadConfig", func(i *IIN) { i.BadConfig = true }, 1, 0x04},
	}
	for _, tt := range tests {
		iin := &IIN{}
		tt.set(iin)
		got := iin.Bytes()[tt.wantB]
		if got != tt.want {
			t.Errorf("%s: byte[%d] = 0x%02X, want 0x%02X", tt.name, tt.wantB, got, tt.want)
		}
	}
}

func TestDecodeIINTooShort(t *testing.T) {
	_, err := DecodeIIN([]byte{0x00})
	if err == nil {
		t.Error("Expected error for too short IIN")
	}
}

func TestResponseEncodeDecode(t *testing.T) {
	resp := NewAppResponse(5, IIN{DeviceTrouble: true, LocalControl: true}, []byte{0x01, 0x02})

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
	if !decoded.IIN.DeviceTrouble {
		t.Error("Expected IIN.DeviceTrouble = true")
	}
	if !decoded.IIN.LocalControl {
		t.Error("Expected IIN.LocalControl = true")
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
