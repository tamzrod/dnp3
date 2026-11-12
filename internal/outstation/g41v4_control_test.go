package outstation

import (
	"math"
	"testing"

	"dnp3/internal/al"
)

// AO-003 — Group 41 Variation 4 (double-precision IEEE-754 float) outstation
// side.
//
// Verifies the outstation decodes the G41V4 record little-endian against
// independent literals (docs/verification/AO-001.md §3.1/§6.1), preserves the
// exact IEEE-754 bits of finite values (including signed zero and subnormals),
// rejects NaN/±Inf and malformed records without invoking the host handler, and
// echoes the object byte-exactly with a SUCCESS command-status octet only after
// an accepted handler call.

// g41v4RecordingHandler records analog outputs the outstation writes, including
// the typed value and variation, so G41V4 dispatch can be asserted exactly.
type g41v4RecordingHandler struct {
	DefaultDataHandler
	calls      int
	gotIndex   uint16
	gotValue   interface{}
	gotVar     uint8
	writeErr   error
	seenCROB   int
	lastAnalog interface{}
}

func (h *g41v4RecordingHandler) WriteAnalogOutput(index uint16, value interface{}, variation uint8) error {
	h.calls++
	h.gotIndex = index
	h.gotValue = value
	h.gotVar = variation
	h.lastAnalog = value
	return h.writeErr
}

func (h *g41v4RecordingHandler) WriteBinaryOutput(index uint16, crob *CROB) error {
	h.seenCROB++
	return nil
}

func newG41V4Outstation(h *g41v4RecordingHandler) *Outstation {
	o := NewOutstation(nil)
	o.Initialize()
	o.Start()
	o.SetDataHandler(h)
	return o
}

func g41v4DirectOperate(data []byte) *al.APDU {
	return &al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: 1},
		FuncCode: al.FuncDirectOperate,
		Data:     data,
	}
}

// g41v4Req builds a G41V4 DirectOperate record: header(5) + index(2 LE) +
// float64 bits(8 LE) + status(1).
func g41v4Req(index uint16, valueBits uint64) []byte {
	out := []byte{
		0x29, 0x04, 0x28, 0x01, 0x00,
		byte(index), byte(index >> 8),
	}
	for i := 0; i < 8; i++ {
		out = append(out, byte(valueBits>>(8*uint(i))))
	}
	return append(out, 0x00)
}

// g41v4Echo builds the expected response data: IIN(2) + G41V4 echo object with a
// SUCCESS status octet.
func g41v4Echo(index uint16, valueBits uint64) []byte {
	return append([]byte{0x00, 0x00}, g41v4Req(index, valueBits)...)
}

// TestOutstationG41V4DecodeGolden confirms the outstation decodes the G41V4
// float64 value little-endian and passes a typed float64 to the handler, against
// the independent golden for 123.456789 (value bytes 0B 0B EE 07 3C DD 5E 40).
// The response echoes the object byte-exactly with a SUCCESS status octet.
func TestOutstationG41V4DecodeGolden(t *testing.T) {
	h := &g41v4RecordingHandler{}
	o := newG41V4Outstation(h)

	bits := math.Float64bits(123.456789)
	resp, err := o.ProcessRequest(g41v4DirectOperate(g41v4Req(0, bits)))
	if err != nil {
		t.Fatalf("DirectOperate G41V4 failed: %v", err)
	}
	if resp == nil {
		t.Fatal("expected a response")
	}
	if h.calls != 1 {
		t.Fatalf("handler calls = %d, want 1", h.calls)
	}
	f, ok := h.gotValue.(float64)
	if !ok {
		t.Fatalf("decoded value type = %T, want float64", h.gotValue)
	}
	if math.Float64bits(f) != bits {
		t.Fatalf("decoded value bits = %016X, want %016X", math.Float64bits(f), bits)
	}
	if h.gotIndex != 0 {
		t.Fatalf("index = %d, want 0", h.gotIndex)
	}
	if h.gotVar != 4 {
		t.Fatalf("variation = %d, want 4", h.gotVar)
	}
	want := g41v4Echo(0, bits)
	if string(resp.Data) != string(want) {
		t.Fatalf("G41V4 echo = %X, want %X", resp.Data, want)
	}
}

