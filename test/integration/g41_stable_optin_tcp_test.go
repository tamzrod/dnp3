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

// g41StableOptInHandler accepts index 0 and reports OUT_OF_RANGE for index 1, so
// a single loopback exercises both the success and out-of-range status paths.
type g41StableOptInHandler struct {
	received []*types.ControlOutput
}

func (h *g41StableOptInHandler) HandleBinaryCommand(cmd *types.ControlOutput) (*types.ControlStatus, error) {
	s := types.ControlSuccess
	return &s, nil
}

func (h *g41StableOptInHandler) HandleAnalogCommand(cmd *types.ControlOutput) (*types.ControlStatus, error) {
	h.received = append(h.received, cmd)
	if cmd.Index == 0 {
		s := types.ControlSuccess
		return &s, nil
	}
	s := types.ControlOutOfRange
	return &s, nil
}

// TestOperateG41StableOptInBothRolesRealTCP drives every G41 variation through
// the stable master.WithAnalogControl() / outstation.WithAnalogControl() pair
// over real TCP: index 0 must surface ControlSuccess, index 1 must surface
// OUT_OF_RANGE, and the typed value must reach the host handler unchanged. This
// pins the stable opt-in name end to end (the legacy experimental name is
// covered by the existing g41v1-v4 real-TCP tests).
func TestOperateG41StableOptInBothRolesRealTCP(t *testing.T) {
	port := getFreePort(t)

	server, err := outstation.NewServer(outstation.NewConfig(
		outstation.WithAddress(1024),
		outstation.WithMasterAddress(0xFFFF),
		outstation.WithTransport(dnp3.TCP, "localhost", port),
		outstation.WithAnalogControl(),
	))
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	handler := &g41StableOptInHandler{}
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
		master.WithAnalogControl(),
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

	cases := []struct {
		name string
		cmd  *types.ControlOutput
	}{
		{"G41V1 int32", types.NewAnalogInt32Control(0, 2300, types.DirectOperate)},
		{"G41V2 int16", types.NewAnalogInt16Control(0, 12345, types.DirectOperate)},
		{"G41V3 float32", types.NewAnalogFloat32Control(0, 1.5, types.DirectOperate)},
		{"G41V4 float64", types.NewAnalogFloat64Control(0, 1.5, types.DirectOperate)},
	}

	for _, tc := range cases {
		t.Run(tc.name+" index 0 success", func(t *testing.T) {
			resp, err := client.Operate(ctx, tc.cmd)
			if err != nil {
				t.Fatalf("Operate: %v", err)
			}
			if resp.Status != types.ControlSuccess {
				t.Fatalf("status = %v, want ControlSuccess", resp.Status)
			}
		})

		t.Run(tc.name+" index 1 out of range", func(t *testing.T) {
			oor := *tc.cmd
			oor.Index = 1
			resp, err := client.Operate(ctx, &oor)
			if err != nil {
				t.Fatalf("Operate: %v", err)
			}
			if resp.Status != types.ControlOutOfRange {
				t.Fatalf("status = %v, want ControlOutOfRange", resp.Status)
			}
			if resp.Status == types.ControlSuccess {
				t.Fatalf("out-of-range index surfaced as ControlSuccess")
			}
		})
	}

	if len(handler.received) != 8 {
		t.Fatalf("host handler received %d commands, want 8", len(handler.received))
	}
}
