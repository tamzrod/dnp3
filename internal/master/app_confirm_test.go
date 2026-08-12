package master

import (
	"testing"

	"dnp3/internal/al"
	"dnp3/internal/dll/frame"
	"dnp3/internal/tl"
)

// conResponseTransport returns a single application response with the CON bit
// set (requesting a master application confirm) echoing the request's
// sequence, then records every frame the master sends so the test can assert
// the confirm was emitted.
type conResponseTransport struct {
	sent      [][]byte
	lastSeq   uint8
	responses [][]byte
	idx       int
}

func (t *conResponseTransport) Send(data []byte) error {
	cp := make([]byte, len(data))
	copy(cp, data)
	t.sent = append(t.sent, cp)
	// Track the request seq so the canned response can echo it.
	if f, err := frame.Decode(data); err == nil {
		if frag, err := tl.DecodeFragment(f.Data); err == nil {
			if apdu, err := al.Decode(frag.Data); err == nil {
				t.lastSeq = apdu.Control.Seq
			}
		}
	}
	return nil
}

func (t *conResponseTransport) Receive() ([]byte, error) {
	if t.idx < len(t.responses) {
		f := t.responses[t.idx]
		t.idx++
		return f, nil
	}
	// Fallback: echo a CON response with the last request seq (for the
	// sendWithRetry request/confirm/response flow).
	return buildCONResponseFrame(t.lastSeq), nil
}

func (t *conResponseTransport) SetTimeout(ms int) {}

// buildCONResponseFrame builds a valid DLL+TL+APDU response frame with the
// application CON bit set (confirmation required) and the given sequence.
func buildCONResponseFrame(seq uint8) []byte {
	apdu := &al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, CON: true, Seq: seq},
		FuncCode: al.FuncResponse,
		Data:     []byte{0x00, 0x00}, // clean IIN
	}
	frag := tl.Fragment{FIR: true, FIN: true, Data: apdu.Encode()}
	tlData := tl.EncodeFragment(frag)
	dllFrame := &frame.Frame{
		Control:  frame.Control{DIR: false, PRM: false, FuncCode: frame.FuncConfirmedUserDataR},
		DestAddr: 1, SrcAddr: 2, Data: tlData,
	}
	raw, _ := frame.Encode(dllFrame)
	return raw
}

// TestSendsApplicationConfirmOnCONResponse verifies DNP3-054: when the
// outstation's response carries CON=1, the master sends an application-layer
// confirm APDU (FuncCode 0, CON=0) with the matching sequence before
// processing the data. The confirm must appear among the frames the master
// sent.
func TestSendsApplicationConfirmOnCONResponse(t *testing.T) {
	m := NewMaster(&Config{MasterAddress: 1, Timeout: 50, MaxRetries: 1, RetryDelay: 0})
	tr := &conResponseTransport{}
	m.SetTransport(tr)

	// Issue a read; the canned response carries CON=1.
	req := &al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: 0},
		FuncCode: al.FuncRead,
		Data:     []byte{0x01, 0x00, 0x06, 0x00}, // G1V0 all-objects
	}
	if _, err := m.SendRequestWithRetryAndGetResponse(req, 2); err != nil {
		t.Fatalf("SendRequestWithRetryAndGetResponse: %v", err)
	}

	// Find a sent frame whose decoded APDU is a confirm (FuncCode 0, CON=0)
	// and whose seq matches the response seq (the request seq, since the
	// response echoes it).
	var gotSeq uint8
	var foundConfirm bool
	for _, raw := range tr.sent {
		f, err := frame.Decode(raw)
		if err != nil {
			continue
		}
		// Only master-originated user-data frames are candidates.
		if !f.Control.PRM {
			continue
		}
		frag, err := tl.DecodeFragment(f.Data)
		if err != nil {
			continue
		}
		apdu, err := al.Decode(frag.Data)
		if err != nil {
			continue
		}
		if apdu.FuncCode == al.FuncConfirm && !apdu.Control.CON {
			foundConfirm = true
			gotSeq = apdu.Control.Seq
			break
		}
	}
	if !foundConfirm {
		t.Fatalf("master did not send an application confirm (CON=0) for the CON=1 response; sent %d frames", len(tr.sent))
	}
	// The response echoed the request seq (0 for a fresh master).
	if gotSeq != 0 {
		t.Errorf("confirm seq = %d, want 0 (matches request/response seq)", gotSeq)
	}
}

// TestNoApplicationConfirmWhenCONClear verifies DNP3-054 does not over-fire:
// a normal response (CON=0) must NOT trigger an application confirm.
func TestNoApplicationConfirmWhenCONClear(t *testing.T) {
	m := NewMaster(&Config{MasterAddress: 1, Timeout: 50, MaxRetries: 1, RetryDelay: 0})
	tr := &conResponseTransport{}
	m.SetTransport(tr)

	req := &al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: 0},
		FuncCode: al.FuncRead,
		Data:     []byte{0x01, 0x00, 0x06, 0x00},
	}
	// The fallback Receive returns a CON response — override to a clean one.
	tr.responses = [][]byte{buildResponseFrameWithSeq(0)}
	if _, err := m.SendRequestWithRetryAndGetResponse(req, 2); err != nil {
		t.Fatalf("SendRequestWithRetryAndGetResponse: %v", err)
	}

	for _, raw := range tr.sent {
		f, err := frame.Decode(raw)
		if err != nil {
			continue
		}
		if !f.Control.PRM {
			continue
		}
		frag, err := tl.DecodeFragment(f.Data)
		if err != nil {
			continue
		}
		apdu, err := al.Decode(frag.Data)
		if err != nil {
			continue
		}
		if apdu.FuncCode == al.FuncConfirm {
			t.Fatalf("master sent an unexpected application confirm for a CON=0 response")
		}
	}
}
