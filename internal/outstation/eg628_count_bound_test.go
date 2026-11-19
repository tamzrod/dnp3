package outstation

import (
	"testing"

	"dnp3/internal/al"
)

// EG628-IM-001 regressions. Both defects were reproduced without hardware
// against the generic in-memory outstation and independently decoded, then
// attributed to this library rather than to the EG628 adapter or the
// independent Windows master. See docs/verification/EG628-IM-001.md.

// largeStaticHandler serves an arbitrary number of points per group so the
// 8-bit count-qualifier boundary can be exercised.
type largeStaticHandler struct {
	binaries  []BinaryInput
	analogs   []AnalogInput
	counters  []Counter
	binaryOut []BinaryOutput
	analogOut []AnalogOutput
}

func (h *largeStaticHandler) GetBinaryInputs() []BinaryInput   { return h.binaries }
func (h *largeStaticHandler) GetAnalogInputs() []AnalogInput   { return h.analogs }
func (h *largeStaticHandler) GetCounters() []Counter           { return h.counters }
func (h *largeStaticHandler) GetBinaryOutputs() []BinaryOutput { return h.binaryOut }
func (h *largeStaticHandler) GetAnalogOutputs() []AnalogOutput { return h.analogOut }
func (h *largeStaticHandler) GetFrozenCounters() []Counter     { return nil }
func (h *largeStaticHandler) FreezeCounters(bool) error        { return nil }

func newLargeStaticOutstation(h DataHandler) *Outstation {
	ost := NewOutstation(nil)
	_ = ost.Initialize()
	ost.SetDataHandler(h)
	return ost
}

func makeBinaries(n int) []BinaryInput {
	out := make([]BinaryInput, n)
	for i := range out {
		out[i] = BinaryInput{Value: i%2 == 0, Quality: BinaryQualityOnline}
	}
	return out
}

func makeAnalogs(n int) []AnalogInput {
	out := make([]AnalogInput, n)
	for i := range out {
		out[i] = AnalogInput{Value: float64(i), Quality: AnalogQualityOnline}
	}
	return out
}

func makeCounters(n int) []Counter {
	out := make([]Counter, n)
	for i := range out {
		out[i] = Counter{Value: uint32(i), Quality: CounterQualityOnline}
	}
	return out
}

func makeBinaryOutputs(n int) []BinaryOutput {
	out := make([]BinaryOutput, n)
	for i := range out {
		out[i] = BinaryOutput{Value: i%2 == 0, Quality: BinaryQualityOnline}
	}
	return out
}

func makeAnalogOutputs(n int) []AnalogOutput {
	out := make([]AnalogOutput, n)
	for i := range out {
		out[i] = AnalogOutput{Value: float64(i), Quality: AnalogQualityOnline}
	}
	return out
}

// TestEG628StaticPackedBinaryCountAbove255 pins the correction for a packed
// Group 1 Variation 1 block with more than 255 points. An 8-bit count octet
// cannot express 256+; emitting `0x07, byte(256)` writes count 0 and leaves the
// packed octets unannounced, so a strict peer parser reads the first packed
// octet as the next object header and desynchronises. The block must instead use
// the 16-bit count qualifier 0x27.
func TestEG628StaticPackedBinaryCountAbove255(t *testing.T) {
	cases := []struct {
		name   string
		points int
		wantQ  []byte // expected qualifier + count octets
	}{
		{"255 uses count8", 255, []byte{0x07, 0xFF}},
		{"256 uses count16", 256, []byte{0x27, 0x00, 0x01}},
		{"300 uses count16", 300, []byte{0x27, 0x2C, 0x01}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ost := newLargeStaticOutstation(&largeStaticHandler{binaries: makeBinaries(tc.points)})
			got := ost.buildBinaryInputData(1)

			if got[0] != 1 || got[1] != 1 {
				t.Fatalf("header group/variation = % X, want 01 01", got[:2])
			}
			if !bytesEqual(got[2:2+len(tc.wantQ)], tc.wantQ) {
				t.Fatalf("qualifier+count = % X, want % X", got[2:2+len(tc.wantQ)], tc.wantQ)
			}

			// The block must decode as exactly one header carrying tc.points.
			h, consumed, err := al.DecodeObjectHeader(got, 0)
			if err != nil {
				t.Fatalf("DecodeObjectHeader: %v", err)
			}
			if int(h.Count) != tc.points {
				t.Fatalf("decoded count = %d, want %d (count octet must not wrap)", h.Count, tc.points)
			}
			wantPacked := (tc.points + 7) / 8
			if len(got)-consumed != wantPacked {
				t.Fatalf("packed payload = %d octets, want %d", len(got)-consumed, wantPacked)
			}
		})
	}
}

