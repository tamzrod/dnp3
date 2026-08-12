package master

import (
	"testing"

	"dnp3/pkg/dnp3/types"
)

func TestParseBinaryInputOnlineTrue(t *testing.T) {
	// g1v1 format: state in bit 7 (0x80), quality in bits 0-6
	// When value=true and quality=ONLINE, byte = 0x80 | 0x01 = 0x81
	data := []byte{
		0x01, // Group 1
		0x01, // Variation 1 (with flags)
		0x00, // Qualifier: index
		0x01, // Count: 1
		0x05, 0x00, // Index: 5 (little-endian)
		0x81, // Value: state=1 (0x80), quality=ONLINE (0x01)
	}

	result := parseBinaryInputs(data)

	if len(result) != 1 {
		t.Fatalf("Expected 1 binary input, got %d", len(result))
	}

	bi := result[0]
	if bi.Index != 5 {
		t.Errorf("Index = %d, want 5", bi.Index)
	}
	if bi.Value != true {
		t.Errorf("Value = %v, want true", bi.Value)
	}
	if bi.Quality&types.QualityOnline == 0 {
		t.Errorf("Quality = %v, want ONLINE bit set", bi.Quality)
	}
	if !bi.Quality.IsGood() {
		t.Errorf("IsGood() = false, want true")
	}
}

func TestParseClass0PackedBinaryInputVector(t *testing.T) {
	// External Group 1 Variation 1 vector: qualifier 0x07 (count8), one
	// packed point with state bit set. Variation 1 is packed per the external
	// device-profile references recorded in active_work/testdata.
	data := []byte{0x01, 0x01, 0x07, 0x01, 0x01}
	result := parseBinaryInputs(data)
	if len(result) != 1 {
		t.Fatalf("expected one packed binary input, got %d", len(result))
	}
	if result[0].Index != 0 || !result[0].Value {
		t.Fatalf("decoded point = index %d value %v, want index 0 value true", result[0].Index, result[0].Value)
	}
	if result[0].Quality&types.QualityOnline == 0 {
		t.Fatalf("packed point quality = %v, want ONLINE default", result[0].Quality)
	}
}

func TestParseBinaryInputOnlineFalse(t *testing.T) {
	// When value=false and quality=ONLINE, byte = 0x00 | 0x01 = 0x01
	data := []byte{
		0x01, // Group 1
		0x01, // Variation 1 (with flags)
		0x00, // Qualifier: index
		0x01, // Count: 1
		0x03, 0x00, // Index: 3
		0x01, // Value: state=0, quality=ONLINE (0x00|0x01)
	}

	result := parseBinaryInputs(data)

	if len(result) != 1 {
		t.Fatalf("Expected 1 binary input, got %d", len(result))
	}

	bi := result[0]
	if bi.Index != 3 {
		t.Errorf("Index = %d, want 3", bi.Index)
	}
	if bi.Value != false {
		t.Errorf("Value = %v, want false", bi.Value)
	}
	if bi.Quality&types.QualityOnline == 0 {
		t.Errorf("Quality = %v, want ONLINE bit set", bi.Quality)
	}
	if !bi.Quality.IsGood() {
		t.Errorf("IsGood() = false, want true")
	}
}

func TestParseBinaryInputOffline(t *testing.T) {
	// When value=true and quality=OFFLINE, byte = 0x80 | 0x80 = 0x80
	data := []byte{
		0x01, // Group 1
		0x01, // Variation 1 (with flags)
		0x00, // Qualifier: index
		0x01, // Count: 1
		0x01, 0x00, // Index: 1
		0x80, // Value: state=1 (0x80), quality=OFFLINE (0x80) - note: OFFLINE in bit 7 conflicts with state, but parsing still works
	}

	result := parseBinaryInputs(data)

	if len(result) != 1 {
		t.Fatalf("Expected 1 binary input, got %d", len(result))
	}

	bi := result[0]
	if bi.Index != 1 {
		t.Errorf("Index = %d, want 1", bi.Index)
	}
	if bi.Value != true {
		t.Errorf("Value = %v, want true", bi.Value)
	}
	// OFFLINE is still in the quality flags (bits 5-7)
	if bi.Quality&types.QualityOnline != 0 {
		t.Errorf("Quality should not have ONLINE, got %v", bi.Quality)
	}
	if bi.Quality.IsGood() {
		t.Errorf("IsGood() = true, want false for OFFLINE")
	}
}

