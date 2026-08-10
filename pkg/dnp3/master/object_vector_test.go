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
