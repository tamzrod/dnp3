package master

import (
	"math"
	"testing"

	"dnp3/pkg/dnp3/types"
)

// TestG30V1Count8Boundary locks Group 30 Variation 1 (signed 32-bit analog
// input with a 1-octet flags byte, 5 octets per point) for the count8
// qualifier: sequential points from index 0, value little-endian.
func TestG30V1Count8Boundary(t *testing.T) {
	// Two points: idx0=1000 ONLINE, idx1=-1 ONLINE.
	// 1000 = E8 03 00 00 ; -1 = FF FF FF FF
	data := []byte{
		0x1E, 0x01, 0x07, 0x02,
		0xE8, 0x03, 0x00, 0x00, 0x01,
		0xFF, 0xFF, 0xFF, 0xFF, 0x01,
	}
	got := parseAnalogInputs(data)
	if len(got) != 2 {
		t.Fatalf("got %d points, want 2", len(got))
	}
	if got[0].Index != 0 || got[0].Value != 1000 || got[0].Quality&types.QualityOnline == 0 {
		t.Errorf("point 0: idx=%d val=%v q=%v, want idx0 val1000 ONLINE", got[0].Index, got[0].Value, got[0].Quality)
	}
	if got[1].Index != 1 || got[1].Value != -1 || got[1].Quality&types.QualityOnline == 0 {
		t.Errorf("point 1: idx=%d val=%v q=%v, want idx1 val-1 ONLINE", got[1].Index, got[1].Value, got[1].Quality)
	}
}

// TestG30V1SignedRange exercises negative values across the int32 range to
// confirm G30V1 is decoded as a signed 32-bit value (not unsigned).
func TestG30V1SignedRange(t *testing.T) {
	cases := []struct {
		name  string
		bytes []byte // 4 value bytes, LSB first
		want  float64
	}{
		{"zero", []byte{0x00, 0x00, 0x00, 0x00}, 0},
		{"max int32", []byte{0xFF, 0xFF, 0xFF, 0x7F}, math.MaxInt32},
		{"min int32", []byte{0x00, 0x00, 0x00, 0x80}, math.MinInt32},
		{"-1", []byte{0xFF, 0xFF, 0xFF, 0xFF}, -1},
		{"-1000", []byte{0x18, 0xFC, 0xFF, 0xFF}, -1000},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data := []byte{0x1E, 0x01, 0x07, 0x01}
			data = append(data, tc.bytes...)
			data = append(data, 0x01) // ONLINE
			got := parseAnalogInputs(data)
			if len(got) != 1 {
				t.Fatalf("got %d points, want 1", len(got))
			}
			if got[0].Value != tc.want {
				t.Errorf("Value = %v, want %v", got[0].Value, tc.want)
			}
		})
	}
}

// TestG30V1Range16Boundary locks the range16 (0x28) qualifier for G30V1:
// sequential points indexed from Start, no per-point index prefix.
func TestG30V1Range16Boundary(t *testing.T) {
	// start=5 stop=6, two points: idx5=1000, idx6=2000, both ONLINE.
	// 1000 = E8 03 00 00 ; 2000 = D0 07 00 00
	data := []byte{
		0x1E, 0x01, 0x28, 0x05, 0x00, 0x06, 0x00,
		0xE8, 0x03, 0x00, 0x00, 0x01,
		0xD0, 0x07, 0x00, 0x00, 0x01,
	}
	got := parseAnalogInputs(data)
	if len(got) != 2 {
		t.Fatalf("got %d points, want 2", len(got))
	}
	if got[0].Index != 5 || got[0].Value != 1000 {
		t.Errorf("point 0: idx=%d val=%v, want idx5 val1000", got[0].Index, got[0].Value)
	}
	if got[1].Index != 6 || got[1].Value != 2000 {
		t.Errorf("point 1: idx=%d val=%v, want idx6 val2000", got[1].Index, got[1].Value)
	}
}

// TestG30V1LSBByteOrder confirms the 32-bit value is decoded little-endian
// (the DNP3-001 BE-negative case, re-asserted for the locked variation).
func TestG30V1LSBByteOrder(t *testing.T) {
	// 1000 as E8 03 00 00 (LSB first). Big-endian read would be 0xE8030000.
	data := []byte{0x1E, 0x01, 0x07, 0x01, 0xE8, 0x03, 0x00, 0x00, 0x01}
	got := parseAnalogInputs(data)
	if len(got) != 1 || got[0].Value != 1000 {
		t.Fatalf("got %+v, want idx0 val1000", got)
	}
}

// TestG30V1QualityByte locks the per-point flags byte semantics for G30V1.
func TestG30V1QualityByte(t *testing.T) {
	cases := []struct {
		name string
		flag byte
		want types.QualityFlags
	}{
		{"online", 0x01, types.QualityOnline},
		{"restart", 0x40, types.QualityRestart},
		{"comm_lost", 0x80, types.QualityCommLost},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data := []byte{0x1E, 0x01, 0x07, 0x01, 0x00, 0x00, 0x00, 0x00, tc.flag}
			got := parseAnalogInputs(data)
			if len(got) != 1 {
				t.Fatalf("got %d points, want 1", len(got))
			}
			if got[0].Quality != tc.want {
				t.Errorf("Quality = %v, want %v", got[0].Quality, tc.want)
			}
		})
	}
}
