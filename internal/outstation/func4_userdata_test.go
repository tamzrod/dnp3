// AO-004 func-4 regression tests.
//
// These drive real encoded DLL frames through the outstation's reassembleMessage
// (and, where a transport is needed, a tiny recording transport) to lock in the
// minimal corrective: primary user data is accepted for BOTH link function 3
// (CONFIRMED_USER_DATA) and function 4 (UNCONFIRMED_USER_DATA); every other
// primary/secondary/mis-addressed/malformed case stays fail-closed.
package outstation

import (
	"testing"

	"dnp3/internal/al"
	"dnp3/internal/dll/frame"
	"dnp3/internal/tl"
)

// recordingTransport captures every frame the outstation sends, so tests can
// assert the link layer produced (or did NOT produce) an ACK. It never blocks.
type recordingTransport struct {
	sent [][]byte
}

func (t *recordingTransport) Send(data []byte) error {
	cp := make([]byte, len(data))
	copy(cp, data)
	t.sent = append(t.sent, cp)
	return nil
}
func (t *recordingTransport) Receive() ([]byte, error) { return nil, nil }
func (t *recordingTransport) SetTimeout(ms int)        {}

// linkFrame builds a real encoded DLL frame carrying one transport fragment of
// the given application APDU. prm/conf select the link function code.
func linkFrame(t *testing.T, ctrl frame.Control, dst, src uint16, apdu *al.APDU) []byte {
	t.Helper()
	enc := apdu.Encode()
	frag := tl.EncodeFragment(tl.Fragment{FIR: true, FIN: true, Seq: 0, Data: enc})
	encoded, err := frame.Encode(&frame.Frame{
		Control:  ctrl,
		DestAddr: dst,
		SrcAddr:  src,
		Data:     frag,
	})
	if err != nil {
		t.Fatalf("frame.Encode: %v", err)
	}
	return encoded
}

// readAPDU is a well-formed Class-0 Read request (function 1).
func readAPDU() *al.APDU {
	return &al.APDU{
		Control:  al.AppControl{FIR: true, FIN: true, Seq: 0},
		FuncCode: al.FuncRead,
		Data:     []byte{0x3C, 0x01, 0x06}, // G60V1, all-objects qualifier 0x06
	}
}

func newTestOutstation(t *testing.T) (*Outstation, *recordingTransport) {
	t.Helper()
	cfg := DefaultConfig()
	cfg.OutstationAddress = 4
	cfg.MasterAddress = 3
	ost := NewOutstation(cfg)
	tr := &recordingTransport{}
	ost.SetTransport(tr)
	return ost, tr
}

// TestReassembleAcceptsConfirmedAndUnconfirmedUserData is the core regression:
// identical payloads carried as wire function 3 and wire function 4 must both be
// reassembled to the same application message, with no link ACK spuriously
// produced by either.
func TestReassembleAcceptsConfirmedAndUnconfirmedUserData(t *testing.T) {
	apdu := readAPDU()

	for _, tc := range []struct {
		name string
		ctrl frame.Control
	}{
		{"confirmed_func3", frame.Control{DIR: true, PRM: true, FCV: true, FuncCode: frame.FuncConfirmedUserData}},
		{"unconfirmed_func4", frame.Control{DIR: true, PRM: true, FuncCode: frame.FuncUnconfirmedUserData}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ost, tr := newTestOutstation(t)
			msg, err := ost.reassembleMessage(linkFrame(t, tc.ctrl, 4, 3, apdu))
			if err != nil {
				t.Fatalf("reassembleMessage error: %v", err)
			}
			if msg == nil {
				t.Fatal("expected reassembled message, got nil (frame dropped)")
			}
			got, err := al.Decode(msg)
			if err != nil {
				t.Fatalf("al.Decode: %v", err)
			}
			if got.FuncCode != al.FuncRead {
				t.Fatalf("func code = %d, want %d", got.FuncCode, al.FuncRead)
			}
			// Neither confirmed nor unconfirmed user data should trigger a bare
			// link ACK merely for carrying data.
			if len(tr.sent) != 0 {
				t.Fatalf("unexpected link response(s) for %s: %d frame(s)", tc.name, len(tr.sent))
			}
		})
	}
}

// TestReassembleRejectsUnsupportedPrimaryFunction verifies the fix did not
// become a blanket acceptance of every primary function code.
func TestReassembleRejectsUnsupportedPrimaryFunction(t *testing.T) {
	apdu := readAPDU()
	// Function 2 (TEST_LINK_STATES) is a primary link-control function, not user
	// data; it must not yield an application message.
	msg, err := mustReassemble(t, frame.Control{DIR: true, PRM: true, FuncCode: frame.FuncTestLinkStates}, 4, 3, apdu)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg != nil {
		t.Fatal("unsupported primary function must not be accepted as user data")
	}
}

