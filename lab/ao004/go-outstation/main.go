// AO-004 reverse-interop Go outstation fixture (memory-only test harness).
//
// PURPOSE: a minimal, memory-only DNP3 outstation built on the EXISTING public
// Go outstation API (dnp3/pkg/dnp3/outstation), intended to later, under a
// SEPARATE owner authorization, receive G41V1/V2/V3/V4 single-point Direct
// Operate with response commands from the independent OpenDNP3 3.1.2 C++ master
// harness (lab/ao004/g41_master). It exists only to prepare a reverse
// interoperability target; it is NOT production code and performs NO hardware
// I/O and NO field/NGCP/SCADA integration.
//
// SCOPE / SAFETY:
//   - Default (no arguments) is OFFLINE READINESS ONLY: it constructs no server,
//     opens no TCP socket/listener, starts no outstation session and sends/
//     receives no DNP3 traffic; it prints a bounded readiness summary and exits 0.
//   - The live "--listen" mode is present in source but is NOT authorized by the
//     microtask that authored this file. It refuses any bind other than exactly
//     10.77.0.2:20000 and must only be invoked on an owner-provided isolated
//     runner under a separate protocol-execution authorization.
//   - Link addresses: outstation = 4, remote master = 3.
//   - The command handler is memory-only, permits index 0 only, returns an
//     explicit non-success status on an invalid index, and rejects binary (G12V1)
//     commands explicitly (this fixture is the G41 analog-output target).
//
// The Go outstation already implements the G41V1-V4 Direct Operate path
// (internal/outstation handleDirectOperate + executeControl; public
// WriteAnalogOutput -> HandleAnalogCommand). This fixture only wires the public
// API and stores accepted commands in memory.
package main

import (
	"context"
	"flag"
	"fmt"
	"math"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"dnp3/pkg/dnp3"
	"dnp3/pkg/dnp3/outstation"
	"dnp3/pkg/dnp3/types"
)

// Lab link/addressing constants (mirror lab/vec01 and lab/ao004/g41_master.cpp).
const (
	labOutstationAddress = "10.77.0.2"
	labOutstationPort    = 20000
	linkOutstationAddr   = 4 // this outstation
	linkMasterAddr       = 3 // the C++ master
	commandIndex         = 0 // the only permitted analog-output index
)

// receivedCommand is the memory-only record of one accepted G41 Direct Operate.
type receivedCommand struct {
	Group     uint8
	Variation uint8
	Index     uint16
	Command   types.CommandType
	// Exactly one of these is populated, matching Variation (G41 V1-V4).
	Int32   int32
	Int16   int16
	Float32 float32
	Float64 float64
}

// fixtureCommandHandler is a memory-only outstation CommandHandler. It accepts
// only Group 41 (analog output) Direct Operate with response at index 0 for
// variations 1-4, stores the accepted command in memory, and returns explicit
// non-success statuses otherwise. It never performs OS I/O.
type fixtureCommandHandler struct {
	mu       sync.Mutex
	seen     bool
	last     receivedCommand
	count    uint32
	resetCnt uint32
}

// Snapshot returns a copy of the last accepted command, the total count and
// whether any command was accepted.
func (h *fixtureCommandHandler) Snapshot() (receivedCommand, uint32, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.last, h.count, h.seen
}

// Reset clears the in-memory state (fixed reset semantics for test isolation).
func (h *fixtureCommandHandler) Reset() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.seen = false
	h.last = receivedCommand{}
	h.count = 0
	h.resetCnt++
}

// HandleBinaryCommand rejects G12V1 binary control: this fixture targets the
// Group 41 analog-output path only.
func (h *fixtureCommandHandler) HandleBinaryCommand(cmd *types.ControlOutput) (*types.ControlStatus, error) {
	status := types.ControlNotSupported
	return &status, fmt.Errorf("fixture is G41 analog-output only; binary control (group=%d variation=%d) is not supported", cmd.Group, cmd.Variation)
}

