package outstation

import (
	"bytes"
	"testing"

	"dnp3/internal/al"
)

// fullStaticHandler serves a configurable number of points for every static
// group so a single read exercises every builder at the same count.
type fullStaticHandler struct {
	binary   []BinaryInput
	analog   []AnalogInput
	counter  []Counter
	binOut   []BinaryOutput
	anaOut   []AnalogOutput
	frozen   []Counter
	freezeOK bool
}

func (h *fullStaticHandler) GetBinaryInputs() []BinaryInput   { return h.binary }
func (h *fullStaticHandler) GetAnalogInputs() []AnalogInput   { return h.analog }
func (h *fullStaticHandler) GetCounters() []Counter           { return h.counter }
func (h *fullStaticHandler) GetBinaryOutputs() []BinaryOutput { return h.binOut }
func (h *fullStaticHandler) GetAnalogOutputs() []AnalogOutput { return h.anaOut }
func (h *fullStaticHandler) GetFrozenCounters() []Counter     { return h.frozen }
func (h *fullStaticHandler) FreezeCounters(clear bool) error  { h.freezeOK = true; return nil }

func newFullStaticOutstation(n int) *Outstation {
	h := &fullStaticHandler{}
	for i := 0; i < n; i++ {
		h.binary = append(h.binary, BinaryInput{Value: i%2 == 0, Quality: BinaryQualityOnline})
		h.analog = append(h.analog, AnalogInput{Value: float64(i), Quality: AnalogQualityOnline})
		h.counter = append(h.counter, Counter{Value: uint32(i), Quality: CounterQualityOnline})
		h.binOut = append(h.binOut, BinaryOutput{Value: i%2 == 0, Quality: 0x01})
		h.anaOut = append(h.anaOut, AnalogOutput{Value: float64(i), Quality: 0x01})
	}
	ost := NewOutstation(nil)
	ost.Initialize()
	ost.SetDataHandler(h)
	return ost
}

// TestStaticCountQualifierBoundaries locks the qualifier chosen for the counts
// on either side of the 8-bit limit: 0x07/0x17 up to 255 points, 0x27 above it.
// A wrapped 8-bit count would announce the wrong number of points and make a
// strict peer mis-parse every following object header.
func TestStaticCountQualifierBoundaries(t *testing.T) {
	for _, count := range []int{255, 256, 300} {
		ost := newFullStaticOutstation(count)
		wantQual := byte(al.QualCount8)
		wantHeader := 4
		if count > 255 {
			wantQual = al.QualCount16
			wantHeader = 5
		}

		// G1V1 packed binary input: count is the bit count, payload is packed.
		g1 := ost.buildBinaryInputData(1)
		if g1[2] != wantQual {
			t.Fatalf("count %d: G1V1 qualifier = 0x%02X, want 0x%02X", count, g1[2], wantQual)
		}
		if got := staticCount(g1); got != count {
			t.Fatalf("count %d: G1V1 declared count = %d", count, got)
		}
		if want := wantHeader + (count+7)/8; len(g1) != want {
			t.Fatalf("count %d: G1V1 length = %d, want %d", count, len(g1), want)
		}

		// G10V1 packed binary output: same shape as G1V1.
		g10 := ost.buildBinaryOutputData(1)
		if got := staticCount(g10); got != count {
			t.Fatalf("count %d: G10V1 declared count = %d", count, got)
		}
		if want := wantHeader + (count+7)/8; len(g10) != want {
			t.Fatalf("count %d: G10V1 length = %d, want %d", count, len(g10), want)
		}

		// G30V1 and G20V1 sequential: flags + int32/uint32 per point.
		g30 := ost.buildAnalogInputData(1)
		if g30[2] != wantQual {
			t.Fatalf("count %d: G30V1 qualifier = 0x%02X, want 0x%02X", count, g30[2], wantQual)
		}
		if want := wantHeader + count*5; len(g30) != want {
			t.Fatalf("count %d: G30V1 length = %d, want %d", count, len(g30), want)
		}

		g20 := ost.buildCounterData(1)
		if got := staticCount(g20); got != count {
			t.Fatalf("count %d: G20V1 declared count = %d", count, got)
		}
		if want := wantHeader + count*5; len(g20) != want {
			t.Fatalf("count %d: G20V1 length = %d, want %d", count, len(g20), want)
		}

		// G40V1 sequential: float32 + flags per point.
		g40 := ost.buildAnalogOutputData(1)
		if got := staticCount(g40); got != count {
			t.Fatalf("count %d: G40V1 declared count = %d", count, got)
		}
		if want := wantHeader + count*5; len(g40) != want {
			t.Fatalf("count %d: G40V1 length = %d, want %d", count, len(g40), want)
		}
	}
}

