package master

import (
	"math"
	"testing"

	"dnp3/pkg/dnp3/types"
)

// VEC-01 Gate C static-object field-order corrective — Master byte-exact
// golden and negative tests.
//
// The §44 original-capture evidence (docs/verification/VEC-01-LAB-EXEC-001.md)
// shows a conforming pinned OpenDNP3 3.1.2 Class-0 response whose static
// objects place the quality-flags octet FIRST, then the little-endian value:
//
//	G30V1 object 0: 01 fc 08 00 00  -> flags 0x01 (ONLINE), int32 LE 2300
//	G20V1 object 0: 01 2a 00 00 00  -> flags 0x01 (ONLINE), uint32 LE 42
//
// This matches IEEE 1815 and the peer's own generated serializer
// (opendnp3 3.1.2 Group30Var1 = (flags, value32); Group20Var1 = (flags, count32)).
// Expected bytes below are stated independently from the peer/spec, never
// generated with the code under test.

// --- decisive §44 goldens -------------------------------------------------

// TestG30V1FlagsFirstGolden pins the exact §44 capture bytes: the flags octet
// precedes the value, so AI[0] decodes to 2300 ONLINE. The historical
// value-then-flags misread of these same bytes was 588801 (quality 0x00); the
// test fails if that regression returns.
func TestG30V1FlagsFirstGolden(t *testing.T) {
	data := []byte{0x1E, 0x01, 0x07, 0x01, 0x01, 0xFC, 0x08, 0x00, 0x00}
	got := parseAnalogInputs(data)
	if len(got) != 1 {
		t.Fatalf("got %d points, want 1", len(got))
	}
	if got[0].Index != 0 {
		t.Errorf("Index = %d, want 0", got[0].Index)
	}
	if got[0].Value != 2300 {
		t.Errorf("Value = %v, want 2300 (588801 means value-then-flags regressed)", got[0].Value)
	}
	if got[0].Value == 588801 {
		t.Errorf("Value = 588801: the §44 value-then-flags regression is present")
	}
	if got[0].Quality != types.QualityOnline {
		t.Errorf("Quality = %v, want ONLINE (0x01)", got[0].Quality)
	}
}

// TestG20V1FlagsFirstGolden pins the exact §44 capture bytes for the counter:
// CT[0] decodes to 42 ONLINE. The historical misread was 10753.
func TestG20V1FlagsFirstGolden(t *testing.T) {
	data := []byte{0x14, 0x01, 0x07, 0x01, 0x01, 0x2A, 0x00, 0x00, 0x00}
	got := parseCounters(data)
	if len(got) != 1 {
		t.Fatalf("got %d points, want 1", len(got))
	}
	if got[0].Index != 0 {
		t.Errorf("Index = %d, want 0", got[0].Index)
	}
	if got[0].Value != 42 {
		t.Errorf("Value = %d, want 42 (10753 means value-then-flags regressed)", got[0].Value)
	}
	if got[0].Value == 10753 {
		t.Errorf("Value = 10753: the §44 value-then-flags regression is present")
	}
	if got[0].Quality != types.QualityOnline {
		t.Errorf("Quality = %v, want ONLINE (0x01)", got[0].Quality)
	}
}

// --- value-then-flags negative guards -------------------------------------

// TestG30V1ValueThenFlagsNotAccepted feeds the value-then-flags ordering
// (value bytes first, flags last). The decoder must not silently produce the
// correct value 2300 or report ONLINE for these bytes; the flags octet is
// consumed from the wrong position, so the result is a different value with the
// online bit clear. This is a fail-closed guard, not a claim that reversed
// bytes are categorically invalid DNP3 data.
func TestG30V1ValueThenFlagsNotAccepted(t *testing.T) {
	// 2300 as LE (FC 08 00 00) then flags 0x01.
	data := []byte{0x1E, 0x01, 0x07, 0x01, 0xFC, 0x08, 0x00, 0x00, 0x01}
	got := parseAnalogInputs(data)
	if len(got) != 1 {
		t.Fatalf("got %d points, want 1", len(got))
	}
	if got[0].Value == 2300 {
		t.Errorf("Value = 2300 for value-then-flags bytes: flags-first ordering is not enforced")
	}
	if got[0].Quality&types.QualityOnline != 0 {
		t.Errorf("Quality = %v reports ONLINE for value-then-flags bytes", got[0].Quality)
	}
}

