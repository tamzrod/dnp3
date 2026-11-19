package outstation

import (
	"errors"
	"testing"

	"dnp3/internal/al"
)

// G12V1 command responses must carry the per-point command-status octet in an
// echoed G12V1 object. Independent masters (including the Windows DNP3 Tester
// and the pinned dscsystems parser) reject an IIN-only response with "the
// outstation returned no command statuses". These tests pin the captured wire
// layout and the fail-closed rejection path.

// crobEchoHandler records the executed CROB and can return a definite IEEE 1815
// command-status refusal through ControlStatusError.
type crobEchoHandler struct {
	DefaultDataHandler
	calls     int
	lastIndex uint16
	lastCROB  CROB
	err       error
}

func (h *crobEchoHandler) WriteBinaryOutput(index uint16, crob *CROB) error {
	h.calls++
	h.lastIndex = index
	h.lastCROB = *crob
	return h.err
}

func newG12EchoOutstation(h *crobEchoHandler) *Outstation {
	o := NewOutstation(nil)
	o.Initialize()
	o.Start()
	o.SetDataHandler(h)
	return o
}

func g12ControlAPDU(fc uint8, data []byte) *al.APDU {
	return &al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: 1},
		FuncCode: fc,
		Data:     append([]byte(nil), data...),
	}
}

// Exact application-layer object data captured from the Windows DNP3 Tester:
// G12V1, qualifier 0x17 (8-bit count + 8-bit index), count 1, index 2,
// control code 0x04 (Latch Off), count 1, onTime=1000, offTime=0, status=0.
// The header CRC and the two data-block CRCs are not part of the APDU.
var windowsTesterCROBIndex2LatchOff = []byte{
	0x0C, 0x01, 0x17, 0x01, 0x02,
	0x04, 0x01, 0xE8, 0x03, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00,
	0x00,
}

// Independent-master request from active_work/testdata/direct-control-crob-vector.hex:
// qualifier 0x00, count 1, index 0x1234, Latch On, onTime=1000, offTime=2000.
var independentCROBVector = []byte{
	0x0C, 0x01, 0x00, 0x01, 0x34, 0x12,
	0x03, 0x01, 0xE8, 0x03, 0x00, 0x00, 0xD0, 0x07, 0x00, 0x00, 0x00,
}

// TestG12V1WindowsTesterQualifier17DecodesAndEchoes is the exact captured
// interoperability regression. Before the fix, qualifier 0x17 was rejected as
// a parameter error; after the fix the command executes and the response echoes
// the request with an explicit SUCCESS status octet.
func TestG12V1WindowsTesterQualifier17DecodesAndEchoes(t *testing.T) {
	h := &crobEchoHandler{}
	o := newG12EchoOutstation(h)

	resp, err := o.ProcessRequest(g12ControlAPDU(al.FuncDirectOperate, windowsTesterCROBIndex2LatchOff))
	if err != nil {
		t.Fatalf("Direct Operate failed: %v", err)
	}
	if resp == nil {
		t.Fatal("expected a response")
	}
	if h.calls != 1 || h.lastIndex != 2 {
		t.Fatalf("handler calls=%d index=%d, want 1 and 2", h.calls, h.lastIndex)
	}
	if h.lastCROB.Code != 0x04 || h.lastCROB.Count != 1 || h.lastCROB.OnTime != 1000 || h.lastCROB.OffTime != 0 {
		t.Fatalf("decoded CROB = %+v, want code=0x04 count=1 on=1000 off=0", h.lastCROB)
	}
	got := assertG12Echo(t, resp, windowsTesterCROBIndex2LatchOff)
	if got[len(got)-1] != 0x00 {
		t.Fatalf("command status = %02X, want SUCCESS 00", got[len(got)-1])
	}
}

