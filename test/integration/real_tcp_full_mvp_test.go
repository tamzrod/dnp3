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

// MEXT-022 — Real-TCP full MVP master path.
//
// Unlike mvp_loopback_test.go (which uses the in-memory
// testutils.NewMVPOutstationSimulator), this test exercises the complete v0
// public API over a REAL TCP master↔outstation loopback with no simulator
// transport: Connect → IntegrityPoll (assert all MVP points) → Operate
// (assert operate policy) → Disconnect → terminal state. Real TCP transport,
// real DNP3 wire framing (AL → TL → DLL), and the real in-repo outstation's
// read/operate handlers are exercised end to end.
//
// Acceptance (MEXT-022): green locally and in
// scripts/verify-external-mvp.sh Tier 1. Points + operate policy asserted.

// TestRealTCPFullMVPPath is the single end-to-end real-TCP MVP path. It uses
// the shared recordingDataHandler / recordingCommandHandler helpers so the
// outstation serves known Class-0 points and records the dispatched CROB.
func TestRealTCPFullMVPPath(t *testing.T) {
	port := getFreePort(t)

	data := &recordingDataHandler{
		binaryInputs: []*types.BinaryInput{
			{Index: 0, Value: true, Quality: types.QualityOnline},
			{Index: 1, Value: false, Quality: types.QualityOnline},
		},
		analogInputs: []*types.AnalogInput{
			// Integer-valued so the G30V1 signed-32-bit encoding round-trips
			// exactly (no float precision loss).
			{Index: 0, Value: 42, Quality: types.QualityOnline},
			{Index: 1, Value: -7, Quality: types.QualityOnline},
		},
		counters: []*types.Counter{
			{Index: 0, Value: 100, Quality: types.QualityOnline},
			{Index: 1, Value: 7, Quality: types.QualityOnline},
		},
	}
	cmd := &recordingCommandHandler{}

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

	// 1. Connect — the public client must reach StateActive over real TCP.
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Connect over real TCP failed: %v", err)
	}
	if got := client.State(); got != dnp3.StateActive {
		t.Fatalf("state after Connect = %v, want Active", got)
	}

	// 2. IntegrityPoll — read all MVP Class-0 static groups over real TCP in
	//    one call and assert the configured points surface (points policy).
	resp, err := client.IntegrityPoll(ctx)
	if err != nil {
		t.Fatalf("IntegrityPoll over real TCP failed: %v", err)
	}
	if resp == nil {
		t.Fatal("IntegrityPoll returned nil response")
	}
	if len(resp.BinaryInputs) != 2 {
		t.Fatalf("BinaryInputs = %+v, want 2 points", resp.BinaryInputs)
	}
	if !resp.BinaryInputs[0].Value || resp.BinaryInputs[1].Value {
		t.Fatalf("BinaryInputs values = %+v, want [true, false]", resp.BinaryInputs)
	}
	if len(resp.AnalogInputs) != 2 {
		t.Fatalf("AnalogInputs = %+v, want 2 points", resp.AnalogInputs)
	}
	if resp.AnalogInputs[0].Value != 42 || resp.AnalogInputs[1].Value != -7 {
		t.Fatalf("AnalogInputs values = %+v, want [42, -7]", resp.AnalogInputs)
	}
	if len(resp.Counters) != 2 {
		t.Fatalf("Counters = %+v, want 2 points", resp.Counters)
	}
	if resp.Counters[0].Value != 100 || resp.Counters[1].Value != 7 {
		t.Fatalf("Counters values = %+v, want [100, 7]", resp.Counters)
	}

	// 3. Operate — DirectOperate a CROB (Latch On, index 0) over real TCP.
	//    Operate policy: the outstation accepts → ControlSuccess (not
	//    ControlTimeout; the MEXT-012 R1 fix on real TCP).
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
		t.Fatalf("Operate status = %v, want ControlSuccess", opResp.Status)
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

	// 4. Disconnect — the client must reach the terminal StateDisconnected
	//    (DNP3-024/050) and leave a clean slate (seq reset) for a reconnect.
	if err := client.Disconnect(ctx); err != nil {
		t.Fatalf("Disconnect over real TCP failed: %v", err)
	}
	if got := client.State(); got != dnp3.StateDisconnected {
		t.Fatalf("state after Disconnect = %v, want Disconnected", got)
	}
}