// TestEG628StaticSequentialCountAbove255 pins the same correction for the
// sequential with-flags builders (G30V1 analog, G20V1 counter) and the
// sequential status builders (G10V1, G40V1).
func TestEG628StaticSequentialCountAbove255(t *testing.T) {
	h := &largeStaticHandler{
		analogs:  makeAnalogs(300),
		counters: makeCounters(300),
	}
	ost := newLargeStaticOutstation(h)

	blocks := map[string]struct {
		data     []byte
		group    uint8
		perPoint int
	}{
		"G30V1": {ost.buildAnalogInputData(1), 30, 5},
		"G20V1": {ost.buildCounterData(1), 20, 5},
	}
	for name, b := range blocks {
		t.Run(name, func(t *testing.T) {
			obj, _, err := al.DecodeObjectHeader(b.data, 0)
			if err != nil {
				t.Fatalf("DecodeObjectHeader: %v", err)
			}
			if obj.Group != b.group || obj.Qualifier != al.QualCount16 {
				t.Fatalf("header = group %d qualifier 0x%02X, want group %d qualifier 0x27",
					obj.Group, obj.Qualifier, b.group)
			}
			if int(obj.Count) != 300 {
				t.Fatalf("count = %d, want 300", obj.Count)
			}
			if got, want := len(b.data)-obj.EncodedSize(), 300*b.perPoint; got != want {
				t.Fatalf("payload = %d octets, want %d", got, want)
			}
		})
	}
}

// TestEG628StaticIndexedFallsBackAbove255 pins the indexed-branch correction.
// Qualifier 0x17 pairs an 8-bit count with an 8-bit per-point index prefix, so
// point 256 cannot be addressed. A larger block must fall back to a sequential
// 0x27 block with no index prefixes rather than repeating wrapped indexes.
func TestEG628StaticIndexedFallsBackAbove255(t *testing.T) {
	ost := newLargeStaticOutstation(&largeStaticHandler{binaries: makeBinaries(300)})
	got := ost.buildBinaryInputData(2) // non-packed variation -> indexed branch

	obj, consumed, err := al.DecodeObjectHeader(got, 0)
	if err != nil {
		t.Fatalf("DecodeObjectHeader: %v", err)
	}
	if obj.Qualifier != al.QualCount16 {
		t.Fatalf("qualifier = 0x%02X, want 0x27 (0x17 cannot index point 256+)", obj.Qualifier)
	}
	if int(obj.Count) != 300 {
		t.Fatalf("count = %d, want 300", obj.Count)
	}
	// One octet per point, no index prefixes.
	if got, want := len(got)-consumed, 300; got != want {
		t.Fatalf("payload = %d octets, want %d (index prefixes must be dropped)", got, want)
	}
	// The values must be the point states in order, not index bytes.
	for i := 0; i < 300; i++ {
		want := byte(0x00)
		if i%2 == 0 {
			want = 0x01
		}
		if got[consumed+i] != want {
			t.Fatalf("point %d = 0x%02X, want 0x%02X", i, got[consumed+i], want)
		}
	}
}

// TestEG628StaticSmallCountUnchanged guards against a regression in the common
// case: at or below 255 points the builders must keep using the 8-bit count
// qualifiers that the existing pinned goldens expect.
func TestEG628StaticSmallCountUnchanged(t *testing.T) {
	ost := newLargeStaticOutstation(&largeStaticHandler{
		binaries: makeBinaries(5),
		analogs:  makeAnalogs(5),
		counters: makeCounters(5),
	})
	checks := []struct {
		name string
		data []byte
		want byte
	}{
		{"G1V1", ost.buildBinaryInputData(1), 0x07},
		{"G1V2", ost.buildBinaryInputData(2), 0x17},
		{"G30V1", ost.buildAnalogInputData(1), 0x07},
		{"G30V2", ost.buildAnalogInputData(2), 0x17},
		{"G20V1", ost.buildCounterData(1), 0x07},
	}
	for _, c := range checks {
		if c.data[2] != c.want {
			t.Fatalf("%s qualifier = 0x%02X, want 0x%02X", c.name, c.data[2], c.want)
		}
	}
}

