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
		{"variation 2", []byte{0x29, 0x02, 0x28, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00}},
		{"variation 3", []byte{0x29, 0x03, 0x28, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
		{"variation 4", []byte{0x29, 0x04, 0x28, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
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
