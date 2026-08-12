package integration

import (
	"context"
	"sync"
	"testing"
	"time"

	"dnp3/pkg/dnp3"
	"dnp3/pkg/dnp3/master"
	"dnp3/pkg/dnp3/outstation"
	"dnp3/pkg/dnp3/types"
)

// outstationSideMVPGate mirrors the Master-side MVP gate
// (mvp_loopback_test.go / public_api_loopback_test.go) from the OUTSTATION's
// perspective (DNP3-091). It asserts the outstation-side observables for the
// same scenarios the master gate exercises:
//   - the outstation serves the configured Class-0 static points (the same
//     data the master read) via its DataHandler getters (read symmetry);
//   - the outstation dispatches a direct-operate CROB to the CommandHandler
//     with the correct Group 12 / Variation 1 / index / value (operate
//     dispatch symmetry);
//   - the outstation transitions to StateRunning on Start and StateDown
//     after a clean Stop.
//
// Acceptance: "Both directions green" — the master-side direction is green
// (mvp_loopback against the simulator); this test adds the outstation-side
// direction by asserting outstation-side observables through a real
// outstation server + master client loopback.
//
// Note (DNP3-091 discovery, updated by MEXT-012/013): the real outstation's
// DirectOperate response carries no control-status object echo — it is an
// IIN-only response. MEXT-012 taught resolveOperateStatus to treat an IIN-only
// response with clear IIN as CommandStatusSuccess, and MEXT-013 proves that fix
// on real TCP (see operate_real_tcp_test.go: TestOperateRealTCPSuccess asserts
// ControlSuccess against this same outstation path). This test still asserts
// only the outstation-side dispatch observable (the CommandHandler receives the
// CROB); the master-side success assertion lives in operate_real_tcp_test.go.

// recordingDataHandler records which MVP getters were called and serves a
// fixed set of points so the outstation-side can be observed.
type recordingDataHandler struct {
	mu             sync.Mutex
	binaryCalled   int
	analogCalled   int
	counterCalled  int
	binaryInputs   []*types.BinaryInput
	analogInputs   []*types.AnalogInput
	counters       []*types.Counter
}

func (h *recordingDataHandler) GetBinaryInputs() []*types.BinaryInput {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.binaryCalled++
	return h.binaryInputs
}

func (h *recordingDataHandler) GetAnalogInputs() []*types.AnalogInput {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.analogCalled++
	return h.analogInputs
}

func (h *recordingDataHandler) GetCounters() []*types.Counter {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.counterCalled++
	return h.counters
}

func (h *recordingDataHandler) GetBinaryOutputs() []*types.BinaryOutput  { return nil }
func (h *recordingDataHandler) GetAnalogOutputs() []*types.AnalogOutput  { return nil }
func (h *recordingDataHandler) GetFrozenCounters() []*types.Counter       { return nil }
func (h *recordingDataHandler) FreezeCounters(clear bool) error           { return nil }

// recordingCommandHandler records the dispatched control command so the
// outstation-side dispatch can be asserted.
type recordingCommandHandler struct {
	mu       sync.Mutex
	received []*types.ControlOutput
}

func (h *recordingCommandHandler) HandleBinaryCommand(cmd *types.ControlOutput) (*types.ControlStatus, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.received = append(h.received, cmd)
	status := types.ControlSuccess
	return &status, nil
}

func (h *recordingCommandHandler) HandleAnalogCommand(cmd *types.ControlOutput) (*types.ControlStatus, error) {
	status := types.ControlNotSupported
	return &status, nil
}

