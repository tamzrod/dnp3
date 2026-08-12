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

func TestPublicAPILoopbackReadAndDirectControl(t *testing.T) {
	port := getFreePort(t)
	server, err := outstation.NewServer(outstation.NewConfig(
		outstation.WithAddress(1024),
		outstation.WithMasterAddress(0xFFFF),
		outstation.WithTransport(dnp3.TCP, "localhost", port),
	))
	if err != nil {
		t.Fatal(err)
	}
	server.SetDataHandler(&comprehensiveDataHandler{
		binaryInputs: []*types.BinaryInput{{Value: true, Quality: types.QualityOnline}},
	})
	server.SetCommandHandler(&comprehensiveCommandHandler{})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := server.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer server.Stop(context.Background())

	client, err := master.NewClient(master.NewConfig(
		master.WithOutstationAddress(1024),
		master.WithTransport(dnp3.TCP, "localhost", port),
		master.WithTimeout(2*time.Second),
		master.WithRetry(1, 0),
	))
	if err != nil {
		t.Fatal(err)
	}
	if err := client.Connect(ctx); err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	resp, err := client.Read(ctx, types.NewReadRequest(types.GroupRequest{Group: 1, Variation: 1}))
	if err != nil {
		t.Fatalf("public read failed: %v", err)
	}
	if len(resp.BinaryInputs) != 1 || !resp.BinaryInputs[0].Value {
		t.Fatalf("unexpected binary inputs: %+v", resp.BinaryInputs)
	}

	// DNP3-012: the public ReadResponse must carry the outstation's IIN,
	// and the master's stored IIN must match the per-response IIN.
	if resp.IIN != client.LastIIN() {
		t.Fatalf("ReadResponse.IIN = %v, LastIIN = %v (expected equal)", resp.IIN, client.LastIIN())
	}

	_, err = client.Operate(ctx, &types.ControlOutput{
		Group: 12, Variation: 1, Index: 0,
		CommandType: types.DirectOperate,
		Value:       &types.BinaryCommandValue{Value: true},
	})
	if err != nil {
		t.Fatalf("public direct control failed: %v", err)
	}
}
