package integration

import (
	"context"
	"sync"
	"testing"
	"time"

	"dnp3/internal/al"
	"dnp3/pkg/dnp3"
	"dnp3/pkg/dnp3/master"
	"dnp3/pkg/dnp3/outstation"
	"dnp3/pkg/dnp3/types"
)

// largeStaticDataHandler serves a fixed number of alternating binary inputs and
// ascending analog inputs/counters, so a Class-0 read must announce more than
// 255 points per group and therefore use the 16-bit count qualifier.
type largeStaticDataHandler struct {
	mu      sync.Mutex
	points  int
	binCall int
	aiCall  int
	ctrCall int
}

func (h *largeStaticDataHandler) GetBinaryInputs() []*types.BinaryInput {
	h.mu.Lock()
	h.binCall++
	h.mu.Unlock()
	out := make([]*types.BinaryInput, h.points)
	for i := range out {
		out[i] = &types.BinaryInput{Index: uint16(i), Value: i%2 == 0, Quality: types.QualityOnline}
	}
	return out
}

func (h *largeStaticDataHandler) GetAnalogInputs() []*types.AnalogInput {
	h.mu.Lock()
	h.aiCall++
	h.mu.Unlock()
	out := make([]*types.AnalogInput, h.points)
	for i := range out {
		out[i] = &types.AnalogInput{Index: uint16(i), Value: float64(i), Quality: types.QualityOnline}
	}
	return out
}

func (h *largeStaticDataHandler) GetCounters() []*types.Counter {
	h.mu.Lock()
	h.ctrCall++
	h.mu.Unlock()
	out := make([]*types.Counter, h.points)
	for i := range out {
		out[i] = &types.Counter{Index: uint16(i), Value: uint32(i), Quality: types.QualityOnline}
	}
	return out
}

func (h *largeStaticDataHandler) GetBinaryOutputs() []*types.BinaryOutput { return nil }
func (h *largeStaticDataHandler) GetAnalogOutputs() []*types.AnalogOutput { return nil }
func (h *largeStaticDataHandler) GetFrozenCounters() []*types.Counter     { return nil }
func (h *largeStaticDataHandler) FreezeCounters(clear bool) error         { return nil }

// TestLargeStaticBlockRealTCP is the end-to-end gate for a static point count
// above 255. The outstation must announce each group with the 16-bit count
// qualifier, the master must parse every point with its correct index and value
// from a multi-fragment Class-0 response, and the event buffers must stay empty
// (no static data leaks into an event class).
func TestLargeStaticBlockRealTCP(t *testing.T) {
	const points = 300

	port := getFreePort(t)
	data := &largeStaticDataHandler{points: points}

	server, err := outstation.NewServer(outstation.NewConfig(
		outstation.WithAddress(1024),
		outstation.WithMasterAddress(0xFFFF),
		outstation.WithTransport(dnp3.TCP, "localhost", port),
	))
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	server.SetDataHandler(data)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := server.Start(ctx); err != nil {
		t.Fatalf("server.Start: %v", err)
	}
	defer server.Stop(context.Background())

	client, err := master.NewClient(master.NewConfig(
		master.WithOutstationAddress(1024),
		master.WithTransport(dnp3.TCP, "localhost", port),
		master.WithTimeout(5*time.Second),
		master.WithRetry(1, 0),
	))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	defer client.Close()

	resp, err := client.IntegrityPoll(ctx)
	if err != nil {
		t.Fatalf("IntegrityPoll: %v", err)
	}

	if len(resp.BinaryInputs) != points {
		t.Fatalf("BinaryInputs = %d, want %d (an 8-bit count would wrap and truncate)", len(resp.BinaryInputs), points)
	}
	for i, p := range resp.BinaryInputs {
		if int(p.Index) != i || p.Value != (i%2 == 0) {
			t.Fatalf("BinaryInputs[%d] = index %d value %v, want index %d value %v", i, p.Index, p.Value, i, i%2 == 0)
		}
	}

	if len(resp.AnalogInputs) != points {
		t.Fatalf("AnalogInputs = %d, want %d", len(resp.AnalogInputs), points)
	}
	for i, p := range resp.AnalogInputs {
		if int(p.Index) != i || float64(p.Value) != float64(i) {
			t.Fatalf("AnalogInputs[%d] = index %d value %v, want index %d value %d", i, p.Index, p.Value, i, i)
		}
	}

	if len(resp.Counters) != points {
		t.Fatalf("Counters = %d, want %d", len(resp.Counters), points)
	}
	for i, p := range resp.Counters {
		if int(p.Index) != i || uint32(p.Value) != uint32(i) {
			t.Fatalf("Counters[%d] = index %d value %d, want index %d value %d", i, p.Index, p.Value, i, i)
		}
	}

	// A static read must not be reported as available event data.
	iin, err := al.DecodeIIN(resp.IIN[:])
	if err != nil {
		t.Fatalf("decode IIN: %v", err)
	}
	if iin.Class1Events || iin.Class2Events || iin.Class3Events {
		t.Fatalf("IIN reports event data after a static read: % X", resp.IIN)
	}
	if iin.ObjectUnknown || iin.ParameterError {
		t.Fatalf("IIN reports an error after a servable large static read: % X", resp.IIN)
	}
}
