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

// EG628-IM-001 — large static block interoperability over real TCP.
//
// The outstation static builders used to write the point count as a single
// octet, so more than 255 points emitted a wrapped count and left the payload
// unannounced; a strict peer parser then desynchronised (the reported class of
// failure was "truncated object data"). The builders now use the 16-bit count
// qualifier 0x27 above 255 points. This test proves the corrected encoding is
// actually interoperable: a real master, over real TCP, must recover every one
// of 300 points per group with the right index and value.
//
// It uses only the public API and does not depend on either side's internal
// helpers, so a matching encoder/decoder mistake cannot mask a wire defect.
func TestRealTCPLargeStaticBlockAbove255Points(t *testing.T) {
	const points = 300
	port := getFreePort(t)

	bins := make([]*types.BinaryInput, points)
	ais := make([]*types.AnalogInput, points)
	ctrs := make([]*types.Counter, points)
	for i := 0; i < points; i++ {
		bins[i] = &types.BinaryInput{Index: uint16(i), Value: i%2 == 0, Quality: types.QualityOnline}
		ais[i] = &types.AnalogInput{Index: uint16(i), Value: float64(i), Quality: types.QualityOnline}
		ctrs[i] = &types.Counter{Index: uint16(i), Value: uint32(i), Quality: types.QualityOnline}
	}

	server, err := outstation.NewServer(outstation.NewConfig(
		outstation.WithAddress(1024),
		outstation.WithMasterAddress(0xFFFF),
		outstation.WithTransport(dnp3.TCP, "localhost", port),
	))
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	server.SetDataHandler(&recordingDataHandler{
		binaryInputs: bins,
		analogInputs: ais,
		counters:     ctrs,
	})

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
		t.Fatalf("Connect over real TCP failed: %v", err)
	}

	resp, err := client.IntegrityPoll(ctx)
	if err != nil {
		t.Fatalf("IntegrityPoll over real TCP failed: %v", err)
	}

	if len(resp.BinaryInputs) != points {
		t.Fatalf("BinaryInputs = %d, want %d (a wrapped 8-bit count truncates the block)", len(resp.BinaryInputs), points)
	}
	for i, p := range resp.BinaryInputs {
		if int(p.Index) != i || p.Value != (i%2 == 0) {
			t.Fatalf("BinaryInputs[%d] = index %d value %v", i, p.Index, p.Value)
		}
	}

	if len(resp.AnalogInputs) != points {
		t.Fatalf("AnalogInputs = %d, want %d", len(resp.AnalogInputs), points)
	}
	for i, p := range resp.AnalogInputs {
		if int(p.Index) != i || float64(p.Value) != float64(i) {
			t.Fatalf("AnalogInputs[%d] = index %d value %v", i, p.Index, p.Value)
		}
	}

	if len(resp.Counters) != points {
		t.Fatalf("Counters = %d, want %d", len(resp.Counters), points)
	}
	for i, p := range resp.Counters {
		if int(p.Index) != i || uint32(p.Value) != uint32(i) {
			t.Fatalf("Counters[%d] = index %d value %d", i, p.Index, p.Value)
		}
	}
}
