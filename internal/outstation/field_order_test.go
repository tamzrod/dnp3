package outstation

import (
	"testing"
)

// VEC-01 Gate C static-object field-order corrective — Outstation byte-exact
// encoder goldens.
//
// IEEE 1815 static analog-input (G30V1) and counter (G20V1) objects place the
// quality-flags octet FIRST, then the little-endian value, matching the pinned
// OpenDNP3 3.1.2 generated serializer (Group30Var1 = (flags, value32);
// Group20Var1 = (flags, count32)) and the §44 original-capture bytes
// (`01 fc 08 00 00` = flags 0x01 + int32 2300; `01 2a 00 00 00` = flags 0x01 +
// uint32 42). Expected bytes below are stated independently from the peer/spec,
// never generated with the encoder under test.

// newFieldOrderOutstation builds an outstation whose analog-input and counter
// points are exactly the requested values, so the emitted object bytes are
// deterministic and byte-checkable.
func newFieldOrderOutstation(analogs []AnalogInput, counters []Counter) *Outstation {
	ost := NewOutstation(nil)
	ost.Initialize()
	ost.data = &staticDataHandler{analogs: analogs, counters: counters}
	return ost
}

// staticDataHandler is a minimal DataHandler returning fixed analog/counter
// points for byte-exact encoder assertions.
type staticDataHandler struct {
	analogs  []AnalogInput
	counters []Counter
}

func (d *staticDataHandler) GetBinaryInputs() []BinaryInput { return nil }
func (d *staticDataHandler) GetAnalogInputs() []AnalogInput { return d.analogs }
func (d *staticDataHandler) GetCounters() []Counter         { return d.counters }
func (d *staticDataHandler) GetBinaryOutputs() []BinaryOutput {
	return nil
}
func (d *staticDataHandler) GetAnalogOutputs() []AnalogOutput { return nil }
func (d *staticDataHandler) GetFrozenCounters() []Counter     { return nil }
func (d *staticDataHandler) FreezeCounters(clear bool) error  { return nil }

// TestBuildAnalogInputFlagsFirstByteExact pins the complete G30V1 encoder output
// for the §44 ground truth: AI[0]=2300 ONLINE must serialize to the header
// `1E 01 07 01` followed by `01 fc 08 00 00` (flags first, then int32 LE 2300).
func TestBuildAnalogInputFlagsFirstByteExact(t *testing.T) {
	ost := newFieldOrderOutstation(
		[]AnalogInput{{Value: 2300, Quality: AnalogQualityOnline}},
		nil,
	)
	got := ost.buildAnalogInputData(1)
	want := []byte{0x1E, 0x01, 0x07, 0x01, 0x01, 0xFC, 0x08, 0x00, 0x00}
	if !bytesEqual(got, want) {
		t.Fatalf("G30V1 bytes = % X, want % X (flags octet must precede the value)", got, want)
	}
}

// TestBuildCounterFlagsFirstByteExact pins the complete G20V1 encoder output for
// the §44 ground truth: CT[0]=42 ONLINE must serialize to the header
// `14 01 07 01` followed by `01 2a 00 00 00` (flags first, then uint32 LE 42).
func TestBuildCounterFlagsFirstByteExact(t *testing.T) {
	ost := newFieldOrderOutstation(
		nil,
		[]Counter{{Value: 42, Quality: CounterQualityOnline}},
	)
	got := ost.buildCounterData(1)
	want := []byte{0x14, 0x01, 0x07, 0x01, 0x01, 0x2A, 0x00, 0x00, 0x00}
	if !bytesEqual(got, want) {
		t.Fatalf("G20V1 bytes = % X, want % X (flags octet must precede the value)", got, want)
	}
}

// TestBuildAnalogInputFlagsFirstOffline checks the flags octet is still emitted
// first for a non-online quality, and that the value bytes are unchanged.
func TestBuildAnalogInputFlagsFirstOffline(t *testing.T) {
	ost := newFieldOrderOutstation(
		[]AnalogInput{{Value: 2300, Quality: AnalogQualityCommLost}},
		nil,
	)
	got := ost.buildAnalogInputData(1)
	want := []byte{0x1E, 0x01, 0x07, 0x01, 0x80, 0xFC, 0x08, 0x00, 0x00}
	if !bytesEqual(got, want) {
		t.Fatalf("G30V1 offline bytes = % X, want % X", got, want)
	}
}

// TestBuildCounterFlagsFirstOffline is the counter analogue.
func TestBuildCounterFlagsFirstOffline(t *testing.T) {
	ost := newFieldOrderOutstation(
		nil,
		[]Counter{{Value: 42, Quality: CounterQualityCommLost}},
	)
	got := ost.buildCounterData(1)
	want := []byte{0x14, 0x01, 0x07, 0x01, 0x80, 0x2A, 0x00, 0x00, 0x00}
	if !bytesEqual(got, want) {
		t.Fatalf("G20V1 offline bytes = % X, want % X", got, want)
	}
}

