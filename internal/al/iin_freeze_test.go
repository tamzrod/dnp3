package al

import (
	"testing"
)

// MEXT-019 — IIN bit-map freeze vs IEEE 1815.
//
// The IIN field is defined in application.go as a frozen 2-octet table
// (IIN1 then IIN2, MSB-first within each octet). These tests lock the table
// against semantic drift before the external interop claim: named critical
// masks are pinned to their verified [2]byte, every possible 16-bit mask
// round-trips losslessly through Bytes/SetIIN, and the reserved IIN2 bits are
// characterized. If any of these breaks, the IIN table drifted.

// TestIINKnownMasksFreeze pins named critical IIN masks to their frozen
// [IIN1, IIN2] byte values (IEEE 1815-2012). Adding, removing, or reordering
// a bit will break the named entry it affects.
func TestIINKnownMasksFreeze(t *testing.T) {
	tests := []struct {
		name string
		iin  IIN
		want [2]byte
	}{
		{"all clear", IIN{}, [2]byte{0x00, 0x00}},
		{"NeedTime", IIN{NeedTime: true}, [2]byte{0x08, 0x00}},
		{"DeviceRestart", IIN{DeviceRestart: true}, [2]byte{0x01, 0x00}},
		{"DeviceTrouble", IIN{DeviceTrouble: true}, [2]byte{0x02, 0x00}},
		{"AllStations+DeviceRestart", IIN{AllStations: true, DeviceRestart: true}, [2]byte{0x81, 0x00}},
		{"Class1+2+3 events", IIN{Class1Events: true, Class2Events: true, Class3Events: true}, [2]byte{0x70, 0x00}},
		{"FuncUnknown", IIN{FuncUnknown: true}, [2]byte{0x00, 0x80}},
		{"ObjectUnknown", IIN{ObjectUnknown: true}, [2]byte{0x00, 0x40}},
		{"ParameterError", IIN{ParameterError: true}, [2]byte{0x00, 0x20}},
		{"all IIN2 errors", IIN{
			FuncUnknown: true, ObjectUnknown: true, ParameterError: true,
			BufferOverflow: true, AlreadyExecuting: true, BadConfig: true,
		}, [2]byte{0x00, 0xFC}},
		{"command rejected (ParameterError)", IIN{ParameterError: true}, [2]byte{0x00, 0x20}},
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
	for _, mask := range [][2]byte{{0x00, 0x02}, {0x00, 0x01}, {0x00, 0x03}} {
		var iin IIN
		iin.SetIIN(mask)
		got := iin.Bytes()
		if got != mask {
			t.Errorf("reserved mask [%02X %02X]: round-trip = [%02X %02X]",
				mask[0], mask[1], got[0], got[1])
		}
		if mask[1]&0x02 != 0 && !iin.Reserved2_6 {
			t.Errorf("reserved mask % X: Reserved2_6 not set", mask)
		}
		if mask[1]&0x01 != 0 && !iin.Reserved2_7 {
			t.Errorf("reserved mask % X: Reserved2_7 not set", mask)
		}
	}
}
