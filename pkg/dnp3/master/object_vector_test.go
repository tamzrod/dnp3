package master

import (
	"testing"

	"dnp3/pkg/dnp3/types"
)

func TestParseClass0AnalogInputVector(t *testing.T) {
	data := []byte{0x1E, 0x01, 0x07, 0x01, 0xE8, 0x03, 0x00, 0x00, 0x01}
	result := parseAnalogInputs(data)
	if len(result) != 1 {
		t.Fatalf("expected one analog input, got %d", len(result))
	}
	if result[0].Index != 0 || result[0].Value != 1000 {
		t.Fatalf("decoded point = index %d value %v, want index 0 value 1000", result[0].Index, result[0].Value)
	}
	if result[0].Quality&types.QualityOnline == 0 {
		t.Fatalf("analog quality = %v, want ONLINE", result[0].Quality)
	}
}

func TestParseClass0CounterVector(t *testing.T) {
	data := []byte{0x14, 0x01, 0x07, 0x01, 0xE8, 0x03, 0x00, 0x00, 0x01}
	result := parseCounters(data)
	if len(result) != 1 { t.Fatalf("expected one counter, got %d", len(result)) }
	if result[0].Index != 0 || result[0].Value != 1000 { t.Fatalf("decoded point = index %d value %d, want index 0 value 1000", result[0].Index, result[0].Value) }
	if result[0].Quality&types.QualityOnline == 0 { t.Fatalf("counter quality = %v, want ONLINE", result[0].Quality) }
}

// TestParseAnalogInputLSBNotBigEndian verifies the G30V1 parser decodes the
// value field as little-endian. The golden vector encodes 1000 as E8 03 00 00
// (LSB first). A big-endian read of those bytes would yield 0xE8030000
// (3892314112), which is clearly wrong. This is the BE-negative case required
// by DNP3-001.
func TestParseAnalogInputLSBNotBigEndian(t *testing.T) {
	data := []byte{0x1E, 0x01, 0x07, 0x01, 0xE8, 0x03, 0x00, 0x00, 0x01}
	result := parseAnalogInputs(data)
	if len(result) != 1 {
		t.Fatalf("expected one analog input, got %d", len(result))
	}
	if result[0].Value != 1000 {
		t.Fatalf("analog value = %v, want 1000 (LSB decode of E8 03 00 00)", result[0].Value)
	}
}

// TestParseCounterLSBNotBigEndian verifies the G20V1 parser decodes the value
// field as little-endian, the BE-negative case for counters (DNP3-001).
func TestParseCounterLSBNotBigEndian(t *testing.T) {
	data := []byte{0x14, 0x01, 0x07, 0x01, 0xE8, 0x03, 0x00, 0x00, 0x01}
	result := parseCounters(data)
	if len(result) != 1 {
		t.Fatalf("expected one counter, got %d", len(result))
	}
	if result[0].Value != 1000 {
		t.Fatalf("counter value = %d, want 1000 (LSB decode of E8 03 00 00)", result[0].Value)
	}
}

// TestParseBinaryOutputIndexLSB verifies the G10 index field is decoded
// little-endian (DNP3-001 BE-negative case for binary output index).
func TestParseBinaryOutputIndexLSB(t *testing.T) {
	// Index 0x1234 encoded LSB-first as 0x34 0x12.
	data := []byte{
		0x0A, 0x01, 0x00, 0x01, // Group 10, Var 1, qualifier index, count 1
		0x34, 0x12, // Index 0x1234 (LSB first)
		0x81,       // state=1, ONLINE
	}
	result := parseBinaryOutputs(data)
	if len(result) != 1 {
		t.Fatalf("expected one binary output, got %d", len(result))
	}
	if result[0].Index != 0x1234 {
		t.Fatalf("binary output index = %d, want 0x1234 (LSB decode of 34 12)", result[0].Index)
	}
	if !result[0].Value {
		t.Fatalf("binary output value = %v, want true", result[0].Value)
	}
}

// TestParseAnalogOutputIndexLSB verifies the G40 index field is decoded
// little-endian (DNP3-001 BE-negative case for analog output index).
func TestParseAnalogOutputIndexLSB(t *testing.T) {
	// Index 0x1234 encoded LSB-first as 0x34 0x12; float 0.0 with ONLINE flags.
	data := []byte{
		0x28, 0x01, 0x00, 0x01, // Group 40, Var 1, qualifier index, count 1
		0x34, 0x12, // Index 0x1234 (LSB first)
		0x00, 0x00, 0x00, 0x00, // float bits = 0.0 (LSB first)
		0x01, // quality ONLINE
	}
	result := parseAnalogOutputs(data)
	if len(result) != 1 {
		t.Fatalf("expected one analog output, got %d", len(result))
	}
	if result[0].Index != 0x1234 {
		t.Fatalf("analog output index = %d, want 0x1234 (LSB decode of 34 12)", result[0].Index)
	}
}
