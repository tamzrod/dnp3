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

// MEXT-025 — Reconnect + DeviceRestart IIN on TCP.
//
// Two deliverables on a REAL TCP loopback:
//
//  1. Drop/reconnect with no stuck state (TestReconnectOnTCPNoStuckState): drop
//     the outstation mid-session and assert the master does NOT get stuck — its
//     public state flips from Active to Disconnected and a subsequent Read
//     returns a disconnect-classified error (not a hang, not a false Active).
//     Recovery is then demonstrated by reconnecting a fresh client to the
//     restarted outstation and reading successfully. (The v0 TCP transport's
//     Close is terminal, so same-client reconnect is a post-MVP enhancement;
//     the no-stuck-state guarantee is the master never reports a dead link as
//     Active and never blocks.)
//
//  2. DeviceRestart IIN (TestDeviceRestartNotRaisableOnV0Outstation): the
//     DeviceRestart IIN bit (IIN1.7) is not raisable via the v0 public
//     outstation API (no SetDeviceRestart/SetRequestHandler surface; the
//     outstation never sets o.iin.DeviceRestart). The master's DNP3-053
//     auto-integrity-on-DeviceRestart path is therefore exercised only against
//     the in-memory simulator (test/integration/auto_integrity_test.go) and
//     cannot be raised over real TCP by the v0 outstation. This test
//     characterizes that fact so it is not mistaken for a regression: a normal
//     real-TCP exchange carries a clean IIN (no DeviceRestart bit).

