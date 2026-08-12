package outstation

import (
	"strings"
	"testing"

	"dnp3/internal/al"
	"dnp3/internal/dll/frame"
	"dnp3/internal/tl"
)

// recTransport captures bytes passed to Send for later decoding.
type recTransport struct {
	sent [][]byte
}

func (t *recTransport) Send(data []byte) error {
	t.sent = append(t.sent, data)
	return nil
}
func (t *recTransport) Receive() ([]byte, error) { return nil, nil }
func (t *recTransport) SetTimeout(ms int)        {}

// TestProcessRequestUnsupportedFunctionCode verifies DNP3-080: an unsupported
// application function code is rejected by ProcessRequest with a descriptive
// error naming the code, and produces no response object (the run loop maps
// the error to an IIN.FuncUnknown response).
func TestProcessRequestUnsupportedFunctionCode(t *testing.T) {
	ost := NewOutstation(nil)
	ost.Initialize()

	// 0x2B (43) is not in the v0 outstation dispatch table.
	const badFC uint8 = 0x2B
	req := &al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: 7},
		FuncCode: badFC,
		Data:     nil,
	}
	resp, err := ost.ProcessRequest(req)
	if err == nil {
		t.Fatal("expected error for unsupported function code, got nil")
	}
	if resp != nil {
		t.Fatalf("expected nil response for unsupported FC, got %+v", resp)
	}
	if !strings.Contains(err.Error(), "unsupported function code") {
		t.Fatalf("error %q must mention 'unsupported function code'", err.Error())
	}
	if !strings.Contains(err.Error(), "43") {
		t.Fatalf("error %q must name the function code (43)", err.Error())
	}
}

// TestSendErrorResponseUnsupportedFCSetsFuncUnknown verifies DNP3-080: the wire
// response the master receives for an unsupported FC is a FuncResponse carrying
// IIN.FuncUnknown (function code cannot be processed), so the master sees a
// clear failure rather than a timeout.
func TestSendErrorResponseUnsupportedFCSetsFuncUnknown(t *testing.T) {
	ost := NewOutstation(nil)
	ost.Initialize()
	tr := &recTransport{}
	ost.SetTransport(tr)

	ost.sendErrorResponse(9, errUnsupportedFCForTest(0x2B))
	if len(tr.sent) == 0 {
		t.Fatal("expected outstation to send an error response on the wire")
	}

	// Decode the single DLL frame → TL fragment → APDU.
	f, err := frame.Decode(tr.sent[0])
	if err != nil {
		t.Fatalf("frame.Decode: %v", err)
	}
	frag, err := tl.DecodeFragment(f.Data)
	if err != nil {
		t.Fatalf("tl.DecodeFragment: %v", err)
	}
	apdu, err := al.Decode(frag.Data)
	if err != nil {
		t.Fatalf("al.Decode: %v", err)
	}
	if apdu.FuncCode != al.FuncResponse {
		t.Fatalf("APDU FuncCode = %d, want FuncResponse (%d)", apdu.FuncCode, al.FuncResponse)
	}
	if apdu.Control.Seq != 9 {
		t.Fatalf("APDU Seq = %d, want 9 (echo request seq)", apdu.Control.Seq)
	}
	iin, err := al.DecodeIIN(apdu.Data)
	if err != nil {
		t.Fatalf("DecodeIIN: %v", err)
	}
	if !iin.FuncUnknown {
		t.Fatalf("expected IIN.FuncUnknown set for unsupported FC, got %+v", iin)
	}
}

// TestProcessRequestUnsupportedFCDoesNotPolluteStationIIN verifies DNP3-080: a
// rejected function code does not leave FuncUnknown/ParameterError set on the
// persistent station IIN, so subsequent valid responses are not polluted.
func TestProcessRequestUnsupportedFCDoesNotPolluteStationIIN(t *testing.T) {
	ost := NewOutstation(nil)
	ost.Initialize()

	_, _ = ost.ProcessRequest(&al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: 1},
		FuncCode: 0x2B,
	})

	// A subsequent valid READ must NOT carry FuncUnknown/ParameterError.
	resp, err := ost.ProcessRequest(&al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: 2},
		FuncCode: al.FuncRead,
		Data:     []byte{1, 1, 0x06, 0x00}, // G1 V1 all-objects
	})
	if err != nil {
		t.Fatalf("subsequent Read: %v", err)
	}
	if resp == nil {
		t.Fatal("expected a Read response")
	}
	iin, err := al.DecodeIIN(resp.Data)
	if err != nil {
		t.Fatalf("DecodeIIN: %v", err)
	}
	if iin.FuncUnknown {
		t.Fatalf("FuncUnknown leaked into subsequent response (persistent IIN pollution)")
	}
	if iin.ParameterError {
		t.Fatalf("ParameterError leaked into subsequent response (persistent IIN pollution)")
	}
}

// errUnsupportedFCForTest mirrors the run-loop's unsupported-FC error shape.
func errUnsupportedFCForTest(fc uint8) error {
	return errUnsupportedFC(fc)
}

type fcErr uint8

func (e fcErr) Error() string { return "unsupported function code" }

func errUnsupportedFC(fc uint8) error { return fcErr(fc) }