// TestOutstationG41V4DecodeBitExact confirms finite G41V4 values round-trip
// bit-exactly through decode and byte-exact echo: signed zero, the smallest
// positive/negative subnormals, the minimum normal and the maximum finite value
// (independent goldens, AO-001 §6.1).
func TestOutstationG41V4DecodeBitExact(t *testing.T) {
	cases := []struct {
		name string
		bits uint64
	}{
		{"+0.0", 0x0000000000000000},
		{"-0.0", 0x8000000000000000},
		{"smallest positive subnormal", 0x0000000000000001},
		{"smallest negative subnormal", 0x8000000000000001},
		{"min normal", 0x0010000000000000},
		{"max finite", 0x7FEFFFFFFFFFFFFF},
		{"negative max finite", 0xFFEFFFFFFFFFFFFF},
		{"-0.5", math.Float64bits(-0.5)},
		{"123.456789", math.Float64bits(123.456789)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := &g41v4RecordingHandler{}
			o := newG41V4Outstation(h)
			resp, err := o.ProcessRequest(g41v4DirectOperate(g41v4Req(0x1234, tc.bits)))
			if err != nil {
				t.Fatalf("DirectOperate G41V4 (%s) failed: %v", tc.name, err)
			}
			f, ok := h.gotValue.(float64)
			if !ok {
				t.Fatalf("decoded value type = %T, want float64", h.gotValue)
			}
			if math.Float64bits(f) != tc.bits {
				t.Fatalf("decoded bits = %016X, want %016X (no rounding/clamping)", math.Float64bits(f), tc.bits)
			}
			want := g41v4Echo(0x1234, tc.bits)
			if string(resp.Data) != string(want) {
				t.Fatalf("echo = %X, want %X", resp.Data, want)
			}
		})
	}
}

// TestOutstationG41V4RejectsNonFinite verifies an inbound NaN/±Inf G41V4 wire
// value is rejected before any handler side effect and never echoed as SUCCESS
// (owner-approved AO-003 float policy).
func TestOutstationG41V4RejectsNonFinite(t *testing.T) {
	cases := []struct {
		name string
		bits uint64
	}{
		{"NaN", 0x7FF8000000000000},
		{"+Inf", 0x7FF0000000000000},
		{"-Inf", 0xFFF0000000000000},
		{"negative NaN", 0xFFF8000000000000},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := &g41v4RecordingHandler{}
			o := newG41V4Outstation(h)
			resp, err := o.ProcessRequest(g41v4DirectOperate(g41v4Req(0, tc.bits)))
			if err == nil {
				t.Fatalf("expected rejection for %s, got resp=%v", tc.name, resp)
			}
			if resp != nil {
				t.Fatalf("expected nil response (no SUCCESS echo), got %v", resp)
			}
			if h.calls != 0 || h.seenCROB != 0 {
				t.Fatalf("handler invoked for a non-finite G41V4 value: analogCalls=%d crobCalls=%d", h.calls, h.seenCROB)
			}
			if h.lastAnalog != nil {
				t.Fatalf("state mutated: lastAnalog=%v", h.lastAnalog)
			}
		})
	}
}

