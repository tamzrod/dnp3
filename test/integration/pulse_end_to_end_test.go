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

// pulseCapturingHandler records the complete public CROB and returns success.
type pulseCapturingHandler struct {
	got *types.ControlOutput
}

func (h *pulseCapturingHandler) HandleBinaryCommand(cmd *types.ControlOutput) (*types.ControlStatus, error) {
	copyCmd := *cmd
	h.got = &copyCmd
	s := types.ControlSuccess
	return &s, nil
}
func (h *pulseCapturingHandler) HandleAnalogCommand(cmd *types.ControlOutput) (*types.ControlStatus, error) {
	s := types.ControlNotSupported
	return &s, nil
}

// TestPulseCROBEndToEndPreservesTiming proves the full operation type survives
// the master encoder, the G12V1 wire, the outstation decoder and the public
// command-handler bridge. The previous G12V3 NewPulseControl assumption and
// boolean-only bridge both collapsed this information.
func TestPulseCROBEndToEndPreservesTiming(t *testing.T) {
	port := getFreePort(t)
	handler := &pulseCapturingHandler{}
	server, err := outstation.NewServer(outstation.NewConfig(
		outstation.WithAddress(1024),
		outstation.WithMasterAddress(0xFFFF),
		outstation.WithTransport(dnp3.TCP, "localhost", port),
	))
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	server.SetDataHandler(&recordingDataHandler{})
	server.SetCommandHandler(handler)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := server.Start(ctx); err != nil {
		t.Fatalf("server.Start: %v", err)
	}
	defer server.Stop(context.Background())
	time.Sleep(150 * time.Millisecond)

	client, err := master.NewClient(master.NewConfig(
		master.WithMasterAddress(0xFFFF),
		master.WithOutstationAddress(1024),
		master.WithTransport(dnp3.TCP, "localhost", port),
		master.WithTimeout(5*time.Second),
		master.WithRetry(1, 0),
	))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	defer client.Close()
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Connect: %v", err)
	}

	resp, err := client.Operate(ctx, types.NewPulseControl(0, 1000, 500, types.DirectOperate))
	if err != nil {
		t.Fatalf("Operate: %v", err)
	}
	if resp.Status != types.ControlSuccess {
		t.Fatalf("pulse status = %v, want success (IIN=%v)", resp.Status, resp.IIN)
	}
	if handler.got == nil {
		t.Fatal("command handler was not invoked")
	}
	c := handler.got.CROB
	if c == nil {
		t.Fatal("handler received no full CROB carrier")
	}
	if c.Code != types.CROBOpPulseOn || c.Count != 1 || c.OnTime != 1000 || c.OffTime != 500 {
		t.Fatalf("handler CROB = %+v, want PulseOn/1/1000/500", c)
	}
	if handler.got.Group != 12 || handler.got.Variation != 1 {
		t.Fatalf("handler object = G%dV%d, want G12V1", handler.got.Group, handler.got.Variation)
	}
}

// TestNULCROBIsEndToEndNoOp proves a NUL code never reaches a command handler
// and is acknowledged, so it cannot be misread as Latch Off.
func TestNULCROBIsEndToEndNoOp(t *testing.T) {
	port := getFreePort(t)
	handler := &pulseCapturingHandler{}
	server, err := outstation.NewServer(outstation.NewConfig(
		outstation.WithAddress(1024),
		outstation.WithMasterAddress(0xFFFF),
		outstation.WithTransport(dnp3.TCP, "localhost", port),
	))
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	server.SetDataHandler(&recordingDataHandler{})
	server.SetCommandHandler(handler)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := server.Start(ctx); err != nil {
		t.Fatalf("server.Start: %v", err)
	}
	defer server.Stop(context.Background())
	time.Sleep(150 * time.Millisecond)

	client, err := master.NewClient(master.NewConfig(
		master.WithMasterAddress(0xFFFF),
		master.WithOutstationAddress(1024),
		master.WithTransport(dnp3.TCP, "localhost", port),
		master.WithTimeout(5*time.Second),
		master.WithRetry(1, 0),
	))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	defer client.Close()
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Connect: %v", err)
	}

	nul := types.NewCROBControl(0, types.CROBOpNUL, 1, 0, 0, types.DirectOperate)
	resp, err := client.Operate(ctx, nul)
	if err != nil {
		t.Fatalf("Operate(NUL): %v", err)
	}
	if resp.Status != types.ControlSuccess {
		t.Fatalf("NUL status = %v, want success (IIN=%v)", resp.Status, resp.IIN)
	}
	if handler.got != nil {
		t.Fatal("NUL CROB reached the command handler")
	}
}
