package outstation

import (
	"testing"

	"dnp3/internal/al"
)

// AO-002 — Group 41 Variation 1 (signed int32) outstation side.
//
// Verifies the outstation decodes the G41V1 record little-endian against
// independent literals (docs/verification/AO-001.md §3/§6), rejects unsupported
// variations and malformed lengths without invoking the host handler, and echoes
// the object with a command-status octet.

// g41RecordingHandler records analog outputs the outstation writes.
type g41RecordingHandler struct {
	DefaultDataHandler
	calls     int
	gotIndex  uint16
	gotValue  interface{}
	gotVar    uint8
	writeErr  error
	seenCROB  int
	lastCROB  *CROB
	lastIndex uint16
}

func (h *g41RecordingHandler) WriteAnalogOutput(index uint16, value interface{}, variation uint8) error {
	h.calls++
	h.gotIndex = index
	h.gotValue = value
	h.gotVar = variation
	return h.writeErr
}

func (h *g41RecordingHandler) WriteBinaryOutput(index uint16, crob *CROB) error {
	h.seenCROB++
	h.lastCROB = crob
	h.lastIndex = index
	return nil
}

func newG41Outstation(h *g41RecordingHandler) *Outstation {
	o := NewOutstation(nil)
	o.Initialize()
	o.Start()
	o.SetDataHandler(h)
	return o
}

// TestOutstationG41V1DecodeGolden confirms the outstation decodes the G41V1
// int32 value little-endian and passes a typed int32 to the handler, against the
// independent golden for 2300 (value bytes FC 08 00 00).
func TestOutstationG41V1DecodeGolden(t *testing.T) {
	h := &g41RecordingHandler{}
	o := newG41Outstation(h)

	// header(5) + index(2) + int32 LE(4) + status(1)
	req := &al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: 1},
		FuncCode: al.FuncDirectOperate,
		Data:     []byte{0x29, 0x01, 0x28, 0x01, 0x00, 0x00, 0x00, 0xFC, 0x08, 0x00, 0x00, 0x00},
	}
	resp, err := o.ProcessRequest(req)
	if err != nil {
		t.Fatalf("DirectOperate G41V1 failed: %v", err)
	}
	if resp == nil {
		t.Fatal("expected a response")
	}
	if h.calls != 1 {
		t.Fatalf("handler calls = %d, want 1", h.calls)
	}
	if h.gotValue != int32(2300) {
		t.Fatalf("decoded value = %v (%T), want int32(2300)", h.gotValue, h.gotValue)
	}
	if h.gotIndex != 0 {
		t.Fatalf("index = %d, want 0", h.gotIndex)
	}
	if h.gotVar != 1 {
		t.Fatalf("variation = %d, want 1", h.gotVar)
	}
	// Response must echo the object with a SUCCESS status octet.
	want := []byte{0x00, 0x00, 0x29, 0x01, 0x28, 0x01, 0x00, 0x00, 0x00, 0xFC, 0x08, 0x00, 0x00, 0x00}
	if string(resp.Data) != string(want) {
		t.Fatalf("G41V1 echo = %X, want %X", resp.Data, want)
	}
}

// TestOutstationG41V1DecodeNegative confirms a negative int32 round-trips
// (little-endian, two's complement): -1 -> FF FF FF FF.
func TestOutstationG41V1DecodeNegative(t *testing.T) {
	h := &g41RecordingHandler{}
	o := newG41Outstation(h)
	req := &al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: 1},
		FuncCode: al.FuncDirectOperate,
		Data:     []byte{0x29, 0x01, 0x28, 0x01, 0x00, 0x00, 0x00, 0xFF, 0xFF, 0xFF, 0xFF, 0x00},
	}
	if _, err := o.ProcessRequest(req); err != nil {
		t.Fatalf("DirectOperate G41V1 (-1) failed: %v", err)
	}
	if h.gotValue != int32(-1) {
		t.Fatalf("decoded value = %v, want int32(-1)", h.gotValue)
	}
}

