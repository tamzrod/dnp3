package integration

import (
	"context"
	"testing"
	"time"

	"dnp3/internal/al"
	"dnp3/pkg/dnp3"
	"dnp3/pkg/dnp3/master"
	"dnp3/pkg/dnp3/outstation"
)

func TestTCPUnsolicitedAndClassRead(t *testing.T) {
	port := getFreePort(t)
	server, err := outstation.NewServer(outstation.NewConfig(
		outstation.WithAddress(1024),
		outstation.WithMasterAddress(1),
		outstation.WithTransport(dnp3.TCP, "127.0.0.1", port),
		outstation.WithExperimentalEventDelivery(),
	))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := server.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer server.Stop(context.Background())

	stamp := time.Date(2020, 1, 2, 3, 4, 5, int(6*time.Millisecond), time.UTC)
	if err := server.PublishBinaryEvent(7, true, al.BinEventOnline, outstation.EventClass1, stamp); err != nil {
		t.Fatal(err)
	}
	if err := server.PublishAnalogEvent(3, -2, al.AnalogEventOnline, outstation.EventClass2, stamp); err != nil {
		t.Fatal(err)
	}
	if err := server.PublishCounterEvent(4, 0x01020304, al.CounterEventOnline|al.CounterEventRollover, outstation.EventClass3, time.Time{}); err != nil {
		t.Fatal(err)
	}

	client, err := master.NewClient(master.NewConfig(
		master.WithMasterAddress(1),
		master.WithOutstationAddress(1024),
		master.WithTransport(dnp3.TCP, "127.0.0.1", port),
		master.WithTimeout(3*time.Second),
		master.WithRetry(1, 0),
		master.WithExperimentalEventDelivery(),
	))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	if err := client.Connect(ctx); err != nil {
		t.Fatal(err)
	}

	polled, err := client.IntegrityPoll(ctx)
	if err != nil {
		t.Fatalf("class 0 poll: %v", err)
	}
	if len(polled.BinaryInputs) == 0 {
		t.Fatal("class 0 returned no binary inputs")
	}

	events, err := client.ReadEvents(ctx)
	if err != nil {
		t.Fatalf("ReadEvents: %v", err)
	}
	if len(events.BinaryInputs) != 1 || !events.BinaryInputs[0].Value || events.BinaryInputs[0].Index != 7 {
		t.Fatalf("binary events %+v", events.BinaryInputs)
	}
	if events.BinaryInputs[0].Time == nil || !events.BinaryInputs[0].Time.Time().Equal(stamp) {
		t.Fatalf("binary time = %v", events.BinaryInputs[0].Time)
	}
	if len(events.AnalogInputs) != 1 || events.AnalogInputs[0].Value != -2 || events.AnalogInputs[0].Index != 3 {
		t.Fatalf("analog events %+v", events.AnalogInputs)
	}
	if len(events.Counters) != 1 || events.Counters[0].Value != 0x01020304 || events.Counters[0].Time != nil {
		t.Fatalf("counter events %+v", events.Counters)
	}
	again, err := client.ReadEvents(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(again.BinaryInputs)+len(again.AnalogInputs)+len(again.Counters) != 0 {
		t.Fatalf("second class read returned events: %+v %+v %+v", again.BinaryInputs, again.AnalogInputs, again.Counters)
	}

	got := make(chan *master.UnsolicitedResponse, 1)
	client.SetUnsolicitedHandler(func(resp *master.UnsolicitedResponse) {
		select {
		case got <- resp:
		default:
		}
	})
	watchCtx, watchCancel := context.WithCancel(ctx)
	defer watchCancel()
	watchErr := make(chan error, 1)
	go func() { watchErr <- client.WatchUnsolicited(watchCtx) }()

	if err := client.EnableUnsolicited(ctx); err != nil {
		t.Fatalf("enable: %v", err)
	}
	live := time.Date(2021, 6, 7, 8, 9, 10, 0, time.UTC)
	if err := server.PublishBinaryEvent(9, false, al.BinEventOnline, outstation.EventClass1, live); err != nil {
		t.Fatal(err)
	}

	select {
	case resp := <-got:
		if len(resp.BinaryInputs) != 1 || resp.BinaryInputs[0].Index != 9 || resp.BinaryInputs[0].Value {
			t.Fatalf("unsolicited binary %+v", resp.BinaryInputs)
		}
		if resp.BinaryInputs[0].Time == nil || !resp.BinaryInputs[0].Time.Time().Equal(live) {
			t.Fatalf("unsolicited time = %v, want %v", resp.BinaryInputs[0].Time, live)
		}
	case err := <-watchErr:
		t.Fatalf("watch ended: %v", err)
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for unsolicited event")
	}

	// The confirmed event is gone. A class read must not see it again.
	left, err := client.ReadEvents(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(left.BinaryInputs) != 0 {
		t.Fatalf("event remained after confirm: %+v", left.BinaryInputs)
	}
	watchCancel()
}

// TestTCPReconnectKeepsUnconfirmedEvent enables unsolicited, drops the master
// before that master confirms, and checks that a new connection still reads
// the original event.
func TestTCPReconnectKeepsUnconfirmedEvent(t *testing.T) {
	port := getFreePort(t)
	server, err := outstation.NewServer(outstation.NewConfig(
		outstation.WithAddress(1024),
		outstation.WithMasterAddress(1),
		outstation.WithTransport(dnp3.TCP, "127.0.0.1", port),
		outstation.WithExperimentalEventDelivery(),
	))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := server.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer server.Stop(context.Background())

	stamp := time.Date(2022, 3, 4, 5, 6, 7, int(8*time.Millisecond), time.UTC)
	if err := server.PublishBinaryEvent(11, true, al.BinEventOnline, outstation.EventClass1, stamp); err != nil {
		t.Fatal(err)
	}

	first, err := master.NewClient(master.NewConfig(
		master.WithMasterAddress(1),
		master.WithOutstationAddress(1024),
		master.WithTransport(dnp3.TCP, "127.0.0.1", port),
		master.WithTimeout(3*time.Second),
		master.WithRetry(1, 0),
		master.WithExperimentalEventDelivery(),
	))
	if err != nil {
		t.Fatal(err)
	}
	if err := first.Connect(ctx); err != nil {
		t.Fatal(err)
	}
	if err := first.EnableUnsolicited(ctx); err != nil {
		first.Close()
		t.Fatal(err)
	}
	// No watch, so this master does not confirm the unsolicited fragment.
	first.Close()

	waitUntil(t, 3*time.Second, func() bool {
		counter, ok := server.(interface{ ActiveConnections() int })
		return ok && counter.ActiveConnections() == 0
	})

	second, err := master.NewClient(master.NewConfig(
		master.WithMasterAddress(1),
		master.WithOutstationAddress(1024),
		master.WithTransport(dnp3.TCP, "127.0.0.1", port),
		master.WithTimeout(3*time.Second),
		master.WithRetry(1, 0),
		master.WithExperimentalEventDelivery(),
	))
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	if err := second.Connect(ctx); err != nil {
		t.Fatal(err)
	}
	events, err := second.ReadEvents(ctx)
	if err != nil {
		t.Fatalf("ReadEvents after reconnect: %v", err)
	}
	if len(events.BinaryInputs) != 1 || events.BinaryInputs[0].Index != 11 || !events.BinaryInputs[0].Value {
		t.Fatalf("retained events %+v", events.BinaryInputs)
	}
	if events.BinaryInputs[0].Time == nil || !events.BinaryInputs[0].Time.Time().Equal(stamp) {
		t.Fatalf("retained time = %v", events.BinaryInputs[0].Time)
	}
}

func waitUntil(t *testing.T, d time.Duration, ready func() bool) {
	t.Helper()
	deadline := time.Now().Add(d)
	for time.Now().Before(deadline) {
		if ready() {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("condition not met before deadline")
}