// TestBuildAnalogInputSignedNegativeByteExact pins the little-endian signed
// representation: -1 must be `ff ff ff ff` after the flags octet.
func TestBuildAnalogInputSignedNegativeByteExact(t *testing.T) {
	ost := newFieldOrderOutstation(
		[]AnalogInput{{Value: -1, Quality: AnalogQualityOnline}},
		nil,
	)
	got := ost.buildAnalogInputData(1)
	want := []byte{0x1E, 0x01, 0x07, 0x01, 0x01, 0xFF, 0xFF, 0xFF, 0xFF}
	if !bytesEqual(got, want) {
		t.Fatalf("G30V1 -1 bytes = % X, want % X", got, want)
	}
}

// TestBuildCounterUnsignedHighBitByteExact pins the unsigned representation: a
// counter with the high bit set must not be treated as negative.
func TestBuildCounterUnsignedHighBitByteExact(t *testing.T) {
	ost := newFieldOrderOutstation(
		nil,
		[]Counter{{Value: 0x80000000, Quality: CounterQualityOnline}},
	)
	got := ost.buildCounterData(1)
	want := []byte{0x14, 0x01, 0x07, 0x01, 0x01, 0x00, 0x00, 0x00, 0x80}
	if !bytesEqual(got, want) {
		t.Fatalf("G20V1 high-bit bytes = % X, want % X", got, want)
	}
}

// TestBuildAnalogInputMultiPointFlagsFirst checks a two-point block: each
// 5-octet object carries its own flags octet first, points sequential.
func TestBuildAnalogInputMultiPointFlagsFirst(t *testing.T) {
	ost := newFieldOrderOutstation(
		[]AnalogInput{
			{Value: 2300, Quality: AnalogQualityOnline},
			{Value: 42, Quality: AnalogQualityOnline},
		},
		nil,
	)
	got := ost.buildAnalogInputData(1)
	want := []byte{
		0x1E, 0x01, 0x07, 0x02,
		0x01, 0xFC, 0x08, 0x00, 0x00,
		0x01, 0x2A, 0x00, 0x00, 0x00,
	}
	if !bytesEqual(got, want) {
		t.Fatalf("G30V1 2-point bytes = % X, want % X", got, want)
	}
}

// TestBuildCounterMultiPointFlagsFirst is the counter analogue.
func TestBuildCounterMultiPointFlagsFirst(t *testing.T) {
	ost := newFieldOrderOutstation(
		nil,
		[]Counter{
			{Value: 42, Quality: CounterQualityOnline},
			{Value: 2300, Quality: CounterQualityOnline},
		},
	)
	got := ost.buildCounterData(1)
	want := []byte{
		0x14, 0x01, 0x07, 0x02,
		0x01, 0x2A, 0x00, 0x00, 0x00,
		0x01, 0xFC, 0x08, 0x00, 0x00,
	}
	if !bytesEqual(got, want) {
		t.Fatalf("G20V1 2-point bytes = % X, want % X", got, want)
	}
}

// TestBuildAnalogInputEmptyNoHeader verifies no object header is emitted when
// there are no points (avoids a malformed empty block).
func TestBuildAnalogInputEmptyNoHeader(t *testing.T) {
	ost := newFieldOrderOutstation(nil, nil)
	if got := ost.buildAnalogInputData(1); len(got) != 0 {
		t.Fatalf("empty analog block = % X, want no bytes", got)
	}
	if got := ost.buildCounterData(1); len(got) != 0 {
		t.Fatalf("empty counter block = % X, want no bytes", got)
	}
}

// TestFlagsFirstRoundTripsThroughMasterEncoderOrder is a self-contained guard:
// the encoder's first object octet after the 4-octet header must be the quality
// flag (0x01 online), never the low byte of the value, for both variations.
func TestFlagsFirstRoundTripsThroughMasterEncoderOrder(t *testing.T) {
	ai := newFieldOrderOutstation([]AnalogInput{{Value: 2300, Quality: AnalogQualityOnline}}, nil).buildAnalogInputData(1)
	if len(ai) < 5 || ai[4] != 0x01 {
		t.Fatalf("G30V1 first object octet = % X, want flags 0x01 first (got % X)", ai, ai)
	}
	ct := newFieldOrderOutstation(nil, []Counter{{Value: 42, Quality: CounterQualityOnline}}).buildCounterData(1)
	if len(ct) < 5 || ct[4] != 0x01 {
		t.Fatalf("G20V1 first object octet = % X, want flags 0x01 first (got % X)", ct, ct)
	}
}
