package master

import (
	"testing"
	"time"

	"dnp3/internal/al"
	"dnp3/internal/dll/frame"
	"dnp3/internal/tl"
)

// fragConfirmTransport serves a fixed list of raw link frames one per Receive,
// recording every frame the master sends so a test can assert the per-fragment
// application confirmations.
type fragConfirmTransport struct {
	frames [][]byte
	idx    int
	sent   [][]byte
}

func (t *fragConfirmTransport) Send(data []byte) error {
	cp := make([]byte, len(data))
	copy(cp, data)
	t.sent = append(t.sent, cp)
	return nil
}

func (t *fragConfirmTransport) Receive() ([]byte, error) {
	if t.idx < len(t.frames) {
		f := t.frames[t.idx]
		t.idx++
		return f, nil
	}
	return nil, ErrTimeout
}

func (t *fragConfirmTransport) SetTimeout(int) {}

// sentConfirmSeqs returns the application sequence of every confirm APDU the
// master sent.
func sentConfirmSeqs(t *testing.T, sent [][]byte) []uint8 {
	t.Helper()
	var seqs []uint8
	for _, raw := range sent {
		f, err := frame.Decode(raw)
		if err != nil {
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
			seqs = append(seqs, apdu.Control.Seq)
		}
	}
	return seqs
}

// TestSolicitedMultiFragmentConfirmPerFragment verifies that a solicited
// multi-fragment response whose fragments each carry CON=1 (OpenDNP3's
// outstation behavior) is confirmed once per fragment with that fragment's own
// sequence, including the final fragment. Confirming only non-final fragments,
// or confirming the final fragment with the first fragment's sequence, leaves
// the outstation in its confirm-wait state and makes the next request observe
// leftover events.
func TestSolicitedMultiFragmentConfirmPerFragment(t *testing.T) {
	m := NewMaster(&Config{MasterAddress: 1, Timeout: 200, MaxRetries: 1})
	m.AddOutstation(2, "RTU")

	// Fragment 1: FIR=1 FIN=0 CON=1 seq=1. Fragment 2: FIR=0 FIN=1 CON=1 seq=2.
	frag1 := &al.APDU{
		Control:  al.AppControl{FIR: true, FIN: false, CON: true, Seq: 1},
		FuncCode: al.FuncResponse,
		Data:     []byte{0x00, 0x00, 0x02, 0x02, 0x17, 0x01, 0x07, 0x81},
	}
	frag2 := &al.APDU{
		Control:  al.AppControl{FIR: false, FIN: true, CON: true, Seq: 2},
		FuncCode: al.FuncResponse,
		Data:     []byte{0x00, 0x00, 0x02, 0x02, 0x17, 0x01, 0x08, 0x81},
	}
	tr := &fragConfirmTransport{frames: [][]byte{
		encodeUserDataFrame(t, frag1, 0),
		encodeUserDataFrame(t, frag2, 1),
	}}
	m.SetTransport(tr)

	data, err := m.waitForResponseTimeout(200*time.Millisecond, 2)
	if err != nil {
		t.Fatalf("waitForResponseTimeout: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("no reassembled response returned")
	}

	seqs := sentConfirmSeqs(t, tr.sent)
	if len(seqs) != 2 || seqs[0] != 1 || seqs[1] != 2 {
		t.Fatalf("confirm sequences = %v, want [1 2]", seqs)
	}
}

// encodeUserDataFrame wraps one complete application fragment in a single
// transport message and a link frame. Each application fragment travels as a
// complete transport message (transport FIR=FIN=1); the FIR/FIN distinction
// that splits a solicited response lives in the application control byte.
func encodeUserDataFrame(t *testing.T, apdu *al.APDU, seq uint8) []byte {
	t.Helper()
	frag := tl.Fragment{FIR: true, FIN: true, Seq: seq, Data: apdu.Encode()}
	dllFrame := &frame.Frame{
		Control:  frame.Control{DIR: false, PRM: false, FuncCode: frame.FuncConfirmedUserDataR},
		DestAddr: 1, SrcAddr: 2, Data: tl.EncodeFragment(frag),
	}
	raw, err := frame.Encode(dllFrame)
	if err != nil {
		t.Fatalf("frame.Encode: %v", err)
	}
	return raw
}
