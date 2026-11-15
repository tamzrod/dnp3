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

// AO-003 — Group 41 Variation 4 (double-precision IEEE-754 float) Direct Operate
// over real TCP loopback.
//
// This is a necessary-but-not-sufficient gate (the in-repo outstation and master
// share the encoder/decoder, so a mutually-consistent bug could pass it; AO-001
// D2). Independent peer verification is AO-004 scope. It proves the G41V4 path
// is wired end-to-end through the public API and the real transport, that the
// echoed float64 reaches the host handler with its exact bits, and that a
// non-success echo is never surfaced as ControlSuccess.

// g41v4AcceptHandler accepts every analog command with ControlSuccess and records
// the typed value the outstation bridge delivered.
type g41v4AcceptHandler struct {
	received []*types.ControlOutput
}

func (h *g41v4AcceptHandler) HandleBinaryCommand(cmd *types.ControlOutput) (*types.ControlStatus, error) {
	s := types.ControlSuccess
	return &s, nil
}

func (h *g41v4AcceptHandler) HandleAnalogCommand(cmd *types.ControlOutput) (*types.ControlStatus, error) {
	h.received = append(h.received, cmd)
	s := types.ControlSuccess
	return &s, nil
}

// g41v4RejectHandler rejects every analog command (non-success status, no error),
// which the bridge maps to a failure response the master must not read as success.
type g41v4RejectHandler struct{}

func (h *g41v4RejectHandler) HandleBinaryCommand(cmd *types.ControlOutput) (*types.ControlStatus, error) {
	s := types.ControlSuccess
	return &s, nil
}

func (h *g41v4RejectHandler) HandleAnalogCommand(cmd *types.ControlOutput) (*types.ControlStatus, error) {
	s := types.ControlNotSupported
	return &s, nil
}

func newG41V4Loopback(t *testing.T, handler outstation.CommandHandler) (master.Client, context.Context, context.CancelFunc) {
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

// TestOperateG41V4RealTCPSuccess drives a G41V4 DirectOperate (float64 123.456789)
// over real TCP and asserts ControlSuccess plus exact dispatch of the typed value.
func TestOperateG41V4RealTCPSuccess(t *testing.T) {
	handler := &g41v4AcceptHandler{}
	client, ctx, _ := newG41V4Loopback(t, handler)

	resp, err := client.Operate(ctx, types.NewAnalogFloat64Control(0, 123.456789, types.DirectOperate))
	if err != nil {
		t.Fatalf("Operate G41V4 over real TCP failed: %v", err)
	}
	if resp.Status != types.ControlSuccess {
		t.Fatalf("Operate status = %v, want ControlSuccess", resp.Status)
	}
	if len(handler.received) != 1 {
		t.Fatalf("handler received %d commands, want 1", len(handler.received))
	}
	got := handler.received[0]
	if got.Group != 41 || got.Variation != 4 || got.Index != 0 {
		t.Errorf("dispatched command = group %d var %d index %d, want 41/4/0", got.Group, got.Variation, got.Index)
	}
	if got.CommandType != types.DirectOperate {
		t.Errorf("dispatched CommandType = %v, want DirectOperate", got.CommandType)
	}
	v, ok := got.Value.(*types.AnalogFloat64CommandValue)
	if !ok {
		t.Fatalf("dispatched value type = %T, want *types.AnalogFloat64CommandValue", got.Value)
	}
	if math.Float64bits(v.Value) != math.Float64bits(123.456789) {
		t.Errorf("dispatched value bits = %016X, want %016X", math.Float64bits(v.Value), math.Float64bits(123.456789))
	}
}

// TestOperateG41V4RealTCPSignedZero asserts a -0.0 setpoint round-trips over real
// TCP with its sign bit intact and surfaces ControlSuccess (the master compares
// the echoed value by raw IEEE-754 bits, not numeric equality).
func TestOperateG41V4RealTCPSignedZero(t *testing.T) {
	handler := &g41v4AcceptHandler{}
	client, ctx, _ := newG41V4Loopback(t, handler)

	negZero := math.Copysign(0, -1)
	resp, err := client.Operate(ctx, types.NewAnalogFloat64Control(0, negZero, types.DirectOperate))
	if err != nil {
		t.Fatalf("Operate G41V4 (-0.0) over real TCP failed: %v", err)
	}
	if resp.Status != types.ControlSuccess {
		t.Fatalf("Operate status = %v, want ControlSuccess", resp.Status)
	}
	if len(handler.received) != 1 {
		t.Fatalf("handler received %d commands, want 1", len(handler.received))
	}
	v, ok := handler.received[0].Value.(*types.AnalogFloat64CommandValue)
	if !ok {
		t.Fatalf("dispatched value type = %T, want *types.AnalogFloat64CommandValue", handler.received[0].Value)
	}
	if math.Float64bits(v.Value) != math.Float64bits(negZero) {
		t.Errorf("dispatched value bits = %016X, want %016X (sign bit lost)", math.Float64bits(v.Value), math.Float64bits(negZero))
	}
}

// TestOperateG41V4RealTCPNonSuccess asserts a rejected G41V4 command is never
// surfaced as ControlSuccess over real TCP. The outstation's handler refusal
// returns an error, so the response carries error IIN and no G41 echo; the G41
// contract deliberately has no IIN-only success fallback, so the master reports
// a fail-closed outcome (an error, never ControlSuccess).
func TestOperateG41V4RealTCPNonSuccess(t *testing.T) {
	client, ctx, _ := newG41V4Loopback(t, &g41v4RejectHandler{})

	resp, err := client.Operate(ctx, types.NewAnalogFloat64Control(0, 1.0, types.DirectOperate))
	if err != nil {
		return
	}
	if resp.Status == types.ControlSuccess {
		t.Fatalf("Operate status = ControlSuccess, want a failure status for a rejected command")
	}
}

// TestOperateG41V4RealTCPNonFiniteRejectedBeforeSend asserts a NaN setpoint is
// rejected by the public client before any wire traffic: no handler call occurs
// and the outcome is never ControlSuccess.
func TestOperateG41V4RealTCPNonFiniteRejectedBeforeSend(t *testing.T) {
	handler := &g41v4AcceptHandler{}
	client, ctx, _ := newG41V4Loopback(t, handler)

	resp, err := client.Operate(ctx, types.NewAnalogFloat64Control(0, math.NaN(), types.DirectOperate))
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
