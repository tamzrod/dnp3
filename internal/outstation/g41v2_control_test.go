package outstation

import (
	"testing"

	"dnp3/internal/al"
)

// AO-003 — Group 41 Variation 2 (signed int16) outstation side.
//
// Verifies the outstation decodes the G41V2 record little-endian against
// independent literals (docs/verification/AO-001.md §3/§6.1), rejects
// unsupported variations and malformed lengths without invoking the host
// handler, and echoes the object with a SUCCESS command-status octet only after
// an accepted handler call.

// g41v2RecordingHandler records analog outputs the outstation writes, including
// the typed value and variation, so G41V2 dispatch can be asserted exactly.
type g41v2RecordingHandler struct {
	DefaultDataHandler
	calls      int
	gotIndex   uint16
	gotValue   interface{}
	gotVar     uint8
	writeErr   error
	seenCROB   int
	lastAnalog interface{}
}

func (h *g41v2RecordingHandler) WriteAnalogOutput(index uint16, value interface{}, variation uint8) error {
	h.calls++
	h.gotIndex = index
	h.gotValue = value
	h.gotVar = variation
	h.lastAnalog = value
	return h.writeErr
}

func (h *g41v2RecordingHandler) WriteBinaryOutput(index uint16, crob *CROB) error {
	h.seenCROB++
	return nil
}

func newG41V2Outstation(h *g41v2RecordingHandler) *Outstation {
	o := NewOutstation(nil)
	o.Initialize()
	o.Start()
	o.SetDataHandler(h)
	return o
}

func g41v2DirectOperate(data []byte) *al.APDU {
	return &al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: 1},
		FuncCode: al.FuncDirectOperate,
		Data:     data,
	}
}

// TestOutstationG41V2DecodeGolden confirms the outstation decodes the G41V2
// int16 value little-endian and passes a typed int16 to the handler, against the
// independent golden for 42 (value bytes 2A 00). The response echoes the object
// byte-exactly with a SUCCESS status octet.
func TestOutstationG41V2DecodeGolden(t *testing.T) {
	h := &g41v2RecordingHandler{}
	o := newG41V2Outstation(h)

	// header(5) + index(2) + int16 LE(2) + status(1)
	req := g41v2DirectOperate([]byte{0x29, 0x02, 0x28, 0x01, 0x00, 0x00, 0x00, 0x2A, 0x00, 0x00})
	resp, err := o.ProcessRequest(req)
	if err != nil {
		t.Fatalf("DirectOperate G41V2 failed: %v", err)
	}
	if resp == nil {
		t.Fatal("expected a response")
	}
	if h.calls != 1 {
		t.Fatalf("handler calls = %d, want 1", h.calls)
	}
	if h.gotValue != int16(42) {
		t.Fatalf("decoded value = %v (%T), want int16(42)", h.gotValue, h.gotValue)
	}
	if h.gotIndex != 0 {
		t.Fatalf("index = %d, want 0", h.gotIndex)
	}
	if h.gotVar != 2 {
		t.Fatalf("variation = %d, want 2", h.gotVar)
	}
	// IIN(2) + G41V2 echo object with SUCCESS status octet.
	want := []byte{0x00, 0x00, 0x29, 0x02, 0x28, 0x01, 0x00, 0x00, 0x00, 0x2A, 0x00, 0x00}
	if string(resp.Data) != string(want) {
		t.Fatalf("G41V2 echo = %X, want %X", resp.Data, want)
	}
}

// TestOutstationG41V2DecodeBoundaries confirms the int16 minimum and maximum
// round-trip little-endian with a byte-exact SUCCESS echo (independent goldens
// for -32768 = 00 80 and 32767 = FF 7F).
func TestOutstationG41V2DecodeBoundaries(t *testing.T) {
	cases := []struct {
		name      string
		record    []byte
		wantValue int16
		wantEchoV [2]byte
	}{
		{"min int16", []byte{0x29, 0x02, 0x28, 0x01, 0x00, 0x00, 0x00, 0x00, 0x80, 0x00}, -32768, [2]byte{0x00, 0x80}},
		{"max int16", []byte{0x29, 0x02, 0x28, 0x01, 0x00, 0x00, 0x00, 0xFF, 0x7F, 0x00}, 32767, [2]byte{0xFF, 0x7F}},
		{"minus1", []byte{0x29, 0x02, 0x28, 0x01, 0x00, 0x00, 0x00, 0xFF, 0xFF, 0x00}, -1, [2]byte{0xFF, 0xFF}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := &g41v2RecordingHandler{}
			o := newG41V2Outstation(h)
			resp, err := o.ProcessRequest(g41v2DirectOperate(tc.record))
			if err != nil {
				t.Fatalf("DirectOperate G41V2 (%s) failed: %v", tc.name, err)
			}
			if h.gotValue != tc.wantValue {
				t.Fatalf("decoded value = %v, want int16(%d)", h.gotValue, tc.wantValue)
			}
			want := []byte{0x00, 0x00, 0x29, 0x02, 0x28, 0x01, 0x00, 0x00, 0x00, tc.wantEchoV[0], tc.wantEchoV[1], 0x00}
			if string(resp.Data) != string(want) {
				t.Fatalf("echo = %X, want %X", resp.Data, want)
			}
		})
	}
}