// TestReassembleRejectsSecondaryUserData ensures PRM=false (secondary) frames
// are never treated as master user data.
func TestReassembleRejectsSecondaryUserData(t *testing.T) {
	apdu := readAPDU()
	// Same function code but PRM=false: a secondary-station frame.
	msg, err := mustReassemble(t, frame.Control{DIR: true, PRM: false, FuncCode: frame.FuncUnconfirmedUserData}, 4, 3, apdu)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg != nil {
		t.Fatal("secondary (PRM=false) user-data frame must be ignored")
	}
}

// TestReassembleIgnoresWrongDestination ensures frames not addressed to this
// outstation are ignored regardless of function code.
func TestReassembleIgnoresWrongDestination(t *testing.T) {
	apdu := readAPDU()
	msg, err := mustReassemble(t, frame.Control{DIR: true, PRM: true, FuncCode: frame.FuncUnconfirmedUserData}, 9, 3, apdu)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg != nil {
		t.Fatal("frame addressed to another outstation must be ignored")
	}
}

// TestReassembleRejectsMalformedAndShortFrames ensures corrupt/short input fails
// closed (no message, no panic).
func TestReassembleRejectsMalformedAndShortFrames(t *testing.T) {
	ost, _ := newTestOutstation(t)
	valid := linkFrame(t, frame.Control{DIR: true, PRM: true, FuncCode: frame.FuncUnconfirmedUserData}, 4, 3, readAPDU())

	cases := map[string][]byte{
		"empty":         {},
		"single_byte":   {0x05},
		"sync_only":     {0x05, 0x64},
		"bad_header_crc": func() []byte { b := append([]byte(nil), valid...); b[8] ^= 0xFF; return b }(),
		"bad_payload_crc": func() []byte {
			b := append([]byte(nil), valid...)
			b[len(b)-1] ^= 0xFF
			return b
		}(),
	}
	for name, in := range cases {
		t.Run(name, func(t *testing.T) {
			msg, err := ost.reassembleMessage(in)
			if err != nil {
				t.Fatalf("reassembleMessage(%s) error: %v", name, err)
			}
			if msg != nil {
				t.Fatalf("malformed/short frame %s must not produce a message", name)
			}
		})
	}
}

// TestReassemblePreservesResetAndLinkStatusHandling confirms the corrective did
// not disturb link-control handling: Reset Link Stations still yields a link ACK
// and a Link Status Request still yields a link status frame.
func TestReassemblePreservesResetAndLinkStatusHandling(t *testing.T) {
	// Reset Link Stations -> ACK.
	ost, tr := newTestOutstation(t)
	resetFrame, err := frame.Encode(&frame.Frame{
		Control:  frame.Control{DIR: true, PRM: true, FuncCode: frame.FuncResetLinkStations},
		DestAddr: 4,
		SrcAddr:  3,
	})
	if err != nil {
		t.Fatalf("encode reset: %v", err)
	}
	if msg, err := ost.reassembleMessage(resetFrame); err != nil || msg != nil {
		t.Fatalf("reset frame: msg=%v err=%v (want nil,nil)", msg, err)
	}
	if len(tr.sent) != 1 {
		t.Fatalf("reset frame must produce exactly one link response, got %d", len(tr.sent))
	}
	ack, err := frame.Decode(tr.sent[0])
	if err != nil {
		t.Fatalf("decode ACK: %v", err)
	}
	if ack.Control.PRM || ack.Control.FuncCode != frame.FuncAck {
		t.Fatalf("reset response = PRM=%v func=%d, want secondary ACK", ack.Control.PRM, ack.Control.FuncCode)
	}

	// Link Status Request -> link status.
	ost2, tr2 := newTestOutstation(t)
	lsFrame, err := frame.Encode(&frame.Frame{
		Control:  frame.Control{DIR: true, PRM: true, FuncCode: frame.FuncRequestLinkStatus},
		DestAddr: 4,
		SrcAddr:  3,
	})
	if err != nil {
		t.Fatalf("encode link status: %v", err)
	}
	if msg, err := ost2.reassembleMessage(lsFrame); err != nil || msg != nil {
		t.Fatalf("link status frame: msg=%v err=%v (want nil,nil)", msg, err)
	}
	if len(tr2.sent) != 1 {
		t.Fatalf("link status request must produce exactly one response, got %d", len(tr2.sent))
	}
	ls, err := frame.Decode(tr2.sent[0])
	if err != nil {
		t.Fatalf("decode link status: %v", err)
	}
	if ls.Control.PRM || ls.Control.FuncCode != frame.FuncLinkStatus {
		t.Fatalf("link status response = PRM=%v func=%d, want secondary LinkStatus", ls.Control.PRM, ls.Control.FuncCode)
	}
}

// mustReassemble is a small helper for the negative-path assertions.
func mustReassemble(t *testing.T, ctrl frame.Control, dst, src uint16, apdu *al.APDU) ([]byte, error) {
	t.Helper()
	ost, _ := newTestOutstation(t)
	return ost.reassembleMessage(linkFrame(t, ctrl, dst, src, apdu))
}
