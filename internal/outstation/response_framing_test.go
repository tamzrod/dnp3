// AO-004 response-framing regression tests.
//
// These drive the outstation's real response path (RunWithContext over a
// recording transport) and decode the link frames it actually emits, to lock in
// the correction that outstation-originating *user data* responses are valid
// primary frames (DIR=false, PRM=true, function 4 UNCONFIRMED_USER_DATA) while
// *link-control* responses (Reset -> ACK, Link Status Request -> LINK_STATUS)
// remain proper secondary frames (PRM=false).
package outstation

import (
	"context"
	"sync"
	"testing"
	"time"

	"dnp3/internal/al"
	"dnp3/internal/dll/frame"
	"dnp3/internal/tl"
)

// scriptedTransport feeds one inbound frame then blocks until the context is
// cancelled, recording everything the outstation sends.
type scriptedTransport struct {
	in     [][]byte
	mu     sync.Mutex
	sent   [][]byte
	closed chan struct{}
}

func newScriptedTransport(in [][]byte) *scriptedTransport {
	return &scriptedTransport{in: in, closed: make(chan struct{})}
}

func (t *scriptedTransport) Send(data []byte) error {
	t.mu.Lock()
	cp := make([]byte, len(data))
	copy(cp, data)
	t.sent = append(t.sent, cp)
	t.mu.Unlock()
	return nil
}

func (t *scriptedTransport) Receive() ([]byte, error) {
	t.mu.Lock()
	if len(t.in) > 0 {
		f := t.in[0]
		t.in = t.in[1:]
		t.mu.Unlock()
		return f, nil
	}
	t.mu.Unlock()
	<-t.closed
	return nil, nil
}

func (t *scriptedTransport) SetTimeout(ms int) {}

func (t *scriptedTransport) sentFrames() [][]byte {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := make([][]byte, len(t.sent))
	copy(out, t.sent)
	return out
}

// runOutstationOnce feeds frames through one RunWithContext cycle, then cancels
// and returns the frames the outstation emitted.
func runOutstationOnce(t *testing.T, ost *Outstation, tr *scriptedTransport, in ...[]byte) [][]byte {
	t.Helper()
	ost.SetTransport(tr)
	for _, f := range in {
		tr.in = append(tr.in, f)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		_ = ost.RunWithContext(ctx)
		close(done)
	}()
	// Let the loop process the scripted input, then stop it.
	time.Sleep(150 * time.Millisecond)
	close(tr.closed)
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("outstation RunWithContext did not stop")
	}
	return tr.sentFrames()
}

// buildAppFrame builds a real encoded DLL frame carrying one transport fragment
// of apdu, with the given link control, addressed master->outstation.
func buildAppFrame(t *testing.T, ctrl frame.Control, apdu *al.APDU) []byte {
	t.Helper()
	enc := apdu.Encode()
	frag := tl.EncodeFragment(tl.Fragment{FIR: true, FIN: true, Seq: 0, Data: enc})
	b, err := frame.Encode(&frame.Frame{Control: ctrl, DestAddr: 4, SrcAddr: 3, Data: frag})
	if err != nil {
		t.Fatalf("frame.Encode: %v", err)
	}
	return b
}

func newResponseTestOutstation() *Outstation {
	cfg := DefaultConfig()
	cfg.OutstationAddress = 4
	cfg.MasterAddress = 3
	ost := NewOutstation(cfg)
	ost.SetDataHandler(NewDefaultDataHandler())
	_ = ost.Initialize()
	_ = ost.Start()
	return ost
}

// reassembleFramePayload extracts the transport fragment carried by an emitted
// DLL frame and reassembles it into the application message.
func reassembleFramePayload(t *testing.T, raw []byte) []byte {
	t.Helper()
	f, err := frame.Decode(raw)
	if err != nil {
		t.Fatalf("decode frame: %v", err)
	}
	if len(f.Data) == 0 {
		t.Fatal("frame carried no transport fragment")
	}
	frag, err := tl.DecodeFragment(f.Data)
	if err != nil {
		t.Fatalf("decode fragment: %v", err)
	}
	msg, err := tl.NewReassembler().Push(frag)
	if err != nil {
		t.Fatalf("reassemble: %v", err)
	}
	if msg == nil {
		t.Fatal("fragment did not yield a complete message")
	}
	return msg
}

