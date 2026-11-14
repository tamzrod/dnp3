// AO-005 negative-interop master harness (LAB-ONLY test harness).
//
// PURPOSE: drive the EXISTING public Go master client (dnp3/pkg/dnp3/master)
// through the bounded negative matrix B scenarios and report the classified
// outcome, so the after-send UNKNOWN / fail-closed / no-blind-retry behavior can
// be observed deterministically against the isolated OpenDNP3 3.1.2 peer.
//
// Modes:
//
//	pass              — one normal DirectOperate (control)
//	presend-cancel    — Connect, then Operate with an already-cancelled context
//	                    (definite pre-send failure; expect zero handler writes)
//	presend-notconn   — Operate before Connect (definite pre-send failure)
//	after-send        — Connect (through the fault proxy), Operate once, then
//	                    Operate a second time to observe the quarantine gate
//
// SAFETY / SCOPE: refuses any target other than exactly 10.77.0.1:20000 (the
// master-side lab address / fault proxy). Test harness ONLY: no production code,
// no hardware I/O, no field/NGCP/SCADA endpoint.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"dnp3/pkg/dnp3"
	"dnp3/pkg/dnp3/master"
	"dnp3/pkg/dnp3/types"
)

const (
	labTargetAddr = "10.77.0.1"
	labTargetPort = 20000
	linkMaster    = 3
	linkOutstn    = 4
)

func main() {
	mode := flag.String("mode", "pass", "pass|presend-cancel|presend-notconn|after-send")
	host := flag.String("host", labTargetAddr, "target address (must be the lab proxy/master address)")
	port := flag.Int("port", labTargetPort, "target port (must be the lab port)")
	timeout := flag.Duration("timeout", 3*time.Second, "operate timeout")
	flag.Parse()

	if *host != labTargetAddr || *port != labTargetPort {
		fmt.Fprintf(os.Stderr, "neg-master: refusing non-lab target %s:%d (only %s:%d is permitted)\n", *host, *port, labTargetAddr, labTargetPort)
		os.Exit(2)
	}
	fmt.Printf("neg-master: mode=%s target=%s:%d\n", *mode, *host, *port)

	if *mode == "presend-notconn" {
		client, err := newClient(*host, *port, *timeout)
		if err != nil {
			fmt.Fprintf(os.Stderr, "neg-master: NewClient: %v\n", err)
			os.Exit(1)
		}
		defer client.Close()
		_, oerr := client.Operate(context.Background(), types.NewAnalogInt32Control(0, 2300, types.DirectOperate))
		report("operate", oerr)
		return
	}

	client, err := newClient(*host, *port, *timeout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "neg-master: NewClient: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), *timeout+6*time.Second)
	defer cancel()
	if cerr := client.Connect(ctx); cerr != nil {
		fmt.Printf("neg-master: connect_error=%v\n", cerr)
		os.Exit(1)
	}
	fmt.Println("neg-master: connected")

	switch *mode {
	case "pass":
		resp, oerr := client.Operate(ctx, types.NewAnalogInt32Control(0, 2300, types.DirectOperate))
		if oerr != nil {
			report("operate", oerr)
			return
		}
		fmt.Printf("neg-master: operate_status=%v\n", resp.Status)
	case "presend-cancel":
		cctx, ccancel := context.WithCancel(ctx)
		ccancel() // already cancelled before Operate
		_, oerr := client.Operate(cctx, types.NewAnalogInt32Control(0, 2300, types.DirectOperate))
		report("operate1(cancelled)", oerr)
		// A definite pre-send failure must not quarantine: a subsequent valid
		// Operate must be admitted (not rejected as quarantined).
		resp, oerr2 := client.Operate(ctx, types.NewAnalogInt32Control(0, 2300, types.DirectOperate))
		if oerr2 != nil {
			report("operate2(after-cancel)", oerr2)
		} else {
			fmt.Printf("neg-master: operate2(after-cancel)_status=%v (not quarantined)\n", resp.Status)
		}
	case "after-send":
		_, oerr := client.Operate(ctx, types.NewAnalogInt32Control(0, 2300, types.DirectOperate))
		report("operate1(after-send)", oerr)
		// The outstation must now be quarantined: a second control is refused and
		// must NOT be executed a second time.
		_, oerr2 := client.Operate(ctx, types.NewAnalogInt32Control(0, 2300, types.DirectOperate))
		report("operate2(quarantine-probe)", oerr2)
	default:
		fmt.Fprintf(os.Stderr, "neg-master: unknown mode %q\n", *mode)
		os.Exit(2)
	}
}

func newClient(host string, port int, timeout time.Duration) (master.Client, error) {
	return master.NewClient(master.NewConfig(
		master.WithMasterAddress(linkMaster),
		master.WithOutstationAddress(linkOutstn),
		master.WithTransport(dnp3.TCP, host, port),
		master.WithExperimentalAnalogControl(),
		master.WithTimeout(timeout),
		master.WithRetry(1, 0),
	))
}

// report prints the error, its public classification, and whether it carries an
// ambiguous/UNKNOWN or quarantine signal.
func report(label string, err error) {
	if err == nil {
		fmt.Printf("neg-master: %s_err=NONE (unexpected)\n", label)
		return
	}
	code := dnp3.ClassifyError(err)
	low := strings.ToLower(err.Error())
	unknown := strings.Contains(low, "unknown") || strings.Contains(low, "quarantin")
	fmt.Printf("neg-master: %s_err=%q classify=%s ambiguous_or_quarantine=%v\n", label, err.Error(), code, unknown)
}