// TestOutstationG41V4RejectionsNoHandlerCall verifies unsupported G41 variations
// and malformed G41V4 lengths are rejected with no host-handler call, no CROB
// side effect and no SUCCESS echo (fail closed).
func TestOutstationG41V4RejectionsNoHandlerCall(t *testing.T) {
	cases := []struct {
		name string
		data []byte
	}{
		{"variation 5 (unsupported)", []byte{0x29, 0x05, 0x28, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xF0, 0x3F, 0x00}},
		{"variation 3 truncated status (11 of 12)", []byte{0x29, 0x03, 0x28, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x80, 0x3F}},
		{"count 0", []byte{0x29, 0x04, 0x28, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xF0, 0x3F, 0x00}},
		{"count 2", []byte{0x29, 0x04, 0x28, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xF0, 0x3F, 0x00}},
		{"truncated value (7 of 8)", []byte{0x29, 0x04, 0x28, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xF0}},
		{"truncated status (8 of 9)", []byte{0x29, 0x04, 0x28, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xF0, 0x3F}},
		{"surplus byte", []byte{0x29, 0x04, 0x28, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xF0, 0x3F, 0x00, 0xEE}},
		{"missing index", []byte{0x29, 0x04, 0x28, 0x01, 0x00}},
		{"bad qualifier", []byte{0x29, 0x04, 0x06, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xF0, 0x3F, 0x00}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := &g41v4RecordingHandler{}
			o := newG41V4Outstation(h)
			resp, err := o.ProcessRequest(g41v4DirectOperate(tc.data))
			if err == nil {
				t.Fatalf("expected rejection for %s, got resp=%v", tc.name, resp)
			}
			if resp != nil {
				t.Fatalf("expected nil response (no SUCCESS echo), got %v", resp)
			}
			if h.calls != 0 || h.seenCROB != 0 {
				t.Fatalf("handler invoked for a rejected G41V4 record: analogCalls=%d crobCalls=%d", h.calls, h.seenCROB)
			}
			if h.lastAnalog != nil {
				t.Fatalf("state mutated: lastAnalog=%v", h.lastAnalog)
			}
		})
	}
}

// TestOutstationG41V4WritePathRejectsWithNoHandlerCall verifies the FuncWrite
// (WRITE) path fails closed for a well-formed G41V4 record: the approved G41
// command profile permits only FuncDirectOperate with response, so an
// application WRITE carrying G41 must reject before any host analog command
// handler call or state mutation, with no success acknowledgement.
// (AO-004 pre-PR corrective F2.)
func TestOutstationG41V4WritePathRejectsWithNoHandlerCall(t *testing.T) {
	h := &g41v4RecordingHandler{}
	o := newG41V4Outstation(h)
	bits := math.Float64bits(123.456789)
	data := []byte{0x29, 0x04, 0x28, 0x01, 0x00, 0x00, 0x00}
	for i := 0; i < 8; i++ {
		data = append(data, byte(bits>>(8*uint(i))))
	}
	data = append(data, 0x00)
	resp, err := o.ProcessRequest(&al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: 1},
		FuncCode: al.FuncWrite,
		Data:     data,
	})
	if err == nil {
		t.Fatalf("expected G41V4 WRITE to be rejected, got resp=%v", resp)
	}
	if resp != nil {
		t.Fatalf("G41V4 WRITE rejection must not acknowledge success, got %v", resp)
	}
	if h.calls != 0 || h.seenCROB != 0 {
		t.Fatalf("handler invoked for a rejected G41V4 WRITE: analogCalls=%d crobCalls=%d", h.calls, h.seenCROB)
	}
	if h.lastAnalog != nil {
		t.Fatalf("state mutated by a rejected G41V4 WRITE: lastAnalog=%v", h.lastAnalog)
	}
	if !o.iin.ParameterError {
		t.Fatalf("expected ParameterError IIN for a rejected G41V4 WRITE")
	}
}

