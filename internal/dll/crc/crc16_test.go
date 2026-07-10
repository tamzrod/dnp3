package crc

import (
	"testing"
)

// Test vectors for CRC-16/DNP
//
// Authority: CRC Catalogue (reveng.sourceforge.io)
//   CRC-16/DNP: poly=0x3D65, init=0x0000, refin=true, refout=true, xorout=0xFFFF
//   Canonical check value: CRC("123456789") = 0xEA82
//
// Verification sources:
//   - CRC Catalogue: Industry-standard CRC reference parameters
//   - IvanGaravito/dnp3-crc: Independent DNP3 CRC implementation (GitHub)
//   - Validated against canonical CRC-16/DNP check value
//
// Previous test vectors claimed to be from "IEEE 1815-2012 Annex B" but
// contained incorrect expected values that did not match CRC-16/DNP.
var testVectors = []struct {
	data   []byte
	expect uint16
}{
	// Empty data
	{[]byte{}, 0xFFFF},

	// Single byte: 0x01
	{[]byte{0x01}, 0xC9A1},

	// Two bytes: 0x01 0x02
	{[]byte{0x01, 0x02}, 0x380D},

	// Three bytes: 0x01 0x02 0x03
	{[]byte{0x01, 0x02, 0x03}, 0xA740},

	// Known pattern (DNP3-like sync bytes)
	{[]byte{0x05, 0x64, 0x04, 0x00, 0xFF, 0xFF}, 0x1967},
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
	// CRC for [0x01, 0x02, 0x03, 0x04] using validated CRC-16/DNP
	validCRC := CRC16([]byte{0x01, 0x02, 0x03, 0x04})

	tests := []struct {
		name      string
		data      []byte
		crcOffset int
		wantValid bool
	}{
		{
			name:      "valid CRC",
			data:      []byte{0x01, 0x02, 0x03, 0x04, byte(validCRC), byte(validCRC >> 8)},
			crcOffset: 4,
			wantValid: true,
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