// HandleAnalogCommand accepts a single G41V1-V4 Direct Operate at index 0 and
// stores it in memory; every other case is rejected with an explicit status.
func (h *fixtureCommandHandler) HandleAnalogCommand(cmd *types.ControlOutput) (*types.ControlStatus, error) {
	if cmd.Group != 41 {
		status := types.ControlNotSupported
		return &status, fmt.Errorf("unsupported analog group %d (only Group 41 is supported)", cmd.Group)
	}
	if cmd.Variation < 1 || cmd.Variation > 4 {
		status := types.ControlNotSupported
		return &status, fmt.Errorf("unsupported G41 variation %d (only V1-V4 are supported)", cmd.Variation)
	}
	if cmd.Index != commandIndex {
		// Explicit non-success on an invalid index; the bridge converts a
		// non-success status into an error response (never a SUCCESS echo).
		status := types.ControlOutOfRange
		return &status, fmt.Errorf("index %d out of range (only index %d is permitted)", cmd.Index, commandIndex)
	}

	rec := receivedCommand{Group: cmd.Group, Variation: cmd.Variation, Index: cmd.Index, Command: cmd.CommandType}

	switch v := cmd.Value.(type) {
	case *types.AnalogInt32CommandValue:
		if cmd.Variation != 1 {
			status := types.ControlBadFormat
			return &status, fmt.Errorf("G41V%d received an int32 value", cmd.Variation)
		}
		rec.Int32 = v.Value
	case *types.AnalogInt16CommandValue:
		if cmd.Variation != 2 {
			status := types.ControlBadFormat
			return &status, fmt.Errorf("G41V%d received an int16 value", cmd.Variation)
		}
		rec.Int16 = v.Value
	case *types.AnalogFloat32CommandValue:
		if cmd.Variation != 3 {
			status := types.ControlBadFormat
			return &status, fmt.Errorf("G41V%d received a float32 value", cmd.Variation)
		}
		if math.IsNaN(float64(v.Value)) || math.IsInf(float64(v.Value), 0) {
			status := types.ControlBadFormat
			return &status, fmt.Errorf("G41V3 value must be finite")
		}
		rec.Float32 = v.Value
	case *types.AnalogFloat64CommandValue:
		if cmd.Variation != 4 {
			status := types.ControlBadFormat
			return &status, fmt.Errorf("G41V%d received a float64 value", cmd.Variation)
		}
		if math.IsNaN(v.Value) || math.IsInf(v.Value, 0) {
			status := types.ControlBadFormat
			return &status, fmt.Errorf("G41V4 value must be finite")
		}
		rec.Float64 = v.Value
	default:
		status := types.ControlNotSupported
		return &status, fmt.Errorf("unsupported G41 value carrier %T", cmd.Value)
	}

	h.mu.Lock()
	h.seen = true
	h.last = rec
	h.count++
	count := h.count
	h.mu.Unlock()

	// Bounded, memory-only observability for the authorized interop microtask: one
	// line per accepted command. It records the variation, index and the typed
	// value, and for floats the exact IEEE-754 bit pattern, so an independent
	// observer can confirm the fixture-side effect without any OS I/O.
	fmt.Printf("ao004-go-outstation: accepted group=%d variation=%d index=%d value=%s count=%d\n",
		rec.Group, rec.Variation, rec.Index, formatValue(rec), count)

	status := types.ControlSuccess
	return &status, nil
}

// formatValue renders a stored command's typed value for the acceptance log,
// preserving exact float bits.
func formatValue(rec receivedCommand) string {
	switch rec.Variation {
	case 1:
		return fmt.Sprintf("int32:%d", rec.Int32)
	case 2:
		return fmt.Sprintf("int16:%d", rec.Int16)
	case 3:
		return fmt.Sprintf("float32:%g bits=0x%08x", rec.Float32, math.Float32bits(rec.Float32))
	case 4:
		return fmt.Sprintf("float64:%g bits=0x%016x", rec.Float64, math.Float64bits(rec.Float64))
	}
	return "unknown"
}

// fixtureDataHandler provides the MVP static input points. No OS I/O.
type fixtureDataHandler struct{}

func (h *fixtureDataHandler) GetBinaryInputs() []*types.BinaryInput {
	return []*types.BinaryInput{{Index: 0, Value: true, Quality: types.QualityOnline}}
}

func (h *fixtureDataHandler) GetAnalogInputs() []*types.AnalogInput {
	return []*types.AnalogInput{{Index: 0, Value: 0, Quality: types.QualityOnline}}
}

