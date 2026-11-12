package outstation

import (
	"errors"
	"testing"

	"dnp3/internal/al"
)

// AO-004 Appendix T — G41V1–V4 OUT_OF_RANGE command-status propagation.
//
// Before the corrective, a DataWriter refusal (fixture returns ControlOutOfRange
// for an invalid index) surfaced only as an error, so the outstation emitted a
// generic error response (IIN FuncUnknown, no command-status object) and the
// master reported point state=INIT status=UNDEFINED. After the corrective, a
// refusal that carries a wire-encodable status is echoed as a normal G41 object
// with the original index, variation, numeric value and the actual OUT_OF_RANGE
// status octet (0x0C). These tests fail on the pre-fix behavior and pass after.

// outOfRangeHandler refuses every analog write with the IEEE 1815 OUT_OF_RANGE
// command status (wire 0x0C), returning the status together with an error, as
// the lab fixture does.
type outOfRangeHandler struct {
	DefaultDataHandler
	analogCalls int
	lastIndex   uint16
	lastValue   interface{}
	lastVar     uint8
	// status is the wire status octet the handler carries.
	status uint8
}

func (h *outOfRangeHandler) WriteAnalogOutput(index uint16, value interface{}, variation uint8) error {
	h.analogCalls++
	h.lastIndex = index
	h.lastValue = value
	h.lastVar = variation
	return NewControlStatusError(h.status, errors.New("index out of range"))
}

// g41Request builds a one-point 0x28-qualified G41 Direct Operate record:
// header(5) + index(2) + value bytes + trailing status octet.
func g41Request(variation uint8, index uint16, valueBytes []byte) *al.APDU {
	data := []byte{0x29, variation, 0x28, 0x01, 0x00, byte(index), byte(index >> 8)}
	data = append(data, valueBytes...)
	data = append(data, 0x00)
	return &al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: 1},
		FuncCode: al.FuncDirectOperate,
		Data:     data,
	}
}

// TestOutstationG41InvalidIndexEchoesOutOfRange is the core regression: for each
// variation at invalid index 1 the response must be a normal G41 echo carrying
// the original index, variation, numeric value and the OUT_OF_RANGE status octet
// (0x0C) — not a generic error response and never a SUCCESS echo.
func TestOutstationG41InvalidIndexEchoesOutOfRange(t *testing.T) {
	cases := []struct {
		name      string
		variation uint8
		value     []byte
	}{
		{"V1 int32", 1, []byte{0xFC, 0x08, 0x00, 0x00}},                           // 2300 LE
		{"V2 int16", 2, []byte{0x39, 0x30}},                                       // 12345 LE
		{"V3 float32", 3, []byte{0x00, 0x00, 0xC0, 0x3F}},                         // 1.5
		{"V4 float64", 4, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xF8, 0x3F}}, // 1.5
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := &outOfRangeHandler{status: 0x0C}
			o := NewOutstation(nil)
			o.Initialize()
			o.Start()
			o.SetDataHandler(h)

			resp, err := o.ProcessRequest(g41Request(tc.variation, 1, tc.value))
			if err != nil {
				t.Fatalf("expected a G41 echo response, got error: %v", err)
			}
			if resp == nil {
				t.Fatal("expected a non-nil response")
			}
			if h.analogCalls != 1 {
				t.Fatalf("handler calls = %d, want 1", h.analogCalls)
			}
			if h.lastIndex != 1 || h.lastVar != tc.variation {
				t.Fatalf("handler saw index=%d var=%d, want index=1 var=%d", h.lastIndex, h.lastVar, tc.variation)
			}

			// Response data is IIN(2) + G41 echo object.
			want := append([]byte{0x00, 0x00}, byte(0x29), tc.variation, 0x28, 0x01, 0x00, 0x01, 0x00)
			want = append(want, tc.value...)
			want = append(want, 0x0C) // OUT_OF_RANGE command status
			if string(resp.Data) != string(want) {
				t.Fatalf("echo = %X, want %X", resp.Data, want)
			}
		})
	}
}