// TestIndexedStaticBuilderFallsBackAbove255 verifies that the indexed builders
// (G1V2, G30V2, G20V5) emit the 0x17 index8 qualifier at 255 points but fall
// back to a sequential 0x27 block above it, because an 8-bit index prefix cannot
// address point 256.
func TestIndexedStaticBuilderFallsBackAbove255(t *testing.T) {
	small := newFullStaticOutstation(255)
	if got := small.buildBinaryInputData(2); got[2] != al.QualIndex8 {
		t.Fatalf("255 points: G1V2 qualifier = 0x%02X, want 0x17", got[2])
	}
	if got := small.buildAnalogInputData(2); got[2] != al.QualIndex8 {
		t.Fatalf("255 points: G30V2 qualifier = 0x%02X, want 0x17", got[2])
	}
	if got := small.buildCounterData(5); got[2] != al.QualIndex8 {
		t.Fatalf("255 points: G20V5 qualifier = 0x%02X, want 0x17", got[2])
	}

	big := newFullStaticOutstation(256)
	g1 := big.buildBinaryInputData(2)
	if g1[2] != al.QualCount16 {
		t.Fatalf("256 points: G1V2 qualifier = 0x%02X, want 0x27", g1[2])
	}
	if got := staticCount(g1); got != 256 {
		t.Fatalf("256 points: G1V2 declared count = %d", got)
	}
	// Sequential: 5 header octets + one flags octet per point, no index prefix.
	if want := 5 + 256; len(g1) != want {
		t.Fatalf("256 points: G1V2 length = %d, want %d (no index prefix)", len(g1), want)
	}

	g30 := big.buildAnalogInputData(2)
	if g30[2] != al.QualCount16 {
		t.Fatalf("256 points: G30V2 qualifier = 0x%02X, want 0x27", g30[2])
	}
	// int16 + flags per point, no index prefix.
	if want := 5 + 256*3; len(g30) != want {
		t.Fatalf("256 points: G30V2 length = %d, want %d", len(g30), want)
	}

	g20 := big.buildCounterData(5)
	if g20[2] != al.QualCount16 {
		t.Fatalf("256 points: G20V5 qualifier = 0x%02X, want 0x27", g20[2])
	}
	// uint16 + flags per point, no index prefix.
	if want := 5 + 256*3; len(g20) != want {
		t.Fatalf("256 points: G20V5 length = %d, want %d", len(g20), want)
	}
}

