package master

import (
	"context"
	"testing"

	"dnp3/pkg/dnp3"
	"dnp3/pkg/dnp3/types"
)

// firstSeqEchoTransport records the application SEQ of the first request sent
// and then echoes the most recent request's SEQ in a minimal valid Read
// response (IIN only). Used to prove a fresh client starts its AC sequence at
// 0 (DNP3-068) while still answering subsequent reads.
type firstSeqEchoTransport struct {
	sent     int
	firstSeq uint8
	lastSeq  uint8
}

func (t *firstSeqEchoTransport) Send(data []byte) error {
	seq := extractPubRequestSeq(data)
	if t.sent == 0 {
		t.firstSeq = seq
	}
	t.lastSeq = seq
	t.sent++
	return nil
}

func (t *firstSeqEchoTransport) Receive() ([]byte, error) {
	return buildPubReadResponse(t.lastSeq), nil
}

func (t *firstSeqEchoTransport) SetTimeout(ms int) {}

// TestMasterRestartAfterCloseIndependent verifies DNP3-068: after Close of a
// client, a new NewClient is fully independent — no global/package state from
// the previous client (public AC sequence, internal master state, per-outstation
// sequence) leaks into the new client. The new client's first request must
// carry AC sequence 0.
func TestMasterRestartAfterCloseIndependent(t *testing.T) {
	// Client A: connect, perform a Read (advances A's public sequence to 1 and
	// the internal master's per-outstation sequence).
	trA := &firstSeqEchoTransport{}
	ccA := newConnectedClientWithTransport(t, trA)
	ccA.config.OutstationAddress = 1024
	ccA.internalMaster.AddOutstation(1024, "RTU-A")

	if _, err := ccA.Read(context.Background(), &types.ReadRequest{
		Groups: []types.GroupRequest{{Group: 1, Variation: 0}},
	}); err != nil {
		t.Fatalf("client A Read: %v", err)
	}
	if trA.firstSeq != 0 {
		t.Fatalf("client A first request SEQ = %d, want 0", trA.firstSeq)
	}
	// A's public sequence is now 1 (advanced from 0 after one Read).
	ccA.mu.RLock()
	seqA := ccA.sequence
	ccA.mu.RUnlock()
	if seqA != 1 {
		t.Fatalf("client A public sequence after one Read = %d, want 1", seqA)
	}

	// Close A cleanly (DNP3-050: resets A's state; must not affect a new client).
	if err := ccA.Close(); err != nil {
		t.Fatalf("client A Close: %v", err)
	}
	if got := ccA.State(); got != dnp3.StateDisconnected {
		t.Fatalf("client A state after Close = %v, want Disconnected", got)
	}

	// Client B: a brand-new NewClient with a brand-new transport. It must NOT
	// inherit A's advanced public sequence (1) nor A's internal per-outstation
	// sequence — its first request must carry AC sequence 0.
	trB := &firstSeqEchoTransport{}
	ccB := newConnectedClientWithTransport(t, trB)
	ccB.config.OutstationAddress = 1024
	ccB.internalMaster.AddOutstation(1024, "RTU-B")

	if _, err := ccB.Read(context.Background(), &types.ReadRequest{
		Groups: []types.GroupRequest{{Group: 1, Variation: 0}},
	}); err != nil {
		t.Fatalf("client B Read: %v", err)
	}
	if trB.firstSeq != 0 {
		t.Fatalf("client B first request SEQ = %d, want 0 (no leakage from client A)", trB.firstSeq)
	}
	// B's public sequence must start from 0 (not A's terminal 1).
	ccB.mu.RLock()
	seqB := ccB.sequence
	ccB.mu.RUnlock()
	if seqB != 1 {
		t.Fatalf("client B public sequence after one Read = %d, want 1 (started from 0)", seqB)
	}

	// Closing A must not have torn down B: B's state is still Connected and a
	// second Read on B still succeeds (B owns its own transport + master).
	if got := ccB.State(); got != dnp3.StateConnected {
		t.Fatalf("client B state after A.Close = %v, want Connected (independent)", got)
	}
	if _, err := ccB.Read(context.Background(), &types.ReadRequest{
		Groups: []types.GroupRequest{{Group: 1, Variation: 0}},
	}); err != nil {
		t.Fatalf("client B second Read after A.Close: %v", err)
	}
}
