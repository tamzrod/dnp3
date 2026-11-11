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

// VEC-01 Gate C static-object field-order corrective — end-to-end regression.
//
// The original defect (docs/verification/VEC-01-LAB-EXEC-001.md §44) was masked
// because the in-repo master and outstation shared the same (wrong)
// value-then-flags order for G30V1/G20V1, so a self-consistent loopback passed.
// This test pins the externally grounded wire contract end to end over real
// TCP: AI[0]=2300 and CT[0]=42, with the flags octet FIRST on the wire
// (IEEE 1815; opendnp3 3.1.2 Group30Var1=(flags,value32),
// Group20Var1=(flags,count32)). The peer-derived object bytes are
// `01 fc 08 00 00` (flags 0x01 + int32 LE 2300) and `01 2a 00 00 00`.
func TestRealTCPStaticFieldOrderCorrective(t *testing.T) {
	port := getFreePort(t)

	data := &recordingDataHandler{
		analogInputs: []*types.AnalogInput{
			{Index: 0, Value: 2300, Quality: types.QualityOnline},
		},
		counters: []*types.Counter{
			{Index: 0, Value: 42, Quality: types.QualityOnline},
		},
	}

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
	if len(resp.AnalogInputs) != 1 {
		t.Fatalf("AnalogInputs = %+v, want 1 point", resp.AnalogInputs)
	}
	if resp.AnalogInputs[0].Value != 2300 {
		t.Fatalf("AnalogInputs[0].Value = %v, want 2300 (flags-first; 588801 means value-then-flags regressed)", resp.AnalogInputs[0].Value)
	}
	if resp.AnalogInputs[0].Quality&types.QualityOnline == 0 {
		t.Fatalf("AnalogInputs[0].Quality = %v, want ONLINE", resp.AnalogInputs[0].Quality)
	}
	if len(resp.Counters) != 1 {
		t.Fatalf("Counters = %+v, want 1 point", resp.Counters)
	}
	if resp.Counters[0].Value != 42 {
		t.Fatalf("Counters[0].Value = %d, want 42 (flags-first; 10753 means value-then-flags regressed)", resp.Counters[0].Value)
	}
	if resp.Counters[0].Quality&types.QualityOnline == 0 {
		t.Fatalf("Counters[0].Quality = %v, want ONLINE", resp.Counters[0].Quality)
	}
}