func TestOutstationSideMVPGate(t *testing.T) {
	port := getFreePort(t)

	data := &recordingDataHandler{
		binaryInputs: []*types.BinaryInput{
			{Index: 0, Value: true, Quality: types.QualityOnline},
			{Index: 1, Value: false, Quality: types.QualityOnline},
		},
		analogInputs: []*types.AnalogInput{
			{Index: 0, Value: 42, Quality: types.QualityOnline},
		},
		counters: []*types.Counter{
			{Index: 0, Value: 100, Quality: types.QualityOnline},
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

	// Outstation-side: Start -> StateRunning.
	if err := server.Start(ctx); err != nil {
		t.Fatalf("server.Start: %v", err)
	}
	if got := server.State(); got != outstation.ServerStateRunning {
		t.Fatalf("state after Start = %v, want Running", got)
	}
	defer server.Stop(context.Background())

	// Short master timeout keeps this outstation-side dispatch test fast; the
	// master-side Operate success is covered by operate_real_tcp_test.go.
	client, err := master.NewClient(master.NewConfig(
		master.WithOutstationAddress(1024),
		master.WithTransport(dnp3.TCP, "localhost", port),
		master.WithTimeout(300*time.Millisecond),
		master.WithRetry(1, 0),
	))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("client.Connect: %v", err)
	}
	defer client.Close()

	// 1. Read Class-0 static MVP groups; the outstation-side getters must be
	//    invoked and the data the master received must match what the
	//    outstation served (read symmetry).
	resp, err := client.Read(ctx, types.NewReadRequest(types.GroupRequest{Group: 1, Variation: 1}))
	if err != nil {
		t.Fatalf("Read G1V1: %v", err)
	}
	if len(resp.BinaryInputs) != 2 || !resp.BinaryInputs[0].Value || resp.BinaryInputs[1].Value {
		t.Fatalf("master received binary inputs %+v, want the outstation's [true,false]", resp.BinaryInputs)
	}

	if _, err := client.Read(ctx, types.NewReadRequest(types.GroupRequest{Group: 30, Variation: 1})); err != nil {
		t.Fatalf("Read G30V1: %v", err)
	}
	if _, err := client.Read(ctx, types.NewReadRequest(types.GroupRequest{Group: 20, Variation: 1})); err != nil {
		t.Fatalf("Read G20V1: %v", err)
	}

	data.mu.Lock()
	binaryCalls := data.binaryCalled
	analogCalls := data.analogCalled
	counterCalls := data.counterCalled
	data.mu.Unlock()
	if binaryCalls == 0 {
		t.Error("outstation GetBinaryInputs was not invoked by the read path")
	}
	if analogCalls == 0 {
		t.Error("outstation GetAnalogInputs was not invoked by the read path")
	}
	if counterCalls == 0 {
		t.Error("outstation GetCounters was not invoked by the read path")
	}

	// 2. Operate a CROB; the outstation-side CommandHandler must receive it
	//    with Group=12, Variation=1, the correct index, and a binary value
	//    (operate dispatch symmetry). The master's Operate status is not
	//    asserted here — see the DNP3-091 discovery note above.
	_, _ = client.Operate(ctx, &types.ControlOutput{
		Group:       12,
		Variation:   1,
		Index:       0,
		CommandType: types.DirectOperate,
		Value:       &types.BinaryCommandValue{Value: true},
	})

	cmd.mu.Lock()
	received := cmd.received
	cmd.mu.Unlock()
	if len(received) != 1 {
		t.Fatalf("outstation command handler received %d commands, want 1", len(received))
	}
	got := received[0]
	if got.Group != 12 || got.Variation != 1 {
		t.Errorf("dispatched command group/var = %d/%d, want 12/1", got.Group, got.Variation)
	}
	if got.Index != 0 {
		t.Errorf("dispatched command index = %d, want 0", got.Index)
	}
	if bv, ok := got.Value.(*types.BinaryCommandValue); !ok || !bv.Value {
		t.Errorf("dispatched command value = %#v, want BinaryCommandValue{true}", got.Value)
	}

	// 3. Clean Stop -> StateDown (outstation-side symmetry with master Close).
	if err := client.Close(); err != nil {
		t.Fatalf("client.Close: %v", err)
	}
	cancel() // release the server's run context
	if err := server.Stop(context.Background()); err != nil {
		t.Fatalf("server.Stop: %v", err)
	}
	if got := server.State(); got != outstation.ServerStateDown {
		t.Fatalf("state after Stop = %v, want Down", got)
	}
}