func TestParseBinaryInputMultiple(t *testing.T) {
	// Test multiple binary inputs in one response
	// State in bit 7 (0x80), ONLINE in bit 0 (0x01)
	data := []byte{
		0x01, // Group 1
		0x01, // Variation 1
		0x00, // Qualifier: index
		0x03, // Count: 3
		0x00, 0x00, // Index 0
		0x81, // state=1, ONLINE
		0x01, 0x00, // Index 1
		0x01, // state=0, ONLINE
		0x02, 0x00, // Index 2
		0x81, // state=1, ONLINE
	}

	result := parseBinaryInputs(data)

	if len(result) != 3 {
		t.Fatalf("Expected 3 binary inputs, got %d", len(result))
	}

	expected := []struct {
		index uint16
		value bool
	}{
		{0, true},
		{1, false},
		{2, true},
	}

	for i, exp := range expected {
		if result[i].Index != exp.index {
			t.Errorf("Input %d: Index = %d, want %d", i, result[i].Index, exp.index)
		}
		if result[i].Value != exp.value {
			t.Errorf("Input %d: Value = %v, want %v", i, result[i].Value, exp.value)
		}
		if result[i].Quality&types.QualityOnline == 0 {
			t.Errorf("Input %d: Quality missing ONLINE", i)
		}
		if !result[i].Quality.IsGood() {
			t.Errorf("Input %d: IsGood() = false, want true", i)
		}
	}
}

func TestBinaryInputEncodeDecodeRoundTrip(t *testing.T) {
	// Test that encode → decode preserves value and quality
	testCases := []struct {
		name       string
		value      bool
		quality    types.QualityFlags
		wantValue  bool
		wantGood   bool
	}{
		{"true, ONLINE", true, types.QualityOnline, true, true},
		{"false, ONLINE", false, types.QualityOnline, false, true},
		{"true, ONLINE|RESTART", true, types.QualityOnline | types.QualityRestart, true, true},
		{"true, COMM_LOST", true, types.QualityCommLost, true, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Simulate outstation encoding (state in bit 7, quality in bits 0-6)
			var encoded byte
			if tc.value {
				encoded = 0x80 // state bit
			}
			encoded |= byte(tc.quality)

			// Parse as master would
			data := []byte{
				0x01, 0x01, 0x00, 0x01, // Header: group 1, var 1, qualifier, count 1
				0x00, 0x00, // Index 0
				encoded,
			}

			result := parseBinaryInputs(data)
			if len(result) != 1 {
				t.Fatalf("Expected 1 result, got %d", len(result))
			}

			if result[0].Value != tc.wantValue {
				t.Errorf("Value = %v, want %v", result[0].Value, tc.wantValue)
			}
			if result[0].Quality.IsGood() != tc.wantGood {
				t.Errorf("IsGood() = %v, want %v", result[0].Quality.IsGood(), tc.wantGood)
			}
		})
	}
}

