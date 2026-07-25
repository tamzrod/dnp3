package frame

import (
	"bytes"
	"testing"
)

// TestEncodeDecodeResetLink tests encoding and decoding a Reset Link Stations frame.
func TestEncodeDecodeResetLink(t *testing.T) {
	f := &Frame{
		Control: Control{
			DIR:      true,
			PRM:      true,
			FCB:      false,
			FCV:      false,
			FuncCode: FuncResetLinkStations,
		},
		DestAddr: 0xFFFF,
		SrcAddr:  0x0000,
		Data:     nil,
	}

	// Encode
	encoded, err := Encode(f)
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	// Verify sync bytes
	if encoded[0] != SyncByte1 || encoded[1] != SyncByte2 {
		t.Errorf("Sync bytes = 0x%02x 0x%02x, want 0x05 0x64", encoded[0], encoded[1])
	}

	// Verify length
	expectedLength := byte(1 + 2 + 2) // Control + Dest + Src (no data)
	if encoded[2] != expectedLength {
		t.Errorf("Length = %d, want %d", encoded[2], expectedLength)
	}

	// Decode
	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	// Verify fields
	if decoded.DestAddr != f.DestAddr {
		t.Errorf("DestAddr = 0x%04X, want 0x%04X", decoded.DestAddr, f.DestAddr)
	}
	if decoded.SrcAddr != f.SrcAddr {
		t.Errorf("SrcAddr = 0x%04X, want 0x%04X", decoded.SrcAddr, f.SrcAddr)
	}
	if decoded.Control.DIR != f.Control.DIR {
		t.Errorf("DIR = %v, want %v", decoded.Control.DIR, f.Control.DIR)
	}
	if decoded.Control.PRM != f.Control.PRM {
		t.Errorf("PRM = %v, want %v", decoded.Control.PRM, f.Control.PRM)
	}
	if decoded.Control.FuncCode != f.Control.FuncCode {
		t.Errorf("FuncCode = %d, want %d", decoded.Control.FuncCode, f.Control.FuncCode)
	}
}

// TestEncodeDecodeWithData tests encoding and decoding a frame with data.
func TestEncodeDecodeWithData(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03, 0x04, 0x05}

	f := &Frame{
		Control: Control{
			DIR:      true,
			PRM:      true,
			FCB:      true,
			FCV:      true,
			FuncCode: FuncConfirmedUserData,
		},
		DestAddr: 0x1234,
		SrcAddr:  0x5678,
		Data:     data,
	}

	// Encode
	encoded, err := Encode(f)
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	// Verify minimum size
	if len(encoded) < MinFrameSize {
		t.Errorf("Encoded length = %d, minimum = %d", len(encoded), MinFrameSize)
	}

	// Verify maximum size
	if len(encoded) > MaxFrameSize {
		t.Errorf("Encoded length = %d, maximum = %d", len(encoded), MaxFrameSize)
	}

	// Decode
	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	// Verify data
	if !bytes.Equal(decoded.Data, data) {
		t.Errorf("Data = %x, want %x", decoded.Data, data)
	}
}

// TestEncodeDataTooLarge tests that encoding fails for data exceeding max size.
func TestEncodeDataTooLarge(t *testing.T) {
	f := &Frame{
		Control: Control{
			DIR:      true,
			PRM:      true,
			FuncCode: FuncConfirmedUserData,
		},
		DestAddr: 0x1234,
		SrcAddr:  0x5678,
		Data:     make([]byte, MaxDataSize+1), // Too large
	}

	_, err := Encode(f)
	if err == nil {
		t.Error("Encode() should return error for data too large")
	}
}