// TestSolicitedResponseIsPrimaryUserData is the core regression: a solicited
// application response (here to a Class-0 Read) must be emitted as a primary
// user-data frame (DIR=false, PRM=true, function 4) addressed outstation->master
// and carrying the response APDU; it must NOT be a secondary frame.
func TestSolicitedResponseIsPrimaryUserData(t *testing.T) {
	ost := newResponseTestOutstation()
	tr := newScriptedTransport(nil)

	// Master -> outstation: primary confirmed user data (wire 3) carrying a Read.
	read := &al.APDU{Control: al.AppControl{FIR: true, FIN: true, Seq: 1}, FuncCode: al.FuncRead,
		Data: []byte{0x3C, 0x01, 0x06}} // G60V1 all-objects
	in := buildAppFrame(t, frame.Control{DIR: true, PRM: true, FCV: true, FuncCode: frame.FuncConfirmedUserData}, read)

	sent := runOutstationOnce(t, ost, tr, in)
	if len(sent) != 1 {
		t.Fatalf("expected exactly 1 emitted frame, got %d", len(sent))
	}
	f, err := frame.Decode(sent[0])
	if err != nil {
		t.Fatalf("decode emitted frame: %v", err)
	}
	if f.Control.DIR {
		t.Errorf("response DIR=true, want false")
	}
	if !f.Control.PRM {
		t.Errorf("response PRM=false, want true (primary user data)")
	}
	if f.Control.FuncCode != frame.FuncUnconfirmedUserData {
		t.Errorf("response func=%d, want %d (UNCONFIRMED_USER_DATA)", f.Control.FuncCode, frame.FuncUnconfirmedUserData)
	}
	if f.Control.FCB || f.Control.FCV {
		t.Errorf("unconfirmed user data must not set FCB/FCV (FCB=%v FCV=%v)", f.Control.FCB, f.Control.FCV)
	}
	if f.SrcAddr != 4 || f.DestAddr != 3 {
		t.Errorf("addresses src=%d dst=%d, want src=4 dst=3", f.SrcAddr, f.DestAddr)
	}
	// The carried fragment must reassemble to the application response.
	msg := reassembleFramePayload(t, sent[0])
	resp, err := al.Decode(msg)
	if err != nil {
		t.Fatalf("decode response APDU: %v", err)
	}
	if resp.FuncCode != al.FuncResponse {
		t.Errorf("application func=0x%02X, want RESPONSE (0x%02X)", resp.FuncCode, al.FuncResponse)
	}
}

// TestApplicationConfirmationIsPrimaryUserData: when the request sets CON, the
// application confirmation must also be a primary user-data frame.
func TestApplicationConfirmationIsPrimaryUserData(t *testing.T) {
	ost := newResponseTestOutstation()
	tr := newScriptedTransport(nil)

	read := &al.APDU{Control: al.AppControl{FIR: true, FIN: true, CON: true, Seq: 2}, FuncCode: al.FuncRead,
		Data: []byte{0x3C, 0x01, 0x06}}
	in := buildAppFrame(t, frame.Control{DIR: true, PRM: true, FCV: true, FuncCode: frame.FuncConfirmedUserData}, read)

	sent := runOutstationOnce(t, ost, tr, in)
	// Expect the solicited response AND the application confirmation.
	if len(sent) != 2 {
		t.Fatalf("expected 2 emitted frames (response + confirmation), got %d", len(sent))
	}
	for i, raw := range sent {
		f, err := frame.Decode(raw)
		if err != nil {
			t.Fatalf("frame %d decode: %v", i, err)
		}
		if f.Control.DIR || !f.Control.PRM || f.Control.FuncCode != frame.FuncUnconfirmedUserData {
			t.Errorf("frame %d = DIR=%v PRM=%v func=%d, want primary func 4", i, f.Control.DIR, f.Control.PRM, f.Control.FuncCode)
		}
	}
}