// TestOutstationG41InvalidIndexNoIINError confirms the OUT_OF_RANGE echo is not
// accompanied by a ParameterError/FuncUnknown IIN — the response is a clean
// per-point rejection, not a generic protocol error.
func TestOutstationG41InvalidIndexNoIINError(t *testing.T) {
	h := &outOfRangeHandler{status: 0x0C}
	o := NewOutstation(nil)
	o.Initialize()
	o.Start()
	o.SetDataHandler(h)

	resp, err := o.ProcessRequest(g41Request(1, 1, []byte{0xFC, 0x08, 0x00, 0x00}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil || len(resp.Data) < 2 {
		t.Fatal("expected a response with IIN")
	}
	// IIN1 = 0x00; IIN2 must not set FuncUnknown(0x80) or ParameterError(0x20).
	if resp.Data[0] != 0x00 {
		t.Fatalf("IIN1 = 0x%02X, want 0x00", resp.Data[0])
	}
	if resp.Data[1]&0x80 != 0 {
		t.Fatalf("IIN2 = 0x%02X has FuncUnknown set, want clean", resp.Data[1])
	}
	if resp.Data[1]&0x20 != 0 {
		t.Fatalf("IIN2 = 0x%02X has ParameterError set, want clean", resp.Data[1])
	}
}

// TestOutstationG41InvalidIndexNoMemoryMutation confirms the rejected index did
// not mutate fixture state: the handler is invoked once and reports no accepted
// value (the handler is the one refusing), and only one call occurs.
func TestOutstationG41InvalidIndexNoMemoryMutation(t *testing.T) {
	h := &outOfRangeHandler{status: 0x0C}
	o := NewOutstation(nil)
	o.Initialize()
	o.Start()
	o.SetDataHandler(h)

	if _, err := o.ProcessRequest(g41Request(1, 1, []byte{0xFC, 0x08, 0x00, 0x00})); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if h.analogCalls != 1 {
		t.Fatalf("handler calls = %d, want exactly 1", h.analogCalls)
	}
}

// TestOutstationG41ValidIndexStillSuccess is the positive control: at valid
// index 0 a handler that accepts returns a SUCCESS (0x00) echo, unchanged.
func TestOutstationG41ValidIndexStillSuccess(t *testing.T) {
	h := &g41v1MutableHandler{}
	o := NewOutstation(nil)
	o.Initialize()
	o.Start()
	o.SetDataHandler(h)

	resp, err := o.ProcessRequest(g41Request(1, 0, []byte{0xFC, 0x08, 0x00, 0x00}))
	if err != nil {
		t.Fatalf("valid index 0 rejected: %v", err)
	}
	want := []byte{0x00, 0x00, 0x29, 0x01, 0x28, 0x01, 0x00, 0x00, 0x00, 0xFC, 0x08, 0x00, 0x00, 0x00}
	if string(resp.Data) != string(want) {
		t.Fatalf("echo = %X, want %X", resp.Data, want)
	}
}

// TestOutstationG41GenericRefusalStillFailsClosed confirms a refusal WITHOUT a
// wire-encodable status (a plain error) still yields an error response and no
// echo — fail-closed is preserved and refusal is never turned into an echo.
func TestOutstationG41GenericRefusalStillFailsClosed(t *testing.T) {
	h := &g41RecordingHandler{writeErr: ErrInvalidRequest}
	o := newG41Outstation(h)

	resp, err := o.ProcessRequest(g41Request(1, 1, []byte{0xFC, 0x08, 0x00, 0x00}))
	if err == nil {
		t.Fatalf("expected an error response for a status-less refusal, got resp=%v", resp)
	}
	if resp != nil {
		t.Fatalf("expected nil response (no echo), got %v", resp)
	}
}

// TestControlStatusErrorUnwrap confirms the carrier preserves the wrapped error
// for existing fail-closed classification via errors.Is/As.
func TestControlStatusErrorUnwrap(t *testing.T) {
	base := errors.New("boom")
	cse := NewControlStatusError(12, base)
	if cse.Status != 12 {
		t.Fatalf("Status = %d, want 12", cse.Status)
	}
	if !errors.Is(cse, base) {
		t.Fatal("errors.Is did not match the wrapped error")
	}
	got, ok := AsControlStatusError(cse)
	if !ok || got.Status != 12 {
		t.Fatalf("AsControlStatusError = %+v, %v; want status 12, true", got, ok)
	}
}
