package master

import (
	"errors"
	"testing"
)

// TestStateTransitionTableLegal verifies every transition in
// legalStateTransitions is accepted by transitionTo (DNP3-039).
func TestStateTransitionTableLegal(t *testing.T) {
	for from, targets := range legalStateTransitions {
		for to := range targets {
			m := NewMaster(nil)
			m.SetState(from) // force the source state via the escape hatch
			if err := m.transitionTo(to); err != nil {
				t.Errorf("legal transition %s -> %s rejected: %v", from, to, err)
			}
			if m.State() != to {
				t.Errorf("after %s -> %s, state = %s", from, to, m.State())
			}
		}
	}
}

// TestStateTransitionTableIllegal verifies a representative set of illegal
// transitions are rejected with ErrIllegalStateTransition (DNP3-039). The key
// invariant: operations-ready states (Initialized/Active) must not be
// reachable from Disconnected without Connect, and Disconnected must not be
// reachable from operational states except via Disconnect.
func TestStateTransitionTableIllegal(t *testing.T) {
	illegal := []struct {
		from, to State
	}{
		{StateDisconnected, StateActive},      // must Connect first
		{StateDisconnected, StateInitialized}, // must Connect + Initialize
		{StateDisconnected, StateError},        // only via a send failure, not direct
		{StateActive, StateConnecting},         // reconnect only from Error
		{StateInitialized, StateConnected},     // backward into handshake state
		{StateError, StateActive},              // Error must re-Connect, not jump
		{StateError, StateInitialized},         // Error must re-Connect
	}
	for _, tc := range illegal {
		m := NewMaster(nil)
		m.SetState(tc.from)
		err := m.transitionTo(tc.to)
		if !errors.Is(err, ErrIllegalStateTransition) {
			t.Errorf("illegal %s -> %s: err = %v, want ErrIllegalStateTransition", tc.from, tc.to, err)
		}
		if m.State() != tc.from {
			t.Errorf("illegal %s -> %s changed state to %s (must be unchanged)", tc.from, tc.to, m.State())
		}
	}
}

// TestStateTransitionSelfIsNoOp verifies a self-transition is an idempotent
// no-op (DNP3-039).
func TestStateTransitionSelfIsNoOp(t *testing.T) {
	for _, s := range []State{
		StateDisconnected, StateConnecting, StateConnected,
		StateInitialized, StateActive, StateError,
	} {
		m := NewMaster(nil)
		m.SetState(s)
		if err := m.transitionTo(s); err != nil {
			t.Errorf("self-transition %s rejected: %v", s, err)
		}
		if m.State() != s {
			t.Errorf("self-transition %s changed state to %s", s, m.State())
		}
	}
}

// TestOperationsRejectedInErrorState is the central DNP3-039 regression: a
// master in StateError must NOT satisfy an operation readiness gate even
// though StateError's iota ordinal (5) exceeds StateActive (4). The legacy
// `State() < StateInitialized` guard would have let operations through on a
// dead link; the formalized isOperational guard rejects them.
func TestOperationsRejectedInErrorState(t *testing.T) {
	m := NewMaster(nil)
	m.SetState(StateError)

	if m.isOperational() {
		t.Fatal("isOperational() = true in StateError; operations would be allowed on a dead link")
	}
	// Poll/Operate/Read* all gate on isOperational; sample via Poll.
	m.AddOutstation(1, "rtu")
	if err := m.Poll(1, PollClass0); err != ErrNotConnected {
		t.Errorf("Poll in StateError: err = %v, want ErrNotConnected", err)
	}
}

// TestOperationsRejectedWhenDisconnected verifies the operational guard
// rejects operations before Connect (DNP3-039).
func TestOperationsRejectedWhenDisconnected(t *testing.T) {
	m := NewMaster(nil)
	m.SetState(StateDisconnected)
	m.AddOutstation(1, "rtu")
	if m.isOperational() {
		t.Fatal("isOperational() = true in StateDisconnected")
	}
	if err := m.Poll(1, PollClass0); err != ErrNotConnected {
		t.Errorf("Poll in StateDisconnected: err = %v, want ErrNotConnected", err)
	}
}

// TestOperationsAcceptedInOperationalStates verifies isOperational admits
// exactly StateInitialized and StateActive (DNP3-039).
func TestOperationsAcceptedInOperationalStates(t *testing.T) {
	for _, s := range []State{StateInitialized, StateActive} {
		m := NewMaster(nil)
		m.SetState(s)
		if !m.isOperational() {
			t.Errorf("isOperational() = false in %s, want true", s)
		}
	}
	for _, s := range []State{StateDisconnected, StateConnecting, StateConnected, StateError} {
		m := NewMaster(nil)
		m.SetState(s)
		if m.isOperational() {
			t.Errorf("isOperational() = true in %s, want false", s)
		}
	}
}

// TestConnectRejectsFromConnecting verifies Connect refuses to re-enter from a
// non-Disconnected/Error state (e.g., a concurrent Connect), surfacing
// ErrIllegalStateTransition instead of silently overwriting state (DNP3-039).
func TestConnectRejectsFromConnecting(t *testing.T) {
	m := NewMaster(nil)
	m.SetTransport(&mockTransport{})
	m.SetState(StateConnecting)
	if err := m.Connect(); !errors.Is(err, ErrIllegalStateTransition) {
		t.Errorf("Connect from Connecting: err = %v, want ErrIllegalStateTransition", err)
	}
	if m.State() != StateConnecting {
		t.Errorf("state changed by rejected Connect: %s", m.State())
	}
}

// TestConnectReconnectFromError verifies Connect is legal from StateError (the
// reconnect path), leaving the master Active (DNP3-032/039). Uses the
// switchableTransport which serves a valid link handshake on each Connect.
func TestConnectReconnectFromError(t *testing.T) {
	m := NewMaster(&Config{MasterAddress: reconnectMasterAddr})
	tr := &switchableTransport{}
	m.SetTransport(tr)
	m.AddOutstation(reconnectOutstationID, "rtu")
	m.SetState(StateError)
	if err := m.Connect(); err != nil {
		t.Fatalf("reconnect from Error: %v", err)
	}
	if m.State() != StateActive {
		t.Fatalf("after reconnect state = %s, want Active", m.State())
	}
}
