// AO-005 isolated G41 direction-B harness: a Go master built on the EXISTING
// public Go API (dnp3/pkg/dnp3/master) that drives the independently implemented
// OpenDNP3 3.1.2 C++ outstation peer (lab/ao004/g41_peer) over the isolated lab
// topology. Test harness ONLY; no production code, no hardware I/O.
//
// SCOPE / SAFETY:
//   - Refuses any target other than exactly 10.77.0.2:20000 (fail closed before
//     constructing a client), mirroring the Go outstation fixture and the C++
//     master harness.
//   - Only Direct Operate with response is issued; no SBO, no
//     DirectOperateNoResponse, no implicit retry beyond the bounded retry option.
//   - Link addresses: master = 3, remote outstation = 4.
//   - The opt-in WithExperimentalAnalogControl() switch is required because the
//     published v0 profile rejects Group 41 by default.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"dnp3/pkg/dnp3"
	"dnp3/pkg/dnp3/master"
	"dnp3/pkg/dnp3/types"
)

const (
	labOutstationAddr  = "10.77.0.2"
	labOutstationPort  = 20000
	linkMasterAddr     = 3 // this master
	linkOutstationAddr = 4 // the C++ peer
)

func main() {
	variation := flag.Int("variation", 0, "G41 variation 1..4")
	index := flag.Int("index", 0, "point index 0..1")
	host := flag.String("host", labOutstationAddr, "target address (must be the lab outstation)")
	port := flag.Int("port", labOutstationPort, "target port (must be the lab port)")
	timeout := flag.Duration("timeout", 6*time.Second, "operate timeout")
	flag.Parse()

	if *variation < 1 || *variation > 4 {
		fmt.Fprintln(os.Stderr, "go-master: --variation must be 1..4")
		os.Exit(2)
	}
	if *index < 0 || *index > 1 {
		fmt.Fprintln(os.Stderr, "go-master: --index must be 0..1")
		os.Exit(2)
	}
	if *host != labOutstationAddr || *port != labOutstationPort {
		fmt.Fprintf(os.Stderr, "go-master: refusing non-lab target %s:%d (only %s:%d is permitted)\n",
			*host, *port, labOutstationAddr, labOutstationPort)
		os.Exit(2)
	}

	var cmd *types.ControlOutput
	switch *variation {
	case 1:
		cmd = types.NewAnalogInt32Control(uint16(*index), 2300, types.DirectOperate)
	case 2:
		cmd = types.NewAnalogInt16Control(uint16(*index), 12345, types.DirectOperate)
	case 3:
		cmd = types.NewAnalogFloat32Control(uint16(*index), 1.5, types.DirectOperate)
	case 4:
		cmd = types.NewAnalogFloat64Control(uint16(*index), 1.5, types.DirectOperate)
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout+8*time.Second)
	defer cancel()

	client, err := master.NewClient(master.NewConfig(
		master.WithMasterAddress(linkMasterAddr),
		master.WithOutstationAddress(linkOutstationAddr),
		master.WithTransport(dnp3.TCP, *host, *port),
		master.WithExperimentalAnalogControl(),
		master.WithTimeout(*timeout),
		master.WithRetry(1, 0),
	))
	if err != nil {
		fmt.Fprintf(os.Stderr, "go-master: NewClient: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	if err := client.Connect(ctx); err != nil {
		fmt.Printf("ao005-go-master: variation=%d index=%d connect_error=%v\n", *variation, *index, err)
		os.Exit(1)
	}

	resp, err := client.Operate(ctx, cmd)
	if err != nil {
		fmt.Printf("ao005-go-master: variation=%d index=%d err=%v\n", *variation, *index, err)
		os.Exit(1)
	}
	fmt.Printf("ao005-go-master: variation=%d index=%d status=%v\n", *variation, *index, resp.Status)
}