// TestOutstationG41RejectionsNoHandlerCall verifies unsupported G41 variations
// and malformed lengths are rejected with no host-handler call and no side
// effect (fail closed).
func TestOutstationG41RejectionsNoHandlerCall(t *testing.T) {
	cases := []struct {
		name string
		data []byte
	}{
		{"variation 3 wrong length (3 of 5 value+status bytes)", []byte{0x29, 0x03, 0x28, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
		{"variation 4 (unsupported)", []byte{0x29, 0x04, 0x28, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
		{"variation 2 truncated value (2 of 3)", []byte{0x29, 0x02, 0x28, 0x01, 0x00, 0x00, 0x00, 0x2A, 0x00}},
		{"truncated value (4 of 5)", []byte{0x29, 0x01, 0x28, 0x01, 0x00, 0x00, 0x00, 0xFC, 0x08, 0x00, 0x00}},
		{"bad qualifier", []byte{0x29, 0x01, 0x06, 0x01, 0x00, 0x00, 0x00, 0xFC, 0x08, 0x00, 0x00, 0x00}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := &g41RecordingHandler{}
			o := newG41Outstation(h)
			req := &al.APDU{
				Control:  al.AppControl{FIR: true, FIN: true, Seq: 1},
				FuncCode: al.FuncDirectOperate,
				Data:     tc.data,
			}
			_, err := o.ProcessRequest(req)
			if err == nil {
				t.Fatalf("expected rejection for %s, got nil", tc.name)
			}
			if h.calls != 0 {
				t.Fatalf("handler invoked %d times for a rejected G41 record", h.calls)
			}
		})
	}
}

// TestOutstationG41V1HandlerRefusalNotSuccess verifies a handler error yields an
// error response (not a SUCCESS echo), so the master cannot misread a refusal as
// success.
func TestOutstationG41V1HandlerRefusalNotSuccess(t *testing.T) {
	h := &g41RecordingHandler{writeErr: ErrInvalidRequest}
	o := newG41Outstation(h)
	req := &al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: 1},
		FuncCode: al.FuncDirectOperate,
		Data:     []byte{0x29, 0x01, 0x28, 0x01, 0x00, 0x00, 0x00, 0xFC, 0x08, 0x00, 0x00, 0x00},
	}
	resp, err := o.ProcessRequest(req)
	if err == nil {
		t.Fatalf("expected handler refusal to surface as an error, got resp=%v", resp)
	}
}

// TestOutstationG41V1DoesNotTouchCROB confirms a G41V1 command never reaches the
// binary-output path (no cross-group leakage).
func TestOutstationG41V1DoesNotTouchCROB(t *testing.T) {
	h := &g41RecordingHandler{}
	o := newG41Outstation(h)
	req := &al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: 1},
		FuncCode: al.FuncDirectOperate,
		Data:     []byte{0x29, 0x01, 0x28, 0x01, 0x00, 0x00, 0x00, 0xFC, 0x08, 0x00, 0x00, 0x00},
	}
	if _, err := o.ProcessRequest(req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if h.seenCROB != 0 {
		t.Fatalf("G41V1 command reached WriteBinaryOutput %d times", h.seenCROB)
	}
}

// g41v1MutableHandler records handler calls and simulates output state so F3 can
// prove a rejected request neither calls the handler nor mutates state.
type g41v1MutableHandler struct {
	DefaultDataHandler
	analogCalls   int
	binaryCalls   int
	lastAnalog    interface{}
	lastVariation uint8
	lastIndex     uint16
}

func (h *g41v1MutableHandler) WriteAnalogOutput(index uint16, value interface{}, variation uint8) error {
	h.analogCalls++
	h.lastAnalog = value
	h.lastVariation = variation
	h.lastIndex = index
	return nil
}

func (h *g41v1MutableHandler) WriteBinaryOutput(index uint16, crob *CROB) error {
	h.binaryCalls++
	return nil
}

