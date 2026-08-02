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
		0x00, 0x05, // Index: 5 (big-endian)
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

func TestParseBinaryInputOnlineFalse(t *testing.T) {
	// When value=false and quality=ONLINE, byte = 0x00 | 0x01 = 0x01
	data := []byte{
		0x01, // Group 1
		0x01, // Variation 1 (with flags)
		0x00, // Qualifier: index
		0x01, // Count: 1
		0x00, 0x03, // Index: 3
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
		0x00, 0x01, // Index: 1
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
		0x00, 0x01, // Index 1
		0x01, // state=0, ONLINE
		0x00, 0x02, // Index 2
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