// TestG20V1ValueThenFlagsNotAccepted is the counter analogue of the guard above.
func TestG20V1ValueThenFlagsNotAccepted(t *testing.T) {
	// 42 as LE (2A 00 00 00) then flags 0x01.
	data := []byte{0x14, 0x01, 0x07, 0x01, 0x2A, 0x00, 0x00, 0x00, 0x01}
	got := parseCounters(data)
	if len(got) != 1 {
		t.Fatalf("got %d points, want 1", len(got))
	}
	if got[0].Value == 42 {
		t.Errorf("Value = 42 for value-then-flags bytes: flags-first ordering is not enforced")
	}
	if got[0].Quality&types.QualityOnline != 0 {
		t.Errorf("Quality = %v reports ONLINE for value-then-flags bytes", got[0].Quality)
	}
}

// --- quality flags --------------------------------------------------------

// TestG30V1FlagsFirstQuality locks the per-point flags octet (first octet of
// the 5-octet object) for online and abnormal quality states.
func TestG30V1FlagsFirstQuality(t *testing.T) {
	cases := []struct {
		name string
		flag byte
		want types.QualityFlags
	}{
		{"online", 0x01, types.QualityOnline},
		{"restart", 0x40, types.QualityRestart},
		{"comm_lost", 0x80, types.QualityCommLost},
		{"no_flags", 0x00, 0x00},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// flags octet first, then value 2300 = FC 08 00 00.
			data := []byte{0x1E, 0x01, 0x07, 0x01, tc.flag, 0xFC, 0x08, 0x00, 0x00}
			got := parseAnalogInputs(data)
			if len(got) != 1 {
				t.Fatalf("got %d points, want 1", len(got))
			}
			if got[0].Value != 2300 {
				t.Errorf("Value = %v, want 2300", got[0].Value)
			}
			if got[0].Quality != tc.want {
				t.Errorf("Quality = %v, want %v", got[0].Quality, tc.want)
			}
		})
	}
}

// TestG20V1FlagsFirstQuality is the counter analogue.
func TestG20V1FlagsFirstQuality(t *testing.T) {
	cases := []struct {
		name string
		flag byte
		want types.QualityFlags
	}{
		{"online", 0x01, types.QualityOnline},
		{"restart", 0x40, types.QualityRestart},
		{"comm_lost", 0x80, types.QualityCommLost},
		{"no_flags", 0x00, 0x00},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data := []byte{0x14, 0x01, 0x07, 0x01, tc.flag, 0x2A, 0x00, 0x00, 0x00}
			got := parseCounters(data)
			if len(got) != 1 {
				t.Fatalf("got %d points, want 1", len(got))
			}
			if got[0].Value != 42 {
				t.Errorf("Value = %d, want 42", got[0].Value)
			}
			if got[0].Quality != tc.want {
				t.Errorf("Quality = %v, want %v", got[0].Quality, tc.want)
			}
		})
	}
}

// --- signed / unsigned and multibyte values -------------------------------

