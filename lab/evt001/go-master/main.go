// Go master harness for the EVT-001 OpenDNP3 outstation lab.
// It dials 127.0.0.1 only and prints the events it decoded.
package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"dnp3/pkg/dnp3"
	"dnp3/pkg/dnp3/master"
	"dnp3/pkg/dnp3/types"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: go-master <read|empty|unsol|negative> <port> [seconds]")
		os.Exit(2)
	}
	mode := os.Args[1]
	port, err := strconv.Atoi(os.Args[2])
	if err != nil || port <= 0 || port > 65535 {
		fmt.Fprintln(os.Stderr, "bad port")
		os.Exit(2)
	}
	seconds := 5
	if len(os.Args) == 4 {
		seconds, err = strconv.Atoi(os.Args[3])
		if err != nil || seconds <= 0 || seconds > 30 {
			fmt.Fprintln(os.Stderr, "bad seconds")
			os.Exit(2)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(seconds+15)*time.Second)
	defer cancel()

	client, err := master.NewClient(master.NewConfig(
		master.WithMasterAddress(1),
		master.WithOutstationAddress(1024),
		master.WithTransport(dnp3.TCP, "127.0.0.1", port),
		master.WithTimeout(3*time.Second),
		master.WithRetry(1, 0),
		master.WithExperimentalEventDelivery(),
	))
	if err != nil {
		fmt.Println("ERROR", err)
		os.Exit(1)
	}
	defer client.Close()
	if err := client.Connect(ctx); err != nil {
		fmt.Println("ERROR connect", err)
		os.Exit(1)
	}
	fmt.Println("CONNECTED")

	switch mode {
	case "read":
		printClass0(ctx, client)
		printEvents("EVENT", mustEvents(ctx, client))
		printEvents("SECOND", mustEvents(ctx, client))
	case "empty":
		printEvents("EVENT", mustEvents(ctx, client))
	case "unsol":
		waitUnsolicited(ctx, client, time.Duration(seconds)*time.Second, true)
	case "drain":
		waitUnsolicitedDrain(ctx, client, time.Duration(seconds)*time.Second)
	case "negative":
		waitUnsolicited(ctx, client, time.Duration(seconds)*time.Second, false)
		printEvents("EVENT", mustEvents(ctx, client))
	default:
		fmt.Println("ERROR mode")
		os.Exit(2)
	}
	fmt.Println("END")
}

func printClass0(ctx context.Context, client master.Client) {
	resp, err := client.IntegrityPoll(ctx)
	if err != nil {
		fmt.Println("CLASS0 error", err)
		return
	}
	fmt.Printf("CLASS0 binaries=%d analogs=%d counters=%d iin=%02x%02x\n",
		len(resp.BinaryInputs), len(resp.AnalogInputs), len(resp.Counters), resp.IIN[0], resp.IIN[1])
}

func mustEvents(ctx context.Context, client master.Client) *master.UnsolicitedResponse {
	resp, err := client.ReadEvents(ctx)
	if err != nil {
		fmt.Println("ERROR read", err)
		os.Exit(1)
	}
	return resp
}

func printEvents(tag string, resp *master.UnsolicitedResponse) {
	if resp == nil {
		fmt.Printf("%s none\n", tag)
		return
	}
	fmt.Printf("%s iin=%02x%02x\n", tag, resp.IIN[0], resp.IIN[1])
	for _, p := range resp.BinaryInputs {
		fmt.Printf("%s kind=bin idx=%d value=%t quality=%02x time=%s\n", tag, p.Index, p.Value, uint8(p.Quality), fmtTime(p.Time))
	}
	for _, p := range resp.AnalogInputs {
		fmt.Printf("%s kind=ana idx=%d value=%.0f quality=%02x time=%s\n", tag, p.Index, p.Value, uint8(p.Quality), fmtTime(p.Time))
	}
	for _, p := range resp.Counters {
		fmt.Printf("%s kind=ctr idx=%d value=%d quality=%02x time=%s\n", tag, p.Index, p.Value, uint8(p.Quality), fmtTime(p.Time))
	}
	if len(resp.BinaryInputs)+len(resp.AnalogInputs)+len(resp.Counters) == 0 {
		fmt.Printf("%s count=0\n", tag)
	}
}

func fmtTime(ts *types.Timestamp) string {
	if ts == nil {
		return "none"
	}
	return ts.Time().UTC().Format(time.RFC3339Nano)
}

func waitUnsolicited(ctx context.Context, client master.Client, wait time.Duration, enable bool) {
	got := make(chan *master.UnsolicitedResponse, 8)
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
	time.Sleep(200 * time.Millisecond)
	if enable {
		if err := client.EnableUnsolicited(ctx); err != nil {
			fmt.Println("ERROR enable", err)
			os.Exit(1)
		}
		fmt.Println("ENABLED")
	} else {
		fmt.Println("WATCHING")
	}
	timer := time.NewTimer(wait)
	defer timer.Stop()
	for {
		select {
		case resp := <-got:
			printEvents("UNSOL", resp)
			if enable {
				watchCancel()
				return
			}
		case <-timer.C:
			fmt.Println("WAIT done")
			watchCancel()
			return
		case err := <-watchErr:
			if watchCtx.Err() != nil {
				return
			}
			fmt.Println("ERROR watch", err)
			os.Exit(1)
		}
	}
}

// waitUnsolicitedDrain enables unsolicited reporting and collects every
// unsolicited fragment delivered within the wait window, across successive
// confirm cycles. Unlike waitUnsolicited it does not return after the first
// response, so it can prove the master drains a batch larger than one fragment.
func waitUnsolicitedDrain(ctx context.Context, client master.Client, wait time.Duration) {
	got := make(chan *master.UnsolicitedResponse, 64)
	client.SetUnsolicitedHandler(func(resp *master.UnsolicitedResponse) {
		got <- resp
	})
	watchCtx, watchCancel := context.WithCancel(ctx)
	defer watchCancel()
	watchErr := make(chan error, 1)
	go func() { watchErr <- client.WatchUnsolicited(watchCtx) }()
	time.Sleep(200 * time.Millisecond)
	if err := client.EnableUnsolicited(ctx); err != nil {
		fmt.Println("ERROR enable", err)
		os.Exit(1)
	}
	fmt.Println("ENABLED")
	timer := time.NewTimer(wait)
	defer timer.Stop()
	for {
		select {
		case resp := <-got:
			printEvents("UNSOL", resp)
		case <-timer.C:
			fmt.Println("WAIT done")
			watchCancel()
			return
		case err := <-watchErr:
			if watchCtx.Err() != nil {
				return
			}
			fmt.Println("ERROR watch", err)
			os.Exit(1)
		}
	}
}
