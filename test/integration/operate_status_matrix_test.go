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

// MEXT-024 — Operate success/fail matrix on TCP (status policy coverage).
//
// A table-driven real-TCP matrix over the v0 public Operate (DirectOperate
// G12V1 CROB) covering the three status-policy branches the supported profile
// must classify correctly:
//
//   - success       -> ControlSuccess            (complete APDU, no false timeout)
//   - not_supported -> a failure status          (complete APDU, no false success,
//                                                 no false timeout)
//   - drop          -> failure, no response      (no false success)
//
// Acceptance: no false success; no false timeout on a complete APDU.
//
// Discovery (MEXT-024): the outstation command-handler bridge
// (internalDataHandler.WriteBinaryOutput/WriteAnalogOutput in
// pkg/dnp3/outstation/server.go) previously ignored a non-nil, non-success
// CommandStatus returned by the handler when err == nil, so a handler that
// signalled "not supported" via (ControlNotSupported, nil) surfaced to the
// master as a false ControlSuccess. The bridge now returns an error for a
// non-success status, so the outstation emits an error response (a complete
// APDU, not a silent timeout) and the master resolves it to a failure status.
// This test locks that fix.

// operateStatusOnlyHandler returns a fixed ControlStatus with nil error — the
// natural way a handler signals a non-success outcome without raising an error.
type operateStatusOnlyHandler struct {
	status types.ControlStatus
	got    *types.ControlOutput
}

func (h *operateStatusOnlyHandler) HandleBinaryCommand(cmd *types.ControlOutput) (*types.ControlStatus, error) {
	h.got = cmd
	s := h.status
	return &s, nil
}
func (h *operateStatusOnlyHandler) HandleAnalogCommand(cmd *types.ControlOutput) (*types.ControlStatus, error) {
	s := types.ControlNotSupported
	return &s, nil
}

// operateStatusMatrixRow is one row of the complete-APDU operate matrix (the
// success and not-supported branches). The drop branch is exercised
// separately below because it tears the transport down before Operate.
type operateStatusMatrixRow struct {
	name        string
	handler     outstation.CommandHandler
	wantSuccess bool
}

// TestOperateStatusMatrixOnTCP runs the complete-APDU operate branches on a
// real TCP loopback. Each row produces a real response APDU (the outstation
// replies), so the master must never report ControlTimeout for these rows
// ("no false timeout on complete APDU"), and a non-success handler must never
// report ControlSuccess ("no false success").
func TestOperateStatusMatrixOnTCP(t *testing.T) {
	rows := []operateStatusMatrixRow{
		{
			name:        "success",
			handler:     &operateSuccessHandler{},
			wantSuccess: true,
		},
		{
			name:        "not_supported",
			handler:     &operateStatusOnlyHandler{status: types.ControlNotSupported},
			wantSuccess: false,
		},
	}

	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
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
			server.SetCommandHandler(row.handler)

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

			opResp, err := client.Operate(ctx, &types.ControlOutput{
				Group:       12,
				Variation:   1,
				Index:       0,
				CommandType: types.DirectOperate,
				Value:       &types.BinaryCommandValue{Value: true},
			})
			if err != nil {
				t.Fatalf("Operate returned error for %s row (expected a response): %v", row.name, err)
			}

			// "no false timeout on complete APDU": the outstation replied with a
			// complete APDU, so the master must not classify it as ControlTimeout.
			if opResp.Status == types.ControlTimeout {
				t.Fatalf("%s: false timeout on complete APDU (status=%v, IIN=%v)",
					row.name, opResp.Status, opResp.IIN)
			}

			if row.wantSuccess {
				// "no false success" does not constrain the success row, but a
				// success row must actually report ControlSuccess.
				if opResp.Status != types.ControlSuccess {
					t.Fatalf("%s: expected ControlSuccess, got %v (IIN=%v)",
						row.name, opResp.Status, opResp.IIN)
				}
			} else {
				// "no false success": a handler that signalled a non-success
				// status must never surface as ControlSuccess.
				if opResp.Status == types.ControlSuccess {
					t.Fatalf("%s: false success (status=ControlSuccess, IIN=%v) — "+
						"the outstation bridge must surface a non-success handler status as a failure",
						row.name, opResp.IIN)
				}
			}
		})
	}
}

// TestOperateStatusMatrixOnTCPDrop exercises the "timeout on drop" branch: the
// TCP peer (outstation) is dropped before Operate, so the master cannot receive
// a response. The master must surface a failure — never ControlSuccess. There is
// no complete APDU, so this row only asserts "no false success".
func TestOperateStatusMatrixOnTCPDrop(t *testing.T) {
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
	server.SetCommandHandler(&operateSuccessHandler{})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := server.Start(ctx); err != nil {
		t.Fatalf("server.Start: %v", err)
	}
	time.Sleep(150 * time.Millisecond)

	client, err := master.NewClient(master.NewConfig(
		master.WithMasterAddress(0xFFFF),
		master.WithOutstationAddress(1024),
		master.WithTransport(dnp3.TCP, "localhost", port),
		master.WithTimeout(2*time.Second),
		master.WithRetry(0, 0),
	))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	defer client.Close()
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Connect: %v", err)
	}

	// Drop the TCP peer before Operate: server shutdown closes all active
	// connection transports (pkg/dnp3/outstation/server.go shutdown).
	if err := server.Stop(context.Background()); err != nil {
		t.Fatalf("server.Stop: %v", err)
	}
	time.Sleep(250 * time.Millisecond)

	opResp, err := client.Operate(ctx, &types.ControlOutput{
		Group:       12,
		Variation:   1,
		Index:       0,
		CommandType: types.DirectOperate,
		Value:       &types.BinaryCommandValue{Value: true},
	})

	// "no false success" on drop: Operate must not report ControlSuccess. With
	// the peer gone, Operate returns an error and a nil response (no status to
	// misread as success). Either an error-with-nil-response or a non-success
	// status is acceptable; ControlSuccess is never acceptable.
	if opResp != nil && opResp.Status == types.ControlSuccess {
		t.Fatalf("drop: false success — Operate reported ControlSuccess after the peer was dropped (err=%v)", err)
	}
	if err == nil {
		t.Fatalf("drop: expected a failure (error/timeout) after the peer was dropped, got err=nil, resp=%v", opResp)
	}
	// The master should reflect the dropped link in its public state so a later
	// Operate does not retry a dead link (DNP3-031).
	if client.State() == dnp3.StateConnected {
		t.Fatalf("drop: master still reports Connected after the peer dropped (state=%v)", client.State())
	}
}
