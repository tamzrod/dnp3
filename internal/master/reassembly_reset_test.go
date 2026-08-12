package master

import (
	"bytes"
	"testing"

	"dnp3/internal/tl"
)

// TestReassemblerNoCrossSessionPollution asserts that a partial fragment left
// in the reassembler from a dropped session does not leak into the next
// session's reassembled message after resetForReconnect clears state
// (DNP3-033).
func TestReassemblerNoCrossSessionPollution(t *testing.T) {
	m := NewMaster(nil)

	// Session 1: a FIR-only fragment is received, then the link drops before the
	// matching FIN arrives. The reassembler retains the partial data.
	stale := []byte{0xDE, 0xAD, 0xBE, 0xEF}
	if _, err := m.reassembler.Push(tl.Fragment{FIR: true, FIN: false, Seq: 1, Data: stale}); err != nil {
		t.Fatalf("session-1 push: %v", err)
	}
	if m.reassembler.BufferLen() != len(stale) {
		t.Fatalf("stale buffer len = %d, want %d", m.reassembler.BufferLen(), len(stale))
	}

	// Reconnect clears TL state (DNP3-032/DNP3-033).
	m.resetForReconnect()
	if m.reassembler.BufferLen() != 0 {
		t.Fatalf("after reset buffer len = %d, want 0", m.reassembler.BufferLen())
	}
	if m.reassembler.IsComplete() {
		t.Fatal("reassembler marked complete after reset")
	}

	// Session 2: a fresh single-fragment message arrives. Its reassembled bytes
	// must equal exactly the new payload — not the new payload prefixed with the
	// stale session-1 bytes.
	fresh := []byte{0x01, 0x02, 0x03}
	got, err := m.reassembler.Push(tl.Fragment{FIR: true, FIN: true, Seq: 1, Data: fresh})
	if err != nil {
		t.Fatalf("session-2 push: %v", err)
	}
	if !bytes.Equal(got, fresh) {
		t.Fatalf("session-2 reassembled = % X, want % X (stale data leaked)", got, fresh)
	}
}

// TestResetForReconnectResetsFragmenter asserts the fragmenter sequence is
// reset on reconnect so the next session's outgoing fragments start from the
// initial sequence (DNP3-033). Verified indirectly: after fragmenting data and
// resetting, a fresh Fragmentize produces fragments whose first Seq is 0.
func TestResetForReconnectResetsFragmenter(t *testing.T) {
	m := NewMaster(nil)
	// Advance the fragmenter by fragmenting data that spans multiple fragments.
	big := make([]byte, tl.MaxFragmentData*2+4)
	for i := range big {
		big[i] = byte(i)
	}
	frags := m.fragmenter.Fragmentize(big)
	if len(frags) < 2 {
		t.Fatalf("expected >=2 fragments, got %d", len(frags))
	}

	m.resetForReconnect()

	// After reset, a fresh Fragmentize must start at Seq 0 again.
	frags2 := m.fragmenter.Fragmentize(big)
	if len(frags2) == 0 {
		t.Fatal("no fragments after reset")
	}
	if frags2[0].Seq != 0 {
		t.Fatalf("first fragment Seq after reset = %d, want 0", frags2[0].Seq)
	}
}
