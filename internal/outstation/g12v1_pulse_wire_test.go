package outstation

import (
	"errors"
	"testing"

	"dnp3/internal/al"
)

// The captured EG628 / Windows DNP3 Tester Pulse Off CROB. This is the exact
// application-layer object data from the passive capture at 2026-11-19
// 09:41:11Z (link header and CRCs are not part of the APDU):
//
// G12V1, qualifier 0x17 (8-bit count + 8-bit index), count 1, index 1,
// code 0x02 (Pulse Off), count 1, OnTime 1000 (E8 03 00 00, LSB first),
// OffTime 0, request status 0x00.
//
// The tester's BAD_FORMAT (0x03) on this request was an EG628 adapter
// pulse-timing decision, not a DNP3 framing defect. The library's contract is
// narrower: echo the caller's full CROB unchanged and report the handler's real
// status. These tests pin that contract so a future change cannot start
// reinterpreting pulse durations inside the generic library.
var capturedPulseOffCROB = []byte{
	0x0C, 0x01, 0x17, 0x01, 0x01,
	0x02, 0x01, 0xE8, 0x03, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00,
	0x00,
}

// TestG12V1CapturedPulseOffDirectOperateEchoesUnchanged verifies a Direct
// Operate of the captured Pulse Off CROB is delivered to the handler with its
// exact fields and echoed byte-for-byte, with SUCCESS for an accepting handler.
func TestG12V1CapturedPulseOffDirectOperateEchoesUnchanged(t *testing.T) {
	h := &crobEchoHandler{}
	o := newG12EchoOutstation(h)

	resp, err := o.ProcessRequest(g12ControlAPDU(al.FuncDirectOperate, capturedPulseOffCROB))
	if err != nil {
		t.Fatalf("Direct Operate failed: %v", err)
	}
	if h.calls != 1 {
		t.Fatalf("handler calls = %d, want 1", h.calls)
	}
	if h.lastIndex != 1 {
		t.Fatalf("handler index = %d, want 1", h.lastIndex)
	}
	want := CROB{Code: 0x02, Count: 1, OnTime: 1000, OffTime: 0, Status: 0x00, Qualifier: 0x17}
	if h.lastCROB != want {
		t.Fatalf("decoded CROB = %+v, want %+v (OnTime/OffTime must not be swapped or rescheduled)", h.lastCROB, want)
	}

	got := assertG12Echo(t, resp, capturedPulseOffCROB)
	if got[len(got)-1] != 0x00 {
		t.Fatalf("command status = %02X, want SUCCESS 00", got[len(got)-1])
	}
}

// TestG12V1CapturedPulseOffSBOEchoesUnchanged covers the SELECT->OPERATE path
// the tester actually used. SELECT must echo status without executing or
// mutating output state; OPERATE must echo the same bytes and execute once.
func TestG12V1CapturedPulseOffSBOEchoesUnchanged(t *testing.T) {
	h := &crobEchoHandler{}
	o := newG12EchoOutstation(h)

	sel, err := o.ProcessRequest(g12ControlAPDU(al.FuncSelect, capturedPulseOffCROB))
	if err != nil {
		t.Fatalf("Select failed: %v", err)
	}
	assertG12Echo(t, sel, capturedPulseOffCROB)
	if h.calls != 0 {
		t.Fatalf("SELECT must not execute the output; handler calls = %d", h.calls)
	}

	op, err := o.ProcessRequest(g12ControlAPDU(al.FuncOperate, capturedPulseOffCROB))
	if err != nil {
		t.Fatalf("Operate failed: %v", err)
	}
	assertG12Echo(t, op, capturedPulseOffCROB)
	if h.calls != 1 {
		t.Fatalf("handler calls = %d, want 1", h.calls)
	}
	if h.lastCROB.OnTime != 1000 || h.lastCROB.OffTime != 0 || h.lastCROB.Code != 0x02 {
		t.Fatalf("OPERATE delivered %+v, want code=0x02 on=1000 off=0", h.lastCROB)
	}
}

// TestG12V1CapturedPulseOffRefusalEchoesBadFormat proves a handler that refuses
// the pulse with BAD_FORMAT (the EG628 adapter's pre-fix behaviour) reaches the
// master as a definite 0x03 status, not as a manufactured SUCCESS.
func TestG12V1CapturedPulseOffRefusalEchoesBadFormat(t *testing.T) {
	h := &crobEchoHandler{err: NewControlStatusError(0x03, errors.New("pulse off without off-time"))}
	o := newG12EchoOutstation(h)

	resp, err := o.ProcessRequest(g12ControlAPDU(al.FuncDirectOperate, capturedPulseOffCROB))
	if err != nil {
		t.Fatalf("expected a G12V1 rejection echo, got error: %v", err)
	}
	want := append([]byte(nil), capturedPulseOffCROB...)
	want[len(want)-1] = 0x03
	got := assertG12Echo(t, resp, want)
	if got[len(got)-1] != 0x03 {
		t.Fatalf("command status = %02X, want BAD_FORMAT 03", got[len(got)-1])
	}
}

// TestG12V1CapturedPulseOffSBOSelectThenRefusedOperate covers the full tester
// sequence with a refusing adapter: SELECT succeeds, OPERATE reports 0x03, and
// the refusal does not fabricate a success echo.
func TestG12V1CapturedPulseOffSBOSelectThenRefusedOperate(t *testing.T) {
	h := &crobEchoHandler{err: NewControlStatusError(0x03, errors.New("pulse off without off-time"))}
	o := newG12EchoOutstation(h)

	sel, err := o.ProcessRequest(g12ControlAPDU(al.FuncSelect, capturedPulseOffCROB))
	if err != nil {
		t.Fatalf("Select failed: %v", err)
	}
	assertG12Echo(t, sel, capturedPulseOffCROB)

	op, err := o.ProcessRequest(g12ControlAPDU(al.FuncOperate, capturedPulseOffCROB))
	if err != nil {
		t.Fatalf("expected a G12V1 rejection echo, got error: %v", err)
	}
	want := append([]byte(nil), capturedPulseOffCROB...)
	want[len(want)-1] = 0x03
	assertG12Echo(t, op, want)
}

// TestG12V1CapturedPulseOffUnclassifiableFailsClosed keeps the fail-closed rule:
// a refusal with no wire-encodable status must produce no object at all rather
// than a fabricated status octet.
func TestG12V1CapturedPulseOffUnclassifiableFailsClosed(t *testing.T) {
	h := &crobEchoHandler{err: errors.New("driver unavailable")}
	o := newG12EchoOutstation(h)

	resp, err := o.ProcessRequest(g12ControlAPDU(al.FuncDirectOperate, capturedPulseOffCROB))
	if err == nil {
		t.Fatalf("expected fail-closed error, got response %X", resp.Data)
	}
	if resp != nil {
		t.Fatalf("expected no response object, got %X", resp.Data)
	}
}