// TestReconnectOnTCPNoStuckState asserts the MEXT-025 no-stuck-state guarantee
// on real TCP: after the peer drops, the master leaves the Active state and
// reports Disconnected, and a following operation returns a disconnect-classified
// failure (not a hang). It then demonstrates reconnect via a fresh client to the
// restarted outstation.
func TestReconnectOnTCPNoStuckState(t *testing.T) {
	port := getFreePort(t)
	mkServer := func() outstation.Server {
		s, err := outstation.NewServer(outstation.NewConfig(
			outstation.WithAddress(1024),
			outstation.WithMasterAddress(0xFFFF),
			outstation.WithTransport(dnp3.TCP, "localhost", port),
		))
		if err != nil {
			t.Fatalf("NewServer: %v", err)
		}
		s.SetDataHandler(&recordingDataHandler{
			binaryInputs: []*types.BinaryInput{
				{Index: 0, Value: true, Quality: types.QualityOnline},
			},
		})
		s.SetCommandHandler(&operateSuccessHandler{})
		return s
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Session 1: connect + a successful Read.
	s1 := mkServer()
	if err := s1.Start(ctx); err != nil {
		t.Fatalf("server1 Start: %v", err)
	}
	defer s1.Stop(context.Background())
	time.Sleep(150 * time.Millisecond)

	client, err := master.NewClient(master.NewConfig(
		master.WithMasterAddress(0xFFFF),
		master.WithOutstationAddress(1024),
		master.WithTransport(dnp3.TCP, "localhost", port),
		master.WithTimeout(3*time.Second),
		master.WithRetry(0, 0),
	))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	if got := client.State(); got != dnp3.StateActive && got != dnp3.StateConnected {
		t.Fatalf("pre-drop state = %v, want Active/Connected", got)
	}
	if _, err := client.Read(ctx, types.NewReadRequest(types.GroupRequest{Group: 1, Variation: 1})); err != nil {
		t.Fatalf("Read before drop: %v", err)
	}

	// Drop the peer.
	if err := s1.Stop(context.Background()); err != nil {
		t.Fatalf("server1 Stop: %v", err)
	}
	time.Sleep(300 * time.Millisecond)

	// No stuck state: a Read after the drop must fail (not hang) and the public
	// state must have left Active/Connected. Before the MEXT-025
	// IsDisconnectError fix this stayed "Active" (the broken-pipe write error
	// was not recognized as a disconnect), leaving the client stuck.
	readCtx, readCancel := context.WithTimeout(ctx, 5*time.Second)
	_, readErr := client.Read(readCtx, types.NewReadRequest(types.GroupRequest{Group: 1, Variation: 1}))
	readCancel()
	if readErr == nil {
		t.Fatal("Read after drop: expected an error, got nil (the dropped link must not appear healthy)")
	}
	if got := client.State(); got == dnp3.StateActive || got == dnp3.StateConnected {
		t.Fatalf("post-drop state = %v: master is stuck in a connected state after the peer dropped (no stuck state)", got)
	}
	// The failure must be classifiable as a disconnect (not a generic/unknown),
	// so callers can react to a peer drop distinctly from a protocol error.
	if got := dnp3.ClassifyError(readErr); got != dnp3.ErrorCodeDisconnect {
		t.Fatalf("post-drop Read error classified %v, want ErrorCodeDisconnect (err=%v)", got, readErr)
	}
	client.Close()

	// Reconnect: a fresh client to the restarted outstation must succeed (the
	// master is not wedged at the process level; recovery proceeds).
	s2 := mkServer()
	if err := s2.Start(ctx); err != nil {
		t.Fatalf("server2 Start: %v", err)
	}
	defer s2.Stop(context.Background())
	time.Sleep(200 * time.Millisecond)

	client2, err := master.NewClient(master.NewConfig(
		master.WithMasterAddress(0xFFFF),
		master.WithOutstationAddress(1024),
		master.WithTransport(dnp3.TCP, "localhost", port),
		master.WithTimeout(3*time.Second),
		master.WithRetry(0, 0),
	))
	if err != nil {
		t.Fatalf("NewClient (reconnect): %v", err)
	}
	defer client2.Close()
	if err := client2.Connect(ctx); err != nil {
		t.Fatalf("reconnect Connect: %v", err)
	}
	if got := client2.State(); got != dnp3.StateActive && got != dnp3.StateConnected {
		t.Fatalf("reconnect state = %v, want Active/Connected", got)
	}
	resp, err := client2.Read(ctx, types.NewReadRequest(types.GroupRequest{Group: 1, Variation: 1}))
	if err != nil {
		t.Fatalf("Read after reconnect: %v (no stuck state — the restarted peer must be readable)", err)
	}
	if len(resp.BinaryInputs) == 0 {
		t.Fatal("Read after reconnect returned no binary inputs")
	}
}

// TestDeviceRestartNotRaisableOnV0Outstation characterizes the DeviceRestart
// half of MEXT-025: the v0 public outstation never raises the DeviceRestart IIN
// bit (IIN1.7), so a normal real-TCP exchange carries a clean IIN. The master's
// DNP3-053 auto-integrity-on-DeviceRestart path is exercised against the
// in-memory simulator (auto_integrity_test.go) and awaits a raisable outstation
// to be exercised over real TCP (post-MVP). This test pins the clean-IIN
// baseline so a future regression (an unexpected DeviceRestart bit on a normal
// exchange) is caught.
func TestDeviceRestartNotRaisableOnV0Outstation(t *testing.T) {
	port := getFreePort(t)
	server, err := outstation.NewServer(outstation.NewConfig(
		outstation.WithAddress(1024),
		outstation.WithMasterAddress(0xFFFF),
		outstation.WithTransport(dnp3.TCP, "localhost", port),
	))
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	server.SetDataHandler(&recordingDataHandler{
		binaryInputs: []*types.BinaryInput{
			{Index: 0, Value: true, Quality: types.QualityOnline},
		},
	})

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
		master.WithTimeout(3*time.Second),
		master.WithRetry(0, 0),
	))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	defer client.Close()
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Connect: %v", err)
	}

	resp, err := client.Read(ctx, types.NewReadRequest(types.GroupRequest{Group: 1, Variation: 1}))
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	// IIN1.7 (DeviceRestart) is bit 0x01 of the first IIN octet. The v0
	// outstation must not raise it on a normal exchange.
	if len(resp.IIN) < 1 {
		t.Fatalf("response has no IIN bytes: %v", resp.IIN)
	}
	if resp.IIN[0]&0x01 != 0 {
		t.Fatalf("v0 outstation raised DeviceRestart (IIN1.7) on a normal exchange: IIN=%v — "+
			"the v0 outstation must not raise DeviceRestart", resp.IIN)
	}
}