// TestG30V1SignedMultibyteValues exercises zero, nontrivial multibyte values and
// the int32 signed range with the flags octet first.
func TestG30V1SignedMultibyteValues(t *testing.T) {
	cases := []struct {
		name  string
		value []byte // 4 value bytes, LSB first
		want  float64
	}{
		{"zero", []byte{0x00, 0x00, 0x00, 0x00}, 0},
		{"one", []byte{0x01, 0x00, 0x00, 0x00}, 1},
		{"2300", []byte{0xFC, 0x08, 0x00, 0x00}, 2300},
		{"-1", []byte{0xFF, 0xFF, 0xFF, 0xFF}, -1},
		{"-1000", []byte{0x18, 0xFC, 0xFF, 0xFF}, -1000},
		{"max int32", []byte{0xFF, 0xFF, 0xFF, 0x7F}, math.MaxInt32},
		{"min int32", []byte{0x00, 0x00, 0x00, 0x80}, math.MinInt32},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data := []byte{0x1E, 0x01, 0x07, 0x01, 0x01} // flags ONLINE first
			data = append(data, tc.value...)
			got := parseAnalogInputs(data)
			if len(got) != 1 {
				t.Fatalf("got %d points, want 1", len(got))
			}
			if got[0].Value != tc.want {
				t.Errorf("Value = %v, want %v", got[0].Value, tc.want)
			}
		})
	}
}

// TestG20V1UnsignedMultibyteValues exercises zero, nontrivial multibyte values
// and the uint32 range (high bit set is not negative) with flags first.
func TestG20V1UnsignedMultibyteValues(t *testing.T) {
	cases := []struct {
		name  string
		value []byte // 4 value bytes, LSB first
		want  uint32
	}{
		{"zero", []byte{0x00, 0x00, 0x00, 0x00}, 0},
		{"one", []byte{0x01, 0x00, 0x00, 0x00}, 1},
		{"42", []byte{0x2A, 0x00, 0x00, 0x00}, 42},
		{"max uint32", []byte{0xFF, 0xFF, 0xFF, 0xFF}, math.MaxUint32},
		{"high bit set", []byte{0x00, 0x00, 0x00, 0x80}, 0x80000000},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data := []byte{0x14, 0x01, 0x07, 0x01, 0x01} // flags ONLINE first
			data = append(data, tc.value...)
			got := parseCounters(data)
			if len(got) != 1 {
				t.Fatalf("got %d points, want 1", len(got))
			}
			if got[0].Value != tc.want {
				t.Errorf("Value = %d, want %d", got[0].Value, tc.want)
			}
		})
	}
}

// --- supported v0 qualifier paths -----------------------------------------

// TestG30V1QualifierCount8MultiPoint decodes two sequential flags-first points
// under the 0x07 count8 qualifier.
func TestG30V1QualifierCount8MultiPoint(t *testing.T) {
	data := []byte{
		0x1E, 0x01, 0x07, 0x02,
		0x01, 0xFC, 0x08, 0x00, 0x00, // idx0 = 2300
		0x01, 0x2A, 0x00, 0x00, 0x00, // idx1 = 42
	}
	got := parseAnalogInputs(data)
	if len(got) != 2 {
		t.Fatalf("got %d points, want 2", len(got))
	}
	if got[0].Index != 0 || got[0].Value != 2300 {
		t.Errorf("point 0: idx=%d val=%v, want idx0 val2300", got[0].Index, got[0].Value)
	}
	if got[1].Index != 1 || got[1].Value != 42 {
		t.Errorf("point 1: idx=%d val=%v, want idx1 val42", got[1].Index, got[1].Value)
	}
}

// TestG30V1QualifierStartStop8 decodes the 0x00 UINT8_START_STOP qualifier with
// flags-first objects indexed sequentially from Start.
func TestG30V1QualifierStartStop8(t *testing.T) {
	data := []byte{
		0x1E, 0x01, 0x00, 0x05, 0x06, // start=5 stop=6
		0x01, 0xFC, 0x08, 0x00, 0x00, // idx5 = 2300
		0x01, 0x2A, 0x00, 0x00, 0x00, // idx6 = 42
	}
	got := parseAnalogInputs(data)
	if len(got) != 2 {
		t.Fatalf("got %d points, want 2", len(got))
	}
	if got[0].Index != 5 || got[0].Value != 2300 {
		t.Errorf("point 0: idx=%d val=%v, want idx5 val2300", got[0].Index, got[0].Value)
	}
	if got[1].Index != 6 || got[1].Value != 42 {
		t.Errorf("point 1: idx=%d val=%v, want idx6 val42", got[1].Index, got[1].Value)
	}
}

