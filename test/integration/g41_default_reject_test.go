package integration

import (
	"context"
	"testing"
	"time"

	"dnp3/pkg/dnp3"
	"dnp3/pkg/dnp3/master"
	"dnp3/pkg/dnp3/outstation"
	"dnp3/pkg/dnp3/types"
)

// AO-005 — the published v0 supported profile rejects Group 41 analog controls.
// This end-to-end gate asserts that a *default* outstation (built without the
// explicit, non-default WithExperimentalAnalogControl switch) fails closed on a
// Group 41 Direct Operate: the host analog handler is never invoked, the command
// is never acknowledged as ControlSuccess, and the master surfaces a fail-closed
// outcome. The experimental master is used so the request reaches the wire; the
// refusal must come from the outstation, not the master gate.

// g41DefaultRejectHandler records any analog command the outstation bridge would
// dispatch. It must stay empty under the default (rejecting) profile.
type g41DefaultRejectHandler struct {
	received []*types.ControlOutput
}

func (h *g41DefaultRejectHandler) HandleBinaryCommand(cmd *types.ControlOutput) (*types.ControlStatus, error) {
	s := types.ControlSuccess
	return &s, nil
}

func (h *g41DefaultRejectHandler) HandleAnalogCommand(cmd *types.ControlOutput) (*types.ControlStatus, error) {
	h.received = append(h.received, cmd)
	s := types.ControlSuccess
	return &s, nil
}

func TestOperateG41DefaultOutstationRejectsEndToEnd(t *testing.T) {
	port := getFreePort(t)

	// Default outstation: no WithExperimentalAnalogControl.
	server, err := outstation.NewServer(outstation.NewConfig(
		outstation.WithAddress(1024),
		outstation.WithMasterAddress(0xFFFF),
		outstation.WithTransport(dnp3.TCP, "localhost", port),
	))
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	handler := &g41DefaultRejectHandler{}
	server.SetDataHandler(&recordingDataHandler{})
	server.SetCommandHandler(handler)

	ctx, cancel := context.WithCancel(context.Background())
	if err := server.Start(ctx); err != nil {
		cancel()
		t.Fatalf("server.Start: %v", err)
	}
	t.Cleanup(func() {
		_ = server.Stop(context.Background())
		cancel()
	})

	time.Sleep(100 * time.Millisecond)

	// Experimental master so the Group 41 request is admitted and transmitted.
	client, err := master.NewClient(master.NewConfig(
		master.WithMasterAddress(0xFFFF),
		master.WithOutstationAddress(1024),
		master.WithTransport(dnp3.TCP, "localhost", port),
		master.WithExperimentalAnalogControl(),
		master.WithTimeout(5*time.Second),
		master.WithRetry(1, 0),
	))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Connect: %v", err)
	}

	resp, err := client.Operate(ctx, types.NewAnalogInt16Control(0, 42, types.DirectOperate))
	if err == nil && resp != nil && resp.Status == types.ControlSuccess {
		t.Fatalf("default outstation acknowledged a Group 41 analog control as ControlSuccess")
	}
	if len(handler.received) != 0 {
		t.Fatalf("default outstation dispatched %d analog commands to the host handler, want 0", len(handler.received))
	}
}