// TestOutstationG41V4WritePathRejectsNonFiniteAndTruncation verifies the FuncWrite
// path fails closed for a non-finite G41V4 value and for a truncated record, with
// no handler call.
func TestOutstationG41V4WritePathRejectsNonFiniteAndTruncation(t *testing.T) {
	cases := []struct {
		name string
		data []byte
	}{
		{"NaN", []byte{0x29, 0x04, 0x28, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xF8, 0x7F, 0x00}},
		{"truncated value", []byte{0x29, 0x04, 0x28, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xF0}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := &g41v4RecordingHandler{}
			o := newG41V4Outstation(h)
			resp, err := o.ProcessRequest(&al.APDU{
				Control:  al.AppControl{FIR: true, FIN: true, Seq: 1},
				FuncCode: al.FuncWrite,
				Data:     tc.data,
			})
			if err == nil {
				t.Fatalf("expected rejection for %s, got resp=%v", tc.name, resp)
			}
			if resp != nil {
				t.Fatalf("expected nil response, got %v", resp)
			}
			if h.calls != 0 || h.seenCROB != 0 {
				t.Fatalf("handler invoked for a rejected WRITE G41V4 record: analogCalls=%d crobCalls=%d", h.calls, h.seenCROB)
			}
		})
	}
}

// TestOutstationG41V4HandlerRefusalNotSuccess verifies a handler error yields an
// error response (never a SUCCESS echo), so the master cannot misread a refusal
// as success.
func TestOutstationG41V4HandlerRefusalNotSuccess(t *testing.T) {
	h := &g41v4RecordingHandler{writeErr: ErrInvalidRequest}
	o := newG41V4Outstation(h)
	resp, err := o.ProcessRequest(g41v4DirectOperate(g41v4Req(0, math.Float64bits(1.0))))
	if err == nil {
		t.Fatalf("expected handler refusal to surface as an error, got resp=%v", resp)
	}
	if resp != nil {
		t.Fatalf("handler refusal must not produce a SUCCESS echo, got %v", resp)
	}
}

// TestOutstationG41V4NoDataWriterFailsClosed verifies a G41V4 DirectOperate to an
// outstation whose data handler is not a DataWriter fails closed: an error is
// returned, no handler runs and no SUCCESS echo is built.
func TestOutstationG41V4NoDataWriterFailsClosed(t *testing.T) {
	o := newReadOnlyOutstation()
	resp, err := o.ProcessRequest(g41v4DirectOperate(g41v4Req(0, math.Float64bits(1.0))))
	if err == nil {
		t.Fatalf("expected fail-closed error for a missing DataWriter, got resp=%v", resp)
	}
	if resp != nil {
		t.Fatalf("expected nil response (no SUCCESS echo), got %v", resp)
	}
}

// TestOutstationG41V4DoesNotTouchCROB confirms a G41V4 command never reaches the
// binary-output path (no cross-group leakage).
func TestOutstationG41V4DoesNotTouchCROB(t *testing.T) {
	h := &g41v4RecordingHandler{}
	o := newG41V4Outstation(h)
	if _, err := o.ProcessRequest(g41v4DirectOperate(g41v4Req(0, math.Float64bits(1.0)))); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if h.seenCROB != 0 {
		t.Fatalf("G41V4 command reached WriteBinaryOutput %d times", h.seenCROB)
	}
}

// TestOutstationG41V4CountOneExecutesOnce is the positive control: a valid
// single-point G41V4 request executes exactly once and returns the byte-exact
// SUCCESS echo.
func TestOutstationG41V4CountOneExecutesOnce(t *testing.T) {
	h := &g41v4RecordingHandler{}
	o := newG41V4Outstation(h)
	resp, err := o.ProcessRequest(g41v4DirectOperate(g41v4Req(0, math.Float64bits(1.0))))
	if err != nil {
		t.Fatalf("valid count=1 G41V4 request rejected: %v", err)
	}
	if h.calls != 1 || h.seenCROB != 0 {
		t.Fatalf("analogCalls=%d crobCalls=%d, want 1/0", h.calls, h.seenCROB)
	}
	f, ok := h.gotValue.(float64)
	if !ok || math.Float64bits(f) != math.Float64bits(1.0) {
		t.Fatalf("executed value = %v (%T), want float64(1.0)", h.gotValue, h.gotValue)
	}
	want := g41v4Echo(0, math.Float64bits(1.0))
	if string(resp.Data) != string(want) {
		t.Fatalf("echo = %X, want %X", resp.Data, want)
	}
}