// TestStaticBlockAboveCountLimitSetsObjectUnknown verifies that a block that no
// DNP3 count qualifier can describe is omitted and reported through IIN
// OBJECT_UNKNOWN rather than sent with a wrapped count. A Class-0 read then
// never looks like a complete (but silently short) integrity response.
func TestStaticBlockAboveCountLimitSetsObjectUnknown(t *testing.T) {
	ost := NewOutstation(nil)
	ost.Initialize()

	// maxStaticPoints+1 packed binary inputs: the count octets cannot express
	// this block, so G1 must be dropped.
	h := &bigStaticHandler{inputs: make([]BinaryInput, maxStaticPoints+1)}
	ost.SetDataHandler(h)

	resp, err := ost.ProcessRequest(&al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: 1},
		FuncCode: al.FuncRead,
		Data:     []byte{1, 1, al.QualAllObjects},
	})
	if err != nil {
		t.Fatalf("ProcessRequest: %v", err)
	}
	respIIN, err := al.DecodeIIN(resp.Data)
	if err != nil {
		t.Fatalf("decode response IIN: %v", err)
	}
	if !respIIN.ObjectUnknown {
		t.Fatal("IIN OBJECT_UNKNOWN not set for an unrepresentable static block")
	}
	if len(resp.Data) != 2 {
		t.Fatalf("response object data = % X, want only the 2 IIN octets", resp.Data[2:])
	}

	// A normal-sized block in the same outstation must still be served, and the
	// error bit must not persist into the next response.
	h.inputs = make([]BinaryInput, 8)
	for i := range h.inputs {
		h.inputs[i] = BinaryInput{Value: i%2 == 0, Quality: BinaryQualityOnline}
	}
	resp2, err := ost.ProcessRequest(&al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: 2},
		FuncCode: al.FuncRead,
		Data:     []byte{1, 1, al.QualAllObjects},
	})
	if err != nil {
		t.Fatalf("second ProcessRequest: %v", err)
	}
	resp2IIN, err := al.DecodeIIN(resp2.Data)
	if err != nil {
		t.Fatalf("decode second response IIN: %v", err)
	}
	if resp2IIN.ObjectUnknown {
		t.Fatal("IIN OBJECT_UNKNOWN leaked into the following read")
	}
	if !bytes.HasPrefix(resp2.Data[2:], []byte{1, 1, al.QualCount8, 8}) {
		t.Fatalf("second response objects = % X, want a G1V1 count-8 header", resp2.Data[2:])
	}
}

// TestClassZeroFragmentsCarryCompleteStaticBlocks verifies that a Class-0
// response larger than one application fragment is split at object boundaries,
// with every declared count intact and every point delivered exactly once. Each
// block fits the budget on its own; only the combined response is oversize, so
// the splitter must advance by the real block span instead of treating the first
// sequential block as the whole remainder.
func TestClassZeroFragmentsCarryCompleteStaticBlocks(t *testing.T) {
	const points = 300
	ost := NewOutstation(&Config{
		OutstationAddress: 1024,
		MasterAddress:     1,
		MaxEventBuffers:   100,
		MaxFragmentSize:   defaultMaxAppFragment,
	})
	ost.Initialize()
	ost.SetDataHandler(&fullStaticHandler{
		binary:  make([]BinaryInput, points),
		analog:  make([]AnalogInput, points),
		counter: make([]Counter, points),
	})

	resp, err := ost.ProcessRequest(&al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: 3},
		FuncCode: al.FuncRead,
		Data:     []byte{60, 1, al.QualAllObjects},
	})
	if err != nil {
		t.Fatalf("ProcessRequest: %v", err)
	}
	if len(resp.Data)-2 <= defaultMaxAppFragment-appFragmentOverhead {
		t.Fatalf("test needs an oversize Class-0 response, got %d object octets", len(resp.Data)-2)
	}
	frags, err := ost.responseFragments(resp)
	if err != nil {
		t.Fatalf("responseFragments: %v", err)
	}
	if len(frags) < 2 {
		t.Fatalf("fragments = %d, want a split response", len(frags))
	}
	for i, f := range frags {
		if len(f.Data) > defaultMaxAppFragment-appFragmentOverhead {
			t.Fatalf("fragment %d is %d octets, over the %d-octet budget", i, len(f.Data), defaultMaxAppFragment-appFragmentOverhead)
		}
		if f.Control.FIR != (i == 0) || f.Control.FIN != (i == len(frags)-1) {
			t.Fatalf("fragment %d FIR/FIN flags wrong", i)
		}
	}

	// Re-walk every fragment's object data: each block must declare a count the
	// fragment actually carries, and the totals must match the served points.
	var gotBinary, gotAnalog, gotCounter int
	for _, f := range frags {
		off := 2 // skip IIN
		for off < len(f.Data) {
			h, consumed, err := al.DecodeObjectHeader(f.Data, off)
			if err != nil {
				t.Fatalf("decode object header at %d: %v (% X)", off, err, f.Data[off:])
			}
			off += consumed
			size, ok := staticValueSize(h.Group, h.Variation)
			if !ok {
				t.Fatalf("unexpected object %s in Class-0 response", al.DescribeObject(h.Group, h.Variation))
			}
			n := int(h.Count)
			body := n * size
			if (h.Group == 1 || h.Group == 10) && h.Variation == 1 {
				body = (n + 7) / 8
			}
			if off+body > len(f.Data) {
				t.Fatalf("block %s at %d declares %d points but the fragment ends at %d", al.DescribeObject(h.Group, h.Variation), off, n, len(f.Data))
			}
			switch h.Group {
			case 1:
				gotBinary += n
			case 30:
				gotAnalog += n
			case 20:
				gotCounter += n
			}
			off += body
		}
	}
	if gotBinary != points || gotAnalog != points || gotCounter != points {
		t.Fatalf("points across fragments: G1=%d G30=%d G20=%d, want %d each", gotBinary, gotAnalog, gotCounter, points)
	}
}

