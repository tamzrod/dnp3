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

// TestOperateG41AsymmetricOptInNoExecution asserts that opting in on only one
// role never executes a control. Two cases are exercised:
//
//  1. master opt-in only (stable WithAnalogControl) against a default outstation:
//     the master transmits, but the default outstation fails closed before the
//     host analog handler is invoked.
//  2. outstation opt-in only against a default master: the default master refuses
//     with ErrUnsupportedOption before any wire traffic, so the host handler is
//     never invoked.
func TestOperateG41AsymmetricOptInNoExecution(t *testing.T) {
	t.Run("master-only opt-in does not execute on default outstation", func(t *testing.T) {
		port := getFreePort(t)
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

		resp, err := client.Operate(ctx, types.NewAnalogInt32Control(0, 2300, types.DirectOperate))
		if err == nil && resp != nil && resp.Status == types.ControlSuccess {
			t.Fatalf("default outstation acknowledged a G41 control as ControlSuccess")
		}
		if len(handler.received) != 0 {
			t.Fatalf("default outstation dispatched %d analog commands, want 0", len(handler.received))
		}
	})

	t.Run("outstation-only opt-in does not reach wire from default master", func(t *testing.T) {
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

		// Default master: no analog-control opt-in.
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

		_, err = client.Operate(ctx, types.NewAnalogInt32Control(0, 2300, types.DirectOperate))
		if err == nil {
			t.Fatalf("default master admitted a G41 control, want ErrUnsupportedOption")
		}
		if len(handler.received) != 0 {
			t.Fatalf("outstation dispatched %d analog commands despite a default master, want 0", len(handler.received))
		}
	})
}
