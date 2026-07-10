package crc

import (
	"testing"
)

// Test vectors from IEEE 1815-2012 Annex B
var testVectors = []struct {
	data   []byte
	expect uint16
}{
	// Empty data
	{[]byte{}, 0xFFFF},

	// Single byte: 0x01
	{[]byte{0x01}, 0x00D9},

	// Two bytes: 0x01 0x02
	{[]byte{0x01, 0x02}, 0x5B02},

	// Three bytes: 0x01 0x02 0x03
	{[]byte{0x01, 0x02, 0x03}, 0x8671},

	// Known pattern
	{[]byte{0x05, 0x64, 0x04, 0x00, 0xFF, 0xFF}, 0x36D9},
}

func TestCRC16(t *testing.T) {
	for _, tv := range testVectors {
		got := CRC16(tv.data)
		if got != tv.expect {
			t.Errorf("CRC16(%x) = 0x%04X, want 0x%04X", tv.data, got, tv.expect)
		}
	}
}

func TestCRC16Quick(t *testing.T) {
	// CRC16 and CRC16Quick should produce the same results
	for _, tv := range testVectors {
		got := CRC16Quick(tv.data)
		if got != tv.expect {
			t.Errorf("CRC16Quick(%x) = 0x%04X, want 0x%04X", tv.data, got, tv.expect)
		}
	}
}

func TestCRC16Consistency(t *testing.T) {
	// Both implementations should be identical
	testData := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}

	crc1 := CRC16(testData)
	crc2 := CRC16Quick(testData)

	if crc1 != crc2 {
		t.Errorf("CRC16 and CRC16Quick produced different results: 0x%04X vs 0x%04X", crc1, crc2)
	}
}

func TestValidateCRC(t *testing.T) {
	tests := []struct {
		name      string
		data      []byte
		crcOffset int
		wantValid bool
	}{
		{
			name:      "valid CRC",
			data:      []byte{0x01, 0x02, 0x03, 0x04, 0x71, 0x86}, // 0x8671 is CRC for 0x01,0x02,0x03,0x04
			crcOffset: 4,
			wantValid:  true,
		},
		{
			name:      "invalid CRC",
			data:      []byte{0x01, 0x02, 0x03, 0x04, 0x00, 0x00},
			crcOffset: 4,
			wantValid: false,
		},
		{
			name:      "too short",
			data:      []byte{0x01},
			crcOffset: 0,
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateCRC(tt.data, tt.crcOffset)
			if got != tt.wantValid {
				t.Errorf("ValidateCRC() = %v, want %v", got, tt.wantValid)
			}
		})
	}
}

// BenchmarkCRC16 benchmarks the CRC16 function.
func BenchmarkCRC16(b *testing.B) {
	data := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CRC16(data)
	}
}

// BenchmarkCRC16Quick benchmarks the CRC16Quick function.
func BenchmarkCRC16Quick(b *testing.B) {
	data := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CRC16Quick(data)
	}
}
