package outstation

import (
	"testing"

	"dnp3/internal/al"
)

// AO-004 pre-PR corrective F2 — Group 41 FuncWrite boundary.
//
// The approved G41 command profile permits only FuncDirectOperate WITH RESPONSE.
// An incoming application WRITE (FuncWrite) carrying any Group 41 object —
// variations 1–4, well-formed, malformed, truncated or multi-point — must be
// rejected before any host analog command handler call or state mutation, and
// must never be acknowledged as success. These tests exercise the reachable
// ProcessRequest -> handleWrite path, not a private helper.

// g41WriteRecordingHandler counts analog/binary handler calls and records the
// last analog value so tests can assert zero side effects on rejection.
type g41WriteRecordingHandler struct {
	DefaultDataHandler
	analogCalls int
	crobCalls   int
	lastAnalog  interface{}
}

func (h *g41WriteRecordingHandler) WriteAnalogOutput(index uint16, value interface{}, variation uint8) error {
	h.analogCalls++
	h.lastAnalog = value
	return nil
}

func (h *g41WriteRecordingHandler) WriteBinaryOutput(index uint16, crob *CROB) error {
	h.crobCalls++
	return nil
}

func newG41WriteOutstation(h *g41WriteRecordingHandler) *Outstation {
	o := NewOutstation(nil)
	o.Initialize()
	o.Start()
	o.SetDataHandler(h)
	return o
}

func g41WriteRequest(data []byte) *al.APDU {
	return &al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: 1},
		FuncCode: al.FuncWrite,
		Data:     data,
	}
}

// g41WriteRecord builds a one-point 0x28-qualified Group 41 WRITE record:
// header(5) + index(2) + value bytes + trailing status octet.
func g41WriteRecord(variation uint8, index uint16, valueBytes []byte) []byte {
	data := []byte{0x29, variation, 0x28, 0x01, 0x00, byte(index), byte(index >> 8)}
	data = append(data, valueBytes...)
	data = append(data, 0x00)
	return data
}

// TestOutstationG41WriteRejectsAllVariationsWithNoHandlerCall verifies that a
// well-formed one-point Group 41 WRITE for each implemented variation (V1 int32,
// V2 int16, V3 float32, V4 float64) is rejected with zero handler calls, zero
// state mutation, no success response and a ParameterError IIN.
func TestOutstationG41WriteRejectsAllVariationsWithNoHandlerCall(t *testing.T) {
	cases := []struct {
		name      string
		variation uint8
		value     []byte
	}{
		{"V1_int32", 1, []byte{0x2A, 0x00, 0x00, 0x00}},
		{"V2_int16", 2, []byte{0x2A, 0x00}},
		{"V3_float32", 3, []byte{0x00, 0x00, 0x80, 0x3F}},
		{"V4_float64", 4, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xF0, 0x3F}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := &g41WriteRecordingHandler{}
			o := newG41WriteOutstation(h)

			resp, err := o.ProcessRequest(g41WriteRequest(g41WriteRecord(tc.variation, 0, tc.value)))
			if err == nil {
				t.Fatalf("expected G41 WRITE to be rejected, got resp=%v", resp)
			}
			if resp != nil {
				t.Fatalf("G41 WRITE rejection must not acknowledge success, got %v", resp)
			}
			if h.analogCalls != 0 || h.crobCalls != 0 {
				t.Fatalf("handler invoked for a rejected G41 WRITE: analogCalls=%d crobCalls=%d", h.analogCalls, h.crobCalls)
			}
			if h.lastAnalog != nil {
				t.Fatalf("state mutated by a rejected G41 WRITE: lastAnalog=%v", h.lastAnalog)
			}
			if !o.iin.ParameterError {
				t.Fatalf("expected ParameterError IIN for a rejected G41 WRITE")
			}
		})
	}
}

// TestOutstationG41WriteRejectsMalformedAndMultiPoint verifies that malformed,
// truncated and multi-point Group 41 WRITE records are rejected with zero
// side effects, and that a non-G41 CROB WRITE on the same outstation still
// works (the gate is G41-specific).
func TestOutstationG41WriteRejectsMalformedAndMultiPoint(t *testing.T) {
	cases := []struct {
		name string
		data []byte
	}{
		{"truncated V4 value", []byte{0x29, 0x04, 0x28, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xF0}},
		{"V4 non-finite (NaN)", []byte{0x29, 0x04, 0x28, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xF8, 0x7F, 0x00}},
		{"multi-point V2 count 2", []byte{0x29, 0x02, 0x28, 0x02, 0x00, 0x00, 0x00, 0x2A, 0x00, 0x00, 0x01, 0x00, 0x2B, 0x00, 0x00}},
		{"unsupported variation 5", g41WriteRecord(5, 0, []byte{0x00, 0x00, 0x00, 0x00})},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := &g41WriteRecordingHandler{}
			o := newG41WriteOutstation(h)

			resp, err := o.ProcessRequest(g41WriteRequest(tc.data))
			if err == nil {
				t.Fatalf("expected rejection for %s, got resp=%v", tc.name, resp)
			}
			if resp != nil {
				t.Fatalf("rejection for %s must not produce a success response, got %v", tc.name, resp)
			}
			if h.analogCalls != 0 || h.crobCalls != 0 {
				t.Fatalf("handler invoked for a rejected G41 WRITE (%s): analogCalls=%d crobCalls=%d", tc.name, h.analogCalls, h.crobCalls)
			}
			if h.lastAnalog != nil {
				t.Fatalf("state mutated by a rejected G41 WRITE (%s): lastAnalog=%v", tc.name, h.lastAnalog)
			}
		})
	}
}

// TestOutstationG41WriteDoesNotBreakCROBWrite verifies the G41 gate does not
// regress the already-supported non-G41 FuncWrite path: a well-formed G12V1
// CROB WRITE still dispatches to the binary handler exactly once.
func TestOutstationG41WriteDoesNotBreakCROBWrite(t *testing.T) {
	h := &g41WriteRecordingHandler{}
	o := newG41WriteOutstation(h)

	data := []byte{0x0C, 0x01, 0x00, 0x01, 0x00, 0x00, 0x01, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	resp, err := o.ProcessRequest(g41WriteRequest(data))
	if err != nil {
		t.Fatalf("CROB WRITE failed: %v", err)
	}
	if resp == nil || resp.FuncCode != al.FuncResponse {
		t.Fatalf("expected a FuncResponse for a CROB WRITE, got %v", resp)
	}
	if h.crobCalls != 1 || h.analogCalls != 0 {
		t.Fatalf("crobCalls=%d analogCalls=%d, want 1/0", h.crobCalls, h.analogCalls)
	}
}
