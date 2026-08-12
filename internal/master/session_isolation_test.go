package master

import (
	"testing"

	"dnp3/internal/al"
)

// TestSessionIsolationNoCrossTalk verifies DNP3-055: each outstation has an
// independent application-layer sequence stream, so interleaved operations on
// two outstations do not cross-talk. Outstation 2's stream must continue 0,1,2
// undisturbed by traffic to outstation 3, which has its own 0,1 stream.
func TestSessionIsolationNoCrossTalk(t *testing.T) {
	m := NewMaster(DefaultConfig())
	tr := &echoSeqTransport{}
	m.SetTransport(tr)
	m.SetState(StateInitialized)
	m.AddOutstation(2, "RTU-A")
	m.AddOutstation(3, "RTU-B")

	req := buildRequest(0, al.FuncRead, []byte{0x00})

	// Outstation 2: two requests → SEQ 0, 1.
	for i, want := range []uint8{0, 1} {
		before := len(tr.sent)
		if err := m.sendWithRetry(req, 2); err != nil {
			t.Fatalf("os2 send[%d]: %v", i, err)
		}
		if got := extractRequestSeq(t, tr.sent[before]); got != want {
			t.Fatalf("os2 request %d SEQ = %d, want %d", i, got, want)
		}
	}

	// Outstation 3: one request → SEQ 0 (independent stream, NOT 2).
	before := len(tr.sent)
	if err := m.sendWithRetry(req, 3); err != nil {
		t.Fatalf("os3 send: %v", err)
	}
	if got := extractRequestSeq(t, tr.sent[before]); got != 0 {
		t.Fatalf("os3 request SEQ = %d, want 0 (independent stream)", got)
	}

	// Outstation 2 again → SEQ 2 (its stream continued; os3 traffic did not
	// advance it, nor did os3's SEQ-0 response confuse os2's seq tracking).
	before = len(tr.sent)
	if err := m.sendWithRetry(req, 2); err != nil {
		t.Fatalf("os2 send[2]: %v", err)
	}
	if got := extractRequestSeq(t, tr.sent[before]); got != 2 {
		t.Fatalf("os2 request 2 SEQ = %d, want 2 (no cross-talk)", got)
	}

	// Final per-outstation sequence state is independent.
	if got := m.currentSequence(2); got != 3 {
		t.Fatalf("os2 currentSequence = %d, want 3", got)
	}
	if got := m.currentSequence(3); got != 1 {
		t.Fatalf("os3 currentSequence = %d, want 1", got)
	}
}

// TestSessionIsolationOutstationSeqWrap verifies DNP3-055 + DNP3-008: an
// outstation's independent sequence wraps 0-15 independently of other
// outstations. Drive outstation 2 through 16 advances (→0) while outstation 3
// is untouched (stays 0).
func TestSessionIsolationOutstationSeqWrap(t *testing.T) {
	m := NewMaster(DefaultConfig())
	tr := &echoSeqTransport{}
	m.SetTransport(tr)
	m.SetState(StateInitialized)
	m.AddOutstation(2, "RTU-A")
	m.AddOutstation(3, "RTU-B")

	req := buildRequest(0, al.FuncRead, []byte{0x00})

	// 16 successful sends to outstation 2 → seq advances 16 times: 0→16≡0.
	for i := 0; i < 16; i++ {
		if err := m.sendWithRetry(req, 2); err != nil {
			t.Fatalf("os2 send[%d]: %v", i, err)
		}
	}
	if got := m.currentSequence(2); got != 0 {
		t.Fatalf("os2 currentSequence after 16 sends = %d, want 0 (wrap)", got)
	}
	// Outstation 3 was never touched: its stream is still 0.
	if got := m.currentSequence(3); got != 0 {
		t.Fatalf("os3 currentSequence = %d, want 0 (no cross-talk from os2)", got)
	}
}
