package outstation

import (
	"testing"
)

// These tests lock the standards-correct static response qualifiers emitted by
// the outstation builders after the Class-0 qualifier correction (VEC-01
// corrective #9): the indexed builders must use 0x17 (UINT8_CNT_UINT8_INDEX,
// 1-octet count with a per-point 8-bit index prefix) and the binary/analog
// output builders must use 0x07 (UINT8_CNT, sequential, no index prefix).

func newDefaultOutstation(t *testing.T) *Outstation {
	t.Helper()
	ost := NewOutstation(nil)
	ost.Initialize()
	ost.data = NewDefaultDataHandler()
	return ost
}

// TestBuildBinaryInputIndexedQualifier verifies G1V2 (non-packed) uses the 0x17
// indexed-count qualifier with an 8-bit per-point index prefix.
func TestBuildBinaryInputIndexedQualifier(t *testing.T) {
	ost := newDefaultOutstation(t)
	got := ost.buildBinaryInputData(2) // variation 2 -> indexed branch
	if len(got) < 4 {
		t.Fatalf("binary input data too short: % X", got)
	}
	if got[0] != 1 || got[1] != 2 || got[2] != 0x17 {
		t.Fatalf("header = % X, want 01 02 17", got[:3])
	}
	count := int(got[3])
	if count != 5 {
		t.Fatalf("count = %d, want 5", count)
	}
	// Each point is index(1) + value(1); the index prefix must be sequential.
	for i := 0; i < count; i++ {
		idx := got[4+i*2]
		if int(idx) != i {
			t.Fatalf("point %d index prefix = %d, want %d", i, idx, i)
		}
	}
}

// TestBuildAnalogInputIndexedQualifier verifies G30V2 uses the 0x17 indexed
// qualifier with an 8-bit per-point index prefix.
func TestBuildAnalogInputIndexedQualifier(t *testing.T) {
	ost := newDefaultOutstation(t)
	got := ost.buildAnalogInputData(2) // variation 2 -> indexed branch
	if len(got) < 4 {
		t.Fatalf("analog input data too short: % X", got)
	}
	if got[0] != 30 || got[1] != 2 || got[2] != 0x17 {
		t.Fatalf("header = % X, want 1E 02 17", got[:3])
	}
	if int(got[3]) != 5 {
		t.Fatalf("count = %d, want 5", got[3])
	}
	// Point layout: index(1) + int16 value(2) + quality(1) = 4 octets each.
	for i := 0; i < 5; i++ {
		idx := got[4+i*4]
		if int(idx) != i {
			t.Fatalf("point %d index prefix = %d, want %d", i, idx, i)
		}
	}
}

// TestBuildCounterIndexedQualifier verifies G20V5 uses the 0x17 indexed
// qualifier with an 8-bit per-point index prefix.
func TestBuildCounterIndexedQualifier(t *testing.T) {
	ost := newDefaultOutstation(t)
	got := ost.buildCounterData(5) // variation 5 -> indexed branch
	if len(got) < 4 {
		t.Fatalf("counter data too short: % X", got)
	}
	if got[0] != 20 || got[1] != 5 || got[2] != 0x17 {
		t.Fatalf("header = % X, want 14 05 17", got[:3])
	}
	if int(got[3]) != 5 {
		t.Fatalf("count = %d, want 5", got[3])
	}
	for i := 0; i < 5; i++ {
		idx := got[4+i*4]
		if int(idx) != i {
			t.Fatalf("point %d index prefix = %d, want %d", i, idx, i)
		}
	}
}

// TestBuildBinaryOutputSequentialQualifier verifies G10 uses the sequential
// 0x07 count8 qualifier with no per-point index prefix.
func TestBuildBinaryOutputSequentialQualifier(t *testing.T) {
	ost := newDefaultOutstation(t)
	got := ost.buildBinaryOutputData(1)
	if len(got) < 4 {
		t.Fatalf("binary output data too short: % X", got)
	}
	if got[0] != 10 || got[1] != 1 || got[2] != 0x07 {
		t.Fatalf("header = % X, want 0A 01 07", got[:3])
	}
	count := int(got[3])
	// Sequential: header(4) + count * 1 value octet, no index prefixes.
	if len(got) != 4+count {
		t.Fatalf("length = %d, want %d (sequential, no index prefix)", len(got), 4+count)
	}
}

// TestBuildAnalogOutputSequentialQualifier verifies G40 uses the sequential
// 0x07 count8 qualifier with no per-point index prefix.
func TestBuildAnalogOutputSequentialQualifier(t *testing.T) {
	ost := newDefaultOutstation(t)
	got := ost.buildAnalogOutputData(1)
	if len(got) < 4 {
		t.Fatalf("analog output data too short: % X", got)
	}
	if got[0] != 40 || got[1] != 1 || got[2] != 0x07 {
		t.Fatalf("header = % X, want 28 01 07", got[:3])
	}
	count := int(got[3])
	// Sequential: header(4) + count * (float32(4) + quality(1)).
	if len(got) != 4+count*5 {
		t.Fatalf("length = %d, want %d (sequential, no index prefix)", len(got), 4+count*5)
	}
}

// TestRequestStartStopHeaderConsumed verifies the request header size table
// treats 0x00 as a 5-octet start/stop header (not a 4-octet count header).
func TestRequestStartStopHeaderConsumed(t *testing.T) {
	if got := requestObjectHeaderSize([]byte{1, 1, 0x00, 0x00, 0x04}, 0); got != 5 {
		t.Fatalf("requestObjectHeaderSize(0x00) = %d, want 5", got)
	}
	if got := requestObjectHeaderSize([]byte{1, 1, 0x17, 0x02}, 0); got != 4 {
		t.Fatalf("requestObjectHeaderSize(0x17) = %d, want 4", got)
	}
}
