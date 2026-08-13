// Command master is a minimal DNP3 master client using the v0 MVP public API.
// It connects to one outstation over TCP, performs a Class-0 integrity poll
// (Binary Input G1V1, Counter G20V1, Analog Input G30V1), prints the points,
// issues one Direct-Operate CROB control (Group 12 Variation 1), and closes.
//
// This example is the copy-paste consumer path for the v0 profile (MEXT-031).
// Pair it with the outstation example:
//
//	# terminal 1: start an outstation on 0.0.0.0:20000 (address 1024)
//	go run ./examples/outstation
//	# terminal 2: connect a master to it
//	go run ./examples/master
//
// The defaults match the outstation example. Override via flags:
//
//	go run ./examples/master -host 192.168.1.50 -port 20000 -master 1 -outstation 1024
//
// Press Ctrl+C to stop early.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"dnp3/pkg/dnp3"
	"dnp3/pkg/dnp3/master"
	"dnp3/pkg/dnp3/types"
)

func main() {
	host := flag.String("host", "localhost", "outstation host/IP")
	port := flag.Int("port", 20000, "outstation TCP port")
	masterAddr := flag.Int("master", 1, "this master's DNP3 link address")
	outstationAddr := flag.Int("outstation", 1024, "target outstation's DNP3 link address")
	timeout := flag.Duration("timeout", 5*time.Second, "response timeout")
	flag.Parse()

	cfg := master.NewConfig(
		master.WithMasterAddress(uint16(*masterAddr)),
		master.WithOutstationAddress(uint16(*outstationAddr)),
		master.WithTransport(dnp3.TCP, *host, *port),
		master.WithTimeout(*timeout),
	)

	client, err := master.NewClient(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "master config error: %v\n", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	fmt.Printf("connecting to %s:%d (outstation %d, master %d)...\n", *host, *port, *outstationAddr, *masterAddr)
	if err := client.Connect(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "connect failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("connected")
	defer client.Close()

	// Class-0 integrity poll: one Read of all v0-supported static groups.
	resp, err := client.IntegrityPoll(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "integrity poll failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("IIN: %02X %02X\n", resp.IIN[0], resp.IIN[1])

	for _, b := range resp.BinaryInputs {
		fmt.Printf("  BI[%d] = %v (quality %s)\n", b.Index, b.Value, b.Quality)
	}
	for _, a := range resp.AnalogInputs {
		fmt.Printf("  AI[%d] = %v (quality %s)\n", a.Index, a.Value, a.Quality)
	}
	for _, c := range resp.Counters {
		fmt.Printf("  CT[%d] = %d (quality %s)\n", c.Index, c.Value, c.Quality)
	}

	// Direct-Operate CROB (Group 12 Variation 1): latch binary output 0 ON.
	op, err := client.Operate(ctx, types.NewBinaryControl(0, true, types.DirectOperate))
	if err != nil {
		fmt.Fprintf(os.Stderr, "operate failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("operate status: %d\n", op.Status)

	fmt.Println("done")
}