// TestG12V1IndependentMasterVectorEchoesStatus locks the qualified response
// layout against the repository's independent OpenDNP3-derived request vector.
func TestG12V1IndependentMasterVectorEchoesStatus(t *testing.T) {
	h := &crobEchoHandler{}
	o := newG12EchoOutstation(h)

	resp, err := o.ProcessRequest(g12ControlAPDU(al.FuncDirectOperate, independentCROBVector))
	if err != nil {
		t.Fatalf("Direct Operate failed: %v", err)
	}
	if resp == nil {
		t.Fatal("expected a response")
	}
	if h.calls != 1 || h.lastIndex != 0x1234 {
		t.Fatalf("handler calls=%d index=%d, want 1 and 0x1234", h.calls, h.lastIndex)
	}
	assertG12Echo(t, resp, independentCROBVector)
}

// assertG12Echo checks the two-octet IIN prefix that ProcessRequest adds to
// every APDU response and then compares the object bytes after it.
func assertG12Echo(t *testing.T, resp *al.APDU, want []byte) []byte {
	t.Helper()
	if resp == nil {
		t.Fatal("expected a response")
	}
	if len(resp.Data) < 2 {
		t.Fatalf("response data = %X, want IIN + object", resp.Data)
	}
	if resp.Data[0] != 0x00 || resp.Data[1] != 0x00 {
		t.Fatalf("response IIN = %02X%02X, want 0000", resp.Data[0], resp.Data[1])
	}
	got := resp.Data[2:]
	if string(got) != string(want) {
		t.Fatalf("response echo = %X, want %X", got, want)
	}
	return got
}

// TestG12V1SBOEchoesStatusOnSelectAndOperate covers the Windows tester's second
// failing mode. Both SELECT and OPERATE must contain a parseable G12V1 status.
func TestG12V1SBOEchoesStatusOnSelectAndOperate(t *testing.T) {
	h := &crobEchoHandler{}
	o := newG12EchoOutstation(h)

	sel, err := o.ProcessRequest(g12ControlAPDU(al.FuncSelect, windowsTesterCROBIndex2LatchOff))
	if err != nil {
		t.Fatalf("Select failed: %v", err)
	}
	assertG12Echo(t, sel, windowsTesterCROBIndex2LatchOff)
	if h.calls != 0 {
		t.Fatalf("Select must not execute the output; handler calls=%d", h.calls)
	}

	op, err := o.ProcessRequest(g12ControlAPDU(al.FuncOperate, windowsTesterCROBIndex2LatchOff))
	if err != nil {
		t.Fatalf("Operate failed: %v", err)
	}
	assertG12Echo(t, op, windowsTesterCROBIndex2LatchOff)
	if h.calls != 1 {
		t.Fatalf("handler calls=%d, want 1", h.calls)
	}
}

// TestG12V1RejectionEchoesActualStatus proves a handler refusal is never
// rewritten as SUCCESS: the carried IEEE 1815 status is echoed in the G12V1
// object so an independent master sees a definite non-success.
func TestG12V1RejectionEchoesActualStatus(t *testing.T) {
	h := &crobEchoHandler{err: NewControlStatusError(0x06, errors.New("hardware error"))}
	o := newG12EchoOutstation(h)

	resp, err := o.ProcessRequest(g12ControlAPDU(al.FuncDirectOperate, independentCROBVector))
	if err != nil {
		t.Fatalf("expected a G12V1 rejection echo, got error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected a response")
	}
	want := append([]byte(nil), independentCROBVector...)
	want[len(want)-1] = 0x06
	assertG12Echo(t, resp, want)
}

// TestG12V1UnclassifiableRejectionFailsClosed keeps the fail-closed behaviour:
// a handler error with no wire command status must not fabricate a SUCCESS echo.
func TestG12V1UnclassifiableRejectionFailsClosed(t *testing.T) {
	h := &crobEchoHandler{err: errors.New("driver unavailable")}
	o := newG12EchoOutstation(h)

	resp, err := o.ProcessRequest(g12ControlAPDU(al.FuncDirectOperate, independentCROBVector))
	if err == nil {
		t.Fatalf("expected fail-closed error, got response %v", resp)
	}
	if resp != nil {
		t.Fatalf("expected no response object on unclassifiable failure, got %X", resp.Data)
	}
}
