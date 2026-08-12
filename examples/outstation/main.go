// Command outstation is a minimal DNP3 outstation server using the v0 MVP
// public API (DNP3-095). It listens on TCP, serves a single master, exposes
// a small set of MVP static points (G1V1 binary input, G20V1 counter, G30V1
// analog input), and accepts Group 12 Variation 1 direct binary control.
//
// This example is build-only (DNP3-095 acceptance: compiles). It is runnable
// for illustration; connect with a DNP3 master to outstation address 1024.
//
// Usage:
//
//	go run ./examples/outstation
//
// Press Ctrl+C to stop.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"dnp3/pkg/dnp3"
	"dnp3/pkg/dnp3/outstation"
	"dnp3/pkg/dnp3/types"
)

// exampleDataHandler provides the MVP static points.
type exampleDataHandler struct{}

func (h *exampleDataHandler) GetBinaryInputs() []*types.BinaryInput {
	return []*types.BinaryInput{
		{Index: 0, Value: true, Quality: types.QualityOnline},
		{Index: 1, Value: false, Quality: types.QualityOnline},
	}
}

func (h *exampleDataHandler) GetAnalogInputs() []*types.AnalogInput {
	return []*types.AnalogInput{
		{Index: 0, Value: 2300, Quality: types.QualityOnline},
	}
}

func (h *exampleDataHandler) GetCounters() []*types.Counter {
	return []*types.Counter{
		{Index: 0, Value: 42, Quality: types.QualityOnline},
	}
}

// The v0 MVP profile serves G1V1/G20V1/G30V1 reads plus G12V1 control; the
// remaining DataHandler methods return empty for the non-MVP groups.
func (h *exampleDataHandler) GetBinaryOutputs() []*types.BinaryOutput { return nil }
func (h *exampleDataHandler) GetAnalogOutputs() []*types.AnalogOutput { return nil }
func (h *exampleDataHandler) GetFrozenCounters() []*types.Counter      { return nil }
func (h *exampleDataHandler) FreezeCounters(clear bool) error          { return nil }

// exampleCommandHandler accepts Group 12 Variation 1 direct binary control
// (the only command profile in v0) and rejects analog commands.
type exampleCommandHandler struct{}

func (h *exampleCommandHandler) HandleBinaryCommand(cmd *types.ControlOutput) (*types.ControlStatus, error) {
	fmt.Printf("binary control: group=%d var=%d index=%d\n", cmd.Group, cmd.Variation, cmd.Index)
	status := types.ControlSuccess
	return &status, nil
}

func (h *exampleCommandHandler) HandleAnalogCommand(cmd *types.ControlOutput) (*types.ControlStatus, error) {
	// Analog output control is outside the v0 MVP profile.
	status := types.ControlNotSupported
	return &status, fmt.Errorf("analog control not supported in v0 MVP profile")
}

func main() {
	cfg := outstation.NewConfig(
		outstation.WithAddress(1024),      // outstation link address
		outstation.WithMasterAddress(1),   // expected master link address
		outstation.WithTransport(dnp3.TCP, "0.0.0.0", 20000),
		outstation.WithMaxConnections(1), // MVP single-master (DNP3-084)
	)

	server, err := outstation.NewServer(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "outstation config error: %v\n", err)
		os.Exit(1)
	}
	server.SetDataHandler(&exampleDataHandler{})
	server.SetCommandHandler(&exampleCommandHandler{})

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := server.Start(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "outstation start error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("DNP3 outstation listening on 0.0.0.0:20000 (address 1024); Ctrl+C to stop")

	<-ctx.Done()
	fmt.Println("shutting down...")
	stopCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := server.Stop(stopCtx); err != nil {
		fmt.Fprintf(os.Stderr, "outstation stop error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("stopped")
}