// TestG30V1QualifierRange16 decodes the 0x28 16-bit range qualifier with
// flags-first objects indexed sequentially from Start.
func TestG30V1QualifierRange16(t *testing.T) {
	data := []byte{
		0x1E, 0x01, 0x28, 0x05, 0x00, 0x06, 0x00, // start=5 stop=6
		0x01, 0xFC, 0x08, 0x00, 0x00, // idx5 = 2300
		0x01, 0x2A, 0x00, 0x00, 0x00, // idx6 = 42
	}
	got := parseAnalogInputs(data)
	if len(got) != 2 {
		t.Fatalf("got %d points, want 2", len(got))
	}
	if got[0].Index != 5 || got[0].Value != 2300 {
		t.Errorf("point 0: idx=%d val=%v, want idx5 val2300", got[0].Index, got[0].Value)
	}
	if got[1].Index != 6 || got[1].Value != 42 {
		t.Errorf("point 1: idx=%d val=%v, want idx6 val42", got[1].Index, got[1].Value)
	}
}

// TestG30V1QualifierCount16 decodes the 0x27 16-bit count qualifier.
func TestG30V1QualifierCount16(t *testing.T) {
	data := []byte{
		0x1E, 0x01, 0x27, 0x02, 0x00, // count16 = 2
		0x01, 0xFC, 0x08, 0x00, 0x00, // idx0 = 2300
		0x01, 0x2A, 0x00, 0x00, 0x00, // idx1 = 42
	}
	got := parseAnalogInputs(data)
	if len(got) != 2 {
		t.Fatalf("got %d points, want 2", len(got))
	}
	if got[0].Index != 0 || got[0].Value != 2300 {
		t.Errorf("point 0: idx=%d val=%v, want idx0 val2300", got[0].Index, got[0].Value)
	}
	if got[1].Index != 1 || got[1].Value != 42 {
		t.Errorf("point 1: idx=%d val=%v, want idx1 val42", got[1].Index, got[1].Value)
	}
}

// TestG30V1QualifierIndex8 decodes the 0x17 UINT8_CNT_UINT8_INDEX qualifier:
// an 8-bit count then a per-point 8-bit index prefix before each flags-first
// object.
func TestG30V1QualifierIndex8(t *testing.T) {
	data := []byte{
		0x1E, 0x01, 0x17, 0x02,
		0x03, 0x01, 0xFC, 0x08, 0x00, 0x00, // idx3 = 2300
		0x07, 0x01, 0x2A, 0x00, 0x00, 0x00, // idx7 = 42
	}
	got := parseAnalogInputs(data)
	if len(got) != 2 {
		t.Fatalf("got %d points, want 2", len(got))
	}
	if got[0].Index != 3 || got[0].Value != 2300 {
		t.Errorf("point 0: idx=%d val=%v, want idx3 val2300", got[0].Index, got[0].Value)
	}
	if got[1].Index != 7 || got[1].Value != 42 {
		t.Errorf("point 1: idx=%d val=%v, want idx7 val42", got[1].Index, got[1].Value)
	}
}

// TestG20V1QualifierIndex8 is the counter analogue of the 0x17 indexed path.
func TestG20V1QualifierIndex8(t *testing.T) {
	data := []byte{
		0x14, 0x01, 0x17, 0x02,
		0x03, 0x01, 0x2A, 0x00, 0x00, 0x00, // idx3 = 42
		0x07, 0x01, 0xFC, 0x08, 0x00, 0x00, // idx7 = 2300
	}
	got := parseCounters(data)
	if len(got) != 2 {
		t.Fatalf("got %d points, want 2", len(got))
	}
	if got[0].Index != 3 || got[0].Value != 42 {
		t.Errorf("point 0: idx=%d val=%d, want idx3 val42", got[0].Index, got[0].Value)
	}
	if got[1].Index != 7 || got[1].Value != 2300 {
		t.Errorf("point 1: idx=%d val=%d, want idx7 val2300", got[1].Index, got[1].Value)
	}
}

