package outstation

import (
	"bytes"
	"testing"
)

// vectorDataHandler supplies deterministic points for byte-exact static
// object-vector tests. It embeds DefaultDataHandler so unexercised interface
// methods keep their existing behavior.
type vectorDataHandler struct {
	DefaultDataHandler
	binaryInputs  []BinaryInput
	analogInputs  []AnalogInput
	counters      []Counter
	binaryOutputs []BinaryOutput
	analogOutputs []AnalogOutput
}

func (h *vectorDataHandler) GetBinaryInputs() []BinaryInput   { return h.binaryInputs }
func (h *vectorDataHandler) GetAnalogInputs() []AnalogInput   { return h.analogInputs }
func (h *vectorDataHandler) GetCounters() []Counter           { return h.counters }
func (h *vectorDataHandler) GetBinaryOutputs() []BinaryOutput { return h.binaryOutputs }
func (h *vectorDataHandler) GetAnalogOutputs() []AnalogOutput { return h.analogOutputs }

func newVectorOutstation(h DataHandler) *Outstation {
	o := NewOutstation(nil)
	o.Initialize()
	o.SetDataHandler(h)
	return o
}

func alternatingInputs(n int, start bool) []BinaryInput {
	out := make([]BinaryInput, n)
	for i := range out {
		on := (i%2 == 0) == start
		out[i] = BinaryInput{Value: on, Quality: BinaryQualityOnline}
	}
	return out
}

// TestBuildBinaryInputPackedG1V1 locks IEEE 1815 Group 1 Variation 1:
// qualifier 0x07 (8-bit count), ceil(n/8) data octets, state packed LSB-first
// with no per-point index or flags octet. The counts include byte boundaries
// (1, 8, 9, 16) that expose off-by-one layout errors.
func TestBuildBinaryInputPackedG1V1(t *testing.T) {
	tests := []struct {
		name   string
		values []bool
		packed []byte
	}{
		{"1 point on", []bool{true}, []byte{0x01}},
		{"4 points mixed", []bool{true, false, true, false}, []byte{0x05}},
		{"8 points alternating", []bool{true, false, true, false, true, false, true, false}, []byte{0x55}},
		{"9 points alternating", []bool{true, false, true, false, true, false, true, false, true}, []byte{0x55, 0x01}},
		{"16 points alternating", alternatingValues(16, true), []byte{0x55, 0x55}},
		{"16 points high/low bits", []bool{
			true, false, false, false, false, false, false, true,
			true, false, false, false, false, false, false, true,
		}, []byte{0x81, 0x81}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputs := make([]BinaryInput, len(tt.values))
			for i, v := range tt.values {
				inputs[i] = BinaryInput{Value: v, Quality: BinaryQualityOnline}
			}
			o := newVectorOutstation(&vectorDataHandler{binaryInputs: inputs})
			got := o.buildBinaryInputData(1)
			want := []byte{1, 1, 0x07, byte(len(tt.values))}
			want = append(want, tt.packed...)
			if !bytes.Equal(got, want) {
				t.Fatalf("G1V1 = % X, want % X", got, want)
			}
			if len(got) != 4+(len(tt.values)+7)/8 {
				t.Fatalf("G1V1 length = %d, want %d", len(got), 4+(len(tt.values)+7)/8)
			}
		})
	}
}

func alternatingValues(n int, start bool) []bool {
	out := make([]bool, n)
	for i := range out {
		out[i] = (i%2 == 0) == start
	}
	return out
}

// TestBuildBinaryInputG1V2Flags verifies that the flagged variation is encoded
// independently: one indexed octet plus one flags octet per point, with the
// state in bit 7 and quality in the low seven bits.
func TestBuildBinaryInputG1V2Flags(t *testing.T) {
	inputs := []BinaryInput{
		{Value: true, Quality: BinaryQualityOnline},
		{Value: false, Quality: BinaryQualityOnline},
		{Value: true, Quality: BinaryQualityOnline | BinaryQualityRestart},
	}
	o := newVectorOutstation(&vectorDataHandler{binaryInputs: inputs})
	got := o.buildBinaryInputData(2)
	want := []byte{1, 2, 0x17, 3, 0x00, 0x81, 0x01, 0x01, 0x02, 0xC1}
	if !bytes.Equal(got, want) {
		t.Fatalf("G1V2 = % X, want % X", got, want)
	}
}

// TestBuildAnalogInputG30V1 verifies flags-first, little-endian 32-bit signed
// values under the sequential count qualifier.
func TestBuildAnalogInputG30V1(t *testing.T) {
	inputs := []AnalogInput{
		{Value: 1000, Quality: AnalogQualityOnline},
		{Value: -2, Quality: AnalogQualityOnline},
		{Value: 0, Quality: AnalogQualityOnline},
	}
	o := newVectorOutstation(&vectorDataHandler{analogInputs: inputs})
	got := o.buildAnalogInputData(1)
	want := []byte{
		30, 1, 0x07, 3,
		0x01, 0xE8, 0x03, 0x00, 0x00,
		0x01, 0xFE, 0xFF, 0xFF, 0xFF,
		0x01, 0x00, 0x00, 0x00, 0x00,
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("G30V1 = % X, want % X", got, want)
	}
}

// TestMixedMultiHeaderClass0Response verifies that a Class 0 request emits one
// correctly bounded object block per static group, in order, with no mixed
// G1V1/G1V2 layouts and no surplus trailing octets.
func TestMixedMultiHeaderClass0Response(t *testing.T) {
	h := &vectorDataHandler{
		binaryInputs:  alternatingInputs(8, true),
		analogInputs:  make([]AnalogInput, 4),
		counters:      []Counter{{Value: 42, Quality: CounterQualityOnline}, {Value: 7, Quality: CounterQualityOnline}},
		binaryOutputs: []BinaryOutput{{Quality: BinaryQualityOnline}, {Quality: BinaryQualityOnline}, {Value: true, Quality: BinaryQualityOnline}, {Quality: BinaryQualityOnline}},
	}
	for i := range h.analogInputs {
		h.analogInputs[i].Quality = AnalogQualityOnline
	}
	o := newVectorOutstation(h)
	got := o.buildReadResponse([]byte{60, 1, 0x06})
	want := []byte{
		1, 1, 0x07, 8, 0x55,
		30, 1, 0x07, 4,
		0x01, 0, 0, 0, 0,
		0x01, 0, 0, 0, 0,
		0x01, 0, 0, 0, 0,
		0x01, 0, 0, 0, 0,
		20, 1, 0x07, 2,
		0x01, 0x2A, 0, 0, 0,
		0x01, 0x07, 0, 0, 0,
		10, 1, 0x07, 4, 0x04,
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("Class 0 objects = % X, want % X", got, want)
	}
}
