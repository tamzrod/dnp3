package outstation

import (
	"math"
	"testing"

	"dnp3/internal/al"
)

// AO-003 — Group 41 Variation 3 (single-precision IEEE-754 float) outstation
// side.
//
// Verifies the outstation decodes the G41V3 record little-endian against
// independent literals (docs/verification/AO-001.md §3/§6.1), preserves the
// exact IEEE-754 bits of finite values (including signed zero and subnormals),
// rejects NaN/±Inf and malformed records without invoking the host handler, and
// echoes the object byte-exactly with a SUCCESS command-status octet only after
// an accepted handler call.

// g41v3RecordingHandler records analog outputs the outstation writes, including
// the typed value and variation, so G41V3 dispatch can be asserted exactly.
type g41v3RecordingHandler struct {
	DefaultDataHandler
	calls      int
	gotIndex   uint16
	gotValue   interface{}
	gotVar     uint8
	writeErr   error
	seenCROB   int
	lastAnalog interface{}
}

func (h *g41v3RecordingHandler) WriteAnalogOutput(index uint16, value interface{}, variation uint8) error {
	h.calls++
	h.gotIndex = index
	h.gotValue = value
	h.gotVar = variation
	h.lastAnalog = value
	return h.writeErr
}

func (h *g41v3RecordingHandler) WriteBinaryOutput(index uint16, crob *CROB) error {
	h.seenCROB++
	return nil
}

func newG41V3Outstation(h *g41v3RecordingHandler) *Outstation {
	o := NewOutstation(nil)
	o.Initialize()
	o.Start()
	o.SetDataHandler(h)
	return o
}

func g41v3DirectOperate(data []byte) *al.APDU {
	return &al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: 1},
		FuncCode: al.FuncDirectOperate,
		Data:     data,
	}
}

// g41v3Req builds a G41V3 DirectOperate record: header(5) + index(2 LE) +
// float32 bits(4 LE) + status(1).
func g41v3Req(index uint16, valueBits uint32) []byte {
	return []byte{
		0x29, 0x03, 0x28, 0x01, 0x00,
		byte(index), byte(index >> 8),
		byte(valueBits), byte(valueBits >> 8), byte(valueBits >> 16), byte(valueBits >> 24),
		0x00,
	}
}

// g41v3Echo builds the expected response data: IIN(2) + G41V3 echo object with a
// SUCCESS status octet.
func g41v3Echo(index uint16, valueBits uint32) []byte {
	return append([]byte{0x00, 0x00}, g41v3Req(index, valueBits)...)
}

// TestOutstationG41V3DecodeGolden confirms the outstation decodes the G41V3
// float32 value little-endian and passes a typed float32 to the handler, against
// the independent golden for +1.0 (value bytes 00 00 80 3F). The response echoes
// the object byte-exactly with a SUCCESS status octet.
func TestOutstationG41V3DecodeGolden(t *testing.T) {
	h := &g41v3RecordingHandler{}
	o := newG41V3Outstation(h)

	resp, err := o.ProcessRequest(g41v3DirectOperate(g41v3Req(0, math.Float32bits(1.0))))
	if err != nil {
		t.Fatalf("DirectOperate G41V3 failed: %v", err)
	}
	if resp == nil {
		t.Fatal("expected a response")
	}
	if h.calls != 1 {
		t.Fatalf("handler calls = %d, want 1", h.calls)
	}
	f, ok := h.gotValue.(float32)
	if !ok {
		t.Fatalf("decoded value type = %T, want float32", h.gotValue)
	}
	if math.Float32bits(f) != math.Float32bits(1.0) {
		t.Fatalf("decoded value bits = %08X, want %08X", math.Float32bits(f), math.Float32bits(1.0))
	}
	if h.gotIndex != 0 {
		t.Fatalf("index = %d, want 0", h.gotIndex)
	}
	if h.gotVar != 3 {
		t.Fatalf("variation = %d, want 3", h.gotVar)
	}
	want := g41v3Echo(0, math.Float32bits(1.0))
	if string(resp.Data) != string(want) {
		t.Fatalf("G41V3 echo = %X, want %X", resp.Data, want)
	}
}

