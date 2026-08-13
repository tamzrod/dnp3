package outstation

import (
	"testing"

	"dnp3/internal/al"
)

// MEXT-022 — IEEE 1815 "variation 0 = default variation" for static groups.
//
// The master's MEXT-015 multi-group integrity read requests the default
// variation per group (G1V0 / G20V0 / G30V0). The v0 outstation's default
// static variation is V1, so a V0 request must be served as V1 (the well-formed
// packed/count8 encoding) rather than emitting a malformed response. This
// locks the normalizeStaticVariation behavior independent of the real-TCP
// integration test.

// TestReadStaticVariationZeroServedAsDefault asserts that a READ of any
// supported static group with variation 0 returns the same well-formed object
// data as variation 1 (the default).
func TestReadStaticVariationZeroServedAsDefault(t *testing.T) {
	ost := NewOutstation(nil)
	ost.Initialize()
	ost.data = NewDefaultDataHandler() // pre-populated with binary/analog/counter points

	cases := []struct {
		name  string
		group uint8
	}{
		{"binary_input_G1", 1},
		{"counter_G20", 20},
		{"analog_input_G30", 30},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reqV0 := &al.APDU{
				Control:  al.AppControl{FIR: true, FIN: true, Seq: 1},
				FuncCode: al.FuncRead,
				Data:     []byte{tc.group, 0, 0x06, 0x00}, // V0 (default), all-objects
			}
			reqV1 := &al.APDU{
				Control:  al.AppControl{FIR: true, FIN: true, Seq: 1},
				FuncCode: al.FuncRead,
				Data:     []byte{tc.group, 1, 0x06, 0x00}, // V1, all-objects
			}
			respV0, err := ost.ProcessRequest(reqV0)
			if err != nil {
				t.Fatalf("ProcessRequest V0: %v", err)
			}
			respV1, err := ost.ProcessRequest(reqV1)
			if err != nil {
				t.Fatalf("ProcessRequest V1: %v", err)
			}
			// The object data (after the 2-byte IIN) must be identical: V0 is
			// served as the default (V1).
			dataV0 := respV0.Data[2:]
			dataV1 := respV1.Data[2:]
			if len(dataV0) == 0 {
				t.Fatalf("%s: V0 returned no object data (expected default V1 data)", tc.name)
			}
			if !bytesEqual(dataV0, dataV1) {
				t.Fatalf("%s: V0 object data = % X, want V1 default % X", tc.name, dataV0, dataV1)
			}
		})
	}
}

// TestNormalizeStaticVariation is a table test for the helper directly,
// pinning V0→V1 for static groups and passthrough otherwise.
func TestNormalizeStaticVariation(t *testing.T) {
	cases := []struct {
		group, in, want uint8
	}{
		{1, 0, 1}, {10, 0, 1}, {20, 0, 1}, {30, 0, 1}, {40, 0, 1},
		{1, 1, 1}, {1, 2, 2}, {30, 5, 5},
		{2, 0, 0}, {21, 0, 0}, {31, 0, 0}, // events not normalized
		{60, 0, 0},
	}
	for _, tc := range cases {
		if got := normalizeStaticVariation(tc.group, tc.in); got != tc.want {
			t.Errorf("normalizeStaticVariation(%d,%d) = %d, want %d", tc.group, tc.in, got, tc.want)
		}
	}
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
