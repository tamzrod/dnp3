package outstation

import (
	"bytes"
	"testing"
)

// Group 40 (Analog Output Status) uses the same IEEE 1815 object shape as
// Group 30: flags octet first, then the value least-significant octet first.
// These byte-exact vectors pin the readback layout used by EG628 virtual
// analog-output setpoints. The vectors are written independently from the
// IEEE 1815 definition, not produced by the encoder under test.

type analogOutputStatusHandler struct {
	DefaultDataHandler
	outputs []AnalogOutput
}

func (h *analogOutputStatusHandler) GetAnalogOutputs() []AnalogOutput { return h.outputs }

// TestG40V1FlagsFirstLittleEndianValue pins two points with distinct flags and
// values. 123.456 == 0x42F6E979 as float32 -> LE 79 E9 F6 42.
func TestG40V1FlagsFirstLittleEndianValue(t *testing.T) {
	o := NewOutstation(nil)
	o.SetDataHandler(&analogOutputStatusHandler{outputs: []AnalogOutput{
		{Value: 123.456, Quality: 0x01},
		{Value: -0.5, Quality: 0x09}, // ONLINE | ROLLOVER-like second flag
	}})

	got := o.buildAnalogOutputData(1)
	want := []byte{
		0x28, 0x01, 0x07, 0x02, // Group 40 V1, count8, 2 points
		0x01, 0x79, 0xE9, 0xF6, 0x42,
		0x09, 0x00, 0x00, 0x00, 0xBF, // -0.5 float32 = BF000000
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("G40V1 wire mismatch\n got: % X\nwant: % X", got, want)
	}
}

// TestG40V2FlagsFirstLittleEndianValue pins the 16-bit integer variation:
// -2 == 0xFFFE -> LE FE FF, flags octet first.
func TestG40V2FlagsFirstLittleEndianValue(t *testing.T) {
	o := NewOutstation(nil)
	o.SetDataHandler(&analogOutputStatusHandler{outputs: []AnalogOutput{
		{Value: -2, Quality: 0x01},
		{Value: 1000, Quality: 0x09},
	}})

	got := o.buildAnalogOutputData(2)
	want := []byte{
		0x28, 0x02, 0x07, 0x02,
		0x01, 0xFE, 0xFF,
		0x09, 0xE8, 0x03,
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("G40V2 wire mismatch\n got: % X\nwant: % X", got, want)
	}
}