// TestOutstationG41V3DecodeBitExact confirms finite G41V3 values round-trip
// bit-exactly through decode and byte-exact echo: signed zero, the smallest
// positive/negative subnormals, the minimum normal and the maximum finite value
// (independent goldens, AO-001 §6.1).
func TestOutstationG41V3DecodeBitExact(t *testing.T) {
	cases := []struct {
		name string
		bits uint32
	}{
		{"+0.0", 0x00000000},
		{"-0.0", 0x80000000},
		{"smallest positive subnormal", 0x00000001},
		{"smallest negative subnormal", 0x80000001},
		{"min normal", 0x00800000},
		{"max finite", 0x7F7FFFFF},
		{"negative max finite", 0xFF7FFFFF},
		{"-0.5", math.Float32bits(-0.5)},
		{"2300.0", math.Float32bits(2300.0)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := &g41v3RecordingHandler{}
			o := newG41V3Outstation(h)
			resp, err := o.ProcessRequest(g41v3DirectOperate(g41v3Req(0x1234, tc.bits)))
			if err != nil {
				t.Fatalf("DirectOperate G41V3 (%s) failed: %v", tc.name, err)
			}
			f, ok := h.gotValue.(float32)
			if !ok {
				t.Fatalf("decoded value type = %T, want float32", h.gotValue)
			}
			if math.Float32bits(f) != tc.bits {
				t.Fatalf("decoded bits = %08X, want %08X (no rounding/clamping)", math.Float32bits(f), tc.bits)
			}
			want := g41v3Echo(0x1234, tc.bits)
			if string(resp.Data) != string(want) {
				t.Fatalf("echo = %X, want %X", resp.Data, want)
			}
		})
	}
}

// TestOutstationG41V3RejectsNonFinite verifies an inbound NaN/±Inf G41V3 wire
// value is rejected before any handler side effect and never echoed as SUCCESS
// (owner-approved AO-003 float policy).
func TestOutstationG41V3RejectsNonFinite(t *testing.T) {
	cases := []struct {
		name string
		bits uint32
	}{
		{"NaN", 0x7FC00000},
		{"+Inf", 0x7F800000},
		{"-Inf", 0xFF800000},
		{"negative NaN", 0xFFC00000},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := &g41v3RecordingHandler{}
			o := newG41V3Outstation(h)
			resp, err := o.ProcessRequest(g41v3DirectOperate(g41v3Req(0, tc.bits)))
			if err == nil {
				t.Fatalf("expected rejection for %s, got resp=%v", tc.name, resp)
			}
			if resp != nil {
				t.Fatalf("expected nil response (no SUCCESS echo), got %v", resp)
			}
			if h.calls != 0 || h.seenCROB != 0 {
				t.Fatalf("handler invoked for a non-finite G41V3 value: analogCalls=%d crobCalls=%d", h.calls, h.seenCROB)
			}
			if h.lastAnalog != nil {
				t.Fatalf("state mutated: lastAnalog=%v", h.lastAnalog)
			}
		})
	}
}

// TestOutstationG41V3RejectionsNoHandlerCall verifies unsupported G41 variations
// and malformed G41V3 lengths are rejected with no host-handler call, no CROB
// side effect and no SUCCESS echo (fail closed).
func TestOutstationG41V3RejectionsNoHandlerCall(t *testing.T) {
	cases := []struct {
		name string
		data []byte
	}{
		{"variation 4 (unsupported)", []byte{0x29, 0x04, 0x28, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xF0, 0x3F, 0x00}},
		{"count 0", []byte{0x29, 0x03, 0x28, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x80, 0x3F, 0x00}},
		{"count 2", []byte{0x29, 0x03, 0x28, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x80, 0x3F, 0x00}},
		{"truncated value (3 of 4)", []byte{0x29, 0x03, 0x28, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x80}},
		{"truncated status (4 of 5)", []byte{0x29, 0x03, 0x28, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x80, 0x3F}},
		{"surplus byte", []byte{0x29, 0x03, 0x28, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x80, 0x3F, 0x00, 0xEE}},
		{"missing index", []byte{0x29, 0x03, 0x28, 0x01, 0x00}},
		{"bad qualifier", []byte{0x29, 0x03, 0x06, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x80, 0x3F, 0x00}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := &g41v3RecordingHandler{}
			o := newG41V3Outstation(h)
			resp, err := o.ProcessRequest(g41v3DirectOperate(tc.data))
			if err == nil {
				t.Fatalf("expected rejection for %s, got resp=%v", tc.name, resp)
			}
			if resp != nil {
				t.Fatalf("expected nil response (no SUCCESS echo), got %v", resp)
			}
			if h.calls != 0 || h.seenCROB != 0 {
				t.Fatalf("handler invoked for a rejected G41V3 record: analogCalls=%d crobCalls=%d", h.calls, h.seenCROB)
			}
			if h.lastAnalog != nil {
				t.Fatalf("state mutated: lastAnalog=%v", h.lastAnalog)
			}
		})
	}
}