// TestOutstationG41V1F3CountValidation verifies F3: this slice supports exactly
// one G41V1 point per Direct Operate. count=0, count>1 (with one or two records),
// truncated and surplus records are rejected before any side effect: no
// WriteAnalogOutput / WriteBinaryOutput call, no state mutation, and never a
// SUCCESS echo.
func TestOutstationG41V1F3CountValidation(t *testing.T) {
	// Header: group 41, var 1, qualifier 0x28, 16-bit count, 16-bit index.
	// After the header the request carries only the 5-byte value+status (the
	// index is part of the header), so a count != 1 with a 5-byte value is what
	// isolates the count check from executeControl's length check.
	hdr := func(count uint16, index uint16) []byte {
		return []byte{0x29, 0x01, 0x28, byte(count), byte(count >> 8), byte(index), byte(index >> 8)}
	}
	value := []byte{0xFC, 0x08, 0x00, 0x00, 0x00} // int32 2300 LE + status
	cases := []struct {
		name string
		data []byte
	}{
		{"count 0 with a 5-byte value", append(hdr(0, 0), value...)},
		{"count 2 with a 5-byte value", append(hdr(2, 0), value...)},
		{"count 2 with two records", append(hdr(2, 0), append(append([]byte{}, value...), value...)...)},
		{"truncated value (count 1, 4 bytes)", append(hdr(1, 0), 0xFC, 0x08, 0x00, 0x00)},
		{"surplus byte (count 1, 6 bytes)", append(hdr(1, 0), 0xFC, 0x08, 0x00, 0x00, 0x00, 0xEE)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := &g41v1MutableHandler{}
			o := NewOutstation(nil)
			o.Initialize()
			o.Start()
			o.SetDataHandler(h)

			resp, err := o.ProcessRequest(&al.APDU{
				Control:  al.AppControl{FIR: true, FIN: true, Seq: 1},
				FuncCode: al.FuncDirectOperate,
				Data:     tc.data,
			})
			if err == nil {
				t.Fatalf("expected rejection for %s, got resp=%v", tc.name, resp)
			}
			if resp != nil {
				t.Fatalf("expected nil response (no SUCCESS echo), got %v", resp)
			}
			if h.analogCalls != 0 || h.binaryCalls != 0 {
				t.Fatalf("side effect: analogCalls=%d binaryCalls=%d, want 0/0", h.analogCalls, h.binaryCalls)
			}
			if h.lastAnalog != nil {
				t.Fatalf("state mutated: lastAnalog=%v", h.lastAnalog)
			}
		})
	}
}

// TestOutstationG41V1F3ValidCountOneStillExecutes is the positive control for F3:
// a valid single-point request (count 1, 5-byte value) still executes exactly
// once and returns the byte-exact SUCCESS echo.
func TestOutstationG41V1F3ValidCountOneStillExecutes(t *testing.T) {
	h := &g41v1MutableHandler{}
	o := NewOutstation(nil)
	o.Initialize()
	o.Start()
	o.SetDataHandler(h)

	req := &al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: 1},
		FuncCode: al.FuncDirectOperate,
		Data:     []byte{0x29, 0x01, 0x28, 0x01, 0x00, 0x00, 0x00, 0xFC, 0x08, 0x00, 0x00, 0x00},
	}
	resp, err := o.ProcessRequest(req)
	if err != nil {
		t.Fatalf("valid count=1 request rejected: %v", err)
	}
	if h.analogCalls != 1 || h.binaryCalls != 0 {
		t.Fatalf("analogCalls=%d binaryCalls=%d, want 1/0", h.analogCalls, h.binaryCalls)
	}
	if h.lastAnalog != int32(2300) {
		t.Fatalf("executed value = %v (%T), want int32(2300)", h.lastAnalog, h.lastAnalog)
	}
	want := []byte{0x00, 0x00, 0x29, 0x01, 0x28, 0x01, 0x00, 0x00, 0x00, 0xFC, 0x08, 0x00, 0x00, 0x00}
	if string(resp.Data) != string(want) {
		t.Fatalf("echo = %X, want %X", resp.Data, want)
	}
}