// TestUnsplittableStaticBlockFailsClosed verifies that a single sequential static
// block larger than the whole fragment budget is refused rather than emitted as
// an oversize fragment. Static blocks carry no per-point index, so splitting one
// would silently renumber its points.
func TestUnsplittableStaticBlockFailsClosed(t *testing.T) {
	ost := NewOutstation(&Config{
		OutstationAddress: 1024,
		MasterAddress:     1,
		MaxEventBuffers:   100,
		MaxFragmentSize:   64,
	})
	ost.Initialize()
	ost.SetDataHandler(&fullStaticHandler{analog: make([]AnalogInput, 300)})

	resp, err := ost.ProcessRequest(&al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: 4},
		FuncCode: al.FuncRead,
		Data:     []byte{60, 1, al.QualAllObjects},
	})
	if err != nil {
		t.Fatalf("ProcessRequest: %v", err)
	}
	if _, err := ost.responseFragments(resp); err == nil {
		t.Fatal("an unsplittable static block must fail closed, not be sent oversize")
	}
}

// TestStaticCountUpperBoundary pins the 16-bit limit itself: 65535 points still
// fit qualifier 0x27, and 65536 cannot be described by any DNP3 count qualifier
// and is therefore refused with IIN OBJECT_UNKNOWN rather than emitted with a
// wrapped count. Only packed G1V1 is exercised here because it keeps the fixture
// small (8192 payload octets) while still crossing the boundary.
func TestStaticCountUpperBoundary(t *testing.T) {
	ost := NewOutstation(nil)
	ost.Initialize()

	atLimit := &bigStaticHandler{inputs: make([]BinaryInput, maxStaticPoints)}
	ost.SetDataHandler(atLimit)

	block := ost.buildBinaryInputData(1)
	if block[2] != al.QualCount16 {
		t.Fatalf("qualifier at %d points = 0x%02X, want 0x27", maxStaticPoints, block[2])
	}
	if got := staticCount(block); got != maxStaticPoints {
		t.Fatalf("declared count at limit = %d, want %d", got, maxStaticPoints)
	}
	if want := 5 + (maxStaticPoints+7)/8; len(block) != want {
		t.Fatalf("block length at limit = %d, want %d", len(block), want)
	}

	overLimit := &bigStaticHandler{inputs: make([]BinaryInput, maxStaticPoints+1)}
	ost.SetDataHandler(overLimit)

	if block := ost.buildBinaryInputData(1); len(block) != 0 {
		t.Fatalf("builder emitted %d octets above the count limit, want none", len(block))
	}

	resp, err := ost.ProcessRequest(&al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: 9},
		FuncCode: al.FuncRead,
		Data:     []byte{1, 1, al.QualAllObjects},
	})
	if err != nil {
		t.Fatalf("ProcessRequest: %v", err)
	}
	iin, err := al.DecodeIIN(resp.Data)
	if err != nil {
		t.Fatalf("decode IIN: %v", err)
	}
	if !iin.ObjectUnknown {
		t.Fatalf("IIN = % X, want OBJECT_UNKNOWN above the count limit", resp.Data[:2])
	}
}

// staticCount reads the declared point count from a static block header.
func staticCount(block []byte) int {
	if len(block) < 4 {
		return -1
	}
	switch block[2] {
	case al.QualCount8, al.QualIndex8:
		return int(block[3])
	case al.QualCount16:
		if len(block) < 5 {
			return -1
		}
		return int(block[3]) | int(block[4])<<8
	}
	return -1
}