// TestOutstationG41V3HandlerRefusalNotSuccess verifies a handler error yields an
// error response (never a SUCCESS echo), so the master cannot misread a refusal
// as success.
func TestOutstationG41V3HandlerRefusalNotSuccess(t *testing.T) {
	h := &g41v3RecordingHandler{writeErr: ErrInvalidRequest}
	o := newG41V3Outstation(h)
	resp, err := o.ProcessRequest(g41v3DirectOperate(g41v3Req(0, math.Float32bits(1.0))))
	if err == nil {
		t.Fatalf("expected handler refusal to surface as an error, got resp=%v", resp)
	}
	if resp != nil {
		t.Fatalf("handler refusal must not produce a SUCCESS echo, got %v", resp)
	}
}

// TestOutstationG41V3NoDataWriterFailsClosed verifies a G41V3 DirectOperate to an
// outstation whose data handler is not a DataWriter fails closed: an error is
// returned, no handler runs and no SUCCESS echo is built.
func TestOutstationG41V3NoDataWriterFailsClosed(t *testing.T) {
	o := newReadOnlyOutstation()
	resp, err := o.ProcessRequest(g41v3DirectOperate(g41v3Req(0, math.Float32bits(1.0))))
	if err == nil {
		t.Fatalf("expected fail-closed error for a missing DataWriter, got resp=%v", resp)
	}
	if resp != nil {
		t.Fatalf("expected nil response (no SUCCESS echo), got %v", resp)
	}
}

// TestOutstationG41V3DoesNotTouchCROB confirms a G41V3 command never reaches the
// binary-output path (no cross-group leakage).
func TestOutstationG41V3DoesNotTouchCROB(t *testing.T) {
	h := &g41v3RecordingHandler{}
	o := newG41V3Outstation(h)
	if _, err := o.ProcessRequest(g41v3DirectOperate(g41v3Req(0, math.Float32bits(1.0)))); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if h.seenCROB != 0 {
		t.Fatalf("G41V3 command reached WriteBinaryOutput %d times", h.seenCROB)
	}
}

// TestOutstationG41V3CountOneExecutesOnce is the positive control: a valid
// single-point G41V3 request executes exactly once and returns the byte-exact
// SUCCESS echo.
func TestOutstationG41V3CountOneExecutesOnce(t *testing.T) {
	h := &g41v3RecordingHandler{}
	o := newG41V3Outstation(h)
	resp, err := o.ProcessRequest(g41v3DirectOperate(g41v3Req(0, math.Float32bits(1.0))))
	if err != nil {
		t.Fatalf("valid count=1 G41V3 request rejected: %v", err)
	}
	if h.calls != 1 || h.seenCROB != 0 {
		t.Fatalf("analogCalls=%d crobCalls=%d, want 1/0", h.calls, h.seenCROB)
	}
	f, ok := h.gotValue.(float32)
	if !ok || math.Float32bits(f) != math.Float32bits(1.0) {
		t.Fatalf("executed value = %v (%T), want float32(1.0)", h.gotValue, h.gotValue)
	}
	want := g41v3Echo(0, math.Float32bits(1.0))
	if string(resp.Data) != string(want) {
		t.Fatalf("echo = %X, want %X", resp.Data, want)
	}
}
