package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"dnp3/pkg/dnp3"
	"dnp3/pkg/dnp3/master"
	"dnp3/pkg/dnp3/outstation"
	"dnp3/pkg/dnp3/types"
)

// MEXT-013 — Operate real-TCP vs in-repo outstation.
//
// The in-repo outstation's DirectOperate handler (internal/outstation
// handleDirectOperate) returns an IIN-only response: it carries the outstation's
// IIN bytes and NO Group 12 Variation 1 control-status echo. Before MEXT-012
// this left the master unable to find a parseable G12V1 status byte, so Operate
// reported ControlTimeout against a real outstation (the DNP3-091 discovery
// recorded in outstation_side_gate_test.go). MEXT-012 taught resolveOperateStatus
// to treat an IIN-only response with clear IIN as CommandStatusSuccess.
//
// This test proves that fix on a REAL TCP master↔outstation loopback (not the
// in-memory simulator): Connect → DirectOperate CROB → assert ControlSuccess
// without ControlTimeout, and document the observed response shape.

// operateSuccessHandler is a CommandHandler that accepts every binary command
// with ControlSuccess so the outstation's DirectOperate path returns a clear
// IIN-only response.
type operateSuccessHandler struct {
	received []*types.ControlOutput
}

func (h *operateSuccessHandler) HandleBinaryCommand(cmd *types.ControlOutput) (*types.ControlStatus, error) {
	h.received = append(h.received, cmd)
	s := types.ControlSuccess
	return &s, nil
}

func (h *operateSuccessHandler) HandleAnalogCommand(cmd *types.ControlOutput) (*types.ControlStatus, error) {
	s := types.ControlNotSupported
	return &s, nil
}

// TestOperateRealTCPSuccess asserts the MEXT-012 R1 fix on real TCP: a
// DirectOperate against the in-repo outstation succeeds (ControlSuccess) even
// though the outstation omits the G12V1 control-status echo. The observed
// response shape is IIN-only (no object data); see the MEXT-013 handoff note.
func TestOperateRealTCPSuccess(t *testing.T) {
	port := getFreePort(t)

	data := &recordingDataHandler{
		binaryInputs: []*types.BinaryInput{
			{Index: 0, Value: true, Quality: types.QualityOnline},
		},
		analogInputs: []*types.AnalogInput{
			{Index: 0, Value: 42, Quality: types.QualityOnline},
		},
		counters: []*types.Counter{
			{Index: 0, Value: 100, Quality: types.QualityOnline},
		},
	}
	cmd := &operateSuccessHandler{}

	server, err := outstation.NewServer(outstation.NewConfig(
		outstation.WithAddress(1024),
		outstation.WithMasterAddress(0xFFFF),
		outstation.WithTransport(dnp3.TCP, "localhost", port),
	))
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	server.SetDataHandler(data)
	server.SetCommandHandler(cmd)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := server.Start(ctx); err != nil {
		t.Fatalf("server.Start: %v", err)
	}
	defer server.Stop(context.Background())

	// Allow the listener to bind before the master dials.
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
	defer client.Close()

	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Connect: %v", err)
	}

	// DirectOperate a CROB (Latch On, index 0). The outstation accepts it and
	// answers with an IIN-only response (no G12V1 echo). MEXT-012 makes this
	// ControlSuccess; before the fix it was ControlTimeout.
	opResp, err := client.Operate(ctx, &types.ControlOutput{
		Group:       12,
		Variation:   1,
		Index:       0,
		CommandType: types.DirectOperate,
		Value:       &types.BinaryCommandValue{Value: true},
	})
	if err != nil {
		t.Fatalf("Operate over real TCP failed: %v", err)
	}
	if opResp.Status != types.ControlSuccess {
		t.Fatalf("Operate status = %v, want ControlSuccess (MEXT-012 R1 fix on real TCP)", opResp.Status)
	}

	// Outstation-side dispatch symmetry: the command handler received the CROB.
	if len(cmd.received) != 1 {
		t.Fatalf("outstation command handler received %d commands, want 1", len(cmd.received))
	}
	got := cmd.received[0]
	if got.Group != 12 || got.Variation != 1 || got.Index != 0 {
		t.Errorf("dispatched command = group %d var %d index %d, want 12/1/0", got.Group, got.Variation, got.Index)
	}
	if bv, ok := got.Value.(*types.BinaryCommandValue); !ok || !bv.Value {
		t.Errorf("dispatched command value = %#v, want BinaryCommandValue{true}", got.Value)
	}
}

// TestOperateRealTCPBlockedStatus asserts a rejected command surfaces a
// non-success status over real TCP (not ControlTimeout): the outstation's
// CommandHandler returns an error, which sets IIN.ParameterError, so the
// IIN-only response maps to a failure status via MEXT-012's commandStatusFromIIN.
func TestOperateRealTCPBlockedStatus(t *testing.T) {
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
	server.SetCommandHandler(&operateRejectingHandler{})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := server.Start(ctx); err != nil {
		t.Fatalf("server.Start: %v", err)
	}
	defer server.Stop(context.Background())
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
	defer client.Close()
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Connect: %v", err)
	}

	opResp, err := client.Operate(ctx, &types.ControlOutput{
		Group:       12,
		Variation:   1,
		Index:       0,
		CommandType: types.DirectOperate,
		Value:       &types.BinaryCommandValue{Value: true},
	})
	if err != nil {
		t.Fatalf("Operate over real TCP failed: %v", err)
	}
	// A rejected command must NEVER be ControlSuccess, and must not be
	// ControlTimeout (the pre-MEXT-012 symptom). The IIN.ParameterError set by
	// the outstation maps to ControlBadFormat via commandStatusFromIIN.
	if opResp.Status == types.ControlSuccess {
		t.Fatalf("Operate status = ControlSuccess, want a failure status for a rejected command")
	}
	if opResp.Status == types.ControlTimeout {
		t.Fatalf("Operate status = ControlTimeout, want a classified failure (MEXT-012 R1 fix on real TCP)")
	}
}

// operateRejectingHandler rejects every binary command with an error, which
// makes the outstation set IIN.ParameterError on its IIN-only DirectOperate
// response.
type operateRejectingHandler struct{}

func (h *operateRejectingHandler) HandleBinaryCommand(cmd *types.ControlOutput) (*types.ControlStatus, error) {
	s := types.ControlBlocked
	return &s, fmt.Errorf("command rejected by test handler")
}

func (h *operateRejectingHandler) HandleAnalogCommand(cmd *types.ControlOutput) (*types.ControlStatus, error) {
	s := types.ControlNotSupported
	return &s, nil
}
