package master

import (
	"context"
	"testing"
	"time"

	"dnp3/internal/al"
	"dnp3/internal/dll/frame"
	"dnp3/internal/testutils"
	"dnp3/internal/tl"
	"dnp3/pkg/dnp3/types"
)

// TestIntegrityPollMatchesExplicitReads verifies IntegrityPoll returns the same
// data as explicit per-group Reads of the MVP-supported Class-0 groups
// (DNP3-037). It uses the deterministic in-memory simulator (DNP3-036) so no
// network I/O is involved.
func TestIntegrityPollMatchesExplicitReads(t *testing.T) {
	sim := testutils.NewMVPOutstationSimulator(1024, 0xFFFF)
	sim.SetBinaryInputs([]*types.BinaryInput{
		{Index: 0, Value: true, Quality: types.QualityOnline},
		{Index: 1, Value: false, Quality: types.QualityOnline},
	})
	sim.SetAnalogInputs([]*types.AnalogInput{
		{Index: 0, Value: 42, Quality: types.QualityOnline},
	})
	sim.SetCounters([]*types.Counter{
		{Index: 0, Value: 100, Quality: types.QualityOnline},
		{Index: 1, Value: 7, Quality: types.QualityOnline},
	})

	client, err := NewClientWithTransport(
		NewConfig(
			WithOutstationAddress(1024),
			WithTimeout(2*time.Second),
			WithRetry(1, 0),
		),
		sim,
	)
	if err != nil {
		t.Fatalf("NewClientWithTransport: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	defer client.Close()

	// Explicit per-group reads (the expected reference data).
	binR, err := client.Read(ctx, types.NewReadRequest(types.GroupRequest{Group: 1, Variation: 1}))
	if err != nil {
		t.Fatalf("explicit Read G1: %v", err)
	}
	anR, err := client.Read(ctx, types.NewReadRequest(types.GroupRequest{Group: 30, Variation: 1}))
	if err != nil {
		t.Fatalf("explicit Read G30: %v", err)
	}
	ctrR, err := client.Read(ctx, types.NewReadRequest(types.GroupRequest{Group: 20, Variation: 1}))
	if err != nil {
		t.Fatalf("explicit Read G20: %v", err)
	}

	// IntegrityPoll returns the same combined data.
	poll, err := client.IntegrityPoll(ctx)
	if err != nil {
		t.Fatalf("IntegrityPoll: %v", err)
	}
	if poll == nil {
		t.Fatal("IntegrityPoll returned nil")
	}

	if len(poll.BinaryInputs) != len(binR.BinaryInputs) {
		t.Fatalf("BinaryInputs len = %d, want %d", len(poll.BinaryInputs), len(binR.BinaryInputs))
	}
	for i, p := range poll.BinaryInputs {
		if p.Value != binR.BinaryInputs[i].Value || p.Index != binR.BinaryInputs[i].Index {
			t.Fatalf("BinaryInputs[%d] = %+v, want %+v", i, p, binR.BinaryInputs[i])
		}
	}
	if len(poll.AnalogInputs) != len(anR.AnalogInputs) {
		t.Fatalf("AnalogInputs len = %d, want %d", len(poll.AnalogInputs), len(anR.AnalogInputs))
	}
	if len(poll.Counters) != len(ctrR.Counters) {
		t.Fatalf("Counters len = %d, want %d", len(poll.Counters), len(ctrR.Counters))
	}
	for i, p := range poll.Counters {
		if p.Value != ctrR.Counters[i].Value || p.Index != ctrR.Counters[i].Index {
			t.Fatalf("Counters[%d] = %+v, want %+v", i, p, ctrR.Counters[i])
		}
	}

	// Unsupported types must not be surfaced by IntegrityPoll (DNP3-035).
	if poll.BinaryOutputs != nil || poll.AnalogOutputs != nil || poll.FrozenCounters != nil {
		t.Fatalf("IntegrityPoll surfaced unsupported types: BO=%v AO=%v FC=%v",
			poll.BinaryOutputs, poll.AnalogOutputs, poll.FrozenCounters)
	}
}

// TestIntegrityPollNotConnected asserts IntegrityPoll returns ErrNotConnected
// when the client is not connected, mirroring Read (DNP3-037).
func TestIntegrityPollNotConnected(t *testing.T) {
	sim := testutils.NewMVPOutstationSimulator(1024, 0xFFFF)
	client, err := NewClientWithTransport(
		NewConfig(WithOutstationAddress(1024), WithTimeout(2*time.Second)),
		sim,
	)
	if err != nil {
		t.Fatalf("NewClientWithTransport: %v", err)
	}
	// Do NOT connect.
	if _, err := client.IntegrityPoll(context.Background()); err == nil {
		t.Fatal("IntegrityPoll on a disconnected client returned nil error")
	}
}

// TestIntegrityPollSingleMultiHeaderExchange asserts the MEXT-015 primary path:
// IntegrityPoll issues ONE Class-0 multi-group Read (a single request carrying
// G1/G20/G30 all-objects headers) and the full MVP set is returned from that
// single exchange — not three per-group reads. The simulator answers a
// multi-group read with one multi-object-header response, so exactly one
// application request frame should be sent for the poll itself.
func TestIntegrityPollSingleMultiHeaderExchange(t *testing.T) {
	sim := testutils.NewMVPOutstationSimulator(1024, 0xFFFF)
	sim.SetBinaryInputs([]*types.BinaryInput{
		{Index: 0, Value: true, Quality: types.QualityOnline},
		{Index: 1, Value: false, Quality: types.QualityOnline},
	})
	sim.SetAnalogInputs([]*types.AnalogInput{
		{Index: 0, Value: 42, Quality: types.QualityOnline},
	})
	sim.SetCounters([]*types.Counter{
		{Index: 0, Value: 100, Quality: types.QualityOnline},
		{Index: 1, Value: 7, Quality: types.QualityOnline},
	})

	client, err := NewClientWithTransport(
		NewConfig(
			WithOutstationAddress(1024),
			WithTimeout(2*time.Second),
			WithRetry(1, 0),
		),
		sim,
	)
	if err != nil {
		t.Fatalf("NewClientWithTransport: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	defer client.Close()

	// Baseline frame count after the link handshake (Connect).
	baseline := len(sim.SentFrames())

	poll, err := client.IntegrityPoll(ctx)
	if err != nil {
		t.Fatalf("IntegrityPoll: %v", err)
	}
	if poll == nil {
		t.Fatal("IntegrityPoll returned nil")
	}

	// Full MVP set must be populated from the single exchange.
	if len(poll.BinaryInputs) != 2 {
		t.Fatalf("BinaryInputs = %d, want 2", len(poll.BinaryInputs))
	}
	if len(poll.AnalogInputs) != 1 {
		t.Fatalf("AnalogInputs = %d, want 1", len(poll.AnalogInputs))
	}
	if len(poll.Counters) != 2 {
		t.Fatalf("Counters = %d, want 2", len(poll.Counters))
	}
	if len(poll.AnalogInputs) == 1 && poll.AnalogInputs[0].Value != 42 {
		t.Fatalf("AnalogInputs[0].Value = %v, want 42", poll.AnalogInputs[0].Value)
	}
	if len(poll.Counters) == 2 && poll.Counters[0].Value != 100 {
		t.Fatalf("Counters[0].Value = %v, want 100", poll.Counters[0].Value)
	}

	// Exactly one application request frame (ConfirmedUserData) should have been
	// sent for the poll — the single multi-group Class-0 read, not three.
	sent := sim.SentFrames()
	var pollRequests int
	for _, f := range sent[baseline:] {
		if f.Control.PRM && f.Control.FuncCode == frame.FuncConfirmedUserData {
			pollRequests++
		}
	}
	if pollRequests != 1 {
		t.Fatalf("IntegrityPoll sent %d application request frames, want 1 (single multi-header exchange)", pollRequests)
	}
}

// TestIntegrityPollFallbackPerGroup asserts the MEXT-015 fallback: when the
// primary single multi-group Class-0 read errors, IntegrityPoll falls back to
// per-group reads and still returns the full set. The failingMultiHeaderTransport
// fails the multi-group primary (3 object headers) with a sequence mismatch —
// retried then surfaced as an error — and answers each single-group fallback
// read with a valid multi-header response.
func TestIntegrityPollFallbackPerGroup(t *testing.T) {
	cc := newConnectedClientWithTransport(t, &failingMultiHeaderTransport{})
	poll, err := cc.IntegrityPoll(context.Background())
	if err != nil {
		t.Fatalf("IntegrityPoll fallback: %v", err)
	}
	if poll == nil {
		t.Fatal("IntegrityPoll returned nil")
	}
	if len(poll.BinaryInputs) != 2 {
		t.Fatalf("BinaryInputs = %d, want 2", len(poll.BinaryInputs))
	}
	if len(poll.AnalogInputs) != 1 {
		t.Fatalf("AnalogInputs = %d, want 1", len(poll.AnalogInputs))
	}
	if len(poll.Counters) != 1 {
		t.Fatalf("Counters = %d, want 1", len(poll.Counters))
	}
}

// failingMultiHeaderTransport fails multi-group (>=3 header) read requests with
// a sequence-mismatched response so the primary IntegrityPoll path errors and
// the per-group fallback is exercised. Single-group reads get a valid
// multi-header response carrying G1+G20+G30 (the parsers populate the requested
// group; the fallback merges only the group it asked for).
type failingMultiHeaderTransport struct {
	lastSeq  uint8
	multiReq bool
}

func (t *failingMultiHeaderTransport) Send(data []byte) error {
	t.lastSeq = extractPubRequestSeq(data)
	t.multiReq = pubRequestHeaderCount(data) >= 3
	return nil
}

func (t *failingMultiHeaderTransport) SetTimeout(ms int) {}

func (t *failingMultiHeaderTransport) Receive() ([]byte, error) {
	if t.multiReq {
		// Wrong-seq response → ErrResponseSeqMismatch (retried, then error).
		return buildMultiHeaderResponse(t.lastSeq ^ 0x0F), nil
	}
	return buildMultiHeaderResponse(t.lastSeq), nil
}

// pubRequestHeaderCount returns the number of object headers in a Read request
// APDU. Each all-objects (0x06) header is 4 octets following the 2-octet APDU
// control+func prefix.
func pubRequestHeaderCount(raw []byte) int {
	f, err := frame.Decode(raw)
	if err != nil {
		return 0
	}
	frag, err := tl.DecodeFragment(f.Data)
	if err != nil {
		return 0
	}
	apdu, err := al.Decode(frag.Data)
	if err != nil {
		return 0
	}
	if len(apdu.Data) == 0 {
		return 0
	}
	return len(apdu.Data) / 4
}