// TestDecodeInvalidSync tests that decoding fails for invalid sync bytes.
func TestDecodeInvalidSync(t *testing.T) {
	data := []byte{0x00, 0x00, 0x05, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

	_, err := Decode(data)
	if err == nil {
		t.Error("Decode() should return error for invalid sync bytes")
	}
}

// TestDecodeTooShort tests that decoding fails for frames that are too short.
func TestDecodeTooShort(t *testing.T) {
	data := []byte{0x05, 0x64, 0x00}

	_, err := Decode(data)
	if err == nil {
		t.Error("Decode() should return error for frame too short")
	}
}

// TestControlByte tests control byte encoding and decoding.
func TestControlByte(t *testing.T) {
	tests := []struct {
		name    string
		control Control
		want    byte
	}{
		{
			// DNP3 Control byte format (IEEE 1815-2012 Section 5.2):
			// Bit 7: DIR, Bit 6: PRM, Bit 5: FCB, Bit 4: FCV, Bits 3-0: FuncCode
			name:    "master to outstation reset",
			control: Control{DIR: true, PRM: true, FuncCode: FuncResetLinkStations},
			want:    0xC0, // 1100 0000: DIR=1, PRM=1, FCB=0, FCV=0, FuncCode=0
		},
		{
			name:    "outstation to master ack",
			control: Control{DIR: false, PRM: false, FuncCode: FuncAck},
			want:    0x00, // 0000 0000: DIR=0, PRM=0, FuncCode=0
		},
		{
			name:    "confirmed data with FCB",
			control: Control{DIR: true, PRM: true, FCB: true, FCV: true, FuncCode: FuncConfirmedUserData},
			want:    0xF4, // 1111 0100: DIR=1, PRM=1, FCB=1, FCV=1, FuncCode=4
		},
		{
			name:    "confirmed data no FCB",
			control: Control{DIR: true, PRM: true, FCB: false, FCV: true, FuncCode: FuncConfirmedUserData},
			want:    0xD4, // 1101 0100: DIR=1, PRM=1, FCB=0, FCV=1, FuncCode=4
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.control.ToByte()
			if got != tt.want {
				t.Errorf("ToByte() = 0x%02X, want 0x%02X", got, tt.want)
			}

			// Round-trip test
			var parsed Control
			parsed.FromByte(got)
			if parsed.DIR != tt.control.DIR {
				t.Errorf("DIR round-trip = %v, want %v", parsed.DIR, tt.control.DIR)
			}
			if parsed.PRM != tt.control.PRM {
				t.Errorf("PRM round-trip = %v, want %v", parsed.PRM, tt.control.PRM)
			}
			if parsed.FCB != tt.control.FCB {
				t.Errorf("FCB round-trip = %v, want %v", parsed.FCB, tt.control.FCB)
			}
			if parsed.FCV != tt.control.FCV {
				t.Errorf("FCV round-trip = %v, want %v", parsed.FCV, tt.control.FCV)
			}
			if parsed.FuncCode != tt.control.FuncCode {
				t.Errorf("FuncCode round-trip = %d, want %d", parsed.FuncCode, tt.control.FuncCode)
			}
		})
	}
}

// TestIsBroadcast tests the IsBroadcast method.
func TestIsBroadcast(t *testing.T) {
	tests := []struct {
		addr uint16
		want bool
	}{
		// 0xFFFF is the broadcast address per IEEE 1815-2012 Section 5.3
		{AddrBroadcast, true},
		// 0xFFFA (All-stations reset) is a special function address,
		// NOT the general broadcast address
		{AddrAllReset, false},
		{0x0001, false},
		{0xFFFD, false}, // Primary channel address
	}

	for _, tt := range tests {
		f := &Frame{DestAddr: tt.addr}
		if got := f.IsBroadcast(); got != tt.want {
			t.Errorf("IsBroadcast(addr=0x%04X) = %v, want %v", tt.addr, got, tt.want)
		}
	}
}

// TestFrameString tests the String method.
func TestFrameString(t *testing.T) {
	f := &Frame{
		Control: Control{
			DIR:      true,
			PRM:      true,
			FuncCode: FuncConfirmedUserData,
		},
		DestAddr: 0x1234,
		SrcAddr:  0x5678,
		Data:     []byte{0x01, 0x02, 0x03},
	}

	s := f.String()
	if s == "" {
		t.Error("String() should not return empty string")
	}

	// Should contain key fields
	if !bytes.Contains([]byte(s), []byte("DIR=true")) {
		t.Error("String() should contain DIR=true")
	}
	if !bytes.Contains([]byte(s), []byte("PRM=true")) {
		t.Error("String() should contain PRM=true")
	}
}

// TestFrameEqual tests the Equal method.
func TestFrameEqual(t *testing.T) {
	f1 := &Frame{
		Control:  Control{DIR: true, PRM: true, FuncCode: FuncResetLinkStations},
		DestAddr: 0x1234,
		SrcAddr:  0x5678,
		Data:     []byte{0x01, 0x02},
	}

	f2 := &Frame{
		Control:  Control{DIR: true, PRM: true, FuncCode: FuncResetLinkStations},
		DestAddr: 0x1234,
		SrcAddr:  0x5678,
		Data:     []byte{0x01, 0x02},
	}

	f3 := &Frame{
		Control:  Control{DIR: true, PRM: true, FuncCode: FuncResetLinkStations},
		DestAddr: 0x1234,
		SrcAddr:  0x5678,
		Data:     []byte{0x01, 0x03}, // Different data
	}

	if !f1.Equal(f2) {
		t.Error("Equal frames should return true")
	}

	if f1.Equal(f3) {
		t.Error("Different frames should return false")
	}
}

// BenchmarkEncode benchmarks frame encoding.
func BenchmarkEncode(b *testing.B) {
	f := &Frame{
		Control: Control{
			DIR:      true,
			PRM:      true,
			FCB:      true,
			FCV:      true,
			FuncCode: FuncConfirmedUserData,
		},
		DestAddr: 0x1234,
		SrcAddr:  0x5678,
		Data:     []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Encode(f)
	}
}

// BenchmarkDecode benchmarks frame decoding.
func BenchmarkDecode(b *testing.B) {
	// Pre-encode a frame
	f := &Frame{
		Control: Control{
			DIR:      true,
			PRM:      true,
			FCB:      true,
			FCV:      true,
			FuncCode: FuncConfirmedUserData,
		},
		DestAddr: 0x1234,
		SrcAddr:  0x5678,
		Data:     []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
	}
	encoded, _ := Encode(f)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Decode(encoded)
	}
}