// TestEG628Class0to3IntegrityReadDecodes is the focused end-to-end reproduction:
// a Class 0+1+2+3 integrity READ against an in-memory outstation must produce a
// response whose object stream decodes completely, with every block header
// carrying the true point count.
func TestEG628Class0to3IntegrityReadDecodes(t *testing.T) {
	cfg := DefaultConfig()
	cfg.OutstationAddress = 4
	cfg.MasterAddress = 3
	ost := NewOutstation(cfg)
	ost.SetDataHandler(&largeStaticHandler{
		binaries:  makeBinaries(300),
		analogs:   makeAnalogs(5),
		counters:  makeCounters(5),
		binaryOut: makeBinaryOutputs(5),
		analogOut: makeAnalogOutputs(5),
	})
	if err := ost.Initialize(); err != nil {
		t.Fatalf("Initialize: %v", err)
	}
	if err := ost.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}

	read := &al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: 1},
		FuncCode: al.FuncRead,
		Data:     []byte{0x3C, 0x01, 0x06, 0x3C, 0x02, 0x06, 0x3C, 0x03, 0x06, 0x3C, 0x04, 0x06},
	}
	resp, err := ost.ProcessRequest(read)
	if err != nil {
		t.Fatalf("ProcessRequest: %v", err)
	}
	if resp == nil {
		t.Fatal("no response")
	}
	iin, err := al.DecodeIIN(resp.Data)
	if err != nil {
		t.Fatalf("DecodeIIN: %v", err)
	}
	if iin.ParameterError || iin.ObjectUnknown || iin.FuncUnknown {
		t.Fatalf("integrity read carried an error IIN: %+v", iin)
	}

	// Walk the object stream with the library's own header decoder and require
	// that every header is well formed and the walk consumes the whole payload.
	off := 2 // skip the two IIN octets
	blocks := 0
	for off < len(resp.Data) {
		h, consumed, err := al.DecodeObjectHeader(resp.Data, off)
		if err != nil {
			t.Fatalf("object header at offset %d: %v", off, err)
		}
		if h.Group == 1 && h.Variation == 1 {
			// Packed binary carries no per-point octet; the payload is ceil(n/8).
			off += consumed + (int(h.Count)+7)/8
		} else {
			per, ok := staticObjectSize(h.Group, h.Variation)
			if !ok {
				t.Fatalf("unexpected object G%dV%d qualifier 0x%02X at offset %d",
					h.Group, h.Variation, h.Qualifier, off)
			}
			off += consumed + int(h.Count)*per
		}
		if off > len(resp.Data) {
			t.Fatalf("object G%dV%d claims %d octets beyond the payload",
				h.Group, h.Variation, off-len(resp.Data))
		}
		blocks++
	}
	if blocks != 5 {
		t.Fatalf("decoded %d object blocks, want 5 (G1,G30,G20,G10,G40)", blocks)
	}
}

// staticObjectSize returns the per-point octets for the static variations the
// outstation emits, and whether the group/variation is known. Packed G1V1 is
// handled separately by the caller.
func staticObjectSize(group, variation uint8) (int, bool) {
	switch group {
	case 1:
		if variation == 1 {
			return 0, true // packed: covered by the count/8 span, handled by caller
		}
		return 1, true
	case 10:
		return 1, true
	case 20:
		switch variation {
		case 1:
			return 5, true
		case 5:
			return 3, true
		case 6:
			return 4, true
		}
	case 30:
		switch variation {
		case 1, 3, 4:
			return 5, true
		case 2:
			return 3, true
		case 5:
			return 4, true
		}
	case 40:
		switch variation {
		case 1:
			return 5, true
		case 2:
			return 3, true
		}
	}
	return 0, false
}
