package master

import (
	"testing"

	"dnp3/pkg/dnp3/types"
)

// TestClass0G1V1StartStop8 decodes a packed G1V1 object whose qualifier is the
// IEEE 1815 UINT8_START_STOP code 0x00: header is group/var/qualifier/start(1)/
// stop(1), and the packed bits map to sequential indices Start..Stop.
func TestClass0G1V1StartStop8(t *testing.T) {
	// G1V1, qualifier 0x00, start=2, stop=4, packed byte 0b00000111 -> bits 0..2 set.
	data := []byte{0x01, 0x01, 0x00, 0x02, 0x04, 0x07}
	got := parseBinaryInputs(data)
	if len(got) != 3 {
		t.Fatalf("got %d points, want 3", len(got))
	}
	for i, want := range []uint16{2, 3, 4} {
		if got[i].Index != want {
			t.Errorf("point %d: Index = %d, want %d", i, got[i].Index, want)
		}
		if !got[i].Value {
			t.Errorf("point %d (idx %d): Value = false, want true", i, got[i].Index)
		}
		if got[i].Quality&types.QualityOnline == 0 {
			t.Errorf("point %d: Quality %v missing ONLINE", i, got[i].Quality)
		}
	}
}

// TestClass0G30V1StartStop8 decodes a G30V1 object with the 0x00 start/stop
// qualifier: sequential signed-32-bit points from Start..Stop, no index prefix.
func TestClass0G30V1StartStop8(t *testing.T) {
	data := []byte{
		0x1E, 0x01, 0x00, 0x05, 0x06,
		0xE8, 0x03, 0x00, 0x00, 0x01, // idx5 = 1000
		0xD0, 0x07, 0x00, 0x00, 0x01, // idx6 = 2000
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

// TestClass0CounterStartStop8 decodes a G20V1 object with the 0x00 start/stop
// qualifier: sequential uint32 counters from Start..Stop.
func TestClass0CounterStartStop8(t *testing.T) {
	data := []byte{
		0x14, 0x01, 0x00, 0x09, 0x0A,
		0x2A, 0x00, 0x00, 0x00, 0x01, // idx9 = 42
		0x63, 0x00, 0x00, 0x00, 0x01, // idx10 = 99
	}
	got := parseCounters(data)
	if len(got) != 2 {
		t.Fatalf("got %d points, want 2", len(got))
	}
	if got[0].Index != 9 || got[0].Value != 42 {
		t.Errorf("point 0: idx=%d val=%d, want idx9 val42", got[0].Index, got[0].Value)
	}
	if got[1].Index != 10 || got[1].Value != 99 {
		t.Errorf("point 1: idx=%d val=%d, want idx10 val99", got[1].Index, got[1].Value)
	}
}

// TestClass0G30V2IndexedCount8 decodes a G30V2 (16-bit signed + flags) object
// with the IEEE 1815 UINT8_CNT_UINT8_INDEX qualifier 0x17: an 8-bit count then
// a per-point 8-bit index prefix before each value.
func TestClass0G30V2IndexedCount8(t *testing.T) {
	data := []byte{
		0x1E, 0x02, 0x17, 0x02,
		0x03, 0xE8, 0x03, 0x01, // idx3 = 1000
		0x07, 0xD0, 0x07, 0x01, // idx7 = 2000
	}
	got := parseAnalogInputs(data)
	if len(got) != 2 {
		t.Fatalf("got %d points, want 2", len(got))
	}
	if got[0].Index != 3 || got[0].Value != 1000 {
		t.Errorf("point 0: idx=%d val=%v, want idx3 val1000", got[0].Index, got[0].Value)
	}
	if got[1].Index != 7 || got[1].Value != 2000 {
		t.Errorf("point 1: idx=%d val=%v, want idx7 val2000", got[1].Index, got[1].Value)
	}
}

// TestClass0StartStop8TruncatedStopsDecode verifies a truncated 0x00 header
// (only 4 octets, missing the stop byte) yields no points instead of a
// misdecoded point or a panic.
func TestClass0StartStop8TruncatedStopsDecode(t *testing.T) {
	if got := parseBinaryInputs([]byte{0x01, 0x01, 0x00, 0x02}); len(got) != 0 {
		t.Fatalf("truncated start/stop header decoded %d points, want 0", len(got))
	}
	if got := parseAnalogInputs([]byte{0x1E, 0x01, 0x00, 0x02}); len(got) != 0 {
		t.Fatalf("truncated start/stop header decoded %d points, want 0", len(got))
	}
}

// TestClass0IndexedCount8TruncatedPrefix verifies a 0x17 object that declares
// more points than the remaining bytes provide stops cleanly at the boundary.
func TestClass0IndexedCount8TruncatedPrefix(t *testing.T) {
	// Declares 2 points but supplies only one full (index+value) record.
	data := []byte{
		0x1E, 0x02, 0x17, 0x02,
		0x03, 0xE8, 0x03, 0x01,
	}
	got := parseAnalogInputs(data)
	if len(got) != 1 {
		t.Fatalf("got %d points, want 1 (truncated second point)", len(got))
	}
	if got[0].Index != 3 || got[0].Value != 1000 {
		t.Errorf("point 0: idx=%d val=%v, want idx3 val1000", got[0].Index, got[0].Value)
	}
}

// TestClass0StartStop8InvertedRejected verifies a 0x00 header with stop<start
// is rejected by the header decoder and therefore yields no points.
func TestClass0StartStop8InvertedRejected(t *testing.T) {
	if got := parseCounters([]byte{0x14, 0x01, 0x00, 0x05, 0x03}); len(got) != 0 {
		t.Fatalf("inverted start/stop decoded %d points, want 0", len(got))
	}
}

// TestClass0MultiHeaderMixedQualifiers verifies a single APDU carrying several
// Class-0 object headers with mixed qualifiers (0x07 packed, 0x00 start/stop,
// 0x17 indexed) decodes every point with no omission or duplication.
func TestClass0MultiHeaderMixedQualifiers(t *testing.T) {
	data := []byte{
		// G1V1 packed, count8, 2 points -> one packed byte 0x03 (bits 0,1 set)
		0x01, 0x01, 0x07, 0x02, 0x03,
		// G30V1, start/stop 0..0, 1 point idx0 = 42
		0x1E, 0x01, 0x00, 0x00, 0x00, 0x2A, 0x00, 0x00, 0x00, 0x01,
		// G20V1, indexed-count 0x17, 1 point idx1 = 7
		0x14, 0x01, 0x17, 0x01, 0x01, 0x07, 0x00, 0x00, 0x00, 0x01,
	}
	bi := parseBinaryInputs(data)
	if len(bi) != 2 {
		t.Fatalf("binary inputs = %d, want 2", len(bi))
	}
	ai := parseAnalogInputs(data)
	if len(ai) != 1 || ai[0].Index != 0 || ai[0].Value != 42 {
		t.Fatalf("analog inputs = %+v, want one idx0 val42", ai)
	}
	ct := parseCounters(data)
	if len(ct) != 1 || ct[0].Index != 1 || ct[0].Value != 7 {
		t.Fatalf("counters = %+v, want one idx1 val7", ct)
	}
}