// --- negative / malformed objects (fail closed) ---------------------------

// TestG30V1TruncatedObjectBodyRejected verifies a count8 block that declares
// more points than the remaining bytes provide is rejected as a whole: no
// partial point is accepted for a truncated declared static block.
func TestG30V1TruncatedObjectBodyRejected(t *testing.T) {
	// Declares 2 points but the second object is one octet short.
	data := []byte{
		0x1E, 0x01, 0x07, 0x02,
		0x01, 0xFC, 0x08, 0x00, 0x00, // complete idx0
		0x01, 0xFC, 0x08, 0x00, // truncated idx1
	}
	if got := parseAnalogInputs(data); len(got) != 0 {
		t.Fatalf("truncated declared block decoded %d points, want 0 (fail closed)", len(got))
	}
}

// TestG20V1TruncatedObjectBodyRejected is the counter analogue.
func TestG20V1TruncatedObjectBodyRejected(t *testing.T) {
	data := []byte{
		0x14, 0x01, 0x07, 0x02,
		0x01, 0x2A, 0x00, 0x00, 0x00,
		0x01, 0x2A, 0x00, 0x00,
	}
	if got := parseCounters(data); len(got) != 0 {
		t.Fatalf("truncated declared block decoded %d points, want 0 (fail closed)", len(got))
	}
}

// TestG30V1TruncatedIndexPrefixRejected verifies a 0x17 block whose declared
// count does not fit (partial index prefix/value) yields no points.
func TestG30V1TruncatedIndexPrefixRejected(t *testing.T) {
	data := []byte{
		0x1E, 0x01, 0x17, 0x02,
		0x03, 0x01, 0xFC, 0x08, 0x00, 0x00, // complete idx3
		0x07, 0x01, 0xFC, // truncated idx7
	}
	if got := parseAnalogInputs(data); len(got) != 0 {
		t.Fatalf("truncated indexed block decoded %d points, want 0 (fail closed)", len(got))
	}
}

// TestG30V1MissingDeclaredObjectsRejected verifies a header that declares a
// count with no object bytes yields no points and does not panic.
func TestG30V1MissingDeclaredObjectsRejected(t *testing.T) {
	if got := parseAnalogInputs([]byte{0x1E, 0x01, 0x07, 0x01}); len(got) != 0 {
		t.Fatalf("count with no objects decoded %d points, want 0", len(got))
	}
	if got := parseCounters([]byte{0x14, 0x01, 0x07, 0x01}); len(got) != 0 {
		t.Fatalf("count with no objects decoded %d points, want 0", len(got))
	}
}

// TestG30V1TruncatedHeaderRejected verifies short headers are rejected without
// a panic or a misdecoded point.
func TestG30V1TruncatedHeaderRejected(t *testing.T) {
	cases := [][]byte{
		{0x1E},                   // group only
		{0x1E, 0x01},             // group + variation
		{0x1E, 0x01, 0x07},       // count8 header missing its count octet
		{0x1E, 0x01, 0x00, 0x05}, // start/stop header missing its stop octet
	}
	for _, data := range cases {
		if got := parseAnalogInputs(data); len(got) != 0 {
			t.Errorf("truncated header % X decoded %d points, want 0", data, len(got))
		}
	}
}

// TestG20V1TruncatedHeaderRejected is the counter analogue.
func TestG20V1TruncatedHeaderRejected(t *testing.T) {
	cases := [][]byte{
		{0x14},
		{0x14, 0x01},
		{0x14, 0x01, 0x07},
		{0x14, 0x01, 0x00, 0x09},
	}
	for _, data := range cases {
		if got := parseCounters(data); len(got) != 0 {
			t.Errorf("truncated header % X decoded %d points, want 0", data, len(got))
		}
	}
}
