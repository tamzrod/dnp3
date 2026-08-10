package frame

import (
	"bytes"
	"encoding/hex"
	"os"
	"strings"
	"testing"
)

// TestDecodeRacomGoldenFrame proves the decoder accepts an independently
// published DNP3 frame. It is intentionally introduced before the wire-format
// repair and must fail against the current implementation.
func TestDecodeRacomGoldenFrame(t *testing.T) {
	raw, err := os.ReadFile("../../../active_work/testdata/racom-dnp3-link-frame.hex")
	if err != nil {
		t.Fatalf("read golden fixture: %v", err)
	}

	var fields []string
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields = append(fields, strings.Fields(line)...)
	}
	encoded, err := hex.DecodeString(strings.Join(fields, ""))
	if err != nil {
		t.Fatalf("decode golden fixture hex: %v", err)
	}

	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatalf("decode published DNP3 frame: %v", err)
	}
	if decoded.Control.FuncCode != FuncUnconfirmedUserData {
		t.Fatalf("function code = %d, want unconfirmed user data (4)", decoded.Control.FuncCode)
	}
	if decoded.DestAddr != 4 || decoded.SrcAddr != 3 {
		t.Fatalf("addresses = %#04x -> %#04x, want 0x0003 -> 0x0004", decoded.SrcAddr, decoded.DestAddr)
	}
	if !bytes.Equal(decoded.Data, []byte{0xE5, 0xC0, 0x01, 0x02, 0x00, 0x06}) {
		t.Fatalf("payload = %x, want e5c001020006", decoded.Data)
	}
}

func TestDecodeRejectsCorruptedHeader(t *testing.T) {
	encoded, err := hex.DecodeString("05640bc404000300e42be5c001020006985c")
	if err != nil {
		t.Fatalf("decode fixture literal: %v", err)
	}
	for i := 0; i < 8; i++ {
		corrupt := append([]byte(nil), encoded...)
		corrupt[i] ^= 0x01
		if _, err := Decode(corrupt); err == nil {
			t.Fatalf("Decode accepted corruption at header byte %d", i)
		}
	}
}

func TestPayloadCRCBoundaryVectors(t *testing.T) {
	tests := []struct {
		name       string
		dataLen    int
		headerCRC  uint16
		payloadCRC []uint16
		frameSize  int
	}{
		{"zero", 0, 0xDAE1, nil, 10},
		{"one", 1, 0x49B1, []uint16{0xFFFF}, 13},
		{"sixteen", 16, 0xA220, []uint16{0x10EC}, 28},
		{"seventeen", 17, 0x3170, []uint16{0x10EC, 0x4D94}, 31},
		{"two-forty-nine", 249, 0xE5D0, []uint16{0x10EC, 0x0327, 0x377A, 0x24B1, 0x5FC0, 0x4C0B, 0x7856, 0x6B9D, 0x8EB4, 0x9D7F, 0xA922, 0xBAE9, 0xC198, 0xD253, 0xE60E, 0x0BE1}, 291},
		{"two-fifty", 250, 0x5037, []uint16{0x10EC, 0x0327, 0x377A, 0x24B1, 0x5FC0, 0x4C0B, 0x7856, 0x6B9D, 0x8EB4, 0x9D7F, 0xA922, 0xBAE9, 0xC198, 0xD253, 0xE60E, 0xA0DC}, 292},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := make([]byte, tt.dataLen)
			for i := range payload { payload[i] = byte(i) }
			encoded, err := Encode(&Frame{Control: Control{DIR: true, PRM: true, FuncCode: FuncUnconfirmedUserData}, DestAddr: 4, SrcAddr: 3, Data: payload})
			if err != nil { t.Fatalf("encode: %v", err) }
			if len(encoded) != tt.frameSize { t.Fatalf("frame size = %d, want %d", len(encoded), tt.frameSize) }
			gotHeader := uint16(encoded[8]) | uint16(encoded[9])<<8
			if gotHeader != tt.headerCRC { t.Fatalf("header CRC = %04X, want %04X", gotHeader, tt.headerCRC) }
			offset := 10
			for i, want := range tt.payloadCRC {
				blockLen := 16
				if remaining := tt.dataLen - i*16; remaining < blockLen { blockLen = remaining }
				crcOffset := offset + blockLen
				got := uint16(encoded[crcOffset]) | uint16(encoded[crcOffset+1])<<8
				if got != want { t.Fatalf("payload CRC[%d] = %04X, want %04X", i, got, want) }
				offset += blockLen + 2
			}
			decoded, err := Decode(encoded)
			if err != nil { t.Fatalf("decode encoded vector: %v", err) }
			if !bytes.Equal(decoded.Data, payload) { t.Fatalf("decoded payload = %x, want %x", decoded.Data, payload) }
		})
	}
}

func TestDecodeRejectsCorruptedPayloadBlocks(t *testing.T) {
	payload := make([]byte, 33)
	for i := range payload { payload[i] = byte(i) }
	encoded, err := Encode(&Frame{Control: Control{DIR: true, PRM: true, FuncCode: FuncUnconfirmedUserData}, DestAddr: 4, SrcAddr: 3, Data: payload})
	if err != nil { t.Fatalf("encode: %v", err) }
	for block := 0; block < 3; block++ {
		corrupt := append([]byte(nil), encoded...)
		crcOffset := 10 + block*18 + 16
		if block == 2 { crcOffset = 10 + 2*18 + 1 }
		corrupt[crcOffset] ^= 0x01
		if _, err := Decode(corrupt); err == nil { t.Fatalf("Decode accepted corrupted payload block %d", block) }
	}
}

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
			want:    0xF3, // 1111 0011: DIR=1, PRM=1, FCB=1, FCV=1, FuncCode=3
		},
		{
			name:    "confirmed data no FCB",
			control: Control{DIR: true, PRM: true, FCB: false, FCV: true, FuncCode: FuncConfirmedUserData},
			want:    0xD3, // 1101 0011: DIR=1, PRM=1, FCB=0, FCV=1, FuncCode=3
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

func TestSecondaryFunctionCodes(t *testing.T) {
	tests := []struct {
		name string
		code uint8
		want uint8
	}{
		{"ack", FuncAck, 0},
		{"nack", FuncNack, 1},
		{"link status", FuncLinkStatus, 2},
		{"not supported", FuncNotSupported, 3},
		{"confirmed user data response", FuncConfirmedUserDataR, 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.code != tt.want {
				t.Fatalf("function code = %d, want %d", tt.code, tt.want)
			}
		})
	}
}

func TestAddressByteOrder(t *testing.T) {
	f := &Frame{
		Control:  Control{DIR: true, PRM: true, FuncCode: FuncResetLinkStations},
		DestAddr: 0x0004,
		SrcAddr:  0x0003,
	}
	encoded, err := Encode(f)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	if got, want := encoded[4:8], []byte{0x04, 0x00, 0x03, 0x00}; !bytes.Equal(got, want) {
		t.Fatalf("address bytes = %x, want %x", got, want)
	}

	// Decode the locally encoded frame to verify both address directions.
	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if decoded.DestAddr != f.DestAddr || decoded.SrcAddr != f.SrcAddr {
		t.Fatalf("decoded addresses = %#04x -> %#04x, want %#04x -> %#04x", decoded.SrcAddr, decoded.DestAddr, f.SrcAddr, f.DestAddr)
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