// TestG1V1PackedBoundary locks the IEEE 1815 "Binary Input - Packed Format"
// (Group 1 Variation 1) decode across byte-boundary and qualifier cases.
// Points are packed 8 per byte (bit 0 of byte 0 = point 0); no per-point
// quality byte is present, so parsed quality is always ONLINE.
func TestG1V1PackedBoundary(t *testing.T) {
	cases := []struct {
		name string
		data []byte
		want []struct {
			index uint16
			value bool
		}
	}{
		{
			name: "count8 single point ON",
			// 01 01 07 01 | 01
			data: []byte{0x01, 0x01, 0x07, 0x01, 0x01},
			want: []struct {
				index uint16
				value bool
			}{{0, true}},
		},
		{
			name: "count8 eight points one full byte",
			// 0xAA = binary 10101010, LSB-first: bit0=0,bit1=1,...,bit7=1
			// -> 0=F,1=T,2=F,3=T,4=F,5=T,6=F,7=T
			data: []byte{0x01, 0x01, 0x07, 0x08, 0xAA},
			want: []struct {
				index uint16
				value bool
			}{
				{0, false}, {1, true}, {2, false}, {3, true},
				{4, false}, {5, true}, {6, false}, {7, true},
			},
		},
		{
			name: "count8 nine points cross byte boundary",
			// byte0=0xFF (points 0..7 ON), byte1=0x01 (point 8 ON)
			data: []byte{0x01, 0x01, 0x07, 0x09, 0xFF, 0x01},
			want: []struct {
				index uint16
				value bool
			}{
				{0, true}, {1, true}, {2, true}, {3, true},
				{4, true}, {5, true}, {6, true}, {7, true}, {8, true},
			},
		},
		{
			name: "count8 sixteen points two bytes",
			// 0x55 = 01010101, LSB-first -> 0=T,1=F,...,7=T
			// 0xAA = 10101010, LSB-first -> 8=F,9=T,...,15=T
			data: []byte{0x01, 0x01, 0x07, 0x10, 0x55, 0xAA},
			want: []struct {
				index uint16
				value bool
			}{
				{0, true}, {1, false}, {2, true}, {3, false},
				{4, true}, {5, false}, {6, true}, {7, false},
				{8, false}, {9, true}, {10, false}, {11, true},
				{12, false}, {13, true}, {14, false}, {15, true},
			},
		},
		{
			name: "range16 base 5 three points",
			// 01 01 28 | start=05 00 | stop=07 00 | packed=05 (bits 0,2 set)
			// points idx5=T, idx6=F, idx7=T
			data: []byte{0x01, 0x01, 0x28, 0x05, 0x00, 0x07, 0x00, 0x05},
			want: []struct {
				index uint16
				value bool
			}{{5, true}, {6, false}, {7, true}},
		},
		{
			name: "range16 base 0 sixteen points all ON",
			// 01 01 28 | start=00 00 | stop=0F 00 | FF FF
			data: []byte{0x01, 0x01, 0x28, 0x00, 0x00, 0x0F, 0x00, 0xFF, 0xFF},
			want: []struct {
				index uint16
				value bool
			}{
				{0, true}, {1, true}, {2, true}, {3, true},
				{4, true}, {5, true}, {6, true}, {7, true},
				{8, true}, {9, true}, {10, true}, {11, true},
				{12, true}, {13, true}, {14, true}, {15, true},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := parseBinaryInputs(tc.data)
			if len(got) != len(tc.want) {
				t.Fatalf("got %d points, want %d", len(got), len(tc.want))
			}
			for i := range tc.want {
				if got[i].Index != tc.want[i].index {
					t.Errorf("point %d: Index = %d, want %d", i, got[i].Index, tc.want[i].index)
				}
				if got[i].Value != tc.want[i].value {
					t.Errorf("point %d (idx %d): Value = %v, want %v", i, got[i].Index, got[i].Value, tc.want[i].value)
				}
				// Packed format carries no flags; master must default to ONLINE.
				if got[i].Quality&types.QualityOnline == 0 {
					t.Errorf("point %d (idx %d): Quality %v missing ONLINE", i, got[i].Index, got[i].Quality)
				}
			}
		})
	}
}

// TestG1V1PackedBitOrder verifies the LSB-first bit ordering within a byte:
// bit 0 of byte 0 is point 0, not bit 7.
func TestG1V1PackedBitOrder(t *testing.T) {
	// count8=1, packed byte 0x01 -> only bit 0 set -> point 0 ON.
	got := parseBinaryInputs([]byte{0x01, 0x01, 0x07, 0x01, 0x01})
	if len(got) != 1 || got[0].Index != 0 || !got[0].Value {
		t.Fatalf("0x01 must decode to point 0=true, got %+v", got)
	}
	// count8=8, packed byte 0x80 -> only bit 7 set -> point 7 ON, rest OFF.
	got = parseBinaryInputs([]byte{0x01, 0x01, 0x07, 0x08, 0x80})
	if len(got) != 8 {
		t.Fatalf("got %d points, want 8", len(got))
	}
	for i, p := range got {
		want := i == 7
		if p.Value != want {
			t.Errorf("point %d: Value = %v, want %v", i, p.Value, want)
		}
	}
}

// TestG1V1PackedMultipleHeaders verifies two packed G1V1 headers in one
// response decode independently with correct index bases.
func TestG1V1PackedMultipleHeaders(t *testing.T) {
	// Header 1: count8=1, byte 0x01 -> idx0=true
	// Header 2: range16 start=10 stop=10, byte 0x01 -> idx10=true
	data := []byte{
		0x01, 0x01, 0x07, 0x01, 0x01,
		0x01, 0x01, 0x28, 0x0A, 0x00, 0x0A, 0x00, 0x01,
	}
	got := parseBinaryInputs(data)
	if len(got) != 2 {
		t.Fatalf("got %d points, want 2", len(got))
	}
	if got[0].Index != 0 || !got[0].Value {
		t.Errorf("point 0: idx=%d value=%v, want idx=0 value=true", got[0].Index, got[0].Value)
	}
	if got[1].Index != 10 || !got[1].Value {
		t.Errorf("point 1: idx=%d value=%v, want idx=10 value=true", got[1].Index, got[1].Value)
	}
}
