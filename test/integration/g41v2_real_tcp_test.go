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

// AO-003 — Group 41 Variation 2 (int16) Direct Operate over real TCP loopback.
//
// This is a necessary-but-not-sufficient gate (the in-repo outstation and master
// share the encoder/decoder, so a mutually-consistent bug could pass it; AO-001
// D2). Independent peer verification is AO-004 scope. It proves the G41V2 path
// is wired end-to-end through the public API and the real transport, that the
// echoed int16 reaches the host handler unchanged, and that a non-success echo
// is never surfaced as ControlSuccess.

// g41v2AcceptHandler accepts every analog command with ControlSuccess and records
// the typed value the outstation bridge delivered.
type g41v2AcceptHandler struct {
	received []*types.ControlOutput
}

func (h *g41v2AcceptHandler) HandleBinaryCommand(cmd *types.ControlOutput) (*types.ControlStatus, error) {
	s := types.ControlSuccess
	return &s, nil
}

func (h *g41v2AcceptHandler) HandleAnalogCommand(cmd *types.ControlOutput) (*types.ControlStatus, error) {
	h.received = append(h.received, cmd)
	s := types.ControlSuccess
	return &s, nil
}

// g41v2RejectHandler rejects every analog command (non-success status, no error),
// which the bridge maps to a failure response the master must not read as success.
type g41v2RejectHandler struct{}

func (h *g41v2RejectHandler) HandleBinaryCommand(cmd *types.ControlOutput) (*types.ControlStatus, error) {
	s := types.ControlSuccess
	return &s, nil
}

func (h *g41v2RejectHandler) HandleAnalogCommand(cmd *types.ControlOutput) (*types.ControlStatus, error) {
	s := types.ControlNotSupported
	return &s, nil
}

func newG41V2Loopback(t *testing.T, handler outstation.CommandHandler) (master.Client, context.Context, context.CancelFunc) {
	t.Helper()
	port := getFreePort(t)

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
	if err := server.Start(ctx); err != nil {
		cancel()
		t.Fatalf("server.Start: %v", err)
	}
	t.Cleanup(func() {
		_ = server.Stop(context.Background())
		cancel()
	})

	time.Sleep(100 * time.Millisecond)

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
	t.Cleanup(func() { _ = client.Close() })

	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	return client, ctx, cancel
}

// TestOperateG41V2RealTCPSuccess drives a G41V2 DirectOperate (int16 -42) over
// real TCP and asserts ControlSuccess plus exact dispatch of the typed value.
func TestOperateG41V2RealTCPSuccess(t *testing.T) {
	handler := &g41v2AcceptHandler{}
	client, ctx, _ := newG41V2Loopback(t, handler)

	resp, err := client.Operate(ctx, types.NewAnalogInt16Control(0, -42, types.DirectOperate))
	if err != nil {
		t.Fatalf("Operate G41V2 over real TCP failed: %v", err)
	}
	if resp.Status != types.ControlSuccess {
		t.Fatalf("Operate status = %v, want ControlSuccess", resp.Status)
	}
	if len(handler.received) != 1 {
		t.Fatalf("handler received %d commands, want 1", len(handler.received))
	}
	got := handler.received[0]
	if got.Group != 41 || got.Variation != 2 || got.Index != 0 {
		t.Errorf("dispatched command = group %d var %d index %d, want 41/2/0", got.Group, got.Variation, got.Index)
	}
	if got.CommandType != types.DirectOperate {
		t.Errorf("dispatched CommandType = %v, want DirectOperate", got.CommandType)
	}
	v, ok := got.Value.(*types.AnalogInt16CommandValue)
	if !ok {
		t.Fatalf("dispatched value type = %T, want *types.AnalogInt16CommandValue", got.Value)
	}
	if v.Value != -42 {
		t.Errorf("dispatched value = %d, want -42", v.Value)
	}
}

// TestOperateG41V2RealTCPNonSuccess asserts a rejected G41V2 command is never
// surfaced as ControlSuccess over real TCP. The outstation's handler refusal
// returns an error, so the response carries error IIN and no G41 echo; the G41
// contract deliberately has no IIN-only success fallback, so the master reports
// a fail-closed outcome (an error, never ControlSuccess).
func TestOperateG41V2RealTCPNonSuccess(t *testing.T) {
	client, ctx, _ := newG41V2Loopback(t, &g41v2RejectHandler{})

	resp, err := client.Operate(ctx, types.NewAnalogInt16Control(0, 42, types.DirectOperate))
	if err != nil {
		// Fail-closed: a rejected command with no conformant G41 echo is an error,
		// which is never a success.
		return
	}
	if resp.Status == types.ControlSuccess {
		t.Fatalf("Operate status = ControlSuccess, want a failure status for a rejected command")
	}
}
