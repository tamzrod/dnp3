package al

import (
	"testing"
)

// EVT-001 — IIN wire masks vs pinned OpenDNP3 3.1.2 IINField.h.
//
// MEXT-019 numbered bit 0 as the MSB. That freeze is lifted for this
// correction. Bit 0 is now the least significant bit, matching OpenDNP3
// LSBMask/MSBMask at 26b4c01e. These tests lock the corrected table:
// named masks, a full 16-bit round-trip, and the two reserved octet-2 bits.

// TestIINKnownMasksFreeze pins named IIN masks to the OpenDNP3 wire bytes.
func TestIINKnownMasksFreeze(t *testing.T) {
	tests := []struct {
		name string
		iin  IIN
		want [2]byte
	}{
		{"all clear", IIN{}, [2]byte{0x00, 0x00}},
		{"NeedTime", IIN{NeedTime: true}, [2]byte{0x10, 0x00}},
		{"DeviceRestart", IIN{DeviceRestart: true}, [2]byte{0x80, 0x00}},
		{"DeviceTrouble", IIN{DeviceTrouble: true}, [2]byte{0x40, 0x00}},
		{"AllStations+DeviceRestart", IIN{AllStations: true, DeviceRestart: true}, [2]byte{0x81, 0x00}},
		{"Class1+2+3 events", IIN{Class1Events: true, Class2Events: true, Class3Events: true}, [2]byte{0x0E, 0x00}},
		{"FuncUnknown", IIN{FuncUnknown: true}, [2]byte{0x00, 0x01}},
		{"ObjectUnknown", IIN{ObjectUnknown: true}, [2]byte{0x00, 0x02}},
		{"ParameterError", IIN{ParameterError: true}, [2]byte{0x00, 0x04}},
		{"BufferOverflow", IIN{BufferOverflow: true}, [2]byte{0x00, 0x08}},
		{"all IIN2 errors", IIN{
			FuncUnknown: true, ObjectUnknown: true, ParameterError: true,
			BufferOverflow: true, AlreadyExecuting: true, BadConfig: true,
		}, [2]byte{0x00, 0x3F}},
		{"command rejected (ParameterError)", IIN{ParameterError: true}, [2]byte{0x00, 0x04}},
	}
	for _, tt := range tests {
		got := tt.iin.Bytes()
		if got != tt.want {
			t.Errorf("%s: Bytes = [%02X %02X], want [%02X %02X]",
				tt.name, got[0], got[1], tt.want[0], tt.want[1])
		}
		// Decode must round-trip back to the same bytes.
		var back IIN
		back.SetIIN(got)
		again := back.Bytes()
		if again != tt.want {
			t.Errorf("%s: decode->encode not stable: [%02X %02X], want [%02X %02X]",
				tt.name, again[0], again[1], tt.want[0], tt.want[1])
		}
	}
}

// TestIINRoundTripAllMasks proves Bytes and SetIIN are exact inverses for every
// 16-bit IIN mask: encode -> decode -> encode yields the original bytes. This
// freezes the entire table (no bit may be dropped, added, or moved).
func TestIINRoundTripAllMasks(t *testing.T) {
	for v := 0; v <= 0xFFFF; v++ {
		orig := [2]byte{byte(v >> 8), byte(v & 0xFF)}
		var iin IIN
		iin.SetIIN(orig)
		got := iin.Bytes()
		if got != orig {
			t.Fatalf("mask %04X: round-trip = [%02X %02X], want [%02X %02X]",
				v, got[0], got[1], orig[0], orig[1])
		}
	}
}

// TestIINDecodeIINAllMasks proves DecodeIIN/EncodeIIN are inverses for every
// 16-bit mask (the public slice-based API path).
func TestIINDecodeIINAllMasks(t *testing.T) {
	for v := 0; v <= 0xFFFF; v++ {
		orig := []byte{byte(v >> 8), byte(v & 0xFF)}
		iin, err := DecodeIIN(orig)
		if err != nil {
			t.Fatalf("mask %04X: DecodeIIN error: %v", v, err)
		}
		got := EncodeIIN(iin)
		if len(got) != 2 || got[0] != orig[0] || got[1] != orig[1] {
			t.Fatalf("mask %04X: EncodeIIN = % X, want % X", v, got, orig)
		}
	}
}

// TestIINReservedBitsRoundTrip characterizes the two reserved IIN2 bits. The
// table documents them as "reserved, always 0"; the encoder/decoder preserve
// whatever is in the wire bytes rather than masking them, so a reserved bit
// set on the wire round-trips. This locks the current behavior so a future
// change (e.g. forcing reserved bits to 0 on decode) is a deliberate, reviewed
// decision.
func TestIINReservedBitsRoundTrip(t *testing.T) {
	for _, mask := range [][2]byte{{0x00, 0x40}, {0x00, 0x80}, {0x00, 0xC0}} {
		var iin IIN
		iin.SetIIN(mask)
		got := iin.Bytes()
		if got != mask {
			t.Errorf("reserved mask [%02X %02X]: round-trip = [%02X %02X]",
				mask[0], mask[1], got[0], got[1])
		}
		if mask[1]&0x40 != 0 && !iin.Reserved2_6 {
			t.Errorf("reserved mask % X: Reserved2_6 not set", mask)
		}
		if mask[1]&0x80 != 0 && !iin.Reserved2_7 {
			t.Errorf("reserved mask % X: Reserved2_7 not set", mask)
		}
	}
}
