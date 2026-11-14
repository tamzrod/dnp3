package integration

import (
	"context"
	"math"
	"testing"
	"time"

	"dnp3/pkg/dnp3"
	"dnp3/pkg/dnp3/master"
	"dnp3/pkg/dnp3/outstation"
	"dnp3/pkg/dnp3/types"
)

// AO-003 — Group 41 Variation 3 (single-precision IEEE-754 float) Direct Operate
// over real TCP loopback.
//
// This is a necessary-but-not-sufficient gate (the in-repo outstation and master
// share the encoder/decoder, so a mutually-consistent bug could pass it; AO-001
// D2). Independent peer verification is AO-004 scope. It proves the G41V3 path
// is wired end-to-end through the public API and the real transport, that the
// echoed float32 reaches the host handler with its exact bits, and that a
// non-success echo is never surfaced as ControlSuccess.

// g41v3AcceptHandler accepts every analog command with ControlSuccess and records
// the typed value the outstation bridge delivered.
type g41v3AcceptHandler struct {
	received []*types.ControlOutput
}

func (h *g41v3AcceptHandler) HandleBinaryCommand(cmd *types.ControlOutput) (*types.ControlStatus, error) {
	s := types.ControlSuccess
	return &s, nil
}

func (h *g41v3AcceptHandler) HandleAnalogCommand(cmd *types.ControlOutput) (*types.ControlStatus, error) {
	h.received = append(h.received, cmd)
	s := types.ControlSuccess
	return &s, nil
}

// g41v3RejectHandler rejects every analog command (non-success status, no error),
// which the bridge maps to a failure response the master must not read as success.
type g41v3RejectHandler struct{}

func (h *g41v3RejectHandler) HandleBinaryCommand(cmd *types.ControlOutput) (*types.ControlStatus, error) {
	s := types.ControlSuccess
	return &s, nil
}

func (h *g41v3RejectHandler) HandleAnalogCommand(cmd *types.ControlOutput) (*types.ControlStatus, error) {
	s := types.ControlNotSupported
	return &s, nil
}

func newG41V3Loopback(t *testing.T, handler outstation.CommandHandler) (master.Client, context.Context, context.CancelFunc) {
	t.Helper()
	port := getFreePort(t)

	server, err := outstation.NewServer(outstation.NewConfig(
		outstation.WithAddress(1024),
		outstation.WithMasterAddress(0xFFFF),
		outstation.WithTransport(dnp3.TCP, "localhost", port),
		outstation.WithExperimentalAnalogControl(),
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
	return client, ctx, cancel
}

// TestOperateG41V3RealTCPSuccess drives a G41V3 DirectOperate (float32 -0.5) over
// real TCP and asserts ControlSuccess plus exact dispatch of the typed value.
func TestOperateG41V3RealTCPSuccess(t *testing.T) {
	handler := &g41v3AcceptHandler{}
	client, ctx, _ := newG41V3Loopback(t, handler)

	resp, err := client.Operate(ctx, types.NewAnalogFloat32Control(0, -0.5, types.DirectOperate))
	if err != nil {
		t.Fatalf("Operate G41V3 over real TCP failed: %v", err)
	}
	if resp.Status != types.ControlSuccess {
		t.Fatalf("Operate status = %v, want ControlSuccess", resp.Status)
	}
	if len(handler.received) != 1 {
		t.Fatalf("handler received %d commands, want 1", len(handler.received))
	}
	got := handler.received[0]
	if got.Group != 41 || got.Variation != 3 || got.Index != 0 {
		t.Errorf("dispatched command = group %d var %d index %d, want 41/3/0", got.Group, got.Variation, got.Index)
	}
	if got.CommandType != types.DirectOperate {
		t.Errorf("dispatched CommandType = %v, want DirectOperate", got.CommandType)
	}
	v, ok := got.Value.(*types.AnalogFloat32CommandValue)
	if !ok {
		t.Fatalf("dispatched value type = %T, want *types.AnalogFloat32CommandValue", got.Value)
	}
	if math.Float32bits(v.Value) != math.Float32bits(-0.5) {
		t.Errorf("dispatched value bits = %08X, want %08X", math.Float32bits(v.Value), math.Float32bits(-0.5))
	}
}

// TestOperateG41V3RealTCPSignedZero asserts a -0.0 setpoint round-trips over real
// TCP with its sign bit intact and surfaces ControlSuccess (the master compares
// the echoed value by raw IEEE-754 bits, not numeric equality).
func TestOperateG41V3RealTCPSignedZero(t *testing.T) {
	handler := &g41v3AcceptHandler{}
	client, ctx, _ := newG41V3Loopback(t, handler)

	negZero := float32(math.Copysign(0, -1))
	resp, err := client.Operate(ctx, types.NewAnalogFloat32Control(0, negZero, types.DirectOperate))
	if err != nil {
		t.Fatalf("Operate G41V3 (-0.0) over real TCP failed: %v", err)
	}
	if resp.Status != types.ControlSuccess {
		t.Fatalf("Operate status = %v, want ControlSuccess", resp.Status)
	}
	if len(handler.received) != 1 {
		t.Fatalf("handler received %d commands, want 1", len(handler.received))
	}
	v, ok := handler.received[0].Value.(*types.AnalogFloat32CommandValue)
	if !ok {
		t.Fatalf("dispatched value type = %T, want *types.AnalogFloat32CommandValue", handler.received[0].Value)
	}
	if math.Float32bits(v.Value) != math.Float32bits(negZero) {
		t.Errorf("dispatched value bits = %08X, want %08X (sign bit lost)", math.Float32bits(v.Value), math.Float32bits(negZero))
	}
}

// TestOperateG41V3RealTCPNonSuccess asserts a rejected G41V3 command is never
// surfaced as ControlSuccess over real TCP. The outstation's handler refusal
// returns an error, so the response carries error IIN and no G41 echo; the G41
// contract deliberately has no IIN-only success fallback, so the master reports
// a fail-closed outcome (an error, never ControlSuccess).
func TestOperateG41V3RealTCPNonSuccess(t *testing.T) {
	client, ctx, _ := newG41V3Loopback(t, &g41v3RejectHandler{})

	resp, err := client.Operate(ctx, types.NewAnalogFloat32Control(0, 1.0, types.DirectOperate))
	if err != nil {
		return
	}
	if resp.Status == types.ControlSuccess {
		t.Fatalf("Operate status = ControlSuccess, want a failure status for a rejected command")
	}
}

// TestOperateG41V3RealTCPNonFiniteRejectedBeforeSend asserts a NaN setpoint is
// rejected by the public client before any wire traffic: no handler call occurs
// and the outcome is never ControlSuccess.
func TestOperateG41V3RealTCPNonFiniteRejectedBeforeSend(t *testing.T) {
	handler := &g41v3AcceptHandler{}
	client, ctx, _ := newG41V3Loopback(t, handler)

	resp, err := client.Operate(ctx, types.NewAnalogFloat32Control(0, float32(math.NaN()), types.DirectOperate))
	if err == nil {
		t.Fatalf("expected NaN to be rejected before sending, got resp=%v", resp)
	}
	if resp != nil && resp.Status == types.ControlSuccess {
		t.Fatalf("rejected NaN setpoint surfaced as ControlSuccess")
	}
	if len(handler.received) != 0 {
		t.Fatalf("handler received %d commands for a NaN setpoint, want 0", len(handler.received))
	}
}
