package integration

import (
	"context"
	"errors"
	"testing"
	"time"

	"dnp3/internal/testutils"
	"dnp3/pkg/dnp3"
	"dnp3/pkg/dnp3/master"
	"dnp3/pkg/dnp3/types"
)

// TestPublicMVPLoopbackFullLifecycle exercises the complete v0 public API
// against the deterministic in-memory outstation simulator in a single
// end-to-end flow (DNP3-045):
//
//	Connect → IntegrityPoll (Class-0) → assert all MVP points → Operate (CROB)
//	         → assert ControlSuccess → Close → terminal state.
//
// Acceptance: the full MVP loopback is green against the simulator only (no
// network I/O, no real outstation process). This consolidates the Connect /
// Integrity / Operate acceptance of DNP3-036/037/021 into one full-MVP test.
func TestPublicMVPLoopbackFullLifecycle(t *testing.T) {
	sim := testutils.NewMVPOutstationSimulator(1024, 0xFFFF)
	sim.SetBinaryInputs([]*types.BinaryInput{
		{Index: 0, Value: true, Quality: types.QualityOnline},
		{Index: 1, Value: false, Quality: types.QualityOnline},
	})
	sim.SetAnalogInputs([]*types.AnalogInput{
		{Index: 0, Value: 42, Quality: types.QualityOnline},
		{Index: 1, Value: -7, Quality: types.QualityOnline},
	})
	sim.SetCounters([]*types.Counter{
		{Index: 0, Value: 100, Quality: types.QualityOnline},
		{Index: 1, Value: 7, Quality: types.QualityOnline},
	})

	client, err := master.NewClientWithTransport(
		master.NewConfig(
			master.WithMasterAddress(0xFFFF),
			master.WithOutstationAddress(1024),
			master.WithTimeout(2*time.Second),
			master.WithRetry(1, 0),
		),
		sim,
	)
	if err != nil {
		t.Fatalf("NewClientWithTransport: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Connect — the public client must reach StateActive.
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Connect against simulator failed: %v", err)
	}
	defer client.Close()
	if got := client.State(); got != dnp3.StateActive {
		t.Fatalf("state after Connect = %v, want Active", got)
	}

	// 2. IntegrityPoll — read all MVP Class-0 static groups in one call.
	resp, err := client.IntegrityPoll(ctx)
	if err != nil {
		t.Fatalf("IntegrityPoll against simulator failed: %v", err)
	}
	if resp == nil {
		t.Fatal("IntegrityPoll returned nil response")
	}

	// Assert all MVP points surfaced through the single merged response.
	if len(resp.BinaryInputs) != 2 {
		t.Fatalf("BinaryInputs = %+v, want 2 points", resp.BinaryInputs)
	}
	if !resp.BinaryInputs[0].Value || resp.BinaryInputs[1].Value {
		t.Fatalf("BinaryInputs values = %+v, want [true, false]", resp.BinaryInputs)
	}
	if len(resp.AnalogInputs) != 2 {
		t.Fatalf("AnalogInputs = %+v, want 2 points", resp.AnalogInputs)
	}
	if resp.AnalogInputs[0].Value != 42 || resp.AnalogInputs[1].Value != -7 {
		t.Fatalf("AnalogInputs values = %+v, want [42, -7]", resp.AnalogInputs)
	}
	if len(resp.Counters) != 2 {
		t.Fatalf("Counters = %+v, want 2 points", resp.Counters)
	}
	if resp.Counters[0].Value != 100 || resp.Counters[1].Value != 7 {
		t.Fatalf("Counters values = %+v, want [100, 7]", resp.Counters)
	}

	// The merged response IIN must match the master's stored LastIIN (DNP3-012).
	if resp.IIN != client.LastIIN() {
		t.Fatalf("IntegrityPoll IIN = %v, LastIIN = %v (expected equal)", resp.IIN, client.LastIIN())
	}

	// 3. Operate — DirectOperate a CROB; the simulator returns ControlSuccess.
	opResp, err := client.Operate(ctx, &types.ControlOutput{
		Group:       12,
		Variation:   1,
		Index:       0,
		CommandType: types.DirectOperate,
		Value:       &types.BinaryCommandValue{Value: true},
	})
	if err != nil {
		t.Fatalf("Operate against simulator failed: %v", err)
	}
	if opResp.Status != types.ControlSuccess {
		t.Fatalf("Operate status = %v, want ControlSuccess", opResp.Status)
	}
}

// TestPublicMVPLoopbackOperateStatus asserts the full-MVP loopback surfaces a
// configured outstation command status through the public Operate path
// (DNP3-045 + DNP3-021): a ControlBlocked outstation must report ControlBlocked.
func TestPublicMVPLoopbackOperateStatus(t *testing.T) {
	sim := testutils.NewMVPOutstationSimulator(1024, 0xFFFF)
	sim.SetCommandStatus(types.ControlBlocked)

	client, err := master.NewClientWithTransport(
		master.NewConfig(
			master.WithOutstationAddress(1024),
			master.WithTimeout(2*time.Second),
			master.WithRetry(0, 0),
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

	resp, err := client.Operate(ctx, &types.ControlOutput{
		Group:       12,
		Variation:   1,
		Index:       0,
		CommandType: types.DirectOperate,
		Value:       &types.BinaryCommandValue{Value: true},
	})
	if err != nil {
		t.Fatalf("Operate failed: %v", err)
	}
	if resp.Status != types.ControlBlocked {
		t.Fatalf("status = %v, want ControlBlocked", resp.Status)
	}
}

// TestPublicMVPLoopbackErrorClassification asserts the full-MVP loopback
// classifies a transport failure through the public error taxonomy (DNP3-045 +
// DNP3-043): a closed peer must surface a classifiable disconnect/protocol
// error, not an opaque string.
func TestPublicMVPLoopbackErrorClassification(t *testing.T) {
	sim := testutils.NewMVPOutstationSimulator(1024, 0xFFFF)
	client, err := master.NewClientWithTransport(
		master.NewConfig(
			master.WithOutstationAddress(1024),
			master.WithTimeout(500*time.Millisecond),
			master.WithRetry(0, 0),
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

	// Tear the simulated peer down so the next Read fails.
	sim.Close()

	_, err = client.Read(ctx, types.NewReadRequest(types.GroupRequest{Group: 1, Variation: 1}))
	if err == nil {
		t.Fatal("expected error after peer close, got nil")
	}
	// The surfaced error must be classifiable (not ErrorCodeUnknown) — the
	// public taxonomy must recognize a transport/protocol failure from a closed
	// peer (DNP3-043).
	if got := dnp3.ClassifyError(err); got == dnp3.ErrorCodeUnknown {
		t.Fatalf("ClassifyError(%v) = unknown, want a classified category", err)
	}
	// A disconnect must also surface the public NotConnected sentinel so callers
	// can detect it without string-matching.
	if !errors.Is(err, dnp3.ErrNotConnected) {
		t.Fatalf("error = %v, want dnp3.ErrNotConnected in chain for a closed peer", err)
	}
}
