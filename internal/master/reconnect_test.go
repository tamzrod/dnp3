package master

import (
	"io"
	"sync"
	"testing"

	"dnp3/internal/dll/frame"
	"dnp3/internal/tl"
)

const (
	reconnectMasterAddr    uint16 = 0x0003
	reconnectOutstationID  uint16 = 0x0004
)

// switchableTransport returns the secondary frames required by the link
// handshake (ACK then Link Status) while "healthy", and io.EOF while "dropped".
// Each Connect consumes one ACK + one Link Status; the cycle repeats so
// reconnects work (DNP3-032).
type switchableTransport struct {
	mu      sync.Mutex
	dropped bool
	calls   int
}

func (t *switchableTransport) Send(data []byte) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.dropped {
		return io.EOF
	}
	return nil
}

func (t *switchableTransport) Receive() ([]byte, error) {
	t.mu.Lock()
	dropped := t.dropped
	n := t.calls
	t.calls++
	t.mu.Unlock()
	if dropped {
		return nil, io.EOF
	}
	// Even-indexed receives answer Reset Link Stations with an ACK; odd-indexed
	// receives answer Request Link Status with a Link Status frame. Both are
	// secondary (DIR=0, PRM=0), from the outstation to the master.
	var fc uint8 = frame.FuncAck
	if n%2 == 1 {
		fc = frame.FuncLinkStatus
	}
	f := &frame.Frame{
		Control:  frame.Control{DIR: false, PRM: false, FuncCode: fc},
		DestAddr: reconnectMasterAddr,
		SrcAddr:  reconnectOutstationID,
	}
	raw, err := frame.Encode(f)
	if err != nil {
		return nil, err
	}
	return raw, nil
}

func (t *switchableTransport) SetTimeout(ms int) {}

func (t *switchableTransport) setDropped(b bool) {
	t.mu.Lock()
	t.dropped = b
	t.mu.Unlock()
}

// TestReconnectReHandshakes asserts that after a mid-session transport drop,
// the master can Connect again: TL state is cleared and the link handshake is
// re-issued, leaving the master Active (DNP3-032).
func TestReconnectReHandshakes(t *testing.T) {
	m := NewMaster(&Config{MasterAddress: reconnectMasterAddr})
	tr := &switchableTransport{}
	m.SetTransport(tr)
	m.AddOutstation(reconnectOutstationID, "rtu-1")

	if err := m.Connect(); err != nil {
		t.Fatalf("initial Connect: %v", err)
	}
	if m.State() != StateActive {
		t.Fatalf("after initial Connect state = %v, want Active", m.State())
	}

	// Simulate a mid-session peer drop.
	tr.setDropped(true)

	// A send attempt now observes the disconnect and transitions to StateError.
	m.SetState(StateError)

	// Recover: transport comes back, reconnect re-handshakes from a clean slate.
	tr.setDropped(false)
	if err := m.Connect(); err != nil {
		t.Fatalf("reconnect: %v", err)
	}
	if m.State() != StateActive {
		t.Fatalf("after reconnect state = %v, want Active", m.State())
	}
}

// TestResetForReconnectClearsTLState asserts that resetForReconnect clears the
// reassembler so no stale fragments survive a reconnect (DNP3-032). The
// detailed cross-session isolation test is DNP3-033.
func TestResetForReconnectClearsTLState(t *testing.T) {
	m := NewMaster(nil)
	// Push a FIR-only fragment so the reassembler has pending state.
	_, _ = m.reassembler.Push(tl.Fragment{FIR: true, FIN: false, Data: []byte{0x01}})
	if m.reassembler.BufferLen() == 0 {
		t.Fatal("expected non-empty reassembler before reset")
	}

	m.resetForReconnect()

	if m.reassembler.BufferLen() != 0 {
		t.Fatalf("reassembler BufferLen after reset = %d, want 0", m.reassembler.BufferLen())
	}
}


