package master

import (
	"context"
	"testing"
	"time"

	"dnp3/internal/testutils"
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
