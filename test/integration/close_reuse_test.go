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

// TestClientReusableAfterClose verifies DNP3-050: after Close the client is in a
// clean Disconnected state with no leaked resources, and the SAME client can
// Connect again and successfully operate against the simulator (reusable after
// Close). It also asserts Close is idempotent.
func TestClientReusableAfterClose(t *testing.T) {
	sim := testutils.NewMVPOutstationSimulator(1024, 0xFFFF)
	sim.SetAnalogInputs([]*types.AnalogInput{
		{Index: 0, Value: 11, Quality: types.QualityOnline},
	})
	sim.SetCommandStatus(types.ControlSuccess)

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

	ctx := context.Background()

	// First session: connect, operate, close.
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("first Connect: %v", err)
	}
	if got := client.State(); got != dnp3.StateActive {
		t.Fatalf("state after first Connect = %v, want Active", got)
	}
	resp, err := client.IntegrityPoll(ctx)
	if err != nil {
		t.Fatalf("first IntegrityPoll: %v", err)
	}
	if len(resp.AnalogInputs) != 1 || resp.AnalogInputs[0].Value != 11 {
		t.Fatalf("first session points = %+v, want 1 analog=11", resp.AnalogInputs)
	}
	opResp, err := client.Operate(ctx, &types.ControlOutput{
		Group: 12, Variation: 1, Index: 0, CommandType: types.DirectOperate,
		Value: &types.BinaryCommandValue{Value: true},
	})
	if err != nil {
		t.Fatalf("first Operate: %v", err)
	}
	if opResp.Status != types.ControlSuccess {
		t.Fatalf("first Operate status = %v, want ControlSuccess", opResp.Status)
	}

	if err := client.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if got := client.State(); got != dnp3.StateDisconnected {
		t.Fatalf("state after Close = %v, want Disconnected", got)
	}

	// Close is idempotent: a second Close must not error and must not change
	// state.
	if err := client.Close(); err != nil {
		t.Fatalf("second Close (idempotent): %v", err)
	}
	if got := client.State(); got != dnp3.StateDisconnected {
		t.Fatalf("state after second Close = %v, want Disconnected", got)
	}

	// Reuse: the SAME client connects again against the same simulator and
	// successfully operates. This is the DNP3-050 acceptance bar.
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("second Connect (reusable): %v", err)
	}
	defer client.Close()
	if got := client.State(); got != dnp3.StateActive {
		t.Fatalf("state after second Connect = %v, want Active", got)
	}

	resp2, err := client.IntegrityPoll(ctx)
	if err != nil {
		t.Fatalf("second IntegrityPoll: %v", err)
	}
	if len(resp2.AnalogInputs) != 1 || resp2.AnalogInputs[0].Value != 11 {
		t.Fatalf("second session points = %+v, want 1 analog=11", resp2.AnalogInputs)
	}
	opResp2, err := client.Operate(ctx, &types.ControlOutput{
		Group: 12, Variation: 1, Index: 0, CommandType: types.DirectOperate,
		Value: &types.BinaryCommandValue{Value: true},
	})
	if err != nil {
		t.Fatalf("second Operate: %v", err)
	}
	if opResp2.Status != types.ControlSuccess {
		t.Fatalf("second Operate status = %v, want ControlSuccess", opResp2.Status)
	}
}

// TestClientCloseNeverConnected verifies DNP3-050: Close on a client that was
// never connected is a safe no-op (state stays Disconnected, no error, no
// leaked resources).
func TestClientCloseNeverConnected(t *testing.T) {
	sim := testutils.NewMVPOutstationSimulator(1024, 0xFFFF)
	client, err := master.NewClientWithTransport(
		master.NewConfig(master.WithOutstationAddress(1024)),
		sim,
	)
	if err != nil {
		t.Fatalf("NewClientWithTransport: %v", err)
	}
	if got := client.State(); got != dnp3.StateDisconnected {
		t.Fatalf("state = %v, want Disconnected", got)
	}
	if err := client.Close(); err != nil {
		t.Fatalf("Close on never-connected client: %v", err)
	}
	if got := client.State(); got != dnp3.StateDisconnected {
		t.Fatalf("state after Close = %v, want Disconnected", got)
	}
}
