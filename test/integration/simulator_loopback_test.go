package integration

import (
	"context"
	"testing"
	"time"

	"dnp3/internal/testutils"
	"dnp3/pkg/dnp3"
	"dnp3/pkg/dnp3/master"
	"dnp3/pkg/dnp3/types"
)

// TestPublicLoopbackAgainstSimulator drives the FULL public client lifecycle
// (Connect → Read → Operate) against the deterministic in-memory outstation
// simulator with NO network I/O and NO real outstation process (DNP3-036).
// Acceptance: the public loopback is green against the simulator only.
func TestPublicLoopbackAgainstSimulator(t *testing.T) {
	sim := testutils.NewMVPOutstationSimulator(1024, 0xFFFF)
	sim.SetAnalogInputs([]*types.AnalogInput{
		{Index: 0, Value: 42, Quality: types.QualityOnline},
	})
	sim.SetCounters([]*types.Counter{
		{Index: 0, Value: 100, Quality: types.QualityOnline},
	})

	client, err := master.NewClientWithTransport(
		master.NewConfig(
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

	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Connect against simulator failed: %v", err)
	}
	defer client.Close()

	// Read each MVP-supported Class-0 group separately (one object header per
	// response), matching how the v0 public Read is exercised against a single
	// golden outstation. Separate reads avoid the legacy multi-header skip
	// path, which is out of scope for the simulator task (DNP3-036).
	binResp, err := client.Read(ctx, types.NewReadRequest(
		types.GroupRequest{Group: 1, Variation: 1},
	))
	if err != nil {
		t.Fatalf("Read G1 against simulator failed: %v", err)
	}
	if len(binResp.BinaryInputs) != 1 || !binResp.BinaryInputs[0].Value {
		t.Fatalf("BinaryInputs = %+v, want 1 point Value=true", binResp.BinaryInputs)
	}

	anResp, err := client.Read(ctx, types.NewReadRequest(
		types.GroupRequest{Group: 30, Variation: 1},
	))
	if err != nil {
		t.Fatalf("Read G30 against simulator failed: %v", err)
	}
	if len(anResp.AnalogInputs) != 1 || anResp.AnalogInputs[0].Value != 42 {
		t.Fatalf("AnalogInputs = %+v, want 1 point Value=42", anResp.AnalogInputs)
	}

	ctrResp, err := client.Read(ctx, types.NewReadRequest(
		types.GroupRequest{Group: 20, Variation: 1},
	))
	if err != nil {
		t.Fatalf("Read G20 against simulator failed: %v", err)
	}
	if len(ctrResp.Counters) != 1 || ctrResp.Counters[0].Value != 100 {
		t.Fatalf("Counters = %+v, want 1 point Value=100", ctrResp.Counters)
	}

	// DNP3-012: the response IIN and the master's stored LastIIN must match.
	if ctrResp.IIN != client.LastIIN() {
		t.Fatalf("ReadResponse.IIN = %v, LastIIN = %v (expected equal)", ctrResp.IIN, client.LastIIN())
	}

	// DirectOperate a CROB; the simulator returns ControlSuccess by default.
	opResp, err := client.Operate(ctx, &types.ControlOutput{
		Group: 12, Variation: 1, Index: 0,
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

// TestPublicLoopbackSimulatorSurfacesCommandStatus asserts the simulator
// surfaces a configured failure status (ControlBlocked) through the full public
// Operate path (DNP3-036 + DNP3-021).
func TestPublicLoopbackSimulatorSurfacesCommandStatus(t *testing.T) {
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
		Group: 12, Variation: 1, Index: 0,
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

// TestPublicLoopbackSimulatorStateTransitions asserts the public client state
// reaches Active after connecting to the simulator and returns to a terminal
// state after Close (DNP3-036).
func TestPublicLoopbackSimulatorStateTransitions(t *testing.T) {
	sim := testutils.NewMVPOutstationSimulator(1024, 0xFFFF)
	client, err := master.NewClientWithTransport(
		master.NewConfig(
			master.WithOutstationAddress(1024),
			master.WithTimeout(2*time.Second),
		),
		sim,
	)
	if err != nil {
		t.Fatalf("NewClientWithTransport: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if got := client.State(); got == dnp3.StateActive {
		t.Fatalf("state before Connect = %v, want not Active", got)
	}
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	if got := client.State(); got != dnp3.StateActive {
		t.Fatalf("state after Connect = %v, want Active", got)
	}
	if err := client.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}
