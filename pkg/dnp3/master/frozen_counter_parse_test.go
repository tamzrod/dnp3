package master

import (
	"testing"
)

// Independent wire-vector tests for parsing Group 21 (Frozen Counter) static
// responses. These vectors are written by hand from the IEEE 1815 object
// layouts (flags octet first, little-endian values), not produced by this
// library's outstation encoder.

// TestParseFrozenCounterG21V1 verifies flags-first little-endian decoding with
// a 16-bit range qualifier and a non-zero index base.
func TestParseFrozenCounterG21V1(t *testing.T) {
	// Group 21 V1, range 5..6, two points: 0x01020304 (ONLINE) and 0xAABBCCDD
	// (ONLINE|ROLLOVER = 0x09).
	data := []byte{
		0x15, 0x01, 0x28, 0x05, 0x00, 0x06, 0x00,
		0x01, 0x04, 0x03, 0x02, 0x01,
		0x09, 0xDD, 0xCC, 0xBB, 0xAA,
	}
	got := parseCounters(data)
	if len(got) != 2 {
		t.Fatalf("parsed %d counters, want 2", len(got))
	}
	if got[0].Index != 5 || got[0].Value != 0x01020304 || got[0].Quality != 0x01 {
		t.Fatalf("point 0 = %+v", got[0])
	}
	if got[1].Index != 6 || got[1].Value != 0xAABBCCDD || got[1].Quality != 0x09 {
		t.Fatalf("point 1 = %+v", got[1])
	}
	for _, c := range got {
		if c.Time != nil {
			t.Fatalf("G21V1 must not carry a timestamp, got %v", c.Time)
		}
	}
}

// TestParseFrozenCounterG21V5 verifies the 48-bit time-tagged variation decodes
// the 6-byte little-endian DNP time to the documented 1970-epoch millisecond
// value.
func TestParseFrozenCounterG21V5(t *testing.T) {
	// Independent vector: 2026-11-19T11:02:56.910Z == 1795086176910 ms.
	data := []byte{
		0x15, 0x05, 0x07, 0x01,
		0x01, 0x07, 0x00, 0x00, 0x00,
		0x8E, 0x52, 0x79, 0xF3, 0xA1, 0x01,
	}
	got := parseCounters(data)
	if len(got) != 1 {
		t.Fatalf("parsed %d counters, want 1", len(got))
	}
	c := got[0]
	if c.Value != 7 || c.Quality != 0x01 {
		t.Fatalf("point = %+v", c)
	}
	if c.Time == nil || c.Time.Value != 1795086176910 {
		t.Fatalf("timestamp = %v, want 1795086176910", c.Time)
	}
}