// TestOutstationG41V2RejectionsNoHandlerCall verifies unsupported G41
// variations and malformed lengths are rejected with no host-handler call, no
// CROB side effect and no SUCCESS echo (fail closed).
func TestOutstationG41V2RejectionsNoHandlerCall(t *testing.T) {
	cases := []struct {
		name string
		data []byte
	}{
		{"variation 3 wrong length (3 of 5 value+status bytes)", []byte{0x29, 0x03, 0x28, 0x01, 0x00, 0x00, 0x00, 0x2A, 0x00, 0x00}},
		{"variation 4 (unsupported)", []byte{0x29, 0x04, 0x28, 0x01, 0x00, 0x00, 0x00, 0x2A, 0x00, 0x00}},
		{"count 0", []byte{0x29, 0x02, 0x28, 0x00, 0x00, 0x00, 0x00, 0x2A, 0x00, 0x00}},
		{"count 2", []byte{0x29, 0x02, 0x28, 0x02, 0x00, 0x00, 0x00, 0x2A, 0x00, 0x00}},
		{"truncated value (1 of 2)", []byte{0x29, 0x02, 0x28, 0x01, 0x00, 0x00, 0x00, 0x2A}},
		{"truncated status (2 of 3)", []byte{0x29, 0x02, 0x28, 0x01, 0x00, 0x00, 0x00, 0x2A, 0x00}},
		{"surplus byte", []byte{0x29, 0x02, 0x28, 0x01, 0x00, 0x00, 0x00, 0x2A, 0x00, 0x00, 0xEE}},
		{"missing index", []byte{0x29, 0x02, 0x28, 0x01, 0x00}},
		{"bad qualifier", []byte{0x29, 0x02, 0x06, 0x01, 0x00, 0x00, 0x00, 0x2A, 0x00, 0x00}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := &g41v2RecordingHandler{}
			o := newG41V2Outstation(h)
			resp, err := o.ProcessRequest(g41v2DirectOperate(tc.data))
			if err == nil {
				t.Fatalf("expected rejection for %s, got resp=%v", tc.name, resp)
			}
			if resp != nil {
				t.Fatalf("expected nil response (no SUCCESS echo), got %v", resp)
			}
			if h.calls != 0 || h.seenCROB != 0 {
				t.Fatalf("handler invoked for a rejected G41V2 record: analogCalls=%d crobCalls=%d", h.calls, h.seenCROB)
			}
			if h.lastAnalog != nil {
				t.Fatalf("state mutated: lastAnalog=%v", h.lastAnalog)
			}
		})
	}
}

// TestOutstationG41V2HandlerRefusalNotSuccess verifies a handler error yields an
// error response (never a SUCCESS echo), so the master cannot misread a refusal
// as success.
func TestOutstationG41V2HandlerRefusalNotSuccess(t *testing.T) {
	h := &g41v2RecordingHandler{writeErr: ErrInvalidRequest}
	o := newG41V2Outstation(h)
	req := g41v2DirectOperate([]byte{0x29, 0x02, 0x28, 0x01, 0x00, 0x00, 0x00, 0x2A, 0x00, 0x00})
	resp, err := o.ProcessRequest(req)
	if err == nil {
		t.Fatalf("expected handler refusal to surface as an error, got resp=%v", resp)
	}
	if resp != nil {
		t.Fatalf("handler refusal must not produce a SUCCESS echo, got %v", resp)
	}
}

// TestOutstationG41V2NoDataWriterFailsClosed verifies a G41V2 DirectOperate to
// an outstation whose data handler is not a DataWriter fails closed: an error is
// returned, no handler runs and no SUCCESS echo is built.
func TestOutstationG41V2NoDataWriterFailsClosed(t *testing.T) {
	o := newReadOnlyOutstation()
	req := g41v2DirectOperate([]byte{0x29, 0x02, 0x28, 0x01, 0x00, 0x00, 0x00, 0x2A, 0x00, 0x00})
	resp, err := o.ProcessRequest(req)
	if err == nil {
		t.Fatalf("expected fail-closed error for a missing DataWriter, got resp=%v", resp)
	}
	if resp != nil {
		t.Fatalf("expected nil response (no SUCCESS echo), got %v", resp)
	}
}

// TestOutstationG41V2DoesNotTouchCROB confirms a G41V2 command never reaches the
// binary-output path (no cross-group leakage).
func TestOutstationG41V2DoesNotTouchCROB(t *testing.T) {
	h := &g41v2RecordingHandler{}
	o := newG41V2Outstation(h)
	req := g41v2DirectOperate([]byte{0x29, 0x02, 0x28, 0x01, 0x00, 0x00, 0x00, 0x2A, 0x00, 0x00})
	if _, err := o.ProcessRequest(req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if h.seenCROB != 0 {
		t.Fatalf("G41V2 command reached WriteBinaryOutput %d times", h.seenCROB)
	}
}

// TestOutstationG41V2CountOneExecutesOnce is the positive control: a valid
// single-point G41V2 request executes exactly once and returns the byte-exact
// SUCCESS echo.
func TestOutstationG41V2CountOneExecutesOnce(t *testing.T) {
	h := &g41v2RecordingHandler{}
	o := newG41V2Outstation(h)
	resp, err := o.ProcessRequest(g41v2DirectOperate([]byte{0x29, 0x02, 0x28, 0x01, 0x00, 0x00, 0x00, 0x2A, 0x00, 0x00}))
	if err != nil {
		t.Fatalf("valid count=1 G41V2 request rejected: %v", err)
	}
	if h.calls != 1 || h.seenCROB != 0 {
		t.Fatalf("analogCalls=%d crobCalls=%d, want 1/0", h.calls, h.seenCROB)
	}
	if h.gotValue != int16(42) {
		t.Fatalf("executed value = %v (%T), want int16(42)", h.gotValue, h.gotValue)
	}
	want := []byte{0x00, 0x00, 0x29, 0x02, 0x28, 0x01, 0x00, 0x00, 0x00, 0x2A, 0x00, 0x00}
	if string(resp.Data) != string(want) {
		t.Fatalf("echo = %X, want %X", resp.Data, want)
	}
}