func (h *fixtureDataHandler) GetCounters() []*types.Counter {
	return []*types.Counter{{Index: 0, Value: 0, Quality: types.QualityOnline}}
}

func (h *fixtureDataHandler) GetBinaryOutputs() []*types.BinaryOutput { return nil }
func (h *fixtureDataHandler) GetAnalogOutputs() []*types.AnalogOutput { return nil }
func (h *fixtureDataHandler) GetFrozenCounters() []*types.Counter     { return nil }
func (h *fixtureDataHandler) FreezeCounters(clear bool) error         { return nil }

// validateListenTarget fails closed unless the bind is exactly the isolated lab
// endpoint. Kept separate from main so it can be unit-tested offline.
func validateListenTarget(host string, port int) error {
	if host != labOutstationAddress || port != labOutstationPort {
		return fmt.Errorf("refusing non-lab listen target %s:%d (only %s:%d is permitted)", host, port, labOutstationAddress, labOutstationPort)
	}
	return nil
}

func printReadiness() {
	fmt.Println("ao004-go-outstation: readiness mode (no server, no socket, no DNP3 traffic).")
	fmt.Printf("  role          : DNP3 outstation (Go, public API)\n")
	fmt.Printf("  link          : outstation=%d <- master=%d\n", linkOutstationAddr, linkMasterAddr)
	fmt.Printf("  target        : %s:%d (isolated lab only)\n", labOutstationAddress, labOutstationPort)
	fmt.Printf("  commands      : G41V1 int32, V2 int16, V3 float32, V4 float64 Direct Operate with response\n")
	fmt.Printf("  index policy  : index %d only; other indices -> OUT_OF_RANGE (explicit non-success)\n", commandIndex)
	fmt.Printf("  binary        : G12V1 rejected explicitly (this fixture is G41 analog-output only)\n")
	fmt.Printf("  state         : memory-only, starts empty (reset clears all state)\n")
	fmt.Printf("  listen mode   : '--listen' binds %s:%d (NOT authorized here)\n", labOutstationAddress, labOutstationPort)
}

func main() {
	var (
		listen = flag.Bool("listen", false, "bind the isolated lab listener (NOT authorized by this microtask)")
		host   = flag.String("host", labOutstationAddress, "bind address (must be the isolated lab address)")
		port   = flag.Int("port", labOutstationPort, "bind port (must be the isolated lab port)")
	)
	flag.Parse()

	if !*listen {
		printReadiness()
		return
	}

	// Live mode: NOT authorized by this microtask. Fail closed on any non-lab
	// target before constructing a server or opening a socket.
	if err := validateListenTarget(*host, *port); err != nil {
		fmt.Fprintf(os.Stderr, "ao004-go-outstation: %v\n", err)
		os.Exit(2)
	}

	handler := &fixtureCommandHandler{}

	cfg := outstation.NewConfig(
		outstation.WithAddress(linkOutstationAddr),
		outstation.WithMasterAddress(linkMasterAddr),
		outstation.WithTransport(dnp3.TCP, *host, *port),
		outstation.WithMaxConnections(1),
		// AO-005: the published v0 supported profile rejects Group 41 analog
		// controls by default. This fixture is the isolated G41 analog-output
		// target, so it explicitly opts into the experimental analog-control
		// path; the default (non-opted-in) behavior is fail-closed rejection.
		outstation.WithExperimentalAnalogControl(),
	)
	server, err := outstation.NewServer(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ao004-go-outstation: config error: %v\n", err)
		os.Exit(1)
	}
	server.SetDataHandler(&fixtureDataHandler{})
	server.SetCommandHandler(handler)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := server.Start(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "ao004-go-outstation: start error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("ao004-go-outstation: listening on %s:%d (link outstation=%d <- master=%d)\n", *host, *port, linkOutstationAddr, linkMasterAddr)

	<-ctx.Done()
	fmt.Println("ao004-go-outstation: shutting down...")
	stopCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := server.Stop(stopCtx); err != nil {
		fmt.Fprintf(os.Stderr, "ao004-go-outstation: stop error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("ao004-go-outstation: stopped")
}
