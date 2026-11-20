package master

import (
	"testing"
)

// Independent wire vectors for Group 40 (Analog Output Status) parsing. They
// are written by hand from the IEEE 1815 definition (flags first, then
// LSB-first value) and are not produced by the library's own encoder.

func TestParseAnalogOutputG40V1FlagsFirst(t *testing.T) {
	// Group 40 V1, range 5..6, two points: 123.456 (ONLINE) and -0.5
	// (ONLINE|0x08). Independent float32 LE bytes from the golden outstation
	// vector, but assembled here by hand with a non-zero index base.
	data := []byte{
		0x28, 0x01, 0x28, 0x05, 0x00, 0x06, 0x00,
		0x01, 0x79, 0xE9, 0xF6, 0x42,
		0x09, 0x00, 0x00, 0x00, 0xBF,
	}
	got := parseAnalogOutputs(data)
	if len(got) != 2 {
		t.Fatalf("parsed %d points, want 2", len(got))
	}
	if got[0].Index != 5 || got[0].Quality != 0x01 || got[0].Value < 123.45 || got[0].Value > 123.47 {
		t.Fatalf("point 0 = %+v", got[0])
	}
	if got[1].Index != 6 || got[1].Quality != 0x09 || got[1].Value != -0.5 {
		t.Fatalf("point 1 = %+v", got[1])
	}
}

func TestParseAnalogOutputG40V2FlagsFirst(t *testing.T) {
	// Group 40 V2, count8, -2 (0xFFFE LE) and 1000 (0x03E8 LE).
	data := []byte{
		0x28, 0x02, 0x07, 0x02,
		0x01, 0xFE, 0xFF,
		0x09, 0xE8, 0x03,
	}
	got := parseAnalogOutputs(data)
	if len(got) != 2 {
		t.Fatalf("parsed %d points, want 2", len(got))
	}
	if got[0].Value != -2 || got[0].Quality != 0x01 {
		t.Fatalf("point 0 = %+v", got[0])
	}
	if got[1].Value != 1000 || got[1].Quality != 0x09 {
		t.Fatalf("point 1 = %+v", got[1])
	}
}

// TestParseAnalogOutputRejectsShortVector proves a truncated G40V1 object is
// not silently accepted as a partial setpoint readback.
func TestParseAnalogOutputRejectsShortVector(t *testing.T) {
	data := []byte{0x28, 0x01, 0x07, 0x02, 0x01, 0x79, 0xE9, 0xF6, 0x42}
	if got := parseAnalogOutputs(data); len(got) != 0 {
		t.Fatalf("truncated vector produced %d points: %+v", len(got), got)
	}
}