// TestLinkControlResponsesRemainSecondary: Reset -> ACK and Link Status Request
// -> LINK_STATUS must keep PRM=false (secondary) and their proper function codes.
func TestLinkControlResponsesRemainSecondary(t *testing.T) {
	reset, err := frame.Encode(&frame.Frame{Control: frame.Control{DIR: true, PRM: true, FuncCode: frame.FuncResetLinkStations}, DestAddr: 4, SrcAddr: 3})
	if err != nil {
		t.Fatalf("encode reset: %v", err)
	}
	linkStatus, err := frame.Encode(&frame.Frame{Control: frame.Control{DIR: true, PRM: true, FuncCode: frame.FuncRequestLinkStatus}, DestAddr: 4, SrcAddr: 3})
	if err != nil {
		t.Fatalf("encode link status: %v", err)
	}

	ost := newResponseTestOutstation()
	tr := newScriptedTransport(nil)
	sent := runOutstationOnce(t, ost, tr, reset, linkStatus)
	if len(sent) != 2 {
		t.Fatalf("expected 2 link responses, got %d", len(sent))
	}
	ack, err := frame.Decode(sent[0])
	if err != nil {
		t.Fatalf("decode ack: %v", err)
	}
	if ack.Control.PRM || ack.Control.DIR || ack.Control.FuncCode != frame.FuncAck {
		t.Errorf("reset response = DIR=%v PRM=%v func=%d, want secondary ACK", ack.Control.DIR, ack.Control.PRM, ack.Control.FuncCode)
	}
	ls, err := frame.Decode(sent[1])
	if err != nil {
		t.Fatalf("decode link status: %v", err)
	}
	if ls.Control.PRM || ls.Control.DIR || ls.Control.FuncCode != frame.FuncLinkStatus {
		t.Errorf("link status response = DIR=%v PRM=%v func=%d, want secondary LINK_STATUS", ls.Control.DIR, ls.Control.PRM, ls.Control.FuncCode)
	}
}

// TestErrorResponseIsPrimaryUserData: an unsupported application function yields
// an error response that is also framed as primary user data (not secondary).
func TestErrorResponseIsPrimaryUserData(t *testing.T) {
	ost := newResponseTestOutstation()
	tr := newScriptedTransport(nil)

	// Application function 0x7E is not handled -> ProcessRequest returns an error,
	// triggering sendErrorResponse.
	bad := &al.APDU{Control: al.AppControl{FIR: true, FIN: true, Seq: 3}, FuncCode: 0x7E}
	in := buildAppFrame(t, frame.Control{DIR: true, PRM: true, FCV: true, FuncCode: frame.FuncConfirmedUserData}, bad)

	sent := runOutstationOnce(t, ost, tr, in)
	if len(sent) != 1 {
		t.Fatalf("expected exactly 1 error-response frame, got %d", len(sent))
	}
	f, err := frame.Decode(sent[0])
	if err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if f.Control.DIR || !f.Control.PRM || f.Control.FuncCode != frame.FuncUnconfirmedUserData {
		t.Errorf("error response = DIR=%v PRM=%v func=%d, want primary func 4", f.Control.DIR, f.Control.PRM, f.Control.FuncCode)
	}
	msg := reassembleFramePayload(t, sent[0])
	resp, err := al.Decode(msg)
	if err != nil {
		t.Fatalf("decode error APDU: %v", err)
	}
	if resp.FuncCode != al.FuncResponse {
		t.Errorf("error application func=0x%02X, want RESPONSE", resp.FuncCode)
	}
	iin, err := al.DecodeIIN(resp.Data)
	if err != nil {
		t.Fatalf("decode IIN: %v", err)
	}
	if !iin.FuncUnknown {
		t.Errorf("error response IIN FuncUnknown=false, want true")
	}
}

// TestMalformedInboundStillFailsClosed confirms the framing change did not
// weaken input validation: a mis-addressed frame and a corrupt frame produce no
// application response.
func TestMalformedInboundStillFailsClosed(t *testing.T) {
	read := &al.APDU{Control: al.AppControl{FIR: true, FIN: true, Seq: 1}, FuncCode: al.FuncRead, Data: []byte{0x3C, 0x01, 0x06}}

	// Wrong destination address (9, not our 4).
	wrongDst, err := frame.Encode(&frame.Frame{
		Control:  frame.Control{DIR: true, PRM: true, FCV: true, FuncCode: frame.FuncConfirmedUserData},
		DestAddr: 9, SrcAddr: 3,
		Data: tl.EncodeFragment(tl.Fragment{FIR: true, FIN: true, Seq: 0, Data: read.Encode()}),
	})
	if err != nil {
		t.Fatalf("encode wrong-dst: %v", err)
	}
	// Corrupt header CRC of a valid frame.
	good := buildAppFrame(t, frame.Control{DIR: true, PRM: true, FCV: true, FuncCode: frame.FuncConfirmedUserData}, read)
	corrupt := append([]byte(nil), good...)
	corrupt[8] ^= 0xFF

	ost := newResponseTestOutstation()
	tr := newScriptedTransport(nil)
	sent := runOutstationOnce(t, ost, tr, wrongDst, corrupt)
	if len(sent) != 0 {
		t.Fatalf("mis-addressed/corrupt input must produce no response, got %d frame(s)", len(sent))
	}
}
